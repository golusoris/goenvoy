package jackett

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

// TestRedactURL verifies that redactURL replaces apikey query parameter values.
func TestRedactURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "replaces apikey preserving other params",
			raw:  "http://h:9117/api/v2.0/indexers?apikey=K1&configured=true",
			want: "http://h:9117/api/v2.0/indexers?apikey=REDACTED&configured=true",
		},
		{
			name: "unchanged without apikey",
			raw:  "http://h:9117/api/v2.0/indexers?configured=true",
			want: "http://h:9117/api/v2.0/indexers?configured=true",
		},
		{
			name: "empty apikey redacted",
			raw:  "http://h:9117/api/v2.0/indexers?apikey=",
			want: "http://h:9117/api/v2.0/indexers?apikey=REDACTED",
		},
		{
			name: "malformed escape in apikey drops query",
			raw:  "http://h:9117/api/v2.0/indexers?apikey=K1%zz&configured=true",
			want: "http://h:9117/api/v2.0/indexers",
		},
		{
			name: "semicolon separator drops query",
			raw:  "http://h:9117/api/v2.0/indexers?apikey=K1;configured=true",
			want: "http://h:9117/api/v2.0/indexers",
		},
		{
			name: "malformed other param drops query",
			raw:  "http://h:9117/api/v2.0/indexers?configured=%zz&apikey=K1",
			want: "http://h:9117/api/v2.0/indexers",
		},
		{
			name: "malformed query without apikey drops query",
			raw:  "http://h:9117/api/v2.0/indexers?x=%zz",
			want: "http://h:9117/api/v2.0/indexers",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := redactURL(tc.raw)
			if got != tc.want {
				t.Errorf("redactURL(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

// TestRedactURLUnparseable verifies that unparseable URLs strip query parameters without leaking keys.
func TestRedactURLUnparseable(t *testing.T) {
	t.Parallel()

	const raw = "http://[::1%zz]/x?apikey=K2"
	got := redactURL(raw)
	if got != "http://[::1%zz]/x" {
		t.Errorf("redactURL(%q) = %q, want %q", raw, got, "http://[::1%zz]/x")
	}
	if strings.Contains(got, "K2") {
		t.Errorf("redactURL(%q) = %q; contains K2", raw, got)
	}
	if strings.Contains(got, "?") {
		t.Errorf("redactURL(%q) = %q; contains ?", raw, got)
	}
}

// TestDoRequestCreateRequestRedactsAPIKey verifies that request creation errors redact the API key.
func TestDoRequestCreateRequestRedactsAPIKey(t *testing.T) {
	t.Parallel()

	const key = `s3cr3t+key/=&x`
	c, err := New("http://127.0.0.1:9117", key)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = c.doRequest(context.Background(), "http://[::1%zz]/x?apikey="+url.QueryEscape(key))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errMsg := err.Error()
	if !strings.HasPrefix(errMsg, "jackett: create request: ") {
		t.Errorf("err.Error() = %q, want prefix %q", errMsg, "jackett: create request: ")
	}
	if strings.Contains(errMsg, key) {
		t.Errorf("err.Error() leaked raw key: %q", errMsg)
	}
	if strings.Contains(errMsg, url.QueryEscape(key)) {
		t.Errorf("err.Error() leaked escaped key: %q", errMsg)
	}
	var uerr *url.Error
	if !errors.As(err, &uerr) {
		t.Errorf("expected *url.Error in chain, got %T", err)
	}
}

// TestScrub verifies scrubbing and detection of raw, query-escaped, and quoted keys, absent text, and empty keys.
func TestScrub(t *testing.T) {
	t.Parallel()

	const special = `k"y\z`
	const control = "key\n\tvalue"
	quoted := strconv.Quote("plain-key")
	cases := []struct {
		name, key, input, want string
		contains               bool
	}{
		{"plain raw", "plain-key", "before plain-key after", "before REDACTED after", true},
		{"plain query", "plain-key", "before " + url.QueryEscape("plain-key") + " after", "before REDACTED after", true},
		{"plain quoted", "plain-key", "before " + quoted[1:len(quoted)-1] + " after", "before REDACTED after", true},
		{"special raw", special, "before " + special + " after", "before REDACTED after", true},
		{"special query", special, "before " + url.QueryEscape(special) + " after", "before REDACTED after", true},
		{"special quoted", special, `before k\"y\\z after`, "before REDACTED after", true},
		{"control raw", control, "before " + control + " after", "before REDACTED after", true},
		{"control query", control, "before " + url.QueryEscape(control) + " after", "before REDACTED after", true},
		{"control quoted", control, `before key\n\tvalue after`, "before REDACTED after", true},
		{"absent", special, "plain error message", "plain error message", false},
		{"empty input", control, "", "", false},
		{"empty key", "", "plain error message", "plain error message", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{apiKey: tc.key}
			if got := c.scrub(tc.input); got != tc.want {
				t.Errorf("scrub(%q) = %q, want %q", tc.input, got, tc.want)
			}
			if got := c.containsKey(tc.input); got != tc.contains {
				t.Errorf("containsKey(%q) = %v, want %v", tc.input, got, tc.contains)
			}
		})
	}
}

// TestRedactErrorWrappers verifies that joined and wrapped URL errors retain their identities and redact all key forms.
func TestRedactErrorWrappers(t *testing.T) {
	t.Parallel()

	for _, key := range []string{"plain-key", `k"y\z`, "key\n\tvalue"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			checkRedactErrorWrappers(t, key)
		})
	}
}

// checkRedactErrorWrappers checks wrapper text and errors.Is identities after redaction.
func checkRedactErrorWrappers(t *testing.T, key string) {
	t.Helper()

	uerr := &url.Error{Op: "Get", URL: "http://h/?apikey=" + key, Err: errors.New("x")}
	cases := []struct {
		name, prefix string
		err, target  error
	}{
		{"joined", io.ErrUnexpectedEOF.Error(), errors.Join(io.ErrUnexpectedEOF, uerr), io.ErrUnexpectedEOF},
		{"wrapped", "read: ", fmt.Errorf("read: %w", uerr), uerr},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := (&Client{apiKey: key}).redactError(tc.err)
			checkRedactedKeyForms(t, got.Error(), key)
			if !strings.HasPrefix(got.Error(), tc.prefix) {
				t.Errorf("error = %q, want prefix %q", got, tc.prefix)
			}
			if !errors.Is(got, tc.err) || !errors.Is(got, tc.target) {
				t.Errorf("errors.Is lost the root error or target %v", tc.target)
			}
		})
	}
}

// checkRedactedKeyForms checks that text contains REDACTED and no raw, query-escaped, or quoted key.
func checkRedactedKeyForms(t *testing.T, text, key string) {
	t.Helper()

	quoted := strconv.Quote(key)
	for _, form := range []string{key, url.QueryEscape(key), quoted[1 : len(quoted)-1]} {
		if strings.Contains(text, form) {
			t.Errorf("error contains key form %q: %q", form, text)
		}
	}
	if !strings.Contains(text, "REDACTED") {
		t.Errorf("error lacks REDACTED: %q", text)
	}
}

// TestRedactErrorRootURL verifies that root URL errors are sanitized copies, retaining their underlying errors.
func TestRedactErrorRootURL(t *testing.T) {
	t.Parallel()

	for _, key := range []string{"plain-key", `k"y\z`, "key\n\tvalue"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			checkRedactErrorRootURL(t, key)
		})
	}
}

// checkRedactErrorRootURL checks the copied URL error and the unchanged original.
func checkRedactErrorRootURL(t *testing.T, key string) {
	t.Helper()

	rawURL := "http://h/?apikey=" + url.QueryEscape(key)
	inner := errors.New("x")
	uerr := &url.Error{Op: "Get", URL: rawURL, Err: inner}
	got := (&Client{apiKey: key}).redactError(uerr)
	var clean *url.Error
	if !errors.As(got, &clean) || any(got) != any(clean) {
		t.Fatalf("redactError = %T, want *url.Error", got)
	}
	if clean == uerr || uerr.URL != rawURL {
		t.Error("redactError did not preserve the original URL error")
	}
	if clean.URL != "http://h/?apikey=REDACTED" || clean.Op != uerr.Op || !errors.Is(got, inner) {
		t.Errorf("redactError = %#v, want sanitized URL with original operation and underlying error", clean)
	}
	checkRedactedKeyForms(t, got.Error(), key)
}

// TestRedactURLErrorNonURLError verifies that non-url.Error values and nil are returned unchanged.
func TestRedactURLErrorNonURLError(t *testing.T) {
	t.Parallel()

	c := &Client{apiKey: "secret"}
	plain := errors.New("boom")
	got := c.redactError(plain)
	if !errors.Is(got, plain) {
		t.Errorf("errors.Is(got, plain) = false, want true")
	}

	if gotNil := c.redactError(nil); gotNil != nil {
		t.Errorf("c.redactError(nil) = %v, want nil", gotNil)
	}
}

// TestScrubErrorHidesOriginal verifies that a scrubbed error hides the original from Unwrap but still matches errors.Is.
func TestScrubErrorHidesOriginal(t *testing.T) {
	t.Parallel()

	c := &Client{apiKey: "secret"}
	orig := errors.New("dial with apikey=secret")
	got := c.scrubError(orig)
	if got.Error() != "dial with apikey=REDACTED" {
		t.Errorf("got.Error() = %q, want %q", got.Error(), "dial with apikey=REDACTED")
	}
	if inner := errors.Unwrap(got); inner != nil {
		t.Errorf("errors.Unwrap(got) = %v, want nil", inner)
	}
	if !errors.Is(got, orig) {
		t.Errorf("errors.Is(got, orig) = false, want true")
	}
}

// quietLink hides the error it wraps from its own message.
type quietLink struct{ err error }

// Error returns a message without the wrapped error's text.
func (quietLink) Error() string { return "link" }

// Unwrap returns the wrapped error.
func (l quietLink) Unwrap() error { return l.err }

// quietChain returns a chain of n errors: n-1 quiet links over leaf.
func quietChain(n int, leaf error) error {
	err := leaf
	for range n - 1 {
		err = quietLink{err: err}
	}
	return err
}

// TestCarriesKeyBound verifies that carriesKey reads exactly maxErrNodes errors and treats a longer tree as carrying the key.
func TestCarriesKeyBound(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "no key within bound", err: quietChain(maxErrNodes, errors.New("plain")), want: false},
		{name: "key at last node read", err: quietChain(maxErrNodes, errors.New("apikey=secret")), want: true},
		{name: "one node past bound", err: quietChain(maxErrNodes+1, errors.New("plain")), want: true},
		{name: "wide join past bound", err: errors.Join(plainLinks(4 * maxErrNodes)...), want: true},
	}

	c := &Client{apiKey: "secret"}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := c.carriesKey(tc.err); got != tc.want {
				t.Errorf("c.carriesKey(chain) = %v, want %v", got, tc.want)
			}
		})
	}
}

// plainLinks returns n quiet links over key-free errors.
func plainLinks(n int) []error {
	errs := make([]error, n)
	for i := range errs {
		errs[i] = quietLink{err: errors.New("plain")}
	}
	return errs
}

// TestCarriesKey verifies raw and query-escaped key detection, including a key equal to the redaction placeholder and an empty key.
func TestCarriesKey(t *testing.T) {
	t.Parallel()

	const escapedKey = `s3cr3t+key/=&x`
	cases := []struct {
		name string
		key  string
		err  error
		want bool
	}{
		{name: "raw key", key: "secret", err: errors.New("dial apikey=secret"), want: true},
		{name: "escaped key only", key: escapedKey, err: errors.New("dial apikey=" + url.QueryEscape(escapedKey)), want: true},
		{name: "key equals placeholder", key: "REDACTED", err: errors.New("dial apikey=REDACTED"), want: true},
		{name: "no key text", key: "secret", err: errors.New("dial refused"), want: false},
		{name: "empty key never matches", key: "", err: errors.New("dial refused"), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{apiKey: tc.key}
			if got := c.carriesKey(tc.err); got != tc.want {
				t.Errorf("c.carriesKey(%q) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// valueKeyError is a value-type error whose field holds key text, the shape %#v prints field by field.
type valueKeyError struct{ text string }

// Error returns the key-bearing text.
func (e valueKeyError) Error() string { return e.text }

// TestRedactedErrorFormatHidesOriginal verifies that every fmt verb prints the sanitized message as it prints a plain string, never the hidden error.
func TestRedactedErrorFormatHidesOriginal(t *testing.T) {
	t.Parallel()

	const sanitized = "dial with apikey=REDACTED"
	c := &Client{apiKey: "secret"}
	got := c.scrubError(valueKeyError{text: "dial with apikey=secret"})
	for _, verb := range []string{"%v", "%+v", "%s", "%q", "%x", "%#v", "%d", "%t", "%10.4v"} {
		if out, want := fmt.Sprintf(verb, got), fmt.Sprintf(verb, sanitized); out != want {
			t.Errorf("fmt.Sprintf(%q, err) = %q, want %q", verb, out, want)
		}
	}
}

// TestAppendWrappedBound verifies that appendWrapped never grows the queue past maxErrNodes+1 errors.
func TestAppendWrappedBound(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		queue   []error
		err     error
		wantLen int
	}{
		{name: "wide join stops past bound", queue: []error{errors.New("root")}, err: errors.Join(plainLinks(4 * maxErrNodes)...), wantLen: maxErrNodes + 1},
		{name: "full queue skips single unwrap", queue: make([]error, maxErrNodes+1), err: quietLink{err: errors.New("plain")}, wantLen: maxErrNodes + 1},
		{name: "narrow join keeps every non-nil child", queue: []error{errors.New("root")}, err: errors.Join(nil, errors.New("a"), errors.New("b")), wantLen: 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			queue := append([]error(nil), tc.queue...)
			if got := len(appendWrapped(queue, tc.err)); got != tc.wantLen {
				t.Errorf("len(appendWrapped(queue, err)) = %d, want %d", got, tc.wantLen)
			}
		})
	}
}
