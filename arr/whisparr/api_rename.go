package whisparr

import (
	"context"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetManualImport returns items available for manual import.
func (c *api) GetManualImport(ctx context.Context, folder string) ([]arr.ManualImportResource, error) {
	path := "/api/v3/manualimport?folder=" + url.QueryEscape(folder)
	return arr.GetList[arr.ManualImportResource](ctx, c.base, path)
}

// ProcessManualImport re-evaluates manual import candidates (POST /api/v3/manualimport)
// and imports nothing; importing is the ManualImport command.
func (c *api) ProcessManualImport(ctx context.Context, items []arr.ManualImportReprocessResource) error {
	return c.base.Post(ctx, "/api/v3/manualimport", items, nil)
}
