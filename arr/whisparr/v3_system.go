package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetHealthByID returns a single health check by ID.
func (c *ClientV3) GetHealthByID(ctx context.Context, id int) (*arr.HealthCheck, error) {
	var out arr.HealthCheck
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/health/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
