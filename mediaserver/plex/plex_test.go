package plex_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/mediaserver/plex"
)

type testResponse struct {
	MediaContainer any `json:"MediaContainer"`
}

func newTestServer(t *testing.T, wantPath, wantToken string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.Header.Get("X-Plex-Token"); got != wantToken {
			t.Errorf("X-Plex-Token = %q, want %q", got, wantToken)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want application/json", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResponse{MediaContainer: response})
	}))
}

func TestGetIdentity(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/identity" {
			t.Errorf("path = %q, want /identity", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResponse{MediaContainer: plex.Identity{
			MachineIdentifier: "abc123",
			Version:           "1.32.0",
		}})
	}))
	defer ts.Close()

	c, err := plex.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	id, err := c.GetIdentity(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if id.MachineIdentifier != "abc123" {
		t.Errorf("MachineIdentifier = %q, want abc123", id.MachineIdentifier)
	}
	if id.Version != "1.32.0" {
		t.Errorf("Version = %q, want 1.32.0", id.Version)
	}
}

func TestGetLibraries(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/library/sections", "lib-token", plex.MediaContainer{
		Directory: []plex.Directory{
			{Key: "/library/sections/1", Title: "Movies", Type: "movie"},
			{Key: "/library/sections/2", Title: "TV Shows", Type: "show"},
		},
	})
	defer ts.Close()

	c, err := plex.New(ts.URL, "lib-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	libs, err := c.GetLibraries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(libs) != 2 {
		t.Fatalf("len = %d, want 2", len(libs))
	}
	if libs[0].Title != "Movies" {
		t.Errorf("Title = %q, want Movies", libs[0].Title)
	}
	if libs[1].Type != "show" {
		t.Errorf("Type = %q, want show", libs[1].Type)
	}
}

func TestGetLibraryContents(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/library/sections/1/all" {
			t.Errorf("path = %q, want /library/sections/1/all", r.URL.Path)
		}
		if got := r.URL.Query().Get("X-Plex-Container-Start"); got != "0" {
			t.Errorf("start = %q, want 0", got)
		}
		if got := r.URL.Query().Get("X-Plex-Container-Size"); got != "50" {
			t.Errorf("size = %q, want 50", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResponse{MediaContainer: plex.MediaContainer{
			Size:      2,
			TotalSize: 100,
			Metadata: []plex.Metadata{
				{RatingKey: "1", Title: "Inception", Year: 2010},
				{RatingKey: "2", Title: "Interstellar", Year: 2014},
			},
		}})
	}))
	defer ts.Close()

	c, err := plex.New(ts.URL, "content-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	mc, err := c.GetLibraryContents(context.Background(), "1", 0, 50)
	if err != nil {
		t.Fatal(err)
	}
	if mc.TotalSize != 100 {
		t.Errorf("TotalSize = %d, want 100", mc.TotalSize)
	}
	if len(mc.Metadata) != 2 {
		t.Fatalf("len = %d, want 2", len(mc.Metadata))
	}
	if mc.Metadata[0].Title != "Inception" {
		t.Errorf("Title = %q, want Inception", mc.Metadata[0].Title)
	}
}

func TestGetMetadata(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/library/metadata/1244", "meta-token", plex.MediaContainer{
		Metadata: []plex.Metadata{{
			RatingKey: "1244",
			Title:     "Riddick",
			Year:      2013,
			Rating:    5.7,
			Duration:  7607642,
			Media: []plex.Media{{
				VideoCodec: "h264",
				AudioCodec: "aac",
				Bitrate:    3557,
			}},
		}},
	})
	defer ts.Close()

	c, err := plex.New(ts.URL, "meta-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m, err := c.GetMetadata(context.Background(), "1244")
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "Riddick" {
		t.Errorf("Title = %q, want Riddick", m.Title)
	}
	if m.Year != 2013 {
		t.Errorf("Year = %d, want 2013", m.Year)
	}
	if len(m.Media) != 1 {
		t.Fatalf("len(Media) = %d, want 1", len(m.Media))
	}
	if m.Media[0].VideoCodec != "h264" {
		t.Errorf("VideoCodec = %q, want h264", m.Media[0].VideoCodec)
	}
}

func TestGetMetadataNotFound(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/library/metadata/99999", "nf-token", plex.MediaContainer{
		Metadata: []plex.Metadata{},
	})
	defer ts.Close()

	c, err := plex.New(ts.URL, "nf-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetMetadata(context.Background(), "99999")
	if err == nil {
		t.Fatal("expected error for empty metadata")
	}
}

func TestGetOnDeck(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/library/onDeck", "deck-token", plex.MediaContainer{
		Metadata: []plex.Metadata{
			{RatingKey: "55", Title: "Episode 5", ViewOffset: 120000},
		},
	})
	defer ts.Close()

	c, err := plex.New(ts.URL, "deck-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	items, err := c.GetOnDeck(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}
	if items[0].ViewOffset != 120000 {
		t.Errorf("ViewOffset = %d, want 120000", items[0].ViewOffset)
	}
}

func TestGetRecentlyAdded(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/library/recentlyAdded", "recent-token", plex.MediaContainer{
		Metadata: []plex.Metadata{
			{RatingKey: "100", Title: "New Movie", Year: 2024},
		},
	})
	defer ts.Close()

	c, err := plex.New(ts.URL, "recent-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	items, err := c.GetRecentlyAdded(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}
	if items[0].Title != "New Movie" {
		t.Errorf("Title = %q, want New Movie", items[0].Title)
	}
}

func TestSearch(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("path = %q, want /search", r.URL.Path)
		}
		if got := r.URL.Query().Get("query"); got != "dark knight" {
			t.Errorf("query = %q, want dark knight", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResponse{MediaContainer: plex.MediaContainer{
			Metadata: []plex.Metadata{
				{RatingKey: "120", Title: "The Dark Knight", Year: 2008},
			},
		}})
	}))
	defer ts.Close()

	c, err := plex.New(ts.URL, "search-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	mc, err := c.Search(context.Background(), "dark knight")
	if err != nil {
		t.Fatal(err)
	}
	if len(mc.Metadata) != 1 {
		t.Fatalf("len = %d, want 1", len(mc.Metadata))
	}
	if mc.Metadata[0].Title != "The Dark Knight" {
		t.Errorf("Title = %q, want The Dark Knight", mc.Metadata[0].Title)
	}
}

func TestGetSessions(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/status/sessions", "sess-token", plex.MediaContainer{
		Metadata: []plex.Metadata{{
			RatingKey: "1244",
			Title:     "Riddick",
			User:      &plex.SessionUser{ID: "1", Title: "admin"},
			Player:    &plex.Player{Title: "Chrome", State: "playing", Platform: "Web"},
			Session:   &plex.SessionInfo{ID: "abc123", Location: "lan"},
		}},
	})
	defer ts.Close()

	c, err := plex.New(ts.URL, "sess-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sessions, err := c.GetSessions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("len = %d, want 1", len(sessions))
	}
	if sessions[0].User.Title != "admin" {
		t.Errorf("User = %q, want admin", sessions[0].User.Title)
	}
	if sessions[0].Player.State != "playing" {
		t.Errorf("State = %q, want playing", sessions[0].Player.State)
	}
}

func TestRefreshLibrary(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/library/sections/1/refresh", "refresh-token", plex.MediaContainer{})
	defer ts.Close()

	c, err := plex.New(ts.URL, "refresh-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.RefreshLibrary(context.Background(), "1"); err != nil {
		t.Fatal(err)
	}
}

func TestMarkWatched(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/:/scrobble" {
			t.Errorf("path = %q, want /:/scrobble", r.URL.Path)
		}
		if got := r.URL.Query().Get("key"); got != "1244" {
			t.Errorf("key = %q, want 1244", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResponse{MediaContainer: plex.MediaContainer{}})
	}))
	defer ts.Close()

	c, err := plex.New(ts.URL, "scrobble-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.MarkWatched(context.Background(), "1244"); err != nil {
		t.Fatal(err)
	}
}

func TestMarkUnwatched(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/:/unscrobble" {
			t.Errorf("path = %q, want /:/unscrobble", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResponse{MediaContainer: plex.MediaContainer{}})
	}))
	defer ts.Close()

	c, err := plex.New(ts.URL, "unscrobble-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.MarkUnwatched(context.Background(), "1244"); err != nil {
		t.Fatal(err)
	}
}

func TestGetServerInfo(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/", "info-token", plex.MediaContainer{
		FriendlyName:      "My Plex Server",
		MachineIdentifier: "abc123",
		Version:           "1.32.0",
		Platform:          "Linux",
	})
	defer ts.Close()

	c, err := plex.New(ts.URL, "info-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	mc, err := c.GetServerInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if mc.FriendlyName != "My Plex Server" {
		t.Errorf("FriendlyName = %q, want My Plex Server", mc.FriendlyName)
	}
	if mc.Platform != "Linux" {
		t.Errorf("Platform = %q, want Linux", mc.Platform)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer ts.Close()

	c, err := plex.New(ts.URL, "bad-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetLibraries(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *plex.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *plex.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
}

func TestAPIErrorMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  plex.APIError
		want string
	}{
		{"with body", plex.APIError{StatusCode: 401, RawBody: "Unauthorized"}, "plex: HTTP 401: Unauthorized"},
		{"code only", plex.APIError{StatusCode: 500}, "plex: HTTP 500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/library/sections", "cancel-token", plex.MediaContainer{})
	defer ts.Close()

	c, err := plex.New(ts.URL, "cancel-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = c.GetLibraries(ctx)
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Plex-Product"); got != "MyApp" {
			t.Errorf("X-Plex-Product = %q, want MyApp", got)
		}
		if got := r.Header.Get("X-Plex-Client-Identifier"); got != "my-app-id" {
			t.Errorf("X-Plex-Client-Identifier = %q, want my-app-id", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResponse{MediaContainer: plex.MediaContainer{}})
	}))
	defer ts.Close()

	c, err := plex.New(ts.URL, "opt-token",
		plex.WithProduct("MyApp"),
		plex.WithClientIdentifier("my-app-id"),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetLibraries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResponse{MediaContainer: map[string]any{}})
	}))
	defer ts.Close()

	c, err := plex.New(ts.URL, "tok", plex.WithUserAgent("myapp/1.2.3"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := c.GetServerInfo(context.Background()); err != nil {
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
			c, err := plex.New(tc.url, "tok")
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}
