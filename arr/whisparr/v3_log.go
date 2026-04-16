package whisparr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetLogs returns log entries (paged).
func (c *ClientV3) GetLogs(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.LogRecord], error) {
	var out arr.PagingResource[arr.LogRecord]
	path := fmt.Sprintf("/api/v3/log?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetLogFiles returns available log files.
func (c *ClientV3) GetLogFiles(ctx context.Context) ([]arr.LogFileResource, error) {
	var out []arr.LogFileResource
	if err := c.base.Get(ctx, "/api/v3/log/file", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetLogFileContent returns the content of a specific log file.
func (c *ClientV3) GetLogFileContent(ctx context.Context, filename string) (string, error) {
	b, err := c.base.GetRaw(ctx, "/api/v3/log/file/"+url.PathEscape(filename))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetUpdateLogFiles returns available update log files.
func (c *ClientV3) GetUpdateLogFiles(ctx context.Context) ([]arr.LogFileResource, error) {
	var out []arr.LogFileResource
	if err := c.base.Get(ctx, "/api/v3/log/file/update", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetUpdateLogFileContent returns the content of a specific update log file.
func (c *ClientV3) GetUpdateLogFileContent(ctx context.Context, filename string) (string, error) {
	b, err := c.base.GetRaw(ctx, "/api/v3/log/file/update/"+url.PathEscape(filename))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ---------- Manual Import ----------.
