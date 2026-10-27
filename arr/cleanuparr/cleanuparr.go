package cleanuparr

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/golusoris/goenvoy/arr/v2"
)

const (
	// MaxRecentStrikes is the maximum count accepted by [Client.GetRecentStrikes].
	MaxRecentStrikes = 50

	summaryPageSize = 500
	maxSummaryPages = 4
)

// Option configures a [Client].
type Option = arr.Option

// WithHTTPClient sets a custom [http.Client].
func WithHTTPClient(c *http.Client) Option {
	return arr.WithHTTPClient(c)
}

// WithTimeout overrides the default HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return arr.WithTimeout(d)
}

// WithUserAgent sets the User-Agent header sent with every request.
func WithUserAgent(ua string) Option {
	return arr.WithUserAgent(ua)
}

// Client is a read-only client for the Cleanuparr API.
type Client struct {
	base *arr.BaseClient
}

// New creates a new Cleanuparr [Client] for the server at baseURL.
// The apiKey is sent in the X-Api-Key header on every request.
func New(baseURL, apiKey string, opts ...arr.Option) (*Client, error) {
	base, err := arr.NewBaseClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, fmt.Errorf("cleanuparr: new client: %w", err)
	}
	return &Client{base: base}, nil
}

// GetStatus returns the current Cleanuparr server status.
func (c *Client) GetStatus(ctx context.Context) (*Status, error) {
	var out Status
	if err := c.base.Get(ctx, "/api/status", &out); err != nil {
		return nil, fmt.Errorf("cleanuparr: get status: %w", err)
	}
	return &out, nil
}

// GetRecentStrikes returns the most recent strikes, clamping count to [1, MaxRecentStrikes].
func (c *Client) GetRecentStrikes(ctx context.Context, count int) ([]Strike, error) {
	if count < 1 {
		count = 1
	} else if count > MaxRecentStrikes {
		count = MaxRecentStrikes
	}
	path := fmt.Sprintf("/api/strikes/recent?count=%d", count)
	var out []Strike
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, fmt.Errorf("cleanuparr: get recent strikes: %w", err)
	}
	return out, nil
}

// GetStrikes returns a paginated page of download item strikes.
func (c *Client) GetStrikes(ctx context.Context, page, pageSize int) (*StrikePage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	} else if pageSize > summaryPageSize {
		pageSize = summaryPageSize
	}
	path := fmt.Sprintf("/api/strikes?page=%d&pageSize=%d", page, pageSize)
	var out StrikePage
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, fmt.Errorf("cleanuparr: get strikes: %w", err)
	}
	return &out, nil
}

// aggregateStrikeItem updates summary counts for a single download item.
func aggregateStrikeItem(sum *StrikeSummary, item DownloadItemStrikes) {
	if item.IsMarkedForRemoval {
		sum.MarkedForRemoval++
	}
	if item.IsRemoved {
		sum.RemovedDownloads++
	}
	if item.IsReturning {
		sum.ReturningDownloads++
	}
}

// aggregateStrikePage processes a page of download item strikes.
func aggregateStrikePage(sum *StrikeSummary, page *StrikePage) {
	sum.StruckDownloads = page.TotalCount
	for _, item := range page.Items {
		aggregateStrikeItem(sum, item)
	}
}

// SummarizeStrikes walks up to 4 pages of strikes with pageSize 500
// (HISS-02 scalar bound to avoid unbounded pagination) and aggregates
// strike counts across all struck downloads.
func (c *Client) SummarizeStrikes(ctx context.Context) (*StrikeSummary, error) {
	var sum StrikeSummary
	for page := 1; page <= maxSummaryPages; page++ {
		out, err := c.GetStrikes(ctx, page, summaryPageSize)
		if err != nil {
			return nil, fmt.Errorf("cleanuparr: summarize strikes: %w", err)
		}
		aggregateStrikePage(&sum, out)
		if page >= out.TotalPages {
			break
		}
		if page == maxSummaryPages {
			sum.Partial = true
		}
	}
	return &sum, nil
}

// GetSeekerConfig returns the seeker configuration from GET /api/configuration/seeker.
func (c *Client) GetSeekerConfig(ctx context.Context) (*SeekerConfig, error) {
	var out SeekerConfig
	if err := c.base.Get(ctx, "/api/configuration/seeker", &out); err != nil {
		return nil, fmt.Errorf("cleanuparr: get seeker config: %w", err)
	}
	return &out, nil
}

// GetGeneralConfig returns the general configuration from GET /api/configuration/general.
func (c *Client) GetGeneralConfig(ctx context.Context) (*GeneralConfig, error) {
	var out GeneralConfig
	if err := c.base.Get(ctx, "/api/configuration/general", &out); err != nil {
		return nil, fmt.Errorf("cleanuparr: get general config: %w", err)
	}
	return &out, nil
}

// GetArrConfig returns the *arr application configuration from GET /api/configuration/{kind}.
func (c *Client) GetArrConfig(ctx context.Context, kind ArrKind) (*ArrConfig, error) {
	switch kind {
	case ArrKindSonarr, ArrKindRadarr, ArrKindLidarr, ArrKindReadarr, ArrKindWhisparr, ArrKindSportarr, ArrKindLazyLibrarian:
	default:
		return nil, fmt.Errorf("cleanuparr: get arr config: unknown kind %q", kind)
	}
	path := "/api/configuration/" + url.PathEscape(string(kind))
	var out ArrConfig
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, fmt.Errorf("cleanuparr: get arr config: %w", err)
	}
	return &out, nil
}
