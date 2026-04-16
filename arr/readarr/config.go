package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetDownloadClientConfig returns the download client configuration.
func (c *Client) GetDownloadClientConfig(ctx context.Context) (*arr.DownloadClientConfigResource, error) {
	var out arr.DownloadClientConfigResource
	if err := c.base.Get(ctx, "/api/v1/config/downloadclient", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateDownloadClientConfig updates the download client configuration.
func (c *Client) UpdateDownloadClientConfig(ctx context.Context, config *arr.DownloadClientConfigResource) (*arr.DownloadClientConfigResource, error) {
	var out arr.DownloadClientConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/config/downloadclient/%d", config.ID), config, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDownloadClientConfigByID returns the download client config by its ID.
func (c *Client) GetDownloadClientConfigByID(ctx context.Context, id int) (*arr.DownloadClientConfigResource, error) {
	var out arr.DownloadClientConfigResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/config/downloadclient/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIndexerConfig returns the indexer configuration.
func (c *Client) GetIndexerConfig(ctx context.Context) (*arr.IndexerConfigResource, error) {
	var out arr.IndexerConfigResource
	if err := c.base.Get(ctx, "/api/v1/config/indexer", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateIndexerConfig updates the indexer configuration.
func (c *Client) UpdateIndexerConfig(ctx context.Context, config *arr.IndexerConfigResource) (*arr.IndexerConfigResource, error) {
	var out arr.IndexerConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/config/indexer/%d", config.ID), config, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIndexerConfigByID returns the indexer config by its ID.
func (c *Client) GetIndexerConfigByID(ctx context.Context, id int) (*arr.IndexerConfigResource, error) {
	var out arr.IndexerConfigResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/config/indexer/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetNamingConfig returns the naming configuration.
func (c *Client) GetNamingConfig(ctx context.Context) (*arr.NamingConfigResource, error) {
	var out arr.NamingConfigResource
	if err := c.base.Get(ctx, "/api/v1/config/naming", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateNamingConfig updates the naming configuration.
func (c *Client) UpdateNamingConfig(ctx context.Context, config *arr.NamingConfigResource) (*arr.NamingConfigResource, error) {
	var out arr.NamingConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/config/naming/%d", config.ID), config, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetNamingConfigByID returns the naming config by its ID.
func (c *Client) GetNamingConfigByID(ctx context.Context, id int) (*arr.NamingConfigResource, error) {
	var out arr.NamingConfigResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/config/naming/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetNamingExamples returns naming format examples based on the current naming config.
func (c *Client) GetNamingExamples(ctx context.Context) (*arr.NamingConfigResource, error) {
	var out arr.NamingConfigResource
	if err := c.base.Get(ctx, "/api/v1/config/naming/examples", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHostConfig returns the host configuration.
func (c *Client) GetHostConfig(ctx context.Context) (*arr.HostConfigResource, error) {
	var out arr.HostConfigResource
	if err := c.base.Get(ctx, "/api/v1/config/host", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateHostConfig updates the host configuration.
func (c *Client) UpdateHostConfig(ctx context.Context, config *arr.HostConfigResource) (*arr.HostConfigResource, error) {
	var out arr.HostConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/config/host/%d", config.ID), config, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHostConfigByID returns the host config by its ID.
func (c *Client) GetHostConfigByID(ctx context.Context, id int) (*arr.HostConfigResource, error) {
	var out arr.HostConfigResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/config/host/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUIConfig returns the UI configuration.
func (c *Client) GetUIConfig(ctx context.Context) (*arr.UIConfigResource, error) {
	var out arr.UIConfigResource
	if err := c.base.Get(ctx, "/api/v1/config/ui", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateUIConfig updates the UI configuration.
func (c *Client) UpdateUIConfig(ctx context.Context, config *arr.UIConfigResource) (*arr.UIConfigResource, error) {
	var out arr.UIConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/config/ui/%d", config.ID), config, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUIConfigByID returns the UI config by its ID.
func (c *Client) GetUIConfigByID(ctx context.Context, id int) (*arr.UIConfigResource, error) {
	var out arr.UIConfigResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/config/ui/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMediaManagementConfig returns the media management configuration.
func (c *Client) GetMediaManagementConfig(ctx context.Context) (*arr.MediaManagementConfigResource, error) {
	var out arr.MediaManagementConfigResource
	if err := c.base.Get(ctx, "/api/v1/config/mediamanagement", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMediaManagementConfig updates the media management configuration.
func (c *Client) UpdateMediaManagementConfig(ctx context.Context, config *arr.MediaManagementConfigResource) (*arr.MediaManagementConfigResource, error) {
	var out arr.MediaManagementConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/config/mediamanagement/%d", config.ID), config, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMediaManagementConfigByID returns the media management config by its ID.
func (c *Client) GetMediaManagementConfigByID(ctx context.Context, id int) (*arr.MediaManagementConfigResource, error) {
	var out arr.MediaManagementConfigResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/config/mediamanagement/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Development Config ----------.

// GetDevelopmentConfig returns the development configuration.
func (c *Client) GetDevelopmentConfig(ctx context.Context) (*DevelopmentConfigResource, error) {
	var out DevelopmentConfigResource
	if err := c.base.Get(ctx, "/api/v1/config/development", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDevelopmentConfigByID returns the development config by its ID.
func (c *Client) GetDevelopmentConfigByID(ctx context.Context, id int) (*DevelopmentConfigResource, error) {
	var out DevelopmentConfigResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/config/development/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateDevelopmentConfig updates the development configuration.
func (c *Client) UpdateDevelopmentConfig(ctx context.Context, config *DevelopmentConfigResource) (*DevelopmentConfigResource, error) {
	var out DevelopmentConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/config/development/%d", config.ID), config, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
