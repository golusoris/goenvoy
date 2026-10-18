package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCustomFilters returns all custom filters.
func (c *api) GetCustomFilters(ctx context.Context) ([]arr.CustomFilterResource, error) {
	return arr.GetList[arr.CustomFilterResource](ctx, c.base, "/api/v3/customfilter")
}

// GetCustomFilter returns a single custom filter by ID.
func (c *api) GetCustomFilter(ctx context.Context, id int) (*arr.CustomFilterResource, error) {
	return arr.GetOne[arr.CustomFilterResource](ctx, c.base, fmt.Sprintf("/api/v3/customfilter/%d", id))
}

// CreateCustomFilter creates a new custom filter.
func (c *api) CreateCustomFilter(ctx context.Context, f *arr.CustomFilterResource) (*arr.CustomFilterResource, error) {
	return arr.PostOne[arr.CustomFilterResource](ctx, c.base, "/api/v3/customfilter", f)
}

// UpdateCustomFilter updates an existing custom filter.
func (c *api) UpdateCustomFilter(ctx context.Context, f *arr.CustomFilterResource) (*arr.CustomFilterResource, error) {
	return arr.PutOne[arr.CustomFilterResource](ctx, c.base, fmt.Sprintf("/api/v3/customfilter/%d", f.ID), f)
}

// DeleteCustomFilter deletes a custom filter by ID.
func (c *api) DeleteCustomFilter(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/customfilter/%d", id), nil, nil)
}
