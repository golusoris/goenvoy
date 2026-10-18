package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetImportListExclusions returns all import list exclusions.
func (c *Client) GetImportListExclusions(ctx context.Context) ([]arr.ImportListExclusionResource, error) {
	var out []arr.ImportListExclusionResource
	if err := c.base.Get(ctx, "/api/v3/importlistexclusion", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetImportListExclusion returns a single import list exclusion by ID.
func (c *Client) GetImportListExclusion(ctx context.Context, id int) (*arr.ImportListExclusionResource, error) {
	var out arr.ImportListExclusionResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/importlistexclusion/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateImportListExclusion creates a new import list exclusion.
func (c *Client) CreateImportListExclusion(ctx context.Context, ex *arr.ImportListExclusionResource) (*arr.ImportListExclusionResource, error) {
	var out arr.ImportListExclusionResource
	if err := c.base.Post(ctx, "/api/v3/importlistexclusion", ex, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateImportListExclusion updates an existing import list exclusion.
func (c *Client) UpdateImportListExclusion(ctx context.Context, ex *arr.ImportListExclusionResource) (*arr.ImportListExclusionResource, error) {
	var out arr.ImportListExclusionResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/importlistexclusion/%d", ex.ID), ex, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteImportListExclusion deletes an import list exclusion by ID.
func (c *Client) DeleteImportListExclusion(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/importlistexclusion/%d", id), nil, nil)
}
