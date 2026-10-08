#!/usr/bin/env bash
#
# Install (or update) this repo's skills (spec, groom-feature-blueprint,
# retro-spec, ...) into the user-level Claude Code configuration so they
# are available across all projects.
#
# Skills are self-contained: shared reference files are vendored into each
# consumer's references/shared/ by scripts/sync-shared-references.sh, so no
# skill reaches outside its own directory. Run that script (or its --check
# mode) before installing if the shared sources have changed.
#
# If a skill is renamed or removed from plugin/skills/, its old
# installation under $DEST_ROOT is cleaned up automatically — but only if
# it still carries the marker this script leaves behind, so a
# same-named skill installed by other means is never touched.
#
# Usage: scripts/install-spec-skill.sh
#   CLAUDE_CONFIG_DIR overrides the target config dir (default: ~/.claude)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC_ROOT="$REPO_ROOT/plugin/skills"
DEST_ROOT="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/skills"
MARKER=".specs-installed"
MANIFEST="$DEST_ROOT/.specs-manifest"

if ! compgen -G "$SRC_ROOT/*/SKILL.md" > /dev/null; then
  echo "error: no skills found under $SRC_ROOT — run from a checkout of specs" >&2
  exit 1
fi

# Defensive: if a reference ever becomes a symlink again, fail loudly
# rather than silently installing a broken link.
broken=0
while IFS= read -r -d '' link; do
  if [[ ! -e "$link" ]]; then
    echo "error: broken reference symlink: $link" >&2
    broken=1
  fi
done < <(find "$SRC_ROOT" -type l -print0)
if [[ $broken -ne 0 ]]; then
  exit 1
fi

mkdir -p "$DEST_ROOT"

current=()
for src in "$SRC_ROOT"/*/; do
  [[ -f "$src/SKILL.md" ]] || continue
  current+=("$(basename "$src")")
done

# Remove previously-installed skills that are no longer in the repo
# (e.g. renamed), as long as this script installed them originally.
if [[ -f "$MANIFEST" ]]; then
  while IFS= read -r name; do
    [[ -z "$name" ]] && continue
    [[ -f "$DEST_ROOT/$name/$MARKER" ]] || continue
    kept=0
    for c in "${current[@]}"; do
      [[ "$c" == "$name" ]] && kept=1 && break
    done
    if [[ $kept -eq 0 ]]; then
      rm -rf "$DEST_ROOT/$name"
      echo "Removed stale $name (renamed or removed upstream)"
    fi
  done < "$MANIFEST"
fi

installed=()
for src in "$SRC_ROOT"/*/; do
  skill="$(basename "$src")"
  [[ -f "$src/SKILL.md" ]] || continue
  dest="$DEST_ROOT/$skill"
  rm -rf "$dest"
  # -L dereferences symlinks, in case one reappears in the source tree
  cp -RL "$src" "$dest"
  touch "$dest/$MARKER"
  installed+=("$skill")
  echo "Installed $skill skill -> $dest"
done

printf '%s\n' "${installed[@]}" > "$MANIFEST"

echo
echo "Contents:"
for skill in "${installed[@]}"; do
  find "$DEST_ROOT/$skill" -type f -not -name "$MARKER" | sed "s|^$DEST_ROOT/|  |" | sort
done

# Claude Code only auto-allows reads inside a project's working
# directory, so skills reading their own reference files from here
# prompt for permission. No plugin-side mechanism can grant this
# (anthropics/claude-code#31036); the fix is a user permission rule.
echo
echo "To skip permission prompts when skills read their reference files,"
echo "add this rule to ~/.claude/settings.json under permissions.allow:"
echo "  \"Read($DEST_ROOT/**)\""
