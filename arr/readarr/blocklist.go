package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetBlocklist returns the blocklist with pagination.
func (c *Client) GetBlocklist(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.BlocklistResource], error) {
	path := fmt.Sprintf("/api/v1/blocklist?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[arr.BlocklistResource]](ctx, c.base, path)
}

// DeleteBlocklistItem deletes a single blocklist item by ID.
func (c *Client) DeleteBlocklistItem(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/blocklist/%d", id), nil, nil)
}

// BulkDeleteBlocklist deletes multiple blocklist items.
func (c *Client) BulkDeleteBlocklist(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v1/blocklist/bulk", &arr.BlocklistBulkResource{IDs: ids}, nil)
}

// ---------- Queue Extended ----------.
