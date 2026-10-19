package readarr_test

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

	"github.com/golusoris/goenvoy/arr/readarr"
	"github.com/golusoris/goenvoy/arr/v2"
)

const readarrCandidateJSON = `[{
	"id": 7,
	"path": "/downloads/complete/Book/book.epub",
	"name": "book.epub",
	"size": 2097152,
	"author": {"id": 12, "authorName": "Author One"},
	"book": {"id": 34, "title": "Book One"},
	"foreignEditionId": "FE123",
	"quality": {"quality": {"id": 7, "name": "EPUB"}, "revision": {"version": 1, "real": 0, "isRepack": false}},
	"releaseGroup": "GRP",
	"qualityWeight": 501,
	"downloadId": "ABCDEF0123456789",
	"indexerFlags": 8,
	"rejections": [{"reason": "Not enough free space", "type": "temporary"}],
	"additionalFile": false,
	"replaceExistingFiles": true,
	"disableReleaseSwitching": false
}]`

func newManualImportServer(t *testing.T, handler http.HandlerFunc) *readarr.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("readarr.New: %v", err)
	}
	return c
}

func runManualImportCandidateQueryTest(t *testing.T, opts readarr.ManualImportOptions, wantPath string) {
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
		opts     readarr.ManualImportOptions
		wantPath string
	}{
		{
			name:     "zero options sends filterExistingFiles and replaceExistingFiles",
			opts:     readarr.ManualImportOptions{},
			wantPath: "/api/v1/manualimport?filterExistingFiles=false&replaceExistingFiles=false",
		},
		{
			name: "download id",
			opts: readarr.ManualImportOptions{
				DownloadID:           "ABC",
				FilterExistingFiles:  true,
				ReplaceExistingFiles: true,
			},
			wantPath: "/api/v1/manualimport?downloadId=ABC&filterExistingFiles=true&replaceExistingFiles=true",
		},
		{
			name:     "folder is escaped",
			opts:     readarr.ManualImportOptions{Folder: "/dl/a b&c"},
			wantPath: "/api/v1/manualimport?filterExistingFiles=false&folder=%2Fdl%2Fa+b%26c&replaceExistingFiles=false",
		},
		{
			name:     "author id",
			opts:     readarr.ManualImportOptions{AuthorID: 12},
			wantPath: "/api/v1/manualimport?authorId=12&filterExistingFiles=false&replaceExistingFiles=false",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runManualImportCandidateQueryTest(t, tc.opts, tc.wantPath)
		})
	}
}

func checkCandidateFields(t *testing.T, m readarr.ManualImportCandidate) {
	t.Helper()
	checks := []struct {
		field string
		got   any
		want  any
	}{
		{"ID", m.ID, 7},
		{"Path", m.Path, "/downloads/complete/Book/book.epub"},
		{"Name", m.Name, "book.epub"},
		{"Size", m.Size, int64(2097152)},
		{"Author.ID", m.Author.ID, 12},
		{"Author.AuthorName", m.Author.AuthorName, "Author One"},
		{"Book.ID", m.Book.ID, 34},
		{"Book.Title", m.Book.Title, "Book One"},
		{"ForeignEditionID", m.ForeignEditionID, "FE123"},
		{"Quality.Quality.ID", m.Quality.Quality.ID, 7},
		{"Quality.Quality.Name", m.Quality.Quality.Name, "EPUB"},
		{"ReleaseGroup", m.ReleaseGroup, "GRP"},
		{"QualityWeight", m.QualityWeight, 501},
		{"DownloadID", m.DownloadID, "ABCDEF0123456789"},
		{"IndexerFlags", m.IndexerFlags, 8},
		{"Rejections", m.Rejections, []readarr.ImportRejection{{Reason: "Not enough free space", Type: "temporary"}}},
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
		_, _ = io.WriteString(w, readarrCandidateJSON)
	})
	got, err := c.GetManualImportCandidates(context.Background(), readarr.ManualImportOptions{DownloadID: "ABCDEF0123456789"})
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

	_, err := c.GetManualImportCandidates(context.Background(), readarr.ManualImportOptions{Folder: "/some/path"})
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

func mappedCandidate() readarr.ManualImportCandidate {
	return readarr.ManualImportCandidate{
		Path:                    "/dl/book/book.epub",
		Author:                  &readarr.Author{ID: 12},
		Book:                    &readarr.Book{ID: 34},
		ForeignEditionID:        "FE123",
		Quality:                 readarr.QualityModel{Quality: readarr.Quality{ID: 4, Name: "EPUB"}},
		ReleaseGroup:            "GRP",
		IndexerFlags:            8,
		DownloadID:              "ABC",
		DisableReleaseSwitching: true,
	}
}

func assertImportFileMatchesCandidate(t *testing.T, f readarr.ManualImportFile, m readarr.ManualImportCandidate) {
	t.Helper()
	want := readarr.ManualImportFile{
		Path:                    m.Path,
		AuthorID:                12,
		BookID:                  34,
		ForeignEditionID:        "FE123",
		Quality:                 m.Quality,
		IndexerFlags:            8,
		DownloadID:              "ABC",
		DisableReleaseSwitching: true,
	}
	if !reflect.DeepEqual(f, want) {
		t.Errorf("file = %#v, want %#v", f, want)
	}
}

func checkImportFileResult(t *testing.T, m readarr.ManualImportCandidate, wantOK bool) {
	t.Helper()
	f, ok := m.ImportFile()
	if ok != wantOK {
		t.Fatalf("ok = %v, want %v", ok, wantOK)
	}
	if !ok {
		if !reflect.DeepEqual(f, readarr.ManualImportFile{}) {
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
		mutate func(*readarr.ManualImportCandidate)
		wantOK bool
	}{
		{name: "mapped", mutate: func(*readarr.ManualImportCandidate) {}, wantOK: true},
		{name: "no author", mutate: func(m *readarr.ManualImportCandidate) { m.Author = nil }, wantOK: false},
		{name: "author id zero", mutate: func(m *readarr.ManualImportCandidate) { m.Author = &readarr.Author{} }, wantOK: false},
		{name: "no book", mutate: func(m *readarr.ManualImportCandidate) { m.Book = nil }, wantOK: false},
		{name: "book id zero", mutate: func(m *readarr.ManualImportCandidate) { m.Book = &readarr.Book{} }, wantOK: false},
		{name: "rejections are the caller's call", mutate: func(m *readarr.ManualImportCandidate) {
			m.Rejections = []readarr.ImportRejection{{Reason: "x", Type: "permanent"}}
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
			"path": "/dl/book/book.epub",
			"authorId": 12,
			"bookId": 34,
			"foreignEditionId": "FE123",
			"quality": {"quality": {"id": 4, "name": "EPUB"}, "revision": {"version": 0, "real": 0, "isRepack": false}},
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

	f, ok := func() (readarr.ManualImportFile, bool) { m := mappedCandidate(); return m.ImportFile() }()
	if !ok {
		t.Fatal("ImportFile: not mapped")
	}
	got, err := c.SendManualImport(context.Background(), readarr.ImportModeAuto, []readarr.ManualImportFile{f})
	if err != nil {
		t.Fatalf("SendManualImport: %v", err)
	}
	if got.ID != 99 || got.Status != "queued" {
		t.Errorf("response = %#v, want id 99 queued", got)
	}
}

type sendManualImportRefusedCase struct {
	name    string
	mode    readarr.ImportMode
	files   []readarr.ManualImportFile
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

	file := []readarr.ManualImportFile{{Path: "/dl/book/book.epub", AuthorID: 1}}
	tests := []sendManualImportRefusedCase{
		{name: "no files", mode: readarr.ImportModeAuto, wantErr: readarr.ErrNoImportFiles},
		{name: "empty mode", mode: "", files: file, wantErr: readarr.ErrInvalidImportMode},
		{name: "unknown mode", mode: "hardlink", files: file, wantErr: readarr.ErrInvalidImportMode},
		{name: "server rejects", mode: readarr.ImportModeCopy, files: file, status: http.StatusBadRequest, wantAPI: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runSendManualImportRefusedTest(t, tc)
		})
	}
}
