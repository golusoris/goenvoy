package whisparr

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/golusoris/goenvoy/arr/v2"
)

// QueueRecord represents a download queue item in Whisparr.
//
// In Whisparr v2 (Sonarr-based), SeriesID, EpisodeID, SeasonNumber, Series, and Episode are populated.
// In Whisparr v3 (Radarr-based), MovieID and Movie are populated.
type QueueRecord struct {
	arr.QueueRecord
	CustomFormatScore int      `json:"customFormatScore"`
	MovieID           int      `json:"movieId,omitempty"`
	Movie             *Movie   `json:"movie,omitempty"`
	SeriesID          int      `json:"seriesId,omitempty"`
	EpisodeID         int      `json:"episodeId,omitempty"`
	SeasonNumber      int      `json:"seasonNumber,omitempty"`
	Series            *Series  `json:"series,omitempty"`
	Episode           *Episode `json:"episode,omitempty"`
}

// QueueOptions specifies filtering and inclusion parameters for GetQueueWithOptions.
//
// Field mappings to Whisparr query parameters:
//   - Page: page
//   - PageSize: pageSize
//   - IncludeUnknown: includeUnknownSeriesItems (v2) / includeUnknownMovieItems (v3)
//   - IncludeParent: includeSeries (v2) / ignored (v3)
//   - IncludeItem: includeEpisode (v2) / includeMovie (v3)
type QueueOptions struct {
	Page           int
	PageSize       int
	IncludeUnknown bool
	IncludeParent  bool
	IncludeItem    bool
}

// QueueDeleteOptions controls deletion behavior when removing an item from the queue.
type QueueDeleteOptions struct {
	RemoveFromClient bool
	Blocklist        bool
	SkipRedownload   bool
	ChangeCategory   bool
}

// GetQueueWithOptions returns the download queue using advanced filtering and inclusion options for Whisparr v2.
func (c *Client) GetQueueWithOptions(ctx context.Context, opts QueueOptions) (*arr.PagingResource[QueueRecord], error) {
	q := make(url.Values)
	q.Set("page", strconv.Itoa(opts.Page))
	q.Set("pageSize", strconv.Itoa(opts.PageSize))
	if opts.IncludeUnknown {
		q.Set("includeUnknownSeriesItems", "true")
	}
	if opts.IncludeParent {
		q.Set("includeSeries", "true")
	}
	if opts.IncludeItem {
		q.Set("includeEpisode", "true")
	}

	path := "/api/v3/queue?" + q.Encode()
	return arr.GetOne[arr.PagingResource[QueueRecord]](ctx, c.base, path)
}

// DeleteQueueItemWithOptions removes an item from the download queue with extended options for Whisparr v2.
func (c *Client) DeleteQueueItemWithOptions(ctx context.Context, id int, opts QueueDeleteOptions) error {
	path := fmt.Sprintf("/api/v3/queue/%d?removeFromClient=%t&blocklist=%t&skipRedownload=%t",
		id, opts.RemoveFromClient, opts.Blocklist, opts.SkipRedownload)
	return c.base.Delete(ctx, path, nil, nil)
}
