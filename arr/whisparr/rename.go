package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRenamePreview returns a rename preview for a series.
func (c *Client) GetRenamePreview(ctx context.Context, seriesID int) ([]arr.RenameEpisodeResource, error) {
	path := fmt.Sprintf("/api/v3/rename?seriesId=%d", seriesID)
	return arr.GetList[arr.RenameEpisodeResource](ctx, c.base, path)
}
