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
#        container is reclaimed, and the export lags the database often enough
#        that it has to be verified rather than assumed;
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
#   1  a step failed — nothing is committed; the message says what to fix.
#      One case to read carefully: the bead is closed in the database but its
#      closure never reached .beads/issues.jsonl. The bead IS closed — the
#      repair is another export plus the commit, never a second `bd update`.
#   2  bad invocation
#
# Tuning (testing knobs; they change how long the verification waits, never
# its verdict)
#   BB_EXPORT_ATTEMPTS     export attempts before deciding   (default 3)
#   BB_EXPORT_RETRY_DELAY  seconds between attempts          (default 1)

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

# 2. Refresh the durable export, then VERIFY that this bead's closure is in it.
#
#    .beads/issues.jsonl is a passive export, and the write behind it is
#    best-effort: a close has been observed reaching the database seconds
#    before it reaches the export. Measured 2026-08-10 on agent-plugins-92q.20
#    (d5a6de5) and agent-plugins-92q.23 (905aee4) — both `closed` in the
#    database and `open` in the ledger that shipped with their commits. This
#    check used to be a warning that fell through to the commit, which is how
#    those two closures were nearly lost.
#
#    So: export, re-export up to BB_EXPORT_ATTEMPTS times about
#    BB_EXPORT_RETRY_DELAY seconds apart, and decide on the bead's RECORDED
#    STATUS, not on the file hash. A close that genuinely produces no ledger
#    delta (the ledger already reads `closed`) must keep passing; a close whose
#    status never lands in the ledger must stop the ceremony.
BB_EXPORT_ATTEMPTS="${BB_EXPORT_ATTEMPTS:-3}"
BB_EXPORT_RETRY_DELAY="${BB_EXPORT_RETRY_DELAY:-1}"

# One export, handling both build flavours (stdout, or writes the file itself).
run_export() {
  if ! bd export > /tmp/bb-export.$$ 2>/dev/null; then
    # This build writes the file itself rather than to stdout.
    rm -f /tmp/bb-export.$$
    bd export >/dev/null 2>&1 || return 1
  elif [ -s /tmp/bb-export.$$ ]; then
    mv /tmp/bb-export.$$ .beads/issues.jsonl
  else
    rm -f /tmp/bb-export.$$
  fi
  return 0
}

# Status of an issue as the LEDGER records it. No jq dependency: the record is
# one JSON line, and quotes inside a description are escaped (\"status\":\"),
# so the unescaped "status":" only ever matches the real field.
ledger_status() {
  [ -f .beads/issues.jsonl ] || return 0
  grep -F "\"id\":\"$1\"" .beads/issues.jsonl 2>/dev/null \
    | grep -o '"status":"[a-z_]*"' | head -1 | cut -d'"' -f4
}

# Status of an issue as the DATABASE records it — the authority on whether the
# close itself took. jq when it is available; the fallback reads the first
# "status" of the pretty-printed record, which is the subject's own.
db_status() {
  if command -v jq >/dev/null 2>&1; then
    bd show "$1" --json 2>/dev/null | jq -r '.[0].status // empty'
  else
    bd show "$1" --json 2>/dev/null | grep -o '"status": *"[a-z_]*"' | head -1 | cut -d'"' -f4
  fi
}

before_snapshot=""; [ -f .beads/issues.jsonl ] && before_snapshot="$(cat .beads/issues.jsonl)"
before=""; [ -f .beads/issues.jsonl ] && before="$(git hash-object .beads/issues.jsonl)"

attempt=1
while : ; do
  run_export || fail "\`bd export\` failed — closure is not durable, refusing to commit"
  [ "$(ledger_status "$id")" = "closed" ] && break
  [ "$attempt" -ge "$BB_EXPORT_ATTEMPTS" ] && break
  echo "bb-close-bead: export attempt $attempt did not carry $id's closure — retrying in ${BB_EXPORT_RETRY_DELAY}s" >&2
  sleep "$BB_EXPORT_RETRY_DELAY"
  attempt=$((attempt + 1))
done

if [ "$(ledger_status "$id")" != "closed" ]; then
  # The hash is not the arbiter — ask the database what actually happened.
  observed="$(db_status "$id")"
  if [ "$observed" != "closed" ]; then
    fail "the close did not take: bd reports $id status '${observed:-unknown}' after \`bd update --status closed\`.
  Nothing staged, nothing committed. Fix the tracker write, then re-run this script."
  fi
  fail "$id is closed in the tracker database but its closure is NOT in .beads/issues.jsonl after $BB_EXPORT_ATTEMPTS export attempt(s).
  Refusing to commit — a closure the committed ledger does not carry is lost when the container is reclaimed.
  The bead IS ALREADY CLOSED: do NOT re-run \`bd update --status closed\`.
  Repair, once \`bd export | grep -F '\"id\":\"$id\"'\` shows status closed:
    bd export > .beads/issues.jsonl
    git add <the paths you passed> .beads/issues.jsonl
    git commit -m \"$id P$priority $title\""
fi

after=""; [ -f .beads/issues.jsonl ] && after="$(git hash-object .beads/issues.jsonl)"
if [ "$before" = "$after" ]; then
  echo "bb-close-bead: .beads/issues.jsonl unchanged, and the ledger already records $id closed — no delta to carry."
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
