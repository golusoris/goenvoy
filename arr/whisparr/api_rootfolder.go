package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetRootFolders returns all configured root folders.
func (c *api) GetRootFolders(ctx context.Context) ([]arr.RootFolder, error) {
	return arr.GetList[arr.RootFolder](ctx, c.base, "/api/v3/rootfolder")
}

// GetRootFolder returns a root folder by ID.
func (c *api) GetRootFolder(ctx context.Context, id int) (*arr.RootFolder, error) {
	return arr.GetOne[arr.RootFolder](ctx, c.base, fmt.Sprintf("/api/v3/rootfolder/%d", id))
}

// CreateRootFolder creates a new root folder.
func (c *api) CreateRootFolder(ctx context.Context, rf *arr.RootFolder) (*arr.RootFolder, error) {
	return arr.PostOne[arr.RootFolder](ctx, c.base, "/api/v3/rootfolder", rf)
}

// DeleteRootFolder deletes a root folder by ID.
func (c *api) DeleteRootFolder(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/rootfolder/%d", id), nil, nil)
}
