package readarr_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lusoris/goenvoy/arr"
	"github.com/lusoris/goenvoy/arr/readarr"
)

func newTestServer(t *testing.T, method, wantPath string, body any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			t.Errorf("method = %s, want %s", r.Method, method)
		}
		if r.Header.Get("X-Api-Key") == "" {
			t.Error("missing X-Api-Key header")
		}
		if wantPath != "" && r.URL.RequestURI() != wantPath {
			t.Errorf("path = %q, want %q", r.URL.RequestURI(), wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if body != nil {
			json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		c, err := readarr.New("http://localhost:8787", "test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		t.Parallel()
		_, err := readarr.New("://bad", "test-key")
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})
}

func TestGetAllAuthors(t *testing.T) {
	t.Parallel()

	want := []readarr.Author{
		{ID: 1, AuthorName: "Brandon Sanderson", ForeignAuthorID: "author-1"},
		{ID: 2, AuthorName: "Stephen King", ForeignAuthorID: "author-2"},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/author", want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetAllAuthors(context.Background())
	if err != nil {
		t.Fatalf("GetAllAuthors: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].AuthorName != "Brandon Sanderson" {
		t.Errorf("AuthorName = %q, want %q", got[0].AuthorName, "Brandon Sanderson")
	}
}

func TestGetAuthor(t *testing.T) {
	t.Parallel()

	want := readarr.Author{ID: 1, AuthorName: "Brandon Sanderson"}

	srv := newTestServer(t, http.MethodGet, "/api/v1/author/1", want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetAuthor(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetAuthor: %v", err)
	}
	if got.AuthorName != "Brandon Sanderson" {
		t.Errorf("AuthorName = %q, want %q", got.AuthorName, "Brandon Sanderson")
	}
}

func TestAddAuthor(t *testing.T) {
	t.Parallel()

	want := readarr.Author{ID: 3, AuthorName: "New Author", ForeignAuthorID: "abc-123"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var body readarr.Author
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.AuthorName != "New Author" {
			t.Errorf("AuthorName = %q, want %q", body.AuthorName, "New Author")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.AddAuthor(context.Background(), &readarr.Author{
		AuthorName:      "New Author",
		ForeignAuthorID: "abc-123",
	})
	if err != nil {
		t.Fatalf("AddAuthor: %v", err)
	}
	if got.ID != 3 {
		t.Errorf("ID = %d, want 3", got.ID)
	}
}

func TestDeleteAuthor(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete,
		"/api/v1/author/1?deleteFiles=true&addImportListExclusion=false",
		nil)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteAuthor(context.Background(), 1, true, false); err != nil {
		t.Fatalf("DeleteAuthor: %v", err)
	}
}

func TestLookupAuthor(t *testing.T) {
	t.Parallel()

	want := []readarr.Author{{ID: 0, AuthorName: "Stephen King"}}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/author/lookup?term=stephen+king",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.LookupAuthor(context.Background(), "stephen king")
	if err != nil {
		t.Fatalf("LookupAuthor: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}

func TestGetBooks(t *testing.T) {
	t.Parallel()

	want := []readarr.Book{
		{ID: 10, Title: "The Way of Kings", AuthorID: 1, ForeignBookID: "book-1"},
		{ID: 11, Title: "Words of Radiance", AuthorID: 1, ForeignBookID: "book-2"},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/book?authorId=1",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetBooks(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetBooks: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Title != "The Way of Kings" {
		t.Errorf("Title = %q, want %q", got[0].Title, "The Way of Kings")
	}
}

func TestGetBook(t *testing.T) {
	t.Parallel()

	want := readarr.Book{ID: 10, Title: "The Way of Kings"}

	srv := newTestServer(t, http.MethodGet, "/api/v1/book/10", want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetBook(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetBook: %v", err)
	}
	if got.Title != "The Way of Kings" {
		t.Errorf("Title = %q, want %q", got.Title, "The Way of Kings")
	}
}

func TestDeleteBook(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete,
		"/api/v1/book/10?deleteFiles=false&addImportListExclusion=true",
		nil)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteBook(context.Background(), 10, false, true); err != nil {
		t.Fatalf("DeleteBook: %v", err)
	}
}

func TestLookupBook(t *testing.T) {
	t.Parallel()

	want := []readarr.Book{{ID: 0, Title: "The Way of Kings"}}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/book/lookup?term=way+of+kings",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.LookupBook(context.Background(), "way of kings")
	if err != nil {
		t.Fatalf("LookupBook: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}

func TestGetBookFiles(t *testing.T) {
	t.Parallel()

	want := []readarr.BookFile{
		{ID: 200, AuthorID: 1, BookID: 10, Path: "/books/Sanderson/The Way of Kings.epub", Size: 5000000},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/bookfile?authorId=1",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetBookFiles(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetBookFiles: %v", err)
	}
	if got[0].Size != 5000000 {
		t.Errorf("Size = %d, want 5000000", got[0].Size)
	}
}

func TestDeleteBookFile(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v1/bookfile/200", nil)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteBookFile(context.Background(), 200); err != nil {
		t.Fatalf("DeleteBookFile: %v", err)
	}
}

func TestGetEditions(t *testing.T) {
	t.Parallel()

	want := []readarr.Edition{
		{ID: 50, BookID: 10, Title: "The Way of Kings (Hardcover)", ForeignEditionID: "ed-1"},
		{ID: 51, BookID: 10, Title: "The Way of Kings (Kindle)", ForeignEditionID: "ed-2", IsEbook: true},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/edition?bookId=10",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetEditions(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetEditions: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if !got[1].IsEbook {
		t.Error("expected second edition to be ebook")
	}
}

func TestSendCommand(t *testing.T) {
	t.Parallel()

	want := arr.CommandResponse{ID: 42, Name: "RefreshAuthor"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var cmd arr.CommandRequest
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if cmd.Name != "RefreshAuthor" {
			t.Errorf("Name = %q, want RefreshAuthor", cmd.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.SendCommand(context.Background(), arr.CommandRequest{Name: "RefreshAuthor"})
	if err != nil {
		t.Fatalf("SendCommand: %v", err)
	}
	if got.ID != 42 {
		t.Errorf("ID = %d, want 42", got.ID)
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	want := readarr.ParseResult{
		Title: "Brandon Sanderson - The Way of Kings (2010) [EPUB]",
		ParsedBookInfo: &readarr.ParsedBookInfo{
			AuthorName: "Brandon Sanderson",
			BookTitle:  "The Way of Kings",
		},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/parse?title=Brandon+Sanderson+-+The+Way+of+Kings+%282010%29+%5BEPUB%5D",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.Parse(context.Background(), "Brandon Sanderson - The Way of Kings (2010) [EPUB]")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.ParsedBookInfo.AuthorName != "Brandon Sanderson" {
		t.Errorf("AuthorName = %q, want %q", got.ParsedBookInfo.AuthorName, "Brandon Sanderson")
	}
}

func TestGetSystemStatus(t *testing.T) {
	t.Parallel()

	want := arr.StatusResponse{AppName: "Readarr", Version: "0.3.0"}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/system/status",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetSystemStatus(context.Background())
	if err != nil {
		t.Fatalf("GetSystemStatus: %v", err)
	}
	if got.AppName != "Readarr" {
		t.Errorf("AppName = %q, want %q", got.AppName, "Readarr")
	}
}

func TestGetQueue(t *testing.T) {
	t.Parallel()

	want := arr.PagingResource[arr.QueueRecord]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []arr.QueueRecord{
			{ID: 1, Title: "Brandon Sanderson - The Way of Kings"},
		},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/queue?page=1&pageSize=10",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetQueue(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetQueue: %v", err)
	}
	if got.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d, want 1", got.TotalRecords)
	}
}

func TestGetTags(t *testing.T) {
	t.Parallel()

	want := []arr.Tag{{ID: 1, Label: "fiction"}, {ID: 2, Label: "non-fiction"}}

	srv := newTestServer(t, http.MethodGet, "/api/v1/tag", want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetTags(context.Background())
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}

func TestCreateTag(t *testing.T) {
	t.Parallel()

	want := arr.Tag{ID: 3, Label: "new-tag"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var tag arr.Tag
		if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if tag.Label != "new-tag" {
			t.Errorf("Label = %q, want %q", tag.Label, "new-tag")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.CreateTag(context.Background(), "new-tag")
	if err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if got.ID != 3 {
		t.Errorf("ID = %d, want 3", got.ID)
	}
}

func TestGetMetadataProfiles(t *testing.T) {
	t.Parallel()

	want := []readarr.MetadataProfile{
		{ID: 1, Name: "Standard"},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/metadataprofile", want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetMetadataProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetMetadataProfiles: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Name != "Standard" {
		t.Errorf("Name = %q, want %q", got[0].Name, "Standard")
	}
}

func TestGetHistory(t *testing.T) {
	t.Parallel()

	want := arr.PagingResource[readarr.HistoryRecord]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []readarr.HistoryRecord{
			{ID: 5, AuthorID: 1, BookID: 10, EventType: "grabbed"},
		},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/history?page=1&pageSize=10",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetHistory(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if got.Records[0].EventType != "grabbed" {
		t.Errorf("EventType = %q, want %q", got.Records[0].EventType, "grabbed")
	}
}

func TestGetWantedMissing(t *testing.T) {
	t.Parallel()

	want := arr.PagingResource[readarr.Book]{
		Page:         1,
		PageSize:     10,
		TotalRecords: 1,
		Records: []readarr.Book{
			{ID: 10, Title: "The Way of Kings"},
		},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/wanted/missing?page=1&pageSize=10",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetWantedMissing(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetWantedMissing: %v", err)
	}
	if got.Records[0].Title != "The Way of Kings" {
		t.Errorf("Title = %q, want %q", got.Records[0].Title, "The Way of Kings")
	}
}

func TestGetSeries(t *testing.T) {
	t.Parallel()

	want := []readarr.Series{
		{ID: 1, Title: "The Stormlight Archive"},
	}

	srv := newTestServer(t, http.MethodGet,
		"/api/v1/series?authorId=1",
		want)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := c.GetSeries(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetSeries: %v", err)
	}
	if got[0].Title != "The Stormlight Archive" {
		t.Errorf("Title = %q, want %q", got[0].Title, "The Stormlight Archive")
	}
}

func TestDeleteQueueItem(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete,
		"/api/v1/queue/5?removeFromClient=true&blocklist=false",
		nil)
	defer srv.Close()

	c, err := readarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteQueueItem(context.Background(), 5, true, false); err != nil {
		t.Fatalf("DeleteQueueItem: %v", err)
	}
}

func TestErrorResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	c, err := readarr.New(srv.URL, "bad-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.GetAllAuthors(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}

	var apiErr *arr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *arr.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}
