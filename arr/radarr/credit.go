package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCredits returns cast and crew credits for a movie.
func (c *Client) GetCredits(ctx context.Context, movieID int) ([]Credit, error) {
	path := fmt.Sprintf("/api/v3/credit?movieId=%d", movieID)
	return arr.GetList[Credit](ctx, c.base, path)
}

// GetCreditByID returns a single credit by its ID.
func (c *Client) GetCreditByID(ctx context.Context, id int) (*Credit, error) {
	return arr.GetOne[Credit](ctx, c.base, fmt.Sprintf("/api/v3/credit/%d", id))
}

// ---------- Config By-ID Gets ----------.
