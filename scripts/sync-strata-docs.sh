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
#   So: no authored copy in this repo at all. Every reference file is
#   generated verbatim from the strata checkout, each stamped with the commit
#   that last touched its source, and `--check` reports when upstream has
#   moved. The stamp is per-file (`git log -1 -- <path>`) rather than repo
#   HEAD, so an unrelated strata commit does not make these look stale.
#
# Where the source comes from — first match wins:
#   $STRATA_SRC_DIR, ../strata (sibling checkout), ~/strata.
#
# Usage
#   scripts/sync-strata-docs.sh            # write the vendored copies
#   scripts/sync-strata-docs.sh --check    # verify they match the checkout
#
# Exit status
#   0  in sync (or written successfully)
#   1  --check found a stale or missing copy
#   2  bad invocation, or no strata checkout to sync from
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEST_DIR="$REPO_ROOT/.claude/skills/strata/references"

CHECK=0
case "${1-}" in
  --check) CHECK=1 ;;
  "")      ;;
  *)       echo "usage: $(basename "$0") [--check]" >&2; exit 2 ;;
esac

# source-path  <TAB>  destination-basename  <TAB>  mode  <TAB>  title
#
# mode `raw`          — copy the source through unchanged (it is already markdown,
#                       and carries its own heading, so `title` is left empty).
# mode `strata-fence` — wrap the source in a ```strata block under `title`. Needed
#                       for `.strata` sources: the language has no comment syntax,
#                       so the provenance header cannot live inside one, and a bare
#                       copy would look like a compilable input to be edited.
#
# The title is what an agent sees when it opens the file directly, so it names
# what the schema IS to a user — not merely which upstream file it came from.
MAP=$(cat <<'EOF'
docs/public/strata.llm.md	strata-llm-reference.md	raw
docs/content/docs/lang.mdx	language-reference.md	raw
strata/options.strata	compiler-options.md	strata-fence	Compiler options — the schema of `strata.config.json`
EOF
)

# ---- locate a strata checkout ----------------------------------------------
STRATA_SRC=""
for candidate in "${STRATA_SRC_DIR:-}" "$REPO_ROOT/../strata" "$HOME/strata"; do
  [ -n "$candidate" ] || continue
  [ -f "$candidate/crates/strata-cli/Cargo.toml" ] || continue
  STRATA_SRC="$(cd "$candidate" && pwd)"
  break
done

if [ -z "$STRATA_SRC" ]; then
  echo "sync-strata-docs: no strata checkout found." >&2
  echo "  Looked in: \$STRATA_SRC_DIR, ../strata, ~/strata" >&2
  echo "  Clone strata-lang/strata somewhere on that list, or set STRATA_SRC_DIR." >&2
  # --check is part of this repo's test suite, and a machine with no strata
  # checkout is not a defect in this repo — it just cannot answer the
  # question. Skipping beats a red suite nobody can turn green. Write mode
  # still fails: there, the missing checkout blocks what was asked for.
  # Spelled as an if, not `[ ... ] && { ... }`: under `set -e` the && form
  # aborts the script when the test is false, so write mode would exit 1
  # here instead of reaching the exit 2 below.
  if [ "$CHECK" -eq 1 ]; then
    echo "RESULT: SKIP (nothing to compare against)."
    exit 0
  fi
  exit 2
fi

# The commit that last touched THIS file, so the stamp only moves when the
# documentation moves. Degrades to "unknown" outside a git checkout rather
# than failing — a tarball copy of the docs is still worth vendoring.
source_stamp() {
  local sha date
  # `|| true` matters under `set -e`: git exits 128 when STRATA_SRC is not a
  # git checkout at all (an extracted tarball, a vendored tree), and a bare
  # failing substitution in an assignment aborts the script — which is the
  # opposite of the graceful "unknown" this function promises.
  sha=$(git -C "$STRATA_SRC" log -1 --format='%H' -- "$1" 2>/dev/null || true)
  date=$(git -C "$STRATA_SRC" log -1 --format='%cs' -- "$1" 2>/dev/null || true)
  if [ -n "$sha" ]; then printf '%s (%s)' "${sha:0:12}" "$date"; else printf 'unknown'; fi
}

render() {  # $1 = source rel path, $2 = mode, $3 = title (strata-fence only)
  printf '<!-- GENERATED COPY — DO NOT EDIT.\n'
  printf '     Source of truth: strata-lang/strata %s\n' "$1"
  printf '     Vendored from commit: %s\n' "$(source_stamp "$1")"
  printf '     Regenerate with: scripts/sync-strata-docs.sh\n'
  printf '     Upstream owns this text. Fix it there, then re-run that script. -->\n\n'
  if [ "$2" = strata-fence ]; then
    printf '# %s\n\n' "$3"
    printf 'Verbatim `%s` from strata-lang/strata. Every field states its\n' "$(basename "$1")"
    printf 'default in its doc block.\n\n'
    printf '```strata\n'
    cat "$STRATA_SRC/$1"
    # Keep the closing fence on its own line even if the source has no
    # trailing newline. Spelled as an if because `[ ] && printf` would abort
    # the script under `set -e` on the false branch.
    if [ -n "$(tail -c1 "$STRATA_SRC/$1")" ]; then printf '\n'; fi
    printf '```\n'
  else
    cat "$STRATA_SRC/$1"
  fi
}

stale=0 wrote=0 checked=0
while IFS=$'\t' read -r src_rel dest_name mode title; do
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
    elif ! render "$src_rel" "$mode" "${title:-}" | diff -q - "$dest" >/dev/null 2>&1; then
      echo "  STALE (differs)  $dest_name  <-  $src_rel"
      stale=$((stale + 1))
    fi
  else
    mkdir -p "$DEST_DIR"
    render "$src_rel" "$mode" "${title:-}" > "$dest"
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
