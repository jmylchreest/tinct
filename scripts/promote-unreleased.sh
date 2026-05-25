#!/usr/bin/env bash
# promote-unreleased.sh — rename CHANGELOG.md's `[Unreleased]` section
# to `[VERSION]`, append auto-generated conventional-commits entries
# from PREV_TAG..HEAD beneath any human-curated content, and re-seed an
# empty `[Unreleased]` at the top.
#
# Usage:
#   scripts/promote-unreleased.sh <version> <date> <previous-tag>
#
# Example:
#   scripts/promote-unreleased.sh 0.3.3 2026-05-06 v0.3.2
#
# Side effects: rewrites CHANGELOG.md in place. The companion script
# scripts/changelog-sections.sh must exist (called for the auto-generated
# bullet list).
#
# Behaviour rules:
#   - If `[Unreleased]` exists with body content, that content is kept
#     and the auto-generated sections are appended underneath it.
#   - If `[Unreleased]` exists with no body content, only the
#     auto-generated sections appear under `[VERSION]`.
#   - If neither human nor auto entries exist, the new `[VERSION]` block
#     contains just header + date (matching today's pre-fix behaviour).
#   - `[Unreleased]` is reset to an empty section at the top.

set -euo pipefail

VERSION="${1:?usage: $0 <version> <date> <previous-tag>}"
DATE="${2:?usage: $0 <version> <date> <previous-tag>}"
PREV_TAG="${3:-}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CHANGELOG="${CHANGELOG:-CHANGELOG.md}"
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

if [ ! -f "$CHANGELOG" ]; then
  echo "error: $CHANGELOG not found" >&2
  exit 1
fi

# Extract any human-curated body under [Unreleased] — everything between
# the `## [Unreleased]` line and the next `## ` heading (or end of file).
human_body=$(awk '
  /^## \[Unreleased\]/ { capture=1; next }
  /^## / { capture=0 }
  capture { print }
' "$CHANGELOG" | sed '/./,$!d' | sed -e :a -e '/^\s*$/{$d;N;ba' -e '}')

# Auto-generate sections from conventional commits.
auto_body=""
if [ -n "$PREV_TAG" ]; then
  auto_body=$("$SCRIPT_DIR/changelog-sections.sh" "${PREV_TAG}..HEAD" || true)
  auto_body=$(printf '%s' "$auto_body" | sed -e :a -e '/^\s*$/{$d;N;ba' -e '}')
fi

# Build the new [VERSION] block.
{
  printf '## [%s]\n' "$VERSION"
  printf '*%s*\n' "$DATE"
  if [ -n "$human_body" ]; then
    printf '\n%s\n' "$human_body"
  fi
  if [ -n "$auto_body" ]; then
    printf '\n%s\n' "$auto_body"
  fi
} > "$TMP.version"

# Rewrite CHANGELOG.md:
#   - Replace [Unreleased] line + its body with a fresh empty [Unreleased]
#     followed by the new [VERSION] block.
#   - Preserve everything from the next `## ` heading onwards.
awk -v version_file="$TMP.version" '
  BEGIN {
    while ((getline line < version_file) > 0) version_block = version_block line "\n"
    close(version_file)
    in_unreleased = 0
    emitted_unreleased = 0
  }
  /^## \[Unreleased\]/ {
    if (!emitted_unreleased) {
      print "## [Unreleased]"
      print ""
      printf "%s", version_block
      print ""
      emitted_unreleased = 1
    }
    in_unreleased = 1
    next
  }
  /^## / {
    in_unreleased = 0
    print
    next
  }
  in_unreleased { next }
  { print }
' "$CHANGELOG" > "$TMP"

mv "$TMP" "$CHANGELOG"
trap - EXIT

echo "Promoted [Unreleased] → [$VERSION] in $CHANGELOG"
if [ -n "$human_body" ]; then
  echo "  (preserved human-curated entries)"
fi
if [ -n "$auto_body" ]; then
  echo "  (added auto-generated sections from $PREV_TAG..HEAD)"
fi
