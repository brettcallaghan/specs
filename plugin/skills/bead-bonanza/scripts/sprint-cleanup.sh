#!/usr/bin/env bash
#
# sprint-cleanup.sh — post-flight. Tear down what a finished run left behind,
# and report what it deliberately did not.
#
# USAGE
#   sprint-cleanup.sh <epic-id> [--dry-run]
#   sprint-cleanup.sh --self-test
#
# EXIT CODES
#   0  everything this run owned is torn down, or was already
#   1  something was LEFT STANDING deliberately — named on stdout
#   3  invocation error
#
# WHO RUNS IT
#   The ORCHESTRATOR, once, after the report and the retro. A recovering
#   session may run it too: it is idempotent, so a second run on a clean sprint
#   reports "nothing to do" rather than failing.
#
# NEVER DESTRUCTIVE ON AMBIGUITY
#   A worktree whose agent still has a queue is a live worker or a corpse, and
#   this script cannot tell which. It leaves it alone and says so.
#
#   It asks the tracker rather than a bookkeeping key: the worktree's checked-
#   out branch is a bead's hash suffix, that bead's `assignee` is the agent,
#   and the agent has DRAINED when `bd ready --assignee <id>` is empty and it
#   holds nothing `in_progress`. Nothing has to be written at dispatch time for
#   that to work, so nothing can be forgotten or go stale.
#
# Every removal prints WHAT was removed and the command that did it, so a
# reader can tell a teardown from a claim of one.

set -uo pipefail

RC_OK=0; RC_LEFT=1; RC_USAGE=3

die()  { echo "sprint-cleanup: $*" >&2; exit "$RC_USAGE"; }
say()  { printf '  %s\n' "$*"; }
line() { printf '  %-8s %-40s %s\n' "$1" "$2" "$3"; }

DRY=no
left=0
note_left() { left=1; }

run() { # run <description> <command...>
  d="$1"; shift
  if [ "$DRY" = yes ]; then line "WOULD" "$d" "$*"; return 0; fi
  if "$@" >/dev/null 2>&1; then line "DONE" "$d" "$*"; else line "FAILED" "$d" "$*"; note_left; fi
}

# --------------------------------------------------------------- worktrees
prune_worktrees() { # prune_worktrees <epic>
  epic="$1"
  # Registrations for directories that no longer exist read as in-flight
  # forever. Pruning them is safe by definition — the directory is gone.
  run "prune stale registrations" git worktree prune

  # bd's JSON is the authority on who still holds what. A worktree whose agent
  # has not drained is left standing: it is either a live agent or a corpse,
  # and only a human can tell those apart.
  wts="$(git worktree list --porcelain 2>/dev/null | parse_worktrees)"
  main="$(git rev-parse --show-toplevel 2>/dev/null)"
  [ -n "$wts" ] || { say "no worktrees registered"; return 0; }

  # Not scoped by `bd list --parent`: it returns DIRECT children only, so on a
  # nested run epic every leaf bead would go unmatched and its worktree be left
  # standing. `--all` because a finished bead is closed and `bd list` defaults
  # to open.
  beads="$(bd list --all --json 2>/dev/null)" || beads='[]'
  pairs="$(git worktree list --porcelain 2>/dev/null | parse_worktree_branches)"
  seen=0
  # A pipe would put this loop in a subshell, where note_left cannot reach
  # $left and a LEFT worktree would print while the script still exited 0.
  while IFS= read -r pair; do
    [ -n "$pair" ] || continue
    wt="${pair%%	*}"; br="${pair#*	}"
    [ "$wt" = "$main" ] && continue
    seen=$((seen+1))
    who="$(printf '%s' "$beads" | assignee_for_branch "$br")"
    if [ -z "$who" ]; then
      line "LEFT" "worktree, no bead owns its branch" "$wt"; note_left
    elif agent_drained "$who"; then
      run "remove worktree ($who drained)" git worktree remove --force "$wt"
    else
      line "LEFT" "worktree, $who still has a queue" "$wt"; note_left
    fi
  done <<WTLIST
$pairs
WTLIST
  [ "$seen" = 0 ] && say "no run worktrees registered"
  return 0
}

# ----------------------------------------------------------------- parsers
# Both read stdin and are pure text transforms, so the self-test can drive
# them with synthetic input and needs neither a repository nor a tracker.

# sed, not `awk '{print $2}'`: a worktree path containing a space is truncated
# at the space by field splitting, and the truncated path then matches no bead
# and is reported LEFT forever.
parse_worktrees() { sed -n 's/^worktree //p'; }

# path<TAB>branch, one line per registration, from `git worktree list
# --porcelain`. A detached worktree has no `branch` line, so it emits an empty
# branch and is reported LEFT rather than silently removed. Substring
# assignment, not field splitting, for the same space-in-path reason.
parse_worktree_branches() {
  awk '
    /^worktree /  { if (p != "") { print p "\t" b } ; p = substr($0, 10); b = "" ; next }
    /^branch /    { b = substr($0, 8); sub(/^refs\/heads\//, "", b); next }
    END           { if (p != "") { print p "\t" b } }
  '
}

# `bd swarm list --json` is the only lookup that works. Molecules carry NO
# metadata (so an `.metadata.epic` predicate matches nothing) and are absent
# from `bd list --type molecule` entirely; the swarm listing carries the epic
# link and the real status as fields. Its TEXT output files closed molecules
# under "Active Swarms" — the JSON's `status` does not lie.
# A worktree's branch is a bead's hash SUFFIX, not its id — the prefix is the
# project's and never appears in a branch name — so the match is on the id
# ending in `-<branch>`. An empty branch (a detached worktree) matches nothing
# by construction, which is the wanted answer.
assignee_for_branch() { # assignee_for_branch <branch>  < beads-json  ->  assignee|""
  [ -n "$1" ] || return 0
  jq -r --arg b "-$1" \
    'first(.[] | select(.id | endswith($b)) | .assignee // "") // ""' 2>/dev/null
}

# Drained = nothing ready and nothing in progress. BOTH are required: `bd
# ready` excludes in_progress, so a bead being worked right now makes the ready
# front empty, and the ready query alone reports a live agent as finished.
# `bd ready --json` returns null rather than [] when empty, so the emptiness
# test has to accept both.
# Shell functions share globals, so every local here is prefixed `_ad_`/`_je_`.
# An unprefixed `_n` silently overwrote the self-test's own loop variable and
# made a passing assertion print the wrong name.
agent_drained() { # agent_drained <assignee>  ->  0 drained / 1 still holding
  _ad_ready="$(bd ready --assignee "$1" --json 2>/dev/null)"
  _ad_prog="$(bd list --assignee "$1" --status in_progress --json 2>/dev/null)"
  json_empty "$_ad_ready" && json_empty "$_ad_prog"
}

json_empty() { # json_empty <json>  ->  0 when null, [] or unreadable-as-a-list
  case "$1" in ''|null|'[]') return 0 ;; esac
  _je_n="$(printf '%s' "$1" | jq -r 'if type == "array" then length else 0 end' 2>/dev/null)"
  [ "${_je_n:-0}" = 0 ]
}

select_molecule() { # select_molecule <epic>  < swarm-list-json
  jq -r --arg e "$1" \
    '[.swarms[]? | select(.epic_id == $e and .status != "closed")] | .[0].id // empty' \
    2>/dev/null
}

# -------------------------------------------------------------- merge slot
release_slot() { # release_slot <epic>
  out="$(bd merge-slot check 2>&1)"; rc=$?
  if [ $rc -ne 0 ]; then say "no merge slot to release (bd merge-slot check exit $rc)"; return 0; fi
  case "$out" in
    *"no holder"*|*"not held"*|*"available"*) say "merge slot already free"; return 0 ;;
  esac
  run "release merge slot" bd merge-slot release
}

# ------------------------------------------------------------------- swarm
close_molecule() { # close_molecule <epic>
  epic="$1"
  mols="$(bd swarm list --json 2>/dev/null)" || mols='{}'
  mol="$(printf '%s' "$mols" | select_molecule "$epic")"
  [ -n "$mol" ] || { say "no open swarm molecule for $epic"; return 0; }
  run "close swarm molecule" bd close "$mol" --reason "sprint complete"
}

# ------------------------------------------------------------------ health
set_health() { # set_health <epic> <value>
  run "set terminal health=$2" bd set-state "$1" "health=$2"
}

# ------------------------------------------------------------------- push
push_tracker() {
  run "push tracker" bd dolt push
}

# --------------------------------------------------------------- self-test
self_test() {
  fails=0
  t() { # t <name> <expected-rc> <command...>
    n="$1"; e="$2"; shift 2
    "$@" >/dev/null 2>&1; r=$?
    if [ "$r" = "$e" ]; then printf '  ok   %s\n' "$n"
    else printf '  FAIL %s (expected %s, got %s)\n' "$n" "$e" "$r"; fails=$((fails+1)); fi
  }
  # Invocation surface only — these fail before any tracker or git read, so the
  # self-test needs neither a database nor a repository.
  t "no argument is a usage error"        "$RC_USAGE" "$0"
  t "unknown flag is a usage error"       "$RC_USAGE" "$0" some-epic --bogus
  t "--self-test takes no epic"           "$RC_USAGE" "$0" --self-test extra
  # The two parsers, driven with synthetic input. Each replaces a form that
  # was measured wrong: `awk '{print $2}'` truncated a path at its first
  # space, and a molecule lookup keyed on `.metadata.epic` matched nothing
  # because molecules carry no metadata at all.
  # Branch -> owning agent, driven through the real function.
  _bl='[{"id":"acme-r7k.4","assignee":"dev-r7k-1"},{"id":"acme-jh6","assignee":"dev-r7k-2"},{"id":"acme-zz9","assignee":null}]'
  for _c in "suffix with a dot:dev-r7k-1:r7k.4" "bare suffix:dev-r7k-2:jh6" "unassigned bead::zz9" "no such branch::nope" "detached worktree::"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _b="${_r#*:}"
    _g="$(printf '%s' "$_bl" | assignee_for_branch "$_b")"
    if [ "$_g" = "$_e" ]; then printf '  ok   branch owner, %s\n' "$_n"
    else printf '  FAIL branch owner, %s (expected [%s], got [%s])\n' "$_n" "$_e" "$_g"; fails=$((fails+1)); fi
  done
  # Emptiness, which decides whether an agent has drained. `bd ready --json`
  # answers null rather than [], so both spellings must read as empty.
  for _c in "null is empty:0:null" "[] is empty:0:[]" "unset is empty:0:" "a one-item list is not:1:[{}]"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _j="${_r#*:}"
    if json_empty "$_j"; then _g=0; else _g=1; fi
    if [ "$_g" = "$_e" ]; then printf '  ok   %s\n' "$_n"
    else printf '  FAIL %s (expected %s, got %s)\n' "$_n" "$_e" "$_g"; fails=$((fails+1)); fi
  done
  # Pairing a worktree with its branch, spaces and detachment included.
  _g="$(printf 'worktree /a/b c\nHEAD deadbeef\nbranch refs/heads/qozp.8\n\nworktree /d/e\nHEAD cafe\n' | parse_worktree_branches)"
  _e="$(printf '/a/b c\tqozp.8\n/d/e\t')"
  if [ "$_g" = "$_e" ]; then printf '  ok   %s\n' "worktree pairs keep spaces; a detached tree has no branch"
  else printf '  FAIL %s (got [%s])\n' "worktree pairs keep spaces; a detached tree has no branch" "$_g"; fails=$((fails+1)); fi
  _sw() { printf '%s' '{"swarms":[{"id":"m-closed","epic_id":"E","status":"closed"},{"id":"m-open","epic_id":"E","status":"active"},{"id":"m-other","epic_id":"F","status":"active"}]}'; }
  r="$(_sw | select_molecule E)"
  if [ "$r" = "m-open" ]; then printf '  ok   %s\n' "molecule lookup skips closed and other epics"
  else printf '  FAIL %s (expected m-open, got %s)\n' "molecule lookup skips closed and other epics" "$r"; fails=$((fails+1)); fi
  r="$(printf 'worktree /a/b c\nHEAD deadbeef\nbranch refs/heads/x\n\nworktree /d/e\n' | parse_worktrees)"
  if [ "$r" = "/a/b c
/d/e" ]; then printf '  ok   %s\n' "a worktree path keeps its spaces"
  else printf '  FAIL %s (got [%s])\n' "a worktree path keeps its spaces" "$r"; fails=$((fails+1)); fi
  # NOT asserted: a dry run's exit status. It reports 1 whenever ANY worktree
  # is registered and unaccounted for, and during a sprint worktrees exist by
  # construction — so asserting 0 here made this self-test, and the declared
  # check that runs it, fail on every bead of every run.
  [ "$fails" = 0 ] || return 1
  return 0
}

# -------------------------------------------------------------------- main
[ $# -ge 1 ] || die "usage: sprint-cleanup.sh <epic-id> [--dry-run] | --self-test"

if [ "$1" = "--self-test" ]; then
  [ $# -eq 1 ] || die "--self-test takes no other arguments"
  echo "sprint-cleanup --self-test"; self_test; exit $?
fi

EPIC="$1"; shift
while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run) DRY=yes ;;
    *) die "unknown argument: $1" ;;
  esac
  shift
done

echo "sprint-cleanup: $EPIC${DRY:+ }$([ "$DRY" = yes ] && echo '(dry run)')"
prune_worktrees "$EPIC"
release_slot    "$EPIC"
close_molecule  "$EPIC"
set_health      "$EPIC" complete
push_tracker

if [ "$left" = 1 ]; then
  echo
  echo "  Something was left standing on purpose — see the LEFT lines above."
  echo "  Resolve each with a human before treating this sprint as torn down."
  exit "$RC_LEFT"
fi
exit "$RC_OK"
