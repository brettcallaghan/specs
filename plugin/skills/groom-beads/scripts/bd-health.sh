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
have() { "$BD" "$1" --help >/dev/null 2>&1; }

run_check() {
  # $1 = subcommand, $2 = human label, $3 = "hard" if failure should fail the run
  local sub="$1" label="$2" severity="${3:-soft}"
  if ! have "$sub"; then
    printf '  - %-18s (skipped: `bd %s` not available)\n' "$label" "$sub"
    return
  fi
  local out rc
  out="$("$BD" "$sub" 2>&1)"; rc=$?
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
echo "RESULT: no hard problems. Address any advisory warnings above where practical."
exit 0
