# bead-bonanza Subagent Crash Recovery

How the orchestrator recovers when a **subagent** dies mid-ceremony. Read this
WHEN a subagent times out, returns an "API Error" instead of a summary, or
otherwise stops before its summary line.

Two neighbouring situations are elsewhere, so this file can be read on its own:
the progress-event schema, the `Monitor`, and the durable out-of-session
trigger are in `progress-streaming.md`; resuming a run whose state you do not
hold — the **orchestrator** was what died — is the **`recover-sprint`** skill,
invoked by name because a fresh session with no bundle loaded cannot resolve a
path into one.

`${TMPDIR:-/tmp}` below is the run root Phase 0's pre-flight script resolved.

## Contents

- Subagent crash / mid-step timeout recovery — the four states a dead subagent
  can leave behind and what to do about each, including the happy crash path
  where only the close is outstanding
- Crash-resilient close — why the handoff file is written before the bead is
  closed, and what it buys on recovery

---

## Subagent crash / mid-step timeout recovery

If a subagent times out, returns an "API Error" instead of a summary, or otherwise dies before its summary line — check its progress JSONL to determine what state it left things in. Then:

1. **Read the progress file** (`<run-id>/<bead-id>.progress.jsonl`). The last phase tells you how far it got.
2. **If `gates-passed` was emitted but `committed` was not** — the work is done in the working tree, only the close+commit ceremony is incomplete. This is the **happy crash path**:
   - Read the handoff file at `<run-id>/<bead-id>.handoff.md`. It names any close-specific caveat (a regen step to run first, a fixture the bead regenerated). Honor the caveat before you stage. It names no paths: everything in the bead's worktree is the bead's work, so the close stages all of it.
   - If the handoff file is missing (subagent crashed before writing it), the bead's worktree is the answer: everything dirty in `git -C <worktree> status --porcelain` belongs to this bead, because nothing else was ever implemented there. Add a `retro:` comment on the bead noting that the handoff was never written.
   - Run `../scripts/bb-close-bead.sh --trailers-file "$RUN_DIR/trailers.txt" <id> <priority> "<title>"` from inside the bead's worktree, passing no paths — the same script the subagent would have run, with the same per-run trailers file Phase 0 resolved (omit `--trailers-file` WHEN the run resolved none). It appends the trailers to the commit body itself: do **not** `git commit --amend` afterward to add them, which is the exact drift the flag exists to avoid. What the ceremony does, in what order, is SKILL.md's published interface for the script.
   - **If the script cannot run**, fix why it cannot run and re-run it. Do not replay its steps by hand: a hand-rolled staging pathspec that excludes the whole `.beads/` directory silently drops tracked tracker config (`config.yaml`, `.gitignore`, the bd hooks) from the commit while still reporting success — the script's own pathspec (`:!.beads/*.jsonl`) is the one correct exclusion, and reproducing it by hand is exactly the kind of drift this ceremony exists to avoid.
   - The canonical message format is `<id> P<priority> <title>` — the script builds it; a by-hand replay must match it exactly.
   - Push the branch (if appropriate for the bucket) — a `git push` of the bead's branch, separate from the `bd dolt push` that lands the tracker.
   - Verify the bead is now `closed` via `bd show <id>`.
   - Record the commit yourself: `../scripts/bb-progress.sh "$RUN_DIR" <id> committed sha=$(git rev-parse --short HEAD)` — read the sha from the branch, not from the script's printout, same rule as the per-bead template's step 8.
3. **If the last phase was `implementing` or earlier** — the work is partial and uncommitted, and it is sitting safely in the bead's own worktree. **Do not stash and do not commit it.** Either re-enter the worktree and finish the work, or leave the worktree exactly as it is, mark the bead `blocked` with a recovery note naming the worktree path, and surface it to the user (`operating-principles.md`, "No exceptions: the worktree holds interrupted work").
4. **If the progress file is empty or missing** — the subagent never started. Check `bd show <id>` to confirm status; if it's still `open`, retry the subagent. If it's `in_progress`, claim it back with `bd update <id> --status open` first. Remove any worktree it created before re-dispatching, so the in-flight signal does not lie.

Read the last phase as a starting hypothesis and nothing more: the ledger is
written by the agent doing the work and stops when that agent stops
(`progress-streaming.md`, "The ledger is a hint, never the authority", carries
the order that settles a status question).

The handoff file is what makes the happy crash path mechanical instead of inferential — that's why it's non-negotiable. Findings are not in it: they are `retro:` comments the subagent writes on the bead, so whatever it managed to record is already durable in the tracker.

## Crash-resilient close: write the handoff file before bd update

The handoff file is written after `gates-passed` and BEFORE the close, so a subagent that dies between the two leaves the orchestrator a procedure instead of a guess. That is the whole of what it buys, and it is why **writing it is non-negotiable**: the happy crash path is mechanical only because the file is already on disk. That path is step 2 of "Subagent crash / mid-step timeout recovery" above.

The file deliberately restates neither the phase — `<bead-id>.progress.jsonl` is the single source for that, and a second copy would give a recovering orchestrator two hints to reconcile — nor the findings, which are `retro:` comments on the bead, durable in the tracker the moment they are written.
