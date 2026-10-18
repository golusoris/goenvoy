package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCustomFilters returns all custom filters.
func (c *Client) GetCustomFilters(ctx context.Context) ([]arr.CustomFilterResource, error) {
	return arr.GetList[arr.CustomFilterResource](ctx, c.base, "/api/v3/customfilter")
}

// GetCustomFilter returns a single custom filter by ID.
func (c *Client) GetCustomFilter(ctx context.Context, id int) (*arr.CustomFilterResource, error) {
	return arr.GetOne[arr.CustomFilterResource](ctx, c.base, fmt.Sprintf("/api/v3/customfilter/%d", id))
}

// CreateCustomFilter creates a new custom filter.
func (c *Client) CreateCustomFilter(ctx context.Context, filter *arr.CustomFilterResource) (*arr.CustomFilterResource, error) {
	return arr.PostOne[arr.CustomFilterResource](ctx, c.base, "/api/v3/customfilter", filter)
}

// UpdateCustomFilter updates an existing custom filter.
func (c *Client) UpdateCustomFilter(ctx context.Context, filter *arr.CustomFilterResource) (*arr.CustomFilterResource, error) {
	return arr.PutOne[arr.CustomFilterResource](ctx, c.base, fmt.Sprintf("/api/v3/customfilter/%d", filter.ID), filter)
}

// DeleteCustomFilter removes a custom filter.
func (c *Client) DeleteCustomFilter(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/customfilter/%d", id), nil, nil)
}

// ---------- Custom Formats ----------.
