# Testing & Releasing Shadowbox

A practical reference for verifying changes locally and cutting a release.

## Prerequisites

- **Go 1.25+** (the module pins `go 1.25`).
- **Runtime tools** on your `PATH` for end-to-end testing: `yt-dlp`, `ffmpeg`,
  and optionally `aria2`. Check with `shadowbox doctor`.
- Optional dev tools: `golangci-lint`, `goreleaser`.

## Local testing

```bash
make test     # go test -race ./...
make vet      # go vet ./...
make lint     # golangci-lint run ./...
make build    # build ./shadowbox with version info baked in
```

A few notes on the test suite:

- API clients (Spotify, iTunes, Last.fm, Genius) and the cover cascade are
  tested with `httptest` mocks — no network needed.
- Tag round-trip tests cover **Opus** and **MP3** with synthetic fixtures and
  always run. **FLAC** and **M4A** tests generate real files with `ffmpeg` and
  are **skipped automatically when `ffmpeg` is not installed**, so install it to
  exercise them.

### Manual end-to-end smoke test

With `yt-dlp` and `ffmpeg` available:

```bash
make build
./shadowbox doctor                          # all required tools green?
./shadowbox download -q "Imagine Dragons Believer" -d /tmp/sbtest -v
ls -R /tmp/sbtest                            # expect Artist/Album/Title.opus
```

For metadata enrichment, set credentials first:

```bash
./shadowbox config set spotify.client_id     YOUR_ID
./shadowbox config set spotify.client_secret YOUR_SECRET
./shadowbox config set genius.access_token   YOUR_TOKEN
./shadowbox download -q "Adele Hello" -s -v
```

### Cross-platform build check (optional)

```bash
make snapshot     # goreleaser build --snapshot --clean
# binaries land in ./dist/<os>_<arch>/
```

## How a release works

Releases are fully automated by GoReleaser and GitHub Actions. Pushing a
`v*` tag triggers `.github/workflows/release.yml`, which:

1. Builds static binaries for linux/darwin/windows on amd64/arm64
   (windows/arm64 is intentionally skipped).
2. Creates the GitHub Release with archives + `checksums.txt`.
3. Pushes an updated Homebrew **cask** to `EnJulian/homebrew-shadowbox`
   (using the `HOMEBREW_TAP_GITHUB_TOKEN` secret).

Separately, when the GitHub Release is *published*,
`.github/workflows/winget.yml` opens a pull request against
`microsoft/winget-pkgs` (using the `WINGET_TOKEN` secret and your
`EnJulian/winget-pkgs` fork).

### Required repository secrets

| Secret | Used by | Token type |
|--------|---------|------------|
| `HOMEBREW_TAP_GITHUB_TOKEN` | GoReleaser cask push | Fine-grained PAT, Contents: read/write on `EnJulian/homebrew-shadowbox` |
| `WINGET_TOKEN` | WinGet PR workflow | Classic PAT with `public_repo` scope |

`GITHUB_TOKEN` is provided automatically by Actions for the release itself.

## Cutting a release — checklist

1. Make sure `main` is green: `make test && make vet && make lint`.
2. Update `CHANGELOG.md`: move items from `[Unreleased]` into a new
   `[X.Y.Z] - YYYY-MM-DD` section.
3. Commit and push to `main`.
4. Tag and push:
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
5. Watch the release run:
   ```bash
   gh run watch --repo EnJulian/shadowbox
   # or: gh run list --repo EnJulian/shadowbox
   ```
6. Verify:
   ```bash
   gh release view vX.Y.Z --repo EnJulian/shadowbox
   ```

### Versioning

Shadowbox follows [Semantic Versioning](https://semver.org): `MAJOR.MINOR.PATCH`.
The version, commit, and build date are injected into the binary at build time
and shown by `shadowbox version`.

## Verifying the published channels

```bash
# Homebrew
brew tap EnJulian/shadowbox
brew install shadowbox
shadowbox version

# WinGet (after the winget-pkgs PR merges)
winget install EnJulian.shadowbox
```

## Troubleshooting

- **Release workflow fails on the Homebrew step** — confirm
  `HOMEBREW_TAP_GITHUB_TOKEN` is set and can write to
  `EnJulian/homebrew-shadowbox`.
- **No WinGet PR appears** — the workflow runs on `release: published` (not
  draft/prerelease). Confirm `WINGET_TOKEN` is set and the
  `EnJulian/winget-pkgs` fork exists. You can also run it manually from the
  Actions tab via `workflow_dispatch`.
- **`goreleaser` config errors** — validate locally with `goreleaser check`.
- **YouTube anti-bot / download failures** — update yt-dlp
  (`yt-dlp -U` or reinstall) and re-run `shadowbox doctor`.
```
