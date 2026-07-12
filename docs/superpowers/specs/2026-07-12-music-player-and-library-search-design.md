# Music Player and Library Search — Design

## Goal

Shadowbox currently has no playback capability at all — it downloads, tags, and organizes audio, but you have to open your files in another application to listen. This adds in-app playback: select a track from Library and it plays, with a now-playing indicator built into the existing ASCII banner (a color sweep across the logo, tracking elapsed time), and background playback that keeps running no matter which screen you're on. Library also gains type-ahead search, since browsing a large collection by drilling through Artist → Album → Track alone doesn't scale.

This deliberately keeps the existing single-screen, menu-driven UI as-is otherwise — no new pane system, no persistent shell. The previous terminal-UI redesign attempt was reverted for being more complex than wanted; this design adds the player as an incremental feature on top of the current simple structure, not a rearchitecture of it.

## Scope

**In scope:**
- Playback of a selected track, with the album it belongs to auto-queued to play through afterward.
- Background playback: continues across all screens (Menu, Settings, Log, etc.), not just Library.
- Controls, available globally: play/pause, stop, skip next/previous track, seek forward/back.
- A now-playing indicator: the ASCII banner sweeps from a muted to an accent color, left to right, proportional to playback progress; the track title and elapsed/total time render on their own line directly below the banner.
- Library type-ahead search: typing while browsing any level (Artists/Albums/Tracks) filters the current list live, no separate search mode.
- Format support matching whatever the user has already downloaded: opus, mp3, flac, m4a/aac, wav, ogg, webm.

**Out of scope (not building in this pass):**
- In-app volume control (use system/OS volume).
- A dedicated "Now Playing" screen — the banner sweep + now-playing line *is* the player UI; you never leave whatever screen you're on to see it.
- Manual queue reordering, playlists, shuffle/repeat modes.
- Crossfade or gapless-playback tuning.
- Any visualizer beyond the banner sweep.
- Volume normalization / ReplayGain.

## Architecture

A new `internal/player` package owns a single long-lived `mpv` process for the whole playback session, controlled over its JSON IPC socket (`--input-ipc-server`; a Unix domain socket on Linux/macOS, a named pipe on Windows). Loading a new track is an IPC `loadfile` command sent to the same running `mpv` process — we do not spawn a new process per track. `internal/ui` polls `internal/player` for state on a render tick to drive the banner sweep and now-playing line, and forwards global key presses (space/seek/skip/stop) to it from the root `Update` loop, so controls work regardless of which screen is currently active.

`mpv` becomes a dependency required only for playback, not for the app as a whole — it's checked lazily the first time you try to play a track (mirroring how `aria2c` is already an optional dependency today), not at startup, since it has nothing to do with downloading.

## Components

### `internal/player` (new package)

- Spawns and owns the `mpv` subprocess and its IPC socket connection.
- `Load(tracks []string, startIndex int) error` — starts (or swaps) playback and sets up the queue for auto-advance; the caller (Library) is responsible for building `tracks` as the full sorted list of audio files in the selected track's album directory.
- `State() State` — returns `{Title, Artist, Elapsed, Duration time.Duration, Playing bool}`, read from `mpv`'s `time-pos`/`duration`/`pause` properties (polled, not pushed).
- `TogglePause()`, `Stop()`, `SeekBy(delta time.Duration)`, `Next()`, `Prev()` — thin wrappers issuing the corresponding IPC commands.
- Detects `mpv` process exit or IPC socket closure and surfaces it as a state change (`Playing: false`, plus an error the caller can log) rather than propagating a panic/crash.
- End-of-file (reported by `mpv` over IPC) triggers auto-advance to the next track in the loaded queue; running out of queue clears state back to idle.

### Banner sweep renderer (`internal/ui`)

- A new render helper, alongside the existing `banner` const, that takes the current `player.State` and the existing theme `Styles` and returns the banner string with each character styled by whether its column falls before or after the elapsed/duration fraction (muted color before the sweep point, accent color at/after it — the "color sweep" treatment).
- When nothing is playing, renders the banner exactly as it does today (no behavior change for the common case).
- Every screen's `view*()` function switches from calling `m.st.title.Render(banner)` directly to calling this helper, so the sweep and now-playing line appear identically everywhere.

### Library search (`internal/ui/library.go`)

- Extends the existing `libState` (Artist/Album/Track drill-down) with a `filter string` field. Typed characters append to it and narrow the current level's `entries` list live (case-insensitive substring match); backspace shortens it; navigating levels resets it.
- `Enter` on a track (currently a no-op at the deepest drill-down level) is wired to: build the queue as the sorted track list of the current album directory, find the selected track's index within it, and call `player.Load(tracks, index)`.

### Global key routing (`internal/ui/program.go`)

- The root `Update` gains a small set of keys handled before per-screen dispatch (same precedence tier as today's global `q`): space (play/pause), `s` (stop), `n`/`p` (skip next/previous), left/right arrow (seek back/forward, e.g. 10s). These must not collide with a screen's own use of the same key when a screen is mid-text-entry (Library's new type-ahead filter, for one) — space and arrows in particular need to only trigger playback control when the active screen isn't currently capturing that keystroke as text/filter input, matching how `q` already coexists with typing today by being screen-scoped where needed. This needs care during implementation; flagged here rather than glossed over, given a very similar class of bug (global shortcuts stealing keystrokes from focused text input) was a real, serious bug in the previous UI redesign attempt.

## Data Flow

1. User presses Enter on a track while browsing Library at the Track level.
2. Library builds the sorted track list for the current album directory and the selected track's index within it.
3. Library calls `player.Load(tracks, index)`. If `mpv` isn't running yet, `internal/player` spawns it (with the IPC socket flag) and connects; otherwise it reuses the existing process and sends `loadfile`.
4. `player.State()` starts reporting the new title and duration once `mpv` confirms the file is loaded.
5. The root model's existing render tick (or a new ~250ms tick, if none suits) reads `player.State()` each cycle and re-renders the banner sweep + now-playing line, regardless of which screen is currently displayed.
6. Global key presses (space/arrows/n/p/s) call the corresponding `internal/player` method, which issues the matching `mpv` IPC command.
7. When `mpv` reports end-of-file, `internal/player` auto-loads the next track in the queue set up in step 3, or clears state if the queue is exhausted.

## Error Handling

- **`mpv` not installed:** caught the first time a track is played (not at app startup), with a clear inline message ("mpv not found — install it to enable playback"). Library remains fully usable for browsing/downloading regardless.
- **`mpv` crashes or the IPC socket drops mid-playback:** detected by `internal/player`; now-playing state clears, a short error surfaces via the existing Log workspace's pattern, and the rest of the app keeps working normally.
- **Corrupt or unsupported file:** `mpv` reports an error over IPC rather than crashing the process; treated like an unplayable track — skip to the next queued track with a message, don't hang waiting on it.
- **Seeking past the start or end of a track:** clamps to `[0, duration]`.

## Testing

- `internal/player`'s IPC client is tested against a fake local socket speaking `mpv`'s JSON-line protocol in-process — no real `mpv` binary required in CI, matching how `internal/download` already avoids invoking real `yt-dlp`/`ffmpeg` in its own tests.
- Library's new filter logic and queue-building (given a selected track, produce its album's sorted track list plus the selected index) are pure functions with direct unit tests.
- The banner sweep renderer is tested by asserting on rendered string output at a few progress fractions (0%, mid-track, ~100%) and the idle (no playback) case, matching the existing pattern for `viewMenu`/`viewLibrary`.

## Dependency Changes

- `mpv` is added as a new, playback-only external dependency (checked lazily, not required for the app to run). `internal/download/deps.go`'s dependency list, the README, and platform install docs (Homebrew, Scoop, apt/source-install instructions) need an entry for it — tracked as implementation work, not detailed further here.
