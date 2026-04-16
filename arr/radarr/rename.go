package radarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRenameList returns proposed renames for a movie.
func (c *Client) GetRenameList(ctx context.Context, movieID int) ([]arr.RenameMovieResource, error) {
	var out []arr.RenameMovieResource
	path := fmt.Sprintf("/api/v3/rename?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Manual Import ----------.

// GetManualImport returns files available for manual import.
func (c *Client) GetManualImport(ctx context.Context, folder, downloadID string) ([]arr.ManualImportResource, error) {
	var out []arr.ManualImportResource
	path := fmt.Sprintf("/api/v3/manualimport?folder=%s&downloadId=%s",
		url.QueryEscape(folder), url.QueryEscape(downloadID))
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ProcessManualImport confirms and processes a manual import.
func (c *Client) ProcessManualImport(ctx context.Context, imports []arr.ManualImportReprocessResource) error {
	return c.base.Post(ctx, "/api/v3/manualimport", imports, nil)
}

// ---------- Logs ----------.
