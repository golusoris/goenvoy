package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetBlocklist returns the blocklist (paged).
func (c *ClientV3) GetBlocklist(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.BlocklistResource], error) {
	var out arr.PagingResource[arr.BlocklistResource]
	path := fmt.Sprintf("/api/v3/blocklist?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBlocklistItem deletes a single blocklist entry.
func (c *ClientV3) DeleteBlocklistItem(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/blocklist/%d", id), nil, nil)
}

// BulkDeleteBlocklist deletes multiple blocklist entries.
func (c *ClientV3) BulkDeleteBlocklist(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v3/blocklist/bulk", arr.BlocklistBulkResource{IDs: ids}, nil)
}

// GetBlocklistMovie returns blocklist entries for a specific movie.
func (c *ClientV3) GetBlocklistMovie(ctx context.Context, movieID int) ([]arr.BlocklistResource, error) {
	var out []arr.BlocklistResource
	path := fmt.Sprintf("/api/v3/blocklist/movie?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Calendar Extended ----------.
