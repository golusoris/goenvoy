package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQueue returns the download queue (paged).
func (c *api) GetQueue(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.QueueRecord], error) {
	var out arr.PagingResource[arr.QueueRecord]
	path := fmt.Sprintf("/api/v3/queue?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteQueueItem removes an item from the download queue.
func (c *api) DeleteQueueItem(ctx context.Context, id int, removeFromClient, blocklist bool) error {
	path := fmt.Sprintf("/api/v3/queue/%d?removeFromClient=%t&blocklist=%t", id, removeFromClient, blocklist)
	return c.base.Delete(ctx, path, nil, nil)
}

// BulkDeleteQueue removes multiple items from the download queue.
func (c *api) BulkDeleteQueue(ctx context.Context, bulk *arr.QueueBulkResource) error {
	return c.base.Delete(ctx, "/api/v3/queue/bulk", bulk, nil)
}

// GrabQueueItem sends a queue item to the download client.
func (c *api) GrabQueueItem(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v3/queue/grab/%d", id), nil, nil)
}

// BulkGrabQueue grabs multiple items from the queue.
func (c *api) BulkGrabQueue(ctx context.Context, ids []int) error {
	return c.base.Post(ctx, "/api/v3/queue/grab/bulk", struct {
		IDs []int `json:"ids"`
	}{IDs: ids}, nil)
}

// GetQueueDetails returns detailed queue information.
func (c *api) GetQueueDetails(ctx context.Context) ([]arr.QueueRecord, error) {
	return arr.GetList[arr.QueueRecord](ctx, c.base, "/api/v3/queue/details")
}

// GetQueueStatus returns the queue status summary.
func (c *api) GetQueueStatus(ctx context.Context) (*arr.QueueStatusResource, error) {
	return arr.GetOne[arr.QueueStatusResource](ctx, c.base, "/api/v3/queue/status")
}
