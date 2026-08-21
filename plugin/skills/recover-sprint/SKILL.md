---
name: recover-sprint
description: Reconstruct the state of an in-flight bead-bonanza sprint from
  durable sources when no session holds it — after the orchestrator was
  interrupted, crashed, compacted past usefulness, or was replaced by a fresh
  session. Use when a durable check-in trigger fires carrying a run epic id,
  when asked to resume, pick back up, or recover a bead-bonanza run or sprint,
  or whenever a session must re-derive which beads closed and which are still
  in flight from git and the tracker rather than from a summary. Requires no
  prior context and takes no required input.
version: "0.2"
user-invocable: false
---

# recover-sprint

Pick a bead-bonanza sprint back up when nobody is holding it any more.

**This skill assumes you know nothing about the run.** It derives the run, its
plan, its completed set and its in-flight set from durable sources — the git
branch, the run epic's manifest, `git worktree list`, and the bead tracker —
and never from a summary, a notification, or an earlier report.

**Nothing reports this to you.** A user interrupt that kills the orchestrator
kills the mechanism that would have announced its children's deaths — kill
notifications sit enqueued and undelivered while the UI still shows the task
running. Assume no notification is coming and no notification you *do* find is
proof.

## Interface

**In: nothing is required.** That is the point of this skill — it is invoked by
name from a session that may hold no context at all, so everything it needs is
derivable. WHEN the caller does supply a run epic id or a branch name, take
them as a starting point and still confirm each against the sources below.

**Out: the reconstructed run state, with a source named for every fact** —

- the run's branch and run id, and where you read them;
- the **completed set** — beads whose close commit is on the branch, from
  `git log`;
- the **in-flight set** — beads whose worktree still exists, from
  `git worktree list`;
- a per-bead classification for every bead in the run epic, with the signal
  that settled it;
- the **excluded set** — beads the manifest's admission record says sat this
  run out, so a narrowed scope is not silently read as a finished one;
- anything you could not establish, named as unestablished rather than
  omitted.

A fact with no source named is not part of the output. A reconstruction that
could not read the manifest is a *reconstruction*, and says so.

## Work these steps in order

0. **Take the run's vital signs first.** Before deriving anything, run
   bead-bonanza's `scripts/bb-status.sh wave` from the sprint branch, resolved
   from that skill's own directory. One command answers whether the tree is
   clean, whether the branch has landed, what is still in flight, and who holds
   the merge slot — and prints the command behind each, so any line can be
   re-derived. It repairs nothing and is safe to run before you know anything
   about the run. Its exit code is the first thing worth reporting: `2` means a
   wave is stopped, not merely paused.

1. **Find the run — derive it from the branch.** A run happens on its own
   dedicated branch, uses one run directory, and that directory is removed
   before the branch merges. So the branch you are on identifies the run, and
   `git log` on it lists the beads that have landed.
   Start there: `git rev-parse --abbrev-ref HEAD`, then
   `git log --oneline <default-branch>..HEAD` — every close commit is
   `<bead-id> P<n> <title>`, so that is the run's completed set, read from the
   record rather than from a report.

   To find the run epic when the caller gave none:

   - `bd swarm list` names the epic behind each registered swarm — but check
     each molecule's own status, because it lists closed ones under "Active".
   - Otherwise the sprint branch's `git log`: every bead's close commit names
     its id, and `bd show <id>` gives the parent.

   **There is no run directory to find, and that is the point.** The run's
   record is its epic, which survives a container reclaim.

2. **Read the run epic in full** — `bd show <epic> --json`. Its `--metadata` is
   the manifest: the standing authorization, the declared check ids, the checks
   baseline, the model lineup, the environmental ledger, and the **admission
   record** (`admitted`, `excluded`, `moved`). Its `amend:`-prefixed comments
   are how the run diverged since it started.

   **The queue is not stored — re-derive it**: `bd ready --parent <epic>
   --exclude-type=epic`. That is deliberate. Membership of the epic IS the
   scope, so a bead added since the run started is in scope now, and the
   wave table was never written down because it drifts the moment the graph
   moves.

   **Compare the live queue against the manifest's `admitted` list and report
   the delta before dispatching anything.** A bead present now and absent then
   was added mid-run; one absent now and present then closed, or was moved out.
   Both are facts the resuming session must state rather than absorb.

3. **Establish what is in flight — `git worktree list` answers "is an agent
   still working", and nothing else.**
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

3b. **Establish what has LANDED — compare the branch against its remote.**

   ```sh
   git fetch "$REMOTE" "$BRANCH"
   git log --oneline "$REMOTE/$BRANCH..HEAD"    # must be empty
   ```

   An orchestrator that dies between committing and pushing, or whose push is
   **rejected**, leaves a state in which every other signal reads "finished":
   no worktrees (they are removed before the push), a clean tree, and the beads
   closed in the tracker. Only this comparison distinguishes a landed wave from
   an unlanded one.

   WHEN commits are unpushed, push them before concluding anything about the
   run, and record why they were not pushed — a rejection has a cause that
   will recur on the next wave. A credential refused for
   `.github/workflows/**` rejects the WHOLE push, so a single workflow file
   strands every commit behind it. `bead-bonanza` Phase 3 step 7 runs the
   same check from the other side; keep the two in the same form.

4. **Classify each in-scope bead from state, in this order.** Resolve every
   status question by reading state, stopping at the first answer, and **name
   which source you used**: `git worktree list` (present means unfinished),
   then `bd show <id>`, then `git log --oneline --grep=<id>`.

   - **Worktree present, bead not closed** — the bead is unfinished. Whether
     the work in it is complete is not knowable from any record, so **re-run
     the project's gate suite in the worktree** rather than inferring. Gates
     clean means only the close is outstanding: read the bead's comments for a
     `caveat:` line, honor it, then run bead-bonanza's close ceremony —
     `scripts/bb-close-bead.sh <id> <priority> "<title>"`, resolved from that
     skill's own directory — from inside the bead's worktree, passing no
     paths. A caveat comment says only what the close must do differently — it
     is not a body for the close, and the close script leaves the bead's
     description as grooming wrote it. Gates failing means the work is partial: leave it in place and
     surface it.
   - **Worktree present, work partial or absent** — it may have died mid-work,
     or be finished and committed. Check `git log --grep=<id>` before you touch
     the tree.
   - **No worktree, bead closed** — verify a commit for it is an ancestor of
     `HEAD` (`git merge-base --is-ancestor <sha> HEAD`) before believing the
     close landed on the sprint branch.
   - **No worktree, bead open** — it may never have started. Confirm with
     `bd show <id>`.

   Whether a bead's own branch was pushed says nothing about whether the sprint
   branch landed; in a concurrent wave that push is the orchestrator's job.
   Verify with step 3b.

5. **Verify every claim against `bd` and `git` yourself.** A completion
   notification is not proof — notifications go missing entirely. Neither is a
   summary, a backup copy of the run directory, or your own earlier report. Read
   the live state, name the source you read, then act.

6. **Report the reconstructed state** in the shape the Interface above
   specifies, then resume the run under bead-bonanza from the point the plan
   and the worktree list put you at. WHEN a bead's worktree holds partial,
   uncommitted work, leave it exactly where it is — do not stash it and do not
   commit it — and either finish it in place or surface it.
