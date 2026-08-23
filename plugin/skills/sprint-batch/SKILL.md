---
name: sprint-batch
description: Take one batch of beads from assignment to handed-back branch —
  the developer tier of a sprint. One subagent, one batch, one worktree, one
  commit per bead. Use when dispatched with several beads to work in order,
  when asked to run a doc/chore fast lane or a shared-surface batch, or when a
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
| `{WORKTREE}` | the worktree you were given, or the instruction to create one per bead |
| `{EPIC}` | the run epic — the manifest, and where incidents land |
| `{DISPATCH_ID}` | yours, shared by every bead in this batch |
| `{SCRIPTS}` | the orchestrator's `scripts/` directory |
| `{CHECKS_IDS}`, `{CHECKS_BASELINE_TEXT}` | what the project declared, and what it printed before the run |
| push mode | `push_per_bead` or `defer_push` |

Infer none of it. A missing input is reported, never guessed.

## D1 · Open — accept the assignment

Four checks, before you touch a bead. Each is cheap and each has caught a run.

1. **The worktree is clean** — `git status --porcelain` in it, empty.
2. **The base ref is the sprint branch, not the default branch.** Verify it:

   ```sh
   git -C "{WORKTREE}" merge-base --is-ancestor "$(git rev-parse <sprint-branch>)" HEAD
   ```

   **This is the one that silently destroys work.** A worktree branched from
   the default branch cannot see anything the sprint branch already carries, so
   every bead in this batch would be written against a tree missing its
   predecessors' commits — and the merge afterwards looks like a conflict
   rather than like a mistake. Report and **do not start**.
3. **Read the manifest** — `bd show {EPIC} --json | jq -r '.[0].metadata'`. The
   checks baseline is in it, and it is what "no findings beyond the baseline"
   means for every bead you are about to run.
4. **Record that you hold them.** For each bead in the batch:

   ```sh
   bd update <id> --set-metadata dispatch_id={DISPATCH_ID} \
     --set-metadata dispatch_state=running \
     --set-metadata worktree={WORKTREE} \
     --set-metadata dispatched_at="$(date -u +%FT%TZ)"
   ```

   One `dispatch_id` for the whole batch. That is what makes "which agent is
   holding this bead" a query rather than an inference.

## D2 · Work — `run-bead`, per bead, in order

**Invoke the `run-bead` skill for each bead.** It carries the ceremony, the
gate step, the findings capture, the incident rule, and the blocked and skipped
paths. Do not restate it and do not shortcut it.

Two things batching does **not** change:

- **Every bead runs the full declared check set before its own close.** Being
  in a batch is not a reason to run less — an unvalidated commit is not
  cheaper, it is just unvalidated.
- **Every bead is its own commit.** Never combine two beads into one.

**Order is load-bearing where you create your own worktrees.** Branch each bead
from the **previous bead's branch**, not from HEAD: inside a batch HEAD does
not move, because the orchestrator merges only after you return, so branching
from HEAD gives beads that cannot see each other's work and collide on any
shared file.

**Fix a finding in the bead that introduced it.** Because every bead runs the
checks, the bead whose run first shows a line absent from the baseline is where
it entered — fix it there, before starting the next, rather than carrying it
forward. WHERE investigating shows an *earlier* bead is at fault, take the
Blocked path for that bead rather than silently repairing it across several
commits.

**A bead that conflicts with an earlier bead in this batch** — one removed what
the next was about to change — takes the Blocked path, and you continue with
the rest.

## D3 · Close — hand off

Your last act per bead, whatever the outcome:

```sh
bd update <id> --set-metadata dispatch_state=returned \
  --set-metadata terminal_state=completed|blocked|skipped
```

**Write it even when the bead is blocked — especially then.** Together with
whether the worktree still exists, that pair is what tells the orchestrator a
live agent from a dead one without guessing.

Then return **one message**: one line per bead, then

```
batch-end: <n> completed, <n> blocked, <n> skipped @ <branch>
```

**Every bead you were given appears in exactly one of those counts.** A bead
you never reached is `blocked` with the reason, not omitted.

**You hand the branch back; you do not merge it.** The orchestrator merges, one
at a time, which is what keeps history linear. In `defer_push` mode you do not
push either.

## What the four states mean

The orchestrator reads your dispatch record against the worktree:

| worktree | `dispatch_state` | reading |
|---|---|---|
| present | `running` | live worker |
| present | `returned` | finished; orphaned tree, the orchestrator removes it |
| absent | `running` | the worker died |
| absent | `returned` | done and cleaned up |

Writing both ends of that record is the whole of your obligation to it. A
standing worktree with no record is three paragraphs of prose to interpret; one
with a record is a table lookup.
