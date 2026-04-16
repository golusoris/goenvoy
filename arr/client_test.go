package arr_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/arr/v2"
)

func TestNewBaseClient(t *testing.T) {
	t.Parallel()

	t.Run("valid URL", func(t *testing.T) {
		t.Parallel()
		c, err := arr.NewBaseClient("http://localhost:8989", "test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("trailing slash stripped", func(t *testing.T) {
		t.Parallel()
		_, err := arr.NewBaseClient("http://localhost:8989/", "test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with options", func(t *testing.T) {
		t.Parallel()
		custom := &http.Client{}
		_, err := arr.NewBaseClient(
			"http://localhost:8989",
			"test-key",
			arr.WithHTTPClient(custom),
			arr.WithUserAgent("custom/1.0"),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestBaseClient_Get(t *testing.T) {
	t.Parallel()

	want := arr.StatusResponse{
		AppName: "Sonarr",
		Version: "4.0.0",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("missing or wrong API key header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := arr.NewBaseClient(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got arr.StatusResponse
	if err := c.Get(context.Background(), "/api/v3/system/status", &got); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.AppName != want.AppName {
		t.Errorf("AppName = %q, want %q", got.AppName, want.AppName)
	}
	if got.Version != want.Version {
		t.Errorf("Version = %q, want %q", got.Version, want.Version)
	}
}

func TestBaseClient_Post(t *testing.T) {
	t.Parallel()

	wantCmd := arr.CommandResponse{
		ID:   1,
		Name: "RefreshSeries",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		var cmd arr.CommandRequest
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if cmd.Name != "RefreshSeries" {
			t.Errorf("command name = %q, want RefreshSeries", cmd.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wantCmd)
	}))
	defer srv.Close()

	c, err := arr.NewBaseClient(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got arr.CommandResponse
	err = c.Post(
		context.Background(),
		"/api/v3/command",
		arr.CommandRequest{Name: "RefreshSeries"},
		&got,
	)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}

	if got.Name != wantCmd.Name {
		t.Errorf("Name = %q, want %q", got.Name, wantCmd.Name)
	}
}

func TestBaseClient_ErrorResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	c, err := arr.NewBaseClient(srv.URL, "bad-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var dst arr.StatusResponse
	err = c.Get(context.Background(), "/api/v3/system/status", &dst)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}

	var apiErr *arr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *arr.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestBaseClient_Delete(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := arr.NewBaseClient(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.Delete(context.Background(), "/api/v3/series/1", nil, nil); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := arr.NewBaseClient(srv.URL, "test-key", arr.WithTimeout(50*time.Millisecond))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.Head(context.Background(), "/api/v3/health")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "Client.Timeout") && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("expected timeout-related error, got: %v", err)
	}
}

func TestBaseClient_Put(t *testing.T) {
	t.Parallel()

	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	want := item{ID: 42, Name: "updated"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		var got item
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if got.Name != "updated" {
			t.Errorf("Name = %q, want %q", got.Name, "updated")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := arr.NewBaseClient(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got item
	err = c.Put(context.Background(), "/api/v3/series/42", item{ID: 42, Name: "updated"}, &got)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %d, want %d", got.ID, want.ID)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
}

func TestBaseClient_Patch(t *testing.T) {
	t.Parallel()

	type patch struct {
		Monitored bool `json:"monitored"`
	}

	want := patch{Monitored: true}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		var got patch
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(got)
	}))
	defer srv.Close()

	c, err := arr.NewBaseClient(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got patch
	err = c.Patch(context.Background(), "/api/v3/series/1", want, &got)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if got.Monitored != want.Monitored {
		t.Errorf("Monitored = %v, want %v", got.Monitored, want.Monitored)
	}
}

func TestBaseClient_Head(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodHead {
				t.Errorf("expected HEAD, got %s", r.Method)
			}
			if r.Header.Get("X-Api-Key") != "test-key" {
				t.Error("missing or wrong API key header")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c, err := arr.NewBaseClient(srv.URL, "test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := c.Head(context.Background(), "/api/v3/health"); err != nil {
			t.Fatalf("Head failed: %v", err)
		}
	})

	t.Run("error status", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		c, err := arr.NewBaseClient(srv.URL, "test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = c.Head(context.Background(), "/api/v3/missing")
		if err == nil {
			t.Fatal("expected error for 404 response")
		}

		var apiErr *arr.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *arr.APIError, got %T", err)
		}
		if apiErr.StatusCode != http.StatusNotFound {
			t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
		}
	})
}

func TestBaseClient_GetRaw(t *testing.T) {
	t.Parallel()

	rawContent := "line1\nline2\nline3"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Error("missing or wrong API key header")
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(rawContent))
	}))
	defer srv.Close()

	c, err := arr.NewBaseClient(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetRaw(context.Background(), "/api/v3/log/file")
	if err != nil {
		t.Fatalf("GetRaw failed: %v", err)
	}
	if string(got) != rawContent {
		t.Errorf("body = %q, want %q", string(got), rawContent)
	}
	if len(got) != len(rawContent) {
		t.Errorf("length = %d, want %d", len(got), len(rawContent))
	}
}

func TestBaseClient_Upload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Errorf("Content-Type = %q, want multipart/form-data", ct)
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("FormFile failed: %v", err)
		}
		defer file.Close()

		if header.Filename != "backup.zip" {
			t.Errorf("filename = %q, want %q", header.Filename, "backup.zip")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := arr.NewBaseClient(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := strings.NewReader("fake-zip-content")
	err = c.Upload(context.Background(), "/api/v3/system/backup/restore", "file", "backup.zip", data)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
}

func TestAPIError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  arr.APIError
		want string
	}{
		{
			name: "with body",
			err: arr.APIError{
				StatusCode: 401,
				Method:     "GET",
				Path:       "/api/v3/system/status",
				Body:       []byte(`{"message":"Unauthorized"}`),
			},
			want: `arr: GET /api/v3/system/status returned 401: {"message":"Unauthorized"}`,
		},
		{
			name: "empty body",
			err: arr.APIError{
				StatusCode: 500,
				Method:     "POST",
				Path:       "/api/v3/command",
				Body:       nil,
			},
			want: "arr: POST /api/v3/command returned 500: ",
		},
		{
			name: "status code only",
			err: arr.APIError{
				StatusCode: 404,
			},
			want: "arr:   returned 404: ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.err.Error()
			if got != tc.want {
				t.Errorf("Error() = %q, want %q", got, tc.want)
			}
			if !strings.Contains(got, "arr:") {
				t.Error("expected error string to contain 'arr:' prefix")
			}
		})
	}
}

// FuzzNewBaseClient guards the NewBaseClient constructor — the one
// user-controlled seam in the whole *arr monorepo — against panics on
// arbitrary baseURL input. url.Parse is surprisingly tolerant and it
// would be easy for downstream callers to feed us a value we don't
// expect; treat any panic here as a bug.
func FuzzNewBaseClient(f *testing.F) {
	f.Add("http://localhost:8989")
	f.Add("https://example.com/sonarr/")
	f.Add("")
	f.Add("://not-a-url")
	f.Add("http://[::1]:9999")
	f.Fuzz(func(_ *testing.T, baseURL string) {
		_, _ = arr.NewBaseClient(baseURL, "k")
	})
}
