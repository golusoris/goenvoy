package nzbget

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "goenvoy/0.0.1"
)

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

// Client is an NZBGet JSON-RPC client.
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	userAgent  string
	requestID  atomic.Int64
}

// New creates an NZBGet [Client] for the given base URL with basic auth credentials.
func New(baseURL, username, password string, opts ...Option) *Client {
	c := &Client{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: &http.Client{Timeout: defaultTimeout},
		userAgent:  defaultUserAgent,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// APIError is returned when NZBGet returns a JSON-RPC error.
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("nzbget: %s (code %d)", e.Message, e.Code)
}

func (c *Client) call(ctx context.Context, method string, params []any, dst any) error {
	id := int(c.requestID.Add(1))
	reqBody := rpcRequest{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("nzbget: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/jsonrpc", bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("nzbget: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nzbget: POST /jsonrpc: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nzbget: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("nzbget: decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return &APIError{Code: rpcResp.Error.Code, Message: rpcResp.Error.Message}
	}

	if dst != nil && len(rpcResp.Result) > 0 {
		if err := json.Unmarshal(rpcResp.Result, dst); err != nil {
			return fmt.Errorf("nzbget: decode result: %w", err)
		}
	}
	return nil
}

// Queue methods.

// ListGroups returns the current download queue.
func (c *Client) ListGroups(ctx context.Context) ([]Group, error) {
	var groups []Group
	if err := c.call(ctx, "listgroups", nil, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// Append adds an NZB by URL to the download queue.
// Returns the NZB ID assigned by NZBGet.
func (c *Client) Append(ctx context.Context, nzbFilename, nzbURL, category string, priority int) (int, error) {
	params := []any{
		nzbFilename, // NZBFilename
		nzbURL,      // NZBContent (URL when not base64)
		category,    // Category
		priority,    // Priority
		false,       // AddToTop
		false,       // AddPaused
		"",          // DupeKey
		0,           // DupeScore
		"score",     // DupeMode
		[]any{},     // PPParameters
	}
	var id int
	if err := c.call(ctx, "append", params, &id); err != nil {
		return 0, err
	}
	return id, nil
}

// EditQueue modifies items in the queue.
// Common commands: GroupPause, GroupResume, GroupDelete, GroupMoveOffset, GroupSetCategory.
func (c *Client) EditQueue(ctx context.Context, command, param string, ids []int) (bool, error) {
	var ok bool
	if err := c.call(ctx, "editqueue", []any{command, param, ids}, &ok); err != nil {
		return false, err
	}
	return ok, nil
}

// PauseDownload pauses the download queue.
func (c *Client) PauseDownload(ctx context.Context) error {
	var ok bool
	if err := c.call(ctx, "pausedownload", nil, &ok); err != nil {
		return err
	}
	return nil
}

// ResumeDownload resumes the download queue.
func (c *Client) ResumeDownload(ctx context.Context) error {
	var ok bool
	if err := c.call(ctx, "resumedownload", nil, &ok); err != nil {
		return err
	}
	return nil
}

// SetDownloadRate sets the download speed limit in KiB/s. Pass 0 to remove the limit.
func (c *Client) SetDownloadRate(ctx context.Context, limitKBs int) error {
	var ok bool
	if err := c.call(ctx, "rate", []any{limitKBs}, &ok); err != nil {
		return err
	}
	return nil
}

// History methods.

// GetHistory returns the download history.
// If hidden is true, hidden items are included.
func (c *Client) GetHistory(ctx context.Context, hidden bool) ([]HistoryItem, error) {
	var items []HistoryItem
	if err := c.call(ctx, "history", []any{hidden}, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// Server methods.

// GetStatus returns the current server status.
func (c *Client) GetStatus(ctx context.Context) (*Status, error) {
	var status Status
	if err := c.call(ctx, "status", nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// GetVersion returns the NZBGet version string.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	var version string
	if err := c.call(ctx, "version", nil, &version); err != nil {
		return "", err
	}
	return version, nil
}

// GetConfig returns the NZBGet configuration.
func (c *Client) GetConfig(ctx context.Context) ([]ConfigEntry, error) {
	var entries []ConfigEntry
	if err := c.call(ctx, "config", nil, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// GetLog returns log entries. Start is the first entry ID, count is the max entries to return.
func (c *Client) GetLog(ctx context.Context, start, count int) ([]LogEntry, error) {
	var entries []LogEntry
	if err := c.call(ctx, "log", []any{start, count}, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// Reload forces NZBGet to reload its configuration.
func (c *Client) Reload(ctx context.Context) error {
	var ok bool
	if err := c.call(ctx, "reload", nil, &ok); err != nil {
		return err
	}
	return nil
}

// ScanNZBDir scans the incoming NZB directory.
func (c *Client) ScanNZBDir(ctx context.Context) error {
	var ok bool
	if err := c.call(ctx, "scan", nil, &ok); err != nil {
		return err
	}
	return nil
}

// SetCategory sets the category for a download.
func (c *Client) SetCategory(ctx context.Context, id int, category string) (bool, error) {
	return c.EditQueue(ctx, "GroupSetCategory", category, []int{id})
}

// SetPriority sets the priority for a download.
func (c *Client) SetPriority(ctx context.Context, id, priority int) (bool, error) {
	return c.EditQueue(ctx, "GroupSetPriority", strconv.Itoa(priority), []int{id})
}
