package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetImportListExclusions returns all import list exclusions.
func (c *Client) GetImportListExclusions(ctx context.Context) ([]ImportListExclusion, error) {
	var out []ImportListExclusion
	if err := c.base.Get(ctx, "/api/v1/importlistexclusion", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetImportLists returns all import list configurations.
func (c *Client) GetImportLists(ctx context.Context) ([]arr.ProviderResource, error) {
	var out []arr.ProviderResource
	if err := c.base.Get(ctx, "/api/v1/importlist", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetImportList returns a single import list by ID.
func (c *Client) GetImportList(ctx context.Context, id int) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/importlist/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateImportList creates a new import list.
func (c *Client) CreateImportList(ctx context.Context, il *arr.ProviderResource) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Post(ctx, "/api/v1/importlist", il, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateImportList updates an existing import list.
func (c *Client) UpdateImportList(ctx context.Context, il *arr.ProviderResource) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/importlist/%d", il.ID), il, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteImportList removes an import list.
func (c *Client) DeleteImportList(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/importlist/%d", id), nil, nil)
}

// GetImportListSchema returns available import list implementations.
func (c *Client) GetImportListSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	var out []arr.ProviderResource
	if err := c.base.Get(ctx, "/api/v1/importlist/schema", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TestImportList sends a test for an import list configuration.
func (c *Client) TestImportList(ctx context.Context, il *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v1/importlist/test", il, nil)
}

// TestAllImportLists tests all configured import lists.
func (c *Client) TestAllImportLists(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v1/importlist/testall", nil, nil)
}

// UpdateImportListsBulk performs a bulk update of import lists.
func (c *Client) UpdateImportListsBulk(ctx context.Context, bulk *arr.ProviderBulkResource) ([]arr.ProviderResource, error) {
	var out []arr.ProviderResource
	if err := c.base.Put(ctx, "/api/v1/importlist/bulk", bulk, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteImportListsBulk bulk-deletes import lists by IDs.
func (c *Client) DeleteImportListsBulk(ctx context.Context, ids []int) error {
	body := struct {
		IDs []int `json:"ids"`
	}{IDs: ids}
	return c.base.Delete(ctx, "/api/v1/importlist/bulk", body, nil)
}

// ---------- Metadata Consumers ----------.

// GetImportListExclusion returns a single import list exclusion by ID.
func (c *Client) GetImportListExclusion(ctx context.Context, id int) (*ImportListExclusion, error) {
	var out ImportListExclusion
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/importlistexclusion/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateImportListExclusion creates a new import list exclusion.
func (c *Client) CreateImportListExclusion(ctx context.Context, e *ImportListExclusion) (*ImportListExclusion, error) {
	var out ImportListExclusion
	if err := c.base.Post(ctx, "/api/v1/importlistexclusion", e, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateImportListExclusion updates an existing import list exclusion.
func (c *Client) UpdateImportListExclusion(ctx context.Context, e *ImportListExclusion) (*ImportListExclusion, error) {
	var out ImportListExclusion
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/importlistexclusion/%d", e.ID), e, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteImportListExclusion deletes an import list exclusion by ID.
func (c *Client) DeleteImportListExclusion(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/importlistexclusion/%d", id), nil, nil)
}

// GetImportListExclusionsPaged returns a paginated list of import list exclusions.
func (c *Client) GetImportListExclusionsPaged(ctx context.Context, page, pageSize int) (*arr.PagingResource[ImportListExclusion], error) {
	var out arr.PagingResource[ImportListExclusion]
	path := fmt.Sprintf("/api/v1/importlistexclusion/paged?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteImportListExclusionsBulk bulk-deletes import list exclusions by IDs.
func (c *Client) DeleteImportListExclusionsBulk(ctx context.Context, ids []int) error {
	body := struct {
		IDs []int `json:"ids"`
	}{IDs: ids}
	return c.base.Delete(ctx, "/api/v1/importlistexclusion/bulk", body, nil)
}

// ---------- Blocklist ----------.
