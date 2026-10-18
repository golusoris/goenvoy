package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// SearchReleases searches for available releases.
func (c *Client) SearchReleases(ctx context.Context, episodeID int) ([]arr.ReleaseResource, error) {
	path := fmt.Sprintf("/api/v3/release?episodeId=%d", episodeID)
	return arr.GetList[arr.ReleaseResource](ctx, c.base, path)
}

// GetReleaseProfiles returns all release profiles.
func (c *Client) GetReleaseProfiles(ctx context.Context) ([]arr.ReleaseProfileResource, error) {
	return arr.GetList[arr.ReleaseProfileResource](ctx, c.base, "/api/v3/releaseprofile")
}

// GetReleaseProfile returns a release profile by ID.
func (c *Client) GetReleaseProfile(ctx context.Context, id int) (*arr.ReleaseProfileResource, error) {
	return arr.GetOne[arr.ReleaseProfileResource](ctx, c.base, fmt.Sprintf("/api/v3/releaseprofile/%d", id))
}

// CreateReleaseProfile creates a new release profile.
func (c *Client) CreateReleaseProfile(ctx context.Context, rp *arr.ReleaseProfileResource) (*arr.ReleaseProfileResource, error) {
	return arr.PostOne[arr.ReleaseProfileResource](ctx, c.base, "/api/v3/releaseprofile", rp)
}

// UpdateReleaseProfile updates an existing release profile.
func (c *Client) UpdateReleaseProfile(ctx context.Context, rp *arr.ReleaseProfileResource) (*arr.ReleaseProfileResource, error) {
	return arr.PutOne[arr.ReleaseProfileResource](ctx, c.base, fmt.Sprintf("/api/v3/releaseprofile/%d", rp.ID), rp)
}

// DeleteReleaseProfile deletes a release profile by ID.
func (c *Client) DeleteReleaseProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/releaseprofile/%d", id), nil, nil)
}
