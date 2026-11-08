package queue_test

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/downloadclient/queue"
)

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("writeJSON encode: %v", err)
	}
}

// droppingURL returns an http URL whose port stays bound for the test and drops every connection.
// Why: a closed server's port can be taken by a parallel test's server, which then answers.
func droppingURL(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for range 256 {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	return "http://" + ln.Addr().String()
}

func rpcMethod(t *testing.T, r *http.Request) string {
	t.Helper()
	var body struct {
		Method string `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Errorf("decode rpc method: %v", err)
	}
	return body.Method
}

// assertSnapshot verifies that the taken snapshot matches the expected snapshot.
func assertSnapshot(t *testing.T, got, want queue.Snapshot) {
	t.Helper()
	if got != want {
		t.Errorf("Take() = %+v, want %+v", got, want)
	}
}

// assertErrorPrefix verifies that err is non-nil and starts with the expected prefix.
func assertErrorPrefix(t *testing.T, err error, prefix string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with prefix %q, got nil", prefix)
	}
	if !strings.HasPrefix(err.Error(), prefix) {
		t.Errorf("err = %q, want prefix %q", err.Error(), prefix)
	}
}

// checkBasicAuth verifies that the request basic auth matches the expected user and pass.
func checkBasicAuth(r *http.Request, wantUser, wantPass string) bool {
	u, p, ok := r.BasicAuth()
	return ok && u == wantUser && p == wantPass
}

// TestTake_QBittorrent verifies taking a queue snapshot from qBittorrent and handling bad credentials.
func TestTake_QBittorrent(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("username") != "admin" || r.FormValue("password") != "secret" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "SID", Value: "session123"})
		_, _ = io.WriteString(w, "Ok.")
	})
	mux.HandleFunc("GET /api/v2/app/version", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "v5.0.1")
	})
	mux.HandleFunc("GET /api/v2/torrents/info", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, []map[string]any{
			{"state": "downloading", "dlspeed": 1000, "upspeed": 100},
			{"state": "metaDL", "dlspeed": 500, "upspeed": 50},
			{"state": "forcedDL", "dlspeed": 200, "upspeed": 20},
			{"state": "forcedMetaDL", "dlspeed": 100, "upspeed": 10},
			{"state": "stalledDL", "dlspeed": 0, "upspeed": 0},
			{"state": "checkingDL", "dlspeed": 50, "upspeed": 5},
			{"state": "queuedDL", "dlspeed": 0, "upspeed": 0},
			{"state": "pausedDL", "dlspeed": 0, "upspeed": 0}, // unmapped
		})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	snap, err := queue.Take(context.Background(), queue.Target{
		Kind:     queue.KindQBittorrent,
		BaseURL:  srv.URL,
		Username: "admin",
		Secret:   "secret",
	})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
	assertSnapshot(t, snap, queue.Snapshot{
		Kind:         queue.KindQBittorrent,
		Active:       6,
		Queued:       1,
		DownloadRate: 1850,
		UploadRate:   185,
		Version:      "v5.0.1",
	})

	// Bad credentials test
	_, errBad := queue.Take(context.Background(), queue.Target{
		Kind:     queue.KindQBittorrent,
		BaseURL:  srv.URL,
		Username: "admin",
		Secret:   "wrong",
	})
	assertErrorPrefix(t, errBad, "queue: qbittorrent: login:")
}

// TestTake_Transmission verifies taking a queue snapshot from Transmission.
func TestTake_Transmission(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /transmission/rpc", func(w http.ResponseWriter, r *http.Request) {
		if !checkBasicAuth(r, "user", "pass") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-Transmission-Session-Id") == "" {
			w.Header().Set("X-Transmission-Session-Id", "sessToken")
			w.WriteHeader(http.StatusConflict)
			return
		}
		switch rpcMethod(t, r) {
		case "session-get":
			writeJSON(t, w, map[string]any{"result": "success", "arguments": map[string]any{"version": "4.0.5"}})
		case "torrent-get":
			writeJSON(t, w, map[string]any{"result": "success", "arguments": map[string]any{"torrents": []map[string]any{
				{"id": 1, "status": 4, "rateDownload": 2000, "rateUpload": 200}, // downloading (mapped active)
				{"id": 2, "status": 3, "rateDownload": 0, "rateUpload": 0},      // downloadWait (mapped queued)
				{"id": 3, "status": 6, "rateDownload": 0, "rateUpload": 50},     // seeding (unmapped)
			}}})
		default:
			writeJSON(t, w, map[string]any{"result": "unknown method"})
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	snap, err := queue.Take(context.Background(), queue.Target{
		Kind:     queue.KindTransmission,
		BaseURL:  srv.URL + "/transmission",
		Username: "user",
		Secret:   "pass",
	})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
	assertSnapshot(t, snap, queue.Snapshot{
		Kind:         queue.KindTransmission,
		Active:       1,
		Queued:       1,
		DownloadRate: 2000,
		UploadRate:   250,
		Version:      "4.0.5",
	})
}

func TestTake_TransmissionURLVariants(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Transmission-Session-Id") == "" {
			w.Header().Set("X-Transmission-Session-Id", "tok")
			w.WriteHeader(http.StatusConflict)
			return
		}
		switch rpcMethod(t, r) {
		case "session-get":
			writeJSON(t, w, map[string]any{"result": "success", "arguments": map[string]any{"version": "4.0.5"}})
		case "torrent-get":
			writeJSON(t, w, map[string]any{"result": "success", "arguments": map[string]any{"torrents": []map[string]any{}}})
		default:
			writeJSON(t, w, map[string]any{"result": "unknown method"})
		}
	}
	mux.HandleFunc("POST /transmission/rpc", handler)
	mux.HandleFunc("POST /custom/rpc", handler)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	variants := []struct {
		name    string
		baseURL string
	}{
		{name: "default /transmission/", baseURL: srv.URL + "/transmission/"},
		{name: "default /transmission", baseURL: srv.URL + "/transmission"},
		{name: "custom /custom", baseURL: srv.URL + "/custom"},
		{name: "custom /custom/rpc", baseURL: srv.URL + "/custom/rpc"},
		{name: "empty path", baseURL: srv.URL},
		{name: "empty path slash", baseURL: srv.URL + "/"},
	}

	for _, v := range variants {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			snap, err := queue.Take(context.Background(), queue.Target{
				Kind:    queue.KindTransmission,
				BaseURL: v.baseURL,
			})
			if err != nil {
				t.Fatalf("Take(%q) error: %v", v.baseURL, err)
			}
			if snap.Version != "4.0.5" {
				t.Errorf("Version = %q, want %q", snap.Version, "4.0.5")
			}
		})
	}
}

// TestTake_Deluge verifies taking a queue snapshot from Deluge.
func TestTake_Deluge(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /json", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
			Params []any  `json:"params"`
			ID     int    `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode json: %v", err)
		}
		reply := func(result any) {
			writeJSON(t, w, map[string]any{"result": result, "error": nil, "id": body.ID})
		}
		switch body.Method {
		case "auth.login":
			reply(len(body.Params) == 1 && body.Params[0] == "delugepass")
		case "daemon.get_version", "daemon.info":
			reply("2.1.1")
		case "core.get_torrents_status":
			reply(map[string]any{
				"t1": map[string]any{"state": "Downloading", "download_payload_rate": 3000, "upload_payload_rate": 300},
				"t2": map[string]any{"state": "Queued", "download_payload_rate": 0, "upload_payload_rate": 0},
				"t3": map[string]any{"state": "Seeding", "download_payload_rate": 0, "upload_payload_rate": 50}, // unmapped
			})
		default:
			writeJSON(t, w, map[string]any{"result": nil, "error": map[string]any{"code": 1, "message": "unknown"}, "id": body.ID})
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	snap, err := queue.Take(context.Background(), queue.Target{
		Kind:    queue.KindDeluge,
		BaseURL: srv.URL,
		Secret:  "delugepass",
	})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
	assertSnapshot(t, snap, queue.Snapshot{
		Kind:         queue.KindDeluge,
		Active:       1,
		Queued:       1,
		DownloadRate: 3000,
		UploadRate:   350,
		Version:      "2.1.1",
	})
}

type rtTorrentFixture struct {
	downRate, upRate       int64
	open, active, complete bool
}

func rtString(s string) string { return "<value><string>" + s + "</string></value>" }
func rtNumber(n int64) string  { return fmt.Sprintf("<value><i8>%d</i8></value>", n) }
func rtFlag(set bool) string {
	if set {
		return rtNumber(1)
	}
	return rtNumber(0)
}

func (d rtTorrentFixture) row() string {
	fields := []string{
		rtString("0123456789ABCDEF0123456789ABCDEF01234567"), rtString("a download"),
		rtNumber(4_000_000_000), rtNumber(1_000_000_000), rtNumber(d.downRate), rtNumber(d.upRate),
		rtNumber(0), rtNumber(0), rtFlag(d.open), rtFlag(d.active), rtFlag(d.complete), rtFlag(false),
		rtString("/downloads/a download"), rtString("/downloads"), rtString(""), rtNumber(1_760_000_000),
		rtString(""), rtNumber(0),
	}
	return "<value><array><data>" + strings.Join(fields, "") + "</data></array></value>"
}

func rtListing(downloads ...rtTorrentFixture) string {
	rows := make([]string, 0, len(downloads))
	for _, d := range downloads {
		rows = append(rows, d.row())
	}
	return "<params><param><value><array><data>" + strings.Join(rows, "") + "</data></array></value></param></params>"
}

// rtXMLResponse returns the method response XML for a given method and multicall listing.
func rtXMLResponse(method, listing string) (string, bool) {
	switch method {
	case "system.client_version":
		return "<params><param>" + rtString("0.16.25") + "</param></params>", true
	case "system.library_version":
		return "<params><param>" + rtString("0.16.25-lib") + "</param></params>", true
	case "system.hostname":
		return "<params><param>" + rtString("seedbox") + "</param></params>", true
	case "system.pid":
		return "<params><param>" + rtNumber(1234) + "</param></params>", true
	case "d.multicall2":
		return listing, true
	default:
		return "", false
	}
}

// TestTake_RTorrent verifies taking a queue snapshot from rTorrent.
func TestTake_RTorrent(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkBasicAuth(r, "rtuser", "rtpass") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var call struct {
			Method string `xml:"methodName"`
		}
		if err := xml.NewDecoder(r.Body).Decode(&call); err != nil {
			t.Errorf("decode call: %v", err)
		}
		result, ok := rtXMLResponse(call.Method, rtListing(
			rtTorrentFixture{downRate: 4000, upRate: 400, open: true, active: true, complete: false}, // active
			rtTorrentFixture{downRate: 0, upRate: 100, open: true, active: true, complete: true},     // seeding (unmapped)
			rtTorrentFixture{downRate: 0, upRate: 0, open: true, active: false, complete: false},     // stopped (unmapped)
			rtTorrentFixture{downRate: 0, upRate: 0, open: false, active: false, complete: false},    // closed (unmapped)
		))
		if !ok {
			t.Errorf("unexpected xmlrpc method %q", call.Method)
		}
		w.Header().Set("Content-Type", "text/xml")
		_, _ = io.WriteString(w, xml.Header+"<methodResponse>"+result+"</methodResponse>")
	}))
	t.Cleanup(srv.Close)

	snap, err := queue.Take(context.Background(), queue.Target{
		Kind:     queue.KindRTorrent,
		BaseURL:  srv.URL,
		Username: "rtuser",
		Secret:   "rtpass",
	})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
	assertSnapshot(t, snap, queue.Snapshot{
		Kind:         queue.KindRTorrent,
		Active:       1,
		Queued:       0,
		DownloadRate: 4000,
		UploadRate:   500,
		Version:      "0.16.25",
	})
}

// TestTake_SABnzbd verifies taking a queue snapshot from SABnzbd.
func TestTake_SABnzbd(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("apikey") != "sabkey" {
			writeJSON(t, w, map[string]any{"status": false, "error": "API Key Incorrect"})
			return
		}
		switch r.URL.Query().Get("mode") {
		case "version":
			writeJSON(t, w, map[string]any{"version": "4.3.2"})
		case "queue":
			writeJSON(t, w, map[string]any{"queue": map[string]any{
				"paused":     true,
				"paused_all": false,
				"speed":      "2.5 M",
				"slots": []map[string]any{
					{"status": "Downloading"},
					{"status": "Queued"},
					{"status": "Queued"},
					{"status": "Paused"}, // unmapped
				},
			}})
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	snap, err := queue.Take(context.Background(), queue.Target{
		Kind:    queue.KindSABnzbd,
		BaseURL: srv.URL,
		Secret:  "sabkey",
	})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
	assertSnapshot(t, snap, queue.Snapshot{
		Kind:         queue.KindSABnzbd,
		Active:       1,
		Queued:       2,
		DownloadRate: 2_621_440,
		UploadRate:   0,
		Version:      "4.3.2",
		Paused:       true,
	})
}

// TestTake_NZBGet verifies taking a queue snapshot from NZBGet.
func TestTake_NZBGet(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jsonrpc", func(w http.ResponseWriter, r *http.Request) {
		if !checkBasicAuth(r, "nzbuser", "nzbpass") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch rpcMethod(t, r) {
		case "version":
			writeJSON(t, w, map[string]any{"result": "24.3"})
		case "status":
			writeJSON(t, w, map[string]any{"result": map[string]any{
				"DownloadPaused": true,
				"DownloadRate":   8192,
			}})
		case "listgroups":
			writeJSON(t, w, map[string]any{"result": []map[string]any{
				{"Status": "DOWNLOADING_NZB", "ActiveDownloads": 1},
				{"Status": "DOWNLOADING_PAR", "ActiveDownloads": 0},
				{"Status": "QUEUED", "ActiveDownloads": 0},
				{"Status": "PAUSED", "ActiveDownloads": 0}, // unmapped
			}})
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	snap, err := queue.Take(context.Background(), queue.Target{
		Kind:     queue.KindNZBGet,
		BaseURL:  srv.URL,
		Username: "nzbuser",
		Secret:   "nzbpass",
	})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
	assertSnapshot(t, snap, queue.Snapshot{
		Kind:         queue.KindNZBGet,
		Active:       2,
		Queued:       1,
		DownloadRate: 8192,
		UploadRate:   0,
		Version:      "24.3",
		Paused:       true,
	})
}

func TestTake_Unreachable(t *testing.T) {
	t.Parallel()

	closedURL := droppingURL(t)

	kinds := []queue.Kind{
		queue.KindQBittorrent,
		queue.KindTransmission,
		queue.KindDeluge,
		queue.KindRTorrent,
		queue.KindSABnzbd,
		queue.KindNZBGet,
	}

	for _, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			t.Parallel()
			_, err := queue.Take(context.Background(), queue.Target{
				Kind:     kind,
				BaseURL:  closedURL,
				Username: "u",
				Secret:   "p",
			}, queue.WithTimeout(500*time.Millisecond))
			if err == nil {
				t.Fatalf("Take(%s) with closed server expected error, got nil", kind)
			}
			prefix := fmt.Sprintf("queue: %s:", kind)
			if !strings.HasPrefix(err.Error(), prefix) {
				t.Errorf("err = %q, want prefix %q", err.Error(), prefix)
			}
		})
	}
}

func TestTake_UnknownKind(t *testing.T) {
	t.Parallel()

	_, err := queue.Take(context.Background(), queue.Target{
		Kind:    "unknown-client",
		BaseURL: "http://127.0.0.1:8080",
	})
	if err == nil {
		t.Fatalf("Take(unknown) expected error, got nil")
	}
	if !errors.Is(err, queue.ErrUnsupportedKind) {
		t.Errorf("errors.Is(err, ErrUnsupportedKind) = false, err = %v", err)
	}
	wantPrefix := "queue: unknown-client:"
	if !strings.HasPrefix(err.Error(), wantPrefix) {
		t.Errorf("err = %q, want prefix %q", err.Error(), wantPrefix)
	}
}

// checkNoWrite verifies that action contains none of the blocked write operations.
func checkNoWrite(t *testing.T, service, action string, blocked []string) {
	t.Helper()
	lowerAction := strings.ToLower(action)
	for _, b := range blocked {
		if strings.Contains(lowerAction, strings.ToLower(b)) {
			t.Fatalf("%s write violation: executed write operation %q", service, action)
		}
	}
}

// testNeverWritesQBittorrent verifies that Take makes no write requests to qBittorrent.
func testNeverWritesQBittorrent(t *testing.T) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		checkNoWrite(t, "qbittorrent", r.URL.Path, []string{"/pause", "/resume", "/delete", "/add"})
		switch r.URL.Path {
		case "/api/v2/app/version":
			_, _ = io.WriteString(w, "5.0.0")
		case "/api/v2/torrents/info":
			writeJSON(t, w, []map[string]any{})
		default:
			http.NotFound(w, r)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, err := queue.Take(context.Background(), queue.Target{Kind: queue.KindQBittorrent, BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
}

// testNeverWritesTransmission verifies that Take makes no write requests to Transmission.
func testNeverWritesTransmission(t *testing.T) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /transmission/rpc", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Transmission-Session-Id") == "" {
			w.Header().Set("X-Transmission-Session-Id", "sess")
			w.WriteHeader(http.StatusConflict)
			return
		}
		method := rpcMethod(t, r)
		checkNoWrite(t, "transmission", method, []string{"torrent-start", "torrent-stop", "torrent-remove", "torrent-add"})
		switch method {
		case "session-get":
			writeJSON(t, w, map[string]any{"result": "success", "arguments": map[string]any{"version": "4.0.0"}})
		case "torrent-get":
			writeJSON(t, w, map[string]any{"result": "success", "arguments": map[string]any{"torrents": []map[string]any{}}})
		default:
			writeJSON(t, w, map[string]any{"result": "unknown method"})
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, err := queue.Take(context.Background(), queue.Target{Kind: queue.KindTransmission, BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
}

// testNeverWritesDeluge verifies that Take makes no write requests to Deluge.
func testNeverWritesDeluge(t *testing.T) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /json", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
			ID     int    `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode deluge body: %v", err)
		}
		checkNoWrite(t, "deluge", body.Method, []string{
			"core.pause", "core.resume", "core.remove", "core.add",
		})
		switch body.Method {
		case "auth.login":
			writeJSON(t, w, map[string]any{"result": true, "error": nil, "id": body.ID})
		case "daemon.get_version", "daemon.info":
			writeJSON(t, w, map[string]any{"result": "2.1.0", "error": nil, "id": body.ID})
		case "core.get_torrents_status":
			writeJSON(t, w, map[string]any{"result": map[string]any{}, "error": nil, "id": body.ID})
		default:
			writeJSON(t, w, map[string]any{"result": nil, "error": map[string]any{"code": 1, "message": "unknown"}, "id": body.ID})
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, err := queue.Take(context.Background(), queue.Target{Kind: queue.KindDeluge, BaseURL: srv.URL, Secret: "p"})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
}

// testNeverWritesRTorrent verifies that Take makes no write requests to rTorrent.
func testNeverWritesRTorrent(t *testing.T) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var call struct {
			Method string `xml:"methodName"`
		}
		if err := xml.NewDecoder(r.Body).Decode(&call); err != nil {
			t.Errorf("decode call: %v", err)
		}
		checkNoWrite(t, "rtorrent", call.Method, []string{"d.stop", "d.start", "d.erase", "d.load"})
		result, ok := rtXMLResponse(call.Method, rtListing())
		if !ok {
			t.Errorf("unexpected xmlrpc call %q", call.Method)
		}
		w.Header().Set("Content-Type", "text/xml")
		_, _ = io.WriteString(w, xml.Header+"<methodResponse>"+result+"</methodResponse>")
	}))
	t.Cleanup(srv.Close)

	_, err := queue.Take(context.Background(), queue.Target{Kind: queue.KindRTorrent, BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
}

// testNeverWritesSABnzbd verifies that Take makes no write requests to SABnzbd.
func testNeverWritesSABnzbd(t *testing.T) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		mode := q.Get("mode")
		name := q.Get("name")
		checkNoWrite(t, "sabnzbd", "mode="+mode, []string{"mode=pause", "mode=resume"})
		checkNoWrite(t, "sabnzbd", "name="+name, []string{"name=delete", "name=addurl", "name=addfile"})
		switch mode {
		case "version":
			writeJSON(t, w, map[string]any{"version": "4.3.0"})
		case "queue":
			writeJSON(t, w, map[string]any{"queue": map[string]any{"paused": false, "speed": "0", "slots": []map[string]any{}}})
		default:
			http.NotFound(w, r)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, err := queue.Take(context.Background(), queue.Target{Kind: queue.KindSABnzbd, BaseURL: srv.URL, Secret: "key"})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
}

// testNeverWritesNZBGet verifies that Take makes no write requests to NZBGet.
func testNeverWritesNZBGet(t *testing.T) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /jsonrpc", func(w http.ResponseWriter, r *http.Request) {
		method := rpcMethod(t, r)
		checkNoWrite(t, "nzbget", method, []string{"pausedownload", "resumedownload", "editqueue", "append"})
		switch method {
		case "version":
			writeJSON(t, w, map[string]any{"result": "24.0"})
		case "status":
			writeJSON(t, w, map[string]any{"result": map[string]any{"DownloadPaused": false, "DownloadRate": 0}})
		case "listgroups":
			writeJSON(t, w, map[string]any{"result": []map[string]any{}})
		default:
			writeJSON(t, w, map[string]any{"result": "unknown"})
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, err := queue.Take(context.Background(), queue.Target{Kind: queue.KindNZBGet, BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
}

// TestNeverWrites verifies that Take makes no write requests across all supported download clients.
func TestNeverWrites(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{name: "qbittorrent", fn: testNeverWritesQBittorrent},
		{name: "transmission", fn: testNeverWritesTransmission},
		{name: "deluge", fn: testNeverWritesDeluge},
		{name: "rtorrent", fn: testNeverWritesRTorrent},
		{name: "sabnzbd", fn: testNeverWritesSABnzbd},
		{name: "nzbget", fn: testNeverWritesNZBGet},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t)
		})
	}
}

func TestParseSABSpeed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  int64
	}{
		{input: "1.5 M", want: 1_572_864},
		{input: "512 K", want: 524_288},
		{input: "2 G", want: 2_147_483_648},
		{input: "1 T", want: 1_099_511_627_776},
		{input: "300", want: 300},
		{input: "0 ", want: 0},
		{input: "", want: 0},
		{input: "   ", want: 0},
		{input: "abc", want: 0},
		{input: "-1 M", want: 0},
		{input: "1.2 k", want: 1229},
		{input: "inf", want: 0},
		{input: "nan", want: 0},
		{input: "1.2.3 M", want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := queue.ParseSABSpeed(tc.input)
			if got != tc.want {
				t.Errorf("ParseSABSpeed(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	var gotUA atomic.Value
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v2/app/version", func(w http.ResponseWriter, r *http.Request) {
		gotUA.Store(r.Header.Get("User-Agent"))
		_, _ = io.WriteString(w, "5.0.0")
	})
	mux.HandleFunc("GET /api/v2/torrents/info", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, []map[string]any{})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	customClient := srv.Client()
	customUA := "CustomAgent/1.0"

	snap, err := queue.Take(
		context.Background(),
		queue.Target{Kind: queue.KindQBittorrent, BaseURL: srv.URL},
		queue.WithHTTPClient(customClient),
		queue.WithUserAgent(customUA),
		queue.WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
	if snap.Version != "5.0.0" {
		t.Errorf("Version = %q, want 5.0.0", snap.Version)
	}
	if ua, ok := gotUA.Load().(string); !ok || ua != customUA {
		t.Errorf("User-Agent = %q, want %q", ua, customUA)
	}
}

func TestTarget_SecretNotSerialized(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(queue.Target{
		Kind:     queue.KindQBittorrent,
		BaseURL:  "http://example.com",
		Username: "user",
		Secret:   "s3cret",
	})
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	s := string(raw)
	if strings.Contains(s, "s3cret") {
		t.Errorf("Marshal(Target{Secret: %q}) = %s, want secret omitted", "s3cret", s)
	}
}

func TestTake_Deluge_EmptySecretSendsLogin(t *testing.T) {
	t.Parallel()

	var loginReceived atomic.Bool
	mux := http.NewServeMux()
	mux.HandleFunc("POST /json", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
			Params []any  `json:"params"`
			ID     int    `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode json: %v", err)
		}
		reply := func(result any) {
			writeJSON(t, w, map[string]any{"result": result, "error": nil, "id": body.ID})
		}
		switch body.Method {
		case "auth.login":
			loginReceived.Store(true)
			reply(len(body.Params) == 1 && body.Params[0] == "")
		case "daemon.get_version", "daemon.info":
			reply("2.1.1")
		case "core.get_torrents_status":
			reply(map[string]any{})
		default:
			writeJSON(t, w, map[string]any{"result": nil, "error": map[string]any{"code": 1, "message": "unknown"}, "id": body.ID})
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, err := queue.Take(context.Background(), queue.Target{
		Kind:    queue.KindDeluge,
		BaseURL: srv.URL,
		Secret:  "",
	})
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}
	if !loginReceived.Load() {
		t.Errorf("auth.login was not sent for empty Secret")
	}
}
