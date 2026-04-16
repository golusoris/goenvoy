package readarr

import (
	"context"
	"fmt"
	"net/url"
)

// GetBooks returns books for the given author.
func (c *Client) GetBooks(ctx context.Context, authorID int) ([]Book, error) {
	var out []Book
	path := fmt.Sprintf("/api/v1/book?authorId=%d", authorID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBook returns a single book by its database ID.
func (c *Client) GetBook(ctx context.Context, id int) (*Book, error) {
	var out Book
	path := fmt.Sprintf("/api/v1/book/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddBook adds a new book to Readarr.
func (c *Client) AddBook(ctx context.Context, book *Book) (*Book, error) {
	var out Book
	if err := c.base.Post(ctx, "/api/v1/book", book, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateBook updates an existing book.
func (c *Client) UpdateBook(ctx context.Context, book *Book) (*Book, error) {
	var out Book
	path := fmt.Sprintf("/api/v1/book/%d", book.ID)
	if err := c.base.Put(ctx, path, book, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBook removes a book.
func (c *Client) DeleteBook(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v1/book/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupBook searches for a book by term.
func (c *Client) LookupBook(ctx context.Context, term string) ([]Book, error) {
	var out []Book
	path := "/api/v1/book/lookup?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MonitorBooks sets the monitored status for a list of books.
func (c *Client) MonitorBooks(ctx context.Context, req *BooksMonitoredResource) error {
	return c.base.Put(ctx, "/api/v1/book/monitor", req, nil)
}

// EditBooks performs a batch update on multiple books.
func (c *Client) EditBooks(ctx context.Context, editor *BookEditorResource) error {
	return c.base.Put(ctx, "/api/v1/book/editor", editor, nil)
}

// DeleteBooks performs a batch delete of multiple books.
func (c *Client) DeleteBooks(ctx context.Context, editor *BookEditorResource) error {
	return c.base.Delete(ctx, "/api/v1/book/editor", editor, nil)
}

// Bookshelf performs batch monitoring changes on authors and books.
func (c *Client) Bookshelf(ctx context.Context, shelf *BookshelfResource) error {
	return c.base.Post(ctx, "/api/v1/bookshelf", shelf, nil)
}

// GetBookOverview returns an overview for a specific book.
func (c *Client) GetBookOverview(ctx context.Context, id int) (*Book, error) {
	var out Book
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/book/%d/overview", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
