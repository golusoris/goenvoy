package sabnzbd

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
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
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := redactionClient(t, key, tc.status, tc.read, tc.err)
			if tc.name == "create request" {
				c.baseURL += "/%zz"
			}
			_, err := c.GetQueue(t.Context())
			checkRedaction(t, err, key, tc.name, sentinel, tc.err != nil)
		})
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
		fmt.Fprint(w, key+" "+url.QueryEscape(key))
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

// checkRedaction checks redaction and preservation of sentinel and timeout errors.
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
}
