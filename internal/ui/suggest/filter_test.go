package suggest

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFilterCaseInsensitiveSubstring(t *testing.T) {
	entries := []string{"Nujabes", "Kanye West", "Aphex Twin"}
	got := Filter(entries, "we")
	want := []string{"Kanye West"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Filter = %v, want %v", got, want)
	}
}

func TestFilterEmptyQueryReturnsAll(t *testing.T) {
	entries := []string{"a", "b"}
	if got := Filter(entries, ""); !reflect.DeepEqual(got, entries) {
		t.Fatalf("Filter with empty query = %v, want %v", got, entries)
	}
}

func TestLibraryMatchesFindsArtistAlbumAndTrack(t *testing.T) {
	dir := t.TempDir()
	trackDir := filepath.Join(dir, "Nujabes", "Modal Soul")
	if err := os.MkdirAll(trackDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(trackDir, "Feather.mp3"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	got := LibraryMatches(dir, "feather", 10)
	if len(got) != 1 || got[0] != "Feather.mp3" {
		t.Fatalf("LibraryMatches = %v, want [Feather.mp3]", got)
	}

	got = LibraryMatches(dir, "nujabes", 10)
	if len(got) != 1 || got[0] != "Nujabes" {
		t.Fatalf("LibraryMatches = %v, want [Nujabes]", got)
	}
}
