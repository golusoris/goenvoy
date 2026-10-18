package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCustomFormats returns all custom formats.
func (c *api) GetCustomFormats(ctx context.Context) ([]arr.CustomFormatResource, error) {
	return arr.GetList[arr.CustomFormatResource](ctx, c.base, "/api/v3/customformat")
}

// GetCustomFormat returns a custom format by ID.
func (c *api) GetCustomFormat(ctx context.Context, id int) (*arr.CustomFormatResource, error) {
	return arr.GetOne[arr.CustomFormatResource](ctx, c.base, fmt.Sprintf("/api/v3/customformat/%d", id))
}

// CreateCustomFormat creates a new custom format.
func (c *api) CreateCustomFormat(ctx context.Context, cf *arr.CustomFormatResource) (*arr.CustomFormatResource, error) {
	return arr.PostOne[arr.CustomFormatResource](ctx, c.base, "/api/v3/customformat", cf)
}

// UpdateCustomFormat updates an existing custom format.
func (c *api) UpdateCustomFormat(ctx context.Context, cf *arr.CustomFormatResource) (*arr.CustomFormatResource, error) {
	return arr.PutOne[arr.CustomFormatResource](ctx, c.base, fmt.Sprintf("/api/v3/customformat/%d", cf.ID), cf)
}

// DeleteCustomFormat deletes a custom format by ID.
func (c *api) DeleteCustomFormat(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/customformat/%d", id), nil, nil)
}

// GetCustomFormatSchema returns the custom format schema.
func (c *api) GetCustomFormatSchema(ctx context.Context) ([]arr.CustomFormatSpecification, error) {
	return arr.GetList[arr.CustomFormatSpecification](ctx, c.base, "/api/v3/customformat/schema")
}
