package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetDownloadClientConfig returns the download client configuration.
func (c *api) GetDownloadClientConfig(ctx context.Context) (*arr.DownloadClientConfigResource, error) {
	return arr.GetOne[arr.DownloadClientConfigResource](ctx, c.base, "/api/v3/config/downloadclient")
}

// UpdateDownloadClientConfig updates the download client configuration.
func (c *api) UpdateDownloadClientConfig(ctx context.Context, cfg *arr.DownloadClientConfigResource) (*arr.DownloadClientConfigResource, error) {
	return arr.PutOne[arr.DownloadClientConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/downloadclient/%d", cfg.ID), cfg)
}

// GetHostConfig returns the host configuration.
func (c *api) GetHostConfig(ctx context.Context) (*arr.HostConfigResource, error) {
	return arr.GetOne[arr.HostConfigResource](ctx, c.base, "/api/v3/config/host")
}

// UpdateHostConfig updates the host configuration.
func (c *api) UpdateHostConfig(ctx context.Context, cfg *arr.HostConfigResource) (*arr.HostConfigResource, error) {
	return arr.PutOne[arr.HostConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/host/%d", cfg.ID), cfg)
}

// GetIndexerConfig returns the indexer configuration.
func (c *api) GetIndexerConfig(ctx context.Context) (*arr.IndexerConfigResource, error) {
	return arr.GetOne[arr.IndexerConfigResource](ctx, c.base, "/api/v3/config/indexer")
}

// UpdateIndexerConfig updates the indexer configuration.
func (c *api) UpdateIndexerConfig(ctx context.Context, cfg *arr.IndexerConfigResource) (*arr.IndexerConfigResource, error) {
	return arr.PutOne[arr.IndexerConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/indexer/%d", cfg.ID), cfg)
}

// GetMediaManagementConfig returns the media management configuration.
func (c *api) GetMediaManagementConfig(ctx context.Context) (*arr.MediaManagementConfigResource, error) {
	return arr.GetOne[arr.MediaManagementConfigResource](ctx, c.base, "/api/v3/config/mediamanagement")
}

// UpdateMediaManagementConfig updates the media management configuration.
func (c *api) UpdateMediaManagementConfig(ctx context.Context, cfg *arr.MediaManagementConfigResource) (*arr.MediaManagementConfigResource, error) {
	return arr.PutOne[arr.MediaManagementConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/mediamanagement/%d", cfg.ID), cfg)
}

// GetNamingConfig returns the naming configuration.
func (c *api) GetNamingConfig(ctx context.Context) (*arr.NamingConfigResource, error) {
	return arr.GetOne[arr.NamingConfigResource](ctx, c.base, "/api/v3/config/naming")
}

// UpdateNamingConfig updates the naming configuration.
func (c *api) UpdateNamingConfig(ctx context.Context, cfg *arr.NamingConfigResource) (*arr.NamingConfigResource, error) {
	return arr.PutOne[arr.NamingConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/naming/%d", cfg.ID), cfg)
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
	return arr.GetOne[arr.UIConfigResource](ctx, c.base, "/api/v3/config/ui")
}

// UpdateUIConfig updates the UI configuration.
func (c *api) UpdateUIConfig(ctx context.Context, cfg *arr.UIConfigResource) (*arr.UIConfigResource, error) {
	return arr.PutOne[arr.UIConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/ui/%d", cfg.ID), cfg)
}
