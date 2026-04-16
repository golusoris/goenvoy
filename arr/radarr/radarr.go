package radarr

import (
	"github.com/golusoris/goenvoy/arr/v2"
)

// Client is a Radarr API client.
type Client struct {
	base *arr.BaseClient
}

// New creates a Radarr [Client] for the instance at baseURL.
func New(baseURL, apiKey string, opts ...arr.Option) (*Client, error) {
	base, err := arr.NewBaseClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{base: base}, nil
}
