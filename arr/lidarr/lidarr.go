package lidarr

import (
	"context"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// Client is a Lidarr API client.
type Client struct {
	base *arr.BaseClient
}

// New creates a Lidarr [Client] for the instance at baseURL.
func New(baseURL, apiKey string, opts ...arr.Option) (*Client, error) {
	return arr.NewServiceClient(baseURL, apiKey, func(b *arr.BaseClient) *Client { return &Client{base: b} }, opts...)
}

// Search performs a general search and returns matching artists and albums.
func (c *Client) Search(ctx context.Context, term string) ([]SearchResult, error) {
	path := "/api/v1/search?term=" + url.QueryEscape(term)
	return arr.GetList[SearchResult](ctx, c.base, path)
}
