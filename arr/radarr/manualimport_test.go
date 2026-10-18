package radarr_test

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

	"github.com/golusoris/goenvoy/arr/radarr"
	"github.com/golusoris/goenvoy/arr/v2"
)

// Why: shape of a Radarr v6 reply for a download its parser could not read but grab history mapped.
const radarrCandidateJSON = `[{
	"id": 7,
	"path": "/downloads/complete/Film.Deutscher.Titel.GERMAN/film.mkv",
	"relativePath": "film.mkv",
	"folderName": "Film.Deutscher.Titel.GERMAN",
	"name": "film",
	"size": 734003200,
	"movie": {"id": 12, "title": "Film"},
	"movieFileId": 0,
	"releaseGroup": "GRP",
	"quality": {"quality": {"id": 7, "name": "Bluray-1080p", "source": "bluray", "resolution": 1080, "modifier": "remux"}, "revision": {"version": 2, "real": 0, "isRepack": true}},
	"languages": [{"id": 4, "name": "German"}],
	"qualityWeight": 501,
	"downloadId": "ABCDEF0123456789",
	"customFormats": [{"id": 3, "name": "German DL"}],
	"customFormatScore": 25,
	"indexerFlags": 8,
	"rejections": [{"reason": "Not enough free space", "type": "temporary"}]
}]`

func intPtr(v int) *int { return &v }

func newManualImportServer(t *testing.T, handler http.HandlerFunc) *radarr.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := radarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("radarr.New: %v", err)
	}
	return c
}

func TestGetManualImportCandidates_QueryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     radarr.ManualImportOptions
		wantPath string
	}{
		{
			name:     "zero options still send filterExistingFiles",
			opts:     radarr.ManualImportOptions{},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false",
		},
		{
			name:     "download id",
			opts:     radarr.ManualImportOptions{DownloadID: "ABC", FilterExistingFiles: true},
			wantPath: "/api/v3/manualimport?downloadId=ABC&filterExistingFiles=true",
		},
		{
			name:     "folder is escaped",
			opts:     radarr.ManualImportOptions{Folder: "/dl/a b&c"},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false&folder=%2Fdl%2Fa+b%26c",
		},
		{
			name:     "movie id",
			opts:     radarr.ManualImportOptions{MovieID: 12},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false&movieId=12",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := newManualImportServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.RequestURI() != tc.wantPath {
					t.Errorf("RequestURI = %q, want %q", r.URL.RequestURI(), tc.wantPath)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, "[]")
			})
			got, err := c.GetManualImportCandidates(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("GetManualImportCandidates: %v", err)
			}
			if len(got) != 0 {
				t.Errorf("len = %d, want 0", len(got))
			}
		})
	}
}

func TestGetManualImportCandidates_Decode(t *testing.T) {
	t.Parallel()

	c := newManualImportServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, radarrCandidateJSON)
	})
	got, err := c.GetManualImportCandidates(context.Background(), radarr.ManualImportOptions{DownloadID: "ABCDEF0123456789"})
	if err != nil {
		t.Fatalf("GetManualImportCandidates: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	m := got[0]

	checks := []struct {
		field string
		got   any
		want  any
	}{
		{"ID", m.ID, 7},
		{"Path", m.Path, "/downloads/complete/Film.Deutscher.Titel.GERMAN/film.mkv"},
		{"RelativePath", m.RelativePath, "film.mkv"},
		{"FolderName", m.FolderName, "Film.Deutscher.Titel.GERMAN"},
		{"Name", m.Name, "film"},
		{"Size", m.Size, int64(734003200)},
		{"Movie.ID", m.Movie.ID, 12},
		{"MovieFileID", *m.MovieFileID, 0},
		{"ReleaseGroup", m.ReleaseGroup, "GRP"},
		{"Quality.Quality.ID", m.Quality.Quality.ID, 7},
		{"Quality.Quality.Modifier", m.Quality.Quality.Modifier, "remux"},
		{"Quality.Revision.Version", m.Quality.Revision.Version, 2},
		{"Languages", m.Languages, []radarr.Language{{ID: 4, Name: "German"}}},
		{"QualityWeight", m.QualityWeight, 501},
		{"DownloadID", m.DownloadID, "ABCDEF0123456789"},
		{"CustomFormats", m.CustomFormats, []arr.CustomFormatResource{{ID: 3, Name: "German DL"}}},
		{"CustomFormatScore", m.CustomFormatScore, 25},
		{"IndexerFlags", m.IndexerFlags, 8},
		{"Rejections", m.Rejections, []radarr.ImportRejection{{Reason: "Not enough free space", Type: "temporary"}}},
	}
	for _, ck := range checks {
		if !reflect.DeepEqual(ck.got, ck.want) {
			t.Errorf("%s = %#v, want %#v", ck.field, ck.got, ck.want)
		}
	}
}

func mappedCandidate() radarr.ManualImportCandidate {
	return radarr.ManualImportCandidate{
		Path:         "/dl/x/x.mkv",
		FolderName:   "x",
		Movie:        &radarr.Movie{ID: 12},
		MovieFileID:  intPtr(5),
		ReleaseGroup: "GRP",
		Quality:      radarr.QualityModel{Quality: radarr.Quality{ID: 4, Modifier: "remux"}},
		Languages:    []radarr.Language{{ID: 4, Name: "German"}},
		IndexerFlags: 8,
		DownloadID:   "ABC",
	}
}

func TestManualImportCandidate_ImportFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*radarr.ManualImportCandidate)
		wantOK bool
	}{
		{name: "mapped", mutate: func(*radarr.ManualImportCandidate) {}, wantOK: true},
		{name: "no movie", mutate: func(m *radarr.ManualImportCandidate) { m.Movie = nil }},
		{name: "movie id zero", mutate: func(m *radarr.ManualImportCandidate) { m.Movie = &radarr.Movie{} }},
		{name: "rejections are the caller's call", mutate: func(m *radarr.ManualImportCandidate) {
			m.Rejections = []radarr.ImportRejection{{Reason: "x", Type: "permanent"}}
		}, wantOK: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := mappedCandidate()
			tc.mutate(&m)
			f, ok := m.ImportFile()
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				if !reflect.DeepEqual(f, radarr.ManualImportFile{}) {
					t.Errorf("file = %#v, want zero value", f)
				}
				return
			}
			if f.MovieID != 12 {
				t.Errorf("MovieID = %d, want 12", f.MovieID)
			}
			if f.Path != m.Path || f.FolderName != m.FolderName ||
				f.ReleaseGroup != "GRP" || f.Quality.Quality.Modifier != "remux" || f.Languages[0].ID != 4 ||
				f.IndexerFlags != 8 || f.DownloadID != "ABC" {
				t.Errorf("file = %#v, want every candidate field carried over", f)
			}
		})
	}
}

func TestSendManualImport_Body(t *testing.T) {
	t.Parallel()

	want := `{
		"name": "ManualImport",
		"importMode": "auto",
		"files": [{
			"path": "/dl/x/x.mkv",
			"folderName": "x",
			"movieId": 12,
			"quality": {"quality": {"id": 4, "name": "", "source": "", "resolution": 0, "modifier": "remux"}, "revision": {"version": 0, "real": 0, "isRepack": false}},
			"languages": [{"id": 4, "name": "German"}],
			"releaseGroup": "GRP",
			"indexerFlags": 8,
			"downloadId": "ABC"
		}]
	}`
	c := newManualImportServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v3/command" {
			t.Errorf("request = %s %s, want POST /api/v3/command", r.Method, r.URL.Path)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id": 99, "name": "ManualImport", "status": "queued"}`)
	})

	f, ok := func() (radarr.ManualImportFile, bool) { m := mappedCandidate(); return m.ImportFile() }()
	if !ok {
		t.Fatal("ImportFile: not mapped")
	}
	got, err := c.SendManualImport(context.Background(), radarr.ImportModeAuto, []radarr.ManualImportFile{f})
	if err != nil {
		t.Fatalf("SendManualImport: %v", err)
	}
	if got.ID != 99 || got.Status != "queued" {
		t.Errorf("response = %#v, want id 99 queued", got)
	}
}

func TestSendManualImport_Refused(t *testing.T) {
	t.Parallel()

	file := []radarr.ManualImportFile{{Path: "/dl/x.mkv", MovieID: 1}}
	tests := []struct {
		name    string
		mode    radarr.ImportMode
		files   []radarr.ManualImportFile
		status  int
		wantErr error
		wantAPI bool
	}{
		{name: "no files", mode: radarr.ImportModeAuto, wantErr: radarr.ErrNoImportFiles},
		{name: "empty mode", mode: "", files: file, wantErr: radarr.ErrInvalidImportMode},
		{name: "unknown mode", mode: "hardlink", files: file, wantErr: radarr.ErrInvalidImportMode},
		{name: "server rejects", mode: radarr.ImportModeCopy, files: file, status: http.StatusBadRequest, wantAPI: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
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
		})
	}
}
