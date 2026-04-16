package whisparr

import (
	"context"
	"net/url"
)

// LookupScene searches for a scene by term.
func (c *ClientV3) LookupScene(ctx context.Context, term string) ([]Movie, error) {
	var out []Movie
	path := "/api/v3/lookup/scene?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
