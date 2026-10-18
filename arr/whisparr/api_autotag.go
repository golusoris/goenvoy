package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetAutoTags returns all auto tagging rules.
func (c *api) GetAutoTags(ctx context.Context) ([]arr.AutoTaggingResource, error) {
	return arr.GetList[arr.AutoTaggingResource](ctx, c.base, "/api/v3/autotagging")
}

// GetAutoTag returns a single auto tag by ID.
func (c *api) GetAutoTag(ctx context.Context, id int) (*arr.AutoTaggingResource, error) {
	return arr.GetOne[arr.AutoTaggingResource](ctx, c.base, fmt.Sprintf("/api/v3/autotagging/%d", id))
}

// CreateAutoTag creates a new auto tagging rule.
func (c *api) CreateAutoTag(ctx context.Context, tag *arr.AutoTaggingResource) (*arr.AutoTaggingResource, error) {
	return arr.PostOne[arr.AutoTaggingResource](ctx, c.base, "/api/v3/autotagging", tag)
}

// UpdateAutoTag updates an existing auto tagging rule.
func (c *api) UpdateAutoTag(ctx context.Context, tag *arr.AutoTaggingResource) (*arr.AutoTaggingResource, error) {
	return arr.PutOne[arr.AutoTaggingResource](ctx, c.base, fmt.Sprintf("/api/v3/autotagging/%d", tag.ID), tag)
}

// DeleteAutoTag deletes an auto tagging rule by ID.
func (c *api) DeleteAutoTag(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/autotagging/%d", id), nil, nil)
}

// GetAutoTagSchema returns the auto tagging schema.
func (c *api) GetAutoTagSchema(ctx context.Context) ([]arr.AutoTaggingSpecification, error) {
	return arr.GetList[arr.AutoTaggingSpecification](ctx, c.base, "/api/v3/autotagging/schema")
}
