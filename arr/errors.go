package arr

import (
	"fmt"
)

// APIError represents a non-2xx HTTP response from an *arr API.
type APIError struct {
	// StatusCode is the HTTP status code returned by the server.
	StatusCode int
	// Method is the HTTP method of the failed request.
	Method string
	// Path is the request path that produced the error.
	Path string
	// Body is the raw response body, useful for debugging.
	Body []byte
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("arr: %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, string(e.Body))
}
