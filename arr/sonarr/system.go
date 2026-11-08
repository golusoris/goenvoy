package sonarr

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetSystemStatus returns Sonarr system information.
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

// RestoreBackup restores the backup with the given ID through the same Sonarr code path as an
// UploadBackup of a .zip archive, so the UploadBackup notes on what Sonarr overwrites, stages and leaves
// behind apply. Sonarr does not restart itself, so call Restart afterwards.
func (c *Client) RestoreBackup(ctx context.Context, id int) error {
	return c.base.Post(ctx, fmt.Sprintf("/api/v3/system/backup/restore/%d", id), nil, nil)
}

// ---------- Quality Profiles (CRUD) ----------.

// GetSystemRoutes returns all registered API routes.
func (c *Client) GetSystemRoutes(ctx context.Context) ([]arr.SystemRouteResource, error) {
	return arr.GetList[arr.SystemRouteResource](ctx, c.base, "/api/v3/system/routes")
}

// GetSystemRoutesDuplicate returns duplicate API routes.
func (c *Client) GetSystemRoutesDuplicate(ctx context.Context) ([]arr.SystemRouteResource, error) {
	return arr.GetList[arr.SystemRouteResource](ctx, c.base, "/api/v3/system/routes/duplicate")
}

// Shutdown sends a shutdown command to Sonarr.
func (c *Client) Shutdown(ctx context.Context) error {
	return c.base.Post(ctx, "/api/v3/system/shutdown", nil, nil)
}

// Restart sends a restart command to Sonarr.
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

// Ping checks connectivity to the Sonarr instance.
func (c *Client) Ping(ctx context.Context) error {
	return c.base.Get(ctx, "/ping", nil)
}

// ---------- Series Folder ----------.

// HeadPing performs a lightweight HEAD request to /ping.
func (c *Client) HeadPing(ctx context.Context) error {
	return c.base.Head(ctx, "/ping")
}

// ---------- Backup Upload ----------.

// UploadBackup uploads a backup archive via multipart form POST and restores it.
// Sonarr picks the restore path from the extension of fileName. A .zip archive is extracted into one
// fixed temporary folder, which Sonarr does not clear first; Sonarr then takes from that folder only
// Config.xml, which overwrites config.xml at once, and sonarr.db or nzbdrone.db, which it stages for its
// next start. Any of these files left in that folder by a failed restore is applied by the next restore
// too. An archive without a database, such as a backup of a PostgreSQL instance, stages none of its own,
// so a database staged by an earlier restore stays pending. A .db or .xml file name skips extraction:
// Sonarr stages the uploaded file itself as the database, so upload config.xml only inside a .zip archive.
// Sonarr does not restart itself, so call Restart afterwards. No upload-only endpoint exists: authorize
// callers for restore, not archive storage. A failed or canceled call can leave the remote outcome unknown
// and the request still running on Sonarr; every upload shares one temporary path there, so do not retry
// while the earlier attempt may still be running.
func (c *Client) UploadBackup(ctx context.Context, fileName string, data io.Reader) error {
	return c.base.Upload(ctx, "/api/v3/system/backup/restore/upload", "file", fileName, data)
}
