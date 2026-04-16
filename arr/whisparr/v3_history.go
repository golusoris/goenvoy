package whisparr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetHistory returns history records (paged).
func (c *ClientV3) GetHistory(ctx context.Context, page, pageSize int) (*arr.PagingResource[HistoryRecordV3], error) {
	var out arr.PagingResource[HistoryRecordV3]
	path := fmt.Sprintf("/api/v3/history?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHistorySince returns history since a specific date.
func (c *ClientV3) GetHistorySince(ctx context.Context, date string) ([]HistoryRecordV3, error) {
	var out []HistoryRecordV3
	path := "/api/v3/history/since?date=" + url.QueryEscape(date)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetHistoryMovie returns history for a specific movie.
func (c *ClientV3) GetHistoryMovie(ctx context.Context, movieID int) ([]HistoryRecordV3, error) {
	var out []HistoryRecordV3
	path := fmt.Sprintf("/api/v3/history/movie?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MarkHistoryFailed marks a history item as failed.
func (c *ClientV3) MarkHistoryFailed(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v3/history/failed/%d", id), nil, nil)
}

// ---------- Host Config ----------.
