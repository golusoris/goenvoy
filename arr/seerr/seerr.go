package seerr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// Client is a Seerr API client.
type Client struct {
	base *arr.BaseClient
}

// New creates a Seerr [Client] for the instance at baseURL.
func New(baseURL, apiKey string, opts ...arr.Option) (*Client, error) {
	base, err := arr.NewBaseClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{base: base}, nil
}

// GetStatus returns the current Seerr server status.
func (c *Client) GetStatus(ctx context.Context) (*StatusResponse, error) {
	return arr.GetOne[StatusResponse](ctx, c.base, "/api/v1/status")
}

// GetMe returns the currently authenticated user.
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	return arr.GetOne[User](ctx, c.base, "/api/v1/auth/me")
}

// Search performs a multi-search for movies, TV shows, and people.
func (c *Client) Search(ctx context.Context, query string, page int) (*SearchResults, error) {
	path := fmt.Sprintf("/api/v1/search?query=%s&page=%d", url.QueryEscape(query), page)
	return arr.GetOne[SearchResults](ctx, c.base, path)
}

// DiscoverMovies returns a paginated list of discovered movies.
func (c *Client) DiscoverMovies(ctx context.Context, page int) (*SearchResults, error) {
	path := fmt.Sprintf("/api/v1/discover/movies?page=%d", page)
	return arr.GetOne[SearchResults](ctx, c.base, path)
}

// DiscoverTV returns a paginated list of discovered TV shows.
func (c *Client) DiscoverTV(ctx context.Context, page int) (*SearchResults, error) {
	path := fmt.Sprintf("/api/v1/discover/tv?page=%d", page)
	return arr.GetOne[SearchResults](ctx, c.base, path)
}

// DiscoverTrending returns currently trending movies and TV shows.
func (c *Client) DiscoverTrending(ctx context.Context, page int) (*SearchResults, error) {
	path := fmt.Sprintf("/api/v1/discover/trending?page=%d", page)
	return arr.GetOne[SearchResults](ctx, c.base, path)
}

// DiscoverUpcomingMovies returns upcoming movies.
func (c *Client) DiscoverUpcomingMovies(ctx context.Context, page int) (*SearchResults, error) {
	path := fmt.Sprintf("/api/v1/discover/movies/upcoming?page=%d", page)
	return arr.GetOne[SearchResults](ctx, c.base, path)
}

// DiscoverUpcomingTV returns upcoming TV shows.
func (c *Client) DiscoverUpcomingTV(ctx context.Context, page int) (*SearchResults, error) {
	path := fmt.Sprintf("/api/v1/discover/tv/upcoming?page=%d", page)
	return arr.GetOne[SearchResults](ctx, c.base, path)
}

// pagedRequests is the list-requests response shape.
type pagedRequests struct {
	PageInfo PageInfo       `json:"pageInfo"`
	Results  []MediaRequest `json:"results,omitempty"`
}

// GetRequests returns all media requests with pagination and optional filter.
func (c *Client) GetRequests(ctx context.Context, take, skip int, filter string) ([]MediaRequest, *PageInfo, error) {
	var out pagedRequests
	path := fmt.Sprintf("/api/v1/request?take=%d&skip=%d", take, skip)
	if filter != "" {
		path += "&filter=" + url.QueryEscape(filter)
	}
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, nil, err
	}
	return out.Results, &out.PageInfo, nil
}

// GetRequest returns a single media request by its ID.
func (c *Client) GetRequest(ctx context.Context, id int) (*MediaRequest, error) {
	path := fmt.Sprintf("/api/v1/request/%d", id)
	return arr.GetOne[MediaRequest](ctx, c.base, path)
}

// CreateRequest creates a new media request.
func (c *Client) CreateRequest(ctx context.Context, body *CreateRequestBody) (*MediaRequest, error) {
	return arr.PostOne[MediaRequest](ctx, c.base, "/api/v1/request", body)
}

// DeleteRequest removes a media request by its ID.
func (c *Client) DeleteRequest(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/request/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// ApproveRequest approves a pending request.
func (c *Client) ApproveRequest(ctx context.Context, id int) (*MediaRequest, error) {
	path := fmt.Sprintf("/api/v1/request/%d/approve", id)
	return arr.PostOne[MediaRequest](ctx, c.base, path, nil)
}

// DeclineRequest declines a pending request.
func (c *Client) DeclineRequest(ctx context.Context, id int) (*MediaRequest, error) {
	path := fmt.Sprintf("/api/v1/request/%d/decline", id)
	return arr.PostOne[MediaRequest](ctx, c.base, path, nil)
}

// RetryRequest retries a failed request.
func (c *Client) RetryRequest(ctx context.Context, id int) (*MediaRequest, error) {
	path := fmt.Sprintf("/api/v1/request/%d/retry", id)
	return arr.PostOne[MediaRequest](ctx, c.base, path, nil)
}

// GetRequestCount returns aggregate request count statistics.
func (c *Client) GetRequestCount(ctx context.Context) (*RequestCount, error) {
	return arr.GetOne[RequestCount](ctx, c.base, "/api/v1/request/count")
}

// GetMovie returns full details for a movie by its TMDB ID.
func (c *Client) GetMovie(ctx context.Context, tmdbID int) (*MovieDetails, error) {
	path := fmt.Sprintf("/api/v1/movie/%d", tmdbID)
	return arr.GetOne[MovieDetails](ctx, c.base, path)
}

// GetMovieRecommendations returns recommended movies for a given TMDB ID.
func (c *Client) GetMovieRecommendations(ctx context.Context, tmdbID, page int) (*SearchResults, error) {
	path := fmt.Sprintf("/api/v1/movie/%d/recommendations?page=%d", tmdbID, page)
	return arr.GetOne[SearchResults](ctx, c.base, path)
}

// GetMovieSimilar returns similar movies for a given TMDB ID.
func (c *Client) GetMovieSimilar(ctx context.Context, tmdbID, page int) (*SearchResults, error) {
	path := fmt.Sprintf("/api/v1/movie/%d/similar?page=%d", tmdbID, page)
	return arr.GetOne[SearchResults](ctx, c.base, path)
}

// GetTV returns full details for a TV show by its TMDB ID.
func (c *Client) GetTV(ctx context.Context, tmdbID int) (*TvDetails, error) {
	path := fmt.Sprintf("/api/v1/tv/%d", tmdbID)
	return arr.GetOne[TvDetails](ctx, c.base, path)
}

// GetTVSeason returns details and episodes for a specific season.
func (c *Client) GetTVSeason(ctx context.Context, tmdbID, seasonNumber int) (*Season, error) {
	path := fmt.Sprintf("/api/v1/tv/%d/season/%d", tmdbID, seasonNumber)
	return arr.GetOne[Season](ctx, c.base, path)
}

// GetTVRecommendations returns recommended TV shows for a given TMDB ID.
func (c *Client) GetTVRecommendations(ctx context.Context, tmdbID, page int) (*SearchResults, error) {
	path := fmt.Sprintf("/api/v1/tv/%d/recommendations?page=%d", tmdbID, page)
	return arr.GetOne[SearchResults](ctx, c.base, path)
}

// GetTVSimilar returns similar TV shows for a given TMDB ID.
func (c *Client) GetTVSimilar(ctx context.Context, tmdbID, page int) (*SearchResults, error) {
	path := fmt.Sprintf("/api/v1/tv/%d/similar?page=%d", tmdbID, page)
	return arr.GetOne[SearchResults](ctx, c.base, path)
}

// pagedMedia is the list-media response shape.
type pagedMedia struct {
	PageInfo PageInfo    `json:"pageInfo"`
	Results  []MediaInfo `json:"results,omitempty"`
}

// GetMedia returns a paginated list of all media items.
func (c *Client) GetMedia(ctx context.Context, take, skip int) ([]MediaInfo, *PageInfo, error) {
	var out pagedMedia
	path := fmt.Sprintf("/api/v1/media?take=%d&skip=%d", take, skip)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, nil, err
	}
	return out.Results, &out.PageInfo, nil
}

// DeleteMedia removes a media item by its ID.
func (c *Client) DeleteMedia(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/media/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// UpdateMediaStatus updates a media item's availability status.
func (c *Client) UpdateMediaStatus(ctx context.Context, id int, status string, is4k bool) (*MediaInfo, error) {
	path := fmt.Sprintf("/api/v1/media/%d/%s", id, status)
	body := map[string]bool{"is4k": is4k}
	return arr.PostOne[MediaInfo](ctx, c.base, path, body)
}

// pagedUsers is the list-users response shape.
type pagedUsers struct {
	PageInfo PageInfo `json:"pageInfo"`
	Results  []User   `json:"results,omitempty"`
}

// GetUsers returns a paginated list of users.
func (c *Client) GetUsers(ctx context.Context, take, skip int) ([]User, *PageInfo, error) {
	var out pagedUsers
	path := fmt.Sprintf("/api/v1/user?take=%d&skip=%d", take, skip)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, nil, err
	}
	return out.Results, &out.PageInfo, nil
}

// GetUser returns a single user by their ID.
func (c *Client) GetUser(ctx context.Context, id int) (*User, error) {
	path := fmt.Sprintf("/api/v1/user/%d", id)
	return arr.GetOne[User](ctx, c.base, path)
}

// DeleteUser removes a user by their ID.
func (c *Client) DeleteUser(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/user/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// GetUserQuota returns the request quota for a specific user.
func (c *Client) GetUserQuota(ctx context.Context, userID int) (*UserQuota, error) {
	path := fmt.Sprintf("/api/v1/user/%d/quota", userID)
	return arr.GetOne[UserQuota](ctx, c.base, path)
}

// pagedIssues is the list-issues response shape.
type pagedIssues struct {
	PageInfo PageInfo `json:"pageInfo"`
	Results  []Issue  `json:"results,omitempty"`
}

// GetIssues returns a paginated list of issues.
func (c *Client) GetIssues(ctx context.Context, take, skip int, filter string) ([]Issue, *PageInfo, error) {
	var out pagedIssues
	path := fmt.Sprintf("/api/v1/issue?take=%d&skip=%d", take, skip)
	if filter != "" {
		path += "&filter=" + url.QueryEscape(filter)
	}
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, nil, err
	}
	return out.Results, &out.PageInfo, nil
}

// GetIssue returns a single issue by its ID.
func (c *Client) GetIssue(ctx context.Context, id int) (*Issue, error) {
	path := fmt.Sprintf("/api/v1/issue/%d", id)
	return arr.GetOne[Issue](ctx, c.base, path)
}

// CreateIssue creates a new issue report.
func (c *Client) CreateIssue(ctx context.Context, body *CreateIssueBody) (*Issue, error) {
	return arr.PostOne[Issue](ctx, c.base, "/api/v1/issue", body)
}

// DeleteIssue removes an issue by its ID.
func (c *Client) DeleteIssue(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/issue/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// AddIssueComment adds a comment to an existing issue.
func (c *Client) AddIssueComment(ctx context.Context, issueID int, message string) (*Issue, error) {
	path := fmt.Sprintf("/api/v1/issue/%d/comment", issueID)
	body := map[string]string{"message": message}
	return arr.PostOne[Issue](ctx, c.base, path, body)
}

// ResolveIssue marks an issue as resolved.
func (c *Client) ResolveIssue(ctx context.Context, id int) (*Issue, error) {
	path := fmt.Sprintf("/api/v1/issue/%d/resolved", id)
	return arr.PostOne[Issue](ctx, c.base, path, nil)
}

// ReopenIssue reopens a resolved issue.
func (c *Client) ReopenIssue(ctx context.Context, id int) (*Issue, error) {
	path := fmt.Sprintf("/api/v1/issue/%d/open", id)
	return arr.PostOne[Issue](ctx, c.base, path, nil)
}

// GetIssueCount returns aggregate issue count statistics.
func (c *Client) GetIssueCount(ctx context.Context) (*IssueCount, error) {
	return arr.GetOne[IssueCount](ctx, c.base, "/api/v1/issue/count")
}

// GetSonarrSettings returns the configured Sonarr server settings in Seerr.
func (c *Client) GetSonarrSettings(ctx context.Context) ([]ServiceSettings, error) {
	return arr.GetList[ServiceSettings](ctx, c.base, "/api/v1/settings/sonarr")
}

// GetRadarrSettings returns the configured Radarr server settings in Seerr.
func (c *Client) GetRadarrSettings(ctx context.Context) ([]ServiceSettings, error) {
	return arr.GetList[ServiceSettings](ctx, c.base, "/api/v1/settings/radarr")
}
