package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetReleaseProfiles returns all release profiles.
func (c *Client) GetReleaseProfiles(ctx context.Context) ([]arr.ReleaseProfileResource, error) {
	var out []arr.ReleaseProfileResource
	if err := c.base.Get(ctx, "/api/v1/releaseprofile", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetReleaseProfile returns a single release profile by its ID.
func (c *Client) GetReleaseProfile(ctx context.Context, id int) (*arr.ReleaseProfileResource, error) {
	var out arr.ReleaseProfileResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/releaseprofile/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateReleaseProfile creates a new release profile.
func (c *Client) CreateReleaseProfile(ctx context.Context, profile *arr.ReleaseProfileResource) (*arr.ReleaseProfileResource, error) {
	var out arr.ReleaseProfileResource
	if err := c.base.Post(ctx, "/api/v1/releaseprofile", profile, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateReleaseProfile updates an existing release profile.
func (c *Client) UpdateReleaseProfile(ctx context.Context, profile *arr.ReleaseProfileResource) (*arr.ReleaseProfileResource, error) {
	var out arr.ReleaseProfileResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/releaseprofile/%d", profile.ID), profile, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteReleaseProfile deletes a release profile by ID.
func (c *Client) DeleteReleaseProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/releaseprofile/%d", id), nil, nil)
}

// ---------- Remote Path Mappings ----------.

// SearchReleases searches for releases using the configured indexers.
func (c *Client) SearchReleases(ctx context.Context, bookID int) ([]arr.ReleaseResource, error) {
	var out []arr.ReleaseResource
	path := fmt.Sprintf("/api/v1/release?bookId=%d", bookID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GrabRelease grabs a release for download.
func (c *Client) GrabRelease(ctx context.Context, release *arr.ReleaseResource) (*arr.ReleaseResource, error) {
	var out arr.ReleaseResource
	if err := c.base.Post(ctx, "/api/v1/release", release, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PushRelease pushes a release to the download client.
func (c *Client) PushRelease(ctx context.Context, release *arr.ReleasePushResource) ([]arr.ReleaseResource, error) {
	var out []arr.ReleaseResource
	if err := c.base.Post(ctx, "/api/v1/release/push", release, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Rename ----------.
