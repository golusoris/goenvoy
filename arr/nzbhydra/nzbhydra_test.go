package nzbhydra_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/golusoris/goenvoy/arr/nzbhydra/v2"
)

const newznabXML = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:newznab="http://www.newznab.com/DTD/2010/feeds/attributes/">
  <channel>
    <item>
      <title>Test.NZB.Release-GRP</title>
      <guid>https://example.com/details/abc</guid>
      <link>https://example.com/getnzb/abc</link>
      <comments>https://example.com/details/abc#comments</comments>
      <pubDate>Mon, 01 Jan 2025 00:00:00 +0000</pubDate>
      <size>1073741824</size>
      <description>A test release</description>
      <newznab:attr name="category" value="2000"/>
      <newznab:attr name="indexer" value="NZBgeek"/>
    </item>
    <item>
      <title>Another.Release-GRP</title>
      <guid>https://example.com/details/def</guid>
      <link>https://example.com/getnzb/def</link>
      <pubDate>Tue, 02 Jan 2025 12:00:00 +0000</pubDate>
      <size>536870912</size>
      <description>Another release</description>
      <newznab:attr name="category" value="5000"/>
    </item>
  </channel>
</rss>`

const capsXML = `<?xml version="1.0" encoding="UTF-8"?>
<caps>
  <server title="NZBHydra2" image="https://example.com/logo.png"/>
  <limits max="100" default="50"/>
  <searching>
    <search available="yes"/>
    <tv-search available="yes"/>
    <movie-search available="yes"/>
    <book-search available="no"/>
  </searching>
  <categories>
    <category id="2000" name="Movies">
      <subcat id="2010" name="Movies/Foreign"/>
    </category>
    <category id="5000" name="TV"/>
  </categories>
</caps>`

func newXMLServer(t *testing.T, xmlBody string) *nzbhydra.Client {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("apikey"); got != "test-key" {
			t.Errorf("apikey = %q, want test-key", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(xmlBody))
	}))
	t.Cleanup(ts.Close)
	c, err := nzbhydra.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func newJSONPostServer(t *testing.T, wantPath string, response any) *nzbhydra.Client {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if wantPath != "" && r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("apikey"); got != "test-key" {
			t.Errorf("apikey = %q, want test-key", got)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/json") {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		w.Header().Set("Content-Type", "application/json")
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
	t.Cleanup(ts.Close)
	c, err := nzbhydra.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func newJSONGetServer(t *testing.T, wantPath string, response any) *nzbhydra.Client {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if wantPath != "" && r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("apikey"); got != "test-key" {
			t.Errorf("apikey = %q, want test-key", got)
		}
		w.Header().Set("Content-Type", "application/json")
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
	t.Cleanup(ts.Close)
	c, err := nzbhydra.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// Why: shares server setup and validates path and query parameters across search tests.
func newQueryValidatingServer(t *testing.T, wantPath string, wantQuery map[string]string) *nzbhydra.Client {
	t.Helper()
	const maxWantQuery = 16
	if len(wantQuery) > maxWantQuery {
		t.Fatalf("len(wantQuery) = %d, want <= %d", len(wantQuery), maxWantQuery)
	}
	wantQuery["apikey"] = "test-key"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		q := r.URL.Query()
		for k, want := range wantQuery {
			if got := q.Get(k); got != want {
				t.Errorf("%s = %q, want %q", k, got, want)
			}
		}
		w.Header().Set("Content-Type", "application/xml")
		if _, err := w.Write([]byte(newznabXML)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(ts.Close)
	c, err := nzbhydra.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// Why: shares server setup and request validation across stats tests.
func newStatsServer(t *testing.T, stats any) *nzbhydra.Client {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/api/stats" {
			t.Errorf("path = %q, want /api/stats", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("invalid request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			t.Errorf("encode stats: %v", err)
		}
	}))
	t.Cleanup(ts.Close)
	c, err := nzbhydra.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// TestSearch tests searching for releases with query parameters and category filters.
func TestSearch(t *testing.T) {
	t.Parallel()

	c := newQueryValidatingServer(t, "/api", map[string]string{
		"t":   "search",
		"q":   "ubuntu",
		"cat": "2000,5000",
	})

	results, err := c.Search(context.Background(), "ubuntu", []int{2000, 5000})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	wantFirst := nzbhydra.SearchResult{
		Title:       "Test.NZB.Release-GRP",
		GUID:        "https://example.com/details/abc",
		Link:        "https://example.com/getnzb/abc",
		Comments:    "https://example.com/details/abc#comments",
		PubDate:     "Mon, 01 Jan 2025 00:00:00 +0000",
		Size:        1073741824,
		Category:    "2000",
		Description: "A test release",
		Indexer:     "NZBgeek",
	}
	if results[0] != wantFirst {
		t.Errorf("results[0] = %+v, want %+v", results[0], wantFirst)
	}
}

func TestSearchNilCategories(t *testing.T) {
	t.Parallel()

	c := newXMLServer(t, newznabXML)
	results, err := c.Search(context.Background(), "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

// TestTVSearch tests searching for TV releases with query parameters and metadata IDs.
func TestTVSearch(t *testing.T) {
	t.Parallel()

	c := newQueryValidatingServer(t, "/api", map[string]string{
		"t":        "tvsearch",
		"season":   "3",
		"ep":       "5",
		"tvdbid":   "12345",
		"imdbid":   "tt0903747",
		"tmdbid":   "1396",
		"tvmazeid": "169",
		"rid":      "999",
	})

	results, err := c.TVSearch(context.Background(), "breaking bad", &nzbhydra.TVSearchOptions{
		Season:   3,
		Episode:  5,
		TVDBID:   "12345",
		IMDBID:   "tt0903747",
		TMDBID:   "1396",
		TVMazeID: "169",
		RID:      "999",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

func TestTVSearchMinimalParams(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Has("season") {
			t.Error("season should not be set")
		}
		if q.Has("ep") {
			t.Error("ep should not be set")
		}
		if q.Has("tvdbid") {
			t.Error("tvdbid should not be set")
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(newznabXML))
	}))
	defer ts.Close()

	c, err := nzbhydra.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.TVSearch(context.Background(), "test", &nzbhydra.TVSearchOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMovieSearch(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("t"); got != "movie" {
			t.Errorf("t = %q, want movie", got)
		}
		if got := q.Get("imdbid"); got != "tt1375666" {
			t.Errorf("imdbid = %q, want tt1375666", got)
		}
		if got := q.Get("tmdbid"); got != "27205" {
			t.Errorf("tmdbid = %q, want 27205", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(newznabXML))
	}))
	defer ts.Close()

	c, err := nzbhydra.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results, err := c.MovieSearch(context.Background(), "inception", "tt1375666", "27205")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

func TestBookSearch(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("t"); got != "book" {
			t.Errorf("t = %q, want book", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(newznabXML))
	}))
	defer ts.Close()

	c, err := nzbhydra.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.BookSearch(context.Background(), "dune")
	if err != nil {
		t.Fatal(err)
	}
}

// TestGetCapabilities tests retrieving and parsing server capabilities from XML.
func TestGetCapabilities(t *testing.T) {
	t.Parallel()

	c := newXMLServer(t, capsXML)
	caps, err := c.GetCapabilities(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := &nzbhydra.Capabilities{
		Server: nzbhydra.ServerInfo{Title: "NZBHydra2", Image: "https://example.com/logo.png"},
		Limits: nzbhydra.Limits{Max: 100, Default: 50},
		Searching: nzbhydra.Searching{
			SearchAvailable:      true,
			TVSearchAvailable:    true,
			MovieSearchAvailable: true,
			BookSearchAvailable:  false,
		},
		Categories: []nzbhydra.Category{
			{
				ID:   2000,
				Name: "Movies",
				SubCategories: []nzbhydra.SubCategory{
					{ID: 2010, Name: "Movies/Foreign"},
				},
			},
			{
				ID:   5000,
				Name: "TV",
			},
		},
	}
	if !reflect.DeepEqual(caps, want) {
		t.Errorf("GetCapabilities() = %+v, want %+v", caps, want)
	}
}

// TestGetStats tests retrieving server statistics from the JSON endpoint.
func TestGetStats(t *testing.T) {
	t.Parallel()

	stats := map[string]any{
		"avgResponseTimes":      []map[string]any{{"indexer": "NZBgeek", "avgResponseTime": 200}},
		"indexerApiAccessStats": []map[string]any{{"indexer": "NZBgeek", "successful": 100, "unsuccessful": 5}},
		"searchesPerDayOfWeek":  map[string]any{"Monday": 10, "Tuesday": 15},
		"downloadsPerDayOfWeek": map[string]any{"Monday": 5},
	}
	c := newStatsServer(t, stats)

	result, err := c.GetStats(context.Background(), nzbhydra.StatsRequest{
		IncludeAverageResponseTimes: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := &nzbhydra.StatsResponse{
		AvgResponseTimes: []nzbhydra.AvgResponseTime{
			{Indexer: "NZBgeek", Avg: 200},
		},
		IndexerAPIAccessStats: []nzbhydra.IndexerAPIStat{
			{Indexer: "NZBgeek", Successful: 100, Unsuccessful: 5},
		},
		SearchesPerDayOfWeek:  map[string]int{"Monday": 10, "Tuesday": 15},
		DownloadsPerDayOfWeek: map[string]int{"Monday": 5},
	}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("GetStats() = %+v, want %+v", result, want)
	}
}

func TestGetSearchHistory(t *testing.T) {
	t.Parallel()

	resp := map[string]any{
		"content": []map[string]any{
			{"id": 1, "source": "INTERNAL", "searchType": "SEARCH", "time": "2025-01-01T00:00:00Z", "query": "test", "ip": "127.0.0.1"},
			{"id": 2, "source": "API", "searchType": "TVSEARCH", "time": "2025-01-02T00:00:00Z", "query": "breaking bad", "season": "5", "episode": "3"},
		},
		"totalElements":    2,
		"totalPages":       1,
		"first":            true,
		"last":             true,
		"numberOfElements": 2,
		"number":           0,
		"size":             25,
	}
	c := newJSONPostServer(t, "/api/history/searches", resp)

	result, err := c.GetSearchHistory(context.Background(), nzbhydra.HistoryRequest{
		Page: 0, Limit: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Content) != 2 {
		t.Fatalf("len(Content) = %d, want 2", len(result.Content))
	}
	if result.Content[0].Source != "INTERNAL" {
		t.Errorf("Source = %q, want INTERNAL", result.Content[0].Source)
	}
	if result.Content[1].Query != "breaking bad" {
		t.Errorf("Query = %q, want breaking bad", result.Content[1].Query)
	}
	if result.TotalElements != 2 {
		t.Errorf("TotalElements = %d, want 2", result.TotalElements)
	}
	if !result.First {
		t.Error("First = false, want true")
	}
	if !result.Last {
		t.Error("Last = false, want true")
	}
}

func TestGetDownloadHistory(t *testing.T) {
	t.Parallel()

	resp := map[string]any{
		"content": []map[string]any{
			{
				"id":            1,
				"searchResult":  map[string]any{"title": "Test.Release", "indexer": "NZBgeek", "link": "https://example.com/nzb"},
				"nzbAccessType": "PROXY",
				"accessSource":  "INTERNAL",
				"time":          "2025-01-01T00:00:00Z",
				"status":        "CONTENT_DOWNLOAD_SUCCESSFUL",
				"age":           5,
			},
		},
		"totalElements":    1,
		"totalPages":       1,
		"first":            true,
		"last":             true,
		"numberOfElements": 1,
		"number":           0,
		"size":             25,
	}
	c := newJSONPostServer(t, "/api/history/downloads", resp)

	result, err := c.GetDownloadHistory(context.Background(), nzbhydra.HistoryRequest{
		Page: 0, Limit: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Content) != 1 {
		t.Fatalf("len(Content) = %d, want 1", len(result.Content))
	}
	if result.Content[0].SearchResult.Title != "Test.Release" {
		t.Errorf("Title = %q, want Test.Release", result.Content[0].SearchResult.Title)
	}
	if result.Content[0].NzbAccessType != "PROXY" {
		t.Errorf("NzbAccessType = %q, want PROXY", result.Content[0].NzbAccessType)
	}
	if result.Content[0].Status != "CONTENT_DOWNLOAD_SUCCESSFUL" {
		t.Errorf("Status = %q, want CONTENT_DOWNLOAD_SUCCESSFUL", result.Content[0].Status)
	}
	if result.Content[0].Age != 5 {
		t.Errorf("Age = %d, want 5", result.Content[0].Age)
	}
}

func TestGetIndexerStatuses(t *testing.T) {
	t.Parallel()

	statuses := []map[string]any{
		{"indexer": "NZBgeek", "state": "ENABLED", "level": "NORMAL", "disabledUntil": "", "lastError": ""},
		{"indexer": "DrunkenSlug", "state": "DISABLED", "level": "ERROR", "disabledUntil": "2025-01-02T00:00:00Z", "lastError": "Connection timeout"},
	}
	c := newJSONGetServer(t, "/api/stats/indexers", statuses)

	result, err := c.GetIndexerStatuses(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].Indexer != "NZBgeek" {
		t.Errorf("Indexer = %q, want NZBgeek", result[0].Indexer)
	}
	if result[0].State != "ENABLED" {
		t.Errorf("State = %q, want ENABLED", result[0].State)
	}
	if result[1].State != "DISABLED" {
		t.Errorf("State = %q, want DISABLED", result[1].State)
	}
	if result[1].LastError != "Connection timeout" {
		t.Errorf("LastError = %q, want Connection timeout", result[1].LastError)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer ts.Close()

	c, err := nzbhydra.New(ts.URL, "bad-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.Search(context.Background(), "test", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *nzbhydra.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
	if apiErr.Body != "Unauthorized" {
		t.Errorf("Body = %q, want Unauthorized", apiErr.Body)
	}
}

func TestAPIErrorNoBody(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c, err := nzbhydra.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetIndexerStatuses(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *nzbhydra.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !strings.Contains(apiErr.Error(), "HTTP 500") {
		t.Errorf("error = %q, want HTTP 500 substring", apiErr.Error())
	}
}

func TestPostAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Forbidden"))
	}))
	defer ts.Close()

	c, err := nzbhydra.New(ts.URL, "bad-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetStats(context.Background(), nzbhydra.StatsRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *nzbhydra.APIError
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

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(newznabXML))
	}))
	defer ts.Close()

	c, err := nzbhydra.New(ts.URL, "k", nzbhydra.WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, _ = c.Search(context.Background(), "test", nil)
	if !called {
		t.Error("custom HTTP client was not used")
	}
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	c, err := nzbhydra.New("http://localhost:5076", "key", nzbhydra.WithTimeout(5*1e9))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_ = c
}

func TestEmptySearchResults(t *testing.T) {
	t.Parallel()

	emptyXML := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel></channel></rss>`
	c := newXMLServer(t, emptyXML)
	results, err := c.Search(context.Background(), "nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(results))
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
			c, err := nzbhydra.New(tc.url, "k")
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}
