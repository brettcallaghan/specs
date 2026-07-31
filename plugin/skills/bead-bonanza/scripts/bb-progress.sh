#!/usr/bin/env bash
# bb-progress.sh — append one progress event to a bead's JSONL progress file.
#
# Why this exists
#   The orchestrator and every subagent emit progress events so a long run is
#   observable. Hand-writing them meant embedding a literal single-quoted JSON
#   string in a printf, because the permission system flags printf operands it
#   cannot prove are constant and prompts on EVERY event write. That rule was
#   easy to state and easy to get wrong, and getting it wrong costs a prompt
#   per event for the rest of the run.
#
#   Here the JSON is built from validated arguments inside one allowlisted
#   script, so the caller never constructs the string and the quoting hazard
#   does not exist.
#
# Usage
#   bb-progress.sh <run-dir> <bead-id> <phase> [key=value ...]
#
#   bb-progress.sh "$RUN_DIR" demo-9h0 claimed
#   bb-progress.sh "$RUN_DIR" demo-9h0 committed sha=a1b2c3d
#   bb-progress.sh "$RUN_DIR" demo-9h0 blocked reason="gates failed twice"
#
# Exit status
#   0  event appended
#   2  bad invocation (unknown phase, missing args, unwritable run dir)

set -euo pipefail

# The phase vocabulary is closed: the orchestrator's Monitor filters on these
# names, and a typo'd phase would be silently invisible rather than an error.
PHASES="claimed implementing gates-running gates-passed retro-written committed pushed blocked"

die() { echo "bb-progress: $*" >&2; exit 2; }

[ $# -ge 3 ] || die "usage: bb-progress.sh <run-dir> <bead-id> <phase> [key=value ...]"

run_dir="$1"; bead="$2"; phase="$3"; shift 3

case " $PHASES " in
  *" $phase "*) ;;
  *) die "unknown phase '$phase'. Valid phases: $PHASES" ;;
esac

[ -n "$bead" ] || die "bead-id must not be empty"
[ -n "$run_dir" ] || die "run-dir must not be empty"

# Validate EVERY argument before touching the filesystem. This loop used to run
# after the mkdir below, so a bad extra field left a freshly created run
# directory behind on a call that then exited 2 — and if the path was wrong,
# that stray directory sat permanently where the real one belonged.
for kv in "$@"; do
  case "$kv" in
    *=*) ;;
    *) die "extra fields must be key=value, got '$kv'" ;;
  esac
  [ -n "${kv%%=*}" ] || die "empty key in '$kv'"
done

# Only now, with the invocation known-good, may we write anything.
if [ ! -d "$run_dir" ]; then
  # Create it rather than failing: a subagent that starts before the
  # orchestrator's mkdir landed should still record its own progress.
  mkdir -p "$run_dir" || die "cannot create run dir: $run_dir"
fi
[ -w "$run_dir" ] || die "run dir is not writable: $run_dir"

# JSON string escaping: backslash and double-quote only. Event values are bead
# ids, short shas, and one-line human reasons — no control characters — so a
# full JSON encoder would be dependency weight for no gain.
esc() { printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'; }

json="{\"bead\":\"$(esc "$bead")\",\"phase\":\"$(esc "$phase")\""
for kv in "$@"; do
  # Shape already validated above, before any filesystem write.
  k="${kv%%=*}"; v="${kv#*=}"
  json="$json,\"$(esc "$k")\":\"$(esc "$v")\""
done
json="$json}"

printf '%s\n' "$json" >> "$run_dir/$bead.progress.jsonl"
