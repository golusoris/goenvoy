package sonarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetAutoTagging returns all auto-tag rules.
func (c *Client) GetAutoTagging(ctx context.Context) ([]arr.AutoTaggingResource, error) {
	var out []arr.AutoTaggingResource
	if err := c.base.Get(ctx, "/api/v3/autotagging", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAutoTag returns a single auto-tag rule by ID.
func (c *Client) GetAutoTag(ctx context.Context, id int) (*arr.AutoTaggingResource, error) {
	var out arr.AutoTaggingResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/autotagging/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateAutoTag creates a new auto-tag rule.
func (c *Client) CreateAutoTag(ctx context.Context, at *arr.AutoTaggingResource) (*arr.AutoTaggingResource, error) {
	var out arr.AutoTaggingResource
	if err := c.base.Post(ctx, "/api/v3/autotagging", at, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAutoTag updates an existing auto-tag rule.
func (c *Client) UpdateAutoTag(ctx context.Context, at *arr.AutoTaggingResource) (*arr.AutoTaggingResource, error) {
	var out arr.AutoTaggingResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/autotagging/%d", at.ID), at, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAutoTag removes an auto-tag rule.
func (c *Client) DeleteAutoTag(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/autotagging/%d", id), nil, nil)
}

// GetAutoTagSchema returns available auto-tag specification implementations.
func (c *Client) GetAutoTagSchema(ctx context.Context) ([]arr.AutoTaggingSpecification, error) {
	var out []arr.AutoTaggingSpecification
	if err := c.base.Get(ctx, "/api/v3/autotagging/schema", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Backup ----------.
