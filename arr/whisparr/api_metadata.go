package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetMetadata returns all metadata consumers.
func (c *api) GetMetadata(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/metadata")
}

// GetMetadataByID returns a metadata consumer by ID.
func (c *api) GetMetadataByID(ctx context.Context, id int) (*arr.ProviderResource, error) {
	return arr.GetOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/metadata/%d", id))
}

// CreateMetadata creates a new metadata consumer.
func (c *api) CreateMetadata(ctx context.Context, m *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PostOne[arr.ProviderResource](ctx, c.base, "/api/v3/metadata", m)
}

// UpdateMetadata updates an existing metadata consumer.
func (c *api) UpdateMetadata(ctx context.Context, m *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PutOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/metadata/%d", m.ID), m)
}

// DeleteMetadata deletes a metadata consumer by ID.
func (c *api) DeleteMetadata(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/metadata/%d", id), nil, nil)
}

// GetMetadataSchema returns the schema for all metadata consumer types.
func (c *api) GetMetadataSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/metadata/schema")
}

// TestMetadata tests a metadata consumer configuration.
func (c *api) TestMetadata(ctx context.Context, m *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/metadata/test", m, nil)
}

// TestAllMetadata tests all configured metadata consumers.
func (c *api) TestAllMetadata(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/metadata/testall", nil, nil)
}
