package ui

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestListTracksFindsNestedAudioFiles(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "01 - Intro.mp3"))
	mustWrite(t, filepath.Join(dir, "CD1", "01 - Track One.mp3"))
	mustWrite(t, filepath.Join(dir, "CD2", "01 - Track Two.mp3"))
	mustWrite(t, filepath.Join(dir, "CD2", "bonus", "Hidden Track.flac"))
	mustWrite(t, filepath.Join(dir, "cover.jpg")) // non-audio, must be excluded

	got := listTracks(dir)
	sort.Strings(got)

	want := []string{
		"01 - Intro.mp3",
		"CD1/01 - Track One.mp3",
		"CD2/01 - Track Two.mp3",
		"CD2/bonus/Hidden Track.flac",
	}
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("listTracks(%q) = %v, want %v", dir, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("listTracks(%q) = %v, want %v", dir, got, want)
		}
	}
}

func TestEnteringAlbumQueuesTracksFromNestedDiscFolders(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "Artist A", "Album B", "CD1", "01.mp3"))
	mustWrite(t, filepath.Join(dir, "Artist A", "Album B", "CD2", "01.mp3"))

	m := newTestModel()
	m.cfg.MusicDirectory = dir

	next, _ := m.openLibrary()
	m = next.(model) // level 0: Artists

	next, _ = m.libraryEnter()
	m = next.(model) // level 1: Albums under "Artist A"

	next, _ = m.libraryEnter()
	m = next.(model) // level 2: Tracks under "Album B" (recursed into CD1/CD2)

	next, cmd := m.libraryEnter()
	m = next.(model)
	if cmd == nil {
		t.Fatal("expected libraryEnter to return a startPlaybackMsg command")
	}
	msg, ok := cmd().(startPlaybackMsg)
	if !ok {
		t.Fatalf("expected startPlaybackMsg, got %T", cmd())
	}
	if len(msg.tracks) != 2 {
		t.Fatalf("expected both CD1 and CD2 tracks queued for the album, got %d: %+v", len(msg.tracks), msg.tracks)
	}
	wantCD1 := filepath.Join(dir, "Artist A", "Album B", "CD1", "01.mp3")
	wantCD2 := filepath.Join(dir, "Artist A", "Album B", "CD2", "01.mp3")
	gotPaths := map[string]bool{msg.tracks[0].Path: true, msg.tracks[1].Path: true}
	if !gotPaths[wantCD1] || !gotPaths[wantCD2] {
		t.Fatalf("expected queue to contain %q and %q, got %+v", wantCD1, wantCD2, msg.tracks)
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}
