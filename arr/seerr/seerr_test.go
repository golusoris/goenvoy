package seerr_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/arr/seerr"
	"github.com/golusoris/goenvoy/arr/v2"
)

func newTestServer(t *testing.T, method, wantPath string, body any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			t.Errorf("method = %s, want %s", r.Method, method)
		}
		if r.Header.Get("X-Api-Key") == "" {
			t.Error("missing X-Api-Key header")
		}
		if wantPath != "" && r.URL.RequestURI() != wantPath {
			t.Errorf("path = %q, want %q", r.URL.RequestURI(), wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if body != nil {
			json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		c, err := seerr.New("http://localhost:5055", "test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		t.Parallel()
		_, err := seerr.New("://bad", "test-key")
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})
}

func TestGetStatus(t *testing.T) {
	t.Parallel()

	want := seerr.StatusResponse{Version: "1.33.2", UpdateAvailable: false}

	srv := newTestServer(t, http.MethodGet, "/api/v1/status", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if got.Version != "1.33.2" {
		t.Errorf("Version = %q, want %q", got.Version, "1.33.2")
	}
}

func TestGetMe(t *testing.T) {
	t.Parallel()

	want := seerr.User{ID: 1, Email: "admin@example.com", Permissions: 2}

	srv := newTestServer(t, http.MethodGet, "/api/v1/auth/me", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if got.Email != "admin@example.com" {
		t.Errorf("Email = %q, want %q", got.Email, "admin@example.com")
	}
}

func TestSearch(t *testing.T) {
	t.Parallel()

	want := seerr.SearchResults{Page: 1, TotalPages: 5, TotalResults: 100}

	srv := newTestServer(t, http.MethodGet, "/api/v1/search?query=Mulan&page=1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.Search(context.Background(), "Mulan", 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if got.TotalResults != 100 {
		t.Errorf("TotalResults = %d, want 100", got.TotalResults)
	}
}

func TestDiscoverMovies(t *testing.T) {
	t.Parallel()

	want := seerr.SearchResults{Page: 1, TotalPages: 10, TotalResults: 200}

	srv := newTestServer(t, http.MethodGet, "/api/v1/discover/movies?page=1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.DiscoverMovies(context.Background(), 1)
	if err != nil {
		t.Fatalf("DiscoverMovies: %v", err)
	}
	if got.TotalResults != 200 {
		t.Errorf("TotalResults = %d, want 200", got.TotalResults)
	}
}

func TestDiscoverTrending(t *testing.T) {
	t.Parallel()

	want := seerr.SearchResults{Page: 1, TotalPages: 1, TotalResults: 20}

	srv := newTestServer(t, http.MethodGet, "/api/v1/discover/trending?page=1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.DiscoverTrending(context.Background(), 1)
	if err != nil {
		t.Fatalf("DiscoverTrending: %v", err)
	}
	if got.TotalResults != 20 {
		t.Errorf("TotalResults = %d, want 20", got.TotalResults)
	}
}

func TestGetRequests(t *testing.T) {
	t.Parallel()

	want := struct {
		PageInfo seerr.PageInfo       `json:"pageInfo"`
		Results  []seerr.MediaRequest `json:"results"`
	}{
		PageInfo: seerr.PageInfo{Page: 1, Pages: 1, Results: 1},
		Results: []seerr.MediaRequest{
			{ID: 1, Status: 2, Is4k: false},
		},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/request?take=20&skip=0&filter=pending", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	results, pageInfo, err := c.GetRequests(context.Background(), 20, 0, "pending")
	if err != nil {
		t.Fatalf("GetRequests: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	if pageInfo.Results != 1 {
		t.Errorf("PageInfo.Results = %d, want 1", pageInfo.Results)
	}
}

func TestCreateRequest(t *testing.T) {
	t.Parallel()

	want := seerr.MediaRequest{ID: 5, Status: 1}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var body seerr.CreateRequestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.MediaType != "movie" {
			t.Errorf("MediaType = %q, want %q", body.MediaType, "movie")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.CreateRequest(context.Background(), &seerr.CreateRequestBody{
		MediaType: "movie",
		MediaID:   337401,
	})
	if err != nil {
		t.Fatalf("CreateRequest: %v", err)
	}
	if got.ID != 5 {
		t.Errorf("ID = %d, want 5", got.ID)
	}
}

func TestDeleteRequest(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v1/request/1", nil)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	if err := c.DeleteRequest(context.Background(), 1); err != nil {
		t.Fatalf("DeleteRequest: %v", err)
	}
}

func TestApproveRequest(t *testing.T) {
	t.Parallel()

	want := seerr.MediaRequest{ID: 1, Status: 2}

	srv := newTestServer(t, http.MethodPost, "/api/v1/request/1/approve", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.ApproveRequest(context.Background(), 1)
	if err != nil {
		t.Fatalf("ApproveRequest: %v", err)
	}
	if got.Status != 2 {
		t.Errorf("Status = %d, want 2", got.Status)
	}
}

func TestDeclineRequest(t *testing.T) {
	t.Parallel()

	want := seerr.MediaRequest{ID: 1, Status: 3}

	srv := newTestServer(t, http.MethodPost, "/api/v1/request/1/decline", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.DeclineRequest(context.Background(), 1)
	if err != nil {
		t.Fatalf("DeclineRequest: %v", err)
	}
	if got.Status != 3 {
		t.Errorf("Status = %d, want 3", got.Status)
	}
}

func TestGetRequestCount(t *testing.T) {
	t.Parallel()

	want := seerr.RequestCount{Total: 50, Movie: 30, TV: 20, Pending: 5}

	srv := newTestServer(t, http.MethodGet, "/api/v1/request/count", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetRequestCount(context.Background())
	if err != nil {
		t.Fatalf("GetRequestCount: %v", err)
	}
	if got.Total != 50 {
		t.Errorf("Total = %d, want 50", got.Total)
	}
}

func TestGetMovie(t *testing.T) {
	t.Parallel()

	want := seerr.MovieDetails{
		ID: 337401, Title: "Mulan", VoteAverage: 7.0,
		Genres: []seerr.Genre{{ID: 28, Name: "Action"}},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/movie/337401", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetMovie(context.Background(), 337401)
	if err != nil {
		t.Fatalf("GetMovie: %v", err)
	}
	if got.Title != "Mulan" {
		t.Errorf("Title = %q, want %q", got.Title, "Mulan")
	}
}

func TestGetTV(t *testing.T) {
	t.Parallel()

	want := seerr.TvDetails{
		ID: 76479, Name: "The Boys", Status: "Returning Series",
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/tv/76479", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetTV(context.Background(), 76479)
	if err != nil {
		t.Fatalf("GetTV: %v", err)
	}
	if got.Name != "The Boys" {
		t.Errorf("Name = %q, want %q", got.Name, "The Boys")
	}
}

func TestGetTVSeason(t *testing.T) {
	t.Parallel()

	want := seerr.Season{
		ID: 1, Name: "Season 1", SeasonNumber: 1,
		Episodes: []seerr.Episode{{ID: 100, Name: "Pilot", EpisodeNumber: 1, SeasonNumber: 1}},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/tv/76479/season/1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetTVSeason(context.Background(), 76479, 1)
	if err != nil {
		t.Fatalf("GetTVSeason: %v", err)
	}
	if len(got.Episodes) != 1 {
		t.Fatalf("Episodes len = %d, want 1", len(got.Episodes))
	}
	if got.Episodes[0].Name != "Pilot" {
		t.Errorf("Episode Name = %q, want %q", got.Episodes[0].Name, "Pilot")
	}
}

func TestGetMedia(t *testing.T) {
	t.Parallel()

	want := struct {
		PageInfo seerr.PageInfo    `json:"pageInfo"`
		Results  []seerr.MediaInfo `json:"results"`
	}{
		PageInfo: seerr.PageInfo{Page: 1, Pages: 1, Results: 2},
		Results: []seerr.MediaInfo{
			{ID: 1, TmdbID: 337401, Status: 5},
			{ID: 2, TmdbID: 76479, Status: 3},
		},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/media?take=20&skip=0", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	results, pageInfo, err := c.GetMedia(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("GetMedia: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len = %d, want 2", len(results))
	}
	if pageInfo.Results != 2 {
		t.Errorf("PageInfo.Results = %d, want 2", pageInfo.Results)
	}
}

func TestDeleteMedia(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v1/media/1", nil)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	if err := c.DeleteMedia(context.Background(), 1); err != nil {
		t.Fatalf("DeleteMedia: %v", err)
	}
}

func TestGetUsers(t *testing.T) {
	t.Parallel()

	want := struct {
		PageInfo seerr.PageInfo `json:"pageInfo"`
		Results  []seerr.User   `json:"results"`
	}{
		PageInfo: seerr.PageInfo{Page: 1, Pages: 1, Results: 1},
		Results:  []seerr.User{{ID: 1, Email: "admin@example.com"}},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/user?take=20&skip=0", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	results, _, err := c.GetUsers(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}
	if results[0].Email != "admin@example.com" {
		t.Errorf("Email = %q, want %q", results[0].Email, "admin@example.com")
	}
}

func TestGetUserQuota(t *testing.T) {
	t.Parallel()

	want := seerr.UserQuota{
		Movie: &seerr.QuotaDetail{Days: 7, Limit: 10, Used: 3, Remaining: 7},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/user/1/quota", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetUserQuota(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUserQuota: %v", err)
	}
	if got.Movie.Remaining != 7 {
		t.Errorf("Movie.Remaining = %d, want 7", got.Movie.Remaining)
	}
}

func TestGetIssues(t *testing.T) {
	t.Parallel()

	want := struct {
		PageInfo seerr.PageInfo `json:"pageInfo"`
		Results  []seerr.Issue  `json:"results"`
	}{
		PageInfo: seerr.PageInfo{Page: 1, Pages: 1, Results: 1},
		Results:  []seerr.Issue{{ID: 1, IssueType: 1}},
	}

	srv := newTestServer(t, http.MethodGet, "/api/v1/issue?take=20&skip=0&filter=open", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	results, _, err := c.GetIssues(context.Background(), 20, 0, "open")
	if err != nil {
		t.Fatalf("GetIssues: %v", err)
	}
	if results[0].IssueType != 1 {
		t.Errorf("IssueType = %d, want 1", results[0].IssueType)
	}
}

func TestCreateIssue(t *testing.T) {
	t.Parallel()

	want := seerr.Issue{ID: 10, IssueType: 2}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var body seerr.CreateIssueBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.IssueType != 2 {
			t.Errorf("IssueType = %d, want 2", body.IssueType)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.CreateIssue(context.Background(), &seerr.CreateIssueBody{
		IssueType: 2,
		Message:   "Audio sync issue",
		MediaID:   337401,
	})
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if got.ID != 10 {
		t.Errorf("ID = %d, want 10", got.ID)
	}
}

func TestAddIssueComment(t *testing.T) {
	t.Parallel()

	want := seerr.Issue{ID: 1, IssueType: 1}

	srv := newTestServer(t, http.MethodPost, "/api/v1/issue/1/comment", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.AddIssueComment(context.Background(), 1, "Still broken")
	if err != nil {
		t.Fatalf("AddIssueComment: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetIssueCount(t *testing.T) {
	t.Parallel()

	want := seerr.IssueCount{Total: 15, Open: 10, Closed: 5, Video: 3, Audio: 2, Subtitles: 5, Others: 5}

	srv := newTestServer(t, http.MethodGet, "/api/v1/issue/count", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetIssueCount(context.Background())
	if err != nil {
		t.Fatalf("GetIssueCount: %v", err)
	}
	if got.Open != 10 {
		t.Errorf("Open = %d, want 10", got.Open)
	}
}

func TestErrorResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "bad-key")
	_, err := c.GetStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}

	var apiErr *arr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *arr.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestDiscoverTV(t *testing.T) {
	t.Parallel()

	want := seerr.SearchResults{Page: 1, TotalPages: 10, TotalResults: 200}

	srv := newTestServer(t, http.MethodGet, "/api/v1/discover/tv?page=1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.DiscoverTV(context.Background(), 1)
	if err != nil {
		t.Fatalf("DiscoverTV: %v", err)
	}
	if got.TotalResults != 200 {
		t.Errorf("TotalResults = %d, want 200", got.TotalResults)
	}
}

func TestDiscoverUpcomingMovies(t *testing.T) {
	t.Parallel()

	want := seerr.SearchResults{Page: 1, TotalResults: 50}

	srv := newTestServer(t, http.MethodGet, "/api/v1/discover/movies/upcoming?page=1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.DiscoverUpcomingMovies(context.Background(), 1)
	if err != nil {
		t.Fatalf("DiscoverUpcomingMovies: %v", err)
	}
	if got.TotalResults != 50 {
		t.Errorf("TotalResults = %d", got.TotalResults)
	}
}

func TestDiscoverUpcomingTV(t *testing.T) {
	t.Parallel()

	want := seerr.SearchResults{Page: 1, TotalResults: 30}

	srv := newTestServer(t, http.MethodGet, "/api/v1/discover/tv/upcoming?page=1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.DiscoverUpcomingTV(context.Background(), 1)
	if err != nil {
		t.Fatalf("DiscoverUpcomingTV: %v", err)
	}
	if got.TotalResults != 30 {
		t.Errorf("TotalResults = %d", got.TotalResults)
	}
}

func TestGetRequest(t *testing.T) {
	t.Parallel()

	want := seerr.MediaRequest{ID: 1, Status: 2}

	srv := newTestServer(t, http.MethodGet, "/api/v1/request/1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetRequest(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetRequest: %v", err)
	}
	if got.Status != 2 {
		t.Errorf("Status = %d", got.Status)
	}
}

func TestRetryRequest(t *testing.T) {
	t.Parallel()

	want := seerr.MediaRequest{ID: 1, Status: 1}

	srv := newTestServer(t, http.MethodPost, "/api/v1/request/1/retry", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.RetryRequest(context.Background(), 1)
	if err != nil {
		t.Fatalf("RetryRequest: %v", err)
	}
	if got.Status != 1 {
		t.Errorf("Status = %d", got.Status)
	}
}

func TestGetMovieRecommendations(t *testing.T) {
	t.Parallel()

	want := seerr.SearchResults{Page: 1, TotalResults: 10}

	srv := newTestServer(t, http.MethodGet, "/api/v1/movie/337401/recommendations?page=1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetMovieRecommendations(context.Background(), 337401, 1)
	if err != nil {
		t.Fatalf("GetMovieRecommendations: %v", err)
	}
	if got.TotalResults != 10 {
		t.Errorf("TotalResults = %d", got.TotalResults)
	}
}

func TestGetMovieSimilar(t *testing.T) {
	t.Parallel()

	want := seerr.SearchResults{Page: 1, TotalResults: 15}

	srv := newTestServer(t, http.MethodGet, "/api/v1/movie/337401/similar?page=1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetMovieSimilar(context.Background(), 337401, 1)
	if err != nil {
		t.Fatalf("GetMovieSimilar: %v", err)
	}
	if got.TotalResults != 15 {
		t.Errorf("TotalResults = %d", got.TotalResults)
	}
}

func TestGetTVRecommendations(t *testing.T) {
	t.Parallel()

	want := seerr.SearchResults{Page: 1, TotalResults: 8}

	srv := newTestServer(t, http.MethodGet, "/api/v1/tv/76479/recommendations?page=1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetTVRecommendations(context.Background(), 76479, 1)
	if err != nil {
		t.Fatalf("GetTVRecommendations: %v", err)
	}
	if got.TotalResults != 8 {
		t.Errorf("TotalResults = %d", got.TotalResults)
	}
}

func TestGetTVSimilar(t *testing.T) {
	t.Parallel()

	want := seerr.SearchResults{Page: 1, TotalResults: 12}

	srv := newTestServer(t, http.MethodGet, "/api/v1/tv/76479/similar?page=1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetTVSimilar(context.Background(), 76479, 1)
	if err != nil {
		t.Fatalf("GetTVSimilar: %v", err)
	}
	if got.TotalResults != 12 {
		t.Errorf("TotalResults = %d", got.TotalResults)
	}
}

func TestUpdateMediaStatus(t *testing.T) {
	t.Parallel()

	want := seerr.MediaInfo{ID: 1, Status: 5}

	srv := newTestServer(t, http.MethodPost, "/api/v1/media/1/available", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.UpdateMediaStatus(context.Background(), 1, "available", false)
	if err != nil {
		t.Fatalf("UpdateMediaStatus: %v", err)
	}
	if got.Status != 5 {
		t.Errorf("Status = %d", got.Status)
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	want := seerr.User{ID: 1, Email: "admin@example.com"}

	srv := newTestServer(t, http.MethodGet, "/api/v1/user/1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got.Email != "admin@example.com" {
		t.Errorf("Email = %q", got.Email)
	}
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v1/user/1", nil)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	if err := c.DeleteUser(context.Background(), 1); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
}

func TestGetIssue(t *testing.T) {
	t.Parallel()

	want := seerr.Issue{ID: 1, IssueType: 1}

	srv := newTestServer(t, http.MethodGet, "/api/v1/issue/1", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.GetIssue(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if got.IssueType != 1 {
		t.Errorf("IssueType = %d", got.IssueType)
	}
}

func TestDeleteIssue(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodDelete, "/api/v1/issue/1", nil)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	if err := c.DeleteIssue(context.Background(), 1); err != nil {
		t.Fatalf("DeleteIssue: %v", err)
	}
}

func TestResolveIssue(t *testing.T) {
	t.Parallel()

	want := seerr.Issue{ID: 1, IssueType: 1}

	srv := newTestServer(t, http.MethodPost, "/api/v1/issue/1/resolved", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.ResolveIssue(context.Background(), 1)
	if err != nil {
		t.Fatalf("ResolveIssue: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d", got.ID)
	}
}

func TestReopenIssue(t *testing.T) {
	t.Parallel()

	want := seerr.Issue{ID: 1, IssueType: 1}

	srv := newTestServer(t, http.MethodPost, "/api/v1/issue/1/open", want)
	defer srv.Close()

	c, _ := seerr.New(srv.URL, "test-key")
	got, err := c.ReopenIssue(context.Background(), 1)
	if err != nil {
		t.Fatalf("ReopenIssue: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d", got.ID)
	}
}

func TestGetSonarrSettings(t *testing.T) {
	t.Parallel()

	want := []seerr.ServiceSettings{
		{
			ID:                1,
			Name:              "Sonarr 4K",
			Hostname:          "sonarr.lan",
			Port:              8989,
			APIKey:            "sk",
			UseSSL:            false,
			BaseURL:           "/sonarr/",
			Is4k:              true,
			IsDefault:         true,
			ActiveProfileID:   1,
			ActiveProfileName: "Ultra-HD",
			ActiveDirectory:   "/media/tv4k",
		},
	}

	t.Run("positive decode", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v1/settings/sonarr", want)
		defer srv.Close()

		c, err := seerr.New(srv.URL, "test-key")
		if err != nil {
			t.Fatalf("seerr.New: %v", err)
		}
		got, err := c.GetSonarrSettings(context.Background())
		if err != nil {
			t.Fatalf("GetSonarrSettings: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
		if got[0].Name != "Sonarr 4K" {
			t.Errorf("Name = %q, want %q", got[0].Name, "Sonarr 4K")
		}
		if got[0].Port != 8989 {
			t.Errorf("Port = %d, want 8989", got[0].Port)
		}
		if got[0].ActiveDirectory != "/media/tv4k" {
			t.Errorf("ActiveDirectory = %q, want %q", got[0].ActiveDirectory, "/media/tv4k")
		}
	})

	t.Run("negative HTTP 500", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"Sonarr unavailable"}`))
		}))
		defer srv.Close()

		c, err := seerr.New(srv.URL, "test-key")
		if err != nil {
			t.Fatalf("seerr.New: %v", err)
		}
		got, err := c.GetSonarrSettings(context.Background())
		if err == nil {
			t.Fatal("expected error on HTTP 500, got nil")
		}
		if got != nil {
			t.Errorf("got = %v, want nil", got)
		}
		var apiErr *arr.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *arr.APIError, got %T", err)
		}
		if apiErr.StatusCode != http.StatusInternalServerError {
			t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
		}
	})

	t.Run("boundary empty array", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v1/settings/sonarr", []seerr.ServiceSettings{})
		defer srv.Close()

		c, err := seerr.New(srv.URL, "test-key")
		if err != nil {
			t.Fatalf("seerr.New: %v", err)
		}
		got, err := c.GetSonarrSettings(context.Background())
		if err != nil {
			t.Fatalf("GetSonarrSettings: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len = %d, want 0", len(got))
		}
		if got == nil {
			t.Error("expected non-nil empty slice")
		}
	})
}

func TestGetRadarrSettings(t *testing.T) {
	t.Parallel()

	want := []seerr.ServiceSettings{
		{
			ID:                0,
			Name:              "Radarr",
			Hostname:          "radarr.lan",
			Port:              7878,
			APIKey:            "rk",
			UseSSL:            true,
			BaseURL:           "",
			Is4k:              false,
			IsDefault:         true,
			ActiveProfileID:   2,
			ActiveProfileName: "HD-1080p",
			ActiveDirectory:   "/media/movies",
		},
	}

	t.Run("positive decode", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v1/settings/radarr", want)
		defer srv.Close()

		c, err := seerr.New(srv.URL, "test-key")
		if err != nil {
			t.Fatalf("seerr.New: %v", err)
		}
		got, err := c.GetRadarrSettings(context.Background())
		if err != nil {
			t.Fatalf("GetRadarrSettings: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
		if got[0].Name != "Radarr" {
			t.Errorf("Name = %q, want %q", got[0].Name, "Radarr")
		}
		if got[0].Port != 7878 {
			t.Errorf("Port = %d, want 7878", got[0].Port)
		}
		if !got[0].UseSSL {
			t.Errorf("UseSSL = %v, want true", got[0].UseSSL)
		}
	})

	t.Run("negative HTTP 500", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"Radarr unavailable"}`))
		}))
		defer srv.Close()

		c, err := seerr.New(srv.URL, "test-key")
		if err != nil {
			t.Fatalf("seerr.New: %v", err)
		}
		got, err := c.GetRadarrSettings(context.Background())
		if err == nil {
			t.Fatal("expected error on HTTP 500, got nil")
		}
		if got != nil {
			t.Errorf("got = %v, want nil", got)
		}
		var apiErr *arr.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *arr.APIError, got %T", err)
		}
		if apiErr.StatusCode != http.StatusInternalServerError {
			t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
		}
	})

	t.Run("boundary empty array", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v1/settings/radarr", []seerr.ServiceSettings{})
		defer srv.Close()

		c, err := seerr.New(srv.URL, "test-key")
		if err != nil {
			t.Fatalf("seerr.New: %v", err)
		}
		got, err := c.GetRadarrSettings(context.Background())
		if err != nil {
			t.Fatalf("GetRadarrSettings: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len = %d, want 0", len(got))
		}
		if got == nil {
			t.Error("expected non-nil empty slice")
		}
	})
}

func TestServiceSettings_URL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		settings seerr.ServiceSettings
		want     string
	}{
		{
			name:     "http with port and empty base",
			settings: seerr.ServiceSettings{Hostname: "sonarr.lan", Port: 8989, UseSSL: false, BaseURL: ""},
			want:     "http://sonarr.lan:8989",
		},
		{
			name:     "https with port and base path with slashes",
			settings: seerr.ServiceSettings{Hostname: "radarr.lan", Port: 7878, UseSSL: true, BaseURL: "/radarr/"},
			want:     "https://radarr.lan:7878/radarr",
		},
		{
			name:     "no port (port 0) omitted",
			settings: seerr.ServiceSettings{Hostname: "10.0.0.5", Port: 0, UseSSL: false, BaseURL: "arr"},
			want:     "http://10.0.0.5/arr",
		},
		{
			name:     "negative port omitted",
			settings: seerr.ServiceSettings{Hostname: "10.0.0.5", Port: -1, UseSSL: false, BaseURL: ""},
			want:     "http://10.0.0.5",
		},
		{
			name:     "IPv6 host with port",
			settings: seerr.ServiceSettings{Hostname: "::1", Port: 8989, UseSSL: false, BaseURL: ""},
			want:     "http://[::1]:8989",
		},
		{
			name:     "IPv6 host with port and ssl",
			settings: seerr.ServiceSettings{Hostname: "2001:db8::1", Port: 8989, UseSSL: true, BaseURL: "/sonarr"},
			want:     "https://[2001:db8::1]:8989/sonarr",
		},
		{
			name:     "IPv6 host without port",
			settings: seerr.ServiceSettings{Hostname: "::1", Port: 0, UseSSL: false, BaseURL: ""},
			want:     "http://::1",
		},
		{
			name:     "base with redundant slashes",
			settings: seerr.ServiceSettings{Hostname: "media.lan", Port: 8080, UseSSL: false, BaseURL: "///sonarr///"},
			want:     "http://media.lan:8080/sonarr",
		},
		{
			name:     "single slash base becomes empty path",
			settings: seerr.ServiceSettings{Hostname: "media.lan", Port: 8080, UseSSL: false, BaseURL: "/"},
			want:     "http://media.lan:8080",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.settings.URL()
			if got != tc.want {
				t.Errorf("URL() = %q, want %q", got, tc.want)
			}
			if len(got) == 0 {
				t.Error("expected non-empty URL string")
			}
		})
	}
}

func TestUser_Name(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		user *seerr.User
		want string
	}{
		{
			name: "nil user returns empty",
			user: nil,
			want: "",
		},
		{
			name: "all empty returns empty",
			user: &seerr.User{},
			want: "",
		},
		{
			name: "all blank whitespace returns empty",
			user: &seerr.User{DisplayName: "   ", Username: "\t", Email: " \n "},
			want: "",
		},
		{
			name: "display name preferred over username and email",
			user: &seerr.User{DisplayName: "Alice", Username: "alice_w", Email: "alice@example.com"},
			want: "Alice",
		},
		{
			name: "username fallback when display name empty",
			user: &seerr.User{DisplayName: "", Username: "bob", Email: "bob@example.com"},
			want: "bob",
		},
		{
			name: "username fallback when display name whitespace",
			user: &seerr.User{DisplayName: "   ", Username: "bob", Email: "bob@example.com"},
			want: "bob",
		},
		{
			name: "email fallback when display name and username empty",
			user: &seerr.User{DisplayName: "", Username: "", Email: "carol@example.com"},
			want: "carol@example.com",
		},
		{
			name: "email fallback when display name and username whitespace",
			user: &seerr.User{DisplayName: "  ", Username: "\t", Email: "carol@example.com"},
			want: "carol@example.com",
		},
		{
			name: "preserves display name with surrounding spaces",
			user: &seerr.User{DisplayName: " Dave ", Username: "dave", Email: "dave@example.com"},
			want: " Dave ",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.user.Name()
			if got != tc.want {
				t.Errorf("Name() = %q, want %q", got, tc.want)
			}
			if tc.want == "" && len(got) != 0 {
				t.Errorf("len(got) = %d, want 0", len(got))
			}
			if tc.want != "" && len(got) == 0 {
				t.Error("expected non-empty string")
			}
		})
	}
}

func TestMediaRequest_DecodeProcessingFields(t *testing.T) {
	t.Parallel()

	fixture := json.RawMessage(`{
		"pageInfo": {"page": 1, "pages": 1, "results": 3},
		"results": [
			{
				"id": 11,
				"type": "movie",
				"status": 2,
				"is4k": false,
				"serverId": 0,
				"createdAt": "2026-09-20T10:00:00.000Z",
				"requestedBy": {
					"displayName": "Alice",
					"username": "alice",
					"email": "alice@example.com"
				},
				"media": {
					"id": 5,
					"mediaType": "movie",
					"tmdbId": 603,
					"status": 3,
					"status4k": 1
				}
			},
			{
				"id": 12,
				"type": "tv",
				"status": 2,
				"is4k": true,
				"serverId": 1,
				"requestedBy": {
					"username": "bob"
				},
				"seasons": [
					{"id": 1, "seasonNumber": 1, "status": 2},
					{"id": 2, "seasonNumber": 2, "status": 2}
				],
				"media": {
					"id": 6,
					"mediaType": "tv",
					"tmdbId": 1396,
					"status": 4,
					"status4k": 2
				}
			},
			{
				"id": 13,
				"type": "tv",
				"status": 2,
				"is4k": false,
				"serverId": 1,
				"requestedBy": {
					"username": "carol"
				},
				"seasons": [],
				"media": {
					"id": 7,
					"mediaType": "tv",
					"tmdbId": 1397,
					"status": 4,
					"status4k": 0
				}
			}
		]
	}`)

	srv := newTestServer(t, http.MethodGet, "/api/v1/request?take=20&skip=0&filter=processing", fixture)
	defer srv.Close()

	c, err := seerr.New(srv.URL, "test-key")
	if err != nil {
		t.Fatalf("seerr.New: %v", err)
	}

	results, pageInfo, err := c.GetRequests(context.Background(), 20, 0, "processing")
	if err != nil {
		t.Fatalf("GetRequests: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("len = %d, want 3", len(results))
	}
	if pageInfo.Results != 3 {
		t.Errorf("PageInfo.Results = %d, want 3", pageInfo.Results)
	}

	// Movie request assertions
	movieReq := results[0]
	if movieReq.Type != "movie" {
		t.Errorf("movieReq.Type = %q, want %q", movieReq.Type, "movie")
	}
	if movieReq.Media == nil || movieReq.Media.MediaType != "movie" {
		t.Errorf("movieReq.Media.MediaType = %v, want %q", movieReq.Media, "movie")
	}
	if movieReq.Media.Status4k != 1 {
		t.Errorf("movieReq.Media.Status4k = %d, want 1", movieReq.Media.Status4k)
	}
	if movieReq.RequestedBy == nil || movieReq.RequestedBy.DisplayName != "Alice" {
		t.Errorf("movieReq.RequestedBy.DisplayName = %v, want %q", movieReq.RequestedBy, "Alice")
	}
	if movieReq.RequestedBy.Name() != "Alice" {
		t.Errorf("movieReq.RequestedBy.Name() = %q, want %q", movieReq.RequestedBy.Name(), "Alice")
	}
	if movieReq.Seasons != nil {
		t.Errorf("movieReq.Seasons = %v, want nil", movieReq.Seasons)
	}

	// TV request assertions
	tvReq := results[1]
	if tvReq.Type != "tv" {
		t.Errorf("tvReq.Type = %q, want %q", tvReq.Type, "tv")
	}
	if tvReq.Media == nil || tvReq.Media.MediaType != "tv" {
		t.Errorf("tvReq.Media.MediaType = %v, want %q", tvReq.Media, "tv")
	}
	if tvReq.Media.Status4k != 2 {
		t.Errorf("tvReq.Media.Status4k = %d, want 2", tvReq.Media.Status4k)
	}
	if tvReq.Seasons == nil {
		t.Fatal("tvReq.Seasons is nil, want non-nil pointer")
	}
	tvSeasons := *tvReq.Seasons
	if len(tvSeasons) != 2 {
		t.Fatalf("tvSeasons len = %d, want 2", len(tvSeasons))
	}
	if tvSeasons[0].SeasonNumber != 1 || tvSeasons[0].Status != 2 || tvSeasons[0].ID != 1 {
		t.Errorf("tvSeasons[0] = %+v, want ID=1 SeasonNumber=1 Status=2", tvSeasons[0])
	}
	if tvSeasons[1].SeasonNumber != 2 || tvSeasons[1].Status != 2 || tvSeasons[1].ID != 2 {
		t.Errorf("tvSeasons[1] = %+v, want ID=2 SeasonNumber=2 Status=2", tvSeasons[1])
	}
	if tvReq.RequestedBy == nil || tvReq.RequestedBy.Name() != "bob" {
		t.Errorf("tvReq.RequestedBy.Name() = %v, want %q", tvReq.RequestedBy, "bob")
	}

	// TV request boundary case: empty seasons slice
	tvEmptyReq := results[2]
	if tvEmptyReq.Type != "tv" {
		t.Errorf("tvEmptyReq.Type = %q, want %q", tvEmptyReq.Type, "tv")
	}
	if tvEmptyReq.Seasons == nil {
		t.Fatal("tvEmptyReq.Seasons is nil, want non-nil pointer to empty slice")
	}
	if len(*tvEmptyReq.Seasons) != 0 {
		t.Errorf("len(*tvEmptyReq.Seasons) = %d, want 0", len(*tvEmptyReq.Seasons))
	}
}

func TestUser_JellyfinFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		json         string
		wantID       string
		wantUsername string
	}{
		{
			name:         "populated jellyfin fields",
			json:         `{"id": 42, "email": "jf@example.com", "jellyfinUserId": "jf-uuid-123", "jellyfinUsername": "jfuser"}`,
			wantID:       "jf-uuid-123",
			wantUsername: "jfuser",
		},
		{
			name:         "null jellyfin fields",
			json:         `{"id": 44, "email": "local@example.com", "jellyfinUserId": null, "jellyfinUsername": null}`,
			wantID:       "",
			wantUsername: "",
		},
		{
			name:         "omitted jellyfin fields",
			json:         `{"id": 43, "email": "plex@example.com"}`,
			wantID:       "",
			wantUsername: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var u seerr.User
			if err := json.Unmarshal([]byte(tc.json), &u); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if u.JellyfinUserID != tc.wantID {
				t.Errorf("JellyfinUserID = %q, want %q", u.JellyfinUserID, tc.wantID)
			}
			if u.JellyfinUsername != tc.wantUsername {
				t.Errorf("JellyfinUsername = %q, want %q", u.JellyfinUsername, tc.wantUsername)
			}
		})
	}
}
