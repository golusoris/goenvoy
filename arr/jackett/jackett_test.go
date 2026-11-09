package jackett_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/arr/jackett/v2"
)

const torznabXML = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed">
  <channel>
    <item>
      <title>Ubuntu 24.04</title>
      <guid>https://example.com/details/123</guid>
      <link>https://example.com/download/123</link>
      <comments>https://example.com/details/123#comments</comments>
      <pubDate>Mon, 01 Jan 2025 00:00:00 +0000</pubDate>
      <size>4294967296</size>
      <torznab:attr name="category" value="4000"/>
      <torznab:attr name="seeders" value="150"/>
      <torznab:attr name="peers" value="42"/>
      <torznab:attr name="infohash" value="abc123def456"/>
      <torznab:attr name="magneturl" value="magnet:?xt=urn:btih:abc123def456"/>
      <torznab:attr name="minimumratio" value="1"/>
      <torznab:attr name="minimumseedtime" value="172800"/>
      <torznab:attr name="downloadvolumefactor" value="0"/>
      <torznab:attr name="uploadvolumefactor" value="1"/>
      <torznab:attr name="indexer" value="TestIndexer"/>
    </item>
    <item>
      <title>Ubuntu 22.04</title>
      <guid>https://example.com/details/456</guid>
      <link>https://example.com/download/456</link>
      <pubDate>Tue, 02 Jan 2025 12:00:00 +0000</pubDate>
      <size>3221225472</size>
      <torznab:attr name="category" value="4000"/>
      <torznab:attr name="seeders" value="75"/>
      <torznab:attr name="peers" value="20"/>
    </item>
  </channel>
</rss>`

const capsXML = `<?xml version="1.0" encoding="UTF-8"?>
<caps>
  <server title="Jackett" image="https://example.com/logo.png"/>
  <limits max="100" default="50"/>
  <searching>
    <search available="yes"/>
    <tv-search available="yes"/>
    <movie-search available="yes"/>
    <music-search available="no"/>
    <book-search available="yes"/>
  </searching>
  <categories>
    <category id="2000" name="Movies">
      <subcat id="2010" name="Movies/Foreign"/>
      <subcat id="2020" name="Movies/Other"/>
    </category>
    <category id="5000" name="TV">
      <subcat id="5010" name="TV/WEB-DL"/>
    </category>
  </categories>
</caps>`

func newXMLServer(t *testing.T, wantPath, xmlBody string) *jackett.Client {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wantPath != "" && !strings.HasPrefix(r.URL.Path, wantPath) {
			t.Errorf("path = %q, want prefix %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("apikey"); got != "test-key" {
			t.Errorf("apikey = %q, want test-key", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(xmlBody))
	}))
	t.Cleanup(ts.Close)
	c, err := jackett.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func newJSONServer(t *testing.T, wantPath string, response any) *jackett.Client {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wantPath != "" && !strings.HasPrefix(r.URL.Path, wantPath) {
			t.Errorf("path = %q, want prefix %q", r.URL.Path, wantPath)
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
	c, err := jackett.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestSearch(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("t"); got != "search" {
			t.Errorf("t = %q, want search", got)
		}
		if got := q.Get("q"); got != "ubuntu" {
			t.Errorf("q = %q, want ubuntu", got)
		}
		if got := q.Get("cat"); got != "4000,5000" {
			t.Errorf("cat = %q, want 4000,5000", got)
		}
		if !strings.HasPrefix(r.URL.Path, "/api/v2.0/indexers/all/results/torznab") {
			t.Errorf("path = %q, want torznab path", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(torznabXML))
	}))
	defer ts.Close()

	c, err := jackett.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results, err := c.Search(context.Background(), "ubuntu", []int{4000, 5000})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	r := results[0]
	if r.Title != "Ubuntu 24.04" {
		t.Errorf("Title = %q, want Ubuntu 24.04", r.Title)
	}
	if r.Size != 4294967296 {
		t.Errorf("Size = %d, want 4294967296", r.Size)
	}
	if r.Seeders != 150 {
		t.Errorf("Seeders = %d, want 150", r.Seeders)
	}
	if r.Peers != 42 {
		t.Errorf("Peers = %d, want 42", r.Peers)
	}
	if r.InfoHash != "abc123def456" {
		t.Errorf("InfoHash = %q, want abc123def456", r.InfoHash)
	}
	if r.MagnetURL != "magnet:?xt=urn:btih:abc123def456" {
		t.Errorf("MagnetURL = %q", r.MagnetURL)
	}
	if r.Category != "4000" {
		t.Errorf("Category = %q, want 4000", r.Category)
	}
	if r.Indexer != "TestIndexer" {
		t.Errorf("Indexer = %q, want TestIndexer", r.Indexer)
	}
	if r.DownloadVolumeFactor != "0" {
		t.Errorf("DownloadVolumeFactor = %q, want 0", r.DownloadVolumeFactor)
	}
}

func TestSearchNilCategories(t *testing.T) {
	t.Parallel()

	c := newXMLServer(t, "/api/v2.0/indexers/all/results/torznab", torznabXML)
	results, err := c.Search(context.Background(), "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

func TestSearchIndexer(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/indexers/myindexer/") {
			t.Errorf("path = %q, want myindexer in path", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(torznabXML))
	}))
	defer ts.Close()

	c, err := jackett.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results, err := c.SearchIndexer(context.Background(), "myindexer", "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

func TestTVSearch(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("t"); got != "tvsearch" {
			t.Errorf("t = %q, want tvsearch", got)
		}
		if got := q.Get("q"); got != "breaking bad" {
			t.Errorf("q = %q, want breaking bad", got)
		}
		if got := q.Get("season"); got != "5" {
			t.Errorf("season = %q, want 5", got)
		}
		if got := q.Get("ep"); got != "3" {
			t.Errorf("ep = %q, want 3", got)
		}
		if got := q.Get("imdbid"); got != "tt0903747" {
			t.Errorf("imdbid = %q, want tt0903747", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(torznabXML))
	}))
	defer ts.Close()

	c, err := jackett.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results, err := c.TVSearch(context.Background(), "breaking bad", 5, 3, "tt0903747")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

func TestTVSearchOptionalParams(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Has("season") {
			t.Error("season should not be set when 0")
		}
		if q.Has("ep") {
			t.Error("ep should not be set when 0")
		}
		if q.Has("imdbid") {
			t.Error("imdbid should not be set when empty")
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(torznabXML))
	}))
	defer ts.Close()

	c, err := jackett.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.TVSearch(context.Background(), "test", 0, 0, "")
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
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(torznabXML))
	}))
	defer ts.Close()

	c, err := jackett.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results, err := c.MovieSearch(context.Background(), "inception", "tt1375666")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

func TestMusicSearch(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("t"); got != "music" {
			t.Errorf("t = %q, want music", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(torznabXML))
	}))
	defer ts.Close()

	c, err := jackett.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.MusicSearch(context.Background(), "radiohead")
	if err != nil {
		t.Fatal(err)
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
		_, _ = w.Write([]byte(torznabXML))
	}))
	defer ts.Close()

	c, err := jackett.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.BookSearch(context.Background(), "dune")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetCapabilities(t *testing.T) {
	t.Parallel()

	c := newXMLServer(t, "/api/v2.0/indexers/all/results/torznab", capsXML)
	caps, err := c.GetCapabilities(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if caps.Server.Title != "Jackett" {
		t.Errorf("Server.Title = %q, want Jackett", caps.Server.Title)
	}
	if caps.Server.Image != "https://example.com/logo.png" {
		t.Errorf("Server.Image = %q", caps.Server.Image)
	}
	if caps.Limits.Max != 100 {
		t.Errorf("Limits.Max = %d, want 100", caps.Limits.Max)
	}
	if caps.Limits.Default != 50 {
		t.Errorf("Limits.Default = %d, want 50", caps.Limits.Default)
	}
	if !caps.Searching.SearchAvailable {
		t.Error("SearchAvailable = false, want true")
	}
	if !caps.Searching.TVSearchAvailable {
		t.Error("TVSearchAvailable = false, want true")
	}
	if !caps.Searching.MovieSearchAvailable {
		t.Error("MovieSearchAvailable = false, want true")
	}
	if caps.Searching.MusicSearchAvailable {
		t.Error("MusicSearchAvailable = true, want false")
	}
	if !caps.Searching.BookSearchAvailable {
		t.Error("BookSearchAvailable = false, want true")
	}
	if len(caps.Categories) != 2 {
		t.Fatalf("len(Categories) = %d, want 2", len(caps.Categories))
	}
	if caps.Categories[0].ID != 2000 {
		t.Errorf("Categories[0].ID = %d, want 2000", caps.Categories[0].ID)
	}
	if caps.Categories[0].Name != "Movies" {
		t.Errorf("Categories[0].Name = %q, want Movies", caps.Categories[0].Name)
	}
	if len(caps.Categories[0].SubCategories) != 2 {
		t.Fatalf("len(SubCategories) = %d, want 2", len(caps.Categories[0].SubCategories))
	}
	if caps.Categories[0].SubCategories[0].ID != 2010 {
		t.Errorf("SubCategories[0].ID = %d, want 2010", caps.Categories[0].SubCategories[0].ID)
	}
}

func TestGetIndexers(t *testing.T) {
	t.Parallel()

	indexers := []map[string]any{
		{"id": "rarbg", "name": "RARBG", "type": "public", "configured": true, "language": "en-US", "site_link": "https://rarbg.to"},
		{"id": "1337x", "name": "1337x", "type": "public", "configured": true, "language": "en-US", "site_link": "https://1337x.to"},
	}
	c := newJSONServer(t, "/api/v2.0/indexers", indexers)

	result, err := c.GetIndexers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].ID != "rarbg" {
		t.Errorf("ID = %q, want rarbg", result[0].ID)
	}
	if result[0].Name != "RARBG" {
		t.Errorf("Name = %q, want RARBG", result[0].Name)
	}
	if !result[0].Configured {
		t.Error("Configured = false, want true")
	}
}

func TestGetServerConfig(t *testing.T) {
	t.Parallel()

	config := map[string]any{
		"api_key": "abc123", "blackholedir": "/downloads", "port": 9117, "instance_id": "inst-1",
	}
	c := newJSONServer(t, "/api/v2.0/server/config", config)

	result, err := c.GetServerConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.APIKey != "abc123" {
		t.Errorf("APIKey = %q, want abc123", result.APIKey)
	}
	if result.Port != 9117 {
		t.Errorf("Port = %d, want 9117", result.Port)
	}
	if result.BlackholeDir != "/downloads" {
		t.Errorf("BlackholeDir = %q, want /downloads", result.BlackholeDir)
	}
	if result.InstanceID != "inst-1" {
		t.Errorf("InstanceID = %q, want inst-1", result.InstanceID)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer ts.Close()

	c, err := jackett.New(ts.URL, "bad-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.Search(context.Background(), "test", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *jackett.APIError
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

	c, err := jackett.New(ts.URL, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetIndexers(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *jackett.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !strings.Contains(apiErr.Error(), "HTTP 500") {
		t.Errorf("error = %q, want HTTP 500 substring", apiErr.Error())
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
		_, _ = w.Write([]byte(torznabXML))
	}))
	defer ts.Close()

	c, err := jackett.New(ts.URL, "k", jackett.WithHTTPClient(custom))
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

	c, err := jackett.New("http://localhost:9117", "key", jackett.WithTimeout(5*1e9))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Just verify it doesn't panic.
	_ = c
}

func TestEmptySearchResults(t *testing.T) {
	t.Parallel()

	emptyXML := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel></channel></rss>`
	c := newXMLServer(t, "", emptyXML)
	results, err := c.Search(context.Background(), "nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(results))
	}
}

// roundTripFunc adapts a function to the http.RoundTripper interface.
type roundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip executes the round trip function.
func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestNew_invalidURL verifies that New rejects malformed URLs and URLs with query or fragment.
func TestNew_invalidURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, url string
	}{
		{"empty", ""},
		{"malformed", "://x"},
		{"ftp", "ftp://x"},
		{"no-scheme", "no-scheme"},
		{"query", "http://h:9117?apikey=K9"},
		{"empty-query", "http://h:9117/?"},
		{"fragment", "http://h:9117#x"},
		{"path-query", "http://h:9117/jackett?x=1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := jackett.New(tc.url, "k")
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
			if strings.Contains(err.Error(), "K9") {
				t.Errorf("New(%q) error leaked key: %q", tc.url, err.Error())
			}
		})
	}
}

// runNumericAttrCase executes a search against a mock XML server and asserts the parsed seeders and peers counts.
func runNumericAttrCase(t *testing.T, xmlBody string, wantSeeders, wantPeers int) {
	t.Helper()
	c := newXMLServer(t, "", xmlBody)
	results, err := c.Search(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Seeders != wantSeeders {
		t.Errorf("Seeders = %d, want %d", results[0].Seeders, wantSeeders)
	}
	if results[0].Peers != wantPeers {
		t.Errorf("Peers = %d, want %d", results[0].Peers, wantPeers)
	}
}

type numericAttrTestCase struct {
	name        string
	xml         string
	wantSeeders int
	wantPeers   int
}

// TestApplyNumericAttr_DefaultZero verifies that malformed, empty, or absent attributes default to zero.
func TestApplyNumericAttr_DefaultZero(t *testing.T) {
	t.Parallel()
	cases := []numericAttrTestCase{
		{
			name: "malformed value yields 0",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed"><channel><item>
  <title>T</title>
  <torznab:attr name="seeders" value="x"/>
  <torznab:attr name="peers" value="invalid"/>
</item></channel></rss>`,
			wantSeeders: 0,
			wantPeers:   0,
		},
		{
			name: "empty value yields 0",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed"><channel><item>
  <title>T</title>
  <torznab:attr name="seeders" value=""/>
  <torznab:attr name="peers" value=""/>
</item></channel></rss>`,
			wantSeeders: 0,
			wantPeers:   0,
		},
		{
			name: "absent attr yields 0",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed"><channel><item>
  <title>T</title>
</item></channel></rss>`,
			wantSeeders: 0,
			wantPeers:   0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runNumericAttrCase(t, tc.xml, tc.wantSeeders, tc.wantPeers)
		})
	}
}

// TestApplyNumericAttr_ValidAndBoundary verifies parsing of valid, negative, zero, and boundary attribute values.
func TestApplyNumericAttr_ValidAndBoundary(t *testing.T) {
	t.Parallel()
	cases := []numericAttrTestCase{
		{
			name: "valid positive values",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed"><channel><item>
  <title>T</title>
  <torznab:attr name="seeders" value="150"/>
  <torznab:attr name="peers" value="42"/>
</item></channel></rss>`,
			wantSeeders: 150,
			wantPeers:   42,
		},
		{
			name: "negative input parses",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed"><channel><item>
  <title>T</title>
  <torznab:attr name="seeders" value="-1"/>
  <torznab:attr name="peers" value="-10"/>
</item></channel></rss>`,
			wantSeeders: -1,
			wantPeers:   -10,
		},
		{
			name: "boundary zero",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed"><channel><item>
  <title>T</title>
  <torznab:attr name="seeders" value="0"/>
  <torznab:attr name="peers" value="0"/>
</item></channel></rss>`,
			wantSeeders: 0,
			wantPeers:   0,
		},
		{
			name: "boundary max int32",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed"><channel><item>
  <title>T</title>
  <torznab:attr name="seeders" value="2147483647"/>
  <torznab:attr name="peers" value="2147483647"/>
</item></channel></rss>`,
			wantSeeders: 2147483647,
			wantPeers:   2147483647,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runNumericAttrCase(t, tc.xml, tc.wantSeeders, tc.wantPeers)
		})
	}
}

// closedServerURL returns the URL of a server that is already closed to force connection errors.
func closedServerURL(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	srv.Close()
	return srv.URL
}

// hangingServerURL returns the URL of a server whose handler hangs until canceled.
func hangingServerURL(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
			t.Errorf("handler timeout waiting for request cancellation")
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// liveServerURL returns the URL of an active server that responds with 200 OK.
func liveServerURL(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// assertRedactedTransportError verifies that err is a redacted transport error not leaking key.
func assertRedactedTransportError(t *testing.T, err error, key string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, key) {
		t.Errorf("error leaked raw key: %q", errMsg)
	}
	if strings.Contains(errMsg, url.QueryEscape(key)) {
		t.Errorf("error leaked escaped key: %q", errMsg)
	}
	if !strings.Contains(errMsg, "apikey=REDACTED") {
		t.Errorf("error missing apikey=REDACTED: %q", errMsg)
	}
	if !strings.HasPrefix(errMsg, "jackett: request: ") {
		t.Errorf("error does not start with %q: %q", "jackett: request: ", errMsg)
	}
	var uerr *url.Error
	if !errors.As(err, &uerr) {
		t.Errorf("expected *url.Error in chain, got %T", err)
	}
}

// transportTestCase defines inputs and expectations for transport error redaction tests.
type transportTestCase struct {
	name     string
	baseURL  func(t *testing.T) string
	opts     []jackett.Option
	canceled bool
	call     func(ctx context.Context, c *jackett.Client) error
}

// runTransportErrorCase executes a single transport error test case and verifies error redaction.
func runTransportErrorCase(t *testing.T, tc transportTestCase, key string) {
	t.Helper()
	c, err := jackett.New(tc.baseURL(t), key, tc.opts...)
	if err != nil {
		t.Fatalf("jackett.New: %v", err)
	}
	ctx := context.Background()
	if tc.canceled {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		cancel()
	}
	err = tc.call(ctx, c)
	assertRedactedTransportError(t, err, key)
	if tc.canceled && !errors.Is(err, context.Canceled) {
		t.Errorf("expected errors.Is(err, context.Canceled), got %v", err)
	}
}

// TestTransportErrorRedactsAPIKey verifies that transport errors do not leak the API key.
func TestTransportErrorRedactsAPIKey(t *testing.T) {
	t.Parallel()

	const secretKey = `s3cr3t+key/=&x`
	cases := []transportTestCase{
		{
			name:    "search torznab path on closed server",
			baseURL: closedServerURL,
			call: func(ctx context.Context, c *jackett.Client) error {
				_, err := c.Search(ctx, "test", nil)
				return err
			},
		},
		{
			name:    "capabilities xml path on closed server",
			baseURL: closedServerURL,
			call: func(ctx context.Context, c *jackett.Client) error {
				_, err := c.GetCapabilities(ctx)
				return err
			},
		},
		{
			name:    "indexers json path on closed server",
			baseURL: closedServerURL,
			call: func(ctx context.Context, c *jackett.Client) error {
				_, err := c.GetIndexers(ctx)
				return err
			},
		},
		{
			name:     "canceled context on live server",
			baseURL:  liveServerURL,
			canceled: true,
			call: func(ctx context.Context, c *jackett.Client) error {
				_, err := c.GetServerConfig(ctx)
				return err
			},
		},
		{
			name:    "timeout on hanging handler",
			baseURL: hangingServerURL,
			opts:    []jackett.Option{jackett.WithTimeout(20 * time.Millisecond)},
			call: func(ctx context.Context, c *jackett.Client) error {
				_, err := c.Search(ctx, "test", nil)
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runTransportErrorCase(t, tc, secretKey)
		})
	}
}

// timeoutError represents a test error that reports a timeout.
type timeoutError struct {
	msg string
}

// Error returns the error message.
func (e timeoutError) Error() string {
	return e.msg
}

// Timeout returns true.
func (timeoutError) Timeout() bool {
	return true
}

// customTransportTestCase defines inputs and expectations for custom transport error tests.
type customTransportTestCase struct {
	name         string
	tripErr      func(r *http.Request) error
	checkTimeout bool
}

// runCustomTransportErrorCase executes a custom transport error case and verifies redaction.
func runCustomTransportErrorCase(t *testing.T, tc customTransportTestCase, key string) {
	t.Helper()
	c, err := jackett.New("http://127.0.0.1:9117", key, jackett.WithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, tc.tripErr(r)
		}),
	}))
	if err != nil {
		t.Fatalf("jackett.New: %v", err)
	}

	_, err = c.GetIndexers(context.Background())
	assertRedactedTransportError(t, err, key)

	var uerr *url.Error
	if !errors.As(err, &uerr) {
		t.Fatalf("expected *url.Error, got %T", err)
	}
	if !strings.Contains(uerr.URL, "apikey=REDACTED") {
		t.Errorf("uerr.URL missing apikey=REDACTED: %q", uerr.URL)
	}
	uerrStr := uerr.Error()
	if strings.Contains(uerrStr, key) {
		t.Errorf("uerr.Error() leaked raw key: %q", uerrStr)
	}
	if strings.Contains(uerrStr, url.QueryEscape(key)) {
		t.Errorf("uerr.Error() leaked escaped key: %q", uerrStr)
	}
	if tc.checkTimeout && !uerr.Timeout() {
		t.Errorf("uerr.Timeout() = false, want true")
	}
}

// TestCustomTransportErrorRedactsAPIKey verifies that custom transport errors do not leak keys.
func TestCustomTransportErrorRedactsAPIKey(t *testing.T) {
	t.Parallel()

	const secretKey = `s3cr3t+key/=&x`
	cases := []customTransportTestCase{
		{
			name: "escaped key in inner text",
			tripErr: func(r *http.Request) error {
				return fmt.Errorf("upstream refused %s", r.URL)
			},
		},
		{
			name: "nested url error",
			tripErr: func(r *http.Request) error {
				return fmt.Errorf("retry: %w", &url.Error{Op: "Get", URL: r.URL.String(), Err: errors.New("reset")})
			},
		},
		{
			name: "raw key",
			tripErr: func(r *http.Request) error {
				return fmt.Errorf("raw key %s", r.URL.Query().Get("apikey"))
			},
		},
		{
			name: "timeout error",
			tripErr: func(r *http.Request) error {
				return timeoutError{msg: "timeout: " + r.URL.String()}
			},
			checkTimeout: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runCustomTransportErrorCase(t, tc, secretKey)
		})
	}
}

// TestAPIErrorBodyRedactsAPIKey verifies that APIError.Body scrubs raw and escaped API keys.
func TestAPIErrorBodyRedactsAPIKey(t *testing.T) {
	t.Parallel()

	const secretKey = `s3cr3t+key/=&x`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("bad request " + r.URL.String() + " key=" + r.URL.Query().Get("apikey")))
	}))
	t.Cleanup(srv.Close)

	c, err := jackett.New(srv.URL, secretKey)
	if err != nil {
		t.Fatalf("jackett.New: %v", err)
	}

	_, err = c.GetIndexers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *jackett.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jackett.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Body, "REDACTED") {
		t.Errorf("apiErr.Body missing REDACTED: %q", apiErr.Body)
	}
	if strings.Contains(apiErr.Body, secretKey) {
		t.Errorf("apiErr.Body leaked raw key: %q", apiErr.Body)
	}
	if strings.Contains(apiErr.Body, url.QueryEscape(secretKey)) {
		t.Errorf("apiErr.Body leaked escaped key: %q", apiErr.Body)
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, secretKey) {
		t.Errorf("err.Error() leaked raw key: %q", errMsg)
	}
	if strings.Contains(errMsg, url.QueryEscape(secretKey)) {
		t.Errorf("err.Error() leaked escaped key: %q", errMsg)
	}
}
