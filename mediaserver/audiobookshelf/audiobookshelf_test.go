package audiobookshelf_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/mediaserver/audiobookshelf/v2"
)

func newTestServer(t *testing.T, wantPath, wantToken string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+wantToken {
			t.Errorf("Authorization = %q, want Bearer %s", got, wantToken)
		}
		w.Header().Set("Content-Type", "application/json")
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
}

func TestGetLibraries(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/libraries", "test-token", map[string]any{
		"libraries": []map[string]any{
			{"id": "lib1", "name": "Audiobooks", "mediaType": "book"},
		},
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	libs, err := c.GetLibraries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(libs) != 1 {
		t.Fatalf("len(libraries) = %d, want 1", len(libs))
	}
	if libs[0].Name != "Audiobooks" {
		t.Errorf("Name = %q, want Audiobooks", libs[0].Name)
	}
}

func TestGetLibrary(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/libraries/lib1", "test-token", map[string]any{
		"id": "lib1", "name": "Audiobooks", "mediaType": "book",
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	lib, err := c.GetLibrary(context.Background(), "lib1")
	if err != nil {
		t.Fatal(err)
	}
	if lib.MediaType != "book" {
		t.Errorf("MediaType = %q, want book", lib.MediaType)
	}
}

func TestGetLibraryItems(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/libraries/lib1/items", "test-token", map[string]any{
		"results": []map[string]any{
			{"id": "item1", "libraryId": "lib1"},
		},
		"total": 1,
		"limit": 20,
		"page":  0,
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	resp, err := c.GetLibraryItems(context.Background(), "lib1", 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 {
		t.Errorf("Total = %d, want 1", resp.Total)
	}
}

func TestGetItem(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/items/item1", "test-token", map[string]any{
		"id": "item1", "libraryId": "lib1",
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	item, err := c.GetItem(context.Background(), "item1")
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "item1" {
		t.Errorf("ID = %q, want item1", item.ID)
	}
}

func newRawServer(t *testing.T, status int, rawBody string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rawBody))
	}))
}

func mustClient(t *testing.T, baseURL, token string) *audiobookshelf.Client {
	t.Helper()
	c, err := audiobookshelf.New(baseURL, token)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// TestGetUsers verifies GET /api/users envelope decoding, User permissions/mediaProgress/librariesAccessible,
// and bare-array rejection (server/routers/ApiRouter.js:134, server/controllers/UserController.js:29-49, server/models/User.js:169-183, 603-638).
func TestGetUsers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		status   int
		rawBody  string
		wantLen  int
		wantUser string
		wantErr  bool
	}{
		{
			name:   "positive envelope with permissions and progress",
			status: http.StatusOK,
			rawBody: `{"users":[{"id":"user1","username":"admin","type":"root","isActive":true,"isLocked":false,` +
				`"permissions":{"download":true,"update":true,"delete":true,"upload":true,"createEreader":true,` +
				`"accessAllLibraries":true,"accessAllTags":true,"accessExplicitContent":true,"selectedTagsNotAccessible":false},` +
				`"librariesAccessible":["lib1"],"mediaProgress":[{"id":"mp1","libraryItemId":"item1","progress":0.5,"isFinished":false}]}]}`,
			wantLen:  1,
			wantUser: "admin",
		},
		{
			name:    "boundary empty users list",
			status:  http.StatusOK,
			rawBody: `{"users":[]}`,
			wantLen: 0,
		},
		{
			name:    "negative bare array regression guard",
			status:  http.StatusOK,
			rawBody: `[{"id":"user1","username":"admin","type":"root"}]`,
			wantErr: true,
		},
		{
			name:    "negative non-2xx status",
			status:  http.StatusForbidden,
			rawBody: `Forbidden`,
			wantErr: true,
		},
		{
			name:    "negative malformed JSON",
			status:  http.StatusOK,
			rawBody: `{"users":`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runGetUsersCase(t, tc.status, tc.rawBody, tc.wantLen, tc.wantUser, tc.wantErr)
		})
	}
}

func runGetUsersCase(t *testing.T, status int, rawBody string, wantLen int, wantUser string, wantErr bool) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users" {
			t.Errorf("path = %q, want /api/users", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rawBody))
	}))
	defer ts.Close()

	users, err := mustClient(t, ts.URL, "test-token").GetUsers(context.Background())
	if (err != nil) != wantErr {
		t.Fatalf("GetUsers error = %v, wantErr %v", err, wantErr)
	}
	if wantErr || len(users) != wantLen {
		if !wantErr {
			t.Fatalf("len(users) = %d, want %d", len(users), wantLen)
		}
		return
	}
	if wantLen > 0 {
		assertAdminUserFields(t, users[0], wantUser)
	}
}

func assertAdminUserFields(t *testing.T, u audiobookshelf.User, wantUsername string) {
	t.Helper()
	if u.Username != wantUsername {
		t.Errorf("Username = %q, want %q", u.Username, wantUsername)
	}
	if !u.Permissions.Download || !u.Permissions.Delete || !u.Permissions.CreateEreader || !u.Permissions.AccessExplicitContent {
		t.Errorf("Permissions = %+v, want root permissions true", u.Permissions)
	}
	if len(u.LibrariesAccessible) != 1 || u.LibrariesAccessible[0] != "lib1" {
		t.Errorf("LibrariesAccessible = %v, want [lib1]", u.LibrariesAccessible)
	}
	if len(u.MediaProgress) != 1 || u.MediaProgress[0].ID != "mp1" {
		t.Errorf("MediaProgress = %+v, want [mp1]", u.MediaProgress)
	}
}

// TestGetUser verifies GET /api/users/:id path escaping, User fields, and error cases
// (server/routers/ApiRouter.js:136, server/controllers/UserController.js:59-105, server/models/User.js:603-638).
func TestGetUser(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		userID      string
		wantEscPath string
		status      int
		rawBody     string
		wantErr     bool
	}{
		{
			name:        "positive path escaped id",
			userID:      "usr/1",
			wantEscPath: "/api/users/usr%2F1",
			status:      http.StatusOK,
			rawBody: `{"id":"usr/1","username":"alice","type":"user","isActive":true,` +
				`"permissions":{"download":true,"accessAllLibraries":true,"accessAllTags":true},` +
				`"librariesAccessible":["lib1"],"mediaProgress":[{"id":"mp1","libraryItemId":"li1","progress":1,"isFinished":true}]}`,
		},
		{
			name:    "boundary empty userID",
			userID:  "",
			wantErr: true,
		},
		{
			name:        "negative 404 not found",
			userID:      "missing",
			wantEscPath: "/api/users/missing",
			status:      http.StatusNotFound,
			rawBody:     `Not Found`,
			wantErr:     true,
		},
		{
			name:        "negative malformed JSON",
			userID:      "u1",
			wantEscPath: "/api/users/u1",
			status:      http.StatusOK,
			rawBody:     `{bad`,
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runGetUserCase(t, tc.userID, tc.wantEscPath, tc.status, tc.rawBody, tc.wantErr)
		})
	}
}

func runGetUserCase(t *testing.T, userID, wantEscPath string, status int, rawBody string, wantErr bool) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != wantEscPath {
			t.Errorf("EscapedPath = %q, want %q", r.URL.EscapedPath(), wantEscPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rawBody))
	}))
	defer ts.Close()

	u, err := mustClient(t, ts.URL, "test-token").GetUser(context.Background(), userID)
	if (err != nil) != wantErr {
		t.Fatalf("GetUser error = %v, wantErr %v", err, wantErr)
	}
	if wantErr {
		return
	}
	if u.ID != userID || !u.Permissions.Download || len(u.MediaProgress) != 1 {
		t.Errorf("unexpected user: %+v", u)
	}
}

// TestLogin verifies POST /login at root with X-Return-Tokens: true, empty-token client support,
// and credential-free error messages (server/Server.js:375, server/Auth.js:96-105, 296-312, 322-328).
func TestLogin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		username string
		password string
		status   int
		rawBody  string
		wantErr  bool
	}{
		{
			name:     "positive returns user and tokens",
			username: "alice",
			password: "correct-horse-battery-staple",
			status:   http.StatusOK,
			rawBody: `{"user":{"id":"u1","username":"alice","type":"admin","token":"legacy-tok",` +
				`"accessToken":"jwt-access","refreshToken":"jwt-refresh","isActive":true,` +
				`"permissions":{"download":true,"update":true,"upload":true}},` +
				`"userDefaultLibraryId":"lib1","Source":"docker"}`,
		},
		{
			name:     "boundary empty username",
			username: "",
			password: "secret-value-not-logged",
			wantErr:  true,
		},
		{
			name:     "negative 401 unauthorized",
			username: "alice",
			password: "secret-value-not-logged",
			status:   http.StatusUnauthorized,
			rawBody:  `Unauthorized`,
			wantErr:  true,
		},
		{
			name:     "negative malformed JSON",
			username: "alice",
			password: "secret-value-not-logged",
			status:   http.StatusOK,
			rawBody:  `{invalid`,
			wantErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runLoginCase(t, tc.username, tc.password, tc.status, tc.rawBody, tc.wantErr)
		})
	}
}

func runLoginCase(t *testing.T, username, password string, status int, rawBody string, wantErr bool) {
	t.Helper()
	ts := newLoginServer(t, status, rawBody)
	defer ts.Close()

	resp, err := mustClient(t, ts.URL, "").Login(context.Background(), username, password)
	if (err != nil) != wantErr {
		t.Fatalf("Login error = %v, wantErr %v", err, wantErr)
	}
	if wantErr {
		if password != "" && strings.Contains(err.Error(), password) {
			t.Errorf("error leaked password: %v", err)
		}
		return
	}
	if resp.User.AccessToken != "jwt-access" || resp.User.RefreshToken != "jwt-refresh" {
		t.Errorf("unexpected LoginResponse: %+v", resp)
	}
}

func newLoginServer(t *testing.T, status int, rawBody string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/login" {
			t.Errorf("request = %s %s, want POST /login", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Return-Tokens") != "true" || r.Header.Get("Authorization") != "" {
			t.Errorf("headers X-Return-Tokens=%q Authorization=%q", r.Header.Get("X-Return-Tokens"), r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rawBody))
	}))
}

// TestLogout verifies POST /logout at root (server/Server.js:375, server/Auth.js:486-524).
func TestLogout(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		status  int
		rawBody string
		wantErr bool
	}{
		{name: "positive 200", status: http.StatusOK, rawBody: `{"redirect_url":null}`},
		{name: "negative 500", status: http.StatusInternalServerError, rawBody: `Internal Server Error`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/logout" {
					t.Errorf("request = %s %s, want POST /logout", r.Method, r.URL.Path)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.rawBody))
			}))
			defer ts.Close()

			err := mustClient(t, ts.URL, "test-token").Logout(context.Background())
			if (err != nil) != tc.wantErr {
				t.Fatalf("Logout error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestGetLibraryItemsWithOptions verifies GET /api/libraries/:id/items query parameters and response fields
// (server/routers/ApiRouter.js:75, server/controllers/LibraryController.js:604-639, 1487-1495).
func TestGetLibraryItemsWithOptions(t *testing.T) {
	t.Parallel()

	fullOpts := &audiobookshelf.LibraryItemsOptions{
		Limit: 25, Page: 2, Sort: "media.metadata.title", Desc: true,
		Filter: "genres.c2NpLWZp", Minified: true, CollapseSeries: true, Include: "rssfeed,episodes",
	}
	fullQuery := url.Values{
		"limit": {"25"}, "page": {"2"}, "sort": {"media.metadata.title"}, "desc": {"1"},
		"filter": {"genres.c2NpLWZp"}, "minified": {"1"}, "collapseseries": {"1"}, "include": {"rssfeed,episodes"},
	}
	fullBody := `{"results":[{"id":"item1","libraryId":"lib1"}],"total":55,"limit":25,"page":2,` +
		`"sortBy":"media.metadata.title","sortDesc":true,"filterBy":"genres.c2NpLWZp",` +
		`"mediaType":"book","minified":true,"collapseseries":true,"include":"rssfeed,episodes","offset":50}`

	cases := []struct {
		name      string
		libraryID string
		opts      *audiobookshelf.LibraryItemsOptions
		wantQuery url.Values
		status    int
		rawBody   string
		wantErr   bool
	}{
		{name: "positive all options and response fields", libraryID: "lib1", opts: fullOpts, wantQuery: fullQuery, status: http.StatusOK, rawBody: fullBody},
		{name: "boundary zero page and limit with nil opts", libraryID: "lib1", wantQuery: url.Values{}, status: http.StatusOK, rawBody: `{"results":[],"total":0,"limit":0,"page":0,"mediaType":"book","offset":0}`},
		{name: "boundary empty libraryID", libraryID: "", wantErr: true},
		{name: "negative 404", libraryID: "lib1", wantQuery: url.Values{}, status: http.StatusNotFound, rawBody: `Library not found`, wantErr: true},
		{name: "negative malformed JSON", libraryID: "lib1", wantQuery: url.Values{}, status: http.StatusOK, rawBody: `{bad`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runGetLibraryItemsOptionsCase(t, tc.libraryID, tc.opts, tc.wantQuery, tc.status, tc.rawBody, tc.wantErr)
		})
	}
}

func runGetLibraryItemsOptionsCase(
	t *testing.T,
	libraryID string,
	opts *audiobookshelf.LibraryItemsOptions,
	wantQuery url.Values,
	status int,
	rawBody string,
	wantErr bool,
) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/libraries/"+libraryID+"/items" {
			t.Errorf("path = %q, want /api/libraries/%s/items", r.URL.Path, libraryID)
		}
		if got := r.URL.Query().Encode(); got != wantQuery.Encode() {
			t.Errorf("query = %q, want %q", got, wantQuery.Encode())
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rawBody))
	}))
	defer ts.Close()

	resp, err := mustClient(t, ts.URL, "test-token").GetLibraryItemsWithOptions(context.Background(), libraryID, opts)
	if (err != nil) != wantErr {
		t.Fatalf("GetLibraryItemsWithOptions error = %v, wantErr %v", err, wantErr)
	}
	if !wantErr && opts != nil && (resp.Offset != 50 || !resp.SortDesc || !resp.CollapseSeries) {
		t.Errorf("unexpected response metadata: %+v", resp)
	}
}

// TestGetPodcastFeed verifies POST /api/podcasts/feed
// (server/routers/ApiRouter.js:249, server/controllers/PodcastController.js:191-216, server/utils/podcastUtils.js:9-59).
func TestGetPodcastFeed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		rssFeed string
		status  int
		rawBody string
		wantErr bool
	}{
		{
			name:    "positive feed parsed",
			rssFeed: "http://feed.example.local/rss.xml",
			status:  http.StatusOK,
			rawBody: `{"podcast":{"metadata":{"title":"Go Time","author":"Changelog","feedUrl":"http://feed.example.local/rss.xml","categories":["Technology"]},` +
				`"episodes":[{"title":"Ep 1","enclosure":{"url":"http://feed.example.local/ep1.mp3","type":"audio/mpeg"}}],"numEpisodes":1}}`,
		},
		{name: "boundary empty rssFeed", rssFeed: "", wantErr: true},
		{name: "negative 400 invalid feed", rssFeed: "bad", status: http.StatusBadRequest, rawBody: `Invalid request body`, wantErr: true},
		{name: "negative malformed JSON", rssFeed: "http://feed.example.local/rss.xml", status: http.StatusOK, rawBody: `{bad`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runGetPodcastFeedCase(t, tc.rssFeed, tc.status, tc.rawBody, tc.wantErr)
		})
	}
}

func runGetPodcastFeedCase(t *testing.T, rssFeed string, status int, rawBody string, wantErr bool) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/podcasts/feed" {
			t.Errorf("request = %s %s, want POST /api/podcasts/feed", r.Method, r.URL.Path)
		}
		var reqBody audiobookshelf.PodcastFeedRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || reqBody.RSSFeed != rssFeed {
			t.Errorf("body rssFeed = %q (err %v), want %q", reqBody.RSSFeed, err, rssFeed)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rawBody))
	}))
	defer ts.Close()

	feed, err := mustClient(t, ts.URL, "test-token").GetPodcastFeed(context.Background(), rssFeed)
	if (err != nil) != wantErr {
		t.Fatalf("GetPodcastFeed error = %v, wantErr %v", err, wantErr)
	}
	if !wantErr && (feed.Metadata.Title != "Go Time" || len(feed.Episodes) != 1) {
		t.Errorf("unexpected feed: %+v", feed)
	}
}

// TestCreatePodcast verifies POST /api/podcasts
// (server/routers/ApiRouter.js:248, server/controllers/PodcastController.js:31-189, server/models/Podcast.js:79-121).
func TestCreatePodcast(t *testing.T) {
	t.Parallel()

	validReq := &audiobookshelf.CreatePodcastRequest{
		LibraryID: "lib1",
		FolderID:  "fold1",
		Path:      "/podcasts/Go Time",
		Media: audiobookshelf.CreatePodcastMedia{
			Metadata:             audiobookshelf.PodcastMetadata{Title: "Go Time", Author: "Changelog", FeedURL: "http://feed.example.local/rss.xml"},
			AutoDownloadEpisodes: true,
		},
	}
	okBody := `{"id":"li-pod-1","libraryId":"lib1","folderId":"fold1","path":"/podcasts/Go Time","mediaType":"podcast",` +
		`"media":{"id":"pod1","metadata":{"title":"Go Time","author":"Changelog","feedUrl":"http://feed.example.local/rss.xml"},"autoDownloadEpisodes":true}}`

	cases := []struct {
		name    string
		req     *audiobookshelf.CreatePodcastRequest
		status  int
		rawBody string
		wantErr bool
	}{
		{name: "positive created", req: validReq, status: http.StatusOK, rawBody: okBody},
		{name: "boundary nil request", req: nil, wantErr: true},
		{name: "boundary empty path", req: &audiobookshelf.CreatePodcastRequest{LibraryID: "lib1", FolderID: "fold1", Path: "   "}, wantErr: true},
		{name: "negative 400", req: validReq, status: http.StatusBadRequest, rawBody: `Podcast already exists`, wantErr: true},
		{name: "negative malformed JSON", req: validReq, status: http.StatusOK, rawBody: `{bad`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ts := newRawServer(t, tc.status, tc.rawBody)
			defer ts.Close()

			item, err := mustClient(t, ts.URL, "test-token").CreatePodcast(context.Background(), tc.req)
			if (err != nil) != tc.wantErr {
				t.Fatalf("CreatePodcast error = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && (item.ID != "li-pod-1" || item.Media.Metadata.Title != "Go Time") {
				t.Errorf("unexpected item: %+v", item)
			}
		})
	}
}

// TestCheckNewEpisodes verifies GET /api/podcasts/:id/checknew
// (server/routers/ApiRouter.js:252, server/controllers/PodcastController.js:276-301).
func TestCheckNewEpisodes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		podcastID   string
		limit       int
		wantEscPath string
		wantLimit   string
		status      int
		rawBody     string
		wantCount   int
		wantErr     bool
	}{
		{
			name: "positive with limit and path escape", podcastID: "pod/1", limit: 5,
			wantEscPath: "/api/podcasts/pod%2F1/checknew", wantLimit: "5", status: http.StatusOK,
			rawBody: `{"episodes":[{"title":"Ep 2","enclosure":{"url":"http://feed.example.local/ep2.mp3"}}]}`, wantCount: 1,
		},
		{
			name: "boundary zero limit and empty episodes", podcastID: "pod1", limit: 0,
			wantEscPath: "/api/podcasts/pod1/checknew", status: http.StatusOK, rawBody: `{"episodes":[]}`,
		},
		{name: "boundary empty podcastID", podcastID: "", wantErr: true},
		{name: "negative 403", podcastID: "pod1", wantEscPath: "/api/podcasts/pod1/checknew", status: http.StatusForbidden, rawBody: `Forbidden`, wantErr: true},
		{name: "negative malformed JSON", podcastID: "pod1", wantEscPath: "/api/podcasts/pod1/checknew", status: http.StatusOK, rawBody: `{bad`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runCheckNewEpisodesCase(t, tc.podcastID, tc.limit, tc.wantEscPath, tc.wantLimit, tc.status, tc.rawBody, tc.wantCount, tc.wantErr)
		})
	}
}

func runCheckNewEpisodesCase(
	t *testing.T,
	podcastID string,
	limit int,
	wantEscPath, wantLimit string,
	status int,
	rawBody string,
	wantCount int,
	wantErr bool,
) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != wantEscPath {
			t.Errorf("EscapedPath = %q, want %q", r.URL.EscapedPath(), wantEscPath)
		}
		if got := r.URL.Query().Get("limit"); got != wantLimit {
			t.Errorf("limit = %q, want %q", got, wantLimit)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rawBody))
	}))
	defer ts.Close()

	eps, err := mustClient(t, ts.URL, "test-token").CheckNewEpisodes(context.Background(), podcastID, limit)
	if (err != nil) != wantErr {
		t.Fatalf("CheckNewEpisodes error = %v, wantErr %v", err, wantErr)
	}
	if !wantErr && len(eps) != wantCount {
		t.Errorf("len(episodes) = %d, want %d", len(eps), wantCount)
	}
}

// TestDownloadEpisodes verifies POST /api/podcasts/:id/download-episodes
// (server/routers/ApiRouter.js:256, server/controllers/PodcastController.js:360-380).
func TestDownloadEpisodes(t *testing.T) {
	t.Parallel()

	sample := []audiobookshelf.PodcastFeedEpisode{
		{Title: "Ep 1", Enclosure: &audiobookshelf.PodcastEnclosure{URL: "http://feed.example.local/ep1.mp3"}},
	}

	cases := []struct {
		name      string
		podcastID string
		episodes  []audiobookshelf.PodcastFeedEpisode
		status    int
		rawBody   string
		wantErr   bool
	}{
		{name: "positive 200 OK plain text", podcastID: "pod/1", episodes: sample, status: http.StatusOK, rawBody: `OK`},
		{name: "boundary empty podcastID", podcastID: "", episodes: sample, wantErr: true},
		{name: "boundary empty episodes slice", podcastID: "pod1", episodes: nil, wantErr: true},
		{name: "negative 400", podcastID: "pod1", episodes: sample, status: http.StatusBadRequest, rawBody: `Bad Request`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.EscapedPath(), "/download-episodes") {
					t.Errorf("request = %s %s", r.Method, r.URL.EscapedPath())
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.rawBody))
			}))
			defer ts.Close()

			err := mustClient(t, ts.URL, "test-token").DownloadEpisodes(context.Background(), tc.podcastID, tc.episodes)
			if (err != nil) != tc.wantErr {
				t.Fatalf("DownloadEpisodes error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestSearchPodcasts verifies GET /api/search/podcast
// (server/routers/ApiRouter.js:288, server/controllers/SearchController.js:135-156, server/providers/iTunes.js:14-28, 139-168).
func TestSearchPodcasts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		term        string
		country     string
		wantCountry string
		status      int
		rawBody     string
		wantLen     int
		wantErr     bool
	}{
		{
			name: "positive with term and country", term: "golang", country: "us", wantCountry: "us", status: http.StatusOK,
			rawBody: `[{"id":12345,"artistId":678,"title":"Go Time","artistName":"Changelog","feedUrl":"http://feed.example.local/rss.xml","trackCount":250,"explicit":false}]`,
			wantLen: 1,
		},
		{name: "boundary empty country and empty results", term: "nonexistent", status: http.StatusOK, rawBody: `[]`},
		{name: "boundary empty term", term: "", country: "us", wantErr: true},
		{name: "negative 500", term: "golang", status: http.StatusInternalServerError, rawBody: `{"error":"Internal server error"}`, wantErr: true},
		{name: "negative malformed JSON", term: "golang", status: http.StatusOK, rawBody: `[bad`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runSearchPodcastsCase(t, tc.term, tc.country, tc.wantCountry, tc.status, tc.rawBody, tc.wantLen, tc.wantErr)
		})
	}
}

func runSearchPodcastsCase(
	t *testing.T,
	term, country, wantCountry string,
	status int,
	rawBody string,
	wantLen int,
	wantErr bool,
) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search/podcast" {
			t.Errorf("path = %q, want /api/search/podcast", r.URL.Path)
		}
		if got := r.URL.Query().Get("term"); got != term {
			t.Errorf("term = %q, want %q", got, term)
		}
		if got := r.URL.Query().Get("country"); got != wantCountry {
			t.Errorf("country = %q, want %q", got, wantCountry)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rawBody))
	}))
	defer ts.Close()

	res, err := mustClient(t, ts.URL, "test-token").SearchPodcasts(context.Background(), term, country)
	if (err != nil) != wantErr {
		t.Fatalf("SearchPodcasts error = %v, wantErr %v", err, wantErr)
	}
	if !wantErr && len(res) != wantLen {
		t.Errorf("len(results) = %d, want %d", len(res), wantLen)
	}
}

// TestGetStatus verifies GET /status at root (server/Server.js:384-400, server/objects/settings/ServerSettings.js:348-357).
func TestGetStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		status      int
		rawBody     string
		wantVersion string
		wantInit    bool
		wantErr     bool
	}{
		{
			name:   "positive initialized server",
			status: http.StatusOK,
			rawBody: `{"app":"audiobookshelf","serverVersion":"2.37.2","isInit":true,"language":"en-us",` +
				`"authMethods":["local","openid"],"authFormData":{"authLoginCustomMessage":"Welcome","authOpenIDButtonText":"SSO","authOpenIDAutoLaunch":false}}`,
			wantVersion: "2.37.2",
			wantInit:    true,
		},
		{
			name:        "boundary uninitialized server with paths",
			status:      http.StatusOK,
			rawBody:     `{"app":"audiobookshelf","serverVersion":"2.37.2","isInit":false,"ConfigPath":"/config","MetadataPath":"/metadata"}`,
			wantVersion: "2.37.2",
			wantInit:    false,
		},
		{name: "negative 500", status: http.StatusInternalServerError, rawBody: `error`, wantErr: true},
		{name: "negative malformed JSON", status: http.StatusOK, rawBody: `{bad`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runGetStatusCase(t, tc.status, tc.rawBody, tc.wantVersion, tc.wantInit, tc.wantErr)
		})
	}
}

func runGetStatusCase(t *testing.T, status int, rawBody, wantVersion string, wantInit, wantErr bool) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/status" {
			t.Errorf("request = %s %s, want GET /status", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rawBody))
	}))
	defer ts.Close()

	st, err := mustClient(t, ts.URL, "").GetStatus(context.Background())
	if (err != nil) != wantErr {
		t.Fatalf("GetStatus error = %v, wantErr %v", err, wantErr)
	}
	if !wantErr && (st.ServerVersion != wantVersion || st.IsInit != wantInit) {
		t.Errorf("unexpected status: %+v", st)
	}
}

func TestGetMe(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/me", "test-token", map[string]any{
		"id": "user1", "username": "admin", "type": "root",
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	user, err := c.GetMe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user.Type != "root" {
		t.Errorf("Type = %q, want root", user.Type)
	}
}

func TestGetCollections(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/libraries/lib1/collections", "test-token", map[string]any{
		"results": []map[string]any{
			{"id": "c1", "libraryId": "lib1", "name": "Favorites"},
		},
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	collections, err := c.GetCollections(context.Background(), "lib1")
	if err != nil {
		t.Fatal(err)
	}
	if collections[0].Name != "Favorites" {
		t.Errorf("Name = %q, want Favorites", collections[0].Name)
	}
}

func TestGetServerInfo(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/server", "test-token", map[string]any{
		"version": "2.5.0",
		"isInit":  true,
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	info, err := c.GetServerInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "2.5.0" {
		t.Errorf("Version = %q, want 2.5.0", info.Version)
	}
}

func TestSearch(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/libraries/lib1/search" {
			t.Errorf("path = %q, want /api/libraries/lib1/search", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != "dune" {
			t.Errorf("q = %q, want dune", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want Bearer test-token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"book": "Dune"})
	}))
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	data, err := c.Search(context.Background(), "lib1", "dune")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if got["book"] != "Dune" {
		t.Errorf("book = %q, want Dune", got["book"])
	}
}

func TestGetSessions(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/sessions", "test-token", map[string]any{
		"sessions": []map[string]any{
			{"id": "s1", "userId": "u1", "displayTitle": "Test Book"},
		},
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sessions, err := c.GetSessions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}
}

func TestGetMediaProgress(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/me/progress/item1", "test-token", map[string]any{
		"id": "p1", "libraryItemId": "item1", "progress": 0.5, "currentTime": 120.0,
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	progress, err := c.GetMediaProgress(context.Background(), "item1")
	if err != nil {
		t.Fatal(err)
	}
	if progress.LibraryItemID != "item1" {
		t.Errorf("LibraryItemID = %q, want item1", progress.LibraryItemID)
	}
}

func TestGetAuthors(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/libraries/lib1/authors", "test-token", map[string]any{
		"authors": []map[string]any{
			{"id": "a1", "name": "Ursula K. Le Guin"},
		},
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	authors, err := c.GetAuthors(context.Background(), "lib1")
	if err != nil {
		t.Fatal(err)
	}
	if authors[0].Name != "Ursula K. Le Guin" {
		t.Errorf("Name = %q, want Ursula K. Le Guin", authors[0].Name)
	}
}

func TestGetSeries(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/libraries/lib1/series", "test-token", map[string]any{
		"results": []map[string]any{
			{"id": "s1", "name": "Earthsea"},
		},
	})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	series, err := c.GetSeries(context.Background(), "lib1")
	if err != nil {
		t.Fatal(err)
	}
	if series[0].Name != "Earthsea" {
		t.Errorf("Name = %q, want Earthsea", series[0].Name)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Forbidden"))
	}))
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "bad-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetLibraries(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *audiobookshelf.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusForbidden)
	}
	if got := apiErr.Error(); got != "audiobookshelf: HTTP 403: Forbidden" {
		t.Errorf("Error() = %q, want audiobookshelf: HTTP 403: Forbidden", got)
	}
	if got := (&audiobookshelf.APIError{StatusCode: http.StatusNotFound}).Error(); got != "audiobookshelf: HTTP 404" {
		t.Errorf("Error() without body = %q, want audiobookshelf: HTTP 404", got)
	}
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	called := false
	custom := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			called = true
			return http.DefaultTransport.RoundTrip(r)
		}),
	}

	ts := newTestServer(t, "/api/libraries", "k", map[string]any{"libraries": []any{}})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "k", audiobookshelf.WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, _ = c.GetLibraries(context.Background())
	if !called {
		t.Error("custom HTTP client was not used")
	}
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"libraries": []any{}})
	}))
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "k", audiobookshelf.WithUserAgent("myapp/1.2.3"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := c.GetLibraries(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotUA != "myapp/1.2.3" {
		t.Errorf("User-Agent = %q, want %q", gotUA, "myapp/1.2.3")
	}
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/api/libraries", "k", map[string]any{"libraries": []any{}})
	defer ts.Close()

	c, err := audiobookshelf.New(ts.URL, "k", audiobookshelf.WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := c.GetLibraries(context.Background()); err != nil {
		t.Fatal(err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

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
			c, err := audiobookshelf.New(tc.url, "k")
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}
