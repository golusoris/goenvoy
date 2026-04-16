package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetAutoTagging returns all auto tagging rules.
func (c *Client) GetAutoTagging(ctx context.Context) ([]arr.AutoTaggingResource, error) {
	var out []arr.AutoTaggingResource
	if err := c.base.Get(ctx, "/api/v1/autotagging", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAutoTag returns a single auto tagging rule by ID.
func (c *Client) GetAutoTag(ctx context.Context, id int) (*arr.AutoTaggingResource, error) {
	var out arr.AutoTaggingResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/autotagging/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateAutoTag creates a new auto tagging rule.
func (c *Client) CreateAutoTag(ctx context.Context, at *arr.AutoTaggingResource) (*arr.AutoTaggingResource, error) {
	var out arr.AutoTaggingResource
	if err := c.base.Post(ctx, "/api/v1/autotagging", at, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAutoTag updates an existing auto tagging rule.
func (c *Client) UpdateAutoTag(ctx context.Context, at *arr.AutoTaggingResource) (*arr.AutoTaggingResource, error) {
	var out arr.AutoTaggingResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/autotagging/%d", at.ID), at, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAutoTag deletes an auto tagging rule by ID.
func (c *Client) DeleteAutoTag(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/autotagging/%d", id), nil, nil)
}

// GetAutoTagSchema returns the auto tagging schema.
func (c *Client) GetAutoTagSchema(ctx context.Context) ([]arr.AutoTaggingResource, error) {
	var out []arr.AutoTaggingResource
	if err := c.base.Get(ctx, "/api/v1/autotagging/schema", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Import List Exclusions Extended ----------.
