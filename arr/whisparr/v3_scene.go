package whisparr

import (
	"context"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// LookupScene searches for a scene by term.
func (c *ClientV3) LookupScene(ctx context.Context, term string) ([]Movie, error) {
	path := "/api/v3/lookup/scene?term=" + url.QueryEscape(term)
	return arr.GetList[Movie](ctx, c.base, path)
}
