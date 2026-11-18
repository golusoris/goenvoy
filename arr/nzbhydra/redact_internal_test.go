package nzbhydra

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

type redactionTransport func(*http.Request) (*http.Response, error)

// RoundTrip returns the configured test response or error.
func (f redactionTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type redactionFailure struct{ error }

// Unwrap exposes the test sentinel.
func (e redactionFailure) Unwrap() error { return e.error }

// Timeout reports a timeout for the test transport.
func (redactionFailure) Timeout() bool { return true }

// Read returns the configured body read error.
func (e redactionFailure) Read([]byte) (int, error) { return 0, e.error }

// Close releases the test body.
func (redactionFailure) Close() error { return nil }

// TestRequestRedaction checks key redaction across request and response failures.
func TestRequestRedaction(t *testing.T) {
	t.Parallel()
	const key = "placeholder+k/y=1"
	sentinel := errors.New("transport sentinel")
	leak := fmt.Errorf("%s %s: %w", key, url.QueryEscape(key), sentinel)
	cases := []struct {
		name   string
		err    error
		status int
		read   bool
	}{
		{name: "closed server"},
		{name: "create request"},
		{name: "transport", err: leak},
		{name: "timeout", err: redactionFailure{leak}},
		{name: "joined transport", err: errors.Join(sentinel, leak)},
		{name: "read success", err: leak, status: 200, read: true},
		{name: "read failure", err: leak, status: 500, read: true},
		{name: "status", status: 500},
	}
	calls := []struct {
		path string
		call func(context.Context, *Client) error
	}{
		{"get", func(ctx context.Context, c *Client) error { _, err := c.GetCapabilities(ctx); return err }},
		{"post", func(ctx context.Context, c *Client) error { _, err := c.GetStats(ctx, StatsRequest{}); return err }},
	}
	for _, tc := range cases {
		for _, p := range calls {
			t.Run(tc.name+"/"+p.path, func(t *testing.T) {
				t.Parallel()
				c := redactionClient(t, key, tc.status, tc.read, tc.err)
				if tc.name == "create request" {
					c.rawBaseURL += "/%zz"
				}
				ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
				defer cancel()
				checkRedaction(t, p.call(ctx, c), key, tc.name, sentinel, tc.err != nil)
			})
		}
	}
}

// assertRedacted checks for key leaks and the requested redaction marker in an error tree.
func assertRedacted(t *testing.T, err error, key string, wantMarker bool) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	if wantMarker && !strings.Contains(err.Error(), "REDACTED") {
		t.Error("missing REDACTED")
	}
	queue := []error{err}
	for i := 0; i < len(queue) && i < 100; i++ {
		text := queue[i].Error()
		if strings.Contains(text, key) || strings.Contains(text, url.QueryEscape(key)) {
			t.Errorf("key leaked: %s", text)
		}
		if child := errors.Unwrap(queue[i]); child != nil {
			queue = append(queue, child)
		}
		if joined, ok := queue[i].(interface{ Unwrap() []error }); ok {
			queue = append(queue, joined.Unwrap()...)
		}
	}
}

// redactionClient creates a client with a test server and optional transport or body errors.
func redactionClient(t *testing.T, key string, status int, read bool, transportErr error) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.Copy(w, strings.NewReader(key+" "+url.QueryEscape(key))); err != nil {
			t.Errorf("nzbhydra test: write redaction response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	hc := srv.Client()
	if status == 0 && transportErr == nil {
		srv.Close()
	}
	if transportErr != nil {
		hc.Transport = redactionTransport(func(r *http.Request) (*http.Response, error) {
			if r.Body != nil {
				r.Body.Close()
			}
			if !read {
				return nil, transportErr
			}
			return &http.Response{StatusCode: status, Body: redactionFailure{transportErr}, Header: make(http.Header)}, nil
		})
	}
	c, err := New(srv.URL, key, WithHTTPClient(hc))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// TestScrubBoundaries checks empty keys, URL error immutability, and error-tree size limits.
func TestScrubBoundaries(t *testing.T) {
	t.Parallel()
	if got := (&Client{}).scrub("plain"); got != "plain" {
		t.Errorf("empty key: scrub = %q", got)
	}
	c := &Client{apiKey: "placeholder"}
	original := &url.Error{URL: "?apikey=placeholder", Err: errors.New("placeholder")}
	assertRedacted(t, c.redactError(original), "placeholder", true)
	if original.URL != "?apikey=placeholder" || original.Err.Error() != "placeholder" {
		t.Error("mutated original URL error")
	}
	for _, width := range []int{maxErrNodes - 1, maxErrNodes, maxErrNodes + 1} {
		children := make([]error, width)
		for i := range children {
			children[i] = errors.New("plain")
		}
		got := c.scrubError(errors.Join(children...))
		var redacted *redactedError
		if hidden := errors.As(got, &redacted); hidden != (width >= maxErrNodes) {
			t.Errorf("width %d: hidden = %v", width, hidden)
		}
	}
}

// checkRedaction checks redaction and preservation of sentinel, timeout, and API errors.
func checkRedaction(t *testing.T, err error, key, name string, sentinel error, hasSentinel bool) {
	t.Helper()
	assertRedacted(t, err, key, name != "create request")
	if hasSentinel && !errors.Is(err, sentinel) {
		t.Error("lost sentinel")
	}
	if name == "timeout" {
		var timed interface{ Timeout() bool }
		if !errors.As(err, &timed) || !timed.Timeout() {
			t.Error("lost timeout")
		}
	}
	if name == "status" {
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Error("missing APIError")
		}
	}
}

// TestRedactError checks root URL sanitization without discarding outer error trees.
func TestRedactError(t *testing.T) {
	t.Parallel()
	const key = "placeholder+k/y=1"
	c := &Client{apiKey: key}
	inner := errors.New("fixture inner sentinel")
	uerr := &url.Error{Op: "Get", URL: "http://example.invalid/?apikey=" + url.QueryEscape(key), Err: inner}
	keyless := errors.Join(io.ErrUnexpectedEOF, errors.New("fixture plain error"))
	cases := []struct {
		name      string
		err       error
		sentinel  error
		hasURL    bool
		unchanged bool
	}{
		{name: "root URL", err: uerr, sentinel: inner, hasURL: true},
		{name: "joined URL", err: errors.Join(io.ErrUnexpectedEOF, uerr), sentinel: io.ErrUnexpectedEOF, hasURL: true},
		{name: "wrapped URL", err: fmt.Errorf("read: %w", uerr), sentinel: inner, hasURL: true},
		{name: "keyless join", err: keyless, sentinel: io.ErrUnexpectedEOF, unchanged: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := c.redactError(tc.err)
			assertNoKey(t, got.Error(), key)
			assertSentinel(t, got, tc.sentinel)
			var matched *url.Error
			if isURL := errors.As(got, &matched); isURL != tc.hasURL {
				t.Errorf("URL error reachable = %v; want %v", isURL, tc.hasURL)
			}
			if same := any(got) == any(tc.err); same != tc.unchanged {
				t.Errorf("same error value = %v; want %v", same, tc.unchanged)
			}
		})
	}
}

// assertSentinel checks that error redaction preserves the expected cause.
func assertSentinel(t *testing.T, got, sentinel error) {
	t.Helper()
	if !errors.Is(got, sentinel) {
		t.Error("lost sentinel")
	}
}

// TestNewRedaction checks constructor errors with keys embedded in invalid URLs.
func TestNewRedaction(t *testing.T) {
	t.Parallel()
	const key = "placeholder+k/y=1"
	for _, base := range []string{"http://%zz/", "ftp://example.invalid/", "https:///"} {
		t.Run(base, func(t *testing.T) {
			t.Parallel()
			c, err := New(base+"?raw="+key+"&escaped="+url.QueryEscape(key), key)
			if c != nil || err == nil {
				t.Fatalf("New = %v, %v; want nil client and error", c, err)
			}
			if base == "http://%zz/" {
				assertNoKey(t, err.Error(), key)
				return
			}
			assertSafeText(t, err.Error(), key)
		})
	}
}

// TestNewRedactsForeignQueryKey checks that constructor errors strip a query apikey different from the client key.
func TestNewRedactsForeignQueryKey(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, baseURL, key string
	}{
		{name: "unsupported scheme", baseURL: "ftp://example.invalid/?apikey=other", key: "k"},
		{name: "parse failure", baseURL: "http://%zz/?apikey=other", key: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := New(tc.baseURL, tc.key)
			if c != nil || err == nil {
				t.Fatalf("New(%q) = %v, %v; want nil client and error", tc.baseURL, c, err)
			}
			if strings.Contains(err.Error(), "other") {
				t.Errorf("error leaked query key: %s", err.Error())
			}
		})
	}
}

// TestCarriesKeyNil checks that a nil error never reports a key or panics.
func TestCarriesKeyNil(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, key string
	}{
		{name: "keyed client", key: "k"},
		{name: "empty key", key: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if (&Client{apiKey: tc.key}).carriesKey(nil) {
				t.Error("carriesKey(nil) = true; want false")
			}
		})
	}
}

// assertNoKey checks text for raw and query-escaped keys.
func assertNoKey(t *testing.T, text, key string) {
	t.Helper()
	if strings.Contains(text, key) || strings.Contains(text, url.QueryEscape(key)) {
		t.Errorf("key leaked: %s", text)
	}
}

// assertSafeText checks both key forms and the redaction marker.
func assertSafeText(t *testing.T, text, key string) {
	t.Helper()
	assertNoKey(t, text, key)
	if !strings.Contains(text, "REDACTED") {
		t.Errorf("missing REDACTED: %s", text)
	}
}

// responseTextClient creates a client serving the supplied response text.
func responseTextClient(t *testing.T, key, body string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := io.Copy(w, strings.NewReader(body)); err != nil {
			t.Errorf("write response text: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	c, err := New(srv.URL, key, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// TestDecodeErrorRedaction checks server text embedded in decoder errors.
func TestDecodeErrorRedaction(t *testing.T) {
	t.Parallel()
	// Why: integer overflow errors echo this oversized numeric fixture key.
	const numericKey = "1234567890123456789012345678901234567890"
	cases := []struct {
		name string
		key  string
		body string
		call func(context.Context, *Client) error
	}{
		{name: "get json", key: numericKey, body: numericKey, call: func(ctx context.Context, c *Client) error {
			var out int
			return c.getJSON(ctx, "/fixture", nil, &out)
		}},
		{name: "post json", key: numericKey, body: numericKey, call: func(ctx context.Context, c *Client) error {
			var out int
			return c.postJSON(ctx, "/fixture", nil, &out)
		}},
		{name: "xml raw", key: "placeholder+k/y=1", body: `<rss><channel><item><size>placeholder+k/y=1</size></item></channel></rss>`, call: func(ctx context.Context, c *Client) error {
			_, err := c.Search(ctx, "fixture", nil)
			return err
		}},
		{name: "xml escaped", key: "placeholder+k/y=1", body: `<rss><channel><item><size>placeholder%2Bk%2Fy%3D1</size></item></channel></rss>`, call: func(ctx context.Context, c *Client) error {
			_, err := c.Search(ctx, "fixture", nil)
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := responseTextClient(t, tc.key, tc.body)
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			err := tc.call(ctx, c)
			if err == nil {
				t.Fatal("expected decoder error")
			}
			assertSafeText(t, err.Error(), tc.key)
		})
	}
}

type responseStatusTransport struct {
	http.RoundTripper
	status string
}

// RoundTrip replaces the fixture response status text.
func (tr responseStatusTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := tr.RoundTripper.RoundTrip(req)
	if err == nil {
		resp.StatusCode = http.StatusInternalServerError
		resp.Status = tr.status
	}
	return resp, err
}

// TestStatusTextRedaction checks server status text exposed through APIError fields.
func TestStatusTextRedaction(t *testing.T) {
	t.Parallel()
	const key = "placeholder+k/y=1"
	cases := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{name: "get", call: func(ctx context.Context, c *Client) error {
			_, err := c.GetCapabilities(ctx)
			return err
		}},
		{name: "post", call: func(ctx context.Context, c *Client) error {
			_, err := c.GetStats(ctx, StatsRequest{})
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := responseTextClient(t, key, "fixture failure")
			c.httpClient.Transport = responseStatusTransport{
				RoundTripper: c.httpClient.Transport,
				status:       "500 " + key + " " + url.QueryEscape(key),
			}
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			err := tc.call(ctx, c)
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected APIError, got %v", err)
			}
			assertSafeText(t, apiErr.Status, key)
			if apiErr.StatusCode != http.StatusInternalServerError {
				t.Errorf("status code = %d", apiErr.StatusCode)
			}
		})
	}
}

// TestRedactQuotedKey checks nested URL errors with Go-quoted keys and preserves joined causes.
func TestRedactQuotedKey(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, key string }{
		{name: "plain", key: "placeholder"},
		{name: "quote and backslash", key: `k"y\z`},
		{name: "control", key: "k\ny\tz"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{apiKey: tc.key}
			original := errors.Join(io.ErrUnexpectedEOF,
				&url.Error{Op: "Get", URL: "http://h/?apikey=" + tc.key, Err: errors.New("x")})
			got := c.redactError(original)
			quoted := strconv.Quote(tc.key)
			text := got.Error()
			if strings.Contains(text, tc.key) || strings.Contains(text, quoted[1:len(quoted)-1]) {
				t.Errorf("key leaked: %s", text)
			}
			if !strings.Contains(text, "REDACTED") {
				t.Error("missing REDACTED")
			}
			if !errors.Is(got, io.ErrUnexpectedEOF) {
				t.Error("lost sentinel")
			}
		})
	}
}

// TestScrubQuotedKey checks raw, query-escaped, Go-quoted, absent, and empty key forms.
func TestScrubQuotedKey(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, key, input, want string
		contains               bool
	}{
		{name: "empty", input: "plain", want: "plain"},
		{name: "absent", key: `k"y\z`, input: "plain", want: "plain"},
		{name: "plain", key: "placeholder", input: "key=placeholder", want: "key=REDACTED", contains: true},
		{name: "raw", key: `k"y\z`, input: `key=k"y\z`, want: "key=REDACTED", contains: true},
		{name: "query", key: `k"y\z`, input: "key=k%22y%5Cz", want: "key=REDACTED", contains: true},
		{name: "quoted", key: `k"y\z`, input: `key=k\"y\\z`, want: "key=REDACTED", contains: true},
		{name: "control", key: "k\ny\tz", input: `key=k\ny\tz`, want: "key=REDACTED", contains: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{apiKey: tc.key}
			if got := c.scrub(tc.input); got != tc.want {
				t.Errorf("scrub = %q; want %q", got, tc.want)
			}
			if got := c.containsKey(tc.input); got != tc.contains {
				t.Errorf("containsKey = %v; want %v", got, tc.contains)
			}
		})
	}
}

// redactionFanout is a keyless error whose Unwrap returns the previous level twice.
type redactionFanout struct{ prev error }

// Error returns constant text so only the tree shape costs time.
func (redactionFanout) Error() string { return "fixture fanout level" }

// Unwrap returns the previous level twice, doubling the tree per level.
func (e redactionFanout) Unwrap() []error { return []error{e.prev, e.prev} }

// TestRedactErrorFanoutBounded checks that a keyless exponential error tree returns promptly.
func TestRedactErrorFanoutBounded(t *testing.T) {
	t.Parallel()
	c := &Client{apiKey: "placeholder+k/y=1"}
	var err error
	err = errors.New("fixture fanout leaf")
	for range 40 {
		err = redactionFanout{prev: err}
	}
	done := make(chan error, 1)
	go func() { done <- c.redactError(err) }()
	select {
	case got := <-done:
		if got == nil {
			t.Error("redactError returned nil for a non-nil error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("redactError did not return within 2s")
	}
}
