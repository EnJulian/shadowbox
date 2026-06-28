# Scoop bucket for Shadowbox

Shadowbox is distributed on Windows through a custom Scoop bucket:
[`EnJulian/scoop-shadowbox`](https://github.com/EnJulian/scoop-shadowbox).

## Installing

```powershell
scoop bucket add shadowbox https://github.com/EnJulian/scoop-shadowbox
scoop install shadowbox
```

This pulls in the required `ffmpeg` and `yt-dlp` packages automatically. Install
`aria2` separately for faster downloads:

```powershell
scoop install aria2
```

## How releases update the bucket

The manifest is generated and pushed automatically by GoReleaser during the
[release workflow](../../.github/workflows/release.yml). On every `v*` tag,
GoReleaser:

1. Builds the cross-platform binaries.
2. Renders `shadowbox.json` with the new version, URL, and checksum.
3. Commits and pushes it to the bucket repository using
   `HOMEBREW_TAP_GITHUB_TOKEN` (the same PAT as the Homebrew tap).

## One-time bucket setup

1. Create an empty public repository named `scoop-shadowbox` under the
   `EnJulian` account.
2. Extend the fine-grained PAT used for `HOMEBREW_TAP_GITHUB_TOKEN` so it has
   **Contents: read/write** on **both** `EnJulian/homebrew-shadowbox` and
   `EnJulian/scoop-shadowbox`.
3. No separate Scoop secret is required.

GoReleaser writes the first manifest on the next tagged release. To backfill an
existing release without cutting a new tag, update `shadowbox.json` in the
bucket manually or push a patch tag.

The manifest GoReleaser produces lives at the bucket repo root as
`shadowbox.json`; you do not need to edit it by hand.
