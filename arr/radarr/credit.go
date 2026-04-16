package radarr

import (
	"context"
	"fmt"
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
	var out Credit
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/credit/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Config By-ID Gets ----------.
