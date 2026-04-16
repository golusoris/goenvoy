package whisparr

import (
	"context"
	"fmt"
)

// GetAlternativeTitles returns alternative titles for a movie.
func (c *ClientV3) GetAlternativeTitles(ctx context.Context, movieID int) ([]AlternativeTitleResource, error) {
	var out []AlternativeTitleResource
	path := fmt.Sprintf("/api/v3/alttitle?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAlternativeTitle returns a single alternative title by ID.
func (c *ClientV3) GetAlternativeTitle(ctx context.Context, id int) (*AlternativeTitleResource, error) {
	var out AlternativeTitleResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/alttitle/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
