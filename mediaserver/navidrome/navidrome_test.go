package navidrome

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// helper wraps a responseBody in the subsonic-response envelope.
func subsonicJSON(t *testing.T, w http.ResponseWriter, rb *responseBody) {
	t.Helper()
	rb.Status = "ok"
	rb.Version = "1.16.1"
	json.NewEncoder(w).Encode(subsonicResponse{Response: *rb})
}

func newTestServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c, err := New(ts.URL, "admin", "password")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestPing(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("u") != "admin" {
			http.Error(w, "bad auth", http.StatusUnauthorized)
			return
		}
		subsonicJSON(t, w, &responseBody{})
	})

	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestGetArtists(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			Artists: &ArtistsID3{
				Index: []IndexID3{{Name: "A", Artist: []ArtistID3{{ID: "1", Name: "ABBA"}}}},
			},
		})
	})

	artists, err := c.GetArtists(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(artists.Index) != 1 || artists.Index[0].Artist[0].Name != "ABBA" {
		t.Fatalf("unexpected artists: %+v", artists)
	}
}

func TestGetArtist(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			Artist: &ArtistID3{ID: "1", Name: "Radiohead", AlbumCount: 9},
		})
	})

	a, err := c.GetArtist(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "Radiohead" {
		t.Fatalf("unexpected artist: %+v", a)
	}
}

func TestGetAlbum(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			Album: &AlbumID3{ID: "a1", Name: "OK Computer", SongCount: 12},
		})
	})

	album, err := c.GetAlbum(context.Background(), "a1")
	if err != nil {
		t.Fatal(err)
	}
	if album.Name != "OK Computer" {
		t.Fatalf("unexpected album: %+v", album)
	}
}

func TestGetAlbumList2(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			AlbumList2: &AlbumList2{Album: []AlbumID3{{ID: "a1", Name: "Abbey Road"}}},
		})
	})

	albums, err := c.GetAlbumList2(context.Background(), "newest", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(albums) != 1 || albums[0].Name != "Abbey Road" {
		t.Fatalf("unexpected albums: %+v", albums)
	}
}

func TestGetSong(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			Song: &Song{ID: "s1", Title: "Paranoid Android"},
		})
	})

	song, err := c.GetSong(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if song.Title != "Paranoid Android" {
		t.Fatalf("unexpected song: %+v", song)
	}
}

func TestGetRandomSongs(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			RandomSongs: &Songs{Song: []Song{{ID: "s1", Title: "Creep"}}},
		})
	})

	songs, err := c.GetRandomSongs(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(songs) != 1 || songs[0].Title != "Creep" {
		t.Fatalf("unexpected songs: %+v", songs)
	}
}

func TestGetTopSongs(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/getTopSongs" {
			t.Errorf("path = %q, want /rest/getTopSongs", r.URL.Path)
		}
		if got := r.URL.Query().Get("artist"); got != "Radiohead" {
			t.Errorf("artist = %q, want Radiohead", got)
		}
		if got := r.URL.Query().Get("count"); got != "2" {
			t.Errorf("count = %q, want 2", got)
		}
		subsonicJSON(t, w, &responseBody{
			TopSongs: &TopSongs{Song: []Song{{ID: "s1", Title: "Creep"}}},
		})
	})

	songs, err := c.GetTopSongs(context.Background(), "Radiohead", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(songs) != 1 || songs[0].Title != "Creep" {
		t.Fatalf("unexpected songs: %+v", songs)
	}
}

func TestSearch3(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			SearchResult: &SearchResult3{
				Artist: []ArtistID3{{ID: "1", Name: "Beatles"}},
				Album:  []AlbumID3{{ID: "a1", Name: "Help!"}},
				Song:   []Song{{ID: "s1", Title: "Yesterday"}},
			},
		})
	})

	res, err := c.Search3(context.Background(), "beatles", 5, 5, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Artist) != 1 || res.Artist[0].Name != "Beatles" {
		t.Fatalf("unexpected search results: %+v", res)
	}
}

func TestGetPlaylists(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			Playlists: &Playlists{Playlist: []Playlist{{ID: "p1", Name: "Favorites"}}},
		})
	})

	playlists, err := c.GetPlaylists(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(playlists) != 1 || playlists[0].Name != "Favorites" {
		t.Fatalf("unexpected playlists: %+v", playlists)
	}
}

func TestGetPlaylist(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			Playlist: &Playlist{ID: "p1", Name: "Road Trip", SongCount: 20},
		})
	})

	pl, err := c.GetPlaylist(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if pl.Name != "Road Trip" {
		t.Fatalf("unexpected playlist: %+v", pl)
	}
}

func TestGetNowPlaying(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/getNowPlaying" {
			t.Errorf("path = %q, want /rest/getNowPlaying", r.URL.Path)
		}
		subsonicJSON(t, w, &responseBody{
			NowPlaying: &NowPlaying{Entry: []NowPlayingEntry{{Song: Song{ID: "s1", Title: "Airbag"}, Username: "admin"}}},
		})
	})

	entries, err := c.GetNowPlaying(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Username != "admin" {
		t.Fatalf("unexpected now playing entries: %+v", entries)
	}
}

func TestGetGenres(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			Genres: &Genres{Genre: []Genre{{Value: "Rock", SongCount: 100, AlbumCount: 20}}},
		})
	})

	genres, err := c.GetGenres(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(genres) != 1 || genres[0].Value != "Rock" {
		t.Fatalf("unexpected genres: %+v", genres)
	}
}

func TestGetScanStatus(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			ScanStatus: &ScanStatus{Scanning: false, Count: 5000},
		})
	})

	ss, err := c.GetScanStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ss.Scanning || ss.Count != 5000 {
		t.Fatalf("unexpected scan status: %+v", ss)
	}
}

func TestStartScan(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/startScan" {
			t.Errorf("path = %q, want /rest/startScan", r.URL.Path)
		}
		subsonicJSON(t, w, &responseBody{
			ScanStatus: &ScanStatus{Scanning: true, Count: 12},
		})
	})

	status, err := c.StartScan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !status.Scanning || status.Count != 12 {
		t.Fatalf("unexpected scan status: %+v", status)
	}
}

func TestScrobble(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/scrobble" {
			t.Errorf("path = %q, want /rest/scrobble", r.URL.Path)
		}
		if got := r.URL.Query().Get("id"); got != "s1" {
			t.Errorf("id = %q, want s1", got)
		}
		if got := r.URL.Query().Get("submission"); got != "true" {
			t.Errorf("submission = %q, want true", got)
		}
		subsonicJSON(t, w, &responseBody{})
	})

	if err := c.Scrobble(context.Background(), "s1", true); err != nil {
		t.Fatal(err)
	}
}

func TestStar(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/star" {
			t.Errorf("path = %q, want /rest/star", r.URL.Path)
		}
		if got := r.URL.Query().Get("id"); got != "s1" {
			t.Errorf("id = %q, want s1", got)
		}
		if got := r.URL.Query().Get("albumId"); got != "a1" {
			t.Errorf("albumId = %q, want a1", got)
		}
		if got := r.URL.Query().Get("artistId"); got != "ar1" {
			t.Errorf("artistId = %q, want ar1", got)
		}
		subsonicJSON(t, w, &responseBody{})
	})

	if err := c.Star(context.Background(), "s1", "a1", "ar1"); err != nil {
		t.Fatal(err)
	}
}

func TestUnstar(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/unstar" {
			t.Errorf("path = %q, want /rest/unstar", r.URL.Path)
		}
		if got := r.URL.Query().Get("albumId"); got != "a1" {
			t.Errorf("albumId = %q, want a1", got)
		}
		subsonicJSON(t, w, &responseBody{})
	})

	if err := c.Unstar(context.Background(), "", "a1", ""); err != nil {
		t.Fatal(err)
	}
}

func TestGetStarred2(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		subsonicJSON(t, w, &responseBody{
			Starred2: &Starred2{
				Song: []Song{{ID: "s1", Title: "Bohemian Rhapsody"}},
			},
		})
	})

	starred, err := c.GetStarred2(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(starred.Song) != 1 || starred.Song[0].Title != "Bohemian Rhapsody" {
		t.Fatalf("unexpected starred: %+v", starred)
	}
}

func TestSubsonicError(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(subsonicResponse{
			Response: responseBody{
				Status:  "failed",
				Version: "1.16.1",
				Error:   &SubsonicError{Code: 40, Message: "Wrong username or password"},
			},
		})
	})

	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var se *SubsonicError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SubsonicError, got %T", err)
	}
	if se.Code != 40 {
		t.Fatalf("unexpected error code: %d", se.Code)
	}
	if got := se.Error(); got != "Wrong username or password" {
		t.Fatalf("Error() = %q, want Wrong username or password", got)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	})

	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status code: %d", apiErr.StatusCode)
	}
	if got := apiErr.Error(); got != "navidrome: 503 Service Unavailable: service unavailable\n" {
		t.Fatalf("Error() = %q, want navidrome service unavailable message", got)
	}
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	custom := &http.Client{}
	c, err := New("http://localhost", "u", "p", WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.http != custom {
		t.Fatal("custom HTTP client not set")
	}
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		subsonicJSON(t, w, &responseBody{})
	}))
	defer ts.Close()

	c, err := New(ts.URL, "admin", "password", WithUserAgent("myapp/1.2.3"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotUA != "myapp/1.2.3" {
		t.Errorf("User-Agent = %q, want %q", gotUA, "myapp/1.2.3")
	}
}

func TestNew_invalidURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, url string
	}{
		{"empty", ""},
		{"malformed", "://x"},
		{"ftp", "ftp://x"},
		{"no-scheme", "no-scheme"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := New(tc.url, "u", "p")
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}
