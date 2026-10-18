package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetTags returns all tags.
func (c *api) GetTags(ctx context.Context) ([]arr.Tag, error) {
	return arr.GetList[arr.Tag](ctx, c.base, "/api/v3/tag")
}

// CreateTag creates a new tag with the given label.
func (c *api) CreateTag(ctx context.Context, label string) (*arr.Tag, error) {
	return arr.PostOne[arr.Tag](ctx, c.base, "/api/v3/tag", arr.Tag{Label: label})
}

// GetTag returns a single tag by ID.
func (c *api) GetTag(ctx context.Context, id int) (*arr.Tag, error) {
	return arr.GetOne[arr.Tag](ctx, c.base, fmt.Sprintf("/api/v3/tag/%d", id))
}

// UpdateTag updates an existing tag.
func (c *api) UpdateTag(ctx context.Context, tag *arr.Tag) (*arr.Tag, error) {
	return arr.PutOne[arr.Tag](ctx, c.base, fmt.Sprintf("/api/v3/tag/%d", tag.ID), tag)
}

// DeleteTag deletes a tag by ID.
func (c *api) DeleteTag(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/tag/%d", id), nil, nil)
}

// GetTagDetails returns all tags with usage details.
func (c *api) GetTagDetails(ctx context.Context) ([]arr.TagDetail, error) {
	return arr.GetList[arr.TagDetail](ctx, c.base, "/api/v3/tag/detail")
}

// GetTagDetail returns a single tag detail by ID.
func (c *api) GetTagDetail(ctx context.Context, id int) (*arr.TagDetail, error) {
	return arr.GetOne[arr.TagDetail](ctx, c.base, fmt.Sprintf("/api/v3/tag/detail/%d", id))
}
