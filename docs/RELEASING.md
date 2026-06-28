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
are **skipped automatically when** `ffmpeg` **is not installed**, so install it to
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
signed `v*` tag triggers `.github/workflows/release.yml`, which:

1. Builds static binaries for linux/darwin/windows on amd64/arm64
  (windows/arm64 is intentionally skipped).
2. Creates the GitHub Release with archives, `checksums.txt`, SBOMs, and
  cosign keyless signatures.
3. Publishes GitHub build provenance attestations for release artifacts.
4. Pushes an updated Homebrew **cask** to `EnJulian/homebrew-shadowbox`
  (using the `HOMEBREW_TAP_GITHUB_TOKEN` secret).
5. Pushes an updated Scoop **manifest** to `EnJulian/scoop-shadowbox`
  (using the `SCOOP_BUCKET_TOKEN` secret).

Separately, when the GitHub Release is *published*,
`.github/workflows/winget.yml` opens a pull request against
`microsoft/winget-pkgs` (using the `WINGET_TOKEN` secret and your
`EnJulian/winget-pkgs` fork). WinGet is optional and not required for Windows
installs — Scoop is the supported path for now.

### Required repository secrets


| Secret                      | Used by              | Token type                                                              |
| --------------------------- | -------------------- | ----------------------------------------------------------------------- |
| `HOMEBREW_TAP_GITHUB_TOKEN` | GoReleaser cask push | Fine-grained PAT, Contents: read/write on `EnJulian/homebrew-shadowbox` |
| `SCOOP_BUCKET_TOKEN`        | GoReleaser Scoop push | Fine-grained PAT, Contents: read/write on `EnJulian/scoop-shadowbox`   |
| `WINGET_TOKEN`              | WinGet PR workflow   | Classic PAT with `public_repo` scope (optional)                        |


`GITHUB_TOKEN` is provided automatically by Actions for the release itself.
Cosign keyless signing and GitHub attestations use the workflow OIDC token and
require no additional repository secrets.

## SSH commit and tag signing

One-time setup on your machine (key generation must run locally):

```bash
# Use an existing ed25519 key or create one
ssh-keygen -t ed25519 -C "your@email" -f ~/.ssh/id_ed25519 -N ""

git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/id_ed25519.pub
git config --global commit.gpgsign true
git config --global tag.gpgsign true

# Allow verifying signatures locally
mkdir -p ~/.config/git
echo "$(git config user.email) namespaces=\"git\",$(cat ~/.ssh/id_ed25519.pub)" \
  >> ~/.config/git/allowed_signers
git config --global gpg.ssh.allowedSignersFile ~/.config/git/allowed_signers
```

Add the **public** key to GitHub under **Settings → SSH and GPG keys → New
signing key** (not as a deploy key). Commits and tags will show as Verified.

Verify a signed tag before pushing:

```bash
git tag -v vX.Y.Z
```



## GitHub repository security

Apply these settings once as a repository admin. They are not stored in the repo.

### Branch protection for `main`

```bash
gh api repos/EnJulian/shadowbox/branches/main/protection \
  --method PUT \
  --input - <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "checks": [
      {"context": "Test", "app_id": null},
      {"context": "govulncheck", "app_id": null},
      {"context": "Lint", "app_id": null},
      {"context": "Secret scan", "app_id": null},
      {"context": "Analyze", "app_id": null}
    ]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1
  },
  "required_signatures": true,
  "required_linear_history": true,
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false
}
EOF
```

If CodeQL check name differs, list open checks with `gh pr checks` after a PR
and adjust the `"Analyze"` entry above.

### Secret scanning and push protection

```bash
# Enable secret scanning (GitHub Advanced Security features may require a paid plan)
gh api repos/EnJulian/shadowbox \
  --method PATCH \
  -f security_and_analysis='{"secret_scanning":{"status":"enabled"},"secret_scanning_push_protection":{"status":"enabled"}}'
```

On github.com: **Settings → Code security and analysis** — enable **Secret
scanning**, **Push protection**, and **Dependabot alerts**.

## Cutting a release — process flow

The whole flow in one line:

```
edit CHANGELOG → commit → push main (CI) → tag vX.Y.Z → push tag (release) → watch → verify
```

Only **pushing a** `vX.Y.Z` **tag** triggers a release. Pushing `main` just runs CI.

### Step 0 — Pre-flight

```bash
make test && make vet && make lint
```

The release builds from the tagged commit, so make sure it is healthy first.

### Step 1 — Update the CHANGELOG

Add a `## [X.Y.Z] - YYYY-MM-DD` section under `## [Unreleased]` describing the
changes (Added / Fixed / Changed / Removed).

### Step 2 — Commit the CHANGELOG

```bash
git add CHANGELOG.md
git commit -m "docs: update changelog for vX.Y.Z"
```



### Step 3 — Push main (runs CI, not a release)

```bash
git push origin main
```

Confirm CI is green before tagging: `gh run list --repo EnJulian/shadowbox`.

### Step 4 — Create a signed tag and push (this triggers the release)

```bash
git tag -s vX.Y.Z -m "vX.Y.Z"
git tag -v vX.Y.Z    # confirm signature before pushing
git push origin vX.Y.Z
```



### Step 5 — Watch the release run

```bash
gh run watch --repo EnJulian/shadowbox
# or: gh run list --repo EnJulian/shadowbox
```



### Step 6 — Verify

```bash
gh release view vX.Y.Z --repo EnJulian/shadowbox
brew update && brew upgrade shadowbox && shadowbox version
```

Confirm the release includes `checksums.txt.sigstore.json`, SBOM files, and
provenance attestations (see [Verifying downloads](#verifying-downloads)).

### If a tag is wrong

Deleting or moving a pushed tag is messy and re-runs the release. The cleanest
fix for a mistake is to bump to the next patch (e.g. `vX.Y.Z+1`) rather than
re-pushing the same tag.

### Versioning

Shadowbox follows [Semantic Versioning](https://semver.org): `MAJOR.MINOR.PATCH`.
The version, commit, and build date are injected into the binary at build time
and shown by `shadowbox version`.

## Verifying downloads

Each GitHub Release ships:

- `checksums.txt` — SHA-256 hashes of all archives
- `checksums.txt.sigstore.json` — cosign keyless signature bundle
- `*.spdx.json` — SBOM per archive (via syft)
- GitHub build provenance attestations (viewable in the release UI)



### Verify checksum signature

Install [cosign](https://docs.sigstore.dev), download `checksums.txt` and
`checksums.txt.sigstore.json` from the release, then:

```bash
cosign verify-blob checksums.txt \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/EnJulian/shadowbox/.github/workflows/release.yml@refs/tags/.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```



### Verify archive attestation

Install the [GitHub CLI](https://cli.github.com/) and download an archive:

```bash
gh release download vX.Y.Z --repo EnJulian/shadowbox \
  --pattern 'shadowbox-linux-amd64.tar.gz'

gh attestation verify shadowbox-linux-amd64.tar.gz \
  --owner EnJulian --repo shadowbox
```



### Verify file integrity manually

```bash
sha256sum -c checksums.txt --ignore-missing
```



## Verifying the published channels

```bash
# Homebrew
brew tap EnJulian/shadowbox
brew install shadowbox
shadowbox version

# WinGet (after the winget-pkgs PR merges; optional)
winget install EnJulian.shadowbox

# Scoop (Windows — supported)
scoop bucket add shadowbox https://github.com/EnJulian/scoop-shadowbox
scoop install shadowbox
```



## Troubleshooting

- **Release workflow fails on the Homebrew step** — confirm
`HOMEBREW_TAP_GITHUB_TOKEN` is set and can write to
`EnJulian/homebrew-shadowbox`.
- **Release workflow fails on the Scoop step** — confirm
`SCOOP_BUCKET_TOKEN` is set and can write to `EnJulian/scoop-shadowbox`.
- **No WinGet PR appears** — the workflow runs on `release: published` (not
draft/prerelease). Confirm `WINGET_TOKEN` is set and the
`EnJulian/winget-pkgs` fork exists. You can also run it manually from the
Actions tab via `workflow_dispatch`.
- `goreleaser` **config errors** — validate locally with `goreleaser check`.
- **YouTube anti-bot / download failures** — update yt-dlp
(`yt-dlp -U` or reinstall) and re-run `shadowbox doctor`.

