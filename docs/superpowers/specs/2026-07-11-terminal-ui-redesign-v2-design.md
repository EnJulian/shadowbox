# Shadowbox Terminal UI Redesign v2 — Design Spec

**Date:** 2026-07-11
**Status:** Approved (brainstorming)
**Supersedes:** `2026-07-01-terminal-ui-redesign-design.md` (proposed a three-pane lazygit-style shell with a dedicated details pane; revised here to a simpler two-pane shell inspired by Claude Code's bordered-box aesthetic, per fresh brainstorming)
**Scope:** Phase 1 implementation plan follows this document.

---

## Summary

Replace Shadowbox's single-column, full-screen-swapping Bubble Tea UI with a persistent **two-pane shell** (Nav | Content), styled with rounded-corner bordered boxes throughout — the visual signature borrowed from Claude Code's terminal UI. Keep the centered ASCII banner at the top. Migrate all existing features into the new layout. A playback bar region is reserved (present but height-0/inactive) from day one so the upcoming playback update slots in without another relayout.

---

## Goals

- Persistent, bordered Nav + Content shell instead of full-screen swaps.
- Left sidebar nav (bordered box) with all sections always visible; content swaps inside the Content box.
- Contextual details (track info, file size, progress stats) render inline within the Content pane — no separate details pane.
- Seamless keyboard navigation: `Tab` always cycles panes, `Esc` always backs out one level, nothing is a dead end.
- Local autocomplete/suggestions on Search, URL/Playlist, and Library inputs.
- Reserved (currently inactive) playback bar slot between Content and the status bar.
- Curated, refined color themes redesigned to read well with the new bordered layout.
- Preserve all existing functionality: search+download, URL/playlist download, library browse, enhance, download log, settings.

## Non-Goals (this phase)

- Audio playback, seek bar and its live logic (slot is reserved structurally; behavior ships in a later update).
- Visualizer mode (may be considered after playback ships; not committed here).
- Multi-download queue (UI slot only; single active job as today).
- Album art rendering (placeholder inline in Library, if any).
- Nerd Font icon glyphs (deliberately out — see Theming).

---

## Design Decisions

| Decision | Choice |
|----------|--------|
| Navigation model | Persistent two-pane shell (Nav \| Content) |
| Details display | Inline within Content pane (no separate details pane) |
| Border style | Rounded corners (`╭─╮` family), matching Claude Code |
| Glyphs | Plain Unicode only — no Nerd Font dependency |
| Nav style | Left sidebar list |
| Theming | Curated set of refined themes (redesigned, not a 1:1 port of the old 5) |
| Playback | Reserved bottom bar slot, inactive this phase |
| Autocomplete | Search history + library fuzzy match, URL clipboard hint, Library type-ahead filter |
| Banner | Full ASCII art, horizontally centered, on every screen |

---

## Layout

```
╭──────────────────────────────────────────────────────────────────────────╮
│                        SHADOWBOX  (centered ASCII)                       │
│                    Music Acquisition Console                             │
├─────────────┬─────────────────────────────────────────────────────────┤
│ ╭─────────╮ │ ╭───────────────────────────────────────────────────────╮ │
│ │ Nav     │ │ │ Content                                               │ │
│ │         │ │ │                                                       │ │
│ │ › Search│ │ │  [active section renders here, incl. inline details] │ │
│ │   URL   │ │ │                                                       │ │
│ │   Playl.│ │ │                                                       │ │
│ │   Library│ │ │                                                      │ │
│ │   Downl.│ │ │                                                       │ │
│ │   Enhance│ │ │                                                      │ │
│ │   Log   │ │ │                                                       │ │
│ │   Settings│ │ │                                                     │ │
│ ╰─────────╯ │ ╰───────────────────────────────────────────────────────╯ │
├─────────────┴─────────────────────────────────────────────────────────┤
│  ▶ (playback bar — reserved region, height 0 / hidden this phase)       │
├──────────────────────────────────────────────────────────────────────────┤
│ SEARCH │ Tab: switch pane │ /: search │ Enter: select │ q: quit          │
╰──────────────────────────────────────────────────────────────────────────╯
```

- **Nav:** fixed-width bordered box, ~18 columns.
- **Content:** fills remaining width, bordered box.
- **Playback bar:** height 0 (hidden) this phase. Shell layout function always accounts for it, so a later phase sets height to 2 (track title, seek bar, elapsed/total) without touching the rest of the layout math.
- **Status bar:** single unboxed line at the very bottom, showing keys for whatever pane/control is focused — mirrors Claude Code's unboxed hint line beneath its input box.
- **Minimum terminal size:** 100×24. Below that: drop Content's inline padding first; if still too narrow, collapse Nav to a top tab bar (rare fallback).

---

## Focus & Keyboard Model

Seamless navigation is a first-class requirement — no dead ends, always something works.

| Action | Behavior |
|--------|----------|
| `Tab` / `Shift+Tab` | Cycle focus Nav → Content → Nav. Always works. |
| `↑`/`↓` or `j`/`k` | Move within the focused pane (nav items, list rows, results) |
| `1`–`8` | Jump directly to a nav section from anywhere + auto-focus Content |
| `Enter` on nav item | Activate section, auto-focus Content on that section's natural first control |
| `→`/`l` from Nav | Same as Enter — jump into Content |
| `←`/`h` from Content | Return focus to Nav; section stays active, no state resets |
| `Esc` | Universal back/cancel — closes overlay, cancels inline edit, or returns focus to Nav; never quits by accident |
| `/` | Jump straight into the Search query input from anywhere |
| `?` | Toggle keybinding help overlay |
| `q` | Quit (confirms if a download is running) |
| `Ctrl+C` | Force quit |

**Library column exception:** while Content is focused on the Library 3-column view, `←`/`→` move between Artist/Album/Track columns first; only falls through to Nav when already at the leftmost column and `←` is pressed again. `Tab` always switches panes regardless of column position.

The status bar always reflects the **focused pane's** available keys.

---

## Suggestions & Autocomplete

- **Search input:** inline suggestion list below the input, blending two sources — recent local search history (small local history file, most-recent-first) and fuzzy matches against the existing library, visually tagged (e.g. "in library") so the user can tell "searched before" from "already have it." `↓` from the input moves into suggestions, `Enter` accepts, `Esc` dismisses.
- **URL / Playlist input:** on focus, if the clipboard currently holds something that looks like a YouTube/Bandcamp/playlist URL, show an inline hint ("Paste from clipboard: `youtu.be/…` — Tab to accept"). Never auto-fills without explicit confirmation.
- **Library browse:** typing while a column is focused jumps to and live-filters matching entries in that column (type-ahead / fuzzy-find). `Esc` or backspacing to empty clears the filter.

All suggestion data is local-only — no new network calls beyond the existing search itself.

---

## Nav Sections & Content

| # | Section | Content pane shows |
|---|---------|---------------------|
| 1 | Search | Query input (with suggestions) + browsable results list. Selecting a result shows duration/uploader/source inline. `Enter` starts the download and auto-switches Nav to Downloads. |
| 2 | URL | URL input (with clipboard hint) + confirm. Submit starts download, switches to Downloads. |
| 3 | Playlist | Same as URL, playlist-flavored validation/copy. |
| 4 | Library | Three columns (Artist / Album / Track) side-by-side inside the Content box, with type-ahead filtering. Selected track's path/format/size shown as a header line above the columns. |
| 5 | Downloads | Active job: heading, progress bar, spinner, stage text (reuses `progress.Update`, `renderProgressBar`, existing spinner). Single active job (queue row present but empty). |
| 6 | Enhance | Directory path input; submit runs `EnhanceDir`, switches to Downloads. Matching file count shown inline once a valid path is entered. |
| 7 | Log | Existing scrollable download log, rendered inside the bordered Content box. `↑↓`, PgUp/PgDn, Home/End, `r` refresh. |
| 8 | Settings | Existing settings list; toggles inline, text/secret settings use inline `textinput`, theme picker opens as an overlay with live preview. |

`Enter`-to-activate-and-auto-focus-Content and per-section auto-focus targets (Search → query input, Library → Artist column first item, etc.) carry over from current behavior, now happening inside a persistent shell instead of full-screen transitions.

**Overlays** (metadata/URL disambiguation picker, result toast, `?` help, theme picker) render on top of the shell, re-skinned with the rounded-border style, and block pane navigation while open.

---

## Theming

- **Glyphs:** plain Unicode only (box-drawing `│─╭╮╰╯`, common symbols `▶ ● ✓ ✗`) — no Nerd Font dependency, renders correctly in any monospace font.
- **Border style:** rounded corners throughout (Nav box, Content box, overlays).
- **Palette:** redesign the existing theme concept into a smaller, curated set of refined themes (roughly 4–5) rather than a 1:1 port of the old 5 — each tuned so border color, active-nav highlight, and status colors stay legible against the new box-heavy layout. Candidates to refine from the current set: Hacker (green), Cyberpunk (magenta/violet), Ocean (blue/cyan), Fire/Ember (amber/orange), Matrix (green, distinct from Hacker). Exact palette values are an implementation-time decision, validated visually against the new layout — not fixed by this spec.
- **Semantic roles** (mapped per theme):

| Role | Usage |
|------|-------|
| Branding | Banner |
| Selection/Active | Active nav item, focused pane border, cursor |
| Primary text | Body |
| Secondary/Muted | Inactive nav, hints, unfocused pane border |
| Success | Completed downloads |
| Warning | Warnings |
| Danger | Errors |

- Theme picker overlay retained with existing live-preview behavior.

---

## Component Architecture

### Directory structure

```
internal/ui/
├── program.go          # tea entry, root model, message routing
├── shell/
│   ├── shell.go        # layout, pane sizes, borders, render tree
│   ├── focus.go        # FocusPane enum + Tab/arrow routing
│   └── banner.go       # centered ASCII art + subtitle
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
├── suggest/
│   ├── history.go       # local search-history persistence
│   └── suggest.go       # fuzzy match against library + history, clipboard URL detection
├── overlay/
│   ├── picker.go
│   ├── help.go
│   ├── result.go
│   └── theme.go
├── playback/
│   └── playback.go      # stub: State{Active bool}, always false this phase
├── themes.go
└── progress.go
```

### Workspace interface

```go
type Workspace interface {
    Update(msg tea.Msg) (Workspace, tea.Cmd)
    View(width, height int) string
    PreferredFocus() FocusTarget
}
```

Root model calls `PreferredFocus()` when the user activates a nav item.

### Message flow

- `KeyMsg` → shell focus routing → active workspace.
- `progressMsg` → update Downloads workspace; auto-switch nav to Downloads.
- `promptRequestMsg` → open picker overlay.
- `taskDoneMsg` → result overlay; stay on current section.

### State ownership

| State | Owner |
|-------|-------|
| `cfg`, `app`, `theme`, dimensions | Root model |
| `focusPane`, `activeSection` | Shell |
| Section-specific state | Each workspace |
| Search/URL suggestion state | `suggest` package, surfaced through Search/URL workspaces |
| Progress, task cancel | Root + Downloads workspace |
| Overlay stack | Root model |

### Playback stub (this phase)

```go
// playback/playback.go
type State struct {
    Active bool // always false this phase
}

// shell.go: PlaybarHeight = 0 this phase, becomes 2 once playback ships
```

---

## Implementation Phases

### Phase 1 — Shell migration (this spec)

1. Shell skeleton (banner, Nav/Content panes, borders, focus routing, status bar, reserved playbar region at height 0)
2. Nav sidebar with auto-focus on section switch
3. Settings + Log workspaces
4. Library 3-column refactor + type-ahead filter
5. Search with browsable results + suggestion list (history + library fuzzy match)
6. URL / Playlist / Enhance workspaces + clipboard hint
7. Downloads progress migration
8. Overlays (picker, result, help, theme picker) re-skinned with rounded borders
9. Curated theme palette redesign
10. Remove old `screenMenu` and per-screen full-page views

### Phase 2 — Playback (future update, not scoped here)

- Backend + IPC approach to be decided when that update starts.
- Playbar grows from height 0 to height 2; track title, seek bar, elapsed/total.
- Library: play a selected track.

---

## Testing

| Layer | Tests |
|-------|-------|
| `shell/focus.go` | Arrow/Tab routing; Library column edge → Nav |
| `workspace/search.go` | Result navigation, suggestion list, empty query |
| `workspace/library.go` | Column drill-down, type-ahead filter |
| `suggest/suggest.go` | History ranking, library fuzzy match, clipboard URL detection |
| `ui_test.go` | Nav Enter → workspace focus; section switch preserves shell state |

All changes must pass `make lint` and `make test`.

---

## Migration Notes

- Remove duplicate banner rendering from individual view functions; shell renders the banner once.
- Replace the `screen` enum's full-screen swaps with `activeSection` + overlay stack.
- Existing theme picker live-preview behavior preserved in the overlay, now against the curated palette.
- Existing filesystem hierarchy (`Artist/Album/Track`) and `libState` logic carry forward into the 3-column workspace.

---

## Inspiration

- **Claude Code** — rounded bordered boxes, plain-Unicode glyphs, seamless keyboard focus model, unboxed status/hint line.
- **lazygit / btop** — persistent pane layout, contextual status bar.
- **ncmpcpp / cmus** — multi-column library browsing, fast keyboard-first navigation.
