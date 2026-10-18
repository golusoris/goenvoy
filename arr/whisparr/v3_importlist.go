package whisparr

import "context"

// GetImportListMovies returns movies available from import lists.
func (c *ClientV3) GetImportListMovies(ctx context.Context) ([]Movie, error) {
	var out []Movie
	if err := c.base.Get(ctx, "/api/v3/importlist/movie", &out); err != nil {
		return nil, err
	}
	return out, nil
}
