package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRenamePreview returns a rename preview for a movie.
func (c *ClientV3) GetRenamePreview(ctx context.Context, movieID int) ([]arr.RenameEpisodeResource, error) {
	path := fmt.Sprintf("/api/v3/rename?movieId=%d", movieID)
	return arr.GetList[arr.RenameEpisodeResource](ctx, c.base, path)
}
