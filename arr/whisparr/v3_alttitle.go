package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetAlternativeTitles returns alternative titles for a movie.
func (c *ClientV3) GetAlternativeTitles(ctx context.Context, movieID int) ([]AlternativeTitleResource, error) {
	path := fmt.Sprintf("/api/v3/alttitle?movieId=%d", movieID)
	return arr.GetList[AlternativeTitleResource](ctx, c.base, path)
}

// GetAlternativeTitle returns a single alternative title by ID.
func (c *ClientV3) GetAlternativeTitle(ctx context.Context, id int) (*AlternativeTitleResource, error) {
	return arr.GetOne[AlternativeTitleResource](ctx, c.base, fmt.Sprintf("/api/v3/alttitle/%d", id))
}
