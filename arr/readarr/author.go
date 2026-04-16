package readarr

import (
	"context"
	"fmt"
	"net/url"
)

// GetAllAuthors returns every author configured in Readarr.
func (c *Client) GetAllAuthors(ctx context.Context) ([]Author, error) {
	var out []Author
	if err := c.base.Get(ctx, "/api/v1/author", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAuthor returns a single author by its database ID.
func (c *Client) GetAuthor(ctx context.Context, id int) (*Author, error) {
	var out Author
	path := fmt.Sprintf("/api/v1/author/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddAuthor adds a new author to Readarr.
func (c *Client) AddAuthor(ctx context.Context, author *Author) (*Author, error) {
	var out Author
	if err := c.base.Post(ctx, "/api/v1/author", author, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAuthor updates an existing author. Set moveFiles to true to relocate
// files when the author path changes.
func (c *Client) UpdateAuthor(ctx context.Context, author *Author, moveFiles bool) (*Author, error) {
	var out Author
	path := fmt.Sprintf("/api/v1/author/%d?moveFiles=%t", author.ID, moveFiles)
	if err := c.base.Put(ctx, path, author, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAuthor removes an author. Set deleteFiles to true to also delete
// downloaded files from disk.
func (c *Client) DeleteAuthor(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v1/author/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupAuthor searches for an author by name.
func (c *Client) LookupAuthor(ctx context.Context, term string) ([]Author, error) {
	var out []Author
	path := "/api/v1/author/lookup?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EditAuthors performs a batch update on multiple authors.
func (c *Client) EditAuthors(ctx context.Context, editor *AuthorEditorResource) error {
	return c.base.Put(ctx, "/api/v1/author/editor", editor, nil)
}

// DeleteAuthors performs a batch delete of multiple authors.
func (c *Client) DeleteAuthors(ctx context.Context, editor *AuthorEditorResource) error {
	return c.base.Delete(ctx, "/api/v1/author/editor", editor, nil)
}
