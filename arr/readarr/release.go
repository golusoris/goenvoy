package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetReleaseProfiles returns all release profiles.
func (c *Client) GetReleaseProfiles(ctx context.Context) ([]arr.ReleaseProfileResource, error) {
	return arr.GetList[arr.ReleaseProfileResource](ctx, c.base, "/api/v1/releaseprofile")
}

// GetReleaseProfile returns a single release profile by its ID.
func (c *Client) GetReleaseProfile(ctx context.Context, id int) (*arr.ReleaseProfileResource, error) {
	return arr.GetOne[arr.ReleaseProfileResource](ctx, c.base, fmt.Sprintf("/api/v1/releaseprofile/%d", id))
}

// CreateReleaseProfile creates a new release profile.
func (c *Client) CreateReleaseProfile(ctx context.Context, profile *arr.ReleaseProfileResource) (*arr.ReleaseProfileResource, error) {
	return arr.PostOne[arr.ReleaseProfileResource](ctx, c.base, "/api/v1/releaseprofile", profile)
}

// UpdateReleaseProfile updates an existing release profile.
func (c *Client) UpdateReleaseProfile(ctx context.Context, profile *arr.ReleaseProfileResource) (*arr.ReleaseProfileResource, error) {
	return arr.PutOne[arr.ReleaseProfileResource](ctx, c.base, fmt.Sprintf("/api/v1/releaseprofile/%d", profile.ID), profile)
}

// DeleteReleaseProfile deletes a release profile by ID.
func (c *Client) DeleteReleaseProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/releaseprofile/%d", id), nil, nil)
}

// ---------- Remote Path Mappings ----------.

// SearchReleases searches for releases using the configured indexers.
func (c *Client) SearchReleases(ctx context.Context, bookID int) ([]arr.ReleaseResource, error) {
	path := fmt.Sprintf("/api/v1/release?bookId=%d", bookID)
	return arr.GetList[arr.ReleaseResource](ctx, c.base, path)
}

// GrabRelease grabs a release for download.
func (c *Client) GrabRelease(ctx context.Context, release *arr.ReleaseResource) (*arr.ReleaseResource, error) {
	return arr.PostOne[arr.ReleaseResource](ctx, c.base, "/api/v1/release", release)
}

// PushRelease pushes a release to the download client.
func (c *Client) PushRelease(ctx context.Context, release *arr.ReleasePushResource) ([]arr.ReleaseResource, error) {
	return arr.PostList[arr.ReleaseResource](ctx, c.base, "/api/v1/release/push", release)
}

// ---------- Rename ----------.
