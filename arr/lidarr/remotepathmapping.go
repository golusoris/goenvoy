package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRemotePathMappings returns all remote path mappings.
func (c *Client) GetRemotePathMappings(ctx context.Context) ([]arr.RemotePathMappingResource, error) {
	return arr.GetList[arr.RemotePathMappingResource](ctx, c.base, "/api/v1/remotepathmapping")
}

// GetRemotePathMapping returns a single remote path mapping by ID.
func (c *Client) GetRemotePathMapping(ctx context.Context, id int) (*arr.RemotePathMappingResource, error) {
	return arr.GetOne[arr.RemotePathMappingResource](ctx, c.base, fmt.Sprintf("/api/v1/remotepathmapping/%d", id))
}

// CreateRemotePathMapping creates a new remote path mapping.
func (c *Client) CreateRemotePathMapping(ctx context.Context, m *arr.RemotePathMappingResource) (*arr.RemotePathMappingResource, error) {
	return arr.PostOne[arr.RemotePathMappingResource](ctx, c.base, "/api/v1/remotepathmapping", m)
}

// UpdateRemotePathMapping updates an existing remote path mapping.
func (c *Client) UpdateRemotePathMapping(ctx context.Context, m *arr.RemotePathMappingResource) (*arr.RemotePathMappingResource, error) {
	return arr.PutOne[arr.RemotePathMappingResource](ctx, c.base, fmt.Sprintf("/api/v1/remotepathmapping/%d", m.ID), m)
}

// DeleteRemotePathMapping deletes a remote path mapping by ID.
func (c *Client) DeleteRemotePathMapping(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/remotepathmapping/%d", id), nil, nil)
}

// ---------- Auto Tagging ----------.
