package whisparr_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lusoris/goenvoy/arr/whisparr"
)

func newV2TestServer(t *testing.T, wantPath, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path+"?"+r.URL.RawQuery != wantPath && r.URL.Path != wantPath {
			t.Errorf("path = %s?%s, want %s", r.URL.Path, r.URL.RawQuery, wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
}

func newV3TestServer(t *testing.T, wantPath, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path+"?"+r.URL.RawQuery != wantPath && r.URL.Path != wantPath {
			t.Errorf("path = %s?%s, want %s", r.URL.Path, r.URL.RawQuery, wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
}

func TestNew(t *testing.T) {
	_, err := whisparr.New("http://localhost:6969", "abc123")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
}

func TestNewInvalidURL(t *testing.T) {
	_, err := whisparr.New("://bad", "key")
	if err == nil {
		t.Fatal("New() with bad URL should fail")
	}
}

func TestNewV3(t *testing.T) {
	_, err := whisparr.NewV3("http://localhost:6969", "abc123")
	if err != nil {
		t.Fatalf("NewV3() error = %v", err)
	}
}

func TestNewV3InvalidURL(t *testing.T) {
	_, err := whisparr.NewV3("://bad", "key")
	if err == nil {
		t.Fatal("NewV3() with bad URL should fail")
	}
}

func TestGetAllSeries(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/series", `[{"id":1,"title":"Test Site","monitored":true}]`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	series, err := c.GetAllSeries(context.Background())
	if err != nil {
		t.Fatalf("GetAllSeries() error = %v", err)
	}
	if len(series) != 1 || series[0].Title != "Test Site" {
		t.Errorf("got %+v", series)
	}
}

func TestGetSeries(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/series/1", `{"id":1,"title":"Site A"}`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	s, err := c.GetSeries(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetSeries() error = %v", err)
	}
	if s.Title != "Site A" {
		t.Errorf("title = %s", s.Title)
	}
}

func TestAddSeries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var s whisparr.Series
		json.NewDecoder(r.Body).Decode(&s)
		s.ID = 42
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	}))
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	out, err := c.AddSeries(context.Background(), &whisparr.Series{Title: "New Site"})
	if err != nil {
		t.Fatalf("AddSeries() error = %v", err)
	}
	if out.ID != 42 {
		t.Errorf("id = %d, want 42", out.ID)
	}
}

func TestDeleteSeries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	if err := c.DeleteSeries(context.Background(), 1, true, false); err != nil {
		t.Fatalf("DeleteSeries() error = %v", err)
	}
}

func TestLookupSeries(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/series/lookup?term=test", `[{"id":1,"title":"Found"}]`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	res, err := c.LookupSeries(context.Background(), "test")
	if err != nil {
		t.Fatalf("LookupSeries() error = %v", err)
	}
	if len(res) != 1 {
		t.Errorf("got %d results", len(res))
	}
}

func TestGetEpisodes(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/episode?seriesId=1", `[{"id":1,"title":"Ep 1","actors":[{"name":"Jane","gender":"female"}]}]`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	eps, err := c.GetEpisodes(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetEpisodes() error = %v", err)
	}
	if len(eps) != 1 || len(eps[0].Actors) != 1 {
		t.Errorf("got %+v", eps)
	}
	if eps[0].Actors[0].Gender != whisparr.GenderFemale {
		t.Errorf("gender = %s, want female", eps[0].Actors[0].Gender)
	}
}

func TestGetEpisodeFiles(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/episodefile?seriesId=1", `[{"id":1,"size":1024}]`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	files, err := c.GetEpisodeFiles(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetEpisodeFiles() error = %v", err)
	}
	if len(files) != 1 {
		t.Errorf("got %d files", len(files))
	}
}

func TestV2SendCommand(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/command", `{"id":1,"name":"RefreshSeries","status":"queued"}`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	_, err := c.SendCommand(context.Background(), struct {
		Name string `json:"name"`
	}{Name: "RefreshSeries"})
	if err != nil {
		t.Fatalf("SendCommand() error = %v", err)
	}
}

func TestV2GetSystemStatus(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/system/status", `{"appName":"Whisparr","version":"2.2.0"}`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	status, err := c.GetSystemStatus(context.Background())
	if err != nil {
		t.Fatalf("GetSystemStatus() error = %v", err)
	}
	if status.AppName != "Whisparr" {
		t.Errorf("appName = %s", status.AppName)
	}
}

func TestV2GetHealth(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/health", `[{"type":"warning","message":"test"}]`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	health, err := c.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("GetHealth() error = %v", err)
	}
	if len(health) != 1 {
		t.Errorf("got %d health checks", len(health))
	}
}

func TestV2GetTags(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/tag", `[{"id":1,"label":"hd"}]`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	tags, err := c.GetTags(context.Background())
	if err != nil {
		t.Fatalf("GetTags() error = %v", err)
	}
	if len(tags) != 1 || tags[0].Label != "hd" {
		t.Errorf("got %+v", tags)
	}
}

func TestV2GetQualityProfiles(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/qualityprofile", `[{"id":1,"name":"Any"}]`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	profiles, err := c.GetQualityProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetQualityProfiles() error = %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("got %d profiles", len(profiles))
	}
}

func TestV2GetRootFolders(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/rootfolder", `[{"id":1,"path":"/data"}]`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	folders, err := c.GetRootFolders(context.Background())
	if err != nil {
		t.Fatalf("GetRootFolders() error = %v", err)
	}
	if len(folders) != 1 {
		t.Errorf("got %d folders", len(folders))
	}
}

func TestV3GetAllMovies(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/movie", `[{"id":1,"title":"Scene 1","stashId":"abc","itemType":"scene"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	movies, err := c.GetAllMovies(context.Background())
	if err != nil {
		t.Fatalf("GetAllMovies() error = %v", err)
	}
	if len(movies) != 1 || movies[0].StashID != "abc" {
		t.Errorf("got %+v", movies)
	}
}

func TestV3GetMovie(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/movie/1", `{"id":1,"title":"Movie A","code":"ABC-123"}`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	m, err := c.GetMovie(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetMovie() error = %v", err)
	}
	if m.Code != "ABC-123" {
		t.Errorf("code = %s", m.Code)
	}
}

func TestV3AddMovie(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var m whisparr.Movie
		json.NewDecoder(r.Body).Decode(&m)
		m.ID = 42
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	out, err := c.AddMovie(context.Background(), &whisparr.Movie{Title: "New Scene"})
	if err != nil {
		t.Fatalf("AddMovie() error = %v", err)
	}
	if out.ID != 42 {
		t.Errorf("id = %d, want 42", out.ID)
	}
}

func TestV3DeleteMovie(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	if err := c.DeleteMovie(context.Background(), 1, true, false); err != nil {
		t.Fatalf("DeleteMovie() error = %v", err)
	}
}

func TestV3LookupScene(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/lookup/scene?term=test", `[{"id":1,"title":"Found"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	res, err := c.LookupScene(context.Background(), "test")
	if err != nil {
		t.Fatalf("LookupScene() error = %v", err)
	}
	if len(res) != 1 {
		t.Errorf("got %d results", len(res))
	}
}

func TestV3GetPerformers(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/performer", `[{"id":1,"name":"Jane Doe","gender":"female","stashId":"xyz"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	perfs, err := c.GetPerformers(context.Background())
	if err != nil {
		t.Fatalf("GetPerformers() error = %v", err)
	}
	if len(perfs) != 1 || perfs[0].Name != "Jane Doe" {
		t.Errorf("got %+v", perfs)
	}
	if perfs[0].Gender != whisparr.GenderFemale {
		t.Errorf("gender = %s, want female", perfs[0].Gender)
	}
}

func TestV3AddPerformer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var p whisparr.Performer
		json.NewDecoder(r.Body).Decode(&p)
		p.ID = 10
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	out, err := c.AddPerformer(context.Background(), &whisparr.Performer{Name: "New"})
	if err != nil {
		t.Fatalf("AddPerformer() error = %v", err)
	}
	if out.ID != 10 {
		t.Errorf("id = %d, want 10", out.ID)
	}
}

func TestV3DeletePerformer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	if err := c.DeletePerformer(context.Background(), 1, false); err != nil {
		t.Fatalf("DeletePerformer() error = %v", err)
	}
}

func TestV3GetStudios(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/studio", `[{"id":1,"title":"Studio X","stashId":"s1"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	studios, err := c.GetStudios(context.Background())
	if err != nil {
		t.Fatalf("GetStudios() error = %v", err)
	}
	if len(studios) != 1 || studios[0].Title != "Studio X" {
		t.Errorf("got %+v", studios)
	}
}

func TestV3AddStudio(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var s whisparr.Studio
		json.NewDecoder(r.Body).Decode(&s)
		s.ID = 5
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	out, err := c.AddStudio(context.Background(), &whisparr.Studio{Title: "New Studio"})
	if err != nil {
		t.Fatalf("AddStudio() error = %v", err)
	}
	if out.ID != 5 {
		t.Errorf("id = %d, want 5", out.ID)
	}
}

func TestV3DeleteStudio(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	if err := c.DeleteStudio(context.Background(), 1, false); err != nil {
		t.Fatalf("DeleteStudio() error = %v", err)
	}
}

func TestV3GetCredits(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/credit?movieId=1", `[{"id":1,"personName":"Jane","type":"cast"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	credits, err := c.GetCredits(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetCredits() error = %v", err)
	}
	if len(credits) != 1 {
		t.Errorf("got %d credits", len(credits))
	}
}

func TestV3GetMoviesByPerformer(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/movie/listbyperformerforeignid?performerForeignId=abc", `[{"id":1,"title":"Scene 1"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	movies, err := c.GetMoviesByPerformer(context.Background(), "abc")
	if err != nil {
		t.Fatalf("GetMoviesByPerformer() error = %v", err)
	}
	if len(movies) != 1 {
		t.Errorf("got %d movies", len(movies))
	}
}

func TestV3GetSystemStatus(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/system/status", `{"appName":"Whisparr","version":"3.3.3"}`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	status, err := c.GetSystemStatus(context.Background())
	if err != nil {
		t.Fatalf("GetSystemStatus() error = %v", err)
	}
	if status.Version != "3.3.3" {
		t.Errorf("version = %s", status.Version)
	}
}

func TestV3GetImportExclusions(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/exclusions", `[{"id":1,"movieTitle":"Excluded"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	exclusions, err := c.GetImportExclusions(context.Background())
	if err != nil {
		t.Fatalf("GetImportExclusions() error = %v", err)
	}
	if len(exclusions) != 1 {
		t.Errorf("got %d exclusions", len(exclusions))
	}
}

func TestV2ErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "bad-key")
	_, err := c.GetAllSeries(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestV3ErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "bad-key")
	_, err := c.GetAllMovies(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

// V2 untested methods.

func TestUpdateSeries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		var s whisparr.Series
		json.NewDecoder(r.Body).Decode(&s)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	}))
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	out, err := c.UpdateSeries(context.Background(), &whisparr.Series{ID: 1, Title: "Updated"}, true)
	if err != nil {
		t.Fatalf("UpdateSeries() error = %v", err)
	}
	if out.Title != "Updated" {
		t.Errorf("title = %s", out.Title)
	}
}

func TestGetEpisode(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/episode/5", `{"id":5,"title":"Scene 5"}`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	ep, err := c.GetEpisode(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetEpisode() error = %v", err)
	}
	if ep.ID != 5 {
		t.Errorf("id = %d, want 5", ep.ID)
	}
}

func TestDeleteEpisodeFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	if err := c.DeleteEpisodeFile(context.Background(), 1); err != nil {
		t.Fatalf("DeleteEpisodeFile() error = %v", err)
	}
}

func TestV2GetCalendar(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/calendar?start=2026-01-01&end=2026-01-31&unmonitored=false", `[{"id":1,"title":"Upcoming"}]`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	eps, err := c.GetCalendar(context.Background(), "2026-01-01", "2026-01-31", false)
	if err != nil {
		t.Fatalf("GetCalendar() error = %v", err)
	}
	if len(eps) != 1 {
		t.Errorf("got %d episodes", len(eps))
	}
}

func TestV2Parse(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/parse?title=test+scene", `{"title":"test scene"}`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	result, err := c.Parse(context.Background(), "test scene")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestV2GetDiskSpace(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/diskspace", `[{"path":"/data","freeSpace":1000}]`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	ds, err := c.GetDiskSpace(context.Background())
	if err != nil {
		t.Fatalf("GetDiskSpace() error = %v", err)
	}
	if len(ds) != 1 {
		t.Errorf("got %d disk spaces", len(ds))
	}
}

func TestV2GetQueue(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/queue?page=1&pageSize=10", `{"page":1,"pageSize":10,"totalRecords":0,"records":[]}`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	q, err := c.GetQueue(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetQueue() error = %v", err)
	}
	if q.Page != 1 {
		t.Errorf("page = %d", q.Page)
	}
}

func TestV2CreateTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":1,"label":"new-tag"}`))
	}))
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	tag, err := c.CreateTag(context.Background(), "new-tag")
	if err != nil {
		t.Fatalf("CreateTag() error = %v", err)
	}
	if tag.Label != "new-tag" {
		t.Errorf("label = %s", tag.Label)
	}
}

func TestV2GetHistory(t *testing.T) {
	ts := newV2TestServer(t, "/api/v3/history?page=1&pageSize=10", `{"page":1,"pageSize":10,"totalRecords":0,"records":[]}`)
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	h, err := c.GetHistory(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if h.Page != 1 {
		t.Errorf("page = %d", h.Page)
	}
}

func TestV2UpdateSeasonPass(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()
	c, _ := whisparr.New(ts.URL, "test-key")
	err := c.UpdateSeasonPass(context.Background(), whisparr.SeasonPassResource{})
	if err != nil {
		t.Fatalf("UpdateSeasonPass() error = %v", err)
	}
}

// V3 untested methods.

func TestV3UpdateMovie(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		var m whisparr.Movie
		json.NewDecoder(r.Body).Decode(&m)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	out, err := c.UpdateMovie(context.Background(), &whisparr.Movie{ID: 1, Title: "Updated"}, true)
	if err != nil {
		t.Fatalf("UpdateMovie() error = %v", err)
	}
	if out.Title != "Updated" {
		t.Errorf("title = %s", out.Title)
	}
}

func TestV3LookupMovie(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/lookup/movie?term=test", `[{"id":1,"title":"Found"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	res, err := c.LookupMovie(context.Background(), "test")
	if err != nil {
		t.Fatalf("LookupMovie() error = %v", err)
	}
	if len(res) != 1 {
		t.Errorf("got %d results", len(res))
	}
}

func TestV3GetMoviesByStudio(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/movie/listbystudioforeignid?studioForeignId=s1", `[{"id":1,"title":"Scene 1"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	movies, err := c.GetMoviesByStudio(context.Background(), "s1")
	if err != nil {
		t.Fatalf("GetMoviesByStudio() error = %v", err)
	}
	if len(movies) != 1 {
		t.Errorf("got %d movies", len(movies))
	}
}

func TestV3GetMovieFile(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/moviefile/1", `{"id":1,"size":2048}`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	f, err := c.GetMovieFile(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetMovieFile() error = %v", err)
	}
	if f.ID != 1 {
		t.Errorf("id = %d", f.ID)
	}
}

func TestV3DeleteMovieFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	if err := c.DeleteMovieFile(context.Background(), 1); err != nil {
		t.Fatalf("DeleteMovieFile() error = %v", err)
	}
}

func TestV3EditMovies(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	err := c.EditMovies(context.Background(), &whisparr.MovieEditorResource{MovieIDs: []int{1, 2}})
	if err != nil {
		t.Fatalf("EditMovies() error = %v", err)
	}
}

func TestV3DeleteMovies(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	err := c.DeleteMovies(context.Background(), &whisparr.MovieEditorResource{MovieIDs: []int{1}})
	if err != nil {
		t.Fatalf("DeleteMovies() error = %v", err)
	}
}

func TestV3GetPerformer(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/performer/1", `{"id":1,"name":"Jane"}`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	p, err := c.GetPerformer(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetPerformer() error = %v", err)
	}
	if p.Name != "Jane" {
		t.Errorf("name = %s", p.Name)
	}
}

func TestV3UpdatePerformer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		var p whisparr.Performer
		json.NewDecoder(r.Body).Decode(&p)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	out, err := c.UpdatePerformer(context.Background(), &whisparr.Performer{ID: 1, Name: "Updated"})
	if err != nil {
		t.Fatalf("UpdatePerformer() error = %v", err)
	}
	if out.Name != "Updated" {
		t.Errorf("name = %s", out.Name)
	}
}

func TestV3GetStudio(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/studio/1", `{"id":1,"title":"Studio A"}`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	s, err := c.GetStudio(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetStudio() error = %v", err)
	}
	if s.Title != "Studio A" {
		t.Errorf("title = %s", s.Title)
	}
}

func TestV3UpdateStudio(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		var s whisparr.Studio
		json.NewDecoder(r.Body).Decode(&s)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	out, err := c.UpdateStudio(context.Background(), &whisparr.Studio{ID: 1, Title: "Updated"})
	if err != nil {
		t.Fatalf("UpdateStudio() error = %v", err)
	}
	if out.Title != "Updated" {
		t.Errorf("title = %s", out.Title)
	}
}

func TestV3GetCalendar(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/calendar?start=2026-01-01&end=2026-01-31&unmonitored=false", `[{"id":1,"title":"Upcoming"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	movies, err := c.GetCalendar(context.Background(), "2026-01-01", "2026-01-31", false)
	if err != nil {
		t.Fatalf("GetCalendar() error = %v", err)
	}
	if len(movies) != 1 {
		t.Errorf("got %d movies", len(movies))
	}
}

func TestV3SendCommand(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/command", `{"id":1,"name":"RefreshMovie","status":"queued"}`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	_, err := c.SendCommand(context.Background(), struct {
		Name string `json:"name"`
	}{Name: "RefreshMovie"})
	if err != nil {
		t.Fatalf("SendCommand() error = %v", err)
	}
}

func TestV3Parse(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/parse?title=test+movie", `{"title":"test movie"}`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	result, err := c.Parse(context.Background(), "test movie")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestV3GetHealth(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/health", `[{"type":"warning","message":"test"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	health, err := c.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("GetHealth() error = %v", err)
	}
	if len(health) != 1 {
		t.Errorf("got %d health checks", len(health))
	}
}

func TestV3GetDiskSpace(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/diskspace", `[{"path":"/data","freeSpace":1000}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	ds, err := c.GetDiskSpace(context.Background())
	if err != nil {
		t.Fatalf("GetDiskSpace() error = %v", err)
	}
	if len(ds) != 1 {
		t.Errorf("got %d disk spaces", len(ds))
	}
}

func TestV3GetQueue(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/queue?page=1&pageSize=10", `{"page":1,"pageSize":10,"totalRecords":0,"records":[]}`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	q, err := c.GetQueue(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetQueue() error = %v", err)
	}
	if q.Page != 1 {
		t.Errorf("page = %d", q.Page)
	}
}

func TestV3GetQualityProfiles(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/qualityprofile", `[{"id":1,"name":"Any"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	profiles, err := c.GetQualityProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetQualityProfiles() error = %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("got %d profiles", len(profiles))
	}
}

func TestV3GetTags(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/tag", `[{"id":1,"label":"hd"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	tags, err := c.GetTags(context.Background())
	if err != nil {
		t.Fatalf("GetTags() error = %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("got %d tags", len(tags))
	}
}

func TestV3CreateTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":1,"label":"new-tag"}`))
	}))
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	tag, err := c.CreateTag(context.Background(), "new-tag")
	if err != nil {
		t.Fatalf("CreateTag() error = %v", err)
	}
	if tag.Label != "new-tag" {
		t.Errorf("label = %s", tag.Label)
	}
}

func TestV3GetRootFolders(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/rootfolder", `[{"id":1,"path":"/data"}]`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	folders, err := c.GetRootFolders(context.Background())
	if err != nil {
		t.Fatalf("GetRootFolders() error = %v", err)
	}
	if len(folders) != 1 {
		t.Errorf("got %d folders", len(folders))
	}
}

func TestV3GetHistory(t *testing.T) {
	ts := newV3TestServer(t, "/api/v3/history?page=1&pageSize=10", `{"page":1,"pageSize":10,"totalRecords":0,"records":[]}`)
	defer ts.Close()
	c, _ := whisparr.NewV3(ts.URL, "test-key")
	h, err := c.GetHistory(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if h.Page != 1 {
		t.Errorf("page = %d", h.Page)
	}
}
