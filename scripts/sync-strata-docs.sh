#!/usr/bin/env bash
# sync-strata-docs.sh — vendor the Strata language documentation into the
# `strata` skill's references/, so the skill teaches the language as the
# strata repo actually documents it.
#
# Why this exists
#   The `strata` skill's value is the language reference it carries. That
#   reference is owned by strata-lang/strata, not by this repo — hand-copying
#   it here would fork it on the first upstream edit and quietly start
#   teaching a language that no longer matches the compiler.
#
#   So: no authored copy in this repo at all. Both reference files are
#   generated verbatim from the strata checkout, each stamped with the commit
#   that last touched its source, and `--check` reports when upstream has
#   moved. The stamp is per-file (`git log -1 -- <path>`) rather than repo
#   HEAD, so an unrelated strata commit does not make these look stale.
#
# Where the source comes from — first match wins:
#   $STRATA_SRC_DIR, ../strata (sibling checkout), ~/strata,
#   ~/.cache/strata-src (the clone .claude/hooks/session-start.sh maintains).
#
# Usage
#   scripts/sync-strata-docs.sh            # write the vendored copies
#   scripts/sync-strata-docs.sh --check    # verify they match the checkout
#
# Exit status
#   0  in sync (or written successfully)
#   1  --check found a stale or missing copy
#   2  bad invocation, or no strata checkout to sync from
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEST_DIR="$REPO_ROOT/.claude/skills/strata/references"

CHECK=0
case "${1-}" in
  --check) CHECK=1 ;;
  "")      ;;
  *)       echo "usage: $(basename "$0") [--check]" >&2; exit 2 ;;
esac

# source-path-in-strata-repo  <TAB>  destination-basename-here
MAP=$(cat <<'EOF'
docs/public/strata.llm.md	strata-llm-reference.md
docs/content/docs/lang.mdx	language-reference.md
EOF
)

# ---- locate a strata checkout ----------------------------------------------
STRATA_SRC=""
for candidate in "${STRATA_SRC_DIR:-}" "$REPO_ROOT/../strata" "$HOME/strata" \
                 "$HOME/.cache/strata-src"; do
  [ -n "$candidate" ] || continue
  [ -f "$candidate/crates/strata-cli/Cargo.toml" ] || continue
  STRATA_SRC="$(cd "$candidate" && pwd)"
  break
done

if [ -z "$STRATA_SRC" ]; then
  echo "sync-strata-docs: no strata checkout found." >&2
  echo "  Looked in: \$STRATA_SRC_DIR, ../strata, ~/strata, ~/.cache/strata-src" >&2
  echo "  Clone strata-lang/strata somewhere on that list, or set STRATA_SRC_DIR." >&2
  # --check is part of this repo's test suite, and a machine with no strata
  # checkout is not a defect in this repo — it just cannot answer the
  # question. Skipping beats a red suite nobody can turn green. Write mode
  # still fails: there, the missing checkout blocks what was asked for.
  [ "$CHECK" -eq 1 ] && { echo "RESULT: SKIP (nothing to compare against)."; exit 0; }
  exit 2
fi

# The commit that last touched THIS file, so the stamp only moves when the
# documentation moves. Degrades to "unknown" outside a git checkout rather
# than failing — a tarball copy of the docs is still worth vendoring.
source_stamp() {
  local sha date
  sha=$(git -C "$STRATA_SRC" log -1 --format='%H' -- "$1" 2>/dev/null)
  date=$(git -C "$STRATA_SRC" log -1 --format='%cs' -- "$1" 2>/dev/null)
  if [ -n "$sha" ]; then printf '%s (%s)' "${sha:0:12}" "$date"; else printf 'unknown'; fi
}

render() {  # $1 = source rel path
  printf '<!-- GENERATED COPY — DO NOT EDIT.\n'
  printf '     Source of truth: strata-lang/strata %s\n' "$1"
  printf '     Vendored from commit: %s\n' "$(source_stamp "$1")"
  printf '     Regenerate with: scripts/sync-strata-docs.sh\n'
  printf '     Upstream owns this text. Fix it there, then re-run that script. -->\n\n'
  cat "$STRATA_SRC/$1"
}

stale=0 wrote=0 checked=0
while IFS=$'\t' read -r src_rel dest_name; do
  [ -n "$src_rel" ] || continue
  if [ ! -f "$STRATA_SRC/$src_rel" ]; then
    echo "error: source not found: $STRATA_SRC/$src_rel" >&2
    exit 2
  fi
  dest="$DEST_DIR/$dest_name"
  checked=$((checked + 1))

  if [ "$CHECK" -eq 1 ]; then
    if [ ! -f "$dest" ]; then
      echo "  STALE (missing)  $dest_name"
      stale=$((stale + 1))
    elif ! render "$src_rel" | diff -q - "$dest" >/dev/null 2>&1; then
      echo "  STALE (differs)  $dest_name  <-  $src_rel"
      stale=$((stale + 1))
    fi
  else
    mkdir -p "$DEST_DIR"
    render "$src_rel" > "$dest"
    wrote=$((wrote + 1))
    echo "  $dest_name  <-  $src_rel @ $(source_stamp "$src_rel")"
  fi
done <<< "$MAP"

if [ "$CHECK" -eq 1 ]; then
  if [ "$stale" -ne 0 ]; then
    echo
    echo "RESULT: $stale of $checked vendored strata doc(s) out of date." >&2
    echo "        Run scripts/sync-strata-docs.sh and commit the result." >&2
    exit 1
  fi
  echo "RESULT: all $checked vendored strata docs match $STRATA_SRC."
  exit 0
fi

echo
echo "RESULT: wrote $wrote vendored doc(s) from $STRATA_SRC."
