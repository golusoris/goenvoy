package arr

import (
	"context"
	"net/http"
)

// GetOne performs an authenticated GET request and decodes the JSON response into a new T.
func GetOne[T any](ctx context.Context, c *BaseClient, path string) (*T, error) {
	return one[T](ctx, c, http.MethodGet, path, nil)
}

// GetList performs an authenticated GET request and decodes the JSON array response into a []T.
func GetList[T any](ctx context.Context, c *BaseClient, path string) ([]T, error) {
	return list[T](ctx, c, http.MethodGet, path, nil)
}

// PostOne performs an authenticated POST request with a JSON body and decodes the response into a new T.
func PostOne[T any](ctx context.Context, c *BaseClient, path string, body any) (*T, error) {
	return one[T](ctx, c, http.MethodPost, path, body)
}

// PostList performs an authenticated POST request with a JSON body and decodes the array response into a []T.
func PostList[T any](ctx context.Context, c *BaseClient, path string, body any) ([]T, error) {
	return list[T](ctx, c, http.MethodPost, path, body)
}

// PutOne performs an authenticated PUT request with a JSON body and decodes the response into a new T.
func PutOne[T any](ctx context.Context, c *BaseClient, path string, body any) (*T, error) {
	return one[T](ctx, c, http.MethodPut, path, body)
}

// PutList performs an authenticated PUT request with a JSON body and decodes the array response into a []T.
func PutList[T any](ctx context.Context, c *BaseClient, path string, body any) ([]T, error) {
	return list[T](ctx, c, http.MethodPut, path, body)
}

func one[T any](ctx context.Context, c *BaseClient, method, path string, body any) (*T, error) {
	var out T
	if err := c.do(ctx, method, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func list[T any](ctx context.Context, c *BaseClient, method, path string, body any) ([]T, error) {
	var out []T
	if err := c.do(ctx, method, path, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetValue performs an authenticated GET request and decodes the JSON response into a T value.
// Use it for endpoints that answer with a bare scalar, such as a JSON string.
func GetValue[T any](ctx context.Context, c *BaseClient, path string) (T, error) {
	var out T
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		var zero T
		return zero, err
	}
	return out, nil
}

// GetText performs an authenticated GET request and returns the response body as a string.
// Use it for endpoints that return plain text, such as log files.
func (c *BaseClient) GetText(ctx context.Context, path string) (string, error) {
	b, err := c.GetRaw(ctx, path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
