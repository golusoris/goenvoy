package metadata_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/metadata"
)

const testPkg = "testpkg"

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return srv
}

func TestNewBaseClient(t *testing.T) {
	t.Parallel()

	c := metadata.NewBaseClient("http://example.com", testPkg)

	if c.BaseURL() != "http://example.com" {
		t.Fatalf("BaseURL = %q, want %q", c.BaseURL(), "http://example.com")
	}

	if c.UserAgent() != metadata.DefaultUserAgent {
		t.Fatalf("UserAgent = %q, want %q", c.UserAgent(), metadata.DefaultUserAgent)
	}

	if c.HTTPClient().Timeout != metadata.DefaultTimeout {
		t.Fatalf("Timeout = %v, want %v", c.HTTPClient().Timeout, metadata.DefaultTimeout)
	}
}

func TestNewBaseClient_Options(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []metadata.Option
		checkFn func(t *testing.T, c *metadata.BaseClient)
	}{
		{
			name: "WithHTTPClient",
			opts: []metadata.Option{metadata.WithHTTPClient(&http.Client{Timeout: 5 * time.Second})},
			checkFn: func(t *testing.T, c *metadata.BaseClient) {
				t.Helper()
				if c.HTTPClient().Timeout != 5*time.Second {
					t.Fatalf("Timeout = %v, want 5s", c.HTTPClient().Timeout)
				}
			},
		},
		{
			name: "WithTimeout",
			opts: []metadata.Option{metadata.WithTimeout(10 * time.Second)},
			checkFn: func(t *testing.T, c *metadata.BaseClient) {
				t.Helper()
				if c.HTTPClient().Timeout != 10*time.Second {
					t.Fatalf("Timeout = %v, want 10s", c.HTTPClient().Timeout)
				}
			},
		},
		{
			name: "WithUserAgent",
			opts: []metadata.Option{metadata.WithUserAgent("custom/1.0")},
			checkFn: func(t *testing.T, c *metadata.BaseClient) {
				t.Helper()
				if c.UserAgent() != "custom/1.0" {
					t.Fatalf("UserAgent = %q, want %q", c.UserAgent(), "custom/1.0")
				}
			},
		},
		{
			name: "WithBaseURL",
			opts: []metadata.Option{metadata.WithBaseURL("http://other.com")},
			checkFn: func(t *testing.T, c *metadata.BaseClient) {
				t.Helper()
				if c.BaseURL() != "http://other.com" {
					t.Fatalf("BaseURL = %q, want %q", c.BaseURL(), "http://other.com")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := metadata.NewBaseClient("http://default.com", testPkg, tt.opts...)
			tt.checkFn(t, c)
		})
	}
}

func TestSetAuth(t *testing.T) {
	t.Parallel()

	var gotHeader string

	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})

	c := metadata.NewBaseClient(srv.URL, testPkg)
	c.SetAuth(func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer token123")
	})

	_, _, err := c.DoRaw(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("DoRaw error: %v", err)
	}

	if gotHeader != "Bearer token123" {
		t.Fatalf("Authorization = %q, want %q", gotHeader, "Bearer token123")
	}
}

func TestDoRaw(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		body   io.Reader
	}{
		{"GET", http.MethodGet, nil},
		{"POST", http.MethodPost, strings.NewReader("payload")},
		{"PUT", http.MethodPut, strings.NewReader("data")},
		{"DELETE", http.MethodDelete, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotMethod, gotPath string

			srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path

				if r.Header.Get("Accept") != "application/json" {
					t.Errorf("Accept = %q, want application/json", r.Header.Get("Accept"))
				}

				if r.Header.Get("User-Agent") != metadata.DefaultUserAgent {
					t.Errorf("User-Agent = %q, want %q", r.Header.Get("User-Agent"), metadata.DefaultUserAgent)
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok":true}`))
			})

			c := metadata.NewBaseClient(srv.URL, testPkg)
			data, status, err := c.DoRaw(context.Background(), tt.method, "/api", tt.body)

			if err != nil {
				t.Fatalf("DoRaw error: %v", err)
			}

			if gotMethod != tt.method {
				t.Errorf("method = %q, want %q", gotMethod, tt.method)
			}

			if gotPath != "/api" {
				t.Errorf("path = %q, want /api", gotPath)
			}

			if status != http.StatusOK {
				t.Errorf("status = %d, want %d", status, http.StatusOK)
			}

			if string(data) != `{"ok":true}` {
				t.Errorf("body = %q, want %q", string(data), `{"ok":true}`)
			}
		})
	}
}

func TestDoRaw_NonOK(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	})

	c := metadata.NewBaseClient(srv.URL, testPkg)
	data, status, err := c.DoRaw(context.Background(), http.MethodGet, "/missing", nil)

	if err != nil {
		t.Fatalf("expected no error for non-2xx in DoRaw, got: %v", err)
	}

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", status, http.StatusNotFound)
	}

	if string(data) != "not found" {
		t.Errorf("body = %q, want %q", string(data), "not found")
	}
}

func TestDoRaw_NetworkError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := metadata.NewBaseClient(srv.URL, testPkg)
	_, _, err := c.DoRaw(context.Background(), http.MethodGet, "/fail", nil)

	if err == nil {
		t.Fatal("expected error for closed server, got nil")
	}

	if !strings.Contains(err.Error(), testPkg) {
		t.Errorf("error %q should contain package name %q", err.Error(), testPkg)
	}
}

func TestDoRawURL(t *testing.T) {
	t.Parallel()

	var gotPath string

	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept = %q, want application/json", r.Header.Get("Accept"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("raw-url-response"))
	})

	c := metadata.NewBaseClient("http://ignored.example.com", testPkg)
	data, status, err := c.DoRawURL(context.Background(), http.MethodGet, srv.URL+"/custom/path", nil)

	if err != nil {
		t.Fatalf("DoRawURL error: %v", err)
	}

	if gotPath != "/custom/path" {
		t.Errorf("path = %q, want /custom/path", gotPath)
	}

	if status != http.StatusOK {
		t.Errorf("status = %d, want %d", status, http.StatusOK)
	}

	if string(data) != "raw-url-response" {
		t.Errorf("body = %q, want %q", string(data), "raw-url-response")
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	type resp struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}

	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp{Name: "test", ID: 42})
	})

	c := metadata.NewBaseClient(srv.URL, testPkg)

	var got resp
	if err := c.Get(context.Background(), "/item", &got); err != nil {
		t.Fatalf("Get error: %v", err)
	}

	if got.Name != "test" {
		t.Errorf("Name = %q, want %q", got.Name, "test")
	}

	if got.ID != 42 {
		t.Errorf("ID = %d, want 42", got.ID)
	}
}

func TestGet_NonOK(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	})

	c := metadata.NewBaseClient(srv.URL, testPkg)

	var dst struct{}
	err := c.Get(context.Background(), "/denied", &dst)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *metadata.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusForbidden)
	}

	if apiErr.RawBody != "forbidden" {
		t.Errorf("RawBody = %q, want %q", apiErr.RawBody, "forbidden")
	}
}

func TestDoJSON(t *testing.T) {
	t.Parallel()

	type reqBody struct {
		Value string `json:"value"`
	}

	type respBody struct {
		Result string `json:"result"`
	}

	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("Content-Type = %q, want application/json prefix", ct)
		}

		var req reqBody
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Value != "hello" {
			t.Errorf("request Value = %q, want %q", req.Value, "hello")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(respBody{Result: "world"})
	})

	c := metadata.NewBaseClient(srv.URL, testPkg)

	var got respBody
	err := c.DoJSON(context.Background(), http.MethodPost, "/submit", reqBody{Value: "hello"}, &got)

	if err != nil {
		t.Fatalf("DoJSON error: %v", err)
	}

	if got.Result != "world" {
		t.Errorf("Result = %q, want %q", got.Result, "world")
	}
}

func TestDoJSON_Methods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
	}{
		{"PUT", http.MethodPut},
		{"PATCH", http.MethodPatch},
		{"DELETE", http.MethodDelete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotMethod string
			var gotBody []byte

			srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotBody, _ = io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true}`))
			})

			c := metadata.NewBaseClient(srv.URL, testPkg)

			type payload struct {
				Key string `json:"key"`
			}

			var dst map[string]bool
			err := c.DoJSON(context.Background(), tt.method, "/resource", payload{Key: "val"}, &dst)

			if err != nil {
				t.Fatalf("DoJSON error: %v", err)
			}

			if gotMethod != tt.method {
				t.Errorf("method = %q, want %q", gotMethod, tt.method)
			}

			if !bytes.Contains(gotBody, []byte(`"key"`)) {
				t.Errorf("body %q should contain key field", gotBody)
			}

			if !dst["ok"] {
				t.Errorf("dst[ok] = false, want true")
			}
		})
	}
}

func TestDoJSON_ErrorResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"BadRequest", http.StatusBadRequest, `{"error":"bad"}`},
		{"InternalServerError", http.StatusInternalServerError, "server error"},
		{"Unauthorized", http.StatusUnauthorized, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			})

			c := metadata.NewBaseClient(srv.URL, testPkg)

			var dst struct{}
			err := c.DoJSON(context.Background(), http.MethodGet, "/err", nil, &dst)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var apiErr *metadata.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *APIError, got %T", err)
			}

			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.statusCode)
			}

			if apiErr.RawBody != tt.body {
				t.Errorf("RawBody = %q, want %q", apiErr.RawBody, tt.body)
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  metadata.APIError
		want string
	}{
		{
			name: "with body",
			err:  metadata.APIError{StatusCode: 404, RawBody: "not found"},
			want: ": HTTP 404: not found",
		},
		{
			name: "without body",
			err:  metadata.APIError{StatusCode: 500, RawBody: ""},
			want: ": HTTP 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.Error()
			if !strings.HasSuffix(got, tt.want) {
				t.Errorf("Error() = %q, want suffix %q", got, tt.want)
			}

			if got == "" {
				t.Error("Error() should not be empty")
			}
		})
	}
}
