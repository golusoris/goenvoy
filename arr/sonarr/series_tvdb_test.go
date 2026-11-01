package sonarr_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/golusoris/goenvoy/arr/sonarr"
)

func TestGetSeriesByTVDBID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body []sonarr.Series
		want int
	}{
		{name: "found", body: []sonarr.Series{{ID: 7, Title: "Breaking Bad", TvdbID: 81189}}, want: 1},
		{name: "not in library", body: []sonarr.Series{}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkSeriesByTVDBID(t, tt.body, tt.want)
		})
	}
}

func checkSeriesByTVDBID(t *testing.T, body []sonarr.Series, want int) {
	t.Helper()
	srv := newTestServer(t, http.MethodGet, "/api/v3/series?tvdbId=81189", body)
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got, err := c.GetSeriesByTVDBID(context.Background(), 81189)
	if err != nil {
		t.Fatalf("GetSeriesByTVDBID: %v", err)
	}
	if len(got) != want {
		t.Fatalf("len = %d, want %d", len(got), want)
	}
	if want == 1 && (got[0].ID != 7 || got[0].TvdbID != 81189) {
		t.Errorf("series = %+v, want ID 7 TVDB 81189", got[0])
	}
}

func TestGetSeriesByTVDBIDRejectsInvalidID(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, id := range []int{0, -1} {
		if _, err := c.GetSeriesByTVDBID(context.Background(), id); !errors.Is(err, sonarr.ErrInvalidTVDBID) {
			t.Errorf("id %d: err = %v, want ErrInvalidTVDBID", id, err)
		}
	}
	if n := calls.Load(); n != 0 {
		t.Errorf("requests sent = %d, want 0", n)
	}
}

func TestGetSeriesByTVDBIDAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c, err := sonarr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetSeriesByTVDBID(context.Background(), 1)
	checkAPIError(t, err, http.StatusUnauthorized)
}
