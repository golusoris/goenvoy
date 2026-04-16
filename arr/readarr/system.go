package readarr

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetSystemStatus returns Readarr system information.
func (c *Client) GetSystemStatus(ctx context.Context) (*arr.StatusResponse, error) {
	var out arr.StatusResponse
	if err := c.base.Get(ctx, "/api/v1/system/status", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHealth returns current health check results.
func (c *Client) GetHealth(ctx context.Context) ([]arr.HealthCheck, error) {
	var out []arr.HealthCheck
	if err := c.base.Get(ctx, "/api/v1/health", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetDiskSpace returns disk usage information for configured paths.
func (c *Client) GetDiskSpace(ctx context.Context) ([]arr.DiskSpace, error) {
	var out []arr.DiskSpace
	if err := c.base.Get(ctx, "/api/v1/diskspace", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBackups returns a list of all available backups.
func (c *Client) GetBackups(ctx context.Context) ([]arr.Backup, error) {
	var out []arr.Backup
	if err := c.base.Get(ctx, "/api/v1/system/backup", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteBackup deletes a backup by ID.
func (c *Client) DeleteBackup(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/system/backup/%d", id), nil, nil)
}

// RestoreBackup triggers a restore from a backup by ID.
func (c *Client) RestoreBackup(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v1/system/backup/restore/%d", id), nil, nil)
}

// ---------- Logs ----------.

// GetTasks returns all scheduled tasks.
func (c *Client) GetTasks(ctx context.Context) ([]arr.TaskResource, error) {
	var out []arr.TaskResource
	if err := c.base.Get(ctx, "/api/v1/system/task", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTask returns a single task by its ID.
func (c *Client) GetTask(ctx context.Context, id int) (*arr.TaskResource, error) {
	var out arr.TaskResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/system/task/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSystemRoutes returns all API routes.
func (c *Client) GetSystemRoutes(ctx context.Context) ([]arr.SystemRouteResource, error) {
	var out []arr.SystemRouteResource
	if err := c.base.Get(ctx, "/api/v1/system/routes", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSystemRoutesDuplicate returns duplicate API routes.
func (c *Client) GetSystemRoutesDuplicate(ctx context.Context) ([]arr.SystemRouteResource, error) {
	var out []arr.SystemRouteResource
	if err := c.base.Get(ctx, "/api/v1/system/routes/duplicate", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Shutdown sends a shutdown command to Readarr.
func (c *Client) Shutdown(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v1/system/shutdown", nil, nil)
}

// Restart sends a restart command to Readarr.
func (c *Client) Restart(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v1/system/restart", nil, nil)
}

// Ping checks connectivity to the Readarr instance.
func (c *Client) Ping(ctx context.Context) error {
	return c.base.Get(ctx, "/ping", nil)
}

// ---------- Indexer Flags ----------.

// BrowseFileSystem returns directory/file listings for the given path.
func (c *Client) BrowseFileSystem(ctx context.Context, path string) (*FileSystemResource, error) {
	var out FileSystemResource
	reqPath := "/api/v1/filesystem?path=" + url.QueryEscape(path)
	if err := c.base.Get(ctx, reqPath, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetFileSystemType returns the filesystem type for the given path.
func (c *Client) GetFileSystemType(ctx context.Context, path string) (string, error) {
	var out string
	reqPath := "/api/v1/filesystem/type?path=" + url.QueryEscape(path)
	if err := c.base.Get(ctx, reqPath, &out); err != nil {
		return "", err
	}
	return out, nil
}

// GetFileSystemMediaFiles returns media files for the given path.
func (c *Client) GetFileSystemMediaFiles(ctx context.Context, path string) ([]FileSystemEntry, error) {
	var out []FileSystemEntry
	reqPath := "/api/v1/filesystem/mediafiles?path=" + url.QueryEscape(path)
	if err := c.base.Get(ctx, reqPath, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Book File Update ----------.

// HeadPing performs a lightweight HEAD request to /ping.
func (c *Client) HeadPing(ctx context.Context) error {
	return c.base.Head(ctx, "/ping")
}

// ---------- Backup Upload ----------.

// UploadBackup uploads a backup file via multipart form POST.
func (c *Client) UploadBackup(ctx context.Context, fileName string, data io.Reader) error {
	return c.base.Upload(ctx, "/api/v1/system/backup/upload", "file", fileName, data)
}
