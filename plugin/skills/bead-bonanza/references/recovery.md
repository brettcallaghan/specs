# bead-bonanza Recovery and Progress Streaming

How the orchestrator sees a run in flight, how it recovers when a subagent
dies mid-ceremony, and how a run is picked back up when the **orchestrator**
is what died. Read this WHEN a subagent times out, returns an API error
instead of a summary, or is expected to run past 5 minutes — and WHEN you
are resuming a bead-bonanza run whose state you do not hold.

`${TMPDIR:-/tmp}` below is the run root Phase 0's pre-flight script resolved.

## Contents

- Orchestrator crash / context loss — the ordered entry procedure for
  resuming a run you have no memory of
- Subagent crash / mid-step timeout recovery — the four states and what to do
- Mid-stream visibility (progress streaming) — the JSONL event schema, the
  rule bounding what that schema proves, and the durable trigger that backs
  the in-session Monitor
- Crash-resilient close — why the handoff file is written before the bead is
  closed, and what it buys on recovery

---

## Orchestrator crash / context loss — resuming a run

The subagent procedure below presumes a live orchestrator that knows the run.
This section is for when that is not true: the orchestrator was interrupted,
compacted past the point of usefulness, or replaced by a fresh session, and
someone has to pick the sprint back up.

**Nothing reports this to you.** A user interrupt that kills the orchestrator
kills the mechanism that would have announced its children's deaths — kill
notifications sit enqueued and undelivered while the UI still shows the task
running. Assume no notification is coming and no notification you *do* find is
proof.

Work these steps in order.

1. **Find the run — derive it from the branch.** A run happens on its own
   dedicated branch, uses one run directory, and that directory is removed
   before the branch merges. So the branch you are on identifies the run, and
   `git log` on it lists the beads that have landed.
   Start there: `git rev-parse --abbrev-ref HEAD`, then
   `git log --oneline <default-branch>..HEAD` — every close commit is
   `<bead-id> P<n> <title>`, so that is the run's completed set, read from the
   record rather than from a report.

   For the plan and the run's artifacts:

   - `ls -l "${TMPDIR:-/tmp}/claude/bead-bonanza/current"` — a symlink
     pre-flight-check.sh repoints at the start of every run. Convenient while
     the container lives; gone after a reclaim.
   - The committed snapshot under `.bonanza/<run-id>/` on the branch, which is
     what survives a reclaim. `ls .bonanza/` and take the directory the branch
     carries. It is written by `scripts/bb-backup-run.sh` after every bead's
     acceptance check, and `.bonanza/<run-id>/SNAPSHOT.md` is self-describing:
     the run id, the as-of time, the live directory's path, the run branch, the
     HEAD when it was taken, and every bead id that had a progress ledger or a
     handoff file. Read that file first — it tells you which run you are
     looking at without your knowing the run id.

   **The live run directory is authoritative; a snapshot never is.** Any
   procedure that reads the backup — this recovery, a retro, a DoD audit —
   first checks whether the live directory still exists and uses it when it
   does. Read the backup only WHEN `$RUN_DIR` is gone, and then **state the
   snapshot's as-of time** in whatever you conclude from it. A copy taken
   mid-run is indistinguishable from a complete one once it is detached from
   its timestamp, and it goes stale within the half hour.

   **Refresh it as you go.** WHEN you resume a run and keep working it, run
   `<skill>/scripts/bb-backup-run.sh "$RUN_DIR"` after each bead's acceptance
   check, exactly as the original orchestrator was to. It re-copies the live
   directory and commits `.bonanza/<run-id>/`, makes no commit when nothing
   changed, and never reads back out of the copy. A snapshot that is committed
   but never pushed dies with the container like everything else — push the
   branch.

   **Never guess the run from directory mtimes.** "Take the newest run-id
   directory under the run root" is wrong whenever an older run's directory is
   touched after a newer one starts. WHEN neither the snapshot nor the symlink
   resolves, reconstruct from `git log` and `bd list`, and **say so in the
   report**: a reconstruction is not a plan.

2. **Read `dispatch.md` in full.** It carries the mandate, the enumerated bead
   list, the standing authorization, the wave table, the gate profile and the
   environmental ledger. **This is the plan — do not re-derive it, and do not
   re-run the filter that produced it.** A filter re-evaluated now returns a
   different set than the one this run committed to, and admits beads nobody
   scoped. Read `## Amendments` too, and reconcile it against the wave table:
   the plan may have moved since it was first written, and the amendments log
   is where it moved.

3. **Establish what is in flight — `git worktree list` is the whole answer.**
   Every dispatched bead runs in its own worktree named for its hash suffix,
   and the orchestrator removes that worktree only when the bead's branch has
   merged. **A worktree that exists is a bead that is not finished.** The state
   is explicit, so nothing has to be inferred:

   ```sh
   cd "$(git rev-parse --show-toplevel)"
   git worktree prune          # drop registrations whose directory is gone
   git worktree list           # every remaining entry is an unfinished bead
   ```

   `git worktree prune` first is not optional. A container reclaimed
   mid-run leaves worktree directories deleted but still registered, and a
   stale entry reads as in-flight forever.

   **Do not infer liveness from file mtimes.** A dirty file's mtime bounds the
   last *write*, not the agent's liveness, so a recency table cannot tell you
   whether an agent is alive — and a second liveness signal that can disagree
   with the worktree list is worse than the one signal that cannot.

   `bd list --status in_progress` is **not** a liveness signal either — a bead
   can be worked without ever being claimed, so the query returns empty while
   an agent is actively editing. Use it only to corroborate what the worktree
   list already told you.

   **To tell an unfinished bead's state, ask its worktree**, not the shared
   checkout: `git -C <worktree> status --porcelain` shows exactly that bead's
   work and nothing else, and `git -C <worktree> log --oneline <base>..` shows
   whether it has committed. No path enumeration and no ownership question —
   everything in that tree belongs to that bead.

4. **Classify each in-scope bead, then corroborate — the tail is a hint.**
   Take `<bead-id>.progress.jsonl`'s last phase as a *starting hypothesis*,
   then resolve it per **The ledger is a hint, never the authority** below
   (`bd show <id>` → `git log --oneline --grep=<id>` → the ledger, first
   answer wins). Never act on the tail alone:

   - **absent** — may never have started, or may have started and written
     nothing. Confirm with `bd show <id>`.
   - **`claimed` / `implementing`** — may have died mid-work, or may be
     finished and committed with nothing appended after the crash. Check
     `git log --grep=<id>` before you touch the tree.
   - **`gates-passed` / `retro-written`** — the work is in the tree and only
     the close is outstanding. Read `<bead-id>.handoff.md` for its caveat,
     honor it, then run
     `../scripts/bb-close-bead.sh <id> <priority> "<title>"` from inside the
     bead's worktree, passing no paths. The handoff file carries **only a
     caveat** — it is not a body for the close, and the close script leaves
     the bead's description as grooming wrote it.
   - **`committed`** — verify the recorded sha is an ancestor of `HEAD`
     (`git merge-base --is-ancestor <sha> HEAD`) before believing it.

5. **Verify every claim against `bd` and `git` yourself.** A completion
   notification is not proof — notifications go missing entirely. Neither is a
   summary, a backup copy of a progress file, or your own earlier report. Read
   the live state, name the source you read, then act.

## Subagent crash / mid-step timeout recovery

If a subagent times out, returns an "API Error" instead of a summary, or otherwise dies before its summary line — check its progress JSONL to determine what state it left things in. Then:

1. **Read the progress file** (`<run-id>/<bead-id>.progress.jsonl`). The last phase tells you how far it got.
2. **If `gates-passed` was emitted but `committed` was not** — the work is done in the working tree, only the close+commit ceremony is incomplete. This is the **happy crash path**:
   - Read the handoff file at `<run-id>/<bead-id>.handoff.md`. It names any close-specific caveat (a regen step to run first, a fixture the bead regenerated). Honor the caveat before you stage. It names no paths: everything in the bead's worktree is the bead's work, so the close stages all of it.
   - If the handoff file is missing (subagent crashed before writing it), the bead's worktree is the answer: everything dirty in `git -C <worktree> status --porcelain` belongs to this bead, because nothing else was ever implemented there. Add a `retro:` comment on the bead noting that the handoff was never written.
   - Run `../scripts/bb-close-bead.sh --trailers-file "$RUN_DIR/trailers.txt" <id> <priority> "<title>"` from inside the bead's worktree, passing no paths — the same script the subagent would have run, with the same per-run trailers file Phase 0 resolved (omit `--trailers-file` WHEN the run resolved none). It closes, verifies the close against the database, stages **code only** (`git add -A -- ':!.beads/*.jsonl'`), commits, and finally runs `bd dolt push` — the push is what makes the closure durable. It appends the trailers to the commit body itself and leaves the bead's description exactly as grooming wrote it. Do not `git commit --amend` afterward to add trailers — that is the exact drift this flag exists to avoid.
   - **If the script cannot run**, fix why it cannot run and re-run it. Do not replay its steps by hand: a hand-rolled staging pathspec that excludes the whole `.beads/` directory silently drops tracked tracker config (`config.yaml`, `.gitignore`, the bd hooks) from the commit while still reporting success — the script's own pathspec (`:!.beads/*.jsonl`) is the one correct exclusion, and reproducing it by hand is exactly the kind of drift this ceremony exists to avoid.
   - The canonical message format is `<id> P<priority> <title>` — the script builds it; a by-hand replay must match it exactly.
   - Push the branch (if appropriate for the bucket) — a `git push` of the bead's branch, separate from the `bd dolt push` that lands the tracker.
   - Verify the bead is now `closed` via `bd show <id>`.
   - Record the commit yourself: `../scripts/bb-progress.sh "$RUN_DIR" <id> committed sha=$(git rev-parse --short HEAD)` — read the sha from the branch, not from the script's printout, same rule as the per-bead template's step 8.
3. **If the last phase was `implementing` or earlier** — the work is partial and uncommitted, and it is sitting safely in the bead's own worktree. **Do not stash and do not commit it.** Leave the worktree exactly as it is, mark the bead `blocked` with a recovery note naming the worktree path, and surface it to the user. The worktree's continued existence is the durable record that this bead is unfinished; stashing would move the work into the shared checkout and destroy that signal, which is the opposite of recovery.

   To resume rather than surrender it, re-enter the worktree and finish the work — there is no WIP commit to hunt for and nothing to squash (`operating-principles.md`, "No exceptions: the worktree holds interrupted work").
4. **If the progress file is empty or missing** — the subagent never started. Check `bd show <id>` to confirm status; if it's still `open`, retry the subagent. If it's `in_progress`, claim it back with `bd update <id> --status open` first. Remove any worktree it created before re-dispatching, so the in-flight signal does not lie.

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

**The ledger is a hint, never the authority.** It is written by the agent
doing the work, so it stops the instant that agent stops. Therefore:

- A trail ending before `committed` does **not** mean the bead is unfinished —
  only that nothing was appended after that point. A trail stopping at
  `implementing` routinely belongs to a bead that commits hours later, several
  beads on; read as a heartbeat it proves nothing but its own last write.
- An absent progress file does **not** mean no agent ran.
- A file's mtime bounds the last *write*, not the agent's liveness.

Resolve any status question in this order, stopping at the first answer:
`git worktree list` (present means unfinished — the one signal that is state
rather than testimony), then `bd show <id>`, then
`git log --oneline --grep=<id>`, then the ledger. **Name which one you used.** Never report a ledger fact you have not just read from
disk, and never from a backup copy while the live directory exists.

Events are written by `../scripts/bb-progress.sh`, never by hand:

```bash
../scripts/bb-progress.sh "$RUN_DIR" <bead-id> <phase> [key=value ...]
```

It validates the phase against the closed vocabulary above, escapes values, and appends one well-formed line. **Never build the event JSON in the caller** — a runtime-substituted `printf` operand trips the permission system's injection heuristic and prompts the user on *every* event write.

Batch subagents (fast lane, shared-surface batch) write events for every bead they process to the same per-bead progress files (one file per bead-id, even from the same batch subagent).

**Monitor any subagent past 5 minutes.** If a subagent is expected to (or actually does) run >5 min — true for any non-trivial bead given full test runs — the orchestrator must arm a `Monitor` on its progress JSONL so the user sees per-phase events live. Don't leave the user staring at silence for 30+ minutes. Pattern:

```
tail -n +1 -F <run-dir>/<bead-id>.progress.jsonl | grep -E --line-buffered '"phase":"(committed|blocked)"'
```

For batch subagents, tail multiple files in one Monitor. **Filter to `committed|blocked` only** — those are the events the orchestrator acts on. Relaying every phase (implementing, gates-running, …) wakes the orchestrator per event and adds a conversation turn each time for narration nobody needs; the per-bead progress files still hold the full phase history for crash recovery.

### Back the Monitor with a durable trigger

A Monitor is in-session: it dies with the turn that armed it, so it cannot
report its own death. A user interrupt that kills the orchestrator kills the
Monitor at the same instant — the "Nothing reports this to you" case above.
A durable run needs a second mechanism that survives that.

Before the first dispatch of a run, arm an out-of-session check-in **in
addition to** the Monitor:

- **Prefer a recurring trigger over a one-shot** — a cron/recurring trigger
  where the harness offers one, otherwise a re-armed `send_later`, at
  15–30 minute intervals. A recurring trigger re-arms itself instead of
  depending on the agent to remember, and the agent forgetting is the
  failure mode being defended against. State once at run start which kind
  the harness offers; WHERE it offers neither, say so once and proceed —
  the only wake path for this run is in-session.
- **The trigger's message must be self-contained.** It wakes a turn that may
  hold no run context: carry the `RUN_DIR` path, the epic id, and an
  instruction to run this file's orchestrator-recovery procedure — not
  "check on the run".
- **Do not key its liveness check on `bd list --status in_progress`.** A
  bead can be worked without ever being claimed, so that query reads
  "nothing in flight" while an agent is actively editing. Use
  `git worktree prune && git worktree list` instead — a worktree that still
  exists is a bead that is not finished.
- **Cancel it in Phase 4**, when the run drains. A trigger left armed after
  the run finishes wakes an agent into a finished sprint, which trains the
  reader to ignore it.

## Crash-resilient close: write the handoff file before bd update

After `gates-passed` and BEFORE the `bd update --status closed` call, the subagent **must** write `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/<bead-id>.handoff.md` — any close-specific caveat, and nothing else; no path list, because the close stages the bead's whole worktree. Then emit `{"phase":"retro-written"}` (the phase keeps that name). Only then proceed to the close ceremony.

This is the recovery anchor. If the subagent times out, crashes, or otherwise dies between `gates-passed` and the commit, the orchestrator can:
1. Read the handoff file for the caveat.
2. Run `../scripts/bb-close-bead.sh --trailers-file "$RUN_DIR/trailers.txt" <id> <priority> "<title>"` from the bead's worktree itself (omit `--trailers-file` WHEN the run resolved none) — no amend needed afterward to add trailers.
3. Record the commit: `../scripts/bb-progress.sh "$RUN_DIR" <id> committed sha=$(git rev-parse --short HEAD)` — read the sha from the branch, not from the script's printout.
4. Verify with `bd show <id>`, then push if the bucket calls for it.

The file deliberately does not restate the phase — `<bead-id>.progress.jsonl` is the single source for that, and a second copy would give a recovering orchestrator two hints to reconcile. It also carries no findings: those are `retro:` comments on the bead, durable in the tracker the moment they are written. **Writing it is non-negotiable.**

