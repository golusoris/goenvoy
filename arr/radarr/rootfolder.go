package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRootFolders returns all configured root folders.
func (c *Client) GetRootFolders(ctx context.Context) ([]arr.RootFolder, error) {
	var out []arr.RootFolder
	if err := c.base.Get(ctx, "/api/v3/rootfolder", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetRootFolder returns a single root folder by ID.
func (c *Client) GetRootFolder(ctx context.Context, id int) (*arr.RootFolder, error) {
	var out arr.RootFolder
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/rootfolder/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateRootFolder adds a new root folder.
func (c *Client) CreateRootFolder(ctx context.Context, rf *arr.RootFolder) (*arr.RootFolder, error) {
	var out arr.RootFolder
	if err := c.base.Post(ctx, "/api/v3/rootfolder", rf, &out); err != nil {
		return nil, err
	}
	return &out, nil
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
