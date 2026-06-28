#!/usr/bin/env bash
# Extract the Keep a Changelog section for a release tag from CHANGELOG.md.
# Usage: extract-release-notes.sh v1.4.1 [CHANGELOG.md]
set -euo pipefail

tag="${1:?usage: extract-release-notes.sh vX.Y.Z [CHANGELOG.md]}"
version="${tag#v}"
changelog="${2:-CHANGELOG.md}"

awk -v ver="$version" '
  /^## \[/ {
    if (capture) exit
    if ($0 ~ "^## \\[" ver "\\]") capture = 1
    next
  }
  capture { print }
' "$changelog"
