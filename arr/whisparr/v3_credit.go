package whisparr

import (
	"context"
	"fmt"
)

// GetCredits returns all credits for a movie/scene.
func (c *ClientV3) GetCredits(ctx context.Context, movieID int) ([]Credit, error) {
	var out []Credit
	path := fmt.Sprintf("/api/v3/credit?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCredit returns a single credit by ID.
func (c *ClientV3) GetCredit(ctx context.Context, id int) (*Credit, error) {
	var out Credit
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/credit/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
