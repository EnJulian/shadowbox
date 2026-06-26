# Homebrew tap for Shadowbox

Shadowbox is distributed through a custom Homebrew tap:
[`EnJulian/homebrew-shadowbox`](https://github.com/EnJulian/homebrew-shadowbox).

## Installing

```bash
brew install EnJulian/shadowbox/shadowbox
```

This pulls in the required `ffmpeg` and `yt-dlp` formulae automatically. Install
`aria2` separately for faster downloads:

```bash
brew install aria2
```

## How releases update the tap

The cask is generated and pushed automatically by GoReleaser during the
[release workflow](../../.github/workflows/release.yml). On every `v*` tag,
GoReleaser:

1. Builds the cross-platform binaries.
2. Renders `Casks/shadowbox.rb` with the new version, URLs, and checksums.
3. Commits and pushes it to the tap repository using the
   `HOMEBREW_TAP_GITHUB_TOKEN` secret.

## One-time tap setup

1. Create an empty public repository named `homebrew-shadowbox` under the
   `EnJulian` account.
2. Create a fine-grained or classic personal access token with `contents: write`
   permission on that repository.
3. Add it to the `shadowbox` repository secrets as `HOMEBREW_TAP_GITHUB_TOKEN`.

The file `shadowbox.rb` in this directory is a reference of what GoReleaser
produces; you do not need to edit it by hand.
