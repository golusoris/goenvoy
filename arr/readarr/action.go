package readarr

import (
	"context"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// NotificationAction triggers a named action on a notification provider.
func (c *Client) NotificationAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v1/notification/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}

// ---------- Download Clients ----------.

// DownloadClientAction triggers a named action on a download client provider.
func (c *Client) DownloadClientAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v1/downloadclient/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}

// ---------- Indexers ----------.

// IndexerAction triggers a named action on an indexer provider.
func (c *Client) IndexerAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v1/indexer/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}

// ---------- Import Lists ----------.

// ImportListAction triggers a named action on an import list provider.
func (c *Client) ImportListAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	path := "/api/v1/importlist/action/" + url.PathEscape(name)
	return c.base.Post(ctx, path, body, nil)
}

// ---------- Metadata Consumers ----------.
