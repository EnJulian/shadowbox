package lastfm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenreFromTrackTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("method") == "track.getinfo" {
			w.Write([]byte(`{"track":{"toptags":{"tag":[{"name":"indie rock"},{"name":"alternative"}]}}}`))
			return
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New()
	c.BaseURL = srv.URL
	genre, err := c.Genre(context.Background(), "Believer", "Imagine Dragons")
	if err != nil {
		t.Fatal(err)
	}
	if genre != "indie rock" {
		t.Errorf("genre = %q, want indie rock", genre)
	}
}

func TestGenreFallsBackToArtistTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("method") {
		case "track.getinfo":
			w.Write([]byte(`{"track":{"toptags":{"tag":[]}}}`))
		case "artist.gettoptags":
			w.Write([]byte(`{"toptags":{"tag":[{"name":"pop"}]}}`))
		default:
			w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	c := New()
	c.BaseURL = srv.URL
	genre, err := c.Genre(context.Background(), "Song", "Artist")
	if err != nil {
		t.Fatal(err)
	}
	if genre != "pop" {
		t.Errorf("genre = %q, want pop", genre)
	}
}
