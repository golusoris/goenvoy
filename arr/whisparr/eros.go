package whisparr

import (
	"github.com/golusoris/goenvoy/arr/v2"
)

// ClientV3 is a Whisparr v3 (Radarr-based) API client.
type ClientV3 struct {
	base *arr.BaseClient
}

// NewV3 creates a Whisparr v3 [ClientV3] for the instance at baseURL.
func NewV3(baseURL, apiKey string, opts ...arr.Option) (*ClientV3, error) {
	base, err := arr.NewBaseClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientV3{base: base}, nil
}
