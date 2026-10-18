package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetMetadataConsumers returns all metadata consumer configurations.
func (c *Client) GetMetadataConsumers(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/metadata")
}

// GetMetadataConsumer returns a single metadata consumer by ID.
func (c *Client) GetMetadataConsumer(ctx context.Context, id int) (*arr.ProviderResource, error) {
	return arr.GetOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/metadata/%d", id))
}

// CreateMetadataConsumer creates a new metadata consumer.
func (c *Client) CreateMetadataConsumer(ctx context.Context, m *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PostOne[arr.ProviderResource](ctx, c.base, "/api/v3/metadata", m)
}

// UpdateMetadataConsumer updates an existing metadata consumer.
func (c *Client) UpdateMetadataConsumer(ctx context.Context, m *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PutOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/metadata/%d", m.ID), m)
}

// DeleteMetadataConsumer removes a metadata consumer.
func (c *Client) DeleteMetadataConsumer(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/metadata/%d", id), nil, nil)
}

// GetMetadataSchema returns available metadata consumer implementations.
func (c *Client) GetMetadataSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/metadata/schema")
}

// TestMetadataConsumer sends a test for a metadata consumer configuration.
func (c *Client) TestMetadataConsumer(ctx context.Context, m *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/metadata/test", m, nil)
}

// ---------- Auto Tagging ----------.

// TestAllMetadataConsumers tests all configured metadata consumers.
func (c *Client) TestAllMetadataConsumers(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/metadata/testall", nil, nil)
}

// ---------- Language ----------.

// GetMetadataConfig returns the metadata global configuration.
func (c *Client) GetMetadataConfig(ctx context.Context) (*MetadataConfigResource, error) {
	return arr.GetOne[MetadataConfigResource](ctx, c.base, "/api/v3/config/metadata")
}

// GetMetadataConfigByID returns the metadata config by its ID.
func (c *Client) GetMetadataConfigByID(ctx context.Context, id int) (*MetadataConfigResource, error) {
	return arr.GetOne[MetadataConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/metadata/%d", id))
}

// UpdateMetadataConfig updates the metadata global configuration.
func (c *Client) UpdateMetadataConfig(ctx context.Context, cfg *MetadataConfigResource) (*MetadataConfigResource, error) {
	return arr.PutOne[MetadataConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/metadata/%d", cfg.ID), cfg)
}

// ---------- Movie File Bulk Update ----------.
