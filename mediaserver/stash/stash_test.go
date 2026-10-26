package stash_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golusoris/goenvoy/mediaserver/stash/v2"
)

func newGQLServer(t *testing.T, wantKey, dataField string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if wantKey != "" {
			if got := r.Header.Get("Apikey"); got != wantKey {
				t.Errorf("ApiKey = %q, want %q", got, wantKey)
			}
		}
		data, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{` + `"` + dataField + `":` + string(data) + `}}`))
	}))
}

func TestFindScene(t *testing.T) {
	t.Parallel()

	scene := stash.Scene{
		ID:        "1",
		Title:     "Test Scene",
		Rating100: intPtr(85),
		Studio:    &stash.Studio{ID: "s1", Name: "Test Studio"},
		Tags:      []stash.Tag{{ID: "t1", Name: "Tag1"}},
	}
	ts := newGQLServer(t, "key-1", "findScene", scene)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-1")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindScene(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Title != "Test Scene" {
		t.Errorf("Title = %q, want %q", result.Title, "Test Scene")
	}
	if result.Rating100 == nil || *result.Rating100 != 85 {
		t.Error("Rating100 not parsed correctly")
	}
}

func TestFindScenes(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count  int           `json:"count"`
		Scenes []stash.Scene `json:"scenes"`
	}{
		Count:  2,
		Scenes: []stash.Scene{{ID: "1", Title: "Scene 1"}, {ID: "2", Title: "Scene 2"}},
	}
	ts := newGQLServer(t, "key-2", "findScenes", resp)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-2")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	scenes, count, err := c.FindScenes(context.Background(), &stash.FindFilter{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	if len(scenes) != 2 {
		t.Fatalf("len = %d, want 2", len(scenes))
	}
}

func TestFindPerformer(t *testing.T) {
	t.Parallel()

	p := stash.Performer{
		ID:       "p1",
		Name:     "Test Performer",
		Gender:   "FEMALE",
		Country:  "US",
		Favorite: true,
	}
	ts := newGQLServer(t, "key-3", "findPerformer", p)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-3")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindPerformer(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Test Performer" {
		t.Errorf("Name = %q, want %q", result.Name, "Test Performer")
	}
	if !result.Favorite {
		t.Error("Favorite should be true")
	}
}

func TestFindPerformers(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count      int               `json:"count"`
		Performers []stash.Performer `json:"performers"`
	}{
		Count:      1,
		Performers: []stash.Performer{{ID: "p1", Name: "Performer 1", SceneCount: 10}},
	}
	ts := newGQLServer(t, "key-4", "findPerformers", resp)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-4")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	performers, count, err := c.FindPerformers(context.Background(), &stash.FindFilter{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(performers) != 1 {
		t.Fatalf("len = %d, want 1", len(performers))
	}
	if performers[0].SceneCount != 10 {
		t.Errorf("SceneCount = %d, want 10", performers[0].SceneCount)
	}
}

func TestFindStudio(t *testing.T) {
	t.Parallel()

	studio := stash.Studio{
		ID:           "s1",
		Name:         "Test Studio",
		ParentStudio: &stash.Studio{ID: "sp1", Name: "Parent"},
		ChildStudios: []stash.Studio{{ID: "sc1", Name: "Child"}},
	}
	ts := newGQLServer(t, "key-5", "findStudio", studio)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-5")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindStudio(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Test Studio" {
		t.Errorf("Name = %q, want %q", result.Name, "Test Studio")
	}
	if result.ParentStudio == nil || result.ParentStudio.Name != "Parent" {
		t.Error("ParentStudio not parsed correctly")
	}
}

func TestFindStudios(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count   int            `json:"count"`
		Studios []stash.Studio `json:"studios"`
	}{
		Count:   1,
		Studios: []stash.Studio{{ID: "s1", Name: "Studio 1", SceneCount: 5}},
	}
	ts := newGQLServer(t, "key-6", "findStudios", resp)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-6")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	studios, count, err := c.FindStudios(context.Background(), &stash.FindFilter{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(studios) != 1 {
		t.Fatalf("len = %d, want 1", len(studios))
	}
}

func TestFindTag(t *testing.T) {
	t.Parallel()

	tag := stash.Tag{
		ID:          "t1",
		Name:        "Test Tag",
		Description: "A test tag",
		SceneCount:  15,
	}
	ts := newGQLServer(t, "key-7", "findTag", tag)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-7")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindTag(context.Background(), "t1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Test Tag" {
		t.Errorf("Name = %q, want %q", result.Name, "Test Tag")
	}
	if result.SceneCount != 15 {
		t.Errorf("SceneCount = %d, want 15", result.SceneCount)
	}
}

func TestFindTags(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count int         `json:"count"`
		Tags  []stash.Tag `json:"tags"`
	}{
		Count: 3,
		Tags:  []stash.Tag{{ID: "t1", Name: "Tag1"}, {ID: "t2", Name: "Tag2"}, {ID: "t3", Name: "Tag3"}},
	}
	ts := newGQLServer(t, "key-8", "findTags", resp)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-8")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	tags, count, err := c.FindTags(context.Background(), &stash.FindFilter{Page: 1, PerPage: 50})
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	if len(tags) != 3 {
		t.Fatalf("len = %d, want 3", len(tags))
	}
}

func TestFindGallery(t *testing.T) {
	t.Parallel()

	gallery := stash.Gallery{
		ID:         "g1",
		Title:      "Test Gallery",
		ImageCount: 42,
		Studio:     &stash.Studio{ID: "s1", Name: "Studio"},
	}
	ts := newGQLServer(t, "key-9", "findGallery", gallery)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-9")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindGallery(context.Background(), "g1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Title != "Test Gallery" {
		t.Errorf("Title = %q, want %q", result.Title, "Test Gallery")
	}
	if result.ImageCount != 42 {
		t.Errorf("ImageCount = %d, want 42", result.ImageCount)
	}
}

func TestFindGalleries(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count     int             `json:"count"`
		Galleries []stash.Gallery `json:"galleries"`
	}{
		Count:     1,
		Galleries: []stash.Gallery{{ID: "g1", Title: "Gallery 1"}},
	}
	ts := newGQLServer(t, "key-10", "findGalleries", resp)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-10")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	galleries, count, err := c.FindGalleries(context.Background(), &stash.FindFilter{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(galleries) != 1 {
		t.Fatalf("len = %d, want 1", len(galleries))
	}
}

func TestFindImage(t *testing.T) {
	t.Parallel()

	img := stash.Image{
		ID:    "i1",
		Title: "Test Image",
		Files: []stash.ImageFile{{Path: "/photos/test.jpg", Width: 1920, Height: 1080}},
	}
	ts := newGQLServer(t, "key-11", "findImage", img)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-11")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindImage(context.Background(), "i1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Title != "Test Image" {
		t.Errorf("Title = %q, want %q", result.Title, "Test Image")
	}
	if len(result.Files) != 1 || result.Files[0].Width != 1920 {
		t.Error("Files not parsed correctly")
	}
}

func TestFindImages(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count  int           `json:"count"`
		Images []stash.Image `json:"images"`
	}{
		Count:  1,
		Images: []stash.Image{{ID: "i1", Title: "Image 1"}},
	}
	ts := newGQLServer(t, "key-12", "findImages", resp)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-12")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	images, count, err := c.FindImages(context.Background(), &stash.FindFilter{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(images) != 1 {
		t.Fatalf("len = %d, want 1", len(images))
	}
}

func TestFindGroup(t *testing.T) {
	t.Parallel()

	group := stash.Group{
		ID:       "m1",
		Name:     "Test Movie",
		Duration: intPtr(7200),
		Director: "Test Director",
		Studio:   &stash.Studio{ID: "s1", Name: "Studio"},
	}
	ts := newGQLServer(t, "key-13", "findGroup", group)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-13")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindGroup(context.Background(), "m1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Test Movie" {
		t.Errorf("Name = %q, want %q", result.Name, "Test Movie")
	}
	if result.Duration == nil || *result.Duration != 7200 {
		t.Error("Duration not parsed correctly")
	}
}

func TestFindGroups(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count  int           `json:"count"`
		Groups []stash.Group `json:"groups"`
	}{
		Count:  1,
		Groups: []stash.Group{{ID: "m1", Name: "Group 1"}},
	}
	ts := newGQLServer(t, "key-14", "findGroups", resp)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-14")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	groups, count, err := c.FindGroups(context.Background(), &stash.FindFilter{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(groups) != 1 {
		t.Fatalf("len = %d, want 1", len(groups))
	}
}

func TestFindSceneMarkers(t *testing.T) {
	t.Parallel()

	markers := struct {
		Count        int                 `json:"count"`
		SceneMarkers []stash.SceneMarker `json:"scene_markers"`
	}{
		Count:        1,
		SceneMarkers: []stash.SceneMarker{{ID: "sm1", Title: "Marker 1", Seconds: 120.5, PrimaryTag: stash.Tag{ID: "t1", Name: "Position"}}},
	}
	ts := newGQLServer(t, "key-15", "findSceneMarkers", markers)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-15")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, count, err := c.FindSceneMarkers(context.Background(), &stash.FindFilter{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Seconds != 120.5 {
		t.Errorf("Seconds = %f, want 120.5", result[0].Seconds)
	}
}

func TestGetStats(t *testing.T) {
	t.Parallel()

	stats := stash.Stats{
		SceneCount:     100,
		PerformerCount: 50,
		StudioCount:    10,
		TagCount:       200,
	}
	ts := newGQLServer(t, "key-16", "stats", stats)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-16")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.GetStats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.SceneCount != 100 {
		t.Errorf("SceneCount = %d, want 100", result.SceneCount)
	}
	if result.PerformerCount != 50 {
		t.Errorf("PerformerCount = %d, want 50", result.PerformerCount)
	}
}

func TestGetVersion(t *testing.T) {
	t.Parallel()

	ver := stash.Version{Version: "0.25.1", Hash: "abc123", BuildType: "official"}
	ts := newGQLServer(t, "key-17", "version", ver)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-17")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.GetVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Version != "0.25.1" {
		t.Errorf("Version = %q, want %q", result.Version, "0.25.1")
	}
}

func TestGetSystemStatus(t *testing.T) {
	t.Parallel()

	status := stash.SystemStatus{
		Status: "OK",
		OS:     "linux",
	}
	ts := newGQLServer(t, "key-18", "systemStatus", status)
	defer ts.Close()

	c, err := stash.New(ts.URL, "key-18")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.GetSystemStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "OK" {
		t.Errorf("Status = %q, want %q", result.Status, "OK")
	}
}

func TestGraphQLError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"scene not found"}]}`))
	}))
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.FindScene(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var gqlErr *stash.GraphQLError
	if !errors.As(err, &gqlErr) {
		t.Fatalf("expected *GraphQLError, got %T", err)
	}
	if gqlErr.Message != "scene not found" {
		t.Errorf("Message = %q, want %q", gqlErr.Message, "scene not found")
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer ts.Close()

	c, err := stash.New(ts.URL, "bad-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.FindScene(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *stash.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestNoAuthRequired(t *testing.T) {
	t.Parallel()

	scene := stash.Scene{ID: "1", Title: "No Auth Scene"}
	ts := newGQLServer(t, "", "findScene", scene)
	defer ts.Close()

	c, err := stash.New(ts.URL, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindScene(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Title != "No Auth Scene" {
		t.Errorf("Title = %q, want %q", result.Title, "No Auth Scene")
	}
}

func intPtr(v int) *int { return &v }

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
			c, err := stash.New(tc.url, "k")
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}

func strPtr(v string) *string { return &v }

type capturedGQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func newMutationServer(t *testing.T, dataField string, response any, onReq func(req capturedGQLRequest)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		var req capturedGQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if onReq != nil {
			onReq(req)
		}
		data, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{` + `"` + dataField + `":` + string(data) + `}}`))
	}))
}

func assertGQLMessage(t *testing.T, err error, wantMsg string) {
	t.Helper()
	var gqlErr *stash.GraphQLError
	if !errors.As(err, &gqlErr) {
		t.Fatalf("expected *GraphQLError, got %T: %v", err, err)
	}
	if gqlErr.Message != wantMsg {
		t.Errorf("Message = %q, want %q", gqlErr.Message, wantMsg)
	}
}

func assertHTTPStatus(t *testing.T, err error, wantStatus int) {
	t.Helper()
	var apiErr *stash.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != wantStatus {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, wantStatus)
	}
}

func assertAbsent(t *testing.T, m map[string]any, key string) {
	t.Helper()
	if _, ok := m[key]; ok {
		t.Errorf("key %q should be absent, got %v", key, m[key])
	}
}

func assertPerformerIDs(t *testing.T, rawInput map[string]any, expected []string) {
	t.Helper()
	ids, ok := rawInput["performer_ids"].([]any)
	if !ok || len(ids) != len(expected) {
		t.Fatalf("performer_ids = %v, want %v", rawInput["performer_ids"], expected)
	}
	for i, exp := range expected {
		if ids[i] != exp {
			t.Errorf("performer_ids[%d] = %v, want %v", i, ids[i], exp)
		}
	}
}

func runSceneUpdateCapture(t *testing.T, input stash.SceneUpdateInput) map[string]any {
	t.Helper()
	var captured capturedGQLRequest
	ts := newMutationServer(t, "sceneUpdate", stash.Scene{ID: "s1"}, func(req capturedGQLRequest) {
		captured = req
	})
	t.Cleanup(ts.Close)

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := c.SceneUpdate(context.Background(), input); err != nil {
		t.Fatalf("SceneUpdate: %v", err)
	}
	raw, ok := captured.Variables["input"].(map[string]any)
	if !ok {
		t.Fatalf("expected map input, got %T", captured.Variables["input"])
	}
	return raw
}

// Stash v0.31.1 graphql/schema/schema.graphql:371.
// Stash v0.31.1 graphql/schema/types/performer.graphql:67.
func TestPerformerCreate_Success(t *testing.T) {
	t.Parallel()

	female := stash.GenderFemale
	input := stash.PerformerCreateInput{
		Name:      "Jane Doe",
		Gender:    &female,
		Rating100: intPtr(95),
		TagIDs:    []string{"t1"},
	}
	expected := stash.Performer{ID: "p1", Name: "Jane Doe"}

	var captured capturedGQLRequest
	ts := newMutationServer(t, "performerCreate", expected, func(req capturedGQLRequest) {
		captured = req
	})
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := c.PerformerCreate(context.Background(), input)
	if err != nil {
		t.Fatalf("PerformerCreate: %v", err)
	}
	if got.ID != "p1" || got.Name != "Jane Doe" {
		t.Errorf("got %+v, want %+v", got, expected)
	}
	if !strings.Contains(captured.Query, "mutation PerformerCreate") || !strings.Contains(captured.Query, "performerCreate(input: $input)") {
		t.Errorf("unexpected query: %q", captured.Query)
	}
	rawInput, ok := captured.Variables["input"].(map[string]any)
	if !ok || rawInput["name"] != "Jane Doe" || rawInput["gender"] != "FEMALE" {
		t.Errorf("unexpected variables input: %+v", captured.Variables)
	}
}

// Stash v0.31.1 graphql/schema/types/performer.graphql:68.
func TestPerformerCreate_Validation(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected HTTP request on validation failure")
	}))
	t.Cleanup(ts.Close)

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	cases := []struct {
		name  string
		input stash.PerformerCreateInput
	}{
		{"empty_name", stash.PerformerCreateInput{Name: ""}},
		{"whitespace_name", stash.PerformerCreateInput{Name: "   "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.PerformerCreate(context.Background(), tc.input)
			if !errors.Is(err, stash.ErrEmptyName) {
				t.Fatalf("got err %v, want ErrEmptyName", err)
			}
		})
	}
}

// Stash v0.31.1 graphql/schema/schema.graphql:371.
func TestPerformerCreate_GraphQLError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"performer already exists"}]}`))
	}))
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.PerformerCreate(context.Background(), stash.PerformerCreateInput{Name: "Valid"})
	assertGQLMessage(t, err, "performer already exists")
}

// Stash v0.31.1 graphql/schema/schema.graphql:371.
func TestPerformerCreate_HTTP500(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.PerformerCreate(context.Background(), stash.PerformerCreateInput{Name: "Valid"})
	assertHTTPStatus(t, err, http.StatusInternalServerError)
}

// Stash v0.31.1 graphql/schema/schema.graphql:372.
// Stash v0.31.1 graphql/schema/types/performer.graphql:107.
func TestPerformerUpdate_Success(t *testing.T) {
	t.Parallel()

	input := stash.PerformerUpdateInput{
		ID:        "p1",
		Name:      strPtr("Updated Performer"),
		Rating100: intPtr(99),
	}
	expected := stash.Performer{ID: "p1", Name: "Updated Performer"}

	var captured capturedGQLRequest
	ts := newMutationServer(t, "performerUpdate", expected, func(req capturedGQLRequest) {
		captured = req
	})
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := c.PerformerUpdate(context.Background(), input)
	if err != nil {
		t.Fatalf("PerformerUpdate: %v", err)
	}
	if got.ID != "p1" || got.Name != "Updated Performer" {
		t.Errorf("got %+v, want %+v", got, expected)
	}
	if !strings.Contains(captured.Query, "mutation PerformerUpdate") || !strings.Contains(captured.Query, "performerUpdate(input: $input)") {
		t.Errorf("unexpected query: %q", captured.Query)
	}
	rawInput, ok := captured.Variables["input"].(map[string]any)
	if !ok || rawInput["id"] != "p1" || rawInput["name"] != "Updated Performer" {
		t.Errorf("unexpected variables input: %+v", captured.Variables)
	}
}

// Stash v0.31.1 graphql/schema/types/performer.graphql:108.
func TestPerformerUpdate_Boundary(t *testing.T) {
	t.Parallel()

	var captured capturedGQLRequest
	ts := newMutationServer(t, "performerUpdate", stash.Performer{ID: "p1"}, func(req capturedGQLRequest) {
		captured = req
	})
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// all-optional-nil sends only id
	_, err = c.PerformerUpdate(context.Background(), stash.PerformerUpdateInput{ID: "p1"})
	if err != nil {
		t.Fatalf("PerformerUpdate: %v", err)
	}
	rawInput, ok := captured.Variables["input"].(map[string]any)
	if !ok {
		t.Fatalf("expected map input, got %T", captured.Variables["input"])
	}
	if len(rawInput) != 1 || rawInput["id"] != "p1" {
		t.Errorf("expected only id in input, got %+v", rawInput)
	}
}

// Stash v0.31.1 graphql/schema/types/performer.graphql:108.
func TestPerformerUpdate_Validation(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected HTTP request on validation failure")
	}))
	t.Cleanup(ts.Close)

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	cases := []struct {
		name  string
		input stash.PerformerUpdateInput
	}{
		{"empty_id", stash.PerformerUpdateInput{ID: ""}},
		{"whitespace_id", stash.PerformerUpdateInput{ID: "   "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.PerformerUpdate(context.Background(), tc.input)
			if !errors.Is(err, stash.ErrEmptyID) {
				t.Fatalf("got err %v, want ErrEmptyID", err)
			}
		})
	}
}

// Stash v0.31.1 graphql/schema/schema.graphql:372.
func TestPerformerUpdate_GraphQLError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"performer not found"}]}`))
	}))
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.PerformerUpdate(context.Background(), stash.PerformerUpdateInput{ID: "p1"})
	assertGQLMessage(t, err, "performer not found")
}

// Stash v0.31.1 graphql/schema/schema.graphql:372.
func TestPerformerUpdate_HTTP500(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server failure"))
	}))
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.PerformerUpdate(context.Background(), stash.PerformerUpdateInput{ID: "p1"})
	assertHTTPStatus(t, err, http.StatusInternalServerError)
}

// Stash v0.31.1 graphql/schema/schema.graphql:291.
// Stash v0.31.1 graphql/schema/types/scene.graphql:129.
// Stash v0.31.1 graphql/schema/types/scene.graphql:146.
func TestSceneUpdate_Success(t *testing.T) {
	t.Parallel()

	input := stash.SceneUpdateInput{
		ID:           "s1",
		Title:        strPtr("Updated Scene"),
		PerformerIDs: []string{"p1", "p2"},
	}
	expected := stash.Scene{ID: "s1", Title: "Updated Scene"}

	var captured capturedGQLRequest
	ts := newMutationServer(t, "sceneUpdate", expected, func(req capturedGQLRequest) {
		captured = req
	})
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := c.SceneUpdate(context.Background(), input)
	if err != nil {
		t.Fatalf("SceneUpdate: %v", err)
	}
	if got.ID != "s1" || got.Title != "Updated Scene" {
		t.Errorf("got %+v, want %+v", got, expected)
	}
	if !strings.Contains(captured.Query, "mutation SceneUpdate") {
		t.Errorf("unexpected query: %q", captured.Query)
	}
	rawInput, ok := captured.Variables["input"].(map[string]any)
	if !ok {
		t.Fatalf("expected map input, got %T", captured.Variables["input"])
	}
	if rawInput["id"] != "s1" || rawInput["title"] != "Updated Scene" {
		t.Errorf("unexpected variables input: %+v", captured.Variables)
	}
	assertPerformerIDs(t, rawInput, []string{"p1", "p2"})
}

// Stash v0.31.1 graphql/schema/types/scene.graphql:131.
// Stash v0.31.1 graphql/schema/types/scene.graphql:146.
func TestSceneUpdate_Boundary(t *testing.T) {
	t.Parallel()

	t.Run("all_optional_nil", func(t *testing.T) {
		t.Parallel()
		raw := runSceneUpdateCapture(t, stash.SceneUpdateInput{ID: "s1"})
		if len(raw) != 1 || raw["id"] != "s1" {
			t.Errorf("expected only id, got %+v", raw)
		}
	})

	t.Run("performer_ids_nil", func(t *testing.T) {
		t.Parallel()
		raw := runSceneUpdateCapture(t, stash.SceneUpdateInput{ID: "s1", PerformerIDs: nil})
		assertAbsent(t, raw, "performer_ids")
	})

	t.Run("performer_ids_empty", func(t *testing.T) {
		t.Parallel()
		raw := runSceneUpdateCapture(t, stash.SceneUpdateInput{ID: "s1", PerformerIDs: []string{}})
		// Why: an empty non-nil slice clears the scene's performers; nil leaves them unchanged.
		assertPerformerIDs(t, raw, []string{})
	})

	t.Run("performer_ids_populated", func(t *testing.T) {
		t.Parallel()
		raw := runSceneUpdateCapture(t, stash.SceneUpdateInput{ID: "s1", PerformerIDs: []string{"p1"}})
		assertPerformerIDs(t, raw, []string{"p1"})
	})
}

// Stash v0.31.1 graphql/schema/types/scene.graphql:131.
func TestSceneUpdate_Validation(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected HTTP request on validation failure")
	}))
	t.Cleanup(ts.Close)

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	cases := []struct {
		name  string
		input stash.SceneUpdateInput
	}{
		{"empty_id", stash.SceneUpdateInput{ID: ""}},
		{"whitespace_id", stash.SceneUpdateInput{ID: "   "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.SceneUpdate(context.Background(), tc.input)
			if !errors.Is(err, stash.ErrEmptyID) {
				t.Fatalf("got err %v, want ErrEmptyID", err)
			}
		})
	}
}

// Stash v0.31.1 graphql/schema/schema.graphql:291.
func TestSceneUpdate_GraphQLError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"scene not found"}]}`))
	}))
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.SceneUpdate(context.Background(), stash.SceneUpdateInput{ID: "s1"})
	assertGQLMessage(t, err, "scene not found")
}

// Stash v0.31.1 graphql/schema/schema.graphql:291.
func TestSceneUpdate_HTTP500(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("fatal server error"))
	}))
	defer ts.Close()

	c, err := stash.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.SceneUpdate(context.Background(), stash.SceneUpdateInput{ID: "s1"})
	assertHTTPStatus(t, err, http.StatusInternalServerError)
}
