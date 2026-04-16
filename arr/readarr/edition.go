package readarr

import (
	"context"
	"fmt"
)

// GetEditions returns all editions for the given book IDs.
func (c *Client) GetEditions(ctx context.Context, bookID int) ([]Edition, error) {
	var out []Edition
	path := fmt.Sprintf("/api/v1/edition?bookId=%d", bookID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
