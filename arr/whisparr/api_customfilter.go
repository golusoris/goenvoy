package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCustomFilters returns all custom filters.
func (c *api) GetCustomFilters(ctx context.Context) ([]arr.CustomFilterResource, error) {
	var out []arr.CustomFilterResource
	if err := c.base.Get(ctx, "/api/v3/customfilter", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCustomFilter returns a single custom filter by ID.
func (c *api) GetCustomFilter(ctx context.Context, id int) (*arr.CustomFilterResource, error) {
	var out arr.CustomFilterResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/customfilter/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateCustomFilter creates a new custom filter.
func (c *api) CreateCustomFilter(ctx context.Context, f *arr.CustomFilterResource) (*arr.CustomFilterResource, error) {
	var out arr.CustomFilterResource
	if err := c.base.Post(ctx, "/api/v3/customfilter", f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateCustomFilter updates an existing custom filter.
func (c *api) UpdateCustomFilter(ctx context.Context, f *arr.CustomFilterResource) (*arr.CustomFilterResource, error) {
	var out arr.CustomFilterResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/customfilter/%d", f.ID), f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCustomFilter deletes a custom filter by ID.
func (c *api) DeleteCustomFilter(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/customfilter/%d", id), nil, nil)
}
