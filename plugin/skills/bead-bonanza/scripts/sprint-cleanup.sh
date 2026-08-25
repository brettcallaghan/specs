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
    handle_worktree "$wt" "$br" "$beads"
  done <<WTLIST
$pairs
WTLIST
  [ "$seen" = 0 ] && say "no run worktrees registered"
  return 0
}

# ONE worktree's fate — decided and acted on. Separate from prune_worktrees so
# the self-test can drive the guard itself with `agent_drained` and `run`
# stubbed, needing neither a tracker nor a repository. The exit-status wiring is
# part of the case, not an afterthought: a guard that decides LEFT correctly and
# still lets the script exit 0 has torn nothing down but has also warned nobody.
# Callers must NOT capture its stdout — a command substitution runs it in a
# subshell, where note_left cannot reach $left.
handle_worktree() { # handle_worktree <path> <branch> <beads-json>
  _hw_who="$(printf '%s' "$3" | assignee_for_branch "$2")"
  _hw_st="$(printf '%s' "$3" | status_for_branch "$2")"
  if [ -z "$_hw_who" ]; then
    line "LEFT" "worktree, no bead owns its branch" "$1"; note_left
  elif [ "$_hw_st" = blocked ]; then
    # Checked BEFORE agent_drained: a blocked bead satisfies agent_drained too
    # (bd ready and bd list --status in_progress both exclude it), and the
    # worktree holds the partial work — the durable record that the bead is
    # unfinished — so it is left standing, not torn down.
    line "LEFT" "worktree, its bead is BLOCKED (partial work)" "$1"; note_left
  else
    agent_drained "$_hw_who"; _hw_ad=$?
    case "$_hw_ad" in
      0) run "remove worktree ($_hw_who drained)" git worktree remove --force "$1" ;;
      2) line "LEFT" "worktree, the tracker could not answer for $_hw_who" "$1"; note_left ;;
      *) line "LEFT" "worktree, $_hw_who still has a queue" "$1"; note_left ;;
    esac
  fi
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

# Same lookup, same branch-is-a-hash-suffix reasoning as assignee_for_branch,
# reading .status instead of .assignee. handle_worktree consults this BEFORE
# agent_drained: a blocked bead's worker has nothing ready and nothing
# in_progress, so agent_drained alone reads it as finished and its worktree —
# the durable record of unfinished work — would be torn down.
status_for_branch() { # status_for_branch <branch>  < beads-json  ->  status|""
  [ -n "$1" ] || return 0
  jq -r --arg b "-$1" \
    'first(.[] | select(.id | endswith($b)) | .status // "") // ""' 2>/dev/null
}

# Drained = nothing ready and nothing in progress. BOTH are required: `bd
# ready` excludes in_progress, so a bead being worked right now makes the ready
# front empty, and the ready query alone reports a live agent as finished.
# `bd ready --json` returns null rather than [] when empty, so the emptiness
# test has to accept both.
# Shell functions share globals, so every local here is prefixed `_ad_`/`_je_`.
# An unprefixed `_n` silently overwrote the self-test's own loop variable and
# made a passing assertion print the wrong name.
# THREE outcomes, not two, because a tracker that cannot ANSWER is not a
# tracker answering "nothing". Both statuses are read: an assignment captures
# the command substitution's own status, so this is not the discarded-status
# form the `2>/dev/null` above would otherwise leave. A failed query is the
# UNKNOWN verdict, and `handle_worktree` treats unknown as LEFT — fail closed,
# because the action on the other branch is `git worktree remove --force` and
# takes uncommitted work with it.
agent_drained() { # agent_drained <assignee>  ->  0 drained / 1 still holding / 2 could not tell
  _ad_ready="$(bd ready --assignee "$1" --json 2>/dev/null)"; _ad_rrc=$?
  [ "$_ad_rrc" -eq 0 ] || return 2
  _ad_prog="$(bd list --assignee "$1" --status in_progress --json 2>/dev/null)"; _ad_prc=$?
  [ "$_ad_prc" -eq 0 ] || return 2
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
  # Same lookup, for status — feeds the blocked guard below.
  _sl='[{"id":"acme-r7k.4","status":"blocked"},{"id":"acme-jh6","status":"in_progress"},{"id":"acme-zz9","status":null}]'
  for _c in "blocked:blocked:r7k.4" "in_progress:in_progress:jh6" "null status::zz9" "no such branch::nope" "detached worktree::"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _b="${_r#*:}"
    _g="$(printf '%s' "$_sl" | status_for_branch "$_b")"
    if [ "$_g" = "$_e" ]; then printf '  ok   branch status, %s\n' "$_n"
    else printf '  FAIL branch status, %s (expected [%s], got [%s])\n' "$_n" "$_e" "$_g"; fails=$((fails+1)); fi
  done
  # Emptiness, which decides whether an agent has drained. `bd ready --json`
  # answers null rather than [], so both spellings must read as empty.
  for _c in "null is empty:0:null" "[] is empty:0:[]" "unset is empty:0:" "a one-item list is not:1:[{}]"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _j="${_r#*:}"
    if json_empty "$_j"; then _g=0; else _g=1; fi
    if [ "$_g" = "$_e" ]; then printf '  ok   %s\n' "$_n"
    else printf '  FAIL %s (expected %s, got %s)\n' "$_n" "$_e" "$_g"; fails=$((fails+1)); fi
  done
  # A FAILED tracker query is not an empty answer. Driven through the REAL
  # `agent_drained`, with `bd` shadowed by a stub whose exit status is the
  # variable under test: reading a failure as "drained" is exactly what
  # force-removes a live agent's worktree, so the verdict has to be UNKNOWN.
  # `run` is stubbed alongside it so the end-to-end case cannot reach
  # `git worktree remove` even if the guard regresses.
  run() { _hw_ran="yes"; }
  _adq='[{"id":"acme-r7k.4","assignee":"dev-live"}]'
  for _c in "both queries fail:2:fail-all" \
            "the in_progress query alone fails:2:fail-prog" \
            "both answer empty:0:empty" \
            "the ready front is non-empty:1:busy"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _adq_mode="${_r#*:}"
    bd() {
      case "$_adq_mode" in
        fail-all)  return 1 ;;
        fail-prog) if [ "$1" = ready ]; then printf '[]'; return 0; fi; return 1 ;;
        empty)     printf '[]'; return 0 ;;
        *)         printf '%s' "$_adq"; return 0 ;;
      esac
    }
    agent_drained dev-live; _g=$?
    if [ "$_g" = "$_e" ]; then printf '  ok   drained verdict, %s\n' "$_n"
    else printf '  FAIL drained verdict, %s (expected %s, got %s)\n' "$_n" "$_e" "$_g"; fails=$((fails+1)); fi
  done
  # End to end. Component coverage above is not this: `agent_drained` can
  # answer UNKNOWN correctly while `handle_worktree` still removes the tree.
  # An unreadable tracker must leave it standing AND raise the flag, so the
  # script exits RC_LEFT instead of reporting a clean teardown.
  _adq_mode=fail-all; left=0; _hw_ran=no
  handle_worktree "/wt/r7k.4" "r7k.4" "$_adq" >/dev/null 2>&1
  if [ "$left" = 1 ] && [ "$_hw_ran" = no ]; then
    printf '  ok   %s\n' "guard, an unreadable tracker leaves the worktree standing"
  else
    printf '  FAIL %s (expected left=1 removal-attempted=no, got left=%s removal-attempted=%s)\n' \
      "guard, an unreadable tracker leaves the worktree standing" "$left" "$_hw_ran"; fails=$((fails+1))
  fi
  unset -f bd
  left=0
  # THE GUARD ITSELF, end to end, through the real decision function. Component
  # coverage above is not this: a lookup and an emptiness test can each be right
  # while the verdict they feed is wrong. A worktree is torn down ONLY when the
  # tracker says its owner has drained; every other answer leaves it standing
  # AND raises the flag that makes the script exit RC_LEFT.
  # `agent_drained` and `run` are stubbed, so nothing here reads a tracker or a
  # repository. Not captured in a `$( )`: that subshell would swallow $left.
  _hw_beads='[{"id":"acme-r7k.4","assignee":"dev-live"},{"id":"acme-blk7","assignee":"dev-done","status":"blocked"},{"id":"acme-jh6","assignee":"dev-done"},{"id":"acme-zz9","assignee":null}]'
  agent_drained() { [ "$1" = dev-done ]; }
  run() { _hw_ran="yes"; }
  # The blocked case MUST run before the drained case: acme-blk7's assignee is
  # dev-done, the same stub agent_drained reports drained, so this is the one
  # arrangement that actually exercises the ordering — a guard that checked
  # agent_drained first would remove this worktree too.
  for _c in "guard, a still-claimed bead keeps its worktree:1:r7k.4" \
            "guard, a blocked bead's worktree is left despite a drained agent:1:blk7" \
            "guard, a drained agent's worktree is removed:0:jh6" \
            "guard, a bead assigned to nobody keeps it:1:zz9" \
            "guard, no bead owns the branch, keep it:1:nope" \
            "guard, a detached worktree is kept:1:"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _b="${_r#*:}"
    left=0; _hw_ran=no
    handle_worktree "/wt/$_b" "$_b" "$_hw_beads" >/dev/null 2>&1
    # left=1 must coincide with "no removal attempted", and left=0 with one.
    if [ "$left" = "$_e" ] && { { [ "$_e" = 1 ] && [ "$_hw_ran" = no ]; } || { [ "$_e" = 0 ] && [ "$_hw_ran" = yes ]; }; }; then
      printf '  ok   %s\n' "$_n"
    else
      printf '  FAIL %s (expected left=%s, got left=%s removal-attempted=%s)\n' "$_n" "$_e" "$left" "$_hw_ran"; fails=$((fails+1))
    fi
  done
  left=0   # the stubs stay defined for the rest of the self-test; nothing below calls them.

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
