package sonarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRenameList returns proposed renames for a series.
func (c *Client) GetRenameList(ctx context.Context, seriesID int) ([]arr.RenameEpisodeResource, error) {
	path := fmt.Sprintf("/api/v3/rename?seriesId=%d", seriesID)
	return arr.GetList[arr.RenameEpisodeResource](ctx, c.base, path)
}

// ---------- Manual Import ----------.

// GetManualImport returns files available for manual import.
func (c *Client) GetManualImport(ctx context.Context, folder, downloadID string) ([]arr.ManualImportResource, error) {
	path := fmt.Sprintf("/api/v3/manualimport?folder=%s&downloadId=%s",
		url.QueryEscape(folder), url.QueryEscape(downloadID))
	return arr.GetList[arr.ManualImportResource](ctx, c.base, path)
}

// ProcessManualImport re-evaluates manual import candidates (POST /api/v3/manualimport)
// and imports nothing; SendManualImport imports files.
func (c *Client) ProcessManualImport(ctx context.Context, imports []arr.ManualImportReprocessResource) error {
	return c.base.Post(ctx, "/api/v3/manualimport", imports, nil)
}

// ---------- Logs ----------.
