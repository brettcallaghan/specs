---
name: sprint-recovery
description: Reconstruct the state of an in-flight sprint from durable sources
  when no session holds it — after the coordinator was interrupted, crashed,
  compacted past usefulness, or was replaced by a fresh session. Invoked by
  `bead-bonanza` as its first act whenever it is given a run epic that already
  exists; a user resumes a run by re-invoking `bead-bonanza`, never this skill
  directly. Re-derives which beads closed and which are still in flight from
  git and the tracker rather than from a summary, and hands the state back
  without deciding what to do with it.
version: "0.5"
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

**In: nothing is required.** That is the point — the coordinator invoking you
may hold no context at all, having just been handed a run epic id by a durable
trigger. WHEN it supplies that id or a branch name, take them as a starting
point and still confirm each against the sources below.

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

1. **Vital signs first — `swarm branch verify-landed`.** It needs no run scope,
   so it runs before you know which run you are in. Route on its exit status
   and report it: it fetches the remote before comparing, so a stale local
   record of where the remote points cannot report a branch as landed, and a
   branch with no counterpart there is reported unlanded rather than clean.
   Add `git status --porcelain` beside it — the working tree is the one vital
   sign the command does not read.

   **This is the step that catches the invisible failure.** A coordinator that
   died between committing and pushing leaves every other signal reading
   "finished" — no worktrees, clean tree, beads closed. Only the comparison
   against the remote separates a landed branch from an unlanded one. WHEN
   commits are unpushed, `swarm branch push` pushes and re-proves them before
   you conclude anything, and record why they were not pushed: a rejection has
   a cause that will recur, and a credential refused for
   `.github/workflows/**` rejects the WHOLE push.

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

3. **Read the run epic in full** — `swarm manifest get --run <epic>` for the
   manifest, `bd show <epic>` for its `amend:` comments. Run both and read what
   they print: the manifest is what the original run recorded about itself, and
   the comments are how it diverged.

   **Then `swarm manifest validate --run <epic>`.** You are reconstructing a
   run no session held, so the record may have been left half-written. **That
   is a fact to hand back, not one to repair.**

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
   swarm status --run <epic>      # run it, read it, route on the status
   ```

   It prunes stale worktree registrations before it counts them, which is not
   optional: a reclaimed container leaves directories deleted but still
   registered, and a stale entry reads as in-flight forever.

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

6. **Hand the state back; do not act on it.** The coordinator re-enters at
   groom & plan, and re-entering is cheap by design: `sprint-plan` grooms only
   what is not already Ready and re-derives the partition from the **current**
   ready set, and closed beads are not in that ready set — so a resumed run is a
   *narrower repeat*, not a duplicate. Scope has nothing to resolve, because
   the epic already exists.

**Verify anything surprising yourself.** A completion notification is not
proof, and neither is a summary or an earlier report. The scripts name the
command behind every line they print — re-run the one you doubt rather than
trusting or discarding the whole report.
