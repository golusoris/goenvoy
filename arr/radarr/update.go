package radarr

import (
	"context"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetUpdates returns available application updates.
func (c *Client) GetUpdates(ctx context.Context) ([]arr.UpdateResource, error) {
	return arr.GetList[arr.UpdateResource](ctx, c.base, "/api/v3/update")
}

// ---------- Collections extras ----------.
