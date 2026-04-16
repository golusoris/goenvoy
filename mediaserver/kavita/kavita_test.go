package kavita

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c, err := New(ts.URL, "test-api-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c.token = "test-token"
	return c
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/Plugin/authenticate" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["apiKey"] != "test-api-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"token": "jwt-token-123"})
	}))
	defer ts.Close()

	c, err := New(ts.URL, "test-api-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.Authenticate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if c.token != "jwt-token-123" {
		t.Fatalf("unexpected token: %s", c.token)
	}
}

func TestAutoAuthenticate(t *testing.T) {
	t.Parallel()

	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/Plugin/authenticate" {
			calls++
			json.NewEncoder(w).Encode(map[string]string{"token": "auto-token"})
			return
		}
		if r.Header.Get("Authorization") != "Bearer auto-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode([]Library{{ID: 1, Name: "Manga"}})
	}))
	defer ts.Close()

	c, err := New(ts.URL, "test-api-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	libs, err := c.GetLibraries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 auth call, got %d", calls)
	}
	if len(libs) != 1 || libs[0].Name != "Manga" {
		t.Fatalf("unexpected libraries: %+v", libs)
	}
}

func TestAuthorizationHeader(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode([]Library{{ID: 1, Name: "Comics"}})
	})

	libs, err := c.GetLibraries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(libs) != 1 || libs[0].Name != "Comics" {
		t.Fatalf("unexpected libraries: %+v", libs)
	}
}

func TestGetLibraries(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]Library{{ID: 1, Name: "Comics"}, {ID: 2, Name: "Manga"}})
	})

	libs, err := c.GetLibraries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(libs) != 2 || libs[0].Name != "Comics" {
		t.Fatalf("unexpected libraries: %+v", libs)
	}
}

func TestGetLibrary(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(Library{ID: 1, Name: "Manga"})
	})

	lib, err := c.GetLibrary(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if lib.Name != "Manga" {
		t.Fatalf("unexpected library: %+v", lib)
	}
}

func TestScanLibrary(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if err := c.ScanLibrary(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
}

func TestGetSeries(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]Series{{ID: 1, Name: "One Piece", LibraryID: 1}})
	})

	series, err := c.GetSeries(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(series) != 1 || series[0].Name != "One Piece" {
		t.Fatalf("unexpected series: %+v", series)
	}
}

func TestGetOneSeries(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(Series{ID: 1, Name: "Naruto"})
	})

	s, err := c.GetOneSeries(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "Naruto" {
		t.Fatalf("unexpected series: %+v", s)
	}
}

func TestGetVolumes(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]Volume{{ID: 1, Name: "Volume 1", SeriesID: 1}})
	})

	vols, err := c.GetVolumes(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(vols) != 1 || vols[0].Name != "Volume 1" {
		t.Fatalf("unexpected volumes: %+v", vols)
	}
}

func TestGetChapter(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(Chapter{ID: 1, Title: "Chapter 1", Number: "1"})
	})

	ch, err := c.GetChapter(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Title != "Chapter 1" {
		t.Fatalf("unexpected chapter: %+v", ch)
	}
}

func TestGetCollections(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]Collection{{ID: 1, Title: "Favorites"}})
	})

	cols, err := c.GetCollections(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 1 || cols[0].Title != "Favorites" {
		t.Fatalf("unexpected collections: %+v", cols)
	}
}

func TestGetReadingLists(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]ReadingList{{ID: 1, Title: "To Read"}})
	})

	rls, err := c.GetReadingLists(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rls) != 1 || rls[0].Title != "To Read" {
		t.Fatalf("unexpected reading lists: %+v", rls)
	}
}

func TestGetUsers(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]User{{ID: 1, Username: "admin", Email: "admin@example.com"}})
	})

	users, err := c.GetUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].Username != "admin" {
		t.Fatalf("unexpected users: %+v", users)
	}
}

func TestGetServerInfo(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(ServerInfo{Version: "0.7.0", Os: "linux"})
	})

	info, err := c.GetServerInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "0.7.0" {
		t.Fatalf("unexpected server info: %+v", info)
	}
}

func TestSearch(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(SearchResult{
			Series: []Series{{ID: 1, Name: "Bleach"}},
		})
	})

	sr, err := c.Search(context.Background(), "Bleach")
	if err != nil {
		t.Fatal(err)
	}
	if len(sr.Series) != 1 || sr.Series[0].Name != "Bleach" {
		t.Fatalf("unexpected search result: %+v", sr)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	})

	_, err := c.GetLibraries(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Fatalf("unexpected status code: %d", apiErr.StatusCode)
	}
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	custom := &http.Client{}
	c, err := New("http://localhost", "key", WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.http != custom {
		t.Fatal("custom HTTP client not set")
	}
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer ts.Close()

	c, err := New(ts.URL, "key", WithUserAgent("myapp/1.2.3"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c.token = "t"
	if _, err := c.GetLibraries(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotUA != "myapp/1.2.3" {
		t.Errorf("User-Agent = %q, want %q", gotUA, "myapp/1.2.3")
	}
}

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
			c, err := New(tc.url, "k")
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}
