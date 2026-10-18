package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetWantedMissing returns missing movies (paged).
func (c *ClientV3) GetWantedMissing(ctx context.Context, page, pageSize int) (*arr.PagingResource[Movie], error) {
	path := fmt.Sprintf("/api/v3/wanted/missing?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[Movie]](ctx, c.base, path)
}

// GetWantedMissingByID returns a single missing movie by ID.
func (c *ClientV3) GetWantedMissingByID(ctx context.Context, id int) (*Movie, error) {
	return arr.GetOne[Movie](ctx, c.base, fmt.Sprintf("/api/v3/wanted/missing/%d", id))
}

// GetWantedCutoff returns cutoff-unmet movies (paged).
func (c *ClientV3) GetWantedCutoff(ctx context.Context, page, pageSize int) (*arr.PagingResource[Movie], error) {
	path := fmt.Sprintf("/api/v3/wanted/cutoff?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[Movie]](ctx, c.base, path)
}

// GetWantedCutoffByID returns a single cutoff-unmet movie by ID.
func (c *ClientV3) GetWantedCutoffByID(ctx context.Context, id int) (*Movie, error) {
	return arr.GetOne[Movie](ctx, c.base, fmt.Sprintf("/api/v3/wanted/cutoff/%d", id))
}
