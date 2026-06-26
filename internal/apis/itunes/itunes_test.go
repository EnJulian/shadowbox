package itunes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
