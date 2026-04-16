package whisparr

import (
	"context"
	"fmt"
)

// GetImportExclusions returns all import exclusions.
func (c *ClientV3) GetImportExclusions(ctx context.Context) ([]ImportExclusion, error) {
	var out []ImportExclusion
	if err := c.base.Get(ctx, "/api/v3/exclusions", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetImportExclusion returns a single import exclusion by ID.
func (c *ClientV3) GetImportExclusion(ctx context.Context, id int) (*ImportExclusion, error) {
	var out ImportExclusion
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/exclusions/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateImportExclusion creates a new import exclusion.
func (c *ClientV3) CreateImportExclusion(ctx context.Context, ex *ImportExclusion) (*ImportExclusion, error) {
	var out ImportExclusion
	if err := c.base.Post(ctx, "/api/v3/exclusions", ex, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateImportExclusion updates an existing import exclusion.
func (c *ClientV3) UpdateImportExclusion(ctx context.Context, ex *ImportExclusion) (*ImportExclusion, error) {
	var out ImportExclusion
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/exclusions/%d", ex.ID), ex, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteImportExclusion deletes an import exclusion by ID.
func (c *ClientV3) DeleteImportExclusion(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/exclusions/%d", id), nil, nil)
}

// BulkCreateImportExclusions creates multiple import exclusions.
func (c *ClientV3) BulkCreateImportExclusions(ctx context.Context, exs []ImportExclusion) ([]ImportExclusion, error) {
	var out []ImportExclusion
	if err := c.base.Post(ctx, "/api/v3/exclusions/bulk", exs, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BulkDeleteImportExclusions deletes multiple import exclusions.
func (c *ClientV3) BulkDeleteImportExclusions(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v3/exclusions/bulk", struct {
		IDs []int `json:"ids"`
	}{IDs: ids}, nil)
}
