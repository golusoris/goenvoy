package readarr

import (
	"context"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// Client is a Readarr API client.
type Client struct {
	base *arr.BaseClient
}

// New creates a Readarr [Client] for the instance at baseURL.
func New(baseURL, apiKey string, opts ...arr.Option) (*Client, error) {
	return arr.NewServiceClient(baseURL, apiKey, func(b *arr.BaseClient) *Client { return &Client{base: b} }, opts...)
}

// Search searches for authors and books by term.
func (c *Client) Search(ctx context.Context, term string) ([]Author, error) {
	path := "/api/v1/search?term=" + url.QueryEscape(term)
	return arr.GetList[Author](ctx, c.base, path)
}

// ReprocessManualImport reprocesses manual imports.
func (c *Client) ReprocessManualImport(ctx context.Context, items []arr.ManualImportReprocessResource) ([]arr.ManualImportResource, error) {
	return arr.PostList[arr.ManualImportResource](ctx, c.base, "/api/v1/manualimport", items)
}
