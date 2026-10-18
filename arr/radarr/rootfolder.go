package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRootFolders returns all configured root folders.
func (c *Client) GetRootFolders(ctx context.Context) ([]arr.RootFolder, error) {
	return arr.GetList[arr.RootFolder](ctx, c.base, "/api/v3/rootfolder")
}

// GetRootFolder returns a single root folder by ID.
func (c *Client) GetRootFolder(ctx context.Context, id int) (*arr.RootFolder, error) {
	return arr.GetOne[arr.RootFolder](ctx, c.base, fmt.Sprintf("/api/v3/rootfolder/%d", id))
}

// CreateRootFolder adds a new root folder.
func (c *Client) CreateRootFolder(ctx context.Context, rf *arr.RootFolder) (*arr.RootFolder, error) {
	return arr.PostOne[arr.RootFolder](ctx, c.base, "/api/v3/rootfolder", rf)
}

// DeleteRootFolder removes a root folder.
func (c *Client) DeleteRootFolder(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/rootfolder/%d", id), nil, nil)
}

// ---------- Release Profiles ----------.

// UpdateRootFolder updates an existing root folder.
func (c *Client) UpdateRootFolder(ctx context.Context, rf *arr.RootFolder) (*arr.RootFolder, error) {
	var out arr.RootFolder
	path := fmt.Sprintf("/api/v3/rootfolder/%d", rf.ID)
	if err := c.base.Put(ctx, path, rf, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- FileSystem ----------.
