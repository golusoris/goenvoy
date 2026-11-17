package spotify

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golusoris/goenvoy/metadata"
)

// TestNewResponseLimit checks the constructor's default response cap.
func TestNewResponseLimit(t *testing.T) {
	t.Parallel()
	if got := New("").maxResponseBytes; got != 32<<20 {
		t.Errorf("maxResponseBytes = %d, want 32 MiB", got)
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

// Why: keep HTTP setup and response assertions outside the case matrix.
func checkGetResponseLimit(t *testing.T, size, status int) {
	t.Helper()
	body := `{"id":"1"}`
	body += strings.Repeat(" ", size-len(body))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = io.Copy(w, strings.NewReader(body))
	}))
	defer srv.Close()
	c := New("", metadata.WithBaseURL(srv.URL))
	c.maxResponseBytes = 16
	artist, err := c.GetArtist(t.Context(), "1")
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
