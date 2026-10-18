package whisparr

import (
	"context"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetIndexerFlags returns all available indexer flags.
func (c *ClientV3) GetIndexerFlags(ctx context.Context) ([]arr.IndexerFlagResource, error) {
	var out []arr.IndexerFlagResource
	if err := c.base.Get(ctx, "/api/v3/indexerflag", &out); err != nil {
		return nil, err
	}
	return out, nil
}
