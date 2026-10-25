package jellyfin_test

import (
	"context"
	"encoding/json"
	"errors"
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
