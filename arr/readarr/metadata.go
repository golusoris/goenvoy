package readarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetMetadataProfiles returns all configured metadata profiles.
func (c *Client) GetMetadataProfiles(ctx context.Context) ([]MetadataProfile, error) {
	return arr.GetList[MetadataProfile](ctx, c.base, "/api/v1/metadataprofile")
}

// GetMetadataConsumers returns all configured metadata consumers.
func (c *Client) GetMetadataConsumers(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v1/metadata")
}

// GetMetadataConsumer returns a single metadata consumer by its ID.
func (c *Client) GetMetadataConsumer(ctx context.Context, id int) (*arr.ProviderResource, error) {
	return arr.GetOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v1/metadata/%d", id))
}

// CreateMetadataConsumer creates a new metadata consumer.
func (c *Client) CreateMetadataConsumer(ctx context.Context, m *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PostOne[arr.ProviderResource](ctx, c.base, "/api/v1/metadata", m)
}

// UpdateMetadataConsumer updates an existing metadata consumer.
func (c *Client) UpdateMetadataConsumer(ctx context.Context, m *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PutOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v1/metadata/%d", m.ID), m)
}

// DeleteMetadataConsumer deletes a metadata consumer by ID.
func (c *Client) DeleteMetadataConsumer(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/metadata/%d", id), nil, nil)
}

// GetMetadataConsumerSchema returns the schema for all metadata consumer types.
func (c *Client) GetMetadataConsumerSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v1/metadata/schema")
}

// TestMetadataConsumer tests a metadata consumer configuration.
func (c *Client) TestMetadataConsumer(ctx context.Context, m *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v1/metadata/test", m, nil)
}

// TestAllMetadataConsumers tests all configured metadata consumers.
func (c *Client) TestAllMetadataConsumers(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v1/metadata/testall", nil, nil)
}

// MetadataConsumerAction triggers a named action on a metadata consumer provider.
func (c *Client) MetadataConsumerAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v1/metadata/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}

// ---------- Config Endpoints ----------.

// GetMetadataProviderConfig returns the metadata provider configuration.
func (c *Client) GetMetadataProviderConfig(ctx context.Context) (*MetadataProviderConfigResource, error) {
	return arr.GetOne[MetadataProviderConfigResource](ctx, c.base, "/api/v1/config/metadataprovider")
}

// GetMetadataProviderConfigByID returns the metadata provider config by its ID.
func (c *Client) GetMetadataProviderConfigByID(ctx context.Context, id int) (*MetadataProviderConfigResource, error) {
	return arr.GetOne[MetadataProviderConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/metadataprovider/%d", id))
}

// UpdateMetadataProviderConfig updates the metadata provider configuration.
func (c *Client) UpdateMetadataProviderConfig(ctx context.Context, config *MetadataProviderConfigResource) (*MetadataProviderConfigResource, error) {
	return arr.PutOne[MetadataProviderConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/metadataprovider/%d", config.ID), config)
}

// ---------- Quality Profiles ----------.

// GetMetadataProfile returns a single metadata profile by its ID.
func (c *Client) GetMetadataProfile(ctx context.Context, id int) (*MetadataProfile, error) {
	return arr.GetOne[MetadataProfile](ctx, c.base, fmt.Sprintf("/api/v1/metadataprofile/%d", id))
}

// CreateMetadataProfile creates a new metadata profile.
func (c *Client) CreateMetadataProfile(ctx context.Context, profile *MetadataProfile) (*MetadataProfile, error) {
	return arr.PostOne[MetadataProfile](ctx, c.base, "/api/v1/metadataprofile", profile)
}

// UpdateMetadataProfile updates an existing metadata profile.
func (c *Client) UpdateMetadataProfile(ctx context.Context, profile *MetadataProfile) (*MetadataProfile, error) {
	return arr.PutOne[MetadataProfile](ctx, c.base, fmt.Sprintf("/api/v1/metadataprofile/%d", profile.ID), profile)
}

// DeleteMetadataProfile deletes a metadata profile by ID.
func (c *Client) DeleteMetadataProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/metadataprofile/%d", id), nil, nil)
}

// GetMetadataProfileSchema returns the metadata profile schema.
func (c *Client) GetMetadataProfileSchema(ctx context.Context) (*MetadataProfile, error) {
	return arr.GetOne[MetadataProfile](ctx, c.base, "/api/v1/metadataprofile/schema")
}

// ---------- Tags extended ----------.
