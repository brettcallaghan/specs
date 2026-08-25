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
failed_hard_labels=""

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
  # Score on EXIT CODE, never on whether the check printed something. Several
  # `bd` checks report success in prose ("✓ No orphaned issues found"), so an
  # emptiness test marks a passing check as failed — and then the summary below
  # names faults that do not exist, which reads as a hard blocker.
  if [ $rc -eq 0 ]; then
    if [ -z "${out//[[:space:]]/}" ]; then
      printf '  ✓ %-18s clean\n' "$label"
    else
      # Passed, but had something to say. Show it; do not score it as failure.
      printf '  ✓ %-18s\n' "$label"
      printf '%s\n' "$out" | sed 's/^/      /'
    fi
  else
    printf '  ✗ %-18s (exit %d)\n' "$label" "$rc"
    # indent the tool output
    printf '%s\n' "$out" | sed 's/^/      /'
    if [ "$severity" = "hard" ]; then
      hard_fail=1
      failed_hard_labels="${failed_hard_labels:+$failed_hard_labels, }$label"
    fi
  fi
}

# `bd swarm validate` validates ONE epic and requires its id; called bare it
# exits non-zero with an arity error. The generic scorer would read that as a
# graph failure, and `have()` cannot catch it because `--help` succeeds. So it
# gets its own runner: enumerate the epics and validate each.
#
# A repo with no epics is the common case and is NOT a failure — there is
# simply nothing for this check to validate. That is a skipped hard check:
# recorded as a gap in coverage so the summary never implies the per-epic
# integrity guarantee was established when it was not.
run_swarm_validate() {
  local label="swarm validate"
  if ! have "swarm validate"; then
    printf '  - %-18s (skipped: `bd %s` not available)\n' "$label" "swarm validate"
    skipped_hard=$((skipped_hard + 1))
    return
  fi

  # Active epics only — no `--all`. A closed epic's subgraph is not something a
  # run can act on, so validating it would be noise at best and a false hard
  # failure at worst.
  #
  # Top-level "id" keys only. Sibling keys are `depends_on_id` / `issue_id`,
  # which this pattern does not match because it anchors the whole key.
  local epics
  epics="$("$BD" list --type epic --json 2>/dev/null \
            | sed -n 's/^[[:space:]]*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')"

  if [ -z "${epics//[[:space:]]/}" ]; then
    printf '  - %-18s (skipped: no epics in this repo to validate)\n' "$label"
    skipped_hard=$((skipped_hard + 1))
    return
  fi

  local epic out rc failed=0 n=0
  while IFS= read -r epic; do
    [ -n "$epic" ] || continue
    n=$((n + 1))
    out="$("$BD" swarm validate "$epic" 2>&1)"; rc=$?
    if [ $rc -ne 0 ]; then
      failed=1
      printf '  ✗ %-18s %s (exit %d)\n' "$label" "$epic" "$rc"
      printf '%s\n' "$out" | sed 's/^/      /'
    fi
  done <<EOF
$epics
EOF

  ran_any=1
  if [ "$failed" -ne 0 ]; then
    hard_fail=1
    failed_hard_labels="${failed_hard_labels:+$failed_hard_labels, }$label"
  else
    printf '  ✓ %-18s clean (%d epic(s))\n' "$label" "$n"
  fi
}

echo "=== bd health check ==="
echo "Graph integrity (hard — must be clean before a run):"
run_swarm_validate                                # cycles / orphans / disconnected per epic
run_check "orphans"        "orphans"        hard
echo
echo "Quality & hygiene (advisory):"
run_check "lint"           "lint"           soft    # missing acceptance / repro / success criteria
run_check "stale"          "stale"          soft    # forgotten in-progress work
run_check "doctor"         "doctor"         soft    # infra / hooks / sync
echo
if [ "$hard_fail" -ne 0 ]; then
  # Name the checks that actually failed. Asserting a fixed pair of faults
  # ("cycles/orphans") sends grooming looking for problems the run never found.
  echo "RESULT: hard problems found in: $failed_hard_labels"
  echo "        Fix what those checks reported before dispatching anything."
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
