package genius

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLyricsSearchAndScrape(t *testing.T) {
	mux := http.NewServeMux()
	var srv *httptest.Server

	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer t0ken" {
			t.Errorf("auth header = %q", got)
		}
		resp := map[string]any{
			"response": map[string]any{
				"hits": []any{
					map[string]any{"result": map[string]any{
						"url":   srv.URL + "/lyrics",
						"title": "Believer",
					}},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/lyrics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body>
		<div data-lyrics-container="true">First line<br/>Second line<br/>You might also like third</div>
		<div>unrelated</div>
		</body></html>`)
	})

	srv = httptest.NewServer(mux)
	defer srv.Close()

	c := New("t0ken")
	c.APIURL = srv.URL

	lyrics, err := c.Lyrics(context.Background(), "Believer", "Imagine Dragons")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(lyrics, "First line") || !strings.Contains(lyrics, "Second line") {
		t.Errorf("lyrics missing content: %q", lyrics)
	}
	if strings.Contains(lyrics, "You might also like") {
		t.Errorf("lyrics should strip 'You might also like': %q", lyrics)
	}
	if strings.Contains(lyrics, "unrelated") {
		t.Errorf("lyrics leaked non-container content: %q", lyrics)
	}
}

func TestLyricsNoToken(t *testing.T) {
	c := New("")
	got, err := c.Lyrics(context.Background(), "a", "b")
	if err != nil || got != "" {
		t.Errorf("expected empty result with no token, got %q err %v", got, err)
	}
}
