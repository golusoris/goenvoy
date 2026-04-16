package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetReleaseProfiles returns all release profiles.
func (c *Client) GetReleaseProfiles(ctx context.Context) ([]arr.ReleaseProfileResource, error) {
	var out []arr.ReleaseProfileResource
	if err := c.base.Get(ctx, "/api/v3/releaseprofile", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetReleaseProfile returns a single release profile by ID.
func (c *Client) GetReleaseProfile(ctx context.Context, id int) (*arr.ReleaseProfileResource, error) {
	var out arr.ReleaseProfileResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/releaseprofile/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateReleaseProfile creates a new release profile.
func (c *Client) CreateReleaseProfile(ctx context.Context, rp *arr.ReleaseProfileResource) (*arr.ReleaseProfileResource, error) {
	var out arr.ReleaseProfileResource
	if err := c.base.Post(ctx, "/api/v3/releaseprofile", rp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateReleaseProfile updates an existing release profile.
func (c *Client) UpdateReleaseProfile(ctx context.Context, rp *arr.ReleaseProfileResource) (*arr.ReleaseProfileResource, error) {
	var out arr.ReleaseProfileResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/releaseprofile/%d", rp.ID), rp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteReleaseProfile removes a release profile.
func (c *Client) DeleteReleaseProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/releaseprofile/%d", id), nil, nil)
}

// ---------- Remote Path Mappings ----------.

// SearchReleases searches for releases matching the given movie ID.
func (c *Client) SearchReleases(ctx context.Context, movieID int) ([]arr.ReleaseResource, error) {
	var out []arr.ReleaseResource
	path := fmt.Sprintf("/api/v3/release?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
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
