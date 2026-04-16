package sonarr

import (
	"context"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// DownloadClientAction triggers a named action on a download client provider.
func (c *Client) DownloadClientAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v3/downloadclient/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}

// ImportListAction triggers a named action on an import list provider.
func (c *Client) ImportListAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v3/importlist/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}

// IndexerAction triggers a named action on an indexer provider.
func (c *Client) IndexerAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v3/indexer/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}

// MetadataAction triggers a named action on a metadata provider.
func (c *Client) MetadataAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v3/metadata/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}

// NotificationAction triggers a named action on a notification provider.
func (c *Client) NotificationAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v3/notification/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}

// ---------- Language Profile ----------.
