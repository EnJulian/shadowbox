package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/download"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestSearch(t *testing.T) *Search {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{MusicDirectory: dir, AudioFormat: "opus"}
	st := style.NewStyles(style.ThemeByName("hacker"))
	return NewSearch(app.New(cfg), cfg, st, filepath.Join(dir, "search_history"))
}

func key(s string) tea.KeyMsg {
	switch s {
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestSearchActivateFocusesQueryInput(t *testing.T) {
	s := newTestSearch(t)
	ws := s.Activate()
	if !ws.(*Search).input.Focused() {
		t.Fatal("Activate() did not focus the query input")
	}
}

func TestSearchResultsMsgPopulatesList(t *testing.T) {
	s := newTestSearch(t)
	ws, _ := s.Update(searchResultsMsg{results: []download.SearchResult{
		{Title: "Feather", Uploader: "Nujabes", Duration: "3:42", URL: "https://youtu.be/1"},
	}})
	view := ws.View(80, 20)
	if !strings.Contains(view, "Feather") || !strings.Contains(view, "Nujabes") {
		t.Fatalf("View missing result details: %q", view)
	}
}

func TestSearchEnterOnResultStartsTaskAndSwitchesSection(t *testing.T) {
	s := newTestSearch(t)
	ws, _ := s.Update(searchResultsMsg{results: []download.SearchResult{
		{Title: "Feather", URL: "https://youtu.be/1"},
	}})
	// Move focus from the input into the results list (no cmd expected —
	// this is a local focus change, not a side effect), then select.
	ws, _ = ws.Update(key("down"))
	if ws.(*Search).focus != searchFocusResults {
		t.Fatal("expected down to move focus into the results list")
	}
	ws, cmd := ws.Update(key("enter"))
	if cmd == nil {
		t.Fatal("expected a cmd after pressing enter on a result")
	}
	msgs := flattenBatch(cmd())
	var sawStart, sawSwitch bool
	for _, m := range msgs {
		if _, ok := m.(StartTaskMsg); ok {
			sawStart = true
		}
		if ss, ok := m.(SwitchSectionMsg); ok && ss.Section == SectionDownloads {
			sawSwitch = true
		}
	}
	if !sawStart || !sawSwitch {
		t.Fatalf("expected StartTaskMsg and SwitchSectionMsg{Downloads}, got %#v", msgs)
	}
	_ = ws
}

// flattenBatch unwraps a tea.Batch'd cmd's messages for assertions.
func flattenBatch(msg tea.Msg) []tea.Msg {
	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, cmd := range batch {
			out = append(out, flattenBatch(cmd())...)
		}
		return out
	}
	return []tea.Msg{msg}
}

func TestSearchTypingShowsHistorySuggestion(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{MusicDirectory: dir, AudioFormat: "opus"}
	st := style.NewStyles(style.ThemeByName("hacker"))
	historyPath := filepath.Join(dir, "search_history")
	os.WriteFile(historyPath, []byte("nujabes feather"), 0o600)

	s := NewSearch(app.New(cfg), cfg, st, historyPath)
	ws := s.Activate()
	for _, r := range "nu" {
		ws, _ = ws.Update(key(string(r)))
	}
	view := ws.View(80, 20)
	if !strings.Contains(view, "nujabes feather") {
		t.Fatalf("expected history suggestion in view, got %q", view)
	}
}
