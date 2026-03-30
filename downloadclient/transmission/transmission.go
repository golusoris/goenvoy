package transmission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "goenvoy/0.0.1"
	defaultRPCPath   = "/transmission/rpc"
	sessionHeader    = "X-Transmission-Session-Id"
)

// defaultTorrentFields are the fields requested for torrent queries.
var defaultTorrentFields = []string{
	"id", "name", "hashString", "status", "error", "errorString",
	"addedDate", "doneDate", "eta", "isFinished", "isStalled",
	"percentDone", "seedRatioMode", "seedRatioLimit",
	"totalSize", "sizeWhenDone", "leftUntilDone",
	"downloadedEver", "uploadedEver", "uploadRatio",
	"rateDownload", "rateUpload", "downloadDir",
	"peersConnected", "peersSendingToUs", "peersGettingFromUs",
	"queuePosition", "labels", "trackers", "files", "fileStats",
}

// Option configures a [Client].
type Option func(*Client)

// WithHTTPClient sets a custom [http.Client].
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.httpClient = c }
}

// WithTimeout overrides the default HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(cl *Client) { cl.httpClient.Timeout = d }
}

// WithUserAgent sets the User-Agent header sent with every request.
func WithUserAgent(ua string) Option {
	return func(cl *Client) { cl.userAgent = ua }
}

// WithRPCPath overrides the default RPC endpoint path.
func WithRPCPath(path string) Option {
	return func(cl *Client) { cl.rpcPath = path }
}

// WithAuth sets HTTP basic authentication credentials.
func WithAuth(username, password string) Option {
	return func(cl *Client) {
		cl.username = username
		cl.password = password
	}
}

// Client is a Transmission RPC client.
type Client struct {
	baseURL    string
	rpcPath    string
	httpClient *http.Client
	userAgent  string
	username   string
	password   string
	mu         sync.Mutex
	sessionID  string
}

// New creates a Transmission [Client] for the given base URL.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:    baseURL,
		rpcPath:    defaultRPCPath,
		httpClient: &http.Client{Timeout: defaultTimeout},
		userAgent:  defaultUserAgent,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// APIError is returned when the RPC response result is not "success".
type APIError struct {
	Result string
}

func (e *APIError) Error() string {
	return "transmission: " + e.Result
}

// HTTPError is returned when the server responds with a non-2xx/non-409 status.
type HTTPError struct {
	StatusCode int
	RawBody    string
}

func (e *HTTPError) Error() string {
	if e.RawBody != "" {
		return fmt.Sprintf("transmission: HTTP %d: %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("transmission: HTTP %d", e.StatusCode)
}

func (c *Client) call(ctx context.Context, method string, args, dst any) error {
	reqBody := rpcRequest{Method: method, Arguments: args}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("transmission: marshal request: %w", err)
	}

	// Try up to two times: first attempt may get 409 with new session ID.
	for range 2 {
		resp, err := c.doHTTP(ctx, buf)
		if err != nil {
			return err
		}

		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("transmission: read response: %w", readErr)
		}

		if resp.StatusCode == http.StatusConflict {
			// Update session ID and retry.
			c.mu.Lock()
			c.sessionID = resp.Header.Get(sessionHeader)
			c.mu.Unlock()
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return &HTTPError{StatusCode: resp.StatusCode, RawBody: string(respBody)}
		}

		var rpcResp rpcResponse
		if err := json.Unmarshal(respBody, &rpcResp); err != nil {
			return fmt.Errorf("transmission: decode response: %w", err)
		}

		if rpcResp.Result != "success" {
			return &APIError{Result: rpcResp.Result}
		}

		if dst != nil && len(rpcResp.Arguments) > 0 {
			if err := json.Unmarshal(rpcResp.Arguments, dst); err != nil {
				return fmt.Errorf("transmission: decode arguments: %w", err)
			}
		}
		return nil
	}
	return &HTTPError{StatusCode: http.StatusConflict, RawBody: "session ID negotiation failed"}
}

func (c *Client) doHTTP(ctx context.Context, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+c.rpcPath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("transmission: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	c.mu.Lock()
	if c.sessionID != "" {
		req.Header.Set(sessionHeader, c.sessionID)
	}
	c.mu.Unlock()

	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transmission: POST %s: %w", c.rpcPath, err)
	}
	return resp, nil
}

// Torrent methods.

// GetTorrents returns all torrents, or specific torrents if IDs are provided.
func (c *Client) GetTorrents(ctx context.Context, ids []int) ([]Torrent, error) {
	args := map[string]any{"fields": defaultTorrentFields}
	if len(ids) > 0 {
		args["ids"] = ids
	}
	var result struct {
		Torrents []Torrent `json:"torrents"`
	}
	if err := c.call(ctx, "torrent-get", args, &result); err != nil {
		return nil, err
	}
	return result.Torrents, nil
}

// AddTorrentURL adds a torrent by URL or magnet link.
func (c *Client) AddTorrentURL(ctx context.Context, url string, opts *AddTorrentOptions) (*AddTorrentResult, error) {
	args := map[string]any{"filename": url}
	if opts != nil {
		if opts.DownloadDir != "" {
			args["download-dir"] = opts.DownloadDir
		}
		if opts.Paused {
			args["paused"] = true
		}
		if opts.PeerLimit > 0 {
			args["peer-limit"] = opts.PeerLimit
		}
		if opts.BandwidthPriority != 0 {
			args["bandwidthPriority"] = opts.BandwidthPriority
		}
		if len(opts.Labels) > 0 {
			args["labels"] = opts.Labels
		}
	}
	var result struct {
		Added     *AddTorrentResult `json:"torrent-added"`
		Duplicate *AddTorrentResult `json:"torrent-duplicate"`
	}
	if err := c.call(ctx, "torrent-add", args, &result); err != nil {
		return nil, err
	}
	if result.Added != nil {
		return result.Added, nil
	}
	return result.Duplicate, nil
}

// StartTorrents starts (resumes) torrents by their IDs.
func (c *Client) StartTorrents(ctx context.Context, ids []int) error {
	return c.call(ctx, "torrent-start", map[string]any{"ids": ids}, nil)
}

// StopTorrents stops (pauses) torrents by their IDs.
func (c *Client) StopTorrents(ctx context.Context, ids []int) error {
	return c.call(ctx, "torrent-stop", map[string]any{"ids": ids}, nil)
}

// RemoveTorrents removes torrents by their IDs.
// If deleteLocalData is true, downloaded data is also deleted.
func (c *Client) RemoveTorrents(ctx context.Context, ids []int, deleteLocalData bool) error {
	args := map[string]any{
		"ids":               ids,
		"delete-local-data": deleteLocalData,
	}
	return c.call(ctx, "torrent-remove", args, nil)
}

// VerifyTorrents verifies torrents by their IDs.
func (c *Client) VerifyTorrents(ctx context.Context, ids []int) error {
	return c.call(ctx, "torrent-verify", map[string]any{"ids": ids}, nil)
}

// ReannounceTorrents reannounces torrents to their trackers.
func (c *Client) ReannounceTorrents(ctx context.Context, ids []int) error {
	return c.call(ctx, "torrent-reannounce", map[string]any{"ids": ids}, nil)
}

// MoveTorrents moves torrent data to a new location.
func (c *Client) MoveTorrents(ctx context.Context, ids []int, location string, move bool) error {
	args := map[string]any{
		"ids":      ids,
		"location": location,
		"move":     move,
	}
	return c.call(ctx, "torrent-set-location", args, nil)
}

// SetTorrentLabels sets labels on torrents.
func (c *Client) SetTorrentLabels(ctx context.Context, ids []int, labels []string) error {
	args := map[string]any{
		"ids":    ids,
		"labels": labels,
	}
	return c.call(ctx, "torrent-set", args, nil)
}

// SetTorrentSpeedLimits sets per-torrent speed limits in KiB/s.
// A value of -1 disables the limit.
func (c *Client) SetTorrentSpeedLimits(ctx context.Context, ids []int, dlLimit, ulLimit int64) error {
	args := map[string]any{"ids": ids}
	if dlLimit >= 0 {
		args["downloadLimit"] = dlLimit
		args["downloadLimited"] = true
	} else {
		args["downloadLimited"] = false
	}
	if ulLimit >= 0 {
		args["uploadLimit"] = ulLimit
		args["uploadLimited"] = true
	} else {
		args["uploadLimited"] = false
	}
	return c.call(ctx, "torrent-set", args, nil)
}

// Session methods.

// GetSession returns the current Transmission session configuration.
func (c *Client) GetSession(ctx context.Context) (*Session, error) {
	var out Session
	if err := c.call(ctx, "session-get", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSessionStats returns current session statistics.
func (c *Client) GetSessionStats(ctx context.Context) (*SessionStats, error) {
	var out SessionStats
	if err := c.call(ctx, "session-stats", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetFreeSpace returns the amount of free space at the given path.
func (c *Client) GetFreeSpace(ctx context.Context, path string) (*FreeSpace, error) {
	var out FreeSpace
	if err := c.call(ctx, "free-space", map[string]any{"path": path}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TestPort checks if the incoming peer port is accessible.
func (c *Client) TestPort(ctx context.Context) (bool, error) {
	var result struct {
		PortIsOpen bool `json:"port-is-open"`
	}
	if err := c.call(ctx, "port-test", nil, &result); err != nil {
		return false, err
	}
	return result.PortIsOpen, nil
}
