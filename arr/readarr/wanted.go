package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetWantedMissing returns books with missing files (paginated).
func (c *Client) GetWantedMissing(ctx context.Context, page, pageSize int) (*arr.PagingResource[Book], error) {
	var out arr.PagingResource[Book]
	path := fmt.Sprintf("/api/v1/wanted/missing?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedCutoff returns books not meeting quality cutoff (paginated).
func (c *Client) GetWantedCutoff(ctx context.Context, page, pageSize int) (*arr.PagingResource[Book], error) {
	var out arr.PagingResource[Book]
	path := fmt.Sprintf("/api/v1/wanted/cutoff?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedMissingByID returns a single wanted missing record by its ID.
func (c *Client) GetWantedMissingByID(ctx context.Context, id int) (*Book, error) {
	var out Book
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/wanted/missing/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedCutoffByID returns a single wanted cutoff record by its ID.
func (c *Client) GetWantedCutoffByID(ctx context.Context, id int) (*Book, error) {
	var out Book
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/wanted/cutoff/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Book Overview ----------.
