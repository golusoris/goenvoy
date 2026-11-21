package mylar

import (
	"errors"
	"fmt"
	"net"
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

// TestRedactedErrorAsOpRedaction checks that errors.As yields a copy whose Op and text lack the key held by a nested URL error's Op.
func TestRedactedErrorAsOpRedaction(t *testing.T) {
	t.Parallel()
	const key = "plain-key"
	c := &Client{apiKey: key}
	inner := &url.Error{Op: "Get " + key, URL: "http://h/", Err: errors.New("fixture transport failure")}
	tree := fmt.Errorf("outer: %w", errors.Join(&redactAsProbe{text: "probe"}, fmt.Errorf("mid: %w", inner)))

	got := c.redactError(tree)
	assertNoKeyForms(t, got.Error(), key)
	var uerr *url.Error
	if !errors.As(got, &uerr) {
		t.Fatal("errors.As into *url.Error = false; want true")
	}
	if uerr == inner {
		t.Error("errors.As returned the original URL error, not a copy")
	}
	assertNoKeyForms(t, uerr.Op, key)
	assertNoKeyForms(t, uerr.Error(), key)
	if inner.Op != "Get "+key {
		t.Errorf("original Op = %q; the copy must not mutate it", inner.Op)
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

// redactCycleErr is a single-unwrap error whose Unwrap returns a cycle.
type redactCycleErr struct {
	text string
	next error
}

// Error returns the fixture text.
func (e *redactCycleErr) Error() string { return e.text }

// Unwrap returns the next error in the cycle.
func (e *redactCycleErr) Unwrap() error { return e.next }

// TestRedactedErrorWalkCycle checks that errors.Is, Timeout, and Temporary terminate on cyclic single-unwrap chains.
func TestRedactedErrorWalkCycle(t *testing.T) {
	t.Parallel()
	const key = "plain-key"
	unrelated := errors.New("unrelated sentinel")
	self := &redactCycleErr{text: "self " + key}
	self.next = self
	first := &redactCycleErr{text: "first " + key}
	second := &redactCycleErr{text: "second", next: first}
	first.next = second
	cases := []struct {
		name string
		err  error
	}{
		{name: "self cycle", err: self},
		{name: "two-node cycle", err: first},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := hiddenRedacted(t, key, tc.err)
			if answerWithinTwoSeconds(t, "Is", func() bool { return errors.Is(got, unrelated) }) {
				t.Error("errors.Is(got, unrelated) = true; want false")
			}
			if answerWithinTwoSeconds(t, "Timeout", func() bool {
				return got.(interface{ Timeout() bool }).Timeout() //nolint:forcetypeassert // wrapper method set under test
			}) {
				t.Error("Timeout() = true; want false")
			}
			if answerWithinTwoSeconds(t, "Temporary", func() bool {
				return got.(interface{ Temporary() bool }).Temporary() //nolint:forcetypeassert // wrapper method set under test
			}) {
				t.Error("Temporary() = true; want false")
			}
		})
	}
}

// TestRedactedErrorIsFindsSentinel checks that errors.Is reaches a sentinel three wraps deep, inside an errors.Join branch, and across a 200-wrap quiet chain.
func TestRedactedErrorIsFindsSentinel(t *testing.T) {
	t.Parallel()
	const key = "plain-key"
	sentinel := errors.New("fixture sentinel")
	unrelatedSentinel := errors.New("unrelated sentinel")
	deep := fmt.Errorf("a %s: %w", key, fmt.Errorf("b: %w", fmt.Errorf("c: %w", sentinel)))
	joined := errors.Join(errors.New("other branch"), fmt.Errorf("k %s: %w", key, sentinel))
	longChain := sentinel
	for range 200 {
		longChain = &redactCycleErr{text: "quiet wrap", next: longChain}
	}
	cases := []struct {
		name string
		tree error
	}{
		{name: "three wraps deep", tree: deep},
		{name: "join branch", tree: joined},
		{name: "long quiet chain", tree: longChain},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			hidden := hiddenRedacted(t, key, tc.tree)
			if !errors.Is(hidden, sentinel) {
				t.Error("errors.Is(hidden, sentinel) = false; want true")
			}
			if errors.Is(hidden, unrelatedSentinel) {
				t.Error("errors.Is(hidden, unrelatedSentinel) = true; want false")
			}
		})
	}
}

// TestRedactedErrorAsTypedNilURLError checks that errors.As skips a typed-nil *url.Error without panicking and still finds a valid *url.Error sibling.
func TestRedactedErrorAsTypedNilURLError(t *testing.T) {
	t.Parallel()
	const key = "plain-key+k/y=1"
	c := &Client{apiKey: key}
	var typedNil *url.Error
	valid := &url.Error{
		Op:  "Get " + key,
		URL: "http://h/?apikey=" + url.QueryEscape(key),
		Err: fmt.Errorf("dial %s", key),
	}
	cases := []struct {
		name      string
		tree      error
		wantMatch bool
	}{
		{
			name:      "typed nil before valid URL error",
			tree:      &redactCycleErr{text: "outer " + key, next: errors.Join(typedNil, valid)},
			wantMatch: true,
		},
		{
			name:      "only typed nil URL error",
			tree:      &redactCycleErr{text: "outer " + key, next: errors.Join(typedNil)},
			wantMatch: false,
		},
		{
			name:      "root typed nil URL error",
			tree:      typedNil,
			wantMatch: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertTypedNilURLErrorAs(t, c, tc.tree, valid, key, tc.wantMatch)
		})
	}
}

// assertTypedNilURLErrorAs checks errors.As against a hidden tree containing a typed-nil *url.Error.
func assertTypedNilURLErrorAs(t *testing.T, c *Client, tree error, valid *url.Error, key string, wantMatch bool) {
	t.Helper()
	got := c.redactError(tree)
	assertNoKeyForms(t, got.Error(), key)
	var uerr *url.Error
	if matched := errors.As(got, &uerr); matched != wantMatch {
		t.Fatalf("errors.As = %v; want %v", matched, wantMatch)
	}
	if !wantMatch {
		if uerr != nil {
			t.Errorf("uerr = %v; want nil", uerr)
		}
		return
	}
	if uerr == nil || uerr == valid {
		t.Fatalf("uerr = %v; want a non-nil sanitized copy", uerr)
	}
	assertNoKeyForms(t, uerr.Error(), key)
	assertNoKeyForms(t, uerr.URL, key)
}

// TestRedactErrorJoinFanoutBounded checks that a 40-level errors.Join fan-out over a key-bearing leaf returns promptly with the fixed over-budget message.
func TestRedactErrorJoinFanoutBounded(t *testing.T) {
	t.Parallel()
	const key = "placeholder+k/y=1"
	c := &Client{apiKey: key}
	tree := fmt.Errorf("leaf %s", key)
	for range 40 {
		tree = errors.Join(tree, tree)
	}
	cases := []struct {
		name string
		call func(error) error
	}{
		{name: "scrubError", call: c.scrubError},
		{name: "redactError", call: c.redactError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got error
			if !answerWithinTwoSeconds(t, tc.name, func() bool {
				got = tc.call(tree)
				return got != nil
			}) {
				t.Fatalf("%s returned nil", tc.name)
			}
			var hidden *redactedError
			if !errors.As(got, &hidden) {
				t.Fatalf("%s(%T) = %T; want *redactedError", tc.name, tree, got)
			}
			if got.Error() != redactedTreeMsg {
				t.Errorf("Error() = %q; want %q", got.Error(), redactedTreeMsg)
			}
			assertNoKeyForms(t, got.Error(), key)
		})
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

// redactIsNode is a tree node that answers errors.Is through its own Is method, for one sentinel it does not equal.
type redactIsNode struct {
	text    string
	matches error
}

// Error returns the fixture text.
func (n redactIsNode) Error() string { return n.text }

// Is reports whether target is the one sentinel this node stands for.
func (n redactIsNode) Is(target error) bool { return target == n.matches }

// TestRedactedErrorIsNodeMethod checks that errors.Is reaches a node that matches only through its own Is method, two wraps deep.
func TestRedactedErrorIsNodeMethod(t *testing.T) {
	t.Parallel()
	const key = "node-key"
	matched := errors.New("fixture matched sentinel")
	other := errors.New("fixture other sentinel")
	node := redactIsNode{text: "node " + key, matches: matched}
	tree := fmt.Errorf("a %s: %w", key, fmt.Errorf("b: %w", node))
	hidden := hiddenRedacted(t, key, tree)
	if !errors.Is(hidden, matched) {
		t.Error("errors.Is through the node's Is method = false; want true")
	}
	if errors.Is(hidden, other) {
		t.Error("errors.Is with a sentinel the node does not report = true; want false")
	}
}

// TestRedactErrorRootPathQuoteAndSpace checks that a root URL error whose path holds a key with quotes and spaces is scrubbed before and after URL parsing.
func TestRedactErrorRootPathQuoteAndSpace(t *testing.T) {
	t.Parallel()
	for _, key := range []string{`k"y z`, `k"y/z 1`} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			assertRootPathRedacted(t, key)
		})
	}
}

// assertRootPathRedacted checks that redactError scrubs raw and escaped forms of key from a root URL error's path and query.
func assertRootPathRedacted(t *testing.T, key string) {
	t.Helper()
	c := &Client{apiKey: key}
	uerr := &url.Error{
		Op:  "Get",
		URL: "http://h/" + key + "?apikey=" + key,
		Err: errors.New("fixture transport failure"),
	}
	got := c.redactError(uerr)
	var clean *url.Error
	if !errors.As(got, &clean) || clean == nil {
		t.Fatal("errors.As into *url.Error = false or nil; want non-nil")
	}
	for _, text := range []string{got.Error(), clean.URL} {
		assertNoKeyForms(t, text, key)
		if strings.Contains(text, url.PathEscape(key)) || strings.Contains(text, "k%22y") {
			t.Errorf("escaped key leaked: %s", text)
		}
	}
}

// TestRedactErrorTypedNilTree checks that trees holding a typed-nil error return the fixed message without panicking.
func TestRedactErrorTypedNilTree(t *testing.T) {
	t.Parallel()
	c := &Client{apiKey: "plain-key"}
	tree := errors.Join((*url.Error)(nil), errors.New("boom"))
	cases := []struct {
		name string
		call func(error) error
	}{
		{name: "redactError", call: c.redactError},
		{name: "scrubError", call: c.scrubError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, input := range []error{tree, (*url.Error)(nil)} {
				got := tc.call(input)
				if got == nil {
					t.Fatalf("%s(%v) = nil; want non-nil", tc.name, input)
				}
				if got.Error() != redactedTreeMsg {
					t.Errorf("%s(%v).Error() = %q; want %q", tc.name, input, got.Error(), redactedTreeMsg)
				}
			}
		})
	}
}

// TestRedactedErrorTypedNilWalk checks that Is, As, Timeout, and Temporary skip a typed-nil node without panicking.
func TestRedactedErrorTypedNilWalk(t *testing.T) {
	t.Parallel()
	c := &Client{apiKey: "plain-key"}
	tree := &redactCycleErr{text: "outer", next: errors.Join((*url.Error)(nil))}
	got := c.redactError(tree)
	if errors.Is(got, errors.New("unrelated")) {
		t.Error("errors.Is = true; want false")
	}
	var uerr *url.Error
	if errors.As(got, &uerr) || uerr != nil {
		t.Errorf("errors.As = true or non-nil (%v); want false and nil", uerr)
	}
	if got.(interface{ Timeout() bool }).Timeout() { //nolint:forcetypeassert // wrapper method set under test
		t.Error("Timeout() = true; want false")
	}
	if got.(interface{ Temporary() bool }).Temporary() { //nolint:forcetypeassert // wrapper method set under test
		t.Error("Temporary() = true; want false")
	}
}

// TestIsNilError checks nil detection across untyped nil, typed-nil values, non-nil pointers, and value errors.
func TestIsNilError(t *testing.T) {
	t.Parallel()
	var typedNil *url.Error
	var sliceNil redactSliceErr
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: true},
		{name: "typed-nil pointer", err: typedNil, want: true},
		{name: "typed-nil slice", err: sliceNil, want: true},
		{name: "non-nil pointer", err: &url.Error{Op: "Get", URL: "http://h/", Err: errors.New("x")}, want: false},
		{name: "value type", err: redactIsNode{text: "v"}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isNilError(tc.err); got != tc.want {
				t.Errorf("isNilError(%v) = %v; want %v", tc.err, got, tc.want)
			}
		})
	}
}

// redactAsClassifier supplies classifiers through a custom As method.
type redactAsClassifier struct{}

// Error returns the fixture error text.
func (redactAsClassifier) Error() string { return "classifier" }

// As supplies the requested classifier.
func (redactAsClassifier) As(target any) bool {
	switch dst := target.(type) {
	case *interface{ Timeout() bool }:
		*dst = redactClassifier{}
		return true
	case *interface{ Temporary() bool }:
		*dst = redactClassifier{}
		return true
	}
	return false
}

// redactClassifier answers both classification methods.
type redactClassifier struct{}

// Timeout reports a timeout.
func (redactClassifier) Timeout() bool { return true }

// Temporary reports a temporary error.
func (redactClassifier) Temporary() bool { return true }

// TestRedactedErrorClassifiers checks custom As and URL errors with nil and non-nil causes.
func TestRedactedErrorClassifiers(t *testing.T) {
	t.Parallel()
	const key = "fixture/part+token"
	cases := []struct {
		name  string
		cause error
		want  bool
	}{
		{"custom As", redactAsClassifier{}, true},
		{"typed nil URL cause", &url.Error{Op: "Get", URL: "http://h/", Err: (*net.DNSError)(nil)}, false},
		{"URL cause", &url.Error{Op: "Get", URL: "http://h/", Err: &net.DNSError{IsTimeout: true, IsTemporary: true}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := hiddenRedacted(t, key, errors.Join(errors.New(key), tc.cause))
			if value := got.(interface{ Timeout() bool }).Timeout(); value != tc.want { //nolint:forcetypeassert // wrapper method under test
				t.Errorf("Timeout() = %v; want %v", value, tc.want)
			}
			if value := got.(interface{ Temporary() bool }).Temporary(); value != tc.want { //nolint:forcetypeassert // wrapper method under test
				t.Errorf("Temporary() = %v; want %v", value, tc.want)
			}
		})
	}
}
