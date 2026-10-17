package readarr_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/arr/readarr"
	"github.com/golusoris/goenvoy/arr/v2"
)

func TestQueueRecord_EmbeddingDecode(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": 401,
		"title": "J.R.R. Tolkien - The Hobbit",
		"status": "downloading",
		"sizeleft": 5242880.0,
		"customFormatScore": 75,
		"authorId": 11,
		"bookId": 22
	}`

	var rec readarr.QueueRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if rec.ID != 401 {
		t.Errorf("rec.ID = %d, want 401", rec.ID)
	}
	if rec.Status != "downloading" {
		t.Errorf("rec.Status = %q, want %q", rec.Status, "downloading")
	}
	if rec.SizeLeft != 5242880.0 {
		t.Errorf("rec.SizeLeft = %f, want 5242880.0", rec.SizeLeft)
	}
	if rec.CustomFormatScore != 75 {
		t.Errorf("rec.CustomFormatScore = %d, want 75", rec.CustomFormatScore)
	}
	if rec.AuthorID != 11 {
		t.Errorf("rec.AuthorID = %d, want 11", rec.AuthorID)
	}
	if rec.BookID != 22 {
		t.Errorf("rec.BookID = %d, want 22", rec.BookID)
	}
}

func TestGetQueueWithOptions_QueryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     readarr.QueueOptions
		wantPath string
	}{
		{
			name:     "all-false",
			opts:     readarr.QueueOptions{Page: 1, PageSize: 10},
			wantPath: "/api/v1/queue?page=1&pageSize=10",
		},
		{
			name:     "includeUnknown alone",
			opts:     readarr.QueueOptions{Page: 1, PageSize: 10, IncludeUnknown: true},
			wantPath: "/api/v1/queue?includeUnknownAuthorItems=true&page=1&pageSize=10",
		},
		{
			name:     "includeParent alone",
			opts:     readarr.QueueOptions{Page: 1, PageSize: 10, IncludeParent: true},
			wantPath: "/api/v1/queue?includeAuthor=true&page=1&pageSize=10",
		},
		{
			name:     "includeItem alone",
			opts:     readarr.QueueOptions{Page: 1, PageSize: 10, IncludeItem: true},
			wantPath: "/api/v1/queue?includeBook=true&page=1&pageSize=10",
		},
		{
			name: "all-true",
			opts: readarr.QueueOptions{
				Page:           1,
				PageSize:       10,
				IncludeUnknown: true,
				IncludeParent:  true,
				IncludeItem:    true,
			},
			wantPath: "/api/v1/queue?includeAuthor=true&includeBook=true&includeUnknownAuthorItems=true&page=1&pageSize=10",
		},
		{
			name:     "boundary zero",
			opts:     readarr.QueueOptions{Page: 0, PageSize: 0},
			wantPath: "/api/v1/queue?page=0&pageSize=0",
		},
		{
			name:     "boundary large",
			opts:     readarr.QueueOptions{Page: 1000, PageSize: 500},
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
				_ = json.NewEncoder(w).Encode(arr.PagingResource[readarr.QueueRecord]{
					Page:         tc.opts.Page,
					PageSize:     tc.opts.PageSize,
					TotalRecords: 0,
					Records:      []readarr.QueueRecord{},
				})
			}))
			defer srv.Close()

			c, err := readarr.New(srv.URL, "test-key")
			if err != nil {
				t.Fatalf("readarr.New: %v", err)
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

	respData := arr.PagingResource[readarr.QueueRecord]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []readarr.QueueRecord{
			{
				QueueRecord: arr.QueueRecord{
					ID:       1,
					Title:    "J.R.R. Tolkien - The Hobbit",
					SizeLeft: 8192,
					Status:   "downloading",
				},
				CustomFormatScore: 115,
				AuthorID:          11,
				BookID:            22,
				Author: &readarr.Author{
					ID:         11,
					AuthorName: "J.R.R. Tolkien",
					Monitored:  true,
				},
				Book: &readarr.Book{
					ID:        22,
					Title:     "The Hobbit",
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

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("readarr.New: %v", err)
	}

	got, err := c.GetQueueWithOptions(context.Background(), readarr.QueueOptions{
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
	if rec.Author == nil {
		t.Fatal("expected non-nil Author")
	}
	if !rec.Author.Monitored {
		t.Errorf("Author.Monitored = false, want true")
	}
	if rec.Book == nil {
		t.Fatal("expected non-nil Book")
	}
	if !rec.Book.Monitored {
		t.Errorf("Book.Monitored = false, want true")
	}
}

func TestGetQueueWithOptions_APIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("readarr.New: %v", err)
	}

	_, err = c.GetQueueWithOptions(context.Background(), readarr.QueueOptions{Page: 1, PageSize: 10})
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
		opts     readarr.QueueDeleteOptions
		wantPath string
	}{
		{
			name:     "all-false",
			opts:     readarr.QueueDeleteOptions{},
			wantPath: "/api/v1/queue/42?removeFromClient=false&blocklist=false&skipRedownload=false&changeCategory=false",
		},
		{
			name: "skipRedownload alone",
			opts: readarr.QueueDeleteOptions{
				SkipRedownload: true,
			},
			wantPath: "/api/v1/queue/42?removeFromClient=false&blocklist=false&skipRedownload=true&changeCategory=false",
		},
		{
			name: "all-true",
			opts: readarr.QueueDeleteOptions{
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

			c, err := readarr.New(srv.URL, "test-key")
			if err != nil {
				t.Fatalf("readarr.New: %v", err)
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

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("readarr.New: %v", err)
	}

	err = c.DeleteQueueItemWithOptions(context.Background(), 7, readarr.QueueDeleteOptions{
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

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("readarr.New: %v", err)
	}

	err = c.DeleteQueueItemWithOptions(context.Background(), 1, readarr.QueueDeleteOptions{})
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
