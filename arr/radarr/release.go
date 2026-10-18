package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetReleaseProfiles returns all release profiles.
func (c *Client) GetReleaseProfiles(ctx context.Context) ([]arr.ReleaseProfileResource, error) {
	return arr.GetList[arr.ReleaseProfileResource](ctx, c.base, "/api/v3/releaseprofile")
}

// GetReleaseProfile returns a single release profile by ID.
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

// DeleteReleaseProfile removes a release profile.
func (c *Client) DeleteReleaseProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/releaseprofile/%d", id), nil, nil)
}

// ---------- Remote Path Mappings ----------.

// SearchReleases searches for releases matching the given movie ID.
func (c *Client) SearchReleases(ctx context.Context, movieID int) ([]arr.ReleaseResource, error) {
	path := fmt.Sprintf("/api/v3/release?movieId=%d", movieID)
	return arr.GetList[arr.ReleaseResource](ctx, c.base, path)
}

// PushRelease manually pushes a download for a release.
func (c *Client) PushRelease(ctx context.Context, release *arr.ReleasePushResource) error {
	return c.base.Post(ctx, "/api/v3/release/push", release, nil)
}

// GrabRelease grabs a release by its GUID.
func (c *Client) GrabRelease(ctx context.Context, guid string, indexerID int) error {
	body := map[string]any{"guid": guid, "indexerId": indexerID}
	return c.base.Post(ctx, "/api/v3/release", body, nil)
}

// ---------- Rename ----------.
