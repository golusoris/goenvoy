package radarr

import (
	"context"
	"fmt"
)

// GetCollections returns all movie collections.
func (c *Client) GetCollections(ctx context.Context) ([]Collection, error) {
	var out []Collection
	if err := c.base.Get(ctx, "/api/v3/collection", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCollection returns a single collection by its database ID.
func (c *Client) GetCollection(ctx context.Context, id int) (*Collection, error) {
	var out Collection
	path := fmt.Sprintf("/api/v3/collection/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateCollection updates an existing collection.
func (c *Client) UpdateCollection(ctx context.Context, collection *Collection) (*Collection, error) {
	var out Collection
	path := fmt.Sprintf("/api/v3/collection/%d", collection.ID)
	if err := c.base.Put(ctx, path, collection, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateCollections performs a bulk update on multiple collections.
func (c *Client) UpdateCollections(ctx context.Context, collections []Collection) error {
	return c.base.Put(ctx, "/api/v3/collection", collections, nil)
}

// ---------- Command ----------.
