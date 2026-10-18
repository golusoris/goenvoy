package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRootFolders returns all configured root folders.
func (c *Client) GetRootFolders(ctx context.Context) ([]arr.RootFolder, error) {
	return arr.GetList[arr.RootFolder](ctx, c.base, "/api/v1/rootfolder")
}

// GetRootFolder returns a single root folder by its ID.
func (c *Client) GetRootFolder(ctx context.Context, id int) (*arr.RootFolder, error) {
	return arr.GetOne[arr.RootFolder](ctx, c.base, fmt.Sprintf("/api/v1/rootfolder/%d", id))
}

// CreateRootFolder creates a new root folder.
func (c *Client) CreateRootFolder(ctx context.Context, folder *arr.RootFolder) (*arr.RootFolder, error) {
	return arr.PostOne[arr.RootFolder](ctx, c.base, "/api/v1/rootfolder", folder)
}

// UpdateRootFolder updates an existing root folder.
func (c *Client) UpdateRootFolder(ctx context.Context, folder *arr.RootFolder) (*arr.RootFolder, error) {
	return arr.PutOne[arr.RootFolder](ctx, c.base, fmt.Sprintf("/api/v1/rootfolder/%d", folder.ID), folder)
}

// DeleteRootFolder deletes a root folder by ID.
func (c *Client) DeleteRootFolder(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/rootfolder/%d", id), nil, nil)
}

// ---------- Custom Filters ----------.
