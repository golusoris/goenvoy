package sabnzbd_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/golusoris/goenvoy/downloadclient/sabnzbd/v2"
)

func newServer(t *testing.T, wantMode string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if got := r.URL.Query().Get("mode"); got != wantMode {
			t.Errorf("mode = %q, want %q", got, wantMode)
		}
		if got := r.URL.Query().Get("output"); got != "json" {
			t.Errorf("output = %q, want json", got)
		}
		if got := r.URL.Query().Get("apikey"); got != "test-key" {
			t.Errorf("apikey = %q, want test-key", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
}

func TestGetQueue(t *testing.T) {
	t.Parallel()

	result := map[string]any{
		"queue": map[string]any{
			"status": "Downloading", "speed": "5.0 M", "paused": false,
			"noofslots": 2,
			"slots": []map[string]any{
				{"nzo_id": "SABnzbd_nzo_abc", "filename": "Ubuntu.nzb", "status": "Downloading", "percentage": "75"},
				{"nzo_id": "SABnzbd_nzo_def", "filename": "Fedora.nzb", "status": "Queued", "percentage": "0"},
			},
		},
	}
	ts := newServer(t, "queue", result)
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	queue, err := c.GetQueue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if queue.Status != "Downloading" {
		t.Errorf("Status = %q, want %q", queue.Status, "Downloading")
	}
	if len(queue.Slots) != 2 {
		t.Fatalf("len(Slots) = %d, want 2", len(queue.Slots))
	}
	if queue.Slots[0].Filename != "Ubuntu.nzb" {
		t.Errorf("Filename = %q, want %q", queue.Slots[0].Filename, "Ubuntu.nzb")
	}
}

func TestAddURL(t *testing.T) {
	t.Parallel()

	result := map[string]any{
		"status":  true,
		"nzo_ids": []string{"SABnzbd_nzo_xyz"},
	}
	ts := newServer(t, "addurl", result)
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ids, err := c.AddURL(context.Background(), "https://example.com/file.nzb", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "SABnzbd_nzo_xyz" {
		t.Errorf("IDs = %v, want [SABnzbd_nzo_xyz]", ids)
	}
}

func TestPause(t *testing.T) {
	t.Parallel()

	ts := newServer(t, "pause", map[string]any{"status": "ok"})
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.Pause(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestResume(t *testing.T) {
	t.Parallel()

	ts := newServer(t, "resume", map[string]any{"status": "ok"})
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.Resume(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestGetHistory(t *testing.T) {
	t.Parallel()

	result := map[string]any{
		"history": map[string]any{
			"total_size": "10.5 GB", "noofslots": 1,
			"slots": []map[string]any{
				{
					"nzo_id": "SABnzbd_nzo_hist1", "name": "Completed Item",
					"status": "Completed", "size": "4.2 GB",
					"category": "movies", "storage": "/downloads/completed",
				},
			},
		},
	}
	ts := newServer(t, "history", result)
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	hist, err := c.GetHistory(context.Background(), 0, 50)
	if err != nil {
		t.Fatal(err)
	}
	if hist.TotalSize != "10.5 GB" {
		t.Errorf("TotalSize = %q, want %q", hist.TotalSize, "10.5 GB")
	}
	if len(hist.Slots) != 1 {
		t.Fatalf("len(Slots) = %d, want 1", len(hist.Slots))
	}
	if hist.Slots[0].Name != "Completed Item" {
		t.Errorf("Name = %q, want %q", hist.Slots[0].Name, "Completed Item")
	}
}

func TestGetVersion(t *testing.T) {
	t.Parallel()

	ts := newServer(t, "version", map[string]any{"version": "4.2.1"})
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	v, err := c.GetVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v != "4.2.1" {
		t.Errorf("version = %q, want %q", v, "4.2.1")
	}
}

func TestGetServerStats(t *testing.T) {
	t.Parallel()

	result := map[string]any{
		"total": 107374182400, "day": 5368709120, "week": 26843545600, "month": 85899345920,
	}
	ts := newServer(t, "server_stats", result)
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	stats, err := c.GetServerStats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 107374182400 {
		t.Errorf("Total = %d, want %d", stats.Total, 107374182400)
	}
}

func TestGetWarnings(t *testing.T) {
	t.Parallel()

	result := map[string]any{
		"warnings": []string{"disk space low", "connection timeout"},
	}
	ts := newServer(t, "warnings", result)
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	warnings, err := c.GetWarnings(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 2 {
		t.Fatalf("len = %d, want 2", len(warnings))
	}
	if warnings[0] != "disk space low" {
		t.Errorf("warning[0] = %q, want %q", warnings[0], "disk space low")
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	result := map[string]any{
		"status": false,
		"error":  "API Key Required",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "bad-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetQueue(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *sabnzbd.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Message != "API Key Required" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "API Key Required")
	}
}

func TestDeleteItem(t *testing.T) {
	t.Parallel()

	ts := newServer(t, "queue", map[string]any{"status": true})
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.DeleteItem(context.Background(), "SABnzbd_nzo_abc"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteHistory(t *testing.T) {
	t.Parallel()

	ts := newServer(t, "history", map[string]any{"status": true})
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.DeleteHistory(context.Background(), "SABnzbd_nzo_abc"); err != nil {
		t.Fatal(err)
	}
}

func TestSetCategory(t *testing.T) {
	t.Parallel()

	ts := newServer(t, "change_cat", map[string]any{"status": true})
	defer ts.Close()

	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.SetCategory(context.Background(), "SABnzbd_nzo_abc", "tv"); err != nil {
		t.Fatal(err)
	}
}

func TestNew_invalidURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, url string
	}{
		{"empty", ""},
		{"malformed", "://x"},
		{"ftp", "ftp://x"},
		{"no-scheme", "no-scheme"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := sabnzbd.New(tc.url, "k")
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}

func checkPriorityQuery(t *testing.T, r *http.Request, nzoID, wantValue2 string) {
	t.Helper()
	q := r.URL.Query()
	want := map[string]string{
		"mode": "queue", "name": "priority", "value": nzoID,
		"value2": wantValue2, "apikey": "test-key", "output": "json",
	}
	for key, value := range want {
		if got := q.Get(key); got != value {
			t.Errorf("%s = %q, want %q", key, got, value)
		}
	}
	if r.Method != http.MethodGet {
		t.Errorf("method = %q, want GET", r.Method)
	}
}

func newClient(t *testing.T, handler http.HandlerFunc) *sabnzbd.Client {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestSetPriority(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		nzoID    string
		priority sabnzbd.Priority
		wantVal2 string
		position int
	}{
		{"default boundary", "nzo-default", sabnzbd.PriorityDefault, "-100", 3},
		{"paused", "nzo-paused", sabnzbd.PriorityPaused, "-2", 4},
		{"low", "nzo-low", sabnzbd.PriorityLow, "-1", 2},
		{"normal boundary zero", "nzo-normal", sabnzbd.PriorityNormal, "0", 0},
		{"high", "nzo-high", sabnzbd.PriorityHigh, "1", 1},
		{"force boundary", "nzo-force", sabnzbd.PriorityForce, "2", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
				checkPriorityQuery(t, r, tc.nzoID, tc.wantVal2)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]int{"position": tc.position})
			})
			pos, err := c.SetPriority(context.Background(), tc.nzoID, tc.priority)
			if err != nil || pos != tc.position {
				t.Fatalf("SetPriority = %d, %v; want %d, nil", pos, err, tc.position)
			}
		})
	}
}

func TestSetPriority_RefusedInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		nzoID    string
		priority sabnzbd.Priority
		want     error
	}{
		{"empty id", "", sabnzbd.PriorityNormal, sabnzbd.ErrEmptyNZOID},
		{"above force", "nzo-123", 3, sabnzbd.ErrInvalidPriority},
		{"below default", "nzo-123", -101, sabnzbd.ErrInvalidPriority},
		{"gap below paused", "nzo-123", -3, sabnzbd.ErrInvalidPriority},
		{"far above", "nzo-123", 100, sabnzbd.ErrInvalidPriority},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var hits atomic.Int64
			c := newClient(t, func(_ http.ResponseWriter, _ *http.Request) { hits.Add(1) })
			_, err := c.SetPriority(context.Background(), tc.nzoID, tc.priority)
			if !errors.Is(err, tc.want) || hits.Load() != 0 {
				t.Fatalf("err = %v, hits = %d; want %v, 0", err, hits.Load(), tc.want)
			}
		})
	}
}

func TestSetPriority_ServerErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		handler http.HandlerFunc
		check   func(error) bool
	}{
		{"unknown job", func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]int{"position": -1})
		}, func(err error) bool { return errors.Is(err, sabnzbd.ErrUnknownJob) }},
		{"api error", func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"status": false, "error": "Cannot change priority"})
		}, func(err error) bool {
			var apiErr *sabnzbd.APIError
			return errors.As(err, &apiErr) && apiErr.Message == "Cannot change priority"
		}},
		{"http 500", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}, func(err error) bool { return err != nil }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pos, err := newClient(t, tc.handler).SetPriority(context.Background(), "nzo-123", sabnzbd.PriorityHigh)
			if !tc.check(err) || pos != -1 {
				t.Fatalf("SetPriority = %d, %v; want -1 and the expected error", pos, err)
			}
		})
	}
}
