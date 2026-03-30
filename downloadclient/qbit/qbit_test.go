package qbit_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lusoris/goenvoy/downloadclient/qbit"
)

func newTestServer(t *testing.T, wantPath string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
}

func newLoginServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			if r.Method != http.MethodPost {
				t.Errorf("login method = %q, want POST", r.Method)
			}
			http.SetCookie(w, &http.Cookie{Name: "SID", Value: "test-session-id", Path: "/"})
			w.WriteHeader(http.StatusOK)
		case "/api/v2/auth/logout":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusForbidden)
		}
	}))
}

func TestLogin(t *testing.T) {
	ts := newLoginServer(t)
	defer ts.Close()

	c := qbit.New(ts.URL)
	if err := c.Login(context.Background(), "admin", "adminadmin"); err != nil {
		t.Fatal(err)
	}
}

func TestLogout(t *testing.T) {
	ts := newLoginServer(t)
	defer ts.Close()

	c := qbit.New(ts.URL)
	_ = c.Login(context.Background(), "admin", "adminadmin")
	if err := c.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("v4.6.7"))
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	v, err := c.Version(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v != "v4.6.7" {
		t.Errorf("version = %q, want %q", v, "v4.6.7")
	}
}

func TestWebAPIVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("2.10.5"))
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	v, err := c.WebAPIVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v != "2.10.5" {
		t.Errorf("version = %q, want %q", v, "2.10.5")
	}
}

func TestGetBuildInfo(t *testing.T) {
	info := qbit.BuildInfo{Qt: "6.7.2", Libtorrent: "2.0.10.0", Boost: "1.86", OpenSSL: "3.3.1", Bitness: 64}
	ts := newTestServer(t, "/api/v2/app/buildInfo", info)
	defer ts.Close()

	c := qbit.New(ts.URL)
	b, err := c.GetBuildInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if b.Qt != "6.7.2" {
		t.Errorf("Qt = %q, want %q", b.Qt, "6.7.2")
	}
	if b.Bitness != 64 {
		t.Errorf("Bitness = %d, want %d", b.Bitness, 64)
	}
}

func TestGetPreferences(t *testing.T) {
	prefs := qbit.Preferences{SavePath: "/downloads", DlLimit: 5000000, QueueingEnabled: true}
	ts := newTestServer(t, "/api/v2/app/preferences", prefs)
	defer ts.Close()

	c := qbit.New(ts.URL)
	p, err := c.GetPreferences(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p.SavePath != "/downloads" {
		t.Errorf("SavePath = %q, want %q", p.SavePath, "/downloads")
	}
	if p.DlLimit != 5000000 {
		t.Errorf("DlLimit = %d, want %d", p.DlLimit, 5000000)
	}
}

func TestDefaultSavePath(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("/downloads/complete"))
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	p, err := c.DefaultSavePath(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p != "/downloads/complete" {
		t.Errorf("path = %q, want %q", p, "/downloads/complete")
	}
}

func TestListTorrents(t *testing.T) {
	torrents := []qbit.Torrent{
		{Hash: "abc123", Name: "Ubuntu 24.04", State: "downloading", Progress: 0.45, Size: 4000000000},
		{Hash: "def456", Name: "Fedora 40", State: "seeding", Progress: 1.0, Size: 2000000000},
	}
	ts := newTestServer(t, "/api/v2/torrents/info", torrents)
	defer ts.Close()

	c := qbit.New(ts.URL)
	result, err := c.ListTorrents(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0].Name != "Ubuntu 24.04" {
		t.Errorf("Name = %q, want %q", result[0].Name, "Ubuntu 24.04")
	}
	if result[1].State != "seeding" {
		t.Errorf("State = %q, want %q", result[1].State, "seeding")
	}
}

func TestListTorrentsWithOptions(t *testing.T) {
	torrents := []qbit.Torrent{{Hash: "abc123", Name: "Test", Category: "movies"}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filter"); got != "downloading" {
			t.Errorf("filter = %q, want %q", got, "downloading")
		}
		if got := r.URL.Query().Get("category"); got != "movies" {
			t.Errorf("category = %q, want %q", got, "movies")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(torrents)
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	result, err := c.ListTorrents(context.Background(), &qbit.ListOptions{Filter: "downloading", Category: "movies"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
}

func TestGetTorrentProperties(t *testing.T) {
	props := qbit.TorrentProperties{SavePath: "/data/movies", TotalSize: 50000000, Seeds: 42, Peers: 10}
	ts := newTestServer(t, "/api/v2/torrents/properties", props)
	defer ts.Close()

	c := qbit.New(ts.URL)
	p, err := c.GetTorrentProperties(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if p.SavePath != "/data/movies" {
		t.Errorf("SavePath = %q, want %q", p.SavePath, "/data/movies")
	}
	if p.Seeds != 42 {
		t.Errorf("Seeds = %d, want %d", p.Seeds, 42)
	}
}

func TestGetTorrentTrackers(t *testing.T) {
	trackers := []qbit.Tracker{{URL: "udp://tracker.example.com:1337", Status: 2, NumPeers: 150}}
	ts := newTestServer(t, "/api/v2/torrents/trackers", trackers)
	defer ts.Close()

	c := qbit.New(ts.URL)
	result, err := c.GetTorrentTrackers(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].NumPeers != 150 {
		t.Errorf("NumPeers = %d, want %d", result[0].NumPeers, 150)
	}
}

func TestGetTorrentFiles(t *testing.T) {
	files := []qbit.TorrentFile{{Index: 0, Name: "movie.mkv", Size: 5000000000, Progress: 0.8, Priority: 1}}
	ts := newTestServer(t, "/api/v2/torrents/files", files)
	defer ts.Close()

	c := qbit.New(ts.URL)
	result, err := c.GetTorrentFiles(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Name != "movie.mkv" {
		t.Errorf("Name = %q, want %q", result[0].Name, "movie.mkv")
	}
}

func TestAddTorrentURLs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/api/v2/torrents/add" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/api/v2/torrents/add")
		}
		_ = r.ParseForm()
		if got := r.FormValue("urls"); got == "" {
			t.Error("urls is empty")
		}
		if got := r.FormValue("category"); got != "movies" {
			t.Errorf("category = %q, want %q", got, "movies")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	err := c.AddTorrentURLs(context.Background(), []string{"magnet:?xt=urn:btih:abc123"}, &qbit.AddTorrentOptions{Category: "movies"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteTorrents(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		_ = r.ParseForm()
		if got := r.FormValue("hashes"); got != "abc123|def456" {
			t.Errorf("hashes = %q, want %q", got, "abc123|def456")
		}
		if got := r.FormValue("deleteFiles"); got != "true" {
			t.Errorf("deleteFiles = %q, want %q", got, "true")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	if err := c.DeleteTorrents(context.Background(), []string{"abc123", "def456"}, true); err != nil {
		t.Fatal(err)
	}
}

func TestPauseTorrents(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/pause" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	if err := c.PauseTorrents(context.Background(), []string{"abc123"}); err != nil {
		t.Fatal(err)
	}
}

func TestResumeTorrents(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/resume" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	if err := c.ResumeTorrents(context.Background(), []string{"abc123"}); err != nil {
		t.Fatal(err)
	}
}

func TestGetTransferInfo(t *testing.T) {
	info := qbit.TransferInfo{
		DlInfoSpeed: 5000000, UpInfoSpeed: 1000000, DHTNodes: 450,
		ConnectionStatus: "connected", FreeSpaceOnDisk: 500000000000,
	}
	ts := newTestServer(t, "/api/v2/transfer/info", info)
	defer ts.Close()

	c := qbit.New(ts.URL)
	result, err := c.GetTransferInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.DlInfoSpeed != 5000000 {
		t.Errorf("DlInfoSpeed = %d, want %d", result.DlInfoSpeed, 5000000)
	}
	if result.ConnectionStatus != "connected" {
		t.Errorf("ConnectionStatus = %q, want %q", result.ConnectionStatus, "connected")
	}
}

func TestListCategories(t *testing.T) {
	cats := map[string]*qbit.Category{
		"movies": {Name: "movies", SavePath: "/data/movies"},
		"tv":     {Name: "tv", SavePath: "/data/tv"},
	}
	ts := newTestServer(t, "/api/v2/torrents/categories", cats)
	defer ts.Close()

	c := qbit.New(ts.URL)
	result, err := c.ListCategories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result["movies"].SavePath != "/data/movies" {
		t.Errorf("SavePath = %q, want %q", result["movies"].SavePath, "/data/movies")
	}
}

func TestListTags(t *testing.T) {
	tags := []string{"4k", "remux", "web-dl"}
	ts := newTestServer(t, "/api/v2/torrents/tags", tags)
	defer ts.Close()

	c := qbit.New(ts.URL)
	result, err := c.ListTags(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 3 {
		t.Fatalf("len = %d, want 3", len(result))
	}
	if result[0] != "4k" {
		t.Errorf("tag = %q, want %q", result[0], "4k")
	}
}

func TestGetSyncMainData(t *testing.T) {
	data := qbit.SyncMainData{
		RID:        1,
		FullUpdate: true,
		Torrents:   map[string]*qbit.Torrent{"abc": {Name: "Test Torrent", Hash: "abc"}},
		Tags:       []string{"hd"},
	}
	ts := newTestServer(t, "/api/v2/sync/maindata", data)
	defer ts.Close()

	c := qbit.New(ts.URL)
	result, err := c.GetSyncMainData(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if !result.FullUpdate {
		t.Error("FullUpdate = false, want true")
	}
	if result.Torrents["abc"].Name != "Test Torrent" {
		t.Errorf("Name = %q, want %q", result.Torrents["abc"].Name, "Test Torrent")
	}
}

func TestGetLog(t *testing.T) {
	logs := []qbit.LogEntry{
		{ID: 1, Message: "qBittorrent started", Timestamp: 1700000000, Type: 1},
		{ID: 2, Message: "Torrent added", Timestamp: 1700000060, Type: 2},
	}
	ts := newTestServer(t, "/api/v2/log/main", logs)
	defer ts.Close()

	c := qbit.New(ts.URL)
	result, err := c.GetLog(context.Background(), -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0].Message != "qBittorrent started" {
		t.Errorf("Message = %q, want %q", result[0].Message, "qBittorrent started")
	}
}

func TestAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Forbidden"))
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	_, err := c.ListTorrents(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *qbit.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusForbidden)
	}
}

func TestRecheckTorrents(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/recheck" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	if err := c.RecheckTorrents(context.Background(), []string{"abc123"}); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCategory(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/createCategory" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = r.ParseForm()
		if got := r.FormValue("category"); got != "movies" {
			t.Errorf("category = %q, want %q", got, "movies")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	if err := c.CreateCategory(context.Background(), "movies", "/data/movies"); err != nil {
		t.Fatal(err)
	}
}

func TestGetPeerLog(t *testing.T) {
	logs := []qbit.PeerLogEntry{
		{ID: 1, IP: "192.168.1.100", Timestamp: 1700000000, Blocked: false},
	}
	ts := newTestServer(t, "/api/v2/log/peers", logs)
	defer ts.Close()

	c := qbit.New(ts.URL)
	result, err := c.GetPeerLog(context.Background(), -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].IP != "192.168.1.100" {
		t.Errorf("IP = %q, want %q", result[0].IP, "192.168.1.100")
	}
}

func TestSetGlobalDownloadLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/transfer/setDownloadLimit" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = r.ParseForm()
		if got := r.FormValue("limit"); got != "10000000" {
			t.Errorf("limit = %q, want %q", got, "10000000")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := qbit.New(ts.URL)
	if err := c.SetGlobalDownloadLimit(context.Background(), 10000000); err != nil {
		t.Fatal(err)
	}
}
