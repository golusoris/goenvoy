package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetWantedMissing returns movies that are monitored but missing files.
func (c *Client) GetWantedMissing(ctx context.Context, page, pageSize int) (*arr.PagingResource[Movie], error) {
	path := fmt.Sprintf("/api/v3/wanted/missing?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[Movie]](ctx, c.base, path)
}

// GetWantedCutoff returns movies that have not met the quality cutoff.
func (c *Client) GetWantedCutoff(ctx context.Context, page, pageSize int) (*arr.PagingResource[Movie], error) {
	path := fmt.Sprintf("/api/v3/wanted/cutoff?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[Movie]](ctx, c.base, path)
}

// ---------- Blocklist ----------.
