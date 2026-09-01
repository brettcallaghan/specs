# Worktree mechanics (coordinator)

**Not a prompt.** This is the coordinator's own procedure: how a dispatch's
worktree is created, verified, merged and removed. No subagent reads it, and
none is told the path either — an agent resolves its own worktree from the tree
it is standing in (`dispatching.md`).

## Contents

- Opening a dispatch — the sprint branch and the worktree, in one command
- Beads inside a worktree — `bd` resolves the shared tracker from any of them
- Accepting a dispatch — what the command does, and the two things it leaves you
- A rebase conflict is a re-planning problem, and who resolves it
- Concurrent dispatch out of one ready set

## Opening a dispatch

**One command creates the sprint branch and the dispatch's worktree**, and the
base each is cut from is an argument rather than a setting:

```sh
swarm dispatch open --run <epic>                                  # the sprint branch, once
swarm dispatch open --run <epic> --dispatch <dispatch-id>         # + its worktree
```

The first invocation of a run records the branch and the commit it was cut from
in one manifest write; every dispatch after it reuses what is recorded and
creates none, which is what makes a resumed run land where the original did.
`--branch` and `--base` name the branch and its base where it does not exist
yet; `--dispatch` stands the worktree up on a branch named for the dispatch.
**The command owns the rest and this document does not restate it**: `--help`
names the flags, `--explain` prints every invocation without running one.
**Route on the exit status** — a branch already carrying this run's name with
the manifest recording none stops the run, before any worktree exists.

**One worktree and ONE BRANCH per DISPATCH.** A batch is one agent working its
beads serially, so a second worktree isolates it from nobody — and a branch per
bead buys nothing either, because **one commit per bead is what separates the
beads**, and the ceremony already guarantees it.

**Presence is the in-flight signal.** A worktree survives until its dispatch is
accepted, so a standing tree means "this dispatch is unfinished"; which of its
beads is unfinished comes from the beads' statuses, not from counting trees.

**Prune before trusting the list.** A reclaimed container deletes worktree
directories without deregistering them, and a stale entry reads as in-flight
forever: `git worktree prune` before `git worktree list` in pre-flight and in
recovery.

## Beads inside a worktree

`bd` is not an obstacle: it resolves the shared tracker database through
git's common directory from any worktree — `bd where`/`bd worktree info`
point at the main checkout, `bd list` returns the full set, and no
`embeddeddolt` or redirect file is created (with `BEADS_DIR` unset). All
worktree agents share one tracker, which is what makes this constraint real:

- **Never read `.beads/*.jsonl` as tracker state.** Ask the database: `bd show`,
  or `bd export` to stdout, which regenerates from it. There is no export step
  to run before committing — the close stages code only.

## Accepting a dispatch

**One command per dispatch, as that dispatch returns** rather than after all of
them return — a branch left unmerged is a base every later dispatch is missing:

```sh
swarm dispatch accept --run <epic> --dispatch <dispatch-id>
```

The drain check, the serialized/concurrent case, the merge, the rebase a
concurrent branch needs, the worktree removal behind its ancestry guard, the
per-bead re-check and the push are all inside it, in the order its guards
require. **The command owns them**: `--help` names the flags, `--explain`
prints every invocation without running one, `--no-push` stops in front of the
push. **Route on the exit status.**

**Two things it does not do.** The pre-merge acceptance check per bead is
still the coordinator's (`dispatching.md`), and a rebase conflict is recorded
and blocked but never resolved — see below.

WHEN a bead is blocked rather than finished, **leave its dispatch's worktree in
place** and say so in the report. It holds the partial work, and its existence
is the durable record that the bead is unfinished — there is nothing to stash
and no WIP commit to make (`operating-principles.md`).

**In a batch that means the dispatch ends there IF the block left work in the
tree.** One worktree cannot both preserve a blocked bead's uncommitted changes
and give the next bead a clean tree to start from, and preserving them wins —
the alternative is the next bead committing someone else's half-finished work.
A bead blocked with a **clean** tree (nothing started, or a prerequisite found
before any edit) costs nothing and the batch continues. The skip-verification
script, run in the worktree, is the test.

## A rebase conflict is a re-planning problem

**The collision signal is the rebase:** a conflict means two beads genuinely
changed the same lines, which is the claim the packing was built on failing.
The acceptance command records it as a `sequencing` finding naming both beads
and every conflicting path, and marks the unmerged bead blocked. **It resolves
nothing, and neither do you.**

**Dispatch ONE agent to resolve it, in that dispatch's worktree.** Resolving is
not yours: you hold no bead's context, you made neither change, and a
resolution you invent is an unreviewed edit landing under someone else's
commit. Hand the agent the conflicted worktree, both beads' ids, and the
instruction to resolve, run the declared checks, and hand the branch back —
nothing else. **One resolver at a time**, which holds for the same reason
accepting does: you dispatch it, and you are one agent.

**Stop DISPATCHING NEW WORK while a resolution is in flight, not the run.**
Every bead already merged stays merged and every branch already returned is
still acceptable; what is unsafe is sending further work onto a base whose
isolation assumption just failed. WHEN the resolver cannot separate the two
changes, leave the blocked bead's worktree standing and report — two beads that
genuinely collided belong in ONE dispatch, sequential, which is what a
collision edge between them expresses.

## Concurrent dispatch out of one ready set (enabled)

Spawn several dispatches at once — `swarm dispatch open --dispatch <id>` per
dispatch, then one `Agent` call each into the worktree it stood up, all in a
single message. What may go out together is whatever the ready set holds:
collision edges keep anything that would conflict out of it.

Opening and accepting are unchanged per dispatch. One addition when running
concurrently: where an acceptance hits a fixture/snapshot conflict and the
project has a regen step, take both sides, regenerate on the merged tree, stage
and amend. A non-fixture conflict on shared code is the re-planning problem above.
