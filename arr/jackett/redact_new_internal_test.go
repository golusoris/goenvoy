package jackett

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewRedactsBaseURLKey checks that constructor errors hold no form of the API key.
func TestNewRedactsBaseURLKey(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, baseURL, key, forbid string }{
		{name: "scheme and host", baseURL: "ftp://plain-key", key: "plain-key", forbid: "plain-key"},
		{name: "parse failure", baseURL: "http://h/%zz/plain-key", key: "plain-key", forbid: "plain-key"},
		{name: "parse failure in host", baseURL: "http://plain-key:80x", key: "plain-key", forbid: "plain-key"},
		{name: "foreign query key on scheme failure", baseURL: "ftp://example.invalid/?apikey=other", key: "k", forbid: "other"},
		{name: "foreign query key on parse failure", baseURL: "http://%zz/?apikey=other", key: "", forbid: "other"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := New(tc.baseURL, tc.key)
			if err == nil {
				t.Fatalf("New(%q) = %v, nil; want error", tc.baseURL, c)
			}
			assertNoKeyForms(t, err.Error(), tc.forbid)
			assertNoKeyForms(t, fmt.Sprintf("%+v %#v", err, err), tc.forbid)
			for i, link := 0, err; link != nil && i < maxErrNodes; i, link = i+1, errors.Unwrap(link) {
				assertNoKeyForms(t, link.Error(), tc.forbid)
			}
		})
	}
}

// TestCarriesKeyNil checks that a nil error never reports a key or panics.
func TestCarriesKeyNil(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, key string
	}{
		{name: "keyed client", key: "k"},
		{name: "empty key", key: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{apiKey: tc.key}
			if c.carriesKey(nil) {
				t.Error("carriesKey(nil) = true; want false")
			}
			if got := c.redactError(nil); got != nil {
				t.Errorf("redactError(nil) = %v; want nil", got)
			}
			if got := c.scrubError(nil); got != nil {
				t.Errorf("scrubError(nil) = %v; want nil", got)
			}
		})
	}
}

// statusRewriteTransport replaces the fixture response status line.
type statusRewriteTransport struct {
	http.RoundTripper
	status string
}

// RoundTrip answers with the wrapped transport's response under the replaced status line.
func (tr statusRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := tr.RoundTripper.RoundTrip(req)
	if err == nil {
		resp.StatusCode = http.StatusInternalServerError
		resp.Status = tr.status
	}
	return resp, err
}

// TestAPIErrorStatusRedaction checks that a status line holding the API key reaches APIError without the key.
func TestAPIErrorStatusRedaction(t *testing.T) {
	t.Parallel()
	const key = "plain-key"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	hc := srv.Client()
	hc.Transport = statusRewriteTransport{RoundTripper: hc.Transport, status: "500 " + key}
	c, err := New(srv.URL, key, WithHTTPClient(hc))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	_, err = c.GetIndexers(ctx)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	assertNoKeyForms(t, apiErr.Status, key)
	assertNoKeyForms(t, apiErr.Error(), key)
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d; want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}
