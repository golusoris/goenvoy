package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetWantedMissing returns albums with missing tracks (paginated).
func (c *Client) GetWantedMissing(ctx context.Context, page, pageSize int) (*arr.PagingResource[Album], error) {
	var out arr.PagingResource[Album]
	path := fmt.Sprintf("/api/v1/wanted/missing?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedCutoff returns albums not meeting quality cutoff (paginated).
func (c *Client) GetWantedCutoff(ctx context.Context, page, pageSize int) (*arr.PagingResource[Album], error) {
	var out arr.PagingResource[Album]
	path := fmt.Sprintf("/api/v1/wanted/cutoff?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedMissingByID returns a single wanted missing record by its ID.
func (c *Client) GetWantedMissingByID(ctx context.Context, id int) (*Album, error) {
	var out Album
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/wanted/missing/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedCutoffByID returns a single wanted cutoff record by its ID.
func (c *Client) GetWantedCutoffByID(ctx context.Context, id int) (*Album, error) {
	var out Album
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/wanted/cutoff/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Config By ID ----------.
