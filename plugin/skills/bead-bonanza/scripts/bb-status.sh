#!/usr/bin/env bash
#
# bb-status.sh — verify a run's acceptance signals from state, not testimony.
#
# USAGE
#   bb-status.sh bead <bead-id> [--worktree <path>] [--pre-merge]
#   bb-status.sh branch [--branch <name>] [--remote <name>]
#   bb-status.sh harvest <bead-id> [<bead-id>...]
#   bb-status.sh --self-test
#
# EXIT CODES — the caller acts on these, so they distinguish three outcomes:
#   0  all clear
#   1  REPAIRABLE failure. Signals 1, 2 and 5: close the bead, or add the
#      missing `retro:` comment from the returned summary. State the repair in
#      the report.
#   2  STOP. Signals 3 and 4: the merge or the acceptance never
#      happened. Not the orchestrator's to repair inline — leave the worktree
#      standing, since it is the in-flight signal, and surface it.
#
# --pre-merge DEFERS SIGNALS 3 AND 4. The acceptance check runs on what a
# subagent returned BEFORE the orchestrator merges the branch and removes the
# worktree, so both signals would fail by construction: the branch cannot be an
# ancestor of HEAD yet, and the worktree is supposed to still be there. That
# run passes --pre-merge; the orchestrator re-runs without it after merging,
# which is what actually verifies 3 and 4.
#   3  invocation error.
#
# RUN-LEVEL STATUS IS NOT HERE. `swarm status --run <epic-id>` measures where
# a run is and whether it is done; `bb-status.sh run` is retired and exits 3
# naming it.
#
# WHO RUNS IT
#   The ORCHESTRATOR, after each bead's acceptance check and at every re-query
#   of the ready set; a recovering session, as its first act; a woken
#   durable trigger.
#
#   **A subagent never runs it.** Inside its own worktree the signals read
#   differently by construction — signal 4 asks whether the worktree is gone,
#   and the subagent is standing in it.
#
# WHAT IT DOES NOT DO
#   It does not check the Phase 4 report's prose. Whether a report NAMES a
#   failing signal is a property of text this script cannot read; what it can
#   do is make the underlying facts cheap to re-derive, which is `branch` mode.
#   It never repairs anything: every mode is read-only.
#
# Every line prints the verified state AND the command that established it, so
# a reader can re-run any single check without reconstructing it.

set -uo pipefail

RC_OK=0; RC_REPAIR=1; RC_STOP=2; RC_USAGE=3

die() { echo "bb-status: $*" >&2; exit "$RC_USAGE"; }
say() { printf '  %s\n' "$*"; }
# state, verdict, and the command that established it — one line, three parts.
line() { printf '  %-4s %-34s %s\n' "$1" "$2" "$3"; }

worst=$RC_OK
note() { # note <rc>
  case "$1" in
    "$RC_STOP")   worst=$RC_STOP ;;
    "$RC_REPAIR") [ "$worst" = "$RC_STOP" ] || worst=$RC_REPAIR ;;
  esac
}

# ------------------------------------------------------------------ bead mode
mode_bead() {
  id=""; wt=""; premerge=no
  [ $# -ge 1 ] || die "usage: bb-status.sh bead <bead-id> [--worktree <path>] [--pre-merge]"
  id="$1"; shift
  while [ $# -gt 0 ]; do
    case "$1" in
      --worktree)  [ $# -ge 2 ] || die "--worktree needs a path"; wt="$2"; shift 2 ;;
      --pre-merge) premerge=yes; shift ;;
      *)           die "unknown argument: $1" ;;
    esac
  done

  echo "bead $id"

  # 1. closed in the tracker.
  #
  # Read the status field, never `bd show | grep -q`: under `pipefail`, grep -q
  # exits at the first match and closes the pipe, `bd show` dies on SIGPIPE,
  # and the pipeline reports failure for a bead that is in fact closed. Same
  # class of trap as reading `$?` through a pager.
  c="bd show $id --json | jq -r .[0].status"
  _st="$(bd show "$id" --json 2>/dev/null | jq -r '(if type=="array" then .[0] else . end).status // empty' 2>/dev/null)"
  if [ "$_st" = closed ]; then
    line OK "1 tracker: closed" "$c"
  else
    line FAIL "1 tracker: ${_st:-unreadable}" "$c"; note $RC_REPAIR
  fi

  # 2. a commit whose subject carries the id
  c="git log --oneline --all --grep=$id -1"
  sha="$(git log --oneline --all --grep="$id" -1 --format=%h 2>/dev/null)"
  if [ -n "$sha" ]; then
    line OK "2 commit: $sha" "$c"
  else
    line FAIL "2 commit: none found" "$c"; note $RC_REPAIR
  fi

  # 3. merged into the current branch. Only meaningful once 2 found a sha, and
  # only AFTER the merge has happened. Acceptance verifies what a subagent returned
  # before the orchestrator merges, so it passes --pre-merge and signals 3 and 4
  # are deferred to the orchestrator's own check after the merge. Without that
  # split every bead fails signal 3 by construction — the branch is not an
  # ancestor of HEAD yet, and cannot be.
  if [ "$premerge" = yes ]; then
    line SKIP "3 merged: deferred" "(--pre-merge; the merge has not happened yet)"
  elif [ -n "$sha" ]; then
    c="git merge-base --is-ancestor $sha HEAD"
    if git merge-base --is-ancestor "$sha" HEAD 2>/dev/null; then
      line OK "3 merged into HEAD" "$c"
    else
      line STOP "3 NOT merged into HEAD" "$c"; note $RC_STOP
    fi
  else
    line SKIP "3 merged: no sha to test" "(signal 2 found nothing)"
  fi

  # 4. the worktree is gone. The load-bearing signal — state, not testimony.
  git worktree prune 2>/dev/null
  c="git worktree prune && git worktree list"
  if [ -n "$wt" ]; then
    if git worktree list 2>/dev/null | grep -qF -- "$wt"; then
      if [ "$premerge" = yes ]; then
        line SKIP "4 worktree: deferred" "(--pre-merge; removal follows the merge)"
      else
        line STOP "4 worktree SURVIVES: $wt" "$c"; note $RC_STOP
      fi
    else
      line OK "4 worktree: gone" "$c"
    fi
  else
    line SKIP "4 worktree: path not given" "$c"
  fi

  # 5. findings recorded
  mode_harvest_one "$id"
}

# --------------------------------------------------------------- harvest mode
#
# `bd show <id> --json` reports comment_count 0 for a bead that HAS comments,
# so a harvest keyed on it fails every bead in the run at once. The export
# pipeline is the one that reads correctly.
harvest_count() {
  # No `head` on the end: it would close the pipe early and, under pipefail,
  # turn a correct answer into a failure.
  bd export 2>/dev/null \
    | jq -r --arg id "$1" 'select(.id==$id) | [.comments[]?|select(.text|startswith("retro:"))] | length' 2>/dev/null \
    | tr -d "[:space:]"
}

mode_harvest_one() {
  _id="$1"
  c="bd export | jq '…select(.text|startswith(\"retro:\"))…'"
  n="$(harvest_count "$_id")"
  [ -n "$n" ] || n=0
  if [ "$n" -gt 0 ] 2>/dev/null; then
    line OK "5 findings: $n retro: comment(s)" "$c"
  else
    # Zero is not automatically wrong — a clean bead may say so explicitly —
    # but it is never silently acceptable, so it is surfaced as repairable.
    line FAIL "5 findings: none recorded" "$c"; note $RC_REPAIR
  fi
}

mode_harvest() {
  [ $# -ge 1 ] || die "usage: bb-status.sh harvest <bead-id> [<bead-id>...]"
  command -v jq >/dev/null 2>&1 || die "harvest needs jq"
  total=0
  for _b in "$@"; do
    echo "bead $_b"
    mode_harvest_one "$_b"
    n="$(harvest_count "$_b")"; [ -n "$n" ] || n=0
    total=$((total + n))
  done
  echo
  # The assertion that matters at run level: a whole run harvesting zero
  # findings is the signature of a broken pipeline, not of a flawless run.
  if [ "$total" -eq 0 ]; then
    line FAIL "harvest total: 0 across $# bead(s)" "a zero-total harvest is a broken pipeline, not a clean run"
    note $RC_REPAIR
  else
    line OK "harvest total: $total" "summed over $# bead(s)"
  fi
}

# ---------------------------------------------------------------- branch mode
mode_branch() {
  branch=""; remote=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --branch) [ $# -ge 2 ] || die "--branch needs a name"; branch="$2"; shift 2 ;;
      --remote) [ $# -ge 2 ] || die "--remote needs a name"; remote="$2"; shift 2 ;;
      *) die "unknown argument: $1" ;;
    esac
  done
  [ -n "$branch" ] || branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null)"
  [ -n "$remote" ] || remote="$(git config --get "branch.$branch.remote" 2>/dev/null || echo origin)"

  echo "branch: $branch, remote $remote"

  c="git status --porcelain"
  if [ -z "$(git status --porcelain 2>/dev/null)" ]; then
    line OK "tree: clean" "$c"
  else
    line FAIL "tree: DIRTY" "$c"; note $RC_REPAIR
  fi

  # The landing check. An orchestrator that dies between committing and pushing
  # leaves every other signal reading "finished": no worktrees, clean tree,
  # beads closed. Only this comparison separates a landed branch from an unlanded
  # one.
  c="git rev-list --count $remote/$branch..$branch"
  if git rev-parse --verify -q "$remote/$branch" >/dev/null 2>&1; then
    ahead="$(git rev-list --count "$remote/$branch..$branch" 2>/dev/null)"
    [ -n "$ahead" ] || ahead=0
    if [ "$ahead" -eq 0 ]; then
      line OK "landed: 0 unpushed commits" "$c"
    else
      line FAIL "landed: $ahead commit(s) UNPUSHED" "$c"; note $RC_REPAIR
    fi
  else
    line FAIL "landed: no $remote/$branch to compare" "$c"; note $RC_REPAIR
  fi

  git worktree prune 2>/dev/null
  c="git worktree prune && git worktree list"
  n="$(git worktree list 2>/dev/null | tail -n +2 | grep -c . || true)"
  [ -n "$n" ] || n=0
  if [ "$n" -eq 0 ]; then
    line OK "in flight: 0 worktrees" "$c"
  else
    # Not a failure by itself: a surviving worktree is the in-flight signal.
    line INFO "in flight: $n worktree(s)" "$c"
    git worktree list 2>/dev/null | tail -n +2 | sed 's/^/      /'
  fi
}

# ------------------------------------------------------------------ self-test
mode_self_test() {
  fails=0
  t() { # t <name> <expected> <actual>
    if [ "$2" = "$3" ]; then printf '  ok   %s\n' "$1"
    else printf '  FAIL %s: expected %s, got %s\n' "$1" "$2" "$3"; fails=$((fails+1)); fi
  }
  worst=$RC_OK; note $RC_REPAIR; t "repair raises worst"      "1" "$worst"
  worst=$RC_OK; note $RC_STOP;   t "stop raises worst"        "2" "$worst"
  worst=$RC_STOP; note $RC_REPAIR; t "repair cannot lower stop" "2" "$worst"
  worst=$RC_REPAIR; note $RC_STOP; t "stop overrides repair"  "2" "$worst"

  # Invocation surface. These fail before any tracker read, so the self-test
  # needs no database — the modes' end-to-end behaviour against real beads is
  # not covered here and is exercised by using them.
  _self="$(cd "$(dirname "$0")" && pwd)/$(basename "$0")"
  "$_self" nosuchmode >/dev/null 2>&1; t "unknown mode exits 3"      "3" "$?"
  "$_self" >/dev/null 2>&1;            t "no mode exits 3"           "3" "$?"
  "$_self" run x >/dev/null 2>&1;      t "retired run mode exits 3"  "3" "$?"
  t "and it names its replacement" "swarm status" \
    "$("$_self" run x 2>&1 | sed -n 's/.*\(swarm status\).*/\1/p')"

  echo
  [ "$fails" -eq 0 ] && { echo "RESULT: PASS"; return 0; }
  echo "RESULT: FAIL"; return 1
}

[ $# -ge 1 ] || die "usage: bb-status.sh bead|branch|harvest|--self-test"
sub="$1"; shift
case "$sub" in
  bead)        mode_bead "$@" ;;
  branch)      mode_branch "$@" ;;
  run)         die "the \`run\` mode is retired; use: swarm status --run <epic-id>" ;;
  harvest)     mode_harvest "$@" ;;
  --self-test) mode_self_test; exit $? ;;
  *)           die "unknown mode: $sub" ;;
esac
exit "$worst"
