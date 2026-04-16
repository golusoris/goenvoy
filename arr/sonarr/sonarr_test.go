package sonarr_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golusoris/goenvoy/arr/sonarr"
	"github.com/golusoris/goenvoy/arr/v2"
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

func newRawTestServer(t *testing.T, method, wantPath, body string) *httptest.Server {
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
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, body)
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

func TestUpdateSeries(t *testing.T) {
	t.Parallel()

	want := sonarr.Series{ID: 1, Title: "Updated"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.URL.RequestURI() != "/api/v3/series/1?moveFiles=true" {
			t.Errorf("path = %q", r.URL.RequestURI())
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateSeries(context.Background(), &sonarr.Series{ID: 1, Title: "Updated"}, true)
	if err != nil {
		t.Fatalf("UpdateSeries: %v", err)
	}
	if got.Title != "Updated" {
		t.Errorf("Title = %q", got.Title)
	}
}

func TestGetEpisode(t *testing.T) {
	t.Parallel()

	want := sonarr.Episode{ID: 10, Title: "Pilot"}

	srv := newTestServer(t, http.MethodGet, "/api/v3/episode/10", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetEpisode(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetEpisode: %v", err)
	}
	if got.Title != "Pilot" {
		t.Errorf("Title = %q", got.Title)
	}
}

func TestUpdateEpisode(t *testing.T) {
	t.Parallel()

	want := sonarr.Episode{ID: 10, Title: "Updated", Monitored: true}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateEpisode(context.Background(), &sonarr.Episode{ID: 10, Monitored: true})
	if err != nil {
		t.Fatalf("UpdateEpisode: %v", err)
	}
	if !got.Monitored {
		t.Error("expected Monitored=true")
	}
}

func TestMonitorEpisodes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.MonitorEpisodes(context.Background(), []int{1, 2, 3}, true); err != nil {
		t.Fatalf("MonitorEpisodes: %v", err)
	}
}

func TestGetEpisodeFile(t *testing.T) {
	t.Parallel()

	want := sonarr.EpisodeFile{ID: 100, SeriesID: 1, Size: 1073741824}

	srv := newTestServer(t, http.MethodGet, "/api/v3/episodefile/100", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetEpisodeFile(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetEpisodeFile: %v", err)
	}
	if got.Size != 1073741824 {
		t.Errorf("Size = %d", got.Size)
	}
}

func TestDeleteEpisodeFile(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v3/episodefile/100", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteEpisodeFile(context.Background(), 100); err != nil {
		t.Fatalf("DeleteEpisodeFile: %v", err)
	}
}

func TestDeleteEpisodeFiles(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v3/episodefile/bulk", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteEpisodeFiles(context.Background(), []int{1, 2, 3}); err != nil {
		t.Fatalf("DeleteEpisodeFiles: %v", err)
	}
}

func TestGetCommands(t *testing.T) {
	t.Parallel()

	want := []arr.CommandResponse{{ID: 1, Name: "RefreshSeries"}}

	srv := newTestServer(t, http.MethodGet, "/api/v3/command", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetCommands(context.Background())
	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestGetCommand(t *testing.T) {
	t.Parallel()

	want := arr.CommandResponse{ID: 42, Name: "RefreshSeries"}

	srv := newTestServer(t, http.MethodGet, "/api/v3/command/42", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetCommand(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.Name != "RefreshSeries" {
		t.Errorf("Name = %q", got.Name)
	}
}

func TestGetCalendar(t *testing.T) {
	t.Parallel()

	want := []sonarr.Episode{{ID: 1, Title: "Upcoming"}}

	srv := newTestServer(t, http.MethodGet, "/api/v3/calendar?start=2026-01-01&end=2026-01-31&unmonitored=false", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetCalendar(context.Background(), "2026-01-01", "2026-01-31", false)
	if err != nil {
		t.Fatalf("GetCalendar: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestGetHealth(t *testing.T) {
	t.Parallel()

	want := []arr.HealthCheck{{Type: "warning", Message: "test"}}

	srv := newTestServer(t, http.MethodGet, "/api/v3/health", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("GetHealth: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestGetDiskSpace(t *testing.T) {
	t.Parallel()

	want := []arr.DiskSpace{{Path: "/data", FreeSpace: 1000}}

	srv := newTestServer(t, http.MethodGet, "/api/v3/diskspace", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetDiskSpace(context.Background())
	if err != nil {
		t.Fatalf("GetDiskSpace: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestGetQualityProfiles(t *testing.T) {
	t.Parallel()

	want := []arr.QualityProfile{{ID: 1, Name: "Any"}}

	srv := newTestServer(t, http.MethodGet, "/api/v3/qualityprofile", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetQualityProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetQualityProfiles: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestGetRootFolders(t *testing.T) {
	t.Parallel()

	want := []arr.RootFolder{{ID: 1, Path: "/tv"}}

	srv := newTestServer(t, http.MethodGet, "/api/v3/rootfolder", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetRootFolders(context.Background())
	if err != nil {
		t.Fatalf("GetRootFolders: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestUpdateSeasonPass(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.UpdateSeasonPass(context.Background(), sonarr.SeasonPassResource{}); err != nil {
		t.Fatalf("UpdateSeasonPass: %v", err)
	}
}

func TestDeleteCommand(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v3/command/1", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteCommand(context.Background(), 1); err != nil {
		t.Fatalf("DeleteCommand: %v", err)
	}
}

func TestUpdateEpisodeFile(t *testing.T) {
	t.Parallel()

	want := sonarr.EpisodeFile{ID: 1}

	srv := newTestServer(t, http.MethodPut, "/api/v3/episodefile/1", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateEpisodeFile(context.Background(), &sonarr.EpisodeFile{ID: 1})
	if err != nil {
		t.Fatalf("UpdateEpisodeFile: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestEditEpisodeFiles(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodPut, "/api/v3/episodefile/editor", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.EditEpisodeFiles(context.Background(), &sonarr.EpisodeFileEditorResource{
		EpisodeFileIDs: []int{1, 2},
	}); err != nil {
		t.Fatalf("EditEpisodeFiles: %v", err)
	}
}

func TestUpdateCustomFormatsBulk(t *testing.T) {
	t.Parallel()

	want := []arr.CustomFormatResource{{ID: 1, Name: "test"}}

	srv := newTestServer(t, http.MethodPut, "/api/v3/customformat/bulk", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateCustomFormatsBulk(context.Background(), &arr.CustomFormatBulkResource{IDs: []int{1}})
	if err != nil {
		t.Fatalf("UpdateCustomFormatsBulk: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestDeleteCustomFormatsBulk(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v3/customformat/bulk", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteCustomFormatsBulk(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("DeleteCustomFormatsBulk: %v", err)
	}
}

func TestUpdateDownloadClientsBulk(t *testing.T) {
	t.Parallel()

	want := []arr.ProviderResource{{ID: 1}}

	srv := newTestServer(t, http.MethodPut, "/api/v3/downloadclient/bulk", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateDownloadClientsBulk(context.Background(), &arr.ProviderBulkResource{IDs: []int{1}})
	if err != nil {
		t.Fatalf("UpdateDownloadClientsBulk: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestDeleteDownloadClientsBulk(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v3/downloadclient/bulk", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteDownloadClientsBulk(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("DeleteDownloadClientsBulk: %v", err)
	}
}

func TestTestAllDownloadClients(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodPost, "/api/v3/downloadclient/testall", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.TestAllDownloadClients(context.Background()); err != nil {
		t.Fatalf("TestAllDownloadClients: %v", err)
	}
}

func TestUpdateIndexersBulk(t *testing.T) {
	t.Parallel()

	want := []arr.ProviderResource{{ID: 1}}

	srv := newTestServer(t, http.MethodPut, "/api/v3/indexer/bulk", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateIndexersBulk(context.Background(), &arr.ProviderBulkResource{IDs: []int{1}})
	if err != nil {
		t.Fatalf("UpdateIndexersBulk: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestDeleteIndexersBulk(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v3/indexer/bulk", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteIndexersBulk(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("DeleteIndexersBulk: %v", err)
	}
}

func TestTestAllIndexers(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodPost, "/api/v3/indexer/testall", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.TestAllIndexers(context.Background()); err != nil {
		t.Fatalf("TestAllIndexers: %v", err)
	}
}

func TestUpdateImportListsBulk(t *testing.T) {
	t.Parallel()

	want := []arr.ProviderResource{{ID: 1}}

	srv := newTestServer(t, http.MethodPut, "/api/v3/importlist/bulk", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateImportListsBulk(context.Background(), &arr.ProviderBulkResource{IDs: []int{1}})
	if err != nil {
		t.Fatalf("UpdateImportListsBulk: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestDeleteImportListsBulk(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v3/importlist/bulk", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteImportListsBulk(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("DeleteImportListsBulk: %v", err)
	}
}

func TestTestAllImportLists(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodPost, "/api/v3/importlist/testall", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.TestAllImportLists(context.Background()); err != nil {
		t.Fatalf("TestAllImportLists: %v", err)
	}
}

func TestGetImportListConfig(t *testing.T) {
	t.Parallel()

	want := sonarr.ImportListConfigResource{ID: 1, ListSyncLevel: "disabled"}

	srv := newTestServer(t, http.MethodGet, "/api/v3/config/importlist", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetImportListConfig(context.Background())
	if err != nil {
		t.Fatalf("GetImportListConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateImportListConfig(t *testing.T) {
	t.Parallel()

	want := sonarr.ImportListConfigResource{ID: 1, ListSyncLevel: "logOnly"}

	srv := newTestServer(t, http.MethodPut, "/api/v3/config/importlist/1", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateImportListConfig(context.Background(), &sonarr.ImportListConfigResource{ID: 1, ListSyncLevel: "logOnly"})
	if err != nil {
		t.Fatalf("UpdateImportListConfig: %v", err)
	}
	if got.ListSyncLevel != "logOnly" {
		t.Errorf("ListSyncLevel = %q, want logOnly", got.ListSyncLevel)
	}
}

func TestGetImportListExclusionsPaged(t *testing.T) {
	t.Parallel()

	want := arr.PagingResource[arr.ImportListExclusionResource]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records:      []arr.ImportListExclusionResource{{ID: 1, TvdbID: 123}},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v3/importlistexclusion/paged?page=1&pageSize=10", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetImportListExclusionsPaged(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetImportListExclusionsPaged: %v", err)
	}
	if got.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d, want 1", got.TotalRecords)
	}
}

func TestDeleteImportListExclusionsBulk(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v3/importlistexclusion/bulk", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteImportListExclusionsBulk(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("DeleteImportListExclusionsBulk: %v", err)
	}
}

func TestTestAllNotifications(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodPost, "/api/v3/notification/testall", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.TestAllNotifications(context.Background()); err != nil {
		t.Fatalf("TestAllNotifications: %v", err)
	}
}

func TestTestAllMetadataConsumers(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodPost, "/api/v3/metadata/testall", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.TestAllMetadataConsumers(context.Background()); err != nil {
		t.Fatalf("TestAllMetadataConsumers: %v", err)
	}
}

func TestGetLanguage(t *testing.T) {
	t.Parallel()

	want := sonarr.Language{ID: 1, Name: "English"}

	srv := newTestServer(t, http.MethodGet, "/api/v3/language/1", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLanguage(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetLanguage: %v", err)
	}
	if got.Name != "English" {
		t.Errorf("Name = %q, want English", got.Name)
	}
}

func TestGetLocalization(t *testing.T) {
	t.Parallel()

	want := sonarr.LocalizationResource{ID: 1, Strings: map[string]string{"key": "val"}}

	srv := newTestServer(t, http.MethodGet, "/api/v3/localization", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLocalization(context.Background())
	if err != nil {
		t.Fatalf("GetLocalization: %v", err)
	}
	if got.Strings["key"] != "val" {
		t.Errorf("Strings[key] = %q, want val", got.Strings["key"])
	}
}

func TestUpdateQualityDefinitions(t *testing.T) {
	t.Parallel()

	want := []arr.QualityDefinitionResource{{ID: 1, Title: "HDTV-720p"}}

	srv := newTestServer(t, http.MethodPut, "/api/v3/qualitydefinition/update", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateQualityDefinitions(context.Background(), []arr.QualityDefinitionResource{{ID: 1, Title: "HDTV-720p"}})
	if err != nil {
		t.Fatalf("UpdateQualityDefinitions: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

func TestGetQualityProfileSchema(t *testing.T) {
	t.Parallel()

	want := arr.QualityProfile{ID: 1, Name: "schema"}

	srv := newTestServer(t, http.MethodGet, "/api/v3/qualityprofile/schema", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetQualityProfileSchema(context.Background())
	if err != nil {
		t.Fatalf("GetQualityProfileSchema: %v", err)
	}
	if got.Name != "schema" {
		t.Errorf("Name = %q, want schema", got.Name)
	}
}

func TestUpdateRootFolder(t *testing.T) {
	t.Parallel()

	want := arr.RootFolder{ID: 1, Path: "/tv"}

	srv := newTestServer(t, http.MethodPut, "/api/v3/rootfolder/1", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateRootFolder(context.Background(), &arr.RootFolder{ID: 1, Path: "/tv"})
	if err != nil {
		t.Fatalf("UpdateRootFolder: %v", err)
	}
	if got.Path != "/tv" {
		t.Errorf("Path = %q, want /tv", got.Path)
	}
}

func TestBrowseFileSystem(t *testing.T) {
	t.Parallel()

	want := sonarr.FileSystemResource{
		Directories: []sonarr.FileSystemEntry{{Path: "/tv", Name: "tv"}},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v3/filesystem?path=%2Ftv&includeFiles=true", want)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.BrowseFileSystem(context.Background(), "/tv", true)
	if err != nil {
		t.Fatalf("BrowseFileSystem: %v", err)
	}
	if len(got.Directories) != 1 {
		t.Errorf("len(Directories) = %d", len(got.Directories))
	}
}

func TestPing(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodGet, "/ping", nil)
	defer srv.Close()

	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestGetCalendarByID(t *testing.T) {
	t.Parallel()
	want := sonarr.Episode{ID: 42, Title: "Pilot"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/calendar/42", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetCalendarByID(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetCalendarByID: %v", err)
	}
	if got.ID != 42 {
		t.Errorf("ID = %d, want 42", got.ID)
	}
}

func TestGetWantedCutoffByID(t *testing.T) {
	t.Parallel()
	want := sonarr.Episode{ID: 7, Title: "Cutoff"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/wanted/cutoff/7", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetWantedCutoffByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetWantedCutoffByID: %v", err)
	}
	if got.ID != 7 {
		t.Errorf("ID = %d, want 7", got.ID)
	}
}

func TestGetWantedMissingByID(t *testing.T) {
	t.Parallel()
	want := sonarr.Episode{ID: 3, Title: "Missing"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/wanted/missing/3", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetWantedMissingByID(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetWantedMissingByID: %v", err)
	}
	if got.ID != 3 {
		t.Errorf("ID = %d, want 3", got.ID)
	}
}

func TestGetDownloadClientConfigByID(t *testing.T) {
	t.Parallel()
	want := arr.DownloadClientConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/downloadclient/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetDownloadClientConfigByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDownloadClientConfigByID: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetHostConfigByID(t *testing.T) {
	t.Parallel()
	want := arr.HostConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/host/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetHostConfigByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetHostConfigByID: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetImportListConfigByID(t *testing.T) {
	t.Parallel()
	want := sonarr.ImportListConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/importlist/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetImportListConfigByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetImportListConfigByID: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetIndexerConfigByID(t *testing.T) {
	t.Parallel()
	want := arr.IndexerConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/indexer/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetIndexerConfigByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetIndexerConfigByID: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetMediaManagementConfigByID(t *testing.T) {
	t.Parallel()
	want := arr.MediaManagementConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/mediamanagement/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetMediaManagementConfigByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetMediaManagementConfigByID: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetNamingConfigByID(t *testing.T) {
	t.Parallel()
	want := arr.NamingConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/naming/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetNamingConfigByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetNamingConfigByID: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetUIConfigByID(t *testing.T) {
	t.Parallel()
	want := arr.UIConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/ui/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetUIConfigByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUIConfigByID: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestDownloadClientAction(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/downloadclient/action/testAction", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DownloadClientAction(context.Background(), "testAction", nil); err != nil {
		t.Fatalf("DownloadClientAction: %v", err)
	}
}

func TestImportListAction(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/importlist/action/testAction", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.ImportListAction(context.Background(), "testAction", nil); err != nil {
		t.Fatalf("ImportListAction: %v", err)
	}
}

func TestIndexerAction(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/indexer/action/testAction", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.IndexerAction(context.Background(), "testAction", nil); err != nil {
		t.Fatalf("IndexerAction: %v", err)
	}
}

func TestMetadataAction(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/metadata/action/testAction", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.MetadataAction(context.Background(), "testAction", nil); err != nil {
		t.Fatalf("MetadataAction: %v", err)
	}
}

func TestNotificationAction(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/notification/action/testAction", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.NotificationAction(context.Background(), "testAction", nil); err != nil {
		t.Fatalf("NotificationAction: %v", err)
	}
}

func TestGetLanguageProfiles(t *testing.T) {
	t.Parallel()
	want := []sonarr.LanguageProfileResource{{ID: 1, Name: "English"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/languageprofile", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLanguageProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetLanguageProfiles: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}

func TestGetLanguageProfile(t *testing.T) {
	t.Parallel()
	want := sonarr.LanguageProfileResource{ID: 1, Name: "English"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/languageprofile/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLanguageProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetLanguageProfile: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateLanguageProfile(t *testing.T) {
	t.Parallel()
	want := sonarr.LanguageProfileResource{ID: 1, Name: "English"}
	srv := newTestServer(t, http.MethodPost, "/api/v3/languageprofile", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateLanguageProfile(context.Background(), &sonarr.LanguageProfileResource{Name: "English"})
	if err != nil {
		t.Fatalf("CreateLanguageProfile: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateLanguageProfile(t *testing.T) {
	t.Parallel()
	want := sonarr.LanguageProfileResource{ID: 1, Name: "Updated"}
	srv := newTestServer(t, http.MethodPut, "/api/v3/languageprofile/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateLanguageProfile(context.Background(), &sonarr.LanguageProfileResource{ID: 1, Name: "Updated"})
	if err != nil {
		t.Fatalf("UpdateLanguageProfile: %v", err)
	}
	if got.Name != "Updated" {
		t.Errorf("Name = %q, want Updated", got.Name)
	}
}

func TestDeleteLanguageProfile(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/languageprofile/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteLanguageProfile(context.Background(), 1); err != nil {
		t.Fatalf("DeleteLanguageProfile: %v", err)
	}
}

func TestGetLanguageProfileSchema(t *testing.T) {
	t.Parallel()
	want := sonarr.LanguageProfileResource{ID: 0, Name: "Schema"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/languageprofile/schema", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLanguageProfileSchema(context.Background())
	if err != nil {
		t.Fatalf("GetLanguageProfileSchema: %v", err)
	}
	if got.Name != "Schema" {
		t.Errorf("Name = %q, want Schema", got.Name)
	}
}

func TestGetLocalizationByID(t *testing.T) {
	t.Parallel()
	want := sonarr.LocalizationResource{ID: 1, Strings: map[string]string{"hello": "world"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/localization/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLocalizationByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetLocalizationByID: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetLocalizationLanguages(t *testing.T) {
	t.Parallel()
	want := []sonarr.LocalizationLanguageResource{{Identifier: "en"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/localization/language", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLocalizationLanguages(context.Background())
	if err != nil {
		t.Fatalf("GetLocalizationLanguages: %v", err)
	}
	if len(got) != 1 || got[0].Identifier != "en" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestGetNamingExamples(t *testing.T) {
	t.Parallel()
	want := arr.NamingConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/naming/examples", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetNamingExamples(context.Background())
	if err != nil {
		t.Fatalf("GetNamingExamples: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetQualityDefinitionLimits(t *testing.T) {
	t.Parallel()
	want := sonarr.QualityDefinitionLimitsResource{Min: 1, Max: 400}
	srv := newTestServer(t, http.MethodGet, "/api/v3/qualitydefinition/limits", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetQualityDefinitionLimits(context.Background())
	if err != nil {
		t.Fatalf("GetQualityDefinitionLimits: %v", err)
	}
	if got.Min != 1 || got.Max != 400 {
		t.Errorf("got Min=%d Max=%d, want 1/400", got.Min, got.Max)
	}
}

func TestUpdateEpisodeFilesBulk(t *testing.T) {
	t.Parallel()
	want := []sonarr.EpisodeFile{{ID: 1}, {ID: 2}}
	srv := newTestServer(t, http.MethodPut, "/api/v3/episodefile/bulk", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateEpisodeFilesBulk(context.Background(), &sonarr.EpisodeFileEditorResource{
		EpisodeFileIDs: []int{1, 2},
	})
	if err != nil {
		t.Fatalf("UpdateEpisodeFilesBulk: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestGetUpdateLogFileContent(t *testing.T) {
	t.Parallel()
	srv := newRawTestServer(t, http.MethodGet, "/api/v3/log/file/update/update.txt", "log content")
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetUpdateLogFileContent(context.Background(), "update.txt")
	if err != nil {
		t.Fatalf("GetUpdateLogFileContent: %v", err)
	}
	if got != "log content" {
		t.Errorf("content = %q, want %q", got, "log content")
	}
}

func TestGetSystemRoutesDuplicate(t *testing.T) {
	t.Parallel()
	want := []arr.SystemRouteResource{{Path: "/api/v3/test", Method: "GET"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/system/routes/duplicate", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetSystemRoutesDuplicate(context.Background())
	if err != nil {
		t.Fatalf("GetSystemRoutesDuplicate: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetLogFileContent(t *testing.T) {
	t.Parallel()
	srv := newRawTestServer(t, http.MethodGet, "/api/v3/log/file/sonarr.txt", "log line 1\nlog line 2")
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLogFileContent(context.Background(), "sonarr.txt")
	if err != nil {
		t.Fatalf("GetLogFileContent: %v", err)
	}
	if got != "log line 1\nlog line 2" {
		t.Errorf("content = %q, want %q", got, "log line 1\nlog line 2")
	}
}

func TestHeadPing(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodHead, "/ping", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.HeadPing(context.Background()); err != nil {
		t.Fatalf("HeadPing: %v", err)
	}
}

func TestUploadBackup(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("X-Api-Key") == "" {
			t.Error("missing X-Api-Key header")
		}
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "multipart/form-data") {
			t.Errorf("Content-Type = %q, want multipart/form-data", ct)
		}
		f, fh, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("FormFile: %v", err)
		}
		defer f.Close()
		if fh.Filename != "backup.zip" {
			t.Errorf("filename = %q, want %q", fh.Filename, "backup.zip")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.UploadBackup(context.Background(), "backup.zip", strings.NewReader("fake-backup-data")); err != nil {
		t.Fatalf("UploadBackup: %v", err)
	}
}

// ---------- Config (CRUD) ----------

func TestGetDownloadClientConfig(t *testing.T) {
	t.Parallel()
	want := arr.DownloadClientConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/downloadclient", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetDownloadClientConfig(context.Background())
	if err != nil {
		t.Fatalf("GetDownloadClientConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateDownloadClientConfig(t *testing.T) {
	t.Parallel()
	want := arr.DownloadClientConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/config/downloadclient/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateDownloadClientConfig(context.Background(), &arr.DownloadClientConfigResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateDownloadClientConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetIndexerConfig(t *testing.T) {
	t.Parallel()
	want := arr.IndexerConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/indexer", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetIndexerConfig(context.Background())
	if err != nil {
		t.Fatalf("GetIndexerConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateIndexerConfig(t *testing.T) {
	t.Parallel()
	want := arr.IndexerConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/config/indexer/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateIndexerConfig(context.Background(), &arr.IndexerConfigResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateIndexerConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetNamingConfig(t *testing.T) {
	t.Parallel()
	want := arr.NamingConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/naming", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetNamingConfig(context.Background())
	if err != nil {
		t.Fatalf("GetNamingConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateNamingConfig(t *testing.T) {
	t.Parallel()
	want := arr.NamingConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/config/naming/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateNamingConfig(context.Background(), &arr.NamingConfigResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateNamingConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetHostConfig(t *testing.T) {
	t.Parallel()
	want := arr.HostConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/host", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetHostConfig(context.Background())
	if err != nil {
		t.Fatalf("GetHostConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateHostConfig(t *testing.T) {
	t.Parallel()
	want := arr.HostConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/config/host/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateHostConfig(context.Background(), &arr.HostConfigResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateHostConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetUIConfig(t *testing.T) {
	t.Parallel()
	want := arr.UIConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/ui", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetUIConfig(context.Background())
	if err != nil {
		t.Fatalf("GetUIConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateUIConfig(t *testing.T) {
	t.Parallel()
	want := arr.UIConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/config/ui/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateUIConfig(context.Background(), &arr.UIConfigResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateUIConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetMediaManagementConfig(t *testing.T) {
	t.Parallel()
	want := arr.MediaManagementConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/config/mediamanagement", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetMediaManagementConfig(context.Background())
	if err != nil {
		t.Fatalf("GetMediaManagementConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateMediaManagementConfig(t *testing.T) {
	t.Parallel()
	want := arr.MediaManagementConfigResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/config/mediamanagement/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateMediaManagementConfig(context.Background(), &arr.MediaManagementConfigResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateMediaManagementConfig: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

// ---------- AutoTag ----------

func TestGetAutoTagging(t *testing.T) {
	t.Parallel()
	want := []arr.AutoTaggingResource{{ID: 1, Name: "test"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/autotagging", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetAutoTagging(context.Background())
	if err != nil {
		t.Fatalf("GetAutoTagging: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetAutoTag(t *testing.T) {
	t.Parallel()
	want := arr.AutoTaggingResource{ID: 1, Name: "test"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/autotagging/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetAutoTag(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetAutoTag: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateAutoTag(t *testing.T) {
	t.Parallel()
	want := arr.AutoTaggingResource{ID: 1, Name: "test"}
	srv := newTestServer(t, http.MethodPost, "/api/v3/autotagging", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateAutoTag(context.Background(), &arr.AutoTaggingResource{Name: "test"})
	if err != nil {
		t.Fatalf("CreateAutoTag: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateAutoTag(t *testing.T) {
	t.Parallel()
	want := arr.AutoTaggingResource{ID: 1, Name: "updated"}
	srv := newTestServer(t, http.MethodPut, "/api/v3/autotagging/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateAutoTag(context.Background(), &arr.AutoTaggingResource{ID: 1, Name: "updated"})
	if err != nil {
		t.Fatalf("UpdateAutoTag: %v", err)
	}
	if got.Name != "updated" {
		t.Errorf("Name = %q, want updated", got.Name)
	}
}

func TestDeleteAutoTag(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/autotagging/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteAutoTag(context.Background(), 1); err != nil {
		t.Fatalf("DeleteAutoTag: %v", err)
	}
}

func TestGetAutoTagSchema(t *testing.T) {
	t.Parallel()
	want := []arr.AutoTaggingSpecification{{Name: "spec"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/autotagging/schema", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetAutoTagSchema(context.Background())
	if err != nil {
		t.Fatalf("GetAutoTagSchema: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

// ---------- Blocklist ----------

func TestGetBlocklist(t *testing.T) {
	t.Parallel()
	want := arr.PagingResource[arr.BlocklistResource]{
		Page: 1, PageSize: 10, TotalRecords: 1,
		Records: []arr.BlocklistResource{{ID: 1}},
	}
	srv := newTestServer(t, http.MethodGet, "/api/v3/blocklist?page=1&pageSize=10", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetBlocklist(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetBlocklist: %v", err)
	}
	if got.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d, want 1", got.TotalRecords)
	}
}

func TestDeleteBlocklistItem(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/blocklist/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteBlocklistItem(context.Background(), 1); err != nil {
		t.Fatalf("DeleteBlocklistItem: %v", err)
	}
}

func TestDeleteBlocklistBulk(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/blocklist/bulk", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteBlocklistBulk(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("DeleteBlocklistBulk: %v", err)
	}
}

// ---------- CustomFilter ----------

func TestGetCustomFilters(t *testing.T) {
	t.Parallel()
	want := []arr.CustomFilterResource{{ID: 1, Label: "test"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/customfilter", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetCustomFilters(context.Background())
	if err != nil {
		t.Fatalf("GetCustomFilters: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetCustomFilter(t *testing.T) {
	t.Parallel()
	want := arr.CustomFilterResource{ID: 1, Label: "test"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/customfilter/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetCustomFilter(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetCustomFilter: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateCustomFilter(t *testing.T) {
	t.Parallel()
	want := arr.CustomFilterResource{ID: 1, Label: "new"}
	srv := newTestServer(t, http.MethodPost, "/api/v3/customfilter", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateCustomFilter(context.Background(), &arr.CustomFilterResource{Label: "new"})
	if err != nil {
		t.Fatalf("CreateCustomFilter: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateCustomFilter(t *testing.T) {
	t.Parallel()
	want := arr.CustomFilterResource{ID: 1, Label: "updated"}
	srv := newTestServer(t, http.MethodPut, "/api/v3/customfilter/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateCustomFilter(context.Background(), &arr.CustomFilterResource{ID: 1, Label: "updated"})
	if err != nil {
		t.Fatalf("UpdateCustomFilter: %v", err)
	}
	if got.Label != "updated" {
		t.Errorf("Label = %q, want updated", got.Label)
	}
}

func TestDeleteCustomFilter(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/customfilter/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteCustomFilter(context.Background(), 1); err != nil {
		t.Fatalf("DeleteCustomFilter: %v", err)
	}
}

// ---------- CustomFormat ----------

func TestGetCustomFormats(t *testing.T) {
	t.Parallel()
	want := []arr.CustomFormatResource{{ID: 1, Name: "test"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/customformat", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetCustomFormats(context.Background())
	if err != nil {
		t.Fatalf("GetCustomFormats: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetCustomFormat(t *testing.T) {
	t.Parallel()
	want := arr.CustomFormatResource{ID: 1, Name: "test"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/customformat/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetCustomFormat(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetCustomFormat: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateCustomFormat(t *testing.T) {
	t.Parallel()
	want := arr.CustomFormatResource{ID: 1, Name: "new"}
	srv := newTestServer(t, http.MethodPost, "/api/v3/customformat", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateCustomFormat(context.Background(), &arr.CustomFormatResource{Name: "new"})
	if err != nil {
		t.Fatalf("CreateCustomFormat: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateCustomFormat(t *testing.T) {
	t.Parallel()
	want := arr.CustomFormatResource{ID: 1, Name: "updated"}
	srv := newTestServer(t, http.MethodPut, "/api/v3/customformat/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateCustomFormat(context.Background(), &arr.CustomFormatResource{ID: 1, Name: "updated"})
	if err != nil {
		t.Fatalf("UpdateCustomFormat: %v", err)
	}
	if got.Name != "updated" {
		t.Errorf("Name = %q, want updated", got.Name)
	}
}

func TestDeleteCustomFormat(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/customformat/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteCustomFormat(context.Background(), 1); err != nil {
		t.Fatalf("DeleteCustomFormat: %v", err)
	}
}

func TestGetCustomFormatSchema(t *testing.T) {
	t.Parallel()
	want := []arr.CustomFormatSpecification{{Name: "spec"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/customformat/schema", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetCustomFormatSchema(context.Background())
	if err != nil {
		t.Fatalf("GetCustomFormatSchema: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

// ---------- DelayProfile ----------

func TestGetDelayProfiles(t *testing.T) {
	t.Parallel()
	want := []arr.DelayProfileResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/delayprofile", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetDelayProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetDelayProfiles: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetDelayProfile(t *testing.T) {
	t.Parallel()
	want := arr.DelayProfileResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/delayprofile/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetDelayProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDelayProfile: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateDelayProfile(t *testing.T) {
	t.Parallel()
	want := arr.DelayProfileResource{ID: 1}
	srv := newTestServer(t, http.MethodPost, "/api/v3/delayprofile", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateDelayProfile(context.Background(), &arr.DelayProfileResource{})
	if err != nil {
		t.Fatalf("CreateDelayProfile: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateDelayProfile(t *testing.T) {
	t.Parallel()
	want := arr.DelayProfileResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/delayprofile/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateDelayProfile(context.Background(), &arr.DelayProfileResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateDelayProfile: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestDeleteDelayProfile(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/delayprofile/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteDelayProfile(context.Background(), 1); err != nil {
		t.Fatalf("DeleteDelayProfile: %v", err)
	}
}

func TestReorderDelayProfile(t *testing.T) {
	t.Parallel()
	want := []arr.DelayProfileResource{{ID: 1}, {ID: 2}}
	srv := newTestServer(t, http.MethodPut, "/api/v3/delayprofile/reorder/1?after=2", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.ReorderDelayProfile(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("ReorderDelayProfile: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

// ---------- DownloadClient ----------

func TestGetDownloadClients(t *testing.T) {
	t.Parallel()
	want := []arr.ProviderResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/downloadclient", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetDownloadClients(context.Background())
	if err != nil {
		t.Fatalf("GetDownloadClients: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetDownloadClient(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/downloadclient/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetDownloadClient(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDownloadClient: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateDownloadClient(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodPost, "/api/v3/downloadclient", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateDownloadClient(context.Background(), &arr.ProviderResource{})
	if err != nil {
		t.Fatalf("CreateDownloadClient: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateDownloadClient(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/downloadclient/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateDownloadClient(context.Background(), &arr.ProviderResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateDownloadClient: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestDeleteDownloadClient(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/downloadclient/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteDownloadClient(context.Background(), 1); err != nil {
		t.Fatalf("DeleteDownloadClient: %v", err)
	}
}

func TestGetDownloadClientSchema(t *testing.T) {
	t.Parallel()
	want := []arr.ProviderResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/downloadclient/schema", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetDownloadClientSchema(context.Background())
	if err != nil {
		t.Fatalf("GetDownloadClientSchema: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestTestDownloadClient(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/downloadclient/test", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.TestDownloadClient(context.Background(), &arr.ProviderResource{ID: 1}); err != nil {
		t.Fatalf("TestDownloadClient: %v", err)
	}
}

// ---------- ImportList ----------

func TestGetImportLists(t *testing.T) {
	t.Parallel()
	want := []arr.ProviderResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/importlist", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetImportLists(context.Background())
	if err != nil {
		t.Fatalf("GetImportLists: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetImportList(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/importlist/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetImportList(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetImportList: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateImportList(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodPost, "/api/v3/importlist", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateImportList(context.Background(), &arr.ProviderResource{})
	if err != nil {
		t.Fatalf("CreateImportList: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateImportList(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/importlist/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateImportList(context.Background(), &arr.ProviderResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateImportList: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestDeleteImportList(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/importlist/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteImportList(context.Background(), 1); err != nil {
		t.Fatalf("DeleteImportList: %v", err)
	}
}

func TestGetImportListSchema(t *testing.T) {
	t.Parallel()
	want := []arr.ProviderResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/importlist/schema", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetImportListSchema(context.Background())
	if err != nil {
		t.Fatalf("GetImportListSchema: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestTestImportList(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/importlist/test", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.TestImportList(context.Background(), &arr.ProviderResource{ID: 1}); err != nil {
		t.Fatalf("TestImportList: %v", err)
	}
}

func TestGetImportListExclusions(t *testing.T) {
	t.Parallel()
	want := []arr.ImportListExclusionResource{{ID: 1, TvdbID: 123}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/importlistexclusion", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetImportListExclusions(context.Background())
	if err != nil {
		t.Fatalf("GetImportListExclusions: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetImportListExclusion(t *testing.T) {
	t.Parallel()
	want := arr.ImportListExclusionResource{ID: 1, TvdbID: 123}
	srv := newTestServer(t, http.MethodGet, "/api/v3/importlistexclusion/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetImportListExclusion(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetImportListExclusion: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateImportListExclusion(t *testing.T) {
	t.Parallel()
	want := arr.ImportListExclusionResource{ID: 1, TvdbID: 123}
	srv := newTestServer(t, http.MethodPost, "/api/v3/importlistexclusion", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateImportListExclusion(context.Background(), &arr.ImportListExclusionResource{TvdbID: 123})
	if err != nil {
		t.Fatalf("CreateImportListExclusion: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateImportListExclusion(t *testing.T) {
	t.Parallel()
	want := arr.ImportListExclusionResource{ID: 1, TvdbID: 456}
	srv := newTestServer(t, http.MethodPut, "/api/v3/importlistexclusion/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateImportListExclusion(context.Background(), &arr.ImportListExclusionResource{ID: 1, TvdbID: 456})
	if err != nil {
		t.Fatalf("UpdateImportListExclusion: %v", err)
	}
	if got.TvdbID != 456 {
		t.Errorf("TvdbID = %d, want 456", got.TvdbID)
	}
}

func TestDeleteImportListExclusion(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/importlistexclusion/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteImportListExclusion(context.Background(), 1); err != nil {
		t.Fatalf("DeleteImportListExclusion: %v", err)
	}
}

// ---------- Indexer ----------

func TestGetIndexers(t *testing.T) {
	t.Parallel()
	want := []arr.ProviderResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/indexer", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetIndexers(context.Background())
	if err != nil {
		t.Fatalf("GetIndexers: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetIndexer(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/indexer/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetIndexer(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetIndexer: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateIndexer(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodPost, "/api/v3/indexer", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateIndexer(context.Background(), &arr.ProviderResource{})
	if err != nil {
		t.Fatalf("CreateIndexer: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateIndexer(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/indexer/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateIndexer(context.Background(), &arr.ProviderResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateIndexer: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestDeleteIndexer(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/indexer/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteIndexer(context.Background(), 1); err != nil {
		t.Fatalf("DeleteIndexer: %v", err)
	}
}

func TestGetIndexerSchema(t *testing.T) {
	t.Parallel()
	want := []arr.ProviderResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/indexer/schema", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetIndexerSchema(context.Background())
	if err != nil {
		t.Fatalf("GetIndexerSchema: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestTestIndexer(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/indexer/test", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.TestIndexer(context.Background(), &arr.ProviderResource{ID: 1}); err != nil {
		t.Fatalf("TestIndexer: %v", err)
	}
}

func TestGetIndexerFlags(t *testing.T) {
	t.Parallel()
	want := []arr.IndexerFlagResource{{ID: 1, Name: "freeleech"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/indexerflag", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetIndexerFlags(context.Background())
	if err != nil {
		t.Fatalf("GetIndexerFlags: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

// ---------- Metadata ----------

func TestGetMetadataConsumers(t *testing.T) {
	t.Parallel()
	want := []arr.ProviderResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/metadata", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetMetadataConsumers(context.Background())
	if err != nil {
		t.Fatalf("GetMetadataConsumers: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetMetadataConsumer(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/metadata/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetMetadataConsumer(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetMetadataConsumer: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateMetadataConsumer(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodPost, "/api/v3/metadata", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateMetadataConsumer(context.Background(), &arr.ProviderResource{})
	if err != nil {
		t.Fatalf("CreateMetadataConsumer: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateMetadataConsumer(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/metadata/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateMetadataConsumer(context.Background(), &arr.ProviderResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateMetadataConsumer: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestDeleteMetadataConsumer(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/metadata/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteMetadataConsumer(context.Background(), 1); err != nil {
		t.Fatalf("DeleteMetadataConsumer: %v", err)
	}
}

func TestGetMetadataSchema(t *testing.T) {
	t.Parallel()
	want := []arr.ProviderResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/metadata/schema", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetMetadataSchema(context.Background())
	if err != nil {
		t.Fatalf("GetMetadataSchema: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestTestMetadataConsumer(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/metadata/test", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.TestMetadataConsumer(context.Background(), &arr.ProviderResource{ID: 1}); err != nil {
		t.Fatalf("TestMetadataConsumer: %v", err)
	}
}

// ---------- Notification ----------

func TestGetNotifications(t *testing.T) {
	t.Parallel()
	want := []arr.ProviderResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/notification", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetNotifications(context.Background())
	if err != nil {
		t.Fatalf("GetNotifications: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetNotification(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/notification/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetNotification(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetNotification: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateNotification(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodPost, "/api/v3/notification", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateNotification(context.Background(), &arr.ProviderResource{})
	if err != nil {
		t.Fatalf("CreateNotification: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateNotification(t *testing.T) {
	t.Parallel()
	want := arr.ProviderResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/notification/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateNotification(context.Background(), &arr.ProviderResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateNotification: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestDeleteNotification(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/notification/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteNotification(context.Background(), 1); err != nil {
		t.Fatalf("DeleteNotification: %v", err)
	}
}

func TestGetNotificationSchema(t *testing.T) {
	t.Parallel()
	want := []arr.ProviderResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/notification/schema", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetNotificationSchema(context.Background())
	if err != nil {
		t.Fatalf("GetNotificationSchema: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestTestNotification(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/notification/test", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.TestNotification(context.Background(), &arr.ProviderResource{ID: 1}); err != nil {
		t.Fatalf("TestNotification: %v", err)
	}
}

// ---------- QualityDefinition ----------

func TestGetQualityDefinitions(t *testing.T) {
	t.Parallel()
	want := []arr.QualityDefinitionResource{{ID: 1, Title: "HDTV-720p"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/qualitydefinition", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetQualityDefinitions(context.Background())
	if err != nil {
		t.Fatalf("GetQualityDefinitions: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetQualityDefinition(t *testing.T) {
	t.Parallel()
	want := arr.QualityDefinitionResource{ID: 1, Title: "HDTV-720p"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/qualitydefinition/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetQualityDefinition(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetQualityDefinition: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateQualityDefinition(t *testing.T) {
	t.Parallel()
	want := arr.QualityDefinitionResource{ID: 1, Title: "Updated"}
	srv := newTestServer(t, http.MethodPut, "/api/v3/qualitydefinition/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateQualityDefinition(context.Background(), &arr.QualityDefinitionResource{ID: 1, Title: "Updated"})
	if err != nil {
		t.Fatalf("UpdateQualityDefinition: %v", err)
	}
	if got.Title != "Updated" {
		t.Errorf("Title = %q, want Updated", got.Title)
	}
}

// ---------- QualityProfile ----------

func TestGetQualityProfile(t *testing.T) {
	t.Parallel()
	want := arr.QualityProfile{ID: 1, Name: "Any"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/qualityprofile/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetQualityProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetQualityProfile: %v", err)
	}
	if got.Name != "Any" {
		t.Errorf("Name = %q, want Any", got.Name)
	}
}

func TestCreateQualityProfile(t *testing.T) {
	t.Parallel()
	want := arr.QualityProfile{ID: 1, Name: "New"}
	srv := newTestServer(t, http.MethodPost, "/api/v3/qualityprofile", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateQualityProfile(context.Background(), &arr.QualityProfile{Name: "New"})
	if err != nil {
		t.Fatalf("CreateQualityProfile: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateQualityProfile(t *testing.T) {
	t.Parallel()
	want := arr.QualityProfile{ID: 1, Name: "Updated"}
	srv := newTestServer(t, http.MethodPut, "/api/v3/qualityprofile/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateQualityProfile(context.Background(), &arr.QualityProfile{ID: 1, Name: "Updated"})
	if err != nil {
		t.Fatalf("UpdateQualityProfile: %v", err)
	}
	if got.Name != "Updated" {
		t.Errorf("Name = %q, want Updated", got.Name)
	}
}

func TestDeleteQualityProfile(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/qualityprofile/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteQualityProfile(context.Background(), 1); err != nil {
		t.Fatalf("DeleteQualityProfile: %v", err)
	}
}

// ---------- Queue ----------

func TestDeleteQueueItems(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/queue/bulk?removeFromClient=true&blocklist=false", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteQueueItems(context.Background(), []int{1, 2}, true, false); err != nil {
		t.Fatalf("DeleteQueueItems: %v", err)
	}
}

func TestGrabQueueItem(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/queue/grab/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.GrabQueueItem(context.Background(), 1); err != nil {
		t.Fatalf("GrabQueueItem: %v", err)
	}
}

func TestGrabQueueItems(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/queue/grab/bulk", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.GrabQueueItems(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("GrabQueueItems: %v", err)
	}
}

func TestGetQueueDetails(t *testing.T) {
	t.Parallel()
	want := []arr.QueueRecord{{ID: 1, Title: "test"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/queue/details", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetQueueDetails(context.Background())
	if err != nil {
		t.Fatalf("GetQueueDetails: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetQueueStatus(t *testing.T) {
	t.Parallel()
	want := arr.QueueStatusResource{TotalCount: 5}
	srv := newTestServer(t, http.MethodGet, "/api/v3/queue/status", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetQueueStatus(context.Background())
	if err != nil {
		t.Fatalf("GetQueueStatus: %v", err)
	}
	if got.TotalCount != 5 {
		t.Errorf("TotalCount = %d, want 5", got.TotalCount)
	}
}

// ---------- Release ----------

func TestGetReleaseProfiles(t *testing.T) {
	t.Parallel()
	want := []arr.ReleaseProfileResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/releaseprofile", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetReleaseProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetReleaseProfiles: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetReleaseProfile(t *testing.T) {
	t.Parallel()
	want := arr.ReleaseProfileResource{ID: 1}
	srv := newTestServer(t, http.MethodGet, "/api/v3/releaseprofile/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetReleaseProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetReleaseProfile: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateReleaseProfile(t *testing.T) {
	t.Parallel()
	want := arr.ReleaseProfileResource{ID: 1}
	srv := newTestServer(t, http.MethodPost, "/api/v3/releaseprofile", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateReleaseProfile(context.Background(), &arr.ReleaseProfileResource{})
	if err != nil {
		t.Fatalf("CreateReleaseProfile: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateReleaseProfile(t *testing.T) {
	t.Parallel()
	want := arr.ReleaseProfileResource{ID: 1}
	srv := newTestServer(t, http.MethodPut, "/api/v3/releaseprofile/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateReleaseProfile(context.Background(), &arr.ReleaseProfileResource{ID: 1})
	if err != nil {
		t.Fatalf("UpdateReleaseProfile: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestDeleteReleaseProfile(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/releaseprofile/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteReleaseProfile(context.Background(), 1); err != nil {
		t.Fatalf("DeleteReleaseProfile: %v", err)
	}
}

func TestSearchReleases(t *testing.T) {
	t.Parallel()
	want := []arr.ReleaseResource{{GUID: "abc"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/release?episodeId=10", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.SearchReleases(context.Background(), 10)
	if err != nil {
		t.Fatalf("SearchReleases: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestPushRelease(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/release/push", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.PushRelease(context.Background(), &arr.ReleasePushResource{}); err != nil {
		t.Fatalf("PushRelease: %v", err)
	}
}

func TestGrabRelease(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/release", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.GrabRelease(context.Background(), "guid-abc", 1); err != nil {
		t.Fatalf("GrabRelease: %v", err)
	}
}

// ---------- RemotePathMapping ----------

func TestGetRemotePathMappings(t *testing.T) {
	t.Parallel()
	want := []arr.RemotePathMappingResource{{ID: 1, Host: "host"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/remotepathmapping", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetRemotePathMappings(context.Background())
	if err != nil {
		t.Fatalf("GetRemotePathMappings: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetRemotePathMapping(t *testing.T) {
	t.Parallel()
	want := arr.RemotePathMappingResource{ID: 1, Host: "host"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/remotepathmapping/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetRemotePathMapping(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetRemotePathMapping: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestCreateRemotePathMapping(t *testing.T) {
	t.Parallel()
	want := arr.RemotePathMappingResource{ID: 1, Host: "host"}
	srv := newTestServer(t, http.MethodPost, "/api/v3/remotepathmapping", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateRemotePathMapping(context.Background(), &arr.RemotePathMappingResource{Host: "host"})
	if err != nil {
		t.Fatalf("CreateRemotePathMapping: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestUpdateRemotePathMapping(t *testing.T) {
	t.Parallel()
	want := arr.RemotePathMappingResource{ID: 1, Host: "updated"}
	srv := newTestServer(t, http.MethodPut, "/api/v3/remotepathmapping/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateRemotePathMapping(context.Background(), &arr.RemotePathMappingResource{ID: 1, Host: "updated"})
	if err != nil {
		t.Fatalf("UpdateRemotePathMapping: %v", err)
	}
	if got.Host != "updated" {
		t.Errorf("Host = %q, want updated", got.Host)
	}
}

func TestDeleteRemotePathMapping(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/remotepathmapping/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteRemotePathMapping(context.Background(), 1); err != nil {
		t.Fatalf("DeleteRemotePathMapping: %v", err)
	}
}

// ---------- Rename / ManualImport ----------

func TestGetRenameList(t *testing.T) {
	t.Parallel()
	want := []arr.RenameEpisodeResource{{EpisodeID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/rename?seriesId=1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetRenameList(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetRenameList: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetManualImport(t *testing.T) {
	t.Parallel()
	want := []arr.ManualImportResource{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/manualimport?folder=%2Ftv&downloadId=abc", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetManualImport(context.Background(), "/tv", "abc")
	if err != nil {
		t.Fatalf("GetManualImport: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestProcessManualImport(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/manualimport", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.ProcessManualImport(context.Background(), []arr.ManualImportReprocessResource{{ID: 1}}); err != nil {
		t.Fatalf("ProcessManualImport: %v", err)
	}
}

// ---------- RootFolder ----------

func TestGetRootFolder(t *testing.T) {
	t.Parallel()
	want := arr.RootFolder{ID: 1, Path: "/tv"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/rootfolder/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetRootFolder(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetRootFolder: %v", err)
	}
	if got.Path != "/tv" {
		t.Errorf("Path = %q, want /tv", got.Path)
	}
}

func TestCreateRootFolder(t *testing.T) {
	t.Parallel()
	want := arr.RootFolder{ID: 1, Path: "/tv"}
	srv := newTestServer(t, http.MethodPost, "/api/v3/rootfolder", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.CreateRootFolder(context.Background(), &arr.RootFolder{Path: "/tv"})
	if err != nil {
		t.Fatalf("CreateRootFolder: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestDeleteRootFolder(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/rootfolder/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteRootFolder(context.Background(), 1); err != nil {
		t.Fatalf("DeleteRootFolder: %v", err)
	}
}

// ---------- Series (editor, import, folder) ----------

func TestEditSeries(t *testing.T) {
	t.Parallel()
	want := []sonarr.Series{{ID: 1, Title: "Edited"}}
	srv := newTestServer(t, http.MethodPut, "/api/v3/series/editor", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.EditSeries(context.Background(), &sonarr.SeriesEditorResource{SeriesIDs: []int{1}})
	if err != nil {
		t.Fatalf("EditSeries: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestDeleteManySeries(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/series/editor", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteManySeries(context.Background(), &sonarr.SeriesEditorResource{SeriesIDs: []int{1, 2}}); err != nil {
		t.Fatalf("DeleteManySeries: %v", err)
	}
}

func TestImportSeries(t *testing.T) {
	t.Parallel()
	want := []sonarr.Series{{ID: 1, Title: "Imported"}}
	srv := newTestServer(t, http.MethodPost, "/api/v3/series/import", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.ImportSeries(context.Background(), []sonarr.Series{{Title: "Imported"}})
	if err != nil {
		t.Fatalf("ImportSeries: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetSeriesFolder(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodGet, "/api/v3/series/1/folder", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.GetSeriesFolder(context.Background(), 1); err != nil {
		t.Fatalf("GetSeriesFolder: %v", err)
	}
}

// ---------- System ----------

func TestGetBackups(t *testing.T) {
	t.Parallel()
	want := []arr.Backup{{ID: 1}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/system/backup", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetBackups(context.Background())
	if err != nil {
		t.Fatalf("GetBackups: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestDeleteBackup(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/system/backup/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteBackup(context.Background(), 1); err != nil {
		t.Fatalf("DeleteBackup: %v", err)
	}
}

func TestRestoreBackup(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/system/backup/restore/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.RestoreBackup(context.Background(), 1); err != nil {
		t.Fatalf("RestoreBackup: %v", err)
	}
}

func TestGetSystemRoutes(t *testing.T) {
	t.Parallel()
	want := []arr.SystemRouteResource{{Path: "/test", Method: "GET"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/system/routes", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetSystemRoutes(context.Background())
	if err != nil {
		t.Fatalf("GetSystemRoutes: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestShutdown(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/system/shutdown", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestRestart(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/system/restart", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.Restart(context.Background()); err != nil {
		t.Fatalf("Restart: %v", err)
	}
}

func TestGetTasks(t *testing.T) {
	t.Parallel()
	want := []arr.TaskResource{{ID: 1, Name: "RefreshSeries"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/system/task", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetTasks(context.Background())
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetTask(t *testing.T) {
	t.Parallel()
	want := arr.TaskResource{ID: 1, Name: "RefreshSeries"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/system/task/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.Name != "RefreshSeries" {
		t.Errorf("Name = %q, want RefreshSeries", got.Name)
	}
}

func TestGetFileSystemType(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodGet, "/api/v3/filesystem/type?path=%2Ftv", "local")
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetFileSystemType(context.Background(), "/tv")
	if err != nil {
		t.Fatalf("GetFileSystemType: %v", err)
	}
	if got != "local" {
		t.Errorf("got = %q, want local", got)
	}
}

func TestGetFileSystemMediaFiles(t *testing.T) {
	t.Parallel()
	want := []string{"file1.mkv", "file2.mkv"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/filesystem/mediafiles?path=%2Ftv", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetFileSystemMediaFiles(context.Background(), "/tv")
	if err != nil {
		t.Fatalf("GetFileSystemMediaFiles: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

// ---------- Tag ----------

func TestGetTag(t *testing.T) {
	t.Parallel()
	want := arr.Tag{ID: 1, Label: "hd"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/tag/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetTag(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTag: %v", err)
	}
	if got.Label != "hd" {
		t.Errorf("Label = %q, want hd", got.Label)
	}
}

func TestUpdateTag(t *testing.T) {
	t.Parallel()
	want := arr.Tag{ID: 1, Label: "updated"}
	srv := newTestServer(t, http.MethodPut, "/api/v3/tag/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.UpdateTag(context.Background(), &arr.Tag{ID: 1, Label: "updated"})
	if err != nil {
		t.Fatalf("UpdateTag: %v", err)
	}
	if got.Label != "updated" {
		t.Errorf("Label = %q, want updated", got.Label)
	}
}

func TestDeleteTag(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodDelete, "/api/v3/tag/1", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.DeleteTag(context.Background(), 1); err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}
}

func TestGetTagDetails(t *testing.T) {
	t.Parallel()
	want := []arr.TagDetail{{ID: 1, Label: "hd"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/tag/detail", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetTagDetails(context.Background())
	if err != nil {
		t.Fatalf("GetTagDetails: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetTagDetail(t *testing.T) {
	t.Parallel()
	want := arr.TagDetail{ID: 1, Label: "hd"}
	srv := newTestServer(t, http.MethodGet, "/api/v3/tag/detail/1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetTagDetail(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTagDetail: %v", err)
	}
	if got.Label != "hd" {
		t.Errorf("Label = %q, want hd", got.Label)
	}
}

// ---------- Update ----------

func TestGetUpdates(t *testing.T) {
	t.Parallel()
	want := []arr.UpdateResource{{Version: "4.0.1"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/update", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetUpdates(context.Background())
	if err != nil {
		t.Fatalf("GetUpdates: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

// ---------- Wanted ----------

func TestGetWantedMissing(t *testing.T) {
	t.Parallel()
	want := arr.PagingResource[sonarr.Episode]{
		Page: 1, PageSize: 10, TotalRecords: 1,
		Records: []sonarr.Episode{{ID: 1, Title: "Missing"}},
	}
	srv := newTestServer(t, http.MethodGet, "/api/v3/wanted/missing?page=1&pageSize=10", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetWantedMissing(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetWantedMissing: %v", err)
	}
	if got.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d, want 1", got.TotalRecords)
	}
}

func TestGetWantedCutoff(t *testing.T) {
	t.Parallel()
	want := arr.PagingResource[sonarr.Episode]{
		Page: 1, PageSize: 10, TotalRecords: 1,
		Records: []sonarr.Episode{{ID: 1, Title: "Cutoff"}},
	}
	srv := newTestServer(t, http.MethodGet, "/api/v3/wanted/cutoff?page=1&pageSize=10", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetWantedCutoff(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetWantedCutoff: %v", err)
	}
	if got.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d, want 1", got.TotalRecords)
	}
}

// ---------- History ----------

func TestGetHistorySeries(t *testing.T) {
	t.Parallel()
	want := []sonarr.HistoryRecord{{ID: 1, SeriesID: 1, EventType: "grabbed"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/history/series?seriesId=1", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetHistorySeries(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetHistorySeries: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetHistorySince(t *testing.T) {
	t.Parallel()
	want := []sonarr.HistoryRecord{{ID: 1, EventType: "grabbed"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/history/since?date=2026-01-01", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetHistorySince(context.Background(), "2026-01-01")
	if err != nil {
		t.Fatalf("GetHistorySince: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestMarkHistoryFailed(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, http.MethodPost, "/api/v3/history/failed/5", nil)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	if err := c.MarkHistoryFailed(context.Background(), 5); err != nil {
		t.Fatalf("MarkHistoryFailed: %v", err)
	}
}

// ---------- Language ----------

func TestGetLanguages(t *testing.T) {
	t.Parallel()
	want := []sonarr.Language{{ID: 1, Name: "English"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/language", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLanguages(context.Background())
	if err != nil {
		t.Fatalf("GetLanguages: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

// ---------- Log ----------

func TestGetLogs(t *testing.T) {
	t.Parallel()
	want := arr.PagingResource[arr.LogRecord]{
		Page: 1, PageSize: 10, TotalRecords: 1,
		Records: []arr.LogRecord{{ID: 1}},
	}
	srv := newTestServer(t, http.MethodGet, "/api/v3/log?page=1&pageSize=10", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLogs(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetLogs: %v", err)
	}
	if got.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d, want 1", got.TotalRecords)
	}
}

func TestGetLogFiles(t *testing.T) {
	t.Parallel()
	want := []arr.LogFileResource{{Filename: "sonarr.txt"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/log/file", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetLogFiles(context.Background())
	if err != nil {
		t.Fatalf("GetLogFiles: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetUpdateLogFiles(t *testing.T) {
	t.Parallel()
	want := []arr.LogFileResource{{Filename: "update.txt"}}
	srv := newTestServer(t, http.MethodGet, "/api/v3/log/file/update", want)
	defer srv.Close()
	c, _ := sonarr.New(srv.URL, "test-key")
	got, err := c.GetUpdateLogFiles(context.Background())
	if err != nil {
		t.Fatalf("GetUpdateLogFiles: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}
