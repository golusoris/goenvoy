package whisparr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetManualImport returns items available for manual import.
func (c *ClientV3) GetManualImport(ctx context.Context, folder string) ([]arr.ManualImportResource, error) {
	var out []arr.ManualImportResource
	path := "/api/v3/manualimport?folder=" + url.QueryEscape(folder)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ProcessManualImport triggers processing of manual import selections.
func (c *ClientV3) ProcessManualImport(ctx context.Context, items []arr.ManualImportReprocessResource) error {
	return c.base.Post(ctx, "/api/v3/manualimport", items, nil)
}

// ---------- Media Management Config ----------.

// GetRenamePreview returns a rename preview for a movie.
func (c *ClientV3) GetRenamePreview(ctx context.Context, movieID int) ([]arr.RenameEpisodeResource, error) {
	var out []arr.RenameEpisodeResource
	path := fmt.Sprintf("/api/v3/rename?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
