---
title: "ADR-026: No Rollback — Validate, Order, Declare Idempotence"
description: A failed plan stops and reports where it stopped rather than
  compensating, because the writes that matter cannot be reversed and a
  partial undo is worse than none
slug: adr-026
type: adr
status: approved
updated_at: 2026-08-28T00:00:00Z
---

# ADR-026: No Rollback — Validate, Order, Declare Idempotence

**Status:** approved  
**Date:** 2026-08-28

## Context

A plan ([[adr-024]]) may write several times, and a step can fail after
an earlier write has landed. The obvious remedy is a compensating action
per step — a saga — so a failure unwinds what preceded it.

What the writes actually are decides whether that is possible. The
family's mutations are: keys into the run manifest, a status or assignee
on a bead, a branch created, a worktree created or removed, a merge into
the sprint branch, and a push. Of those, one is genuinely reversible in
isolation (a freshly created branch), several are reversible only if
nothing has observed them (a merge, before anything builds on it), and
two are not reversible at all — a push has left the machine, and a
removed worktree may have held uncommitted work that is now gone.

[[frd-swarm-core]] does ask for atomicity in one place:
AC-SWARM-CORE-002.14 requires a multi-key manifest write that fails on
one key to write none of them. That is a real obligation, and it is worth
noticing that it does not require undo.

## Decision

We will not implement a rollback or compensation mechanism. Instead a
plan achieves the same guarantees three ways: every check a step can make
runs before the first write, so a plan that will fail fails before it has
changed anything; irreversible writes are ordered last, after everything
that could still refuse; and each step declares whether re-running it is
safe, so recovery from a partial plan is re-running it rather than
undoing it. A plan that stops reports which steps ran and which did not.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| A compensating action per step (saga) | A failure unwinds cleanly where compensation is possible | Compensation is impossible for a push and unsafe for a worktree removal, so it would exist for some steps and not others — and a caller cannot tell which kind of failure it got without reading the implementation. The compensating writes themselves run at the moment the world is known not to be as expected | A guarantee that holds for some steps is not a guarantee |
| A transaction over the tracker | Familiar semantics | The tracker offers none, and git offers none across a merge and a push; a transaction would have to be simulated by the compensation above | The substrate cannot provide it |
| Snapshot and restore the repository and tracker | Complete, mechanical | Snapshotting a git repository plus a Dolt database per plan costs more than the operations being guarded, and restoring one discards concurrent work from other agents in the same run | Cost and blast radius both exceed the failure it guards |
| Validate first, order writes, declare idempotence | Every guarantee it offers actually holds; the multi-key atomicity requirement is met by validating the whole write before applying any of it; re-running is a mechanical recovery an agent can be told to do | A failed plan leaves the world part-changed, and the caller must read the report to learn what landed | N/A — selected |

## Consequences

**Positive:**

- Every atomicity guarantee the requirements ask for is met by checking
  before writing, which is both simpler and stronger than undoing after:
  the multi-key manifest write validates all keys and then applies one
  merge, so a refused key means nothing was stored.
- Recovery is re-running, which an agent can be instructed to do without
  understanding what already happened — that is what the idempotence
  declaration on each step is for, and it is the property a resumed run
  depends on.
- Nothing in the CLI performs a mutation whose purpose is to reverse
  another mutation, so there is no path that writes while the world is in
  a state the plan did not expect.
- Ordering irreversible writes last converts most partial failures into
  no-write failures.

**Negative:**

- A plan that fails midway leaves the world part-changed. The report
  names which steps ran, and that report is the only record — a caller
  that ignores it cannot reconstruct the state from the exit status
  alone.
- Some sequences cannot put every irreversible write last. Accepting a
  dispatch merges, then removes a worktree, then pushes; a failure at the
  removal leaves a merged branch and a standing worktree, which is a
  state a human or a later command resolves rather than one the CLI
  reverses.
- Each step's idempotence declaration is a claim that must be true, and
  nothing but a test enforces it. A step declared safe to re-run that is
  not turns a mechanical recovery into a second failure.

## References

- [[frd-swarm-core]] — AC-SWARM-CORE-002.14, the one atomicity
  obligation, and the dry-run mode that makes a write previewable
- [[adr-024]] — the plan whose failure policy this decides
