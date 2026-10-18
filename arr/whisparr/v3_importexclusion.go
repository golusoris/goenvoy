package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetImportExclusions returns all import exclusions.
func (c *ClientV3) GetImportExclusions(ctx context.Context) ([]ImportExclusion, error) {
	return arr.GetList[ImportExclusion](ctx, c.base, "/api/v3/exclusions")
}

// GetImportExclusion returns a single import exclusion by ID.
func (c *ClientV3) GetImportExclusion(ctx context.Context, id int) (*ImportExclusion, error) {
	return arr.GetOne[ImportExclusion](ctx, c.base, fmt.Sprintf("/api/v3/exclusions/%d", id))
}

// CreateImportExclusion creates a new import exclusion.
func (c *ClientV3) CreateImportExclusion(ctx context.Context, ex *ImportExclusion) (*ImportExclusion, error) {
	return arr.PostOne[ImportExclusion](ctx, c.base, "/api/v3/exclusions", ex)
}

// UpdateImportExclusion updates an existing import exclusion.
func (c *ClientV3) UpdateImportExclusion(ctx context.Context, ex *ImportExclusion) (*ImportExclusion, error) {
	return arr.PutOne[ImportExclusion](ctx, c.base, fmt.Sprintf("/api/v3/exclusions/%d", ex.ID), ex)
}

// DeleteImportExclusion deletes an import exclusion by ID.
func (c *ClientV3) DeleteImportExclusion(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/exclusions/%d", id), nil, nil)
}

// BulkCreateImportExclusions creates multiple import exclusions.
func (c *ClientV3) BulkCreateImportExclusions(ctx context.Context, exs []ImportExclusion) ([]ImportExclusion, error) {
	return arr.PostList[ImportExclusion](ctx, c.base, "/api/v3/exclusions/bulk", exs)
}

// BulkDeleteImportExclusions deletes multiple import exclusions.
func (c *ClientV3) BulkDeleteImportExclusions(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v3/exclusions/bulk", struct {
		IDs []int `json:"ids"`
	}{IDs: ids}, nil)
}
