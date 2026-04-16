package whisparr

import (
	"context"
	"fmt"
	"net/url"
)

// GetAllSeries returns every series configured in the instance.
func (c *Client) GetAllSeries(ctx context.Context) ([]Series, error) {
	var out []Series
	if err := c.base.Get(ctx, "/api/v3/series", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSeries returns a single series by its database ID.
func (c *Client) GetSeries(ctx context.Context, id int) (*Series, error) {
	var out Series
	path := fmt.Sprintf("/api/v3/series/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddSeries adds a new series.
func (c *Client) AddSeries(ctx context.Context, series *Series) (*Series, error) {
	var out Series
	if err := c.base.Post(ctx, "/api/v3/series", series, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSeries updates an existing series. Set moveFiles to true to relocate
// files when the path changes.
func (c *Client) UpdateSeries(ctx context.Context, series *Series, moveFiles bool) (*Series, error) {
	var out Series
	path := fmt.Sprintf("/api/v3/series/%d?moveFiles=%t", series.ID, moveFiles)
	if err := c.base.Put(ctx, path, series, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSeries removes a series by ID.
func (c *Client) DeleteSeries(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v3/series/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupSeries searches for a series by term.
func (c *Client) LookupSeries(ctx context.Context, term string) ([]Series, error) {
	var out []Series
	path := "/api/v3/series/lookup?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateSeasonPass updates monitoring for multiple series at once.
func (c *Client) UpdateSeasonPass(ctx context.Context, pass SeasonPassResource) error {
	return c.base.Post(ctx, "/api/v3/seasonpass", pass, nil)
}

// ---------- AutoTagging ----------.

// EditSeries applies bulk edits to multiple series.
func (c *Client) EditSeries(ctx context.Context, editor *SeriesEditorResource) error {
	return c.base.Put(ctx, "/api/v3/series/editor", editor, nil)
}

// ImportSeries imports one or more series.
func (c *Client) ImportSeries(ctx context.Context, series []Series) error {
	return c.base.Post(ctx, "/api/v3/series/import", series, nil)
}

// ---------- System Extended ----------.

// DeleteSeriesBulk deletes multiple series according to the editor payload.
func (c *Client) DeleteSeriesBulk(ctx context.Context, editor *SeriesEditorResource) error {
	return c.base.Delete(ctx, "/api/v3/series/editor", editor, nil)
}
