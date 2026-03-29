package arr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "goenvoy/0.0.1"
)

// Option configures a [BaseClient].
type Option func(*BaseClient)

// WithHTTPClient sets a custom [http.Client] for the [BaseClient].
func WithHTTPClient(c *http.Client) Option {
	return func(b *BaseClient) { b.httpClient = c }
}

// WithTimeout overrides the default HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(b *BaseClient) { b.httpClient.Timeout = d }
}

// WithUserAgent sets the User-Agent header sent with every request.
func WithUserAgent(ua string) Option {
	return func(b *BaseClient) { b.userAgent = ua }
}

// BaseClient is a low-level HTTP client shared by all *arr service clients.
// It handles authentication, JSON marshaling, and error wrapping.
type BaseClient struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	userAgent  string
}

// NewBaseClient creates a [BaseClient] targeting the given base URL.
// The apiKey is sent in the X-Api-Key header on every request.
func NewBaseClient(baseURL, apiKey string, opts ...Option) (*BaseClient, error) {
	baseURL = strings.TrimRight(baseURL, "/")

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("arr: invalid base URL %q: %w", baseURL, err)
	}

	c := &BaseClient{
		baseURL:    u,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout},
		userAgent:  defaultUserAgent,
	}

	for _, o := range opts {
		o(c)
	}

	return c, nil
}

// Get performs an authenticated GET request and decodes the JSON response into dst.
func (c *BaseClient) Get(ctx context.Context, path string, dst any) error {
	return c.do(ctx, http.MethodGet, path, nil, dst)
}

// Post performs an authenticated POST request with a JSON body and decodes the response into dst.
func (c *BaseClient) Post(ctx context.Context, path string, body, dst any) error {
	return c.do(ctx, http.MethodPost, path, body, dst)
}

// Put performs an authenticated PUT request with a JSON body and decodes the response into dst.
func (c *BaseClient) Put(ctx context.Context, path string, body, dst any) error {
	return c.do(ctx, http.MethodPut, path, body, dst)
}

// Delete performs an authenticated DELETE request and decodes the response into dst.
func (c *BaseClient) Delete(ctx context.Context, path string, dst any) error {
	return c.do(ctx, http.MethodDelete, path, nil, dst)
}

// do is the internal method that executes every HTTP request.
func (c *BaseClient) do(ctx context.Context, method, path string, body, dst any) error {
	u := c.baseURL.JoinPath(path)

	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("arr: marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return fmt.Errorf("arr: create request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("arr: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("arr: read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Body:       respBody,
		}
	}

	if dst != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return fmt.Errorf("arr: decode response: %w", err)
		}
	}

	return nil
}
