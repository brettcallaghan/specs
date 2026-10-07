#!/usr/bin/env bash
# Stop hook: push the tracker to origin's beads/data branch at the end of every turn.
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

# If metadata.json names a database that is not on disk, bd would create an
# empty one under that name and push it. That happens on a checkout whose
# .beads/ predates a tracker move; the SessionStart hook moves it aside.
_db="$(python3 -c 'import json;print(json.load(open(".beads/metadata.json")).get("dolt_database",""))' 2>/dev/null)" || _db=""
if [ -n "$_db" ] && [ ! -d ".beads/embeddeddolt/$_db" ]; then
  printf '[dolt-push] skipped: database %s is not in .beads/embeddeddolt/ — run the SessionStart hook.\n' "$_db" >&2
  exit 0
fi

# Push only to the beads/data branch. A remote without that ref — a clone
# scripts/beads-setup.sh has not configured — pushes to Dolt's default
# refs/dolt/data instead, which recreates the retired tracker.
_ref="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["remotes"]["origin"]["params"].get("git_ref",""))' ".beads/embeddeddolt/$_db/.dolt/repo_state.json" 2>/dev/null)" || _ref=""
if [ "$_ref" != "refs/heads/beads/data" ]; then
  printf '[dolt-push] skipped: the Dolt remote does not target refs/heads/beads/data — run ./scripts/beads-setup.sh.\n' >&2
  exit 0
fi

if ! out="$(timeout 120 bd dolt push 2>&1)"; then
  printf '[dolt-push] WARNING: `bd dolt push` failed — tracker changes are local only.\n%s\n' \
    "$(printf '%s' "$out" | tail -3)" >&2
fi
exit 0
