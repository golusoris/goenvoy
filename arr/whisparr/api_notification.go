package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetNotifications returns all configured notifications.
func (c *api) GetNotifications(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/notification")
}

// GetNotification returns a single notification by ID.
func (c *api) GetNotification(ctx context.Context, id int) (*arr.ProviderResource, error) {
	return arr.GetOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/notification/%d", id))
}

// CreateNotification creates a new notification.
func (c *api) CreateNotification(ctx context.Context, n *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PostOne[arr.ProviderResource](ctx, c.base, "/api/v3/notification", n)
}

// UpdateNotification updates an existing notification.
func (c *api) UpdateNotification(ctx context.Context, n *arr.ProviderResource) (*arr.ProviderResource, error) {
	return arr.PutOne[arr.ProviderResource](ctx, c.base, fmt.Sprintf("/api/v3/notification/%d", n.ID), n)
}

// DeleteNotification deletes a notification by ID.
func (c *api) DeleteNotification(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/notification/%d", id), nil, nil)
}

// GetNotificationSchema returns the schema for all notification types.
func (c *api) GetNotificationSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	return arr.GetList[arr.ProviderResource](ctx, c.base, "/api/v3/notification/schema")
}

// TestNotification tests a notification configuration.
func (c *api) TestNotification(ctx context.Context, n *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/notification/test", n, nil)
}

// TestAllNotifications tests all configured notifications.
func (c *api) TestAllNotifications(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/notification/testall", nil, nil)
}
