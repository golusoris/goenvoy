package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetAutoTagging returns all auto tagging rules.
func (c *Client) GetAutoTagging(ctx context.Context) ([]arr.AutoTaggingResource, error) {
	return arr.GetList[arr.AutoTaggingResource](ctx, c.base, "/api/v1/autotagging")
}

// GetAutoTag returns a single auto tagging rule by ID.
func (c *Client) GetAutoTag(ctx context.Context, id int) (*arr.AutoTaggingResource, error) {
	return arr.GetOne[arr.AutoTaggingResource](ctx, c.base, fmt.Sprintf("/api/v1/autotagging/%d", id))
}

// CreateAutoTag creates a new auto tagging rule.
func (c *Client) CreateAutoTag(ctx context.Context, at *arr.AutoTaggingResource) (*arr.AutoTaggingResource, error) {
	return arr.PostOne[arr.AutoTaggingResource](ctx, c.base, "/api/v1/autotagging", at)
}

// UpdateAutoTag updates an existing auto tagging rule.
func (c *Client) UpdateAutoTag(ctx context.Context, at *arr.AutoTaggingResource) (*arr.AutoTaggingResource, error) {
	return arr.PutOne[arr.AutoTaggingResource](ctx, c.base, fmt.Sprintf("/api/v1/autotagging/%d", at.ID), at)
}

// DeleteAutoTag deletes an auto tagging rule by ID.
func (c *Client) DeleteAutoTag(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/autotagging/%d", id), nil, nil)
}

// GetAutoTagSchema returns the auto tagging schema.
func (c *Client) GetAutoTagSchema(ctx context.Context) ([]arr.AutoTaggingResource, error) {
	return arr.GetList[arr.AutoTaggingResource](ctx, c.base, "/api/v1/autotagging/schema")
}

// ---------- Import List Exclusions Extended ----------.
