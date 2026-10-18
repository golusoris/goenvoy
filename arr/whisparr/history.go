package whisparr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetHistory returns history records (paged).
func (c *Client) GetHistory(ctx context.Context, page, pageSize int) (*arr.PagingResource[V2HistoryRecord], error) {
	path := fmt.Sprintf("/api/v3/history?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[V2HistoryRecord]](ctx, c.base, path)
}

// GetHistorySince returns history since a specific date.
func (c *Client) GetHistorySince(ctx context.Context, date string) ([]V2HistoryRecord, error) {
	path := "/api/v3/history/since?date=" + url.QueryEscape(date)
	return arr.GetList[V2HistoryRecord](ctx, c.base, path)
}

// GetHistorySeries returns history for a specific series.
func (c *Client) GetHistorySeries(ctx context.Context, seriesID int) ([]V2HistoryRecord, error) {
	path := fmt.Sprintf("/api/v3/history/series?seriesId=%d", seriesID)
	return arr.GetList[V2HistoryRecord](ctx, c.base, path)
}
