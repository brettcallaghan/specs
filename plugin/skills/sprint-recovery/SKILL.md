---
name: sprint-recovery
description: Reconstruct the state of an in-flight sprint from durable sources
  when no session holds it — after the orchestrator was interrupted, crashed,
  compacted past usefulness, or was replaced by a fresh session. Use when a
  durable check-in trigger fires carrying a run epic id, when asked to resume,
  pick back up, or recover a sprint or bead run, or whenever a session must
  re-derive which beads closed and which are still in flight from git and the
  tracker rather than from a summary. Requires no prior context and takes no
  required input.
version: "0.3"
user-invocable: false
---

# sprint-recovery

Pick a sprint back up when nobody is holding it any more.

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

0. **Take the run's vital signs first — `bb-status.sh branch`.** From the sprint
   branch, resolved from the sprint orchestrator's own `scripts/`. It needs no epic and no
   context, so it runs before you know anything: it answers whether the tree is
   clean and whether the branch has **landed**, and prints the command behind
   each line so any of them can be re-derived. It repairs nothing.

   Its exit code is the first thing worth reporting: `2` means the run is
   stopped, not merely paused.

1. **Find the run — derive it from the branch.** A run happens on its own
   dedicated **sprint branch**, created once before its first worktree. So the
   branch you are on identifies the run, and `git log` on it lists the beads
   that have landed.
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

   **Never create a sprint branch here.** The manifest's `branch` and
   `base_sha` are what the original run wrote; a resumed run checks that branch
   out and adds to it. Creating one would fork the run's history from the work
   it is meant to continue.

2. **Read the run epic in full** — `bd show <epic> --json`. Its `--metadata` is
   the manifest: the standing authorization, the declared check ids, the checks
   baseline, the model lineup, the environmental ledger, the **admission
   record** (`admitted`, `excluded`, `moved`), and the review counter
   (`review_passes`, `review_budget`). Its `amend:`-prefixed comments are how
   the run diverged since it started.

   **Carry `review_passes` forward — do not restart it.** It is durable for
   exactly this moment: a resumed run that reset the counter re-enters the
   review loop from zero and never exhausts its budget. Hand it back to the
   orchestrator with the rest, and say what it read.

   **The queue is not stored — re-derive it**: `bd ready --parent <epic>
   --exclude-type=epic`. That is deliberate. Membership of the epic IS the
   scope, so a bead added since the run started is in scope now. Nothing wrote
   a dispatch order down, because it drifts the moment the graph moves.

   **Re-entering costs almost nothing, and that is by design.** Grooming and
   planning are one step (`sprint-plan`) that grooms only what is not already
   Ready and re-derives the partition from the **current** ready front. Closed
   beads are not in that front, so they are not re-dispatched. Scope resolution
   has nothing to resolve, because the epic already exists. Re-running the
   whole process against the same epic id is therefore a *narrower repeat*, not
   a duplicate run.

   **Compare the live queue against the manifest's `admitted` list and report
   the delta before dispatching anything.** A bead present now and absent then
   was added mid-run; one absent now and present then closed, or was moved out.
   Both are facts the resuming session must state rather than absorb.

3. **Derive the rest by running the scripts, not by hand.** Now that you have
   the epic:

   ```sh
   git worktree prune                      # stale registrations read as in-flight forever
   "$SCRIPTS/bb-status.sh" run  <epic>     # scope, counts, health, readiness, DoD
   "$SCRIPTS/bb-status.sh" slot <epic>     # is a held merge slot safe to release
   ```

   `git worktree prune` first is not optional: a reclaimed container leaves
   directories deleted but still registered, and a stale entry reads as
   in-flight forever.

   **Step 0's `branch` is the one that catches the invisible failure.** An
   orchestrator that dies between committing and pushing leaves every other
   signal reading "finished" — no worktrees, clean tree, beads closed in the
   tracker. Only the comparison against the remote separates a landed branch
   from an unlanded one. WHEN commits are unpushed, push them before concluding
   anything, and record why they were not: a rejection has a cause that will
   recur. A credential refused for `.github/workflows/**` rejects the WHOLE
   push, stranding every commit behind one file.

4. **Classify what the scripts flagged — this is the judgment, and it is what
   remains yours.** For each bead not closed, read its status against the
   worktree whose branch names it. `git worktree list` prints each
   registration's checked-out branch; a bead's branch is its hash suffix, so
   the branch names the bead and the bead's `assignee` names the agent.

   | bead status | its branch's worktree | what it means | what to do |
   |---|---|---|---|
   | `in_progress` | present | the agent died holding it | re-run the project's checks IN that worktree |
   | `open` | present | started and abandoned, or never claimed | read the tree: dirty means partial work, clean means unstarted |
   | `in_progress` | absent | died before or without a worktree | confirm with `bd show`, re-dispatch |
   | `open` | absent | never started | re-dispatch under its existing `assignee` |
   | `blocked` | either | it stopped on purpose | read its notes for the blocker; do not re-dispatch blind |

   That table replaces inference, and nothing had to be written at dispatch
   time for it to work. **An agent has drained when `bd ready --assignee <id>`
   is empty AND it holds nothing `in_progress`** — both, because `bd ready`
   excludes `in_progress`, so the ready query alone reports a live agent as
   finished. Where a worktree's branch matches no bead, say so and leave the
   tree standing; `git log --oneline --grep=<id>` is the last resort, and name
   which source answered.

   **Checks clean in a standing worktree means only the close is outstanding.**
   Read the bead's comments for a `caveat:` line, honour it, then run
   `$SCRIPTS/bb-close-bead.sh <id> <priority> "<title>"` from inside that
   worktree, passing no paths. A caveat says what the close must do
   differently; it is not a body for the close. Checks failing means the work
   is partial: leave it and surface it.

   **Do not infer liveness from file mtimes.** A dirty file's mtime bounds the
   last write, not the agent's liveness. And `bd list --status in_progress` is
   not a liveness signal either — a bead can be worked without being claimed.

5. **Verify anything surprising yourself.** A completion notification is not
   proof; notifications go missing. Neither is a summary or your own earlier
   report. The scripts name the command behind every line they print — re-run
   the one you doubt rather than trusting or discarding the whole report.

6. **Report the reconstructed state** in the shape the Interface above
   specifies, then resume the run under the sprint orchestrator from the point
   the plan and the worktree list put you at. WHEN a bead's worktree holds partial,
   uncommitted work, leave it exactly where it is — do not stash it and do not
   commit it — and either finish it in place or surface it.
