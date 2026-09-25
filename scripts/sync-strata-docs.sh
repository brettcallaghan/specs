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
#   generated verbatim from the strata RELEASE this repo builds with — the
#   one scripts/install-strata-cli.sh installs — and stamped with its tag.
#   The docs therefore describe the compiler on PATH, not whatever upstream
#   `main` says today, and `--check` goes stale only when the release does.
#
# Which release — first match wins:
#   the TAG argument, $STRATA_VERSION, or the installed `strata --version`.
#
# Where the files come from — first match wins:
#   a strata checkout holding that tag ($STRATA_SRC_DIR, ../strata, ~/strata),
#   else the GitHub API (needs $STRATA_ACCESS_TOKEN while strata is private).
#
# Each MAP entry names a path SUFFIX, resolved against the release's tree, so
# a file that upstream moves between directories is still found. A suffix
# matching no file, or more than one, is an error naming the entry to fix.
#
# Usage
#   scripts/sync-strata-docs.sh [TAG]            # write the vendored copies
#   scripts/sync-strata-docs.sh --check [TAG]    # verify they match
#
# Exit status
#   0  in sync (or written successfully); --check also exits 0, printing
#      SKIP, when no release or source is reachable
#   1  --check found a stale or missing copy
#   2  bad invocation, no source to sync from, or a MAP entry that does not
#      resolve in the release
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEST_DIR="$REPO_ROOT/.claude/skills/strata/references"
API="https://api.github.com/repos/strata-lang/strata"

CHECK=0 TAG=""
for arg in "$@"; do
  case "$arg" in
    --check) CHECK=1 ;;
    v[0-9]*) TAG="$arg" ;;
    *)       echo "usage: $(basename "$0") [--check] [TAG]" >&2; exit 2 ;;
  esac
done

# path-suffix  <TAB>  destination-basename  <TAB>  mode  <TAB>  title
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
public/strata.llm.md	strata-llm-reference.md	raw
content/docs/lang.mdx	language-reference.md	raw
strata/options.strata	compiler-options.md	strata-fence	Compiler options — the schema of `strata.config.json`
EOF
)

# --check is part of this repo's test suite, and a machine that cannot reach
# the release is not a defect in this repo — it just cannot answer the
# question. Skipping beats a red suite nobody can turn green. Write mode
# still fails: there, the missing source blocks what was asked for.
unreachable() {
  echo "sync-strata-docs: $1" >&2
  if [ "$CHECK" -eq 1 ]; then
    echo "RESULT: SKIP (nothing to compare against)."
    exit 0
  fi
  exit 2
}

# ---- which release ---------------------------------------------------------
TAG="${TAG:-${STRATA_VERSION:-}}"
if [ -z "$TAG" ] || [ "$TAG" = latest ]; then
  version="$(strata --version 2>/dev/null || true)"   # "strata-cli 0.8.0"
  [ -n "$version" ] || unreachable "no strata on PATH to take the release from — pass a TAG or run scripts/install-strata-cli.sh."
  TAG="v${version##* }"
fi

# ---- where the files come from ----------------------------------------------
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

STRATA_SRC=""
for candidate in "${STRATA_SRC_DIR:-}" "$REPO_ROOT/../strata" "$HOME/strata"; do
  [ -n "$candidate" ] || continue
  if git -C "$candidate" rev-parse -q --verify "$TAG^{commit}" >/dev/null 2>&1; then
    STRATA_SRC="$(cd "$candidate" && pwd)"
    break
  fi
done

# GitHub API request. Tried through the environment's proxy settings first,
# then directly: an egress proxy that brokers its own GitHub credential
# replaces the Authorization header and answers 403 for strata.
gh_curl() {
  local out="$1" url="$2" accept="$3"
  local args=(-fsSL --retry 2 -o "$out" -H "Accept: $accept"
              -H "X-GitHub-Api-Version: 2022-11-28")
  [ -n "${STRATA_ACCESS_TOKEN:-}" ] \
    && args+=(-H "Authorization: Bearer $STRATA_ACCESS_TOKEN")
  curl "${args[@]}" "$url" 2>/dev/null \
    || curl --noproxy '*' "${args[@]}" "$url" 2>/dev/null
}

if [ -n "$STRATA_SRC" ]; then
  SOURCE_DESC="$STRATA_SRC"
  git -C "$STRATA_SRC" ls-tree -r --name-only "$TAG" >"$WORK/tree"
else
  SOURCE_DESC="the GitHub API"
  gh_curl "$WORK/tree.json" "$API/git/trees/$TAG?recursive=1" application/vnd.github+json \
    || unreachable "no local checkout holds $TAG, and the GitHub API did not answer for it (is STRATA_ACCESS_TOKEN set?)."
  python3 -c 'import json,sys
for e in json.load(open(sys.argv[1]))["tree"]:
    if e["type"] == "blob": print(e["path"])' "$WORK/tree.json" >"$WORK/tree"
fi

# The one path in the release ending in /<suffix>, or a named error.
resolve() {
  local suffix="$1" dest="$2" matches
  matches="$(grep -E "(^|/)${suffix//./\\.}\$" "$WORK/tree" || true)"
  if [ -z "$matches" ]; then
    echo "error: strata-lang/strata $TAG has no file ending in '$suffix' (vendored as $dest)." >&2
    echo "  Upstream moved or renamed it. Find its new path in $TAG and update" >&2
    echo "  that MAP entry in scripts/sync-strata-docs.sh." >&2
    exit 2
  fi
  if [ "$(printf '%s\n' "$matches" | wc -l)" -gt 1 ]; then
    echo "error: '$suffix' (vendored as $dest) matches more than one file in $TAG:" >&2
    printf '  %s\n' $matches >&2
    echo "  Lengthen that MAP entry's suffix in scripts/sync-strata-docs.sh." >&2
    exit 2
  fi
  printf '%s' "$matches"
}

# Fetch one release file into $WORK, once.
fetch() {
  local path="$1" out="$WORK/src/$1"
  if [ ! -f "$out" ]; then
    mkdir -p "$(dirname "$out")"
    if [ -n "$STRATA_SRC" ]; then
      git -C "$STRATA_SRC" show "$TAG:$path" >"$out"
    else
      gh_curl "$out" "$API/contents/$path?ref=$TAG" application/vnd.github.raw \
        || { echo "error: could not fetch $path at $TAG from the GitHub API." >&2; exit 2; }
    fi
  fi
  printf '%s' "$out"
}

render() {  # $1 = source path in the release, $2 = mode, $3 = title (strata-fence only)
  local file
  file="$(fetch "$1")"
  printf '<!-- GENERATED COPY — DO NOT EDIT.\n'
  printf '     Source of truth: strata-lang/strata %s\n' "$1"
  printf '     Vendored from release: %s\n' "$TAG"
  printf '     Regenerate with: scripts/sync-strata-docs.sh\n'
  printf '     Upstream owns this text. Fix it there, then re-run that script. -->\n\n'
  if [ "$2" = strata-fence ]; then
    printf '# %s\n\n' "$3"
    printf 'Verbatim `%s` from strata-lang/strata. Every field states its\n' "$(basename "$1")"
    printf 'default in its doc block.\n\n'
    printf '```strata\n'
    cat "$file"
    # Keep the closing fence on its own line even if the source has no
    # trailing newline. Spelled as an if because `[ ] && printf` would abort
    # the script under `set -e` on the false branch.
    if [ -n "$(tail -c1 "$file")" ]; then printf '\n'; fi
    printf '```\n'
  else
    cat "$file"
  fi
}

stale=0 wrote=0 checked=0
while IFS=$'\t' read -r suffix dest_name mode title; do
  [ -n "$suffix" ] || continue
  src_rel="$(resolve "$suffix" "$dest_name")"
  dest="$DEST_DIR/$dest_name"
  checked=$((checked + 1))
  render "$src_rel" "$mode" "${title:-}" >"$WORK/rendered"

  if [ "$CHECK" -eq 1 ]; then
    if [ ! -f "$dest" ]; then
      echo "  STALE (missing)  $dest_name"
      stale=$((stale + 1))
    elif ! diff -q "$WORK/rendered" "$dest" >/dev/null 2>&1; then
      echo "  STALE (differs)  $dest_name  <-  $src_rel @ $TAG"
      stale=$((stale + 1))
    fi
  else
    mkdir -p "$DEST_DIR"
    cp -f "$WORK/rendered" "$dest"
    wrote=$((wrote + 1))
    echo "  $dest_name  <-  $src_rel @ $TAG"
  fi
done <<< "$MAP"

if [ "$CHECK" -eq 1 ]; then
  if [ "$stale" -ne 0 ]; then
    echo
    echo "RESULT: $stale of $checked vendored strata doc(s) do not match $TAG." >&2
    echo "        Run scripts/sync-strata-docs.sh and commit the result." >&2
    exit 1
  fi
  echo "RESULT: all $checked vendored strata docs match $TAG (via $SOURCE_DESC)."
  exit 0
fi

echo
echo "RESULT: wrote $wrote vendored doc(s) from $TAG (via $SOURCE_DESC)."
