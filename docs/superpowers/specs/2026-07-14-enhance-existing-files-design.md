# Enhance Existing Files — Folder Picker + Unattended Batch Matching — Design

## Goal

"Enhance Existing Files" currently asks the user to paste a directory path into a raw text field, and silently turns into a one-decision-per-song chore for any folder with more than a couple of ambiguous matches, since each file's metadata match pops an interactive "pick a candidate" prompt. This design replaces the text prompt with a folder browser, and makes batch matching fully unattended by auto-accepting the top-ranked candidate instead of prompting per file.

## Scope

**In scope:**
- A new folder-picker screen for the "Enhance Existing Files" menu entry, replacing the free-text `openInput("enhance", ...)` prompt.
- An `Options.AutoAcceptTopMatch` flag, set by `EnhanceDir`, that makes `choose()` skip the interactive prompt (both the TUI picker and CLI stdin) and return the top-ranked candidate.
- Tests for both.

**Out of scope:**
- Metadata match quality/correctness (iTunes-vs-MusicBrainz ordering, filtering logic) — explicitly deferred pending the user's own batch testing.
- Any change to Search & Download or Playlist flows — both keep their existing interactive picker.
- A live audio-file count per folder in the picker (deferred; adds a recursive walk on every render for no confirmed need).
- A CLI flag to opt back into interactive per-file prompts for `shadowbox enhance` (not requested; `EnhanceDir` behavior is now unattended-by-default for both entry points).

## Architecture

### Folder picker

A new screen, `screenEnhancePicker`, backed by a new `enhancePickerState` struct (mirrors `libState` in `internal/ui/library.go`, but browses arbitrary directories instead of the fixed Artist/Album/Track hierarchy):

```go
type enhancePickerState struct {
	cwd     string   // absolute path currently being browsed
	cursor  int
	entries []string // subdirectory names in cwd, sorted, dotfiles excluded
	filter  string
}
```

- Opens at `cfg.MusicDirectory`.
- The rendered list always shows a pinned `[ Use this folder ]` entry first (not part of `entries`, not subject to `filter`), followed by `visible()` — the filtered subdirectory list, reusing the exact substring-filter pattern `libState.visible()` already implements.
- **Enter** on `[ Use this folder ]`: dispatches the Enhance task for `cwd` (same `startTask`/`EnhanceDir` call `forms.go`'s `"enhance"` case already makes, just sourced from `cwd` instead of a typed value) and switches to `screenRunning`.
- **Enter** on a subdirectory: descends (`cwd = filepath.Join(cwd, selected)`, reset `cursor`/`filter`, re-list).
- **Esc**: goes to the parent directory (`filepath.Dir(cwd)`), reset `cursor`/`filter`, re-list — unless `cwd` has no parent (filesystem root), in which case it returns to the main menu. This makes going up and canceling the same intuitive gesture used throughout the rest of the app (mirrors `libraryBack`'s per-level Esc semantics), generalized to arbitrary depth via a root check instead of a fixed 3-level `switch`.
- Typing filters visible subdirectories, same as Library's type-ahead (backspace edits, cursor resets on every keystroke).
- No `..` list entry — Esc already covers "go up," avoiding a redundant, sometimes-confusing row.

Listing a directory (`listSubdirs`) is a thin, non-recursive `os.ReadDir` returning sorted directory names with dotfiles excluded — cheap enough to call on every navigation without caching.

### Unattended batch matching

`app.Options` (in `internal/app/app.go`) gains one field:

```go
AutoAcceptTopMatch bool
```

`choose()` (in `internal/app/prompt.go`) gets one new branch, checked after the existing 0/1-candidate cases and before falling through to `opts.Select`/`chooseCLI`:

```go
func choose(ctx context.Context, opts Options, req PromptRequest) (int, error) {
	n := len(req.Options)
	switch n {
	case 0:
		return -1, fmt.Errorf("no options to choose from")
	case 1:
		return 0, nil
	}
	if opts.AutoAcceptTopMatch {
		return 0, nil
	}
	if opts.Select != nil {
		return opts.Select(ctx, req)
	}
	return chooseCLI(ctx, req)
}
```

`EnhanceDir` (in `internal/app/enhance.go`) sets `opts.AutoAcceptTopMatch = true` once, before its file loop — `Options` is already passed by value into both `EnhanceDir` and each `enhanceFile` call, so this is a local copy with no shared-mutable-state risk. Because `matchFromITunes`/`matchFromMusicBrainz` → `enrich` → `enhanceFile` → `EnhanceDir` is the only call chain that sets this flag, `Run`/`RunPlaylist` (Search & Download, Playlist) are untouched and keep prompting interactively as before. This affects both the TUI (`internal/ui/forms.go`'s `"enhance"` case → replaced by the picker) and the CLI `shadowbox enhance` command (`internal/cmd/enhance.go`), since both call `EnhanceDir` directly.

## Components

- **`internal/ui/enhance_picker.go`** (new): `enhancePickerState` struct, `(l enhancePickerState) visible() []string`, `(m model) openEnhancePicker() (tea.Model, tea.Cmd)`, `(m model) updateEnhancePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd)`, `(m model) viewEnhancePicker() string`, `listSubdirs(dir string) []string`.
- **`internal/ui/program.go` changes**: `screenEnhancePicker` added to the `screen` const block; `model.enhancePicker enhancePickerState` field; `screenEnhancePicker` cases added to `handleKey`'s and `View`'s screen switches.
- **`internal/ui/menu.go` change**: the `"Enhance Existing Files"` menu case calls `m.openEnhancePicker()` instead of `m.openInput("enhance", ...)`.
- **`internal/ui/forms.go` change**: the `"enhance"` case is removed from `submitInput` (dead code once the picker replaces the text-input path) — the same `a.EnhanceDir(...)` call moves into `enhance_picker.go`'s Enter-on-`[ Use this folder ]` handler.
- **`internal/app/app.go` change**: `Options.AutoAcceptTopMatch bool` field added.
- **`internal/app/prompt.go` change**: `choose()` gains the `AutoAcceptTopMatch` branch described above.
- **`internal/app/enhance.go` change**: `EnhanceDir` sets `opts.AutoAcceptTopMatch = true` before its loop.

## Data Flow

1. User selects "Enhance Existing Files" from the main menu → `openEnhancePicker()` sets `m.enhancePicker = enhancePickerState{cwd: cfg.MusicDirectory, entries: listSubdirs(cfg.MusicDirectory)}`, `m.screen = screenEnhancePicker`.
2. User navigates (Enter descends, Esc goes up) until the desired folder is showing, optionally typing to filter.
3. User selects `[ Use this folder ]` → the same `startTask("Enhancement", ...)` machinery already used by `forms.go` runs, calling `a.EnhanceDir(ctx, m.enhancePicker.cwd, true, []string{"opus","mp3","m4a","flac"}, false, opts)`.
4. Inside `EnhanceDir`, `opts.AutoAcceptTopMatch = true` is set once; each file's `enhanceFile` → `enrich` → `matchFromITunes`/`matchFromMusicBrainz` → `choose()` call now returns the top candidate immediately, with no picker screen and no stdin read — the whole batch runs to completion unattended, exactly as `EnhanceDir`'s existing per-file logging (`applog.Successf("ITUNES", "Matched: %s by %s", ...)`) already reports each match for later review via the Download Log.

## Error Handling

- `listSubdirs` returns `nil` on a read error (permission denied, etc.) — same fail-quiet convention as `listDirs`/`listTracks` in `library.go`; the picker just shows an empty list rather than surfacing a raw OS error, and Esc still works to back out.
- The filesystem-root check for Esc (`filepath.Dir(cwd) == cwd`) is the standard Go idiom and works identically on Unix (`/`) and Windows (`C:\`).
- No new failure modes in `choose()` — `AutoAcceptTopMatch` only removes a branch that could block, it doesn't add I/O.

## Testing

- `internal/ui`: navigating into a subdirectory and back out via Esc returns to the same listing; filtering narrows correctly; selecting `[ Use this folder ]` dispatches a task with the current `cwd`; Esc at the filesystem root falls back to the menu instead of erroring.
- `internal/app`: a `choose()` test confirming `AutoAcceptTopMatch: true` returns index 0 for a multi-candidate request without touching `opts.Select` or stdin (e.g. leave `opts.Select` as a func that fails the test if called).
- `internal/app`: an `EnhanceDir` test (or extension of the existing one) confirming a directory with multiple files and an injected multi-candidate `Select` func that would normally be invoked is never called — i.e. the batch completes without prompting.
