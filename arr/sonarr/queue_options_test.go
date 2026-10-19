package sonarr_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/arr/sonarr"
	"github.com/golusoris/goenvoy/arr/v2"
)

func TestQueueRecord_EmbeddingDecode(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": 101,
		"title": "Breaking Bad - S01E01 - Pilot",
		"status": "downloading",
		"sizeleft": 52428800.0,
		"customFormatScore": 85,
		"seriesId": 12,
		"episodeId": 34,
		"seasonNumber": 1
	}`

	var rec sonarr.QueueRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if rec.ID != 101 {
		t.Errorf("rec.ID = %d, want 101", rec.ID)
	}
	if rec.Status != "downloading" {
		t.Errorf("rec.Status = %q, want %q", rec.Status, "downloading")
	}
	if rec.SizeLeft != 52428800.0 {
		t.Errorf("rec.SizeLeft = %f, want 52428800.0", rec.SizeLeft)
	}
	if rec.CustomFormatScore != 85 {
		t.Errorf("rec.CustomFormatScore = %d, want 85", rec.CustomFormatScore)
	}
	if rec.SeriesID != 12 {
		t.Errorf("rec.SeriesID = %d, want 12", rec.SeriesID)
	}
	if rec.EpisodeID != 34 {
		t.Errorf("rec.EpisodeID = %d, want 34", rec.EpisodeID)
	}
	if rec.SeasonNumber != 1 {
		t.Errorf("rec.SeasonNumber = %d, want 1", rec.SeasonNumber)
	}
}

func newQueueServer(t *testing.T, handler http.HandlerFunc) *sonarr.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("sonarr.New: %v", err)
	}
	return c
}

type queueQueryTestCase struct {
	name     string
	opts     sonarr.QueueOptions
	wantPath string
}

func queueQueryTests() []queueQueryTestCase {
	return []queueQueryTestCase{
		{
			name:     "all-false",
			opts:     sonarr.QueueOptions{Page: 1, PageSize: 10},
			wantPath: "/api/v3/queue?page=1&pageSize=10",
		},
		{
			name:     "includeUnknown alone",
			opts:     sonarr.QueueOptions{Page: 1, PageSize: 10, IncludeUnknown: true},
			wantPath: "/api/v3/queue?includeUnknownSeriesItems=true&page=1&pageSize=10",
		},
		{
			name:     "includeParent alone",
			opts:     sonarr.QueueOptions{Page: 1, PageSize: 10, IncludeParent: true},
			wantPath: "/api/v3/queue?includeSeries=true&page=1&pageSize=10",
		},
		{
			name:     "includeItem alone",
			opts:     sonarr.QueueOptions{Page: 1, PageSize: 10, IncludeItem: true},
			wantPath: "/api/v3/queue?includeEpisode=true&page=1&pageSize=10",
		},
		{
			name: "all-true",
			opts: sonarr.QueueOptions{
				Page:           1,
				PageSize:       10,
				IncludeUnknown: true,
				IncludeParent:  true,
				IncludeItem:    true,
			},
			wantPath: "/api/v3/queue?includeEpisode=true&includeSeries=true&includeUnknownSeriesItems=true&page=1&pageSize=10",
		},
		{
			name:     "boundary zero",
			opts:     sonarr.QueueOptions{Page: 0, PageSize: 0},
			wantPath: "/api/v3/queue?page=0&pageSize=0",
		},
		{
			name:     "boundary large",
			opts:     sonarr.QueueOptions{Page: 1000, PageSize: 500},
			wantPath: "/api/v3/queue?page=1000&pageSize=500",
		},
	}
}

func checkGetQueueQueryString(t *testing.T, opts sonarr.QueueOptions, wantPath string) {
	t.Helper()
	c := newQueueServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.RequestURI() != wantPath {
			t.Errorf("RequestURI = %q, want %q", r.URL.RequestURI(), wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(arr.PagingResource[sonarr.QueueRecord]{
			Page:         opts.Page,
			PageSize:     opts.PageSize,
			TotalRecords: 0,
			Records:      []sonarr.QueueRecord{},
		})
	})

	res, err := c.GetQueueWithOptions(context.Background(), opts)
	if err != nil {
		t.Fatalf("GetQueueWithOptions: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil PagingResource")
	}
}

func TestGetQueueWithOptions_QueryString(t *testing.T) {
	t.Parallel()

	for _, tc := range queueQueryTests() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkGetQueueQueryString(t, tc.opts, tc.wantPath)
		})
	}
}

func nestedMonitoredQueueResponse() arr.PagingResource[sonarr.QueueRecord] {
	return arr.PagingResource[sonarr.QueueRecord]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []sonarr.QueueRecord{
			{
				QueueRecord: arr.QueueRecord{
					ID:       1,
					Title:    "Breaking Bad - S01E01",
					SizeLeft: 1024,
					Status:   "downloading",
				},
				CustomFormatScore: 100,
				SeriesID:          10,
				EpisodeID:         20,
				SeasonNumber:      1,
				Series: &sonarr.Series{
					ID:        10,
					Title:     "Breaking Bad",
					Monitored: true,
				},
				Episode: &sonarr.Episode{
					ID:        20,
					Title:     "Pilot",
					Monitored: true,
				},
			},
		},
	}
}

func assertNestedMonitoredRecord(t *testing.T, got *arr.PagingResource[sonarr.QueueRecord]) {
	t.Helper()
	if len(got.Records) != 1 {
		t.Fatalf("got %d records, want 1", len(got.Records))
	}
	rec := got.Records[0]
	if rec.Series == nil {
		t.Fatal("expected non-nil Series")
	}
	if !rec.Series.Monitored {
		t.Errorf("Series.Monitored = false, want true")
	}
	if rec.Episode == nil {
		t.Fatal("expected non-nil Episode")
	}
	if !rec.Episode.Monitored {
		t.Errorf("Episode.Monitored = false, want true")
	}
}

func TestGetQueueWithOptions_NestedMonitored(t *testing.T) {
	t.Parallel()

	respData := nestedMonitoredQueueResponse()
	c := newQueueServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(respData)
	})

	got, err := c.GetQueueWithOptions(context.Background(), sonarr.QueueOptions{
		Page:          1,
		PageSize:      10,
		IncludeParent: true,
		IncludeItem:   true,
	})
	if err != nil {
		t.Fatalf("GetQueueWithOptions: %v", err)
	}

	assertNestedMonitoredRecord(t, got)
}

func TestGetQueueWithOptions_APIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("sonarr.New: %v", err)
	}

	_, err = c.GetQueueWithOptions(context.Background(), sonarr.QueueOptions{Page: 1, PageSize: 10})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *arr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *arr.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}

func checkDeleteQueueItemWithOptions(t *testing.T, opts sonarr.QueueDeleteOptions, wantPath string) {
	t.Helper()
	c := newQueueServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.RequestURI() != wantPath {
			t.Errorf("RequestURI = %q, want %q", r.URL.RequestURI(), wantPath)
		}
		w.WriteHeader(http.StatusOK)
	})

	if err := c.DeleteQueueItemWithOptions(context.Background(), 42, opts); err != nil {
		t.Fatalf("DeleteQueueItemWithOptions: %v", err)
	}
}

func TestDeleteQueueItemWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     sonarr.QueueDeleteOptions
		wantPath string
	}{
		{
			name:     "all-false",
			opts:     sonarr.QueueDeleteOptions{},
			wantPath: "/api/v3/queue/42?removeFromClient=false&blocklist=false&skipRedownload=false&changeCategory=false",
		},
		{
			name: "skipRedownload alone",
			opts: sonarr.QueueDeleteOptions{
				SkipRedownload: true,
			},
			wantPath: "/api/v3/queue/42?removeFromClient=false&blocklist=false&skipRedownload=true&changeCategory=false",
		},
		{
			name: "all-true",
			opts: sonarr.QueueDeleteOptions{
				RemoveFromClient: true,
				Blocklist:        true,
				SkipRedownload:   true,
				ChangeCategory:   true,
			},
			wantPath: "/api/v3/queue/42?removeFromClient=true&blocklist=true&skipRedownload=true&changeCategory=true",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkDeleteQueueItemWithOptions(t, tc.opts, tc.wantPath)
		})
	}
}

func TestDeleteQueueItemWithOptions_SkipRedownloadReachesServer(t *testing.T) {
	t.Parallel()

	serverReceivedSkipRedownload := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.URL.Query().Get("skipRedownload")
		if val == "true" {
			serverReceivedSkipRedownload = true
			w.WriteHeader(http.StatusOK)
			return
		}
		t.Errorf("expected skipRedownload=true, got %q", val)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("sonarr.New: %v", err)
	}

	err = c.DeleteQueueItemWithOptions(context.Background(), 7, sonarr.QueueDeleteOptions{
		SkipRedownload: true,
	})
	if err != nil {
		t.Fatalf("DeleteQueueItemWithOptions: %v", err)
	}
	if !serverReceivedSkipRedownload {
		t.Fatal("server did not receive skipRedownload=true")
	}
}

func TestDeleteQueueItemWithOptions_APIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("sonarr.New: %v", err)
	}

	err = c.DeleteQueueItemWithOptions(context.Background(), 1, sonarr.QueueDeleteOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *arr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *arr.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
}
