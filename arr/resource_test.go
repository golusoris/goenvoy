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

// tagServer answers on /tag with the method and body it was sent, echoed as tags.
func tagServer(t *testing.T, status int) (*arr.BaseClient, *[]string) {
	t.Helper()
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		var in arr.Tag
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&in)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/tag/list" {
			_ = json.NewEncoder(w).Encode([]arr.Tag{{ID: 1, Label: r.Method + in.Label}, {ID: 2}})
			return
		}
		_ = json.NewEncoder(w).Encode(arr.Tag{ID: 7, Label: r.Method + in.Label})
	}))
	t.Cleanup(srv.Close)
	c, err := arr.NewBaseClient(srv.URL, "k")
	if err != nil {
		t.Fatal(err)
	}
	return c, &seen
}

func TestResourceHelpers(t *testing.T) {
	t.Parallel()
	c, seen := tagServer(t, http.StatusOK)
	ctx := context.Background()
	body := arr.Tag{Label: "x"}

	ones := map[string]func() (*arr.Tag, error){
		"GET":  func() (*arr.Tag, error) { return arr.GetOne[arr.Tag](ctx, c, "/tag") },
		"POST": func() (*arr.Tag, error) { return arr.PostOne[arr.Tag](ctx, c, "/tag", body) },
		"PUT":  func() (*arr.Tag, error) { return arr.PutOne[arr.Tag](ctx, c, "/tag", body) },
	}
	for method, call := range ones {
		got, err := call()
		if err != nil {
			t.Fatalf("%s one: %v", method, err)
		}
		want := method
		if method != "GET" {
			want += "x"
		}
		if got.ID != 7 || got.Label != want {
			t.Errorf("%s one = %+v, want ID 7 label %q", method, got, want)
		}
	}

	lists := map[string]func() ([]arr.Tag, error){
		"GET":  func() ([]arr.Tag, error) { return arr.GetList[arr.Tag](ctx, c, "/tag/list") },
		"POST": func() ([]arr.Tag, error) { return arr.PostList[arr.Tag](ctx, c, "/tag/list", body) },
		"PUT":  func() ([]arr.Tag, error) { return arr.PutList[arr.Tag](ctx, c, "/tag/list", body) },
	}
	for method, call := range lists {
		got, err := call()
		if err != nil {
			t.Fatalf("%s list: %v", method, err)
		}
		if len(got) != 2 || got[0].ID != 1 {
			t.Errorf("%s list = %+v, want two tags starting at ID 1", method, got)
		}
	}
	if len(*seen) != 6 {
		t.Errorf("requests = %v, want 6", *seen)
	}
}

func TestResourceHelpers_Error(t *testing.T) {
	t.Parallel()
	c, _ := tagServer(t, http.StatusNotFound)
	ctx := context.Background()

	one, err := arr.GetOne[arr.Tag](ctx, c, "/tag")
	var apiErr *arr.APIError
	if one != nil || !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("GetOne = %v, %v; want nil and a 404 APIError", one, err)
	}
	many, err := arr.PutList[arr.Tag](ctx, c, "/tag/list", nil)
	if many != nil || !errors.As(err, &apiErr) {
		t.Errorf("PutList = %v, %v; want nil and an APIError", many, err)
	}
}

func TestGetList_NullBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("null"))
	}))
	defer srv.Close()
	c, err := arr.NewBaseClient(srv.URL, "k")
	if err != nil {
		t.Fatal(err)
	}
	got, err := arr.GetList[arr.Tag](context.Background(), c, "/tag")
	if err != nil || got != nil {
		t.Errorf("GetList(null) = %v, %v; want nil, nil", got, err)
	}
}
