package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetIndexers returns all indexer configurations.
func (c *Client) GetIndexers(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v1/indexer")
}

// GetIndexer returns a single indexer by ID.
func (c *Client) GetIndexer(ctx context.Context, id int) (*arr.ProviderResource, error) {
	return arr.GetOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v1/indexer/%d", id))
}

// CreateIndexer creates a new indexer.
func (c *Client) CreateIndexer(ctx context.Context, idx *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PostOne[arr.ProviderResource](ctx, c.base, "/api/v1/indexer", idx)
}

// UpdateIndexer updates an existing indexer.
func (c *Client) UpdateIndexer(ctx context.Context, idx *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PutOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v1/indexer/%d", idx.ID), idx)
}

// DeleteIndexer removes an indexer.
func (c *Client) DeleteIndexer(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/indexer/%d", id), nil, nil)
}

// GetIndexerSchema returns available indexer implementations.
func (c *Client) GetIndexerSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v1/indexer/schema")
}

// TestIndexer sends a test for an indexer configuration.
func (c *Client) TestIndexer(ctx context.Context, idx *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v1/indexer/test", idx, nil)
}

// TestAllIndexers tests all configured indexers.
func (c *Client) TestAllIndexers(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v1/indexer/testall", nil, nil)
}

// UpdateIndexersBulk performs a bulk update of indexers.
func (c *Client) UpdateIndexersBulk(ctx context.Context, bulk *arr.ProviderBulkResource) ([]arr.ProviderResource, error) {
	return arr.PutList[arr.ProviderResource](ctx, c.base, "/api/v1/indexer/bulk", bulk)
}

// DeleteIndexersBulk bulk-deletes indexers by IDs.
func (c *Client) DeleteIndexersBulk(ctx context.Context, ids []int) error {
	body := struct {
		IDs []int `json:"ids"`
	}{IDs: ids}
	return c.base.Delete(ctx, "/api/v1/indexer/bulk", body, nil)
}

// ---------- Import Lists ----------.

// GetIndexerFlags returns the list of indexer flags.
func (c *Client) GetIndexerFlags(ctx context.Context) ([]arr.IndexerFlagResource, error) {
	return arr.GetList[arr.IndexerFlagResource](ctx, c.base, "/api/v1/indexerflag")
}

// ---------- Metadata Profile Schema ----------.
