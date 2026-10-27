package deluge_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/golusoris/goenvoy/downloadclient/deluge/v2"
)

type rpcRequest struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

func newRPCServer(t *testing.T, wantMethod string, result any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("HTTP method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/json" {
			t.Errorf("path = %q, want /json", r.URL.Path)
		}
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != wantMethod {
			t.Errorf("RPC method = %q, want %q", req.Method, wantMethod)
		}
		w.Header().Set("Content-Type", "application/json")
		resultJSON, _ := json.Marshal(result)
		resp := map[string]any{
			"id":     req.ID,
			"result": json.RawMessage(resultJSON),
			"error":  nil,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func loginServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{"id": req.ID, "result": true, "error": nil}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestLogin(t *testing.T) {
	t.Parallel()

	ts := loginServer(t)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.Login(context.Background(), "deluge"); err != nil {
		t.Fatal(err)
	}
}

func TestLoginFailed(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{"id": req.ID, "result": false, "error": nil}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = c.Login(context.Background(), "wrong")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *deluge.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
}

func TestGetTorrentsStatus(t *testing.T) {
	t.Parallel()

	result := map[string]any{
		"abc123": map[string]any{
			"hash": "abc123", "name": "Ubuntu", "state": "Seeding",
			"progress": 100.0, "total_size": 4000000000,
		},
		"def456": map[string]any{
			"hash": "def456", "name": "Fedora", "state": "Downloading",
			"progress": 50.0, "total_size": 2000000000,
		},
	}
	ts := newRPCServer(t, "core.get_torrents_status", result)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	torrents, err := c.GetTorrentsStatus(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(torrents) != 2 {
		t.Fatalf("len = %d, want 2", len(torrents))
	}
	if torrents["abc123"].Name != "Ubuntu" {
		t.Errorf("Name = %q, want %q", torrents["abc123"].Name, "Ubuntu")
	}
}

func TestGetTorrentStatus(t *testing.T) {
	t.Parallel()

	result := map[string]any{
		"hash": "abc123", "name": "Ubuntu", "state": "Seeding", "progress": 100.0,
	}
	ts := newRPCServer(t, "core.get_torrent_status", result)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	torrent, err := c.GetTorrentStatus(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if torrent.State != "Seeding" {
		t.Errorf("State = %q, want %q", torrent.State, "Seeding")
	}
}

func TestAddTorrentURL(t *testing.T) {
	t.Parallel()

	ts := newRPCServer(t, "core.add_torrent_url", "abc123def456")
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	hash, err := c.AddTorrentURL(context.Background(), "magnet:?xt=urn:btih:abc123", nil)
	if err != nil {
		t.Fatal(err)
	}
	if hash != "abc123def456" {
		t.Errorf("hash = %q, want %q", hash, "abc123def456")
	}
}

func TestRemoveTorrent(t *testing.T) {
	t.Parallel()

	ts := newRPCServer(t, "core.remove_torrent", true)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.RemoveTorrent(context.Background(), "abc123", true); err != nil {
		t.Fatal(err)
	}
}

func TestPauseTorrent(t *testing.T) {
	t.Parallel()

	ts := newRPCServer(t, "core.pause_torrent", nil)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.PauseTorrent(context.Background(), "abc123"); err != nil {
		t.Fatal(err)
	}
}

func TestResumeTorrent(t *testing.T) {
	t.Parallel()

	ts := newRPCServer(t, "core.resume_torrent", nil)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.ResumeTorrent(context.Background(), "def456"); err != nil {
		t.Fatal(err)
	}
}

func TestForceRecheck(t *testing.T) {
	t.Parallel()

	ts := newRPCServer(t, "core.force_recheck", nil)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.ForceRecheck(context.Background(), []string{"abc123"}); err != nil {
		t.Fatal(err)
	}
}

type rpcMethodResponse struct {
	result any
	err    *deluge.APIError
}

type versionTestCase struct {
	name           string
	responses      map[string]rpcMethodResponse
	wantVersion    string
	wantErr        bool
	wantErrCode    int
	wantInfoCalled bool
}

func newFakeRPCServer(t *testing.T, responses map[string]rpcMethodResponse, calls map[string]int, mu *sync.Mutex) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		calls[req.Method]++
		mu.Unlock()

		respDef, ok := responses[req.Method]
		if !ok {
			t.Errorf("unexpected RPC method: %s", req.Method)
			http.Error(w, "unexpected method", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if respDef.err != nil {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     req.ID,
				"result": nil,
				"error": map[string]any{
					"message": respDef.err.Message,
					"code":    respDef.err.Code,
				},
			})
			return
		}

		resultJSON, err := json.Marshal(respDef.result)
		if err != nil {
			t.Errorf("marshal result: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     req.ID,
			"result": json.RawMessage(resultJSON),
			"error":  nil,
		})
	}))
}

func assertVersionSuccess(t *testing.T, got, want string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("version = %q, want %q", got, want)
	}
}

func assertVersionError(t *testing.T, err error, wantCode int) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if wantCode == 0 {
		return
	}
	var apiErr *deluge.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *deluge.APIError, got %T: %v", err, err)
	}
	if apiErr.Code != wantCode {
		t.Errorf("apiErr.Code = %d, want %d", apiErr.Code, wantCode)
	}
}

func assertVersionCallCount(t *testing.T, calls map[string]int, mu *sync.Mutex, wantInfoCalled bool) {
	t.Helper()
	mu.Lock()
	count := calls["daemon.info"]
	mu.Unlock()

	if wantInfoCalled && count == 0 {
		t.Error("daemon.info was not called, want called")
	}
	if !wantInfoCalled && count > 0 {
		t.Errorf("daemon.info was called %d times, want 0", count)
	}
}

func runGetVersionTestCase(t *testing.T, tc versionTestCase) {
	t.Helper()
	var (
		mu    sync.Mutex
		calls = make(map[string]int)
	)

	ts := newFakeRPCServer(t, tc.responses, calls, &mu)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	v, err := c.GetVersion(context.Background())
	if tc.wantErr {
		assertVersionError(t, err, tc.wantErrCode)
	} else {
		assertVersionSuccess(t, v, tc.wantVersion, err)
	}

	assertVersionCallCount(t, calls, &mu, tc.wantInfoCalled)
}

func TestGetVersion(t *testing.T) {
	t.Parallel()

	tests := []versionTestCase{
		{
			name: "deluge 2 daemon",
			responses: map[string]rpcMethodResponse{
				"daemon.get_version": {result: "2.2.0"},
				"daemon.info":        {err: &deluge.APIError{Code: 2, Message: "Unknown method"}},
			},
			wantVersion:    "2.2.0",
			wantErr:        false,
			wantInfoCalled: false,
		},
		{
			name: "deluge 1 daemon",
			responses: map[string]rpcMethodResponse{
				"daemon.get_version": {err: &deluge.APIError{Code: 2, Message: "Unknown method"}},
				"daemon.info":        {result: "1.3.15"},
			},
			wantVersion:    "1.3.15",
			wantErr:        false,
			wantInfoCalled: true,
		},
		{
			name: "negative non-2 error",
			responses: map[string]rpcMethodResponse{
				"daemon.get_version": {err: &deluge.APIError{Code: 1, Message: "Not authenticated"}},
				"daemon.info":        {result: "1.3.15"},
			},
			wantErr:        true,
			wantErrCode:    1,
			wantInfoCalled: false,
		},
		{
			name: "negative both methods unknown",
			responses: map[string]rpcMethodResponse{
				"daemon.get_version": {err: &deluge.APIError{Code: 2, Message: "Unknown method"}},
				"daemon.info":        {err: &deluge.APIError{Code: 2, Message: "Unknown method"}},
			},
			wantErr:        true,
			wantErrCode:    2,
			wantInfoCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runGetVersionTestCase(t, tc)
		})
	}
}

func TestGetSessionStatus(t *testing.T) {
	t.Parallel()

	result := map[string]any{
		"payload_download_rate": 50000, "payload_upload_rate": 10000,
		"dht_nodes": 200,
	}
	ts := newRPCServer(t, "core.get_session_status", result)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	stats, err := c.GetSessionStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.DownloadRate != 50000 {
		t.Errorf("DownloadRate = %d, want 50000", stats.DownloadRate)
	}
}

func TestGetFreeSpace(t *testing.T) {
	t.Parallel()

	ts := newRPCServer(t, "core.get_free_space", 500000000000)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	space, err := c.GetFreeSpace(context.Background(), "/downloads")
	if err != nil {
		t.Fatal(err)
	}
	if space != 500000000000 {
		t.Errorf("space = %d, want 500000000000", space)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"id":     req.ID,
			"result": nil,
			"error":  map[string]any{"message": "torrent not found", "code": 2},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetTorrentsStatus(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *deluge.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Message != "torrent not found" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "torrent not found")
	}
}

func TestConnected(t *testing.T) {
	t.Parallel()

	ts := newRPCServer(t, "web.connected", true)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ok, err := c.Connected(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("connected = false, want true")
	}
}

func TestSetTorrentLabel(t *testing.T) {
	t.Parallel()

	ts := newRPCServer(t, "label.set_torrent", nil)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.SetTorrentLabel(context.Background(), "abc123", "movies"); err != nil {
		t.Fatal(err)
	}
}

func TestMoveTorrent(t *testing.T) {
	t.Parallel()

	ts := newRPCServer(t, "core.move_storage", nil)
	defer ts.Close()

	c, err := deluge.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.MoveTorrent(context.Background(), []string{"abc123"}, "/new/path"); err != nil {
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
			c, err := deluge.New(tc.url)
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}
