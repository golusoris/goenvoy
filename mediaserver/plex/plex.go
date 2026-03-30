package plex

import (
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
	defaultTimeout  = 30 * time.Second
	defaultProduct  = "goenvoy"
	defaultVersion  = "0.0.1"
	defaultClientID = "goenvoy-client"
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

// WithProduct sets the X-Plex-Product header.
func WithProduct(product string) Option {
	return func(cl *Client) { cl.product = product }
}

// WithClientIdentifier sets the X-Plex-Client-Identifier header.
func WithClientIdentifier(id string) Option {
	return func(cl *Client) { cl.clientID = id }
}

// Client is a Plex Media Server API client.
type Client struct {
	rawBaseURL string
	token      string
	httpClient *http.Client
	product    string
	version    string
	clientID   string
}

// New creates a Plex [Client] for the server at baseURL with the given token.
func New(baseURL, token string, opts ...Option) *Client {
	c := &Client{
		rawBaseURL: baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: defaultTimeout},
		product:    defaultProduct,
		version:    defaultVersion,
		clientID:   defaultClientID,
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
		return fmt.Sprintf("plex: HTTP %d: %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("plex: HTTP %d", e.StatusCode)
}

// containerResponse wraps the JSON response which nests MediaContainer.
type containerResponse struct {
	MediaContainer MediaContainer `json:"MediaContainer"`
}

func (c *Client) get(ctx context.Context, path string, params url.Values) (*MediaContainer, error) {
	u, err := url.Parse(c.rawBaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("plex: parse URL: %w", err)
	}
	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("plex: create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Token", c.token)
	req.Header.Set("X-Plex-Product", c.product)
	req.Header.Set("X-Plex-Version", c.version)
	req.Header.Set("X-Plex-Client-Identifier", c.clientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("plex: GET %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("plex: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{StatusCode: resp.StatusCode, RawBody: string(body)}
	}

	var wrapper containerResponse
	if len(body) > 0 {
		if err := json.Unmarshal(body, &wrapper); err != nil {
			return nil, fmt.Errorf("plex: decode response: %w", err)
		}
	}
	return &wrapper.MediaContainer, nil
}

func (c *Client) getRaw(ctx context.Context, path string) ([]byte, error) {
	u, err := url.Parse(c.rawBaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("plex: parse URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("plex: create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Token", c.token)
	req.Header.Set("X-Plex-Product", c.product)
	req.Header.Set("X-Plex-Version", c.version)
	req.Header.Set("X-Plex-Client-Identifier", c.clientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("plex: GET %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("plex: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{StatusCode: resp.StatusCode, RawBody: string(body)}
	}
	return body, nil
}

// Server information.

// GetIdentity returns the server identity (no authentication required).
func (c *Client) GetIdentity(ctx context.Context) (*Identity, error) {
	data, err := c.getRaw(ctx, "/identity")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		MediaContainer Identity `json:"MediaContainer"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("plex: decode identity: %w", err)
	}
	return &wrapper.MediaContainer, nil
}

// GetServerInfo returns the server root information.
func (c *Client) GetServerInfo(ctx context.Context) (*MediaContainer, error) {
	return c.get(ctx, "/", nil)
}

// Libraries.

// GetLibraries returns all library sections.
func (c *Client) GetLibraries(ctx context.Context) ([]Directory, error) {
	mc, err := c.get(ctx, "/library/sections", nil)
	if err != nil {
		return nil, err
	}
	return mc.Directory, nil
}

// GetLibraryContents returns items in a library section with pagination.
func (c *Client) GetLibraryContents(ctx context.Context, sectionID string, start, size int) (*MediaContainer, error) {
	p := url.Values{}
	if start >= 0 {
		p.Set("X-Plex-Container-Start", strconv.Itoa(start))
	}
	if size > 0 {
		p.Set("X-Plex-Container-Size", strconv.Itoa(size))
	}
	return c.get(ctx, "/library/sections/"+url.PathEscape(sectionID)+"/all", p)
}

// RefreshLibrary triggers a library scan for the given section.
func (c *Client) RefreshLibrary(ctx context.Context, sectionID string) error {
	_, err := c.get(ctx, "/library/sections/"+url.PathEscape(sectionID)+"/refresh", nil)
	return err
}

// Media items.

// GetMetadata returns metadata for a specific item by rating key.
func (c *Client) GetMetadata(ctx context.Context, ratingKey string) (*Metadata, error) {
	mc, err := c.get(ctx, "/library/metadata/"+url.PathEscape(ratingKey), nil)
	if err != nil {
		return nil, err
	}
	if len(mc.Metadata) == 0 {
		return nil, fmt.Errorf("plex: no metadata found for key %s", ratingKey)
	}
	return &mc.Metadata[0], nil
}

// GetOnDeck returns the on-deck (continue watching) items.
func (c *Client) GetOnDeck(ctx context.Context) ([]Metadata, error) {
	mc, err := c.get(ctx, "/library/onDeck", nil)
	if err != nil {
		return nil, err
	}
	return mc.Metadata, nil
}

// GetRecentlyAdded returns recently added items.
func (c *Client) GetRecentlyAdded(ctx context.Context) ([]Metadata, error) {
	mc, err := c.get(ctx, "/library/recentlyAdded", nil)
	if err != nil {
		return nil, err
	}
	return mc.Metadata, nil
}

// Search searches across all libraries.
func (c *Client) Search(ctx context.Context, query string) (*MediaContainer, error) {
	p := url.Values{"query": {query}}
	return c.get(ctx, "/search", p)
}

// Sessions.

// GetSessions returns currently active playback sessions.
func (c *Client) GetSessions(ctx context.Context) ([]Metadata, error) {
	mc, err := c.get(ctx, "/status/sessions", nil)
	if err != nil {
		return nil, err
	}
	return mc.Metadata, nil
}

// GetTranscodeSessions returns active transcode sessions.
func (c *Client) GetTranscodeSessions(ctx context.Context) (*MediaContainer, error) {
	return c.get(ctx, "/transcode/sessions", nil)
}

// Playback control.

// MarkWatched marks an item as watched (scrobble).
func (c *Client) MarkWatched(ctx context.Context, ratingKey string) error {
	p := url.Values{
		"key":        {ratingKey},
		"identifier": {"com.plexapp.plugins.library"},
	}
	_, err := c.get(ctx, "/:/scrobble", p)
	return err
}

// MarkUnwatched marks an item as unwatched.
func (c *Client) MarkUnwatched(ctx context.Context, ratingKey string) error {
	p := url.Values{
		"key":        {ratingKey},
		"identifier": {"com.plexapp.plugins.library"},
	}
	_, err := c.get(ctx, "/:/unscrobble", p)
	return err
}
