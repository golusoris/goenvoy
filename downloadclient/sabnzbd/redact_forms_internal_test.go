package sabnzbd

import (
	"context"
	"errors"
	"fmt"
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

// TestAPIErrorJSONFormRedaction checks that a response holding the JSON-escaped key is scrubbed, as a status body and as an application error.
func TestAPIErrorJSONFormRedaction(t *testing.T) {
	t.Parallel()
	const key = "fixture<key"
	const escaped = `fixture\u003ckey`
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{name: "status body", status: http.StatusInternalServerError, body: `{"detail":"` + escaped + `"}`},
		{name: "application error", status: http.StatusOK, body: `{"status":false,"error":"` + escaped + `"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkJSONFormResponse(t, key, escaped, tc.status, tc.body)
		})
	}
}

// checkJSONFormResponse serves body with status and checks that the GetQueue error holds no key form.
func checkJSONFormResponse(t *testing.T, key, escaped string, status int, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		if _, err := fmt.Fprint(w, body); err != nil {
			t.Errorf("sabnzbd test: write JSON form response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	c, err := New(srv.URL, key, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	_, err = c.GetQueue(ctx)
	if err == nil {
		t.Fatal("GetQueue error = nil; want an error")
	}
	if strings.Contains(err.Error(), key) || strings.Contains(err.Error(), escaped) {
		t.Errorf("key leaked: %s", err.Error())
	}
	var apiErr *APIError
	if status == http.StatusOK && !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %v", err)
	}
}
