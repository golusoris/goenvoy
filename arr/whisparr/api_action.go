package whisparr

import (
	"context"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// DownloadClientAction triggers a named action on a download client.
func (c *api) DownloadClientAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/downloadclient/action/"+url.PathEscape(name), body, nil)
}

// ImportListAction triggers a named action on an import list.
func (c *api) ImportListAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/importlist/action/"+url.PathEscape(name), body, nil)
}

// IndexerAction triggers a named action on an indexer.
func (c *api) IndexerAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/indexer/action/"+url.PathEscape(name), body, nil)
}

// MetadataAction triggers a named action on a metadata consumer.
func (c *api) MetadataAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/metadata/action/"+url.PathEscape(name), body, nil)
}

// NotificationAction triggers a named action on a notification.
func (c *api) NotificationAction(ctx context.Context, name string, body *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/notification/action/"+url.PathEscape(name), body, nil)
}
