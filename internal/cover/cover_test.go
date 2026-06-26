package cover

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EnJulian/shadowbox/internal/apis/itunes"
	"github.com/EnJulian/shadowbox/internal/apis/spotify"
)

func TestURLPrefersSpotify(t *testing.T) {
	sp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Write([]byte(`{"access_token":"t","expires_in":3600}`))
		default: // /search
			w.Write([]byte(`{"tracks":{"items":[{"name":"S","artists":[{"name":"A"}],"album":{"name":"Al","images":[{"url":"spotify-cover","width":640,"height":640}]}}]}}`))
		}
	}))
	defer sp.Close()

	spc := spotify.New("id", "secret")
	spc.TokenURL = sp.URL + "/token"
	spc.APIURL = sp.URL

	itc := itunes.New()
	itc.BaseURL = sp.URL // unused; spotify should win

	r := New(spc, itc)
	got := r.URL(context.Background(), "Song", "Artist")
	if got != "spotify-cover" {
		t.Errorf("cover = %q, want spotify-cover", got)
	}
}

func TestURLFallsBackToITunes(t *testing.T) {
	itServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"results":[{"artworkUrl100":"https://cdn/100x100bb.jpg"}]}`))
	}))
	defer itServer.Close()

	itc := itunes.New()
	itc.BaseURL = itServer.URL

	// No Spotify configured -> straight to iTunes.
	r := New(spotify.New("", ""), itc)
	got := r.URL(context.Background(), "Song", "Artist")
	if got != "https://cdn/600x600bb.jpg" {
		t.Errorf("cover = %q, want upgraded itunes url", got)
	}
}

func TestDownloadDetectsMIME(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("PNGDATA"))
	}))
	defer srv.Close()

	r := New(nil, nil)
	data, mime, err := r.Download(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "PNGDATA" || mime != "image/png" {
		t.Errorf("got data=%q mime=%q", data, mime)
	}
}
