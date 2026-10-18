package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetBlocklistMovie returns blocklist entries for a specific movie.
func (c *ClientV3) GetBlocklistMovie(ctx context.Context, movieID int) ([]arr.BlocklistResource, error) {
	var out []arr.BlocklistResource
	path := fmt.Sprintf("/api/v3/blocklist/movie?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
