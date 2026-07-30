#!/usr/bin/env bash
# sync-shared-references.sh — vendor spec-owned reference files into the
# skills that consume them, so every skill's references live inside its own
# directory instead of behind a `../<sibling>/` path.
#
# Why this exists
#   A skill's bundled resources are supposed to sit under its own SKILL.md.
#   Cross-skill `../spec/references/...` paths break that: copy one skill
#   somewhere and its references dangle, and the paths are undiscoverable
#   from the reading skill's own tree. But duplicating the *authored* copies
#   is worse — that is how the two RFC templates drifted apart.
#
#   So: exactly one authored copy, N generated copies, and a check that
#   fails the build when they diverge. Each generated file carries a header
#   naming its source, so nobody edits it by mistake.
#
# Usage
#   scripts/sync-shared-references.sh            # write the vendored copies
#   scripts/sync-shared-references.sh --check    # verify they are current
#
# Exit status
#   0  in sync (or written successfully)
#   1  --check found a stale or missing copy
#   2  bad invocation / missing source file

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKILLS="$REPO_ROOT/plugin/skills"
VENDOR_DIR="references/shared"

CHECK=0
case "${1-}" in
  --check) CHECK=1 ;;
  "")      ;;
  *)       echo "usage: $(basename "$0") [--check]" >&2; exit 2 ;;
esac

# consumer-skill  <TAB>  owning-skill/references/file.md
# Keep this list in sync with what the SKILL.md files actually load; the
# consistency check cross-references it against their `references/shared/`
# citations.
MAP=$(cat <<'EOF'
bead-bonanza	spec/references/work-order-definition-of-ready.md
groom-beads	spec/references/work-order-definition-of-ready.md
groom-beads	spec/references/work-order-field-guide.md
groom-beads	spec/references/work-order-template.md
groom-feature-blueprint	spec/references/ac-traceability-standards.md
groom-feature-blueprint	spec/references/blueprint-readiness-checklist.md
groom-feature-blueprint	spec/references/feature-blueprint-template.md
groom-feature-blueprint	spec/references/work-order-template.md
groom-feature-blueprint	retro-spec/references/findings-capture-protocol.md
retro-skill	retro-spec/references/findings-capture-protocol.md
retro-spec	spec/references/template-authoring-standards.md
EOF
)

# Extract `version: "X.Y"` from a document's frontmatter, so the stamp records
# WHICH revision was vendored. Falls back to "unversioned" rather than failing:
# a few reference files are templates whose version lives in the instruction
# block instead, and a missing version is not a reason to block the build.
source_version() {
  local v
  v=$(sed -n '1,/^---$/p' "$1" | sed -n 's/^version:[[:space:]]*"\{0,1\}\([^"]*\)"\{0,1\}[[:space:]]*$/\1/p' | head -1)
  [ -n "$v" ] || v=$(sed -n 's/^Version:[[:space:]]*\(.*\)$/\1/p' "$1" | head -1)
  printf '%s' "${v:-unversioned}"
}

render() {  # $1 = source path, $2 = source rel path
  printf '<!-- GENERATED COPY — DO NOT EDIT.\n'
  printf '     Source of truth: plugin/skills/%s (version %s)\n' "$2" "$(source_version "$1")"
  printf '     Regenerate with: scripts/sync-shared-references.sh\n'
  printf '     Edit the source, then re-run that script. -->\n\n'
  cat "$1"
}

stale=0 wrote=0 checked=0
while IFS=$'\t' read -r consumer src_rel; do
  [ -n "$consumer" ] || continue
  src="$SKILLS/$src_rel"
  if [ ! -f "$src" ]; then
    echo "error: source not found: plugin/skills/$src_rel" >&2
    exit 2
  fi
  dest="$SKILLS/$consumer/$VENDOR_DIR/$(basename "$src_rel")"
  checked=$((checked + 1))

  if [ "$CHECK" -eq 1 ]; then
    if [ ! -f "$dest" ]; then
      echo "  STALE (missing)  $consumer/$VENDOR_DIR/$(basename "$src_rel")"
      stale=$((stale + 1))
    elif ! render "$src" "$src_rel" | diff -q - "$dest" >/dev/null 2>&1; then
      echo "  STALE (differs)  $consumer/$VENDOR_DIR/$(basename "$src_rel")"
      stale=$((stale + 1))
    fi
  else
    mkdir -p "$(dirname "$dest")"
    render "$src" "$src_rel" > "$dest"
    wrote=$((wrote + 1))
    echo "  $consumer/$VENDOR_DIR/$(basename "$src_rel")  <-  $src_rel"
  fi
done <<< "$MAP"

if [ "$CHECK" -eq 1 ]; then
  if [ "$stale" -ne 0 ]; then
    echo
    echo "RESULT: $stale of $checked vendored reference(s) out of date." >&2
    echo "        Run scripts/sync-shared-references.sh and commit the result." >&2
    exit 1
  fi
  echo "RESULT: all $checked vendored references are current."
  exit 0
fi

echo
echo "RESULT: wrote $wrote vendored reference(s)."
