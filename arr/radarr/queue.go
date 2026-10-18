package radarr

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/golusoris/goenvoy/arr/v2"
)

// QueueRecord represents a download queue item in Radarr with movie details.
type QueueRecord struct {
	arr.QueueRecord
	CustomFormatScore int    `json:"customFormatScore"`
	MovieID           int    `json:"movieId"`
	Movie             *Movie `json:"movie,omitempty"`
}

// QueueOptions specifies filtering and inclusion parameters for GetQueueWithOptions.
//
// Field mappings to Radarr query parameters:
//   - Page: page
//   - PageSize: pageSize
//   - IncludeUnknown: includeUnknownMovieItems
//   - IncludeParent: ignored (Radarr has no parent entity in queue)
//   - IncludeItem: includeMovie
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

// GetQueue returns the current download queue with pagination.
func (c *Client) GetQueue(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.QueueRecord], error) {
	path := fmt.Sprintf("/api/v3/queue?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[arr.QueueRecord]](ctx, c.base, path)
}

// GetQueueWithOptions returns the download queue using advanced filtering and inclusion options.
func (c *Client) GetQueueWithOptions(ctx context.Context, opts QueueOptions) (*arr.PagingResource[QueueRecord], error) {
	q := make(url.Values)
	q.Set("page", strconv.Itoa(opts.Page))
	q.Set("pageSize", strconv.Itoa(opts.PageSize))
	if opts.IncludeUnknown {
		q.Set("includeUnknownMovieItems", "true")
	}
	// opts.IncludeParent is ignored because Radarr has no parent entity in queue.
	if opts.IncludeItem {
		q.Set("includeMovie", "true")
	}

	path := "/api/v3/queue?" + q.Encode()
	return arr.GetOne[arr.PagingResource[QueueRecord]](ctx, c.base, path)
}

// DeleteQueueItem removes an item from the download queue.
func (c *Client) DeleteQueueItem(ctx context.Context, id int, removeFromClient, blocklist bool) error {
	path := fmt.Sprintf("/api/v3/queue/%d?removeFromClient=%t&blocklist=%t", id, removeFromClient, blocklist)
	return c.base.Delete(ctx, path, nil, nil)
}

// DeleteQueueItemWithOptions removes an item from the download queue with extended options.
func (c *Client) DeleteQueueItemWithOptions(ctx context.Context, id int, opts QueueDeleteOptions) error {
	path := fmt.Sprintf("/api/v3/queue/%d?removeFromClient=%t&blocklist=%t&skipRedownload=%t&changeCategory=%t",
		id, opts.RemoveFromClient, opts.Blocklist, opts.SkipRedownload, opts.ChangeCategory)
	return c.base.Delete(ctx, path, nil, nil)
}

// DeleteQueueItems removes multiple items from the download queue.
func (c *Client) DeleteQueueItems(ctx context.Context, ids []int, removeFromClient, blocklist bool) error {
	path := fmt.Sprintf("/api/v3/queue/bulk?removeFromClient=%t&blocklist=%t", removeFromClient, blocklist)
	return c.base.Delete(ctx, path, &arr.QueueBulkResource{IDs: ids}, nil)
}

// GrabQueueItem forces a grab for a queued item.
func (c *Client) GrabQueueItem(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v3/queue/grab/%d", id), nil, nil)
}

// GrabQueueItems forces grabs for multiple queued items.
func (c *Client) GrabQueueItems(ctx context.Context, ids []int) error {
	return c.base.Post(ctx, "/api/v3/queue/grab/bulk", &arr.QueueBulkResource{IDs: ids}, nil)
}

// GetQueueDetails returns all queue items without pagination.
func (c *Client) GetQueueDetails(ctx context.Context) ([]arr.QueueRecord, error) {
	return arr.GetList[arr.QueueRecord](ctx, c.base, "/api/v3/queue/details")
}

// GetQueueStatus returns the overall queue status counts.
func (c *Client) GetQueueStatus(ctx context.Context) (*arr.QueueStatusResource, error) {
	return arr.GetOne[arr.QueueStatusResource](ctx, c.base, "/api/v3/queue/status")
}

// ---------- History extras ----------.
