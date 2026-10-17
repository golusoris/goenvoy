package lidarr_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/arr/lidarr"
	"github.com/golusoris/goenvoy/arr/v2"
)

func TestQueueRecord_EmbeddingDecode(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": 301,
		"title": "Pink Floyd - The Dark Side of the Moon",
		"status": "downloading",
		"sizeleft": 209715200.0,
		"customFormatScore": 95,
		"artistId": 77,
		"albumId": 88
	}`

	var rec lidarr.QueueRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if rec.ID != 301 {
		t.Errorf("rec.ID = %d, want 301", rec.ID)
	}
	if rec.Status != "downloading" {
		t.Errorf("rec.Status = %q, want %q", rec.Status, "downloading")
	}
	if rec.SizeLeft != 209715200.0 {
		t.Errorf("rec.SizeLeft = %f, want 209715200.0", rec.SizeLeft)
	}
	if rec.CustomFormatScore != 95 {
		t.Errorf("rec.CustomFormatScore = %d, want 95", rec.CustomFormatScore)
	}
	if rec.ArtistID != 77 {
		t.Errorf("rec.ArtistID = %d, want 77", rec.ArtistID)
	}
	if rec.AlbumID != 88 {
		t.Errorf("rec.AlbumID = %d, want 88", rec.AlbumID)
	}
}

func TestGetQueueWithOptions_QueryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     lidarr.QueueOptions
		wantPath string
	}{
		{
			name:     "all-false",
			opts:     lidarr.QueueOptions{Page: 1, PageSize: 10},
			wantPath: "/api/v1/queue?page=1&pageSize=10",
		},
		{
			name:     "includeUnknown alone",
			opts:     lidarr.QueueOptions{Page: 1, PageSize: 10, IncludeUnknown: true},
			wantPath: "/api/v1/queue?includeUnknownArtistItems=true&page=1&pageSize=10",
		},
		{
			name:     "includeParent alone",
			opts:     lidarr.QueueOptions{Page: 1, PageSize: 10, IncludeParent: true},
			wantPath: "/api/v1/queue?includeArtist=true&page=1&pageSize=10",
		},
		{
			name:     "includeItem alone",
			opts:     lidarr.QueueOptions{Page: 1, PageSize: 10, IncludeItem: true},
			wantPath: "/api/v1/queue?includeAlbum=true&page=1&pageSize=10",
		},
		{
			name: "all-true",
			opts: lidarr.QueueOptions{
				Page:           1,
				PageSize:       10,
				IncludeUnknown: true,
				IncludeParent:  true,
				IncludeItem:    true,
			},
			wantPath: "/api/v1/queue?includeAlbum=true&includeArtist=true&includeUnknownArtistItems=true&page=1&pageSize=10",
		},
		{
			name:     "boundary zero",
			opts:     lidarr.QueueOptions{Page: 0, PageSize: 0},
			wantPath: "/api/v1/queue?page=0&pageSize=0",
		},
		{
			name:     "boundary large",
			opts:     lidarr.QueueOptions{Page: 1000, PageSize: 500},
			wantPath: "/api/v1/queue?page=1000&pageSize=500",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.RequestURI() != tc.wantPath {
					t.Errorf("RequestURI = %q, want %q", r.URL.RequestURI(), tc.wantPath)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(arr.PagingResource[lidarr.QueueRecord]{
					Page:         tc.opts.Page,
					PageSize:     tc.opts.PageSize,
					TotalRecords: 0,
					Records:      []lidarr.QueueRecord{},
				})
			}))
			defer srv.Close()

			c, err := lidarr.New(srv.URL, "test-key")
			if err != nil {
				t.Fatalf("lidarr.New: %v", err)
			}

			res, err := c.GetQueueWithOptions(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("GetQueueWithOptions: %v", err)
			}
			if res == nil {
				t.Fatal("expected non-nil PagingResource")
			}
		})
	}
}

func TestGetQueueWithOptions_NestedMonitored(t *testing.T) {
	t.Parallel()

	respData := arr.PagingResource[lidarr.QueueRecord]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []lidarr.QueueRecord{
			{
				QueueRecord: arr.QueueRecord{
					ID:       1,
					Title:    "Pink Floyd - The Dark Side of the Moon",
					SizeLeft: 4096,
					Status:   "downloading",
				},
				CustomFormatScore: 110,
				ArtistID:          77,
				AlbumID:           88,
				Artist: &lidarr.Artist{
					ID:         77,
					ArtistName: "Pink Floyd",
					Monitored:  true,
				},
				Album: &lidarr.Album{
					ID:        88,
					Title:     "The Dark Side of the Moon",
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

	c, err := lidarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("lidarr.New: %v", err)
	}

	got, err := c.GetQueueWithOptions(context.Background(), lidarr.QueueOptions{
		Page:          1,
		PageSize:      10,
		IncludeParent: true,
		IncludeItem:   true,
	})
	if err != nil {
		t.Fatalf("GetQueueWithOptions: %v", err)
	}

	if len(got.Records) != 1 {
		t.Fatalf("got %d records, want 1", len(got.Records))
	}
	rec := got.Records[0]
	if rec.Artist == nil {
		t.Fatal("expected non-nil Artist")
	}
	if !rec.Artist.Monitored {
		t.Errorf("Artist.Monitored = false, want true")
	}
	if rec.Album == nil {
		t.Fatal("expected non-nil Album")
	}
	if !rec.Album.Monitored {
		t.Errorf("Album.Monitored = false, want true")
	}
}

func TestGetQueueWithOptions_APIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer srv.Close()

	c, err := lidarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("lidarr.New: %v", err)
	}

	_, err = c.GetQueueWithOptions(context.Background(), lidarr.QueueOptions{Page: 1, PageSize: 10})
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

func TestDeleteQueueItemWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     lidarr.QueueDeleteOptions
		wantPath string
	}{
		{
			name:     "all-false",
			opts:     lidarr.QueueDeleteOptions{},
			wantPath: "/api/v1/queue/42?removeFromClient=false&blocklist=false&skipRedownload=false&changeCategory=false",
		},
		{
			name: "skipRedownload alone",
			opts: lidarr.QueueDeleteOptions{
				SkipRedownload: true,
			},
			wantPath: "/api/v1/queue/42?removeFromClient=false&blocklist=false&skipRedownload=true&changeCategory=false",
		},
		{
			name: "all-true",
			opts: lidarr.QueueDeleteOptions{
				RemoveFromClient: true,
				Blocklist:        true,
				SkipRedownload:   true,
				ChangeCategory:   true,
			},
			wantPath: "/api/v1/queue/42?removeFromClient=true&blocklist=true&skipRedownload=true&changeCategory=true",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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

			c, err := lidarr.New(srv.URL, "test-key")
			if err != nil {
				t.Fatalf("lidarr.New: %v", err)
			}

			if err := c.DeleteQueueItemWithOptions(context.Background(), 42, tc.opts); err != nil {
				t.Fatalf("DeleteQueueItemWithOptions: %v", err)
			}
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

	c, err := lidarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("lidarr.New: %v", err)
	}

	err = c.DeleteQueueItemWithOptions(context.Background(), 7, lidarr.QueueDeleteOptions{
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

	c, err := lidarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("lidarr.New: %v", err)
	}

	err = c.DeleteQueueItemWithOptions(context.Background(), 1, lidarr.QueueDeleteOptions{})
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
