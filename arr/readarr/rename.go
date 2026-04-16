package readarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetManualImport returns a list of potential imports for the given path.
func (c *Client) GetManualImport(ctx context.Context, folder string) ([]arr.ManualImportResource, error) {
	var out []arr.ManualImportResource
	path := "/api/v1/manualimport?folder=" + url.QueryEscape(folder)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetRenamePreview returns a preview of book file renames for an author.
func (c *Client) GetRenamePreview(ctx context.Context, authorID, bookID int) ([]RenameBookResource, error) {
	var out []RenameBookResource
	path := fmt.Sprintf("/api/v1/rename?authorId=%d&bookId=%d", authorID, bookID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetRetagPreview returns a preview of book file retags for an author.
func (c *Client) GetRetagPreview(ctx context.Context, authorID, bookID int) ([]RetagBookResource, error) {
	var out []RetagBookResource
	path := fmt.Sprintf("/api/v1/retag?authorId=%d&bookId=%d", authorID, bookID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
