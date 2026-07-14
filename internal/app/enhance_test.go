package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/EnJulian/shadowbox/internal/apis/genius"
	"github.com/EnJulian/shadowbox/internal/apis/itunes"
	"github.com/EnJulian/shadowbox/internal/apis/lastfm"
	"github.com/EnJulian/shadowbox/internal/apis/musicbrainz"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/cover"
	"github.com/EnJulian/shadowbox/internal/tag"
)

// TestEnhanceDirAutoAcceptsMatchesWithoutPrompting verifies that batch
// enhancement never blocks on an interactive candidate picker, even when a
// file's metadata search returns multiple artist-matching candidates.
func TestEnhanceDirAutoAcceptsMatchesWithoutPrompting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Imagine Dragons - Believer.mp3")
	// A short run of silent MPEG frames; enough for id3v2 to read a header and
	// prepend a tag (same fixture technique as internal/tag's MP3 round-trip test).
	frame := []byte{0xFF, 0xFB, 0x90, 0x00}
	buf := bytes.Repeat(append(frame, make([]byte, 413)...), 4)
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		t.Fatal(err)
	}

	itunesSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// artworkUrl100 deliberately uses http (not https): cover.Download
		// rejects non-https URLs before making any network call, so this
		// keeps the test hermetic without a real (slow) DNS lookup.
		w.Write([]byte(`{"results":[
			{"trackName":"Believer","artistName":"Imagine Dragons","collectionName":"Evolve","primaryGenreName":"Alternative","releaseDate":"2017-06-23T12:00:00Z","artworkUrl100":"http://cdn/a/100x100bb.jpg"},
			{"trackName":"Believer (Live)","artistName":"Imagine Dragons","collectionName":"Live","primaryGenreName":"Alternative","releaseDate":"2018-01-01T12:00:00Z","artworkUrl100":"http://cdn/b/100x100bb.jpg"}
		]}`))
	}))
	defer itunesSrv.Close()

	var mbCalled int32
	mbSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&mbCalled, 1)
		w.Write([]byte(`{"recordings":[]}`))
	}))
	defer mbSrv.Close()

	var lastfmCalled int32
	lastfmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&lastfmCalled, 1)
		w.Write([]byte(`{}`))
	}))
	defer lastfmSrv.Close()

	it := itunes.New()
	it.BaseURL = itunesSrv.URL
	mb := musicbrainz.New()
	mb.BaseURL = mbSrv.URL
	lf := lastfm.New()
	lf.BaseURL = lastfmSrv.URL

	a := &App{
		cfg:         &config.Config{},
		itunes:      it,
		musicbrainz: mb,
		lastfm:      lf,
		genius:      genius.New(""),
		cover:       cover.New(mb, it),
	}

	promptCalled := false
	opts := Options{
		Select: func(ctx context.Context, req PromptRequest) (int, error) {
			promptCalled = true
			return 0, nil
		},
	}

	if err := a.EnhanceDir(context.Background(), dir, false, []string{"mp3"}, false, opts); err != nil {
		t.Fatalf("EnhanceDir: %v", err)
	}

	if promptCalled {
		t.Error("expected batch enhancement to auto-accept the top match without prompting")
	}
	if atomic.LoadInt32(&mbCalled) != 0 {
		t.Error("expected MusicBrainz not to be queried once iTunes already matched")
	}

	got, err := tag.Read(path)
	if err != nil {
		t.Fatalf("Read tagged file: %v", err)
	}
	if got.Title != "Believer" || got.Artist != "Imagine Dragons" {
		t.Errorf("tags = %+v, want Believer / Imagine Dragons", got)
	}
}
