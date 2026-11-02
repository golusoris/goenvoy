package steamgriddb_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golusoris/goenvoy/metadata"
	"github.com/golusoris/goenvoy/metadata/game/steamgriddb"
)

func setup(t *testing.T, handler http.HandlerFunc) *steamgriddb.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return steamgriddb.New("test-key", metadata.WithBaseURL(srv.URL))
}

func TestGetGameByID(t *testing.T) {
	t.Parallel()

	c := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/games/id/12345" {
			t.Errorf("path = %q, want /games/id/12345", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q, want Bearer test-key", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    map[string]any{"id": 12345, "name": "Half-Life 2", "verified": true},
		})
	})

	game, err := c.GetGameByID(context.Background(), 12345)
	if err != nil {
		t.Fatal(err)
	}
	if game.Name != "Half-Life 2" {
		t.Errorf("Name = %q", game.Name)
	}
	if !game.Verified {
		t.Error("expected Verified to be true")
	}
}

func TestGetGameByPlatformID(t *testing.T) {
	t.Parallel()

	c := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/games/steam/220" {
			t.Errorf("path = %q, want /games/steam/220", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    map[string]any{"id": 12345, "name": "Half-Life 2"},
		})
	})

	game, err := c.GetGameByPlatformID(context.Background(), "steam", 220)
	if err != nil {
		t.Fatal(err)
	}
	if game.ID != 12345 {
		t.Errorf("ID = %d, want 12345", game.ID)
	}
}

func TestSearchGames(t *testing.T) {
	t.Parallel()

	c := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/autocomplete/half-life" {
			t.Errorf("path = %q, want /search/autocomplete/half-life", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": []map[string]any{
				{"id": 12345, "name": "Half-Life 2", "verified": true},
			},
		})
	})

	results, err := c.SearchGames(context.Background(), "half-life")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Name != "Half-Life 2" {
		t.Errorf("Name = %q", results[0].Name)
	}
}

func TestGetGrids(t *testing.T) {
	t.Parallel()

	c := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/grids/game/12345" {
			t.Errorf("path = %q, want /grids/game/12345", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": []map[string]any{
				{"id": 1, "url": "https://example.com/grid.png", "width": 600, "height": 900, "style": "alternate"},
			},
		})
	})

	images, err := c.GetGrids(context.Background(), 12345, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 {
		t.Fatalf("len(images) = %d, want 1", len(images))
	}
	if images[0].Style != "alternate" {
		t.Errorf("Style = %q", images[0].Style)
	}
}

func TestGetHeroes(t *testing.T) {
	t.Parallel()

	c := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/heroes/game/12345" {
			t.Errorf("path = %q, want /heroes/game/12345", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": []map[string]any{
				{"id": 2, "url": "https://example.com/hero.png", "width": 1920, "height": 620},
			},
		})
	})

	images, err := c.GetHeroes(context.Background(), 12345, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 {
		t.Fatalf("len(images) = %d, want 1", len(images))
	}
	if images[0].Width != 1920 {
		t.Errorf("Width = %d, want 1920", images[0].Width)
	}
}

func TestGetLogos(t *testing.T) {
	t.Parallel()

	c := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/logos/game/12345" {
			t.Errorf("path = %q, want /logos/game/12345", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": []map[string]any{
				{"id": 3, "url": "https://example.com/logo.png"},
			},
		})
	})

	images, err := c.GetLogos(context.Background(), 12345, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 {
		t.Fatalf("len(images) = %d, want 1", len(images))
	}
}

func TestGetIcons(t *testing.T) {
	t.Parallel()

	c := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/icons/game/12345" {
			t.Errorf("path = %q, want /icons/game/12345", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": []map[string]any{
				{"id": 4, "url": "https://example.com/icon.png"},
			},
		})
	})

	images, err := c.GetIcons(context.Background(), 12345, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 {
		t.Fatalf("len(images) = %d, want 1", len(images))
	}
}

func TestImageOptions(t *testing.T) {
	t.Parallel()

	c := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("styles"); got != "alternate,material" {
			t.Errorf("styles = %q, want alternate,material", got)
		}
		if got := r.URL.Query().Get("nsfw"); got != "false" {
			t.Errorf("nsfw = %q, want false", got)
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("limit = %q, want 10", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    []map[string]any{},
		})
	})

	nsfw := false
	_, err := c.GetGrids(context.Background(), 1, &steamgriddb.ImageOptions{
		Styles: []string{"alternate", "material"},
		NSFW:   &nsfw,
		Limit:  10,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	c := setup(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success":false,"errors":["Unauthorized"]}`))
	})

	_, err := c.GetGameByID(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *steamgriddb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
}

type closeErrTransport struct {
	base http.RoundTripper
	err  error
}

func (t closeErrTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp.Body = &closeErrBody{rc: resp.Body, err: t.err}
	return resp, nil
}

type closeErrBody struct {
	rc  io.ReadCloser
	err error
}

func (b *closeErrBody) Read(p []byte) (int, error) {
	return b.rc.Read(p)
}

func (b *closeErrBody) Close() error {
	cerr := b.rc.Close()
	return errors.Join(cerr, b.err)
}

func runCloseTest(t *testing.T, statusCode int, response string) {
	t.Helper()
	errBoom := errors.New("close boom")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if _, err := io.WriteString(w, response); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(srv.Close)

	client := &http.Client{
		Transport: closeErrTransport{
			base: srv.Client().Transport,
			err:  errBoom,
		},
	}
	c := steamgriddb.New("key", metadata.WithHTTPClient(client), metadata.WithBaseURL(srv.URL))

	_, err := c.GetGameByID(context.Background(), 123)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errBoom) {
		t.Fatalf("errors.Is(err, errBoom) = false; err: %v", err)
	}
	if statusCode == http.StatusOK {
		if !strings.Contains(err.Error(), "steamgriddb: close response body") {
			t.Fatalf("error %q does not contain close response body", err.Error())
		}
		return
	}
	var apiErr *steamgriddb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As(err, &apiErr) = false; err: %v", err)
	}
}

func TestCloseResponseBodyError(t *testing.T) {
	t.Parallel()

	t.Run("status_200", func(t *testing.T) {
		t.Parallel()
		runCloseTest(t, http.StatusOK, `{"success":true,"data":{"id":123}}`)
	})
	t.Run("status_500", func(t *testing.T) {
		t.Parallel()
		runCloseTest(t, http.StatusInternalServerError, `{"error":"internal error"}`)
	})
}
