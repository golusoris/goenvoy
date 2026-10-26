package audiobookshelf

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "goenvoy/0.0.1"
)

var (
	// ErrEmptyID is returned when a required resource identifier is empty.
	ErrEmptyID = errors.New("id must not be empty")
	// ErrEmptyUsername is returned when a login username is empty.
	ErrEmptyUsername = errors.New("username must not be empty")
	// ErrEmptyRSSFeed is returned when a podcast RSS feed URL is empty.
	ErrEmptyRSSFeed = errors.New("rssFeed must not be empty")
	// ErrEmptySearchTerm is returned when a podcast search term is empty.
	ErrEmptySearchTerm = errors.New("search term must not be empty")
	// ErrEmptyEpisodes is returned when no episodes are provided for download.
	ErrEmptyEpisodes = errors.New("episodes must not be empty")
	// ErrInvalidPodcastRequest is returned when a CreatePodcastRequest is nil or missing required fields.
	ErrInvalidPodcastRequest = errors.New("podcast request requires libraryId, folderId, and path")
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

// WithUserAgent sets the User-Agent header for all requests.
func WithUserAgent(ua string) Option {
	return func(cl *Client) { cl.userAgent = ua }
}

// Client is an Audiobookshelf API client.
type Client struct {
	rawBaseURL string
	token      string
	httpClient *http.Client
	userAgent  string
}

// New creates an Audiobookshelf [Client] for the instance at baseURL with the given token.
// Pass an empty token when constructing a client for [Client.Login] or [Client.GetStatus].
// It returns an error if baseURL is not a valid HTTP/HTTPS URL.
func New(baseURL, token string, opts ...Option) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("audiobookshelf: invalid base URL %q: %w", baseURL, err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("audiobookshelf: invalid base URL %q: must be http(s) with a host", baseURL)
	}

	c := &Client{
		rawBaseURL: baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: defaultTimeout, Transport: sharedTransport},
		userAgent:  defaultUserAgent,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// APIError is returned when the API responds with a non-2xx status.
type APIError struct {
	StatusCode int    `json:"-"`
	RawBody    string `json:"-"`
}

func (e *APIError) Error() string {
	if e.RawBody != "" {
		return fmt.Sprintf("audiobookshelf: HTTP %d: %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("audiobookshelf: HTTP %d", e.StatusCode)
}

func marshalBody(reqBody any) (io.Reader, error) {
	if reqBody == nil {
		return http.NoBody, nil
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("audiobookshelf: marshal request: %w", err)
	}
	return bytes.NewReader(data), nil
}

func (c *Client) newRequest(
	ctx context.Context,
	method, fullPath string,
	params url.Values,
	reqBody any,
) (*http.Request, error) {
	u, err := url.Parse(c.rawBaseURL + fullPath)
	if err != nil {
		return nil, fmt.Errorf("audiobookshelf: parse URL: %w", err)
	}
	if len(params) > 0 {
		u.RawQuery = params.Encode()
	}

	bodyReader, err := marshalBody(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("audiobookshelf: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return req, nil
}

func (c *Client) execute(req *http.Request, label string) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("audiobookshelf: %s %s: %w", req.Method, label, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("audiobookshelf: read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &APIError{StatusCode: resp.StatusCode, RawBody: string(body)}
	}
	return body, nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api"+path, params, nil)
	if err != nil {
		return nil, err
	}
	return c.execute(req, path)
}

func (c *Client) post(ctx context.Context, path string, reqBody any) ([]byte, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/api"+path, nil, reqBody)
	if err != nil {
		return nil, err
	}
	return c.execute(req, path)
}

func (c *Client) doRoot(
	ctx context.Context,
	method, path string,
	reqBody any,
	headers map[string]string,
) ([]byte, error) {
	req, err := c.newRequest(ctx, method, path, nil, reqBody)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.execute(req, path)
}

// Authentication (root routes).

// Login authenticates with username and password via POST /login (root route, not /api)
// and sends X-Return-Tokens: true so accessToken and refreshToken are returned in [User]
// (server/Server.js:375, server/Auth.js:96-105, 296-312, 322-328).
func (c *Client) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	if username == "" {
		return nil, fmt.Errorf("audiobookshelf: login: %w", ErrEmptyUsername)
	}
	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	}
	data, err := c.doRoot(ctx, http.MethodPost, "/login", payload, map[string]string{
		"X-Return-Tokens": "true",
	})
	if err != nil {
		return nil, err
	}
	var resp LoginResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode login: %w", err)
	}
	return &resp, nil
}

// Logout terminates the current session via POST /logout (root route, not /api)
// (server/Server.js:375, server/Auth.js:486-524).
func (c *Client) Logout(ctx context.Context) error {
	_, err := c.doRoot(ctx, http.MethodPost, "/logout", nil, nil)
	return err
}

// Libraries.

// GetLibraries returns all libraries.
func (c *Client) GetLibraries(ctx context.Context) ([]Library, error) {
	data, err := c.get(ctx, "/libraries", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Libraries []Library `json:"libraries"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode libraries: %w", err)
	}
	return resp.Libraries, nil
}

// GetLibrary returns a library by ID.
func (c *Client) GetLibrary(ctx context.Context, libraryID string) (*Library, error) {
	if libraryID == "" {
		return nil, fmt.Errorf("audiobookshelf: get library: %w", ErrEmptyID)
	}
	data, err := c.get(ctx, "/libraries/"+url.PathEscape(libraryID), nil)
	if err != nil {
		return nil, err
	}
	var lib Library
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode library: %w", err)
	}
	return &lib, nil
}

func (o *LibraryItemsOptions) values() url.Values {
	if o == nil {
		return nil
	}
	p := url.Values{}
	if o.Page > 0 {
		p.Set("page", strconv.Itoa(o.Page))
	}
	if o.Limit > 0 {
		p.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Sort != "" {
		p.Set("sort", o.Sort)
	}
	if o.Desc {
		p.Set("desc", "1")
	}
	if o.Filter != "" {
		p.Set("filter", o.Filter)
	}
	if o.Minified {
		p.Set("minified", "1")
	}
	if o.CollapseSeries {
		p.Set("collapseseries", "1")
	}
	if o.Include != "" {
		p.Set("include", o.Include)
	}
	return p
}

// GetLibraryItems returns items in a library with pagination
// (server/routers/ApiRouter.js:75, server/controllers/LibraryController.js:604-639).
func (c *Client) GetLibraryItems(ctx context.Context, libraryID string, page, limit int) (*LibraryItemsResponse, error) {
	return c.GetLibraryItemsWithOptions(ctx, libraryID, &LibraryItemsOptions{
		Page:  page,
		Limit: limit,
	})
}

// GetLibraryItemsWithOptions returns items in a library using all supported query options
// (server/routers/ApiRouter.js:75, server/controllers/LibraryController.js:604-639, 1487-1495).
func (c *Client) GetLibraryItemsWithOptions(
	ctx context.Context,
	libraryID string,
	opts *LibraryItemsOptions,
) (*LibraryItemsResponse, error) {
	if libraryID == "" {
		return nil, fmt.Errorf("audiobookshelf: get library items: %w", ErrEmptyID)
	}
	data, err := c.get(ctx, "/libraries/"+url.PathEscape(libraryID)+"/items", opts.values())
	if err != nil {
		return nil, err
	}
	var resp LibraryItemsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode library items: %w", err)
	}
	return &resp, nil
}

// Items.

// GetItem returns a library item by ID.
func (c *Client) GetItem(ctx context.Context, itemID string) (*LibraryItem, error) {
	if itemID == "" {
		return nil, fmt.Errorf("audiobookshelf: get item: %w", ErrEmptyID)
	}
	data, err := c.get(ctx, "/items/"+url.PathEscape(itemID), nil)
	if err != nil {
		return nil, err
	}
	var item LibraryItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode item: %w", err)
	}
	return &item, nil
}

// Users.

// GetUsers returns all users (admin only) from GET /api/users, unwrapping the {"users": [...]} envelope
// (server/routers/ApiRouter.js:134, server/controllers/UserController.js:29-49).
func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	data, err := c.get(ctx, "/users", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Users []User `json:"users"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode users: %w", err)
	}
	return resp.Users, nil
}

// GetUser returns a single user by ID via GET /api/users/:id (admin or self)
// (server/routers/ApiRouter.js:136, server/controllers/UserController.js:59-105).
func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, fmt.Errorf("audiobookshelf: get user: %w", ErrEmptyID)
	}
	data, err := c.get(ctx, "/users/"+url.PathEscape(userID), nil)
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode user: %w", err)
	}
	return &user, nil
}

// GetMe returns the authenticated user.
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	data, err := c.get(ctx, "/me", nil)
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode user: %w", err)
	}
	return &user, nil
}

// Collections.

// GetCollections returns all collections for a library.
func (c *Client) GetCollections(ctx context.Context, libraryID string) ([]Collection, error) {
	if libraryID == "" {
		return nil, fmt.Errorf("audiobookshelf: get collections: %w", ErrEmptyID)
	}
	data, err := c.get(ctx, "/libraries/"+url.PathEscape(libraryID)+"/collections", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Results []Collection `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode collections: %w", err)
	}
	return resp.Results, nil
}

// Sessions.

// GetSessions returns active listening sessions.
func (c *Client) GetSessions(ctx context.Context) ([]PlaybackSession, error) {
	data, err := c.get(ctx, "/sessions", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Sessions []PlaybackSession `json:"sessions"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode sessions: %w", err)
	}
	return resp.Sessions, nil
}

// Server.

// GetStatus returns server initialization and version status via GET /status (root route, not /api)
// (server/Server.js:384-400).
func (c *Client) GetStatus(ctx context.Context) (*ServerStatus, error) {
	data, err := c.doRoot(ctx, http.MethodGet, "/status", nil, nil)
	if err != nil {
		return nil, err
	}
	var status ServerStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode status: %w", err)
	}
	return &status, nil
}

// GetServerInfo returns server information.
//
// Deprecated: Audiobookshelf has no /api/server route; use GetStatus.
func (c *Client) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	data, err := c.get(ctx, "/server", nil)
	if err != nil {
		return nil, err
	}
	var info ServerInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode server info: %w", err)
	}
	return &info, nil
}

// Search searches within a library.
func (c *Client) Search(ctx context.Context, libraryID, query string) ([]byte, error) {
	if libraryID == "" {
		return nil, fmt.Errorf("audiobookshelf: search: %w", ErrEmptyID)
	}
	p := url.Values{"q": {query}}
	return c.get(ctx, "/libraries/"+url.PathEscape(libraryID)+"/search", p)
}

// GetMediaProgress returns listening progress for an item.
func (c *Client) GetMediaProgress(ctx context.Context, itemID string) (*MediaProgress, error) {
	if itemID == "" {
		return nil, fmt.Errorf("audiobookshelf: get media progress: %w", ErrEmptyID)
	}
	data, err := c.get(ctx, "/me/progress/"+url.PathEscape(itemID), nil)
	if err != nil {
		return nil, err
	}
	var progress MediaProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode progress: %w", err)
	}
	return &progress, nil
}

// GetAuthors returns all authors in a library.
func (c *Client) GetAuthors(ctx context.Context, libraryID string) ([]Author, error) {
	if libraryID == "" {
		return nil, fmt.Errorf("audiobookshelf: get authors: %w", ErrEmptyID)
	}
	data, err := c.get(ctx, "/libraries/"+url.PathEscape(libraryID)+"/authors", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Authors []Author `json:"authors"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode authors: %w", err)
	}
	return resp.Authors, nil
}

// GetSeries returns all series in a library.
func (c *Client) GetSeries(ctx context.Context, libraryID string) ([]Series, error) {
	if libraryID == "" {
		return nil, fmt.Errorf("audiobookshelf: get series: %w", ErrEmptyID)
	}
	data, err := c.get(ctx, "/libraries/"+url.PathEscape(libraryID)+"/series", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Results []Series `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode series: %w", err)
	}
	return resp.Results, nil
}

// Podcasts.

// GetPodcastFeed fetches and parses a podcast RSS feed via POST /api/podcasts/feed
// (server/routers/ApiRouter.js:249, server/controllers/PodcastController.js:191-216).
func (c *Client) GetPodcastFeed(ctx context.Context, rssFeed string) (*PodcastFeed, error) {
	if rssFeed == "" {
		return nil, fmt.Errorf("audiobookshelf: get podcast feed: %w", ErrEmptyRSSFeed)
	}
	data, err := c.post(ctx, "/podcasts/feed", PodcastFeedRequest{RSSFeed: rssFeed})
	if err != nil {
		return nil, err
	}
	var resp PodcastFeedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode podcast feed: %w", err)
	}
	return &resp.Podcast, nil
}

// CreatePodcast creates a new podcast library item via POST /api/podcasts
// (server/routers/ApiRouter.js:248, server/controllers/PodcastController.js:31-189).
func (c *Client) CreatePodcast(ctx context.Context, req *CreatePodcastRequest) (*LibraryItem, error) {
	if req == nil || req.LibraryID == "" || req.FolderID == "" || strings.TrimSpace(req.Path) == "" {
		return nil, fmt.Errorf("audiobookshelf: create podcast: %w", ErrInvalidPodcastRequest)
	}
	data, err := c.post(ctx, "/podcasts", req)
	if err != nil {
		return nil, err
	}
	var item LibraryItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode podcast: %w", err)
	}
	return &item, nil
}

// CheckNewEpisodes checks a podcast feed for new episodes via GET /api/podcasts/:id/checknew
// (server/routers/ApiRouter.js:252, server/controllers/PodcastController.js:276-301).
func (c *Client) CheckNewEpisodes(ctx context.Context, podcastID string, limit int) ([]PodcastFeedEpisode, error) {
	if podcastID == "" {
		return nil, fmt.Errorf("audiobookshelf: check new episodes: %w", ErrEmptyID)
	}
	var p url.Values
	if limit > 0 {
		p = url.Values{"limit": {strconv.Itoa(limit)}}
	}
	data, err := c.get(ctx, "/podcasts/"+url.PathEscape(podcastID)+"/checknew", p)
	if err != nil {
		return nil, err
	}
	var resp CheckNewEpisodesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode new episodes: %w", err)
	}
	return resp.Episodes, nil
}

// DownloadEpisodes queues podcast episodes for download via POST /api/podcasts/:id/download-episodes
// (server/routers/ApiRouter.js:256, server/controllers/PodcastController.js:360-380).
func (c *Client) DownloadEpisodes(ctx context.Context, podcastID string, episodes []PodcastFeedEpisode) error {
	if podcastID == "" {
		return fmt.Errorf("audiobookshelf: download episodes: %w", ErrEmptyID)
	}
	if len(episodes) == 0 {
		return fmt.Errorf("audiobookshelf: download episodes: %w", ErrEmptyEpisodes)
	}
	_, err := c.post(ctx, "/podcasts/"+url.PathEscape(podcastID)+"/download-episodes", episodes)
	return err
}

// SearchPodcasts searches iTunes for podcasts via GET /api/search/podcast
// (server/routers/ApiRouter.js:288, server/controllers/SearchController.js:135-156, server/providers/iTunes.js:139-168).
func (c *Client) SearchPodcasts(ctx context.Context, term, country string) ([]PodcastSearchResult, error) {
	if term == "" {
		return nil, fmt.Errorf("audiobookshelf: search podcasts: %w", ErrEmptySearchTerm)
	}
	p := url.Values{"term": {term}}
	if country != "" {
		p.Set("country", country)
	}
	data, err := c.get(ctx, "/search/podcast", p)
	if err != nil {
		return nil, err
	}
	var results []PodcastSearchResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("audiobookshelf: decode podcast search: %w", err)
	}
	return results, nil
}
