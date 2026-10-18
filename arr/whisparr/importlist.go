package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetImportListExclusions returns all import list exclusions.
func (c *Client) GetImportListExclusions(ctx context.Context) ([]arr.ImportListExclusionResource, error) {
	return arr.GetList[arr.ImportListExclusionResource](ctx, c.base, "/api/v3/importlistexclusion")
}

// GetImportListExclusion returns a single import list exclusion by ID.
func (c *Client) GetImportListExclusion(ctx context.Context, id int) (*arr.ImportListExclusionResource, error) {
	return arr.GetOne[arr.ImportListExclusionResource](ctx, c.base, fmt.Sprintf("/api/v3/importlistexclusion/%d", id))
}

// CreateImportListExclusion creates a new import list exclusion.
func (c *Client) CreateImportListExclusion(ctx context.Context, ex *arr.ImportListExclusionResource) (*arr.ImportListExclusionResource, error) {
	return arr.PostOne[arr.ImportListExclusionResource](ctx, c.base, "/api/v3/importlistexclusion", ex)
}

// UpdateImportListExclusion updates an existing import list exclusion.
func (c *Client) UpdateImportListExclusion(ctx context.Context, ex *arr.ImportListExclusionResource) (*arr.ImportListExclusionResource, error) {
	return arr.PutOne[arr.ImportListExclusionResource](ctx, c.base, fmt.Sprintf("/api/v3/importlistexclusion/%d", ex.ID), ex)
}

// DeleteImportListExclusion deletes an import list exclusion by ID.
func (c *Client) DeleteImportListExclusion(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/importlistexclusion/%d", id), nil, nil)
}
