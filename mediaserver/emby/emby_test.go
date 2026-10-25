package emby_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/mediaserver/emby/v3"
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
		if r.URL.Path != "/emby/Users/AuthenticateByName" {
			t.Errorf("path = %q, want /emby/Users/AuthenticateByName", r.URL.Path)
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
		_ = json.NewEncoder(w).Encode(emby.AuthenticationResult{
			AccessToken: "test-token-123",
			ServerID:    "server-1",
		})
	}))
	defer ts.Close()

	c, err := emby.New(ts.URL)
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

	c, err := emby.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = c.AuthenticateByName(context.Background(), "bad", "creds")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *emby.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *emby.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
}

func TestGetPublicSystemInfo(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/emby/System/Info/Public", emby.PublicSystemInfo{
		ServerName:   "My Emby Server",
		Version:      "4.7.11.0",
		ID:           "abc123",
		LocalAddress: "192.168.1.100",
	})
	defer ts.Close()

	c, err := emby.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	info, err := c.GetPublicSystemInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.ServerName != "My Emby Server" {
		t.Errorf("ServerName = %q, want My Emby Server", info.ServerName)
	}
	if info.Version != "4.7.11.0" {
		t.Errorf("Version = %q, want 4.7.11.0", info.Version)
	}
}

func TestGetSystemInfo(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/emby/System/Info", emby.SystemInfo{
		ServerName:                 "My Server",
		Version:                    "4.7.11.0",
		ID:                         "abc123",
		OperatingSystem:            "Linux",
		OperatingSystemDisplayName: "Debian GNU/Linux 12",
	})
	defer ts.Close()

	c, err := emby.New(ts.URL, emby.WithAccessToken("token"))
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

	ts := newTestServer(t, "/emby/System/Ping", "")
	defer ts.Close()

	c, err := emby.New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestGetUsers(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, "/emby/Users", []emby.UserDto{
		{Name: "admin", ID: "user-1", HasPassword: true},
		{Name: "guest", ID: "user-2", HasPassword: false},
	})
	defer ts.Close()

	c, err := emby.New(ts.URL, emby.WithAccessToken("token"))
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

	ts := newTestServer(t, "/emby/Users/Me", emby.UserDto{
		Name:        "admin",
		ID:          "user-1",
		ServerID:    "server-1",
		HasPassword: true,
	})
	defer ts.Close()

	c, err := emby.New(ts.URL, emby.WithAccessToken("token"))
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

	ts := newTestServer(t, "/emby/Sessions", []emby.SessionInfoDto{
		{
			ID:         "session-1",
			UserID:     "user-1",
			UserName:   "admin",
			DeviceName: "Chrome",
			Client:     "Emby Web",
			IsActive:   true,
		},
	})
	defer ts.Close()

	c, err := emby.New(ts.URL, emby.WithAccessToken("token"))
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

	ts := newTestServer(t, "/emby/Items", emby.ItemsResult{
		Items: []emby.BaseItemDto{
			{ID: "item-1", Name: "Inception", Type: "Movie"},
			{ID: "item-2", Name: "Interstellar", Type: "Movie"},
		},
		TotalRecordCount: 2,
		StartIndex:       0,
	})
	defer ts.Close()

	c, err := emby.New(ts.URL, emby.WithAccessToken("token"))
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
		if r.URL.Path != "/emby/Items" {
			t.Errorf("path = %q, want /emby/Items", r.URL.Path)
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
		_ = json.NewEncoder(w).Encode(emby.ItemsResult{
			Items:            []emby.BaseItemDto{{ID: "item-1", Name: "Movie 1"}},
			TotalRecordCount: 100,
			StartIndex:       0,
		})
	}))
	defer ts.Close()

	c, err := emby.New(ts.URL, emby.WithAccessToken("token"))
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

	ts := newTestServer(t, "/emby/Items/item-1", emby.BaseItemDto{
		ID:           "item-1",
		Name:         "The Matrix",
		Type:         "Movie",
		RunTimeTicks: 8208000000,
		Genres:       []string{"Action", "Sci-Fi"},
	})
	defer ts.Close()

	c, err := emby.New(ts.URL, emby.WithAccessToken("token"))
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
		if r.URL.Path != "/emby/UserViews" {
			t.Errorf("path = %q, want /emby/UserViews", r.URL.Path)
		}
		if got := r.URL.Query().Get("userId"); got != "user-1" {
			t.Errorf("userId = %q, want user-1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(emby.ItemsResult{
			Items: []emby.BaseItemDto{
				{ID: "lib-1", Name: "Movies", Type: "CollectionFolder", IsFolder: true},
				{ID: "lib-2", Name: "TV Shows", Type: "CollectionFolder", IsFolder: true},
			},
			TotalRecordCount: 2,
		})
	}))
	defer ts.Close()

	c, err := emby.New(ts.URL, emby.WithAccessToken("token"))
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

	c, err := emby.New(ts.URL, emby.WithAccessToken("bad"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.GetUsers(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *emby.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *emby.APIError, got %T", err)
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
		err  emby.APIError
		want string
	}{
		{"with message", emby.APIError{StatusCode: 401, Message: "Unauthorized"}, "emby: HTTP 401: Unauthorized"},
		{"raw body", emby.APIError{StatusCode: 502, RawBody: "gateway error"}, "emby: HTTP 502: gateway error"},
		{"code only", emby.APIError{StatusCode: 500}, "emby: HTTP 500"},
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

	ts := newTestServer(t, "/emby/Users", []emby.UserDto{})
	defer ts.Close()

	c, err := emby.New(ts.URL, emby.WithAccessToken("token"))
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
		if got := r.Header.Get("X-Emby-Device-Id"); got != "my-device-123" {
			t.Errorf("X-Emby-Device-Id = %q, want my-device-123", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]emby.UserDto{})
	}))
	defer ts.Close()

	c, err := emby.New(
		ts.URL,
		emby.WithAccessToken("token"),
		emby.WithDeviceID("my-device-123"),
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
		_ = json.NewEncoder(w).Encode([]emby.UserDto{})
	}))
	defer ts.Close()

	c, err := emby.New(
		ts.URL,
		emby.WithAccessToken("token"),
		emby.WithUserAgent("myapp/1.2.3"),
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
			c, err := emby.New(tc.url)
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
			var data emby.UserItemDataDto
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
			var item emby.BaseItemDto
			if err := json.Unmarshal([]byte(tc.json), &item); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			assertProviderIDs(t, item.ProviderIDs, tc.wantNil, tc.wantLen, tc.wantEntries)
		})
	}
}
