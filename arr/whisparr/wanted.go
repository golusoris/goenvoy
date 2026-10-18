package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetWantedMissing returns missing episodes (paged).
func (c *Client) GetWantedMissing(ctx context.Context, page, pageSize int) (*arr.PagingResource[Episode], error) {
	path := fmt.Sprintf("/api/v3/wanted/missing?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[Episode]](ctx, c.base, path)
}

// GetWantedMissingByID returns a single missing episode by ID.
func (c *Client) GetWantedMissingByID(ctx context.Context, id int) (*Episode, error) {
	return arr.GetOne[Episode](ctx, c.base, fmt.Sprintf("/api/v3/wanted/missing/%d", id))
}

// GetWantedCutoff returns cutoff unmet episodes (paged).
func (c *Client) GetWantedCutoff(ctx context.Context, page, pageSize int) (*arr.PagingResource[Episode], error) {
	path := fmt.Sprintf("/api/v3/wanted/cutoff?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[Episode]](ctx, c.base, path)
}

// GetWantedCutoffByID returns a single cutoff-unmet episode by ID.
func (c *Client) GetWantedCutoffByID(ctx context.Context, id int) (*Episode, error) {
	return arr.GetOne[Episode](ctx, c.base, fmt.Sprintf("/api/v3/wanted/cutoff/%d", id))
}
