package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetDownloadClientConfig returns the download client configuration.
func (c *Client) GetDownloadClientConfig(ctx context.Context) (*arr.DownloadClientConfigResource, error) {
	return arr.GetOne[arr.DownloadClientConfigResource](ctx, c.base, "/api/v1/config/downloadclient")
}

// UpdateDownloadClientConfig updates the download client configuration.
func (c *Client) UpdateDownloadClientConfig(ctx context.Context, config *arr.DownloadClientConfigResource) (*arr.DownloadClientConfigResource, error) {
	return arr.PutOne[arr.DownloadClientConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/downloadclient/%d", config.ID), config)
}

// GetDownloadClientConfigByID returns the download client config by its ID.
func (c *Client) GetDownloadClientConfigByID(ctx context.Context, id int) (*arr.DownloadClientConfigResource, error) {
	return arr.GetOne[arr.DownloadClientConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/downloadclient/%d", id))
}

// GetIndexerConfig returns the indexer configuration.
func (c *Client) GetIndexerConfig(ctx context.Context) (*arr.IndexerConfigResource, error) {
	return arr.GetOne[arr.IndexerConfigResource](ctx, c.base, "/api/v1/config/indexer")
}

// UpdateIndexerConfig updates the indexer configuration.
func (c *Client) UpdateIndexerConfig(ctx context.Context, config *arr.IndexerConfigResource) (*arr.IndexerConfigResource, error) {
	return arr.PutOne[arr.IndexerConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/indexer/%d", config.ID), config)
}

// GetIndexerConfigByID returns the indexer config by its ID.
func (c *Client) GetIndexerConfigByID(ctx context.Context, id int) (*arr.IndexerConfigResource, error) {
	return arr.GetOne[arr.IndexerConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/indexer/%d", id))
}

// GetNamingConfig returns the naming configuration.
func (c *Client) GetNamingConfig(ctx context.Context) (*arr.NamingConfigResource, error) {
	return arr.GetOne[arr.NamingConfigResource](ctx, c.base, "/api/v1/config/naming")
}

// UpdateNamingConfig updates the naming configuration.
func (c *Client) UpdateNamingConfig(ctx context.Context, config *arr.NamingConfigResource) (*arr.NamingConfigResource, error) {
	return arr.PutOne[arr.NamingConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/naming/%d", config.ID), config)
}

// GetNamingConfigByID returns the naming config by its ID.
func (c *Client) GetNamingConfigByID(ctx context.Context, id int) (*arr.NamingConfigResource, error) {
	return arr.GetOne[arr.NamingConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/naming/%d", id))
}

// GetNamingExamples returns naming format examples based on the current naming config.
func (c *Client) GetNamingExamples(ctx context.Context) (*arr.NamingConfigResource, error) {
	return arr.GetOne[arr.NamingConfigResource](ctx, c.base, "/api/v1/config/naming/examples")
}

// GetHostConfig returns the host configuration.
func (c *Client) GetHostConfig(ctx context.Context) (*arr.HostConfigResource, error) {
	return arr.GetOne[arr.HostConfigResource](ctx, c.base, "/api/v1/config/host")
}

// UpdateHostConfig updates the host configuration.
func (c *Client) UpdateHostConfig(ctx context.Context, config *arr.HostConfigResource) (*arr.HostConfigResource, error) {
	return arr.PutOne[arr.HostConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/host/%d", config.ID), config)
}

// GetHostConfigByID returns the host config by its ID.
func (c *Client) GetHostConfigByID(ctx context.Context, id int) (*arr.HostConfigResource, error) {
	return arr.GetOne[arr.HostConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/host/%d", id))
}

// GetUIConfig returns the UI configuration.
func (c *Client) GetUIConfig(ctx context.Context) (*arr.UIConfigResource, error) {
	return arr.GetOne[arr.UIConfigResource](ctx, c.base, "/api/v1/config/ui")
}

// UpdateUIConfig updates the UI configuration.
func (c *Client) UpdateUIConfig(ctx context.Context, config *arr.UIConfigResource) (*arr.UIConfigResource, error) {
	return arr.PutOne[arr.UIConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/ui/%d", config.ID), config)
}

// GetUIConfigByID returns the UI config by its ID.
func (c *Client) GetUIConfigByID(ctx context.Context, id int) (*arr.UIConfigResource, error) {
	return arr.GetOne[arr.UIConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/ui/%d", id))
}

// GetMediaManagementConfig returns the media management configuration.
func (c *Client) GetMediaManagementConfig(ctx context.Context) (*arr.MediaManagementConfigResource, error) {
	return arr.GetOne[arr.MediaManagementConfigResource](ctx, c.base, "/api/v1/config/mediamanagement")
}

// UpdateMediaManagementConfig updates the media management configuration.
func (c *Client) UpdateMediaManagementConfig(ctx context.Context, config *arr.MediaManagementConfigResource) (*arr.MediaManagementConfigResource, error) {
	return arr.PutOne[arr.MediaManagementConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/mediamanagement/%d", config.ID), config)
}

// GetMediaManagementConfigByID returns the media management config by its ID.
func (c *Client) GetMediaManagementConfigByID(ctx context.Context, id int) (*arr.MediaManagementConfigResource, error) {
	return arr.GetOne[arr.MediaManagementConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/mediamanagement/%d", id))
}

// ---------- Development Config ----------.

// GetDevelopmentConfig returns the development configuration.
func (c *Client) GetDevelopmentConfig(ctx context.Context) (*DevelopmentConfigResource, error) {
	return arr.GetOne[DevelopmentConfigResource](ctx, c.base, "/api/v1/config/development")
}

// GetDevelopmentConfigByID returns the development config by its ID.
func (c *Client) GetDevelopmentConfigByID(ctx context.Context, id int) (*DevelopmentConfigResource, error) {
	return arr.GetOne[DevelopmentConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/development/%d", id))
}

// UpdateDevelopmentConfig updates the development configuration.
func (c *Client) UpdateDevelopmentConfig(ctx context.Context, config *DevelopmentConfigResource) (*DevelopmentConfigResource, error) {
	return arr.PutOne[DevelopmentConfigResource](ctx, c.base, fmt.Sprintf("/api/v1/config/development/%d", config.ID), config)
}
