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
	// The title always shows in the results list regardless of focus.
	view := ws.View(80, 20)
	if !strings.Contains(view, "Feather") {
		t.Fatalf("View missing result title: %q", view)
	}
	// The uploader/duration detail line only shows for the selected result
	// (spec: "Selecting a result shows duration/uploader/source inline"),
	// so move focus into the results list first to check it.
	ws, _ = ws.Update(key("down"))
	if ws.(*Search).focus != searchFocusResults {
		t.Fatal("expected down to move focus into the results list")
	}
	view = ws.View(80, 20)
	if !strings.Contains(view, "Nujabes") {
		t.Fatalf("View missing selected result's detail line: %q", view)
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

func TestSearchDownFromInputEntersSuggestionsBeforeResults(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{MusicDirectory: dir, AudioFormat: "opus"}
	st := style.NewStyles(style.ThemeByName("hacker"))
	historyPath := filepath.Join(dir, "search_history")
	os.WriteFile(historyPath, []byte("nujabes feather"), 0o600)

	s := NewSearch(app.New(cfg), cfg, st, historyPath)
	ws := s.Activate()
	// Populate results too, to prove suggestions win priority while typing.
	ws, _ = ws.Update(searchResultsMsg{results: []download.SearchResult{{Title: "Old result", URL: "https://youtu.be/old"}}})
	for _, r := range "nu" {
		ws, _ = ws.Update(key(string(r)))
	}

	ws, _ = ws.Update(key("down"))
	sr := ws.(*Search)
	if sr.focus != searchFocusSuggestions {
		t.Fatalf("focus = %v, want searchFocusSuggestions", sr.focus)
	}
	if len(sr.activeSuggestions) == 0 || sr.activeSuggestions[0] != "nujabes feather" {
		t.Fatalf("activeSuggestions = %v, want first entry 'nujabes feather'", sr.activeSuggestions)
	}
}

func TestSearchSuggestionEnterFillsInputAndReturnsFocus(t *testing.T) {
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
	ws, _ = ws.Update(key("down")) // enter suggestions
	ws, _ = ws.Update(key("enter"))

	sr := ws.(*Search)
	if sr.focus != searchFocusInput {
		t.Fatalf("focus = %v, want searchFocusInput after accepting a suggestion", sr.focus)
	}
	if !sr.input.Focused() {
		t.Fatal("expected the query input to be focused again after accepting a suggestion")
	}
	if sr.input.Value() != "nujabes feather" {
		t.Fatalf("input value = %q, want %q", sr.input.Value(), "nujabes feather")
	}
}

func TestSearchSuggestionEscDismissesWithoutChangingInput(t *testing.T) {
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
	ws, _ = ws.Update(key("down")) // enter suggestions
	ws, _ = ws.Update(key("esc"))

	sr := ws.(*Search)
	if sr.focus != searchFocusInput {
		t.Fatalf("focus = %v, want searchFocusInput after esc", sr.focus)
	}
	if sr.input.Value() != "nu" {
		t.Fatalf("input value = %q, want unchanged %q", sr.input.Value(), "nu")
	}
}
