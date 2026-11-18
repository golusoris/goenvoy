package mylar

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

type redactTransport struct{ err error }

// RoundTrip returns the fixture error without sending a request.
func (r redactTransport) RoundTrip(*http.Request) (*http.Response, error) { return nil, r.err }

type redactTimeout struct{ error }

// Unwrap returns the fixture cause.
func (e redactTimeout) Unwrap() error { return e.error }

// Timeout reports the fixture timeout.
func (redactTimeout) Timeout() bool { return true }

// TestRequestRedaction checks transport, timeout, and status error sanitization.
func TestRequestRedaction(t *testing.T) {
	t.Parallel()
	const key = "fixture-k+y/=1"
	sentinel := errors.New("fixture transport failure")
	sensitive := fmt.Errorf("raw %s escaped %s: %w", key, url.QueryEscape(key), sentinel)
	cases := []struct {
		name      string
		closed    bool
		transport http.RoundTripper
		timeout   bool
	}{
		{name: "closed server", closed: true},
		{name: "custom transport", transport: redactTransport{err: sensitive}},
		{name: "timeout", transport: redactTransport{err: redactTimeout{error: sensitive}}, timeout: true},
		{name: "status body"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkRequestRedaction(t, key, sentinel, tc.closed, tc.transport, tc.timeout)
		})
	}
}

// checkRequestRedaction checks redaction and error kinds using the configured fixture.
func checkRequestRedaction(t *testing.T, key string, sentinel error, closed bool, transport http.RoundTripper, timedOut bool) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		body := fmt.Sprintf("raw %s escaped %s", key, url.QueryEscape(key))
		if _, err := io.Copy(w, strings.NewReader(body)); err != nil {
			t.Errorf("mylar test: write redaction response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	if closed {
		server.Close()
	}
	client := server.Client()
	if transport != nil {
		client.Transport = transport
	}
	c, err := New(server.URL, key, WithHTTPClient(client))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	_, err = c.GetIndex(ctx)
	assertRedactedTree(t, err, key)
	assertRedactionKind(t, err, sentinel, closed, transport, timedOut)
}

// Why: sentinel, timeout, and status contracts survive text sanitization.
func assertRedactionKind(t *testing.T, err, sentinel error, closed bool, transport http.RoundTripper, timedOut bool) {
	t.Helper()
	if transport != nil && !errors.Is(err, sentinel) {
		t.Error("sentinel match lost")
	}
	if timedOut {
		var timeout interface{ Timeout() bool }
		if !errors.As(err, &timeout) || !timeout.Timeout() {
			t.Error("timeout lost")
		}
	}
	if !closed && transport == nil {
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected HTTP 500 APIError, got %v", err)
		}
	}
}

// Why: inspect every reachable error and diagnostic format for fixture key text.
func assertRedactedTree(t *testing.T, err error, key string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "REDACTED") {
		t.Error("redaction marker missing")
	}
	queue := []error{err}
	for i := 0; i < len(queue) && i < 256; i++ {
		current := queue[i]
		text := fmt.Sprintf("%v %+v %#v", current, current, current)
		if strings.Contains(text, key) || strings.Contains(text, url.QueryEscape(key)) {
			t.Errorf("key leaked: %s", text)
		}
		if child := errors.Unwrap(current); child != nil {
			queue = append(queue, child)
		}
		if joined, ok := current.(interface{ Unwrap() []error }); ok {
			queue = append(queue, joined.Unwrap()...)
		}
	}
	if len(queue) > 256 {
		t.Fatal("error tree exceeded test bound")
	}
}

// TestScrub checks raw, escaped, empty, and absent key text.
func TestScrub(t *testing.T) {
	t.Parallel()
	const key = "fixture-k+y/=1"
	cases := []struct{ name, key, input, want string }{
		{"empty key", "", "unchanged", "unchanged"},
		{"absent key", key, "unchanged", "unchanged"},
		{"raw and escaped", key, key + ":" + url.QueryEscape(key), "REDACTED:REDACTED"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{apiKey: tc.key}
			if got := c.scrub(tc.input); got != tc.want {
				t.Errorf("scrub = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestWideErrorTree checks fail-closed handling beyond the traversal bound.
func TestWideErrorTree(t *testing.T) {
	t.Parallel()
	for _, count := range []int{maxErrNodes - 1, 64, 65, 256} {
		children := make([]error, count)
		for i := range children {
			children[i] = errors.New("fixture plain error")
		}
		joined := errors.Join(children...)
		c := &Client{apiKey: "fixture-k+y/=1"}
		got := c.scrubError(joined)
		var hidden *redactedError
		if count == maxErrNodes-1 {
			assertVisibleErrorTree(t, got, children[0])
			continue
		}
		if !errors.As(got, &hidden) || errors.Unwrap(got) != nil {
			t.Fatal("wide tree exposed")
		}
		if !errors.Is(got, children[0]) {
			t.Error("sentinel match lost")
		}
	}
}

// assertVisibleErrorTree checks visibility and sentinel matching below the node bound.
func assertVisibleErrorTree(t *testing.T, got, sentinel error) {
	t.Helper()
	var hidden *redactedError
	if errors.As(got, &hidden) {
		t.Fatal("below-bound tree hidden")
	}
	if !errors.Is(got, sentinel) {
		t.Error("below-bound sentinel match lost")
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
		{name: "json number", key: numericKey, body: numericKey, call: func(ctx context.Context, c *Client) error {
			var out int
			return c.get(ctx, "getComic", nil, &out)
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
			_, err := c.GetComic(ctx, "fixture")
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
