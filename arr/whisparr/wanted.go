package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetWantedMissing returns missing episodes (paged).
func (c *Client) GetWantedMissing(ctx context.Context, page, pageSize int) (*arr.PagingResource[Episode], error) {
	var out arr.PagingResource[Episode]
	path := fmt.Sprintf("/api/v3/wanted/missing?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedMissingByID returns a single missing episode by ID.
func (c *Client) GetWantedMissingByID(ctx context.Context, id int) (*Episode, error) {
	var out Episode
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/wanted/missing/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedCutoff returns cutoff unmet episodes (paged).
func (c *Client) GetWantedCutoff(ctx context.Context, page, pageSize int) (*arr.PagingResource[Episode], error) {
	var out arr.PagingResource[Episode]
	path := fmt.Sprintf("/api/v3/wanted/cutoff?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedCutoffByID returns a single cutoff-unmet episode by ID.
func (c *Client) GetWantedCutoffByID(ctx context.Context, id int) (*Episode, error) {
	var out Episode
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/wanted/cutoff/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Delay Profiles ----------.
