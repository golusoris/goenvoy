package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetBlocklistMovie returns blocklist entries for a specific movie.
func (c *ClientV3) GetBlocklistMovie(ctx context.Context, movieID int) ([]arr.BlocklistResource, error) {
	path := fmt.Sprintf("/api/v3/blocklist/movie?movieId=%d", movieID)
	return arr.GetList[arr.BlocklistResource](ctx, c.base, path)
}
