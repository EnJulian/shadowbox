package organize

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	cases := map[string]string{
		`AC/DC`:            "AC_DC",
		`What: A Day`:      "What_ A Day",
		`  .trim. `:        "trim",
		``:                 "Unknown",
		`a<b>c|d?`:         "a_b_c_d_",
		`Song "Title"`:     "Song _Title_",
	}
	for in, want := range cases {
		if got := SanitizeFilename(in); got != want {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestArtistAndAlbumDir(t *testing.T) {
	base := t.TempDir()
	ad, err := ArtistDir(base, "Imagine Dragons, Someone")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(ad) != "Imagine Dragons" {
		t.Errorf("artist dir = %q, want first artist only", filepath.Base(ad))
	}

	bd, err := AlbumDir(ad, "Album - Evolve", "Believer")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(bd) != "Evolve" {
		t.Errorf("album dir = %q, want 'Album - ' prefix stripped", filepath.Base(bd))
	}

	// Empty album falls back to title.
	bd2, _ := AlbumDir(ad, "", "OnlyTitle")
	if filepath.Base(bd2) != "OnlyTitle" {
		t.Errorf("album dir = %q, want fallback to title", filepath.Base(bd2))
	}
}

func TestUniquePath(t *testing.T) {
	dir := t.TempDir()
	p1 := UniquePath(dir, "song", ".opus")
	if filepath.Base(p1) != "song.opus" {
		t.Fatalf("got %q", p1)
	}
	os.WriteFile(p1, []byte("x"), 0o644)
	p2 := UniquePath(dir, "song", ".opus")
	if filepath.Base(p2) != "song_1.opus" {
		t.Errorf("got %q, want song_1.opus", p2)
	}
}
