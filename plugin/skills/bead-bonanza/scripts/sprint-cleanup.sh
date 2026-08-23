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
#   A worktree whose bead still reads dispatch_state=running is a live worker
#   or a dead one, and this script cannot tell which. It leaves it alone and
#   says so. The four-state table in `sprint-batch` is how a human reads it;
#   this script acts only on the two rows that are unambiguous.
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

  # bd's JSON is the authority on who still holds what. A worktree whose bead
  # reads `running` is left standing: it is either a live agent or a corpse,
  # and only a human can tell those apart.
  wts="$(git worktree list --porcelain 2>/dev/null | awk '/^worktree /{print $2}')"
  main="$(git rev-parse --show-toplevel 2>/dev/null)"
  [ -n "$wts" ] || { say "no worktrees registered"; return 0; }

  beads="$(bd list --parent "$epic" --json 2>/dev/null)" || beads='[]'
  seen=0
  # A pipe would put this loop in a subshell, where note_left cannot reach
  # $left and a LEFT worktree would print while the script still exited 0.
  while IFS= read -r wt; do
    [ -n "$wt" ] || continue
    [ "$wt" = "$main" ] && continue
    seen=$((seen+1))
    state="$(printf '%s' "$beads" | jq -r --arg w "$wt" \
      '[.[] | select(.metadata.worktree == $w)] | .[0].metadata.dispatch_state // "unknown"' 2>/dev/null)"
    case "$state" in
      returned) run "remove worktree" git worktree remove --force "$wt" ;;
      running)  line "LEFT" "worktree, dispatch_state=running" "$wt"; note_left ;;
      *)        line "LEFT" "worktree, no dispatch record" "$wt"; note_left ;;
    esac
  done <<WTLIST
$wts
WTLIST
  [ "$seen" = 0 ] && say "no run worktrees registered"
  return 0
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
  # `bd swarm list` reports CLOSED molecules under "Active", so read status
  # from the molecule itself rather than trusting that listing.
  mols="$(bd list --type molecule --status open --json 2>/dev/null)" || mols='[]'
  mol="$(printf '%s' "$mols" | jq -r --arg e "$epic" \
        '[.[] | select((.metadata.epic // "") == $e)] | .[0].id // empty' 2>/dev/null)"
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
