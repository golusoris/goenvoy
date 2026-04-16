package whisparr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetHistory returns history records (paged).
func (c *Client) GetHistory(ctx context.Context, page, pageSize int) (*arr.PagingResource[V2HistoryRecord], error) {
	var out arr.PagingResource[V2HistoryRecord]
	path := fmt.Sprintf("/api/v3/history?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHistorySince returns history since a specific date.
func (c *Client) GetHistorySince(ctx context.Context, date string) ([]V2HistoryRecord, error) {
	var out []V2HistoryRecord
	path := "/api/v3/history/since?date=" + url.QueryEscape(date)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetHistorySeries returns history for a specific series.
func (c *Client) GetHistorySeries(ctx context.Context, seriesID int) ([]V2HistoryRecord, error) {
	var out []V2HistoryRecord
	path := fmt.Sprintf("/api/v3/history/series?seriesId=%d", seriesID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MarkHistoryFailed marks a history item as failed.
func (c *Client) MarkHistoryFailed(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v3/history/failed/%d", id), nil, nil)
}

// ---------- Import Lists ----------.
