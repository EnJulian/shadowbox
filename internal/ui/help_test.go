package ui

import (
	"strings"
	"testing"
)

func TestHelpContentCoversEveryScreensShortcuts(t *testing.T) {
	full := strings.Join(helpContent, "\n")
	for _, want := range []string{
		"Ctrl+C", "playback",
		"Menu", "quit",
		"Library", "filter",
		"Settings",
		"Theme picker",
		"Download Log", "reload",
		"picker",
		"input", "submit",
		"Wizard",
		"config.yaml",
		"opus",
		"Artist",
	} {
		if !strings.Contains(full, want) {
			t.Errorf("helpContent missing expected reference to %q", want)
		}
	}
}

func TestQuestionMarkOpensHelpFromMenu(t *testing.T) {
	m := newTestModel()
	m.screen = screenMenu

	next, _ := m.handleKey(key("?"))
	m = next.(model)
	if m.screen != screenHelp {
		t.Fatalf("expected '?' to open help from Menu, got screen %v", m.screen)
	}
	if m.helpReturnTo != screenMenu {
		t.Fatalf("expected helpReturnTo = screenMenu, got %v", m.helpReturnTo)
	}
}

func TestQuestionMarkDoesNotOpenHelpOnLibraryScreen(t *testing.T) {
	m := newTestModel()
	m.screen = screenLibrary
	m.lib = libState{level: 0, entries: []string{"Nujabes"}}

	next, _ := m.handleKey(key("?"))
	m = next.(model)
	if m.screen != screenLibrary {
		t.Fatalf("'?' must not open help while Library's filter is active, got screen %v", m.screen)
	}
	if m.lib.filter != "?" {
		t.Errorf("filter = %q, want %q ('?' must reach the filter instead)", m.lib.filter, "?")
	}
}

func TestQuestionMarkDoesNotStealFromURLInput(t *testing.T) {
	m := newTestModel()
	next, _ := m.openInput("url", "Enter a URL")
	m = next.(model)

	valueBefore := m.input.Value()
	next, _ = m.handleKey(key("?"))
	m = next.(model)
	if m.screen == screenHelp {
		t.Fatal("'?' must not open help while the URL input is focused (URLs commonly contain '?')")
	}
	if m.input.Value() == valueBefore {
		t.Error("expected '?' to reach the text input")
	}
}

func TestQuestionMarkTogglesClosedFromHelpScreen(t *testing.T) {
	m := newTestModel()
	m.screen = screenSettings

	next, _ := m.handleKey(key("?"))
	m = next.(model)
	if m.screen != screenHelp {
		t.Fatalf("expected help to open, got %v", m.screen)
	}

	next, _ = m.handleKey(key("?"))
	m = next.(model)
	if m.screen != screenSettings {
		t.Fatalf("expected second '?' to close help back to Settings, got %v", m.screen)
	}
}

func TestEscAndQAlsoCloseHelp(t *testing.T) {
	for _, k := range []string{"esc", "q"} {
		m := newTestModel()
		m.screen = screenMenu
		next, _ := m.handleKey(key("?"))
		m = next.(model)
		if m.screen != screenHelp {
			t.Fatalf("[%s] expected help to open first, got %v", k, m.screen)
		}

		next, _ = m.updateHelp(key(k))
		m = next.(model)
		if m.screen != screenMenu {
			t.Errorf("[%s] expected help to close back to Menu, got %v", k, m.screen)
		}
	}
}

func TestHelpScrollClampsAtBothEnds(t *testing.T) {
	m := newTestModel()
	m.height = 20 // small viewport, guarantees content taller than it
	next, _ := m.openHelp(screenMenu)
	m = next.(model)

	next, _ = m.updateHelp(key("up")) // already at top
	m = next.(model)
	if m.helpScroll != 0 {
		t.Fatalf("scroll = %d, want 0 (can't go above top)", m.helpScroll)
	}

	for range 500 {
		next, _ = m.updateHelp(key("down"))
		m = next.(model)
	}
	if m.helpScroll != maxHelpScroll(helpViewport(m.height)) {
		t.Fatalf("scroll = %d, want clamped at %d", m.helpScroll, maxHelpScroll(helpViewport(m.height)))
	}
}

func TestQuestionMarkDoesNotOverrideResultScreensAnyKeyDismiss(t *testing.T) {
	m := newTestModel()
	m.screen = screenResult
	m.result = "Download complete"

	next, _ := m.handleKey(key("?"))
	m = next.(model)
	if m.screen != screenMenu {
		t.Fatalf("'?' on the result screen must still dismiss it like any other key, got screen %v", m.screen)
	}
}

func TestQuestionMarkDoesNotInterruptRunningScreen(t *testing.T) {
	m := newTestModel()
	m.screen = screenRunning

	next, _ := m.handleKey(key("?"))
	m = next.(model)
	if m.screen != screenRunning {
		t.Fatalf("'?' must not open help mid-task (running screen ignores all keys), got screen %v", m.screen)
	}
}

func TestFooterHintsMentionHelpKeyWhereItActuallyWorks(t *testing.T) {
	m := newTestModel()

	m.screen = screenMenu
	if !strings.Contains(m.viewMenu(), "?: help") {
		t.Error("expected the menu footer to mention ?: help")
	}

	next, _ := m.openWizard(screenMenu)
	m = next.(model)
	if !strings.Contains(m.viewWizard(), "?: help") {
		t.Error("expected the wizard footer to mention ?: help")
	}
}
