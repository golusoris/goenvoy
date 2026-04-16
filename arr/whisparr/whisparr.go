package whisparr

import (
	"github.com/golusoris/goenvoy/arr/v2"
)

// Client is a Whisparr v2 (Sonarr-based) API client.
type Client struct {
	base *arr.BaseClient
}

// New creates a Whisparr v2 [Client] for the instance at baseURL.
func New(baseURL, apiKey string, opts ...arr.Option) (*Client, error) {
	base, err := arr.NewBaseClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{base: base}, nil
}
