package readarr

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/golusoris/goenvoy/arr/v2"
)

// QueueRecord represents a download queue item in Readarr with author and book details.
type QueueRecord struct {
	arr.QueueRecord
	CustomFormatScore int     `json:"customFormatScore"`
	AuthorID          int     `json:"authorId"`
	BookID            int     `json:"bookId"`
	Author            *Author `json:"author,omitempty"`
	Book              *Book   `json:"book,omitempty"`
}

// QueueOptions specifies filtering and inclusion parameters for GetQueueWithOptions.
//
// Field mappings to Readarr query parameters:
//   - Page: page
//   - PageSize: pageSize
//   - IncludeUnknown: includeUnknownAuthorItems
//   - IncludeParent: includeAuthor
//   - IncludeItem: includeBook
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
	path := fmt.Sprintf("/api/v1/queue?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[arr.QueueRecord]](ctx, c.base, path)
}

// GetQueueWithOptions returns the download queue using advanced filtering and inclusion options.
func (c *Client) GetQueueWithOptions(ctx context.Context, opts QueueOptions) (*arr.PagingResource[QueueRecord], error) {
	q := make(url.Values)
	q.Set("page", strconv.Itoa(opts.Page))
	q.Set("pageSize", strconv.Itoa(opts.PageSize))
	if opts.IncludeUnknown {
		q.Set("includeUnknownAuthorItems", "true")
	}
	if opts.IncludeParent {
		q.Set("includeAuthor", "true")
	}
	if opts.IncludeItem {
		q.Set("includeBook", "true")
	}

	path := "/api/v1/queue?" + q.Encode()
	return arr.GetOne[arr.PagingResource[QueueRecord]](ctx, c.base, path)
}

// DeleteQueueItem removes an item from the download queue.
func (c *Client) DeleteQueueItem(ctx context.Context, id int, removeFromClient, blocklist bool) error {
	path := fmt.Sprintf("/api/v1/queue/%d?removeFromClient=%t&blocklist=%t", id, removeFromClient, blocklist)
	return c.base.Delete(ctx, path, nil, nil)
}

// DeleteQueueItemWithOptions removes an item from the download queue with extended options.
func (c *Client) DeleteQueueItemWithOptions(ctx context.Context, id int, opts QueueDeleteOptions) error {
	path := fmt.Sprintf("/api/v1/queue/%d?removeFromClient=%t&blocklist=%t&skipRedownload=%t&changeCategory=%t",
		id, opts.RemoveFromClient, opts.Blocklist, opts.SkipRedownload, opts.ChangeCategory)
	return c.base.Delete(ctx, path, nil, nil)
}

// BulkDeleteQueue deletes multiple items from the download queue.
func (c *Client) BulkDeleteQueue(ctx context.Context, bulk *arr.QueueBulkResource, removeFromClient, blocklist bool) error {
	path := fmt.Sprintf("/api/v1/queue/bulk?removeFromClient=%t&blocklist=%t", removeFromClient, blocklist)
	return c.base.Delete(ctx, path, bulk, nil)
}

// GrabQueueItem grabs a pending release from the queue by its ID.
func (c *Client) GrabQueueItem(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v1/queue/grab/%d", id), nil, nil)
}

// GrabQueueItemsBulk grabs multiple pending releases from the queue.
func (c *Client) GrabQueueItemsBulk(ctx context.Context, ids []int) error {
	return c.base.Post(ctx, "/api/v1/queue/grab/bulk", struct {
		IDs []int `json:"ids"`
	}{IDs: ids}, nil)
}

// GetQueueDetails returns detailed information about all items in the queue.
func (c *Client) GetQueueDetails(ctx context.Context) ([]arr.QueueRecord, error) {
	return arr.GetList[arr.QueueRecord](ctx, c.base, "/api/v1/queue/details")
}

// GetQueueStatus returns the overall status of the download queue.
func (c *Client) GetQueueStatus(ctx context.Context) (*arr.QueueStatusResource, error) {
	return arr.GetOne[arr.QueueStatusResource](ctx, c.base, "/api/v1/queue/status")
}

// ---------- History Extended ----------.
