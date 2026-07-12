package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/setup"
)

func TestOpenWizardBuildsItemsFromSetupTools(t *testing.T) {
	m := newTestModel()
	next, _ := m.openWizard(screenMenu)
	m = next.(model)

	if m.screen != screenSetupWizard {
		t.Fatalf("expected screenSetupWizard, got %v", m.screen)
	}
	if m.wizardReturnTo != screenMenu {
		t.Fatalf("expected wizardReturnTo = screenMenu, got %v", m.wizardReturnTo)
	}
	if len(m.wizardItems) != len(setup.Tools)+1 {
		t.Fatalf("wizardItems = %d, want %d (tools + Done)", len(m.wizardItems), len(setup.Tools)+1)
	}
	if !m.wizardItems[len(m.wizardItems)-1].isDone {
		t.Error("expected the last item to be the Done entry")
	}
}

func TestWizardNavigationClampsAtBounds(t *testing.T) {
	m := newTestModel()
	next, _ := m.openWizard(screenMenu)
	m = next.(model)

	next, _ = m.updateWizard(key("up")) // already at top
	m = next.(model)
	if m.wizardCursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.wizardCursor)
	}

	for range 5 {
		next, _ = m.updateWizard(key("down"))
		m = next.(model)
	}
	if m.wizardCursor != len(m.wizardItems)-1 {
		t.Fatalf("cursor = %d, want %d (clamped at the last item)", m.wizardCursor, len(m.wizardItems)-1)
	}
}

func TestWizardEnterOnDoneSavesConfigAndReturns(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := newTestModel()
	next, _ := m.openWizard(screenSettings)
	m = next.(model)
	m.wizardCursor = len(m.wizardItems) - 1 // Done

	next, _ = m.updateWizard(key("enter"))
	m = next.(model)
	if m.screen != screenSettings {
		t.Fatalf("expected to return to screenSettings, got %v", m.screen)
	}
	path, _ := config.Path()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected config.Save to have created %s, got error: %v", path, err)
	}
}

func TestWizardEscBehavesLikeDone(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := newTestModel()
	next, _ := m.openWizard(screenMenu)
	m = next.(model)

	next, _ = m.updateWizard(key("esc"))
	m = next.(model)
	if m.screen != screenMenu {
		t.Fatalf("expected esc to return to screenMenu, got %v", m.screen)
	}
}

func TestWizardEnterOnAlreadyInstalledToolIsNoOp(t *testing.T) {
	m := newTestModel()
	next, _ := m.openWizard(screenMenu)
	m = next.(model)
	item := m.wizardItems[0]
	if !setup.Available(item.tool.Name) {
		t.Skip("tool not installed on this machine; nothing to verify for the already-installed path")
	}
	before := m.wizardItems[0]
	next, _ = m.updateWizard(key("enter"))
	m = next.(model)
	if m.wizardItems[0] != before {
		t.Error("expected no state change when activating an already-installed tool")
	}
}

func TestWizardEnterOnNotInstalledToolTakesAppropriateAction(t *testing.T) {
	m := newTestModel()
	next, _ := m.openWizard(screenMenu)
	m = next.(model)
	tool := m.wizardItems[0].tool
	if setup.Available(tool.Name) {
		t.Skip("tool is actually installed on this machine; the not-installed path isn't exercised here")
	}

	result := setup.InstallCommand(tool)
	next, cmd := m.updateWizard(key("enter"))
	m2 := next.(model)

	if result.Automatable {
		if cmd == nil {
			t.Fatal("expected a cmd (tea.ExecProcess) when the install is automatable")
		}
	} else {
		if !m2.wizardItems[0].showHint {
			t.Error("expected the instructions hint to toggle on when install isn't automatable")
		}
	}
}

func TestWizardViewShowsInstructionWhenHintToggled(t *testing.T) {
	m := newTestModel()
	next, _ := m.openWizard(screenMenu)
	m = next.(model)
	m.wizardItems[0].showHint = true

	view := m.viewWizard()
	result := setup.InstallCommand(m.wizardItems[0].tool)
	if !strings.Contains(view, result.Instruction) {
		t.Errorf("expected the view to show the instruction %q", result.Instruction)
	}
}

func TestHandleWizardInstallDoneSetsResultMessage(t *testing.T) {
	m := newTestModel()
	next, _ := m.openWizard(screenMenu)
	m = next.(model)

	m = m.handleWizardInstallDone(wizardInstallDoneMsg{index: 0, tool: "does-not-exist-tool", err: nil})
	if m.wizardItems[0].resultMsg == "" {
		t.Error("expected a result message to be set")
	}
}

func TestInitialModelTriggersWizardOnlyWhenConfigFileMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &config.Config{AudioFormat: "opus", MusicDirectory: ".", Theme: "hacker"}
	theme := themeByName(cfg.Theme)

	m := initialModel(cfg, theme, textinput.New(), spinner.New())
	if m.screen != screenSetupWizard {
		t.Fatalf("expected the wizard to auto-trigger with no config file present, got screen %v", m.screen)
	}

	// Simulate the file now existing, as if a previous run already saved it.
	shadowboxDir := filepath.Join(dir, "shadowbox")
	if err := os.MkdirAll(shadowboxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(shadowboxDir, "config.yaml"), []byte("theme: hacker\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	m2 := initialModel(cfg, theme, textinput.New(), spinner.New())
	if m2.screen != screenMenu {
		t.Fatalf("expected the menu when a config file already exists, got screen %v", m2.screen)
	}
}

func TestWizardHelpLineWarnsAboutPasswordWhenAutomatableAndNotInstalled(t *testing.T) {
	m := newTestModel()
	next, _ := m.openWizard(screenMenu)
	m = next.(model)

	current := m.wizardItems[m.wizardCursor].tool
	if setup.Available(current.Name) {
		t.Skip("tool is actually installed on this machine; the not-installed path isn't exercised here")
	}
	if !setup.InstallCommand(current).Automatable {
		t.Skip("no automated install path on this machine/OS for this tool; the automatable path isn't exercised here")
	}

	view := m.viewWizard()
	if !strings.Contains(view, "may ask for your password") {
		t.Errorf("expected the help line to warn about a password prompt, got:\n%s", view)
	}
}
