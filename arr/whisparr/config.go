package whisparr

import (
	"context"
	"fmt"
)

// GetImportListConfig returns the import list configuration.
func (c *Client) GetImportListConfig(ctx context.Context) (*ImportListConfigResource, error) {
	var out ImportListConfigResource
	if err := c.base.Get(ctx, "/api/v3/config/importlist", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateImportListConfig updates the import list configuration.
func (c *Client) UpdateImportListConfig(ctx context.Context, cfg *ImportListConfigResource) (*ImportListConfigResource, error) {
	var out ImportListConfigResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/config/importlist/%d", cfg.ID), cfg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
