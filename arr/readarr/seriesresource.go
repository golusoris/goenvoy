package readarr

import (
	"context"
	"fmt"
)

// GetSeries returns book series for the given author.
func (c *Client) GetSeries(ctx context.Context, authorID int) ([]Series, error) {
	var out []Series
	path := fmt.Sprintf("/api/v1/series?authorId=%d", authorID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
