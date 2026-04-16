package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetImportListExclusions returns all import list exclusions.
func (c *Client) GetImportListExclusions(ctx context.Context) ([]ImportListExclusion, error) {
	var out []ImportListExclusion
	if err := c.base.Get(ctx, "/api/v3/exclusions", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetImportLists returns all import list configurations.
func (c *Client) GetImportLists(ctx context.Context) ([]arr.ProviderResource, error) {
	var out []arr.ProviderResource
	if err := c.base.Get(ctx, "/api/v3/importlist", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetImportList returns a single import list by ID.
func (c *Client) GetImportList(ctx context.Context, id int) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/importlist/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateImportList creates a new import list.
func (c *Client) CreateImportList(ctx context.Context, il *arr.ProviderResource) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Post(ctx, "/api/v3/importlist", il, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateImportList updates an existing import list.
func (c *Client) UpdateImportList(ctx context.Context, il *arr.ProviderResource) (*arr.ProviderResource, error) {
	var out arr.ProviderResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/importlist/%d", il.ID), il, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteImportList removes an import list.
func (c *Client) DeleteImportList(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/importlist/%d", id), nil, nil)
}

// GetImportListSchema returns available import list implementations.
func (c *Client) GetImportListSchema(ctx context.Context) ([]arr.ProviderResource, error) {
	var out []arr.ProviderResource
	if err := c.base.Get(ctx, "/api/v3/importlist/schema", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TestImportList sends a test for an import list configuration.
func (c *Client) TestImportList(ctx context.Context, il *arr.ProviderResource) error {
	return c.base.Post(ctx, "/api/v3/importlist/test", il, nil)
}

// ---------- Import List Exclusions (full CRUD) ----------.

// CreateImportListExclusion creates a new import list exclusion.
func (c *Client) CreateImportListExclusion(ctx context.Context, excl *ImportListExclusion) (*ImportListExclusion, error) {
	var out ImportListExclusion
	if err := c.base.Post(ctx, "/api/v3/exclusions", excl, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteImportListExclusion removes an import list exclusion.
func (c *Client) DeleteImportListExclusion(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/exclusions/%d", id), nil, nil)
}

// ---------- Metadata ----------.

// UpdateImportListsBulk performs a bulk update of import lists.
func (c *Client) UpdateImportListsBulk(ctx context.Context, body *arr.ProviderBulkResource) ([]arr.ProviderResource, error) {
	var out []arr.ProviderResource
	if err := c.base.Put(ctx, "/api/v3/importlist/bulk", body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteImportListsBulk bulk-deletes import lists by IDs.
func (c *Client) DeleteImportListsBulk(ctx context.Context, ids []int) error {
	body := &arr.ProviderBulkResource{IDs: ids}
	return c.base.Delete(ctx, "/api/v3/importlist/bulk", body, nil)
}

// TestAllImportLists tests all configured import lists.
func (c *Client) TestAllImportLists(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/importlist/testall", nil, nil)
}

// ---------- Import List Config ----------.

// GetImportListExclusionsPaged returns a paginated list of import list exclusions.
func (c *Client) GetImportListExclusionsPaged(ctx context.Context, page, pageSize int) (*arr.PagingResource[ImportListExclusion], error) {
	var out arr.PagingResource[ImportListExclusion]
	path := fmt.Sprintf("/api/v3/exclusions/paged?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteImportListExclusionsBulk bulk-deletes import list exclusions by IDs.
func (c *Client) DeleteImportListExclusionsBulk(ctx context.Context, ids []int) error {
	body := struct {
		IDs []int `json:"ids"`
	}{IDs: ids}
	return c.base.Delete(ctx, "/api/v3/exclusions/bulk", body, nil)
}

// ---------- Notification / Metadata TestAll ----------.

// GetImportListExclusion returns a single import list exclusion by ID.
func (c *Client) GetImportListExclusion(ctx context.Context, id int) (*ImportListExclusion, error) {
	var out ImportListExclusion
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/exclusions/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateImportListExclusionsBulk creates multiple import list exclusions in bulk.
func (c *Client) CreateImportListExclusionsBulk(ctx context.Context, exclusions []ImportListExclusion) ([]ImportListExclusion, error) {
	var out []ImportListExclusion
	if err := c.base.Post(ctx, "/api/v3/exclusions/bulk", exclusions, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateImportListExclusion updates an existing import list exclusion.
func (c *Client) UpdateImportListExclusion(ctx context.Context, exclusion *ImportListExclusion) (*ImportListExclusion, error) {
	var out ImportListExclusion
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/exclusions/%d", exclusion.ID), exclusion, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Import List Movies ----------.

// GetImportListMovies returns movies from import lists.
func (c *Client) GetImportListMovies(ctx context.Context) ([]Movie, error) {
	var out []Movie
	if err := c.base.Get(ctx, "/api/v3/importlist/movie", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateImportListMovies adds movies from import lists.
func (c *Client) CreateImportListMovies(ctx context.Context, movies []Movie) ([]Movie, error) {
	var out []Movie
	if err := c.base.Post(ctx, "/api/v3/importlist/movie", movies, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Provider Action endpoints ----------.
