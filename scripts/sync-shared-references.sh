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
#   `--check` also audits the reverse direction: every file in the repository
#   carrying that header must be one something actually regenerates. See the
#   orphan guard near the bottom for why that half is not optional.
#
# Usage
#   scripts/sync-shared-references.sh            # write the vendored copies
#   scripts/sync-shared-references.sh --check    # verify they are current
#
# Exit status
#   0  in sync (or written successfully)
#   1  --check found a stale copy, a missing copy, or an orphaned header
#   2  bad invocation / missing source file

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKILLS="$REPO_ROOT/plugin/skills"
VENDOR_DIR="references/shared"

# The path this script is invoked as, spelled once: `render` stamps it into
# every header it writes and the orphan guard matches headers against it, so
# the two cannot drift apart.
SELF_REL="scripts/sync-shared-references.sh"
HEADER_OPEN="<!-- GENERATED COPY"

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

# repo-relative destination  <TAB>  owning-skill/references/file.md
#
# Destinations that are not a consumer skill's `references/shared/`. MAP above
# derives its destination from the consumer's name; an entry here has no skill
# to derive one from, so it spells the path out in full.
#
# rfcs/findings-capture.md is this project's own copy of the findings-capture
# protocol — the one work orders and bead-bonanza send agents to read. It
# carried the GENERATED COPY header while no entry generated it, so `--check`
# passed over it and it drifted by hand instead (agent-plugins-92q.18 had to
# hand-edit it after the sync silently left it behind; agent-plugins-92q.28).
PATH_MAP=$(cat <<'EOF'
rfcs/findings-capture.md	retro-spec/references/findings-capture-protocol.md
EOF
)

# Both maps reduce to the same pair — repo-relative destination, source path
# under plugin/skills/ — and everything below works on that, so a destination
# outside plugin/skills/ is not a special case anywhere but here.
ENTRIES=$(
  while IFS=$'\t' read -r consumer src_rel; do
    [ -n "$consumer" ] || continue
    printf 'plugin/skills/%s/%s/%s\t%s\n' \
      "$consumer" "$VENDOR_DIR" "$(basename "$src_rel")" "$src_rel"
  done <<< "$MAP"
  while IFS=$'\t' read -r dest_rel src_rel; do
    [ -n "$dest_rel" ] || continue
    printf '%s\t%s\n' "$dest_rel" "$src_rel"
  done <<< "$PATH_MAP"
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
  printf '%s — DO NOT EDIT.\n' "$HEADER_OPEN"
  printf '     Source of truth: plugin/skills/%s (version %s)\n' "$2" "$(source_version "$1")"
  printf '     Regenerate with: %s\n' "$SELF_REL"
  printf '     Edit the source, then re-run that script. -->\n\n'
  cat "$1"
}

stale=0 wrote=0 checked=0
while IFS=$'\t' read -r dest_rel src_rel; do
  [ -n "$dest_rel" ] || continue
  src="$SKILLS/$src_rel"
  if [ ! -f "$src" ]; then
    echo "error: source not found: plugin/skills/$src_rel" >&2
    exit 2
  fi
  dest="$REPO_ROOT/$dest_rel"
  checked=$((checked + 1))

  if [ "$CHECK" -eq 1 ]; then
    if [ ! -f "$dest" ]; then
      echo "  STALE (missing)  $dest_rel"
      stale=$((stale + 1))
    elif ! render "$src" "$src_rel" | diff -q - "$dest" >/dev/null 2>&1; then
      echo "  STALE (differs)  $dest_rel"
      stale=$((stale + 1))
    fi
  else
    mkdir -p "$(dirname "$dest")"
    render "$src" "$src_rel" > "$dest"
    wrote=$((wrote + 1))
    echo "  $dest_rel  <-  $src_rel"
  fi
done <<< "$ENTRIES"

if [ "$CHECK" -eq 1 ]; then
  # ---- orphan guard: a GENERATED COPY header that nothing generates --------
  #
  # Two mechanisms are supposed to keep a vendored file honest: the header
  # ("DO NOT EDIT — regenerate with this script") and this --check. On a file
  # with the header but no MAP entry, BOTH fail open and point away from it.
  # The header tells readers not to touch it, --check never looks at it, and
  # the run still ends "all N vendored references are current" — so the file
  # drifts by hand and the only two things that could notice both report
  # health. rfcs/findings-capture.md sat in exactly that state (agent-plugins
  # -92q.28); the loop above now covers it, and this covers the next one.
  #
  # The rule: carrying the header obliges you to name a generator this repo
  # can actually run. Either this script maintains you and a MAP entry proves
  # it, or another script here does, or the header is a claim nobody honours.
  managed_dests=$(printf '%s\n' "$ENTRIES" | cut -f1)
  orphans=0 elsewhere=0

  # Tracked files only. This gate speaks for what is committed; an untracked
  # scratch copy in somebody's working tree is not the repo's drift. Falls
  # back to a filesystem walk outside a git checkout (an extracted tarball).
  if git -C "$REPO_ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    candidates=$(git -C "$REPO_ROOT" ls-files -- '*.md')
  else
    candidates=$(cd "$REPO_ROOT" && find . -name '*.md' -not -path './.git/*' | sed 's|^\./||')
  fi

  while IFS= read -r rel; do
    [ -n "$rel" ] || continue
    [ -f "$REPO_ROOT/$rel" ] || continue
    first=""; IFS= read -r first < "$REPO_ROOT/$rel" || true
    case "$first" in "$HEADER_OPEN"*) ;; *) continue ;; esac
    if printf '%s\n' "$managed_dests" | grep -qxF -- "$rel"; then
      continue    # generated by the loop above, and diffed there
    fi

    regen=$(sed -n '1,10{s|^[[:space:]]*Regenerate with:[[:space:]]*\([^[:space:]]*\).*|\1|p;}' \
              "$REPO_ROOT/$rel" | head -1)
    if [ "$regen" = "$SELF_REL" ]; then
      echo "  ORPHAN  $rel" >&2
      echo "          header names $SELF_REL, but no MAP entry generates it." >&2
      echo "          Add it to MAP or PATH_MAP, or rewrite the header to say" >&2
      echo "          how the file is really maintained." >&2
      orphans=$((orphans + 1))
    elif [ -n "$regen" ] && [ -x "$REPO_ROOT/$regen" ]; then
      elsewhere=$((elsewhere + 1))    # another script here owns it; its --check diffs it
    else
      echo "  ORPHAN  $rel" >&2
      echo "          carries a GENERATED COPY header naming no generator this" >&2
      echo "          repo can run${regen:+ (\"$regen\")}. Wire it to one, or drop the header." >&2
      orphans=$((orphans + 1))
    fi
  done <<< "$candidates"

  if [ "$stale" -ne 0 ] || [ "$orphans" -ne 0 ]; then
    echo
    if [ "$stale" -ne 0 ]; then
      echo "RESULT: $stale of $checked vendored reference(s) out of date." >&2
      echo "        Run scripts/sync-shared-references.sh and commit the result." >&2
    fi
    if [ "$orphans" -ne 0 ]; then
      echo "RESULT: $orphans file(s) claim to be generated copies that nothing regenerates." >&2
    fi
    exit 1
  fi
  echo "RESULT: all $checked vendored references are current."
  echo "        $elsewhere further generated file(s) name a generator this repo runs; 0 unmanaged."
  exit 0
fi

echo
echo "RESULT: wrote $wrote vendored reference(s)."
