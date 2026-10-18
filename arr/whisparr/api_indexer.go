package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetIndexers returns all configured indexers.
func (c *api) GetIndexers(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/indexer")
}

// GetIndexer returns a single indexer by ID.
func (c *api) GetIndexer(ctx context.Context, id int) (*arr.ProviderResource, error) {
	return arr.GetOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/indexer/%d", id))
}

// CreateIndexer creates a new indexer.
func (c *api) CreateIndexer(ctx context.Context, idx *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PostOne[arr.ProviderResource](ctx, c.base, "/api/v3/indexer", idx)
}

// UpdateIndexer updates an existing indexer.
func (c *api) UpdateIndexer(ctx context.Context, idx *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PutOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/indexer/%d", idx.ID), idx)
}

// DeleteIndexer deletes an indexer by ID.
func (c *api) DeleteIndexer(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/indexer/%d", id), nil, nil)
}

// GetIndexerSchema returns the schema for all indexer types.
func (c *api) GetIndexerSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/indexer/schema")
}

// TestIndexer tests an indexer configuration.
func (c *api) TestIndexer(ctx context.Context, idx *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/indexer/test", idx, nil)
}

// TestAllIndexers tests all configured indexers.
func (c *api) TestAllIndexers(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/indexer/testall", nil, nil)
}

// BulkUpdateIndexers updates multiple indexers.
func (c *api) BulkUpdateIndexers(ctx context.Context, bulk *arr.ProviderBulkResource) ([]arr.ProviderResource, error) {
	return arr.PutList[arr.ProviderResource](ctx, c.base, "/api/v3/indexer/bulk", bulk)
}

// BulkDeleteIndexers deletes multiple indexers.
func (c *api) BulkDeleteIndexers(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v3/indexer/bulk", struct {
		IDs []int `json:"ids"`
	}{IDs: ids}, nil)
}
