package radarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetHistory returns the download history with pagination.
func (c *Client) GetHistory(ctx context.Context, page, pageSize int) (*arr.PagingResource[HistoryRecord], error) {
	var out arr.PagingResource[HistoryRecord]
	path := fmt.Sprintf("/api/v3/history?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHistoryMovie returns the history for a specific movie.
func (c *Client) GetHistoryMovie(ctx context.Context, movieID int) ([]HistoryRecord, error) {
	var out []HistoryRecord
	path := fmt.Sprintf("/api/v3/history/movie?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetHistorySince returns history records since a given date.
func (c *Client) GetHistorySince(ctx context.Context, date string) ([]HistoryRecord, error) {
	var out []HistoryRecord
	path := "/api/v3/history/since?date=" + url.QueryEscape(date)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MarkHistoryFailed marks a history record as failed to trigger re-download.
func (c *Client) MarkHistoryFailed(ctx context.Context, historyID int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v3/history/failed/%d", historyID), nil, nil)
}

// ---------- Languages ----------.
