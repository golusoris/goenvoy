package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetImportListConfig returns the import list configuration.
func (c *Client) GetImportListConfig(ctx context.Context) (*ImportListConfigResource, error) {
	return arr.GetOne[ImportListConfigResource](ctx, c.base, "/api/v3/config/importlist")
}

// UpdateImportListConfig updates the import list configuration.
func (c *Client) UpdateImportListConfig(ctx context.Context, cfg *ImportListConfigResource) (*ImportListConfigResource, error) {
	return arr.PutOne[ImportListConfigResource](ctx, c.base, fmt.Sprintf("/api/v3/config/importlist/%d", cfg.ID), cfg)
}
