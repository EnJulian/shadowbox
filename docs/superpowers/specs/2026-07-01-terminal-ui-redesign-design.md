# Shadowbox Terminal UI Redesign — Design Spec

**Date:** 2026-07-01  
**Status:** Superseded by `2026-07-11-terminal-ui-redesign-v2-design.md` (three-pane shell replaced with a simpler two-pane, Claude Code-inspired bordered layout)  
**Scope:** Phase 1 implementation plan follows this document.

---

## Summary

Replace Shadowbox's single-column, full-screen-swapping Bubble Tea UI with a persistent three-pane shell inspired by lazygit, ncmpcpp, and spotatui. Keep the centered ASCII banner at the top. Migrate all existing features into the new layout in Phase 1. Playback, a persistent playbar with seeker, and a full-width Library visualizer mode follow in Phases 2 and 3.

---

## Goals

- Preserve the retro terminal aesthetic and centered ASCII art banner.
- Replace the main menu with a persistent pane-based interface.
- Make navigation faster and more intuitive.
- Display useful contextual information at all times (details pane + status bar).
- Support browsable YouTube search results before download.
- Architect for future playback and visualizer without rewriting the shell.

## Non-Goals (Phase 1)

- Audio playback, seek bar, or visualizer implementation.
- Multi-download queue (UI slot only; single active job like today).
- Album art rendering (placeholder in details pane).
- Replacing or removing existing CLI commands.

---

## Design Decisions

| Decision | Choice |
|----------|--------|
| Navigation model | Three-pane shell (Nav \| Workspace \| Details) |
| Library visualizer | Full-width workspace mode (Phase 3); details pane hidden |
| Implementation phasing | A: Shell first → Playback → Visualizer |
| Architecture | Composed shell with sub-packages (Approach 2) |
| Search UX | Browse results before download |
| Banner | Full ASCII art, horizontally centered, on every screen |

---

## Layout

### Persistent shell (Phase 1)

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         SHADOWBOX  (centered ASCII)                      │
│                    Music Acquisition Console                             │
├─────────────┬────────────────────────────────────────┬───────────────────┤
│ Navigation  │  Workspace                             │  Details          │
│             │                                        │                   │
│ › Search    │  [content for active section]          │  [contextual]     │
│   URL       │                                        │                   │
│   Playlist  │                                        │                   │
│   Library   │                                        │                   │
│   Downloads │                                        │                   │
│   Enhance   │                                        │                   │
│   Log       │                                        │                   │
│   Settings  │                                        │                   │
├─────────────┴────────────────────────────────────────┴───────────────────┤
│ SEARCH │ Tab: panes │ /: focus search │ Enter: action │ q: quit          │
└──────────────────────────────────────────────────────────────────────────┘
```

### Pane sizing

- **Nav:** 18 columns fixed.
- **Details:** 28 columns fixed (hidden when empty or in Phase 3 visualizer mode).
- **Workspace:** remaining width.
- **Minimum terminal:** 100×24. Below minimum: hide details pane first; degrade nav to a compact top tab bar if needed.

### Phase 2 playbar slot

One reserved row between workspace and status bar (height 0 in Phase 1, height 2 in Phase 2):

```
▶ Feather — Nujabes          [====●────]  2:14 / 4:32
```

Playbar remains visible on all screens when playback is active.

### Phase 3 visualizer mode

Triggered by `v` in Library while a track is playing (Phase 3). Details pane hides; workspace expands to full width. `v` or `Esc` returns to three-column library browse; playback and playbar continue.

---

## Navigation sidebar

| # | Item | Workspace | Details pane |
|---|------|-----------|--------------|
| 1 | Search | Query input + results list | Source, duration, uploader |
| 2 | URL | URL input + confirm | Detected source, hostname |
| 3 | Playlist | Playlist URL input | Source, track count (when known) |
| 4 | Library | 3-column Artists / Albums / Tracks | Path, format, file size |
| 5 | Downloads | Active job + queue slot | Stage, progress, speed, ETA |
| 6 | Enhance | Directory path input | File count preview |
| 7 | Log | Scrollable download log | Line count, last error |
| 8 | Settings | Settings list | Description + current value |

**Exit:** `q` from any pane (confirm if download running). No dedicated nav item.

**Active nav item:** cyan. Inactive items: gray.

---

## Focus & pane switching

Seamless navigation is a first-class requirement.

| Action | Behavior |
|--------|----------|
| `Tab` / `Shift+Tab` | Cycle focus: Nav → Workspace → Details → Nav (always works) |
| `←` / `→` and `h` / `l` | Switch panes horizontally |
| `↑` / `↓` and `j` / `k` | Navigate within the focused pane only |
| `Enter` on nav item | Activate section **and** auto-focus Workspace |
| `→` from nav | Same as Enter — jump into Workspace |
| `←` from workspace | Return focus to Nav (section stays active) |

### Library column exception

When Workspace is focused on the Library 3-column view:

- `←` / `→` move between Artist / Album / Track columns.
- `←` from the leftmost (Artist) column returns focus to Nav.
- `Tab` always switches panes regardless of column position.

### Auto-focus targets on nav activation

| Section | Workspace focus lands on |
|---------|--------------------------|
| Search | Query input |
| URL / Playlist / Enhance | Respective input field |
| Library | Artist column, first item |
| Downloads | Current job area |
| Log | Log viewport |
| Settings | Settings list |

---

## Keyboard reference

### Global (no modal open)

| Key | Action |
|-----|--------|
| `1`–`8` | Jump to nav item + auto-focus Workspace |
| `Tab` / `Shift+Tab` | Cycle pane focus |
| `←` `→` `h` `l` | Switch panes (Library: columns when in Workspace) |
| `?` | Toggle keybinding help overlay |
| `q` | Quit (confirm if download running) |
| `Ctrl+C` | Force quit |

Status bar shows keys for the **focused pane only**.

---

## Workspace specifications

### Search

Interactive search with browsable results before download.

```
Query:
> nujabes_

Results
► Feather — Nujabes          3:42
  Reflection Eternal         4:01
  Luv(sic) Pt.3              4:15
```

| Key (workspace focused) | Action |
|-------------------------|--------|
| `/` | Focus query input |
| `Enter` (in input) | Run `SearchYouTube` async |
| `↑↓` | Move through results |
| `Enter` (on result) | Start download for selected URL |

**Flow:** Enter on result → `startTask` with selected URL (skip in-task YouTube picker) → nav switches to Downloads → Workspace shows progress.

**Details pane:** Updates as result cursor moves (title, uploader, duration, source).

### URL & Playlist

Single URL input each. Submit starts download task and switches to Downloads.

### Library

Three columns visible simultaneously. Phase 1 is browse-only.

| Key | Action |
|-----|--------|
| `←→` | Move active column |
| `↑↓` | Move cursor in active column |
| `Enter` | Drill into selection |
| `Space` | Play track *(Phase 2)* |
| `v` | Visualizer full-width *(Phase 3)* |

Existing filesystem hierarchy (`Artist/Album/Track`) and `libState` logic carry forward, refactored into 3-column simultaneous view.

### Downloads

Home for active jobs. Any download start auto-switches nav to Downloads and focuses Workspace.

Reuses `progress.Update`, `renderProgressBar`, and spinner. Single active job in Phase 1 (queue row UI present but empty).

| Key | Action |
|-----|--------|
| `c` | Cancel running download |
| `l` | Jump to Log workspace |

### Enhance

Directory path input. Submit runs `EnhanceDir` → Downloads progress view. Details shows matching file count preview.

### Log

Existing scrollable download log (`applog`). Keys: `↑↓`, `PgUp`/`PgDn`, `Home`/`End`, `r` refresh.

### Settings

Existing settings list and edit flows, rendered inside Workspace. Details shows highlighted setting description + value.

- Toggle settings: Enter toggles inline.
- Text/secret settings: inline `textinput` in Workspace.
- Theme picker: overlay modal (existing live preview behavior).

`Esc` from edit returns to settings list; does not exit the shell.

---

## Overlays

Render on top of the shell; block pane navigation while open.

| Overlay | Trigger | Dismiss |
|---------|---------|---------|
| Metadata picker | Pipeline disambiguation (non-Search paths) | Select option |
| Result toast | Task complete/fail | Any key |
| Help | `?` | `?` or `Esc` |
| Theme picker | Settings → Theme | Enter / Esc |

---

## Color palette

Semantic roles mapped onto existing theme system (`hacker`, `matrix`, etc.):

| Role | Color | Usage |
|------|-------|-------|
| Branding | Magenta | Banner |
| Selection | Cyan | Active nav, borders, cursor |
| Primary text | White | Body |
| Secondary | Gray | Inactive nav, hints |
| Success | Green | Completed downloads |
| Warning | Yellow | Warnings |
| Error | Red | Errors |

---

## Component architecture

### Directory structure

```
internal/ui/
├── program.go          # tea entry, root model, message routing
├── shell/
│   ├── shell.go        # layout, pane sizes, borders, render tree
│   ├── focus.go        # FocusPane enum + Tab/arrow routing
│   └── banner.go       # centered ASCII art + subtitle
├── nav/
│   └── nav.go          # sidebar list, section switching
├── details/
│   └── details.go      # section-specific detail renderers
├── status/
│   └── status.go       # context key bar + stats
├── workspace/
│   ├── workspace.go    # interface + registry
│   ├── search.go
│   ├── url.go
│   ├── playlist.go
│   ├── library.go
│   ├── downloads.go
│   ├── enhance.go
│   ├── log.go
│   └── settings.go
├── overlay/
│   ├── picker.go
│   ├── help.go
│   ├── result.go
│   └── theme.go
├── playback/
│   └── playback.go     # Phase 1: stub interface only
├── themes.go
└── progress.go
```

### Workspace interface

```go
type Workspace interface {
    Update(msg tea.Msg) (Workspace, tea.Cmd)
    View(width, height int) string
    Details() string
    PreferredFocus() FocusTarget
}
```

Root model calls `PreferredFocus()` when user activates a nav item.

### Message flow

- `KeyMsg` → shell focus routing → active sub-component.
- `progressMsg` → update Downloads workspace; auto-switch nav to Downloads.
- `promptRequestMsg` → open picker overlay.
- `taskDoneMsg` → result overlay; stay on current section.

### State ownership

| State | Owner |
|-------|-------|
| `cfg`, `app`, `theme`, dimensions | Root model |
| `focusPane`, `activeSection` | Shell |
| Section-specific state | Each workspace |
| Progress, task cancel | Root + Downloads workspace |
| Overlay stack | Root model |

### Phase 2/3 stubs (Phase 1)

```go
// playback/playback.go
type State struct {
    Active bool // always false in Phase 1
}

// shell.go: PlaybarHeight = 0 (Phase 1), 2 (Phase 2)
// shell.go: DetailsWidth = 0 when visualizer active (Phase 3)
```

---

## Implementation phases

### Phase 1 — Shell migration (this spec)

1. Shell skeleton (banner, panes, focus, status bar)
2. Nav with auto-focus on section switch
3. Settings + Log workspaces
4. Library 3-column refactor
5. Search with browsable results
6. URL / Playlist / Enhance workspaces
7. Downloads progress migration
8. Overlays (picker, result, help)
9. Remove old `screenMenu` and per-screen full-page views

### Phase 2 — Playback

- mpv subprocess with JSON IPC (recommended backend)
- Persistent playbar with seeker on all screens
- Library: Space to play selected track
- Playbar shows track title, progress, elapsed/total

### Phase 3 — Visualizer

- `v` in Library toggles full-width spectrum/bars
- Exit visualizer returns to 3-column browse; playback continues
- FFT via audio output monitor (similar to spotatui/ncmpcpp pattern)

---

## Testing

| Layer | Tests |
|-------|-------|
| `shell/focus.go` | Arrow/Tab routing; Library column edge → nav |
| `workspace/search.go` | Result navigation, empty query |
| `workspace/library.go` | Column drill-down, breadcrumbs |
| `ui_test.go` | Nav Enter → workspace focus; section switch preserves shell |

All changes must pass `make lint` and `make test`.

---

## Migration notes

- Remove duplicate `banner` rendering from individual view functions; shell renders banner once.
- Replace `screen` enum full-screen swaps with `activeSection` + overlay stack.
- Search path bypasses in-task `youtubePrompt` picker; URL/metadata pickers remain for other flows.
- Existing theme picker live preview behavior preserved in overlay.

---

## Inspiration

- **lazygit / btop** — pane layout, status bar, contextual details
- **ncmpcpp** — multi-column library, visualizer as dedicated mode
- **cmus** — fast keyboard nav, minimal chrome
- **spotatui** — persistent playbar across panel switches, `v` visualizer toggle
