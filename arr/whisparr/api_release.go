package whisparr

import (
	"context"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GrabRelease sends a release to the download client.
func (c *api) GrabRelease(ctx context.Context, release *arr.ReleaseResource) (*arr.ReleaseResource, error) {
	return arr.PostOne[arr.ReleaseResource](ctx, c.base, "/api/v3/release", release)
}

// PushRelease pushes a release for processing.
func (c *api) PushRelease(ctx context.Context, push *arr.ReleasePushResource) ([]arr.ReleaseResource, error) {
	return arr.PostList[arr.ReleaseResource](ctx, c.base, "/api/v3/release/push", push)
}
