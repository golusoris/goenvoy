package whisparr

import (
	"context"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetImportListMovies returns movies available from import lists.
func (c *ClientV3) GetImportListMovies(ctx context.Context) ([]Movie, error) {
	return arr.GetList[Movie](ctx, c.base, "/api/v3/importlist/movie")
}
