#!/usr/bin/env bash
# Extract the Keep a Changelog section for a release tag from CHANGELOG.md.
#
# Usage:
#   bash scripts/extract-release-notes.sh v1.4.2 [CHANGELOG.md] [output.md]
#
# When output.md is omitted, writes to stdout. Exits non-zero with a helpful
# message if the tag is invalid, the changelog is missing, or no section exists.
set -euo pipefail

tag="${1:?usage: extract-release-notes.sh vX.Y.Z [CHANGELOG.md] [output.md]}"
changelog="${2:-CHANGELOG.md}"
output="${3:-}"

if [[ ! "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo "error: tag must look like v1.2.3 (got: $tag)" >&2
  exit 1
fi

version="${tag#v}"

if [[ ! -f "$changelog" ]]; then
  echo "error: changelog file not found: $changelog" >&2
  exit 1
fi

notes="$(
  awk -v ver="$version" '
    /^## \[Unreleased\]/ { next }
    /^## \[/ {
      if (capture) exit
      if ($0 ~ "^## \\[" ver "\\]") capture = 1
      next
    }
    capture { print }
  ' "$changelog"
)"

if [[ -z "${notes//[$' \t\r\n']/}" ]]; then
  echo "error: no changelog section for version ${version}" >&2
  echo "Add a '## [${version}] - YYYY-MM-DD' section to ${changelog} before pushing tag ${tag}." >&2
  echo "Existing release sections:" >&2
  grep -E '^## \[[0-9]' "$changelog" >&2 || echo "  (none)" >&2
  exit 1
fi

if [[ -n "$output" ]]; then
  printf '%s\n' "$notes" > "$output"
else
  printf '%s\n' "$notes"
fi
