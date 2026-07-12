package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/ui/shell"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	"github.com/EnJulian/shadowbox/internal/ui/workspace"
)

func newTestModel(t *testing.T) *model {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := &config.Config{AudioFormat: "opus", MusicDirectory: t.TempDir(), Theme: "hacker", UseGenius: true}
	theme := style.ThemeByName(cfg.Theme)
	m := &model{
		cfg:     cfg,
		app:     app.New(cfg),
		theme:   theme,
		st:      style.NewStyles(theme),
		spinner: spinner.New(),
	}
	m.buildWorkspaces()
	return m
}

func key(s string) tea.KeyMsg {
	switch s {
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestNavDownThenEnterSwitchesToSettingsAndFocusesContent(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40

	for i := 0; i < 7; i++ { // Settings is index 7 in workspace.Order
		next, _ := m.handleKey(key("down"))
		m = next.(*model)
	}
	next, _ := m.handleKey(key("enter"))
	m = next.(*model)

	if m.activeSection != workspace.SectionSettings {
		t.Fatalf("activeSection = %v, want SectionSettings", m.activeSection)
	}
	view := m.View()
	if !strings.Contains(view, "Audio format") {
		t.Errorf("expected settings content in view, got %q", view)
	}
}

func TestDigitJumpsToSection(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40

	next, _ := m.handleKey(key("4")) // Library is section index 3, key "4"
	m = next.(*model)
	if m.activeSection != workspace.SectionLibrary {
		t.Fatalf("activeSection = %v, want SectionLibrary", m.activeSection)
	}
}

func TestTabTogglesFocusAndActivatesOnContentEntry(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40

	before := m.pane
	next, _ := m.handleKey(key("tab"))
	m = next.(*model)
	if m.pane == before {
		t.Fatal("expected Tab to toggle the focused pane")
	}
	view := m.View()
	if !strings.Contains(view, "Query:") { // Search is the default active section
		t.Errorf("expected Search workspace content, got %q", view)
	}
}

func TestHelpOverlayTogglesWithQuestionMark(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40

	next, _ := m.handleKey(key("?"))
	m = next.(*model)
	view := m.View()
	if !strings.Contains(view, "Keybindings") {
		t.Fatalf("expected help overlay content, got %q", view)
	}

	next, _ = m.handleKey(key("esc"))
	m = next.(*model)
	if m.ov != overlayNone {
		t.Fatal("expected esc to close the help overlay")
	}
}

func TestLibraryBackAtTopLevelReturnsFocusToNav(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40

	next, _ := m.handleKey(key("4")) // jump to Library and focus its content
	m = next.(*model)
	if m.pane != shell.PaneContent {
		t.Fatalf("pane = %v, want PaneContent after jumping into Library", m.pane)
	}

	next, cmd := m.handleKey(key("left")) // back out of the top (Artists) level
	m = next.(*model)
	if cmd == nil {
		t.Fatal("expected Library.back() at level 0 to return a FocusNavMsg cmd")
	}

	next, _ = m.Update(cmd())
	m = next.(*model)
	if m.pane != shell.PaneNav {
		t.Fatalf("pane = %v, want PaneNav after shell.FocusNavMsg", m.pane)
	}
}

func TestQuitKeyReturnsQuitCmd(t *testing.T) {
	m := newTestModel(t)
	_, cmd := m.handleKey(key("q"))
	if cmd == nil {
		t.Fatal("expected a quit cmd")
	}
}
