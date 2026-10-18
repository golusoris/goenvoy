package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCredits returns all credits for a movie/scene.
func (c *ClientV3) GetCredits(ctx context.Context, movieID int) ([]Credit, error) {
	path := fmt.Sprintf("/api/v3/credit?movieId=%d", movieID)
	return arr.GetList[Credit](ctx, c.base, path)
}

// GetCredit returns a single credit by ID.
func (c *ClientV3) GetCredit(ctx context.Context, id int) (*Credit, error) {
	return arr.GetOne[Credit](ctx, c.base, fmt.Sprintf("/api/v3/credit/%d", id))
}
