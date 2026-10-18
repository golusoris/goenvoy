package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetDownloadClientConfig returns the download client configuration.
func (c *api) GetDownloadClientConfig(ctx context.Context) (*arr.DownloadClientConfigResource, error) {
	var out arr.DownloadClientConfigResource
	if err := c.base.Get(ctx, "/api/v3/config/downloadclient", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateDownloadClientConfig updates the download client configuration.
func (c *api) UpdateDownloadClientConfig(ctx context.Context, cfg *arr.DownloadClientConfigResource) (*arr.DownloadClientConfigResource, error) {
	var out arr.DownloadClientConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/config/downloadclient/%d", cfg.ID), cfg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHostConfig returns the host configuration.
func (c *api) GetHostConfig(ctx context.Context) (*arr.HostConfigResource, error) {
	var out arr.HostConfigResource
	if err := c.base.Get(ctx, "/api/v3/config/host", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateHostConfig updates the host configuration.
func (c *api) UpdateHostConfig(ctx context.Context, cfg *arr.HostConfigResource) (*arr.HostConfigResource, error) {
	var out arr.HostConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/config/host/%d", cfg.ID), cfg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIndexerConfig returns the indexer configuration.
func (c *api) GetIndexerConfig(ctx context.Context) (*arr.IndexerConfigResource, error) {
	var out arr.IndexerConfigResource
	if err := c.base.Get(ctx, "/api/v3/config/indexer", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateIndexerConfig updates the indexer configuration.
func (c *api) UpdateIndexerConfig(ctx context.Context, cfg *arr.IndexerConfigResource) (*arr.IndexerConfigResource, error) {
	var out arr.IndexerConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/config/indexer/%d", cfg.ID), cfg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMediaManagementConfig returns the media management configuration.
func (c *api) GetMediaManagementConfig(ctx context.Context) (*arr.MediaManagementConfigResource, error) {
	var out arr.MediaManagementConfigResource
	if err := c.base.Get(ctx, "/api/v3/config/mediamanagement", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMediaManagementConfig updates the media management configuration.
func (c *api) UpdateMediaManagementConfig(ctx context.Context, cfg *arr.MediaManagementConfigResource) (*arr.MediaManagementConfigResource, error) {
	var out arr.MediaManagementConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/config/mediamanagement/%d", cfg.ID), cfg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetNamingConfig returns the naming configuration.
func (c *api) GetNamingConfig(ctx context.Context) (*arr.NamingConfigResource, error) {
	var out arr.NamingConfigResource
	if err := c.base.Get(ctx, "/api/v3/config/naming", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateNamingConfig updates the naming configuration.
func (c *api) UpdateNamingConfig(ctx context.Context, cfg *arr.NamingConfigResource) (*arr.NamingConfigResource, error) {
	var out arr.NamingConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/config/naming/%d", cfg.ID), cfg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetNamingExamples returns naming examples based on the current config.
func (c *api) GetNamingExamples(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.base.Get(ctx, "/api/v3/config/naming/examples", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetUIConfig returns the UI configuration.
func (c *api) GetUIConfig(ctx context.Context) (*arr.UIConfigResource, error) {
	var out arr.UIConfigResource
	if err := c.base.Get(ctx, "/api/v3/config/ui", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateUIConfig updates the UI configuration.
func (c *api) UpdateUIConfig(ctx context.Context, cfg *arr.UIConfigResource) (*arr.UIConfigResource, error) {
	var out arr.UIConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/config/ui/%d", cfg.ID), cfg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
