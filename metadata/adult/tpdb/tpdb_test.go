package tpdb_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/metadata"
	"github.com/golusoris/goenvoy/metadata/adult/tpdb"
)

func newTestServer(t *testing.T, wantPath, wantToken string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		wantAuth := "Bearer " + wantToken
		if got := r.Header.Get("Authorization"); got != wantAuth {
			t.Errorf("Authorization = %q, want %q", got, wantAuth)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want %q", got, "application/json")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
}

type paginatedResponse struct {
	Data any             `json:"data"`
	Meta tpdb.Pagination `json:"meta"`
}

func newListServer(t *testing.T, wantPath, wantToken string, data any) *httptest.Server {
	t.Helper()
	resp := paginatedResponse{
		Data: data,
		Meta: tpdb.Pagination{Total: 1, PerPage: 25, CurrentPage: 1, LastPage: 1},
	}
	return newTestServer(t, wantPath, wantToken, resp)
}

type itemResponseWrapper struct {
	Data any `json:"data"`
}

func newItemServer(t *testing.T, wantPath, wantToken string, data any) *httptest.Server {
	t.Helper()
	return newTestServer(t, wantPath, wantToken, itemResponseWrapper{Data: data})
}

func TestSearchScenes(t *testing.T) {
	t.Parallel()

	scene := tpdb.Scene{
		ID:    42,
		PID:   "abc123",
		Title: "Test Scene",
		Date:  "2024-01-15",
		Tags:  []tpdb.TagRef{{ID: 1, Name: "tag1", Slug: "tag1"}},
	}
	ts := newListServer(t, "/scenes", "test-token", []tpdb.Scene{scene})
	defer ts.Close()

	c := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	scenes, pg, err := c.SearchScenes(context.Background(), &tpdb.SceneSearchParams{Query: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 1 {
		t.Fatalf("len = %d, want 1", len(scenes))
	}
	if scenes[0].Title != "Test Scene" {
		t.Errorf("Title = %q, want %q", scenes[0].Title, "Test Scene")
	}
	if scenes[0].PID != "abc123" {
		t.Errorf("PID = %q, want %q", scenes[0].PID, "abc123")
	}
	if pg.Total != 1 {
		t.Errorf("Total = %d, want 1", pg.Total)
	}
}

func TestSearchScenesNilParams(t *testing.T) {
	t.Parallel()

	ts := newListServer(t, "/scenes", "test-token", []tpdb.Scene{{ID: 1, Title: "Scene"}})
	defer ts.Close()

	c := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	scenes, _, err := c.SearchScenes(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 1 {
		t.Fatalf("len = %d, want 1", len(scenes))
	}
}

func TestGetScene(t *testing.T) {
	t.Parallel()

	scene := tpdb.Scene{
		ID:          99,
		PID:         "def456",
		Title:       "Specific Scene",
		Description: "A description",
		Duration:    1800,
		Site:        &tpdb.SiteRef{ID: 5, Name: "TestSite"},
	}
	ts := newItemServer(t, "/scenes/def456", "scene-token", scene)
	defer ts.Close()

	c := tpdb.New("scene-token", metadata.WithBaseURL(ts.URL))
	s, err := c.GetScene(context.Background(), "def456")
	if err != nil {
		t.Fatal(err)
	}
	if s.Title != "Specific Scene" {
		t.Errorf("Title = %q, want %q", s.Title, "Specific Scene")
	}
	if s.Duration != 1800 {
		t.Errorf("Duration = %d, want %d", s.Duration, 1800)
	}
	if s.Site == nil || s.Site.Name != "TestSite" {
		t.Error("Site not parsed correctly")
	}
}

func TestGetSimilarScenes(t *testing.T) {
	t.Parallel()

	scenes := []tpdb.Scene{{ID: 10, Title: "Similar 1"}, {ID: 11, Title: "Similar 2"}}
	ts := newListServer(t, "/scenes/abc/similar", "sim-token", scenes)
	defer ts.Close()

	c := tpdb.New("sim-token", metadata.WithBaseURL(ts.URL))
	result, err := c.GetSimilarScenes(context.Background(), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
}

func TestSearchPerformers(t *testing.T) {
	t.Parallel()

	perf := tpdb.Performer{
		ID:   1,
		PID:  "perf1",
		Name: "Test Performer",
		Extras: tpdb.ExtraData{
			Gender:    "female",
			Ethnicity: "caucasian",
		},
	}
	ts := newListServer(t, "/performers", "perf-token", []tpdb.Performer{perf})
	defer ts.Close()

	c := tpdb.New("perf-token", metadata.WithBaseURL(ts.URL))
	perfs, pg, err := c.SearchPerformers(context.Background(), &tpdb.PerformerSearchParams{Query: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(perfs) != 1 {
		t.Fatalf("len = %d, want 1", len(perfs))
	}
	if perfs[0].Name != "Test Performer" {
		t.Errorf("Name = %q, want %q", perfs[0].Name, "Test Performer")
	}
	if perfs[0].Extras.Gender != "female" {
		t.Errorf("Gender = %q, want %q", perfs[0].Extras.Gender, "female")
	}
	if pg.Total != 1 {
		t.Errorf("Total = %d, want 1", pg.Total)
	}
}

func TestSearchPerformersNilParams(t *testing.T) {
	t.Parallel()

	ts := newListServer(t, "/performers", "perf-token", []tpdb.Performer{{ID: 1, Name: "Performer"}})
	defer ts.Close()

	c := tpdb.New("perf-token", metadata.WithBaseURL(ts.URL))
	perfs, _, err := c.SearchPerformers(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(perfs) != 1 {
		t.Fatalf("len = %d, want 1", len(perfs))
	}
}

func TestGetPerformer(t *testing.T) {
	t.Parallel()

	perf := tpdb.Performer{
		ID:       2,
		PID:      "perf2",
		Name:     "Specific Performer",
		FullName: "Specific J. Performer",
		Bio:      "A bio",
		Rating:   4.5,
		Aliases:  []string{"Alias1", "Alias2"},
	}
	ts := newItemServer(t, "/performers/perf2", "p-token", perf)
	defer ts.Close()

	c := tpdb.New("p-token", metadata.WithBaseURL(ts.URL))
	p, err := c.GetPerformer(context.Background(), "perf2")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Specific Performer" {
		t.Errorf("Name = %q, want %q", p.Name, "Specific Performer")
	}
	if p.Rating != 4.5 {
		t.Errorf("Rating = %f, want %f", p.Rating, 4.5)
	}
	if len(p.Aliases) != 2 {
		t.Errorf("len(Aliases) = %d, want 2", len(p.Aliases))
	}
}

func TestGetPerformerScenes(t *testing.T) {
	t.Parallel()

	scenes := []tpdb.Scene{{ID: 20, Title: "Scene With Performer"}}
	ts := newListServer(t, "/performers/p1/scenes", "ps-token", scenes)
	defer ts.Close()

	c := tpdb.New("ps-token", metadata.WithBaseURL(ts.URL))
	result, pg, err := c.GetPerformerScenes(context.Background(), "p1", 1, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if pg.CurrentPage != 1 {
		t.Errorf("CurrentPage = %d, want 1", pg.CurrentPage)
	}
}

func TestSearchSites(t *testing.T) {
	t.Parallel()

	site := tpdb.Site{ID: 3, UUID: "uuid-123", Name: "Test Studio", URL: "https://example.com"}
	ts := newListServer(t, "/sites", "site-token", []tpdb.Site{site})
	defer ts.Close()

	c := tpdb.New("site-token", metadata.WithBaseURL(ts.URL))
	sites, pg, err := c.SearchSites(context.Background(), "test", 1, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(sites) != 1 {
		t.Fatalf("len = %d, want 1", len(sites))
	}
	if sites[0].Name != "Test Studio" {
		t.Errorf("Name = %q, want %q", sites[0].Name, "Test Studio")
	}
	if pg.Total != 1 {
		t.Errorf("Total = %d, want 1", pg.Total)
	}
}

func TestGetSite(t *testing.T) {
	t.Parallel()

	site := tpdb.Site{ID: 4, UUID: "uuid-456", Name: "Specific Site", Rating: 3.8}
	ts := newItemServer(t, "/sites/uuid-456", "s-token", site)
	defer ts.Close()

	c := tpdb.New("s-token", metadata.WithBaseURL(ts.URL))
	s, err := c.GetSite(context.Background(), "uuid-456")
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "Specific Site" {
		t.Errorf("Name = %q, want %q", s.Name, "Specific Site")
	}
	if s.Rating != 3.8 {
		t.Errorf("Rating = %f, want %f", s.Rating, 3.8)
	}
}

func TestListTags(t *testing.T) {
	t.Parallel()

	tags := []tpdb.Tag{{ID: 1, Name: "Brunette", Slug: "brunette"}}
	ts := newListServer(t, "/tags", "tag-token", tags)
	defer ts.Close()

	c := tpdb.New("tag-token", metadata.WithBaseURL(ts.URL))
	result, pg, err := c.ListTags(context.Background(), "brunette", 1, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Slug != "brunette" {
		t.Errorf("Slug = %q, want %q", result[0].Slug, "brunette")
	}
	if pg.Total != 1 {
		t.Errorf("Total = %d, want 1", pg.Total)
	}
}

func TestGetTag(t *testing.T) {
	t.Parallel()

	tag := tpdb.Tag{ID: 7, Name: "Blonde", Slug: "blonde"}
	ts := newItemServer(t, "/tags/blonde", "gt-token", tag)
	defer ts.Close()

	c := tpdb.New("gt-token", metadata.WithBaseURL(ts.URL))
	result, err := c.GetTag(context.Background(), "blonde")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Blonde" {
		t.Errorf("Name = %q, want %q", result.Name, "Blonde")
	}
}

func TestListDirectors(t *testing.T) {
	t.Parallel()

	dirs := []tpdb.Director{{ID: 1, Name: "John Doe", Slug: "john-doe"}}
	ts := newListServer(t, "/directors", "dir-token", dirs)
	defer ts.Close()

	c := tpdb.New("dir-token", metadata.WithBaseURL(ts.URL))
	result, pg, err := c.ListDirectors(context.Background(), "john", 1, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Name != "John Doe" {
		t.Errorf("Name = %q, want %q", result[0].Name, "John Doe")
	}
	if pg.Total != 1 {
		t.Errorf("Total = %d, want 1", pg.Total)
	}
}

func TestSearchMovies(t *testing.T) {
	t.Parallel()

	movie := tpdb.Movie{ID: 5, Title: "Test Movie", SKU: "ABC-123"}
	ts := newListServer(t, "/movies", "mov-token", []tpdb.Movie{movie})
	defer ts.Close()

	c := tpdb.New("mov-token", metadata.WithBaseURL(ts.URL))
	movies, pg, err := c.SearchMovies(context.Background(), "test", 1, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(movies) != 1 {
		t.Fatalf("len = %d, want 1", len(movies))
	}
	if movies[0].SKU != "ABC-123" {
		t.Errorf("SKU = %q, want %q", movies[0].SKU, "ABC-123")
	}
	if pg.Total != 1 {
		t.Errorf("Total = %d, want 1", pg.Total)
	}
}

func TestGetMovie(t *testing.T) {
	t.Parallel()

	movie := tpdb.Movie{ID: 6, Title: "Specific Movie", Duration: 7200}
	ts := newItemServer(t, "/movies/6", "gm-token", movie)
	defer ts.Close()

	c := tpdb.New("gm-token", metadata.WithBaseURL(ts.URL))
	m, err := c.GetMovie(context.Background(), "6")
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "Specific Movie" {
		t.Errorf("Title = %q, want %q", m.Title, "Specific Movie")
	}
	if m.Duration != 7200 {
		t.Errorf("Duration = %d, want %d", m.Duration, 7200)
	}
}

func TestSearchJav(t *testing.T) {
	t.Parallel()

	jav := tpdb.Jav{ID: 8, Title: "Test JAV", SKU: "JAV-001"}
	ts := newListServer(t, "/jav", "jav-token", []tpdb.Jav{jav})
	defer ts.Close()

	c := tpdb.New("jav-token", metadata.WithBaseURL(ts.URL))
	result, pg, err := c.SearchJav(context.Background(), "test", 1, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].SKU != "JAV-001" {
		t.Errorf("SKU = %q, want %q", result[0].SKU, "JAV-001")
	}
	if pg.Total != 1 {
		t.Errorf("Total = %d, want 1", pg.Total)
	}
}

func TestGetJav(t *testing.T) {
	t.Parallel()

	jav := tpdb.Jav{ID: 9, Title: "Specific JAV", Duration: 5400}
	ts := newItemServer(t, "/jav/9", "gj-token", jav)
	defer ts.Close()

	c := tpdb.New("gj-token", metadata.WithBaseURL(ts.URL))
	j, err := c.GetJav(context.Background(), "9")
	if err != nil {
		t.Fatal(err)
	}
	if j.Title != "Specific JAV" {
		t.Errorf("Title = %q, want %q", j.Title, "Specific JAV")
	}
}

func TestGetChanges(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/changes/scenes" {
			t.Errorf("path = %q, want /changes/scenes", r.URL.Path)
		}
		if got := r.URL.Query().Get("from"); got != "2024-01-01T00:00:00Z" {
			t.Errorf("from = %q, want 2024-01-01T00:00:00Z", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []tpdb.ChangeRecord{{ID: 100, UUID: "00000000-0000-4000-8000-000000000001"}},
			"meta": tpdb.Pagination{Total: 1, PerPage: 25, CurrentPage: 1, LastPage: 1},
		})
	}))
	defer ts.Close()

	c := tpdb.New("ch-token", metadata.WithBaseURL(ts.URL))
	result, pg, err := c.GetChanges(context.Background(), "2024-01-01T00:00:00Z", 1, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].ID != 100 || result[0].PID != "00000000-0000-4000-8000-000000000001" {
		t.Errorf("result[0] = %+v, want ID 100 and PID 00000000-0000-4000-8000-000000000001", result[0])
	}
	if pg.Total != 1 {
		t.Errorf("Total = %d, want 1", pg.Total)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Unauthenticated."})
	}))
	defer ts.Close()

	c := tpdb.New("bad-token", metadata.WithBaseURL(ts.URL))
	_, err := c.GetScene(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *tpdb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
	if apiErr.Message != "Unauthenticated." {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Unauthenticated.")
	}
}

func TestAPIErrorRawBody(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	c := tpdb.New("token", metadata.WithBaseURL(ts.URL))
	_, err := c.GetScene(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *tpdb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.RawBody != "internal error" {
		t.Errorf("RawBody = %q, want %q", apiErr.RawBody, "internal error")
	}
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	ts := newItemServer(t, "/scenes/1", "custom-token", tpdb.Scene{ID: 1, Title: "Custom"})
	defer ts.Close()

	custom := &http.Client{}
	c := tpdb.New("custom-token", metadata.WithBaseURL(ts.URL), metadata.WithHTTPClient(custom))
	s, err := c.GetScene(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if s.Title != "Custom" {
		t.Errorf("Title = %q, want %q", s.Title, "Custom")
	}
}

func TestGetSimilarPerformers(t *testing.T) {
	t.Parallel()

	perfs := []tpdb.Performer{{ID: 50, Name: "Similar Performer"}}
	ts := newListServer(t, "/performers/p1/similar", "sp-token", perfs)
	defer ts.Close()

	c := tpdb.New("sp-token", metadata.WithBaseURL(ts.URL))
	result, err := c.GetSimilarPerformers(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Name != "Similar Performer" {
		t.Errorf("Name = %q, want %q", result[0].Name, "Similar Performer")
	}
}

func TestGetPerformerMovies(t *testing.T) {
	t.Parallel()

	movies := []tpdb.Movie{{ID: 60, Title: "Performer Movie"}}
	ts := newListServer(t, "/performers/p2/movies", "pm-token", movies)
	defer ts.Close()

	c := tpdb.New("pm-token", metadata.WithBaseURL(ts.URL))
	result, pg, err := c.GetPerformerMovies(context.Background(), "p2", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Title != "Performer Movie" {
		t.Errorf("Title = %q, want %q", result[0].Title, "Performer Movie")
	}
	if pg.PerPage != 25 {
		t.Errorf("PerPage = %d, want 25", pg.PerPage)
	}
}

func TestGetDirector(t *testing.T) {
	t.Parallel()

	dir := tpdb.Director{ID: 12, Name: "Test Director", Slug: "test-director"}
	ts := newItemServer(t, "/directors/test-director", "gd-token", dir)
	defer ts.Close()

	c := tpdb.New("gd-token", metadata.WithBaseURL(ts.URL))
	d, err := c.GetDirector(context.Background(), "test-director")
	if err != nil {
		t.Fatal(err)
	}
	if d.Name != "Test Director" {
		t.Errorf("Name = %q, want %q", d.Name, "Test Director")
	}
	if d.Slug != "test-director" {
		t.Errorf("Slug = %q, want %q", d.Slug, "test-director")
	}
}

func TestFindSceneByHash(t *testing.T) {
	t.Parallel()

	scenes := []tpdb.Scene{{ID: 77, Title: "Hashed Scene"}}
	ts := newListServer(t, "/scenes", "hash-token", scenes)
	defer ts.Close()

	c := tpdb.New("hash-token", metadata.WithBaseURL(ts.URL))
	result, err := c.FindSceneByHash(context.Background(), "abc123def456")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Title != "Hashed Scene" {
		t.Errorf("Title = %q, want %q", result[0].Title, "Hashed Scene")
	}
}

func TestLookupByHash(t *testing.T) {
	t.Parallel()

	scenes := []tpdb.Scene{{ID: 78, Title: "Looked Up"}}
	ts := newListServer(t, "/scenes", "hash-token", scenes)
	defer ts.Close()

	c := tpdb.New("hash-token", metadata.WithBaseURL(ts.URL))
	result, err := c.LookupByHash(context.Background(), "abc123def456")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].ID != 78 {
		t.Fatalf("result = %+v, want one scene with ID 78", result)
	}
}

func assertScenePosters(t *testing.T, s *tpdb.Scene) {
	t.Helper()
	if s.Title != "Fixture Scene" {
		t.Errorf("Title = %q, want Fixture Scene", s.Title)
	}
	if s.PosterSizes.Full != "https://cdn.example.test/scene/poster-full.jpg" ||
		s.PosterSizes.Large != "https://cdn.example.test/scene/poster-large.jpg" ||
		s.PosterSizes.Medium != "https://cdn.example.test/scene/poster-medium.jpg" ||
		s.PosterSizes.Small != "" {
		t.Errorf("PosterSizes = %+v", s.PosterSizes)
	}
	if s.Background.Full != "https://cdn.example.test/scene/bg-full.jpg" ||
		s.Background.Small != "https://cdn.example.test/scene/bg-small.jpg" {
		t.Errorf("Background = %+v", s.Background)
	}
	if len(s.Posters) != 0 || len(s.Backgrounds) != 0 {
		t.Errorf("deprecated slices must be empty, got posters=%d backgrounds=%d", len(s.Posters), len(s.Backgrounds))
	}
}

func TestGetScene_PostersFixture(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/scene_posters.json")
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/scenes/00000000-0000-4000-8000-000000000001" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer ts.Close()

	c := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	scene, err := c.GetScene(context.Background(), "00000000-0000-4000-8000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	assertScenePosters(t, scene)
}

func assertPerformerPosters(t *testing.T, p *tpdb.Performer) {
	t.Helper()
	if p.Name != "Fixture Performer" {
		t.Errorf("Name = %q, want Fixture Performer", p.Name)
	}
	if len(p.Posters) != 2 {
		t.Fatalf("len(Posters) = %d, want 2", len(p.Posters))
	}
	if p.Posters[0].ID != 101 || p.Posters[0].Size != 102400 || p.Posters[0].Order != 1 {
		t.Errorf("Posters[0] = %+v", p.Posters[0])
	}
	if p.Posters[1].ID != 102 || p.Posters[1].Size != 204800 || p.Posters[1].Order != 2 {
		t.Errorf("Posters[1] = %+v", p.Posters[1])
	}
}

func TestGetPerformer_PostersFixture(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/performer_posters.json")
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/performers/00000000-0000-4000-8000-000000000002" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer ts.Close()

	c := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	p, err := c.GetPerformer(context.Background(), "00000000-0000-4000-8000-000000000002")
	if err != nil {
		t.Fatal(err)
	}
	assertPerformerPosters(t, p)
}

type listChangesCase struct {
	name      string
	entity    tpdb.ChangeEntity
	from      time.Time
	page      int
	perPage   int
	wantPath  string
	wantQuery string
	respBody  []byte
	wantErr   bool
	wantCount int
	wantTotal int
}

func newChangesServer(t *testing.T, wantPath, wantQuery string, body []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != wantPath || r.URL.Query().Encode() != wantQuery {
			t.Errorf("got %s %s?%s, want GET %s?%s", r.Method, r.URL.Path, r.URL.Query().Encode(), wantPath, wantQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
}

func assertChangeRecords(t *testing.T, records []tpdb.ChangeRecord, pg *tpdb.Pagination, wantCount, wantTotal int) {
	t.Helper()
	if len(records) != wantCount || pg.Total != wantTotal {
		t.Fatalf("records=%d, total=%d, want %d/%d", len(records), pg.Total, wantCount, wantTotal)
	}
	if wantCount > 0 && records[0].ID != 101 {
		t.Errorf("records[0].ID = %d, want 101", records[0].ID)
	}
}

func checkListChangesCase(t *testing.T, tc listChangesCase) {
	t.Helper()
	if tc.wantErr {
		c := tpdb.New("test-token")
		if _, _, err := c.ListChanges(context.Background(), tc.entity, tc.from, tc.page, tc.perPage); err == nil {
			t.Fatal("expected error, got nil")
		}
		return
	}

	ts := newChangesServer(t, tc.wantPath, tc.wantQuery, tc.respBody)
	defer ts.Close()

	c := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	records, pg, err := c.ListChanges(context.Background(), tc.entity, tc.from, tc.page, tc.perPage)
	if err != nil {
		t.Fatal(err)
	}
	assertChangeRecords(t, records, pg, tc.wantCount, tc.wantTotal)
}

func TestListChanges(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/changes_scenes.json")
	if err != nil {
		t.Fatal(err)
	}

	tests := []listChangesCase{
		{
			name:      "positive",
			entity:    tpdb.ChangeScenes,
			from:      time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			page:      1,
			perPage:   25,
			wantPath:  "/changes/scenes",
			wantQuery: "from=2024-01-01T12%3A00%3A00Z&page=1&per_page=25",
			respBody:  fixture,
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name:      "boundary_zero_from_and_per_page",
			entity:    tpdb.ChangeMovies,
			wantPath:  "/changes/movies",
			wantQuery: "",
			respBody:  []byte(`{"data":[],"meta":{"total":0}}`),
			wantCount: 0,
			wantTotal: 0,
		},
		{
			name:    "negative_invalid_entity",
			entity:  tpdb.ChangeEntity("invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkListChangesCase(t, tt)
		})
	}
}

func TestGetChanges_NegativeAndBoundary(t *testing.T) {
	t.Parallel()

	t.Run("negative_bad_since", func(t *testing.T) {
		t.Parallel()
		c := tpdb.New("test-token")
		if _, _, err := c.GetChanges(context.Background(), "invalid-date", 1, 25); err == nil {
			t.Fatal("expected error for unparseable since, got nil")
		}
	})

	t.Run("boundary_empty_since", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Has("from") {
				t.Errorf("unexpected from param: %q", r.URL.Query().Get("from"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[],"meta":{}}`))
		}))
		defer ts.Close()

		c := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
		scenes, _, err := c.GetChanges(context.Background(), "", 1, 25)
		if err != nil {
			t.Fatal(err)
		}
		if len(scenes) != 0 {
			t.Errorf("len(scenes) = %d, want 0", len(scenes))
		}
	})
}

type searchPerformerSitesCase struct {
	name      string
	params    *tpdb.PerformerSiteSearchParams
	wantQuery string
	respBody  []byte
	wantCount int
}

func checkSearchPerformerSitesCase(t *testing.T, tc searchPerformerSitesCase) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/performer-sites" || r.URL.Query().Encode() != tc.wantQuery {
			t.Errorf("got %s %s?%s, want GET /performer-sites?%s", r.Method, r.URL.Path, r.URL.Query().Encode(), tc.wantQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(tc.respBody)
	}))
	defer ts.Close()

	c := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	sites, pg, err := c.SearchPerformerSites(context.Background(), tc.params)
	if err != nil {
		t.Fatal(err)
	}
	if len(sites) != tc.wantCount {
		t.Fatalf("len(sites) = %d, want %d", len(sites), tc.wantCount)
	}
	if tc.wantCount > 0 {
		assertPerformerSiteFixture(t, &sites[0], pg)
	}
}

func assertPerformerSiteFixture(t *testing.T, s *tpdb.PerformerSite, pg *tpdb.Pagination) {
	t.Helper()
	if s.ID != "00000000-0000-4000-8000-000000000001" || s.NumericID != 1 || s.SiteID != 10 {
		t.Errorf("IDs = %q/%d/%d", s.ID, s.NumericID, s.SiteID)
	}
	if s.Name != "Fixture Performer Site 1" || !s.IsParent || s.Extras.Gender != "Female" {
		t.Errorf("content = %+v", s)
	}
	if pg.Total != 1 {
		t.Errorf("Total = %d, want 1", pg.Total)
	}
}

func TestSearchPerformerSites(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/performer_sites.json")
	if err != nil {
		t.Fatal(err)
	}

	tests := []searchPerformerSitesCase{
		{
			name: "positive_all_params",
			params: &tpdb.PerformerSiteSearchParams{
				Query:           "Fixture",
				OrderBy:         "name",
				Letter:          "F",
				Parent:          "Parent",
				ParentID:        5,
				ParentOperation: "include",
				Site:            "Site",
				SiteID:          10,
				Page:            2,
				PerPage:         15,
			},
			wantQuery: "letter=F&orderBy=name&page=2&parent=Parent&parent_id=5&parent_operation=include&per_page=15&q=Fixture&site=Site&site_id=10",
			respBody:  fixture,
			wantCount: 1,
		},
		{
			name:      "boundary_nil_params",
			params:    nil,
			wantQuery: "",
			respBody:  []byte(`{"data":[],"meta":{"total":0}}`),
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkSearchPerformerSitesCase(t, tt)
		})
	}
}

type getPerformerSiteCase struct {
	name       string
	id         string
	statusCode int
	respBody   []byte
	wantErr    bool
}

func checkGetPerformerSiteCase(t *testing.T, tc getPerformerSiteCase) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/performer-sites/"+tc.id {
			t.Errorf("got %s %s, want GET /performer-sites/%s", r.Method, r.URL.Path, tc.id)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(tc.statusCode)
		_, _ = w.Write(tc.respBody)
	}))
	defer ts.Close()

	c := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	ps, err := c.GetPerformerSite(context.Background(), tc.id)
	if tc.wantErr {
		var apiErr *tpdb.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
			t.Fatalf("expected HTTP 404 APIError, got: %v", err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	assertSinglePerformerSite(t, ps)
}

func assertPerformerSiteBasics(t *testing.T, ps *tpdb.PerformerSite) {
	t.Helper()
	if ps.ID != "00000000-0000-4000-8000-000000000001" || ps.NumericID != 1 || ps.SiteID != 10 {
		t.Errorf("IDs = %q/%d/%d", ps.ID, ps.NumericID, ps.SiteID)
	}
	if ps.Name != "Fixture Performer Site" || ps.Bio != "Fixture Bio" || ps.IsParent {
		t.Errorf("content = %+v", ps)
	}
	if ps.Extras.Gender != "Female" || ps.Extras.Birthday != "1995-01-01" {
		t.Errorf("Extras = %+v", ps.Extras)
	}
}

func assertPerformerSiteRelations(t *testing.T, ps *tpdb.PerformerSite) {
	t.Helper()
	if ps.Image == "" || ps.Thumbnail == "" || ps.Face == "" {
		t.Errorf("images empty: %+v", ps)
	}
	if ps.Parent == nil || ps.Parent.Name != "Fixture Parent Performer" {
		t.Errorf("Parent = %+v", ps.Parent)
	}
	if ps.Site == nil || ps.Site.ID != 10 || ps.Site.Name != "Fixture Site" {
		t.Errorf("Site = %+v", ps.Site)
	}
}

func assertSinglePerformerSite(t *testing.T, ps *tpdb.PerformerSite) {
	t.Helper()
	assertPerformerSiteBasics(t, ps)
	assertPerformerSiteRelations(t, ps)
}

func TestGetPerformerSite(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/performer_site.json")
	if err != nil {
		t.Fatal(err)
	}

	tests := []getPerformerSiteCase{
		{
			name:       "positive",
			id:         "00000000-0000-4000-8000-000000000001",
			statusCode: http.StatusOK,
			respBody:   fixture,
		},
		{
			name:       "negative_404",
			id:         "not-found",
			statusCode: http.StatusNotFound,
			respBody:   []byte(`{"message":"Performer site not found."}`),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkGetPerformerSiteCase(t, tt)
		})
	}
}

const favouritesEndpoint = "/favourites" //nolint:misspell // Upstream API endpoint uses British spelling

type listFavouritesCase struct {
	name      string
	favType   tpdb.FavouriteType
	wantQuery string
	respBody  []byte
	wantErr   bool
	wantIDs   []int
}

func checkListFavouritesCase(t *testing.T, tc listFavouritesCase) {
	t.Helper()
	if tc.wantErr {
		c := tpdb.New("test-token")
		if _, err := c.ListFavourites(context.Background(), tc.favType); err == nil {
			t.Fatal("expected error, got nil")
		}
		return
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != favouritesEndpoint || r.URL.Query().Encode() != tc.wantQuery {
			t.Errorf("got %s %s?%s, want GET %s?%s", r.Method, r.URL.Path, r.URL.Query().Encode(), favouritesEndpoint, tc.wantQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(tc.respBody)
	}))
	defer ts.Close()

	c := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	ids, err := c.ListFavourites(context.Background(), tc.favType)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != len(tc.wantIDs) {
		t.Fatalf("len(ids) = %d, want %d", len(ids), len(tc.wantIDs))
	}
	for i, wantID := range tc.wantIDs {
		if ids[i] != wantID {
			t.Errorf("ids[%d] = %d, want %d", i, ids[i], wantID)
		}
	}
}

func TestListFavourites(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/favourites_list.json") //nolint:misspell // Fixture file name matches API path
	if err != nil {
		t.Fatal(err)
	}

	tests := []listFavouritesCase{
		{
			name:      "positive_performer",
			favType:   tpdb.FavouritePerformer,
			wantQuery: "type=performer",
			respBody:  fixture,
			wantIDs:   []int{101, 102, 103},
		},
		{
			name:      "boundary_empty_data",
			favType:   tpdb.FavouriteSite,
			wantQuery: "type=site",
			respBody:  []byte(`{"data":[]}`),
			wantIDs:   []int{},
		},
		{
			name:    "negative_invalid_type",
			favType: tpdb.FavouriteType("invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkListFavouritesCase(t, tt)
		})
	}
}

type toggleFavouriteCase struct {
	name      string
	favType   tpdb.FavouriteType
	id        int
	wantQuery string
	respBody  []byte
	wantErr   bool
	wantVal   bool
}

func checkToggleFavouriteCase(t *testing.T, tc toggleFavouriteCase) {
	t.Helper()
	if tc.wantErr {
		c := tpdb.New("test-token")
		if _, err := c.ToggleFavourite(context.Background(), tc.favType, tc.id); err == nil {
			t.Fatal("expected error, got nil")
		}
		return
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != favouritesEndpoint || r.URL.Query().Encode() != tc.wantQuery {
			t.Errorf("got %s %s?%s, want POST %s?%s", r.Method, r.URL.Path, r.URL.Query().Encode(), favouritesEndpoint, tc.wantQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(tc.respBody)
	}))
	defer ts.Close()

	c := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	val, err := c.ToggleFavourite(context.Background(), tc.favType, tc.id)
	if err != nil {
		t.Fatal(err)
	}
	if val != tc.wantVal {
		t.Errorf("val = %v, want %v", val, tc.wantVal)
	}
}

func TestToggleFavourite(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/favourite_toggle.json") //nolint:misspell // Fixture file name matches API path
	if err != nil {
		t.Fatal(err)
	}

	tests := []toggleFavouriteCase{
		{
			name:      "positive_performer",
			favType:   tpdb.FavouritePerformer,
			id:        101,
			wantQuery: "type=performer&value=101",
			respBody:  fixture,
			wantVal:   true,
		},
		{
			name:      "positive_site",
			favType:   tpdb.FavouriteSite,
			id:        202,
			wantQuery: "type=site&value=202",
			respBody:  []byte(`{"value":false}`),
			wantVal:   false,
		},
		{
			name:    "negative_invalid_type",
			favType: tpdb.FavouriteType("invalid"),
			id:      101,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkToggleFavouriteCase(t, tt)
		})
	}
}
