package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetExtraFiles returns extra files for a movie.
func (c *ClientV3) GetExtraFiles(ctx context.Context, movieID int) ([]ExtraFileResource, error) {
	path := fmt.Sprintf("/api/v3/extrafile?movieId=%d", movieID)
	return arr.GetList[ExtraFileResource](ctx, c.base, path)
}
