package lidarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetHistory returns the download history with pagination.
func (c *Client) GetHistory(ctx context.Context, page, pageSize int) (*arr.PagingResource[HistoryRecord], error) {
	path := fmt.Sprintf("/api/v1/history?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[HistoryRecord]](ctx, c.base, path)
}

// GetHistoryArtist returns history for a specific artist.
func (c *Client) GetHistoryArtist(ctx context.Context, artistID int) ([]HistoryRecord, error) {
	path := fmt.Sprintf("/api/v1/history/artist?artistId=%d", artistID)
	return arr.GetList[HistoryRecord](ctx, c.base, path)
}

// GetHistorySince returns history records since a given date.
func (c *Client) GetHistorySince(ctx context.Context, date string) ([]HistoryRecord, error) {
	path := "/api/v1/history/since?date=" + url.QueryEscape(date)
	return arr.GetList[HistoryRecord](ctx, c.base, path)
}

// MarkHistoryFailed marks a history item as failed.
func (c *Client) MarkHistoryFailed(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v1/history/failed/%d", id), nil, nil)
}

// ---------- Releases ----------.
