#!/usr/bin/env bash
# sync-shared-references.sh — vendor authored reference documents into the
# skills that consume them, so every skill's references live inside its own
# directory instead of behind a `../<sibling>/` path.
#
# Why this exists
#   A skill's bundled resources are supposed to sit under its own SKILL.md.
#   Cross-skill `../spec/references/...` paths break that: copy one skill
#   somewhere and its references dangle, and the paths are undiscoverable
#   from the reading skill's own tree. But duplicating the *authored* copies
#   is worse: two hand-kept copies of one template drift apart.
#
#   So: exactly one authored copy, N generated copies, and a check that
#   fails the build when they diverge. Each generated file carries a header
#   naming its source, so nobody edits it by mistake.
#
#   A MAP source is either a `specs/...` path (the authored copy is a
#   corpus member under `specs/`, and every skill that ships it — including
#   the skill that used to author it directly — receives a generated copy)
#   or a `<skill>/references/...` path (the authored copy is still a
#   template or prose file that stays inside a skill's own tree, not a
#   corpus member). Editing a generated copy is always wrong either way:
#   fix the source the header names, then re-run this script.
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

# consumer-skill  <TAB>  source path — `specs/...` for a corpus-member
# standard/guide, or `owning-skill/references/file.md` for a template or
# prose file that still lives inside a skill's own tree.
# Keep this list in sync with what the SKILL.md files actually load; the
# consistency check cross-references it against their `references/shared/`
# citations.
MAP=$(cat <<'EOF'
groom-beads	specs/standards/work-order-definition-of-ready.md
groom-beads	specs/guides/work-order-field-guide.md
groom-beads	spec/references/work-order-template.md
groom-feature-blueprint	specs/standards/ac-traceability-standards.md
groom-feature-blueprint	specs/standards/blueprint-readiness-checklist.md
groom-feature-blueprint	spec/references/feature-blueprint-template.md
groom-feature-blueprint	spec/references/work-order-template.md
groom-feature-blueprint	specs/standards/findings-capture-protocol.md
retro-skill	specs/standards/findings-capture-protocol.md
retro-skill	spec/references/issue-template.md
retro-spec	specs/standards/findings-capture-protocol.md
retro-spec	spec/references/issue-template.md
spec	specs/standards/ac-traceability-standards.md
spec	specs/standards/blueprint-readiness-checklist.md
spec	specs/standards/blueprint-content-standards.md
spec	specs/standards/c4-standards.md
spec	specs/standards/cross-document-references-standards.md
spec	specs/standards/document-readiness-checklist.md
spec	specs/standards/frontmatter-standards.md
spec	specs/standards/template-authoring-standards.md
spec	specs/standards/work-order-definition-of-ready.md
spec	specs/guides/anti-patterns-guide.md
spec	specs/guides/design-principles-guide.md
spec	specs/guides/design-principles-systems-guide.md
spec	specs/guides/design-principles-containers-guide.md
spec	specs/guides/design-principles-components-guide.md
spec	specs/guides/document-type-boundaries-guide.md
spec	specs/guides/ears-guide.md
spec	specs/standards/gherkin-guide.md
spec	specs/guides/spec-cli-reference.md
spec	specs/guides/work-order-field-guide.md
EOF
)

# repo-relative destination  <TAB>  owning-skill/references/file.md
#
# Destinations that are not a consumer skill's `references/shared/`. MAP above
# derives its destination from the consumer's name; an entry here has no skill
# to derive one from, so it spells the path out in full.
#
# Empty is a valid state, and the current one: no vendored copy outside a
# consumer skill's `references/shared/` is wanted right now. Any file added
# here MUST be one this repo genuinely needs at a fixed path — a copy placed
# somewhere merely so readers can find it is a second thing to keep current
# saying what the first already says.
PATH_MAP=$(cat <<'EOF'
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

# A MAP/PATH_MAP source is repo-root-relative when it starts with `specs/`
# (a corpus member) and plugin/skills-relative otherwise (a template or
# prose file still authored inside a skill's own tree). Both forms reduce
# to an absolute path and a repo-relative label here, so nothing below
# this pair needs to know which kind a given entry is.
src_abs() {  # $1 = source rel path as stored in MAP/PATH_MAP
  case "$1" in
    specs/*) printf '%s/%s' "$REPO_ROOT" "$1" ;;
    *)       printf '%s/%s' "$SKILLS" "$1" ;;
  esac
}
src_label() {  # $1 = source rel path as stored in MAP/PATH_MAP
  case "$1" in
    specs/*) printf '%s' "$1" ;;
    *)       printf 'plugin/skills/%s' "$1" ;;
  esac
}

render() {  # $1 = source path, $2 = source rel path
  printf '%s — DO NOT EDIT.\n' "$HEADER_OPEN"
  printf '     Source of truth: %s\n' "$(src_label "$2")"
  printf '     Regenerate with: %s\n' "$SELF_REL"
  printf '     Edit the source, then re-run that script. -->\n\n'
  cat "$1"
}

stale=0 wrote=0 checked=0
while IFS=$'\t' read -r dest_rel src_rel; do
  [ -n "$dest_rel" ] || continue
  src="$(src_abs "$src_rel")"
  if [ ! -f "$src" ]; then
    echo "error: source not found: $(src_label "$src_rel")" >&2
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
  # health.
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
    # Herestring, not `printf ... | grep -qxF`: under `set -o pipefail`, a
    # match makes grep -q exit as soon as it finds one, closing the pipe out
    # from under printf — printf then dies on SIGPIPE, pipefail reports the
    # whole pipeline as failed, and this `if` reads a real match as "not
    # managed", declaring a correctly-mapped file an ORPHAN.
    if grep -qxF -- "$rel" <<< "$managed_dests"; then
      continue    # generated by the loop above, and diffed there
    fi

    # No `| head -1` on a live pipe: under `set -o pipefail` a SIGPIPE'd
    # sed would trip `set -e`. Take the first line in the shell instead.
    regen=$(sed -n '1,10{s|^[[:space:]]*Regenerate with:[[:space:]]*\([^[:space:]]*\).*|\1|p;}' \
              "$REPO_ROOT/$rel")
    regen=${regen%%$'\n'*}
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
