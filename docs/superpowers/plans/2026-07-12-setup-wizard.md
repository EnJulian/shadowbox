# Setup Wizard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a setup wizard that offers to install Shadowbox's two optional external tools (`mpv` for playback, `aria2` for faster downloads) — shown automatically the very first time Shadowbox ever runs, and reachable any time afterward from Settings.

**Architecture:** A new `internal/setup` package holds pure, testable tool-detection and install-command logic (which OS, which package manager, is it safe to run without sudo). A new Bubble Tea screen in `internal/ui` (`wizard.go`) drives the interactive list and runs installs via `tea.ExecProcess` where safe, or shows copy-paste instructions otherwise. `program.go` triggers it automatically on first run (no `config.yaml` present yet) and Settings gets a manual trigger.

**Tech Stack:** Go 1.25, existing `charmbracelet/bubbletea`, `internal/player.Available()`, `internal/download.HasAria2()`.

## Global Constraints

- Automated install only runs where it doesn't need elevated permissions: macOS via `brew`, Windows via `scoop` (preferred) or `winget`. Linux is always instructions-only — no sudo/password handling inside the TUI, ever.
- No new persisted config field for "has the wizard run" — first-run detection is "does `config.yaml` not exist yet," checked once via `config.Path()` + `os.Stat`. If that check errors for any reason other than "file does not exist," treat it as not-first-run (fail safe).
- The wizard's list navigation (mpv, aria2, Done) uses the same up/down/enter pattern as every other screen (Menu, Settings, Library) — no new keybinding invented for "finish." `Esc` behaves identically to selecting "Done."
- Every task must leave `make lint` and `make test` passing before its commit message is written. Per this project's rules, do NOT run `git commit` — write out the exact commit message and stop; the user commits it themselves.
- Tests must be hermetic: no test may depend on real `brew`/`scoop`/`winget`/`mpv`/`aria2c` being installed. Any test touching `internal/config` file I/O must sandbox via `t.Setenv("XDG_CONFIG_HOME", t.TempDir())` as its first statement(s).
- `aria2`'s binary name (`aria2c`, what `download.HasAria2()` checks on PATH) and its package-manager name (`aria2`, what `brew install`/`scoop install` expect) are different — don't conflate them.

---

## File Structure

```
internal/setup/
├── setup.go        # NEW: Tool, Tools, Available, InstallCommand, InstallResult
└── setup_test.go   # NEW

internal/ui/
├── wizard.go        # NEW: the setup wizard screen (state, update, view)
├── wizard_test.go   # NEW
├── program.go       # MODIFIED: screenSetupWizard, model fields, initialModel(), dispatch wiring
└── settings.go      # MODIFIED: settingAction kind, "Run setup wizard" entry
```

---

### Task 1: `internal/setup` — tool detection and install-command logic

**Files:**
- Create: `internal/setup/setup.go`
- Create: `internal/setup/setup_test.go`

**Interfaces:**
- Consumes: `player.Available() bool` (existing, `internal/player`), `download.HasAria2() bool` (existing, `internal/download`).
- Produces: `setup.Tool{Name, PackageName, Purpose string}`, `setup.Tools []Tool`, `setup.Available(name string) bool`, `setup.InstallResult{Automatable bool, Command []string, Instruction string}`, `setup.InstallCommand(tool Tool) InstallResult`.

This package is pure logic — no subprocess execution, no terminal interaction. `InstallCommand`'s real OS/PATH-lookup dependencies are injected through a small internal function (`installCommandFor`) so every platform branch is testable regardless of what's actually installed on the machine running the tests.

- [ ] **Step 1: Write the failing test**

```go
// internal/setup/setup_test.go
package setup

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func fakeLookPathFound(string) (string, error) { return "/usr/bin/x", nil }
func fakeLookPathNotFound(string) (string, error) {
	return "", errors.New("not found")
}

func TestInstallCommandDarwinWithBrew(t *testing.T) {
	res := installCommandFor("darwin", fakeLookPathFound, Tool{Name: "mpv", PackageName: "mpv"})
	if !res.Automatable {
		t.Fatal("expected automatable when brew is found")
	}
	want := []string{"brew", "install", "mpv"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v", res.Command, want)
	}
}

func TestInstallCommandDarwinWithoutBrew(t *testing.T) {
	res := installCommandFor("darwin", fakeLookPathNotFound, Tool{Name: "mpv", PackageName: "mpv"})
	if res.Automatable {
		t.Fatal("expected not automatable when brew is missing")
	}
	if !strings.Contains(res.Instruction, "brew.sh") {
		t.Errorf("Instruction = %q, want a pointer to installing Homebrew", res.Instruction)
	}
}

func TestInstallCommandWindowsPrefersScoopOverWinget(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "scoop" {
			return "/scoop", nil
		}
		return "", errors.New("not found")
	}
	res := installCommandFor("windows", lookup, Tool{Name: "aria2c", PackageName: "aria2"})
	if !res.Automatable {
		t.Fatal("expected automatable via scoop")
	}
	want := []string{"scoop", "install", "aria2"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v", res.Command, want)
	}
}

func TestInstallCommandWindowsFallsBackToWinget(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "winget" {
			return "/winget", nil
		}
		return "", errors.New("not found")
	}
	res := installCommandFor("windows", lookup, Tool{Name: "aria2c", PackageName: "aria2"})
	if !res.Automatable {
		t.Fatal("expected automatable via winget")
	}
	want := []string{"winget", "install", "aria2"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v", res.Command, want)
	}
}

func TestInstallCommandWindowsNoManagerFound(t *testing.T) {
	res := installCommandFor("windows", fakeLookPathNotFound, Tool{Name: "mpv", PackageName: "mpv"})
	if res.Automatable {
		t.Fatal("expected not automatable when neither scoop nor winget is found")
	}
}

func TestInstallCommandLinuxNeverAutomatable(t *testing.T) {
	res := installCommandFor("linux", fakeLookPathFound, Tool{Name: "mpv", PackageName: "mpv"})
	if res.Automatable {
		t.Fatal("expected Linux to never be automatable, even if a lookup would succeed")
	}
	if !strings.Contains(res.Instruction, "mpv") {
		t.Errorf("Instruction = %q, want it to mention the package name", res.Instruction)
	}
}

func TestAvailableReturnsFalseForUnrecognizedTool(t *testing.T) {
	if Available("does-not-exist") {
		t.Error("expected false for an unrecognized tool name")
	}
}

func TestToolsListsMpvAndAria2(t *testing.T) {
	names := map[string]bool{}
	for _, tool := range Tools {
		names[tool.Name] = true
	}
	if !names["mpv"] || !names["aria2c"] {
		t.Errorf("Tools = %+v, want it to include mpv and aria2c", Tools)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/setup/... -v`
Expected: FAIL — `package setup` doesn't exist yet.

- [ ] **Step 3: Write the implementation**

```go
// internal/setup/setup.go

// Package setup detects Shadowbox's optional external tools (mpv, aria2)
// and figures out how to install whichever ones are missing, for the
// first-run and Settings-triggered setup wizard.
package setup

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/EnJulian/shadowbox/internal/download"
	"github.com/EnJulian/shadowbox/internal/player"
)

// Tool describes an optional external dependency the setup wizard can help
// install. Name is the binary Shadowbox actually looks for on PATH;
// PackageName is what package managers call it — not always the same
// (aria2's binary is aria2c, but Homebrew/Scoop's package is "aria2").
type Tool struct {
	Name        string
	PackageName string
	Purpose     string
}

// Tools is the fixed list of optional dependencies the wizard offers.
var Tools = []Tool{
	{Name: "mpv", PackageName: "mpv", Purpose: "in-app audio playback"},
	{Name: "aria2c", PackageName: "aria2", Purpose: "accelerated multi-connection downloads"},
}

// Available reports whether the named tool (Tool.Name) is already usable.
func Available(name string) bool {
	switch name {
	case "mpv":
		return player.Available()
	case "aria2c":
		return download.HasAria2()
	}
	return false
}

// InstallResult describes how to get a tool installed on this machine.
type InstallResult struct {
	// Automatable is true if Command can be run directly, without elevated
	// permissions.
	Automatable bool
	// Command is the argv to execute; only meaningful when Automatable.
	Command []string
	// Instruction is a human-readable line to show either way.
	Instruction string
}

// InstallCommand returns the install approach for tool on this machine.
func InstallCommand(tool Tool) InstallResult {
	return installCommandFor(runtime.GOOS, exec.LookPath, tool)
}

// installCommandFor is InstallCommand's testable core: goos and lookPath
// are injected so every platform branch is exercisable without real
// package managers installed or being locked to the test runner's own OS.
func installCommandFor(goos string, lookPath func(string) (string, error), tool Tool) InstallResult {
	switch goos {
	case "darwin":
		cmd := []string{"brew", "install", tool.PackageName}
		if _, err := lookPath("brew"); err == nil {
			return InstallResult{Automatable: true, Command: cmd, Instruction: strings.Join(cmd, " ")}
		}
		return InstallResult{Instruction: "install Homebrew (https://brew.sh), then run: " + strings.Join(cmd, " ")}
	case "windows":
		if _, err := lookPath("scoop"); err == nil {
			cmd := []string{"scoop", "install", tool.PackageName}
			return InstallResult{Automatable: true, Command: cmd, Instruction: strings.Join(cmd, " ")}
		}
		if _, err := lookPath("winget"); err == nil {
			cmd := []string{"winget", "install", tool.PackageName}
			return InstallResult{Automatable: true, Command: cmd, Instruction: strings.Join(cmd, " ")}
		}
		return InstallResult{Instruction: "install via Scoop (https://scoop.sh) or winget — package name: " + tool.PackageName}
	default:
		return InstallResult{Instruction: "install `" + tool.PackageName + "` using your distribution's package manager, then ensure it's on your PATH"}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/setup/... -v`
Expected: PASS (8 tests)

- [ ] **Step 5: Verify and stop (do not commit)**

```bash
go vet ./internal/setup/...
make lint
```
Expected: both clean. Per project rules, do NOT run `git commit`. Write out this commit message and stop:

```
feat(setup): add tool detection and install-command logic

Pure logic for the setup wizard: which of mpv/aria2 are already
available, and for the current OS, whether an install can run
directly (macOS via brew, Windows via scoop/winget) or needs to be
shown as a manual command instead (always the case on Linux, since
virtually every native package manager there needs sudo).
```

---

### Task 2: `internal/ui` — the setup wizard screen

**Files:**
- Create: `internal/ui/wizard.go`
- Create: `internal/ui/wizard_test.go`
- Modify: `internal/ui/program.go`

**Interfaces:**
- Consumes: `setup.Tool`, `setup.Tools`, `setup.Available`, `setup.InstallResult`, `setup.InstallCommand` (Task 1); existing `model`, `screen`, `renderBannerWithPlayback`, `config.Save`, `config.Path` (all pre-existing in `internal/ui`/`internal/config`).
- Produces: `screenSetupWizard` (new `screen` value), `model.wizardReturnTo screen`, `model.wizardCursor int`, `model.wizardItems []wizardItem` (new fields), `wizardItem{tool setup.Tool, isDone, showHint bool, resultMsg string}`, `wizardInstallDoneMsg{index int, tool string, err error}`, `(model).openWizard(returnTo screen) (tea.Model, tea.Cmd)`, `(model).updateWizard(msg tea.KeyMsg) (tea.Model, tea.Cmd)`, `(model).viewWizard() string`, `(model).handleWizardInstallDone(msg wizardInstallDoneMsg) model`, `initialModel(cfg *config.Config, theme Theme, ti textinput.Model, sp spinner.Model) model`.

This task adds the screen and its model plumbing together, so it's a complete, independently buildable and testable unit on its own (reachable via first-run even before Task 3 adds the manual Settings trigger) — avoiding a half-compiling intermediate state.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/wizard_test.go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/... -run 'TestOpenWizard|TestWizard|TestHandleWizard|TestInitialModel' -v`
Expected: FAIL — `screenSetupWizard`, `model.wizardItems`, `openWizard`, `initialModel`, etc. undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/wizard.go
package ui

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/setup"
)

// wizardItem is one row in the setup wizard: either a tool to install, or
// the trailing "Done" entry.
type wizardItem struct {
	tool      setup.Tool
	isDone    bool
	showHint  bool
	resultMsg string
}

// wizardInstallDoneMsg reports the outcome of running an install command
// via tea.ExecProcess for the item at index.
type wizardInstallDoneMsg struct {
	index int
	tool  string
	err   error
}

// openWizard switches to the setup wizard screen, remembering returnTo so
// "Done" (or Esc) knows where to go back to.
func (m model) openWizard(returnTo screen) (tea.Model, tea.Cmd) {
	m.screen = screenSetupWizard
	m.wizardReturnTo = returnTo
	m.wizardCursor = 0
	items := make([]wizardItem, 0, len(setup.Tools)+1)
	for _, t := range setup.Tools {
		items = append(items, wizardItem{tool: t})
	}
	items = append(items, wizardItem{isDone: true})
	m.wizardItems = items
	return m, nil
}

func (m model) updateWizard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.wizardDone()
	case "up", "k":
		if m.wizardCursor > 0 {
			m.wizardCursor--
		}
	case "down", "j":
		if m.wizardCursor < len(m.wizardItems)-1 {
			m.wizardCursor++
		}
	case "enter":
		return m.wizardActivate()
	}
	return m, nil
}

func (m model) wizardActivate() (tea.Model, tea.Cmd) {
	item := m.wizardItems[m.wizardCursor]
	if item.isDone {
		return m.wizardDone()
	}
	if setup.Available(item.tool.Name) {
		return m, nil
	}
	result := setup.InstallCommand(item.tool)
	if !result.Automatable {
		items := make([]wizardItem, len(m.wizardItems))
		copy(items, m.wizardItems)
		items[m.wizardCursor].showHint = !items[m.wizardCursor].showHint
		m.wizardItems = items
		return m, nil
	}
	cmd := exec.Command(result.Command[0], result.Command[1:]...)
	index := m.wizardCursor
	toolName := item.tool.Name
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return wizardInstallDoneMsg{index: index, tool: toolName, err: err}
	})
}

func (m model) wizardDone() (tea.Model, tea.Cmd) {
	_ = config.Save(m.cfg)
	m.screen = m.wizardReturnTo
	return m, nil
}

// handleWizardInstallDone applies the result of a finished install attempt.
func (m model) handleWizardInstallDone(msg wizardInstallDoneMsg) model {
	items := make([]wizardItem, len(m.wizardItems))
	copy(items, m.wizardItems)
	switch {
	case msg.err != nil:
		items[msg.index].resultMsg = "install failed — see output above"
	case setup.Available(msg.tool):
		items[msg.index].resultMsg = "installed"
	default:
		items[msg.index].resultMsg = "install finished, but still not found on PATH"
	}
	m.wizardItems = items
	return m
}

func (m model) viewWizard() string {
	var b strings.Builder
	b.WriteString(renderBannerWithPlayback(m.st, m.theme, m.playback))
	b.WriteString("\n\n")
	b.WriteString("  " + m.st.subtitle.Render("Setup — optional tools") + "\n\n")

	for i, item := range m.wizardItems {
		label := "Done"
		if !item.isDone {
			status := "not installed"
			if setup.Available(item.tool.Name) {
				status = "installed"
			} else if item.resultMsg != "" {
				status = item.resultMsg
			}
			label = item.tool.Name + " — " + item.tool.Purpose + " (" + status + ")"
		}

		cursor := "  "
		line := m.st.item.Render(label)
		if i == m.wizardCursor {
			cursor = m.st.accent.Render("> ")
			line = m.st.selected.Render(label)
		}
		b.WriteString("  " + cursor + line + "\n")

		if !item.isDone && item.showHint {
			result := setup.InstallCommand(item.tool)
			b.WriteString("      " + m.st.help.Render(result.Instruction) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render("  up/down: navigate   enter: install/toggle   esc: done"))
	return b.String()
}
```

Now modify `internal/ui/program.go`:

1. Add `"os"` to the import block (alongside the existing `"context"`, `"io"`, `"time"`).

2. Add `screenSetupWizard` to the `screen` const block, after the existing last entry (`screenResult`):

```go
const (
	screenMenu screen = iota
	screenInput
	screenSettings
	screenSettingEdit
	screenThemePicker
	screenLibrary
	screenDownloadLog
	screenRunning
	screenPicker
	screenResult
	screenSetupWizard
)
```

3. Add fields to the `model` struct (alongside the existing `playback`/`player`/`playerErr` fields):

```go
	// setup wizard
	wizardReturnTo screen
	wizardCursor   int
	wizardItems    []wizardItem
```

4. Replace `runProgram`'s model-construction block. It currently reads:

```go
	m := model{
		cfg:     cfg,
		app:     app.New(cfg),
		theme:   theme,
		st:      newStyles(theme),
		input:   ti,
		spinner: sp,
		screen:  screenMenu,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
```

Replace with:

```go
	m := initialModel(cfg, theme, ti, sp)

	p := tea.NewProgram(m, tea.WithAltScreen())
```

And add this function (near `runProgram`, e.g. directly below it):

```go
// initialModel builds the model runProgram starts with, triggering the
// setup wizard automatically the very first time Shadowbox ever runs (no
// config file on disk yet). If the check itself errors for a reason other
// than "file does not exist," this fails safe and skips the auto-trigger —
// the wizard is still reachable manually from Settings either way.
func initialModel(cfg *config.Config, theme Theme, ti textinput.Model, sp spinner.Model) model {
	m := model{
		cfg:     cfg,
		app:     app.New(cfg),
		theme:   theme,
		st:      newStyles(theme),
		input:   ti,
		spinner: sp,
		screen:  screenMenu,
	}
	if path, err := config.Path(); err == nil {
		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			next, _ := m.openWizard(screenMenu)
			m = next.(model)
		}
	}
	return m
}
```

5. Add a case to `handleKey`'s screen switch:

```go
	case screenSetupWizard:
		return m.updateWizard(msg)
```

6. Add a case to `View`'s screen switch:

```go
	case screenSetupWizard:
		return m.viewWizard()
```

7. Add a case to `Update`'s message switch (alongside `taskDoneMsg`/`progressMsg`):

```go
	case wizardInstallDoneMsg:
		m = m.handleWizardInstallDone(msg)
		return m, nil
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/... -v`
Expected: PASS (all `internal/ui` tests, including the new wizard ones)

- [ ] **Step 5: Verify and stop (do not commit)**

```bash
go build ./...
go vet ./...
make lint
make test
```
Expected: all clean. Write out this commit message and stop:

```
feat(ui): add the setup wizard screen, triggered on first run

New screenSetupWizard lists mpv/aria2 with installed state; Enter
runs a real install via tea.ExecProcess where it's safe to do so
(no sudo needed), or toggles inline instructions otherwise. Enter
on "Done" (or Esc) saves config and returns to wherever the wizard
was opened from. runProgram now builds its initial model through
initialModel(), which auto-opens the wizard exactly once — the
very first time Shadowbox runs on a machine, detected by config.yaml
not existing yet — and fails safe (no auto-trigger) if that check
errors for any other reason. Not yet reachable from Settings (Task 3).
```

---

### Task 3: `internal/ui/settings.go` — manual "Run setup wizard" entry

**Files:**
- Modify: `internal/ui/settings.go`
- Modify: `internal/ui/ui_test.go` (new test appended)

**Interfaces:**
- Consumes: `screenSetupWizard`, `(model).openWizard` (Task 2).
- Produces: a new `settingAction` `settingKind` value; a new `settingItem{"setup_wizard", "Run setup wizard", settingAction}`, appended after the existing `genius.access_token` entry; `activateSetting()` handles `settingAction` by calling `m.openWizard(screenSettings)`.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/ui_test.go — append this test

func TestSettingsSetupWizardEntryOpensWizard(t *testing.T) {
	m := newTestModel()
	m.screen = screenSettings
	m.settingsCursor = len(settingItems) - 1 // the new entry, appended last

	next, _ := m.activateSetting()
	m = next.(model)
	if m.screen != screenSetupWizard {
		t.Fatalf("expected screenSetupWizard, got %v", m.screen)
	}
	if m.wizardReturnTo != screenSettings {
		t.Fatalf("expected wizardReturnTo = screenSettings, got %v", m.wizardReturnTo)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/... -run TestSettingsSetupWizardEntryOpensWizard -v`
Expected: FAIL — `m.settingsCursor = len(settingItems)-1` selects `genius.access_token` (a `settingSecret`), so `m.screen` ends up as `screenSettingEdit`, not `screenSetupWizard`.

- [ ] **Step 3: Write the implementation**

In `internal/ui/settings.go`, add `settingAction` to the `settingKind` const block:

```go
const (
	settingToggle settingKind = iota
	settingText
	settingSecret
	settingTheme
	settingAction
)
```

Append a new entry to `settingItems` (after `genius.access_token`):

```go
var settingItems = []settingItem{
	{"audio_format", "Audio format", settingText},
	{"music_directory", "Music directory", settingText},
	{"use_genius", "Use Genius lyrics", settingToggle},
	{"verbose", "Verbose logging", settingToggle},
	{"theme", "Theme", settingTheme},
	{"genius.access_token", "Genius access token", settingSecret},
	{"setup_wizard", "Run setup wizard", settingAction},
}
```

Add a case to `activateSetting()`'s switch:

```go
	case settingAction:
		return m.openWizard(screenSettings)
```

(`settingValue("setup_wizard")` needs no change — the function's existing fallthrough already returns `""` for any key it doesn't recognize, which is the correct empty value column for an action item.)

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/... -v`
Expected: PASS (all `internal/ui` tests, including the new one)

- [ ] **Step 5: Verify and stop (do not commit)**

```bash
go build ./...
go vet ./...
make lint
make test
```
Expected: all clean. Write out this commit message and stop:

```
feat(ui): add "Run setup wizard" entry to Settings

Lets a user re-open the setup wizard any time — to install
something they skipped on first run, or just re-check status —
via the same settingAction pattern the theme picker already uses.
```

---

## Testing Summary

| Layer | Tests |
|-------|-------|
| `internal/setup` | Install-command branching for every OS/package-manager combination (darwin+brew, darwin without brew, windows+scoop, windows+winget fallback, windows with neither, linux always non-automatable), `Available` for an unrecognized tool, `Tools` contents |
| `internal/ui` (wizard) | Item construction from `setup.Tools`, navigation bounds, Done/Esc save-and-return, already-installed no-op, not-installed automatable-vs-instructions branching, instruction text rendering, install-result message handling, first-run auto-trigger vs. existing-config-file skip |
| `internal/ui` (settings) | The new Settings entry opens the wizard with the right return screen |

All changes must pass `make lint` and `make test` (per Global Constraints).

## Manual Smoke Test (before shipping)

Since real subprocess installs aren't unit-tested (matching this project's existing convention — see `internal/download`), verify by hand before considering this done:
- Delete (or rename) `~/.config/shadowbox/config.yaml` temporarily, launch Shadowbox, confirm the wizard appears automatically; press Esc, confirm it returns to the menu and the config file now exists (so it won't auto-trigger again).
- From Settings, select "Run setup wizard," confirm it opens with the correct list and returns to Settings on Done/Esc.
- On this Linux machine, confirm both tools show the instructions-only path (Enter toggles inline text, nothing executes) — expected, since Linux is never automatable per this plan.
- Restore the real `config.yaml` (or its `music_directory`/`audio_format`/`theme` values) afterward if it was moved aside rather than just renamed back, per the config-safety lesson from earlier in this project.

## Not Covered By This Plan

Installing `yt-dlp`/`ffmpeg` (already required, out of scope), any sudo/elevated-permission automated install path, uninstalling or upgrading already-installed tools, and any "don't show this again" mechanism beyond the natural one-time first-run trigger.
