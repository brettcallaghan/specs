# bead-bonanza Subagent Crash Recovery

How the orchestrator recovers when a **subagent** dies mid-ceremony. Read this
WHEN a subagent times out, returns an "API Error" instead of a summary, or
otherwise stops before its summary line.

Two neighbouring situations are elsewhere, so this file can be read on its own:
the event stream, the `Monitor`, and the durable out-of-session
trigger are in `progress-streaming.md`; resuming a run whose state you do not
hold — the **orchestrator** was what died — is the **`sprint-recovery`** skill,
invoked by name because a fresh session with no bundle loaded cannot resolve a
path into one.

`${TMPDIR:-/tmp}` below is the run root Phase 0's pre-flight script resolved.

## Contents

- Subagent crash / mid-step timeout recovery — the four states a dead subagent
  can leave behind and what to do about each, including the happy crash path
  where only the close is outstanding

---

## Subagent crash / mid-step timeout recovery

If a subagent times out, returns an "API Error" instead of a summary, or otherwise dies before its summary line — read the state it left behind, never a report of it. Two places hold that state: the bead's worktree (`git -C <worktree> status --porcelain`) and the tracker (`bd show <id>`).

**One worktree per dispatch, so first find which bead it died on.** A batch holds ONE worktree for all of its beads, and the beads it finished are already committed on their own branches inside it. The bead whose `dispatch_state` is not `returned` is where it died — `bd list --all --has-metadata-key worktree --json`, filtered to that path. Beads before it in the batch order need nothing. Recover that one bead in the tree it is already standing in, then re-dispatch the remainder of the batch as a fresh batch.

Then, for that bead:

1. **Is there work in its worktree?** Everything dirty there belongs to this bead: an earlier bead in the same batch would have committed before this one started, so uncommitted changes can only be this one's.
2. **If the worktree holds work and the bead is not closed** — the change is in the tree and the close ceremony is incomplete. **Two things gate the close, and they answer different questions.** This is the **happy crash path**:
   - **Do the checks pass?** Run BOTH kinds in the worktree: the project's declared checks (`sf.toml`'s `[[checks]]`, via the runner) AND any targeted tests the bead itself names. `run-bead` runs both before its own close, so a recovery that runs only the first closes on a weaker bar than the ceremony it is standing in for. Do not skip either on the strength of anything the dead subagent recorded: a run that recorded a clean check and exited 1 with 222 findings is the measured failure this ordering exists to prevent. The cost is one check suite per crash — across 31 beads measured, one crash — against accepting a close whose only evidence is the testimony of an agent that then died.
   - **Is the WORK actually finished?** Green checks say the repository is not broken. They say nothing about whether this bead is done — a subagent that died halfway leaves a tree that compiles and tests clean and satisfies none of its acceptance criteria. **Dispatch a subagent** to read the bead's acceptance criteria against `git -C <worktree> status --porcelain` and the diff, and return `complete` or `incomplete` with the criterion it failed. This is a read of source, so it is not yours to do — you route.
   - **Either one failing means this is not the happy path**: treat it as step 3. A bead is closed only when the checks are green AND the work is verified complete.
   - Read the bead's comments for a `caveat:` line — anything the close must do differently, such as a regen step to run before staging. Honor it before you stage. Most beads carry none.
   - The worktree names the paths: everything dirty in `git -C <worktree> status --porcelain` belongs to this bead, for the reason in step 1. The close stages all of it.
   - Run `../scripts/bb-close-bead.sh --trailers-from <epic> <id> <priority> "<title>"` from inside the bead's worktree, passing no paths — the same script the subagent would have run, reading the same per-run trailers Phase 0 recorded in the run epic's manifest (omit `--trailers-from` WHEN the run resolved none). It appends the trailers to the commit body itself: do **not** `git commit --amend` afterward to add them, which is the exact drift the flag exists to avoid. What the ceremony does, in what order, is SKILL.md's published interface for the script.
   - **If the script cannot run**, fix why it cannot run and re-run it. Do not replay its steps by hand: a hand-rolled staging pathspec that excludes the whole `.beads/` directory silently drops tracked tracker config (`config.yaml`, `.gitignore`, the bd hooks) from the commit while still reporting success — the script's own pathspec (`:!.beads/*.jsonl`) is the one correct exclusion, and reproducing it by hand is exactly the kind of drift this ceremony exists to avoid.
   - The canonical message format is `<id> P<priority> <title>` — the script builds it; a by-hand replay must match it exactly.
   - Push the branch (if appropriate for the batch) — a `git push` of the bead's branch, separate from the `bd dolt push` that lands the tracker.
   - Verify the bead is now `closed` via `bd show <id>`.
3. **If the checks fail, or the acceptance check returned `incomplete`** — the work is partial and uncommitted, and it is sitting safely in the bead's own worktree. **Do not stash and do not commit it.** Either re-enter the worktree and finish the work, or leave the worktree exactly as it is, mark the bead `blocked` with a recovery note naming the worktree path, and surface it to the user. The **`run-bead`** skill's Blocked path is the rule: the worktree IS the durable record of interrupted work, and stashing or committing it to tidy up destroys the signal.
4. **If the worktree is clean or was never created** — the subagent never started. Check `bd show <id>` to confirm status; if it's still `open`, retry the subagent. If it's `in_progress`, claim it back with `bd update <id> --status open` first. Remove any worktree it created before re-dispatching, so the in-flight signal does not lie.

The bead is the anchor. Whatever the subagent managed to record — `retro:`
comments, a `caveat:` line — is in the tracker the moment it was written, and
survives the container that wrote it. State is read from the worktree and the
tracker; there is no per-bead file to find, and nothing to reconcile against.
