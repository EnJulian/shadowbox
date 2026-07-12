package ui

import (
	"errors"
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

func TestPickerRequestOpensOverlayAndEnterSendsResponse(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40

	respCh := make(chan promptResult, 1)
	out := promptOutgoing{
		req: app.PromptRequest{
			Title: "Select track",
			Options: []app.PromptOption{
				{Label: "Track A"},
				{Label: "Track B"},
				{Label: "Track C"},
			},
		},
		resp: respCh,
	}

	next, _ := m.Update(promptRequestMsg{out: out})
	m = next.(*model)

	if m.ov != overlayPicker {
		t.Fatalf("ov = %v, want overlayPicker", m.ov)
	}
	if len(m.picker.Options) != 3 {
		t.Fatalf("picker.Options = %d, want 3", len(m.picker.Options))
	}
	if m.pending == nil {
		t.Fatal("expected pending prompt to be set")
	}

	next, _ = m.handleKey(key("down"))
	m = next.(*model)
	if m.picker.Cursor != 1 {
		t.Fatalf("picker.Cursor = %d, want 1 after down", m.picker.Cursor)
	}

	next, _ = m.handleKey(key("enter"))
	m = next.(*model)

	if m.ov != overlayNone {
		t.Fatalf("ov = %v, want overlayNone after enter", m.ov)
	}
	if m.pending != nil {
		t.Fatal("expected pending to be cleared after enter")
	}

	select {
	case r := <-respCh:
		if r.idx != 1 {
			t.Errorf("resp.idx = %d, want 1", r.idx)
		}
		if r.err != nil {
			t.Errorf("resp.err = %v, want nil", r.err)
		}
	default:
		t.Fatal("expected a response on resp channel")
	}
}

func TestPickerEscCancelsAndSendsError(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40

	respCh := make(chan promptResult, 1)
	out := promptOutgoing{
		req: app.PromptRequest{
			Title:   "Select track",
			Options: []app.PromptOption{{Label: "Track A"}, {Label: "Track B"}},
		},
		resp: respCh,
	}

	next, _ := m.Update(promptRequestMsg{out: out})
	m = next.(*model)

	next, _ = m.handleKey(key("esc"))
	m = next.(*model)

	if m.ov != overlayNone {
		t.Fatalf("ov = %v, want overlayNone after esc", m.ov)
	}
	if m.pending != nil {
		t.Fatal("expected pending to be cleared after esc")
	}

	select {
	case r := <-respCh:
		if r.idx != -1 {
			t.Errorf("resp.idx = %d, want -1", r.idx)
		}
		if r.err != app.ErrSelectionCancelled {
			t.Errorf("resp.err = %v, want ErrSelectionCancelled", r.err)
		}
	default:
		t.Fatal("expected a cancellation response on resp channel")
	}
}

func TestThemeOverlayLivePreviewAndEnterSavesTheme(t *testing.T) {
	m := newTestModel(t) // sandboxes XDG_CONFIG_HOME so config.Save never touches real user data
	m.width, m.height = 120, 40

	next, _ := m.Update(workspace.SettingsThemeRequestMsg{})
	m = next.(*model)

	if m.ov != overlayTheme {
		t.Fatalf("ov = %v, want overlayTheme", m.ov)
	}
	if m.themeOv.Cursor != 0 { // cfg.Theme is "hacker", index 0 in style.ThemeOrder
		t.Fatalf("themeOv.Cursor = %d, want 0", m.themeOv.Cursor)
	}

	next, _ = m.handleKey(key("down"))
	m = next.(*model)

	if m.themeOv.Cursor != 1 {
		t.Fatalf("themeOv.Cursor = %d, want 1 after down", m.themeOv.Cursor)
	}
	wantPreview := style.ThemeOrder[1]
	if m.theme.Name != wantPreview {
		t.Fatalf("theme.Name = %q, want live preview %q", m.theme.Name, wantPreview)
	}
	if m.cfg.Theme != "hacker" {
		t.Fatalf("cfg.Theme = %q, want unchanged %q before enter", m.cfg.Theme, "hacker")
	}

	next, _ = m.handleKey(key("enter"))
	m = next.(*model)

	if m.ov != overlayNone {
		t.Fatalf("ov = %v, want overlayNone after enter", m.ov)
	}
	if m.cfg.Theme != wantPreview {
		t.Fatalf("cfg.Theme = %q, want %q after enter", m.cfg.Theme, wantPreview)
	}
	if m.theme.Name != wantPreview {
		t.Fatalf("theme.Name = %q, want %q after enter", m.theme.Name, wantPreview)
	}
}

func TestThemeOverlayEscRevertsWithoutSaving(t *testing.T) {
	m := newTestModel(t) // sandboxes XDG_CONFIG_HOME so config.Save never touches real user data
	m.width, m.height = 120, 40

	next, _ := m.Update(workspace.SettingsThemeRequestMsg{})
	m = next.(*model)

	next, _ = m.handleKey(key("down"))
	m = next.(*model)
	if m.theme.Name == "hacker" {
		t.Fatal("expected live preview to have moved off the original theme")
	}

	next, _ = m.handleKey(key("esc"))
	m = next.(*model)

	if m.ov != overlayNone {
		t.Fatalf("ov = %v, want overlayNone after esc", m.ov)
	}
	if m.cfg.Theme != "hacker" {
		t.Fatalf("cfg.Theme = %q, want unchanged %q", m.cfg.Theme, "hacker")
	}
	if m.theme.Name != "hacker" {
		t.Fatalf("theme.Name = %q, want reverted to %q", m.theme.Name, "hacker")
	}
}

func TestTaskDoneMsgUpdatesDownloadsOnSuccess(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40

	next, _ := m.Update(taskDoneMsg{summary: "Believer complete", err: nil})
	m = next.(*model)

	if m.ov != overlayResult {
		t.Fatalf("ov = %v, want overlayResult", m.ov)
	}
	view := m.downloads.View(80, 10)
	if !strings.Contains(view, "Believer complete") {
		t.Errorf("expected downloads view to contain summary, got %q", view)
	}
}

func TestTaskDoneMsgUpdatesDownloadsOnFailure(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40

	wantErr := errors.New("boom")
	next, _ := m.Update(taskDoneMsg{summary: "Believer failed", err: wantErr})
	m = next.(*model)

	if m.ov != overlayResult {
		t.Fatalf("ov = %v, want overlayResult", m.ov)
	}
	view := m.downloads.View(80, 10)
	if !strings.Contains(view, "Believer failed") {
		t.Errorf("expected downloads view to contain summary, got %q", view)
	}
	if !strings.Contains(view, "boom") {
		t.Errorf("expected downloads view to contain error text, got %q", view)
	}
}
