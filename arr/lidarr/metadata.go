package lidarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetMetadataProfiles returns all configured metadata profiles.
func (c *Client) GetMetadataProfiles(ctx context.Context) ([]MetadataProfile, error) {
	var out []MetadataProfile
	if err := c.base.Get(ctx, "/api/v1/metadataprofile", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMetadataConsumers returns all metadata consumer configurations.
func (c *Client) GetMetadataConsumers(ctx context.Context) ([]arr.ProviderResource, error) {
	var out []arr.ProviderResource
	if err := c.base.Get(ctx, "/api/v1/metadata", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMetadataConsumer returns a single metadata consumer by ID.
func (c *Client) GetMetadataConsumer(ctx context.Context, id int) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/metadata/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateMetadataConsumer creates a new metadata consumer.
func (c *Client) CreateMetadataConsumer(ctx context.Context, m *arr.ProviderResource) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Post(ctx, "/api/v1/metadata", m, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMetadataConsumer updates an existing metadata consumer.
func (c *Client) UpdateMetadataConsumer(ctx context.Context, m *arr.ProviderResource) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/metadata/%d", m.ID), m, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMetadataConsumer removes a metadata consumer.
func (c *Client) DeleteMetadataConsumer(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/metadata/%d", id), nil, nil)
}

// GetMetadataSchema returns available metadata consumer implementations.
func (c *Client) GetMetadataSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	var out []arr.ProviderResource
	if err := c.base.Get(ctx, "/api/v1/metadata/schema", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TestMetadataConsumer sends a test for a metadata consumer configuration.
func (c *Client) TestMetadataConsumer(ctx context.Context, m *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v1/metadata/test", m, nil)
}

// TestAllMetadataConsumers tests all configured metadata consumers.
func (c *Client) TestAllMetadataConsumers(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v1/metadata/testall", nil, nil)
}

// ---------- Config: Download Client ----------.

// GetMetadataProfile returns a single metadata profile by ID.
func (c *Client) GetMetadataProfile(ctx context.Context, id int) (*MetadataProfile, error) {
	var out MetadataProfile
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/metadataprofile/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateMetadataProfile creates a new metadata profile.
func (c *Client) CreateMetadataProfile(ctx context.Context, p *MetadataProfile) (*MetadataProfile, error) {
	var out MetadataProfile
	if err := c.base.Post(ctx, "/api/v1/metadataprofile", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMetadataProfile updates an existing metadata profile.
func (c *Client) UpdateMetadataProfile(ctx context.Context, p *MetadataProfile) (*MetadataProfile, error) {
	var out MetadataProfile
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/metadataprofile/%d", p.ID), p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMetadataProfile deletes a metadata profile by ID.
func (c *Client) DeleteMetadataProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/metadataprofile/%d", id), nil, nil)
}

// ---------- Tags ----------.

// GetMetadataProviderConfig returns the metadata provider configuration.
func (c *Client) GetMetadataProviderConfig(ctx context.Context) (*MetadataProviderConfigResource, error) {
	var out MetadataProviderConfigResource
	if err := c.base.Get(ctx, "/api/v1/config/metadataprovider", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMetadataProviderConfigByID returns the metadata provider config by its ID.
func (c *Client) GetMetadataProviderConfigByID(ctx context.Context, id int) (*MetadataProviderConfigResource, error) {
	var out MetadataProviderConfigResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/config/metadataprovider/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMetadataProviderConfig updates the metadata provider configuration.
func (c *Client) UpdateMetadataProviderConfig(ctx context.Context, config *MetadataProviderConfigResource) (*MetadataProviderConfigResource, error) {
	var out MetadataProviderConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/config/metadataprovider/%d", config.ID), config, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Indexer Flags ----------.

// GetMetadataProfileSchema returns the metadata profile schema.
func (c *Client) GetMetadataProfileSchema(ctx context.Context) (*MetadataProfile, error) {
	var out MetadataProfile
	if err := c.base.Get(ctx, "/api/v1/metadataprofile/schema", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Naming Examples ----------.

// MetadataConsumerAction triggers a named action on a metadata consumer provider.
func (c *Client) MetadataConsumerAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v1/metadata/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}
