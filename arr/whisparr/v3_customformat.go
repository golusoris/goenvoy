package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCustomFormats returns all custom formats.
func (c *ClientV3) GetCustomFormats(ctx context.Context) ([]arr.CustomFormatResource, error) {
	var out []arr.CustomFormatResource
	if err := c.base.Get(ctx, "/api/v3/customformat", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCustomFormat returns a custom format by ID.
func (c *ClientV3) GetCustomFormat(ctx context.Context, id int) (*arr.CustomFormatResource, error) {
	var out arr.CustomFormatResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/customformat/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateCustomFormat creates a new custom format.
func (c *ClientV3) CreateCustomFormat(ctx context.Context, cf *arr.CustomFormatResource) (*arr.CustomFormatResource, error) {
	var out arr.CustomFormatResource
	if err := c.base.Post(ctx, "/api/v3/customformat", cf, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateCustomFormat updates an existing custom format.
func (c *ClientV3) UpdateCustomFormat(ctx context.Context, cf *arr.CustomFormatResource) (*arr.CustomFormatResource, error) {
	var out arr.CustomFormatResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/customformat/%d", cf.ID), cf, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCustomFormat deletes a custom format by ID.
func (c *ClientV3) DeleteCustomFormat(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/customformat/%d", id), nil, nil)
}

// GetCustomFormatSchema returns the custom format schema.
func (c *ClientV3) GetCustomFormatSchema(ctx context.Context) ([]arr.CustomFormatSpecification, error) {
	var out []arr.CustomFormatSpecification
	if err := c.base.Get(ctx, "/api/v3/customformat/schema", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Delay Profiles ----------.
