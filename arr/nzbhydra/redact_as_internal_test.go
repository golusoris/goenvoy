package nzbhydra

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

// redactSliceErr is an uncomparable error type; comparing two values of it with == panics.
type redactSliceErr []string

// Error returns constant fixture text.
func (redactSliceErr) Error() string { return "fixture slice error" }

// hiddenRedacted returns the redactedError that scrubError builds for err and fails the test when err is not hidden.
func hiddenRedacted(t *testing.T, key string, err error) error {
	t.Helper()
	got := (&Client{apiKey: key}).scrubError(err)
	var hidden *redactedError
	if !errors.As(got, &hidden) {
		t.Fatalf("scrubError(%v) did not hide the tree", err)
	}
	return got
}

// answerWithinTwoSeconds runs ask and fails the test, instead of hanging, when it does not answer within 2s.
func answerWithinTwoSeconds(t *testing.T, name string, ask func() bool) bool {
	t.Helper()
	done := make(chan bool, 1)
	go func() { done <- ask() }()
	select {
	case got := <-done:
		return got
	case <-time.After(2 * time.Second):
		t.Fatalf("%s did not return within 2s", name)
		return false
	}
}

// TestRedactedErrorSearchesFanoutBounded checks that Timeout, Temporary and errors.Is on a hidden exponential tree return promptly.
func TestRedactedErrorSearchesFanoutBounded(t *testing.T) {
	t.Parallel()
	var err error
	err = errors.New("fixture fanout leaf")
	for range 30 {
		err = redactionFanout{prev: err}
	}
	hidden := hiddenRedacted(t, "plain-key", err)
	absent := errors.New("not in the tree")
	checks := map[string]func() bool{
		"Timeout":   func() bool { return hidden.(interface{ Timeout() bool }).Timeout() },     //nolint:forcetypeassert // the wrapper's own method set under test
		"Temporary": func() bool { return hidden.(interface{ Temporary() bool }).Temporary() }, //nolint:forcetypeassert // the wrapper's own method set under test
		"Is":        func() bool { return errors.Is(hidden, absent) },
	}
	for name, ask := range checks {
		if answerWithinTwoSeconds(t, name, ask) {
			t.Errorf("%s = true on a tree with no matching node", name)
		}
	}
}

// TestRedactedErrorIsFindsSentinel checks that errors.Is reaches a sentinel three wraps deep and one inside an errors.Join branch.
func TestRedactedErrorIsFindsSentinel(t *testing.T) {
	t.Parallel()
	const key = "plain-key"
	sentinel := errors.New("fixture sentinel")
	deep := fmt.Errorf("a %s: %w", key, fmt.Errorf("b: %w", fmt.Errorf("c: %w", sentinel)))
	joined := errors.Join(errors.New("other branch"), fmt.Errorf("k %s: %w", key, sentinel))
	for name, tree := range map[string]error{"three wraps deep": deep, "join branch": joined} {
		if !errors.Is(hiddenRedacted(t, key, tree), sentinel) {
			t.Errorf("%s: errors.Is = false; want true", name)
		}
	}
}

// TestRedactedErrorIsUncomparableTarget checks that errors.Is with an uncomparable target reports false and never panics.
func TestRedactedErrorIsUncomparableTarget(t *testing.T) {
	t.Parallel()
	const key = "plain-key"
	leaf := redactSliceErr{"leaf"}
	hidden := hiddenRedacted(t, key, fmt.Errorf("a %s: %w", key, leaf))
	if errors.Is(hidden, redactSliceErr{"x"}) {
		t.Error("errors.Is with an uncomparable target = true; want false")
	}
}
