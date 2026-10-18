package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCredits returns cast and crew credits for a movie.
func (c *Client) GetCredits(ctx context.Context, movieID int) ([]Credit, error) {
	var out []Credit
	path := fmt.Sprintf("/api/v3/credit?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCreditByID returns a single credit by its ID.
func (c *Client) GetCreditByID(ctx context.Context, id int) (*Credit, error) {
	return arr.GetOne[Credit](ctx, c.base, fmt.Sprintf("/api/v3/credit/%d", id))
}

// ---------- Config By-ID Gets ----------.
