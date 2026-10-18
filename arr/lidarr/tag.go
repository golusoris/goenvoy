package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetTags returns all tags.
func (c *Client) GetTags(ctx context.Context) ([]arr.Tag, error) {
	return arr.GetList[arr.Tag](ctx, c.base, "/api/v1/tag")
}

// CreateTag creates a new tag and returns it with its assigned ID.
func (c *Client) CreateTag(ctx context.Context, label string) (*arr.Tag, error) {
	return arr.PostOne[arr.Tag](ctx, c.base, "/api/v1/tag", arr.Tag{Label: label})
}

// GetTag returns a single tag by ID.
func (c *Client) GetTag(ctx context.Context, id int) (*arr.Tag, error) {
	return arr.GetOne[arr.Tag](ctx, c.base, fmt.Sprintf("/api/v1/tag/%d", id))
}

// UpdateTag updates an existing tag.
func (c *Client) UpdateTag(ctx context.Context, tag *arr.Tag) (*arr.Tag, error) {
	return arr.PutOne[arr.Tag](ctx, c.base, fmt.Sprintf("/api/v1/tag/%d", tag.ID), tag)
}

// DeleteTag deletes a tag by ID.
func (c *Client) DeleteTag(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/tag/%d", id), nil, nil)
}

// GetTagDetails returns all tags with usage details.
func (c *Client) GetTagDetails(ctx context.Context) ([]arr.TagDetail, error) {
	return arr.GetList[arr.TagDetail](ctx, c.base, "/api/v1/tag/detail")
}

// GetTagDetail returns a single tag with usage details.
func (c *Client) GetTagDetail(ctx context.Context, id int) (*arr.TagDetail, error) {
	return arr.GetOne[arr.TagDetail](ctx, c.base, fmt.Sprintf("/api/v1/tag/detail/%d", id))
}

// ---------- Root Folders ----------.
