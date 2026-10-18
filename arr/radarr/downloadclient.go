package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetDownloadClients returns all download client configurations.
func (c *Client) GetDownloadClients(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/downloadclient")
}

// GetDownloadClient returns a single download client by ID.
func (c *Client) GetDownloadClient(ctx context.Context, id int) (*arr.ProviderResource, error) {
	return arr.GetOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/downloadclient/%d", id))
}

// CreateDownloadClient creates a new download client.
func (c *Client) CreateDownloadClient(ctx context.Context, dc *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PostOne[arr.ProviderResource](ctx, c.base, "/api/v3/downloadclient", dc)
}

// UpdateDownloadClient updates an existing download client.
func (c *Client) UpdateDownloadClient(ctx context.Context, dc *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PutOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/downloadclient/%d", dc.ID), dc)
}

// DeleteDownloadClient removes a download client.
func (c *Client) DeleteDownloadClient(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/downloadclient/%d", id), nil, nil)
}

// GetDownloadClientSchema returns available download client implementations.
func (c *Client) GetDownloadClientSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/downloadclient/schema")
}

// TestDownloadClient sends a test for a download client configuration.
func (c *Client) TestDownloadClient(ctx context.Context, dc *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/downloadclient/test", dc, nil)
}

// UpdateDownloadClientsBulk performs a bulk update of download clients.
func (c *Client) UpdateDownloadClientsBulk(ctx context.Context, body *arr.ProviderBulkResource) ([]arr.ProviderResource, error) {
	return arr.PutList[arr.ProviderResource](ctx, c.base, "/api/v3/downloadclient/bulk", body)
}

// DeleteDownloadClientsBulk bulk-deletes download clients by IDs.
func (c *Client) DeleteDownloadClientsBulk(ctx context.Context, ids []int) error {
	body := &arr.ProviderBulkResource{IDs: ids}
	return c.base.Delete(ctx, "/api/v3/downloadclient/bulk", body, nil)
}

// TestAllDownloadClients tests all configured download clients.
func (c *Client) TestAllDownloadClients(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/downloadclient/testall", nil, nil)
}

// ---------- Indexer Bulk ----------.
