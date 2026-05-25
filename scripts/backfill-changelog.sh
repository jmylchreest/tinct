#!/usr/bin/env bash
# backfill-changelog.sh — one-off cleanup. For each `## [X.Y.Z]` header
# in CHANGELOG.md whose body is empty (just header + date), generate
# conventional-commits sections from the appropriate `vPREV..vTHIS`
# range and inject them under the header.
#
# Idempotent: skips versions that already have a non-empty body.
# Safe: only writes if there's content to add.
#
# Usage:
#   scripts/backfill-changelog.sh                     # in-place edit
#   DRY_RUN=1 scripts/backfill-changelog.sh           # preview only
#
# Versions for which no conventional commits are found in their range
# are left untouched (still empty) — there's nothing to add.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CHANGELOG="${CHANGELOG:-CHANGELOG.md}"
DRY_RUN="${DRY_RUN:-0}"

# Extract every `## [X.Y.Z]` version header (skipping [Unreleased]) in
# the order they appear in the file. Bash array would be nicer; mapfile
# keeps the ordering.
mapfile -t versions < <(grep -E '^## \[[0-9]' "$CHANGELOG" | sed 's/^## \[\([^]]*\)\]$/\1/')

if [ ${#versions[@]} -eq 0 ]; then
  echo "No version headers found in $CHANGELOG"
  exit 0
fi

# Walk version pairs (this → previous = older). For each version, check
# if its body is empty; if so, backfill from the range.
updated=0
skipped_full=0
skipped_empty=0
for i in "${!versions[@]}"; do
  this_ver="${versions[$i]}"
  # The "previous" version is the next in the file (older). On the last
  # version, there's nothing older — leave it alone.
  next_i=$((i + 1))
  if [ "$next_i" -ge "${#versions[@]}" ]; then
    continue
  fi
  prev_ver="${versions[$next_i]}"
  this_tag="v${this_ver}"
  prev_tag="v${prev_ver}"

  # Check if this version's body is empty. Body = everything between
  # `## [this_ver]` and the next `## ` heading. Use exact string match
  # to avoid awk regex headaches with `[` / `.` in version numbers.
  body=$(awk -v target="## [${this_ver}]" '
    $0 == target { capture=1; next }
    /^## / { capture=0 }
    capture {
      # Strip the *DATE* line and blank lines for the emptiness check.
      if ($0 ~ /^\*[0-9-]+\*$/) next
      if ($0 ~ /^[[:space:]]*$/) next
      print
    }
  ' "$CHANGELOG")

  if [ -n "$body" ]; then
    skipped_full=$((skipped_full + 1))
    continue
  fi

  # Check both tags exist before invoking the section script.
  if ! git rev-parse --verify "$this_tag" >/dev/null 2>&1; then
    echo "skip $this_ver — tag $this_tag does not exist"
    continue
  fi
  if ! git rev-parse --verify "$prev_tag" >/dev/null 2>&1; then
    echo "skip $this_ver — tag $prev_tag does not exist"
    continue
  fi

  sections=$("$SCRIPT_DIR/changelog-sections.sh" "${prev_tag}..${this_tag}" || true)
  # Trim trailing blank lines.
  sections=$(printf '%s' "$sections" | sed -e :a -e '/^\s*$/{$d;N;ba' -e '}')

  if [ -z "$sections" ]; then
    echo "$this_ver: no qualifying commits in ${prev_tag}..${this_tag} — leaving empty"
    skipped_empty=$((skipped_empty + 1))
    continue
  fi

  echo "$this_ver: backfilling from ${prev_tag}..${this_tag}"
  if [ "$DRY_RUN" = "1" ]; then
    echo "--- would insert under ## [${this_ver}] ---"
    printf '%s\n' "$sections"
    echo "----------------------------------------"
    continue
  fi

  # Inject the sections under the version header. Use a tmp file as the
  # awk var trick from promote-unreleased.sh.
  TMP_SECTIONS=$(mktemp)
  printf '%s\n' "$sections" > "$TMP_SECTIONS"
  TMP_OUT=$(mktemp)
  awk -v target="## [${this_ver}]" -v sections_file="$TMP_SECTIONS" '
    BEGIN {
      while ((getline line < sections_file) > 0) sections = sections line "\n"
      close(sections_file)
    }
    {
      print
    }
    $0 == target {
      # After the header, also print the *date* line if present, then inject.
      getline next_line
      print next_line
      if (next_line ~ /^\*[0-9-]+\*$/) {
        print ""
        printf "%s", sections
      } else {
        printf "\n%s", sections
      }
    }
  ' "$CHANGELOG" > "$TMP_OUT"
  mv "$TMP_OUT" "$CHANGELOG"
  rm -f "$TMP_SECTIONS"
  updated=$((updated + 1))
done

echo
echo "Backfill summary:"
echo "  updated:        $updated"
echo "  already filled: $skipped_full"
echo "  no commits:     $skipped_empty"

if [ "$updated" -gt 0 ] && [ -f docs/docs/changelog.md ] && [ "$DRY_RUN" != "1" ]; then
  cp CHANGELOG.md docs/docs/changelog.md
  echo "  mirrored to docs/docs/changelog.md"
fi
