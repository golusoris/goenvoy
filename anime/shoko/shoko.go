package shoko

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "goenvoy/0.0.1"
	defaultDevice    = "goenvoy"
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

// WithAPIKey sets a pre-existing API key, skipping the need to call [Client.Login].
func WithAPIKey(key string) Option {
	return func(cl *Client) { cl.apiKey = key }
}

// Client is a Shoko Server API v3 client.
type Client struct {
	rawBaseURL string
	httpClient *http.Client
	userAgent  string
	apiKey     string
}

// New creates a Shoko [Client] for the server at baseURL (e.g. "http://localhost:8111").
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		rawBaseURL: baseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
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
	Title      string `json:"title"`
	RawBody    string `json:"-"`
}

func (e *APIError) Error() string {
	if e.Title != "" {
		return fmt.Sprintf("shoko: HTTP %d: %s", e.StatusCode, e.Title)
	}
	if e.RawBody != "" {
		return fmt.Sprintf("shoko: HTTP %d: %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("shoko: HTTP %d", e.StatusCode)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body, dst any) error {
	u, err := url.Parse(c.rawBaseURL + path)
	if err != nil {
		return fmt.Errorf("shoko: parse URL: %w", err)
	}

	var reqBody io.Reader = http.NoBody
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("shoko: marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return fmt.Errorf("shoko: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.apiKey != "" {
		req.Header.Set("apikey", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("shoko: %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("shoko: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if jsonErr := json.Unmarshal(respBody, apiErr); jsonErr != nil {
			apiErr.RawBody = string(respBody)
		}
		return apiErr
	}

	if dst != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return fmt.Errorf("shoko: decode response: %w", err)
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

func (c *Client) post(ctx context.Context, path string, body, dst any) error {
	return c.doRequest(ctx, http.MethodPost, path, body, dst)
}

// pageParams builds pagination query parameters.
func pageParams(page, pageSize int) url.Values {
	p := url.Values{}
	if page > 0 {
		p.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		p.Set("pageSize", strconv.Itoa(pageSize))
	}
	return p
}

// Authentication.

// Login authenticates with the Shoko server and stores the API key on the client.
func (c *Client) Login(ctx context.Context, username, password string) error {
	var resp loginResponse
	err := c.post(ctx, "/api/auth", &loginRequest{
		User:   username,
		Pass:   password,
		Device: defaultDevice,
	}, &resp)
	if err != nil {
		return err
	}
	c.apiKey = resp.APIKey
	return nil
}

// Series endpoints.

// GetSeries returns a series by its Shoko ID.
func (c *Client) GetSeries(ctx context.Context, seriesID int) (*Series, error) {
	var out Series
	if err := c.get(ctx, "/api/v3/Series/"+strconv.Itoa(seriesID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSeries returns a paginated list of series.
func (c *Client) ListSeries(ctx context.Context, page, pageSize int) ([]Series, error) {
	var out []Series
	if err := c.get(ctx, "/api/v3/Series", pageParams(page, pageSize), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchSeries searches for series by name.
func (c *Client) SearchSeries(ctx context.Context, query string, fuzzy bool, page, pageSize int) ([]Series, error) {
	p := pageParams(page, pageSize)
	p.Set("search", query)
	if fuzzy {
		p.Set("fuzzy", "true")
	}
	var out []Series
	if err := c.get(ctx, "/api/v3/Series", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSeriesWithoutFiles returns series that have no associated files.
func (c *Client) GetSeriesWithoutFiles(ctx context.Context) ([]Series, error) {
	var out []Series
	if err := c.get(ctx, "/api/v3/Series/WithoutFiles", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSeriesAniDB returns the AniDB metadata for a series.
func (c *Client) GetSeriesAniDB(ctx context.Context, seriesID int) (*AniDBAnime, error) {
	var out AniDBAnime
	if err := c.get(ctx, "/api/v3/Series/"+strconv.Itoa(seriesID)+"/AniDB", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSeriesSimilar returns similar anime for a series from AniDB.
func (c *Client) GetSeriesSimilar(ctx context.Context, seriesID int) ([]AniDBAnime, error) {
	var out []AniDBAnime
	if err := c.get(ctx, "/api/v3/Series/"+strconv.Itoa(seriesID)+"/AniDB/Similar", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSeriesTags returns tags for a series.
func (c *Client) GetSeriesTags(ctx context.Context, seriesID int) ([]Tag, error) {
	var out []Tag
	if err := c.get(ctx, "/api/v3/Series/"+strconv.Itoa(seriesID)+"/Tags", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSeriesEpisodes returns episodes for a series.
func (c *Client) GetSeriesEpisodes(ctx context.Context, seriesID, page, pageSize int) ([]Episode, error) {
	var out []Episode
	if err := c.get(ctx, "/api/v3/Series/"+strconv.Itoa(seriesID)+"/Episode", pageParams(page, pageSize), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RefreshSeriesAniDB queues an AniDB data refresh for a series.
func (c *Client) RefreshSeriesAniDB(ctx context.Context, seriesID int) error {
	return c.post(ctx, "/api/v3/Series/"+strconv.Itoa(seriesID)+"/AniDB/Refresh", nil, nil)
}

// AniDB lookup endpoints.

// GetAniDBAnime returns AniDB anime metadata by AniDB ID.
func (c *Client) GetAniDBAnime(ctx context.Context, anidbID int) (*AniDBAnime, error) {
	var out AniDBAnime
	if err := c.get(ctx, "/api/v3/Series/AniDB/"+strconv.Itoa(anidbID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSeriesByAniDBID returns the Shoko series linked to an AniDB anime.
func (c *Client) GetSeriesByAniDBID(ctx context.Context, anidbID int) (*Series, error) {
	var out Series
	if err := c.get(ctx, "/api/v3/Series/AniDB/"+strconv.Itoa(anidbID)+"/Series", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAniDBRelations returns relations for an AniDB anime.
func (c *Client) GetAniDBRelations(ctx context.Context, anidbID int) ([]AniDBRelation, error) {
	var out []AniDBRelation
	if err := c.get(ctx, "/api/v3/Series/AniDB/"+strconv.Itoa(anidbID)+"/Relations", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Episode endpoints.

// GetEpisode returns an episode by its Shoko ID.
func (c *Client) GetEpisode(ctx context.Context, episodeID int) (*Episode, error) {
	var out Episode
	if err := c.get(ctx, "/api/v3/Episode/"+strconv.Itoa(episodeID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAniDBEpisode returns AniDB episode metadata by AniDB episode ID.
func (c *Client) GetAniDBEpisode(ctx context.Context, anidbEpisodeID int) (*AniDBEpisode, error) {
	var out AniDBEpisode
	if err := c.get(ctx, "/api/v3/Episode/AniDB/"+strconv.Itoa(anidbEpisodeID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// File endpoints.

// GetFile returns a file by its Shoko ID.
func (c *Client) GetFile(ctx context.Context, fileID int) (*File, error) {
	var out File
	if err := c.get(ctx, "/api/v3/File/"+strconv.Itoa(fileID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListFiles returns a paginated list of files.
func (c *Client) ListFiles(ctx context.Context, page, pageSize int) ([]File, error) {
	var out []File
	if err := c.get(ctx, "/api/v3/File", pageParams(page, pageSize), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RescanFile queues a file for rescanning on AniDB.
func (c *Client) RescanFile(ctx context.Context, fileID int) error {
	return c.post(ctx, "/api/v3/File/"+strconv.Itoa(fileID)+"/Rescan", nil, nil)
}

// RehashFile queues a file for rehashing.
func (c *Client) RehashFile(ctx context.Context, fileID int) error {
	return c.post(ctx, "/api/v3/File/"+strconv.Itoa(fileID)+"/Rehash", nil, nil)
}

// Managed folder endpoints.

// ListManagedFolders returns all managed (import) folders.
func (c *Client) ListManagedFolders(ctx context.Context) ([]ManagedFolder, error) {
	var out []ManagedFolder
	if err := c.get(ctx, "/api/v3/ManagedFolder", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetManagedFolder returns a managed folder by ID.
func (c *Client) GetManagedFolder(ctx context.Context, folderID int) (*ManagedFolder, error) {
	var out ManagedFolder
	if err := c.get(ctx, "/api/v3/ManagedFolder/"+strconv.Itoa(folderID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ScanManagedFolder triggers a scan on a managed folder.
func (c *Client) ScanManagedFolder(ctx context.Context, folderID int) error {
	return c.get(ctx, "/api/v3/ManagedFolder/"+strconv.Itoa(folderID)+"/Scan", nil, nil)
}

// Server action endpoints.

// RunImport triggers a full import (hash, scan, match, download images).
func (c *Client) RunImport(ctx context.Context) error {
	return c.get(ctx, "/api/v3/Action/RunImport", nil, nil)
}

// ImportNewFiles scans managed folders for new files only.
func (c *Client) ImportNewFiles(ctx context.Context) error {
	return c.get(ctx, "/api/v3/Action/ImportNewFiles", nil, nil)
}

// Dashboard endpoints.

// GetDashboardStats returns server dashboard statistics.
func (c *Client) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	var out DashboardStats
	if err := c.get(ctx, "/api/v3/Dashboard/Stats", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
