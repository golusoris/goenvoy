package musicbrainz_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/metadata"
	"github.com/golusoris/goenvoy/metadata/music/musicbrainz"
)

func newTestServer(t *testing.T, wantPath string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Error("User-Agent header is empty")
		}
		if got := r.URL.Query().Get("fmt"); got != "json" {
			t.Errorf("fmt = %q, want %q", got, "json")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
}

func TestLookupArtist(t *testing.T) {
	t.Parallel()

	artist := musicbrainz.Artist{
		ID:       "5b11f4ce-a62d-471e-81fc-a69a8278c7da",
		Name:     "Nirvana",
		SortName: "Nirvana",
		Type:     "Group",
		Country:  "US",
		LifeSpan: &musicbrainz.LifeSpan{Begin: "1987", End: "1994-04-05", Ended: true},
	}
	ts := newTestServer(t, "/artist/5b11f4ce-a62d-471e-81fc-a69a8278c7da", artist)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	a, err := c.LookupArtist(context.Background(), "5b11f4ce-a62d-471e-81fc-a69a8278c7da", nil)
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "Nirvana" {
		t.Errorf("Name = %q, want %q", a.Name, "Nirvana")
	}
	if a.Country != "US" {
		t.Errorf("Country = %q, want %q", a.Country, "US")
	}
	if a.LifeSpan == nil || !a.LifeSpan.Ended {
		t.Error("LifeSpan.Ended should be true")
	}
}

func TestLookupArtistWithIncludes(t *testing.T) {
	t.Parallel()

	artist := musicbrainz.Artist{
		ID:   "a1",
		Name: "Radiohead",
		Recordings: []musicbrainz.Recording{
			{ID: "r1", Title: "Creep"},
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/artist/a1" {
			t.Errorf("path = %q, want /artist/a1", r.URL.Path)
		}
		if got := r.URL.Query().Get("inc"); got != "recordings+releases" {
			t.Errorf("inc = %q, want recordings+releases", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(artist)
	}))
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	a, err := c.LookupArtist(context.Background(), "a1", []string{"recordings", "releases"})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Recordings) != 1 {
		t.Fatalf("len(recordings) = %d, want 1", len(a.Recordings))
	}
}

func TestLookupRelease(t *testing.T) {
	t.Parallel()

	release := musicbrainz.Release{
		ID:     "b84ee12a-09ef-421b-82de-0441a926375b",
		Title:  "Nevermind",
		Status: "Official",
		Date:   "1991-09-24",
	}
	ts := newTestServer(t, "/release/b84ee12a-09ef-421b-82de-0441a926375b", release)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	r, err := c.LookupRelease(context.Background(), "b84ee12a-09ef-421b-82de-0441a926375b", nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Title != "Nevermind" {
		t.Errorf("Title = %q, want %q", r.Title, "Nevermind")
	}
	if r.Status != "Official" {
		t.Errorf("Status = %q, want %q", r.Status, "Official")
	}
}

func TestLookupReleaseGroup(t *testing.T) {
	t.Parallel()

	rg := musicbrainz.ReleaseGroup{
		ID:               "1b022e01-4da6-387b-8658-8678046e4cef",
		Title:            "Nevermind",
		PrimaryType:      "Album",
		FirstReleaseDate: "1991-09-24",
	}
	ts := newTestServer(t, "/release-group/1b022e01-4da6-387b-8658-8678046e4cef", rg)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	r, err := c.LookupReleaseGroup(context.Background(), "1b022e01-4da6-387b-8658-8678046e4cef", nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Title != "Nevermind" {
		t.Errorf("Title = %q, want %q", r.Title, "Nevermind")
	}
	if r.PrimaryType != "Album" {
		t.Errorf("PrimaryType = %q, want %q", r.PrimaryType, "Album")
	}
}

func TestLookupRecording(t *testing.T) {
	t.Parallel()

	rec := musicbrainz.Recording{
		ID:     "87ec0c32-6035-476e-a7a6-8543b4bfbb65",
		Title:  "Smells Like Teen Spirit",
		Length: 301000,
	}
	ts := newTestServer(t, "/recording/87ec0c32-6035-476e-a7a6-8543b4bfbb65", rec)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	r, err := c.LookupRecording(context.Background(), "87ec0c32-6035-476e-a7a6-8543b4bfbb65", nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Title != "Smells Like Teen Spirit" {
		t.Errorf("Title = %q, want %q", r.Title, "Smells Like Teen Spirit")
	}
	if r.Length != 301000 {
		t.Errorf("Length = %d, want %d", r.Length, 301000)
	}
}

func TestLookupLabel(t *testing.T) {
	t.Parallel()

	label := musicbrainz.Label{
		ID:   "50c384a2-0b44-401b-b893-8181571d90e7",
		Name: "DGC Records",
		Type: "Original Production",
	}
	ts := newTestServer(t, "/label/50c384a2-0b44-401b-b893-8181571d90e7", label)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	l, err := c.LookupLabel(context.Background(), "50c384a2-0b44-401b-b893-8181571d90e7", nil)
	if err != nil {
		t.Fatal(err)
	}
	if l.Name != "DGC Records" {
		t.Errorf("Name = %q, want %q", l.Name, "DGC Records")
	}
}

func TestLookupWork(t *testing.T) {
	t.Parallel()

	work := musicbrainz.Work{ID: "abc-123", Title: "Bohemian Rhapsody", Type: "Song"}
	ts := newTestServer(t, "/work/abc-123", work)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	w, err := c.LookupWork(context.Background(), "abc-123", nil)
	if err != nil {
		t.Fatal(err)
	}
	if w.Title != "Bohemian Rhapsody" {
		t.Errorf("Title = %q, want %q", w.Title, "Bohemian Rhapsody")
	}
}

func TestLookupArea(t *testing.T) {
	t.Parallel()

	area := musicbrainz.Area{ID: "area-1", Name: "United Kingdom", ISO31661Codes: []string{"GB"}}
	ts := newTestServer(t, "/area/area-1", area)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	a, err := c.LookupArea(context.Background(), "area-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "United Kingdom" {
		t.Errorf("Name = %q, want %q", a.Name, "United Kingdom")
	}
}

func TestLookupEvent(t *testing.T) {
	t.Parallel()

	event := musicbrainz.Event{ID: "evt-1", Name: "Live Aid"}
	ts := newTestServer(t, "/event/evt-1", event)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	e, err := c.LookupEvent(context.Background(), "evt-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if e.Name != "Live Aid" {
		t.Errorf("Name = %q, want %q", e.Name, "Live Aid")
	}
}

func TestLookupGenre(t *testing.T) {
	t.Parallel()

	genre := musicbrainz.Genre{ID: "genre-1", Name: "rock"}
	ts := newTestServer(t, "/genre/genre-1", genre)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	g, err := c.LookupGenre(context.Background(), "genre-1")
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "rock" {
		t.Errorf("Name = %q, want %q", g.Name, "rock")
	}
}

func TestLookupInstrument(t *testing.T) {
	t.Parallel()

	inst := musicbrainz.Instrument{ID: "inst-1", Name: "guitar", Type: "String instrument"}
	ts := newTestServer(t, "/instrument/inst-1", inst)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	i, err := c.LookupInstrument(context.Background(), "inst-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if i.Name != "guitar" {
		t.Errorf("Name = %q, want %q", i.Name, "guitar")
	}
}

func TestLookupPlace(t *testing.T) {
	t.Parallel()

	place := musicbrainz.Place{
		ID:          "place-1",
		Name:        "Abbey Road Studios",
		Address:     "3 Abbey Road, London",
		Coordinates: &musicbrainz.Coordinates{Latitude: 51.5320, Longitude: -0.1767},
	}
	ts := newTestServer(t, "/place/place-1", place)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	p, err := c.LookupPlace(context.Background(), "place-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Abbey Road Studios" {
		t.Errorf("Name = %q, want %q", p.Name, "Abbey Road Studios")
	}
	if p.Coordinates == nil {
		t.Fatal("Coordinates should not be nil")
	}
}

func TestSearchArtists(t *testing.T) {
	t.Parallel()

	resp := struct {
		Created string               `json:"created"`
		Count   int                  `json:"count"`
		Offset  int                  `json:"offset"`
		Artists []musicbrainz.Artist `json:"artists"`
	}{
		Created: "2024-01-01T00:00:00Z",
		Count:   100,
		Offset:  0,
		Artists: []musicbrainz.Artist{{ID: "a1", Name: "Radiohead", SortName: "Radiohead"}},
	}
	ts := newTestServer(t, "/artist", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.SearchArtists(context.Background(), "radiohead", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 100 {
		t.Errorf("Count = %d, want 100", result.Count)
	}
	if len(result.Entities) != 1 {
		t.Fatalf("len = %d, want 1", len(result.Entities))
	}
	if result.Entities[0].Name != "Radiohead" {
		t.Errorf("Name = %q, want %q", result.Entities[0].Name, "Radiohead")
	}
}

func TestSearchReleases(t *testing.T) {
	t.Parallel()

	resp := struct {
		Created  string                `json:"created"`
		Count    int                   `json:"count"`
		Offset   int                   `json:"offset"`
		Releases []musicbrainz.Release `json:"releases"`
	}{
		Count:    50,
		Releases: []musicbrainz.Release{{ID: "r1", Title: "OK Computer"}},
	}
	ts := newTestServer(t, "/release", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.SearchReleases(context.Background(), "ok computer", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 50 {
		t.Errorf("Count = %d, want 50", result.Count)
	}
	if len(result.Entities) != 1 {
		t.Fatalf("len = %d, want 1", len(result.Entities))
	}
}

func TestSearchRecordings(t *testing.T) {
	t.Parallel()

	resp := struct {
		Created    string                  `json:"created"`
		Count      int                     `json:"count"`
		Offset     int                     `json:"offset"`
		Recordings []musicbrainz.Recording `json:"recordings"`
	}{
		Count:      200,
		Recordings: []musicbrainz.Recording{{ID: "rec1", Title: "Paranoid Android", Length: 386000}},
	}
	ts := newTestServer(t, "/recording", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.SearchRecordings(context.Background(), "paranoid android", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 200 {
		t.Errorf("Count = %d, want 200", result.Count)
	}
}

func TestSearchReleaseGroups(t *testing.T) {
	t.Parallel()

	resp := struct {
		Created       string                     `json:"created"`
		Count         int                        `json:"count"`
		Offset        int                        `json:"offset"`
		ReleaseGroups []musicbrainz.ReleaseGroup `json:"release-groups"`
	}{
		Count:         10,
		ReleaseGroups: []musicbrainz.ReleaseGroup{{ID: "rg1", Title: "The Bends", PrimaryType: "Album"}},
	}
	ts := newTestServer(t, "/release-group", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.SearchReleaseGroups(context.Background(), "the bends", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entities) != 1 {
		t.Fatalf("len = %d, want 1", len(result.Entities))
	}
}

func TestSearchLabels(t *testing.T) {
	t.Parallel()

	resp := struct {
		Created string              `json:"created"`
		Count   int                 `json:"count"`
		Offset  int                 `json:"offset"`
		Labels  []musicbrainz.Label `json:"labels"`
	}{
		Count:  5,
		Labels: []musicbrainz.Label{{ID: "l1", Name: "Parlophone"}},
	}
	ts := newTestServer(t, "/label", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.SearchLabels(context.Background(), "parlophone", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 5 {
		t.Errorf("Count = %d, want 5", result.Count)
	}
}

func TestSearchWorks(t *testing.T) {
	t.Parallel()

	resp := struct {
		Created string             `json:"created"`
		Count   int                `json:"count"`
		Offset  int                `json:"offset"`
		Works   []musicbrainz.Work `json:"works"`
	}{
		Count: 7,
		Works: []musicbrainz.Work{{ID: "w1", Title: "Exit Music (For a Film)"}},
	}
	ts := newTestServer(t, "/work", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.SearchWorks(context.Background(), "exit music", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Entities[0].Title != "Exit Music (For a Film)" {
		t.Errorf("Title = %q, want Exit Music (For a Film)", result.Entities[0].Title)
	}
}

func TestSearchAreas(t *testing.T) {
	t.Parallel()

	resp := struct {
		Created string             `json:"created"`
		Count   int                `json:"count"`
		Offset  int                `json:"offset"`
		Areas   []musicbrainz.Area `json:"areas"`
	}{
		Count: 1,
		Areas: []musicbrainz.Area{{ID: "area-1", Name: "Iceland"}},
	}
	ts := newTestServer(t, "/area", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.SearchAreas(context.Background(), "iceland", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 1 || result.Entities[0].Name != "Iceland" {
		t.Fatalf("unexpected areas: %+v", result)
	}
}

func TestSearchEvents(t *testing.T) {
	t.Parallel()

	resp := struct {
		Created string              `json:"created"`
		Count   int                 `json:"count"`
		Offset  int                 `json:"offset"`
		Events  []musicbrainz.Event `json:"events"`
	}{
		Count:  2,
		Events: []musicbrainz.Event{{ID: "event-1", Name: "Glastonbury"}},
	}
	ts := newTestServer(t, "/event", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.SearchEvents(context.Background(), "glastonbury", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Entities[0].Name != "Glastonbury" {
		t.Errorf("Name = %q, want Glastonbury", result.Entities[0].Name)
	}
}

func TestSearchInstruments(t *testing.T) {
	t.Parallel()

	resp := struct {
		Created     string                   `json:"created"`
		Count       int                      `json:"count"`
		Offset      int                      `json:"offset"`
		Instruments []musicbrainz.Instrument `json:"instruments"`
	}{
		Count:       3,
		Instruments: []musicbrainz.Instrument{{ID: "inst-1", Name: "guitar"}},
	}
	ts := newTestServer(t, "/instrument", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.SearchInstruments(context.Background(), "guitar", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Entities[0].Name != "guitar" {
		t.Errorf("Name = %q, want guitar", result.Entities[0].Name)
	}
}

func TestSearchSeries(t *testing.T) {
	t.Parallel()

	resp := struct {
		Created string               `json:"created"`
		Count   int                  `json:"count"`
		Offset  int                  `json:"offset"`
		Series  []musicbrainz.Series `json:"series"`
	}{
		Count:  4,
		Series: []musicbrainz.Series{{ID: "series-1", Name: "BBC Proms"}},
	}
	ts := newTestServer(t, "/series", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.SearchSeries(context.Background(), "proms", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Entities[0].Name != "BBC Proms" {
		t.Errorf("Name = %q, want BBC Proms", result.Entities[0].Name)
	}
}

func TestBrowseReleasesByArtist(t *testing.T) {
	t.Parallel()

	resp := struct {
		ReleaseCount  int                   `json:"release-count"`
		ReleaseOffset int                   `json:"release-offset"`
		Releases      []musicbrainz.Release `json:"releases"`
	}{
		ReleaseCount: 30,
		Releases:     []musicbrainz.Release{{ID: "r1", Title: "Pablo Honey"}},
	}
	ts := newTestServer(t, "/release", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.BrowseReleasesByArtist(context.Background(), "a74b1b7f-71a5-4011-9441-d0b5e4122711", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 30 {
		t.Errorf("Count = %d, want 30", result.Count)
	}
	if len(result.Entities) != 1 {
		t.Fatalf("len = %d, want 1", len(result.Entities))
	}
}

func TestBrowseArtistsByReleaseGroup(t *testing.T) {
	t.Parallel()

	resp := struct {
		ArtistCount  int                  `json:"artist-count"`
		ArtistOffset int                  `json:"artist-offset"`
		Artists      []musicbrainz.Artist `json:"artists"`
	}{
		ArtistCount: 1,
		Artists:     []musicbrainz.Artist{{ID: "artist-1", Name: "Radiohead"}},
	}
	ts := newTestServer(t, "/artist", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.BrowseArtistsByReleaseGroup(context.Background(), "rg1", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Entities[0].Name != "Radiohead" {
		t.Errorf("Name = %q, want Radiohead", result.Entities[0].Name)
	}
}

func TestBrowseArtistsByRecording(t *testing.T) {
	t.Parallel()

	resp := struct {
		ArtistCount  int                  `json:"artist-count"`
		ArtistOffset int                  `json:"artist-offset"`
		Artists      []musicbrainz.Artist `json:"artists"`
	}{
		ArtistCount: 1,
		Artists:     []musicbrainz.Artist{{ID: "artist-1", Name: "Radiohead"}},
	}
	ts := newTestServer(t, "/artist", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.BrowseArtistsByRecording(context.Background(), "recording-1", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 1 {
		t.Errorf("Count = %d, want 1", result.Count)
	}
}

func TestBrowseReleasesByLabel(t *testing.T) {
	t.Parallel()

	resp := struct {
		ReleaseCount  int                   `json:"release-count"`
		ReleaseOffset int                   `json:"release-offset"`
		Releases      []musicbrainz.Release `json:"releases"`
	}{
		ReleaseCount: 12,
		Releases:     []musicbrainz.Release{{ID: "r1", Title: "OK Computer"}},
	}
	ts := newTestServer(t, "/release", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.BrowseReleasesByLabel(context.Background(), "label-1", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 12 {
		t.Errorf("Count = %d, want 12", result.Count)
	}
}

func TestBrowseReleaseGroupsByArtist(t *testing.T) {
	t.Parallel()

	resp := struct {
		ReleaseGroupCount  int                        `json:"release-group-count"`
		ReleaseGroupOffset int                        `json:"release-group-offset"`
		ReleaseGroups      []musicbrainz.ReleaseGroup `json:"release-groups"`
	}{
		ReleaseGroupCount: 15,
		ReleaseGroups:     []musicbrainz.ReleaseGroup{{ID: "rg1", Title: "A Moon Shaped Pool"}},
	}
	ts := newTestServer(t, "/release-group", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.BrowseReleaseGroupsByArtist(context.Background(), "a74b1b7f-71a5-4011-9441-d0b5e4122711", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 15 {
		t.Errorf("Count = %d, want 15", result.Count)
	}
}

func TestBrowseRecordingsByArtist(t *testing.T) {
	t.Parallel()

	resp := struct {
		RecordingCount  int                     `json:"recording-count"`
		RecordingOffset int                     `json:"recording-offset"`
		Recordings      []musicbrainz.Recording `json:"recordings"`
	}{
		RecordingCount: 500,
		Recordings:     []musicbrainz.Recording{{ID: "rec1", Title: "Creep"}},
	}
	ts := newTestServer(t, "/recording", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.BrowseRecordingsByArtist(context.Background(), "a74b1b7f-71a5-4011-9441-d0b5e4122711", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 500 {
		t.Errorf("Count = %d, want 500", result.Count)
	}
}

func TestBrowseRecordingsByRelease(t *testing.T) {
	t.Parallel()

	resp := struct {
		RecordingCount  int                     `json:"recording-count"`
		RecordingOffset int                     `json:"recording-offset"`
		Recordings      []musicbrainz.Recording `json:"recordings"`
	}{
		RecordingCount: 9,
		Recordings:     []musicbrainz.Recording{{ID: "rec1", Title: "No Surprises"}},
	}
	ts := newTestServer(t, "/recording", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.BrowseRecordingsByRelease(context.Background(), "release-1", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Entities[0].Title != "No Surprises" {
		t.Errorf("Title = %q, want No Surprises", result.Entities[0].Title)
	}
}

func TestLookupByISRC(t *testing.T) {
	t.Parallel()

	resp := struct {
		Recordings []musicbrainz.Recording `json:"recordings"`
	}{
		Recordings: []musicbrainz.Recording{{ID: "rec1", Title: "Karma Police"}},
	}
	ts := newTestServer(t, "/isrc/GBAYE9700104", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	recs, err := c.LookupByISRC(context.Background(), "GBAYE9700104", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("len = %d, want 1", len(recs))
	}
	if recs[0].Title != "Karma Police" {
		t.Errorf("Title = %q, want %q", recs[0].Title, "Karma Police")
	}
}

func TestLookupByDiscID(t *testing.T) {
	t.Parallel()

	resp := struct {
		Releases []musicbrainz.Release `json:"releases"`
	}{
		Releases: []musicbrainz.Release{{ID: "r1", Title: "Kid A"}},
	}
	ts := newTestServer(t, "/discid/test-disc-id", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	rels, err := c.LookupByDiscID(context.Background(), "test-disc-id", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rels) != 1 {
		t.Fatalf("len = %d, want 1", len(rels))
	}
}

func TestListGenres(t *testing.T) {
	t.Parallel()

	genres := []musicbrainz.Genre{{ID: "g1", Name: "rock"}, {ID: "g2", Name: "jazz"}}
	ts := newTestServer(t, "/genre/all", genres)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.ListGenres(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Not Found"})
	}))
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	_, err := c.LookupArtist(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *musicbrainz.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
	if apiErr.Message != "Not Found" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Not Found")
	}
	if got := apiErr.Error(); got != "musicbrainz: HTTP 404: Not Found" {
		t.Errorf("Error() = %q, want musicbrainz: HTTP 404: Not Found", got)
	}
}

func TestAPIErrorRawBody(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("service unavailable"))
	}))
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	_, err := c.LookupArtist(context.Background(), "x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *musicbrainz.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.RawBody != "service unavailable" {
		t.Errorf("RawBody = %q, want %q", apiErr.RawBody, "service unavailable")
	}
	if got := apiErr.Error(); got != "musicbrainz: HTTP 503: service unavailable" {
		t.Errorf("Error() = %q, want musicbrainz: HTTP 503: service unavailable", got)
	}
	if got := (&musicbrainz.APIError{StatusCode: http.StatusTooManyRequests}).Error(); got != "musicbrainz: HTTP 429" {
		t.Errorf("Error() without body = %q, want musicbrainz: HTTP 429", got)
	}
}

func TestLookupSeries(t *testing.T) {
	t.Parallel()

	series := musicbrainz.Series{ID: "s1", Name: "BBC Proms", Type: "Festival"}
	ts := newTestServer(t, "/series/s1", series)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	s, err := c.LookupSeries(context.Background(), "s1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "BBC Proms" {
		t.Errorf("Name = %q, want %q", s.Name, "BBC Proms")
	}
}

func TestLookupURL(t *testing.T) {
	t.Parallel()

	u := musicbrainz.URLEntity{ID: "url-1", Resource: "https://example.com"}
	ts := newTestServer(t, "/url/url-1", u)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.LookupURL(context.Background(), "url-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Resource != "https://example.com" {
		t.Errorf("Resource = %q, want %q", result.Resource, "https://example.com")
	}
}

func TestBrowseWorksByArtist(t *testing.T) {
	t.Parallel()

	resp := struct {
		WorkCount  int                `json:"work-count"`
		WorkOffset int                `json:"work-offset"`
		Works      []musicbrainz.Work `json:"works"`
	}{
		WorkCount: 100,
		Works:     []musicbrainz.Work{{ID: "w1", Title: "Exit Music (For a Film)"}},
	}
	ts := newTestServer(t, "/work", resp)
	defer ts.Close()

	c := musicbrainz.New(metadata.WithBaseURL(ts.URL))
	result, err := c.BrowseWorksByArtist(context.Background(), "a74b1b7f-71a5-4011-9441-d0b5e4122711", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 100 {
		t.Errorf("Count = %d, want 100", result.Count)
	}
}
