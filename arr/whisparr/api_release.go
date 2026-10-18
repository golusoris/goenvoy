package whisparr

import (
	"context"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GrabRelease sends a release to the download client.
func (c *api) GrabRelease(ctx context.Context, release *arr.ReleaseResource) (*arr.ReleaseResource, error) {
	var out arr.ReleaseResource
	if err := c.base.Post(ctx, "/api/v3/release", release, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PushRelease pushes a release for processing.
func (c *api) PushRelease(ctx context.Context, push *arr.ReleasePushResource) ([]arr.ReleaseResource, error) {
	var out []arr.ReleaseResource
	if err := c.base.Post(ctx, "/api/v3/release/push", push, &out); err != nil {
		return nil, err
	}
	return out, nil
}
