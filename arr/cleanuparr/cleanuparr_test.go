package cleanuparr_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/arr/cleanuparr"
)

func newMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func jsonOK(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		t.Fatalf("failed to encode mock response: %v", err)
	}
}

// TestClientNeverWrites verifies the read-only contract: every exported client
// method MUST execute only HTTP GET requests. Any non-GET method causes a test failure.
func TestClientNeverWrites(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("non-GET request received: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/status":
			fmt.Fprintln(w, `{"application":{"version":"2.10.11","startTime":"2026-09-20T08:00:00Z"},"mediaManagers":{}}`)
		case "/api/strikes/recent":
			fmt.Fprintln(w, `[]`)
		case "/api/strikes":
			fmt.Fprintln(w, `{"items":[],"page":1,"pageSize":500,"totalCount":0,"totalPages":0}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	client, err := cleanuparr.New(server.URL, "test-api-key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	ctx := context.Background()

	if _, err := client.GetStatus(ctx); err != nil {
		t.Errorf("GetStatus failed: %v", err)
	}
	if _, err := client.GetRecentStrikes(ctx, 10); err != nil {
		t.Errorf("GetRecentStrikes failed: %v", err)
	}
	if _, err := client.GetStrikes(ctx, 1, 50); err != nil {
		t.Errorf("GetStrikes failed: %v", err)
	}
	if _, err := client.SummarizeStrikes(ctx); err != nil {
		t.Errorf("SummarizeStrikes failed: %v", err)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		apiKey  string
		opts    []cleanuparr.Option
		wantErr bool
	}{
		{
			name:    "valid base url",
			baseURL: "http://127.0.0.1:11011",
			apiKey:  "valid-key",
			wantErr: false,
		},
		{
			name:    "valid with options",
			baseURL: "http://127.0.0.1:11011",
			apiKey:  "valid-key",
			opts: []cleanuparr.Option{
				cleanuparr.WithTimeout(5 * time.Second),
				cleanuparr.WithUserAgent("test-agent/1.0"),
				cleanuparr.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
			},
			wantErr: false,
		},
		{
			name:    "invalid base url",
			baseURL: "::invalid-url::",
			apiKey:  "valid-key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := cleanuparr.New(tt.baseURL, tt.apiKey, tt.opts...)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c == nil {
				t.Fatalf("expected client, got nil")
			}
		})
	}
}

func TestGetStatus_Success(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "clean-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/api/status" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		jsonOK(t, w, map[string]any{
			"application": map[string]any{
				"version":       "2.10.11.0",
				"startTime":     "2026-09-20T08:00:00.1234567+00:00",
				"upTime":        "1.02:03:04",
				"memoryUsageMB": 128.5,
				"processorTime": "00:01:23.456",
			},
			"mediaManagers": map[string]any{
				"Sonarr": map[string]any{"instanceCount": 2},
				"Radarr": map[string]any{"instanceCount": 1},
			},
		})
	})

	client, err := cleanuparr.New(server.URL, "clean-secret")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	st, err := client.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}

	if st.Application.Version != "2.10.11.0" {
		t.Errorf("expected app version 2.10.11.0, got %q", st.Application.Version)
	}
	if st.Application.StartTime != "2026-09-20T08:00:00.1234567+00:00" {
		t.Errorf("expected StartTime, got %q", st.Application.StartTime)
	}

	parsed, err := st.Application.StartTimeParsed()
	if err != nil {
		t.Errorf("StartTimeParsed failed: %v", err)
	}
	if parsed.IsZero() {
		t.Errorf("expected non-zero parsed time")
	}

	if len(st.MediaManagers) != 2 {
		t.Fatalf("expected 2 media managers, got %d", len(st.MediaManagers))
	}
	if st.MediaManagers["Sonarr"].InstanceCount != 2 {
		t.Errorf("expected Sonarr count 2, got %d", st.MediaManagers["Sonarr"].InstanceCount)
	}
	if st.MediaManagers["Radarr"].InstanceCount != 1 {
		t.Errorf("expected Radarr count 1, got %d", st.MediaManagers["Radarr"].InstanceCount)
	}
}

func TestGetStatus_InvalidTime(t *testing.T) {
	t.Parallel()

	app := cleanuparr.ApplicationStatus{
		StartTime: "invalid-timestamp",
	}
	if _, err := app.StartTimeParsed(); err == nil {
		t.Fatalf("expected error for invalid StartTime, got nil")
	}
}

func TestApplicationStatus_StartTimeParsed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		startTime string
		wantErr   bool
		checkUTC  bool
	}{
		{
			name:      "rfc3339nano with offset",
			startTime: "2026-09-20T08:00:00.1234567+00:00",
			wantErr:   false,
		},
		{
			name:      "rfc3339 with Z",
			startTime: "2026-09-20T08:00:00Z",
			wantErr:   false,
			checkUTC:  true,
		},
		{
			name:      "dotnet datetime local naive without offset",
			startTime: "2026-09-20T08:00:00.1234567",
			wantErr:   false,
			checkUTC:  true,
		},
		{
			name:      "garbage string",
			startTime: "invalid-timestamp-or-garbage",
			wantErr:   true,
		},
		{
			name:      "empty string",
			startTime: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := cleanuparr.ApplicationStatus{StartTime: tt.startTime}
			parsed, err := app.StartTimeParsed()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tt.startTime)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.startTime, err)
			}
			if parsed.IsZero() {
				t.Fatalf("expected non-zero time for %q", tt.startTime)
			}
			if parsed.Year() != 2026 || parsed.Month() != time.September || parsed.Day() != 20 {
				t.Errorf("expected date 2026-09-20, got %v", parsed)
			}
			if tt.checkUTC && parsed.Location() != time.UTC {
				t.Errorf("expected time in UTC, got %v", parsed.Location())
			}
		})
	}
}

func TestGetRecentStrikes_Success(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/strikes/recent" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		countParam := r.URL.Query().Get("count")
		if countParam != "2" {
			t.Errorf("expected count=2, got %s", countParam)
		}

		jsonOK(t, w, []map[string]any{
			{
				"id":         "strike-1",
				"type":       "Stalled",
				"createdAt":  "2026-09-21T10:00:00Z",
				"downloadId": "DL-101",
				"title":      "Test.Release.1",
				"isDryRun":   false,
			},
			{
				"id":         "strike-2",
				"type":       "SlowSpeed",
				"createdAt":  "2026-09-21T09:00:00Z",
				"downloadId": "DL-102",
				"title":      "Test.Release.2",
				"isDryRun":   true,
			},
		})
	})

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	strikes, err := client.GetRecentStrikes(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetRecentStrikes error: %v", err)
	}

	if len(strikes) != 2 {
		t.Fatalf("expected 2 strikes, got %d", len(strikes))
	}
	if strikes[0].ID != "strike-1" || strikes[0].Type != "Stalled" || strikes[0].DownloadID != "DL-101" {
		t.Errorf("strike 0 mismatch: %+v", strikes[0])
	}
	if strikes[0].IsDryRun {
		t.Errorf("expected isDryRun false")
	}
	if strikes[1].ID != "strike-2" || !strikes[1].IsDryRun {
		t.Errorf("strike 1 mismatch: %+v", strikes[1])
	}
	if strikes[0].CreatedAt.IsZero() || strikes[1].CreatedAt.IsZero() {
		t.Errorf("expected non-zero CreatedAt")
	}
}

func TestGetRecentStrikes_ClampBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		inputCount    int
		expectedCount string
	}{
		{
			name:          "count zero clamped to 1",
			inputCount:    0,
			expectedCount: "1",
		},
		{
			name:          "count negative clamped to 1",
			inputCount:    -5,
			expectedCount: "1",
		},
		{
			name:          "count 999 clamped to 50",
			inputCount:    999,
			expectedCount: "50",
		},
		{
			name:          "count within range unchanged",
			inputCount:    25,
			expectedCount: "25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
				actualCount := r.URL.Query().Get("count")
				if actualCount != tt.expectedCount {
					t.Errorf("expected count=%s, got %s", tt.expectedCount, actualCount)
				}
				jsonOK(t, w, []cleanuparr.Strike{})
			})

			client, err := cleanuparr.New(server.URL, "key")
			if err != nil {
				t.Fatalf("unexpected New error: %v", err)
			}

			_, err = client.GetRecentStrikes(context.Background(), tt.inputCount)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetStrikes_Success(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/strikes" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("pageSize") != "50" {
			t.Errorf("unexpected query params: %s", r.URL.RawQuery)
		}

		lastDownloaded := int64(1048576)
		jsonOK(t, w, map[string]any{
			"page":       1,
			"pageSize":   50,
			"totalCount": 1,
			"totalPages": 1,
			"items": []map[string]any{
				{
					"downloadItemId":     "item-uuid-1",
					"downloadId":         "dl-hash-abc",
					"title":              "Linux.ISO.2026",
					"totalStrikes":       2,
					"strikesByType":      map[string]int{"Stalled": 2},
					"latestStrikeAt":     "2026-09-21T12:00:00Z",
					"firstStrikeAt":      "2026-09-21T10:00:00Z",
					"isMarkedForRemoval": true,
					"isRemoved":          false,
					"isReturning":        false,
					"hasDryRunStrikes":   false,
					"strikes": []map[string]any{
						{
							"id":                  "strike-detail-1",
							"type":                "Stalled",
							"createdAt":           "2026-09-21T12:00:00Z",
							"lastDownloadedBytes": lastDownloaded,
							"jobRunId":            "job-uuid-1",
							"isDryRun":            false,
						},
					},
				},
			},
		})
	})

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	page, err := client.GetStrikes(context.Background(), 1, 50)
	if err != nil {
		t.Fatalf("GetStrikes error: %v", err)
	}

	if page.TotalCount != 1 || page.TotalPages != 1 || len(page.Items) != 1 {
		t.Fatalf("page summary mismatch: %+v", page)
	}

	item := page.Items[0]
	if item.DownloadItemID != "item-uuid-1" || item.DownloadID != "dl-hash-abc" || item.Title != "Linux.ISO.2026" {
		t.Errorf("item fields mismatch: %+v", item)
	}
	if !item.IsMarkedForRemoval || item.IsRemoved || item.IsReturning {
		t.Errorf("item status flags mismatch: %+v", item)
	}
	if len(item.Strikes) != 1 || item.Strikes[0].ID != "strike-detail-1" {
		t.Errorf("strikes details mismatch: %+v", item.Strikes)
	}
	if item.Strikes[0].LastDownloadedBytes == nil || *item.Strikes[0].LastDownloadedBytes != 1048576 {
		t.Errorf("last downloaded bytes mismatch: %v", item.Strikes[0].LastDownloadedBytes)
	}
}

func TestGetStrikes_Clamping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		page             int
		pageSize         int
		expectedPage     string
		expectedPageSize string
	}{
		{
			name:             "page < 1 clamps to 1",
			page:             0,
			pageSize:         50,
			expectedPage:     "1",
			expectedPageSize: "50",
		},
		{
			name:             "pageSize < 1 clamps to 50",
			page:             1,
			pageSize:         0,
			expectedPage:     "1",
			expectedPageSize: "50",
		},
		{
			name:             "pageSize > 500 clamps to 500",
			page:             2,
			pageSize:         1000,
			expectedPage:     "2",
			expectedPageSize: "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("page") != tt.expectedPage {
					t.Errorf("expected page %s, got %s", tt.expectedPage, r.URL.Query().Get("page"))
				}
				if r.URL.Query().Get("pageSize") != tt.expectedPageSize {
					t.Errorf("expected pageSize %s, got %s", tt.expectedPageSize, r.URL.Query().Get("pageSize"))
				}
				jsonOK(t, w, cleanuparr.StrikePage{})
			})

			client, err := cleanuparr.New(server.URL, "key")
			if err != nil {
				t.Fatalf("unexpected New error: %v", err)
			}

			_, err = client.GetStrikes(context.Background(), tt.page, tt.pageSize)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSummarizeStrikes_Success(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		pageSize := r.URL.Query().Get("pageSize")
		if pageSize != "500" {
			t.Errorf("expected pageSize=500, got %s", pageSize)
		}

		var items []cleanuparr.DownloadItemStrikes
		switch page {
		case "1":
			items = []cleanuparr.DownloadItemStrikes{
				{IsMarkedForRemoval: true, IsRemoved: false, IsReturning: false},
				{IsMarkedForRemoval: false, IsRemoved: true, IsReturning: false},
			}
		case "2":
			items = []cleanuparr.DownloadItemStrikes{
				{IsMarkedForRemoval: false, IsRemoved: false, IsReturning: true},
			}
		default:
			t.Errorf("unexpected page requested: %s", page)
		}

		jsonOK(t, w, cleanuparr.StrikePage{
			Items:      items,
			Page:       1,
			PageSize:   500,
			TotalCount: 3,
			TotalPages: 2,
		})
	})

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	summary, err := client.SummarizeStrikes(context.Background())
	if err != nil {
		t.Fatalf("SummarizeStrikes error: %v", err)
	}

	if summary.StruckDownloads != 3 {
		t.Errorf("expected StruckDownloads 3, got %d", summary.StruckDownloads)
	}
	if summary.MarkedForRemoval != 1 {
		t.Errorf("expected MarkedForRemoval 1, got %d", summary.MarkedForRemoval)
	}
	if summary.RemovedDownloads != 1 {
		t.Errorf("expected RemovedDownloads 1, got %d", summary.RemovedDownloads)
	}
	if summary.ReturningDownloads != 1 {
		t.Errorf("expected ReturningDownloads 1, got %d", summary.ReturningDownloads)
	}
	if summary.Partial {
		t.Errorf("expected Partial false, got true")
	}
}

func TestSummarizeStrikes_StopsAtTotalPages(t *testing.T) {
	t.Parallel()

	var reqCount atomic.Int32
	server := newMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		page := reqCount.Add(1)
		jsonOK(t, w, cleanuparr.StrikePage{
			Items:      []cleanuparr.DownloadItemStrikes{{IsMarkedForRemoval: true}},
			Page:       int(page),
			PageSize:   500,
			TotalCount: 1,
			TotalPages: 1, // TotalPages is 1, so summary MUST stop after page 1
		})
	})

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	sum, err := client.SummarizeStrikes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count := reqCount.Load(); count != 1 {
		t.Errorf("expected exactly 1 page requested, got %d", count)
	}
	if sum.MarkedForRemoval != 1 {
		t.Errorf("expected MarkedForRemoval=1, got %d", sum.MarkedForRemoval)
	}
	if sum.Partial {
		t.Errorf("expected Partial=false, got true")
	}
}

func TestSummarizeStrikes_StopsAtFourPageBoundEvenWhenTotalPages10(t *testing.T) {
	t.Parallel()

	var reqCount atomic.Int32
	server := newMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		page := reqCount.Add(1)
		jsonOK(t, w, cleanuparr.StrikePage{
			Items:      []cleanuparr.DownloadItemStrikes{{IsMarkedForRemoval: true}},
			Page:       int(page),
			PageSize:   500,
			TotalCount: 10,
			TotalPages: 10, // TotalPages is 10, but scalar bound is 4 pages!
		})
	})

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	sum, err := client.SummarizeStrikes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count := reqCount.Load(); count != 4 {
		t.Errorf("expected exactly 4 requests (HISS-02 scalar bound), got %d", count)
	}
	if sum.MarkedForRemoval != 4 {
		t.Errorf("expected 4 marked for removal across 4 fetched pages, got %d", sum.MarkedForRemoval)
	}
	if !sum.Partial {
		t.Errorf("expected Partial=true, got false")
	}
}

func TestSummarizeStrikes_EmptyItems(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		jsonOK(t, w, cleanuparr.StrikePage{
			Items:      []cleanuparr.DownloadItemStrikes{},
			Page:       1,
			PageSize:   500,
			TotalCount: 0,
			TotalPages: 0,
		})
	})

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	sum, err := client.SummarizeStrikes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sum.StruckDownloads != 0 || sum.MarkedForRemoval != 0 || sum.RemovedDownloads != 0 || sum.ReturningDownloads != 0 || sum.Partial {
		t.Errorf("expected all zero summary, got %+v", sum)
	}
}

func TestNegative_APIError500(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
	})

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	ctx := context.Background()

	// Test GetStatus 500
	_, err = client.GetStatus(ctx)
	if err == nil {
		t.Fatalf("expected error from GetStatus 500, got nil")
	}
	var apiErr *cleanuparr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected errors.As *cleanuparr.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code 500, got %d", apiErr.StatusCode)
	}
	if !strings.HasPrefix(err.Error(), "cleanuparr: get status:") {
		t.Errorf("expected prefix 'cleanuparr: get status:', got %q", err.Error())
	}

	// Test GetRecentStrikes 500
	_, err = client.GetRecentStrikes(ctx, 5)
	if err == nil || !errors.As(err, &apiErr) {
		t.Fatalf("expected *cleanuparr.APIError from GetRecentStrikes, got %v", err)
	}

	// Test GetStrikes 500
	_, err = client.GetStrikes(ctx, 1, 50)
	if err == nil || !errors.As(err, &apiErr) {
		t.Fatalf("expected *cleanuparr.APIError from GetStrikes, got %v", err)
	}

	// Test SummarizeStrikes 500
	_, err = client.SummarizeStrikes(ctx)
	if err == nil || !errors.As(err, &apiErr) {
		t.Fatalf("expected *cleanuparr.APIError from SummarizeStrikes, got %v", err)
	}
}

func TestNegative_MalformedJSON(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"invalid-json-content:`)
	})

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	ctx := context.Background()

	if _, err := client.GetStatus(ctx); err == nil {
		t.Errorf("expected error for malformed JSON in GetStatus, got nil")
	}
	if _, err := client.GetRecentStrikes(ctx, 5); err == nil {
		t.Errorf("expected error for malformed JSON in GetRecentStrikes, got nil")
	}
	if _, err := client.GetStrikes(ctx, 1, 50); err == nil {
		t.Errorf("expected error for malformed JSON in GetStrikes, got nil")
	}
	if _, err := client.SummarizeStrikes(ctx); err == nil {
		t.Errorf("expected error for malformed JSON in SummarizeStrikes, got nil")
	}
}

func TestNegative_NoApiKeyLeakInErrors(t *testing.T) {
	t.Parallel()

	secretKey := "super-secret-cleanuparr-key-999"
	server := newMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "server boom", http.StatusInternalServerError)
	})

	client, err := cleanuparr.New(server.URL, secretKey)
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	_, err = client.GetStatus(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if strings.Contains(err.Error(), secretKey) {
		t.Fatalf("security violation: secret API key found in error message: %q", err.Error())
	}
}
