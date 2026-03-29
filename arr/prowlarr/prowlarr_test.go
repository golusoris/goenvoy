package prowlarr_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lusoris/goenvoy/arr"
	"github.com/lusoris/goenvoy/arr/prowlarr"
)

func newTestServer(t *testing.T, method, wantPath string, body any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			t.Errorf("method = %s, want %s", r.Method, method)
		}
		if r.Header.Get("X-Api-Key") == "" {
			t.Error("missing X-Api-Key header")
		}
		if wantPath != "" && r.URL.RequestURI() != wantPath {
			t.Errorf("path = %q, want %q", r.URL.RequestURI(), wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if body != nil {
			json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		c, err := prowlarr.New("http://localhost:9696", "test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		t.Parallel()
		_, err := prowlarr.New("://bad", "test-key")
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})
}

func TestGetIndexers(t *testing.T) {
	t.Parallel()

	want := []prowlarr.Indexer{
		{ID: 1, Name: "NZBgeek", Enable: true, Protocol: "usenet"},
		{ID: 2, Name: "TorrentLeech", Enable: true, Protocol: "torrent"},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/indexer", want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetIndexers(context.Background())
	if err != nil {
		t.Fatalf("GetIndexers: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Name != "NZBgeek" {
		t.Errorf("Name = %q, want %q", got[0].Name, "NZBgeek")
	}
}

func TestGetIndexer(t *testing.T) {
	t.Parallel()

	want := prowlarr.Indexer{ID: 1, Name: "NZBgeek", Enable: true}

	srv := newTestServer(t, http.MethodGet, "/api/v1/indexer/1", want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetIndexer(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetIndexer: %v", err)
	}
	if got.Name != "NZBgeek" {
		t.Errorf("Name = %q, want %q", got.Name, "NZBgeek")
	}
}

func TestAddIndexer(t *testing.T) {
	t.Parallel()

	want := prowlarr.Indexer{ID: 3, Name: "Jackett", Enable: true}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var body prowlarr.Indexer
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Name != "Jackett" {
			t.Errorf("Name = %q, want %q", body.Name, "Jackett")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.AddIndexer(context.Background(), &prowlarr.Indexer{Name: "Jackett"})
	if err != nil {
		t.Fatalf("AddIndexer: %v", err)
	}
	if got.ID != 3 {
		t.Errorf("ID = %d, want 3", got.ID)
	}
}

func TestDeleteIndexer(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v1/indexer/1", nil)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteIndexer(context.Background(), 1); err != nil {
		t.Fatalf("DeleteIndexer: %v", err)
	}
}

func TestGetIndexerCategories(t *testing.T) {
	t.Parallel()

	want := []prowlarr.IndexerCategory{
		{ID: 2000, Name: "Movies"},
		{ID: 5000, Name: "TV", SubCategories: []prowlarr.IndexerCategory{
			{ID: 5010, Name: "TV/WEB-DL"},
		}},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/indexer/categories", want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetIndexerCategories(context.Background())
	if err != nil {
		t.Fatalf("GetIndexerCategories: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if len(got[1].SubCategories) != 1 {
		t.Errorf("SubCategories len = %d, want 1", len(got[1].SubCategories))
	}
}

func TestGetApplications(t *testing.T) {
	t.Parallel()

	want := []prowlarr.Application{
		{ID: 1, Name: "Sonarr", SyncLevel: "fullSync"},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/applications", want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetApplications(context.Background())
	if err != nil {
		t.Fatalf("GetApplications: %v", err)
	}
	if got[0].Name != "Sonarr" {
		t.Errorf("Name = %q, want %q", got[0].Name, "Sonarr")
	}
}

func TestDeleteApplication(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v1/applications/1", nil)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteApplication(context.Background(), 1); err != nil {
		t.Fatalf("DeleteApplication: %v", err)
	}
}

func TestGetAppProfiles(t *testing.T) {
	t.Parallel()

	want := []prowlarr.AppProfile{
		{ID: 1, Name: "Standard", EnableRss: true, EnableAutomaticSearch: true, EnableInteractiveSearch: true},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/appprofile", want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetAppProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetAppProfiles: %v", err)
	}
	if got[0].Name != "Standard" {
		t.Errorf("Name = %q, want %q", got[0].Name, "Standard")
	}
}

func TestSearch(t *testing.T) {
	t.Parallel()

	want := []prowlarr.Release{
		{ID: 1, Title: "Ubuntu 24.04 LTS", Size: 4500000000, IndexerID: 1, Indexer: "NZBgeek"},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/search?categories=2000&categories=2010&indexerIds=1&limit=25&query=ubuntu&type=search",
		want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.Search(context.Background(), &prowlarr.SearchOptions{
		Query:      "ubuntu",
		Type:       "search",
		IndexerIDs: []int{1},
		Categories: []int{2000, 2010},
		Limit:      25,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Title != "Ubuntu 24.04 LTS" {
		t.Errorf("Title = %q, want %q", got[0].Title, "Ubuntu 24.04 LTS")
	}
}

func TestGrabRelease(t *testing.T) {
	t.Parallel()

	want := prowlarr.Release{ID: 5, Title: "Grabbed", IndexerID: 1}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GrabRelease(context.Background(), &prowlarr.Release{
		GUID:      "abc-123",
		IndexerID: 1,
	})
	if err != nil {
		t.Fatalf("GrabRelease: %v", err)
	}
	if got.ID != 5 {
		t.Errorf("ID = %d, want 5", got.ID)
	}
}

func TestGetDownloadClients(t *testing.T) {
	t.Parallel()

	want := []prowlarr.DownloadClientResource{
		{ID: 1, Name: "qBittorrent", Enable: true, Protocol: "torrent"},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/downloadclient", want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetDownloadClients(context.Background())
	if err != nil {
		t.Fatalf("GetDownloadClients: %v", err)
	}
	if got[0].Name != "qBittorrent" {
		t.Errorf("Name = %q, want %q", got[0].Name, "qBittorrent")
	}
}

func TestSendCommand(t *testing.T) {
	t.Parallel()

	want := arr.CommandResponse{ID: 1, Name: "ApplicationIndexerSync"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var cmd arr.CommandRequest
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if cmd.Name != "ApplicationIndexerSync" {
			t.Errorf("Name = %q, want ApplicationIndexerSync", cmd.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.SendCommand(context.Background(), arr.CommandRequest{Name: "ApplicationIndexerSync"})
	if err != nil {
		t.Fatalf("SendCommand: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetSystemStatus(t *testing.T) {
	t.Parallel()

	want := arr.StatusResponse{AppName: "Prowlarr", Version: "1.25.0"}

	srv := newTestServer(t, http.MethodGet, "/api/v1/system/status", want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetSystemStatus(context.Background())
	if err != nil {
		t.Fatalf("GetSystemStatus: %v", err)
	}
	if got.AppName != "Prowlarr" {
		t.Errorf("AppName = %q, want %q", got.AppName, "Prowlarr")
	}
}

func TestGetTags(t *testing.T) {
	t.Parallel()

	want := []arr.Tag{{ID: 1, Label: "usenet"}, {ID: 2, Label: "torrent"}}

	srv := newTestServer(t, http.MethodGet, "/api/v1/tag", want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetTags(context.Background())
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}

func TestGetHistory(t *testing.T) {
	t.Parallel()

	want := arr.PagingResource[prowlarr.HistoryRecord]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []prowlarr.HistoryRecord{
			{ID: 1, IndexerID: 1, EventType: "releaseGrabbed", Successful: true},
		},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/history?page=1&pageSize=10", want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetHistory(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if got.Records[0].EventType != "releaseGrabbed" {
		t.Errorf("EventType = %q, want %q", got.Records[0].EventType, "releaseGrabbed")
	}
}

func TestGetHistoryByIndexer(t *testing.T) {
	t.Parallel()

	want := []prowlarr.HistoryRecord{
		{ID: 5, IndexerID: 1, EventType: "indexerQuery"},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/history/indexer?indexerId=1&limit=50",
		want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetHistoryByIndexer(context.Background(), 1, 50)
	if err != nil {
		t.Fatalf("GetHistoryByIndexer: %v", err)
	}
	if got[0].EventType != "indexerQuery" {
		t.Errorf("EventType = %q, want %q", got[0].EventType, "indexerQuery")
	}
}

func TestGetIndexerStats(t *testing.T) {
	t.Parallel()

	want := prowlarr.IndexerStats{
		ID: 0,
		Indexers: []prowlarr.IndexerStatistic{
			{IndexerID: 1, IndexerName: "NZBgeek", NumberOfQueries: 100, NumberOfGrabs: 10},
		},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/indexerstats?startDate=2025-01-01&endDate=2025-01-31",
		want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetIndexerStats(context.Background(), "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetIndexerStats: %v", err)
	}
	if got.Indexers[0].NumberOfQueries != 100 {
		t.Errorf("NumberOfQueries = %d, want 100", got.Indexers[0].NumberOfQueries)
	}
}

func TestGetIndexerStatuses(t *testing.T) {
	t.Parallel()

	want := []prowlarr.IndexerStatus{
		{ID: 1, IndexerID: 5, DisabledTill: "2025-01-01T12:00:00Z"},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/indexerstatus", want)
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetIndexerStatuses(context.Background())
	if err != nil {
		t.Fatalf("GetIndexerStatuses: %v", err)
	}
	if got[0].IndexerID != 5 {
		t.Errorf("IndexerID = %d, want 5", got[0].IndexerID)
	}
}

func TestErrorResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	c, err := prowlarr.New(srv.URL, "bad-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.GetIndexers(context.Background())
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
