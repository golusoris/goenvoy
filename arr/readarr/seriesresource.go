package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetSeries returns book series for the given author.
func (c *Client) GetSeries(ctx context.Context, authorID int) ([]Series, error) {
	path := fmt.Sprintf("/api/v1/series?authorId=%d", authorID)
	return arr.GetList[Series](ctx, c.base, path)
}
