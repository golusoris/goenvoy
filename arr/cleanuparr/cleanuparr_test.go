package cleanuparr_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
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

// readOnlyHandler serves a minimal valid body for every read endpoint and
// fails the test on any non-GET request or unknown path.
func readOnlyHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	bodies := map[string]string{
		"/api/status":                `{"application":{"version":"2.10.11","startTime":"2026-09-20T08:00:00Z"},"mediaManagers":{}}`,
		"/api/strikes/recent":        `[]`,
		"/api/strikes":               `{"items":[],"page":1,"pageSize":500,"totalCount":0,"totalPages":0}`,
		"/api/configuration/seeker":  `{"searchEnabled":false,"searchInterval":60,"proactiveSearchEnabled":false,"selectionStrategy":"BalancedWeighted","useRoundRobin":false,"postReleaseGraceHours":0,"instances":[]}`,
		"/api/configuration/general": `{"displaySupportBanner":false,"dryRun":false,"httpMaxRetries":3,"httpTimeout":30,"httpCertificateValidation":"Enabled","httpSendUserAgent":true,"statusCheckEnabled":false,"ignoredDownloads":[],"connectivityCheckEnabled":false,"connectivityCheckUrls":[],"strikeInactivityWindowHours":24,"historyRetentionDays":7,"log":{"level":"info","rollingSizeMB":10,"retainedFileCount":5,"timeLimitHours":24,"archiveEnabled":false,"archiveRetainedCount":0,"archiveTimeLimitHours":0},"auth":{"disableAuthForLocalAddresses":false,"trustForwardedHeaders":false,"trustedNetworks":[]}}`,
		"/api/configuration/sonarr":  `{"id":"00000000-0000-0000-0000-000000000001","type":"Sonarr","failedImportMaxStrikes":3,"instances":[]}`,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("non-GET request received: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, ok := bodies[r.URL.Path]
		if !ok {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	}
}

// TestClientNeverWrites verifies the read-only contract: every exported client
// method MUST execute only HTTP GET requests. Any non-GET method causes a test failure.
func TestClientNeverWrites(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, readOnlyHandler(t))

	client, err := cleanuparr.New(server.URL, "test-api-key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	calls := []struct {
		name string
		call func(context.Context) error
	}{
		{"GetStatus", func(ctx context.Context) error { _, err := client.GetStatus(ctx); return err }},
		{"GetRecentStrikes", func(ctx context.Context) error { _, err := client.GetRecentStrikes(ctx, 10); return err }},
		{"GetStrikes", func(ctx context.Context) error { _, err := client.GetStrikes(ctx, 1, 50); return err }},
		{"SummarizeStrikes", func(ctx context.Context) error { _, err := client.SummarizeStrikes(ctx); return err }},
		{"GetSeekerConfig", func(ctx context.Context) error { _, err := client.GetSeekerConfig(ctx); return err }},
		{"GetGeneralConfig", func(ctx context.Context) error { _, err := client.GetGeneralConfig(ctx); return err }},
		{"GetArrConfig", func(ctx context.Context) error {
			_, err := client.GetArrConfig(ctx, cleanuparr.ArrKindSonarr)
			return err
		}},
	}

	ctx := context.Background()
	for _, c := range calls {
		if err := c.call(ctx); err != nil {
			t.Errorf("%s failed: %v", c.name, err)
		}
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

// statusHandler serves an authenticated /api/status fixture.
func statusHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func TestGetStatus_Success(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, statusHandler(t))

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

// startTimeCase is one StartTimeParsed input and its expected outcome.
type startTimeCase struct {
	name      string
	startTime string
	wantErr   bool
	checkUTC  bool
}

// assertStartTimeParsed checks StartTimeParsed against one startTimeCase.
func assertStartTimeParsed(t *testing.T, tc startTimeCase) {
	t.Helper()
	parsed, err := cleanuparr.ApplicationStatus{StartTime: tc.startTime}.StartTimeParsed()
	if tc.wantErr {
		if err == nil {
			t.Fatalf("expected error for input %q, got nil", tc.startTime)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error for input %q: %v", tc.startTime, err)
	}
	if parsed.Year() != 2026 || parsed.Month() != time.September || parsed.Day() != 20 {
		t.Errorf("expected date 2026-09-20, got %v", parsed)
	}
	if tc.checkUTC && parsed.Location() != time.UTC {
		t.Errorf("expected time in UTC, got %v", parsed.Location())
	}
}

func TestApplicationStatus_StartTimeParsed(t *testing.T) {
	t.Parallel()

	tests := []startTimeCase{
		{name: "rfc3339nano with offset", startTime: "2026-09-20T08:00:00.1234567+00:00"},
		{name: "rfc3339 with Z", startTime: "2026-09-20T08:00:00Z", checkUTC: true},
		{name: "dotnet datetime local naive without offset", startTime: "2026-09-20T08:00:00.1234567", checkUTC: true},
		{name: "garbage string", startTime: "invalid-timestamp-or-garbage", wantErr: true},
		{name: "empty string", startTime: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertStartTimeParsed(t, tt)
		})
	}
}

// recentStrikesHandler serves two recent strikes and checks the count query.
func recentStrikesHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func TestGetRecentStrikes_Success(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, recentStrikesHandler(t))

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	strikes, err := client.GetRecentStrikes(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetRecentStrikes error: %v", err)
	}

	want := []cleanuparr.Strike{
		{
			ID:         "strike-1",
			Type:       "Stalled",
			CreatedAt:  time.Date(2026, time.September, 21, 10, 0, 0, 0, time.UTC),
			DownloadID: "DL-101",
			Title:      "Test.Release.1",
		},
		{
			ID:         "strike-2",
			Type:       "SlowSpeed",
			CreatedAt:  time.Date(2026, time.September, 21, 9, 0, 0, 0, time.UTC),
			DownloadID: "DL-102",
			Title:      "Test.Release.2",
			IsDryRun:   true,
		},
	}
	if !reflect.DeepEqual(strikes, want) {
		t.Errorf("strikes mismatch:\n got %+v\nwant %+v", strikes, want)
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

// strikesPageHandler serves one strikes page and checks the paging query.
func strikesPageHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func TestGetStrikes_Success(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, strikesPageHandler(t))

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	page, err := client.GetStrikes(context.Background(), 1, 50)
	if err != nil {
		t.Fatalf("GetStrikes error: %v", err)
	}

	lastDownloaded := int64(1048576)
	want := &cleanuparr.StrikePage{
		Page:       1,
		PageSize:   50,
		TotalCount: 1,
		TotalPages: 1,
		Items: []cleanuparr.DownloadItemStrikes{
			{
				DownloadItemID:     "item-uuid-1",
				DownloadID:         "dl-hash-abc",
				Title:              "Linux.ISO.2026",
				TotalStrikes:       2,
				StrikesByType:      map[string]int{"Stalled": 2},
				LatestStrikeAt:     time.Date(2026, time.September, 21, 12, 0, 0, 0, time.UTC),
				FirstStrikeAt:      time.Date(2026, time.September, 21, 10, 0, 0, 0, time.UTC),
				IsMarkedForRemoval: true,
				Strikes: []cleanuparr.StrikeDetail{
					{
						ID:                  "strike-detail-1",
						Type:                "Stalled",
						CreatedAt:           time.Date(2026, time.September, 21, 12, 0, 0, 0, time.UTC),
						LastDownloadedBytes: &lastDownloaded,
						JobRunID:            "job-uuid-1",
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(page, want) {
		t.Errorf("strike page mismatch:\n got %+v\nwant %+v", page, want)
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

// summaryPagesHandler serves a two-page strike listing for SummarizeStrikes.
func summaryPagesHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func TestSummarizeStrikes_Success(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, summaryPagesHandler(t))

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

// assertAPIError500 checks err is a 500 *cleanuparr.APIError whose message starts with prefix.
func assertAPIError500(t *testing.T, err error, prefix string) {
	t.Helper()
	var apiErr *cleanuparr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected errors.As *cleanuparr.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code 500, got %d", apiErr.StatusCode)
	}
	if !strings.HasPrefix(err.Error(), prefix) {
		t.Errorf("expected prefix %q, got %q", prefix, err.Error())
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

	calls := []struct {
		prefix string
		call   func(context.Context) error
	}{
		{"cleanuparr: get status:", func(ctx context.Context) error { _, err := client.GetStatus(ctx); return err }},
		{"cleanuparr: get recent strikes:", func(ctx context.Context) error { _, err := client.GetRecentStrikes(ctx, 5); return err }},
		{"cleanuparr: get strikes:", func(ctx context.Context) error { _, err := client.GetStrikes(ctx, 1, 50); return err }},
		{"cleanuparr: summarize strikes:", func(ctx context.Context) error { _, err := client.SummarizeStrikes(ctx); return err }},
	}

	ctx := context.Background()
	for _, c := range calls {
		assertAPIError500(t, c.call(ctx), c.prefix)
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
	if _, err := client.GetSeekerConfig(ctx); err == nil {
		t.Errorf("expected error for malformed JSON in GetSeekerConfig, got nil")
	}
	if _, err := client.GetGeneralConfig(ctx); err == nil {
		t.Errorf("expected error for malformed JSON in GetGeneralConfig, got nil")
	}
	if _, err := client.GetArrConfig(ctx, cleanuparr.ArrKindSonarr); err == nil {
		t.Errorf("expected error for malformed JSON in GetArrConfig, got nil")
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
	_, err = client.GetSeekerConfig(context.Background())
	if err != nil && strings.Contains(err.Error(), secretKey) {
		t.Fatalf("security violation: secret API key found in error message: %q", err.Error())
	}
	_, err = client.GetGeneralConfig(context.Background())
	if err != nil && strings.Contains(err.Error(), secretKey) {
		t.Fatalf("security violation: secret API key found in error message: %q", err.Error())
	}
	_, err = client.GetArrConfig(context.Background(), cleanuparr.ArrKindSonarr)
	if err != nil && strings.Contains(err.Error(), secretKey) {
		t.Fatalf("security violation: secret API key found in error message: %q", err.Error())
	}
}

// configHandler serves body at path to GET requests that carry the clean-secret API key.
func configHandler(t *testing.T, path, body string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-Api-Key") != "clean-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	}
}

func TestGetSeekerConfig_Success(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, configHandler(t, "/api/configuration/seeker", `{
		"searchEnabled": true,
		"searchInterval": 15,
		"proactiveSearchEnabled": true,
		"selectionStrategy": "BalancedWeighted",
		"useRoundRobin": true,
		"postReleaseGraceHours": 4,
		"instances": [
			{
				"arrInstanceId": "11111111-1111-1111-1111-111111111111",
				"instanceName": "sonarr-main",
				"instanceType": "Sonarr",
				"enabled": true,
				"skipTags": ["archive", "slow"],
				"lastProcessedAt": "2026-09-20T08:00:00Z",
				"arrInstanceEnabled": true,
				"activeDownloadLimit": 10,
				"ignoreStruckDownloads": true,
				"minCycleTimeDays": 2,
				"monitoredOnly": true,
				"useCutoff": true,
				"useCustomFormatScore": false
			},
			{
				"arrInstanceId": "22222222-2222-2222-2222-222222222222",
				"instanceName": "radarr-main",
				"instanceType": "Radarr",
				"enabled": false,
				"skipTags": [],
				"lastProcessedAt": null,
				"arrInstanceEnabled": false,
				"activeDownloadLimit": 5,
				"ignoreStruckDownloads": false,
				"minCycleTimeDays": 1,
				"monitoredOnly": false,
				"useCutoff": false,
				"useCustomFormatScore": true
			}
		]
	}`))

	client, err := cleanuparr.New(server.URL, "clean-secret")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	cfg, err := client.GetSeekerConfig(context.Background())
	if err != nil {
		t.Fatalf("GetSeekerConfig error: %v", err)
	}

	lastProcessed := time.Date(2026, time.September, 20, 8, 0, 0, 0, time.UTC)
	want := &cleanuparr.SeekerConfig{
		SearchEnabled:          true,
		SearchInterval:         15,
		ProactiveSearchEnabled: true,
		SelectionStrategy:      "BalancedWeighted",
		UseRoundRobin:          true,
		PostReleaseGraceHours:  4,
		Instances: []cleanuparr.SeekerInstance{
			{
				ArrInstanceID:         "11111111-1111-1111-1111-111111111111",
				InstanceName:          "sonarr-main",
				InstanceType:          "Sonarr",
				Enabled:               true,
				SkipTags:              []string{"archive", "slow"},
				LastProcessedAt:       &lastProcessed,
				ArrInstanceEnabled:    true,
				ActiveDownloadLimit:   10,
				IgnoreStruckDownloads: true,
				MinCycleTimeDays:      2,
				MonitoredOnly:         true,
				UseCutoff:             true,
			},
			{
				// Why: a null lastProcessedAt must decode to a nil pointer.
				ArrInstanceID:        "22222222-2222-2222-2222-222222222222",
				InstanceName:         "radarr-main",
				InstanceType:         "Radarr",
				SkipTags:             []string{},
				ActiveDownloadLimit:  5,
				MinCycleTimeDays:     1,
				UseCustomFormatScore: true,
			},
		},
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("seeker config mismatch:\n got %+v\nwant %+v", cfg, want)
	}
}

func TestGetGeneralConfig_Success(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, configHandler(t, "/api/configuration/general", `{
		"displaySupportBanner": true,
		"dryRun": true,
		"httpMaxRetries": 5,
		"httpTimeout": 60,
		"httpCertificateValidation": "Enabled",
		"httpSendUserAgent": true,
		"statusCheckEnabled": true,
		"ignoredDownloads": ["dl-1", "dl-2"],
		"connectivityCheckEnabled": true,
		"connectivityCheckUrls": ["http://connectivity.test/check"],
		"strikeInactivityWindowHours": 48,
		"historyRetentionDays": 30,
		"log": {
			"level": "debug",
			"rollingSizeMB": 20,
			"retainedFileCount": 10,
			"timeLimitHours": 72,
			"archiveEnabled": true,
			"archiveRetainedCount": 3,
			"archiveTimeLimitHours": 168
		},
		"auth": {
			"disableAuthForLocalAddresses": true,
			"trustForwardedHeaders": true,
			"trustedNetworks": ["10.0.0.0/8", "192.168.1.0/24"]
		}
	}`))

	client, err := cleanuparr.New(server.URL, "clean-secret")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	cfg, err := client.GetGeneralConfig(context.Background())
	if err != nil {
		t.Fatalf("GetGeneralConfig error: %v", err)
	}

	want := &cleanuparr.GeneralConfig{
		DisplaySupportBanner:        true,
		DryRun:                      true,
		HTTPMaxRetries:              5,
		HTTPTimeout:                 60,
		HTTPCertificateValidation:   "Enabled",
		HTTPSendUserAgent:           true,
		StatusCheckEnabled:          true,
		IgnoredDownloads:            []string{"dl-1", "dl-2"},
		ConnectivityCheckEnabled:    true,
		ConnectivityCheckURLs:       []string{"http://connectivity.test/check"},
		StrikeInactivityWindowHours: 48,
		HistoryRetentionDays:        30,
		Log: cleanuparr.GeneralLogConfig{
			Level:                 "debug",
			RollingSizeMB:         20,
			RetainedFileCount:     10,
			TimeLimitHours:        72,
			ArchiveEnabled:        true,
			ArchiveRetainedCount:  3,
			ArchiveTimeLimitHours: 168,
		},
		Auth: cleanuparr.GeneralAuthConfig{
			DisableAuthForLocalAddresses: true,
			TrustForwardedHeaders:        true,
			TrustedNetworks:              []string{"10.0.0.0/8", "192.168.1.0/24"},
		},
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("general config mismatch:\n got %+v\nwant %+v", cfg, want)
	}
}

func TestGetArrConfig_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind       cleanuparr.ArrKind
		targetPath string
		serverType string
		instName   string
		instURL    string
	}{
		{
			kind:       cleanuparr.ArrKindSonarr,
			targetPath: "/api/configuration/sonarr",
			serverType: "Sonarr",
			instName:   "sonarr-main",
			instURL:    "http://sonarr.test:8989",
		},
		{
			kind:       cleanuparr.ArrKindRadarr,
			targetPath: "/api/configuration/radarr",
			serverType: "Radarr",
			instName:   "radarr-main",
			instURL:    "http://radarr.test:7878",
		},
		{
			kind:       cleanuparr.ArrKindLazyLibrarian,
			targetPath: "/api/configuration/lazylibrarian",
			serverType: "LazyLibrarian",
			instName:   "lazylibrarian-main",
			instURL:    "http://lazylibrarian.test:5299",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			t.Parallel()

			secretAPIKeyToLeak := "fixture-key-must-not-decode"
			body := fmt.Sprintf(`{
			"id": "33333333-3333-3333-3333-333333333333",
			"type": "%s",
			"failedImportMaxStrikes": 3,
			"instances": [
				{
					"id": "44444444-4444-4444-4444-444444444444",
					"enabled": true,
					"version": 3.0,
					"name": "%s",
					"url": "%s",
					"externalUrl": "http://sonarr.public:8989",
					"apiKey": "%s"
				}
			]
		}`, tt.serverType, tt.instName, tt.instURL, secretAPIKeyToLeak)
			server := newMockServer(t, configHandler(t, tt.targetPath, body))

			client, err := cleanuparr.New(server.URL, "clean-secret")
			if err != nil {
				t.Fatalf("unexpected New error: %v", err)
			}

			cfg, err := client.GetArrConfig(context.Background(), tt.kind)
			if err != nil {
				t.Fatalf("GetArrConfig error: %v", err)
			}

			externalURL := "http://sonarr.public:8989"
			want := &cleanuparr.ArrConfig{
				ID:                     "33333333-3333-3333-3333-333333333333",
				Type:                   tt.serverType,
				FailedImportMaxStrikes: 3,
				Instances: []cleanuparr.ArrInstance{
					{
						ID:          "44444444-4444-4444-4444-444444444444",
						Enabled:     true,
						Version:     3.0,
						Name:        tt.instName,
						URL:         tt.instURL,
						ExternalURL: &externalURL,
					},
				},
			}
			if !reflect.DeepEqual(cfg, want) {
				t.Errorf("arr config mismatch:\n got %+v\nwant %+v", cfg, want)
			}

			// Security: apiKey must not decode into struct anywhere
			marshaled, err := json.Marshal(cfg)
			if err != nil {
				t.Fatalf("json.Marshal failed: %v", err)
			}
			if strings.Contains(string(marshaled), secretAPIKeyToLeak) {
				t.Fatalf("security violation: secret apiKey found in decoded struct JSON: %s", string(marshaled))
			}
		})
	}
}

func TestGetArrConfig_UnknownKind(t *testing.T) {
	t.Parallel()

	var handlerHit atomic.Bool
	server := newMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		handlerHit.Store(true)
		w.WriteHeader(http.StatusOK)
	})

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	invalidKinds := []cleanuparr.ArrKind{
		"prowlarr",
		"../x",
		"invalid",
		"",
	}

	for _, k := range invalidKinds {
		_, err := client.GetArrConfig(context.Background(), k)
		if err == nil {
			t.Errorf("expected error for invalid kind %q, got nil", k)
		}
	}

	if handlerHit.Load() {
		t.Errorf("server handler was hit for invalid kinds, expected no request sent")
	}
}

func TestNegative_Config_APIError500(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
	})

	client, err := cleanuparr.New(server.URL, "key")
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	calls := []struct {
		prefix string
		call   func(context.Context) error
	}{
		{"cleanuparr: get seeker config:", func(ctx context.Context) error { _, err := client.GetSeekerConfig(ctx); return err }},
		{"cleanuparr: get general config:", func(ctx context.Context) error { _, err := client.GetGeneralConfig(ctx); return err }},
		{"cleanuparr: get arr config:", func(ctx context.Context) error {
			_, err := client.GetArrConfig(ctx, cleanuparr.ArrKindSonarr)
			return err
		}},
	}

	ctx := context.Background()
	for _, c := range calls {
		assertAPIError500(t, c.call(ctx), c.prefix)
	}
}
