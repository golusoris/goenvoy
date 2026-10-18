package lidarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRenameList returns a list of track files that would be renamed.
func (c *Client) GetRenameList(ctx context.Context, artistID int) ([]RenameTrackResource, error) {
	path := fmt.Sprintf("/api/v1/rename?artistId=%d", artistID)
	return arr.GetList[RenameTrackResource](ctx, c.base, path)
}

// ---------- Retag ----------.

// GetManualImport returns a list of items available for manual import.
func (c *Client) GetManualImport(ctx context.Context, folder string) ([]arr.ManualImportResource, error) {
	path := "/api/v1/manualimport?folder=" + url.QueryEscape(folder)
	return arr.GetList[arr.ManualImportResource](ctx, c.base, path)
}

// ProcessManualImport re-evaluates manual import candidates (POST /api/v1/manualimport)
// and imports nothing; importing is the ManualImport command.
func (c *Client) ProcessManualImport(ctx context.Context, imports []arr.ManualImportReprocessResource) error {
	return c.base.Post(ctx, "/api/v1/manualimport", imports, nil)
}

// ---------- Backups ----------.

// GetRetag returns a list of track files that would be retagged.
func (c *Client) GetRetag(ctx context.Context, artistID int) ([]RetagResource, error) {
	path := fmt.Sprintf("/api/v1/retag?artistId=%d", artistID)
	return arr.GetList[RetagResource](ctx, c.base, path)
}
