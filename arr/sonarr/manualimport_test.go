package sonarr_test

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

	"github.com/golusoris/goenvoy/arr/sonarr"
	"github.com/golusoris/goenvoy/arr/v2"
)

// Why: shape of a Sonarr v4 reply for a download its parser could not read but grab history mapped.
const sonarrCandidateJSON = `[{
	"id": 7,
	"path": "/downloads/complete/Serie.Folgentitel.GERMAN/serie.mkv",
	"relativePath": "serie.mkv",
	"folderName": "Serie.Folgentitel.GERMAN",
	"name": "serie",
	"size": 734003200,
	"series": {"id": 12, "title": "Serie"},
	"seasonNumber": 2,
	"episodes": [{"id": 340, "seriesId": 12}, {"id": 341, "seriesId": 12}],
	"episodeFileId": 0,
	"releaseGroup": "GRP",
	"quality": {"quality": {"id": 4, "name": "HDTV-720p", "source": "television", "resolution": 720}, "revision": {"version": 2, "real": 0, "isRepack": true}},
	"languages": [{"id": 4, "name": "German"}],
	"qualityWeight": 501,
	"downloadId": "ABCDEF0123456789",
	"customFormats": [{"id": 3, "name": "German DL"}],
	"customFormatScore": 25,
	"indexerFlags": 8,
	"releaseType": "multiEpisode",
	"rejections": [{"reason": "Not enough free space", "type": "temporary"}]
}]`

func intPtr(v int) *int { return &v }

func newManualImportServer(t *testing.T, handler http.HandlerFunc) *sonarr.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("sonarr.New: %v", err)
	}
	return c
}

func TestGetManualImportCandidates_QueryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     sonarr.ManualImportOptions
		wantPath string
	}{
		{
			name:     "zero options still send filterExistingFiles",
			opts:     sonarr.ManualImportOptions{},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false",
		},
		{
			name:     "download id",
			opts:     sonarr.ManualImportOptions{DownloadID: "ABC", FilterExistingFiles: true},
			wantPath: "/api/v3/manualimport?downloadId=ABC&filterExistingFiles=true",
		},
		{
			name:     "folder is escaped",
			opts:     sonarr.ManualImportOptions{Folder: "/dl/a b&c"},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false&folder=%2Fdl%2Fa+b%26c",
		},
		{
			name:     "series and season zero are sent",
			opts:     sonarr.ManualImportOptions{SeriesID: 12, SeasonNumber: intPtr(0)},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false&seasonNumber=0&seriesId=12",
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
		_, _ = io.WriteString(w, sonarrCandidateJSON)
	})
	got, err := c.GetManualImportCandidates(context.Background(), sonarr.ManualImportOptions{DownloadID: "ABCDEF0123456789"})
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
		{"Path", m.Path, "/downloads/complete/Serie.Folgentitel.GERMAN/serie.mkv"},
		{"RelativePath", m.RelativePath, "serie.mkv"},
		{"FolderName", m.FolderName, "Serie.Folgentitel.GERMAN"},
		{"Name", m.Name, "serie"},
		{"Size", m.Size, int64(734003200)},
		{"Series.ID", m.Series.ID, 12},
		{"SeasonNumber", *m.SeasonNumber, 2},
		{"Episodes", []int{m.Episodes[0].ID, m.Episodes[1].ID}, []int{340, 341}},
		{"EpisodeFileID", *m.EpisodeFileID, 0},
		{"ReleaseGroup", m.ReleaseGroup, "GRP"},
		{"Quality.Quality.ID", m.Quality.Quality.ID, 4},
		{"Quality.Revision.Version", m.Quality.Revision.Version, 2},
		{"Languages", m.Languages, []sonarr.Language{{ID: 4, Name: "German"}}},
		{"QualityWeight", m.QualityWeight, 501},
		{"DownloadID", m.DownloadID, "ABCDEF0123456789"},
		{"CustomFormats", m.CustomFormats, []arr.CustomFormatResource{{ID: 3, Name: "German DL"}}},
		{"CustomFormatScore", m.CustomFormatScore, 25},
		{"IndexerFlags", m.IndexerFlags, 8},
		{"ReleaseType", m.ReleaseType, "multiEpisode"},
		{"Rejections", m.Rejections, []sonarr.ImportRejection{{Reason: "Not enough free space", Type: "temporary"}}},
	}
	for _, ck := range checks {
		if !reflect.DeepEqual(ck.got, ck.want) {
			t.Errorf("%s = %#v, want %#v", ck.field, ck.got, ck.want)
		}
	}
}

func mappedCandidate() sonarr.ManualImportCandidate {
	return sonarr.ManualImportCandidate{
		Path:          "/dl/x/x.mkv",
		FolderName:    "x",
		Series:        &sonarr.Series{ID: 12},
		Episodes:      []sonarr.Episode{{ID: 340}, {ID: 341}},
		EpisodeFileID: intPtr(5),
		ReleaseGroup:  "GRP",
		Quality:       sonarr.QualityModel{Quality: sonarr.Quality{ID: 4}},
		Languages:     []sonarr.Language{{ID: 4, Name: "German"}},
		IndexerFlags:  8,
		ReleaseType:   "multiEpisode",
		DownloadID:    "ABC",
	}
}

func TestManualImportCandidate_ImportFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*sonarr.ManualImportCandidate)
		wantOK bool
	}{
		{name: "mapped", mutate: func(*sonarr.ManualImportCandidate) {}, wantOK: true},
		{name: "no series", mutate: func(m *sonarr.ManualImportCandidate) { m.Series = nil }},
		{name: "series id zero", mutate: func(m *sonarr.ManualImportCandidate) { m.Series = &sonarr.Series{} }},
		{name: "no episodes", mutate: func(m *sonarr.ManualImportCandidate) { m.Episodes = nil }},
		{name: "one episode id zero", mutate: func(m *sonarr.ManualImportCandidate) { m.Episodes[1].ID = 0 }},
		{name: "single episode", mutate: func(m *sonarr.ManualImportCandidate) { m.Episodes = m.Episodes[:1] }, wantOK: true},
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
				if !reflect.DeepEqual(f, sonarr.ManualImportFile{}) {
					t.Errorf("file = %#v, want zero value", f)
				}
				return
			}
			if f.SeriesID != 12 || f.EpisodeIDs[0] != 340 || len(f.EpisodeIDs) != len(m.Episodes) {
				t.Errorf("mapping = %d %v, want 12 and the candidate's episodes", f.SeriesID, f.EpisodeIDs)
			}
			if f.Path != m.Path || f.FolderName != m.FolderName || *f.EpisodeFileID != 5 ||
				f.ReleaseGroup != "GRP" || f.Quality.Quality.ID != 4 || f.Languages[0].ID != 4 ||
				f.IndexerFlags != 8 || f.ReleaseType != "multiEpisode" || f.DownloadID != "ABC" {
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
			"seriesId": 12,
			"episodeIds": [340, 341],
			"episodeFileId": 5,
			"quality": {"quality": {"id": 4, "name": "", "source": "", "resolution": 0}, "revision": {"version": 0, "real": 0, "isRepack": false}},
			"languages": [{"id": 4, "name": "German"}],
			"releaseGroup": "GRP",
			"indexerFlags": 8,
			"releaseType": "multiEpisode",
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

	f, ok := func() (sonarr.ManualImportFile, bool) { m := mappedCandidate(); return m.ImportFile() }()
	if !ok {
		t.Fatal("ImportFile: not mapped")
	}
	got, err := c.SendManualImport(context.Background(), sonarr.ImportModeAuto, []sonarr.ManualImportFile{f})
	if err != nil {
		t.Fatalf("SendManualImport: %v", err)
	}
	if got.ID != 99 || got.Status != "queued" {
		t.Errorf("response = %#v, want id 99 queued", got)
	}
}

func TestSendManualImport_Refused(t *testing.T) {
	t.Parallel()

	file := []sonarr.ManualImportFile{{Path: "/dl/x.mkv", SeriesID: 1, EpisodeIDs: []int{1}}}
	tests := []struct {
		name    string
		mode    sonarr.ImportMode
		files   []sonarr.ManualImportFile
		status  int
		wantErr error
		wantAPI bool
	}{
		{name: "no files", mode: sonarr.ImportModeAuto, wantErr: sonarr.ErrNoImportFiles},
		{name: "empty mode", mode: "", files: file, wantErr: sonarr.ErrInvalidImportMode},
		{name: "unknown mode", mode: "hardlink", files: file, wantErr: sonarr.ErrInvalidImportMode},
		{name: "server rejects", mode: sonarr.ImportModeCopy, files: file, status: http.StatusBadRequest, wantAPI: true},
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
