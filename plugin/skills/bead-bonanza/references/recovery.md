# bead-bonanza Recovery and Progress Streaming

How the orchestrator sees a run in flight, and how it recovers when a
subagent dies mid-ceremony. Read this WHEN a subagent times out, returns an
API error instead of a summary, or is expected to run past 5 minutes.

`${TMPDIR:-/tmp}` below is the run root resolved in Phase 0 step 4.

## Contents

- Subagent crash / mid-step timeout recovery — the four states and what to do
- Mid-stream visibility (progress streaming) — the JSONL event schema and the
  literal-quoting rule that avoids a permission prompt per event
- Crash-resilient close — why the retro file is written before the bead is
  closed, and what it buys on recovery

---

## Subagent crash / mid-step timeout recovery

If a subagent times out, returns an "API Error" instead of a summary, or otherwise dies before its summary line — check its progress JSONL to determine what state it left things in. Then:

1. **Read the progress file** (`<run-id>/<bead-id>.progress.jsonl`). The last phase tells you how far it got.
2. **If `gates-passed` was emitted but `committed` was not** — the work is done in the working tree, only the close+commit ceremony is incomplete. This is the **happy crash path**:
   - Read the retro file at `<run-id>/<bead-id>.retro.md`. If present, use it as the close body.
   - If the retro file is missing (subagent crashed before writing it), reconstruct minimally: the original bead description plus a short Retro noting "Subagent timed out before retro write; orchestrator finished close from working-tree state."
   - Run `bd update <id> --status closed --body-file <retro-file-or-tmp>`.
   - Run `bd export` into `.beads/issues.jsonl` first (the subagent may have crashed before exporting). Then stage source paths, any regenerated fixtures, `.beads/issues.jsonl`, `.beads/interactions.jsonl`. Inspect `git status --porcelain` carefully — only stage what belongs to this bead.
   - Commit with the canonical message format (`<id> P<priority> <title>`).
   - Push (if appropriate for the bucket).
   - Verify the bead is now `closed` via `bd show <id>`.
3. **If the last phase was `implementing` or earlier** — the work is partial. Stash with `git stash push -m "bead-bonanza recovery: <id>"`, mark the bead `blocked` with a recovery note, and surface to the user. Don't try to finish someone else's half-done work.
4. **If the progress file is empty or missing** — the subagent never started. Check `bd show <id>` to confirm status; if it's still `open`, retry the subagent. If it's `in_progress`, claim it back with `bd update <id> --status open` first.

The retro file is what makes the happy crash path mechanical instead of inferential — that's why it's non-negotiable.

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

**Events are FULLY LITERAL single-quoted strings** — `printf '%s\n' '{"bead":"demo-x","phase":"committed","sha":"abc1234"}' >> <file>` with the sha typed literally from the git output. Never embed `$(date ...)`, `"$var"`, or any runtime substitution: the permission system flags runtime-determined printf operands (zsh arith-eval injection heuristic) and prompts the user on EVERY event write. No timestamps — the orchestrator's monitor notification is timestamped on arrival, so `ts` never earned its keep.

Batch subagents (fast lane, shared-surface batch) write events for every bead they process to the same per-bead progress files (one file per bead-id, even from the same batch subagent).

**Monitor any subagent past 5 minutes.** If a subagent is expected to (or actually does) run >5 min — true for any non-trivial bead given full test runs — the orchestrator must arm a `Monitor` on its progress JSONL so the user sees per-phase events live. Don't leave the user staring at silence for 30+ minutes. Pattern:

```
tail -n +1 -F <run-dir>/<bead-id>.progress.jsonl | grep -E --line-buffered '"phase":"(committed|blocked)"'
```

For batch subagents, tail multiple files in one Monitor. **Filter to `committed|blocked` only** — those are the events the orchestrator acts on. Relaying every phase (implementing, gates-running, …) wakes the orchestrator per event and adds a conversation turn each time for narration nobody needs; the per-bead progress files still hold the full phase history for crash recovery.

## Crash-resilient close: write the retro to a file before bd update

After `gates-passed` and BEFORE the `bd update --status closed` call, the subagent **must** write its proposed Retro section to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/<bead-id>.retro.md`. Then emit `{"phase":"retro-written"}`. Only then proceed to the close ceremony.

This is the recovery anchor. If the subagent times out, crashes, or otherwise dies between `gates-passed` and the commit, the orchestrator can:
1. Read the retro file.
2. Run `bd update <id> --status closed --body-file -` itself (concatenating original description + retro).
3. Stage, commit, push.

No information is lost; recovery is mechanical instead of inferential. **This is non-negotiable** — the cost is one extra file write per bead; the savings on a single recovery pay for years of runs.

