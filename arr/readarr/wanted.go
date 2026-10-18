package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetWantedMissing returns books with missing files (paginated).
func (c *Client) GetWantedMissing(ctx context.Context, page, pageSize int) (*arr.PagingResource[Book], error) {
	path := fmt.Sprintf("/api/v1/wanted/missing?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[Book]](ctx, c.base, path)
}

// GetWantedCutoff returns books not meeting quality cutoff (paginated).
func (c *Client) GetWantedCutoff(ctx context.Context, page, pageSize int) (*arr.PagingResource[Book], error) {
	path := fmt.Sprintf("/api/v1/wanted/cutoff?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[Book]](ctx, c.base, path)
}

// GetWantedMissingByID returns a single wanted missing record by its ID.
func (c *Client) GetWantedMissingByID(ctx context.Context, id int) (*Book, error) {
	return arr.GetOne[Book](ctx, c.base, fmt.Sprintf("/api/v1/wanted/missing/%d", id))
}

// GetWantedCutoffByID returns a single wanted cutoff record by its ID.
func (c *Client) GetWantedCutoffByID(ctx context.Context, id int) (*Book, error) {
	return arr.GetOne[Book](ctx, c.base, fmt.Sprintf("/api/v1/wanted/cutoff/%d", id))
}

// ---------- Book Overview ----------.
