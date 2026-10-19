package whisparr

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQueueDetailsByMovieID returns detailed queue info for specific movie IDs.
func (c *ClientV3) GetQueueDetailsByMovieID(ctx context.Context, movieIDs []int) ([]arr.QueueRecord, error) {
	vals := make(url.Values)
	for _, id := range movieIDs {
		vals.Add("movieIds", strconv.Itoa(id))
	}
	path := "/api/v3/queue/details?" + vals.Encode()
	return arr.GetList[arr.QueueRecord](ctx, c.base, path)
}

// GetQueueWithOptions returns the download queue using advanced filtering and inclusion options for Whisparr v3.
func (c *ClientV3) GetQueueWithOptions(ctx context.Context, opts QueueOptions) (*arr.PagingResource[QueueRecord], error) {
	// opts.IncludeParent is ignored because Whisparr v3 (Radarr-based) has no parent entity in queue.
	return arr.GetQueuePage[QueueRecord](ctx, c.base, "/api/v3/queue", opts.Page, opts.PageSize, map[string]bool{
		"includeUnknownMovieItems": opts.IncludeUnknown,
		"includeMovie":             opts.IncludeItem,
	})
}

// DeleteQueueItemWithOptions removes an item from the download queue with extended options for Whisparr v3.
func (c *ClientV3) DeleteQueueItemWithOptions(ctx context.Context, id int, opts QueueDeleteOptions) error {
	path := fmt.Sprintf("/api/v3/queue/%d?removeFromClient=%t&blocklist=%t&skipRedownload=%t&changeCategory=%t",
		id, opts.RemoveFromClient, opts.Blocklist, opts.SkipRedownload, opts.ChangeCategory)
	return c.base.Delete(ctx, path, nil, nil)
}
