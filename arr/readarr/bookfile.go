package readarr

import (
	"context"
	"fmt"
)

// GetBookFiles returns all book files for the given author.
func (c *Client) GetBookFiles(ctx context.Context, authorID int) ([]BookFile, error) {
	var out []BookFile
	path := fmt.Sprintf("/api/v1/bookfile?authorId=%d", authorID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBookFile returns a single book file by its database ID.
func (c *Client) GetBookFile(ctx context.Context, id int) (*BookFile, error) {
	var out BookFile
	path := fmt.Sprintf("/api/v1/bookfile/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBookFile removes a single book file by its database ID.
func (c *Client) DeleteBookFile(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/bookfile/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// DeleteBookFiles removes multiple book files by their IDs.
func (c *Client) DeleteBookFiles(ctx context.Context, ids []int) error {
	body := BookFileListResource{BookFileIDs: ids}
	return c.base.Delete(ctx, "/api/v1/bookfile/bulk", &body, nil)
}

// UpdateBookFile updates a single book file.
func (c *Client) UpdateBookFile(ctx context.Context, file *BookFile) (*BookFile, error) {
	var out BookFile
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/bookfile/%d", file.ID), file, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EditBookFilesBulk updates multiple book files at once.
func (c *Client) EditBookFilesBulk(ctx context.Context, editor *BookFileListResource) ([]BookFile, error) {
	var out []BookFile
	if err := c.base.Put(ctx, "/api/v1/bookfile/editor", editor, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Bookshelf ----------.
