---
name: sprint-recovery
description: Reconstruct the state of an in-flight sprint from durable sources
  when no session holds it — after the coordinator was interrupted, crashed,
  compacted past usefulness, or was replaced by a fresh session. Use when a
  durable check-in trigger fires carrying a run epic id, when asked to resume,
  pick back up, or recover a sprint or bead run, or whenever a session must
  re-derive which beads closed and which are still in flight from git and the
  tracker rather than from a summary. Requires no prior context and takes no
  required input.
version: "0.4"
user-invocable: false
---

# sprint-recovery

Pick a sprint back up when nobody is holding it any more.

**This skill assumes you know nothing about the run.** Three durable sources
hold all of it — the sprint **branch**, the run **epic**, and `git worktree
list` — and between them there is nothing a summary, a notification or an
earlier report could add. Read them; do not reconstruct from testimony.

**Nothing reports this to you.** A user interrupt that kills the coordinator
kills the mechanism that would have announced its children's deaths — kill
notifications sit enqueued and undelivered while the UI still shows the task
running. Assume no notification is coming and no notification you *do* find is
proof.

## Interface

**In: nothing is required.** That is the point — it is invoked by name from a
session that may hold no context at all. WHEN the caller supplies a run epic id
or a branch name, take them as a starting point and still confirm each against
the sources below.

**Out: the reconstructed run state, with a source named for every fact** — the
run's branch and id; the **completed set** from `git log`; the **in-flight
set** from `git worktree list`; a per-bead classification with the signal that
settled it; the **excluded set**, so a narrowed scope is not read as a finished
one; and anything you could not establish, named as unestablished rather than
omitted.

A fact with no source named is not part of the output.

**Your actor is the coordinator's**, and you cannot export it until step 2
gives you the run id: `export BEADS_ACTOR=coordinator-<run-id>`, the epic id
with the project prefix dropped. That is the same string the interrupted
session used, which is the whole reason ids are derived rather than minted —
it is what lets you find the authorization gate you already hold. The roster
and the derivation rule are the coordinator's dispatch policy's, their only
copy.

## Work these steps in order

1. **Vital signs first — `bb-status.sh branch`.** It needs no epic and no
   context, so it runs before you know anything: tree cleanliness, and whether
   the branch has **landed**. Its exit code is the first thing worth reporting;
   `2` means the run is stopped, not merely paused.

   **This is the step that catches the invisible failure.** A coordinator that
   died between committing and pushing leaves every other signal reading
   "finished" — no worktrees, clean tree, beads closed. Only the comparison
   against the remote separates a landed branch from an unlanded one. WHEN
   commits are unpushed, push them before concluding anything, and record why
   they were not: a rejection has a cause that will recur, and a credential
   refused for `.github/workflows/**` rejects the WHOLE push.

2. **Find the run — derive it from the branch.** A run happens on its own
   dedicated sprint branch, so the branch identifies the run and `git log` on
   it lists what landed:

   ```sh
   git rev-parse --abbrev-ref HEAD
   git log --oneline <default-branch>..HEAD     # every close commit is `<id> P<n> <title>`
   ```

   That is the completed set, read from the record rather than from a report.
   `bd show <id>` on any of them gives the parent, hence the run epic.

   **Never create a sprint branch here.** The manifest's `branch` and
   `base_sha` are what the original run wrote; a resumed run checks that branch
   out and adds to it. Creating one forks the run's history from the work it is
   meant to continue.

3. **Read the run epic in full** — `bd show <epic> --json`. Its metadata is the
   manifest: standing authorization, declared check ids, the checks baseline,
   the model lineup, the environmental ledger, the **admission record**
   (`admitted`, `excluded`, `moved`) and the review counter (`review_passes`,
   `review_budget`). Its `amend:` comments are how the run diverged.

   **Carry `review_passes` forward — do not restart it.** It is durable for
   exactly this moment: a resumed run that reset the counter re-enters the
   review loop from zero and never exhausts its budget.

   **The queue is not stored — re-derive it**: `bd ready --parent <epic>
   --exclude-type=epic`. Membership of the epic IS the scope, so a bead added
   since the run started is in scope now, and nothing wrote a dispatch order
   down because it drifts the moment the graph moves. **Compare the live queue
   against `manifest.admitted` and report the delta before dispatching
   anything** — a bead present now and absent then was added mid-run; one
   absent now and present then closed or was moved out. Both are facts to state
   rather than absorb.

4. **Run the scripts; do not derive by hand.**

   ```sh
   git worktree prune                      # stale registrations read as in-flight forever
   "$SCRIPTS/bb-status.sh" run <epic>      # scope, counts, derived health, readiness, DoD
   ```

   `git worktree prune` first is not optional: a reclaimed container leaves
   directories deleted but still registered, and a stale entry reads as
   in-flight forever.

5. **Classify what the scripts flagged — this is the judgment, and it is what
   remains yours.** Each surviving worktree's checked-out branch is **named for
   its dispatch id**, so the branch names the agent, and the agent's queue
   (`bd ready --assignee <id>` plus `bd list --assignee <id> --status
   in_progress`, both empty means DRAINED) says whether it is still working.
   Both halves are required: `bd ready` excludes `in_progress`, so the ready
   query alone reports a live agent as finished.

   | bead status | its dispatch's worktree | what it means | what to do |
   |---|---|---|---|
   | `in_progress` | present | the agent died holding it | re-run the project's checks IN that worktree |
   | `open` | present | started and abandoned, or never claimed | read the tree: dirty means partial work, clean means unstarted |
   | `in_progress` | absent | died before or without a worktree | confirm with `bd show`, re-dispatch |
   | `open` | absent | never started | re-dispatch under its existing `assignee` |
   | `blocked` | either | it stopped on purpose | read its notes for the blocker; do not re-dispatch blind |

   **Checks clean in a standing worktree means only the close is outstanding.**
   Read the bead's comments for a `caveat:` line, honour it, then run
   `$SCRIPTS/bb-close-bead.sh <id> <priority> "<title>"` from inside that
   worktree, passing no paths. **Green checks are not enough on their own** —
   they say the repository is not broken, not that this bead is done. Dispatch
   an agent to read the bead's acceptance criteria against the worktree's diff
   and return `complete` or `incomplete`; that is a read of source, so it is
   not yours. Either failing means the work is partial: leave the tree exactly
   as it is, mark the bead `blocked` with a note naming the worktree, and
   surface it. **Do not stash and do not commit it** — the worktree IS the
   durable record.

   **Do not infer liveness from file mtimes**, and do not read
   `bd list --status in_progress` as a liveness signal: a bead can be worked
   without being claimed.

6. **Resume, and re-entering is cheap by design.** `sprint-plan` grooms only
   what is not already Ready and re-derives the partition from the **current**
   ready front, and closed beads are not in that front — so re-running the
   process against the same epic is a *narrower repeat*, not a duplicate run.
   Scope resolution has nothing to resolve, because the epic already exists.

**Verify anything surprising yourself.** A completion notification is not
proof, and neither is a summary or an earlier report. The scripts name the
command behind every line they print — re-run the one you doubt rather than
trusting or discarding the whole report.
