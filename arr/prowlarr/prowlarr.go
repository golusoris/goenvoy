package prowlarr

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/lusoris/goenvoy/arr"
)

// Client is a Prowlarr API client.
type Client struct {
	base *arr.BaseClient
}

// New creates a Prowlarr [Client] for the instance at baseURL.
func New(baseURL, apiKey string, opts ...arr.Option) (*Client, error) {
	base, err := arr.NewBaseClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{base: base}, nil
}

// GetIndexers returns all configured indexers.
func (c *Client) GetIndexers(ctx context.Context) ([]Indexer, error) {
	var out []Indexer
	if err := c.base.Get(ctx, "/api/v1/indexer", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetIndexer returns a single indexer by its database ID.
func (c *Client) GetIndexer(ctx context.Context, id int) (*Indexer, error) {
	var out Indexer
	path := fmt.Sprintf("/api/v1/indexer/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddIndexer adds a new indexer to Prowlarr.
func (c *Client) AddIndexer(ctx context.Context, indexer *Indexer) (*Indexer, error) {
	var out Indexer
	if err := c.base.Post(ctx, "/api/v1/indexer", indexer, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateIndexer updates an existing indexer.
func (c *Client) UpdateIndexer(ctx context.Context, indexer *Indexer) (*Indexer, error) {
	var out Indexer
	path := fmt.Sprintf("/api/v1/indexer/%d", indexer.ID)
	if err := c.base.Put(ctx, path, indexer, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteIndexer removes an indexer by its database ID.
func (c *Client) DeleteIndexer(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/indexer/%d", id)
	return c.base.Delete(ctx, path, nil)
}

// GetIndexerCategories returns the global list of Newznab/Torznab categories.
func (c *Client) GetIndexerCategories(ctx context.Context) ([]IndexerCategory, error) {
	var out []IndexerCategory
	if err := c.base.Get(ctx, "/api/v1/indexer/categories", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetApplications returns all connected PVR applications.
func (c *Client) GetApplications(ctx context.Context) ([]Application, error) {
	var out []Application
	if err := c.base.Get(ctx, "/api/v1/applications", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetApplication returns a single application by its database ID.
func (c *Client) GetApplication(ctx context.Context, id int) (*Application, error) {
	var out Application
	path := fmt.Sprintf("/api/v1/applications/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddApplication adds a new application to Prowlarr.
func (c *Client) AddApplication(ctx context.Context, app *Application) (*Application, error) {
	var out Application
	if err := c.base.Post(ctx, "/api/v1/applications", app, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateApplication updates an existing application.
func (c *Client) UpdateApplication(ctx context.Context, app *Application) (*Application, error) {
	var out Application
	path := fmt.Sprintf("/api/v1/applications/%d", app.ID)
	if err := c.base.Put(ctx, path, app, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteApplication removes a connected application by its ID.
func (c *Client) DeleteApplication(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/applications/%d", id)
	return c.base.Delete(ctx, path, nil)
}

// GetAppProfiles returns all app profiles.
func (c *Client) GetAppProfiles(ctx context.Context) ([]AppProfile, error) {
	var out []AppProfile
	if err := c.base.Get(ctx, "/api/v1/appprofile", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAppProfile returns a single app profile by its ID.
func (c *Client) GetAppProfile(ctx context.Context, id int) (*AppProfile, error) {
	var out AppProfile
	path := fmt.Sprintf("/api/v1/appprofile/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddAppProfile creates a new app profile.
func (c *Client) AddAppProfile(ctx context.Context, profile *AppProfile) (*AppProfile, error) {
	var out AppProfile
	if err := c.base.Post(ctx, "/api/v1/appprofile", profile, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAppProfile updates an existing app profile.
func (c *Client) UpdateAppProfile(ctx context.Context, profile *AppProfile) (*AppProfile, error) {
	var out AppProfile
	path := fmt.Sprintf("/api/v1/appprofile/%d", profile.ID)
	if err := c.base.Put(ctx, path, profile, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAppProfile removes an app profile by its ID.
func (c *Client) DeleteAppProfile(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/appprofile/%d", id)
	return c.base.Delete(ctx, path, nil)
}

// SearchOptions configures a search request.
type SearchOptions struct {
	Query      string
	Type       string // "search", "tvsearch", "movie", "music", "book"
	IndexerIDs []int
	Categories []int
	Limit      int
	Offset     int
}

// Search performs a search across configured indexers.
func (c *Client) Search(ctx context.Context, opts *SearchOptions) ([]Release, error) {
	var out []Release
	params := url.Values{}
	if opts.Query != "" {
		params.Set("query", opts.Query)
	}
	if opts.Type != "" {
		params.Set("type", opts.Type)
	}
	if len(opts.IndexerIDs) > 0 {
		for _, id := range opts.IndexerIDs {
			params.Add("indexerIds", strconv.Itoa(id))
		}
	}
	if len(opts.Categories) > 0 {
		for _, cat := range opts.Categories {
			params.Add("categories", strconv.Itoa(cat))
		}
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	path := "/api/v1/search"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GrabRelease sends a release to the download client.
func (c *Client) GrabRelease(ctx context.Context, release *Release) (*Release, error) {
	var out Release
	if err := c.base.Post(ctx, "/api/v1/search", release, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDownloadClients returns all configured download clients.
func (c *Client) GetDownloadClients(ctx context.Context) ([]DownloadClientResource, error) {
	var out []DownloadClientResource
	if err := c.base.Get(ctx, "/api/v1/downloadclient", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SendCommand triggers a named command.
func (c *Client) SendCommand(ctx context.Context, cmd arr.CommandRequest) (*arr.CommandResponse, error) {
	var out arr.CommandResponse
	if err := c.base.Post(ctx, "/api/v1/command", cmd, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCommands returns all currently queued or running commands.
func (c *Client) GetCommands(ctx context.Context) ([]arr.CommandResponse, error) {
	var out []arr.CommandResponse
	if err := c.base.Get(ctx, "/api/v1/command", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCommand returns the status of a single command by its ID.
func (c *Client) GetCommand(ctx context.Context, id int) (*arr.CommandResponse, error) {
	var out arr.CommandResponse
	path := fmt.Sprintf("/api/v1/command/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHealth returns current health check results.
func (c *Client) GetHealth(ctx context.Context) ([]arr.HealthCheck, error) {
	var out []arr.HealthCheck
	if err := c.base.Get(ctx, "/api/v1/health", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSystemStatus returns Prowlarr system information.
func (c *Client) GetSystemStatus(ctx context.Context) (*arr.StatusResponse, error) {
	var out arr.StatusResponse
	if err := c.base.Get(ctx, "/api/v1/system/status", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTags returns all tags.
func (c *Client) GetTags(ctx context.Context) ([]arr.Tag, error) {
	var out []arr.Tag
	if err := c.base.Get(ctx, "/api/v1/tag", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateTag creates a new tag and returns it with its assigned ID.
func (c *Client) CreateTag(ctx context.Context, label string) (*arr.Tag, error) {
	var out arr.Tag
	if err := c.base.Post(ctx, "/api/v1/tag", arr.Tag{Label: label}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHistory returns the indexer history with pagination.
func (c *Client) GetHistory(ctx context.Context, page, pageSize int) (*arr.PagingResource[HistoryRecord], error) {
	var out arr.PagingResource[HistoryRecord]
	path := fmt.Sprintf("/api/v1/history?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHistorySince returns history events since a specific date (RFC 3339 timestamp).
func (c *Client) GetHistorySince(ctx context.Context, date string) ([]HistoryRecord, error) {
	var out []HistoryRecord
	path := "/api/v1/history/since?date=" + url.QueryEscape(date)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetHistoryByIndexer returns history events for a specific indexer.
func (c *Client) GetHistoryByIndexer(ctx context.Context, indexerID, limit int) ([]HistoryRecord, error) {
	var out []HistoryRecord
	path := fmt.Sprintf("/api/v1/history/indexer?indexerId=%d&limit=%d", indexerID, limit)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetIndexerStats returns aggregated indexer statistics for the given date range.
func (c *Client) GetIndexerStats(ctx context.Context, startDate, endDate string) (*IndexerStats, error) {
	var out IndexerStats
	params := []string{}
	if startDate != "" {
		params = append(params, "startDate="+url.QueryEscape(startDate))
	}
	if endDate != "" {
		params = append(params, "endDate="+url.QueryEscape(endDate))
	}
	path := "/api/v1/indexerstats"
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIndexerStatuses returns the status of all indexers (including disabled ones).
func (c *Client) GetIndexerStatuses(ctx context.Context) ([]IndexerStatus, error) {
	var out []IndexerStatus
	if err := c.base.Get(ctx, "/api/v1/indexerstatus", &out); err != nil {
		return nil, err
	}
	return out, nil
}
