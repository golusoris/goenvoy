package jackett

import (
	"errors"
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
	if strings.Contains(got, "K2") {
		t.Errorf("redactURL(%q) = %q; contains K2", raw, got)
	}
	if strings.Contains(got, "?") {
		t.Errorf("redactURL(%q) = %q; contains ?", raw, got)
	}
}

// TestRedactURLErrorNonURLError verifies that non-url.Error values and nil are returned unchanged.
func TestRedactURLErrorNonURLError(t *testing.T) {
	t.Parallel()

	plain := errors.New("boom")
	got := redactURLError(plain)
	if got != plain {
		t.Errorf("redactURLError(plain) = %v, want same pointer %v", got, plain)
	}
	if !errors.Is(got, plain) {
		t.Errorf("errors.Is(got, plain) = false, want true")
	}

	if gotNil := redactURLError(nil); gotNil != nil {
		t.Errorf("redactURLError(nil) = %v, want nil", gotNil)
	}
}
