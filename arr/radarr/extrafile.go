package radarr

import (
	"context"
	"fmt"
)

// GetExtraFiles returns extra files for a given movie.
func (c *Client) GetExtraFiles(ctx context.Context, movieID int) ([]ExtraFileResource, error) {
	var out []ExtraFileResource
	path := fmt.Sprintf("/api/v3/extrafile?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Import List Exclusion extras ----------.
