// internal/ui/workspace/library_test.go
package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func newTestLibrary(t *testing.T) *Library {
	t.Helper()
	dir := t.TempDir()
	trackDir := filepath.Join(dir, "Nujabes", "Modal Soul")
	if err := os.MkdirAll(trackDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(trackDir, "Feather.mp3"), []byte{}, 0o644)
	os.MkdirAll(filepath.Join(dir, "Kanye West"), 0o755)

	cfg := &config.Config{MusicDirectory: dir}
	st := style.NewStyles(style.ThemeByName("hacker"))
	return NewLibrary(cfg, st)
}

func TestLibraryActivateListsArtists(t *testing.T) {
	l := newTestLibrary(t)
	ws := l.Activate()
	view := ws.View(80, 20)
	if !strings.Contains(view, "Nujabes") || !strings.Contains(view, "Kanye West") {
		t.Fatalf("expected both artists listed, got %q", view)
	}
}

func TestLibraryDrillDownToTrack(t *testing.T) {
	l := newTestLibrary(t)
	ws := l.Activate()
	// Nujabes sorts before "Kanye West"? No: alphabetical, "Kanye West" < "Nujabes".
	// Move cursor down to Nujabes explicitly rather than assuming order.
	for i := 0; i < 5; i++ {
		view := ws.View(80, 20)
		if strings.Contains(view, "> Nujabes") {
			break
		}
		ws, _ = ws.Update(key("down"))
	}
	ws, _ = ws.Update(key("enter")) // into albums
	ws, _ = ws.Update(key("enter")) // into tracks
	view := ws.View(80, 20)
	if !strings.Contains(view, "Feather.mp3") {
		t.Fatalf("expected track listed after drill-down, got %q", view)
	}
}

func TestLibraryLeftFromArtistColumnRequestsNavFocus(t *testing.T) {
	l := newTestLibrary(t)
	ws := l.Activate()
	_, cmd := ws.Update(key("left"))
	if cmd == nil {
		t.Fatal("expected a FocusNavMsg cmd when pressing left at the Artist column")
	}
}

func TestLibraryTypeAheadFiltersEntries(t *testing.T) {
	l := newTestLibrary(t)
	ws := l.Activate()
	for _, r := range "kan" {
		ws, _ = ws.Update(key(string(r)))
	}
	// Assert the actual accumulated filter string, not just the visible
	// outcome — a prior version of this test typed "kan" but the buggy
	// production code swallowed the "k" as an "up" navigation key, leaving
	// the real filter as "an"; the view assertions below still happened to
	// pass by coincidence (both "an" and "kan" match "Kanye West" but not
	// "Nujabes"), silently masking the bug. Asserting the filter value
	// directly closes that false-positive gap.
	lib := ws.(*Library)
	if lib.filter != "kan" {
		t.Fatalf("filter = %q, want %q (letter keys must all reach the filter, not be swallowed as navigation)", lib.filter, "kan")
	}
	view := ws.View(80, 20)
	if strings.Contains(view, "Nujabes") {
		t.Fatalf("expected Nujabes filtered out, got %q", view)
	}
	if !strings.Contains(view, "Kanye West") {
		t.Fatalf("expected Kanye West to remain, got %q", view)
	}
}

func TestLibraryTextFocusedAlwaysTrue(t *testing.T) {
	l := newTestLibrary(t)
	ws := l.Activate()
	if !ws.(*Library).TextFocused() {
		t.Fatal("TextFocused() = false, want true right after Activate")
	}

	// Still true after typing into the filter, and after drilling down —
	// every level of this workspace routes typed runes to the filter, so
	// global shortcuts must never be allowed to intercept them.
	for _, r := range "kan" {
		ws, _ = ws.Update(key(string(r)))
	}
	if !ws.(*Library).TextFocused() {
		t.Fatal("TextFocused() = false, want true after typing into the filter")
	}
	ws, _ = ws.Update(key("enter")) // drill into Kanye West
	if !ws.(*Library).TextFocused() {
		t.Fatal("TextFocused() = false, want true after drilling into a level")
	}
}

func TestLibraryLetterKeysAlwaysFilterNeverNavigate(t *testing.T) {
	l := newTestLibrary(t)
	ws := l.Activate()
	// h, j, k, l are also vim navigation mnemonics elsewhere in this app,
	// but Library is a type-ahead filter: every one of these letters must
	// be treated as filter text, never as back/down/up/enter shortcuts.
	for _, r := range "hjkl" {
		ws, _ = ws.Update(key(string(r)))
	}
	lib := ws.(*Library)
	if lib.filter != "hjkl" {
		t.Fatalf("filter = %q, want %q — h/j/k/l must never be swallowed as navigation while filtering", lib.filter, "hjkl")
	}
	if lib.level != 0 {
		t.Fatalf("level = %d, want 0 — typing letters must never change drill-down level", lib.level)
	}
}
