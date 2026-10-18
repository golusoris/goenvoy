package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
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
	return arr.GetOne[Credit](ctx, c.base, fmt.Sprintf("/api/v3/credit/%d", id))
}
