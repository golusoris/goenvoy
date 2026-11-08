package launchbox_test

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/metadata"
	"github.com/golusoris/goenvoy/metadata/game/launchbox"
)

func makeTestZip(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	metadataXML := `<?xml version="1.0" encoding="utf-8"?>
<LaunchBox>
  <Game>
    <DatabaseID>1</DatabaseID>
    <Name>Super Mario Bros.</Name>
    <Platform>Nintendo Entertainment System</Platform>
    <Developer>Nintendo</Developer>
    <Publisher>Nintendo</Publisher>
    <Genres>Platform</Genres>
    <MaxPlayers>2</MaxPlayers>
    <Overview>A classic platformer.</Overview>
  </Game>
  <Game>
    <DatabaseID>2</DatabaseID>
    <Name>The Legend of Zelda</Name>
    <Platform>Nintendo Entertainment System</Platform>
    <Developer>Nintendo</Developer>
    <Publisher>Nintendo</Publisher>
  </Game>
  <Game>
    <DatabaseID>3</DatabaseID>
    <Name>Sonic the Hedgehog</Name>
    <Platform>Sega Genesis</Platform>
    <Developer>Sonic Team</Developer>
    <Publisher>Sega</Publisher>
  </Game>
  <GameAlternateName>
    <DatabaseID>1</DatabaseID>
    <AlternateNameID>100</AlternateNameID>
    <AlternateName>Super Mario Bros</AlternateName>
    <Region>Japan</Region>
  </GameAlternateName>
  <GameImage>
    <DatabaseID>1</DatabaseID>
    <FileName>Images/Nintendo Entertainment System/Super Mario Bros./Box - Front.jpg</FileName>
    <Type>Box - Front</Type>
    <Region>North America</Region>
  </GameImage>
</LaunchBox>`

	platformsXML := `<?xml version="1.0" encoding="utf-8"?>
<LaunchBox>
  <Platform>
    <Name>Nintendo Entertainment System</Name>
    <Emulated>true</Emulated>
    <Developer>Nintendo</Developer>
    <Manufacturer>Nintendo</Manufacturer>
  </Platform>
  <Platform>
    <Name>Sega Genesis</Name>
    <Emulated>true</Emulated>
    <Manufacturer>Sega</Manufacturer>
  </Platform>
</LaunchBox>`

	f1, err := w.Create("Metadata.xml")
	if err != nil {
		t.Fatal(err)
	}
	f1.Write([]byte(metadataXML))

	f2, err := w.Create("Platforms.xml")
	if err != nil {
		t.Fatal(err)
	}
	f2.Write([]byte(platformsXML))

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func setup(t *testing.T) *launchbox.Client {
	t.Helper()

	zipData := makeTestZip(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(zipData)
	}))
	t.Cleanup(srv.Close)

	c := launchbox.New(metadata.WithBaseURL(srv.URL))
	if err := c.Download(context.Background()); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestDownloadAndGameCount(t *testing.T) {
	t.Parallel()

	c := setup(t)
	if c.GameCount() != 3 {
		t.Errorf("GameCount = %d, want 3", c.GameCount())
	}
}

func TestGetGameByID(t *testing.T) {
	t.Parallel()

	c := setup(t)

	game := c.GetGameByID(1)
	if game == nil {
		t.Fatal("game not found")
	}
	if game.Name != "Super Mario Bros." {
		t.Errorf("Name = %q", game.Name)
	}
	if game.Platform != "Nintendo Entertainment System" {
		t.Errorf("Platform = %q", game.Platform)
	}
	if game.Developer != "Nintendo" {
		t.Errorf("Developer = %q", game.Developer)
	}

	if c.GetGameByID(999) != nil {
		t.Error("expected nil for unknown ID")
	}
}

func TestSearchGames(t *testing.T) {
	t.Parallel()

	c := setup(t)

	// Search by name.
	results := c.SearchGames("mario", "")
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	if results[0].DatabaseID != 1 {
		t.Errorf("DatabaseID = %d, want 1", results[0].DatabaseID)
	}

	// Search with platform filter.
	results = c.SearchGames("", "Nintendo Entertainment System")
	if len(results) != 2 {
		t.Errorf("len = %d, want 2", len(results))
	}

	// Search no match.
	results = c.SearchGames("nonexistent", "")
	if len(results) != 0 {
		t.Errorf("len = %d, want 0", len(results))
	}
}

func TestGetAlternateNames(t *testing.T) {
	t.Parallel()

	c := setup(t)

	names := c.GetAlternateNames(1)
	if len(names) != 1 {
		t.Fatalf("len = %d, want 1", len(names))
	}
	if names[0].Name != "Super Mario Bros" {
		t.Errorf("Name = %q", names[0].Name)
	}
	if names[0].Region != "Japan" {
		t.Errorf("Region = %q", names[0].Region)
	}
}

func TestGetImages(t *testing.T) {
	t.Parallel()

	c := setup(t)

	images := c.GetImages(1)
	if len(images) != 1 {
		t.Fatalf("len = %d, want 1", len(images))
	}
	if images[0].Type != "Box - Front" {
		t.Errorf("Type = %q", images[0].Type)
	}
}

func TestGetPlatforms(t *testing.T) {
	t.Parallel()

	c := setup(t)

	platforms := c.GetPlatforms()
	if len(platforms) != 2 {
		t.Fatalf("len = %d, want 2", len(platforms))
	}
	if platforms[0].Name != "Nintendo Entertainment System" {
		t.Errorf("Name = %q", platforms[0].Name)
	}
}

// runDownloadHTTPErrorTest verifies that Download returns the expected error when the server responds with an HTTP error.
func runDownloadHTTPErrorTest(t *testing.T, handler http.HandlerFunc, wantErrIs error, check func(*testing.T, error)) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c := launchbox.New(metadata.WithBaseURL(srv.URL))
	err := c.Download(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if wantErrIs != nil && !errors.Is(err, wantErrIs) {
		t.Fatalf("errors.Is(err, %v) = false; err: %v", wantErrIs, err)
	}
	if check != nil {
		check(t, err)
	}
}

type downloadTestCase struct {
	name      string
	handler   func(t *testing.T) http.HandlerFunc
	wantErrIs error
	check     func(*testing.T, error)
}

// downloadErrorCases returns test cases for TestDownloadError.
func downloadErrorCases() []downloadTestCase {
	return []downloadTestCase{
		{
			name: "read body error",
			handler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				return func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Length", "100")
					w.WriteHeader(http.StatusInternalServerError)
					if _, err := w.Write([]byte("short")); err != nil {
						t.Errorf("write body: %v", err)
					}
				}
			},
			wantErrIs: io.ErrUnexpectedEOF,
			check: func(t *testing.T, err error) {
				t.Helper()
				if !strings.Contains(err.Error(), "launchbox: HTTP 500: read body:") {
					t.Fatalf("err = %q, want containing %q", err.Error(), "launchbox: HTTP 500: read body:")
				}
			},
		},
		{
			name: "non-2xx with body",
			handler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				return func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					if _, err := w.Write([]byte("server error")); err != nil {
						t.Errorf("write body: %v", err)
					}
				}
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				var apiErr *launchbox.APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("error type = %T, want *launchbox.APIError", err)
				}
				if apiErr.StatusCode != http.StatusInternalServerError {
					t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
				}
			},
		},
	}
}

// TestDownloadError verifies error handling for HTTP failures and non-2xx status codes during download.
func TestDownloadError(t *testing.T) {
	t.Parallel()
	cases := downloadErrorCases()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runDownloadHTTPErrorTest(t, tc.handler(t), tc.wantErrIs, tc.check)
		})
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

func runCloseTest(t *testing.T, statusCode int, response []byte) {
	t.Helper()
	errBoom := errors.New("close boom")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if statusCode >= 400 {
			http.Error(w, string(response), statusCode)
			return
		}
		http.ServeContent(w, r, "Metadata.zip", time.Time{}, bytes.NewReader(response))
	}))
	t.Cleanup(srv.Close)

	client := &http.Client{
		Transport: closeErrTransport{
			base: srv.Client().Transport,
			err:  errBoom,
		},
	}
	c := launchbox.New(metadata.WithHTTPClient(client), metadata.WithBaseURL(srv.URL))

	err := c.Download(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errBoom) {
		t.Fatalf("errors.Is(err, errBoom) = false; err: %v", err)
	}
	if statusCode == http.StatusOK {
		if !strings.Contains(err.Error(), "launchbox: close response body") {
			t.Fatalf("error %q does not contain close response body", err.Error())
		}
		return
	}
	var apiErr *launchbox.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As(err, &apiErr) = false; err: %v", err)
	}
}

func TestCloseResponseBodyError(t *testing.T) {
	t.Parallel()

	t.Run("status_200", func(t *testing.T) {
		t.Parallel()
		runCloseTest(t, http.StatusOK, makeTestZip(t))
	})
	t.Run("status_500", func(t *testing.T) {
		t.Parallel()
		runCloseTest(t, http.StatusInternalServerError, []byte("internal error"))
	})
}
