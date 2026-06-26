package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(handler http.Handler) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	c := New("id", "secret")
	c.TokenURL = srv.URL + "/token"
	c.APIURL = srv.URL
	return c, srv
}

func TestSearchAndAlbumEnrichment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
	})
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("missing bearer token, got %q", got)
		}
		w.Write([]byte(`{"tracks":{"items":[{
			"name":"Believer","track_number":3,"disc_number":1,
			"artists":[{"name":"Imagine Dragons"}],
			"album":{"id":"alb1","name":"Evolve","release_date":"2017-06-23","total_tracks":11,
			  "images":[{"url":"small","width":64,"height":64},{"url":"big","width":640,"height":640}]}
		}]}}`))
	})
	mux.HandleFunc("/albums/alb1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"genres":["Pop Rock"],"tracks":{"items":[{"disc_number":1},{"disc_number":2}]}}`))
	})

	c, srv := newTestClient(mux)
	defer srv.Close()

	meta, err := c.Search(context.Background(), "Believer", "Imagine Dragons")
	if err != nil {
		t.Fatal(err)
	}
	if meta == nil {
		t.Fatal("expected metadata")
	}
	if meta.Title != "Believer" || meta.Album != "Evolve" {
		t.Errorf("unexpected meta: %+v", meta)
	}
	if meta.Genre != "Pop Rock" {
		t.Errorf("genre = %q, want Pop Rock", meta.Genre)
	}
	if meta.TotalDiscs != "2" {
		t.Errorf("total discs = %q, want 2", meta.TotalDiscs)
	}
	if meta.CoverURL != "big" {
		t.Errorf("cover = %q, want big (largest image)", meta.CoverURL)
	}
	if meta.Year() != "2017" {
		t.Errorf("year = %q", meta.Year())
	}
}

func TestSearchFallsBackToTitleOnly(t *testing.T) {
	var queries []string
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
	})
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		queries = append(queries, q)
		if strings.Contains(q, "artist:") {
			w.Write([]byte(`{"tracks":{"items":[]}}`)) // no results for qualified query
			return
		}
		w.Write([]byte(`{"tracks":{"items":[{"name":"X","artists":[{"name":"Y"}],"album":{"name":"Z"}}]}}`))
	})
	c, srv := newTestClient(mux)
	defer srv.Close()

	meta, err := c.Search(context.Background(), "X", "Y")
	if err != nil {
		t.Fatal(err)
	}
	if meta == nil || meta.Title != "X" {
		t.Fatalf("expected fallback hit, got %+v", meta)
	}
	if len(queries) != 2 {
		t.Errorf("expected 2 search attempts, got %d: %v", len(queries), queries)
	}
}

func TestSearchNoCredentials(t *testing.T) {
	c := New("", "")
	if _, err := c.Search(context.Background(), "a", "b"); err != ErrNoCredentials {
		t.Errorf("expected ErrNoCredentials, got %v", err)
	}
}
