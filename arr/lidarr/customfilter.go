package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCustomFilters returns all custom filters.
func (c *Client) GetCustomFilters(ctx context.Context) ([]arr.CustomFilterResource, error) {
	return arr.GetList[arr.CustomFilterResource](ctx, c.base, "/api/v1/customfilter")
}

// GetCustomFilter returns a single custom filter by ID.
func (c *Client) GetCustomFilter(ctx context.Context, id int) (*arr.CustomFilterResource, error) {
	return arr.GetOne[arr.CustomFilterResource](ctx, c.base, fmt.Sprintf("/api/v1/customfilter/%d", id))
}

// CreateCustomFilter creates a new custom filter.
func (c *Client) CreateCustomFilter(ctx context.Context, cf *arr.CustomFilterResource) (*arr.CustomFilterResource, error) {
	return arr.PostOne[arr.CustomFilterResource](ctx, c.base, "/api/v1/customfilter", cf)
}

// UpdateCustomFilter updates an existing custom filter.
func (c *Client) UpdateCustomFilter(ctx context.Context, cf *arr.CustomFilterResource) (*arr.CustomFilterResource, error) {
	return arr.PutOne[arr.CustomFilterResource](ctx, c.base, fmt.Sprintf("/api/v1/customfilter/%d", cf.ID), cf)
}

// DeleteCustomFilter deletes a custom filter by ID.
func (c *Client) DeleteCustomFilter(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/customfilter/%d", id), nil, nil)
}

// ---------- Custom Formats ----------.
