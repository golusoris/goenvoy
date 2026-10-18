package whisparr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetLogs returns log entries (paged).
func (c *api) GetLogs(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.LogRecord], error) {
	var out arr.PagingResource[arr.LogRecord]
	path := fmt.Sprintf("/api/v3/log?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetLogFiles returns available log files.
func (c *api) GetLogFiles(ctx context.Context) ([]arr.LogFileResource, error) {
	return arr.GetList[arr.LogFileResource](ctx, c.base, "/api/v3/log/file")
}

// GetLogFileContent returns the content of a specific log file.
func (c *api) GetLogFileContent(ctx context.Context, filename string) (string, error) {
	b, err := c.base.GetRaw(ctx, "/api/v3/log/file/"+url.PathEscape(filename))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetUpdateLogFiles returns available update log files.
func (c *api) GetUpdateLogFiles(ctx context.Context) ([]arr.LogFileResource, error) {
	return arr.GetList[arr.LogFileResource](ctx, c.base, "/api/v3/log/file/update")
}

// GetUpdateLogFileContent returns the content of a specific update log file.
func (c *api) GetUpdateLogFileContent(ctx context.Context, filename string) (string, error) {
	b, err := c.base.GetRaw(ctx, "/api/v3/log/file/update/"+url.PathEscape(filename))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
