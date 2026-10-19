package whisparr_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/arr/v2"
	"github.com/golusoris/goenvoy/arr/whisparr"
)

func TestQueueRecord_EmbeddingDecode_V3(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": 201,
		"title": "Scene 1",
		"status": "downloading",
		"sizeleft": 104857600.0,
		"customFormatScore": 90,
		"movieId": 55
	}`

	var rec whisparr.QueueRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if rec.ID != 201 {
		t.Errorf("rec.ID = %d, want 201", rec.ID)
	}
	if rec.Status != "downloading" {
		t.Errorf("rec.Status = %q, want %q", rec.Status, "downloading")
	}
	if rec.SizeLeft != 104857600.0 {
		t.Errorf("rec.SizeLeft = %f, want 104857600.0", rec.SizeLeft)
	}
	if rec.CustomFormatScore != 90 {
		t.Errorf("rec.CustomFormatScore = %d, want 90", rec.CustomFormatScore)
	}
	if rec.MovieID != 55 {
		t.Errorf("rec.MovieID = %d, want 55", rec.MovieID)
	}
}

func TestQueueRecord_EmbeddingDecode_V2(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": 301,
		"title": "Scene 2",
		"status": "downloading",
		"sizeleft": 209715200.0,
		"customFormatScore": 85,
		"seriesId": 12,
		"episodeId": 34,
		"seasonNumber": 1
	}`

	var rec whisparr.QueueRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if rec.ID != 301 {
		t.Errorf("rec.ID = %d, want 301", rec.ID)
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
	if rec.CustomFormatScore != 85 {
		t.Errorf("rec.CustomFormatScore = %d, want 85", rec.CustomFormatScore)
	}
}

type queueQueryTestCase struct {
	name     string
	opts     whisparr.QueueOptions
	wantPath string
}

func v2QueueQueryTestCases() []queueQueryTestCase {
	return []queueQueryTestCase{
		{
			name:     "all-false",
			opts:     whisparr.QueueOptions{Page: 1, PageSize: 10},
			wantPath: "/api/v3/queue?page=1&pageSize=10",
		},
		{
			name:     "includeUnknown alone",
			opts:     whisparr.QueueOptions{Page: 1, PageSize: 10, IncludeUnknown: true},
			wantPath: "/api/v3/queue?includeUnknownSeriesItems=true&page=1&pageSize=10",
		},
		{
			name:     "includeParent alone",
			opts:     whisparr.QueueOptions{Page: 1, PageSize: 10, IncludeParent: true},
			wantPath: "/api/v3/queue?includeSeries=true&page=1&pageSize=10",
		},
		{
			name:     "includeItem alone",
			opts:     whisparr.QueueOptions{Page: 1, PageSize: 10, IncludeItem: true},
			wantPath: "/api/v3/queue?includeEpisode=true&page=1&pageSize=10",
		},
		{
			name: "all-true",
			opts: whisparr.QueueOptions{
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
			opts:     whisparr.QueueOptions{Page: 0, PageSize: 0},
			wantPath: "/api/v3/queue?page=0&pageSize=0",
		},
		{
			name:     "boundary large",
			opts:     whisparr.QueueOptions{Page: 1000, PageSize: 500},
			wantPath: "/api/v3/queue?page=1000&pageSize=500",
		},
	}
}

func v3QueueQueryTestCases() []queueQueryTestCase {
	return []queueQueryTestCase{
		{
			name:     "all-false",
			opts:     whisparr.QueueOptions{Page: 1, PageSize: 10},
			wantPath: "/api/v3/queue?page=1&pageSize=10",
		},
		{
			name:     "includeUnknown alone",
			opts:     whisparr.QueueOptions{Page: 1, PageSize: 10, IncludeUnknown: true},
			wantPath: "/api/v3/queue?includeUnknownMovieItems=true&page=1&pageSize=10",
		},
		{
			name:     "includeParent alone ignored",
			opts:     whisparr.QueueOptions{Page: 1, PageSize: 10, IncludeParent: true},
			wantPath: "/api/v3/queue?page=1&pageSize=10",
		},
		{
			name:     "includeItem alone",
			opts:     whisparr.QueueOptions{Page: 1, PageSize: 10, IncludeItem: true},
			wantPath: "/api/v3/queue?includeMovie=true&page=1&pageSize=10",
		},
		{
			name: "all-true",
			opts: whisparr.QueueOptions{
				Page:           1,
				PageSize:       10,
				IncludeUnknown: true,
				IncludeParent:  true,
				IncludeItem:    true,
			},
			wantPath: "/api/v3/queue?includeMovie=true&includeUnknownMovieItems=true&page=1&pageSize=10",
		},
		{
			name:     "boundary zero",
			opts:     whisparr.QueueOptions{Page: 0, PageSize: 0},
			wantPath: "/api/v3/queue?page=0&pageSize=0",
		},
		{
			name:     "boundary large",
			opts:     whisparr.QueueOptions{Page: 1000, PageSize: 500},
			wantPath: "/api/v3/queue?page=1000&pageSize=500",
		},
	}
}

func runV2GetQueueWithOptionsQueryTest(t *testing.T, tc queueQueryTestCase) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.RequestURI() != tc.wantPath {
			t.Errorf("RequestURI = %q, want %q", r.URL.RequestURI(), tc.wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(arr.PagingResource[whisparr.QueueRecord]{
			Page:         tc.opts.Page,
			PageSize:     tc.opts.PageSize,
			TotalRecords: 0,
			Records:      []whisparr.QueueRecord{},
		})
	}))
	defer srv.Close()

	c, err := whisparr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.New: %v", err)
	}

	res, err := c.GetQueueWithOptions(context.Background(), tc.opts)
	if err != nil {
		t.Fatalf("GetQueueWithOptions: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil PagingResource")
	}
}

func TestGetQueueWithOptions_V2QueryString(t *testing.T) {
	t.Parallel()

	for _, tc := range v2QueueQueryTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runV2GetQueueWithOptionsQueryTest(t, tc)
		})
	}
}

func runV3GetQueueWithOptionsQueryTest(t *testing.T, tc queueQueryTestCase) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.RequestURI() != tc.wantPath {
			t.Errorf("RequestURI = %q, want %q", r.URL.RequestURI(), tc.wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(arr.PagingResource[whisparr.QueueRecord]{
			Page:         tc.opts.Page,
			PageSize:     tc.opts.PageSize,
			TotalRecords: 0,
			Records:      []whisparr.QueueRecord{},
		})
	}))
	defer srv.Close()

	c, err := whisparr.NewV3(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.NewV3: %v", err)
	}

	res, err := c.GetQueueWithOptions(context.Background(), tc.opts)
	if err != nil {
		t.Fatalf("GetQueueWithOptions: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil PagingResource")
	}
}

func TestGetQueueWithOptions_V3QueryString(t *testing.T) {
	t.Parallel()

	for _, tc := range v3QueueQueryTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runV3GetQueueWithOptionsQueryTest(t, tc)
		})
	}
}

func TestGetQueueWithOptions_NestedMonitored_V2(t *testing.T) {
	t.Parallel()

	respData := arr.PagingResource[whisparr.QueueRecord]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []whisparr.QueueRecord{
			{
				QueueRecord: arr.QueueRecord{
					ID:       1,
					Title:    "Scene 1",
					SizeLeft: 2048,
					Status:   "downloading",
				},
				CustomFormatScore: 120,
				SeriesID:          55,
				Series: &whisparr.Series{
					ID:        55,
					Title:     "Series Title",
					Monitored: true,
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(respData)
	}))
	defer srv.Close()

	c, err := whisparr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.New: %v", err)
	}

	got, err := c.GetQueueWithOptions(context.Background(), whisparr.QueueOptions{
		Page:          1,
		PageSize:      10,
		IncludeParent: true,
	})
	if err != nil {
		t.Fatalf("GetQueueWithOptions: %v", err)
	}

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
}

func TestGetQueueWithOptions_NestedMonitored_V3(t *testing.T) {
	t.Parallel()

	respData := arr.PagingResource[whisparr.QueueRecord]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []whisparr.QueueRecord{
			{
				QueueRecord: arr.QueueRecord{
					ID:       1,
					Title:    "Movie 1",
					SizeLeft: 2048,
					Status:   "downloading",
				},
				CustomFormatScore: 120,
				MovieID:           55,
				Movie: &whisparr.Movie{
					ID:        55,
					Title:     "Movie Title",
					Monitored: true,
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(respData)
	}))
	defer srv.Close()

	c, err := whisparr.NewV3(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.NewV3: %v", err)
	}

	got, err := c.GetQueueWithOptions(context.Background(), whisparr.QueueOptions{
		Page:        1,
		PageSize:    10,
		IncludeItem: true,
	})
	if err != nil {
		t.Fatalf("GetQueueWithOptions: %v", err)
	}

	if len(got.Records) != 1 {
		t.Fatalf("got %d records, want 1", len(got.Records))
	}
	rec := got.Records[0]
	if rec.Movie == nil {
		t.Fatal("expected non-nil Movie")
	}
	if !rec.Movie.Monitored {
		t.Errorf("Movie.Monitored = false, want true")
	}
}

func TestGetQueueWithOptions_APIError_V2(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer srv.Close()

	c, err := whisparr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.New: %v", err)
	}

	_, err = c.GetQueueWithOptions(context.Background(), whisparr.QueueOptions{Page: 1, PageSize: 10})
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

func TestGetQueueWithOptions_APIError_V3(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer srv.Close()

	c, err := whisparr.NewV3(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.NewV3: %v", err)
	}

	_, err = c.GetQueueWithOptions(context.Background(), whisparr.QueueOptions{Page: 1, PageSize: 10})
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

type deleteQueueTestCase struct {
	name     string
	opts     whisparr.QueueDeleteOptions
	wantPath string
}

func runV2DeleteQueueTest(t *testing.T, tc deleteQueueTestCase) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.RequestURI() != tc.wantPath {
			t.Errorf("RequestURI = %q, want %q", r.URL.RequestURI(), tc.wantPath)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := whisparr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.New: %v", err)
	}

	if err := c.DeleteQueueItemWithOptions(context.Background(), 42, tc.opts); err != nil {
		t.Fatalf("DeleteQueueItemWithOptions: %v", err)
	}
}

func TestDeleteQueueItemWithOptions_V2(t *testing.T) {
	t.Parallel()

	tests := []deleteQueueTestCase{
		{
			name:     "all-false",
			opts:     whisparr.QueueDeleteOptions{},
			wantPath: "/api/v3/queue/42?removeFromClient=false&blocklist=false&skipRedownload=false",
		},
		{
			name: "skipRedownload alone",
			opts: whisparr.QueueDeleteOptions{
				SkipRedownload: true,
			},
			wantPath: "/api/v3/queue/42?removeFromClient=false&blocklist=false&skipRedownload=true",
		},
		{
			name: "all-true",
			opts: whisparr.QueueDeleteOptions{
				RemoveFromClient: true,
				Blocklist:        true,
				SkipRedownload:   true,
				ChangeCategory:   true,
			},
			wantPath: "/api/v3/queue/42?removeFromClient=true&blocklist=true&skipRedownload=true",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runV2DeleteQueueTest(t, tc)
		})
	}
}

func runV3DeleteQueueTest(t *testing.T, tc deleteQueueTestCase) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.RequestURI() != tc.wantPath {
			t.Errorf("RequestURI = %q, want %q", r.URL.RequestURI(), tc.wantPath)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := whisparr.NewV3(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.NewV3: %v", err)
	}

	if err := c.DeleteQueueItemWithOptions(context.Background(), 42, tc.opts); err != nil {
		t.Fatalf("DeleteQueueItemWithOptions: %v", err)
	}
}

func TestDeleteQueueItemWithOptions_V3(t *testing.T) {
	t.Parallel()

	tests := []deleteQueueTestCase{
		{
			name:     "all-false",
			opts:     whisparr.QueueDeleteOptions{},
			wantPath: "/api/v3/queue/42?removeFromClient=false&blocklist=false&skipRedownload=false&changeCategory=false",
		},
		{
			name: "skipRedownload alone",
			opts: whisparr.QueueDeleteOptions{
				SkipRedownload: true,
			},
			wantPath: "/api/v3/queue/42?removeFromClient=false&blocklist=false&skipRedownload=true&changeCategory=false",
		},
		{
			name: "all-true",
			opts: whisparr.QueueDeleteOptions{
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
			runV3DeleteQueueTest(t, tc)
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

	c, err := whisparr.NewV3(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.NewV3: %v", err)
	}

	err = c.DeleteQueueItemWithOptions(context.Background(), 7, whisparr.QueueDeleteOptions{
		SkipRedownload: true,
	})
	if err != nil {
		t.Fatalf("DeleteQueueItemWithOptions: %v", err)
	}
	if !serverReceivedSkipRedownload {
		t.Fatal("server did not receive skipRedownload=true")
	}
}

func TestDeleteQueueItemWithOptions_APIError_V2(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer srv.Close()

	c, err := whisparr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.New: %v", err)
	}

	err = c.DeleteQueueItemWithOptions(context.Background(), 1, whisparr.QueueDeleteOptions{})
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

func TestDeleteQueueItemWithOptions_APIError_V3(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer srv.Close()

	c, err := whisparr.NewV3(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.NewV3: %v", err)
	}

	err = c.DeleteQueueItemWithOptions(context.Background(), 1, whisparr.QueueDeleteOptions{})
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
