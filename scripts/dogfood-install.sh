#!/usr/bin/env bash
# dogfood-install.sh — make this repo use the plugin it develops.
#
# Why this exists
#   A session working in this repo should have these skills loaded, so the
#   framework is exercised by the people building it rather than only by
#   downstream projects. The catch is WHICH version gets exercised:
#
#     - Installing from the marketplace (or `scripts/install-spec-skill.sh`)
#       loads the *pushed* version. That tests packaging, but a session
#       editing `plugin/skills/` would not be running its own edits.
#     - Claude Code discovers project skills in `.claude/skills/`. Copying
#       the working tree there means the session runs exactly what is on
#       disk, which is what dogfooding has to mean while developing.
#
#   So this materializes the working tree into `.claude/skills/`. The copies
#   are generated and gitignored; `plugin/skills/` stays the only source.
#
# Usage
#   scripts/dogfood-install.sh            # materialize working-tree skills
#   scripts/dogfood-install.sh --check    # verify copies match the source
#   scripts/dogfood-install.sh --clean    # remove generated copies only
#
# Exit status
#   0  installed / in sync / cleaned
#   1  --check found a stale or missing copy
#   2  bad invocation

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="$REPO_ROOT/plugin/skills"
DEST="$REPO_ROOT/.claude/skills"
# Left behind in every generated directory so --clean can never delete a
# hand-authored project skill (e.g. triage-rfc, which lives here natively).
MARKER=".dogfood-generated"

MODE=install
case "${1-}" in
  --check) MODE=check ;;
  --clean) MODE=clean ;;
  "")      ;;
  *)       echo "usage: $(basename "$0") [--check|--clean]" >&2; exit 2 ;;
esac

[ -d "$SRC" ] || { echo "dogfood-install: no plugin/skills/ at $SRC" >&2; exit 2; }

# Vendored shared references must be current, or the dogfooded skills would
# read stale copies of spec-owned documents.
if [ "$MODE" != "clean" ]; then
  if ! "$REPO_ROOT/scripts/sync-shared-references.sh" --check >/dev/null 2>&1; then
    echo "dogfood-install: vendored references are out of date." >&2
    echo "  Run scripts/sync-shared-references.sh first." >&2
    exit 1
  fi
fi

if [ "$MODE" = clean ]; then
  n=0
  for d in "$DEST"/*/; do
    [ -f "$d/$MARKER" ] || continue
    rm -rf "$d"; n=$((n + 1))
  done
  echo "dogfood-install: removed $n generated skill(s)."
  exit 0
fi

mkdir -p "$DEST"
stale=0 count=0
for src in "$SRC"/*/; do
  [ -f "$src/SKILL.md" ] || continue
  skill="$(basename "$src")"
  dest="$DEST/$skill"
  count=$((count + 1))

  if [ -d "$dest" ] && [ ! -f "$dest/$MARKER" ]; then
    # A hand-authored project skill of the same name. Never clobber it.
    echo "  SKIP  $skill — a project-authored skill already exists here" >&2
    continue
  fi

  if [ "$MODE" = check ]; then
    if [ ! -d "$dest" ] || ! diff -rq --exclude="$MARKER" "$src" "$dest" >/dev/null 2>&1; then
      echo "  STALE $skill"
      stale=$((stale + 1))
    fi
  else
    rm -rf "$dest"
    cp -R "$src" "$dest"
    touch "$dest/$MARKER"
    echo "  $skill"
  fi
done

if [ "$MODE" = check ]; then
  if [ "$stale" -ne 0 ]; then
    echo
    echo "RESULT: $stale of $count dogfooded skill(s) out of date." >&2
    echo "        Run scripts/dogfood-install.sh to refresh." >&2
    exit 1
  fi
  echo "RESULT: all $count dogfooded skills match plugin/skills/."
  exit 0
fi

echo
echo "RESULT: materialized $count skill(s) into .claude/skills/ (generated, gitignored)."
echo "        plugin/skills/ remains the only source — edit there, re-run this."
