package ui

import (
	"context"
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

// TestTabWorksFromTextFocusedWorkspace confirms Tab always cycles panes, even
// when the active workspace's TextFocused() is true (e.g. Search's query
// input, focused by default). Tab is never a printable rune a user would
// type into a text field, so it must not be gated behind the TextFocused
// early-return the way q/?// /digits below are (see
// TestGlobalShortcutsDoNotStealKeystrokesFromFocusedText for that gating).
func TestTabWorksFromTextFocusedWorkspace(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40
	m.switchSection(workspace.SectionSearch) // focuses Content, Activate()s Search

	ws, ok := m.workspaces[workspace.SectionSearch].(workspace.TextFocused)
	if !ok || !ws.TextFocused() {
		t.Fatal("test setup invalid: Search is not TextFocused after switchSection")
	}
	if m.pane != shell.PaneContent {
		t.Fatalf("pane = %v, want PaneContent before pressing tab", m.pane)
	}

	next, _ := m.handleKey(key("tab"))
	m = next.(*model)
	if m.pane != shell.PaneNav {
		t.Fatalf("pane = %v, want PaneNav after tab from a TextFocused workspace", m.pane)
	}

	// Negative control: this fix must not regress the earlier critical fix
	// that q/digits are swallowed by focused text rather than stealing the
	// keystroke as global shortcuts. Broad coverage lives in
	// TestGlobalShortcutsDoNotStealKeystrokesFromFocusedText; this is a
	// quick spot-check scoped to this test's own setup.
	m.switchSection(workspace.SectionSearch)
	next, cmd := m.handleKey(key("q"))
	m = next.(*model)
	if cmd != nil {
		if _, isQuit := cmd().(tea.QuitMsg); isQuit {
			t.Fatal("q should be swallowed by the focused query input, not quit the app")
		}
	}
	view := m.workspaces[workspace.SectionSearch].View(80, 20)
	if !strings.Contains(view, "q") {
		t.Fatalf("expected %q to reach the query input, view = %q", "q", view)
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

// TestQuitWithDownloadRunningRequiresConfirmation proves that pressing q
// while a download is active (m.taskCancel != nil) does not quit outright:
// it arms a confirmation instead, per the design spec's Global Keys table
// ("q | Quit (confirms if a download is running)").
func TestQuitWithDownloadRunningRequiresConfirmation(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40
	_, cancel := context.WithCancel(context.Background())
	m.taskCancel = cancel
	defer cancel()

	next, cmd := m.handleKey(key("q"))
	m = next.(*model)

	if !m.confirmQuit {
		t.Fatal("expected confirmQuit to be set after q with a download running")
	}
	if cmd != nil {
		if _, isQuit := cmd().(tea.QuitMsg); isQuit {
			t.Fatal("expected the first q to NOT quit while a download is running")
		}
	}
	view := m.View()
	if !strings.Contains(view, "Press q again") {
		t.Errorf("expected status bar to show the confirm-quit prompt, got %q", view)
	}
}

// TestQuitConfirmedCancelsTaskAndQuits proves a second q after confirmQuit
// is armed actually quits, and cancels the running task's context so it
// doesn't leak.
func TestQuitConfirmedCancelsTaskAndQuits(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40
	ctx, cancel := context.WithCancel(context.Background())
	m.taskCancel = cancel

	next, _ := m.handleKey(key("q")) // arm confirmation
	m = next.(*model)
	if !m.confirmQuit {
		t.Fatal("expected confirmQuit to be armed after first q")
	}

	_, cmd := m.handleKey(key("q")) // confirm

	if cmd == nil {
		t.Fatal("expected a quit cmd after confirming q")
	}
	if _, isQuit := cmd().(tea.QuitMsg); !isQuit {
		t.Fatal("expected tea.QuitMsg after confirming q")
	}
	if ctx.Err() == nil {
		t.Fatal("expected the running task's context to be cancelled after confirming quit")
	}
}

// TestAnyOtherKeyClearsConfirmQuitWithoutQuittingOrActing proves the
// confirmation modally consumes the next keypress as a yes/no answer: any
// key other than q clears confirmQuit and does nothing else — in
// particular it must not also perform that key's normal global action
// (e.g. "?" must not open the help overlay).
func TestAnyOtherKeyClearsConfirmQuitWithoutQuittingOrActing(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 120, 40
	_, cancel := context.WithCancel(context.Background())
	m.taskCancel = cancel
	defer cancel()

	next, _ := m.handleKey(key("q")) // arm confirmation
	m = next.(*model)
	if !m.confirmQuit {
		t.Fatal("expected confirmQuit to be armed after first q")
	}

	next, cmd := m.handleKey(key("?"))
	m = next.(*model)

	if m.confirmQuit {
		t.Fatal("expected confirmQuit to be cleared after a non-q key")
	}
	if m.ov == overlayHelp {
		t.Fatal("expected the non-q key's normal action (opening help) to be suppressed")
	}
	if cmd != nil {
		if _, isQuit := cmd().(tea.QuitMsg); isQuit {
			t.Fatal("expected a non-q key to not quit")
		}
	}
}

// TestGlobalShortcutsDoNotStealKeystrokesFromFocusedText proves the fix for
// the reported bug: typing "q" (or any other single-char global shortcut)
// into a workspace that currently has a live text cursor must reach that
// workspace's text control, not trigger the global action (quit/help/jump).
//
// It reads back the typed text via View() rather than an internal field:
// in this test environment lipgloss renders with no ANSI escapes (no TTY),
// so a growing plain substring in the rendered view is reliable proof the
// keystroke reached the text control, and it only exercises exported API.
func TestGlobalShortcutsDoNotStealKeystrokesFromFocusedText(t *testing.T) {
	cases := []struct {
		name    string
		section workspace.Section
		setup   func(t *testing.T, m *model) // drives the workspace into a text-focused state
	}{
		{
			name:    "Search query input",
			section: workspace.SectionSearch,
			setup:   func(t *testing.T, m *model) {},
		},
		{
			name:    "Input workspace (URL)",
			section: workspace.SectionURL,
			setup:   func(t *testing.T, m *model) {},
		},
		{
			name:    "Settings inline edit",
			section: workspace.SectionSettings,
			setup: func(t *testing.T, m *model) {
				t.Helper()
				// audio_format (cursor 0) is a settingText item; enter opens
				// inline edit, putting the workspace into a text-focused state.
				next, _ := m.handleKey(key("enter"))
				*m = *next.(*model)
			},
		},
		{
			name:    "Library type-ahead filter",
			section: workspace.SectionLibrary,
			setup:   func(t *testing.T, m *model) {},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModel(t)
			m.width, m.height = 120, 40
			m.switchSection(tc.section) // focuses Content and Activate()s the workspace
			tc.setup(t, m)

			ws, ok := m.workspaces[tc.section].(workspace.TextFocused)
			if !ok || !ws.TextFocused() {
				t.Fatalf("test setup invalid: %s is not TextFocused after setup", tc.name)
			}

			var typed string
			for _, k := range []string{"q", "?", "/", "3"} {
				typed += k
				next, cmd := m.handleKey(key(k))
				m = next.(*model)

				view := m.workspaces[tc.section].View(80, 20)
				if !strings.Contains(view, typed) {
					t.Fatalf("after pressing %q: view = %q, want it to contain accumulated typed text %q (keystroke did not reach the text control)", k, view, typed)
				}
				if m.ov == overlayHelp {
					t.Fatalf("after pressing %q: help overlay opened, want it to stay closed", k)
				}
				if m.activeSection != tc.section {
					t.Fatalf("after pressing %q: activeSection = %v, want unchanged %v (digit-jump must not fire)", k, m.activeSection, tc.section)
				}
				if cmd != nil {
					if _, isQuit := cmd().(tea.QuitMsg); isQuit {
						t.Fatalf("after pressing %q: got tea.Quit, want the workspace to consume the key instead", k)
					}
				}
			}
		})
	}
}

// TestGlobalShortcutsStillWorkWhenNoTextIsFocused is the negative control for
// the above: workspaces with no text entry (Downloads, Log) must still let
// q/?// /digit-jump work as global shortcuts. This proves the fix is scoped
// to text-focused states, not a blanket "global keys stopped working."
// (Search's own arrow-key-navigation states — results/suggestions focused,
// not the query input — are covered by TestSearchTextFocused in the
// workspace package, proving the same scoping one level down.)
func TestGlobalShortcutsStillWorkWhenNoTextIsFocused(t *testing.T) {
	t.Run("Downloads: digit-jump", func(t *testing.T) {
		m := newTestModel(t)
		m.width, m.height = 120, 40
		m.switchSection(workspace.SectionDownloads)

		next, _ := m.handleKey(key("3")) // Playlist is section index 2, key "3"
		m = next.(*model)
		if m.activeSection != workspace.SectionPlaylist {
			t.Fatalf("activeSection = %v, want SectionPlaylist after digit-jump from Downloads", m.activeSection)
		}
	})

	t.Run("Log: help overlay", func(t *testing.T) {
		m := newTestModel(t)
		m.width, m.height = 120, 40
		m.switchSection(workspace.SectionLog)

		next, _ := m.handleKey(key("?"))
		m = next.(*model)
		if m.ov != overlayHelp {
			t.Fatalf("ov = %v, want overlayHelp after ? from Log", m.ov)
		}
	})

	t.Run("Log: q quits", func(t *testing.T) {
		m := newTestModel(t)
		m.width, m.height = 120, 40
		m.switchSection(workspace.SectionLog)

		_, cmd := m.handleKey(key("q"))
		if cmd == nil {
			t.Fatal("expected a quit cmd from Log")
		}
		if _, isQuit := cmd().(tea.QuitMsg); !isQuit {
			t.Fatal("expected tea.QuitMsg from q while Log is active")
		}
	})
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
