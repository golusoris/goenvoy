package whisparr

import (
	"context"
	"net/url"
	"strconv"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQueueDetailsByMovieID returns detailed queue info for specific movie IDs.
func (c *ClientV3) GetQueueDetailsByMovieID(ctx context.Context, movieIDs []int) ([]arr.QueueRecord, error) {
	var out []arr.QueueRecord
	vals := make(url.Values)
	for _, id := range movieIDs {
		vals.Add("movieIds", strconv.Itoa(id))
	}
	path := "/api/v3/queue/details?" + vals.Encode()
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
