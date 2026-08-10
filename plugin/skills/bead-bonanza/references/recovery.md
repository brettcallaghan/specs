# bead-bonanza Recovery and Progress Streaming

How the orchestrator sees a run in flight, and how it recovers when a
subagent dies mid-ceremony. Read this WHEN a subagent times out, returns an
API error instead of a summary, or is expected to run past 5 minutes.

`${TMPDIR:-/tmp}` below is the run root Phase 0's pre-flight script resolved.

## Contents

- Subagent crash / mid-step timeout recovery — the four states and what to do
- Mid-stream visibility (progress streaming) — the JSONL event schema and the
  literal-quoting rule that avoids a permission prompt per event
- Crash-resilient close — why the handoff file is written before the bead is
  closed, and what it buys on recovery

---

## Subagent crash / mid-step timeout recovery

If a subagent times out, returns an "API Error" instead of a summary, or otherwise dies before its summary line — check its progress JSONL to determine what state it left things in. Then:

1. **Read the progress file** (`<run-id>/<bead-id>.progress.jsonl`). The last phase tells you how far it got.
2. **If `gates-passed` was emitted but `committed` was not** — the work is done in the working tree, only the close+commit ceremony is incomplete. This is the **happy crash path**:
   - Read the handoff file at `<run-id>/<bead-id>.handoff.md`. It names the paths to stage and any close-specific caveat (a `WIP <id>` commit to squash, a regen step to run first). Honor the caveat before you stage.
   - If the handoff file is missing (subagent crashed before writing it), derive the paths yourself from `git status --porcelain` — and be strict about it, because in a shared tree some of that dirt belongs to another agent. Add a `retro:` comment on the bead noting that the handoff was never written.
   - Run `../scripts/bb-close-bead.sh <id> <priority> "<title>" <paths>` — the same script the subagent would have run. It closes, exports, stages, and commits in the required order, and leaves the bead's description exactly as grooming wrote it.
   - Run `bd export` into `.beads/issues.jsonl` first (the subagent may have crashed before exporting). Then stage source paths, any regenerated fixtures, `.beads/issues.jsonl`, `.beads/interactions.jsonl`. Inspect `git status --porcelain` carefully — only stage what belongs to this bead.
   - Commit with the canonical message format (`<id> P<priority> <title>`).
   - Push (if appropriate for the bucket).
   - Verify the bead is now `closed` via `bd show <id>`.
3. **If the last phase was `implementing` or earlier** — the work is partial. Stash with `git stash push -m "bead-bonanza recovery: <id>"`, mark the bead `blocked` with a recovery note, and surface to the user. Don't try to finish someone else's half-done work.
4. **If the progress file is empty or missing** — the subagent never started. Check `bd show <id>` to confirm status; if it's still `open`, retry the subagent. If it's `in_progress`, claim it back with `bd update <id> --status open` first.

The handoff file is what makes the happy crash path mechanical instead of inferential — that's why it's non-negotiable. Findings are not in it: they are `retro:` comments the subagent writes on the bead, so whatever it managed to record is already durable in the tracker.

## Mid-stream visibility (progress streaming)

Each subagent appends progress events to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/<bead-id>.progress.jsonl`. Required event phases (one line each):

```jsonl
{"bead":"<id>","phase":"claimed"}
{"bead":"<id>","phase":"implementing"}
{"bead":"<id>","phase":"gates-running"}
{"bead":"<id>","phase":"gates-passed"}
{"bead":"<id>","phase":"retro-written"}
{"bead":"<id>","phase":"committed","sha":"<short>"}
{"bead":"<id>","phase":"pushed"}             # omitted in defer-push mode
{"bead":"<id>","phase":"blocked","reason":"..."}  # only on failure
```

Events are written by `../scripts/bb-progress.sh`, never by hand:

```bash
../scripts/bb-progress.sh "$RUN_DIR" <bead-id> <phase> [key=value ...]
```

It validates the phase against the closed vocabulary above, escapes values, and appends one well-formed line. Building the JSON in the caller is what used to make this fragile — a runtime-substituted `printf` operand trips the permission system's injection heuristic and prompts the user on *every* event write, so the old rule was to hand-type fully literal strings. One allowlisted script removes both the hazard and the rule. No timestamps: the orchestrator's monitor notification is timestamped on arrival, so `ts` never earned its keep.

Batch subagents (fast lane, shared-surface batch) write events for every bead they process to the same per-bead progress files (one file per bead-id, even from the same batch subagent).

**Monitor any subagent past 5 minutes.** If a subagent is expected to (or actually does) run >5 min — true for any non-trivial bead given full test runs — the orchestrator must arm a `Monitor` on its progress JSONL so the user sees per-phase events live. Don't leave the user staring at silence for 30+ minutes. Pattern:

```
tail -n +1 -F <run-dir>/<bead-id>.progress.jsonl | grep -E --line-buffered '"phase":"(committed|blocked)"'
```

For batch subagents, tail multiple files in one Monitor. **Filter to `committed|blocked` only** — those are the events the orchestrator acts on. Relaying every phase (implementing, gates-running, …) wakes the orchestrator per event and adds a conversation turn each time for narration nobody needs; the per-bead progress files still hold the full phase history for crash recovery.

## Crash-resilient close: write the handoff file before bd update

After `gates-passed` and BEFORE the `bd update --status closed` call, the subagent **must** write `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/<bead-id>.handoff.md` — the paths the close must stage, plus any close-specific caveat, and nothing else. Then emit `{"phase":"retro-written"}` (the phase vocabulary predates the rename and is unchanged). Only then proceed to the close ceremony.

This is the recovery anchor. If the subagent times out, crashes, or otherwise dies between `gates-passed` and the commit, the orchestrator can:
1. Read the handoff file for the paths and the caveat.
2. Run `../scripts/bb-close-bead.sh <id> <priority> "<title>" <paths>` itself.
3. Verify with `bd show <id>`, then push if the bucket calls for it.

The file deliberately does not restate the phase — `<bead-id>.progress.jsonl` is the single source for that, and a second copy would give a recovering orchestrator two hints to reconcile. It also carries no findings: those are `retro:` comments on the bead, durable in the tracker the moment they are written. **This is non-negotiable** — the cost is one extra file write per bead; the savings on a single recovery pay for years of runs.

