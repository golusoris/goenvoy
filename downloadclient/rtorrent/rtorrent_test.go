package rtorrent_test

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golusoris/goenvoy/downloadclient/rtorrent/v2"
)

type xmlMethodCall struct {
	XMLName    xml.Name `xml:"methodCall"`
	MethodName string   `xml:"methodName"`
}

func xmlStringResponse(s string) string {
	return fmt.Sprintf(`<?xml version="1.0"?>
<methodResponse><params><param><value><string>%s</string></value></param></params></methodResponse>`, s)
}

func xmlIntResponse(n int) string {
	return fmt.Sprintf(`<?xml version="1.0"?>
<methodResponse><params><param><value><i8>%d</i8></value></param></params></methodResponse>`, n)
}

func xmlVoidResponse() string {
	return `<?xml version="1.0"?>
<methodResponse><params><param><value><i8>0</i8></value></param></params></methodResponse>`
}

func xmlFaultResponse(value string) string {
	return `<?xml version="1.0"?>
<methodResponse><fault><value>` + value + `</value></fault></methodResponse>`
}

func xmlMember(name, value string) string {
	return "<member><name>" + name + "</name><value>" + value + "</value></member>"
}

// xmlTorrentRow is one d.multicall2 row with the given values for d.is_open, d.is_active,
// d.complete and d.is_hash_checking, plus an optional d.hashing value (defaulting to 0).
// An empty flag or hashing string is a value without a typed element.
func xmlTorrentRow(flags [4]string, hashing ...string) string {
	const (
		str = "<value><string>x</string></value>"
		num = "<value><i8>7</i8></value>"
	)
	hashVal := "<value><i8>0</i8></value>"
	if len(hashing) > 0 {
		hashVal = "<value>" + hashing[0] + "</value>"
	}
	row := str + str + num + num + num + num + num + num +
		"<value>" + strings.Join(flags[:], "</value><value>") + "</value>" +
		str + str + str + num + str + hashVal
	return `<?xml version="1.0"?>
<methodResponse><params><param><value><array><data><value><array><data>` + row +
		`</data></array></value></data></array></value></param></params></methodResponse>`
}

func xmlMulticallResponse() string {
	return `<?xml version="1.0"?>
<methodResponse><params><param><value><array><data>
<value><array><data>
  <value><string>ABC123</string></value>
  <value><string>Ubuntu 24.04</string></value>
  <value><i8>4000000000</i8></value>
  <value><i8>4000000000</i8></value>
  <value><i8>0</i8></value>
  <value><i8>50000</i8></value>
  <value><i8>8000000000</i8></value>
  <value><i8>2000</i8></value>
  <value><i8>1</i8></value>
  <value><i8>1</i8></value>
  <value><i8>1</i8></value>
  <value><i8>0</i8></value>
  <value><string>/downloads/Ubuntu</string></value>
  <value><string>/downloads</string></value>
  <value><string>linux</string></value>
  <value><i8>1700000000</i8></value>
  <value><string></string></value>
  <value><i8>0</i8></value>
</data></array></value>
<value><array><data>
  <value><string>DEF456</string></value>
  <value><string>Fedora 40</string></value>
  <value><i8>2000000000</i8></value>
  <value><i8>1000000000</i8></value>
  <value><i8>5000000</i8></value>
  <value><i8>10000</i8></value>
  <value><i8>500000000</i8></value>
  <value><i8>500</i8></value>
  <value><i8>1</i8></value>
  <value><i8>1</i8></value>
  <value><i8>0</i8></value>
  <value><i8>0</i8></value>
  <value><string>/downloads/Fedora</string></value>
  <value><string>/downloads</string></value>
  <value><string>linux</string></value>
  <value><i8>1700000001</i8></value>
  <value><string></string></value>
  <value><i8>0</i8></value>
</data></array></value>
</data></array></value></param></params></methodResponse>`
}

func newXMLRPCServer(t *testing.T, wantMethod, response string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("HTTP method = %q, want POST", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var call xmlMethodCall
		if err := xml.Unmarshal(body, &call); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if call.MethodName != wantMethod {
			t.Errorf("method = %q, want %q", call.MethodName, wantMethod)
		}
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(response))
	}))
}

func TestGetTorrents(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "d.multicall2", xmlMulticallResponse())
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	torrents, err := c.GetTorrents(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(torrents) != 2 {
		t.Fatalf("len = %d, want 2", len(torrents))
	}
	if torrents[0].Hash != "ABC123" {
		t.Errorf("Hash = %q, want %q", torrents[0].Hash, "ABC123")
	}
	if torrents[0].Name != "Ubuntu 24.04" {
		t.Errorf("Name = %q, want %q", torrents[0].Name, "Ubuntu 24.04")
	}
	if got := flagsOf(&torrents[0]); got != [4]bool{true, true, true, false} {
		t.Errorf("flags = %v, want open, active and complete", got)
	}
	if torrents[1].Hash != "DEF456" {
		t.Errorf("Hash = %q, want %q", torrents[1].Hash, "DEF456")
	}
	if got := flagsOf(&torrents[1]); got != [4]bool{true, true, false, false} {
		t.Errorf("flags = %v, want open and active", got)
	}
}

func flagsOf(t *rtorrent.Torrent) [4]bool {
	return [4]bool{t.IsOpen, t.IsActive, t.IsComplete, t.IsHashChecking}
}

func getFlags(t *testing.T, flags [4]string) [4]bool {
	t.Helper()
	ts := newXMLRPCServer(t, "d.multicall2", xmlTorrentRow(flags))
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	torrents, err := c.GetTorrents(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(torrents) != 1 {
		t.Fatalf("len = %d, want 1", len(torrents))
	}
	return flagsOf(&torrents[0])
}

// rTorrent writes every number as an i8, the 0 and 1 of the d.is_* commands and of d.complete
// included (rakshasa/rtorrent v0.16.25, src/rpc/xmlrpc_tinyxml2.cc, print_object_xml). The i4
// and boolean rows are the other two encodings the client has always read.
func TestGetTorrents_flagEncodings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		flag string
		want bool
	}{
		{"i8 one", "<i8>1</i8>", true},
		{"i8 zero", "<i8>0</i8>", false},
		{"i4 one", "<i4>1</i4>", true},
		{"i4 zero", "<i4>0</i4>", false},
		{"boolean one", "<boolean>1</boolean>", true},
		{"boolean zero", "<boolean>0</boolean>", false},
		{"no typed value", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			want := [4]bool{tc.want, tc.want, tc.want, tc.want}
			if got := getFlags(t, [4]string{tc.flag, tc.flag, tc.flag, tc.flag}); got != want {
				t.Errorf("flags = %v, want %v", got, want)
			}
		})
	}
}

// Each flag is read from its own column: one set flag must not show up in another field.
func TestGetTorrents_flagColumns(t *testing.T) {
	t.Parallel()

	names := [4]string{"open", "active", "complete", "hash checking"}
	for i, name := range names {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			flags := [4]string{"<i8>0</i8>", "<i8>0</i8>", "<i8>0</i8>", "<i8>0</i8>"}
			flags[i] = "<i8>1</i8>"
			var want [4]bool
			want[i] = true
			if got := getFlags(t, flags); got != want {
				t.Errorf("flags = %v, want %v", got, want)
			}
		})
	}
}

func TestGetTorrents_hashing(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		flags       [4]string
		hashing     string
		wantHashing int64
	}{
		{
			name:        "closed inactive incomplete queued for hash check",
			flags:       [4]string{"<i8>0</i8>", "<i8>0</i8>", "<i8>0</i8>", "<i8>0</i8>"},
			hashing:     "<i8>1</i8>",
			wantHashing: 1,
		},
		{
			name:        "hashing zero boundary",
			flags:       [4]string{"<i8>0</i8>", "<i8>0</i8>", "<i8>0</i8>", "<i8>0</i8>"},
			hashing:     "<i8>0</i8>",
			wantHashing: 0,
		},
		{
			name:        "missing or empty typed value",
			flags:       [4]string{"<i8>0</i8>", "<i8>0</i8>", "<i8>0</i8>", "<i8>0</i8>"},
			hashing:     "",
			wantHashing: 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ts := newXMLRPCServer(t, "d.multicall2", xmlTorrentRow(tc.flags, tc.hashing))
			defer ts.Close()

			c, err := rtorrent.New(ts.URL)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			torrents, err := c.GetTorrents(context.Background(), "")
			if err != nil {
				t.Fatal(err)
			}
			if len(torrents) != 1 {
				t.Fatalf("len = %d, want 1", len(torrents))
			}
			got := torrents[0]
			if got.IsOpen || got.IsActive || got.IsComplete || got.IsHashChecking {
				t.Errorf("flags = {%v, %v, %v, %v}, want all false",
					got.IsOpen, got.IsActive, got.IsComplete, got.IsHashChecking)
			}
			if got.Hashing != tc.wantHashing {
				t.Errorf("Hashing = %d, want %d", got.Hashing, tc.wantHashing)
			}
		})
	}
}

func TestAddTorrentURL(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "load.start", xmlVoidResponse())
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.AddTorrentURL(context.Background(), "magnet:?xt=urn:btih:abc123"); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveTorrent(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "d.erase", xmlVoidResponse())
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.RemoveTorrent(context.Background(), "ABC123"); err != nil {
		t.Fatal(err)
	}
}

func TestStartTorrent(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "d.start", xmlVoidResponse())
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.StartTorrent(context.Background(), "ABC123"); err != nil {
		t.Fatal(err)
	}
}

func TestStopTorrent(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "d.stop", xmlVoidResponse())
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.StopTorrent(context.Background(), "ABC123"); err != nil {
		t.Fatal(err)
	}
}

func TestPauseTorrent(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "d.pause", xmlVoidResponse())
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.PauseTorrent(context.Background(), "DEF456"); err != nil {
		t.Fatal(err)
	}
}

func TestResumeTorrent(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "d.resume", xmlVoidResponse())
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.ResumeTorrent(context.Background(), "DEF456"); err != nil {
		t.Fatal(err)
	}
}

func TestRecheckTorrent(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "d.check_hash", xmlVoidResponse())
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.RecheckTorrent(context.Background(), "ABC123"); err != nil {
		t.Fatal(err)
	}
}

func TestSetLabel(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "d.custom1.set", xmlVoidResponse())
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.SetLabel(context.Background(), "ABC123", "movies"); err != nil {
		t.Fatal(err)
	}
}

func TestGetSystemInfo(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/xml")
		switch {
		case strings.Contains(string(body), "system.client_version"):
			_, _ = w.Write([]byte(xmlStringResponse("0.9.8")))
		case strings.Contains(string(body), "system.library_version"):
			_, _ = w.Write([]byte(xmlStringResponse("0.13.8")))
		case strings.Contains(string(body), "system.hostname"):
			_, _ = w.Write([]byte(xmlStringResponse("seedbox")))
		case strings.Contains(string(body), "system.pid"):
			_, _ = w.Write([]byte(xmlIntResponse(1234)))
		}
	}))
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	info, err := c.GetSystemInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.ClientVersion != "0.9.8" {
		t.Errorf("ClientVersion = %q, want %q", info.ClientVersion, "0.9.8")
	}
	if info.Hostname != "seedbox" {
		t.Errorf("Hostname = %q, want %q", info.Hostname, "seedbox")
	}
	if info.PID != 1234 {
		t.Errorf("PID = %d, want 1234", info.PID)
	}
}

func TestGetDownloadRate(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "throttle.global_down.rate", xmlIntResponse(5000000))
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rate, err := c.GetDownloadRate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rate != 5000000 {
		t.Errorf("rate = %d, want 5000000", rate)
	}
}

func TestGetUploadRate(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "throttle.global_up.rate", xmlIntResponse(1000000))
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rate, err := c.GetUploadRate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rate != 1000000 {
		t.Errorf("rate = %d, want 1000000", rate)
	}
}

func TestSetDownloadLimit(t *testing.T) {
	t.Parallel()

	ts := newXMLRPCServer(t, "throttle.global_down.max_rate.set", xmlVoidResponse())
	defer ts.Close()

	c, err := rtorrent.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.SetDownloadLimit(context.Background(), 5000000); err != nil {
		t.Fatal(err)
	}
}

// A fault is a struct of faultCode and faultString; rTorrent writes the code as an i8
// (rakshasa/rtorrent v0.16.25, src/rpc/xmlrpc_tinyxml2.cc, print_xmlrpc_fault).
func TestAPIError(t *testing.T) {
	t.Parallel()

	const text = "Could not find info-hash."
	code8 := xmlMember("faultCode", "<i8>-501</i8>")
	code4 := xmlMember("faultCode", "<i4>-501</i4>")
	message := xmlMember("faultString", "<string>"+text+"</string>")
	cases := []struct {
		name      string
		fault     string
		wantCode  int
		wantText  string
		wantError string
	}{
		{"struct as rTorrent writes it", "<struct>" + code8 + message + "</struct>", -501, text, "rtorrent: " + text},
		{"struct with an i4 code", "<struct>" + code4 + message + "</struct>", -501, text, "rtorrent: " + text},
		{"struct with the text first", "<struct>" + message + code8 + "</struct>", -501, text, "rtorrent: " + text},
		{"struct with a code only", "<struct>" + code8 + "</struct>", -501, "", "rtorrent: fault code -501"},
		{"struct without members", "<struct></struct>", -1, "unknown fault", "rtorrent: unknown fault"},
		{"bare string", "<string>" + text + "</string>", 0, text, "rtorrent: " + text},
		{"bare integer", "<i8>-501</i8>", -501, "", "rtorrent: fault code -501"},
		{"no typed value", "", -1, "unknown fault", "rtorrent: unknown fault"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ts := newXMLRPCServer(t, "d.start", xmlFaultResponse(tc.fault))
			defer ts.Close()

			c, err := rtorrent.New(ts.URL)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			err = c.StartTorrent(context.Background(), "INVALID")
			var apiErr *rtorrent.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected APIError, got %T (%v)", err, err)
			}
			if apiErr.FaultCode != tc.wantCode || apiErr.FaultString != tc.wantText {
				t.Errorf("fault = %d, %q, want %d, %q", apiErr.FaultCode, apiErr.FaultString, tc.wantCode, tc.wantText)
			}
			if err.Error() != tc.wantError {
				t.Errorf("Error() = %q, want %q", err.Error(), tc.wantError)
			}
		})
	}
}

func TestWithAuth(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "admin" || p != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(xmlStringResponse("0.9.8")))
	}))
	defer ts.Close()

	c, err := rtorrent.New(ts.URL, rtorrent.WithAuth("admin", "secret"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// GetTorrents will work if auth succeeds (even though response is wrong format)
	// but we test that auth is sent. Use a system call that returns a string.
	info, err := c.GetSystemInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.ClientVersion != "0.9.8" {
		t.Errorf("ClientVersion = %q, want %q", info.ClientVersion, "0.9.8")
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
			c, err := rtorrent.New(tc.url)
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}
