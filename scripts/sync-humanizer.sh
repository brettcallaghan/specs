#!/usr/bin/env bash
# sync-humanizer.sh — vendor blader/humanizer (MIT) into
# .claude/skills/humanizer/ as a project skill.
#
# The skill is owned upstream, so the copy here is verbatim: SKILL.md and
# LICENSE from one pinned tag, plus an UPSTREAM file naming that tag.
# check-skills.sh skips any skill directory holding UPSTREAM, because a file
# whose source of truth is another repository cannot be fixed by editing it
# here. Change the copy by bumping the tag, never by hand.
#
# humanizer-tech-writer amends humanizer by section number, so re-read its
# § references against the new SKILL.md whenever the tag moves.
#
# Usage
#   scripts/sync-humanizer.sh [TAG]            # write the vendored copy
#   scripts/sync-humanizer.sh --check [TAG]    # verify it matches
#
# TAG defaults to the one recorded in UPSTREAM.
#
# Exit status
#   0  in sync (or written); --check also exits 0, printing SKIP, when
#      upstream is unreachable
#   1  --check found a stale or missing copy
#   2  bad invocation, or the fetch failed while writing
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEST="$REPO_ROOT/.claude/skills/humanizer"
REPO="blader/humanizer"
FILES=(SKILL.md LICENSE)

CHECK=0 TAG=""
for arg in "$@"; do
  case "$arg" in
    --check) CHECK=1 ;;
    v[0-9]*) TAG="$arg" ;;
    *)       echo "usage: $(basename "$0") [--check] [TAG]" >&2; exit 2 ;;
  esac
done

if [ -z "$TAG" ]; then
  TAG="$(sed -n 's/^tag: //p' "$DEST/UPSTREAM" 2>/dev/null || true)"
  [ -n "$TAG" ] || { echo "sync-humanizer: no TAG given and none in $DEST/UPSTREAM" >&2; exit 2; }
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

for f in "${FILES[@]}"; do
  url="https://raw.githubusercontent.com/$REPO/$TAG/$f"
  if ! curl -fsSL "$url" -o "$tmp/$f"; then
    if [ "$CHECK" = 1 ]; then
      echo "SKIP: $url unreachable"; exit 0
    fi
    echo "sync-humanizer: could not fetch $url" >&2; exit 2
  fi
done
printf 'repo: %s\ntag: %s\nlicense: MIT\n' "$REPO" "$TAG" > "$tmp/UPSTREAM"

if [ "$CHECK" = 1 ]; then
  rc=0
  for f in "${FILES[@]}" UPSTREAM; do
    if ! cmp -s "$tmp/$f" "$DEST/$f"; then
      echo "stale: .claude/skills/humanizer/$f differs from $REPO@$TAG" >&2; rc=1
    fi
  done
  [ "$rc" = 0 ] && echo "humanizer: in sync with $REPO@$TAG"
  exit "$rc"
fi

mkdir -p "$DEST"
for f in "${FILES[@]}" UPSTREAM; do cp -f "$tmp/$f" "$DEST/$f"; done
echo "humanizer: wrote $REPO@$TAG to .claude/skills/humanizer/"
