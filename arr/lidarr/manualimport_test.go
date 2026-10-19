package lidarr_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/golusoris/goenvoy/arr/lidarr"
	"github.com/golusoris/goenvoy/arr/v2"
)

const lidarrCandidateJSON = `[{
	"id": 7,
	"path": "/downloads/complete/Album/01.mp3",
	"name": "01.mp3",
	"size": 10485760,
	"artist": {"id": 12, "artistName": "Artist One"},
	"album": {"id": 34, "title": "Album One"},
	"albumReleaseId": 56,
	"tracks": [{"id": 78, "title": "Track One", "trackNumber": "1"}],
	"quality": {"quality": {"id": 7, "name": "MP3-320"}, "revision": {"version": 1, "real": 0, "isRepack": false}},
	"releaseGroup": "GRP",
	"qualityWeight": 501,
	"downloadId": "ABCDEF0123456789",
	"indexerFlags": 8,
	"rejections": [{"reason": "Not enough free space", "type": "temporary"}],
	"additionalFile": false,
	"replaceExistingFiles": true,
	"disableReleaseSwitching": false
}]`

func newManualImportServer(t *testing.T, handler http.HandlerFunc) *lidarr.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := lidarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("lidarr.New: %v", err)
	}
	return c
}

func runManualImportCandidateQueryTest(t *testing.T, opts lidarr.ManualImportOptions, wantPath string) {
	t.Helper()
	c := newManualImportServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.RequestURI() != wantPath {
			t.Errorf("RequestURI = %q, want %q", r.URL.RequestURI(), wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, "[]")
	})
	got, err := c.GetManualImportCandidates(context.Background(), opts)
	if err != nil {
		t.Fatalf("GetManualImportCandidates: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestGetManualImportCandidates_QueryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     lidarr.ManualImportOptions
		wantPath string
	}{
		{
			name:     "zero options sends filterExistingFiles and replaceExistingFiles",
			opts:     lidarr.ManualImportOptions{},
			wantPath: "/api/v1/manualimport?filterExistingFiles=false&replaceExistingFiles=false",
		},
		{
			name: "download id",
			opts: lidarr.ManualImportOptions{
				DownloadID:           "ABC",
				FilterExistingFiles:  true,
				ReplaceExistingFiles: true,
			},
			wantPath: "/api/v1/manualimport?downloadId=ABC&filterExistingFiles=true&replaceExistingFiles=true",
		},
		{
			name:     "folder is escaped",
			opts:     lidarr.ManualImportOptions{Folder: "/dl/a b&c"},
			wantPath: "/api/v1/manualimport?filterExistingFiles=false&folder=%2Fdl%2Fa+b%26c&replaceExistingFiles=false",
		},
		{
			name:     "artist id",
			opts:     lidarr.ManualImportOptions{ArtistID: 12},
			wantPath: "/api/v1/manualimport?artistId=12&filterExistingFiles=false&replaceExistingFiles=false",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runManualImportCandidateQueryTest(t, tc.opts, tc.wantPath)
		})
	}
}

func checkCandidateFields(t *testing.T, m lidarr.ManualImportCandidate) {
	t.Helper()
	checks := []struct {
		field string
		got   any
		want  any
	}{
		{"ID", m.ID, 7},
		{"Path", m.Path, "/downloads/complete/Album/01.mp3"},
		{"Name", m.Name, "01.mp3"},
		{"Size", m.Size, int64(10485760)},
		{"Artist.ID", m.Artist.ID, 12},
		{"Artist.ArtistName", m.Artist.ArtistName, "Artist One"},
		{"Album.ID", m.Album.ID, 34},
		{"Album.Title", m.Album.Title, "Album One"},
		{"AlbumReleaseID", m.AlbumReleaseID, 56},
		{"Tracks", m.Tracks, []lidarr.Track{{ID: 78, Title: "Track One", TrackNumber: "1"}}},
		{"Quality.Quality.ID", m.Quality.Quality.ID, 7},
		{"Quality.Quality.Name", m.Quality.Quality.Name, "MP3-320"},
		{"ReleaseGroup", m.ReleaseGroup, "GRP"},
		{"QualityWeight", m.QualityWeight, 501},
		{"DownloadID", m.DownloadID, "ABCDEF0123456789"},
		{"IndexerFlags", m.IndexerFlags, 8},
		{"Rejections", m.Rejections, []lidarr.ImportRejection{{Reason: "Not enough free space", Type: "temporary"}}},
		{"AdditionalFile", m.AdditionalFile, false},
		{"ReplaceExistingFiles", m.ReplaceExistingFiles, true},
		{"DisableReleaseSwitching", m.DisableReleaseSwitching, false},
	}
	for _, ck := range checks {
		if !reflect.DeepEqual(ck.got, ck.want) {
			t.Errorf("%s = %#v, want %#v", ck.field, ck.got, ck.want)
		}
	}
}

func TestGetManualImportCandidates_Decode(t *testing.T) {
	t.Parallel()

	c := newManualImportServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, lidarrCandidateJSON)
	})
	got, err := c.GetManualImportCandidates(context.Background(), lidarr.ManualImportOptions{DownloadID: "ABCDEF0123456789"})
	if err != nil {
		t.Fatalf("GetManualImportCandidates: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	checkCandidateFields(t, got[0])
}

func TestGetManualImportCandidates_APIError(t *testing.T) {
	t.Parallel()

	c := newManualImportServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	})

	_, err := c.GetManualImportCandidates(context.Background(), lidarr.ManualImportOptions{Folder: "/some/path"})
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

func mappedCandidate() lidarr.ManualImportCandidate {
	return lidarr.ManualImportCandidate{
		Path:                    "/dl/album/01.mp3",
		Artist:                  &lidarr.Artist{ID: 12},
		Album:                   &lidarr.Album{ID: 34},
		AlbumReleaseID:          56,
		Tracks:                  []lidarr.Track{{ID: 78}},
		Quality:                 lidarr.QualityModel{Quality: lidarr.Quality{ID: 4, Name: "FLAC"}},
		ReleaseGroup:            "GRP",
		IndexerFlags:            8,
		DownloadID:              "ABC",
		DisableReleaseSwitching: true,
	}
}

func assertImportFileMatchesCandidate(t *testing.T, f lidarr.ManualImportFile, m lidarr.ManualImportCandidate) {
	t.Helper()
	want := lidarr.ManualImportFile{
		Path:                    m.Path,
		ArtistID:                12,
		AlbumID:                 34,
		AlbumReleaseID:          56,
		TrackIDs:                []int{78},
		Quality:                 m.Quality,
		ReleaseGroup:            "GRP",
		IndexerFlags:            8,
		DownloadID:              "ABC",
		DisableReleaseSwitching: true,
	}
	if !reflect.DeepEqual(f, want) {
		t.Errorf("file = %#v, want %#v", f, want)
	}
}

func checkImportFileResult(t *testing.T, m lidarr.ManualImportCandidate, wantOK bool) {
	t.Helper()
	f, ok := m.ImportFile()
	if ok != wantOK {
		t.Fatalf("ok = %v, want %v", ok, wantOK)
	}
	if !ok {
		if !reflect.DeepEqual(f, lidarr.ManualImportFile{}) {
			t.Errorf("file = %#v, want zero value", f)
		}
		return
	}
	assertImportFileMatchesCandidate(t, f, m)
}

func TestManualImportCandidate_ImportFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*lidarr.ManualImportCandidate)
		wantOK bool
	}{
		{name: "mapped", mutate: func(*lidarr.ManualImportCandidate) {}, wantOK: true},
		{name: "no artist", mutate: func(m *lidarr.ManualImportCandidate) { m.Artist = nil }, wantOK: false},
		{name: "artist id zero", mutate: func(m *lidarr.ManualImportCandidate) { m.Artist = &lidarr.Artist{} }, wantOK: false},
		{name: "no album", mutate: func(m *lidarr.ManualImportCandidate) { m.Album = nil }, wantOK: false},
		{name: "album id zero", mutate: func(m *lidarr.ManualImportCandidate) { m.Album = &lidarr.Album{} }, wantOK: false},
		{name: "no tracks", mutate: func(m *lidarr.ManualImportCandidate) { m.Tracks = nil }, wantOK: false},
		{name: "track id zero", mutate: func(m *lidarr.ManualImportCandidate) { m.Tracks = []lidarr.Track{{ID: 0}} }, wantOK: false},
		{name: "rejections are the caller's call", mutate: func(m *lidarr.ManualImportCandidate) {
			m.Rejections = []lidarr.ImportRejection{{Reason: "x", Type: "permanent"}}
		}, wantOK: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := mappedCandidate()
			tc.mutate(&m)
			checkImportFileResult(t, m, tc.wantOK)
		})
	}
}

func verifyManualImportRequestBody(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.Method != http.MethodPost || r.URL.Path != "/api/v1/command" {
		t.Errorf("request = %s %s, want POST /api/v1/command", r.Method, r.URL.Path)
	}
	var gotBody, wantBody any
	if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
		t.Errorf("decode body: %v", err)
	}
	if err := json.Unmarshal([]byte(want), &wantBody); err != nil {
		t.Errorf("decode want: %v", err)
	}
	if !reflect.DeepEqual(gotBody, wantBody) {
		t.Errorf("body = %v\nwant %v", gotBody, wantBody)
	}
}

func TestSendManualImport_Body(t *testing.T) {
	t.Parallel()

	want := `{
		"name": "ManualImport",
		"importMode": "auto",
		"files": [{
			"path": "/dl/album/01.mp3",
			"artistId": 12,
			"albumId": 34,
			"albumReleaseId": 56,
			"trackIds": [78],
			"quality": {"quality": {"id": 4, "name": "FLAC"}, "revision": {"version": 0, "real": 0, "isRepack": false}},
			"releaseGroup": "GRP",
			"indexerFlags": 8,
			"downloadId": "ABC",
			"disableReleaseSwitching": true
		}]
	}`
	c := newManualImportServer(t, func(w http.ResponseWriter, r *http.Request) {
		verifyManualImportRequestBody(t, r, want)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id": 99, "name": "ManualImport", "status": "queued"}`)
	})

	f, ok := func() (lidarr.ManualImportFile, bool) { m := mappedCandidate(); return m.ImportFile() }()
	if !ok {
		t.Fatal("ImportFile: not mapped")
	}
	got, err := c.SendManualImport(context.Background(), lidarr.ImportModeAuto, []lidarr.ManualImportFile{f})
	if err != nil {
		t.Fatalf("SendManualImport: %v", err)
	}
	if got.ID != 99 || got.Status != "queued" {
		t.Errorf("response = %#v, want id 99 queued", got)
	}
}

type sendManualImportRefusedCase struct {
	name    string
	mode    lidarr.ImportMode
	files   []lidarr.ManualImportFile
	status  int
	wantErr error
	wantAPI bool
}

func runSendManualImportRefusedTest(t *testing.T, tc sendManualImportRefusedCase) {
	t.Helper()
	var called atomic.Bool
	c := newManualImportServer(t, func(w http.ResponseWriter, _ *http.Request) {
		called.Store(true)
		w.WriteHeader(tc.status)
	})
	_, err := c.SendManualImport(context.Background(), tc.mode, tc.files)
	if tc.wantAPI {
		var apiErr *arr.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != tc.status {
			t.Fatalf("err = %v, want APIError %d", err, tc.status)
		}
		return
	}
	if !errors.Is(err, tc.wantErr) {
		t.Fatalf("err = %v, want %v", err, tc.wantErr)
	}
	if called.Load() {
		t.Error("request sent, want refusal before any request")
	}
}

func TestSendManualImport_Refused(t *testing.T) {
	t.Parallel()

	file := []lidarr.ManualImportFile{{Path: "/dl/album/01.mp3", ArtistID: 1}}
	tests := []sendManualImportRefusedCase{
		{name: "no files", mode: lidarr.ImportModeAuto, wantErr: lidarr.ErrNoImportFiles},
		{name: "empty mode", mode: "", files: file, wantErr: lidarr.ErrInvalidImportMode},
		{name: "unknown mode", mode: "hardlink", files: file, wantErr: lidarr.ErrInvalidImportMode},
		{name: "server rejects", mode: lidarr.ImportModeCopy, files: file, status: http.StatusBadRequest, wantAPI: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runSendManualImportRefusedTest(t, tc)
		})
	}
}
