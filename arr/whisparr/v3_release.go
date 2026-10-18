package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// SearchReleases searches for available releases.
func (c *ClientV3) SearchReleases(ctx context.Context, movieID int) ([]arr.ReleaseResource, error) {
	path := fmt.Sprintf("/api/v3/release?movieId=%d", movieID)
	return arr.GetList[arr.ReleaseResource](ctx, c.base, path)
}
