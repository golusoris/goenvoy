package whisparr_test

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

	"github.com/golusoris/goenvoy/arr/v2"
	"github.com/golusoris/goenvoy/arr/whisparr"
)

// Why: shape of a Whisparr v2 candidate reply with synthetic paths.
const whisparrV2CandidateJSON = `[{
	"id": 7,
	"path": "/downloads/test/scene.mkv",
	"relativePath": "scene.mkv",
	"folderName": "scene",
	"name": "scene",
	"size": 734003200,
	"series": {"id": 12, "title": "Series Title"},
	"seasonNumber": 2,
	"episodes": [{"id": 340, "seriesId": 12}, {"id": 341, "seriesId": 12}],
	"episodeFileId": 0,
	"releaseGroup": "GRP",
	"quality": {"quality": {"id": 4, "name": "HDTV-720p", "source": "television", "resolution": 720}, "revision": {"version": 2, "real": 0, "isRepack": true}},
	"languages": [{"id": 4, "name": "English"}],
	"qualityWeight": 501,
	"downloadId": "ABC123",
	"customFormats": [{"id": 3, "name": "German DL"}],
	"customFormatScore": 25,
	"rejections": [{"reason": "Not enough free space", "type": "temporary"}]
}]`

// Why: shape of a Whisparr v3 candidate reply with synthetic paths.
const whisparrV3CandidateJSON = `[{
	"id": 7,
	"path": "/downloads/test/scene.mkv",
	"relativePath": "scene.mkv",
	"folderName": "scene",
	"name": "scene",
	"size": 734003200,
	"movie": {"id": 12, "title": "Movie Title"},
	"movieFileId": 0,
	"releaseGroup": "GRP",
	"quality": {"quality": {"id": 4, "name": "HDTV-720p", "source": "television", "resolution": 720}, "revision": {"version": 2, "real": 0, "isRepack": true}},
	"languages": [{"id": 4, "name": "English"}],
	"qualityWeight": 501,
	"downloadId": "ABC123",
	"customFormats": [{"id": 3, "name": "German DL"}],
	"customFormatScore": 25,
	"indexerFlags": 8,
	"rejections": [{"reason": "Not enough free space", "type": "temporary"}]
}]`

func intPtr(v int) *int { return &v }

func newV2ManualImportServer(t *testing.T, handler http.HandlerFunc) *whisparr.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := whisparr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.New: %v", err)
	}
	return c
}

func newV3ManualImportServer(t *testing.T, handler http.HandlerFunc) *whisparr.ClientV3 {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := whisparr.NewV3(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("whisparr.NewV3: %v", err)
	}
	return c
}

func checkV2ManualImportQuery(t *testing.T, opts whisparr.ManualImportOptions, wantPath string) {
	t.Helper()
	c := newV2ManualImportServer(t, func(w http.ResponseWriter, r *http.Request) {
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

func TestGetManualImportCandidates_QueryString_V2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     whisparr.ManualImportOptions
		wantPath string
	}{
		{
			name:     "zero options still send filterExistingFiles",
			opts:     whisparr.ManualImportOptions{},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false",
		},
		{
			name:     "download id",
			opts:     whisparr.ManualImportOptions{DownloadID: "ABC123", FilterExistingFiles: true},
			wantPath: "/api/v3/manualimport?downloadId=ABC123&filterExistingFiles=true",
		},
		{
			name:     "folder is escaped",
			opts:     whisparr.ManualImportOptions{Folder: "/downloads/test/a b&c"},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false&folder=%2Fdownloads%2Ftest%2Fa+b%26c",
		},
		{
			name:     "series and season zero are sent",
			opts:     whisparr.ManualImportOptions{SeriesID: 12, SeasonNumber: intPtr(0)},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false&seasonNumber=0&seriesId=12",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkV2ManualImportQuery(t, tc.opts, tc.wantPath)
		})
	}
}

func checkV3ManualImportQuery(t *testing.T, opts whisparr.ManualImportOptionsV3, wantPath string) {
	t.Helper()
	c := newV3ManualImportServer(t, func(w http.ResponseWriter, r *http.Request) {
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

func TestGetManualImportCandidates_QueryString_V3(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     whisparr.ManualImportOptionsV3
		wantPath string
	}{
		{
			name:     "zero options still send filterExistingFiles",
			opts:     whisparr.ManualImportOptionsV3{},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false",
		},
		{
			name:     "download id",
			opts:     whisparr.ManualImportOptionsV3{DownloadID: "ABC123", FilterExistingFiles: true},
			wantPath: "/api/v3/manualimport?downloadId=ABC123&filterExistingFiles=true",
		},
		{
			name:     "folder is escaped",
			opts:     whisparr.ManualImportOptionsV3{Folder: "/downloads/test/a b&c"},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false&folder=%2Fdownloads%2Ftest%2Fa+b%26c",
		},
		{
			name:     "movie id is sent",
			opts:     whisparr.ManualImportOptionsV3{MovieID: 12},
			wantPath: "/api/v3/manualimport?filterExistingFiles=false&movieId=12",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkV3ManualImportQuery(t, tc.opts, tc.wantPath)
		})
	}
}

func TestGetManualImportCandidates_Decode_V2(t *testing.T) {
	t.Parallel()

	c := newV2ManualImportServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, whisparrV2CandidateJSON)
	})
	got, err := c.GetManualImportCandidates(context.Background(), whisparr.ManualImportOptions{DownloadID: "ABC123"})
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
		{"Path", m.Path, "/downloads/test/scene.mkv"},
		{"RelativePath", m.RelativePath, "scene.mkv"},
		{"FolderName", m.FolderName, "scene"},
		{"Name", m.Name, "scene"},
		{"Size", m.Size, int64(734003200)},
		{"Series.ID", m.Series.ID, 12},
		{"SeasonNumber", *m.SeasonNumber, 2},
		{"Episodes", []int{m.Episodes[0].ID, m.Episodes[1].ID}, []int{340, 341}},
		{"EpisodeFileID", *m.EpisodeFileID, 0},
		{"ReleaseGroup", m.ReleaseGroup, "GRP"},
		{"Quality.Quality.ID", m.Quality.Quality.ID, 4},
		{"Quality.Quality.Name", m.Quality.Quality.Name, "HDTV-720p"},
		{"Quality.Revision.Version", m.Quality.Revision.Version, 2},
		{"Quality.Revision.IsRepack", m.Quality.Revision.IsRepack, true},
		{"Languages", m.Languages, []whisparr.Language{{ID: 4, Name: "English"}}},
		{"QualityWeight", m.QualityWeight, 501},
		{"DownloadID", m.DownloadID, "ABC123"},
		{"CustomFormats", m.CustomFormats, []arr.CustomFormatResource{{ID: 3, Name: "German DL"}}},
		{"CustomFormatScore", m.CustomFormatScore, 25},
		{"Rejections", m.Rejections, []whisparr.ImportRejection{{Reason: "Not enough free space", Type: "temporary"}}},
	}
	for _, ck := range checks {
		if !reflect.DeepEqual(ck.got, ck.want) {
			t.Errorf("%s = %#v, want %#v", ck.field, ck.got, ck.want)
		}
	}
}

func TestGetManualImportCandidates_Decode_V3(t *testing.T) {
	t.Parallel()

	c := newV3ManualImportServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, whisparrV3CandidateJSON)
	})
	got, err := c.GetManualImportCandidates(context.Background(), whisparr.ManualImportOptionsV3{DownloadID: "ABC123"})
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
		{"Path", m.Path, "/downloads/test/scene.mkv"},
		{"RelativePath", m.RelativePath, "scene.mkv"},
		{"FolderName", m.FolderName, "scene"},
		{"Name", m.Name, "scene"},
		{"Size", m.Size, int64(734003200)},
		{"Movie.ID", m.Movie.ID, 12},
		{"MovieFileID", *m.MovieFileID, 0},
		{"ReleaseGroup", m.ReleaseGroup, "GRP"},
		{"Quality.Quality.ID", m.Quality.Quality.ID, 4},
		{"Quality.Quality.Name", m.Quality.Quality.Name, "HDTV-720p"},
		{"Quality.Revision.Version", m.Quality.Revision.Version, 2},
		{"Quality.Revision.IsRepack", m.Quality.Revision.IsRepack, true},
		{"Languages", m.Languages, []whisparr.Language{{ID: 4, Name: "English"}}},
		{"QualityWeight", m.QualityWeight, 501},
		{"DownloadID", m.DownloadID, "ABC123"},
		{"CustomFormats", m.CustomFormats, []arr.CustomFormatResource{{ID: 3, Name: "German DL"}}},
		{"CustomFormatScore", m.CustomFormatScore, 25},
		{"IndexerFlags", m.IndexerFlags, 8},
		{"Rejections", m.Rejections, []whisparr.ImportRejection{{Reason: "Not enough free space", Type: "temporary"}}},
	}
	for _, ck := range checks {
		if !reflect.DeepEqual(ck.got, ck.want) {
			t.Errorf("%s = %#v, want %#v", ck.field, ck.got, ck.want)
		}
	}
}

func mappedCandidateV2() whisparr.ManualImportCandidate {
	return whisparr.ManualImportCandidate{
		Path:          "/downloads/test/scene.mkv",
		FolderName:    "scene",
		Series:        &whisparr.Series{ID: 12},
		Episodes:      []whisparr.Episode{{ID: 340}, {ID: 341}},
		EpisodeFileID: intPtr(5),
		ReleaseGroup:  "GRP",
		Quality:       whisparr.QualityModel{Quality: whisparr.Quality{ID: 4}},
		Languages:     []whisparr.Language{{ID: 4, Name: "English"}},
		DownloadID:    "ABC123",
	}
}

func assertImportFileMappingV2(t *testing.T, f whisparr.ManualImportFile, m whisparr.ManualImportCandidate) {
	t.Helper()
	if f.SeriesID != 12 || len(f.EpisodeIDs) != len(m.Episodes) || f.EpisodeIDs[0] != 340 {
		t.Errorf("mapping = %d %v, want 12 and the candidate's episodes", f.SeriesID, f.EpisodeIDs)
	}
}

func assertImportFileFieldsV2(t *testing.T, f whisparr.ManualImportFile, m whisparr.ManualImportCandidate) {
	t.Helper()
	if f.Path != m.Path || f.FolderName != m.FolderName || *f.EpisodeFileID != 5 {
		t.Errorf("file = %#v, want candidate fields carried over", f)
		return
	}
	if f.ReleaseGroup != "GRP" || f.Quality.Quality.ID != 4 || f.Languages[0].ID != 4 || f.DownloadID != "ABC123" {
		t.Errorf("file = %#v, want candidate fields carried over", f)
	}
}

func checkV2CandidateImportFile(t *testing.T, mutate func(*whisparr.ManualImportCandidate), wantOK bool) {
	t.Helper()
	m := mappedCandidateV2()
	mutate(&m)
	f, ok := m.ImportFile()
	if ok != wantOK {
		t.Fatalf("ok = %v, want %v", ok, wantOK)
	}
	if !ok {
		if !reflect.DeepEqual(f, whisparr.ManualImportFile{}) {
			t.Errorf("file = %#v, want zero value", f)
		}
		return
	}
	assertImportFileMappingV2(t, f, m)
	assertImportFileFieldsV2(t, f, m)
}

func TestManualImportCandidate_ImportFile_V2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*whisparr.ManualImportCandidate)
		wantOK bool
	}{
		{name: "mapped", mutate: func(*whisparr.ManualImportCandidate) {}, wantOK: true},
		{name: "no series", mutate: func(m *whisparr.ManualImportCandidate) { m.Series = nil }},
		{name: "series id zero", mutate: func(m *whisparr.ManualImportCandidate) { m.Series = &whisparr.Series{} }},
		{name: "no episodes", mutate: func(m *whisparr.ManualImportCandidate) { m.Episodes = nil }},
		{name: "one episode id zero", mutate: func(m *whisparr.ManualImportCandidate) { m.Episodes[1].ID = 0 }},
		{name: "single episode", mutate: func(m *whisparr.ManualImportCandidate) { m.Episodes = m.Episodes[:1] }, wantOK: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkV2CandidateImportFile(t, tc.mutate, tc.wantOK)
		})
	}
}

func mappedCandidateV3() whisparr.ManualImportCandidateV3 {
	return whisparr.ManualImportCandidateV3{
		Path:         "/downloads/test/scene.mkv",
		FolderName:   "scene",
		Movie:        &whisparr.Movie{ID: 12},
		ReleaseGroup: "GRP",
		Quality:      whisparr.QualityModel{Quality: whisparr.Quality{ID: 4}},
		Languages:    []whisparr.Language{{ID: 4, Name: "English"}},
		IndexerFlags: 8,
		DownloadID:   "ABC123",
	}
}

func assertImportFileFieldsV3(t *testing.T, f whisparr.ManualImportFileV3, m whisparr.ManualImportCandidateV3) {
	t.Helper()
	if f.Path != m.Path || f.FolderName != m.FolderName || f.MovieID != 12 {
		t.Errorf("file = %#v, want candidate fields carried over", f)
		return
	}
	if f.ReleaseGroup != "GRP" || f.Quality.Quality.ID != 4 || f.Languages[0].ID != 4 {
		t.Errorf("file = %#v, want candidate fields carried over", f)
		return
	}
	if f.IndexerFlags != 8 || f.DownloadID != "ABC123" {
		t.Errorf("file = %#v, want candidate fields carried over", f)
	}
}

func checkV3CandidateImportFile(t *testing.T, mutate func(*whisparr.ManualImportCandidateV3), wantOK bool) {
	t.Helper()
	m := mappedCandidateV3()
	mutate(&m)
	f, ok := m.ImportFile()
	if ok != wantOK {
		t.Fatalf("ok = %v, want %v", ok, wantOK)
	}
	if !ok {
		if !reflect.DeepEqual(f, whisparr.ManualImportFileV3{}) {
			t.Errorf("file = %#v, want zero value", f)
		}
		return
	}
	assertImportFileFieldsV3(t, f, m)
}

func TestManualImportCandidateV3_ImportFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*whisparr.ManualImportCandidateV3)
		wantOK bool
	}{
		{name: "mapped", mutate: func(*whisparr.ManualImportCandidateV3) {}, wantOK: true},
		{name: "no movie", mutate: func(m *whisparr.ManualImportCandidateV3) { m.Movie = nil }},
		{name: "movie id zero", mutate: func(m *whisparr.ManualImportCandidateV3) { m.Movie = &whisparr.Movie{} }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkV3CandidateImportFile(t, tc.mutate, tc.wantOK)
		})
	}
}

func assertManualImportBody(t *testing.T, r *http.Request, wantJSON string) {
	t.Helper()
	if r.Method != http.MethodPost || r.URL.Path != "/api/v3/command" {
		t.Errorf("request = %s %s, want POST /api/v3/command", r.Method, r.URL.Path)
	}
	var gotBody, wantBody any
	if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
		t.Errorf("decode body: %v", err)
	}
	if err := json.Unmarshal([]byte(wantJSON), &wantBody); err != nil {
		t.Errorf("decode want: %v", err)
	}
	if !reflect.DeepEqual(gotBody, wantBody) {
		t.Errorf("body = %v\nwant %v", gotBody, wantBody)
	}
}

func TestSendManualImport_Body_V2(t *testing.T) {
	t.Parallel()

	want := `{
		"name": "ManualImport",
		"importMode": "auto",
		"files": [{
			"path": "/downloads/test/scene.mkv",
			"folderName": "scene",
			"seriesId": 12,
			"episodeIds": [340, 341],
			"episodeFileId": 5,
			"quality": {"quality": {"id": 4, "name": "", "source": "", "resolution": 0}, "revision": {"version": 0, "real": 0, "isRepack": false}},
			"languages": [{"id": 4, "name": "English"}],
			"releaseGroup": "GRP",
			"downloadId": "ABC123"
		}]
	}`
	c := newV2ManualImportServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertManualImportBody(t, r, want)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id": 99, "name": "ManualImport", "status": "queued"}`)
	})

	m := mappedCandidateV2()
	f, ok := m.ImportFile()
	if !ok {
		t.Fatal("ImportFile: not mapped")
	}
	got, err := c.SendManualImport(context.Background(), whisparr.ImportModeAuto, []whisparr.ManualImportFile{f})
	if err != nil {
		t.Fatalf("SendManualImport: %v", err)
	}
	if got.ID != 99 || got.Status != "queued" {
		t.Errorf("response = %#v, want id 99 queued", got)
	}
}

func TestSendManualImport_Body_V3(t *testing.T) {
	t.Parallel()

	want := `{
		"name": "ManualImport",
		"importMode": "auto",
		"files": [{
			"path": "/downloads/test/scene.mkv",
			"folderName": "scene",
			"movieId": 12,
			"quality": {"quality": {"id": 4, "name": "", "source": "", "resolution": 0}, "revision": {"version": 0, "real": 0, "isRepack": false}},
			"languages": [{"id": 4, "name": "English"}],
			"releaseGroup": "GRP",
			"indexerFlags": 8,
			"downloadId": "ABC123"
		}]
	}`
	c := newV3ManualImportServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertManualImportBody(t, r, want)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id": 100, "name": "ManualImport", "status": "queued"}`)
	})

	m := mappedCandidateV3()
	f, ok := m.ImportFile()
	if !ok {
		t.Fatal("ImportFile: not mapped")
	}
	got, err := c.SendManualImport(context.Background(), whisparr.ImportModeAuto, []whisparr.ManualImportFileV3{f})
	if err != nil {
		t.Fatalf("SendManualImport: %v", err)
	}
	if got.ID != 100 || got.Status != "queued" {
		t.Errorf("response = %#v, want id 100 queued", got)
	}
}

func TestSendManualImport_ZeroIndexerFlags_V3(t *testing.T) {
	t.Parallel()

	want := `{
		"name": "ManualImport",
		"importMode": "move",
		"files": [{
			"path": "/downloads/test/scene.mkv",
			"movieId": 12,
			"quality": {"quality": {"id": 0, "name": "", "source": "", "resolution": 0}, "revision": {"version": 0, "real": 0, "isRepack": false}},
			"languages": null,
			"indexerFlags": 0
		}]
	}`
	c := newV3ManualImportServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertManualImportBody(t, r, want)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id": 101, "name": "ManualImport", "status": "queued"}`)
	})

	// Why: upstream IndexerFlags is a non-nullable int, so 0 must still be sent.
	f := whisparr.ManualImportFileV3{Path: "/downloads/test/scene.mkv", MovieID: 12}
	if _, err := c.SendManualImport(context.Background(), whisparr.ImportModeMove, []whisparr.ManualImportFileV3{f}); err != nil {
		t.Fatalf("SendManualImport: %v", err)
	}
}

func checkV2SendRefusal(t *testing.T, mode whisparr.ImportMode, files []whisparr.ManualImportFile, status int, wantErr error, wantAPI bool) {
	t.Helper()
	var called atomic.Bool
	c := newV2ManualImportServer(t, func(w http.ResponseWriter, _ *http.Request) {
		called.Store(true)
		w.WriteHeader(status)
	})
	_, err := c.SendManualImport(context.Background(), mode, files)
	if wantAPI {
		var apiErr *arr.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != status {
			t.Fatalf("err = %v, want APIError %d", err, status)
		}
		return
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if called.Load() {
		t.Error("request sent, want refusal before any request")
	}
}

func TestSendManualImport_Refused_V2(t *testing.T) {
	t.Parallel()

	file := []whisparr.ManualImportFile{{Path: "/downloads/test/scene.mkv", SeriesID: 1, EpisodeIDs: []int{1}}}
	tests := []struct {
		name    string
		mode    whisparr.ImportMode
		files   []whisparr.ManualImportFile
		status  int
		wantErr error
		wantAPI bool
	}{
		{name: "no files", mode: whisparr.ImportModeAuto, wantErr: whisparr.ErrNoImportFiles},
		{name: "empty mode", mode: "", files: file, wantErr: whisparr.ErrInvalidImportMode},
		{name: "unknown mode", mode: "hardlink", files: file, wantErr: whisparr.ErrInvalidImportMode},
		{name: "server 500", mode: whisparr.ImportModeCopy, files: file, status: http.StatusInternalServerError, wantAPI: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkV2SendRefusal(t, tc.mode, tc.files, tc.status, tc.wantErr, tc.wantAPI)
		})
	}
}

func checkV3SendRefusal(t *testing.T, mode whisparr.ImportMode, files []whisparr.ManualImportFileV3, status int, wantErr error, wantAPI bool) {
	t.Helper()
	var called atomic.Bool
	c := newV3ManualImportServer(t, func(w http.ResponseWriter, _ *http.Request) {
		called.Store(true)
		w.WriteHeader(status)
	})
	_, err := c.SendManualImport(context.Background(), mode, files)
	if wantAPI {
		var apiErr *arr.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != status {
			t.Fatalf("err = %v, want APIError %d", err, status)
		}
		return
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if called.Load() {
		t.Error("request sent, want refusal before any request")
	}
}

func TestSendManualImport_Refused_V3(t *testing.T) {
	t.Parallel()

	file := []whisparr.ManualImportFileV3{{Path: "/downloads/test/scene.mkv", MovieID: 1}}
	tests := []struct {
		name    string
		mode    whisparr.ImportMode
		files   []whisparr.ManualImportFileV3
		status  int
		wantErr error
		wantAPI bool
	}{
		{name: "no files", mode: whisparr.ImportModeAuto, wantErr: whisparr.ErrNoImportFiles},
		{name: "empty mode", mode: "", files: file, wantErr: whisparr.ErrInvalidImportMode},
		{name: "unknown mode", mode: "hardlink", files: file, wantErr: whisparr.ErrInvalidImportMode},
		{name: "server 500", mode: whisparr.ImportModeCopy, files: file, status: http.StatusInternalServerError, wantAPI: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkV3SendRefusal(t, tc.mode, tc.files, tc.status, tc.wantErr, tc.wantAPI)
		})
	}
}

func TestGetManualImportCandidates_Negative_V2(t *testing.T) {
	t.Parallel()

	c := newV2ManualImportServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error": "internal server error"}`)
	})
	_, err := c.GetManualImportCandidates(context.Background(), whisparr.ManualImportOptions{SeriesID: 1})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *arr.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("err = %v, want *arr.APIError 500", err)
	}
}

func TestGetManualImportCandidates_Negative_V3(t *testing.T) {
	t.Parallel()

	c := newV3ManualImportServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error": "internal server error"}`)
	})
	_, err := c.GetManualImportCandidates(context.Background(), whisparr.ManualImportOptionsV3{MovieID: 1})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *arr.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("err = %v, want *arr.APIError 500", err)
	}
}
