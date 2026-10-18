package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetDelayProfiles returns all delay profiles.
func (c *Client) GetDelayProfiles(ctx context.Context) ([]arr.DelayProfileResource, error) {
	return arr.GetList[arr.DelayProfileResource](ctx, c.base, "/api/v1/delayprofile")
}

// GetDelayProfile returns a single delay profile by its ID.
func (c *Client) GetDelayProfile(ctx context.Context, id int) (*arr.DelayProfileResource, error) {
	return arr.GetOne[arr.DelayProfileResource](ctx, c.base, fmt.Sprintf("/api/v1/delayprofile/%d", id))
}

// CreateDelayProfile creates a new delay profile.
func (c *Client) CreateDelayProfile(ctx context.Context, profile *arr.DelayProfileResource) (*arr.DelayProfileResource, error) {
	return arr.PostOne[arr.DelayProfileResource](ctx, c.base, "/api/v1/delayprofile", profile)
}

// UpdateDelayProfile updates an existing delay profile.
func (c *Client) UpdateDelayProfile(ctx context.Context, profile *arr.DelayProfileResource) (*arr.DelayProfileResource, error) {
	return arr.PutOne[arr.DelayProfileResource](ctx, c.base, fmt.Sprintf("/api/v1/delayprofile/%d", profile.ID), profile)
}

// DeleteDelayProfile deletes a delay profile by ID.
func (c *Client) DeleteDelayProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/delayprofile/%d", id), nil, nil)
}

// ReorderDelayProfile changes the order of a delay profile.
func (c *Client) ReorderDelayProfile(ctx context.Context, id, afterID int) ([]arr.DelayProfileResource, error) {
	path := fmt.Sprintf("/api/v1/delayprofile/reorder/%d?after=%d", id, afterID)
	return arr.PutList[arr.DelayProfileResource](ctx, c.base, path, nil)
}

// ---------- Release Profiles ----------.
