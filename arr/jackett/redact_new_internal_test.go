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
	const key = "plain-key"
	cases := []struct{ name, baseURL string }{
		{name: "scheme and host", baseURL: "ftp://plain-key"},
		{name: "parse failure", baseURL: "http://h/%zz/plain-key"},
		{name: "parse failure in host", baseURL: "http://plain-key:80x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := New(tc.baseURL, key)
			if err == nil {
				t.Fatalf("New(%q) = %v, nil; want error", tc.baseURL, c)
			}
			assertNoKeyForms(t, err.Error(), key)
			assertNoKeyForms(t, fmt.Sprintf("%+v %#v", err, err), key)
			for i, link := 0, err; link != nil && i < maxErrNodes; i, link = i+1, errors.Unwrap(link) {
				assertNoKeyForms(t, link.Error(), key)
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
