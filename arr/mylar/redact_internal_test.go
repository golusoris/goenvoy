package mylar

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
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

// Why: keep fixture setup and error-kind checks outside the table traversal.
func checkRequestRedaction(t *testing.T, key string, sentinel error, closed bool, transport http.RoundTripper, timedOut bool) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "raw %s escaped %s", key, url.QueryEscape(key))
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
	_, err = c.GetIndex(t.Context())
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
