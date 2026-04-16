package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetBlocklist returns the blocklisted releases with pagination.
func (c *Client) GetBlocklist(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.BlocklistResource], error) {
	var out arr.PagingResource[arr.BlocklistResource]
	path := fmt.Sprintf("/api/v3/blocklist?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBlocklistItem removes a single blocklist entry.
func (c *Client) DeleteBlocklistItem(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/blocklist/%d", id), nil, nil)
}

// DeleteBlocklistBulk removes multiple blocklist entries at once.
func (c *Client) DeleteBlocklistBulk(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v3/blocklist/bulk", &arr.BlocklistBulkResource{IDs: ids}, nil)
}

// ---------- Custom Filters ----------.

// GetBlocklistMovies returns blocklisted movies.
func (c *Client) GetBlocklistMovies(ctx context.Context) ([]arr.BlocklistResource, error) {
	var out []arr.BlocklistResource
	if err := c.base.Get(ctx, "/api/v3/blocklist/movie", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Credit By ID ----------.
