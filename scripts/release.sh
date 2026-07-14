#!/usr/bin/env bash
# Cut a new shadowbox release: bump CHANGELOG.md, commit, tag, and push.
#
# Pushing the tag triggers .github/workflows/release.yml (GoReleaser: build,
# sign, SBOM, GitHub Release, Homebrew, Scoop), which in turn triggers
# .github/workflows/winget.yml on publish.
#
# Usage:
#   scripts/release.sh <patch|minor|major|X.Y.Z> [--yes] [--dry-run]
#
#   patch|minor|major   Bump the latest git tag (vX.Y.Z) accordingly.
#   X.Y.Z / vX.Y.Z       Use this exact version instead of bumping.
#   --yes, -y            Skip confirmation prompts.
#   --dry-run             Show what would happen; don't write, commit, or push.
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

changelog="CHANGELOG.md"
branch="main"
assume_yes=0
dry_run=0
bump_arg=""

usage() {
  sed -n '2,14p' "$0" | sed 's/^# \{0,1\}//'
}

for arg in "$@"; do
  case "$arg" in
    -h|--help) usage; exit 0 ;;
    -y|--yes) assume_yes=1 ;;
    --dry-run) dry_run=1 ;;
    *)
      if [[ -n "$bump_arg" ]]; then
        echo "error: unexpected extra argument: $arg" >&2
        exit 1
      fi
      bump_arg="$arg"
      ;;
  esac
done

if [[ -z "$bump_arg" ]]; then
  usage
  exit 1
fi

confirm() {
  local prompt="$1"
  if [[ "$assume_yes" -eq 1 ]]; then
    return 0
  fi
  read -r -p "$prompt [y/N] " reply
  [[ "$reply" =~ ^[Yy]$ ]]
}

# --- Preflight ---------------------------------------------------------

if [[ -n "$(git status --porcelain)" ]]; then
  echo "error: working tree is not clean. Commit or stash your changes first." >&2
  exit 1
fi

current_branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$current_branch" != "$branch" ]]; then
  echo "error: must be on '$branch' to release (currently on '$current_branch')." >&2
  exit 1
fi

echo "Fetching origin/$branch..."
git fetch origin "$branch" --quiet

local_head="$(git rev-parse HEAD)"
remote_head="$(git rev-parse "origin/$branch")"
if [[ "$local_head" != "$remote_head" ]]; then
  base="$(git merge-base HEAD "origin/$branch")"
  if [[ "$base" == "$remote_head" ]]; then
    echo "Local $branch is ahead of origin/$branch; will push those commits too."
  else
    echo "error: local $branch has diverged from origin/$branch. Pull/rebase first." >&2
    exit 1
  fi
fi

# --- Compute new version -------------------------------------------------

latest_tag="$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -1)"
latest_tag="${latest_tag:-v0.0.0}"
current_version="${latest_tag#v}"
IFS=. read -r major minor patch <<<"$current_version"

if [[ "$bump_arg" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  new_version="${bump_arg#v}"
else
  case "$bump_arg" in
    major) new_version="$((major + 1)).0.0" ;;
    minor) new_version="$major.$((minor + 1)).0" ;;
    patch) new_version="$major.$minor.$((patch + 1))" ;;
    *)
      echo "error: bump must be one of patch|minor|major or an explicit X.Y.Z version (got: $bump_arg)" >&2
      exit 1
      ;;
  esac
fi
new_tag="v$new_version"

if git rev-parse "$new_tag" >/dev/null 2>&1; then
  echo "error: tag $new_tag already exists." >&2
  exit 1
fi

echo "Current version: $current_version"
echo "New version:     $new_version"

# --- Rewrite CHANGELOG.md -------------------------------------------------

if [[ ! -f "$changelog" ]]; then
  echo "error: $changelog not found." >&2
  exit 1
fi

release_date="$(date -u +%Y-%m-%d)"
tmp_changelog="$(mktemp)"
trap 'rm -f "$tmp_changelog"' EXIT

awk -v ver="$new_version" -v date="$release_date" '
  state == 0 && /^## \[Unreleased\]/ {
    print
    print ""
    print "## [" ver "] - " date
    state = 1
    next
  }
  state == 1 && /^## \[/ {
    printf "%s", body
    state = 2
    print
    next
  }
  state == 1 {
    body = body $0 "\n"
    next
  }
  { print }
' "$changelog" > "$tmp_changelog"

if ! grep -q "^## \[Unreleased\]" "$changelog"; then
  echo "error: no '## [Unreleased]' section found in $changelog." >&2
  exit 1
fi

unreleased_body="$(awk '
  /^## \[Unreleased\]/ { capture = 1; next }
  capture && /^## \[/ { exit }
  capture { print }
' "$changelog")"

if [[ -z "${unreleased_body//[$' \t\r\n']/}" ]]; then
  echo "error: '## [Unreleased]' section is empty. Add changelog entries before releasing." >&2
  exit 1
fi

echo
echo "--- CHANGELOG.md diff -----------------------------------------------"
diff -u "$changelog" "$tmp_changelog" || true
echo "------------------------------------------------------------------------"
echo

if [[ "$dry_run" -eq 1 ]]; then
  echo "Dry run: would commit, tag $new_tag, and push. No changes made."
  exit 0
fi

if ! confirm "Apply this CHANGELOG.md update and commit?"; then
  echo "Aborted."
  exit 1
fi

cp "$tmp_changelog" "$changelog"

# --- Commit and tag --------------------------------------------------------

git add "$changelog"
git commit -m "docs: update changelog for $new_tag" --quiet

notes_file="$(mktemp)"
trap 'rm -f "$tmp_changelog" "$notes_file"' EXIT
bash "$repo_root/scripts/extract-release-notes.sh" "$new_tag" "$changelog" "$notes_file"

git tag -a "$new_tag" -F "$notes_file"

echo "Committed and tagged $new_tag locally."

if ! confirm "Push commit + tag $new_tag to origin/$branch? This triggers the public release pipeline."; then
  echo "Not pushed. To finish later, run:"
  echo "  git push origin $branch"
  echo "  git push origin $new_tag"
  exit 0
fi

git push origin "$branch"
git push origin "$new_tag"

echo
echo "Pushed $new_tag. Release pipeline triggered:"
if command -v gh >/dev/null 2>&1; then
  run_url="$(gh run list --workflow=release.yml --branch="$branch" --limit=1 --json url --jq '.[0].url' 2>/dev/null || true)"
  if [[ -n "$run_url" ]]; then
    echo "  $run_url"
  else
    echo "  https://github.com/EnJulian/shadowbox/actions/workflows/release.yml"
  fi
  if confirm "Watch the release run with 'gh run watch'?"; then
    sleep 3
    run_id="$(gh run list --workflow=release.yml --branch="$branch" --limit=1 --json databaseId --jq '.[0].databaseId' 2>/dev/null || true)"
    if [[ -n "$run_id" ]]; then
      gh run watch "$run_id"
    fi
  fi
else
  echo "  https://github.com/EnJulian/shadowbox/actions/workflows/release.yml"
fi
