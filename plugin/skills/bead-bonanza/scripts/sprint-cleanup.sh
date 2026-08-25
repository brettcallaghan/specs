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
#   out branch is `bb/<dispatch-id>`, the dispatch id IS the agent, and the
#   agent has DRAINED when `bd ready --assignee <id>` is empty and it holds
#   nothing `in_progress`. Nothing has to be written at dispatch time for that
#   to work, so nothing can be forgotten or go stale.
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
  run_id="${epic##*-}"
  # Registrations for directories that no longer exist read as in-flight
  # forever. Pruning them is safe by definition — the directory is gone.
  run "prune stale registrations" git worktree prune

  main="$(git rev-parse --show-toplevel 2>/dev/null)"
  pairs="$(git worktree list --porcelain 2>/dev/null | parse_worktree_branches)"
  [ -n "$pairs" ] || { say "no worktrees registered"; return 0; }

  seen=0
  # A pipe would put this loop in a subshell, where note_left cannot reach
  # $left and a LEFT worktree would print while the script still exited 0.
  while IFS= read -r pair; do
    [ -n "$pair" ] || continue
    wt="${pair%%	*}"; br="${pair#*	}"
    [ "$wt" = "$main" ] && continue
    seen=$((seen+1))
    handle_worktree "$wt" "$br" "$run_id"
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
handle_worktree() { # handle_worktree <path> <branch> <run-id>
  _hw_who="$(owner_for_branch "$2")"
  if [ -z "$_hw_who" ]; then
    line "LEFT" "worktree, its branch names no dispatch" "$1"; note_left
  elif ! in_run "$_hw_who" "$3"; then
    # SCOPE. A dispatch id carries its run id by construction, so a worktree
    # belonging to a CONCURRENT sprint is recognised without reading the
    # tracker at all — and is left standing rather than force-removed with its
    # uncommitted work.
    line "LEFT" "worktree, $_hw_who belongs to another run" "$1"; note_left
  else
    holds_blocked "$_hw_who"; _hw_bl=$?
    case "$_hw_bl" in
      2) line "LEFT" "worktree, the tracker could not answer for $_hw_who" "$1"; note_left; return ;;
      0) # Checked BEFORE agent_drained: a blocked bead satisfies agent_drained
         # too (bd ready and bd list --status in_progress both exclude it), and
         # the worktree holds the partial work — the durable record that the
         # bead is unfinished — so it is left standing, not torn down.
         line "LEFT" "worktree, $_hw_who holds a BLOCKED bead (partial work)" "$1"; note_left; return ;;
    esac
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

# A dispatch's branch is `bb/<dispatch-id>` and the dispatch id IS the agent,
# so ownership is a string strip rather than a tracker lookup. A branch that
# does not carry the prefix belongs to something this run did not create — a
# detached worktree emits an empty branch and lands here too — and is reported
# LEFT rather than removed.
owner_for_branch() { # owner_for_branch <branch>  ->  dispatch-id|""
  case "$1" in
    bb/?*) printf '%s' "${1#bb/}" ;;
    *)     : ;;
  esac
}

# SCOPE, without reading the tracker. Every dispatch id derives from the run
# epic id — `dev-<run-id>-<n>` for a developer, `<role>-<run-id>` for a
# ceremony phase — so a CONCURRENT sprint's id cannot match, and a scope that
# cannot be established is short rather than wide. A worktree this misses reads
# as unowned and is LEFT standing, which is the safe direction; the wide answer
# force-removes another run's uncommitted work.
in_run() { # in_run <dispatch-id> <run-id>  ->  0 in this run / 1 not
  [ -n "$2" ] || return 1
  case "$1" in
    *-"$2"-*|*-"$2") return 0 ;;
  esac
  return 1
}

# Does this agent hold a bead it gave up on? Consulted BEFORE agent_drained: a
# blocked bead's worker has nothing ready and nothing in_progress, so
# agent_drained alone reads it as finished and its worktree — the durable
# record of unfinished work — would be torn down. THREE outcomes, for the same
# reason agent_drained has three: a query that fails is not a query answering
# "none".
holds_blocked() { # holds_blocked <assignee>  ->  0 holds one / 1 none / 2 could not tell
  _hb_out="$(bd list --assignee "$1" --status blocked --json 2>/dev/null)"; _hb_rc=$?
  [ "$_hb_rc" -eq 0 ] || return 2
  json_empty "$_hb_out" && return 1
  return 0
}

# Drained = nothing ready and nothing in progress. BOTH are required: `bd
# ready` excludes in_progress, so a bead being worked right now makes the ready
# front empty, and the ready query alone reports a live agent as finished.
# `bd ready --json` prints [] when the front is empty -- measured on bd 1.2.2
# in every configuration tried, and null in none of them. json_empty accepts
# null and the empty string as well: it is a shape guard rather than a
# restatement of what bd emits, and what separates "nothing there" from
# "could not tell" is the query's exit status, read below before json_empty
# is reached at all.
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
  # Branch -> owning agent, driven through the real function. A branch is
  # `bb/<dispatch-id>`; anything else names no dispatch this run created.
  for _c in "dev branch:dev-r7k-1:bb/dev-r7k-1" \
            "ceremony branch:retro-r7k:bb/retro-r7k" \
            "a bare branch name owns nothing::main" \
            "the prefix alone owns nothing::bb/" \
            "detached worktree::"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _b="${_r#*:}"
    _g="$(owner_for_branch "$_b")"
    if [ "$_g" = "$_e" ]; then printf '  ok   branch owner, %s\n' "$_n"
    else printf '  FAIL branch owner, %s (expected [%s], got [%s])\n' "$_n" "$_e" "$_g"; fails=$((fails+1)); fi
  done
  # RUN SCOPING, driven through the real function. This is what keeps a
  # CONCURRENT sprint's worktree standing, and it replaces a tracker walk — so
  # every way a foreign id could sneak through is asserted, not just the
  # matching case.
  for _c in "a dev id of this run:0:dev-r7k-1:r7k" \
            "a ceremony id of this run:0:retro-r7k:r7k" \
            "a two-digit batch:0:dev-r7k-12:r7k" \
            "another run's dev id:1:dev-abc-1:r7k" \
            "another run's ceremony id:1:retro-abc:r7k" \
            "a run id that is only a SUBSTRING:1:dev-r7king-1:r7k" \
            "a run id that only PREFIXES the segment:1:dev-xr7k-1:r7k" \
            "an empty run id matches nothing:1:dev-r7k-1:"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _r="${_r#*:}"
    _i="${_r%%:*}"; _ru="${_r#*:}"
    if in_run "$_i" "$_ru"; then _g=0; else _g=1; fi
    if [ "$_g" = "$_e" ]; then printf '  ok   scope, %s\n' "$_n"
    else printf '  FAIL scope, %s (expected %s, got %s)\n' "$_n" "$_e" "$_g"; fails=$((fails+1)); fi
  done
  # Emptiness, which decides whether an agent has drained. The measured shape
  # is [], but json_empty must read every non-list payload as empty too, so
  # each spelling is exercised rather than only the one bd emits.
  for _c in "null is empty:0:null" "[] is empty:0:[]" "unset is empty:0:" "a one-item list is not:1:[{}]"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _j="${_r#*:}"
    if json_empty "$_j"; then _g=0; else _g=1; fi
    if [ "$_g" = "$_e" ]; then printf '  ok   %s\n' "$_n"
    else printf '  FAIL %s (expected %s, got %s)\n' "$_n" "$_e" "$_g"; fails=$((fails+1)); fi
  done
  # A FAILED tracker query is not an empty answer. Driven through the REAL
  # `agent_drained` and `holds_blocked`, with `bd` shadowed by a stub whose
  # exit status is the variable under test: reading a failure as "drained" or
  # as "holds nothing blocked" is exactly what force-removes a live agent's
  # worktree, so the verdict has to be UNKNOWN.
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
        *)         printf '[{"id":"acme-r7k.4","assignee":"dev-live"}]'; return 0 ;;
      esac
    }
    agent_drained dev-live; _g=$?
    if [ "$_g" = "$_e" ]; then printf '  ok   drained verdict, %s\n' "$_n"
    else printf '  FAIL drained verdict, %s (expected %s, got %s)\n' "$_n" "$_e" "$_g"; fails=$((fails+1)); fi
  done
  for _c in "a failed query is not 'none blocked':2:fail-all" \
            "an empty answer is none blocked:1:empty" \
            "a non-empty answer holds one:0:busy"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _adq_mode="${_r#*:}"
    holds_blocked dev-live; _g=$?
    if [ "$_g" = "$_e" ]; then printf '  ok   blocked verdict, %s\n' "$_n"
    else printf '  FAIL blocked verdict, %s (expected %s, got %s)\n' "$_n" "$_e" "$_g"; fails=$((fails+1)); fi
  done
  unset -f bd
  # THE GUARD ITSELF, end to end, through the real decision function. Component
  # coverage above is not this: a lookup and an emptiness test can each be right
  # while the verdict they feed is wrong. A worktree is torn down ONLY when the
  # tracker says its owner has drained AND the owner belongs to this run; every
  # other answer leaves it standing AND raises the flag that makes the script
  # exit RC_LEFT.
  # `agent_drained`, `holds_blocked` and `run` are stubbed, so nothing here
  # reads a tracker or a repository. Not captured in a `$( )`: that subshell
  # would swallow $left.
  run() { _hw_ran="yes"; }
  agent_drained() { [ "$1" = "dev-r7k-2" ] || [ "$1" = "dev-r7k-3" ] || [ "$1" = "dev-abc-9" ]; }
  holds_blocked() { [ "$1" = "dev-r7k-3" ]; }
  # The blocked case and the foreign case both name agents `agent_drained`
  # reports DRAINED, which is the only arrangement that actually exercises the
  # ordering: a guard consulting agent_drained first would remove both.
  for _c in "guard, a still-claimed agent keeps its worktree:1:no:bb/dev-r7k-1" \
            "guard, a drained agent's worktree is removed:0:yes:bb/dev-r7k-2" \
            "guard, a BLOCKED bead's worktree is left despite a drained agent:1:no:bb/dev-r7k-3" \
            "guard, a CONCURRENT run's worktree is left despite a drained agent:1:no:bb/dev-abc-9" \
            "guard, a branch naming no dispatch keeps it:1:no:main" \
            "guard, a detached worktree is kept:1:no:"; do
    _n="${_c%%:*}"; _r="${_c#*:}"; _e="${_r%%:*}"; _r="${_r#*:}"
    _x="${_r%%:*}"; _b="${_r#*:}"
    left=0; _hw_ran=no
    handle_worktree "/wt/$_b" "$_b" "r7k" >/dev/null 2>&1
    if [ "$left" = "$_e" ] && [ "$_hw_ran" = "$_x" ]; then printf '  ok   %s\n' "$_n"
    else printf '  FAIL %s (expected left=%s removal-attempted=%s, got left=%s removal-attempted=%s)\n' "$_n" "$_e" "$_x" "$left" "$_hw_ran"; fails=$((fails+1)); fi
  done
  # THE NEGATIVE CONTROL that names the bug: with scoping disabled, the same
  # foreign worktree IS removed. Without this the scoping assertion above could
  # pass for the wrong reason.
  in_run() { return 0; }
  left=0; _hw_ran=no
  handle_worktree "/wt/foreign" "bb/dev-abc-9" "r7k" >/dev/null 2>&1
  if [ "$left" = 0 ] && [ "$_hw_ran" = yes ]; then
    printf '  ok   %s\n' "UNSCOPED (the bug), the same foreign worktree IS removed"
  else
    printf '  FAIL %s (expected left=0 removal-attempted=yes, got left=%s removal-attempted=%s)\n' \
      "UNSCOPED (the bug), the same foreign worktree IS removed" "$left" "$_hw_ran"; fails=$((fails+1))
  fi
  unset -f in_run agent_drained holds_blocked
  left=0

  # Pairing a worktree with its branch, spaces and detachment included.
  _g="$(printf 'worktree /a/b c\nHEAD deadbeef\nbranch refs/heads/qozp.8\n\nworktree /d/e\nHEAD cafe\n' | parse_worktree_branches)"
  _e="$(printf '/a/b c\tqozp.8\n/d/e\t')"
  if [ "$_g" = "$_e" ]; then printf '  ok   %s\n' "worktree pairs keep spaces; a detached tree has no branch"
  else printf '  FAIL %s (got [%s])\n' "worktree pairs keep spaces; a detached tree has no branch" "$_g"; fails=$((fails+1)); fi
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
push_tracker

if [ "$left" = 1 ]; then
  echo
  echo "  Something was left standing on purpose — see the LEFT lines above."
  echo "  Resolve each with a human before treating this sprint as torn down."
  exit "$RC_LEFT"
fi
exit "$RC_OK"
