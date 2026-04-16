package whisparr

import (
	"context"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// DownloadClientAction triggers a named action on a download client.
func (c *ClientV3) DownloadClientAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/downloadclient/action/"+url.PathEscape(name), body, nil)
}

// ---------- Download Client Config ----------.

// ImportListAction triggers a named action on an import list.
func (c *ClientV3) ImportListAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/importlist/action/"+url.PathEscape(name), body, nil)
}

// ---------- Import List Movies ----------.

// IndexerAction triggers a named action on an indexer.
func (c *ClientV3) IndexerAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/indexer/action/"+url.PathEscape(name), body, nil)
}

// ---------- Indexer Config ----------.

// MetadataAction triggers a named action on a metadata consumer.
func (c *ClientV3) MetadataAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/metadata/action/"+url.PathEscape(name), body, nil)
}

// ---------- Movie Extended ----------.

// NotificationAction triggers a named action on a notification.
func (c *ClientV3) NotificationAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/notification/action/"+url.PathEscape(name), body, nil)
}

// ---------- Performer Editor ----------.
