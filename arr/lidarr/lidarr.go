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
	base, err := arr.NewBaseClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{base: base}, nil
}

// Search performs a general search and returns matching artists and albums.
func (c *Client) Search(ctx context.Context, term string) ([]SearchResult, error) {
	var out []SearchResult
	path := "/api/v1/search?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
