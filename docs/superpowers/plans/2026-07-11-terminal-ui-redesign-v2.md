# Terminal UI Redesign v2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace Shadowbox's single-column, full-screen-swapping Bubble Tea UI with a persistent, rounded-border two-pane shell (Nav | Content), matching the approved spec at `docs/superpowers/specs/2026-07-11-terminal-ui-redesign-v2-design.md`.

**Architecture:** Split `internal/ui` into leaf packages (`style`, `shell`, `workspace`, `suggest`, `overlay`, `playback`) that the root `internal/ui` package composes. Workspaces communicate side effects (start a task, switch section, cancel) to the root model via typed `tea.Msg` values returned from `tea.Cmd`, the standard Bubble Tea pattern — this avoids a Go import cycle between the root model and the workspace package.

**Tech Stack:** Go 1.25, `charmbracelet/bubbletea` v1.3.10, `charmbracelet/bubbles` (textinput, spinner), `charmbracelet/lipgloss` v1.1.0, `atotto/clipboard` v0.1.4 (already an indirect dependency — this plan promotes it to direct).

## Global Constraints

- Minimum supported terminal size is 100×24 (from spec §Layout).
- Border style is rounded corners everywhere (`lipgloss.RoundedBorder()`), no exceptions.
- Glyphs are plain Unicode only — no Nerd Font icons (spec §Theming).
- All autocomplete/suggestion data is local-only; no new network calls beyond the existing YouTube search (spec §Suggestions & Autocomplete).
- The playback bar region is structurally reserved but `PlaybarHeight` stays `0` this phase — no audio playback code ships in this plan (spec §Non-Goals).
- Every task must leave `make lint` and `make test` passing before its commit.
- Package boundaries: `style`, `shell`, `suggest`, `playback` must not import `internal/ui` (the root package) or `internal/ui/workspace`, to avoid import cycles — the root package imports them, never the reverse.

---

## File Structure

```
internal/app/
└── search.go                     # NEW: App.SearchTracks (thin wrapper for browsable search)

internal/ui/
├── ui.go                         # unchanged: Run(cfg) entry point
├── helpers.go                    # unchanged: rebuildApp
├── program.go                    # REWRITTEN: root model, message routing, key routing
├── ui_test.go                    # REWRITTEN: root model integration tests
├── style/
│   └── style.go                  # NEW: Theme, Styles, curated palette, banner
├── shell/
│   ├── focus.go                  # NEW: Pane, FocusNavMsg
│   ├── focus_test.go             # NEW
│   ├── shell.go                  # NEW: Layout, Compute, Render
│   └── shell_test.go             # NEW
├── workspace/
│   ├── workspace.go              # NEW: Workspace interface, Section, messages
│   ├── workspace_test.go         # NEW
│   ├── search.go                 # NEW
│   ├── search_test.go            # NEW
│   ├── input.go                  # NEW: shared URL/Playlist/Enhance input workspace
│   ├── input_test.go             # NEW
│   ├── library.go                # NEW
│   ├── library_test.go           # NEW
│   ├── downloads.go              # NEW
│   ├── downloads_test.go         # NEW
│   ├── log.go                    # NEW
│   ├── log_test.go               # NEW
│   ├── settings.go               # NEW
│   └── settings_test.go          # NEW
├── suggest/
│   ├── history.go                # NEW
│   ├── history_test.go           # NEW
│   ├── filter.go                 # NEW: Filter + LibraryMatches
│   ├── filter_test.go            # NEW
│   ├── clipboard.go              # NEW
│   └── clipboard_test.go         # NEW
├── overlay/
│   ├── picker.go                 # NEW
│   ├── picker_test.go            # NEW
│   ├── help.go                   # NEW
│   ├── result.go                 # NEW
│   └── theme.go                  # NEW
└── playback/
    └── playback.go                # NEW: State{Active bool} stub

# DELETED at the end of the plan (Task 16):
internal/ui/forms.go
internal/ui/library.go
internal/ui/log_view.go
internal/ui/menu.go
internal/ui/picker.go
internal/ui/settings.go
internal/ui/theme_picker.go
internal/ui/themes.go
internal/ui/progress.go          # renderProgressBar moved to workspace/downloads.go (Task 10)
internal/ui/progress_test.go     # its tests moved to workspace/downloads_test.go (Task 10)
```

---

### Task 1: `style` package — theme palette, styles, banner

**Files:**
- Create: `internal/ui/style/style.go`
- Create: `internal/ui/style/style_test.go`

**Interfaces:**
- Produces: `style.Theme{Name, Primary, Secondary, Accent, Muted, Danger lipgloss.Color}`, `style.Styles{Title, Subtitle, Selected, Item, Help, Accent, Danger, Box lipgloss.Style}`, `style.NewStyles(t Theme) Styles`, `style.ThemeByName(name string) Theme`, `style.ThemeOrder []string`, `style.Banner string`, `style.RenderBanner(st Styles) string`.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/style/style_test.go
package style

import "testing"

func TestThemeByNameFallsBackToHacker(t *testing.T) {
	got := ThemeByName("does-not-exist")
	if got.Name != "hacker" {
		t.Fatalf("ThemeByName fallback = %q, want %q", got.Name, "hacker")
	}
}

func TestThemeOrderMatchesRegisteredThemes(t *testing.T) {
	for _, name := range ThemeOrder {
		if ThemeByName(name).Name != name {
			t.Errorf("ThemeOrder entry %q not registered", name)
		}
	}
}

func TestRenderBannerContainsSubtitle(t *testing.T) {
	st := NewStyles(ThemeByName("hacker"))
	out := RenderBanner(st)
	if len(out) == 0 {
		t.Fatal("RenderBanner returned empty string")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/style/... -v`
Expected: FAIL — `package style is not in std` / no such file, since `style.go` does not exist yet.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/style/style.go

// Package style holds Shadowbox's terminal color themes, derived lipgloss
// styles, and the ASCII banner — the one leaf package every other ui
// sub-package (shell, workspace, overlay) depends on for rendering.
package style

import "github.com/charmbracelet/lipgloss"

// Theme defines the colour palette for the interface.
type Theme struct {
	Name      string
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Muted     lipgloss.Color
	Danger    lipgloss.Color
}

// themes is the curated v2 palette, retuned so border colours and active-nav
// highlights stay legible against the new rounded-border, box-heavy layout.
var themes = map[string]Theme{
	"hacker": {
		Name: "hacker", Primary: "10", Secondary: "2", Accent: "14", Muted: "245", Danger: "9",
	},
	"cyberpunk": {
		Name: "cyberpunk", Primary: "201", Secondary: "135", Accent: "51", Muted: "240", Danger: "197",
	},
	"ocean": {
		Name: "ocean", Primary: "39", Secondary: "33", Accent: "87", Muted: "244", Danger: "203",
	},
	"ember": {
		Name: "ember", Primary: "208", Secondary: "166", Accent: "220", Muted: "242", Danger: "160",
	},
	"matrix": {
		Name: "matrix", Primary: "46", Secondary: "34", Accent: "118", Muted: "238", Danger: "196",
	},
}

// ThemeOrder is the stable ordering used when cycling themes in the picker.
var ThemeOrder = []string{"hacker", "cyberpunk", "ocean", "ember", "matrix"}

// ThemeByName returns the named theme, or the "hacker" default if unknown.
func ThemeByName(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["hacker"]
}

// Styles bundles the lipgloss styles derived from a Theme.
type Styles struct {
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Selected lipgloss.Style
	Item     lipgloss.Style
	Help     lipgloss.Style
	Accent   lipgloss.Style
	Danger   lipgloss.Style
	Box      lipgloss.Style
}

// NewStyles derives the full style set from a Theme.
func NewStyles(t Theme) Styles {
	return Styles{
		Title:    lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Subtitle: lipgloss.NewStyle().Foreground(t.Secondary),
		Selected: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		Item:     lipgloss.NewStyle().Foreground(t.Muted),
		Help:     lipgloss.NewStyle().Foreground(t.Muted).Italic(true),
		Accent:   lipgloss.NewStyle().Foreground(t.Accent),
		Danger:   lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		Box:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1),
	}
}

// Banner is the ASCII logo shown at the top of the shell on every screen.
const Banner = `
 ███████╗██╗  ██╗ █████╗ ██████╗  ██████╗ ██╗    ██╗██████╗  ██████╗ ██╗  ██╗
 ██╔════╝██║  ██║██╔══██╗██╔══██╗██╔═══██╗██║    ██║██╔══██╗██╔═══██╗╚██╗██╔╝
 ███████╗███████║███████║██║  ██║██║   ██║██║ █╗ ██║██████╔╝██║   ██║ ╚███╔╝
 ╚════██║██╔══██║██╔══██║██║  ██║██║   ██║██║███╗██║██╔══██╗██║   ██║ ██╔██╗
 ███████║██║  ██║██║  ██║██████╔╝╚██████╔╝╚███╔███╔╝██████╔╝╚██████╔╝██╔╝ ██╗
 ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝  ╚══╝╚══╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝`

// RenderBanner renders the centered banner plus subtitle line, styled.
func RenderBanner(st Styles) string {
	return st.Title.Render(Banner) + "\n" + st.Subtitle.Render("                    Music Acquisition Console")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/style/... -v`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/style/style.go internal/ui/style/style_test.go
git commit -m "ui: add style package with curated theme palette"
```

---

### Task 2: `shell` package — focus model

**Files:**
- Create: `internal/ui/shell/focus.go`
- Create: `internal/ui/shell/focus_test.go`

**Interfaces:**
- Produces: `shell.Pane` (`PaneNav`, `PaneContent`), `(Pane).Toggle() Pane`, `shell.FocusNavMsg{}`, `shell.RequestNavFocus() tea.Cmd`.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/shell/focus_test.go
package shell

import "testing"

func TestPaneToggle(t *testing.T) {
	if PaneNav.Toggle() != PaneContent {
		t.Fatalf("PaneNav.Toggle() = %v, want PaneContent", PaneNav.Toggle())
	}
	if PaneContent.Toggle() != PaneNav {
		t.Fatalf("PaneContent.Toggle() = %v, want PaneNav", PaneContent.Toggle())
	}
}

func TestRequestNavFocusReturnsFocusNavMsg(t *testing.T) {
	cmd := RequestNavFocus()
	msg := cmd()
	if _, ok := msg.(FocusNavMsg); !ok {
		t.Fatalf("RequestNavFocus() cmd returned %T, want FocusNavMsg", msg)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/shell/... -v`
Expected: FAIL — package does not exist.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/shell/focus.go

// Package shell renders Shadowbox's persistent two-pane frame (Nav |
// Content) and owns which pane currently has keyboard focus.
package shell

import tea "github.com/charmbracelet/bubbletea"

// Pane identifies which region of the shell currently has keyboard focus.
type Pane int

const (
	PaneNav Pane = iota
	PaneContent
)

// Toggle returns the other pane. With exactly two panes, Tab and Shift+Tab
// both cycle to the same place, matching the spec's "always works, no dead
// ends" focus model.
func (p Pane) Toggle() Pane {
	if p == PaneNav {
		return PaneContent
	}
	return PaneNav
}

// FocusNavMsg is emitted by a workspace to ask the root model to return
// focus to Nav (e.g. pressing "left" from the leftmost Library column).
type FocusNavMsg struct{}

// RequestNavFocus returns a tea.Cmd that emits FocusNavMsg.
func RequestNavFocus() tea.Cmd {
	return func() tea.Msg { return FocusNavMsg{} }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/shell/... -v`
Expected: PASS (2 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/shell/focus.go internal/ui/shell/focus_test.go
git commit -m "ui: add shell focus model"
```

---

### Task 3: `shell` package — layout & rendering

**Files:**
- Create: `internal/ui/shell/shell.go`
- Create: `internal/ui/shell/shell_test.go`

**Interfaces:**
- Consumes: `style.Theme`, `style.Styles`, `style.RenderBanner` (Task 1).
- Produces: `shell.Layout{NavWidth, ContentWidth, ContentHeight, PlaybarHeight int}`, `shell.Compute(width, height int) Layout`, `shell.Render(st style.Styles, theme style.Theme, layout Layout, navFocused bool, navBody, contentBody, playbarBody, statusBody string) string`, `shell.NavWidth`, `shell.MinWidth`, `shell.MinHeight` constants.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/shell/shell_test.go
package shell

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestComputeSplitsWidthBetweenNavAndContent(t *testing.T) {
	l := Compute(120, 40)
	if l.NavWidth != NavWidth {
		t.Fatalf("NavWidth = %d, want %d", l.NavWidth, NavWidth)
	}
	if l.ContentWidth != 120-NavWidth {
		t.Fatalf("ContentWidth = %d, want %d", l.ContentWidth, 120-NavWidth)
	}
	if l.PlaybarHeight != 0 {
		t.Fatalf("PlaybarHeight = %d, want 0 (reserved but inactive this phase)", l.PlaybarHeight)
	}
}

func TestComputeClampsContentWidthOnNarrowTerminals(t *testing.T) {
	l := Compute(50, 40)
	if l.ContentWidth < 20 {
		t.Fatalf("ContentWidth = %d, want at least 20 even on a narrow terminal", l.ContentWidth)
	}
}

func TestComputeClampsContentHeightOnShortTerminals(t *testing.T) {
	l := Compute(120, 10)
	if l.ContentHeight < 5 {
		t.Fatalf("ContentHeight = %d, want at least 5 even on a short terminal", l.ContentHeight)
	}
}

func TestRenderIncludesBannerNavAndStatus(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	out := Render(st, style.ThemeByName("hacker"), Compute(120, 40), true, "NAV BODY", "CONTENT BODY", "", "STATUS LINE")
	for _, want := range []string{"NAV BODY", "CONTENT BODY", "STATUS LINE"} {
		if !strings.Contains(out, want) {
			t.Errorf("Render output missing %q", want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/shell/... -v`
Expected: FAIL — `Compute`, `Render`, `NavWidth` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/shell/shell.go
package shell

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

const (
	// NavWidth is the fixed outer width (including border) of the Nav pane.
	NavWidth = 20
	// MinWidth and MinHeight are the terminal dimensions the shell is
	// designed for; smaller terminals still render but content is clamped.
	MinWidth  = 100
	MinHeight = 24

	minContentWidth  = 20
	minContentHeight = 5
	// reservedRows accounts for the banner (2 lines), a blank line, and the
	// unboxed status bar (1 line) that always surround the pane row.
	reservedRows = 4
)

// Layout describes the computed pixel dimensions of each shell region for a
// given terminal size.
type Layout struct {
	NavWidth      int
	ContentWidth  int
	ContentHeight int
	// PlaybarHeight is 0 while playback is inactive (always, this phase).
	// The field exists so a later phase can grow it without touching the
	// rest of the layout math.
	PlaybarHeight int
}

// Compute derives a Layout from the terminal size reported by tea.WindowSizeMsg.
func Compute(width, height int) Layout {
	contentWidth := width - NavWidth
	if contentWidth < minContentWidth {
		contentWidth = minContentWidth
	}
	contentHeight := height - reservedRows
	if contentHeight < minContentHeight {
		contentHeight = minContentHeight
	}
	return Layout{
		NavWidth:      NavWidth,
		ContentWidth:  contentWidth,
		ContentHeight: contentHeight,
		PlaybarHeight: 0,
	}
}

func borderColor(theme style.Theme, focused bool) lipgloss.Color {
	if focused {
		return theme.Accent
	}
	return theme.Muted
}

// Render composes the full frame: banner, Nav | Content row (each in a
// rounded-border box, the focused pane's border in the theme's accent
// colour), the reserved playback bar row (only when playbarBody is
// non-empty and layout.PlaybarHeight > 0), and the unboxed status line.
func Render(st style.Styles, theme style.Theme, layout Layout, navFocused bool, navBody, contentBody, playbarBody, statusBody string) string {
	navStyle := st.Box.Width(layout.NavWidth - 2).Height(layout.ContentHeight - 2).
		BorderForeground(borderColor(theme, navFocused))
	contentStyle := st.Box.Width(layout.ContentWidth - 2).Height(layout.ContentHeight - 2).
		BorderForeground(borderColor(theme, !navFocused))

	row := lipgloss.JoinHorizontal(lipgloss.Top, navStyle.Render(navBody), contentStyle.Render(contentBody))

	parts := []string{style.RenderBanner(st), row}
	if layout.PlaybarHeight > 0 && playbarBody != "" {
		parts = append(parts, playbarBody)
	}
	parts = append(parts, statusBody)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/shell/... -v`
Expected: PASS (6 tests total in the package)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/shell/shell.go internal/ui/shell/shell_test.go
git commit -m "ui: add shell layout and rendering"
```

---

### Task 4: `workspace` package — interface, sections, messages

**Files:**
- Create: `internal/ui/workspace/workspace.go`
- Create: `internal/ui/workspace/workspace_test.go`

**Interfaces:**
- Consumes: `app.Options` (existing, `internal/app`).
- Produces: `workspace.Section` (`SectionSearch` … `SectionSettings`, 8 values), `workspace.Meta{Section, Label}`, `workspace.Order []Meta`, `workspace.Workspace` interface, `workspace.StartTaskMsg{Summary string, Run func(context.Context, app.Options) error}`, `workspace.SwitchSectionMsg{Section Section}`, `workspace.CancelTaskMsg{}`, `workspace.StartTask(summary string, run func(context.Context, app.Options) error) tea.Cmd`, `workspace.SwitchSection(s Section) tea.Cmd`, `workspace.CancelTask() tea.Cmd`.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/workspace/workspace_test.go
package workspace

import (
	"context"
	"testing"

	"github.com/EnJulian/shadowbox/internal/app"
)

func TestOrderCoversEverySection(t *testing.T) {
	seen := map[Section]bool{}
	for _, m := range Order {
		seen[m.Section] = true
		if m.Label == "" {
			t.Errorf("section %v has empty label", m.Section)
		}
	}
	want := []Section{SectionSearch, SectionURL, SectionPlaylist, SectionLibrary, SectionDownloads, SectionEnhance, SectionLog, SectionSettings}
	if len(seen) != len(want) {
		t.Fatalf("Order has %d unique sections, want %d", len(seen), len(want))
	}
	for _, s := range want {
		if !seen[s] {
			t.Errorf("Order missing section %v", s)
		}
	}
}

func TestStartTaskCmdCarriesSummary(t *testing.T) {
	called := false
	cmd := StartTask("Download", func(ctx context.Context, opts app.Options) error {
		called = true
		return nil
	})
	msg := cmd()
	stMsg, ok := msg.(StartTaskMsg)
	if !ok {
		t.Fatalf("StartTask() cmd returned %T, want StartTaskMsg", msg)
	}
	if stMsg.Summary != "Download" {
		t.Fatalf("Summary = %q, want %q", stMsg.Summary, "Download")
	}
	_ = stMsg.Run(context.Background(), app.Options{})
	if !called {
		t.Fatal("StartTaskMsg.Run was not the closure passed to StartTask")
	}
}

func TestSwitchSectionCmd(t *testing.T) {
	msg := SwitchSection(SectionDownloads)()
	ss, ok := msg.(SwitchSectionMsg)
	if !ok || ss.Section != SectionDownloads {
		t.Fatalf("SwitchSection cmd = %#v, want SwitchSectionMsg{SectionDownloads}", msg)
	}
}

func TestCancelTaskCmd(t *testing.T) {
	if _, ok := CancelTask()().(CancelTaskMsg); !ok {
		t.Fatal("CancelTask() cmd did not return CancelTaskMsg")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/workspace/... -v`
Expected: FAIL — package does not exist.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/workspace/workspace.go

// Package workspace implements the content pane's per-section behavior
// (Search, URL, Playlist, Library, Downloads, Enhance, Log, Settings). Each
// Workspace reports side effects (starting a task, switching section,
// cancelling) to the root model via tea.Cmd-returned messages rather than
// calling back directly, so this package never imports the root ui package.
package workspace

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
)

// Section identifies one nav sidebar entry / content pane.
type Section int

const (
	SectionSearch Section = iota
	SectionURL
	SectionPlaylist
	SectionLibrary
	SectionDownloads
	SectionEnhance
	SectionLog
	SectionSettings
)

// Meta describes one nav sidebar entry.
type Meta struct {
	Section Section
	Label   string
}

// Order is the fixed nav sidebar ordering, matching the spec's Nav Sections table.
var Order = []Meta{
	{SectionSearch, "Search"},
	{SectionURL, "URL"},
	{SectionPlaylist, "Playlist"},
	{SectionLibrary, "Library"},
	{SectionDownloads, "Downloads"},
	{SectionEnhance, "Enhance"},
	{SectionLog, "Log"},
	{SectionSettings, "Settings"},
}

// Workspace is a single nav section's content-pane behavior.
type Workspace interface {
	// Activate resets internal focus/cursor state (e.g. focuses the query
	// input). Called whenever this section becomes active, and whenever the
	// Content pane regains focus while this section is already active.
	Activate() Workspace
	Update(msg tea.Msg) (Workspace, tea.Cmd)
	View(width, height int) string
}

// StartTaskMsg asks the root model to run a background pipeline operation,
// identical in shape to the pre-redesign model.startTask contract.
type StartTaskMsg struct {
	Summary string
	Run     func(ctx context.Context, opts app.Options) error
}

// SwitchSectionMsg asks the root model to change the active nav section and
// focus Content on it.
type SwitchSectionMsg struct {
	Section Section
}

// CancelTaskMsg asks the root model to cancel the running task.
type CancelTaskMsg struct{}

// StartTask returns a tea.Cmd that emits StartTaskMsg.
func StartTask(summary string, run func(ctx context.Context, opts app.Options) error) tea.Cmd {
	return func() tea.Msg { return StartTaskMsg{Summary: summary, Run: run} }
}

// SwitchSection returns a tea.Cmd that emits SwitchSectionMsg.
func SwitchSection(s Section) tea.Cmd {
	return func() tea.Msg { return SwitchSectionMsg{Section: s} }
}

// CancelTask returns a tea.Cmd that emits CancelTaskMsg.
func CancelTask() tea.Cmd {
	return func() tea.Msg { return CancelTaskMsg{} }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/workspace/... -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/workspace/workspace.go internal/ui/workspace/workspace_test.go
git commit -m "ui: add workspace interface, section registry, and root messages"
```

---

### Task 5: `suggest` package — history, filtering, clipboard detection

**Files:**
- Create: `internal/ui/suggest/history.go`
- Create: `internal/ui/suggest/history_test.go`
- Create: `internal/ui/suggest/filter.go`
- Create: `internal/ui/suggest/filter_test.go`
- Create: `internal/ui/suggest/clipboard.go`
- Create: `internal/ui/suggest/clipboard_test.go`

**Interfaces:**
- Produces: `suggest.History`, `suggest.LoadHistory(path string) (*History, error)`, `(*History).Add(query string)`, `(*History).Save(path string) error`, `(*History).Matches(prefix string, limit int) []string`, `suggest.Filter(entries []string, query string) []string`, `suggest.LibraryMatches(musicDir, query string, limit int) []string`, `suggest.ClipboardURL() string`, `suggest.LooksLikeURL(s string) bool`.

- [ ] **Step 1: Write the failing tests**

```go
// internal/ui/suggest/history_test.go
package suggest

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestHistoryAddDedupesAndPrepends(t *testing.T) {
	h := &History{}
	h.Add("nujabes feather")
	h.Add("kanye")
	h.Add("nujabes feather") // re-adding moves it to the front, no duplicate

	got := h.Matches("", 10)
	want := []string{"nujabes feather", "kanye"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Matches = %v, want %v", got, want)
	}
}

func TestHistoryMatchesFiltersByPrefixCaseInsensitive(t *testing.T) {
	h := &History{}
	h.Add("Nujabes Feather")
	h.Add("Kanye Stronger")

	got := h.Matches("nu", 10)
	if len(got) != 1 || got[0] != "Nujabes Feather" {
		t.Fatalf("Matches(\"nu\") = %v, want [Nujabes Feather]", got)
	}
}

func TestHistorySaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "search_history")

	h := &History{}
	h.Add("first query")
	h.Add("second query")
	if err := h.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}
	got := loaded.Matches("", 10)
	want := []string{"second query", "first query"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded Matches = %v, want %v", got, want)
	}
}

func TestLoadHistoryMissingFileReturnsEmpty(t *testing.T) {
	h, err := LoadHistory(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("LoadHistory() error = %v, want nil for a missing file", err)
	}
	if len(h.Matches("", 10)) != 0 {
		t.Fatalf("expected empty history, got %v", h.Matches("", 10))
	}
}
```

```go
// internal/ui/suggest/filter_test.go
package suggest

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFilterCaseInsensitiveSubstring(t *testing.T) {
	entries := []string{"Nujabes", "Kanye West", "Aphex Twin"}
	got := Filter(entries, "we")
	want := []string{"Kanye West"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Filter = %v, want %v", got, want)
	}
}

func TestFilterEmptyQueryReturnsAll(t *testing.T) {
	entries := []string{"a", "b"}
	if got := Filter(entries, ""); !reflect.DeepEqual(got, entries) {
		t.Fatalf("Filter with empty query = %v, want %v", got, entries)
	}
}

func TestLibraryMatchesFindsArtistAlbumAndTrack(t *testing.T) {
	dir := t.TempDir()
	trackDir := filepath.Join(dir, "Nujabes", "Modal Soul")
	if err := os.MkdirAll(trackDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(trackDir, "Feather.mp3"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	got := LibraryMatches(dir, "feather", 10)
	if len(got) != 1 || got[0] != "Feather.mp3" {
		t.Fatalf("LibraryMatches = %v, want [Feather.mp3]", got)
	}

	got = LibraryMatches(dir, "nujabes", 10)
	if len(got) != 1 || got[0] != "Nujabes" {
		t.Fatalf("LibraryMatches = %v, want [Nujabes]", got)
	}
}
```

```go
// internal/ui/suggest/clipboard_test.go
package suggest

import "testing"

func TestLooksLikeURL(t *testing.T) {
	cases := map[string]bool{
		"https://youtu.be/abc123":        true,
		"http://example.com/track":       true,
		"just some search text":          false,
		"":                               false,
		"  https://youtu.be/abc123  ":    true,
	}
	for in, want := range cases {
		if got := LooksLikeURL(in); got != want {
			t.Errorf("LooksLikeURL(%q) = %v, want %v", in, got, want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/suggest/... -v`
Expected: FAIL — package does not exist.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/suggest/history.go

// Package suggest provides local-only autocomplete data sources: recent
// search history, library fuzzy matching, and clipboard URL detection.
package suggest

import (
	"os"
	"strings"
)

const maxHistoryEntries = 20

// History stores recent search queries, most-recent-first.
type History struct {
	entries []string
}

// LoadHistory reads a newline-delimited history file. A missing file is not
// an error — it just means no history yet.
func LoadHistory(path string) (*History, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &History{}, nil
		}
		return nil, err
	}
	h := &History{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}
	return h, nil
}

// Save writes the history to path, most-recent-first, one query per line.
func (h *History) Save(path string) error {
	return os.WriteFile(path, []byte(strings.Join(h.entries, "\n")), 0o600)
}

// Add records a query as most recent, de-duplicating and capping the list.
func (h *History) Add(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}
	filtered := h.entries[:0]
	for _, e := range h.entries {
		if !strings.EqualFold(e, query) {
			filtered = append(filtered, e)
		}
	}
	h.entries = append([]string{query}, filtered...)
	if len(h.entries) > maxHistoryEntries {
		h.entries = h.entries[:maxHistoryEntries]
	}
}

// Matches returns up to limit history entries whose prefix case-insensitively
// matches prefix, most-recent-first. An empty prefix matches everything.
func (h *History) Matches(prefix string, limit int) []string {
	var out []string
	lowerPrefix := strings.ToLower(prefix)
	for _, e := range h.entries {
		if strings.HasPrefix(strings.ToLower(e), lowerPrefix) {
			out = append(out, e)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}
```

```go
// internal/ui/suggest/filter.go
package suggest

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Filter returns the entries whose text case-insensitively contains query.
// An empty query returns entries unchanged.
func Filter(entries []string, query string) []string {
	if query == "" {
		return entries
	}
	lowerQuery := strings.ToLower(query)
	var out []string
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e), lowerQuery) {
			out = append(out, e)
		}
	}
	return out
}

var audioExts = map[string]bool{".opus": true, ".mp3": true, ".m4a": true, ".flac": true, ".wav": true, ".ogg": true}

// LibraryMatches scans musicDir (Artist/Album/Track, one level of each) and
// returns up to limit artist, album, and track names whose text
// case-insensitively contains query.
func LibraryMatches(musicDir, query string, limit int) []string {
	var out []string
	appendMatch := func(name string) bool {
		if strings.Contains(strings.ToLower(name), strings.ToLower(query)) {
			out = append(out, name)
		}
		return len(out) >= limit
	}

	artists, err := os.ReadDir(musicDir)
	if err != nil {
		return nil
	}
	sort.Slice(artists, func(i, j int) bool { return artists[i].Name() < artists[j].Name() })

	for _, artist := range artists {
		if !artist.IsDir() {
			continue
		}
		if appendMatch(artist.Name()) {
			return out
		}
		artistDir := filepath.Join(musicDir, artist.Name())
		albums, err := os.ReadDir(artistDir)
		if err != nil {
			continue
		}
		for _, album := range albums {
			if album.IsDir() {
				if appendMatch(album.Name()) {
					return out
				}
				albumDir := filepath.Join(artistDir, album.Name())
				tracks, err := os.ReadDir(albumDir)
				if err != nil {
					continue
				}
				for _, track := range tracks {
					if !track.IsDir() && audioExts[strings.ToLower(filepath.Ext(track.Name()))] {
						if appendMatch(track.Name()) {
							return out
						}
					}
				}
			}
		}
	}
	return out
}
```

```go
// internal/ui/suggest/clipboard.go
package suggest

import (
	"strings"

	"github.com/atotto/clipboard"
)

// LooksLikeURL reports whether s (after trimming whitespace) starts with
// http:// or https://.
func LooksLikeURL(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// ClipboardURL returns the current clipboard contents if they look like a
// URL, or "" if the clipboard is empty, unreadable, or not URL-shaped.
func ClipboardURL() string {
	text, err := clipboard.ReadAll()
	if err != nil {
		return ""
	}
	text = strings.TrimSpace(text)
	if LooksLikeURL(text) {
		return text
	}
	return ""
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/suggest/... -v`
Expected: PASS (all tests; `TestLibraryMatchesFindsArtistAlbumAndTrack` covers both directory and file matches)

- [ ] **Step 5: Promote the clipboard dependency and commit**

```bash
go mod tidy
git add internal/ui/suggest/ go.mod go.sum
git commit -m "ui: add suggest package for search history, filtering, and clipboard detection"
```

---

### Task 6: `App.SearchTracks` — browsable search without the interactive picker

**Files:**
- Create: `internal/app/search.go`

**Interfaces:**
- Consumes: `a.newDownloader(Options) *download.Downloader` (existing, private, same package), `(*download.Downloader).SearchYouTube(ctx, query, limit) ([]download.SearchResult, error)` (existing).
- Produces: `(a *App).SearchTracks(ctx context.Context, query string, limit int) ([]download.SearchResult, error)`.

This is a thin delegation with no branching logic of its own — `SearchYouTube` already has its own tests in `internal/download/search_test.go`, so no new test is added here, matching this repo's existing convention of not re-testing pure pass-through wrappers (see e.g. `app.New`).

- [ ] **Step 1: Write the implementation**

```go
// internal/app/search.go
package app

import (
	"context"

	"github.com/EnJulian/shadowbox/internal/download"
)

// SearchTracks runs a YouTube search and returns the raw results for the
// caller to browse, without downloading or prompting. Used by the Search
// workspace to show results before the user picks one.
func (a *App) SearchTracks(ctx context.Context, query string, limit int) ([]download.SearchResult, error) {
	dl := a.newDownloader(Options{})
	return dl.SearchYouTube(ctx, query, limit)
}
```

- [ ] **Step 2: Verify it builds**

Run: `go build ./internal/app/...`
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/app/search.go
git commit -m "app: add SearchTracks for browsable search results"
```

---

### Task 7: `workspace.Search`

**Files:**
- Create: `internal/ui/workspace/search.go`
- Create: `internal/ui/workspace/search_test.go`

**Interfaces:**
- Consumes: `app.SearchTracks` (Task 6), `suggest.History`, `suggest.LibraryMatches`, `suggest.Filter` (Task 5), `style.Styles` (Task 1), `workspace.StartTask`, `workspace.SwitchSection` (Task 4).
- Produces: `workspace.NewSearch(a *app.App, cfg *config.Config, st style.Styles, historyPath string) *Search`, implementing `workspace.Workspace`.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/workspace/search_test.go
package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/download"
	"github.com/EnJulian/shadowbox/internal/ui/style"
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/workspace/... -run TestSearch -v`
Expected: FAIL — `Search`, `NewSearch`, `searchResultsMsg` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/workspace/search.go
package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/download"
	"github.com/EnJulian/shadowbox/internal/ui/shell"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	"github.com/EnJulian/shadowbox/internal/ui/suggest"
)

type searchFocus int

const (
	searchFocusInput searchFocus = iota
	searchFocusResults
)

// searchResultsMsg carries browsable results back from an async SearchTracks call.
type searchResultsMsg struct {
	results []download.SearchResult
	err     error
}

// Search is the workspace.Workspace for the Search nav section: a query
// input with autocomplete, then a browsable results list.
type Search struct {
	app         *app.App
	cfg         *config.Config
	st          style.Styles
	historyPath string
	history     *suggest.History

	input   textinput.Model
	focus   searchFocus
	results []download.SearchResult
	cursor  int
	loading bool
	errMsg  string
}

// NewSearch builds the Search workspace.
func NewSearch(a *app.App, cfg *config.Config, st style.Styles, historyPath string) *Search {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Width = 60
	ti.Placeholder = "Enter title by artist (e.g. High Speed Chasing by BØRNS)"

	history, _ := suggest.LoadHistory(historyPath)

	return &Search{app: a, cfg: cfg, st: st, historyPath: historyPath, history: history, input: ti}
}

func (s *Search) Activate() Workspace {
	s.focus = searchFocusInput
	s.input.Focus()
	return s
}

func (s *Search) suggestions() []string {
	q := strings.TrimSpace(s.input.Value())
	if q == "" {
		return nil
	}
	out := s.history.Matches(q, 5)
	for _, m := range suggest.LibraryMatches(s.cfg.MusicDirectory, q, 3) {
		out = append(out, m+" (in library)")
	}
	return out
}

func (s *Search) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	switch msg := msg.(type) {
	case searchResultsMsg:
		s.loading = false
		s.results = msg.results
		s.cursor = 0
		if msg.err != nil {
			s.errMsg = msg.err.Error()
		} else {
			s.errMsg = ""
		}
		return s, nil

	case tea.KeyMsg:
		return s.handleKey(msg)
	}
	return s, nil
}

func (s *Search) handleKey(msg tea.KeyMsg) (Workspace, tea.Cmd) {
	if s.focus == searchFocusInput {
		switch msg.String() {
		case "enter":
			query := strings.TrimSpace(s.input.Value())
			if query == "" {
				return s, nil
			}
			s.history.Add(query)
			_ = s.history.Save(s.historyPath)
			s.loading = true
			a := s.app
			return s, func() tea.Msg {
				results, err := a.SearchTracks(context.Background(), query, 10)
				return searchResultsMsg{results: results, err: err}
			}
		case "down":
			if len(s.results) > 0 {
				s.focus = searchFocusResults
				s.cursor = 0
				s.input.Blur()
			}
			return s, nil
		}
		var cmd tea.Cmd
		s.input, cmd = s.input.Update(msg)
		return s, cmd
	}

	// focus == searchFocusResults
	switch msg.String() {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		} else {
			s.focus = searchFocusInput
			s.input.Focus()
		}
	case "down", "j":
		if s.cursor < len(s.results)-1 {
			s.cursor++
		}
	case "left", "h":
		s.focus = searchFocusInput
		s.input.Focus()
		return s, shell.RequestNavFocus()
	case "enter":
		if len(s.results) == 0 {
			return s, nil
		}
		selected := s.results[s.cursor]
		return s, tea.Batch(
			StartTask("Download", func(ctx context.Context, opts app.Options) error {
				return s.app.Run(ctx, selected.URL, opts)
			}),
			SwitchSection(SectionDownloads),
		)
	}
	return s, nil
}

func (s *Search) View(width, height int) string {
	var b strings.Builder
	b.WriteString(s.st.Subtitle.Render("Query:") + "\n")
	b.WriteString(s.input.View() + "\n")

	if s.loading {
		b.WriteString("\n" + s.st.Item.Render("searching…"))
		return b.String()
	}

	if suggestions := s.suggestions(); s.focus == searchFocusInput && len(suggestions) > 0 {
		b.WriteString("\n" + s.st.Help.Render("Suggestions") + "\n")
		for _, sug := range suggestions {
			b.WriteString("  " + s.st.Item.Render(sug) + "\n")
		}
	}

	if s.errMsg != "" {
		b.WriteString("\n" + s.st.Danger.Render(s.errMsg) + "\n")
	}

	if len(s.results) > 0 {
		b.WriteString("\n" + s.st.Subtitle.Render("Results") + "\n")
		for i, r := range s.results {
			cursor := "  "
			label := s.st.Item.Render(r.Title)
			if s.focus == searchFocusResults && i == s.cursor {
				cursor = s.st.Accent.Render("> ")
				label = s.st.Selected.Render(r.Title)
			}
			detail := fmt.Sprintf("%s · %s", r.Uploader, r.Duration)
			b.WriteString(cursor + label + "\n")
			if s.focus == searchFocusResults && i == s.cursor {
				b.WriteString("      " + s.st.Help.Render(detail) + "\n")
			}
		}
	}
	return b.String()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/workspace/... -run TestSearch -v`
Expected: PASS (5 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/workspace/search.go internal/ui/workspace/search_test.go
git commit -m "ui: add Search workspace with browsable results and suggestions"
```

---

### Task 8: `workspace.Input` (URL, Playlist, Enhance)

These three sections are near-identical: one text field, one submit action, one background task. They're implemented as a single parameterized type rather than three near-duplicate files, per the plan's file structure.

**Files:**
- Create: `internal/ui/workspace/input.go`
- Create: `internal/ui/workspace/input_test.go`

**Interfaces:**
- Consumes: `suggest.ClipboardURL` (Task 5), `style.Styles` (Task 1), `workspace.StartTask`, `workspace.SwitchSection` (Task 4).
- Produces: `workspace.NewURL(a *app.App, st style.Styles) *Input`, `workspace.NewPlaylist(a *app.App, st style.Styles) *Input`, `workspace.NewEnhance(a *app.App, st style.Styles) *Input`, `*Input` implementing `workspace.Workspace`.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/workspace/input_test.go
package workspace

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestInputActivateFocusesField(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewURL(a, st).Activate()
	if !ws.(*Input).input.Focused() {
		t.Fatal("Activate() did not focus the input")
	}
}

func TestInputSubmitStartsTaskAndSwitchesToDownloads(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewURL(a, st).Activate()

	for _, r := range "https://youtu.be/abc" {
		ws, _ = ws.Update(key(string(r)))
	}
	_, cmd := ws.Update(key("enter"))
	if cmd == nil {
		t.Fatal("expected a cmd after submitting a URL")
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
}

func TestInputEmptySubmitDoesNothing(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewURL(a, st).Activate()
	_, cmd := ws.Update(key("enter"))
	if cmd != nil {
		t.Fatal("expected no cmd when submitting an empty field")
	}
}

func TestEnhanceRunsEnhanceDir(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewEnhance(a, st).Activate()
	for _, r := range "." {
		ws, _ = ws.Update(key(string(r)))
	}
	_, cmd := ws.Update(key("enter"))
	msgs := flattenBatch(cmd())
	for _, m := range msgs {
		if st, ok := m.(StartTaskMsg); ok {
			if !strings.Contains(st.Summary, "Enhance") {
				t.Fatalf("Summary = %q, want it to mention Enhance", st.Summary)
			}
			return
		}
	}
	t.Fatal("expected a StartTaskMsg")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/workspace/... -run TestInput -v`
Expected: FAIL — `Input`, `NewURL`, `NewEnhance` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/workspace/input.go
package workspace

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	"github.com/EnJulian/shadowbox/internal/ui/suggest"
)

// Input is the shared workspace.Workspace behind URL, Playlist, and Enhance:
// a single text field, an optional clipboard hint, and a submit action that
// starts a background task.
type Input struct {
	app         *app.App
	st          style.Styles
	title       string
	summary     string
	clipboard   bool // offer a clipboard-URL hint on Activate
	input       textinput.Model
	clipHint    string
	submit      func(a *app.App, value string) func(ctx context.Context, opts app.Options) error
}

// NewURL builds the URL-download Input workspace.
func NewURL(a *app.App, st style.Styles) *Input {
	return newInput(a, st, "Enter a YouTube or Bandcamp URL", "Download", true,
		func(a *app.App, value string) func(context.Context, app.Options) error {
			return func(ctx context.Context, opts app.Options) error { return a.Run(ctx, value, opts) }
		})
}

// NewPlaylist builds the playlist-download Input workspace.
func NewPlaylist(a *app.App, st style.Styles) *Input {
	return newInput(a, st, "Enter a YouTube playlist URL", "Playlist download", true,
		func(a *app.App, value string) func(context.Context, app.Options) error {
			return func(ctx context.Context, opts app.Options) error { return a.RunPlaylist(ctx, value, opts) }
		})
}

// NewEnhance builds the directory-enhance Input workspace.
func NewEnhance(a *app.App, st style.Styles) *Input {
	return newInput(a, st, "Enter a directory of audio files to enhance", "Enhancement", false,
		func(a *app.App, value string) func(context.Context, app.Options) error {
			return func(ctx context.Context, opts app.Options) error {
				return a.EnhanceDir(ctx, value, true, []string{"opus", "mp3", "m4a", "flac"}, false, opts)
			}
		})
}

func newInput(a *app.App, st style.Styles, title, summary string, clipboard bool, submit func(*app.App, string) func(context.Context, app.Options) error) *Input {
	ti := textinput.New()
	ti.CharLimit = 512
	ti.Width = 60
	return &Input{app: a, st: st, title: title, summary: summary, clipboard: clipboard, input: ti, submit: submit}
}

func (in *Input) Activate() Workspace {
	in.input.SetValue("")
	in.input.Focus()
	in.clipHint = ""
	if in.clipboard {
		in.clipHint = suggest.ClipboardURL()
	}
	return in
}

func (in *Input) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return in, nil
	}
	switch keyMsg.String() {
	case "tab":
		if in.clipHint != "" {
			in.input.SetValue(in.clipHint)
			in.input.CursorEnd()
		}
		return in, nil
	case "enter":
		value := strings.TrimSpace(in.input.Value())
		if value == "" {
			return in, nil
		}
		run := in.submit(in.app, value)
		return in, tea.Batch(StartTask(in.summary, run), SwitchSection(SectionDownloads))
	}
	var cmd tea.Cmd
	in.input, cmd = in.input.Update(keyMsg)
	return in, cmd
}

func (in *Input) View(width, height int) string {
	var b strings.Builder
	b.WriteString(in.st.Subtitle.Render(in.title) + "\n\n")
	b.WriteString(in.input.View() + "\n")
	if in.clipHint != "" {
		b.WriteString("\n" + in.st.Help.Render("Paste from clipboard: "+in.clipHint+" — Tab to accept"))
	}
	return b.String()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/workspace/... -run TestInput -v` and `go test ./internal/ui/workspace/... -run TestEnhance -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/workspace/input.go internal/ui/workspace/input_test.go
git commit -m "ui: add shared Input workspace for URL, Playlist, and Enhance"
```

---

### Task 9: `workspace.Library`

**Files:**
- Create: `internal/ui/workspace/library.go`
- Create: `internal/ui/workspace/library_test.go`

**Interfaces:**
- Consumes: `suggest.Filter` (Task 5), `style.Styles` (Task 1), `shell.RequestNavFocus` (Task 2).
- Produces: `workspace.NewLibrary(cfg *config.Config, st style.Styles) *Library`, implementing `workspace.Workspace`.

- [ ] **Step 1: Write the failing test**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/workspace/... -run TestLibrary -v`
Expected: FAIL — `Library`, `NewLibrary` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/workspace/library.go
package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/ui/shell"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	"github.com/EnJulian/shadowbox/internal/ui/suggest"
)

var audioExts = map[string]bool{".opus": true, ".mp3": true, ".m4a": true, ".flac": true, ".wav": true, ".ogg": true}

// Library is the workspace.Workspace for the Library nav section: a
// drill-down Artist -> Album -> Track browser with type-ahead filtering.
type Library struct {
	cfg *config.Config
	st  style.Styles

	level    int // 0 = artists, 1 = albums, 2 = tracks
	cursor   int
	artist   string
	album    string
	allNames []string // unfiltered entries at the current level
	filter   string
}

// NewLibrary builds the Library workspace.
func NewLibrary(cfg *config.Config, st style.Styles) *Library {
	return &Library{cfg: cfg, st: st}
}

func (l *Library) Activate() Workspace {
	l.level, l.cursor, l.artist, l.album, l.filter = 0, 0, "", "", ""
	l.allNames = listDirs(l.cfg.MusicDirectory)
	return l
}

func (l *Library) entries() []string {
	return suggest.Filter(l.allNames, l.filter)
}

func (l *Library) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return l, nil
	}
	switch keyMsg.String() {
	case "up", "k":
		if l.cursor > 0 {
			l.cursor--
		}
	case "down", "j":
		if l.cursor < len(l.entries())-1 {
			l.cursor++
		}
	case "left", "h", "esc":
		return l.back()
	case "right", "l", "enter":
		return l.enter()
	case "backspace":
		if l.filter != "" {
			l.filter = l.filter[:len(l.filter)-1]
			l.cursor = 0
		}
	default:
		if len(keyMsg.Runes) == 1 {
			l.filter += string(keyMsg.Runes)
			l.cursor = 0
		}
	}
	return l, nil
}

func (l *Library) back() (Workspace, tea.Cmd) {
	l.filter = ""
	l.cursor = 0
	switch l.level {
	case 0:
		return l, shell.RequestNavFocus()
	case 1:
		l.level = 0
		l.artist = ""
		l.allNames = listDirs(l.cfg.MusicDirectory)
	case 2:
		l.level = 1
		l.album = ""
		l.allNames = listDirs(filepath.Join(l.cfg.MusicDirectory, l.artist))
	}
	return l, nil
}

func (l *Library) enter() (Workspace, tea.Cmd) {
	entries := l.entries()
	if len(entries) == 0 || l.level == 2 {
		return l, nil
	}
	selected := entries[l.cursor]
	l.filter = ""
	l.cursor = 0
	switch l.level {
	case 0:
		l.artist = selected
		l.level = 1
		l.allNames = listDirs(filepath.Join(l.cfg.MusicDirectory, l.artist))
	case 1:
		l.album = selected
		l.level = 2
		l.allNames = listTracks(filepath.Join(l.cfg.MusicDirectory, l.artist, l.album))
	}
	return l, nil
}

func (l *Library) View(width, height int) string {
	var b strings.Builder
	var crumb string
	switch l.level {
	case 0:
		crumb = "Library / Artists"
	case 1:
		crumb = "Library / " + l.artist
	case 2:
		crumb = "Library / " + l.artist + " / " + l.album
	}
	b.WriteString(l.st.Subtitle.Render(crumb) + "\n")
	if l.filter != "" {
		b.WriteString(l.st.Help.Render("filter: "+l.filter) + "\n")
	}
	b.WriteString("\n")

	entries := l.entries()
	if len(entries) == 0 {
		b.WriteString(l.st.Item.Render("(empty)") + "\n")
	}
	for i, e := range entries {
		cursor := "  "
		label := l.st.Item.Render(e)
		if i == l.cursor {
			cursor = l.st.Accent.Render("> ")
			label = l.st.Selected.Render(e)
		}
		b.WriteString(cursor + label + "\n")
	}
	return b.String()
}

func listDirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

func listTracks(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if audioExts[strings.ToLower(filepath.Ext(e.Name()))] {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}
```

Note: `TestLibraryDrillDownToTrack`'s cursor-seeking loop and the "> Nujabes" check depend on `View` prefixing the selected line with `"> "` after style rendering strips ANSI codes in tests run without a TTY color profile; `lipgloss` styles fall back to plain text when `NO_COLOR`/non-TTY is detected, so `s.st.Accent.Render("> ")` renders as literal `"> "`. This matches the existing test suite's behavior (see `ui_test.go`'s pre-existing `strings.Contains` assertions against styled output).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/workspace/... -run TestLibrary -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/workspace/library.go internal/ui/workspace/library_test.go
git commit -m "ui: add Library workspace with drill-down browsing and type-ahead filter"
```

---

### Task 10: `workspace.Downloads`

**Files:**
- Create: `internal/ui/workspace/downloads.go`
- Create: `internal/ui/workspace/downloads_test.go`

**Interfaces:**
- Consumes: `progress.Update` (existing, `internal/progress`), `style.Styles` (Task 1).
- Produces: `workspace.NewDownloads(st style.Styles) *Downloads`, `(*Downloads).SetSpinnerFrame(frame string)`, `(*Downloads).SetRunning(heading string)`, `(*Downloads).Finish(summary string, err error)`, `(*Downloads).ProgressBar() string` (used by root, exported for the renderer), implementing `workspace.Workspace`. `Downloads.Update` handles `progress.Update` messages directly (root forwards `progressMsg` — a `progress.Update` type alias defined in root — by converting it to `progress.Update` before calling `Update`).

To avoid the root package needing a type alias just to satisfy this workspace, `Downloads.Update` accepts `progress.Update` directly as a `tea.Msg` (progress.Update is a plain struct, safe to use as a message type on its own).

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/workspace/downloads_test.go
package workspace

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/progress"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestDownloadsActivateShowsIdleState(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	d := NewDownloads(st).Activate()
	view := d.View(80, 20)
	if !strings.Contains(view, "No active download") {
		t.Fatalf("expected idle message, got %q", view)
	}
}

func TestDownloadsSetRunningThenProgressUpdate(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	d := NewDownloads(st)
	d.SetRunning("Downloading audio")
	ws, _ := d.Update(progress.Update{Heading: "Tagging track 2/5", Current: 2, Total: 5})
	view := ws.View(80, 20)
	if !strings.Contains(view, "Tagging track 2/5") {
		t.Fatalf("expected heading in view, got %q", view)
	}
	if !strings.Contains(view, "2/5") {
		t.Fatalf("expected numbered progress in view, got %q", view)
	}
}

func TestDownloadsFinishShowsResult(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	d := NewDownloads(st)
	d.SetRunning("Downloading audio")
	d.Finish("Download complete", nil)
	view := d.View(80, 20)
	if !strings.Contains(view, "Download complete") {
		t.Fatalf("expected completion summary, got %q", view)
	}
}

func TestDownloadsCancelKeyEmitsCancelTaskMsg(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	d := NewDownloads(st)
	d.SetRunning("Downloading audio")
	_, cmd := d.Update(key("c"))
	if cmd == nil {
		t.Fatal("expected a cmd for the cancel key")
	}
	if _, ok := cmd().(CancelTaskMsg); !ok {
		t.Fatal("expected CancelTaskMsg")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/workspace/... -run TestDownloads -v`
Expected: FAIL — `Downloads`, `NewDownloads` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/workspace/downloads.go
package workspace

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/progress"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

const progressBarWidth = 24

// Downloads is the workspace.Workspace for the Downloads nav section: the
// active job's heading, progress bar, and completion result.
type Downloads struct {
	st style.Styles

	running       bool
	heading       string
	progress      progress.Update
	spinnerFrame  string
	resultSummary string
	resultErr     error
}

// NewDownloads builds the Downloads workspace.
func NewDownloads(st style.Styles) *Downloads {
	return &Downloads{st: st}
}

func (d *Downloads) Activate() Workspace { return d }

// SetSpinnerFrame is called by root each spinner tick so the running view
// can show an animated frame without Downloads owning its own spinner.Model.
func (d *Downloads) SetSpinnerFrame(frame string) { d.spinnerFrame = frame }

// SetRunning transitions to the running state with the given initial heading.
func (d *Downloads) SetRunning(heading string) {
	d.running = true
	d.heading = heading
	d.progress = progress.Update{}
	d.resultSummary = ""
	d.resultErr = nil
}

// Finish transitions out of the running state with a result.
func (d *Downloads) Finish(summary string, err error) {
	d.running = false
	d.resultSummary = summary
	d.resultErr = err
}

func (d *Downloads) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	switch msg := msg.(type) {
	case progress.Update:
		if msg.Heading != "" {
			d.heading = msg.Heading
		}
		d.progress = msg
		return d, nil
	case tea.KeyMsg:
		if d.running && msg.String() == "c" {
			return d, CancelTask()
		}
	}
	return d, nil
}

func (d *Downloads) View(width, height int) string {
	var b strings.Builder
	switch {
	case d.running:
		heading := d.heading
		if heading == "" {
			heading = "Initializing"
		}
		b.WriteString(d.spinnerFrame + " " + d.st.Accent.Render(heading) + "\n\n")
		b.WriteString(renderProgressBar(d.progress, d.st) + "\n\n")
		b.WriteString(d.st.Help.Render("c: cancel"))
	case d.resultSummary != "":
		if d.resultErr != nil {
			b.WriteString(d.st.Danger.Render("x " + d.resultSummary) + "\n")
			b.WriteString(d.st.Item.Render(d.resultErr.Error()))
		} else {
			b.WriteString(d.st.Selected.Render("+ " + d.resultSummary))
		}
	default:
		b.WriteString(d.st.Item.Render("No active download — start one from Search, URL, or Playlist."))
	}
	return b.String()
}

// renderProgressBar returns a minimal text progress bar, reusing the same
// visual language as the pre-redesign UI.
func renderProgressBar(u progress.Update, st style.Styles) string {
	filled := progressBarWidth / 4
	if u.Total > 0 && u.Current > 0 {
		current := u.Current
		if current > u.Total {
			current = u.Total
		}
		filled = (current * progressBarWidth) / u.Total
	}
	bar := "[" + st.Accent.Render(strings.Repeat("─", filled)) + st.Help.Render(strings.Repeat("·", progressBarWidth-filled)) + "]"
	if u.Total > 0 && u.Current > 0 {
		return bar + st.Help.Render(fmt.Sprintf(" %d/%d", u.Current, u.Total))
	}
	return bar
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/workspace/... -run TestDownloads -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/workspace/downloads.go internal/ui/workspace/downloads_test.go
git commit -m "ui: add Downloads workspace"
```

---

### Task 11: `workspace.Log`

**Files:**
- Create: `internal/ui/workspace/log.go`
- Create: `internal/ui/workspace/log_test.go`

**Interfaces:**
- Consumes: `applog.LoadDownloadLog() error`, `applog.DownloadLogLines() []string` (existing, `internal/log`), `style.Styles` (Task 1).
- Produces: `workspace.NewLog(st style.Styles) *Log`, implementing `workspace.Workspace`.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/workspace/log_test.go
package workspace

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestLogActivateWithNoLogsShowsHint(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	l := NewLog(st).Activate()
	view := l.View(80, 20)
	if !strings.Contains(view, "no download logs yet") {
		t.Fatalf("expected empty-log hint, got %q", view)
	}
}

func TestLogScrollClampsToBounds(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	l := NewLog(st)
	l.lines = []string{"a", "b", "c", "d", "e"}
	l.scroll = 0

	ws, _ := l.Update(key("up")) // already at top, should not go negative
	if ws.(*Log).scroll != 0 {
		t.Fatalf("scroll = %d, want 0", ws.(*Log).scroll)
	}

	ws, _ = ws.Update(key("end"))
	if ws.(*Log).scroll < 0 {
		t.Fatalf("scroll went negative: %d", ws.(*Log).scroll)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/workspace/... -run TestLog -v`
Expected: FAIL — `Log`, `NewLog` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/workspace/log.go
package workspace

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	applog "github.com/EnJulian/shadowbox/internal/log"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

// Log is the workspace.Workspace for the Log nav section: the scrollable
// download log viewer.
type Log struct {
	st     style.Styles
	lines  []string
	scroll int
	height int // last known viewport height, for scroll clamping
}

// NewLog builds the Log workspace.
func NewLog(st style.Styles) *Log {
	return &Log{height: 10}
}

func (l *Log) Activate() Workspace {
	_ = applog.LoadDownloadLog()
	l.lines = applog.DownloadLogLines()
	l.scroll = l.maxScroll()
	return l
}

func (l *Log) maxScroll() int {
	if len(l.lines) <= l.height {
		return 0
	}
	return len(l.lines) - l.height
}

func (l *Log) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return l, nil
	}
	switch keyMsg.String() {
	case "up", "k":
		if l.scroll > 0 {
			l.scroll--
		}
	case "down", "j":
		if l.scroll < l.maxScroll() {
			l.scroll++
		}
	case "pgup":
		l.scroll = max(0, l.scroll-l.height)
	case "pgdown":
		l.scroll = min(l.maxScroll(), l.scroll+l.height)
	case "home":
		l.scroll = 0
	case "end", "G":
		l.scroll = l.maxScroll()
	case "r":
		_ = applog.LoadDownloadLog()
		l.lines = applog.DownloadLogLines()
		l.scroll = l.maxScroll()
	}
	return l, nil
}

func (l *Log) View(width, height int) string {
	if height > 4 {
		l.height = height - 4
	}
	if l.scroll > l.maxScroll() {
		l.scroll = l.maxScroll()
	}

	var b strings.Builder
	if len(l.lines) == 0 {
		b.WriteString(l.st.Item.Render("(no download logs yet — run a download first)"))
		return b.String()
	}

	end := l.scroll + l.height
	if end > len(l.lines) {
		end = len(l.lines)
	}
	for _, line := range l.lines[l.scroll:end] {
		b.WriteString(l.st.Item.Render(line) + "\n")
	}
	if len(l.lines) > l.height {
		b.WriteString("\n" + l.st.Help.Render("lines "+strconv.Itoa(l.scroll+1)+"-"+strconv.Itoa(end)+" of "+strconv.Itoa(len(l.lines))))
	}
	return b.String()
}
```

Go 1.25's builtin `min`/`max` (available since Go 1.21) are used directly — no local redeclaration needed, unlike the pre-redesign `log_view.go`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/workspace/... -run TestLog -v`
Expected: PASS (2 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/workspace/log.go internal/ui/workspace/log_test.go
git commit -m "ui: add Log workspace"
```

---

### Task 12: `workspace.Settings`

**Files:**
- Create: `internal/ui/workspace/settings.go`
- Create: `internal/ui/workspace/settings_test.go`

**Interfaces:**
- Consumes: `config.Config`, `config.Save` (existing, `internal/config`), `style.Styles` (Task 1).
- Produces: `workspace.NewSettings(cfg *config.Config, st style.Styles) *Settings`, `workspace.SettingsThemeRequestMsg{}` (asks root to open the theme picker overlay), `workspace.SettingsChangedMsg{}` (asks root to reload theme/app/workspaces after a save), implementing `workspace.Workspace`.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/workspace/settings_test.go
package workspace

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestSettingsViewListsItems(t *testing.T) {
	cfg := &config.Config{AudioFormat: "opus", Theme: "hacker"}
	st := style.NewStyles(style.ThemeByName("hacker"))
	s := NewSettings(cfg, st).Activate()
	view := s.View(80, 20)
	if !strings.Contains(view, "Audio format") {
		t.Fatalf("expected settings list, got %q", view)
	}
}

func TestSettingsToggleFlipsBoolAndEmitsChanged(t *testing.T) {
	cfg := &config.Config{UseGenius: true}
	st := style.NewStyles(style.ThemeByName("hacker"))
	s := NewSettings(cfg, st).Activate()

	// use_genius is index 2 in the fixed settingItems ordering.
	for i := 0; i < 2; i++ {
		s, _ = s.Update(key("down"))
	}
	_, cmd := s.Update(key("enter"))
	if cfg.UseGenius {
		t.Fatal("expected UseGenius to flip to false")
	}
	if cmd == nil {
		t.Fatal("expected a SettingsChangedMsg cmd after toggling")
	}
	if _, ok := cmd().(SettingsChangedMsg); !ok {
		t.Fatal("expected SettingsChangedMsg")
	}
}

func TestSettingsThemeItemRequestsThemePicker(t *testing.T) {
	cfg := &config.Config{Theme: "hacker"}
	st := style.NewStyles(style.ThemeByName("hacker"))
	s := NewSettings(cfg, st).Activate()

	// theme is index 4.
	for i := 0; i < 4; i++ {
		s, _ = s.Update(key("down"))
	}
	_, cmd := s.Update(key("enter"))
	if cmd == nil {
		t.Fatal("expected a SettingsThemeRequestMsg cmd")
	}
	if _, ok := cmd().(SettingsThemeRequestMsg); !ok {
		t.Fatal("expected SettingsThemeRequestMsg")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/workspace/... -run TestSettings -v`
Expected: FAIL — `Settings`, `NewSettings` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/workspace/settings.go
package workspace

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

type settingKind int

const (
	settingToggle settingKind = iota
	settingText
	settingSecret
	settingTheme
)

type settingItem struct {
	key   string
	label string
	kind  settingKind
}

var settingItems = []settingItem{
	{"audio_format", "Audio format", settingText},
	{"music_directory", "Music directory", settingText},
	{"use_genius", "Use Genius lyrics", settingToggle},
	{"verbose", "Verbose logging", settingToggle},
	{"theme", "Theme", settingTheme},
	{"genius.access_token", "Genius access token", settingSecret},
}

// SettingsThemeRequestMsg asks root to open the theme picker overlay.
type SettingsThemeRequestMsg struct{}

// SettingsChangedMsg asks root to reload theme/app/workspaces from cfg after a save.
type SettingsChangedMsg struct{}

// Settings is the workspace.Workspace for the Settings nav section.
type Settings struct {
	cfg *config.Config
	st  style.Styles

	cursor  int
	editing bool
	editKey string
	input   textinput.Model
}

// NewSettings builds the Settings workspace.
func NewSettings(cfg *config.Config, st style.Styles) *Settings {
	ti := textinput.New()
	ti.CharLimit = 512
	ti.Width = 60
	return &Settings{cfg: cfg, st: st, input: ti}
}

func (s *Settings) Activate() Workspace {
	s.editing = false
	return s
}

func (s *Settings) value(key string) string {
	switch key {
	case "audio_format":
		return s.cfg.AudioFormat
	case "music_directory":
		return s.cfg.MusicDirectory
	case "use_genius":
		return boolLabel(s.cfg.UseGenius)
	case "verbose":
		return boolLabel(s.cfg.Verbose)
	case "theme":
		return s.cfg.Theme
	case "genius.access_token":
		return maskSecret(s.cfg.Genius.AccessToken)
	}
	return ""
}

func (s *Settings) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}
	if s.editing {
		return s.updateEdit(keyMsg)
	}

	switch keyMsg.String() {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down", "j":
		if s.cursor < len(settingItems)-1 {
			s.cursor++
		}
	case "enter":
		return s.activate()
	}
	return s, nil
}

func (s *Settings) activate() (Workspace, tea.Cmd) {
	item := settingItems[s.cursor]
	switch item.kind {
	case settingToggle:
		s.toggle(item.key)
		_ = config.Save(s.cfg)
		return s, func() tea.Msg { return SettingsChangedMsg{} }
	case settingTheme:
		return s, func() tea.Msg { return SettingsThemeRequestMsg{} }
	case settingText, settingSecret:
		s.editing = true
		s.editKey = item.key
		s.input.SetValue("")
		s.input.Placeholder = "type a new value, enter to save"
		s.input.Focus()
	}
	return s, nil
}

func (s *Settings) toggle(key string) {
	switch key {
	case "use_genius":
		s.cfg.UseGenius = !s.cfg.UseGenius
	case "verbose":
		s.cfg.Verbose = !s.cfg.Verbose
	}
}

func (s *Settings) updateEdit(msg tea.KeyMsg) (Workspace, tea.Cmd) {
	switch msg.String() {
	case "esc":
		s.editing = false
		return s, nil
	case "enter":
		value := strings.TrimSpace(s.input.Value())
		s.editing = false
		if value == "" {
			return s, nil
		}
		s.apply(s.editKey, value)
		_ = config.Save(s.cfg)
		return s, func() tea.Msg { return SettingsChangedMsg{} }
	}
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

func (s *Settings) apply(key, value string) {
	switch key {
	case "audio_format":
		s.cfg.AudioFormat = value
	case "music_directory":
		s.cfg.MusicDirectory = value
	case "genius.access_token":
		s.cfg.Genius.AccessToken = value
	}
}

func (s *Settings) View(width, height int) string {
	if s.editing {
		var b strings.Builder
		b.WriteString(s.st.Subtitle.Render("Edit "+s.editKey) + "\n\n")
		b.WriteString(s.input.View() + "\n\n")
		b.WriteString(s.st.Help.Render("enter: save   esc: cancel"))
		return b.String()
	}

	var b strings.Builder
	for i, item := range settingItems {
		cursor := "  "
		label := s.st.Item.Render(item.label)
		if i == s.cursor {
			cursor = s.st.Accent.Render("> ")
			label = s.st.Selected.Render(item.label)
		}
		value := s.st.Accent.Render(s.value(item.key))
		fmt.Fprintf(&b, "%s%-26s %s\n", cursor, label, value)
	}
	return b.String()
}

func boolLabel(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func maskSecret(v string) string {
	if v == "" {
		return "(not set)"
	}
	if len(v) <= 4 {
		return strings.Repeat("*", len(v))
	}
	return strings.Repeat("*", len(v)-4) + v[len(v)-4:]
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/workspace/... -run TestSettings -v`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/workspace/settings.go internal/ui/workspace/settings_test.go
git commit -m "ui: add Settings workspace"
```

---

### Task 13: `overlay` package — picker, help, result, theme

**Files:**
- Create: `internal/ui/overlay/picker.go`
- Create: `internal/ui/overlay/picker_test.go`
- Create: `internal/ui/overlay/help.go`
- Create: `internal/ui/overlay/result.go`
- Create: `internal/ui/overlay/theme.go`

**Interfaces:**
- Consumes: `app.PromptOption` (existing, `internal/app`), `style.Styles`, `style.ThemeOrder`, `style.ThemeByName` (Task 1).
- Produces: `overlay.Picker{Title string, Options []app.PromptOption, Cursor int}`, `(*Picker).MoveUp()`, `(*Picker).MoveDown()`, `(*Picker).View(st style.Styles) string`; `overlay.Help{}`, `(Help).View(st style.Styles) string`; `overlay.Result{Summary string, Err error}`, `(Result).View(st style.Styles) string`; `overlay.Theme{Cursor int}`, `(*Theme).MoveUp()`, `(*Theme).MoveDown()`, `(Theme).Selected() string`, `(Theme).View(st style.Styles) string`.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/overlay/picker_test.go
package overlay

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestPickerViewRendersOptions(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	p := Picker{
		Title: "Select track",
		Options: []app.PromptOption{
			{Label: "Song A — Artist", Detail: "Album (2020) · 3:30"},
			{Label: "Song B — Artist", Detail: "Single (2021)"},
		},
	}
	view := p.View(st)
	if !strings.Contains(view, "Select track") || !strings.Contains(view, "Song A — Artist") {
		t.Fatalf("missing expected content: %q", view)
	}
}

func TestPickerMoveClampsAtBounds(t *testing.T) {
	p := Picker{Options: []app.PromptOption{{Label: "a"}, {Label: "b"}}}
	p.MoveUp()
	if p.Cursor != 0 {
		t.Fatalf("Cursor = %d, want 0 (clamped)", p.Cursor)
	}
	p.MoveDown()
	if p.Cursor != 1 {
		t.Fatalf("Cursor = %d, want 1", p.Cursor)
	}
	p.MoveDown()
	if p.Cursor != 1 {
		t.Fatalf("Cursor = %d, want 1 (clamped)", p.Cursor)
	}
}

func TestThemeMoveAndSelected(t *testing.T) {
	th := Theme{}
	if th.Selected() != style.ThemeOrder[0] {
		t.Fatalf("Selected() = %q, want %q", th.Selected(), style.ThemeOrder[0])
	}
	th.MoveDown()
	if th.Selected() != style.ThemeOrder[1] {
		t.Fatalf("Selected() = %q, want %q", th.Selected(), style.ThemeOrder[1])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/overlay/... -v`
Expected: FAIL — package does not exist.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/overlay/picker.go

// Package overlay renders modal content shown on top of the shell: the
// metadata/URL disambiguation picker, the keybinding help screen, the
// task-result toast, and the theme picker.
package overlay

import (
	"strings"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

// Picker is the disambiguation overlay shown when a pipeline step needs the
// user to choose among several candidates (metadata match, URL variant, …).
type Picker struct {
	Title   string
	Options []app.PromptOption
	Cursor  int
}

func (p *Picker) MoveUp() {
	if p.Cursor > 0 {
		p.Cursor--
	}
}

func (p *Picker) MoveDown() {
	if p.Cursor < len(p.Options)-1 {
		p.Cursor++
	}
}

func (p Picker) View(st style.Styles) string {
	var b strings.Builder
	b.WriteString(st.Subtitle.Render(p.Title) + "\n\n")
	if len(p.Options) == 0 {
		b.WriteString(st.Item.Render("(no options)") + "\n")
	}
	for i, o := range p.Options {
		cursor := "  "
		label := st.Item.Render(o.Label)
		if i == p.Cursor {
			cursor = st.Accent.Render("> ")
			label = st.Selected.Render(o.Label)
		}
		b.WriteString(cursor + label + "\n")
		if o.Detail != "" {
			b.WriteString("    " + st.Help.Render(o.Detail) + "\n")
		}
	}
	b.WriteString("\n" + st.Help.Render("up/down: navigate   enter: select   esc: cancel"))
	return b.String()
}
```

```go
// internal/ui/overlay/help.go
package overlay

import "github.com/EnJulian/shadowbox/internal/ui/style"

// Help is the static keybinding reference overlay, toggled by "?".
type Help struct{}

var helpLines = []string{
	"1-8       jump to a nav section",
	"Tab       cycle focus: Nav <-> Content",
	"<- -> h l switch panes (Library: columns first)",
	"up down   move within the focused pane",
	"/         focus the Search query input",
	"enter     activate / submit / select",
	"esc       back / cancel / close",
	"?         toggle this help",
	"q         quit (confirms if a download is running)",
	"ctrl+c    force quit",
}

func (Help) View(st style.Styles) string {
	out := st.Subtitle.Render("Keybindings") + "\n\n"
	for _, line := range helpLines {
		out += st.Item.Render(line) + "\n"
	}
	return out + "\n" + st.Help.Render("? or esc: close")
}
```

```go
// internal/ui/overlay/result.go
package overlay

import "github.com/EnJulian/shadowbox/internal/ui/style"

// Result is the task-completion toast, dismissed by any key.
type Result struct {
	Summary string
	Err     error
}

func (r Result) View(st style.Styles) string {
	var body string
	if r.Err != nil {
		body = st.Danger.Render("x "+r.Summary) + "\n" + st.Item.Render(r.Err.Error())
	} else {
		body = st.Selected.Render("+ " + r.Summary)
	}
	return body + "\n\n" + st.Help.Render("press any key to dismiss")
}
```

```go
// internal/ui/overlay/theme.go
package overlay

import (
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

// Theme is the theme picker overlay, applied live for instant preview as
// the cursor moves.
type Theme struct {
	Cursor int
}

func (t *Theme) MoveUp() {
	if t.Cursor > 0 {
		t.Cursor--
	}
}

func (t *Theme) MoveDown() {
	if t.Cursor < len(style.ThemeOrder)-1 {
		t.Cursor++
	}
}

// Selected returns the name of the currently highlighted theme.
func (t Theme) Selected() string {
	return style.ThemeOrder[t.Cursor]
}

func (t Theme) View(st style.Styles) string {
	out := st.Subtitle.Render("Choose a theme") + "\n\n"
	for i, name := range style.ThemeOrder {
		cursor := "  "
		label := st.Item.Render(name)
		if i == t.Cursor {
			cursor = st.Accent.Render("> ")
			label = st.Selected.Render(name)
		}
		out += cursor + label + "\n"
	}
	return out + "\n" + st.Help.Render("up/down: preview   enter: apply   esc: cancel")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/overlay/... -v`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/overlay/
git commit -m "ui: add overlay package for picker, help, result, and theme"
```

---

### Task 14: `playback` package stub

**Files:**
- Create: `internal/ui/playback/playback.go`

**Interfaces:**
- Produces: `playback.State{Active bool}`.

No test needed — this is a data-only stub with no logic (`Active` is always the zero value `false` this phase, per spec §Non-Goals). It exists purely so `shell.Layout.PlaybarHeight` and the root model's render call have a concrete type to reference, avoiding a relayout when playback ships.

- [ ] **Step 1: Write the implementation**

```go
// internal/ui/playback/playback.go

// Package playback is a Phase 1 stub. It exists so the shell's reserved
// playback bar slot (shell.Layout.PlaybarHeight) and the root model have a
// concrete type to hold, without any decoding/IPC logic yet — that ships in
// a later update (see docs/superpowers/specs/2026-07-11-terminal-ui-redesign-v2-design.md).
package playback

// State describes the current playback state. Active is always false this
// phase.
type State struct {
	Active bool
}
```

- [ ] **Step 2: Verify it builds**

Run: `go build ./internal/ui/playback/...`
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/ui/playback/playback.go
git commit -m "ui: add playback state stub for the reserved playbar slot"
```

---

### Task 15: Root model rewrite — wire the shell together

**Files:**
- Modify: `internal/ui/program.go` (full rewrite)
- Modify: `internal/ui/ui_test.go` (full rewrite)
- Keep: `internal/ui/helpers.go`, `internal/ui/ui.go` unchanged
- Note: `internal/ui/progress.go` and `internal/ui/progress_test.go` are untouched by this task — `renderProgressBar` already moved to `workspace/downloads.go` in Task 10 with its own tests, so these two files are now dead code, deleted outright in Task 16.

**Interfaces:**
- Consumes everything from Tasks 1–14: `style`, `shell`, `workspace`, `overlay`, `playback`.
- Produces: the `model` type implementing `tea.Model` (`Init`, `Update`, `View`), unchanged external contract `ui.Run(cfg *config.Config) error`.

This task has no isolated unit-testable slice on its own — it's the integration point — so instead of a red/green TDD cycle, write the implementation directly and verify it with the test suite in Step 2, following the existing `ui_test.go` pattern of driving the model's `Update`/`View` methods directly.

- [ ] **Step 1: Write the implementation**

```go
// internal/ui/program.go
package ui

import (
	"context"
	"io"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	applog "github.com/EnJulian/shadowbox/internal/log"
	"github.com/EnJulian/shadowbox/internal/progress"
	"github.com/EnJulian/shadowbox/internal/ui/overlay"
	"github.com/EnJulian/shadowbox/internal/ui/playback"
	"github.com/EnJulian/shadowbox/internal/ui/shell"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	"github.com/EnJulian/shadowbox/internal/ui/workspace"
)

type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayPicker
	overlayHelp
	overlayResult
	overlayTheme
)

// promptOutgoing/promptResult wire app.SelectFunc through to the picker overlay.
type promptOutgoing struct {
	req  app.PromptRequest
	resp chan promptResult
}

type promptResult struct {
	idx int
	err error
}

type promptRequestMsg struct{ out promptOutgoing }

// taskDoneMsg is emitted when a background pipeline operation completes.
type taskDoneMsg struct {
	summary string
	err     error
}

type model struct {
	cfg   *config.Config
	app   *app.App
	theme style.Theme
	st    style.Styles

	width, height int

	pane          shell.Pane
	activeSection workspace.Section
	workspaces    map[workspace.Section]workspace.Workspace
	downloads     *workspace.Downloads

	spinner  spinner.Model
	playback playback.State

	taskCancel  context.CancelFunc
	progressCh  chan progress.Update
	promptReqCh chan promptOutgoing

	ov      overlayKind
	picker  overlay.Picker
	pending *promptOutgoing
	themeOv overlay.Theme
	result  overlay.Result
}

func historyPath(cfg *config.Config) string {
	dir, err := config.Dir()
	if err != nil {
		return ""
	}
	return dir + "/search_history"
}

func (m *model) buildWorkspaces() {
	downloads := workspace.NewDownloads(m.st)
	m.downloads = downloads
	m.workspaces = map[workspace.Section]workspace.Workspace{
		workspace.SectionSearch:    workspace.NewSearch(m.app, m.cfg, m.st, historyPath(m.cfg)),
		workspace.SectionURL:       workspace.NewURL(m.app, m.st),
		workspace.SectionPlaylist:  workspace.NewPlaylist(m.app, m.st),
		workspace.SectionLibrary:   workspace.NewLibrary(m.cfg, m.st),
		workspace.SectionDownloads: downloads,
		workspace.SectionEnhance:   workspace.NewEnhance(m.app, m.st),
		workspace.SectionLog:       workspace.NewLog(m.st),
		workspace.SectionSettings:  workspace.NewSettings(m.cfg, m.st),
	}
}

// runProgram builds and runs the Bubble Tea program.
func runProgram(cfg *config.Config) error {
	capture := applog.DownloadLogWriter()
	_ = applog.LoadDownloadLog()
	applog.SetWriters(io.MultiWriter(io.Discard, capture), io.MultiWriter(io.Discard, capture))
	applog.SetVerbose(true)
	defer applog.Reset()

	theme := style.ThemeByName(cfg.Theme)
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := &model{
		cfg:           cfg,
		app:           app.New(cfg),
		theme:         theme,
		st:            style.NewStyles(theme),
		spinner:       sp,
		activeSection: workspace.SectionSearch,
	}
	m.buildWorkspaces()
	m.workspaces[workspace.SectionSearch] = m.workspaces[workspace.SectionSearch].Activate()

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m *model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.downloads.SetSpinnerFrame(m.spinner.View())
		return m, cmd

	case progress.Update:
		m.downloads.Update(msg)
		return m, m.taskListenCmd()

	case promptRequestMsg:
		m.ov = overlayPicker
		m.picker = overlay.Picker{Title: msg.out.req.Title, Options: msg.out.req.Options}
		m.pending = &msg.out
		return m, m.taskListenCmd()

	case taskDoneMsg:
		m.ov = overlayResult
		m.result = overlay.Result{Summary: msg.summary, Err: msg.err}
		m.downloads.Finish(msg.summary, msg.err)
		m.taskCancel = nil
		m.promptReqCh = nil
		return m, nil

	case workspace.StartTaskMsg:
		return m, m.startTask(msg.Summary, msg.Run)

	case workspace.SwitchSectionMsg:
		m.switchSection(msg.Section)
		return m, nil

	case workspace.CancelTaskMsg:
		if m.taskCancel != nil {
			m.taskCancel()
		}
		return m, nil

	case workspace.SettingsThemeRequestMsg:
		m.ov = overlayTheme
		m.themeOv = overlay.Theme{Cursor: themeIndex(m.cfg.Theme)}
		return m, nil

	case workspace.SettingsChangedMsg:
		m.applyTheme(m.cfg.Theme)
		m.app = rebuildApp(m.cfg)
		m.buildWorkspaces()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func themeIndex(name string) int {
	for i, n := range style.ThemeOrder {
		if n == name {
			return i
		}
	}
	return 0
}

func (m *model) applyTheme(name string) {
	m.theme = style.ThemeByName(name)
	m.st = style.NewStyles(m.theme)
}

func (m *model) switchSection(s workspace.Section) {
	m.activeSection = s
	m.pane = shell.PaneContent
	m.workspaces[s] = m.workspaces[s].Activate()
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.ov != overlayNone {
		return m.handleOverlayKey(msg)
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "?":
		m.ov = overlayHelp
		return m, nil
	case "/":
		m.switchSection(workspace.SectionSearch)
		return m, nil
	case "tab", "shift+tab":
		m.pane = m.pane.Toggle()
		if m.pane == shell.PaneContent {
			m.workspaces[m.activeSection] = m.workspaces[m.activeSection].Activate()
		}
		return m, nil
	}

	if n := sectionForDigit(msg.String()); n >= 0 {
		m.switchSection(workspace.Section(n))
		return m, nil
	}

	if m.pane == shell.PaneNav {
		return m.handleNavKey(msg)
	}
	return m.handleContentKey(msg)
}

func sectionForDigit(s string) int {
	if len(s) == 1 && s[0] >= '1' && s[0] <= '8' {
		return int(s[0] - '1')
	}
	return -1
}

func (m *model) handleNavKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.activeSection > 0 {
			m.activeSection--
		}
	case "down", "j":
		if int(m.activeSection) < len(workspace.Order)-1 {
			m.activeSection++
		}
	case "right", "l", "enter":
		m.switchSection(m.activeSection)
	}
	return m, nil
}

func (m *model) handleContentKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ws, cmd := m.workspaces[m.activeSection].Update(msg)
	m.workspaces[m.activeSection] = ws
	return m, cmd
}

func (m *model) handleOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.ov {
	case overlayPicker:
		return m.handlePickerKey(msg)
	case overlayHelp:
		if msg.String() == "?" || msg.String() == "esc" {
			m.ov = overlayNone
		}
	case overlayResult:
		m.ov = overlayNone
	case overlayTheme:
		return m.handleThemeKey(msg)
	}
	return m, nil
}

func (m *model) handlePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pending == nil {
		m.ov = overlayNone
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.pending.resp <- promptResult{idx: -1, err: app.ErrSelectionCancelled}
		m.pending = nil
		m.ov = overlayNone
		if m.taskCancel != nil {
			m.taskCancel()
		}
	case "up", "k":
		m.picker.MoveUp()
	case "down", "j":
		m.picker.MoveDown()
	case "enter":
		idx := m.picker.Cursor
		pending := m.pending
		m.pending = nil
		m.ov = overlayNone
		pending.resp <- promptResult{idx: idx, err: nil}
		return m, m.taskListenCmd()
	}
	return m, nil
}

func (m *model) handleThemeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.applyTheme(m.cfg.Theme) // revert live preview
		m.ov = overlayNone
	case "up", "k":
		m.themeOv.MoveUp()
		m.applyTheme(m.themeOv.Selected())
	case "down", "j":
		m.themeOv.MoveDown()
		m.applyTheme(m.themeOv.Selected())
	case "enter":
		m.cfg.Theme = m.themeOv.Selected()
		_ = config.Save(m.cfg)
		m.applyTheme(m.cfg.Theme)
		m.ov = overlayNone
	}
	return m, nil
}

// startTask transitions Downloads to the running state and runs fn in the background.
func (m *model) startTask(summary string, fn func(ctx context.Context, opts app.Options) error) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.taskCancel = cancel
	m.downloads.SetRunning("Initializing")
	applog.BeginDownloadSession("Initializing")

	ch := make(chan progress.Update, 32)
	m.progressCh = ch
	promptCh := make(chan promptOutgoing, 1)
	m.promptReqCh = promptCh

	report := func(u progress.Update) {
		select {
		case ch <- u:
		default:
		}
	}

	run := func() tea.Msg {
		opts := app.Options{Progress: report, Select: makeSelectFunc(promptCh)}
		err := fn(ctx, opts)
		cancel()
		close(promptCh)
		close(ch)
		summaryMsg := summary + " complete"
		if err != nil {
			summaryMsg = summary + " failed"
		}
		return taskDoneMsg{summary: summaryMsg, err: err}
	}

	return tea.Batch(m.spinner.Tick, run, m.taskListenCmd())
}

func (m *model) taskListenCmd() tea.Cmd {
	var cmds []tea.Cmd
	if m.progressCh != nil {
		cmds = append(cmds, waitForProgress(m.progressCh))
	}
	if m.promptReqCh != nil {
		cmds = append(cmds, waitForPrompt(m.promptReqCh))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func waitForProgress(ch chan progress.Update) tea.Cmd {
	return func() tea.Msg {
		u, ok := <-ch
		if !ok {
			return nil
		}
		return u
	}
}

func waitForPrompt(reqCh chan promptOutgoing) tea.Cmd {
	return func() tea.Msg {
		out, ok := <-reqCh
		if !ok {
			return nil
		}
		return promptRequestMsg{out: out}
	}
}

func makeSelectFunc(reqCh chan promptOutgoing) app.SelectFunc {
	return func(ctx context.Context, req app.PromptRequest) (int, error) {
		respCh := make(chan promptResult, 1)
		out := promptOutgoing{req: req, resp: respCh}
		select {
		case reqCh <- out:
		case <-ctx.Done():
			return -1, ctx.Err()
		}
		select {
		case r := <-respCh:
			return r.idx, r.err
		case <-ctx.Done():
			return -1, ctx.Err()
		}
	}
}

func (m *model) navBody() string {
	var b string
	for i, item := range workspace.Order {
		cursor := "  "
		label := m.st.Item.Render(item.Label)
		if workspace.Section(i) == m.activeSection {
			label = m.st.Selected.Render(item.Label)
			if m.pane == shell.PaneNav {
				cursor = m.st.Accent.Render("> ")
			} else {
				cursor = "  "
			}
		}
		b += cursor + label + "\n"
	}
	return b
}

func (m *model) statusBody() string {
	if m.ov != overlayNone {
		return m.st.Help.Render("esc: close")
	}
	if m.pane == shell.PaneNav {
		return m.st.Help.Render(workspace.Order[m.activeSection].Label + " │ up/down: navigate │ enter/→: open │ tab: content │ ?: help │ q: quit")
	}
	return m.st.Help.Render(workspace.Order[m.activeSection].Label + " │ tab: nav │ ←: back │ ?: help │ q: quit")
}

func (m *model) View() string {
	layout := shell.Compute(m.width, m.height)
	content := m.workspaces[m.activeSection].View(layout.ContentWidth-2, layout.ContentHeight-2)

	if m.ov != overlayNone {
		content = m.overlayBody(layout)
	}

	return shell.Render(m.st, m.theme, layout, m.pane == shell.PaneNav, m.navBody(), content, "", m.statusBody())
}

func (m *model) overlayBody(layout shell.Layout) string {
	switch m.ov {
	case overlayPicker:
		return m.picker.View(m.st)
	case overlayHelp:
		return overlay.Help{}.View(m.st)
	case overlayResult:
		return m.result.View(m.st)
	case overlayTheme:
		return m.themeOv.View(m.st)
	}
	return ""
}
```

Note: `historyPath`'s unused `cfg` parameter shadowing note — it's intentionally named to document intent even though it currently just calls `config.Dir()`; keep the parameter for symmetry with future per-user history paths, but if `go vet`/lint flags an unused parameter, drop it and call `config.Dir()` directly (parameters aren't flagged as unused by standard `go vet`, so this is not expected to trigger lint — verify in Step 2).

- [ ] **Step 2: Update ui_test.go and run the full package test suite**

```go
// internal/ui/ui_test.go
package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
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

	next, _ := m.handleKey(key("tab"))
	m = next.(*model)
	if m.pane.String() == "" { // sanity: pane changed without panicking
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

func TestQuitKeyReturnsQuitCmd(t *testing.T) {
	m := newTestModel(t)
	_, cmd := m.handleKey(key("q"))
	if cmd == nil {
		t.Fatal("expected a quit cmd")
	}
}
```

`Pane.String()` doesn't exist yet — replace that sanity check with a direct field comparison instead:

```go
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
```

Run: `go test ./internal/ui/... -v`
Expected: PASS across all sub-packages and the root `ui` package (the root package won't compile until Task 16 deletes the old files that reference removed identifiers like `screenMenu` — run this from a state where Task 16's deletions have already happened, or run `go vet ./internal/ui` first to confirm which old files still reference removed types before deleting them in Task 16. In practice, run Tasks 15 and 16 back-to-back: write `program.go` and `ui_test.go` here, then immediately delete the old files in the same working session before attempting `go test`, since the package won't build with both old and new model definitions present.)

- [ ] **Step 3: Commit**

```bash
git add internal/ui/program.go internal/ui/ui_test.go
git commit -m "ui: rewrite root model around the two-pane shell"
```

---

### Task 16: Remove the old screen-based files and finish integration

**Files:**
- Delete: `internal/ui/forms.go`, `internal/ui/library.go`, `internal/ui/log_view.go`, `internal/ui/menu.go`, `internal/ui/picker.go`, `internal/ui/settings.go`, `internal/ui/theme_picker.go`, `internal/ui/themes.go`
- Modify: `internal/ui/progress_test.go` (delete — its tests either moved to `workspace/downloads_test.go` in Task 10, or tested removed screen-view functions)
- Modify: `internal/ui/progress.go` (delete — `renderProgressBar` moved to `workspace/downloads.go` in Task 10)

**Interfaces:** none new — this task only removes dead code now superseded by Tasks 1–15.

- [ ] **Step 1: Delete the superseded files**

```bash
git rm internal/ui/forms.go internal/ui/library.go internal/ui/log_view.go \
       internal/ui/menu.go internal/ui/picker.go internal/ui/settings.go \
       internal/ui/theme_picker.go internal/ui/themes.go \
       internal/ui/progress.go internal/ui/progress_test.go
```

- [ ] **Step 2: Build and fix any remaining references**

Run: `go build ./...`
Expected: no errors. If any old identifier (`screenMenu`, `banner`, `mainMenu`, etc.) is still referenced, it means a leftover file wasn't deleted or `program.go` still calls something removed — grep for it and fix:

```bash
grep -rn "screenMenu\|screenInput\|screenSettings\|screenLibrary\|screenDownloadLog\|screenRunning\|screenPicker\|screenResult\|mainMenu\b" internal/ui/*.go
```

Expected: no matches outside of comments.

- [ ] **Step 3: Run the full test suite and lint**

Run: `go test ./... -v`
Expected: PASS for every package, including `internal/ui`, `internal/ui/style`, `internal/ui/shell`, `internal/ui/workspace`, `internal/ui/suggest`, `internal/ui/overlay`.

Run: `make lint`
Expected: no findings. If `golangci-lint` flags unused parameters or exported-without-doc-comment issues introduced in Tasks 1–15, fix them in place (this is expected cleanup, not a plan gap — e.g. add doc comments to any exported type Step 3 of earlier tasks left undocumented).

Run: `make test`
Expected: same as `go test ./... -v`, confirms the Makefile target matches.

- [ ] **Step 4: Manual smoke test**

Use the `run` skill (or `go run ./cmd/shadowbox` / the project's existing entry point — check `cmd/` or `main.go` for the exact invocation) to launch the app in a real terminal and verify by hand:

- The banner, Nav, and Content panes render with rounded borders at a normal terminal size (≥100×24).
- `Tab` toggles focus between Nav and Content; the focused pane's border highlights in the theme's accent colour.
- `1`-`8` jump directly to each section; `Enter` on a Nav item focuses Content.
- Typing in Search shows history/library suggestions; submitting a query shows results; selecting a result switches to Downloads and starts a real download.
- Library drill-down (Artist → Album → Track) works, `←` from the Artist column returns focus to Nav, and typing filters the current column.
- Settings toggles a boolean, edits a text field, and opens the theme picker with live preview; changing the theme updates border colours across all panes immediately.
- `?` opens the help overlay; `Esc` closes it. `q` quits (with an active download running, confirm behavior matches the pre-redesign confirm-on-quit expectation — if this plan's `handleKey` doesn't yet implement the confirmation prompt, note it as a follow-up rather than silently shipping a regression).

If the manual pass finds the missing "confirm quit while a download is running" behavior from the old menu screen (`handleKey`'s old comment: "Global quit on the menu screen... confirm if download running"), add it as a fast-follow: a `confirmQuit bool` on `model`, set when `q` is pressed while `m.taskCancel != nil`, requiring a second `q` or `y` to actually quit. This was flagged in the original spec's Nav table ("Exit: q from any pane (confirm if download running)") — implement it before merging if the manual pass confirms it's missing.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "ui: remove superseded screen-based files, finish shell integration"
```

---

## Testing Summary

| Layer | Tests |
|-------|-------|
| `style` | Theme fallback, theme registry consistency, banner rendering |
| `shell` | Pane toggle, layout computation and clamping, render composition |
| `workspace` | Section registry, message constructors, per-workspace Activate/Update/View behavior (Search, Input, Library, Downloads, Log, Settings) |
| `suggest` | History add/dedupe/match/persist, filter, library fuzzy match, clipboard URL shape detection |
| `overlay` | Picker/theme cursor movement and view rendering |
| `ui` (root) | Nav→Content activation, digit jump, Tab toggle, help overlay, quit |

All changes must pass `make lint` and `make test` (per Global Constraints).

---

## Not Covered By This Plan

Per spec §Non-Goals: audio playback/decoding, the visualizer, multi-download queue, album art rendering, and the "confirm quit while downloading" behavior are out of scope unless flagged as a fast-follow in Task 16 Step 4.
