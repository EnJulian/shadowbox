package cover

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/apis/itunes"
	"github.com/EnJulian/shadowbox/internal/apis/musicbrainz"
)

func TestURLPrefersMusicBrainz(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/recording") {
			w.Write([]byte(`{"recordings":[{
				"id":"rec1","score":100,"title":"Song",
				"artist-credit":[{"name":"Artist","artist":{"name":"Artist"}}],
				"releases":[{"id":"rel1","status":"Official","title":"Album",
					"release-group":{"title":"Album"},
					"media":[{"track":[{"number":"1","title":"Song"}]}]
				}]
			}]}`))
			return
		}
		w.Write([]byte(`{"images":[{"front":true,"thumbnails":{"500":"https://cdn/mb-cover.jpg"}}]}`))
	}))
	defer srv.Close()

	mb := musicbrainz.New()
	mb.BaseURL = srv.URL
	mb.CoverBase = srv.URL

	itc := itunes.New()
	itc.BaseURL = srv.URL

	r := New(mb, itc)
	got := r.URL(context.Background(), "Song", "Artist")
	if got != "https://cdn/mb-cover.jpg" {
		t.Errorf("cover = %q, want mb-cover", got)
	}
}

func TestURLFallsBackToITunes(t *testing.T) {
	itServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"results":[{"artworkUrl100":"https://cdn/100x100bb.jpg"}]}`))
	}))
	defer itServer.Close()

	mb := musicbrainz.New()
	mb.BaseURL = itServer.URL

	itc := itunes.New()
	itc.BaseURL = itServer.URL

	r := New(mb, itc)
	got := r.URL(context.Background(), "Song", "Artist")
	if got != "https://cdn/600x600bb.jpg" {
		t.Errorf("cover = %q, want upgraded itunes url", got)
	}
}

func TestDownloadDetectsMIME(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("PNGDATA"))
	}))
	defer srv.Close()

	r := New(nil, nil)
	r.HTTP = srv.Client()
	data, mime, err := r.Download(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "PNGDATA" || mime != "image/png" {
		t.Errorf("got data=%q mime=%q", data, mime)
	}
}

func TestDownloadRejectsHTTP(t *testing.T) {
	r := New(nil, nil)
	_, _, err := r.Download(context.Background(), "http://example.com/cover.jpg")
	if err == nil || !strings.Contains(err.Error(), "HTTPS") {
		t.Fatalf("Download(http) = %v, want HTTPS error", err)
	}
}

func TestDownloadRejectsNonOKStatus(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	r := New(nil, nil)
	r.HTTP = srv.Client()
	_, _, err := r.Download(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("Download(404) = %v, want status error", err)
	}
}

func TestDownloadRejectsNonImageContentType(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html></html>"))
	}))
	defer srv.Close()

	r := New(nil, nil)
	r.HTTP = srv.Client()
	_, _, err := r.Download(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "not an image") {
		t.Fatalf("Download(html) = %v, want content-type error", err)
	}
}

func TestDownloadRejectsOversizedImage(t *testing.T) {
	payload := strings.Repeat("x", maxCoverBytes+1)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = io.WriteString(w, payload)
	}))
	defer srv.Close()

	r := New(nil, nil)
	r.HTTP = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	_, _, err := r.Download(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("Download(oversized) = %v, want size limit error", err)
	}
}
