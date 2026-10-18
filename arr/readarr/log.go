package readarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetLogs returns log entries with pagination.
func (c *Client) GetLogs(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.LogRecord], error) {
	path := fmt.Sprintf("/api/v1/log?page=%d&pageSize=%d", page, pageSize)
	return arr.GetOne[arr.PagingResource[arr.LogRecord]](ctx, c.base, path)
}

// GetLogFiles returns a list of log files.
func (c *Client) GetLogFiles(ctx context.Context) ([]arr.LogFileResource, error) {
	return arr.GetList[arr.LogFileResource](ctx, c.base, "/api/v1/log/file")
}

// GetLogFileContent returns the content of a specific log file by filename.
func (c *Client) GetLogFileContent(ctx context.Context, filename string) (string, error) {
	path := "/api/v1/log/file/" + url.PathEscape(filename)
	b, err := c.base.GetRaw(ctx, path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetUpdateLogFiles returns a list of update log files.
func (c *Client) GetUpdateLogFiles(ctx context.Context) ([]arr.LogFileResource, error) {
	return arr.GetList[arr.LogFileResource](ctx, c.base, "/api/v1/log/file/update")
}

// GetUpdateLogFileContent returns the content of a specific update log file.
func (c *Client) GetUpdateLogFileContent(ctx context.Context, filename string) (string, error) {
	path := "/api/v1/log/file/update/" + url.PathEscape(filename)
	b, err := c.base.GetRaw(ctx, path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ---------- System ----------.
