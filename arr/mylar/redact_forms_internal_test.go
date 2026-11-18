package mylar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// TestRedactErrorRootIdentity checks that a root URL error is copied only when sanitizing changes it.
func TestRedactErrorRootIdentity(t *testing.T) {
	t.Parallel()
	const key = "plain-key"
	sentinel := errors.New("fixture transport failure")
	cases := []struct {
		name string
		uerr *url.Error
		keep bool
	}{
		{name: "keyless", uerr: &url.Error{Op: "Get", URL: "http://example.invalid/path?q=1", Err: sentinel}, keep: true},
		{name: "key in URL", uerr: &url.Error{Op: "Get", URL: "http://example.invalid/?apikey=" + key, Err: sentinel}},
		{name: "key in cause", uerr: &url.Error{Op: "Get", URL: "http://example.invalid/", Err: fmt.Errorf("fixture %s: %w", key, sentinel)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertRootIdentity(t, &Client{apiKey: key}, tc.uerr, sentinel, tc.keep)
		})
	}
}

// assertRootIdentity checks the identity, cause, and text of a redacted root URL error.
func assertRootIdentity(t *testing.T, c *Client, uerr *url.Error, sentinel error, keep bool) {
	t.Helper()
	got := c.redactError(uerr)
	if same := any(got) == any(uerr); same != keep {
		t.Errorf("same error value = %v; want %v", same, keep)
	}
	if keep && !errors.Is(got, uerr) {
		t.Error("errors.Is lost the original root URL error")
	}
	if !errors.Is(got, sentinel) {
		t.Error("errors.Is lost the sentinel cause")
	}
	if c.containsKey(got.Error()) {
		t.Errorf("key form leaked: %s", got.Error())
	}
}

// TestScrubJSONForm checks that the HTML-escaped JSON form of the API key is scrubbed and detected.
func TestScrubJSONForm(t *testing.T) {
	t.Parallel()
	const key = "fixture<key"
	const escaped = `fixture\u003ckey`
	cases := []struct{ name, text string }{
		{name: "json escaped", text: `{"error":"` + escaped + `"}`},
		{name: "raw", text: `{"error":"` + key + `"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{apiKey: key}
			got := c.scrub(tc.text)
			if strings.Contains(got, key) || strings.Contains(got, escaped) {
				t.Errorf("scrub kept the key: %s", got)
			}
			if !strings.Contains(got, "REDACTED") {
				t.Errorf("scrub = %q; want REDACTED marker", got)
			}
			if !c.containsKey(tc.text) {
				t.Errorf("containsKey(%q) = false; want true", tc.text)
			}
		})
	}
}

// TestAPIErrorJSONFormRedaction checks that a status body holding the JSON-escaped key is scrubbed.
func TestAPIErrorJSONFormRedaction(t *testing.T) {
	t.Parallel()
	const key = "fixture<key"
	const escaped = `fixture\u003ckey`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.Copy(w, strings.NewReader(`{"error":"`+escaped+`"}`)); err != nil {
			t.Errorf("mylar test: write JSON form response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	c, err := New(srv.URL, key, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var out any
	err = c.get(ctx, "fixture", nil, &out)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	for _, text := range []string{apiErr.Body, apiErr.Error()} {
		if strings.Contains(text, key) || strings.Contains(text, escaped) {
			t.Errorf("key leaked: %s", text)
		}
	}
}

// TestScrubPercentVariants checks that lower-case hex and path-escaped spellings of the API key are scrubbed and detected.
func TestScrubPercentVariants(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, key, text, want string }{
		{name: "lower-case hex", key: "fixture+k/y=1", text: "apikey=fixture%2bk%2fy%3d1", want: "apikey=REDACTED"},
		{name: "upper-case hex", key: "fixture+k/y=1", text: "apikey=fixture%2Bk%2Fy%3D1", want: "apikey=REDACTED"},
		{name: "path-escaped space", key: "fixture key", text: "fixture%20key", want: "REDACTED"},
		{name: "query-escaped space", key: "fixture key", text: "fixture+key", want: "REDACTED"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertScrubbed(t, &Client{apiKey: tc.key}, tc.text, tc.want)
		})
	}
}

// assertScrubbed checks the scrubbed text, the absence of every key form in it, and detection in the input.
func assertScrubbed(t *testing.T, c *Client, text, want string) {
	t.Helper()
	got := c.scrub(text)
	if got != want {
		t.Errorf("scrub(%q) = %q; want %q", text, got, want)
	}
	for _, form := range c.keyForms() {
		if strings.Contains(got, form) {
			t.Errorf("scrub kept key form %q: %s", form, got)
		}
	}
	if !c.containsKey(text) {
		t.Errorf("containsKey(%q) = false; want true", text)
	}
	if c.containsKey(got) {
		t.Errorf("containsKey(%q) = true; want false", got)
	}
}

// TestKeyFormsDistinct checks that a key whose escaped forms all equal the raw key yields one form.
func TestKeyFormsDistinct(t *testing.T) {
	t.Parallel()
	forms := (&Client{apiKey: "Plain-Key"}).keyForms()
	if len(forms) != 1 || forms[0] != "Plain-Key" {
		t.Errorf("keyForms() = %q; want [Plain-Key]", forms)
	}
}

// TestAPIErrorPercentVariantRedaction checks that a status body holding a lower- or upper-case percent-hex key is scrubbed.
func TestAPIErrorPercentVariantRedaction(t *testing.T) {
	t.Parallel()
	const key = "fixture+k/y=1"
	cases := []struct{ name, spelling string }{
		{name: "lower-case hex", spelling: "fixture%2bk%2fy%3d1"},
		{name: "upper-case hex", spelling: "fixture%2Bk%2Fy%3D1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			apiErr := percentFormAPIError(t, key, tc.spelling)
			for _, text := range []string{apiErr.Body, apiErr.Error()} {
				if strings.Contains(text, key) || strings.Contains(text, tc.spelling) {
					t.Errorf("key leaked: %s", text)
				}
			}
		})
	}
}

// percentFormAPIError returns the APIError for a 500 response whose body echoes spelling.
func percentFormAPIError(t *testing.T, key, spelling string) *APIError {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.Copy(w, strings.NewReader("apikey="+spelling)); err != nil {
			t.Errorf("mylar test: write percent form response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	c, err := New(srv.URL, key, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var out any
	err = c.get(ctx, "fixture", nil, &out)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	return apiErr
}

// temporaryCause is a fixture error whose Temporary method returns a fixed value.
type temporaryCause struct{ temporary bool }

func (temporaryCause) Error() string { return "fixture cause" }

func (e temporaryCause) Temporary() bool { return e.temporary }

// TestRedactedErrorTemporary checks that the wrapper forwards Temporary from the hidden error.
func TestRedactedErrorTemporary(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		cause error
		want  bool
	}{
		{name: "temporary cause", cause: temporaryCause{temporary: true}, want: true},
		{name: "permanent cause", cause: temporaryCause{temporary: false}, want: false},
		{name: "wrapped temporary cause", cause: fmt.Errorf("fixture wrap: %w", temporaryCause{temporary: true}), want: true},
		{name: "cause without method", cause: errors.New("fixture plain"), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e := &redactedError{msg: "fixture", err: tc.cause, c: &Client{apiKey: "fixture-key"}}
			if got := e.Temporary(); got != tc.want {
				t.Errorf("Temporary() = %v; want %v", got, tc.want)
			}
		})
	}
}
