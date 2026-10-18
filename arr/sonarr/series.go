package sonarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetAllSeries returns every series configured in Sonarr.
func (c *Client) GetAllSeries(ctx context.Context) ([]Series, error) {
	return arr.GetList[Series](ctx, c.base, "/api/v3/series")
}

// GetSeries returns a single series by its database ID.
func (c *Client) GetSeries(ctx context.Context, id int) (*Series, error) {
	path := fmt.Sprintf("/api/v3/series/%d", id)
	return arr.GetOne[Series](ctx, c.base, path)
}

// AddSeries adds a new series to Sonarr.
func (c *Client) AddSeries(ctx context.Context, series *Series) (*Series, error) {
	return arr.PostOne[Series](ctx, c.base, "/api/v3/series", series)
}

// UpdateSeries updates an existing series. Set moveFiles to true to relocate
// files when the series path changes.
func (c *Client) UpdateSeries(ctx context.Context, series *Series, moveFiles bool) (*Series, error) {
	path := fmt.Sprintf("/api/v3/series/%d?moveFiles=%t", series.ID, moveFiles)
	return arr.PutOne[Series](ctx, c.base, path, series)
}

// DeleteSeries removes a series. Set deleteFiles to true to also delete
// downloaded episode files from disk.
func (c *Client) DeleteSeries(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v3/series/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupSeries searches for a series by term (title or TVDB ID slug).
func (c *Client) LookupSeries(ctx context.Context, term string) ([]Series, error) {
	path := "/api/v3/series/lookup?term=" + url.QueryEscape(term)
	return arr.GetList[Series](ctx, c.base, path)
}

// UpdateSeasonPass updates monitored status for multiple series and seasons.
func (c *Client) UpdateSeasonPass(ctx context.Context, pass SeasonPassResource) error {
	return c.base.Post(ctx, "/api/v3/seasonpass", pass, nil)
}

// ---------- Series Editor ----------.

// EditSeries applies bulk edits to multiple series.
func (c *Client) EditSeries(ctx context.Context, editor *SeriesEditorResource) ([]Series, error) {
	return arr.PutList[Series](ctx, c.base, "/api/v3/series/editor", editor)
}

// DeleteManySeries deletes multiple series at once.
func (c *Client) DeleteManySeries(ctx context.Context, editor *SeriesEditorResource) error {
	return c.base.Delete(ctx, "/api/v3/series/editor", editor, nil)
}

// ImportSeries imports one or more series in bulk.
func (c *Client) ImportSeries(ctx context.Context, series []Series) ([]Series, error) {
	return arr.PostList[Series](ctx, c.base, "/api/v3/series/import", series)
}

// ---------- Wanted ----------.

// GetSeriesFolder returns folder information for a series.
func (c *Client) GetSeriesFolder(ctx context.Context, seriesID int) error {
	return c.base.Get(ctx, fmt.Sprintf("/api/v3/series/%d/folder", seriesID), nil)
}

// ---------- Calendar By ID ----------.
