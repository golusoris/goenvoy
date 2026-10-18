package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCustomFormats returns all custom formats.
func (c *Client) GetCustomFormats(ctx context.Context) ([]arr.CustomFormatResource, error) {
	return arr.GetList[arr.CustomFormatResource](ctx, c.base, "/api/v1/customformat")
}

// GetCustomFormat returns a single custom format by its ID.
func (c *Client) GetCustomFormat(ctx context.Context, id int) (*arr.CustomFormatResource, error) {
	return arr.GetOne[arr.CustomFormatResource](ctx, c.base, fmt.Sprintf("/api/v1/customformat/%d", id))
}

// CreateCustomFormat creates a new custom format.
func (c *Client) CreateCustomFormat(ctx context.Context, format *arr.CustomFormatResource) (*arr.CustomFormatResource, error) {
	return arr.PostOne[arr.CustomFormatResource](ctx, c.base, "/api/v1/customformat", format)
}

// UpdateCustomFormat updates an existing custom format.
func (c *Client) UpdateCustomFormat(ctx context.Context, format *arr.CustomFormatResource) (*arr.CustomFormatResource, error) {
	return arr.PutOne[arr.CustomFormatResource](ctx, c.base, fmt.Sprintf("/api/v1/customformat/%d", format.ID), format)
}

// DeleteCustomFormat deletes a custom format by ID.
func (c *Client) DeleteCustomFormat(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/customformat/%d", id), nil, nil)
}

// GetCustomFormatSchema returns the custom format schema.
func (c *Client) GetCustomFormatSchema(ctx context.Context) ([]arr.CustomFormatResource, error) {
	return arr.GetList[arr.CustomFormatResource](ctx, c.base, "/api/v1/customformat/schema")
}

// ---------- Delay Profiles ----------.
