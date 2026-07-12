# Installing Shadowbox from source

This guide covers building and running Shadowbox locally — useful for testing
before (or instead of) installing through Homebrew or Scoop.

For releasing, see [RELEASING.md](RELEASING.md).

## Prerequisites

- **Go 1.25+** — the module pins `go 1.25`.
- **Runtime tools** on your `PATH`: `yt-dlp` and `ffmpeg` (required), `aria2`
  (optional, for faster downloads), `mpv` (optional, for in-app playback).
  Verify later with `shadowbox doctor`.

### Make sure `go` is on your PATH

If `go version` fails with "command not found", Go isn't on your PATH. On this
machine the toolchain lives in `~/.local/go-toolchain/bin`. Add it permanently:

```bash
echo 'export PATH="$HOME/.local/go-toolchain/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
go version   # should now print go1.25.x
```

## Option A — Run without installing (fastest dev loop)

From the project root:

```bash
go run ./cmd/shadowbox            # launches the interactive TUI
go run ./cmd/shadowbox doctor     # runs a subcommand
```

## Option B — Build a local binary

```bash
make build        # produces ./shadowbox with version info baked in
./shadowbox       # launch the TUI
```

`make build` injects the version, commit, and build date via ldflags, so
`./shadowbox version` shows real values.

## Option C — Install onto your PATH (run `shadowbox` from anywhere)

```bash
go install ./cmd/shadowbox
```

This compiles and copies the binary to Go's install directory:

- `$GOBIN` if set, otherwise
- `$(go env GOPATH)/bin` (commonly `~/go/bin`).

On this machine it installs to `~/.local/bin`, which is already on the PATH.

Find the exact location and confirm it's on your PATH:

```bash
go env GOBIN GOPATH        # where binaries go
echo "$PATH" | tr ':' '\n' | grep -E 'go/bin|\.local/bin'
```

If the install directory is **not** on your PATH, add it (adjust the path to
match the output above):

```bash
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Then open a new terminal (or run `hash -r` to clear the shell's command cache)
and run:

```bash
shadowbox            # TUI
shadowbox version
shadowbox doctor
```

> Note: binaries installed with `go install` report their version as `dev`
> because the version stamp is only injected by `make build` and by official
> release builds. This is cosmetic and does not affect functionality.

## Verifying your setup

```bash
shadowbox doctor
```

You should see `yt-dlp` and `ffmpeg` marked `[+]`. Configure Genius credentials
(optional, for lyrics) with:

```bash
shadowbox config set genius.access_token YOUR_TOKEN
```

## Updating

After pulling new changes, re-run whichever option you use:

```bash
git pull
go install ./cmd/shadowbox    # or: make build
hash -r                        # so the shell picks up the new binary
```

## Troubleshooting

- **`go: command not found`** — Go isn't on your PATH. See
  "Make sure `go` is on your PATH" above. This is the most common reason
  `go install` appears to "do nothing".
- **`shadowbox: command not found` after install** — the install directory
  isn't on your PATH, or your shell cached the old lookup. Add the directory to
  PATH and run `hash -r` or open a new terminal.
- **`could not open a new TTY` / blank screen** — the interactive TUI needs a
  real terminal. Run `shadowbox` directly in an interactive shell, not piped
  (`shadowbox | less`), redirected (`shadowbox < /dev/null`), or from a
  non-interactive output pane.
- **`doctor` shows tools missing** — install `yt-dlp` and `ffmpeg` and ensure
  they're on your PATH, then re-run `shadowbox doctor`.
```
