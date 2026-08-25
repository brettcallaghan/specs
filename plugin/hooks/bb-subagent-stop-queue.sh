#!/usr/bin/env bash
# SubagentStop hook: refuse a worker's completion claim while its own queue
# still holds work it neither finished nor declared blocked. Exit 2 blocks the
# stop, which moves acceptance from "the coordinator verifies testimony" to
# "the harness will not accept the claim".
#
# WHAT IT CANNOT DO, stated first so nobody expects it: it cannot detect a
# STALL. A stalled subagent never finishes, so SubagentStop never fires, and
# the absence of an event is not an event. This catches a worker that stops
# EARLY, not one that stops never.
#
# THE VERDICT, and why these two queries and no others (all measured on bd
# 1.2.2, transcript in the bead):
#
#   ready       = bd ready --assignee <id> --json                 | length
#   in_progress = bd list  --assignee <id> --status in_progress --json | length
#   REFUSE iff ready > 0 or in_progress > 0.
#
#   BLOCKED IS NOT INCOMPLETE, and this is the failure mode that would trap a
#   worker forever. A worker that took `run-bead`'s Blocked path leaves its
#   bead at status `blocked`. `bd ready` excludes blocked beads and
#   `--status in_progress` does not match them, so both counts read 0 and the
#   worker is ALLOWED to stop. Measured, one bead, four states:
#     open        -> ready=1 in_progress=0  REFUSE   (never started)
#     in_progress -> ready=0 in_progress=1  REFUSE   (claimed, not finished)
#     blocked     -> ready=0 in_progress=0  ALLOW    <-- the one that matters
#     closed      -> ready=0 in_progress=0  ALLOW
#   And a blocked bead beside an unstarted one still reads ready=1: REFUSE.
#
# FOUR PROPERTIES, in the order they matter.
#
# 1. INERT UNTIL ARMED. Every adopter of this plugin inherits this hook. It
#    does nothing until `--arm` has written state for this project. Absent
#    state is the normal case, not an error.
#
# 2. SCOPED, OR SILENT. The agent id comes from the SubagentStop payload on
#    stdin and from nowhere else — a hook is spawned by the harness, so an
#    id the coordinator exported never reaches this process, and a single
#    id on disk could not say WHICH of several concurrent workers stopped.
#    With no id in the payload the hook exits 0 and says why. It must never
#    fall back to an unscoped query: measured, `bd ready --assignee ""`
#    returned 3 rows and `bd list --assignee "" --status in_progress`
#    returned 2 on a tracker where the intended agent held none.
#
# 3. BOUNDED, PER AGENT. At most $BB_SUBAGENT_HOOK_MAX_BLOCKS consecutive
#    blocks for any one agent id (default 3 — lower than the Stop hook's,
#    because a worker cannot dispatch its way out of a queue the way a
#    coordinator can). The counter is on disk, keyed by agent, and is NOT
#    reset at the cap: past it the hook stands down for that agent until its
#    queue drains or an operator re-arms. A bound that re-arms itself is not
#    a bound. Per agent, so one wedged worker cannot spend another's budget.
#
# 4. FAILS OPEN, ALWAYS. Every path that is not "armed, scoped to a real
#    agent id, and provably holding work" exits 0. A hook that blocks on its
#    own malfunction is a session an operator has to kill.
#
# KILL SWITCH, two forms, both printed on every block:
#   bb-subagent-stop-queue.sh --disarm    remove this project's state
#   BB_SUBAGENT_HOOK=0                    inert everywhere, no files touched
#
# Modes: --arm | --disarm | --status | --self-test | (no args) = hook.
#
# Environment:
#   BB_SUBAGENT_HOOK=0             inert; the kill switch that touches no files
#   BB_SUBAGENT_HOOK_MAX_BLOCKS    the bound (default 3)
#   BB_SUBAGENT_HOOK_DIR           override the state directory
#   BB_SUBAGENT_HOOK_BD            the `bd` executable to ask (default `bd`)
set -uo pipefail

ME="$(basename "$0")"
MAX="${BB_SUBAGENT_HOOK_MAX_BLOCKS:-3}"
case "$MAX" in ''|*[!0-9]*) MAX=3 ;; esac
BD="${BB_SUBAGENT_HOOK_BD:-bd}"

# ------------------------------------------------------------------ state dir
# Outside the repository on purpose: an arming file inside the working tree
# would show as an untracked change, and this project's close ceremony stages
# everything dirty in the worktree — the kill switch would end up committed.
state_dir() {
  if [ -n "${BB_SUBAGENT_HOOK_DIR:-}" ]; then printf '%s\n' "$BB_SUBAGENT_HOOK_DIR"; return; fi
  _root="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}}"
  _proj="${CLAUDE_PROJECT_DIR:-$PWD}"
  _key="$(basename "$_proj")-$(printf '%s' "$_proj" | cksum | cut -d' ' -f1)"
  printf '%s\n' "$_root/bb-subagent-stop-queue/$_key"
}

warn() { printf '[bb-subagent-stop-queue] %s\n' "$*" >&2; }

# A counter file per agent id. The id is agent-chosen text, so it is reduced to
# a filename rather than used as one.
slug() { printf '%s' "$1" | tr -c 'A-Za-z0-9._-' '_' | cut -c1-64; }

# ------------------------------------------------------- agent id from stdin
# The payload is JSON on stdin. Field name is not something this hook can
# assume, so it accepts any of the plausible spellings and treats "none of
# them" as "not scoped" rather than as an error. Reading stdin is wrapped so a
# hook invoked with no stdin at all cannot hang: callers that have nothing to
# say close it, and the redirect in mode_hook guarantees a source either way.
agent_id_from() { # agent_id_from <json>  ->  id | ""
  [ -n "$1" ] || return 0
  if command -v python3 >/dev/null 2>&1; then
    printf '%s' "$1" | python3 -c '
import json, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(0)
if not isinstance(d, dict):
    sys.exit(0)
for k in ("agent_id", "subagent_id", "agentId", "subagentId", "agent"):
    v = d.get(k)
    if isinstance(v, str) and v.strip():
        print(v.strip())
        break
' 2>/dev/null
  fi
}

# ------------------------------------------------------------------- counting
# Never `jq length` bare: `bd ready --json` returns `null`, not `[]`, when the
# queue is empty. Anything that is not an array counts as 0 — the fail-open
# direction. A non-zero exit or empty output returns 1 so the caller can tell
# "could not count" from "counted zero"; conflating those two is what makes a
# hook block on a missing dependency.
count_json() { # count_json <bd args...>  ->  prints a number, or returns 1
  _out="$("$BD" "$@" --json 2>/dev/null)" || return 1
  [ -n "$_out" ] || return 1
  if command -v jq >/dev/null 2>&1; then
    printf '%s' "$_out" | jq -r 'if type=="array" then length else 0 end' 2>/dev/null
  elif command -v python3 >/dev/null 2>&1; then
    printf '%s' "$_out" | python3 -c 'import json,sys
try: d = json.load(sys.stdin)
except Exception: d = None
print(len(d) if isinstance(d, list) else 0)' 2>/dev/null
  else
    return 1
  fi
}

# ---------------------------------------------------------------------- modes
mode_arm() {
  d="$(state_dir)"
  mkdir -p "$d" || return 1
  : > "$d/armed" || return 1
  rm -f "$d"/blocks-*
  printf 'armed: SubagentStop will refuse a worker that stops holding ready or\n'
  printf '       in-progress beads. A BLOCKED bead is not held against it.\n'
  printf 'bound: at most %s consecutive blocks PER AGENT, then it stands down.\n' "$MAX"
  printf 'state: %s   (disarm with: %s --disarm)\n' "$d" "$ME"
}

mode_disarm() {
  d="$(state_dir)"
  rm -rf "$d"
  printf 'disarmed: %s removed. The SubagentStop hook is inert for this project.\n' "$d"
}

mode_status() {
  d="$(state_dir)"
  if [ ! -f "$d/armed" ]; then
    printf 'inert: not armed. (state would live at %s)\n' "$d"
    return 0
  fi
  printf 'armed:  yes\n'
  printf 'bound:  %s consecutive blocks per agent\n' "$MAX"
  for f in "$d"/blocks-*; do
    [ -f "$f" ] || continue
    printf 'blocks: %s used by %s\n' "$(cat "$f" 2>/dev/null || echo 0)" "$(basename "$f" | sed 's/^blocks-//')"
  done
  printf 'state:  %s\n' "$d"
}

# ------------------------------------------------------------------ hook mode
mode_hook() {
  [ "${BB_SUBAGENT_HOOK:-1}" = 0 ] && return 0

  d="$(state_dir)"
  [ -f "$d/armed" ] || return 0                   # not armed: the normal case

  payload="$(cat 2>/dev/null)"
  agent="$(agent_id_from "$payload")"
  agent="${agent%%[[:space:]]*}"
  # Unscoped is not "empty" — it is the whole tracker. Never query it.
  if [ -z "$agent" ]; then
    warn "the SubagentStop payload carried no agent id; not blocking."
    return 0
  fi

  command -v "$BD" >/dev/null 2>&1 || { warn "$BD not on PATH; not blocking."; return 0; }

  # bd resolves its database from the repository, so ask from the project root.
  cd "${CLAUDE_PROJECT_DIR:-.}" 2>/dev/null || true

  ready="$(count_json ready --assignee "$agent")"
  case "$ready" in ''|*[!0-9]*) warn "could not count $agent's ready queue; not blocking."; return 0 ;; esac

  busy="$(count_json list --assignee "$agent" --status in_progress)"
  case "$busy" in ''|*[!0-9]*) warn "could not count $agent's in-progress beads; not blocking."; return 0 ;; esac

  if [ "$ready" -eq 0 ] && [ "$busy" -eq 0 ]; then
    rm -f "$d/blocks-$(slug "$agent")"            # a drained queue restores the budget
    return 0
  fi

  cf="$d/blocks-$(slug "$agent")"
  n="$(cat "$cf" 2>/dev/null)"
  case "$n" in ''|*[!0-9]*) n=0 ;; esac
  n=$((n + 1))
  if ! printf '%s\n' "$n" > "$cf" 2>/dev/null; then
    # Can't persist the bound, so it cannot be enforced. Standing down beats
    # blocking on an unaccounted budget -- see property 4 above.
    warn "cannot write the block counter ($cf); not blocking."
    return 0
  fi

  if [ "$n" -gt "$MAX" ]; then
    # Deliberately does NOT reset the counter. Past the cap this stays quiet
    # for this agent until its queue drains or someone re-arms.
    warn "$agent still holds work (ready=$ready in_progress=$busy), but this hook"
    warn "has already blocked it $MAX times. Standing down so it can return."
    warn "Re-arm with \`$ME --arm\` once the reason its queue will not drain is fixed."
    return 0
  fi

  # stderr on exit 2 is what the subagent reads as the reason.
  warn "$agent is stopping with work still open: ready=$ready, in_progress=$busy."
  warn "Finish each bead, or set it \`blocked\` with a reason — a BLOCKED bead is"
  warn "accepted here, an unfinished one is not. (block $n of $MAX)"
  warn "To stop anyway: \`$ME --disarm\`, or set BB_SUBAGENT_HOOK=0."
  return 2
}

# ------------------------------------------------------------------ self-test
# No database, no repository, no network: `bd` is stubbed on PATH and the state
# directory is a temp dir. THE BOUND IS ASSERTED FIRST, before the happy path,
# because a hook that can block is a trap until its bound is proven.
mode_self_test() {
  t=0 f=0
  ok()  { t=$((t+1)); printf '  ok   %s\n' "$1"; }
  bad() { t=$((t+1)); f=$((f+1)); printf '  FAIL %s\n' "$1"; }
  is()  { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (expected $3, got $2)"; fi; }

  tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
  bin="$tmp/bin"; mkdir -p "$bin"
  export BB_SUBAGENT_HOOK_DIR="$tmp/state"
  export BB_SUBAGENT_HOOK_MAX_BLOCKS=3
  MAX=3
  self="$(cd "$(dirname "$0")" && pwd)/$(basename "$0")"

  # The stub answers `ready` and `list` differently, which is the whole point:
  # a blocked bead shows in NEITHER, and that must read as drained.
  stub() { # stub <ready-json> <in-progress-json>
    { printf '#!/bin/sh\n'
      printf 'case "$1" in\n'
      printf "  ready) printf '%%s' '%s' ;;\n" "$1"
      printf "  list)  printf '%%s' '%s' ;;\n" "$2"
      printf "  *)     printf '%%s' 'null' ;;\n"
      printf 'esac\n'
    } > "$bin/bd"
    chmod +x "$bin/bd"
  }
  # Every hook invocation gets a payload on stdin, as the harness would send.
  run() { # run [agent-id]
    printf '{"hook_event_name":"SubagentStop","agent_id":"%s"}' "${1-dev-r7k-1}" \
      | PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?
  }

  echo "bb-subagent-stop-queue --self-test"

  # ---- THE BOUND, first. A worker whose queue never drains must be released.
  stub '[{"id":"a"}]' 'null'
  PATH="$bin:$PATH" "$self" --arm >/dev/null 2>&1
  seq=""
  i=0; while [ "$i" -lt 6 ]; do seq="$seq$(run)"; i=$((i+1)); done
  is "blocks exactly MAX times, then stands down" "$seq" "222000"
  is "stays stood down" "$(run)$(run)" "00"

  # ---- the bound is PER AGENT: one wedged worker must not spend another's
  is "a different agent still has its full budget" "$(run dev-r7k-2)" 2

  # ---- THE TRAP THIS BEAD EXISTS FOR: blocked is not incomplete.
  stub 'null' 'null'          # a blocked bead appears in NEITHER query
  is "a worker holding only a BLOCKED bead may stop" "$(run dev-r7k-9)" 0
  stub '[]' '[]'
  is "empty lists also read as drained" "$(run dev-r7k-9)" 0

  # ---- the refusals
  stub '[{"id":"a"}]' 'null'
  is "an unstarted ready bead is refused" "$(run dev-r7k-3)" 2
  stub 'null' '[{"id":"a"}]'
  is "an in_progress bead is refused" "$(run dev-r7k-4)" 2
  # A blocked bead is invisible to both queries, so "blocked beside unfinished
  # work" IS "unfinished work" — the blocked one must not soften the verdict.
  stub '[{"id":"a"}]' '[{"id":"b"}]'
  is "blocked BESIDE unfinished work is still refused" "$(run dev-r7k-5)" 2

  # ---- a drained queue restores the budget
  stub '[{"id":"a"}]' 'null'
  is "budget spent again" "$(run dev-r7k-6)$(run dev-r7k-6)" "22"
  stub 'null' 'null'
  is "drained restores it" "$(run dev-r7k-6)" 0
  stub '[{"id":"a"}]' 'null'
  is "and it blocks from zero again" "$(run dev-r7k-6)" 2

  # ---- fail-open paths: none of these may return 2
  is "no agent id in the payload does not block" \
     "$(printf '{"hook_event_name":"SubagentStop"}' | PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?)" 0
  is "an empty agent id does not query unscoped" \
     "$(printf '{"agent_id":""}' | PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?)" 0
  is "an unparseable payload does not block" \
     "$(printf 'not json at all' | PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?)" 0
  is "an empty payload does not block" \
     "$(printf '' | PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?)" 0
  is "a JSON array payload does not block" \
     "$(printf '[1,2]' | PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?)" 0
  is "bd missing does not block" \
     "$(printf '{"agent_id":"dev-r7k-7"}' | BB_SUBAGENT_HOOK_BD=bb-no-such-command PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?)" 0
  stub 'not json at all' 'null'
  is "unparseable bd output does not block" "$(run dev-r7k-7)" 0
  stub '' ''
  is "empty bd output does not block" "$(run dev-r7k-7)" 0

  # -- the counter cannot be written: fail OPEN, not closed. A directory at
  # the counter path fails the write for every user including root, unlike
  # chmod'ing the parent (which a root-run self-test would not observe).
  stub '[{"id":"a"}]' 'null'
  mkdir -p "$BB_SUBAGENT_HOOK_DIR/blocks-$(printf '%s' dev-r7k-unwritable | tr -c 'A-Za-z0-9._-' '_' | cut -c1-64)"
  seq=""
  i=0; while [ "$i" -lt 6 ]; do seq="$seq$(run dev-r7k-unwritable)"; i=$((i+1)); done
  is "unwritable counter never blocks, over repeated calls" "$seq" "000000"
  stub '[{"id":"a"}]' 'null'
  is "kill switch BB_SUBAGENT_HOOK=0" \
     "$(printf '{"agent_id":"dev-r7k-8"}' | BB_SUBAGENT_HOOK=0 PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?)" 0
  PATH="$bin:$PATH" "$self" --disarm >/dev/null 2>&1
  is "disarm makes it inert" "$(run dev-r7k-8)" 0
  is "unarmed does not block" "$(run dev-r7k-1)" 0

  # ---- invocation surface
  PATH="$bin:$PATH" "$self" --arm >/dev/null 2>&1
  is "--status runs clean" \
     "$(PATH="$bin:$PATH" "$self" --status >/dev/null 2>&1; echo $?)" 0
  is "--arm takes no argument" \
     "$(PATH="$bin:$PATH" "$self" --arm extra >/dev/null 2>&1; echo $?)" 2
  is "unknown option is a usage error" \
     "$(PATH="$bin:$PATH" "$self" --nope >/dev/null 2>&1; echo $?)" 2

  echo
  if [ "$f" -eq 0 ]; then echo "RESULT: PASS ($t assertions)"; return 0; fi
  echo "RESULT: FAIL ($f of $t)"; return 1
}

case "${1:-}" in
  --arm)       [ $# -eq 1 ] || { echo "usage: $ME --arm   (it takes no argument)" >&2; exit 2; }
               mode_arm ;;
  --disarm)    mode_disarm ;;
  --status)    mode_status ;;
  --self-test) mode_self_test ;;
  "")          mode_hook </dev/stdin ;;
  *)           echo "usage: $ME [--arm | --disarm | --status | --self-test]" >&2; exit 2 ;;
esac
