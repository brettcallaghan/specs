---
title: "ADR-024: A Command Builds a Plan; Three Interpreters Read It"
description: Every swarm command produces an ordered plan of typed step
  values that explain, dry-run and execute all read, rather than writing the
  sequence three times or deriving two modes from a third
slug: adr-024
type: adr
status: approved
updated_at: 2026-08-28T00:00:00Z
---

# ADR-024: A Command Builds a Plan; Three Interpreters Read It

**Status:** approved  
**Date:** 2026-08-28

## Context

[[frd-swarm-core]] requires every command to behave three ways over one
sequence of external invocations. It executes them while printing each as
it runs (AC-SWARM-CORE-001.7); it prints what it would run and executes
none of it (AC-SWARM-CORE-001.4); and it performs every read while
writing nothing (AC-SWARM-CORE-001.6). It also requires an ordering
guard rather than a shell conjunction — the incident that motivates the
whole feature is a worktree removed after a merge exited 128 — and
per-step timing, since the progress measurements are ages.

Written as straight-line Go that calls a process runner, none of that
follows. Explain mode becomes a second function printing strings that
must be kept in agreement with the first by hand, which is the same
failure the CLI exists to remove one level up. A guard becomes an `if err
!= nil` an author can forget exactly as they can forget `&&`.

The composition the family needs is modest and worth stating before
choosing a mechanism for it. Accepting a dispatch is eight steps where
each guards the next, with one conditional retry. Writing the manifest is
read, validate, write. Reporting status is a fan of independent reads
joined at the end. Nothing here is a long-running, resumable, distributed
workflow.

## Decision

We will have every command build a **plan**: an ordered list of step
values, each describing one external invocation, how to decode its
result, what to do when it fails, and whether re-running it is safe. Three
interpreters consume that one plan — `Explain` renders it and runs
nothing, `DryRun` runs the steps marked as reads and refuses the rest,
and `Execute` runs all of them, printing each as it goes. Branching is a
property of a step (an alternative plan attached to a named outcome), not
of the code that built it, so a plan is fully renderable before anything
runs. No workflow, DAG or state-machine library is adopted.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Straight-line Go calling a runner | Simplest to write; ordinary control flow | Explain mode is a second implementation of the same sequence, kept in agreement by hand; a guard is a line an author can omit; there is no value to attach timing or idempotence to | Recreates, inside the CLI, the drift the CLI exists to remove |
| A closure per step, executed by a driver | Guards and timing live in the driver | A closure cannot be rendered — explain mode must call it to find out what it would do, which is exactly what it must not do | Cannot satisfy explain without executing |
| Explain by running with stubbed results | One code path | The stub decides which branch is taken, so the printed plan is the plan for an imaginary repository; it diverges precisely where the real run is interesting | Prints a transcript that is confidently wrong |
| A workflow engine (Temporal, Cadence, `go-workflows`) | Durable execution, retries, visibility | All are built for distributed, resumable, hours-long workflows with a server or a worker pool and a persistence model. This CLI is one process that lives for seconds and holds no state of its own ([[frd-swarm-core]]) | Adds a runtime model and a dependency for none of the properties that motivate it |
| A generic DAG library | Dependency ordering for free | The family's real shape is a short chain with one conditional and one fan of reads; a DAG buys expressiveness nobody needs and costs a second vocabulary for readers | Solves a problem the inventory does not contain |
| Plan of typed step values, three interpreters | Explain, dry-run and execute cannot drift because they read one structure; the guard, the timing, the audit line and the idempotence declaration are fields rather than conventions; a plan is a testable value | An indirection to learn before writing a command; a step's arguments that depend on an earlier result need an explicit reference type rather than a variable | N/A — selected |

## Consequences

**Positive:**

- The three modes cannot disagree about what a command does, because
  there is one description and three readers of it. A step added for
  execution appears in explain output with no second edit.
- The ordering guard is structural. A step declares that failure stops
  the plan, and the interpreter honours it; there is no conjunction to
  write and none to forget.
- Timing, the printed invocation, and the audit record are all derived
  from the step value, so every command gets them without its author
  doing anything.
- A plan is a value, so a unit test can assert the sequence a command
  builds without running a single subprocess — which is what makes the
  guard itself testable rather than only its effects.
- Independent reads are visible as such in the plan, so parallelising the
  status fan later is a change to one interpreter rather than to every
  command. `golang.org/x/sync/errgroup` covers it when the measurement
  says it is worth doing; for a handful of subprocesses it may not be.

**Negative:**

- An argument produced by an earlier step is a reference value rather
  than a Go variable, which is more ceremony to write and to read. This
  is what AC-SWARM-CORE-001.5 asks for and it does not come free.
- Explain output for a branching plan is not a single runnable script.
  The transcript names the condition and renders both arms, so a reader
  follows the one their situation matches; AC-SWARM-CORE-001.13 is
  satisfied by a labelled transcript rather than by a linear one, and any
  command whose plan branches must be read that way.
- A command whose logic genuinely needs a result before it can decide
  what to do next must either express that as a branch on the step or
  split into two plans. Where it splits, explain shows the first plan and
  names what the second depends on.
- The indirection is a cost paid by every command, including the ones
  simple enough not to need it.

## References

- [[frd-swarm-core]] — the three modes, the ordering guard and the
  timing this structure exists to provide
- [[adr-025]] — how one step's invocation is run and its result typed
- [[adr-026]] — why the failure policy is stop-and-report rather than
  compensate
- `docs/research/swarm-cli/2026-08-25-capability-audit.md` §3.3 — the
  five properties a composition mechanism was asked to carry
