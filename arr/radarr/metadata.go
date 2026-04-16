package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetMetadataConsumers returns all metadata consumer configurations.
func (c *Client) GetMetadataConsumers(ctx context.Context) ([]arr.ProviderResource, error) {
	var out []arr.ProviderResource
	if err := c.base.Get(ctx, "/api/v3/metadata", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMetadataConsumer returns a single metadata consumer by ID.
func (c *Client) GetMetadataConsumer(ctx context.Context, id int) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/metadata/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateMetadataConsumer creates a new metadata consumer.
func (c *Client) CreateMetadataConsumer(ctx context.Context, m *arr.ProviderResource) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Post(ctx, "/api/v3/metadata", m, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMetadataConsumer updates an existing metadata consumer.
func (c *Client) UpdateMetadataConsumer(ctx context.Context, m *arr.ProviderResource) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/metadata/%d", m.ID), m, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMetadataConsumer removes a metadata consumer.
func (c *Client) DeleteMetadataConsumer(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/metadata/%d", id), nil, nil)
}

// GetMetadataSchema returns available metadata consumer implementations.
func (c *Client) GetMetadataSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	var out []arr.ProviderResource
	if err := c.base.Get(ctx, "/api/v3/metadata/schema", &out); err != nil {
		return nil, err
	}
	return out, nil
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
	var out MetadataConfigResource
	if err := c.base.Get(ctx, "/api/v3/config/metadata", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMetadataConfigByID returns the metadata config by its ID.
func (c *Client) GetMetadataConfigByID(ctx context.Context, id int) (*MetadataConfigResource, error) {
	var out MetadataConfigResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/config/metadata/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMetadataConfig updates the metadata global configuration.
func (c *Client) UpdateMetadataConfig(ctx context.Context, cfg *MetadataConfigResource) (*MetadataConfigResource, error) {
	var out MetadataConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/config/metadata/%d", cfg.ID), cfg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Movie File Bulk Update ----------.
