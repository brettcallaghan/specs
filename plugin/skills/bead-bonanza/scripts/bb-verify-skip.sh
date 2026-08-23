#!/usr/bin/env bash
# bb-verify-skip.sh — check that a "skipped" claim is consistent with a clean
# working tree.
#
# Why this exists
#   A subagent that reports a bead as `skipped` while leaving edits in the tree
#   has produced a contradictory state: either it did work it did not record,
#   or it left the tree dirty for the next bead to trip over. Both sides of the
#   handoff have to check this — the subagent before it claims the skip, and
#   the orchestrator when the claim comes back. Two prose procedures describing
#   the same check drift; one script cannot.
#
# Usage
#   bb-verify-skip.sh [pathspec ...]     # no pathspec = whole working tree
#
# Exit status
#   0  clean — the skip is consistent, proceed
#   1  dirty — the skip contradicts the tree; caller must reconcile
#   2  bad invocation (not a git repo)

set -uo pipefail

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "bb-verify-skip: not inside a git work tree" >&2
  exit 2
fi

# Each git call's status is read, not just its stdout. A failing git — a
# pathspec that matches nothing tracked, a corrupt index — prints to stderr and
# returns EMPTY stdout, which is byte-for-byte what a clean tree returns. Left
# unchecked, the one case where the check could not run reads as the one case
# where it passed.
unstaged="$(git diff --stat -- "$@")"           || { echo "bb-verify-skip: git diff failed" >&2; exit 2; }
staged="$(git diff --staged --stat -- "$@")"    || { echo "bb-verify-skip: git diff --staged failed" >&2; exit 2; }
# Untracked files count as work too: a subagent that wrote a new file and
# claimed a skip is the same contradiction as one that edited an existing file.
untracked="$(git ls-files --others --exclude-standard -- "$@")" \
  || { echo "bb-verify-skip: git ls-files failed" >&2; exit 2; }

if [ -z "$unstaged" ] && [ -z "$staged" ] && [ -z "$untracked" ]; then
  echo "bb-verify-skip: CLEAN — no unstaged, staged, or untracked changes."
  echo "  The skip is consistent. Remember the summary line still needs a"
  echo "  citation: 'skipped - already addressed by <bead-id> @ <short-sha>'."
  exit 0
fi

echo "bb-verify-skip: DIRTY — a 'skipped' claim contradicts this tree." >&2
[ -n "$unstaged"  ] && { echo "  unstaged:" >&2;  printf '%s\n' "$unstaged"  | sed 's/^/    /' >&2; }
[ -n "$staged"    ] && { echo "  staged:" >&2;    printf '%s\n' "$staged"    | sed 's/^/    /' >&2; }
[ -n "$untracked" ] && { echo "  untracked:" >&2; printf '%s\n' "$untracked" | sed 's/^/    /' >&2; }
cat >&2 <<'EOF'

  Choose one — do NOT proceed to push or to the next bead until reconciled:

  (a) The work belongs to this bead. It is not a skip. Return to the
      gate step and finish the close ceremony as `completed`.

  (b) The changes are accidental, pre-existing, or out-of-scope. Revert
      them, then re-run this check until it reports CLEAN:
        git restore --staged <files> && git checkout -- <files>
      or, to keep them for inspection:
        git stash push -m "out-of-scope before skip: <bead-id>"
EOF
exit 1
