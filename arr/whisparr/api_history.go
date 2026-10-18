package whisparr

import (
	"context"
	"fmt"
)

// MarkHistoryFailed marks a history item as failed.
func (c *api) MarkHistoryFailed(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v3/history/failed/%d", id), nil, nil)
}
