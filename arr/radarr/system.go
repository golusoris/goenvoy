package radarr

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetSystemStatus returns Radarr system information.
func (c *Client) GetSystemStatus(ctx context.Context) (*arr.StatusResponse, error) {
	return arr.GetOne[arr.StatusResponse](ctx, c.base, "/api/v3/system/status")
}

// GetHealth returns current health check results.
func (c *Client) GetHealth(ctx context.Context) ([]arr.HealthCheck, error) {
	return arr.GetList[arr.HealthCheck](ctx, c.base, "/api/v3/health")
}

// GetDiskSpace returns disk usage information for configured paths.
func (c *Client) GetDiskSpace(ctx context.Context) ([]arr.DiskSpace, error) {
	return arr.GetList[arr.DiskSpace](ctx, c.base, "/api/v3/diskspace")
}

// GetBackups returns all available backup files.
func (c *Client) GetBackups(ctx context.Context) ([]arr.Backup, error) {
	return arr.GetList[arr.Backup](ctx, c.base, "/api/v3/system/backup")
}

// DeleteBackup removes a backup file by ID.
func (c *Client) DeleteBackup(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/system/backup/%d", id), nil, nil)
}

// RestoreBackup triggers a restore from a backup by ID.
func (c *Client) RestoreBackup(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v3/system/backup/restore/%d", id), nil, nil)
}

// ---------- Quality Profiles (full CRUD) ----------.

// GetSystemRoutes returns all registered API routes.
func (c *Client) GetSystemRoutes(ctx context.Context) ([]arr.SystemRouteResource, error) {
	return arr.GetList[arr.SystemRouteResource](ctx, c.base, "/api/v3/system/routes")
}

// Shutdown sends a shutdown command to Radarr.
func (c *Client) Shutdown(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/system/shutdown", nil, nil)
}

// Restart sends a restart command to Radarr.
func (c *Client) Restart(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/system/restart", nil, nil)
}

// ---------- Tasks ----------.

// GetTasks returns all scheduled tasks.
func (c *Client) GetTasks(ctx context.Context) ([]arr.TaskResource, error) {
	return arr.GetList[arr.TaskResource](ctx, c.base, "/api/v3/system/task")
}

// GetTask returns a single scheduled task by ID.
func (c *Client) GetTask(ctx context.Context, id int) (*arr.TaskResource, error) {
	return arr.GetOne[arr.TaskResource](ctx, c.base, fmt.Sprintf("/api/v3/system/task/%d", id))
}

// ---------- Updates ----------.

// BrowseFileSystem returns directories and files at the given path.
func (c *Client) BrowseFileSystem(ctx context.Context, path string, includeFiles bool) (*FileSystemResource, error) {
	endpoint := fmt.Sprintf("/api/v3/filesystem?path=%s&includeFiles=%t", url.QueryEscape(path), includeFiles)
	return arr.GetOne[FileSystemResource](ctx, c.base, endpoint)
}

// GetFileSystemType returns the filesystem type (e.g. local, network) for a path.
func (c *Client) GetFileSystemType(ctx context.Context, path string) (string, error) {
	return arr.GetValue[string](ctx, c.base, "/api/v3/filesystem/type?path="+url.QueryEscape(path))
}

// GetFileSystemMediaFiles returns media files at the given path.
func (c *Client) GetFileSystemMediaFiles(ctx context.Context, path string) ([]string, error) {
	endpoint := "/api/v3/filesystem/mediafiles?path=" + url.QueryEscape(path)
	return arr.GetList[string](ctx, c.base, endpoint)
}

// ---------- Ping ----------.

// Ping checks connectivity to the Radarr instance.
func (c *Client) Ping(ctx context.Context) error {
	return c.base.Get(ctx, "/ping", nil)
}

// ---------- Alternative Titles ----------.

// GetSystemRoutesDuplicate returns duplicate API routes.
func (c *Client) GetSystemRoutesDuplicate(ctx context.Context) ([]arr.SystemRouteResource, error) {
	return arr.GetList[arr.SystemRouteResource](ctx, c.base, "/api/v3/system/routes/duplicate")
}

// ---------- Update Log File Content ----------.

// HeadPing performs a lightweight HEAD request to /ping.
func (c *Client) HeadPing(ctx context.Context) error {
	return c.base.Head(ctx, "/ping")
}

// ---------- Backup Upload ----------.

// UploadBackup uploads a backup file via multipart form POST.
func (c *Client) UploadBackup(ctx context.Context, fileName string, data io.Reader) error {
	return c.base.Upload(ctx, "/api/v3/system/backup/upload", "file", fileName, data)
}
