package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetImportLists returns all configured import lists.
func (c *api) GetImportLists(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/importlist")
}

// GetImportList returns a single import list by ID.
func (c *api) GetImportList(ctx context.Context, id int) (*arr.ProviderResource, error) {
	return arr.GetOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/importlist/%d", id))
}

// CreateImportList creates a new import list.
func (c *api) CreateImportList(ctx context.Context, il *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PostOne[arr.ProviderResource](ctx, c.base, "/api/v3/importlist", il)
}

// UpdateImportList updates an existing import list.
func (c *api) UpdateImportList(ctx context.Context, il *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PutOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/importlist/%d", il.ID), il)
}

// DeleteImportList deletes an import list by ID.
func (c *api) DeleteImportList(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/importlist/%d", id), nil, nil)
}

// GetImportListSchema returns the schema for all import list types.
func (c *api) GetImportListSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/importlist/schema")
}

// TestImportList tests an import list configuration.
func (c *api) TestImportList(ctx context.Context, il *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/importlist/test", il, nil)
}

// TestAllImportLists tests all configured import lists.
func (c *api) TestAllImportLists(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/importlist/testall", nil, nil)
}

// BulkUpdateImportLists updates multiple import lists.
func (c *api) BulkUpdateImportLists(ctx context.Context, bulk *arr.ProviderBulkResource) ([]arr.ProviderResource, error) {
	return arr.PutList[arr.ProviderResource](ctx, c.base, "/api/v3/importlist/bulk", bulk)
}

// BulkDeleteImportLists deletes multiple import lists.
func (c *api) BulkDeleteImportLists(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v3/importlist/bulk", struct {
		IDs []int `json:"ids"`
	}{IDs: ids}, nil)
}
