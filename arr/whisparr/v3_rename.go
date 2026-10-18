package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRenamePreview returns a rename preview for a movie.
func (c *ClientV3) GetRenamePreview(ctx context.Context, movieID int) ([]arr.RenameEpisodeResource, error) {
	var out []arr.RenameEpisodeResource
	path := fmt.Sprintf("/api/v3/rename?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
