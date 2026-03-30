package deluge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"sync/atomic"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "goenvoy/0.0.1"
)

// Option configures a [Client].
type Option func(*Client)

// WithHTTPClient sets a custom [http.Client]. A cookie jar will be added
// if the provided client does not already have one.
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

// Client is a Deluge JSON-RPC client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
	requestID  atomic.Int64
}

// New creates a Deluge [Client] for the given base URL.
func New(baseURL string, opts ...Option) *Client {
	jar, _ := cookiejar.New(nil)
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
			Jar:     jar,
		},
		userAgent: defaultUserAgent,
	}
	for _, o := range opts {
		o(c)
	}
	if c.httpClient.Jar == nil {
		c.httpClient.Jar = jar
	}
	return c
}

// APIError is returned when Deluge returns a JSON-RPC error.
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("deluge: %s (code %d)", e.Message, e.Code)
}

func (c *Client) call(ctx context.Context, method string, params []any, dst any) error {
	id := int(c.requestID.Add(1))
	reqBody := rpcRequest{
		ID:     id,
		Method: method,
		Params: params,
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("deluge: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/json", bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("deluge: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("deluge: POST /json: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("deluge: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("deluge: decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return &APIError{Code: rpcResp.Error.Code, Message: rpcResp.Error.Message}
	}

	if dst != nil && len(rpcResp.Result) > 0 {
		if err := json.Unmarshal(rpcResp.Result, dst); err != nil {
			return fmt.Errorf("deluge: decode result: %w", err)
		}
	}
	return nil
}

// Login authenticates with the Deluge web UI.
func (c *Client) Login(ctx context.Context, password string) error {
	var ok bool
	if err := c.call(ctx, "auth.login", []any{password}, &ok); err != nil {
		return err
	}
	if !ok {
		return &APIError{Code: 0, Message: "authentication failed"}
	}
	return nil
}

// Connected checks if the Deluge web UI is connected to a daemon.
func (c *Client) Connected(ctx context.Context) (bool, error) {
	var connected bool
	if err := c.call(ctx, "web.connected", []any{}, &connected); err != nil {
		return false, err
	}
	return connected, nil
}

// GetTorrentsStatus returns all torrents matching the given filter.
// If filter is nil, all torrents are returned.
func (c *Client) GetTorrentsStatus(ctx context.Context, filter map[string]string) (map[string]*Torrent, error) {
	if filter == nil {
		filter = map[string]string{}
	}
	var result map[string]*Torrent
	if err := c.call(ctx, "core.get_torrents_status", []any{filter, defaultTorrentFields}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetTorrentStatus returns the status of a single torrent by hash.
func (c *Client) GetTorrentStatus(ctx context.Context, hash string) (*Torrent, error) {
	var result Torrent
	if err := c.call(ctx, "core.get_torrent_status", []any{hash, defaultTorrentFields}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddTorrentURL adds a torrent by URL or magnet link.
func (c *Client) AddTorrentURL(ctx context.Context, url string, options map[string]any) (string, error) {
	if options == nil {
		options = map[string]any{}
	}
	var hash string
	if err := c.call(ctx, "core.add_torrent_url", []any{url, options}, &hash); err != nil {
		return "", err
	}
	return hash, nil
}

// RemoveTorrent removes a torrent by hash.
// If removeData is true, downloaded files are also deleted.
func (c *Client) RemoveTorrent(ctx context.Context, hash string, removeData bool) error {
	var ok bool
	if err := c.call(ctx, "core.remove_torrent", []any{hash, removeData}, &ok); err != nil {
		return err
	}
	return nil
}

// PauseTorrent pauses a torrent by hash.
func (c *Client) PauseTorrent(ctx context.Context, hash string) error {
	return c.call(ctx, "core.pause_torrent", []any{hash}, nil)
}

// ResumeTorrent resumes a torrent by hash.
func (c *Client) ResumeTorrent(ctx context.Context, hash string) error {
	return c.call(ctx, "core.resume_torrent", []any{hash}, nil)
}

// PauseAll pauses all torrents.
func (c *Client) PauseAll(ctx context.Context) error {
	return c.call(ctx, "core.pause_all_torrents", []any{}, nil)
}

// ResumeAll resumes all torrents.
func (c *Client) ResumeAll(ctx context.Context) error {
	return c.call(ctx, "core.resume_all_torrents", []any{}, nil)
}

// ForceRecheck rechecks a torrent by hash.
func (c *Client) ForceRecheck(ctx context.Context, hashes []string) error {
	return c.call(ctx, "core.force_recheck", []any{hashes}, nil)
}

// SetTorrentLabel sets a label on a torrent.
func (c *Client) SetTorrentLabel(ctx context.Context, hash, label string) error {
	return c.call(ctx, "label.set_torrent", []any{hash, label}, nil)
}

// MoveTorrent moves torrent data to a new path.
func (c *Client) MoveTorrent(ctx context.Context, hashes []string, dest string) error {
	return c.call(ctx, "core.move_storage", []any{hashes, dest}, nil)
}

// GetSessionStatus returns current session statistics.
func (c *Client) GetSessionStatus(ctx context.Context) (*SessionStatus, error) {
	keys := []string{
		"payload_download_rate", "payload_upload_rate", "dht_nodes",
		"has_incoming_connections", "total_payload_download", "total_payload_upload",
	}
	var result SessionStatus
	if err := c.call(ctx, "core.get_session_status", []any{keys}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFreeSpace returns free disk space in bytes at the default download location.
func (c *Client) GetFreeSpace(ctx context.Context, path string) (int64, error) {
	var freeSpace int64
	if err := c.call(ctx, "core.get_free_space", []any{path}, &freeSpace); err != nil {
		return 0, err
	}
	return freeSpace, nil
}

// GetVersion returns the Deluge daemon version string.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	var version string
	if err := c.call(ctx, "daemon.info", []any{}, &version); err != nil {
		return "", err
	}
	return version, nil
}
