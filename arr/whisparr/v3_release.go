package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// SearchReleases searches for available releases.
func (c *ClientV3) SearchReleases(ctx context.Context, movieID int) ([]arr.ReleaseResource, error) {
	var out []arr.ReleaseResource
	path := fmt.Sprintf("/api/v3/release?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
