#!/usr/bin/env bash
# sync-vendored-skills.sh — vendor third-party skills verbatim into
# .claude/skills/<skill>/ as project skills.
#
# Each skill is owned upstream, so the copy here is byte-for-byte: the files
# MAP names, fetched at one pinned ref, plus an UPSTREAM file recording the
# repository and ref. check-skills.sh skips any skill directory holding
# UPSTREAM, because a file whose source of truth is another repository cannot
# be fixed by editing it here. Change a copy by moving its ref in MAP and
# re-running this script, never by hand.
#
# tech-writer amends humanizer by section number, so re-read its
# § references against the new SKILL.md whenever humanizer's ref moves.
#
# Usage
#   scripts/sync-vendored-skills.sh            # write the vendored copies
#   scripts/sync-vendored-skills.sh --check    # verify they match
#
# Exit status
#   0  in sync (or written); --check also exits 0, printing SKIP, when
#      upstream is unreachable
#   1  --check found a stale or missing copy
#   2  bad invocation, or a fetch failed while writing
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKILLS_DIR="$REPO_ROOT/.claude/skills"

# skill  <TAB>  owner/repo  <TAB>  ref  <TAB>  upstream path  <TAB>  path in skill
#
# Pin a tag where upstream publishes one, else a full commit SHA. Ship the
# license, and any notice the license or upstream attribution requires.
MAP="$(cat <<'EOF'
humanizer	blader/humanizer	v3.1.0	SKILL.md	SKILL.md
humanizer	blader/humanizer	v3.1.0	LICENSE	LICENSE
nbj-write-clearly	daniel-p-green/nbj-write-clearly	56e3a27f920bcd772f7b03fa476b5cfb9268332c	.agents/skills/nbj-write-clearly/SKILL.md	SKILL.md
nbj-write-clearly	daniel-p-green/nbj-write-clearly	56e3a27f920bcd772f7b03fa476b5cfb9268332c	.agents/skills/nbj-write-clearly/references/guide.md	references/guide.md
nbj-write-clearly	daniel-p-green/nbj-write-clearly	56e3a27f920bcd772f7b03fa476b5cfb9268332c	.agents/skills/nbj-write-clearly/references/official-index.md	references/official-index.md
nbj-write-clearly	daniel-p-green/nbj-write-clearly	56e3a27f920bcd772f7b03fa476b5cfb9268332c	LICENSE	LICENSE
nbj-write-clearly	daniel-p-green/nbj-write-clearly	56e3a27f920bcd772f7b03fa476b5cfb9268332c	NOTICE.md	NOTICE.md
EOF
)"

CHECK=0
case "${1-}" in
  --check) CHECK=1 ;;
  "")      ;;
  *)       echo "usage: $(basename "$0") [--check]" >&2; exit 2 ;;
esac

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

while IFS=$'\t' read -r skill repo ref src dest; do
  mkdir -p "$tmp/$skill/$(dirname "$dest")"
  url="https://raw.githubusercontent.com/$repo/$ref/$src"
  if ! curl -fsSL "$url" -o "$tmp/$skill/$dest"; then
    if [ "$CHECK" = 1 ]; then
      echo "SKIP: $url unreachable"; exit 0
    fi
    echo "sync-vendored-skills: could not fetch $url" >&2; exit 2
  fi
  printf 'repo: %s\nref: %s\n' "$repo" "$ref" > "$tmp/$skill/UPSTREAM"
done <<<"$MAP"

rc=0
for staged in "$tmp"/*/; do
  skill="$(basename "$staged")"
  if [ "$CHECK" = 1 ]; then
    # diff -r also catches a file added by hand next to the vendored ones.
    if ! diff -r "$staged" "$SKILLS_DIR/$skill" >/dev/null 2>&1; then
      echo "stale: .claude/skills/$skill differs from upstream (run $(basename "$0"))" >&2; rc=1
    else
      echo "$skill: in sync"
    fi
  else
    rm -rf "${SKILLS_DIR:?}/$skill"
    cp -rf "$staged" "$SKILLS_DIR/$skill"
    echo "$skill: written"
  fi
done
exit "$rc"
