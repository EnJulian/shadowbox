# Setup Wizard — Design

> **Superseded in part by [Revision 1](#revision-1-linux-auto-install) below**, which enables real automated installs on Linux (originally instructions-only) once the wizard shipped and the user found the Linux experience unhelpfully passive. Everything else on this page is still accurate; read Revision 1 for what changed and why.

## Goal

Shadowbox has two optional external dependencies — `mpv` (in-app playback) and `aria2` (accelerated downloads) — that are currently invisible unless a user thinks to run `shadowbox doctor` or reads the README. This adds a setup wizard that automatically appears the very first time Shadowbox is ever run (no `config.yaml` present yet) offering to install them, and is also reachable any time afterward from Settings, so a user can install something they skipped earlier or re-run it whenever they want.

## Scope

**In scope:**
- A wizard screen listing `mpv` and `aria2`, each showing installed/not-installed.
- Auto-triggered on the very first run (config file doesn't exist yet); reachable manually via a new "Run setup wizard" item in Settings.
- Where installation can be automated without elevated permissions (macOS via `brew`, Windows via `scoop` or `winget`), the wizard runs the real install command and shows live output.
- Everywhere else (Linux, or no known package manager found), the wizard shows the exact command to run manually instead of attempting an automated install.
- Re-checking a tool's availability after an install attempt.

**Out of scope:**
- Installing `yt-dlp` or `ffmpeg` (both already required; this wizard is for the two optional tools only).
- Any automated install path that requires sudo/elevated permissions or interactive credential entry.
- Uninstalling tools, or managing versions/updates of already-installed tools.
- A "don't show this again" toggle beyond the natural one-time trigger described below.

## First-Run Detection

The wizard auto-triggers when `config.yaml` does not exist yet at `config.Path()` — checked once, before the initial screen is chosen in `runProgram`. No new persisted field is needed: the wizard's own "Done" action always calls `config.Save()`, which creates the file, so the auto-trigger naturally never fires again on the same machine. This also means an existing user upgrading to this version won't suddenly see the wizard (they already have a config file) — they can still open it manually from Settings.

If the `os.Stat` check itself errors for a reason other than "file does not exist" (e.g. a permissions problem on the config directory), the wizard does **not** auto-trigger — fail safe, rather than risk showing it on every launch or blocking startup. This has no effect on the manual Settings trigger.

## Architecture

A new `internal/setup` package holds pure, testable detection logic: which tools are missing, and for the current OS, whether an automated install command exists and is safe to run without elevated permissions. A new `screenSetupWizard` screen in `internal/ui` renders the list and drives installs through `tea.ExecProcess` (which suspends the Bubble Tea alt-screen, hands the real terminal to the subprocess so its output — and, if it ever needed one, a password prompt — is visible, then resumes). The wizard remembers which screen it was opened from (`screenMenu` on first run, `screenSettings` when opened manually) and returns there when done.

The list is three entries — "mpv", "aria2", and "Done" — navigated and selected with the same up/down/enter pattern every other list screen in this app already uses (Menu, Settings, Library), rather than introducing a new dedicated keybinding for finishing the wizard. `Esc` also saves and returns, identically to selecting "Done" — consistent with every other screen's back behavior. This is safe even mid-consideration of an install, since `tea.ExecProcess` blocks the whole Bubble Tea event loop while the subprocess runs, so no keypress can reach the wizard (or interrupt a live install) until the subprocess actually finishes.

## Components

- **`internal/setup`** (new package): `Tool{Name, Purpose string}`, `Available(tool string) bool` (thin wrappers around the existing `player.Available()` / `download.HasAria2()`), and `InstallCommand(tool string) (cmd []string, automatable bool)` — returns the right command for `runtime.GOOS` and whether it can run without elevated permissions. macOS → `brew install <tool>` (if `brew` is on PATH); Windows → `scoop install <tool>` (if `scoop` is on PATH) else `winget install <tool>` (if `winget` is on PATH); anything else (including no manager found on macOS/Windows) → `automatable = false`, with an instructional command string to display instead. Linux is always `automatable = false` — virtually every native package manager there needs sudo, and reliably detecting "can this sudo without a password right now" isn't worth the complexity for this feature.
- **`internal/ui/wizard.go`** (new): the screen — a three-item list (mpv, aria2, Done), each tool item showing its installed state; Enter on an already-installed tool item is a no-op (nothing to do); Enter on an automatable, not-yet-installed tool item runs the install via `tea.ExecProcess` and re-checks availability afterward; Enter on a non-automatable, not-yet-installed tool item toggles inline display of the manual install command; Enter on "Done" saves config and returns to the stored origin screen.
- **`internal/ui/program.go` changes**: a new `screenSetupWizard` value; the first-run check in `runProgram`; a `wizardReturnTo screen` field on `model` so "Done" knows where to go back to.
- **Settings changes** (`internal/ui/settings.go`): one new `settingItem{"setup_wizard", "Run setup wizard", settingAction}` (a new `settingKind`), wired the same way the existing theme-picker entry switches screens.

## Data Flow

1. `runProgram(cfg)` checks whether `config.Path()` exists. If not, the initial model's screen is `screenSetupWizard` with `wizardReturnTo: screenMenu`; otherwise `screenMenu` as today.
2. From Settings, selecting "Run setup wizard" switches to `screenSetupWizard` with `wizardReturnTo: screenSettings`.
3. On entering the wizard, it computes each tool's installed state and install capability once (not re-computed every render).
4. Enter on an automatable item runs `tea.ExecProcess(exec.Command(cmd[0], cmd[1:]...), callback)`; the callback re-checks `Available()` and sets a per-item result message (success/failure).
5. Enter on a non-automatable item toggles an inline instructions block showing the exact command to copy-paste.
6. Enter on "Done" calls `config.Save(m.cfg)` and switches to `m.wizardReturnTo`.

## Error Handling

- **Install command fails** (non-zero exit): the tool stays marked not-installed, a "install failed — see output above" line shows, the wizard stays open — no crash, no silent success claim.
- **No known package manager found** on macOS/Windows: falls back to the same instructions-only path as Linux, no special-cased error message.
- **First-run `os.Stat` check errors** for a reason other than not-existing: treated as "not first run," fails safe.
- **`config.Save()` fails** on "Done": swallowed the same way the rest of Settings already does today (`_ = config.Save(m.cfg)`) — not a new failure mode.

## Testing

- `internal/setup`'s `InstallCommand`/`Available` logic is unit-tested as pure functions, given an injectable "is this on PATH" lookup and a `GOOS` value — no real package managers needed, following the same fake-injection pattern already used for `internal/player`'s IPC client.
- The wizard screen's navigation, item toggling, and "Done" behavior, plus `program.go`'s first-run screen selection, are tested by driving `Update`/`handleKey` directly against a temp-dir config path (this project's existing sandboxing rule for anything touching `internal/config`).
- The real `tea.ExecProcess` install run itself is not unit-tested, consistent with this project's existing convention of not testing actual subprocess invocation (see `internal/download`) — verified instead with a manual smoke test before this ships.

## Revision 1: Linux auto-install

After shipping, the Linux experience (always instructions-only) turned out to be unhelpfully passive — a user tried it and asked for the exact command plus real help running it, the same as macOS/Windows already get. This revision changes exactly one thing: `installCommandFor`'s Linux branch (previously always `Automatable: false` with a generic "use your distro's package manager" message) now detects a real package manager and, if found, becomes automatable — same shape and same `tea.ExecProcess` execution path already used for brew/scoop/winget, no changes needed anywhere outside `internal/setup`.

**Detection, in priority order:** `apt-get`, `dnf`, `pacman`, `zypper` (covers Debian/Ubuntu, Fedora/RHEL, Arch, openSUSE — the large majority of desktop Linux). If none of these four are on PATH, falls back to today's generic instructions-only message — this part is unchanged.

**Commands:**
- `apt-get` → `sudo apt-get install -y <PackageName>`
- `dnf` → `sudo dnf install -y <PackageName>`
- `pacman` → `sudo pacman -S --noconfirm <PackageName>`
- `zypper` → `sudo zypper install -y <PackageName>`

All four use the existing `Tool.PackageName` field unchanged — verified `aria2`'s package name is `"aria2"` on all four (not `aria2c`), consistent with the existing brew/scoop entries, so no new per-manager name mapping is needed.

**Sudo inside the TUI:** this is safe via the same mechanism already used for brew/scoop/winget — `tea.ExecProcess` fully suspends Bubble Tea's rendering and hands the real terminal to the subprocess for its entire run, so a `sudo` password prompt behaves exactly as it would running the command directly in a terminal (reads/writes the real stdin/stdout, no pty tricks needed). The wizard's help line gains one clause when the highlighted item is automatable, so this isn't a surprise: `"enter: install (may ask for your password)   esc: done"`.

**macOS/Windows are unchanged** — they already auto-run via brew/scoop/winget; the user confirmed no further expansion (e.g. MacPorts, Chocolatey) is wanted there.

**Testing:** four new `installCommandFor("linux", ...)` cases (one per manager, each verifying the exact command and package name), plus a "none found" case confirming the existing generic fallback still applies — all via the same injected `lookPath` pattern already used for the darwin/windows tests, no real package managers needed.

