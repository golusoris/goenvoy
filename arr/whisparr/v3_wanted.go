package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetWantedMissing returns missing movies (paged).
func (c *ClientV3) GetWantedMissing(ctx context.Context, page, pageSize int) (*arr.PagingResource[Movie], error) {
	var out arr.PagingResource[Movie]
	path := fmt.Sprintf("/api/v3/wanted/missing?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedMissingByID returns a single missing movie by ID.
func (c *ClientV3) GetWantedMissingByID(ctx context.Context, id int) (*Movie, error) {
	var out Movie
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/wanted/missing/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedCutoff returns cutoff-unmet movies (paged).
func (c *ClientV3) GetWantedCutoff(ctx context.Context, page, pageSize int) (*arr.PagingResource[Movie], error) {
	var out arr.PagingResource[Movie]
	path := fmt.Sprintf("/api/v3/wanted/cutoff?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedCutoffByID returns a single cutoff-unmet movie by ID.
func (c *ClientV3) GetWantedCutoffByID(ctx context.Context, id int) (*Movie, error) {
	var out Movie
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/wanted/cutoff/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Ping ----------.
