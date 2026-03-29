package sonarr_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lusoris/goenvoy/arr"
	"github.com/lusoris/goenvoy/arr/sonarr"
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
		c, err := sonarr.New("http://localhost:8989", "test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		t.Parallel()
		_, err := sonarr.New("://bad", "test-key")
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})
}

func TestGetAllSeries(t *testing.T) {
	t.Parallel()

	want := []sonarr.Series{
		{ID: 1, Title: "Breaking Bad", TvdbID: 81189},
		{ID: 2, Title: "The Wire", TvdbID: 79126},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v3/series", want)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetAllSeries(context.Background())
	if err != nil {
		t.Fatalf("GetAllSeries: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Title != "Breaking Bad" {
		t.Errorf("Title = %q, want %q", got[0].Title, "Breaking Bad")
	}
}

func TestGetSeries(t *testing.T) {
	t.Parallel()

	want := sonarr.Series{ID: 1, Title: "Breaking Bad"}

	srv := newTestServer(t, http.MethodGet, "/api/v3/series/1", want)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetSeries(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetSeries: %v", err)
	}
	if got.Title != "Breaking Bad" {
		t.Errorf("Title = %q, want %q", got.Title, "Breaking Bad")
	}
}

func TestAddSeries(t *testing.T) {
	t.Parallel()

	want := sonarr.Series{ID: 3, Title: "New Show", TvdbID: 99999}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var body sonarr.Series
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Title != "New Show" {
			t.Errorf("Title = %q, want %q", body.Title, "New Show")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.AddSeries(context.Background(), &sonarr.Series{
		Title:  "New Show",
		TvdbID: 99999,
	})
	if err != nil {
		t.Fatalf("AddSeries: %v", err)
	}
	if got.ID != 3 {
		t.Errorf("ID = %d, want 3", got.ID)
	}
}

func TestDeleteSeries(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete,
		"/api/v3/series/1?deleteFiles=true&addImportListExclusion=false",
		nil)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteSeries(context.Background(), 1, true, false); err != nil {
		t.Fatalf("DeleteSeries: %v", err)
	}
}

func TestLookupSeries(t *testing.T) {
	t.Parallel()

	want := []sonarr.Series{{ID: 0, Title: "Breaking Bad", TvdbID: 81189}}

	srv := newTestServer(t, http.MethodGet,
		"/api/v3/series/lookup?term=breaking+bad",
		want)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.LookupSeries(context.Background(), "breaking bad")
	if err != nil {
		t.Fatalf("LookupSeries: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}

func TestGetEpisodes(t *testing.T) {
	t.Parallel()

	want := []sonarr.Episode{
		{ID: 10, SeriesID: 1, SeasonNumber: 1, EpisodeNumber: 1, Title: "Pilot"},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v3/episode?seriesId=1",
		want)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetEpisodes(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetEpisodes: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Title != "Pilot" {
		t.Errorf("Title = %q, want %q", got[0].Title, "Pilot")
	}
}

func TestGetEpisodeFiles(t *testing.T) {
	t.Parallel()

	want := []sonarr.EpisodeFile{
		{ID: 100, SeriesID: 1, RelativePath: "S01E01.mkv", Size: 1073741824},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v3/episodefile?seriesId=1",
		want)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetEpisodeFiles(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetEpisodeFiles: %v", err)
	}
	if got[0].Size != 1073741824 {
		t.Errorf("Size = %d, want 1073741824", got[0].Size)
	}
}

func TestSendCommand(t *testing.T) {
	t.Parallel()

	want := arr.CommandResponse{ID: 42, Name: "RefreshSeries"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var cmd arr.CommandRequest
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if cmd.Name != "RefreshSeries" {
			t.Errorf("Name = %q, want RefreshSeries", cmd.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.SendCommand(context.Background(), arr.CommandRequest{Name: "RefreshSeries"})
	if err != nil {
		t.Fatalf("SendCommand: %v", err)
	}
	if got.ID != 42 {
		t.Errorf("ID = %d, want 42", got.ID)
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	want := sonarr.ParseResult{
		Title: "Breaking.Bad.S01E01.720p",
		ParsedEpisodeInfo: &sonarr.ParsedEpisodeInfo{
			SeriesTitle:    "Breaking Bad",
			SeasonNumber:   1,
			EpisodeNumbers: []int{1},
		},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v3/parse?title=Breaking.Bad.S01E01.720p",
		want)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.Parse(context.Background(), "Breaking.Bad.S01E01.720p")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.ParsedEpisodeInfo.SeriesTitle != "Breaking Bad" {
		t.Errorf("SeriesTitle = %q, want %q", got.ParsedEpisodeInfo.SeriesTitle, "Breaking Bad")
	}
}

func TestGetSystemStatus(t *testing.T) {
	t.Parallel()

	want := arr.StatusResponse{AppName: "Sonarr", Version: "4.0.0"}

	srv := newTestServer(t, http.MethodGet,
		"/api/v3/system/status",
		want)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetSystemStatus(context.Background())
	if err != nil {
		t.Fatalf("GetSystemStatus: %v", err)
	}
	if got.AppName != "Sonarr" {
		t.Errorf("AppName = %q, want %q", got.AppName, "Sonarr")
	}
}

func TestGetQueue(t *testing.T) {
	t.Parallel()

	want := arr.PagingResource[arr.QueueRecord]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []arr.QueueRecord{
			{ID: 1, Title: "Breaking Bad - S01E01"},
		},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v3/queue?page=1&pageSize=10",
		want)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetQueue(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetQueue: %v", err)
	}
	if got.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d, want 1", got.TotalRecords)
	}
}

func TestGetTags(t *testing.T) {
	t.Parallel()

	want := []arr.Tag{{ID: 1, Label: "hd"}, {ID: 2, Label: "anime"}}

	srv := newTestServer(t, http.MethodGet,
		"/api/v3/tag",
		want)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
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

func TestCreateTag(t *testing.T) {
	t.Parallel()

	want := arr.Tag{ID: 3, Label: "new-tag"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var tag arr.Tag
		if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if tag.Label != "new-tag" {
			t.Errorf("Label = %q, want %q", tag.Label, "new-tag")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.CreateTag(context.Background(), "new-tag")
	if err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if got.ID != 3 {
		t.Errorf("ID = %d, want 3", got.ID)
	}
}

func TestErrorResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "bad-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.GetAllSeries(context.Background())
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

func TestGetHistory(t *testing.T) {
	t.Parallel()

	want := arr.PagingResource[sonarr.HistoryRecord]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []sonarr.HistoryRecord{
			{ID: 5, EpisodeID: 10, SeriesID: 1, EventType: "grabbed"},
		},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v3/history?page=1&pageSize=10",
		want)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetHistory(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if got.Records[0].EventType != "grabbed" {
		t.Errorf("EventType = %q, want %q", got.Records[0].EventType, "grabbed")
	}
}

func TestDeleteQueueItem(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete,
		"/api/v3/queue/5?removeFromClient=true&blocklist=false",
		nil)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteQueueItem(context.Background(), 5, true, false); err != nil {
		t.Fatalf("DeleteQueueItem: %v", err)
	}
}
