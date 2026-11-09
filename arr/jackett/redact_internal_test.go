package jackett

import (
	"context"
	"errors"
	"net/url"
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

// TestScrub verifies that scrub redacts raw and escaped keys and handles boundary cases.
func TestScrub(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		apiKey string
		input  string
		want   string
	}{
		{
			name:   "empty key boundary leaves input unchanged",
			apiKey: "",
			input:  "http://example.com/api?foo=bar",
			want:   "http://example.com/api?foo=bar",
		},
		{
			name:   "raw key replaced",
			apiKey: "my-secret-key",
			input:  "error with my-secret-key present",
			want:   "error with REDACTED present",
		},
		{
			name:   "query escaped key replaced",
			apiKey: "key with spaces&symbols+",
			input:  "url: ?key=" + url.QueryEscape("key with spaces&symbols+"),
			want:   "url: ?key=REDACTED",
		},
		{
			name:   "text without key unchanged",
			apiKey: "secret",
			input:  "plain error message",
			want:   "plain error message",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{apiKey: tc.apiKey}
			got := c.scrub(tc.input)
			if got != tc.want {
				t.Errorf("c.scrub(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
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
