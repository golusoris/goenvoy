package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetBlocklist returns the blocklist (paged).
func (c *api) GetBlocklist(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.BlocklistResource], error) {
	path := fmt.Sprintf("/api/v3/blocklist?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[arr.BlocklistResource]](ctx, c.base, path)
}

// DeleteBlocklistItem deletes a single blocklist entry.
func (c *api) DeleteBlocklistItem(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/blocklist/%d", id), nil, nil)
}

// BulkDeleteBlocklist deletes multiple blocklist entries.
func (c *api) BulkDeleteBlocklist(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v3/blocklist/bulk", arr.BlocklistBulkResource{IDs: ids}, nil)
}
