package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetHealthByID returns a single health check by ID.
func (c *ClientV3) GetHealthByID(ctx context.Context, id int) (*arr.HealthCheck, error) {
	return arr.GetOne[arr.HealthCheck](ctx, c.base, fmt.Sprintf("/api/v3/health/%d", id))
}
