package itunes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const sampleSearch = `{
  "results": [
    {
      "trackName": "Believer",
      "artistName": "Imagine Dragons",
      "collectionName": "Evolve",
      "primaryGenreName": "Alternative",
      "releaseDate": "2017-06-23T12:00:00Z",
      "trackNumber": 3,
      "trackCount": 11,
      "artworkUrl100": "https://cdn/a/100x100bb.jpg"
    },
    {
      "trackName": "Believer",
      "artistName": "Other Artist",
      "collectionName": "Covers",
      "primaryGenreName": "Rock",
      "releaseDate": "2020-01-01T12:00:00Z",
      "trackNumber": 1,
      "trackCount": 1,
      "artworkUrl100": "https://cdn/b/100x100bb.jpg"
    }
  ]
}`

func TestCoverURLUpgradesResolution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("term") == "" {
			t.Error("missing term query")
		}
		w.Write([]byte(`{"results":[{"artworkUrl100":"https://cdn/a/100x100bb.jpg","primaryGenreName":"Rock"}]}`))
	}))
	defer srv.Close()

	c := New()
	c.BaseURL = srv.URL

	url, err := c.CoverURL(context.Background(), "Believer Imagine Dragons")
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://cdn/a/600x600bb.jpg" {
		t.Errorf("cover url = %q, want upgraded 600x600", url)
	}
}

func TestCoverURLNoResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()
	c := New()
	c.BaseURL = srv.URL
	url, err := c.CoverURL(context.Background(), "nothing")
	if err != nil {
		t.Fatal(err)
	}
	if url != "" {
		t.Errorf("expected empty url, got %q", url)
	}
}

func TestSearchCandidatesMapsMetadata(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("entity") != "song" {
			t.Errorf("entity = %q", r.URL.Query().Get("entity"))
		}
		w.Write([]byte(sampleSearch))
	}))
	defer srv.Close()

	c := New()
	c.BaseURL = srv.URL
	candidates, err := c.SearchCandidates(context.Background(), "Believer", "Imagine Dragons")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatalf("len = %d, want 2", len(candidates))
	}
	if candidates[0].Title != "Believer" || candidates[0].Artist != "Imagine Dragons" {
		t.Errorf("first candidate: %+v", candidates[0])
	}
	if candidates[0].Album != "Evolve" || candidates[0].Genre != "Alternative" {
		t.Errorf("album/genre: %+v", candidates[0])
	}
	if candidates[0].CoverURL != "https://cdn/a/600x600bb.jpg" {
		t.Errorf("cover = %q", candidates[0].CoverURL)
	}

	meta := c.MetaFromCandidate(candidates[0])
	if meta.Title != "Believer" || meta.Artist != "Imagine Dragons" || meta.Genre != "Alternative" {
		t.Errorf("meta = %+v", meta)
	}
}
