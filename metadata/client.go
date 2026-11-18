package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"
)

const (
	// DefaultTimeout is the default HTTP request timeout used by all providers.
	DefaultTimeout = 30 * time.Second
	// DefaultUserAgent is the default User-Agent header for all providers.
	DefaultUserAgent = "goenvoy/0.0.1"
	// DefaultMaxResponseBytes is the default response body limit of 32 MiB.
	DefaultMaxResponseBytes int64 = 32 << 20
)

// ErrResponseTooLarge indicates that a response body exceeds the configured limit.
var ErrResponseTooLarge = errors.New("response too large")

// AuthFunc is called before each HTTP request to apply provider-specific
// authentication (e.g. setting headers or query parameters).
type AuthFunc func(req *http.Request)

// Option configures a [BaseClient].
type Option func(*BaseClient)

// WithHTTPClient sets a custom [http.Client].
func WithHTTPClient(c *http.Client) Option {
	return func(b *BaseClient) { b.httpClient = c }
}

// WithTimeout overrides the default HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(b *BaseClient) { b.httpClient.Timeout = d }
}

// WithUserAgent sets the User-Agent header for all requests.
func WithUserAgent(ua string) Option {
	return func(b *BaseClient) { b.userAgent = ua }
}

// WithBaseURL overrides the provider's default API base URL.
func WithBaseURL(u string) Option {
	return func(b *BaseClient) { b.baseURL = u }
}

// WithMaxResponseBytes sets the response body limit; nonpositive values keep the default.
func WithMaxResponseBytes(n int64) Option {
	return func(b *BaseClient) {
		if n > 0 {
			b.maxResponseBytes = n
		}
	}
}

// BaseClient is a shared low-level HTTP client for metadata providers.
// Providers embed BaseClient and use its methods to avoid duplicating
// HTTP plumbing. The [AuthFunc] hook is called on every request to
// apply provider-specific authentication.
type BaseClient struct {
	httpClient       *http.Client
	baseURL          string
	userAgent        string
	pkgName          string
	authFunc         AuthFunc
	maxResponseBytes int64
}

// NewBaseClient creates a [BaseClient] with the given default base URL and
// package name (used for error prefixes). Apply shared [Option] values to
// configure HTTP client, timeout, user agent, or base URL.
func NewBaseClient(defaultBaseURL, pkgName string, opts ...Option) *BaseClient {
	c := &BaseClient{
		httpClient:       &http.Client{Timeout: DefaultTimeout, Transport: sharedTransport},
		baseURL:          defaultBaseURL,
		userAgent:        DefaultUserAgent,
		pkgName:          pkgName,
		maxResponseBytes: DefaultMaxResponseBytes,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// SetAuth configures the authentication callback applied to every request.
func (c *BaseClient) SetAuth(fn AuthFunc) { c.authFunc = fn }

// BaseURL returns the configured base URL.
func (c *BaseClient) BaseURL() string { return c.baseURL }

// HTTPClient returns the underlying [http.Client].
func (c *BaseClient) HTTPClient() *http.Client { return c.httpClient }

// UserAgent returns the configured User-Agent string.
func (c *BaseClient) UserAgent() string { return c.userAgent }

// DoRaw executes an HTTP request to baseURL+path and returns the raw response
// body and status code. Auth, User-Agent, and Accept headers are applied
// automatically. Use this for providers that need custom error handling while
// still benefiting from shared HTTP plumbing.
func (c *BaseClient) DoRaw(ctx context.Context, method, path string, body io.Reader) (data []byte, statusCode int, err error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: create request: %w", c.pkgName, err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.authFunc != nil {
		c.authFunc(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %s %s: %w", c.pkgName, method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := c.readResponse(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return respBody, resp.StatusCode, nil
}

// DoRawURL is like [DoRaw] but accepts a fully-constructed URL instead of
// a path relative to baseURL. Use for providers with query-param auth or
// custom URL schemes.
func (c *BaseClient) DoRawURL(ctx context.Context, method, rawURL string, body io.Reader) (data []byte, statusCode int, err error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: create request: %w", c.pkgName, err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.authFunc != nil {
		c.authFunc(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %s: %w", c.pkgName, method, err)
	}
	defer resp.Body.Close()

	respBody, err := c.readResponse(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return respBody, resp.StatusCode, nil
}

// Get performs a GET request to baseURL+path, checks for a non-2xx status,
// and decodes the JSON response into dst.
func (c *BaseClient) Get(ctx context.Context, path string, dst any) error {
	return c.DoJSON(ctx, http.MethodGet, path, nil, dst)
}

// newRequest prepares an authorized HTTP request with an optional JSON payload.
func (c *BaseClient) newRequest(ctx context.Context, method, targetURL string, payload any) (*http.Request, error) {
	var bodyReader io.Reader = http.NoBody
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("%s: marshal body: %w", c.pkgName, err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", c.pkgName, err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.authFunc != nil {
		c.authFunc(req)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json;charset=utf-8")
	}
	return req, nil
}

// decodeResponse verifies HTTP status and unmarshals JSON response data.
func (c *BaseClient) decodeResponse(statusCode int, body []byte, dst any) error {
	if statusCode < 200 || statusCode >= 300 {
		return &APIError{
			StatusCode: statusCode,
			RawBody:    string(body),
			pkgName:    c.pkgName,
		}
	}

	if dst != nil && len(body) > 0 {
		if err := json.Unmarshal(body, dst); err != nil {
			return fmt.Errorf("%s: decode response: %w", c.pkgName, err)
		}
	}
	return nil
}

// DoJSON performs an HTTP request with an optional JSON payload. On success
// (2xx), it decodes the response into dst. On failure, it returns an [*APIError].
func (c *BaseClient) DoJSON(ctx context.Context, method, path string, payload, dst any) error {
	req, err := c.newRequest(ctx, method, c.baseURL+path, payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %s %s: %w", c.pkgName, method, path, err)
	}
	defer resp.Body.Close()

	body, err := c.readResponse(resp.Body)
	if err != nil {
		return err
	}

	return c.decodeResponse(resp.StatusCode, body, dst)
}

// Why: bound response allocation and preserve one error contract for all request paths.
func (c *BaseClient) readResponse(body io.Reader) ([]byte, error) {
	readLimit := c.maxResponseBytes
	if readLimit < math.MaxInt64 {
		readLimit++
	}
	data, err := io.ReadAll(io.LimitReader(body, readLimit))
	if int64(len(data)) > c.maxResponseBytes {
		return nil, fmt.Errorf("%s: read response: %w (limit %d bytes)", c.pkgName, ErrResponseTooLarge, c.maxResponseBytes)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: read response: %w", c.pkgName, err)
	}
	return data, nil
}

// APIError is returned when a metadata provider API responds with a non-2xx
// HTTP status code.
type APIError struct {
	// StatusCode is the HTTP response status code.
	StatusCode int `json:"-"`
	// RawBody holds the raw response body.
	RawBody string `json:"-"`
	// pkgName is the provider package name for error formatting.
	pkgName string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.RawBody != "" {
		body := e.RawBody
		if len(body) > 512 {
			end := 512
			for end > 0 && !utf8.RuneStart(body[end]) {
				end--
			}
			body = body[:end] + "... (truncated)"
		}
		return e.pkgName + ": HTTP " + strconv.Itoa(e.StatusCode) + ": " + body
	}
	return e.pkgName + ": HTTP " + strconv.Itoa(e.StatusCode)
}
