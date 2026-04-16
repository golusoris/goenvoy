package whisparr

import (
	"context"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// DownloadClientAction triggers a named action on a download client.
func (c *Client) DownloadClientAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/downloadclient/action/"+url.PathEscape(name), body, nil)
}

// ---------- Download Client Config ----------.

// ImportListAction triggers a named action on an import list.
func (c *Client) ImportListAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/importlist/action/"+url.PathEscape(name), body, nil)
}

// ---------- Import List Config ----------.

// IndexerAction triggers a named action on an indexer.
func (c *Client) IndexerAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/indexer/action/"+url.PathEscape(name), body, nil)
}

// ---------- Indexer Config ----------.

// MetadataAction triggers a named action on a metadata consumer.
func (c *Client) MetadataAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/metadata/action/"+url.PathEscape(name), body, nil)
}

// ---------- Naming Config ----------.

// NotificationAction triggers a named action on a notification.
func (c *Client) NotificationAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/notification/action/"+url.PathEscape(name), body, nil)
}

// ---------- Quality Definitions ----------.
