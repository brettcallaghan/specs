#!/usr/bin/env bash
# bd-health.sh — composite graph/quality health check for a bead backlog.
#
# Runs whichever Beads health commands exist in the installed `bd` and
# summarizes them. Designed for phase 6 of the groom-beads skill.
#
# Exit status:
#   0  no hard problems (lint warnings may still be present)
#   1  hard problems found (DAG cycles, orphans, or a failing validate)
#   2  `bd` not found / not a beads repo
#
# Usage: bd-health.sh            # check the current repo's backlog
#        BD=bd bd-health.sh      # override the bd binary
#
# The script never mutates beads; it only reports. Grooming should fix what it
# surfaces and re-run.

set -uo pipefail
BD="${BD:-bd}"

if ! command -v "$BD" >/dev/null 2>&1; then
  echo "bd-health: '$BD' not found on PATH" >&2
  exit 2
fi
if [ ! -d ".beads" ]; then
  echo "bd-health: no .beads/ directory here — run from a beads repo root" >&2
  exit 2
fi

hard_fail=0
ran_any=0
skipped_hard=0

# $1 is a subcommand that may be MULTIPLE words ("swarm validate"), so it must
# word-split into separate argv entries. Quoting it would invoke a single
# subcommand literally named "swarm validate", which no `bd` has — every such
# check would then report itself unavailable and silently never run.
# shellcheck disable=SC2086
have() { "$BD" $1 --help >/dev/null 2>&1; }

run_check() {
  # $1 = subcommand, $2 = human label, $3 = "hard" if failure should fail the run
  local sub="$1" label="$2" severity="${3:-soft}"
  if ! have "$sub"; then
    printf '  - %-18s (skipped: `bd %s` not available)\n' "$label" "$sub"
    # A skipped hard check is a gap in coverage, not a pass. Track it so the
    # summary can say which integrity guarantees this run did NOT establish.
    if [ "$severity" = "hard" ]; then skipped_hard=$((skipped_hard + 1)); fi
    return
  fi
  ran_any=1
  local out rc
  # shellcheck disable=SC2086  # same deliberate word-split as `have`
  out="$("$BD" $sub 2>&1)"; rc=$?
  if [ $rc -eq 0 ] && [ -z "${out//[[:space:]]/}" ]; then
    printf '  ✓ %-18s clean\n' "$label"
  else
    printf '  ✗ %-18s\n' "$label"
    # indent the tool output
    printf '%s\n' "$out" | sed 's/^/      /'
    if [ "$severity" = "hard" ]; then hard_fail=1; fi
  fi
}

echo "=== bd health check ==="
echo "Graph integrity (hard — must be clean before a run):"
run_check "swarm validate" "swarm validate" hard   # cycles / orphans / disconnected per epic
run_check "orphans"        "orphans"        hard
echo
echo "Quality & hygiene (advisory):"
run_check "lint"           "lint"           soft    # missing acceptance / repro / success criteria
run_check "stale"          "stale"          soft    # forgotten in-progress work
run_check "doctor"         "doctor"         soft    # infra / hooks / sync
echo
if [ "$hard_fail" -ne 0 ]; then
  echo "RESULT: hard problems found — fix cycles/orphans before dispatching waves."
  exit 1
fi
if [ "$ran_any" -eq 0 ]; then
  echo "RESULT: NO CHECKS RAN — this installed \`bd\` exposes none of them." >&2
  echo "        This is not a clean bill of health. Validate the graph another way." >&2
  exit 2
fi
if [ "$skipped_hard" -ne 0 ]; then
  echo "RESULT: no hard problems among the checks that ran, but $skipped_hard hard"
  echo "        check(s) were unavailable — cycles/orphans are NOT fully verified."
  exit 0
fi
echo "RESULT: no hard problems. Address any advisory warnings above where practical."
exit 0
