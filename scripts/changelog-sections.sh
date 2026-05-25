#!/usr/bin/env bash
# changelog-sections.sh — generate Keep-a-Changelog `### Added` /
# `### Changed` / `### Fixed` sections from conventional-commit history.
#
# Usage:
#   scripts/changelog-sections.sh <range>
#
# Where <range> is anything `git log` understands (e.g. `v0.3.0..v0.3.1`,
# `v0.3.0..HEAD`, or empty for full history).
#
# Output: Markdown with `### <Section>` headings and bulleted entries.
# Sections with no entries are omitted. Output is empty (zero bytes) when
# no qualifying commits are found in the range — callers can detect this
# and skip the heading insertion.
#
# Sections are aligned with Keep a Changelog vocabulary, not goreleaser's
# (so CHANGELOG.md reads naturally to humans following the convention):
#
#   feat       → Added
#   fix        → Fixed
#   perf       → Changed (performance)
#   refactor   → Changed
#   revert     → Reverted
#   feat!/fix! → Breaking changes (with original section underneath)
#
# Conventional commits with `chore:`, `docs:`, `test:`, `ci:`,
# `build(deps):`, and `style:` prefixes are filtered out — they're noise
# in a user-facing changelog.

set -euo pipefail

RANGE="${1:-}"

if [ -z "$RANGE" ]; then
  RANGE_ARGS=()
else
  RANGE_ARGS=("$RANGE")
fi

# Emit `### Section` plus bullets if any matching commits exist in range.
section() {
  local title="$1"
  local pattern="$2"
  local entries
  entries=$(git log "${RANGE_ARGS[@]}" --no-merges --pretty=format:'- %s (%h)' \
    | grep -E "$pattern" || true)
  if [ -n "$entries" ]; then
    printf '### %s\n\n%s\n\n' "$title" "$entries"
  fi
}

# Breaking changes section — surface BREAKING CHANGE bodies and `!` markers.
breaking_entries=$(git log "${RANGE_ARGS[@]}" --no-merges --pretty=format:'- %s (%h)' \
  | grep -E '^.*?[a-z]+(\([^)]+\))?!:' || true)
breaking_body=$(git log "${RANGE_ARGS[@]}" --no-merges --grep='BREAKING CHANGE' --pretty=format:'- %s (%h)' || true)
if [ -n "$breaking_entries$breaking_body" ]; then
  printf '### Breaking changes\n\n'
  [ -n "$breaking_entries" ] && printf '%s\n' "$breaking_entries"
  [ -n "$breaking_body" ] && [ "$breaking_body" != "$breaking_entries" ] && printf '%s\n' "$breaking_body"
  printf '\n'
fi

section 'Added' '^.*?feat(\([^)]+\))?!?:'
section 'Fixed' '^.*?fix(\([^)]+\))?!?:'
section 'Changed' '^.*?(perf|refactor)(\([^)]+\))?!?:'
section 'Reverted' '^.*?revert(\([^)]+\))?!?:'
