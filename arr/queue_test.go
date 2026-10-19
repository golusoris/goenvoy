package arr_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/arr/v2"
)

// queueServer answers every request with status and body, and records the last raw query.
func queueServer(t *testing.T, status int, body string) (*arr.BaseClient, *string) {
	t.Helper()
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Path + "?" + r.URL.RawQuery
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

func TestGetQueuePage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		include map[string]bool
		want    string
	}{
		{"flags set", map[string]bool{"includeSeries": true, "includeEpisode": true}, "/api/v3/queue?includeEpisode=true&includeSeries=true&page=2&pageSize=50"},
		{"false flag left out", map[string]bool{"includeSeries": false, "includeEpisode": true}, "/api/v3/queue?includeEpisode=true&page=2&pageSize=50"},
		{"no flags", nil, "/api/v3/queue?page=2&pageSize=50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, got := queueServer(t, http.StatusOK, `{"page":2,"pageSize":50,"totalRecords":1,"records":[{"id":7}]}`)
			res, err := arr.GetQueuePage[arr.QueueRecord](context.Background(), c, "/api/v3/queue", 2, 50, tt.include)
			if err != nil {
				t.Fatalf("GetQueuePage: %v", err)
			}
			if *got != tt.want {
				t.Errorf("request = %q, want %q", *got, tt.want)
			}
			if res.TotalRecords != 1 || len(res.Records) != 1 || res.Records[0].ID != 7 {
				t.Errorf("result = %+v", res)
			}
		})
	}
}

func TestGetQueuePageError(t *testing.T) {
	t.Parallel()
	c, _ := queueServer(t, http.StatusInternalServerError, `{"message":"boom"}`)
	if _, err := arr.GetQueuePage[arr.QueueRecord](context.Background(), c, "/api/v3/queue", 1, 10, nil); err == nil {
		t.Fatal("GetQueuePage on 500 = nil error, want error")
	}
}
