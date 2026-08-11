#!/usr/bin/env bash
# bb-close-bead.sh — the per-bead close ceremony: atomic close, tracker export,
# stage, single commit.
#
# Why this exists
#   Every closed bead must land as exactly one commit that also carries the
#   tracker's durable export. Getting there has four ordering constraints that
#   are easy to state and easy to reorder under pressure:
#
#     1. close BEFORE staging, so the closure is inside the implementation
#        commit rather than trailing it as a second commit;
#     2. `bd export` must actually land in .beads/issues.jsonl — some builds
#        write to stdout — because an unexported closure is lost if the
#        container is reclaimed;
#     3. the export must be staged WITH the code, not after;
#     4. one commit, no --amend of anything pushed, no --no-verify.
#
#   The same ceremony appears in three subagent prompt templates and has to be
#   replayed by hand during crash recovery. One script means one behavior.
#
# Usage
#   bb-close-bead.sh <bead-id> <priority> <title> [path ...]
#
#   Paths are the code files this bead changed. WHEN omitted, everything
#   currently modified is staged — pass explicit paths whenever another agent
#   may be sharing the tree.
#
#   The close does NOT touch the bead's description: grooming wrote it, and
#   closing must not rewrite it. Findings belong in `retro:`-prefixed
#   `bd comment`s on the bead; the run directory's <bead-id>.handoff.md is an
#   artifact the orchestrator's acceptance check reads, not an input here.
#
# Exit status
#   0  bead closed, exported, committed (short sha on stdout)
#   1  a step failed — nothing is committed; the message says what to fix
#   2  bad invocation

set -uo pipefail

die()  { echo "bb-close-bead: $*" >&2; exit 2; }
fail() { echo "bb-close-bead: $*" >&2; exit 1; }

[ $# -ge 3 ] || die "usage: bb-close-bead.sh <bead-id> <priority> <title> [path ...]"
id="$1"; priority="$2"; title="$3"; shift 3

[ -n "$id" ]    || die "bead-id must not be empty"
[ -n "$title" ] || die "title must not be empty"
case "$priority" in ''|*[!0-4]*|??*) die "priority must be a single digit 0-4, got '$priority'" ;; esac
command -v bd >/dev/null 2>&1 || fail "\`bd\` not on PATH"
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "not inside a git work tree"

branch="$(git rev-parse --abbrev-ref HEAD)"
case "$branch" in
  main|master) fail "refusing to commit directly to '$branch' — switch to a working branch" ;;
esac

# 1. Atomic close — status only. `bd close` would be a second write and a
#    second commit; `--body-file` would overwrite the groomed description with
#    a hand-retyped copy, which is how descriptions used to lose content.
bd update "$id" --status closed \
  || fail "\`bd update $id --status closed\` failed — bead not closed, nothing staged"

# 2. Refresh the durable export. .beads/issues.jsonl is a passive export that
#    bd does not necessarily rewrite on every write, so confirm it changed
#    rather than assuming.
before_snapshot=""; [ -f .beads/issues.jsonl ] && before_snapshot="$(cat .beads/issues.jsonl)"
before=""; [ -f .beads/issues.jsonl ] && before="$(git hash-object .beads/issues.jsonl)"
if ! bd export > /tmp/bb-export.$$ 2>/dev/null; then
  # This build writes the file itself rather than to stdout.
  rm -f /tmp/bb-export.$$
  bd export >/dev/null 2>&1 || fail "\`bd export\` failed — closure is not durable, refusing to commit"
elif [ -s /tmp/bb-export.$$ ]; then
  mv /tmp/bb-export.$$ .beads/issues.jsonl
else
  rm -f /tmp/bb-export.$$
fi
after=""; [ -f .beads/issues.jsonl ] && after="$(git hash-object .beads/issues.jsonl)"
if [ "$before" = "$after" ]; then
  echo "bb-close-bead: warning — .beads/issues.jsonl did not change after export." >&2
  echo "  The close may not have reached the export. Verify with: bd show $id" >&2
fi

# Report the export delta: how many issue records actually changed. The
# tracker ledger is a shared-database projection (see
# references/operating-principles.md, "the shared-database race") — this
# bead's own status flip accounts for exactly one changed record, so a
# count above one is the signal that this commit is carrying another
# actor's tracker write. No jq dependency: extract "<id><TAB><line-hash>"
# per record with sed + `git hash-object` (already required by this
# script) and diff the two id/hash sets.
export_record_map() {
  while IFS= read -r line || [ -n "$line" ]; do
    rid="$(printf '%s' "$line" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')"
    [ -n "$rid" ] || continue
    printf '%s\t%s\n' "$rid" "$(printf '%s' "$line" | git hash-object --stdin)"
  done
}
changed_records=0
if [ -f .beads/issues.jsonl ]; then
  before_map="$(printf '%s\n' "$before_snapshot" | export_record_map | sort)"
  after_map="$(export_record_map < .beads/issues.jsonl | sort)"
  changed_records="$(comm -3 <(printf '%s\n' "$before_map") <(printf '%s\n' "$after_map") 2>/dev/null \
    | cut -f1 | sort -u | wc -l | tr -d ' ')"
fi
echo "bb-close-bead: export: ${changed_records:-0} record(s) changed (this bead: $id)"

# 3. Stage code + tracker together, so one commit carries both.
if [ $# -gt 0 ]; then
  git add -- "$@" || fail "git add failed for: $*"
else
  git add -A -- ':!.beads' || fail "git add failed"
fi
[ -f .beads/issues.jsonl ]      && git add .beads/issues.jsonl
[ -f .beads/interactions.jsonl ] && git add .beads/interactions.jsonl

git diff --staged --quiet && fail "nothing staged — no changes to commit for $id.
  If the bead needed no code change, this is a SKIP, not a close:
  run bb-verify-skip.sh and take the skipped path instead."

# 4. One commit. No --no-verify (hooks are a gate), no --amend (the commit may
#    already be pushed; a missed step becomes a follow-up commit instead).
git commit -q -m "$id P$priority $title" \
  || fail "git commit failed — hooks may have rejected it. Fix and re-run; the bead is already closed, so do not re-run bd update."

sha="$(git rev-parse --short HEAD)"
echo "$sha"
