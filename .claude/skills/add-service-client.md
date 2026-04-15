---
name: add-service-client
description: Scaffold a new pure-stdlib Go API-client module under a given category.
---

# Skill — `/add-service-client`

Scaffold a new service-client module from a one-line prompt.

## When to use

The user says "add a client for <service>" and you need to create an entire new module (directory, `go.mod`, doc/types/impl/test/example files, `AGENTS.md`, `docs/upstream/<service>.md`).

## Expected arguments

- `$1` — category, one of `arr | metadata/video | metadata/anime | metadata/music | metadata/tracking | metadata/book | metadata/game | metadata/adult | downloadclient | mediaserver | anime`.
- `$2` — service slug (kebab-case → used as both directory name and Go package name, e.g. `aria2`, `jellyseerr`).
- `$3` — upstream-API docs URL (pinned).
- (optional) `$4` — auth model: one of `apikey | basic | oauth-device | oauth-auth-code-pkce | jwt | none`. Default: `apikey`.

## Steps

1. Verify the target directory does not exist (`ls <category>/<service>` → must error).
2. Create `<category>/<service>/` and inside it:
   - `go.mod` — `module github.com/golusoris/goenvoy/<category>/<service>` · `go 1.26.1`. No `require` block (pure stdlib).
   - `doc.go` — package-level comment: one sentence, ends with a period.
   - `types.go` — placeholder struct for `<Service>Response` + `APIError`.
   - `<service>.go` — `New(baseURL, apiKey string, opts ...Option) (*Client, error)` + `Option` type + `WithHTTPClient` + `WithTimeout` + `WithHeader` + helper `do(ctx, method, path, body, out) error`.
   - `<service>_test.go` — table-driven `TestNew` + one HTTP method test using `httptest.NewServer`.
   - `example_test.go` — `func ExampleNew()` that shows the idiomatic construction.
   - `AGENTS.md` — auth model, pagination style, error shape, pinned upstream URL, last-verified date.
3. Append the new module to `go.work` (`use ./<category>/<service>`).
4. Add `docs/upstream/<category>-<service>.md` with URL + version + today's date + a one-paragraph "what this API does".
5. Add a `## Unreleased` stub under the module in the root `CHANGELOG.md`.
6. From the new module directory run: `go build ./... && go vet ./... && go test -race ./... && golangci-lint run --config ../../.golangci.yml ./...`.
7. Report to the user: module path, files created, next steps (typically: wire the upstream API's methods).

## Template — `<service>.go`

```go
// Package <service> is a pure-stdlib Go client for the <Service> API.
package <service>

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client talks to a <Service> server.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	headers    http.Header
}

// Option configures a Client in New.
type Option func(*Client)

// WithHTTPClient replaces the default *http.Client.
func WithHTTPClient(c *http.Client) Option {
	return func(x *Client) {
		if c != nil {
			x.httpClient = c
		}
	}
}

// WithTimeout sets the client request timeout.
func WithTimeout(d time.Duration) Option {
	return func(x *Client) { x.httpClient.Timeout = d }
}

// WithHeader adds a request header sent on every call.
func WithHeader(k, v string) Option {
	return func(x *Client) { x.headers.Set(k, v) }
}

// New constructs a Client. baseURL must be absolute http or https.
func New(baseURL, apiKey string, opts ...Option) (*Client, error) {
	u, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("<service>: parse baseURL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("<service>: baseURL scheme must be http or https, got %q", u.Scheme)
	}
	c := &Client{
		baseURL:    u,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		headers:    http.Header{},
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// APIError is returned when the <Service> API responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("<service>: HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("<service>: HTTP %d", e.StatusCode)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	u := *c.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + "/" + strings.TrimLeft(path, "/")
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("<service>: marshal body: %w", err)
		}
		rdr = strings.NewReader(string(b))
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), rdr)
	if err != nil {
		return fmt.Errorf("<service>: new request: %w", err)
	}
	for k, vs := range c.headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	// TODO: authentication scheme (api-key query, bearer, basic, etc.).
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("<service>: %s %s: %w", method, u.Path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("<service>: decode response: %w", err)
	}
	return nil
}
```

## Don't

- Don't add any `require` block to the new `go.mod`. goenvoy is pure stdlib.
- Don't copy-paste another service's auth handling blindly — check the upstream docs.
- Don't commit placeholder URLs in `docs/upstream/*.md`. Pin a real versioned URL.
