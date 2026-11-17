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
	c.maxResponseBytes = 16
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	artist, err := c.GetArtist(ctx, "1")
	if size > 16 {
		checkOversizedResponse(t, err)
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
func checkOversizedResponse(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, errResponseTooLarge) {
		t.Fatalf("error = %v, want errResponseTooLarge", err)
	}
	var apiErr *APIError
	if !strings.Contains(err.Error(), "limit 16 bytes") || errors.As(err, &apiErr) {
		t.Errorf("error = %v, want limit 16 bytes without APIError", err)
	}
}
