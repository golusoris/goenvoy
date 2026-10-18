package metadata

import (
	"context"
	"net/url"
	"strconv"
)

// PostFunc is the shape of a service's own POST helper: it sends body to path and decodes the answer into dst.
type PostFunc func(ctx context.Context, path string, body, dst any) error

// PostAs sends body to path with post and returns the decoded T.
// A service method whose body is one POST plus a decode calls it instead of repeating the var out block.
func PostAs[T any](ctx context.Context, post PostFunc, path string, body any) (*T, error) {
	var out T
	if err := post(ctx, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PageParams returns page and limit query parameters; a value of zero or less is left out.
func PageParams(page, limit int) url.Values {
	p := url.Values{}
	if page > 0 {
		p.Set("page", strconv.Itoa(page))
	}
	if limit > 0 {
		p.Set("limit", strconv.Itoa(limit))
	}
	return p
}
