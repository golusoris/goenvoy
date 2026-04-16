package whisparr

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQueue returns the download queue (paged).
func (c *ClientV3) GetQueue(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.QueueRecord], error) {
	var out arr.PagingResource[arr.QueueRecord]
	path := fmt.Sprintf("/api/v3/queue?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteQueueItem removes an item from the download queue.
func (c *ClientV3) DeleteQueueItem(ctx context.Context, id int, removeFromClient, blocklist bool) error {
	path := fmt.Sprintf("/api/v3/queue/%d?removeFromClient=%t&blocklist=%t", id, removeFromClient, blocklist)
	return c.base.Delete(ctx, path, nil, nil)
}

// BulkDeleteQueue removes multiple items from the download queue.
func (c *ClientV3) BulkDeleteQueue(ctx context.Context, bulk *arr.QueueBulkResource) error {
	return c.base.Delete(ctx, "/api/v3/queue/bulk", bulk, nil)
}

// GrabQueueItem sends a queue item to the download client.
func (c *ClientV3) GrabQueueItem(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v3/queue/grab/%d", id), nil, nil)
}

// BulkGrabQueue grabs multiple items from the queue.
func (c *ClientV3) BulkGrabQueue(ctx context.Context, ids []int) error {
	return c.base.Post(ctx, "/api/v3/queue/grab/bulk", struct {
		IDs []int `json:"ids"`
	}{IDs: ids}, nil)
}

// GetQueueDetails returns detailed queue information.
func (c *ClientV3) GetQueueDetails(ctx context.Context) ([]arr.QueueRecord, error) {
	var out []arr.QueueRecord
	if err := c.base.Get(ctx, "/api/v3/queue/details", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetQueueDetailsByMovieID returns detailed queue info for specific movie IDs.
func (c *ClientV3) GetQueueDetailsByMovieID(ctx context.Context, movieIDs []int) ([]arr.QueueRecord, error) {
	var out []arr.QueueRecord
	vals := make(url.Values)
	for _, id := range movieIDs {
		vals.Add("movieIds", strconv.Itoa(id))
	}
	path := "/api/v3/queue/details?" + vals.Encode()
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetQueueStatus returns the queue status summary.
func (c *ClientV3) GetQueueStatus(ctx context.Context) (*arr.QueueStatusResource, error) {
	var out arr.QueueStatusResource
	if err := c.base.Get(ctx, "/api/v3/queue/status", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Releases ----------.
