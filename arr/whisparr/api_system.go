package whisparr

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetSystemStatus returns system information.
func (c *api) GetSystemStatus(ctx context.Context) (*arr.StatusResponse, error) {
	var out arr.StatusResponse
	if err := c.base.Get(ctx, "/api/v3/system/status", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHealth returns a list of health check results.
func (c *api) GetHealth(ctx context.Context) ([]arr.HealthCheck, error) {
	var out []arr.HealthCheck
	if err := c.base.Get(ctx, "/api/v3/health", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetDiskSpace returns disk space information for all root folders.
func (c *api) GetDiskSpace(ctx context.Context) ([]arr.DiskSpace, error) {
	var out []arr.DiskSpace
	if err := c.base.Get(ctx, "/api/v3/diskspace", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBackups returns all available backups.
func (c *api) GetBackups(ctx context.Context) ([]arr.Backup, error) {
	var out []arr.Backup
	if err := c.base.Get(ctx, "/api/v3/system/backup", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteBackup deletes a backup by ID.
func (c *api) DeleteBackup(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/system/backup/%d", id), nil, nil)
}

// RestoreBackup restores from a backup by ID.
func (c *api) RestoreBackup(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v3/system/backup/restore/%d", id), nil, nil)
}

// BrowseFileSystem returns a directory listing.
func (c *api) BrowseFileSystem(ctx context.Context, path string) (map[string]any, error) {
	var out map[string]any
	reqPath := "/api/v3/filesystem?path=" + url.QueryEscape(path)
	if err := c.base.Get(ctx, reqPath, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetFileSystemType returns the filesystem type for a path.
func (c *api) GetFileSystemType(ctx context.Context, path string) (string, error) {
	var out string
	reqPath := "/api/v3/filesystem/type?path=" + url.QueryEscape(path)
	if err := c.base.Get(ctx, reqPath, &out); err != nil {
		return "", err
	}
	return out, nil
}

// GetFileSystemMediaFiles returns media files in a path.
func (c *api) GetFileSystemMediaFiles(ctx context.Context, path string) ([]map[string]any, error) {
	var out []map[string]any
	reqPath := "/api/v3/filesystem/mediafiles?path=" + url.QueryEscape(path)
	if err := c.base.Get(ctx, reqPath, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSystemRoutes returns all API routes.
func (c *api) GetSystemRoutes(ctx context.Context) ([]arr.SystemRouteResource, error) {
	var out []arr.SystemRouteResource
	if err := c.base.Get(ctx, "/api/v3/system/routes", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSystemRoutesDuplicate returns duplicate API routes.
func (c *api) GetSystemRoutesDuplicate(ctx context.Context) ([]arr.SystemRouteResource, error) {
	var out []arr.SystemRouteResource
	if err := c.base.Get(ctx, "/api/v3/system/routes/duplicate", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Shutdown sends a shutdown command.
func (c *api) Shutdown(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/system/shutdown", nil, nil)
}

// Restart sends a restart command.
func (c *api) Restart(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/system/restart", nil, nil)
}

// GetTasks returns all scheduled tasks.
func (c *api) GetTasks(ctx context.Context) ([]arr.TaskResource, error) {
	var out []arr.TaskResource
	if err := c.base.Get(ctx, "/api/v3/system/task", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTask returns a scheduled task by ID.
func (c *api) GetTask(ctx context.Context, id int) (*arr.TaskResource, error) {
	var out arr.TaskResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/system/task/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Ping checks if the Whisparr instance is reachable.
func (c *api) Ping(ctx context.Context) error {
	return c.base.Get(ctx, "/ping", nil)
}

// HeadPing performs a lightweight HEAD request to /ping.
func (c *api) HeadPing(ctx context.Context) error {
	return c.base.Head(ctx, "/ping")
}

// UploadBackup uploads a backup file via multipart form POST.
func (c *api) UploadBackup(ctx context.Context, fileName string, data io.Reader) error {
	return c.base.Upload(ctx, "/api/v3/system/backup/upload", "file", fileName, data)
}
