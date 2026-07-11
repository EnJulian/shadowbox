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
	view := ws.View(80, 20)
	if strings.Contains(view, "Nujabes") {
		t.Fatalf("expected Nujabes filtered out, got %q", view)
	}
	if !strings.Contains(view, "Kanye West") {
		t.Fatalf("expected Kanye West to remain, got %q", view)
	}
}
