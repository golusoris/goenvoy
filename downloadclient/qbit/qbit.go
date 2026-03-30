package qbit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "goenvoy/0.0.1"
)

// Option configures a [Client].
type Option func(*Client)

// WithHTTPClient sets a custom [http.Client].
// The client must have a cookie jar configured for session management.
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

// Client is a qBittorrent WebUI API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// New creates a qBittorrent [Client] for the given base URL (e.g. "http://localhost:8080").
func New(baseURL string, opts ...Option) *Client {
	jar, _ := cookiejar.New(nil)
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: defaultTimeout, Jar: jar},
		userAgent:  defaultUserAgent,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// APIError is returned when the API responds with a non-2xx status.
type APIError struct {
	StatusCode int    `json:"-"`
	RawBody    string `json:"-"`
}

func (e *APIError) Error() string {
	if e.RawBody != "" {
		return fmt.Sprintf("qbit: HTTP %d: %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("qbit: HTTP %d", e.StatusCode)
}

func (c *Client) doRequest(ctx context.Context, method, path string, form url.Values, dst any) error {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("qbit: create request: %w", err)
	}

	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("User-Agent", c.userAgent)
	// Referer header is required by qBittorrent for CSRF protection.
	req.Header.Set("Referer", c.baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("qbit: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("qbit: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, RawBody: string(respBody)}
	}

	if dst != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return fmt.Errorf("qbit: decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, dst any) error {
	if params != nil {
		path += "?" + params.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, dst)
}

func (c *Client) post(ctx context.Context, path string, form url.Values) error {
	return c.doRequest(ctx, http.MethodPost, path, form, nil)
}

func (c *Client) getRaw(ctx context.Context, path string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("qbit: create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("qbit: GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("qbit: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &APIError{StatusCode: resp.StatusCode, RawBody: string(buf)}
	}
	return string(buf), nil
}

// Authentication.

// Login authenticates with the qBittorrent WebUI.
// The session cookie (SID) is automatically stored in the client's cookie jar.
func (c *Client) Login(ctx context.Context, username, password string) error {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	return c.post(ctx, "/api/v2/auth/login", form)
}

// Logout ends the current session.
func (c *Client) Logout(ctx context.Context) error {
	return c.post(ctx, "/api/v2/auth/logout", nil)
}

// Application.

// Version returns the qBittorrent application version string.
func (c *Client) Version(ctx context.Context) (string, error) {
	return c.getRaw(ctx, "/api/v2/app/version")
}

// WebAPIVersion returns the WebUI API version string.
func (c *Client) WebAPIVersion(ctx context.Context) (string, error) {
	return c.getRaw(ctx, "/api/v2/app/webapiVersion")
}

// GetBuildInfo returns qBittorrent build information.
func (c *Client) GetBuildInfo(ctx context.Context) (*BuildInfo, error) {
	var out BuildInfo
	if err := c.get(ctx, "/api/v2/app/buildInfo", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPreferences returns the current qBittorrent preferences.
func (c *Client) GetPreferences(ctx context.Context) (*Preferences, error) {
	var out Preferences
	if err := c.get(ctx, "/api/v2/app/preferences", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DefaultSavePath returns the default save path for downloads.
func (c *Client) DefaultSavePath(ctx context.Context) (string, error) {
	return c.getRaw(ctx, "/api/v2/app/defaultSavePath")
}

// Torrents.

// ListTorrents returns a list of torrents matching the given options.
// Pass nil for opts to list all torrents.
func (c *Client) ListTorrents(ctx context.Context, opts *ListOptions) ([]Torrent, error) {
	p := url.Values{}
	if opts != nil {
		if opts.Filter != "" {
			p.Set("filter", opts.Filter)
		}
		if opts.Category != "" {
			p.Set("category", opts.Category)
		}
		if opts.Tag != "" {
			p.Set("tag", opts.Tag)
		}
		if opts.Sort != "" {
			p.Set("sort", opts.Sort)
		}
		if opts.Reverse {
			p.Set("reverse", "true")
		}
		if opts.Limit > 0 {
			p.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Offset > 0 {
			p.Set("offset", strconv.Itoa(opts.Offset))
		}
		if opts.Hashes != "" {
			p.Set("hashes", opts.Hashes)
		}
	}
	var out []Torrent
	if err := c.get(ctx, "/api/v2/torrents/info", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTorrentProperties returns detailed properties for a specific torrent.
func (c *Client) GetTorrentProperties(ctx context.Context, hash string) (*TorrentProperties, error) {
	var out TorrentProperties
	p := url.Values{}
	p.Set("hash", hash)
	if err := c.get(ctx, "/api/v2/torrents/properties", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTorrentTrackers returns trackers for a specific torrent.
func (c *Client) GetTorrentTrackers(ctx context.Context, hash string) ([]Tracker, error) {
	var out []Tracker
	p := url.Values{}
	p.Set("hash", hash)
	if err := c.get(ctx, "/api/v2/torrents/trackers", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTorrentWebSeeds returns web seeds for a specific torrent.
func (c *Client) GetTorrentWebSeeds(ctx context.Context, hash string) ([]WebSeed, error) {
	var out []WebSeed
	p := url.Values{}
	p.Set("hash", hash)
	if err := c.get(ctx, "/api/v2/torrents/webseeds", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTorrentFiles returns files within a specific torrent.
func (c *Client) GetTorrentFiles(ctx context.Context, hash string) ([]TorrentFile, error) {
	var out []TorrentFile
	p := url.Values{}
	p.Set("hash", hash)
	if err := c.get(ctx, "/api/v2/torrents/files", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AddTorrentURLs adds torrents by URL (magnet links or HTTP URLs to .torrent files).
func (c *Client) AddTorrentURLs(ctx context.Context, urls []string, opts *AddTorrentOptions) error {
	form := url.Values{}
	form.Set("urls", strings.Join(urls, "\n"))
	if opts != nil {
		setAddOptions(form, opts)
	}
	return c.post(ctx, "/api/v2/torrents/add", form)
}

func setAddOptions(form url.Values, opts *AddTorrentOptions) {
	if opts.SavePath != "" {
		form.Set("savepath", opts.SavePath)
	}
	if opts.Category != "" {
		form.Set("category", opts.Category)
	}
	if opts.Tags != "" {
		form.Set("tags", opts.Tags)
	}
	if opts.SkipChecking {
		form.Set("skip_checking", "true")
	}
	if opts.Paused {
		form.Set("paused", "true")
	}
	if opts.RootFolder {
		form.Set("root_folder", "true")
	}
	if opts.Rename != "" {
		form.Set("rename", opts.Rename)
	}
	if opts.UpLimit > 0 {
		form.Set("upLimit", strconv.FormatInt(opts.UpLimit, 10))
	}
	if opts.DlLimit > 0 {
		form.Set("dlLimit", strconv.FormatInt(opts.DlLimit, 10))
	}
	if opts.AutoTMM {
		form.Set("autoTMM", "true")
	}
	if opts.SequentialDownload {
		form.Set("sequentialDownload", "true")
	}
	if opts.FirstLastPiecePrio {
		form.Set("firstLastPiecePrio", "true")
	}
}

// DeleteTorrents removes torrents by their hashes.
// If deleteFiles is true, downloaded data is also deleted from disk.
func (c *Client) DeleteTorrents(ctx context.Context, hashes []string, deleteFiles bool) error {
	form := url.Values{}
	form.Set("hashes", strings.Join(hashes, "|"))
	form.Set("deleteFiles", strconv.FormatBool(deleteFiles))
	return c.post(ctx, "/api/v2/torrents/delete", form)
}

// PauseTorrents pauses torrents by their hashes.
// Use "all" as single element to pause all torrents.
func (c *Client) PauseTorrents(ctx context.Context, hashes []string) error {
	form := url.Values{}
	form.Set("hashes", strings.Join(hashes, "|"))
	return c.post(ctx, "/api/v2/torrents/pause", form)
}

// ResumeTorrents resumes torrents by their hashes.
// Use "all" as single element to resume all torrents.
func (c *Client) ResumeTorrents(ctx context.Context, hashes []string) error {
	form := url.Values{}
	form.Set("hashes", strings.Join(hashes, "|"))
	return c.post(ctx, "/api/v2/torrents/resume", form)
}

// RecheckTorrents rechecks torrents by their hashes.
func (c *Client) RecheckTorrents(ctx context.Context, hashes []string) error {
	form := url.Values{}
	form.Set("hashes", strings.Join(hashes, "|"))
	return c.post(ctx, "/api/v2/torrents/recheck", form)
}

// ReannounceTorrents reannounces torrents to their trackers.
func (c *Client) ReannounceTorrents(ctx context.Context, hashes []string) error {
	form := url.Values{}
	form.Set("hashes", strings.Join(hashes, "|"))
	return c.post(ctx, "/api/v2/torrents/reannounce", form)
}

// SetTorrentLocation moves torrents to the specified path.
func (c *Client) SetTorrentLocation(ctx context.Context, hashes []string, location string) error {
	form := url.Values{}
	form.Set("hashes", strings.Join(hashes, "|"))
	form.Set("location", location)
	return c.post(ctx, "/api/v2/torrents/setLocation", form)
}

// RenameTorrent renames a torrent.
func (c *Client) RenameTorrent(ctx context.Context, hash, name string) error {
	form := url.Values{}
	form.Set("hash", hash)
	form.Set("name", name)
	return c.post(ctx, "/api/v2/torrents/rename", form)
}

// Categories and Tags.

// ListCategories returns all categories.
func (c *Client) ListCategories(ctx context.Context) (map[string]*Category, error) {
	out := make(map[string]*Category)
	if err := c.get(ctx, "/api/v2/torrents/categories", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateCategory creates a new category with an optional save path.
func (c *Client) CreateCategory(ctx context.Context, name, savePath string) error {
	form := url.Values{}
	form.Set("category", name)
	form.Set("savePath", savePath)
	return c.post(ctx, "/api/v2/torrents/createCategory", form)
}

// SetTorrentCategory assigns a category to torrents.
func (c *Client) SetTorrentCategory(ctx context.Context, hashes []string, category string) error {
	form := url.Values{}
	form.Set("hashes", strings.Join(hashes, "|"))
	form.Set("category", category)
	return c.post(ctx, "/api/v2/torrents/setCategory", form)
}

// ListTags returns all tags.
func (c *Client) ListTags(ctx context.Context) ([]string, error) {
	var out []string
	if err := c.get(ctx, "/api/v2/torrents/tags", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AddTorrentTags adds tags to torrents.
func (c *Client) AddTorrentTags(ctx context.Context, hashes, tags []string) error {
	form := url.Values{}
	form.Set("hashes", strings.Join(hashes, "|"))
	form.Set("tags", strings.Join(tags, ","))
	return c.post(ctx, "/api/v2/torrents/addTags", form)
}

// RemoveTorrentTags removes tags from torrents.
func (c *Client) RemoveTorrentTags(ctx context.Context, hashes, tags []string) error {
	form := url.Values{}
	form.Set("hashes", strings.Join(hashes, "|"))
	form.Set("tags", strings.Join(tags, ","))
	return c.post(ctx, "/api/v2/torrents/removeTags", form)
}

// Transfer.

// GetTransferInfo returns global transfer statistics.
func (c *Client) GetTransferInfo(ctx context.Context) (*TransferInfo, error) {
	var out TransferInfo
	if err := c.get(ctx, "/api/v2/transfer/info", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetGlobalDownloadLimit returns the global download speed limit in bytes/second.
// A value of 0 means no limit.
func (c *Client) GetGlobalDownloadLimit(ctx context.Context) (int64, error) {
	var out int64
	if err := c.get(ctx, "/api/v2/transfer/downloadLimit", nil, &out); err != nil {
		return 0, err
	}
	return out, nil
}

// SetGlobalDownloadLimit sets the global download speed limit in bytes/second.
// Set to 0 to remove the limit.
func (c *Client) SetGlobalDownloadLimit(ctx context.Context, limit int64) error {
	form := url.Values{}
	form.Set("limit", strconv.FormatInt(limit, 10))
	return c.post(ctx, "/api/v2/transfer/setDownloadLimit", form)
}

// GetGlobalUploadLimit returns the global upload speed limit in bytes/second.
func (c *Client) GetGlobalUploadLimit(ctx context.Context) (int64, error) {
	var out int64
	if err := c.get(ctx, "/api/v2/transfer/uploadLimit", nil, &out); err != nil {
		return 0, err
	}
	return out, nil
}

// SetGlobalUploadLimit sets the global upload speed limit in bytes/second.
func (c *Client) SetGlobalUploadLimit(ctx context.Context, limit int64) error {
	form := url.Values{}
	form.Set("limit", strconv.FormatInt(limit, 10))
	return c.post(ctx, "/api/v2/transfer/setUploadLimit", form)
}

// Sync.

// GetSyncMainData returns the sync main data.
// Pass rid=0 for a full update, or the previous rid for incremental updates.
func (c *Client) GetSyncMainData(ctx context.Context, rid int) (*SyncMainData, error) {
	var out SyncMainData
	p := url.Values{}
	p.Set("rid", strconv.Itoa(rid))
	if err := c.get(ctx, "/api/v2/sync/maindata", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Log.

// GetLog returns application log entries.
func (c *Client) GetLog(ctx context.Context, lastKnownID int) ([]LogEntry, error) {
	var out []LogEntry
	p := url.Values{}
	p.Set("last_known_id", strconv.Itoa(lastKnownID))
	if err := c.get(ctx, "/api/v2/log/main", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPeerLog returns peer log entries.
func (c *Client) GetPeerLog(ctx context.Context, lastKnownID int) ([]PeerLogEntry, error) {
	var out []PeerLogEntry
	p := url.Values{}
	p.Set("last_known_id", strconv.Itoa(lastKnownID))
	if err := c.get(ctx, "/api/v2/log/peers", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}
