package spotify

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/metadata"
)

// TestNewResponseLimit checks the constructor's default response cap and base URL option.
func TestNewResponseLimit(t *testing.T) {
	t.Parallel()
	const baseURL = "https://spotify.example.test/v1"
	c := New("", metadata.WithBaseURL(baseURL))
	if got := c.maxResponseBytes; got != 32<<20 {
		t.Errorf("maxResponseBytes = %d, want 32 MiB", got)
	}
	if got := c.BaseURL(); got != baseURL {
		t.Errorf("BaseURL() = %q, want %q", got, baseURL)
	}
}

// TestGetResponseLimit checks success and error bodies around the local cap.
func TestGetResponseLimit(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		size   int
		status int
	}{
		{"under", 15, http.StatusOK},
		{"exact", 16, http.StatusOK},
		{"over", 17, http.StatusOK},
		{"error_within", 16, http.StatusBadRequest},
		{"error_over", 17, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkGetResponseLimit(t, tc.size, tc.status)
		})
	}
}

// checkGetResponseLimit checks response decoding around the configured cap.
func checkGetResponseLimit(t *testing.T, size, status int) {
	t.Helper()
	body := `{"id":"1"}`
	body += strings.Repeat(" ", size-len(body))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		if _, err := io.Copy(w, strings.NewReader(body)); err != nil {
			t.Errorf("spotify test: write response fixture: %v", err)
		}
	}))
	defer srv.Close()
	c := New("", metadata.WithBaseURL(srv.URL))
	c.SetMaxResponseBytes(16)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	artist, err := c.GetArtist(ctx, "1")
	if size > 16 {
		checkOversizedResponse(t, err, "limit 16 bytes")
		return
	}
	if status != http.StatusOK {
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != status || apiErr.Body != body {
			t.Errorf("error = %v, want APIError with status %d and original body", err, status)
		}
		return
	}
	if err != nil {
		t.Fatalf("GetArtist: %v", err)
	}
	if artist.ID != "1" {
		t.Errorf("artist ID = %q, want 1", artist.ID)
	}
}

// Why: keep oversized-body assertions shared across success and error responses.
func checkOversizedResponse(t *testing.T, err error, wantLimit string) {
	t.Helper()
	if !errors.Is(err, errResponseTooLarge) {
		t.Fatalf("error = %v, want errResponseTooLarge", err)
	}
	var apiErr *APIError
	if !strings.Contains(err.Error(), wantLimit) || errors.As(err, &apiErr) {
		t.Errorf("error = %v, want %s without APIError", err, wantLimit)
	}
}

// responseLimitCase describes a response-cap configuration and body size scenario.
type responseLimitCase struct {
	name      string
	zeroValue bool
	override  int64
	bodySize  int
	wantErr   string
	wantLimit int64
}

// TestSetMaxResponseBytes checks 1 MiB overrides, nonpositive overrides, and zero-value clients.
func TestSetMaxResponseBytes(t *testing.T) {
	t.Parallel()
	const oneMiB = 1 << 20
	for _, tc := range []responseLimitCase{
		{name: "override_1mib_rejects_2mib", override: oneMiB, bodySize: 2 << 20, wantErr: "limit 1048576 bytes", wantLimit: oneMiB},
		{name: "override_1mib_accepts_at_cap", override: oneMiB, bodySize: oneMiB, wantLimit: oneMiB},
		{name: "zero_override_keeps_default", override: 0, bodySize: 2 << 20, wantLimit: defaultMaxResponseBytes},
		{name: "negative_override_keeps_default", override: -1, bodySize: 64, wantLimit: defaultMaxResponseBytes},
		{name: "zero_value_client_without_new", zeroValue: true, bodySize: 64, wantLimit: defaultMaxResponseBytes},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkSetMaxResponseBytes(t, tc)
		})
	}
}

// checkSetMaxResponseBytes executes one response-limit scenario against an httptest server.
func checkSetMaxResponseBytes(t *testing.T, tc responseLimitCase) {
	t.Helper()
	body := `{"id":"1","name":"Spot"}`
	if tc.bodySize > len(body) {
		body += strings.Repeat(" ", tc.bodySize-len(body))
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := io.Copy(w, strings.NewReader(body)); err != nil && tc.wantErr == "" {
			t.Errorf("spotify test: write response fixture: %v", err)
		}
	}))
	defer srv.Close()
	c := newLimitTestClient(srv.URL, tc)
	if got := c.responseLimit(); got != tc.wantLimit {
		t.Errorf("responseLimit() = %d, want %d", got, tc.wantLimit)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	artist, err := c.GetArtist(ctx, "1")
	assertArtistResponse(t, artist, err, tc.wantErr)
}

// assertArtistResponse checks the decoded artist or oversized response error.
func assertArtistResponse(t *testing.T, artist *Artist, err error, wantErr string) {
	t.Helper()
	if wantErr != "" {
		checkOversizedResponse(t, err, wantErr)
		if artist != nil {
			t.Errorf("artist = %v, want nil on oversized response", artist)
		}
		return
	}
	if err != nil {
		t.Fatalf("GetArtist: %v", err)
	}
	if artist.ID != "1" || artist.Name != "Spot" {
		t.Errorf("artist = %+v, want ID 1 and Name Spot", artist)
	}
}

// newLimitTestClient constructs either a zero-value Client or a configured Client.
func newLimitTestClient(baseURL string, tc responseLimitCase) *Client {
	if tc.zeroValue {
		return &Client{BaseClient: metadata.NewBaseClient(baseURL, "spotify")}
	}
	c := New("", metadata.WithBaseURL(baseURL))
	c.SetMaxResponseBytes(tc.override)
	return c
}
