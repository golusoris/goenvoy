package radarr

import (
	"context"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetUpdates returns available application updates.
func (c *Client) GetUpdates(ctx context.Context) ([]arr.UpdateResource, error) {
	var out []arr.UpdateResource
	if err := c.base.Get(ctx, "/api/v3/update", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Collections extras ----------.
