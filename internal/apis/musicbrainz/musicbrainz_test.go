package musicbrainz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const sampleSearch = `{
  "recordings": [{
    "id": "rec1",
    "score": 100,
    "title": "Believer",
    "first-release-date": "2017",
    "artist-credit": [{"name": "Imagine Dragons", "joinphrase": "", "artist": {"name": "Imagine Dragons"}}],
    "releases": [{
      "id": "rel-official",
      "title": "Evolve",
      "status": "Official",
      "date": "2017-06-23",
      "track-count": 11,
      "release-group": {"title": "Evolve", "primary-type": "Album"},
      "media": [{
        "position": 1,
        "format": "CD",
        "track-count": 11,
        "track": [{"number": "3", "title": "Believer", "length": 204000}]
      }]
    }, {
      "id": "rel-bootleg",
      "title": "Live Bootleg",
      "status": "Bootleg",
      "track-count": 5,
      "release-group": {"title": "Live Bootleg", "primary-type": "Album", "secondary-types": ["Live"]},
      "media": [{"position": 1, "track-count": 5, "track": [{"number": "1", "title": "Believer"}]}]
    }],
    "tags": [{"name": "pop rock", "count": 5}]
  }]
}`

const sampleCover = `{
  "images": [{
    "front": true,
    "image": "https://archive.org/full.jpg",
    "thumbnails": {"500": "https://archive.org/thumb500.jpg", "250": "https://archive.org/thumb250.jpg"}
  }]
}`

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New()
	c.BaseURL = srv.URL
	c.CoverBase = srv.URL
	return c
}

const sampleMultiSearch = `{
  "recordings": [{
    "id": "rec1",
    "score": 100,
    "title": "Believer",
    "length": 204000,
    "first-release-date": "2017",
    "artist-credit": [{"name": "Imagine Dragons", "joinphrase": "", "artist": {"name": "Imagine Dragons"}}],
    "releases": [{
      "id": "rel-official",
      "title": "Evolve",
      "status": "Official",
      "date": "2017-06-23",
      "track-count": 11,
      "release-group": {"title": "Evolve", "primary-type": "Album"},
      "media": [{"position": 1, "track-count": 11, "track": [{"number": "3", "title": "Believer"}]}]
    }]
  }, {
    "id": "rec2",
    "score": 80,
    "title": "Believer",
    "length": 310000,
    "first-release-date": "2018",
    "artist-credit": [{"name": "Imagine Dragons", "joinphrase": "", "artist": {"name": "Imagine Dragons"}}],
    "releases": [{
      "id": "rel-live",
      "title": "Live at Red Rocks",
      "status": "Official",
      "date": "2018-01-01",
      "track-count": 8,
      "release-group": {"title": "Live at Red Rocks", "primary-type": "Album", "secondary-types": ["Live"]},
      "media": [{"position": 1, "track-count": 8, "track": [{"number": "1", "title": "Believer"}]}]
    }]
  }]
}`

func TestSearchCandidatesReturnsMultiple(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/recording") {
			w.Write([]byte(sampleMultiSearch))
			return
		}
		w.Write([]byte(sampleCover))
	}))

	candidates, err := c.SearchCandidates(context.Background(), "Believer", "Imagine Dragons")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatalf("len = %d, want 2", len(candidates))
	}
	if candidates[0].Artist != "Imagine Dragons" {
		t.Errorf("artist = %q", candidates[0].Artist)
	}
	if candidates[0].Length != "3:24" {
		t.Errorf("length = %q, want 3:24", candidates[0].Length)
	}
}

func TestMetaFromCandidate(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/recording") {
			w.Write([]byte(sampleMultiSearch))
			return
		}
		w.Write([]byte(sampleCover))
	}))

	candidates, err := c.SearchCandidates(context.Background(), "Believer", "Imagine Dragons")
	if err != nil {
		t.Fatal(err)
	}
	meta, err := c.MetaFromCandidate(context.Background(), candidates[1], "Believer")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Album != "Live at Red Rocks" {
		t.Errorf("album = %q, want Live at Red Rocks", meta.Album)
	}
}

func TestSearchMapsMetadata(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}
		switch {
		case strings.Contains(r.URL.Path, "/recording"):
			w.Write([]byte(sampleSearch))
		case strings.Contains(r.URL.Path, "/release/"):
			w.Write([]byte(sampleCover))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))

	meta, err := c.Search(context.Background(), "Believer", "Imagine Dragons")
	if err != nil {
		t.Fatal(err)
	}
	if meta == nil {
		t.Fatal("expected metadata")
	}
	if meta.Title != "Believer" || meta.Artist != "Imagine Dragons" || meta.Album != "Evolve" {
		t.Errorf("unexpected meta: %+v", meta)
	}
	if meta.ReleaseDate != "2017-06-23" || meta.TrackNumber != "3" || meta.TotalTracks != "11" {
		t.Errorf("track/date fields: %+v", meta)
	}
	if meta.Genre != "pop rock" {
		t.Errorf("genre = %q, want pop rock", meta.Genre)
	}
	if meta.CoverURL != "https://archive.org/thumb500.jpg" {
		t.Errorf("cover = %q", meta.CoverURL)
	}
}

func TestSearchPrefersOfficialRelease(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/recording") {
			w.Write([]byte(sampleSearch))
			return
		}
		w.Write([]byte(`{"images":[]}`))
	}))

	meta, err := c.Search(context.Background(), "Believer", "Imagine Dragons")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Album != "Evolve" {
		t.Errorf("album = %q, want Evolve (official over bootleg)", meta.Album)
	}
}

func TestSearchTitleOnlyFallback(t *testing.T) {
	calls := 0
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		q := r.URL.Query().Get("query")
		if strings.Contains(q, "artist:") {
			w.Write([]byte(`{"recordings":[]}`))
			return
		}
		w.Write([]byte(sampleSearch))
	}))

	meta, err := c.Search(context.Background(), "Believer", "Imagine Dragons")
	if err != nil {
		t.Fatal(err)
	}
	if meta == nil {
		t.Fatal("expected metadata from title-only fallback")
	}
	if calls < 2 {
		t.Errorf("expected at least 2 search attempts, got %d", calls)
	}
}

func TestSearchEmptyResult(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"recordings":[]}`))
	}))

	meta, err := c.Search(context.Background(), "Unknown Track", "Nobody")
	if err != nil {
		t.Fatal(err)
	}
	if meta != nil {
		t.Errorf("expected nil meta, got %+v", meta)
	}
}

func TestSearchUsesCache(t *testing.T) {
	calls := 0
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if strings.Contains(r.URL.Path, "/recording") {
			w.Write([]byte(sampleSearch))
			return
		}
		w.Write([]byte(sampleCover))
	}))

	ctx := context.Background()
	if _, err := c.Search(ctx, "Believer", "Imagine Dragons"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Search(ctx, "Believer", "Imagine Dragons"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Errorf("expected 2 HTTP calls (search + cover), got %d", calls)
	}
}

func TestThrottleEnforcesGap(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"recordings":[]}`))
	}))

	start := time.Now()
	ctx := context.Background()
	if err := c.getJSON(ctx, c.BaseURL+"/recording?query=test&fmt=json", &searchResponse{}); err != nil {
		t.Fatal(err)
	}
	if err := c.getJSON(ctx, c.BaseURL+"/recording?query=test2&fmt=json", &searchResponse{}); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed < minRequestGap {
		t.Errorf("requests too fast: %v", elapsed)
	}
}

func TestCoverURLMissingRelease(t *testing.T) {
	c := New()
	u, err := c.CoverURL(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if u != "" {
		t.Errorf("cover = %q, want empty", u)
	}
}
