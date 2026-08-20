#!/usr/bin/env bash
# Stop hook: push the tracker to origin refs/dolt/data at the end of every turn.
#
# The Dolt remote is this repository's source of truth and the container is
# ephemeral, so a bead created or closed locally is lost on reclaim until it is
# pushed. Making that mechanical is the point: an instruction telling agents to
# remember `bd dolt push` outside the close ceremony is a rule that gets
# followed most of the time, which for durability is the same as not at all.
#
# A no-op push costs ~2s, so this runs unconditionally rather than trying to
# detect whether anything changed — a wrong "nothing to push" loses beads,
# which is worse than the two seconds.
#
# Never fails the turn: a push failure is reported and exits 0.
set -uo pipefail

command -v bd >/dev/null 2>&1 || exit 0
[ -d .beads ] || exit 0
[ "${BB_DOLT_AUTOPUSH:-1}" = 0 ] && exit 0

if ! out="$(timeout 120 bd dolt push 2>&1)"; then
  printf '[dolt-push] WARNING: `bd dolt push` failed — tracker changes are local only.\n%s\n' \
    "$(printf '%s' "$out" | tail -3)" >&2
fi
exit 0
