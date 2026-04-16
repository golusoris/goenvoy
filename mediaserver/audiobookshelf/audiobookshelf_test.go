package audiobookshelf_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/mediaserver/audiobookshelf"
)

func newTestServer(t *testing.T, wantPath, wantToken string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+wantToken {
			t.Errorf("Authorization = %q, want Bearer %s", got, wantToken)
		}
		w.Header().Set("Content-Type", "application/json")
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
}

func TestGetLibraries(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/libraries", "test-token", map[string]any{
		"libraries": []map[string]any{
			{"id": "lib1", "name": "Audiobooks", "mediaType": "book"},
		},
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	libs, err := c.GetLibraries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(libs) != 1 {
		t.Fatalf("len(libraries) = %d, want 1", len(libs))
	}
	if libs[0].Name != "Audiobooks" {
		t.Errorf("Name = %q, want Audiobooks", libs[0].Name)
	}
}

func TestGetLibrary(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/libraries/lib1", "test-token", map[string]any{
		"id": "lib1", "name": "Audiobooks", "mediaType": "book",
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	lib, err := c.GetLibrary(context.Background(), "lib1")
	if err != nil {
		t.Fatal(err)
	}
	if lib.MediaType != "book" {
		t.Errorf("MediaType = %q, want book", lib.MediaType)
	}
}

func TestGetLibraryItems(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/libraries/lib1/items", "test-token", map[string]any{
		"results": []map[string]any{
			{"id": "item1", "libraryId": "lib1"},
		},
		"total": 1,
		"limit": 20,
		"page":  0,
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	resp, err := c.GetLibraryItems(context.Background(), "lib1", 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 {
		t.Errorf("Total = %d, want 1", resp.Total)
	}
}

func TestGetItem(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/items/item1", "test-token", map[string]any{
		"id": "item1", "libraryId": "lib1",
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	item, err := c.GetItem(context.Background(), "item1")
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "item1" {
		t.Errorf("ID = %q, want item1", item.ID)
	}
}

func TestGetUsers(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/users", "test-token", []map[string]any{
		{"id": "user1", "username": "admin", "type": "root"},
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	users, err := c.GetUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 {
		t.Fatalf("len(users) = %d, want 1", len(users))
	}
	if users[0].Username != "admin" {
		t.Errorf("Username = %q, want admin", users[0].Username)
	}
}

func TestGetMe(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/me", "test-token", map[string]any{
		"id": "user1", "username": "admin", "type": "root",
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	user, err := c.GetMe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user.Type != "root" {
		t.Errorf("Type = %q, want root", user.Type)
	}
}

func TestGetServerInfo(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/server", "test-token", map[string]any{
		"version": "2.5.0",
		"isInit":  true,
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	info, err := c.GetServerInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "2.5.0" {
		t.Errorf("Version = %q, want 2.5.0", info.Version)
	}
}

func TestGetSessions(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/sessions", "test-token", map[string]any{
		"sessions": []map[string]any{
			{"id": "s1", "userId": "u1", "displayTitle": "Test Book"},
		},
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sessions, err := c.GetSessions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Forbidden"))
	}))
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "bad-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetLibraries(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *audiobookshelf.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusForbidden)
	}
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	called := false
	custom := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			called = true
			return http.DefaultTransport.RoundTrip(r)
		}),
	}

	ts := newTestServer(t, "/api/libraries", "k", map[string]any{"libraries": []any{}})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "k", audiobookshelf.WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, _ = c.GetLibraries(context.Background())
	if !called {
		t.Error("custom HTTP client was not used")
	}
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"libraries": []any{}})
	}))
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "k", audiobookshelf.WithUserAgent("myapp/1.2.3"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := c.GetLibraries(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotUA != "myapp/1.2.3" {
		t.Errorf("User-Agent = %q, want %q", gotUA, "myapp/1.2.3")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNew_invalidURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, url string
	}{
		{"empty", ""},
		{"malformed", "://x"},
		{"ftp", "ftp://x"},
		{"no-scheme", "no-scheme"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := audiobookshelf.New(tc.url, "k")
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}
