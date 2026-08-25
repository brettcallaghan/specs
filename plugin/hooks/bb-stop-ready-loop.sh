#!/usr/bin/env bash
# Stop hook: refuse to end the coordinator's turn while its run epic still has
# ready beads. Exit 2 blocks the stop and continues the conversation, which
# turns the orchestrator's ready-loop from something the model must remember
# into something the harness will not let it skip.
#
# THREE PROPERTIES, IN THE ORDER THEY MATTER.
#
# 1. INERT UNTIL ARMED. The hook does nothing until `--arm <epic>` has written
#    a run epic into its state directory. Every adopter of this plugin inherits
#    this hook; one that acted on installation would seize sessions that have
#    no run in them. Absent state is the normal case, not an error.
#
# 2. BOUNDED. At most $BB_STOP_HOOK_MAX_BLOCKS consecutive blocks per arming
#    (default 8). The counter is on disk, so it outlives the hook process, and
#    it is NOT reset at the cap: past the cap the hook stays quiet until the
#    queue drains or an operator re-arms it. A bound that re-arms itself is not
#    a bound.
#
# 3. FAILS OPEN, ALWAYS. Every path that is not "armed, and the queue is
#    provably non-empty" exits 0. That is the whole safety argument: a hook
#    that blocks on its own malfunction is a session an operator must kill.
#
#    `left=$(bd ready --parent "$RUN_EPIC" --exclude-type=epic --json |
#     jq length); [ "$left" -eq 0 ] || exit 2` fails CLOSED in three ways:
#      a. `[ "$left" -eq 0 ]` returns 2 when $left is not a number, and 2 is
#         the block code -- so `bd` or `jq` missing wedges the session.
#      b. `bd ready --parent ""` is repo-wide rather than empty, so an unset
#         epic id blocks on unrelated work forever.
#      c. Hooks are spawned by the harness, not by the coordinator's shell, so
#         an eval'd `BB_EPIC=` / `RUN_EPIC=` assignment never reaches this
#         process. The epic MUST come from disk -- which makes (b) the default
#         case rather than an edge case.
#
# KILL SWITCH, two forms, both printed on every block:
#   bb-stop-ready-loop.sh --disarm      remove this project's state
#   BB_STOP_HOOK=0                      inert everywhere, no files touched
#
# Modes: --arm <epic> | --disarm | --status | --self-test | (no args) = hook.
#
# Environment:
#   BB_STOP_HOOK=0            inert; the kill switch that touches no files
#   BB_STOP_HOOK_MAX_BLOCKS   the bound (default 8)
#   BB_STOP_HOOK_DIR          override the state directory
#   BB_STOP_HOOK_BD           the `bd` executable to ask (default `bd`)
set -uo pipefail

ME="$(basename "$0")"
MAX="${BB_STOP_HOOK_MAX_BLOCKS:-8}"
case "$MAX" in ''|*[!0-9]*) MAX=8 ;; esac
BD="${BB_STOP_HOOK_BD:-bd}"

# ------------------------------------------------------------------ state dir
# Outside the repository on purpose. An arming file inside the working tree
# would show as an untracked change, and this project's close ceremony stages
# everything dirty in the worktree -- the kill switch would end up committed.
state_dir() {
  if [ -n "${BB_STOP_HOOK_DIR:-}" ]; then printf '%s\n' "$BB_STOP_HOOK_DIR"; return; fi
  _root="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}}"
  _proj="${CLAUDE_PROJECT_DIR:-$PWD}"
  # Keyed by project so two projects on one machine cannot share a counter.
  _key="$(basename "$_proj")-$(printf '%s' "$_proj" | cksum | cut -d' ' -f1)"
  printf '%s\n' "$_root/bb-stop-ready-loop/$_key"
}

warn() { printf '[bb-stop-ready-loop] %s\n' "$*" >&2; }

# ------------------------------------------------------------------- counting
# Never `jq 'length'` bare: `bd ready --json` returns `null`, not `[]`, when the
# queue is empty, and a consumer that assumes a list is exactly the bug this
# hook must not have. Anything that is not an array counts as 0 -- the
# fail-open direction.
count_ready() {
  _epic="$1" _out=""
  _out="$("$BD" ready --parent "$_epic" --exclude-type=epic --json 2>/dev/null)" || return 1
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
  [ -n "${1:-}" ] || { echo "usage: $ME --arm <run-epic-id>" >&2; return 2; }
  d="$(state_dir)"
  mkdir -p "$d" || return 1
  printf '%s\n' "$1" > "$d/epic" || return 1
  rm -f "$d/blocks"
  printf 'armed: the Stop hook will hold this session while %s has ready beads.\n' "$1"
  printf 'bound: at most %s consecutive blocks, then it stands down on its own.\n' "$MAX"
  printf 'state: %s   (disarm with: %s --disarm)\n' "$d" "$ME"
}

mode_disarm() {
  d="$(state_dir)"
  rm -rf "$d"
  printf 'disarmed: %s removed. The Stop hook is inert for this project.\n' "$d"
}

mode_status() {
  d="$(state_dir)"
  if [ ! -f "$d/epic" ]; then
    printf 'inert: no run armed. (state would live at %s)\n' "$d"
    return 0
  fi
  printf 'armed:  %s\n' "$(cat "$d/epic")"
  printf 'blocks: %s of %s used\n' "$(cat "$d/blocks" 2>/dev/null || echo 0)" "$MAX"
  printf 'state:  %s\n' "$d"
}

# ------------------------------------------------------------------ hook mode
mode_hook() {
  [ "${BB_STOP_HOOK:-1}" = 0 ] && return 0

  d="$(state_dir)"
  [ -f "$d/epic" ] || return 0                    # not armed: the normal case

  epic="$(cat "$d/epic" 2>/dev/null)"
  epic="${epic%%[[:space:]]*}"
  # An empty epic id would make `bd ready --parent ""` repo-wide. Never query
  # unscoped from a hook that can block.
  [ -n "$epic" ] || { warn "state file names no epic; not blocking."; return 0; }

  command -v "$BD" >/dev/null 2>&1 || { warn "$BD not on PATH; not blocking."; return 0; }

  # bd resolves its database from the repository, so ask from the project root.
  cd "${CLAUDE_PROJECT_DIR:-.}" 2>/dev/null || true

  left="$(count_ready "$epic")"
  case "$left" in
    ''|*[!0-9]*) warn "could not count ready beads for $epic; not blocking."; return 0 ;;
  esac

  if [ "$left" -eq 0 ]; then
    rm -f "$d/blocks"                             # a drained queue restores the budget
    return 0
  fi

  cf="$d/blocks"
  n="$(cat "$cf" 2>/dev/null)"
  case "$n" in ''|*[!0-9]*) n=0 ;; esac
  n=$((n + 1))
  if ! printf '%s\n' "$n" > "$cf" 2>/dev/null; then
    # Can't persist the bound, so it cannot be enforced. Standing down beats
    # blocking on an unaccounted budget -- see property 3 above.
    warn "cannot write the block counter ($cf); not blocking."
    return 0
  fi

  if [ "$n" -gt "$MAX" ]; then
    # Deliberately does NOT reset the counter. Past the cap this stays quiet
    # until the queue drains or someone re-arms it.
    warn "$epic still has $left ready bead(s), but this hook has already blocked"
    warn "$MAX times. Standing down so the session can end. Re-arm with"
    warn "\`$ME --arm $epic\` once the reason the queue will not drain is fixed."
    return 0
  fi

  # stderr on exit 2 is what the coordinator reads as the reason.
  warn "$epic has $left ready bead(s): the run is not finished."
  warn "Dispatch or account for them, then stop. (block $n of $MAX)"
  warn "To stop anyway: \`$ME --disarm\`, or set BB_STOP_HOOK=0."
  return 2
}

# ------------------------------------------------------------------ self-test
# No database, no repository, no network: `bd` is stubbed on PATH and the state
# directory is a temp dir. What it proves is the bound -- a queue that never
# drains must stop blocking on its own.
mode_self_test() {
  t=0 f=0
  ok()  { t=$((t+1)); printf '  ok   %s\n' "$1"; }
  bad() { t=$((t+1)); f=$((f+1)); printf '  FAIL %s\n' "$1"; }
  is()  { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (expected $3, got $2)"; fi; }

  tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
  bin="$tmp/bin"; mkdir -p "$bin"
  export BB_STOP_HOOK_DIR="$tmp/state"
  export BB_STOP_HOOK_MAX_BLOCKS=3
  MAX=3
  self="$(cd "$(dirname "$0")" && pwd)/$(basename "$0")"

  stub() {   # $1 = what the fake `bd ready --json` prints
    { printf '#!/bin/sh\n'; printf "printf '%%s' '%s'\n" "$1"; } > "$bin/bd"
    chmod +x "$bin/bd"
  }
  run() { PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?; }

  echo "bb-stop-ready-loop --self-test"

  # -- unarmed is inert, whatever the queue says
  stub '[{"id":"x"}]'
  is "unarmed does not block" "$(run)" 0

  # -- armed, and the queue never drains: the bound must end the loop
  PATH="$bin:$PATH" "$self" --arm demo-epic >/dev/null 2>&1
  seq=""
  i=0; while [ "$i" -lt 6 ]; do seq="$seq$(run)"; i=$((i+1)); done
  is "blocks exactly MAX times, then stands down" "$seq" "222000"

  # -- the bound does not re-arm itself on later turns
  is "stays stood down" "$(run)$(run)" "00"

  # -- a drained queue restores the budget
  stub 'null'
  is "empty queue (null, not []) does not block" "$(run)" 0
  stub '[{"id":"x"}]'
  is "budget restored after the queue drained" "$(run)" 2

  # -- fail-open paths: none of these may return 2
  PATH="$bin:$PATH" "$self" --disarm >/dev/null 2>&1
  is "disarm makes it inert" "$(run)" 0

  PATH="$bin:$PATH" "$self" --arm demo-epic >/dev/null 2>&1
  is "kill switch BB_STOP_HOOK=0" \
     "$(BB_STOP_HOOK=0 PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?)" 0

  : > "$BB_STOP_HOOK_DIR/epic"
  is "empty epic id does not query unscoped" "$(run)" 0

  PATH="$bin:$PATH" "$self" --arm demo-epic >/dev/null 2>&1
  is "bd missing does not block" \
     "$(BB_STOP_HOOK_BD=bb-no-such-command PATH="$bin:$PATH" "$self" >/dev/null 2>&1; echo $?)" 0

  stub 'not json at all'
  is "unparseable output does not block" "$(run)" 0
  stub ''
  is "empty output does not block" "$(run)" 0

  # -- the counter cannot be written: fail OPEN, not closed. A directory at
  # the counter path fails the write for every user including root, unlike
  # chmod'ing the parent (which a root-run self-test would not observe).
  PATH="$bin:$PATH" "$self" --arm demo-epic >/dev/null 2>&1
  rm -f "$BB_STOP_HOOK_DIR/blocks"
  mkdir -p "$BB_STOP_HOOK_DIR/blocks"
  stub '[{"id":"x"}]'
  seq=""
  i=0; while [ "$i" -lt 6 ]; do seq="$seq$(run)"; i=$((i+1)); done
  is "unwritable counter never blocks, over repeated calls" "$seq" "000000"
  rmdir "$BB_STOP_HOOK_DIR/blocks" 2>/dev/null

  # -- invocation surface
  is "--arm with no epic is a usage error" \
     "$(PATH="$bin:$PATH" "$self" --arm >/dev/null 2>&1; echo $?)" 2
  is "--status runs clean" \
     "$(PATH="$bin:$PATH" "$self" --status >/dev/null 2>&1; echo $?)" 0
  is "unknown option is a usage error" \
     "$(PATH="$bin:$PATH" "$self" --nope >/dev/null 2>&1; echo $?)" 2

  echo
  if [ "$f" -eq 0 ]; then echo "RESULT: PASS ($t assertions)"; return 0; fi
  echo "RESULT: FAIL ($f of $t)"; return 1
}

case "${1:-}" in
  --arm)       mode_arm "${2:-}" ;;
  --disarm)    mode_disarm ;;
  --status)    mode_status ;;
  --self-test) mode_self_test ;;
  "")          mode_hook ;;
  *)           echo "usage: $ME [--arm <epic> | --disarm | --status | --self-test]" >&2; exit 2 ;;
esac
