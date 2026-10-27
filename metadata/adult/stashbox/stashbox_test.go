package stashbox_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/metadata"
	"github.com/golusoris/goenvoy/metadata/adult/stashbox/v2"
)

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

func newGQLServer(t *testing.T, wantKey, dataField string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if got := r.Header.Get("Apikey"); got != wantKey {
			t.Errorf("ApiKey = %q, want %q", got, wantKey)
		}
		var req gqlRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		data, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{` + `"` + dataField + `":` + string(data) + `}}`))
	}))
}

func TestFindPerformer(t *testing.T) {
	t.Parallel()

	p := stashbox.Performer{
		ID:      "abc-123",
		Name:    "Test Performer",
		Gender:  "FEMALE",
		Country: "US",
		Aliases: []string{"Alias1"},
	}
	ts := newGQLServer(t, "key-1", "findPerformer", p)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-1")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindPerformer(context.Background(), "abc-123")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Test Performer" {
		t.Errorf("Name = %q, want %q", result.Name, "Test Performer")
	}
	if result.Gender != "FEMALE" {
		t.Errorf("Gender = %q, want %q", result.Gender, "FEMALE")
	}
}

func TestQueryPerformers(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count      int                  `json:"count"`
		Performers []stashbox.Performer `json:"performers"`
	}{
		Count:      1,
		Performers: []stashbox.Performer{{ID: "p1", Name: "Performer 1"}},
	}
	ts := newGQLServer(t, "key-2", "queryPerformers", resp)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-2")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	performers, count, err := c.QueryPerformers(context.Background(), &stashbox.QueryInput{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(performers) != 1 {
		t.Fatalf("len = %d, want 1", len(performers))
	}
}

func TestSearchPerformers(t *testing.T) {
	t.Parallel()

	performers := []stashbox.Performer{{ID: "sp1", Name: "Searched Performer"}}
	ts := newGQLServer(t, "key-3", "searchPerformer", performers)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-3")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.SearchPerformers(context.Background(), "test", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Name != "Searched Performer" {
		t.Errorf("Name = %q, want %q", result[0].Name, "Searched Performer")
	}
}

func TestFindScene(t *testing.T) {
	t.Parallel()

	s := stashbox.Scene{
		ID:       "scene-1",
		Title:    "Test Scene",
		Duration: intPtr(3600),
		Studio:   &stashbox.Studio{ID: "studio-1", Name: "Test Studio"},
		Tags:     []stashbox.Tag{{ID: "tag-1", Name: "Tag1"}},
	}
	ts := newGQLServer(t, "key-4", "findScene", s)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-4")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindScene(context.Background(), "scene-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Title != "Test Scene" {
		t.Errorf("Title = %q, want %q", result.Title, "Test Scene")
	}
	if result.Studio == nil || result.Studio.Name != "Test Studio" {
		t.Error("Studio not parsed correctly")
	}
}

func TestQueryScenes(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count  int              `json:"count"`
		Scenes []stashbox.Scene `json:"scenes"`
	}{
		Count:  2,
		Scenes: []stashbox.Scene{{ID: "s1", Title: "Scene 1"}, {ID: "s2", Title: "Scene 2"}},
	}
	ts := newGQLServer(t, "key-5", "queryScenes", resp)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-5")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	scenes, count, err := c.QueryScenes(context.Background(), &stashbox.QueryInput{Page: 1, PerPage: 10})
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

func TestSearchScenes(t *testing.T) {
	t.Parallel()

	scenes := []stashbox.Scene{{ID: "ss1", Title: "Found Scene"}}
	ts := newGQLServer(t, "key-6", "searchScene", scenes)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-6")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.SearchScenes(context.Background(), "test query")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
}

func TestFindScenesByFingerprints(t *testing.T) {
	t.Parallel()

	scenes := [][]stashbox.Scene{
		{{ID: "fp-1", Title: "Fingerprint Match"}},
	}
	ts := newGQLServer(t, "key-7", "findScenesBySceneFingerprints", scenes)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-7")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	fp := [][]stashbox.FingerprintInput{{{Hash: "abc123", Algorithm: "MD5"}}}
	result, err := c.FindScenesByFingerprints(context.Background(), fp)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || len(result[0]) != 1 {
		t.Fatalf("unexpected result shape")
	}
	if result[0][0].Title != "Fingerprint Match" {
		t.Errorf("Title = %q, want %q", result[0][0].Title, "Fingerprint Match")
	}
}

func TestFindStudio(t *testing.T) {
	t.Parallel()

	studio := stashbox.Studio{
		ID:     "studio-1",
		Name:   "Test Studio",
		Parent: &stashbox.Studio{ID: "parent-1", Name: "Parent Studio"},
	}
	ts := newGQLServer(t, "key-8", "findStudio", studio)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-8")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindStudio(context.Background(), "studio-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Test Studio" {
		t.Errorf("Name = %q, want %q", result.Name, "Test Studio")
	}
	if result.Parent == nil || result.Parent.Name != "Parent Studio" {
		t.Error("Parent not parsed correctly")
	}
}

func TestQueryStudios(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count   int               `json:"count"`
		Studios []stashbox.Studio `json:"studios"`
	}{
		Count:   1,
		Studios: []stashbox.Studio{{ID: "qs1", Name: "Queried Studio"}},
	}
	ts := newGQLServer(t, "key-9", "queryStudios", resp)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-9")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	studios, count, err := c.QueryStudios(context.Background(), &stashbox.QueryInput{Page: 1, PerPage: 25})
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

func TestSearchStudios(t *testing.T) {
	t.Parallel()

	studios := []stashbox.Studio{{ID: "ss1", Name: "Searched Studio"}}
	ts := newGQLServer(t, "key-10", "searchStudio", studios)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-10")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.SearchStudios(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
}

func TestFindTag(t *testing.T) {
	t.Parallel()

	tag := stashbox.Tag{
		ID:          "tag-1",
		Name:        "Test Tag",
		Description: "A tag for testing",
		Category:    &stashbox.TagCategory{ID: "cat-1", Name: "Category"},
	}
	ts := newGQLServer(t, "key-11", "findTag", tag)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-11")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindTag(context.Background(), "tag-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Test Tag" {
		t.Errorf("Name = %q, want %q", result.Name, "Test Tag")
	}
	if result.Category == nil || result.Category.Name != "Category" {
		t.Error("Category not parsed correctly")
	}
}

func TestQueryTags(t *testing.T) {
	t.Parallel()

	resp := struct {
		Count int            `json:"count"`
		Tags  []stashbox.Tag `json:"tags"`
	}{
		Count: 3,
		Tags:  []stashbox.Tag{{ID: "t1", Name: "Tag1"}, {ID: "t2", Name: "Tag2"}, {ID: "t3", Name: "Tag3"}},
	}
	ts := newGQLServer(t, "key-12", "queryTags", resp)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-12")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	tags, count, err := c.QueryTags(context.Background(), &stashbox.QueryInput{Page: 1, PerPage: 50})
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

func TestSearchTags(t *testing.T) {
	t.Parallel()

	tags := []stashbox.Tag{{ID: "st1", Name: "Searched Tag"}}
	ts := newGQLServer(t, "key-13", "searchTag", tags)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-13")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.SearchTags(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
}

func TestListSites(t *testing.T) {
	t.Parallel()

	sites := []stashbox.Site{
		{ID: "site-1", Name: "Test Site", URL: "https://example.com"},
		{ID: "site-2", Name: "Other Site", URL: "https://other.com", ValidTypes: []string{"SCENE"}},
	}
	ts := newGQLServer(t, "key-14", "querySites", sites)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-14")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.ListSites(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0].Name != "Test Site" {
		t.Errorf("Name = %q, want %q", result[0].Name, "Test Site")
	}
}

func TestGetConfig(t *testing.T) {
	t.Parallel()

	cfg := stashbox.Config{
		Host:          "https://stashdb.org",
		RequireInvite: true,
		GuidelinesURL: "https://guidelines.stashdb.org",
	}
	ts := newGQLServer(t, "key-15", "getConfig", cfg)
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key-15")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.GetConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Host != "https://stashdb.org" {
		t.Errorf("Host = %q, want %q", result.Host, "https://stashdb.org")
	}
	if !result.RequireInvite {
		t.Error("RequireInvite should be true")
	}
}

func TestGraphQLError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"performer not found"}]}`))
	}))
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.FindPerformer(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var gqlErr *stashbox.GraphQLError
	if !errors.As(err, &gqlErr) {
		t.Fatalf("expected *GraphQLError, got %T", err)
	}
	if gqlErr.Message != "performer not found" {
		t.Errorf("Message = %q, want %q", gqlErr.Message, "performer not found")
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "bad-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.FindPerformer(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *stashbox.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestWithCustomHTTPClient(t *testing.T) {
	t.Parallel()

	p := stashbox.Performer{ID: "p1", Name: "Custom"}
	ts := newGQLServer(t, "custom-key", "findPerformer", p)
	defer ts.Close()

	custom := &http.Client{}
	c, err := stashbox.New(ts.URL, "custom-key", metadata.WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.FindPerformer(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Custom" {
		t.Errorf("Name = %q, want %q", result.Name, "Custom")
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
			c, err := stashbox.New(tc.url, "k")
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}

// Why: stub mirrors stash-box v0.10.3 schema.graphql favoritePerformer/favoriteStudio(id: ID!, favorite: Boolean!): Boolean!.
func newFavoriteServer(t *testing.T, wantField, wantID string, wantFav bool, body string, status int) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if got := r.Header.Get("Apikey"); got != "key" {
			t.Errorf("Apikey = %q, want %q", got, "key")
		}
		var req gqlRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		wantQuery := "mutation ($id: ID!, $favorite: Boolean!) { " + wantField + "(id: $id, favorite: $favorite) }"
		if req.Query != wantQuery {
			t.Errorf("query = %q, want %q", req.Query, wantQuery)
		}
		if got, ok := req.Variables["id"].(string); !ok || got != wantID {
			t.Errorf("variables.id = %#v, want %q", req.Variables["id"], wantID)
		}
		if got, ok := req.Variables["favorite"].(bool); !ok || got != wantFav {
			t.Errorf("variables.favorite = %#v (present=%v), want %v", req.Variables["favorite"], ok, wantFav)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(ts.Close)
	return ts, &calls
}

type favoriteMethod struct {
	name  string
	field string
	call  func(c *stashbox.Client, ctx context.Context, id string, fav bool) (bool, error)
}

func favoriteMethods() []favoriteMethod {
	return []favoriteMethod{
		{"performer", "favoritePerformer", (*stashbox.Client).FavoritePerformer},
		{"studio", "favoriteStudio", (*stashbox.Client).FavoriteStudio},
	}
}

type favoriteCase struct {
	name    string
	fav     bool
	body    string
	status  int
	want    bool
	wantErr func(error) bool
}

func favoriteCases(field string) []favoriteCase {
	missing := func(err error) bool { return errors.Is(err, stashbox.ErrMissingResult) }
	return []favoriteCase{
		{name: "set true", fav: true, body: `{"data":{"` + field + `":true}}`, status: http.StatusOK, want: true},
		{name: "unset sends explicit false", fav: false, body: `{"data":{"` + field + `":true}}`, status: http.StatusOK, want: true},
		{name: "server returns false", fav: true, body: `{"data":{"` + field + `":false}}`, status: http.StatusOK, want: false},
		{
			name: "graphql error", fav: true, status: http.StatusOK,
			body: `{"errors":[{"message":"not found"}],"data":null}`,
			wantErr: func(err error) bool {
				var ge *stashbox.GraphQLError
				return errors.As(err, &ge) && ge.Message == "not found"
			},
		},
		{
			name: "http 500", fav: true, status: http.StatusInternalServerError, body: `boom`,
			wantErr: func(err error) bool {
				var ae *stashbox.APIError
				return errors.As(err, &ae) && ae.StatusCode == http.StatusInternalServerError
			},
		},
		{name: "null result", fav: true, status: http.StatusOK, body: `{"data":{"` + field + `":null}}`, wantErr: missing},
		{name: "result field absent", fav: true, status: http.StatusOK, body: `{"data":{}}`, wantErr: missing},
	}
}

func runFavoriteCase(t *testing.T, m favoriteMethod, tt favoriteCase) {
	t.Helper()
	ts, _ := newFavoriteServer(t, m.field, "uuid-1", tt.fav, tt.body, tt.status)
	c, err := stashbox.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got, err := m.call(c, context.Background(), "uuid-1", tt.fav)
	if tt.wantErr != nil {
		if err == nil || !tt.wantErr(err) {
			t.Fatalf("err = %v, want matching error", err)
		}
		if !strings.Contains(err.Error(), "favorite "+m.name) {
			t.Errorf("err = %q, want operation context", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != tt.want {
		t.Errorf("result = %v, want %v", got, tt.want)
	}
}

func TestFavoriteMutations(t *testing.T) {
	t.Parallel()

	for _, m := range favoriteMethods() {
		for _, tt := range favoriteCases(m.field) {
			t.Run(m.name+"/"+tt.name, func(t *testing.T) {
				t.Parallel()
				runFavoriteCase(t, m, tt)
			})
		}
	}
}

func TestFavoriteEmptyIDSendsNoRequest(t *testing.T) {
	t.Parallel()

	for _, m := range favoriteMethods() {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			ts, calls := newFavoriteServer(t, m.field, "", true, `{"data":{}}`, http.StatusOK)
			c, err := stashbox.New(ts.URL, "key")
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if _, err := m.call(c, context.Background(), "", true); !errors.Is(err, stashbox.ErrEmptyID) {
				t.Fatalf("err = %v, want ErrEmptyID", err)
			}
			if n := calls.Load(); n != 0 {
				t.Errorf("requests = %d, want 0", n)
			}
		})
	}
}

func TestFavoriteContextDeadline(t *testing.T) {
	t.Parallel()

	for _, m := range favoriteMethods() {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			release := make(chan struct{})
			ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				select {
				case <-release:
				case <-r.Context().Done():
				}
			}))
			defer ts.Close()
			defer close(release)

			c, err := stashbox.New(ts.URL, "key")
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			if _, err := m.call(c, ctx, "uuid-1", true); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("err = %v, want context.DeadlineExceeded", err)
			}
		})
	}
}

type queryCapture struct {
	mu    sync.Mutex
	query string
}

func (qc *queryCapture) set(q string) {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	qc.query = q
}

func (qc *queryCapture) get() string {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	return qc.query
}

func checkPerformerQueryTokens(t *testing.T, query string) {
	t.Helper()
	forbidden := []string{"birthdate", "deathdate", "weight", "measurements"}
	for _, tok := range forbidden {
		re := regexp.MustCompile(`\b` + tok + `\b`)
		if re.MatchString(query) {
			t.Errorf("sent query contains forbidden token %q: %s", tok, query)
		}
	}
	required := []string{"birth_date", "merged_ids", "merged_into_id", "cup_size"}
	for _, tok := range required {
		re := regexp.MustCompile(`\b` + tok + `\b`)
		if !re.MatchString(query) {
			t.Errorf("sent query missing required token %q: %s", tok, query)
		}
	}
}

func TestPerformerQuery_Tokens(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		call func(context.Context, *stashbox.Client) error
	}{
		{
			name: "FindPerformer",
			call: func(ctx context.Context, c *stashbox.Client) error {
				_, err := c.FindPerformer(ctx, "00000000-0000-4000-8000-000000000001")
				return err
			},
		},
		{
			name: "QueryPerformers",
			call: func(ctx context.Context, c *stashbox.Client) error {
				_, _, err := c.QueryPerformers(ctx, &stashbox.QueryInput{Page: 1, PerPage: 10})
				return err
			},
		},
		{
			name: "SearchPerformers",
			call: func(ctx context.Context, c *stashbox.Client) error {
				_, err := c.SearchPerformers(ctx, "test", 10)
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var qc queryCapture
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req gqlRequest
				_ = json.NewDecoder(r.Body).Decode(&req)
				qc.set(req.Query)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":{"findPerformer":{"id":"p1"},"queryPerformers":{"count":0,"performers":[]},"searchPerformer":[]}}`))
			}))
			defer ts.Close()

			c, err := stashbox.New(ts.URL, "key")
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if err := tc.call(context.Background(), c); err != nil {
				t.Fatalf("call: %v", err)
			}
			checkPerformerQueryTokens(t, qc.get())
		})
	}
}

func assertPerformerV0103Strings(t *testing.T, res *stashbox.Performer) {
	t.Helper()
	if res.BirthDate != "1990-01-02" {
		t.Errorf("BirthDate = %q, want %q", res.BirthDate, "1990-01-02")
	}
	if res.DeathDate != "" {
		t.Errorf("DeathDate = %q, want empty", res.DeathDate)
	}
	if res.CupSize != "C" {
		t.Errorf("CupSize = %q, want %q", res.CupSize, "C")
	}
	if res.SceneCount != 42 {
		t.Errorf("SceneCount = %d, want 42", res.SceneCount)
	}
}

func assertPerformerV0103Sizes(t *testing.T, res *stashbox.Performer) {
	t.Helper()
	if res.BandSize == nil || *res.BandSize != 34 {
		t.Errorf("BandSize = %v, want 34", res.BandSize)
	}
	if res.WaistSize == nil || *res.WaistSize != 26 {
		t.Errorf("WaistSize = %v, want 26", res.WaistSize)
	}
	if res.HipSize == nil || *res.HipSize != 36 {
		t.Errorf("HipSize = %v, want 36", res.HipSize)
	}
}

func assertPerformerV0103Merged(t *testing.T, res *stashbox.Performer) {
	t.Helper()
	wantIDs := []string{
		"00000000-0000-4000-8000-000000000002",
		"00000000-0000-4000-8000-000000000003",
	}
	if !slices.Equal(res.MergedIDs, wantIDs) {
		t.Errorf("MergedIDs = %v, want %v", res.MergedIDs, wantIDs)
	}
	if res.MergedIntoID != "" {
		t.Errorf("MergedIntoID = %q, want empty", res.MergedIntoID)
	}
}

func assertPerformerV0103Fields(t *testing.T, res *stashbox.Performer) {
	t.Helper()
	assertPerformerV0103Strings(t, res)
	assertPerformerV0103Sizes(t, res)
	assertPerformerV0103Merged(t, res)
}

func TestFindPerformer_DecodeV0103(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/find_performer_v0103.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer ts.Close()

	c, err := stashbox.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := c.FindPerformer(context.Background(), "00000000-0000-4000-8000-000000000001")
	if err != nil {
		t.Fatalf("FindPerformer: %v", err)
	}
	assertPerformerV0103Fields(t, res)
}

func assertEmptyMergedIDs(t *testing.T, res *stashbox.Performer) {
	t.Helper()
	if len(res.MergedIDs) != 0 {
		t.Errorf("len(MergedIDs) = %d, want 0", len(res.MergedIDs))
	}
	if res.MergedIntoID != "" {
		t.Errorf("MergedIntoID = %q, want empty", res.MergedIntoID)
	}
}

func assertMergedIntoIDSet(t *testing.T, res *stashbox.Performer) {
	t.Helper()
	if res.MergedIntoID != "00000000-0000-4000-8000-000000000099" {
		t.Errorf("MergedIntoID = %q, want 00000000-0000-4000-8000-000000000099", res.MergedIntoID)
	}
	if len(res.MergedIDs) != 0 {
		t.Errorf("len(MergedIDs) = %d, want 0", len(res.MergedIDs))
	}
}

func assertNullMeasurements(t *testing.T, res *stashbox.Performer) {
	t.Helper()
	if res.BandSize != nil {
		t.Errorf("BandSize = %v, want nil", res.BandSize)
	}
	if res.WaistSize != nil {
		t.Errorf("WaistSize = %v, want nil", res.WaistSize)
	}
	if res.HipSize != nil {
		t.Errorf("HipSize = %v, want nil", res.HipSize)
	}
}

func TestFindPerformer_Boundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		jsonBody string
		assertFn func(t *testing.T, res *stashbox.Performer)
	}{
		{
			name:     "empty merged_ids array",
			jsonBody: `{"data":{"findPerformer":{"id":"p1","name":"P1","merged_ids":[],"merged_into_id":null,"scene_count":0}}}`,
			assertFn: assertEmptyMergedIDs,
		},
		{
			name:     "merged_into_id set",
			jsonBody: `{"data":{"findPerformer":{"id":"p1","name":"P1","merged_ids":[],"merged_into_id":"00000000-0000-4000-8000-000000000099","scene_count":0}}}`,
			assertFn: assertMergedIntoIDSet,
		},
		{
			name:     "null measurements pointers",
			jsonBody: `{"data":{"findPerformer":{"id":"p1","name":"P1","band_size":null,"waist_size":null,"hip_size":null,"scene_count":0}}}`,
			assertFn: assertNullMeasurements,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.jsonBody))
			}))
			defer ts.Close()

			c, err := stashbox.New(ts.URL, "key")
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			res, err := c.FindPerformer(context.Background(), "p1")
			if err != nil {
				t.Fatalf("FindPerformer: %v", err)
			}
			tc.assertFn(t, res)
		})
	}
}

func ExampleClient_FavoritePerformer() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"favoritePerformer":true}}`))
	}))
	defer ts.Close()

	client, err := stashbox.New(ts.URL, "your-api-key")
	if err != nil {
		fmt.Println(err)
		return
	}
	ok, err := client.FavoritePerformer(context.Background(), "performer-uuid", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ok)
	// Output: true
}
