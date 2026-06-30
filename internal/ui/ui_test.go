package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
)

func newTestModel() model {
	cfg := &config.Config{AudioFormat: "opus", MusicDirectory: ".", Theme: "hacker", UseGenius: true}
	theme := themeByName(cfg.Theme)
	ti := textinput.New()
	return model{
		cfg:     cfg,
		app:     app.New(cfg),
		theme:   theme,
		st:      newStyles(theme),
		input:   ti,
		spinner: spinner.New(),
		screen:  screenMenu,
	}
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

func TestMenuRendersAllScreens(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // isolate config writes
	m := newTestModel()

	if !strings.Contains(m.viewMenu(), "Search & Download") {
		t.Error("menu missing first item")
	}

	// Navigate to Settings (index 6) and open it.
	for i := 0; i < 6; i++ {
		next, _ := m.updateMenu(key("down"))
		m = next.(model)
	}
	next, _ := m.handleKey(key("enter"))
	m = next.(model)
	if m.screen != screenSettings {
		t.Fatalf("expected settings screen, got %v", m.screen)
	}
	if !strings.Contains(m.viewSettings(), "Audio format") {
		t.Error("settings view missing items")
	}

	// Toggle a boolean setting (use_genius at index 2).
	m.settingsCursor = 2
	before := m.cfg.UseGenius
	next, _ = m.activateSetting()
	m = next.(model)
	if m.cfg.UseGenius == before {
		t.Error("toggle did not flip use_genius")
	}

	// Open theme picker and apply a different theme.
	m.settingsCursor = 4
	next, _ = m.activateSetting()
	m = next.(model)
	if m.screen != screenThemePicker {
		t.Fatalf("expected theme picker, got %v", m.screen)
	}
	next, _ = m.updateThemePicker(key("down"))
	m = next.(model)
	if !strings.Contains(m.viewThemePicker(), "matrix") {
		t.Error("theme picker missing themes")
	}
}

func TestInputAndLibraryViews(t *testing.T) {
	m := newTestModel()
	next, _ := m.openInput("search", "Enter query")
	m = next.(model)
	if m.screen != screenInput || !strings.Contains(m.viewInput(), "Enter query") {
		t.Error("input screen not shown")
	}

	next, _ = m.openLibrary()
	m = next.(model)
	if m.screen != screenLibrary {
		t.Fatalf("expected library screen, got %v", m.screen)
	}
	if v := m.viewLibrary(); !strings.Contains(v, "Artists") {
		t.Errorf("library view missing breadcrumb: %q", v)
	}

	next, _ = m.openDownloadLog()
	m = next.(model)
	if m.screen != screenDownloadLog {
		t.Fatalf("expected download log screen, got %v", m.screen)
	}
	if !strings.Contains(m.viewDownloadLog(), "Download log") {
		t.Error("download log view missing title")
	}
}
