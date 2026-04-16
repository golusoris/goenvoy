package lidarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRenameList returns a list of track files that would be renamed.
func (c *Client) GetRenameList(ctx context.Context, artistID int) ([]RenameTrackResource, error) {
	var out []RenameTrackResource
	path := fmt.Sprintf("/api/v1/rename?artistId=%d", artistID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Retag ----------.

// GetManualImport returns a list of items available for manual import.
func (c *Client) GetManualImport(ctx context.Context, folder string) ([]arr.ManualImportResource, error) {
	var out []arr.ManualImportResource
	path := "/api/v1/manualimport?folder=" + url.QueryEscape(folder)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ProcessManualImport processes manual imports.
func (c *Client) ProcessManualImport(ctx context.Context, imports []arr.ManualImportReprocessResource) error {
	return c.base.Post(ctx, "/api/v1/manualimport", imports, nil)
}

// ---------- Backups ----------.

// GetRetag returns a list of track files that would be retagged.
func (c *Client) GetRetag(ctx context.Context, artistID int) ([]RetagResource, error) {
	var out []RetagResource
	path := fmt.Sprintf("/api/v1/retag?artistId=%d", artistID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
