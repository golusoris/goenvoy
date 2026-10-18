package whisparr

import (
	"context"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetIndexerFlags returns all available indexer flags.
func (c *ClientV3) GetIndexerFlags(ctx context.Context) ([]arr.IndexerFlagResource, error) {
	return arr.GetList[arr.IndexerFlagResource](ctx, c.base, "/api/v3/indexerflag")
}
