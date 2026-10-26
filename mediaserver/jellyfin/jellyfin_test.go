package jellyfin_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golusoris/goenvoy/mediaserver/jellyfin/v3"
)

func newTestServer(t *testing.T, wantPath string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
}

func TestAuthenticateByName(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/Users/AuthenticateByName" {
			t.Errorf("path = %q, want /Users/AuthenticateByName", r.URL.Path)
		}
		var body struct {
			Username string `json:"Username"`
			Pw       string `json:"Pw"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Username != "admin" {
			t.Errorf("username = %q, want admin", body.Username)
		}
		if body.Pw != "secret" {
			t.Errorf("password = %q, want secret", body.Pw)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jellyfin.AuthenticationResult{
			AccessToken: "test-token-123",
			ServerID:    "server-1",
		})
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.AuthenticateByName(context.Background(), "admin", "secret"); err != nil {
		t.Fatal(err)
	}
}

func TestAuthenticationError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid username or password"})
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = c.AuthenticateByName(context.Background(), "bad", "creds")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
}

func TestGetPublicSystemInfo(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/System/Info/Public", jellyfin.PublicSystemInfo{
		ServerName:   "My Jellyfin Server",
		Version:      "10.8.13",
		ID:           "abc123",
		LocalAddress: "192.168.1.100",
	})
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	info, err := c.GetPublicSystemInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.ServerName != "My Jellyfin Server" {
		t.Errorf("ServerName = %q, want My Jellyfin Server", info.ServerName)
	}
	if info.Version != "10.8.13" {
		t.Errorf("Version = %q, want 10.8.13", info.Version)
	}
}

func TestGetSystemInfo(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/System/Info", jellyfin.SystemInfo{
		ServerName:                 "My Server",
		Version:                    "10.8.13",
		ID:                         "abc123",
		OperatingSystem:            "Linux",
		OperatingSystemDisplayName: "Debian GNU/Linux 12",
	})
	defer ts.Close()

	c, err := jellyfin.New(ts.URL, jellyfin.WithAccessToken("token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	info, err := c.GetSystemInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.OperatingSystem != "Linux" {
		t.Errorf("OS = %q, want Linux", info.OperatingSystem)
	}
}

func TestPing(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/System/Ping", "")
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestGetUsers(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/Users", []jellyfin.UserDto{
		{Name: "admin", ID: "user-1", HasPassword: true},
		{Name: "guest", ID: "user-2", HasPassword: false},
	})
	defer ts.Close()

	c, err := jellyfin.New(ts.URL, jellyfin.WithAccessToken("token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	users, err := c.GetUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("len = %d, want 2", len(users))
	}
	if users[0].Name != "admin" {
		t.Errorf("Name = %q, want admin", users[0].Name)
	}
}

func TestGetCurrentUser(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/Users/Me", jellyfin.UserDto{
		Name:        "admin",
		ID:          "user-1",
		ServerID:    "server-1",
		HasPassword: true,
	})
	defer ts.Close()

	c, err := jellyfin.New(ts.URL, jellyfin.WithAccessToken("token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	user, err := c.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user.Name != "admin" {
		t.Errorf("Name = %q, want admin", user.Name)
	}
}

func TestGetSessions(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/Sessions", []jellyfin.SessionInfoDto{
		{
			ID:         "session-1",
			UserID:     "user-1",
			UserName:   "admin",
			DeviceName: "Chrome",
			Client:     "Jellyfin Web",
			IsActive:   true,
		},
	})
	defer ts.Close()

	c, err := jellyfin.New(ts.URL, jellyfin.WithAccessToken("token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sessions, err := c.GetSessions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("len = %d, want 1", len(sessions))
	}
	if sessions[0].UserName != "admin" {
		t.Errorf("UserName = %q, want admin", sessions[0].UserName)
	}
}

func TestGetItems(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/Items", jellyfin.ItemsResult{
		Items: []jellyfin.BaseItemDto{
			{ID: "item-1", Name: "Inception", Type: "Movie"},
			{ID: "item-2", Name: "Interstellar", Type: "Movie"},
		},
		TotalRecordCount: 2,
		StartIndex:       0,
	})
	defer ts.Close()

	c, err := jellyfin.New(ts.URL, jellyfin.WithAccessToken("token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.GetItems(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalRecordCount != 2 {
		t.Errorf("TotalRecordCount = %d, want 2", result.TotalRecordCount)
	}
	if len(result.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(result.Items))
	}
	if result.Items[0].Name != "Inception" {
		t.Errorf("Name = %q, want Inception", result.Items[0].Name)
	}
}

func TestGetItemsByParent(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Items" {
			t.Errorf("path = %q, want /Items", r.URL.Path)
		}
		if got := r.URL.Query().Get("ParentId"); got != "lib-1" {
			t.Errorf("ParentId = %q, want lib-1", got)
		}
		if got := r.URL.Query().Get("StartIndex"); got != "0" {
			t.Errorf("StartIndex = %q, want 0", got)
		}
		if got := r.URL.Query().Get("Limit"); got != "50" {
			t.Errorf("Limit = %q, want 50", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jellyfin.ItemsResult{
			Items:            []jellyfin.BaseItemDto{{ID: "item-1", Name: "Movie 1"}},
			TotalRecordCount: 100,
			StartIndex:       0,
		})
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL, jellyfin.WithAccessToken("token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.GetItemsByParent(context.Background(), "lib-1", 0, 50)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalRecordCount != 100 {
		t.Errorf("TotalRecordCount = %d, want 100", result.TotalRecordCount)
	}
}

func TestGetItem(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/Items/item-1", jellyfin.BaseItemDto{
		ID:           "item-1",
		Name:         "The Matrix",
		Type:         "Movie",
		RunTimeTicks: 8208000000,
		Genres:       []string{"Action", "Sci-Fi"},
	})
	defer ts.Close()

	c, err := jellyfin.New(ts.URL, jellyfin.WithAccessToken("token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	item, err := c.GetItem(context.Background(), "item-1")
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "The Matrix" {
		t.Errorf("Name = %q, want The Matrix", item.Name)
	}
	if len(item.Genres) != 2 {
		t.Errorf("len(Genres) = %d, want 2", len(item.Genres))
	}
}

func TestGetUserViews(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/UserViews" {
			t.Errorf("path = %q, want /UserViews", r.URL.Path)
		}
		if got := r.URL.Query().Get("userId"); got != "user-1" {
			t.Errorf("userId = %q, want user-1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jellyfin.ItemsResult{
			Items: []jellyfin.BaseItemDto{
				{ID: "lib-1", Name: "Movies", Type: "CollectionFolder", IsFolder: true},
				{ID: "lib-2", Name: "TV Shows", Type: "CollectionFolder", IsFolder: true},
			},
			TotalRecordCount: 2,
		})
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL, jellyfin.WithAccessToken("token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := c.GetUserViews(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("len = %d, want 2", len(result.Items))
	}
	if result.Items[0].Name != "Movies" {
		t.Errorf("Name = %q, want Movies", result.Items[0].Name)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("{\"message\":\"Access denied\"}"))
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL, jellyfin.WithAccessToken("bad"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetUsers(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want 403", apiErr.StatusCode)
	}
	if apiErr.Message != "Access denied" {
		t.Errorf("Message = %q, want Access denied", apiErr.Message)
	}
}

func TestAPIErrorMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  jellyfin.APIError
		want string
	}{
		{"with message", jellyfin.APIError{StatusCode: 401, Message: "Unauthorized"}, "jellyfin: HTTP 401: Unauthorized"},
		{"raw body", jellyfin.APIError{StatusCode: 502, RawBody: "gateway error"}, "jellyfin: HTTP 502: gateway error"},
		{"code only", jellyfin.APIError{StatusCode: 500}, "jellyfin: HTTP 500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/Users", []jellyfin.UserDto{})
	defer ts.Close()

	c, err := jellyfin.New(ts.URL, jellyfin.WithAccessToken("token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = c.GetUsers(ctx)
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.Contains(auth, `DeviceId="my-device-123"`) {
			t.Errorf("Authorization = %q, want DeviceId my-device-123", auth)
		}
		if !strings.Contains(auth, `Token="token"`) {
			t.Errorf("Authorization = %q, want Token", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]jellyfin.UserDto{})
	}))
	defer ts.Close()

	c, err := jellyfin.New(
		ts.URL,
		jellyfin.WithAccessToken("token"),
		jellyfin.WithDeviceID("my-device-123"),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]jellyfin.UserDto{})
	}))
	defer ts.Close()

	c, err := jellyfin.New(
		ts.URL,
		jellyfin.WithAccessToken("token"),
		jellyfin.WithUserAgent("myapp/1.2.3"),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := c.GetUsers(context.Background()); err != nil {
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
			c, err := jellyfin.New(tc.url)
			if err == nil {
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("expected nil client")
			}
		})
	}
}

func assertPlayedPercentage(t *testing.T, got, want *float64) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("PlayedPercentage = %v, want nil", *got)
		}
		return
	}
	if got == nil {
		t.Fatalf("PlayedPercentage = nil, want %v", *want)
	}
	if *got != *want {
		t.Errorf("PlayedPercentage = %v, want %v", *got, *want)
	}
}

func TestUserItemDataDto_PlayedFields(t *testing.T) {
	t.Parallel()

	pct85 := 85.5
	pct0 := 0.0
	pct100 := 100.0

	cases := []struct {
		name       string
		json       string
		wantPlayed bool
		wantPct    *float64
	}{
		{"played with pct", `{"PlayCount": 3, "Played": true, "PlayedPercentage": 85.5}`, true, &pct85},
		{"unplayed null pct", `{"PlayCount": 0, "Played": false, "PlayedPercentage": null}`, false, nil},
		{"boundary 0 pct", `{"PlayCount": 1, "Played": false, "PlayedPercentage": 0.0}`, false, &pct0},
		{"boundary 100 pct", `{"PlayCount": 2, "Played": true, "PlayedPercentage": 100.0}`, true, &pct100},
		{"omitted pct", `{"PlayCount": 1, "Played": true}`, true, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var data jellyfin.UserItemDataDto
			if err := json.Unmarshal([]byte(tc.json), &data); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if data.Played != tc.wantPlayed {
				t.Errorf("Played = %v, want %v", data.Played, tc.wantPlayed)
			}
			assertPlayedPercentage(t, data.PlayedPercentage, tc.wantPct)
		})
	}
}

func checkCreateCollectionRequest(t *testing.T, r *http.Request, wantName, wantIDs, wantParentID string) {
	t.Helper()
	if r.Method != http.MethodPost || r.URL.Path != "/Collections" {
		t.Errorf("unexpected method/path: %s %s", r.Method, r.URL.Path)
	}
	q := r.URL.Query()
	if q.Get("name") != wantName || q.Get("ids") != wantIDs || q.Get("parentId") != wantParentID {
		t.Errorf("unexpected query: %v", q)
	}
}

func TestCreateCollection_Success(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkCreateCollectionRequest(t, r, "Favorites", "item-1,item-2,item-3", "parent-uuid")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jellyfin.CollectionCreationResult{ID: "created-col-uuid"})
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := c.CreateCollection(context.Background(), "Favorites", []string{"item-1", "item-2", "item-3"}, "parent-uuid")
	if err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}
	if res.ID != "created-col-uuid" {
		t.Errorf("ID = %q, want created-col-uuid", res.ID)
	}
}

func TestCreateCollection_Boundary(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("ids"); got != "single-id" {
			t.Errorf("ids = %q, want single-id", got)
		}
		if r.URL.Query().Has("parentId") {
			t.Errorf("parentId unexpectedly present: %q", r.URL.Query().Get("parentId"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jellyfin.CollectionCreationResult{ID: "col-1"})
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := c.CreateCollection(context.Background(), "Action", []string{"single-id"}, "")
	if err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}
	if res.ID != "col-1" {
		t.Errorf("ID = %q, want col-1", res.ID)
	}
}

func TestCreateCollection_EmptyName(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("request sent unexpectedly")
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.CreateCollection(context.Background(), "   ", []string{"item-1"}, "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !errors.Is(err, jellyfin.ErrEmptyName) {
		t.Errorf("err = %v, want ErrEmptyName", err)
	}
}

func TestCreateCollection_BlankItemID(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("request sent unexpectedly")
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.CreateCollection(context.Background(), "Col", []string{"a", " "}, "")
	if err == nil {
		t.Fatal("expected error for a blank item ID")
	}
	if !errors.Is(err, jellyfin.ErrEmptyItemIDs) {
		t.Errorf("err = %v, want ErrEmptyItemIDs", err)
	}
}

func TestCreateCollection_NoItems(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/Collections" {
			t.Errorf("request = %s %s, want POST /Collections", r.Method, r.URL.Path)
		}
		if _, ok := r.URL.Query()["ids"]; ok {
			t.Errorf("ids = %q, want the parameter omitted", r.URL.Query().Get("ids"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"Id": "col-empty"})
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got, err := c.CreateCollection(context.Background(), "Col", nil, "")
	if err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}
	if got.ID != "col-empty" {
		t.Errorf("ID = %q, want col-empty", got.ID)
	}
}

func TestCreateCollection_ServerError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.CreateCollection(context.Background(), "Col", []string{"item-1"}, "")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
}

func TestAddToCollection_Success(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.EscapedPath() != "/Collections/col%2F1/Items" {
			t.Errorf("EscapedPath = %q, want /Collections/col%%2F1/Items", r.URL.EscapedPath())
		}
		if got := r.URL.Query().Get("ids"); got != "id-1,id-2" {
			t.Errorf("ids = %q, want id-1,id-2", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.AddToCollection(context.Background(), "col/1", []string{"id-1", "id-2"}); err != nil {
		t.Fatalf("AddToCollection: %v", err)
	}
}

func TestAddToCollection_BoundarySingleID(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("ids"); got != "id-single" {
			t.Errorf("ids = %q, want id-single", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.AddToCollection(context.Background(), "col-123", []string{"id-single"}); err != nil {
		t.Fatalf("AddToCollection: %v", err)
	}
}

func TestAddToCollection_EmptyCollectionID(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("request sent unexpectedly")
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = c.AddToCollection(context.Background(), "", []string{"id-1"})
	if err == nil {
		t.Fatal("expected error for empty collectionID")
	}
	if !errors.Is(err, jellyfin.ErrEmptyCollectionID) {
		t.Errorf("err = %v, want ErrEmptyCollectionID", err)
	}
}

func TestAddToCollection_EmptyItemIDs(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("request sent unexpectedly")
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = c.AddToCollection(context.Background(), "col-1", []string{" "})
	if err == nil {
		t.Fatal("expected error for blank itemIDs")
	}
	if !errors.Is(err, jellyfin.ErrEmptyItemIDs) {
		t.Errorf("err = %v, want ErrEmptyItemIDs", err)
	}
}

func TestAddToCollection_NotFound(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "collection not found"})
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = c.AddToCollection(context.Background(), "col-missing", []string{"id-1"})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func TestRemoveFromCollection_Success(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.EscapedPath() != "/Collections/col%2F1/Items" {
			t.Errorf("EscapedPath = %q, want /Collections/col%%2F1/Items", r.URL.EscapedPath())
		}
		if got := r.URL.Query().Get("ids"); got != "id-1,id-2,id-3" {
			t.Errorf("ids = %q, want id-1,id-2,id-3", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.RemoveFromCollection(context.Background(), "col/1", []string{"id-1", "id-2", "id-3"}); err != nil {
		t.Fatalf("RemoveFromCollection: %v", err)
	}
}

func TestRemoveFromCollection_BoundarySingleID(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("ids"); got != "single-id" {
			t.Errorf("ids = %q, want single-id", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.RemoveFromCollection(context.Background(), "col-1", []string{"single-id"}); err != nil {
		t.Fatalf("RemoveFromCollection: %v", err)
	}
}

func TestRemoveFromCollection_EmptyCollectionID(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("request sent unexpectedly")
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = c.RemoveFromCollection(context.Background(), "   ", []string{"id-1"})
	if err == nil {
		t.Fatal("expected error for empty collectionID")
	}
	if !errors.Is(err, jellyfin.ErrEmptyCollectionID) {
		t.Errorf("err = %v, want ErrEmptyCollectionID", err)
	}
}

func TestRemoveFromCollection_EmptyItemIDs(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("request sent unexpectedly")
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = c.RemoveFromCollection(context.Background(), "col-1", []string{})
	if err == nil {
		t.Fatal("expected error for empty itemIDs")
	}
	if !errors.Is(err, jellyfin.ErrEmptyItemIDs) {
		t.Errorf("err = %v, want ErrEmptyItemIDs", err)
	}
}

func TestRemoveFromCollection_Forbidden(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "forbidden"})
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = c.RemoveFromCollection(context.Background(), "col-1", []string{"id-1"})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want 403", apiErr.StatusCode)
	}
}

func assertProviderIDs(t *testing.T, got map[string]string, wantNil bool, wantLen int, wantEntries map[string]string) {
	t.Helper()
	if wantNil {
		if got != nil {
			t.Errorf("ProviderIDs = %v, want nil", got)
		}
		return
	}
	if got == nil {
		t.Fatal("ProviderIDs = nil, want non-nil")
	}
	if len(got) != wantLen {
		t.Errorf("len(ProviderIDs) = %d, want %d", len(got), wantLen)
	}
	for k, wantVal := range wantEntries {
		gotVal, ok := got[k]
		if !ok {
			t.Errorf("ProviderIDs[%q] missing, want key present", k)
			continue
		}
		if gotVal != wantVal {
			t.Errorf("ProviderIDs[%q] = %q, want %q", k, gotVal, wantVal)
		}
	}
}

func TestBaseItemDto_ProviderIDs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		json        string
		wantNil     bool
		wantLen     int
		wantEntries map[string]string
	}{
		{
			name:    "negative field absent",
			json:    `{}`,
			wantNil: true,
		},
		{
			name:    "positive",
			json:    `{"ProviderIds":{"Tvdb":"81189","Tmdb":"1396","Imdb":"tt0903747"}}`,
			wantLen: 3,
			wantEntries: map[string]string{
				"Tvdb": "81189",
				"Tmdb": "1396",
				"Imdb": "tt0903747",
			},
		},
		{
			name:        "boundary empty object",
			json:        `{"ProviderIds":{}}`,
			wantLen:     0,
			wantEntries: map[string]string{},
		},
		{
			name:    "boundary mixed-case key",
			json:    `{"ProviderIds":{"TvDb":"1","tmdb":"2"}}`,
			wantLen: 2,
			wantEntries: map[string]string{
				"TvDb": "1",
				"tmdb": "2",
			},
		},
		{
			name:    "boundary null value",
			json:    `{"ProviderIds":{"Imdb":null}}`,
			wantLen: 1,
			wantEntries: map[string]string{
				"Imdb": "",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var item jellyfin.BaseItemDto
			if err := json.Unmarshal([]byte(tc.json), &item); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			assertProviderIDs(t, item.ProviderIDs, tc.wantNil, tc.wantLen, tc.wantEntries)
		})
	}
}

func assertMaxParentalRating(t *testing.T, got *int, wantNil bool, wantVal int) {
	t.Helper()
	if wantNil {
		if got != nil {
			t.Errorf("got %v, want nil", *got)
		}
		return
	}
	if got == nil {
		t.Fatal("got nil, want non-nil")
	}
	if *got != wantVal {
		t.Errorf("got %d, want %d", *got, wantVal)
	}
}

func TestUserPolicy_MaxParentalRating(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		jsonStr   string
		wantNil   bool
		wantValue int
	}{
		{name: "absent -> nil", jsonStr: `{"IsAdministrator":false,"EnableAllFolders":true}`, wantNil: true},
		{name: "0 -> non-nil 0", jsonStr: `{"MaxParentalRating":0,"EnableAllFolders":false}`, wantNil: false, wantValue: 0},
		{name: "17 -> 17", jsonStr: `{"MaxParentalRating":17,"EnableAllFolders":true}`, wantNil: false, wantValue: 17},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var p jellyfin.UserPolicy
			if err := json.Unmarshal([]byte(tc.jsonStr), &p); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			assertMaxParentalRating(t, p.MaxParentalRating, tc.wantNil, tc.wantValue)
		})
	}
}

func assertAuthResult(t *testing.T, res *jellyfin.AuthenticationResult) {
	t.Helper()
	if res.AccessToken != "test-user-token-456" {
		t.Errorf("AccessToken = %q, want test-user-token-456", res.AccessToken)
	}
	if res.ServerID != "srv-123" {
		t.Errorf("ServerID = %q, want srv-123", res.ServerID)
	}
	if res.User == nil || res.User.ID != "usr-999" || res.User.Name != "admin" {
		t.Errorf("User = %+v, want ID usr-999, Name admin", res.User)
	}
	if res.User.Policy == nil || res.User.Policy.MaxParentalRating == nil || *res.User.Policy.MaxParentalRating != 17 {
		t.Errorf("User.Policy.MaxParentalRating = %v, want 17", res.User.Policy)
	}
	if !res.User.Policy.EnableAllFolders {
		t.Errorf("User.Policy.EnableAllFolders = false, want true")
	}
}

func newAuthTestServer(t *testing.T, expectedToken string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Users/AuthenticateByName":
			var body struct {
				Username string `json:"Username"`
				Pw       string `json:"Pw"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Username != "admin" || body.Pw != "secret" {
				t.Errorf("unexpected body: %+v", body)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"AccessToken": "` + expectedToken + `",
				"ServerId": "srv-123",
				"User": {
					"Id": "usr-999",
					"Name": "admin",
					"Policy": {
						"IsAdministrator": true,
						"MaxParentalRating": 17,
						"EnableAllFolders": true
					}
				}
			}`))
		case "/System/Ping":
			auth := r.Header.Get("Authorization")
			if !strings.Contains(auth, `Token="`+expectedToken+`"`) {
				t.Errorf("Authorization header missing stored token: %q", auth)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`"pong"`))
		}
	}))
}

func TestAuthenticateUserByName(t *testing.T) {
	t.Parallel()

	ts := newAuthTestServer(t, "test-user-token-456")
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.AuthenticateUserByName(context.Background(), "admin", "secret")
	if err != nil {
		t.Fatalf("AuthenticateUserByName: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
	assertAuthResult(t, res)

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping follow-up: %v", err)
	}
}

func TestAuthenticateUserByName_Error(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "Invalid username or password"}`))
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.AuthenticateUserByName(context.Background(), "bad", "creds")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if res != nil {
		t.Errorf("res = %v, want nil", res)
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
}

func assertUserDto(t *testing.T, user *jellyfin.UserDto) {
	t.Helper()
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.ID != "usr-123" || user.Name != "alice" || user.ServerID != "srv-1" {
		t.Errorf("unexpected user: %+v", user)
	}
	if user.Policy == nil || len(user.Policy.AllowedTags) != 2 || user.Policy.AllowedTags[0] != "family" {
		t.Errorf("unexpected user policy: %+v", user.Policy)
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/Users/usr-123" {
			t.Errorf("path = %q, want /Users/usr-123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Id": "usr-123",
			"Name": "alice",
			"ServerId": "srv-1",
			"HasPassword": true,
			"Policy": {
				"IsAdministrator": false,
				"EnableAllFolders": false,
				"AllowedTags": ["family", "pg"],
				"BlockedMediaFolders": ["fld-1", "fld-2"]
			}
		}`))
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	user, err := c.GetUser(context.Background(), "usr-123")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	assertUserDto(t, user)
}

func TestGetUser_BlankInput(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("server was hit for path %q, expected no request", r.URL.Path)
	}))
	t.Cleanup(ts.Close)

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	blanks := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "spaces", input: "   "},
		{name: "tabs and newlines", input: "\t\n "},
	}

	for _, tc := range blanks {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			user, err := c.GetUser(context.Background(), tc.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if user != nil {
				t.Errorf("user = %v, want nil", user)
			}
			if !errors.Is(err, jellyfin.ErrEmptyUserID) {
				t.Errorf("err = %v, want errors.Is ErrEmptyUserID", err)
			}
		})
	}
}

func TestGetUser_NotFound(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "User not found"}`))
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	user, err := c.GetUser(context.Background(), "unknown-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if user != nil {
		t.Errorf("user = %v, want nil", user)
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func TestGetUser_PathEscaping(t *testing.T) {
	t.Parallel()

	hit := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		if r.URL.EscapedPath() != "/Users/a%2Fb" {
			t.Errorf("EscapedPath = %q, want /Users/a%%2Fb", r.URL.EscapedPath())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Id":"a/b","Name":"slashuser"}`))
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	user, err := c.GetUser(context.Background(), "a/b")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if !hit {
		t.Fatal("server was not hit")
	}
	if user.ID != "a/b" {
		t.Errorf("user.ID = %q, want a/b", user.ID)
	}
}

func runQuickConnectEnabledTest(t *testing.T, respJSON string, want bool) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/QuickConnect/Enabled" {
			t.Errorf("path = %q, want /QuickConnect/Enabled", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(respJSON))
	}))
	t.Cleanup(ts.Close)

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got, err := c.QuickConnectEnabled(context.Background())
	if err != nil {
		t.Fatalf("QuickConnectEnabled: %v", err)
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestQuickConnectEnabled(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		respJSON string
		want     bool
	}{
		{name: "enabled true", respJSON: "true", want: true},
		{name: "enabled false", respJSON: "false", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runQuickConnectEnabledTest(t, tc.respJSON, tc.want)
		})
	}
}

func TestQuickConnectEnabled_Disabled503(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"message": "Quick connect is unavailable"}`))
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := c.QuickConnectEnabled(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got {
		t.Errorf("got = true, want false")
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("StatusCode = %d, want 503", apiErr.StatusCode)
	}
}

func assertInitiateResult(t *testing.T, res *jellyfin.QuickConnectResult) {
	t.Helper()
	if res == nil {
		t.Fatal("expected non-nil result")
	}
	if res.Authenticated {
		t.Errorf("Authenticated = true, want false")
	}
	if res.Secret != "sec-init-123" || res.Code != "654321" {
		t.Errorf("Secret/Code = %q / %q, want sec-init-123 / 654321", res.Secret, res.Code)
	}
	if res.DeviceID != "dev-456" || res.DeviceName != "Living Room TV" {
		t.Errorf("DeviceID/Name = %q / %q, want dev-456 / Living Room TV", res.DeviceID, res.DeviceName)
	}
	if res.AppName != "Jellyfin Web" || res.AppVersion != "10.8.13" {
		t.Errorf("AppName/Version = %q / %q, want Jellyfin Web / 10.8.13", res.AppName, res.AppVersion)
	}
	if res.DateAdded != "2026-10-26T01:23:45Z" {
		t.Errorf("DateAdded = %q, want 2026-10-26T01:23:45Z", res.DateAdded)
	}
}

func TestInitiateQuickConnect(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/QuickConnect/Initiate" {
			t.Errorf("path = %q, want /QuickConnect/Initiate", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			t.Errorf("expected empty body, got %q", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Authenticated": false,
			"Secret": "sec-init-123",
			"Code": "654321",
			"DeviceId": "dev-456",
			"DeviceName": "Living Room TV",
			"AppName": "Jellyfin Web",
			"AppVersion": "10.8.13",
			"DateAdded": "2026-10-26T01:23:45Z"
		}`))
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.InitiateQuickConnect(context.Background())
	if err != nil {
		t.Fatalf("InitiateQuickConnect: %v", err)
	}
	assertInitiateResult(t, res)
}

func TestInitiateQuickConnect_Error401(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "Unauthorized"}`))
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.InitiateQuickConnect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if res != nil {
		t.Errorf("res = %v, want nil", res)
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
}

func runGetQuickConnectStateTest(t *testing.T, secret, respJSON string, wantAuth bool) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/QuickConnect/Connect" {
			t.Errorf("path = %q, want /QuickConnect/Connect", r.URL.Path)
		}
		if r.URL.Query().Get("secret") != secret {
			t.Errorf("secret query = %q, want %q", r.URL.Query().Get("secret"), secret)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(respJSON))
	}))
	t.Cleanup(ts.Close)

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.GetQuickConnectState(context.Background(), secret)
	if err != nil {
		t.Fatalf("GetQuickConnectState: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
	if res.Authenticated != wantAuth {
		t.Errorf("Authenticated = %v, want %v", res.Authenticated, wantAuth)
	}
	if res.Secret != secret {
		t.Errorf("Secret = %q, want %q", res.Secret, secret)
	}
}

func TestGetQuickConnectState(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name              string
		secret            string
		respJSON          string
		wantAuthenticated bool
	}{
		{
			name:              "authenticated false",
			secret:            "sec-pending-1",
			respJSON:          `{"Authenticated":false,"Secret":"sec-pending-1","Code":"111222"}`,
			wantAuthenticated: false,
		},
		{
			name:              "authenticated true",
			secret:            "sec-authed-2",
			respJSON:          `{"Authenticated":true,"Secret":"sec-authed-2","Code":"333444"}`,
			wantAuthenticated: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runGetQuickConnectStateTest(t, tc.secret, tc.respJSON, tc.wantAuthenticated)
		})
	}
}

func TestGetQuickConnectState_BlankSecret(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("server was hit for path %q, expected no request", r.URL.Path)
	}))
	t.Cleanup(ts.Close)

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	blanks := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "spaces", input: "   "},
		{name: "tabs and newlines", input: "\t\n "},
	}

	for _, tc := range blanks {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := c.GetQuickConnectState(context.Background(), tc.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if res != nil {
				t.Errorf("res = %v, want nil", res)
			}
			if !errors.Is(err, jellyfin.ErrEmptySecret) {
				t.Errorf("err = %v, want errors.Is ErrEmptySecret", err)
			}
		})
	}
}

func TestGetQuickConnectState_NotFound(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Secret not found"}`))
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.GetQuickConnectState(context.Background(), "unknown-secret")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if res != nil {
		t.Errorf("res = %v, want nil", res)
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func assertAuthorizeQueryParams(t *testing.T, r *http.Request, code, userID string, wantHasUID bool) {
	t.Helper()
	if r.URL.Query().Get("code") != code {
		t.Errorf("code query = %q, want %q", r.URL.Query().Get("code"), code)
	}
	if wantHasUID {
		if r.URL.Query().Get("userId") != userID {
			t.Errorf("userId query = %q, want %q", r.URL.Query().Get("userId"), userID)
		}
		return
	}
	if r.URL.Query().Has("userId") {
		t.Errorf("userId query present unexpectedly: %q", r.URL.Query().Get("userId"))
	}
}

func runAuthorizeQuickConnectTest(t *testing.T, code, userID string, wantHasUID bool, respBody string, wantResult bool) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/QuickConnect/Authorize" {
			t.Errorf("path = %q, want /QuickConnect/Authorize", r.URL.Path)
		}
		assertAuthorizeQueryParams(t, r, code, userID, wantHasUID)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(respBody))
	}))
	t.Cleanup(ts.Close)

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ok, err := c.AuthorizeQuickConnect(context.Background(), code, userID)
	if err != nil {
		t.Fatalf("AuthorizeQuickConnect: %v", err)
	}
	if ok != wantResult {
		t.Errorf("AuthorizeQuickConnect = %v, want %v", ok, wantResult)
	}
}

func TestAuthorizeQuickConnect(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		code       string
		userID     string
		wantHasUID bool
		respBody   string
		wantResult bool
	}{
		{name: "with userID true", code: "123456", userID: "usr-abc", wantHasUID: true, respBody: "true", wantResult: true},
		{name: "without userID true", code: "654321", userID: "", wantHasUID: false, respBody: "true", wantResult: true},
		{name: "with userID false", code: "111111", userID: "usr-xyz", wantHasUID: true, respBody: "false", wantResult: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runAuthorizeQuickConnectTest(t, tc.code, tc.userID, tc.wantHasUID, tc.respBody, tc.wantResult)
		})
	}
}

func TestAuthorizeQuickConnect_BlankCode(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("server was hit for path %q, expected no request", r.URL.Path)
	}))
	t.Cleanup(ts.Close)

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	blanks := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "spaces", input: "   "},
		{name: "tabs and newlines", input: "\t\n "},
	}

	for _, tc := range blanks {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ok, err := c.AuthorizeQuickConnect(context.Background(), tc.input, "user1")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if ok {
				t.Errorf("ok = true, want false")
			}
			if !errors.Is(err, jellyfin.ErrEmptyCode) {
				t.Errorf("err = %v, want errors.Is ErrEmptyCode", err)
			}
		})
	}
}

func runAuthorizeErrorTest(t *testing.T, statusCode int) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(`{"message": "error"}`))
	}))
	t.Cleanup(ts.Close)

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ok, err := c.AuthorizeQuickConnect(context.Background(), "123456", "user1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if ok {
		t.Errorf("ok = true, want false")
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != statusCode {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, statusCode)
	}
}

func TestAuthorizeQuickConnect_Errors(t *testing.T) {
	t.Parallel()

	codes := []struct {
		name       string
		statusCode int
	}{
		{name: "401 unauthorized", statusCode: http.StatusUnauthorized},
		{name: "403 forbidden", statusCode: http.StatusForbidden},
	}

	for _, tc := range codes {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runAuthorizeErrorTest(t, tc.statusCode)
		})
	}
}

func newQuickConnectAuthServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Users/AuthenticateWithQuickConnect":
			var body struct {
				Secret string `json:"Secret"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Secret != "sec-authorized-789" {
				t.Errorf("Secret = %q, want sec-authorized-789", body.Secret)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"AccessToken": "qc-jwt-token-999",
				"ServerId": "srv-qc",
				"User": {"Id": "usr-qc", "Name": "qcuser"}
			}`))
		case "/System/Ping":
			auth := r.Header.Get("Authorization")
			if !strings.Contains(auth, `Token="qc-jwt-token-999"`) {
				t.Errorf("Authorization header missing stored token: %q", auth)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`"pong"`))
		}
	}))
}

func TestAuthenticateWithQuickConnect(t *testing.T) {
	t.Parallel()

	ts := newQuickConnectAuthServer(t)
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.AuthenticateWithQuickConnect(context.Background(), "sec-authorized-789")
	if err != nil {
		t.Fatalf("AuthenticateWithQuickConnect: %v", err)
	}
	if res == nil || res.AccessToken != "qc-jwt-token-999" || res.ServerID != "srv-qc" {
		t.Fatalf("unexpected res: %+v", res)
	}
	if res.User == nil || res.User.ID != "usr-qc" || res.User.Name != "qcuser" {
		t.Errorf("User = %+v, want ID usr-qc, Name qcuser", res.User)
	}

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping follow-up: %v", err)
	}
}

func TestAuthenticateWithQuickConnect_BlankSecret(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("server was hit for path %q, expected no request", r.URL.Path)
	}))
	t.Cleanup(ts.Close)

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	blanks := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "spaces", input: "   "},
		{name: "tabs and newlines", input: "\t\n "},
	}

	for _, tc := range blanks {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := c.AuthenticateWithQuickConnect(context.Background(), tc.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if res != nil {
				t.Errorf("res = %v, want nil", res)
			}
			if !errors.Is(err, jellyfin.ErrEmptySecret) {
				t.Errorf("err = %v, want errors.Is ErrEmptySecret", err)
			}
		})
	}
}

func TestAuthenticateWithQuickConnect_BadRequest400(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "Invalid secret"}`))
	}))
	defer ts.Close()

	c, err := jellyfin.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.AuthenticateWithQuickConnect(context.Background(), "invalid-secret")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if res != nil {
		t.Errorf("res = %v, want nil", res)
	}
	var apiErr *jellyfin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *jellyfin.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
}

func assertPolicyRatings(t *testing.T, p *jellyfin.UserPolicy) {
	t.Helper()
	if p.MaxParentalRating == nil || *p.MaxParentalRating != 12 {
		t.Errorf("MaxParentalRating = %v, want 12", p.MaxParentalRating)
	}
	if p.MaxParentalSubRating == nil || *p.MaxParentalSubRating != 3 {
		t.Errorf("MaxParentalSubRating = %v, want 3", p.MaxParentalSubRating)
	}
}

func assertPolicyFoldersAndTags(t *testing.T, p *jellyfin.UserPolicy) {
	t.Helper()
	if p.EnableAllFolders {
		t.Errorf("EnableAllFolders = true, want false")
	}
	if len(p.AllowedTags) != 2 || p.AllowedTags[0] != "safe" || p.AllowedTags[1] != "family" {
		t.Errorf("AllowedTags = %v", p.AllowedTags)
	}
	if len(p.EnabledFolders) != 2 || p.EnabledFolders[0] != "fld-a" || p.EnabledFolders[1] != "fld-b" {
		t.Errorf("EnabledFolders = %v", p.EnabledFolders)
	}
	if len(p.BlockedMediaFolders) != 1 || p.BlockedMediaFolders[0] != "fld-c" {
		t.Errorf("BlockedMediaFolders = %v", p.BlockedMediaFolders)
	}
}

func TestUserPolicy_FullDecode(t *testing.T) {
	t.Parallel()

	jsonStr := `{
		"IsAdministrator": true,
		"IsHidden": false,
		"IsDisabled": false,
		"BlockedTags": ["tag1"],
		"EnableContentDeletion": true,
		"MaxParentalRating": 12,
		"MaxParentalSubRating": 3,
		"AllowedTags": ["safe", "family"],
		"EnableAllFolders": false,
		"EnabledFolders": ["fld-a", "fld-b"],
		"BlockedMediaFolders": ["fld-c"]
	}`

	var p jellyfin.UserPolicy
	if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if !p.IsAdministrator || p.IsHidden || p.IsDisabled || !p.EnableContentDeletion {
		t.Errorf("base flags mismatch: %+v", p)
	}
	if len(p.BlockedTags) != 1 || p.BlockedTags[0] != "tag1" {
		t.Errorf("BlockedTags = %v", p.BlockedTags)
	}
	assertPolicyRatings(t, &p)
	assertPolicyFoldersAndTags(t, &p)
}
