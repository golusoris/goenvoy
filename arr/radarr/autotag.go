package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetAutoTagging returns all auto-tag rules.
func (c *Client) GetAutoTagging(ctx context.Context) ([]arr.AutoTaggingResource, error) {
	return arr.GetList[arr.AutoTaggingResource](ctx, c.base, "/api/v3/autotagging")
}

// GetAutoTag returns a single auto-tag rule by ID.
func (c *Client) GetAutoTag(ctx context.Context, id int) (*arr.AutoTaggingResource, error) {
	return arr.GetOne[arr.AutoTaggingResource](ctx, c.base, fmt.Sprintf("/api/v3/autotagging/%d", id))
}

// CreateAutoTag creates a new auto-tag rule.
func (c *Client) CreateAutoTag(ctx context.Context, at *arr.AutoTaggingResource) (*arr.AutoTaggingResource, error) {
	return arr.PostOne[arr.AutoTaggingResource](ctx, c.base, "/api/v3/autotagging", at)
}

// UpdateAutoTag updates an existing auto-tag rule.
func (c *Client) UpdateAutoTag(ctx context.Context, at *arr.AutoTaggingResource) (*arr.AutoTaggingResource, error) {
	return arr.PutOne[arr.AutoTaggingResource](ctx, c.base, fmt.Sprintf("/api/v3/autotagging/%d", at.ID), at)
}

// DeleteAutoTag removes an auto-tag rule.
func (c *Client) DeleteAutoTag(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/autotagging/%d", id), nil, nil)
}

// GetAutoTagSchema returns available auto-tag specification implementations.
func (c *Client) GetAutoTagSchema(ctx context.Context) ([]arr.AutoTaggingSpecification, error) {
	return arr.GetList[arr.AutoTaggingSpecification](ctx, c.base, "/api/v3/autotagging/schema")
}

// ---------- Backup ----------.
