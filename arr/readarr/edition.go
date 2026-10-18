package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetEditions returns all editions for the given book IDs.
func (c *Client) GetEditions(ctx context.Context, bookID int) ([]Edition, error) {
	path := fmt.Sprintf("/api/v1/edition?bookId=%d", bookID)
	return arr.GetList[Edition](ctx, c.base, path)
}
