package sabnzbd

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// redactAsProbe is a second error type in the tree, used to prove As answers only a URL error target.
type redactAsProbe struct{ text string }

// Error returns the fixture text.
func (e *redactAsProbe) Error() string { return e.text }

// assertNoKeyForms reports every raw, query-escaped, or Go-quoted form of key found in text.
func assertNoKeyForms(t *testing.T, text, key string) {
	t.Helper()
	quoted := strconv.Quote(key)
	for _, form := range []string{key, url.QueryEscape(key), quoted[1 : len(quoted)-1]} {
		if strings.Contains(text, form) {
			t.Errorf("key form %q leaked: %s", form, text)
		}
	}
}

// TestRedactedErrorAsURLError checks that errors.As reaches a sanitized copy of a nested URL error and nothing else.
func TestRedactedErrorAsURLError(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, key string }{
		{name: "plain", key: "plain-key"},
		{name: "query escaped and quoted", key: `plain-key+k/y="1"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkRedactedErrorAs(t, tc.key)
		})
	}
}

// checkRedactedErrorAs redacts a tree that nests a key-bearing URL error and checks what errors.As reaches.
func checkRedactedErrorAs(t *testing.T, key string) {
	t.Helper()
	c := &Client{apiKey: key}
	inner := &url.Error{
		Op:  "Get",
		URL: "http://h/?apikey=" + url.QueryEscape(key),
		Err: fmt.Errorf("dial %s %q", key, key),
	}
	probe := &redactAsProbe{text: "probe " + key}
	tree := fmt.Errorf("outer: %w", errors.Join(probe, fmt.Errorf("mid: %w", inner)))
	var present *redactAsProbe
	if !errors.As(tree, &present) {
		t.Fatal("fixture broken: probe missing from the original tree")
	}

	got := c.redactError(tree)
	assertNoKeyForms(t, got.Error(), key)
	var uerr *url.Error
	if !errors.As(got, &uerr) {
		t.Fatal("errors.As into *url.Error = false; want true")
	}
	if uerr == inner {
		t.Error("errors.As returned the original URL error, not a copy")
	}
	if uerr.Op != inner.Op {
		t.Errorf("Op = %q; want %q", uerr.Op, inner.Op)
	}
	assertNoKeyForms(t, uerr.Error(), key)
	assertNoKeyForms(t, uerr.URL, key)
	var hidden *redactAsProbe
	if errors.As(got, &hidden) {
		t.Error("errors.As into another tree type = true; want false")
	}
}

// TestRedactedErrorAsFanoutBounded checks that errors.As on a hidden exponential tree with a key-bearing leaf returns promptly.
func TestRedactedErrorAsFanoutBounded(t *testing.T) {
	t.Parallel()
	c := &Client{apiKey: "plain-key"}
	var err error
	err = errors.New("fixture fanout leaf plain-key")
	for range 40 {
		err = redactionFanout{prev: err}
	}
	done := make(chan bool, 1)
	go func() {
		var uerr *url.Error
		done <- errors.As(c.redactError(err), &uerr)
	}()
	select {
	case found := <-done:
		if found {
			t.Error("errors.As found a *url.Error in a tree that holds none")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("errors.As did not return within 2s")
	}
}
