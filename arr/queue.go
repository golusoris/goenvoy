package arr

import (
	"context"
	"net/url"
	"strconv"
)

// GetQueuePage gets one page of the download queue at path (such as /api/v3/queue).
// Each include flag set to true is added to the query as name=true; false flags are left
// out, so a service maps its own option fields to its own parameter names.
func GetQueuePage[T any](ctx context.Context, c *BaseClient, path string, page, pageSize int, include map[string]bool) (*PagingResource[T], error) {
	q := make(url.Values, len(include)+2)
	q.Set("page", strconv.Itoa(page))
	q.Set("pageSize", strconv.Itoa(pageSize))
	for name, on := range include {
		if on {
			q.Set(name, "true")
		}
	}
	return GetOne[PagingResource[T]](ctx, c, path+"?"+q.Encode())
}
