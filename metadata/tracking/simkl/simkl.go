package simkl

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
	defaultBaseURL   = "https://api.simkl.com"
	defaultCalURL    = "https://data.simkl.in"
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "goenvoy/0.0.1"
	defaultFilter    = "all"
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

// WithBaseURL overrides the default Simkl API base URL.
func WithBaseURL(u string) Option {
	return func(cl *Client) { cl.rawBaseURL = u }
}

// WithCalendarURL overrides the default Simkl calendar (CDN) base URL.
func WithCalendarURL(u string) Option {
	return func(cl *Client) { cl.calBaseURL = u }
}

// Client is a Simkl API client.
type Client struct {
	clientID   string
	rawBaseURL string
	calBaseURL string
	httpClient *http.Client
	userAgent  string
}

// New creates a Simkl [Client] using the given client ID (API key).
func New(clientID string, opts ...Option) *Client {
	c := &Client{
		clientID:   clientID,
		rawBaseURL: defaultBaseURL,
		calBaseURL: defaultCalURL,
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
	Error_     string `json:"error"`
	Code       int    `json:"code"`
	Message    string `json:"message"`
	// RawBody holds the raw response body when the error could not be parsed as JSON.
	RawBody string `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("simkl: HTTP %d: %s: %s", e.StatusCode, e.Error_, e.Message)
	}
	if e.Error_ != "" {
		return fmt.Sprintf("simkl: HTTP %d: %s", e.StatusCode, e.Error_)
	}
	if e.RawBody != "" {
		return fmt.Sprintf("simkl: HTTP %d: %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("simkl: HTTP %d", e.StatusCode)
}

func (c *Client) doGet(ctx context.Context, baseURL, path string, params url.Values, dst any) error {
	u, err := url.Parse(baseURL + path)
	if err != nil {
		return fmt.Errorf("simkl: parse URL: %w", err)
	}
	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("simkl: create request: %w", err)
	}

	req.Header.Set("simkl-api-key", c.clientID)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("simkl: GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("simkl: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if err := json.Unmarshal(body, apiErr); err != nil {
			apiErr.RawBody = string(body)
		}
		return apiErr
	}

	if dst != nil && len(body) > 0 {
		if err := json.Unmarshal(body, dst); err != nil {
			return fmt.Errorf("simkl: decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, dst any) error {
	return c.doGet(ctx, c.rawBaseURL, path, params, dst)
}

func (c *Client) getCal(ctx context.Context, path string, dst any) error {
	return c.doGet(ctx, c.calBaseURL, path, nil, dst)
}

func pageParams(page, limit int) url.Values {
	p := url.Values{}
	if page > 0 {
		p.Set("page", strconv.Itoa(page))
	}
	if limit > 0 {
		p.Set("limit", strconv.Itoa(limit))
	}
	return p
}

// Movies.

// GetMovie returns detail information about a movie.
// The id can be a Simkl ID or an IMDB ID.
func (c *Client) GetMovie(ctx context.Context, id string) (*Movie, error) {
	var out Movie
	p := url.Values{}
	p.Set("extended", "full")
	if err := c.get(ctx, "/movies/"+url.PathEscape(id), p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TrendingMovies returns trending movies, optionally filtered by time interval.
// Valid intervals: "today", "week", "month" (empty string for default).
func (c *Client) TrendingMovies(ctx context.Context, interval string) ([]TrendingMovie, error) {
	path := "/movies/trending"
	if interval != "" {
		path += "/" + url.PathEscape(interval)
	}
	var out []TrendingMovie
	if err := c.get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TV Shows.

// GetShow returns detail information about a TV show.
func (c *Client) GetShow(ctx context.Context, id string) (*Show, error) {
	var out Show
	p := url.Values{}
	p.Set("extended", "full")
	if err := c.get(ctx, "/tv/"+url.PathEscape(id), p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetShowEpisodes returns episodes for a TV show.
func (c *Client) GetShowEpisodes(ctx context.Context, id string) ([]Episode, error) {
	var out []Episode
	p := url.Values{}
	p.Set("extended", "full")
	if err := c.get(ctx, "/tv/episodes/"+url.PathEscape(id), p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TrendingShows returns trending TV shows.
func (c *Client) TrendingShows(ctx context.Context, interval string) ([]TrendingShow, error) {
	path := "/tv/trending"
	if interval != "" {
		path += "/" + url.PathEscape(interval)
	}
	var out []TrendingShow
	if err := c.get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ShowGenres returns shows filtered by genre and other criteria.
func (c *Client) ShowGenres(ctx context.Context, genre string, page, limit int) ([]GenreItem, error) {
	if genre == "" {
		genre = defaultFilter
	}
	var out []GenreItem
	if err := c.get(ctx, "/tv/genres/"+url.PathEscape(genre), pageParams(page, limit), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ShowPremieres returns latest TV show premieres.
// param should be "new" or "soon".
func (c *Client) ShowPremieres(ctx context.Context, param string, page, limit int) ([]PremiereItem, error) {
	if param == "" {
		param = "new"
	}
	var out []PremiereItem
	if err := c.get(ctx, "/tv/premieres/"+url.PathEscape(param), pageParams(page, limit), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AiringShows returns currently airing TV shows.
func (c *Client) AiringShows(ctx context.Context, date, sort string) ([]AiringItem, error) {
	p := url.Values{}
	if date != "" {
		p.Set("date", date)
	}
	if sort != "" {
		p.Set("sort", sort)
	}
	var out []AiringItem
	if err := c.get(ctx, "/tv/airing", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BestShows returns the best of all TV shows.
// filter: "year", "month", "all", "voted", "watched".
func (c *Client) BestShows(ctx context.Context, filter string) ([]BestItem, error) {
	if filter == "" {
		filter = defaultFilter
	}
	var out []BestItem
	if err := c.get(ctx, "/tv/best/"+url.PathEscape(filter), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Anime.

// GetAnime returns detail information about an anime.
func (c *Client) GetAnime(ctx context.Context, id string) (*Anime, error) {
	var out Anime
	p := url.Values{}
	p.Set("extended", "full")
	if err := c.get(ctx, "/anime/"+url.PathEscape(id), p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAnimeEpisodes returns episodes for an anime.
func (c *Client) GetAnimeEpisodes(ctx context.Context, id string) ([]Episode, error) {
	var out []Episode
	p := url.Values{}
	p.Set("extended", "full")
	if err := c.get(ctx, "/anime/episodes/"+url.PathEscape(id), p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TrendingAnime returns trending anime.
func (c *Client) TrendingAnime(ctx context.Context, interval string) ([]TrendingAnime, error) {
	path := "/anime/trending"
	if interval != "" {
		path += "/" + url.PathEscape(interval)
	}
	var out []TrendingAnime
	if err := c.get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AnimeGenres returns anime filtered by genre.
func (c *Client) AnimeGenres(ctx context.Context, genre string, page, limit int) ([]GenreItem, error) {
	if genre == "" {
		genre = defaultFilter
	}
	var out []GenreItem
	if err := c.get(ctx, "/anime/genres/"+url.PathEscape(genre), pageParams(page, limit), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AnimePremieres returns latest anime premieres.
func (c *Client) AnimePremieres(ctx context.Context, param string, page, limit int) ([]PremiereItem, error) {
	if param == "" {
		param = "new"
	}
	var out []PremiereItem
	if err := c.get(ctx, "/anime/premieres/"+url.PathEscape(param), pageParams(page, limit), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AiringAnime returns currently airing anime.
func (c *Client) AiringAnime(ctx context.Context, date, sort string) ([]AiringItem, error) {
	p := url.Values{}
	if date != "" {
		p.Set("date", date)
	}
	if sort != "" {
		p.Set("sort", sort)
	}
	var out []AiringItem
	if err := c.get(ctx, "/anime/airing", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BestAnime returns the best of all anime.
func (c *Client) BestAnime(ctx context.Context, filter string) ([]BestItem, error) {
	if filter == "" {
		filter = defaultFilter
	}
	var out []BestItem
	if err := c.get(ctx, "/anime/best/"+url.PathEscape(filter), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Search.

// SearchByID resolves an external ID to Simkl entries.
// Supported ID types: imdb, tvdb, tmdb, anidb, mal, anilist, hulu, netflix,
// crunchyroll, kitsu, livechart, anisearch, animeplanet, traktslug, letterboxd.
func (c *Client) SearchByID(ctx context.Context, idType, idValue string) ([]SearchIDResult, error) {
	p := url.Values{}
	p.Set(idType, idValue)
	var out []SearchIDResult
	if err := c.get(ctx, "/search/id", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchText searches for movies, shows, or anime by text query.
// searchType should be "tv", "anime", or "movie".
func (c *Client) SearchText(ctx context.Context, searchType, query string, page, limit int) ([]SearchResult, error) {
	p := pageParams(page, limit)
	p.Set("q", query)
	p.Set("extended", "full")
	var out []SearchResult
	if err := c.get(ctx, "/search/"+url.PathEscape(searchType), p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Calendar.
// Calendar endpoints use a separate CDN URL and do not require API authentication.

// CalendarShows returns the next 33 days of TV show releases.
func (c *Client) CalendarShows(ctx context.Context) ([]CalendarShow, error) {
	var out []CalendarShow
	if err := c.getCal(ctx, "/calendar/tv.json", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CalendarShowsMonth returns TV show releases for a specific month.
func (c *Client) CalendarShowsMonth(ctx context.Context, year, month int) ([]CalendarShow, error) {
	path := fmt.Sprintf("/calendar/%d/%d/tv.json", year, month)
	var out []CalendarShow
	if err := c.getCal(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CalendarAnime returns the next 33 days of anime releases.
func (c *Client) CalendarAnime(ctx context.Context) ([]CalendarAnime, error) {
	var out []CalendarAnime
	if err := c.getCal(ctx, "/calendar/anime.json", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CalendarAnimeMonth returns anime releases for a specific month.
func (c *Client) CalendarAnimeMonth(ctx context.Context, year, month int) ([]CalendarAnime, error) {
	path := fmt.Sprintf("/calendar/%d/%d/anime.json", year, month)
	var out []CalendarAnime
	if err := c.getCal(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CalendarMovies returns the next 33 days of movie releases.
func (c *Client) CalendarMovies(ctx context.Context) ([]CalendarMovie, error) {
	var out []CalendarMovie
	if err := c.getCal(ctx, "/calendar/movie_release.json", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CalendarMoviesMonth returns movie releases for a specific month.
func (c *Client) CalendarMoviesMonth(ctx context.Context, year, month int) ([]CalendarMovie, error) {
	path := fmt.Sprintf("/calendar/%d/%d/movie_release.json", year, month)
	var out []CalendarMovie
	if err := c.getCal(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
