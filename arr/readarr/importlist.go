package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetImportListExclusions returns all import list exclusions.
func (c *Client) GetImportListExclusions(ctx context.Context) ([]ImportListExclusion, error) {
	return arr.GetList[ImportListExclusion](ctx, c.base, "/api/v1/importlistexclusion")
}

// GetImportLists returns all configured import lists.
func (c *Client) GetImportLists(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v1/importlist")
}

// GetImportList returns a single import list by its ID.
func (c *Client) GetImportList(ctx context.Context, id int) (*arr.ProviderResource, error) {
	return arr.GetOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v1/importlist/%d", id))
}

// CreateImportList creates a new import list.
func (c *Client) CreateImportList(ctx context.Context, il *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PostOne[arr.ProviderResource](ctx, c.base, "/api/v1/importlist", il)
}

// UpdateImportList updates an existing import list.
func (c *Client) UpdateImportList(ctx context.Context, il *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PutOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v1/importlist/%d", il.ID), il)
}

// DeleteImportList deletes an import list by ID.
func (c *Client) DeleteImportList(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/importlist/%d", id), nil, nil)
}

// GetImportListSchema returns the schema for all import list types.
func (c *Client) GetImportListSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v1/importlist/schema")
}

// TestImportList tests an import list configuration.
func (c *Client) TestImportList(ctx context.Context, il *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v1/importlist/test", il, nil)
}

// TestAllImportLists tests all configured import lists.
func (c *Client) TestAllImportLists(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v1/importlist/testall", nil, nil)
}

// BulkUpdateImportLists updates multiple import lists at once.
func (c *Client) BulkUpdateImportLists(ctx context.Context, bulk *arr.ProviderBulkResource) ([]arr.ProviderResource, error) {
	return arr.PutList[arr.ProviderResource](ctx, c.base, "/api/v1/importlist/bulk", bulk)
}

// BulkDeleteImportLists deletes multiple import lists at once.
func (c *Client) BulkDeleteImportLists(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v1/importlist/bulk", struct {
		IDs []int `json:"ids"`
	}{IDs: ids}, nil)
}

// GetImportListExclusion returns a single import list exclusion by its ID.
func (c *Client) GetImportListExclusion(ctx context.Context, id int) (*ImportListExclusion, error) {
	return arr.GetOne[ImportListExclusion](ctx, c.base, fmt.Sprintf("/api/v1/importlistexclusion/%d", id))
}

// CreateImportListExclusion creates a new import list exclusion.
func (c *Client) CreateImportListExclusion(ctx context.Context, exclusion *ImportListExclusion) (*ImportListExclusion, error) {
	return arr.PostOne[ImportListExclusion](ctx, c.base, "/api/v1/importlistexclusion", exclusion)
}

// UpdateImportListExclusion updates an existing import list exclusion.
func (c *Client) UpdateImportListExclusion(ctx context.Context, exclusion *ImportListExclusion) (*ImportListExclusion, error) {
	return arr.PutOne[ImportListExclusion](ctx, c.base, fmt.Sprintf("/api/v1/importlistexclusion/%d", exclusion.ID), exclusion)
}

// DeleteImportListExclusion deletes an import list exclusion by ID.
func (c *Client) DeleteImportListExclusion(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/importlistexclusion/%d", id), nil, nil)
}

// ---------- Blocklist ----------.
