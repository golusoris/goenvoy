package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetDownloadClientConfig returns the download client global configuration.
func (c *Client) GetDownloadClientConfig(ctx context.Context) (*arr.DownloadClientConfigResource, error) {
	return arr.GetOne[arr.DownloadClientConfigResource](ctx, c.base, "/api/v3/config/downloadclient")
}

// UpdateDownloadClientConfig updates the download client global configuration.
func (c *Client) UpdateDownloadClientConfig(ctx context.Context, cfg *arr.DownloadClientConfigResource) (*arr.DownloadClientConfigResource, error) {
	return arr.PutOne[arr.DownloadClientConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/downloadclient/%d", cfg.ID), cfg)
}

// ---------- Indexers ----------.

// GetIndexerConfig returns the indexer global configuration.
func (c *Client) GetIndexerConfig(ctx context.Context) (*arr.IndexerConfigResource, error) {
	return arr.GetOne[arr.IndexerConfigResource](ctx, c.base, "/api/v3/config/indexer")
}

// UpdateIndexerConfig updates the indexer global configuration.
func (c *Client) UpdateIndexerConfig(ctx context.Context, cfg *arr.IndexerConfigResource) (*arr.IndexerConfigResource, error) {
	return arr.PutOne[arr.IndexerConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/indexer/%d", cfg.ID), cfg)
}

// GetNamingConfig returns the file naming configuration.
func (c *Client) GetNamingConfig(ctx context.Context) (*arr.NamingConfigResource, error) {
	return arr.GetOne[arr.NamingConfigResource](ctx, c.base, "/api/v3/config/naming")
}

// UpdateNamingConfig updates the file naming configuration.
func (c *Client) UpdateNamingConfig(ctx context.Context, cfg *arr.NamingConfigResource) (*arr.NamingConfigResource, error) {
	return arr.PutOne[arr.NamingConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/naming/%d", cfg.ID), cfg)
}

// GetHostConfig returns the host configuration.
func (c *Client) GetHostConfig(ctx context.Context) (*arr.HostConfigResource, error) {
	return arr.GetOne[arr.HostConfigResource](ctx, c.base, "/api/v3/config/host")
}

// UpdateHostConfig updates the host configuration.
func (c *Client) UpdateHostConfig(ctx context.Context, cfg *arr.HostConfigResource) (*arr.HostConfigResource, error) {
	return arr.PutOne[arr.HostConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/host/%d", cfg.ID), cfg)
}

// GetUIConfig returns the UI configuration.
func (c *Client) GetUIConfig(ctx context.Context) (*arr.UIConfigResource, error) {
	return arr.GetOne[arr.UIConfigResource](ctx, c.base, "/api/v3/config/ui")
}

// UpdateUIConfig updates the UI configuration.
func (c *Client) UpdateUIConfig(ctx context.Context, cfg *arr.UIConfigResource) (*arr.UIConfigResource, error) {
	return arr.PutOne[arr.UIConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/ui/%d", cfg.ID), cfg)
}

// GetMediaManagementConfig returns the media management configuration.
func (c *Client) GetMediaManagementConfig(ctx context.Context) (*arr.MediaManagementConfigResource, error) {
	return arr.GetOne[arr.MediaManagementConfigResource](ctx, c.base, "/api/v3/config/mediamanagement")
}

// UpdateMediaManagementConfig updates the media management configuration.
func (c *Client) UpdateMediaManagementConfig(ctx context.Context, cfg *arr.MediaManagementConfigResource) (*arr.MediaManagementConfigResource, error) {
	return arr.PutOne[arr.MediaManagementConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/mediamanagement/%d", cfg.ID), cfg)
}

// ---------- Queue extras ----------.

// GetImportListConfig returns the global import list configuration.
func (c *Client) GetImportListConfig(ctx context.Context) (*ImportListConfigResource, error) {
	return arr.GetOne[ImportListConfigResource](ctx, c.base, "/api/v3/config/importlist")
}

// UpdateImportListConfig updates the global import list configuration.
func (c *Client) UpdateImportListConfig(ctx context.Context, cfg *ImportListConfigResource) (*ImportListConfigResource, error) {
	var out ImportListConfigResource
	path := fmt.Sprintf("/api/v3/config/importlist/%d", cfg.ID)
	if err := c.base.Put(ctx, path, cfg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Import List Exclusion Bulk ----------.

// GetDownloadClientConfigByID returns the download client config by its ID.
func (c *Client) GetDownloadClientConfigByID(ctx context.Context, id int) (*arr.DownloadClientConfigResource, error) {
	return arr.GetOne[arr.DownloadClientConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/downloadclient/%d", id))
}

// GetHostConfigByID returns the host config by its ID.
func (c *Client) GetHostConfigByID(ctx context.Context, id int) (*arr.HostConfigResource, error) {
	return arr.GetOne[arr.HostConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/host/%d", id))
}

// GetImportListConfigByID returns the import list config by its ID.
func (c *Client) GetImportListConfigByID(ctx context.Context, id int) (*ImportListConfigResource, error) {
	return arr.GetOne[ImportListConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/importlist/%d", id))
}

// GetIndexerConfigByID returns the indexer config by its ID.
func (c *Client) GetIndexerConfigByID(ctx context.Context, id int) (*arr.IndexerConfigResource, error) {
	return arr.GetOne[arr.IndexerConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/indexer/%d", id))
}

// GetMediaManagementConfigByID returns the media management config by its ID.
func (c *Client) GetMediaManagementConfigByID(ctx context.Context, id int) (*arr.MediaManagementConfigResource, error) {
	return arr.GetOne[arr.MediaManagementConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/mediamanagement/%d", id))
}

// GetNamingConfigByID returns the naming config by its ID.
func (c *Client) GetNamingConfigByID(ctx context.Context, id int) (*arr.NamingConfigResource, error) {
	return arr.GetOne[arr.NamingConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/naming/%d", id))
}

// GetUIConfigByID returns the UI config by its ID.
func (c *Client) GetUIConfigByID(ctx context.Context, id int) (*arr.UIConfigResource, error) {
	return arr.GetOne[arr.UIConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/ui/%d", id))
}

// ---------- Extra Files ----------.

// GetNamingExamples returns naming format examples based on the current naming config.
func (c *Client) GetNamingExamples(ctx context.Context) (*arr.NamingConfigResource, error) {
	return arr.GetOne[arr.NamingConfigResource](ctx, c.base, "/api/v3/config/naming/examples")
}

// ---------- Quality Definition Limits ----------.
