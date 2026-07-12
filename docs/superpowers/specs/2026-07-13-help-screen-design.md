# Help Screen — Design

## Goal

Shadowbox currently has no in-app reference for its keyboard shortcuts or a few things worth knowing (where the config file lives, which external tools are optional, what audio formats/library layouts are supported). This adds a Claude-Code-`/help`-style reference screen: press `?` from anywhere to see everything, `?`/`Esc`/`q` again to close and return to exactly where you were.

## Scope

**In scope:**
- A new `screenHelp` screen listing every keybinding in the app, verified against each screen's actual current key handling (not assumed/remembered).
- A short "important things to remember" section: config file location, external tool dependencies (+ the setup wizard), supported audio formats, Library folder-layout expectations.
- Global `?` key to open, from any screen except Library/Input/Settings-edit (same gating already used for global playback keys, since those screens need `?` as literal text — e.g. YouTube URLs like `?v=abc123`).
- `?`, `Esc`, or `q` closes the help screen and returns to `helpReturnTo` (the screen it was opened from).
- Scrolling (↑/↓, PgUp/PgDn), since the full content doesn't fit a typical terminal height — reusing the Download Log viewer's existing scroll-with-clamping pattern.

**Out of scope:**
- A real Unix `man` page (no packaging/release changes).
- A separate `docs/*.md` reference file (in-app only, per this decision).
- Any interactivity beyond scrolling — no search/filter within the help screen itself.

## Verified Keybinding Reference

Audited directly against each screen's current `case`/`tea.Key*` handling (not from memory):

- **Global:** `Ctrl+C` force quit (always) · `?` open/close help · Space/`n`/`p`/`s`/`←`/`→` playback pause/next/prev/stop/seek — active everywhere **except** Library, Input (URL/Playlist/Search/Enhance), and Settings-edit, which need those keys for text/filter entry.
- **Menu:** `↑`/`↓`/`k`/`j` navigate, `Enter` select, `q` quit (Menu-specific, not global — confirmed `handleKey` only treats `q` as quit when `m.screen == screenMenu`).
- **Library:** `↑`/`↓`/`←`/`→`/`Enter` navigate/open, any other printable character types into the live filter, `Backspace` removes the last filter character.
- **Settings:** `↑`/`↓`/`k`/`j` navigate, `Enter` edit/toggle/open, `Esc`/`q` back to menu.
- **Settings edit** (text/secret field): `Enter` save, `Esc` cancel.
- **Theme picker:** `↑`/`↓`/`k`/`j` live-preview, `Enter` apply (saves), `Esc`/`q` cancel (reverts the preview).
- **Download Log:** `↑`/`↓`/`k`/`j` scroll, `PgUp`/`PgDn`, `Home`, `End`/`G`, `r` reload, `Esc`/`q` back.
- **Search result picker:** `↑`/`↓`/`k`/`j` navigate, `Enter` select, `Esc` cancel (also cancels the running task if one is active).
- **URL / Playlist / Search / Enhance input:** `Enter` submit, `Esc` cancel/back.
- **Setup Wizard:** `↑`/`↓`/`k`/`j` navigate, `Enter` install (or toggle instructions), `Esc` done (saves and returns).

## Important Things to Remember (content)

1. Config lives at `~/.config/shadowbox/config.yaml`; edit it through Settings (auto-saved) or directly.
2. `yt-dlp`/`ffmpeg` are required; `aria2` (faster downloads) and `mpv` (in-app playback) are optional — the Setup Wizard (first run, or Settings → "Run setup wizard") can install them for you where it's safe to do so, or shows you the exact command otherwise.
3. Supported audio formats: `opus`, `mp3`, `flac`, `m4a`, `wav`, `ogg`, `webm`, `aac`.
4. Library expects `Artist/Album/Track` folders, but a track sitting directly under an Artist folder (no album) works too — selecting it plays it immediately.

## Architecture

`screenHelp` is a new screen value. `model.helpReturnTo screen` remembers where `?` was pressed from (mirrors `wizardReturnTo`'s existing pattern). The global `?` handling sits in `handleKey`, gated by the same `screenCapturesText` check that already protects playback keys — so this doesn't introduce a new gating mechanism, just reuses the established one. The reference content describes fixed facts about the app (keybindings, config location, supported formats) with no dependency on live state (e.g. it does NOT show whether mpv/aria2 are currently installed — that's what the Setup Wizard itself is for), so it's a package-level constant computed once at load time, never recomputed. `updateHelp` only adjusts a scroll offset; `viewHelp` slices the constant content by the current viewport height, matching `internal/ui/log_view.go`'s existing scroll approach (`logViewport`/`maxLogScroll`/`visibleLogLines`) — mirrored here as `helpViewport`/`maxHelpScroll`/`visibleHelpLines`.

## Components

- **`internal/ui/help.go`** (new): `helpContent []string` (a package-level var — the reference text as lines, built from the verified keybinding list and the four "remember" bullets above, split once so the same line-based scroll math `log_view.go` already uses applies directly), `(model).openHelp() (tea.Model, tea.Cmd)`, `(model).updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd)`, `(model).viewHelp() string`, `helpViewport() int`, `maxHelpScroll() int`, `visibleHelpLines(scroll int) []string`.
- **`internal/ui/program.go` changes**: `screenHelp` added to the `screen` const block; `model.helpReturnTo screen` and `model.helpScroll int` fields; a `?` case in `handleKey` (gated by `screenCapturesText`, same as the existing playback-key check) that calls `openHelp()` when not already on `screenHelp`, or closes back to `helpReturnTo` when already there; `screenHelp` cases added to `handleKey`'s and `View`'s screen switches.

## Data Flow

1. User presses `?` on any screen where `screenCapturesText` is false.
2. `handleKey` checks: is `m.screen == screenHelp` already? If so, close (restore `m.screen = m.helpReturnTo`). If not, call `openHelp()`, which stores `m.helpReturnTo = m.screen`, resets `m.helpScroll = 0`, and switches to `screenHelp`.
3. `Esc` or `q` on `screenHelp` does the same close-and-return as pressing `?` again.
4. `↑`/`↓`/`PgUp`/`PgDn` on `screenHelp` adjust `m.helpScroll`, clamped to `[0, max(0, contentLines - viewportHeight)]`.
5. `viewHelp()` renders the banner (via the existing `renderBannerWithPlayback`, so the now-playing sweep keeps animating even while help is open — background playback isn't interrupted by checking help) plus the visible slice of `helpContent()`.

## Error Handling

Nothing here does file I/O, network calls, or subprocess execution — there's no failure mode beyond ordinary bounds-clamping on scroll, which is the same defensive pattern the Download Log viewer already has.

## Testing

- `helpContent()` is tested by asserting it contains the expected key strings for each screen's shortcuts (spot-checks, not asserting the entire rendered banner+content string verbatim).
- `?` correctly opens from a non-text-capturing screen (e.g. Menu) and correctly does *not* open from a text-capturing one (e.g. typing `?` into the Library filter, or into a URL field, must add the literal character instead) — this is the single most important test given this project's history with exactly this class of bug (global keys stealing keystrokes from text entry).
- Round-trip: opening from Menu and from Settings both return to the correct origin screen on close.
- Scroll clamping at both ends (can't scroll above the top or past the bottom).
