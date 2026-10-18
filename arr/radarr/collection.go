package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCollections returns all movie collections.
func (c *Client) GetCollections(ctx context.Context) ([]Collection, error) {
	return arr.GetList[Collection](ctx, c.base, "/api/v3/collection")
}

// GetCollection returns a single collection by its database ID.
func (c *Client) GetCollection(ctx context.Context, id int) (*Collection, error) {
	path := fmt.Sprintf("/api/v3/collection/%d", id)
	return arr.GetOne[Collection](ctx, c.base, path)
}

// UpdateCollection updates an existing collection.
func (c *Client) UpdateCollection(ctx context.Context, collection *Collection) (*Collection, error) {
	path := fmt.Sprintf("/api/v3/collection/%d", collection.ID)
	return arr.PutOne[Collection](ctx, c.base, path, collection)
}

// UpdateCollections performs a bulk update on multiple collections.
func (c *Client) UpdateCollections(ctx context.Context, collections []Collection) error {
	return c.base.Put(ctx, "/api/v3/collection", collections, nil)
}

// ---------- Command ----------.
