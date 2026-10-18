package arr_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/arr/v2"
)

// echoServer answers every request with status and body, and records the last request body.
func echoServer(t *testing.T, status int, body string) (*arr.BaseClient, *[]byte) {
	t.Helper()
	var got []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in json.RawMessage
		if json.NewDecoder(r.Body).Decode(&in) == nil {
			got = in
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	c, err := arr.NewBaseClient(srv.URL, "k")
	if err != nil {
		t.Fatal(err)
	}
	return c, &got
}

func TestGetValue(t *testing.T) {
	t.Parallel()
	c, _ := echoServer(t, http.StatusOK, `"ext4"`)
	got, err := arr.GetValue[string](context.Background(), c, "/api/v3/filesystem/type?path=%2F")
	if err != nil || got != "ext4" {
		t.Fatalf("GetValue = %q, %v; want ext4", got, err)
	}
}

func TestGetValue_Error(t *testing.T) {
	t.Parallel()
	c, _ := echoServer(t, http.StatusNotFound, `{"message":"no"}`)
	got, err := arr.GetValue[string](context.Background(), c, "/x")
	var apiErr *arr.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound || got != "" {
		t.Fatalf("GetValue = %q, %v; want zero value and 404 APIError", got, err)
	}
}

func TestGetText(t *testing.T) {
	t.Parallel()
	c, _ := echoServer(t, http.StatusOK, "line one\nline two\n")
	got, err := c.GetText(context.Background(), "/api/v3/log/file/sonarr.txt")
	if err != nil || got != "line one\nline two\n" {
		t.Fatalf("GetText = %q, %v", got, err)
	}
}

func TestGetText_EmptyAndError(t *testing.T) {
	t.Parallel()
	empty, _ := echoServer(t, http.StatusOK, "")
	if got, err := empty.GetText(context.Background(), "/x"); err != nil || got != "" {
		t.Fatalf("GetText empty = %q, %v", got, err)
	}
	failing, _ := echoServer(t, http.StatusInternalServerError, "boom")
	var apiErr *arr.APIError
	if _, err := failing.GetText(context.Background(), "/x"); !errors.As(err, &apiErr) {
		t.Fatalf("GetText 500 err = %v; want APIError", err)
	}
}

type importMode string

type importFile struct {
	Path string `json:"path"`
}

var (
	errNoFiles = errors.New("no files")
	errMode    = errors.New("bad mode")
)

func TestSendManualImport(t *testing.T) {
	t.Parallel()
	c, sent := echoServer(t, http.StatusOK, `{"id":3,"name":"ManualImport"}`)
	files := []importFile{{Path: "/a.mkv"}}
	for _, mode := range []importMode{"auto", "move", "copy"} {
		got, err := arr.SendManualImport(context.Background(), c, "/api/v3/command", mode, files, errNoFiles, errMode)
		if err != nil || got.ID != 3 {
			t.Fatalf("mode %s: %+v, %v", mode, got, err)
		}
		want := `{"name":"ManualImport","importMode":"` + string(mode) + `","files":[{"path":"/a.mkv"}]}`
		if string(*sent) != want {
			t.Fatalf("mode %s body = %s; want %s", mode, *sent, want)
		}
	}
}

func TestSendManualImport_Refused(t *testing.T) {
	t.Parallel()
	c, sent := echoServer(t, http.StatusOK, `{}`)
	ctx := context.Background()
	if _, err := arr.SendManualImport(ctx, c, "/c", importMode("auto"), []importFile{}, errNoFiles, errMode); !errors.Is(err, errNoFiles) {
		t.Fatalf("empty files err = %v; want errNoFiles", err)
	}
	for _, mode := range []importMode{"", "Move", "hardlink"} {
		_, err := arr.SendManualImport(ctx, c, "/c", mode, []importFile{{Path: "/a"}}, errNoFiles, errMode)
		if !errors.Is(err, errMode) {
			t.Fatalf("mode %q err = %v; want errMode", mode, err)
		}
	}
	if *sent != nil {
		t.Fatalf("refused import sent a request: %s", *sent)
	}
}
