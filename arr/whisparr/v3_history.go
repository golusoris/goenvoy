package whisparr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetHistory returns history records (paged).
func (c *ClientV3) GetHistory(ctx context.Context, page, pageSize int) (*arr.PagingResource[HistoryRecordV3], error) {
	path := fmt.Sprintf("/api/v3/history?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[HistoryRecordV3]](ctx, c.base, path)
}

// GetHistorySince returns history since a specific date.
func (c *ClientV3) GetHistorySince(ctx context.Context, date string) ([]HistoryRecordV3, error) {
	path := "/api/v3/history/since?date=" + url.QueryEscape(date)
	return arr.GetList[HistoryRecordV3](ctx, c.base, path)
}

// GetHistoryMovie returns history for a specific movie.
func (c *ClientV3) GetHistoryMovie(ctx context.Context, movieID int) ([]HistoryRecordV3, error) {
	path := fmt.Sprintf("/api/v3/history/movie?movieId=%d", movieID)
	return arr.GetList[HistoryRecordV3](ctx, c.base, path)
}
