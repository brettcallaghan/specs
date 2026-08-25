---
name: sprint-batch
description: Take one batch of beads from assignment to handed-back branch —
  the developer tier of a sprint. One subagent, one batch, one worktree, one
  commit per bead. Use when dispatched with several beads to work in order,
  when asked to run a batch of beads over a shared surface, or when a
  batch must account for every bead it was given. Covers accepting the
  assignment (base ref, clean tree, manifest), running run-bead per bead, and
  the return line the orchestrator reads. It merges nothing and pushes only
  where the batch's push mode says to.
version: "0.1"
user-invocable: false
---

# sprint-batch

You are one developer holding one batch. **Work the beads in order, one commit
each, and account for every one of them** — a batch that returns fewer lines
than it was given beads is the failure this tier exists to prevent.

## Inputs

| | |
|---|---|
| the bead list | ids, priorities, titles, and each bead's decision — **in order** |
| `{WORKTREE}` | the ONE worktree for this batch — you create no others |
| `{EPIC}` | the run epic — the manifest, and where incidents land |
| `{DISPATCH_ID}` | yours, shared by every bead in this batch |
| `{BRANCH}` | the sprint branch — your worktree descends from it |
| `{SCRIPTS}` | the orchestrator's `scripts/` directory |
| `{CHECKS_IDS}`, `{CHECKS_BASELINE_TEXT}` | what the project declared, and what it printed before the run |
| push mode | `push_per_bead` or `defer_push` |

Infer none of it. A missing input is reported, never guessed.

## D1 · Open — accept the assignment

Four checks, before you touch a bead. Each is cheap and each has caught a run.

Before any bd call, in every shell you run `bd` from, export your dispatch id:

```sh
export BEADS_ACTOR={DISPATCH_ID}
```

A shell that loses it claims as the harness default, and an unassigned bead
claimed that way is reassigned with no error anywhere.

1. **The worktree is clean** — `git status --porcelain` in it, empty.
2. **The base ref is the sprint branch, not the default branch.** Verify it:

   ```sh
   git -C "{WORKTREE}" merge-base --is-ancestor "$(git rev-parse {BRANCH})" HEAD
   ```

   **This is the one that silently destroys work.** A worktree branched from
   the default branch cannot see anything the sprint branch already carries, so
   every bead in this batch would be written against a tree missing its
   predecessors' commits — and the merge afterwards looks like a conflict
   rather than like a mistake. Report and **do not start**.
3. **Read the manifest** — `bd show {EPIC} --json | jq -r '.[0].metadata'`. The
   checks baseline is in it, and it is what "no findings beyond the baseline"
   means for every bead you are about to run.
4. **Confirm the batch is yours** — `bd ready --assignee {DISPATCH_ID}` plus
   `bd list --assignee {DISPATCH_ID} --status blocked`. Between them they are
   your queue, and their union must be the beads you were handed.

   **Assignment is the only record, so there is nothing to write here.** A bead
   in your list that is not assigned to you is a dispatch error: report it and
   do not start, rather than claiming your way into someone else's queue. A
   bead assigned to you that is missing from your list is blocked behind
   something — say which, from `bd show <id>`.

## D2 · Work — `run-bead`, per bead, in order

**Invoke the `run-bead` skill for each bead.** It carries the ceremony, the
gate step, the findings capture, the incident rule, and the blocked and skipped
paths. Do not restate it and do not shortcut it.

Two things batching does **not** change:

- **Every bead runs the full declared check set before its own close.** Being
  in a batch is not a reason to run less — an unvalidated commit is not
  cheaper, it is just unvalidated.
- **Every bead is its own commit.** Never combine two beads into one.

**One worktree, one branch per bead, cut in order.** In `{WORKTREE}`, before
each bead:

```sh
git checkout -b <bead-suffix>
```

That branches from where the previous bead committed, which is what makes each
bead build on the last. **Do not create a second worktree** — the batch is one
agent working serially, so a second tree isolates it from nobody, and it is the
*branches* that keep the beads separate. **Do not branch from the sprint branch
for each bead**: the orchestrator merges only after you return, so the sprint
head does not move while you work, and beads cut from it cannot see each
other's work and collide on any shared file.

**Fix a finding in the bead that introduced it.** Because every bead runs the
checks, the bead whose run first shows a line absent from the baseline is where
it entered — fix it there, before starting the next, rather than carrying it
forward. WHERE investigating shows an *earlier* bead is at fault, take the
Blocked path for that bead rather than silently repairing it across several
commits.

**A bead that conflicts with an earlier bead in this batch** — one removed what
the next was about to change — takes the Blocked path.

**Whether you continue after a block depends on the tree, not the reason.**
`run-bead`'s Blocked path leaves the partial work uncommitted, and you have one
worktree, so that work is sitting where the next bead would start. Run
`{SCRIPTS}/bb-verify-skip.sh` in `{WORKTREE}`:

- **CLEAN (exit 0)** — nothing was started. Continue with the next bead.
- **DIRTY (exit 1)** — the batch ends here. Return the blocked bead with its
  reason and every remaining bead as `blocked — not reached (batch ended at
  <id>)`. Leave the tree exactly as it is: it is the durable record of the
  partial work, and committing or stashing it to free the tree destroys the
  thing the block exists to preserve.

## D3 · Close — hand off

**Every bead ends in a tracker status, and that status is the hand-off.**
`run-bead` closes a completed bead and sets a blocked one to `blocked`; a
skipped bead is closed with its rationale in the notes. Leave none of them
`in_progress` — an agent that walks away from a claimed bead is
indistinguishable from one still working, which is the single reading the
orchestrator cannot recover from.

Then return **one message**: one line per bead, then

```
batch-end: <n> completed, <n> blocked, <n> skipped @ <branch>
```

**Every bead you were given appears in exactly one of those counts.** A bead
you never reached is `blocked` with the reason, not omitted.

**You hand the branch back; you do not merge it.** The orchestrator merges, one
at a time, which is what keeps history linear. In `defer_push` mode you do not
push either.

## How the orchestrator reads you

Not from your return line — from your queue. **You have DRAINED when
`bd ready --assignee {DISPATCH_ID}` is empty and you hold nothing
`in_progress`.** Both halves are required: `bd ready` excludes `in_progress`,
so a bead you are working right now empties the ready front, and the ready
query alone would report you finished mid-bead.

| your queue | your worktree | reading |
|---|---|---|
| not drained | present | live worker |
| drained | present | finished; orphaned tree, the orchestrator removes it |
| not drained | absent | the worker died |
| drained | absent | done and cleaned up |

Nothing in that is yours to write, which is why nothing in it can go stale.
**Your obligation is to leave every bead in a truthful status** — claimed while
you work it, closed or `blocked` when you stop. A worktree is matched to you
through its checked-out branch, which is a bead's hash suffix, so the branch
names the bead and the bead names you.
