package readarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetManualImport returns a list of potential imports for the given path.
func (c *Client) GetManualImport(ctx context.Context, folder string) ([]arr.ManualImportResource, error) {
	path := "/api/v1/manualimport?folder=" + url.QueryEscape(folder)
	return arr.GetList[arr.ManualImportResource](ctx, c.base, path)
}

// GetRenamePreview returns a preview of book file renames for an author.
func (c *Client) GetRenamePreview(ctx context.Context, authorID, bookID int) ([]RenameBookResource, error) {
	path := fmt.Sprintf("/api/v1/rename?authorId=%d&bookId=%d", authorID, bookID)
	return arr.GetList[RenameBookResource](ctx, c.base, path)
}

// GetRetagPreview returns a preview of book file retags for an author.
func (c *Client) GetRetagPreview(ctx context.Context, authorID, bookID int) ([]RetagBookResource, error) {
	path := fmt.Sprintf("/api/v1/retag?authorId=%d&bookId=%d", authorID, bookID)
	return arr.GetList[RetagBookResource](ctx, c.base, path)
}
