---
title: Findings Capture Protocol
description: How agents record retro evidence in-flight during
  multi-agent phases — the finding format, its kinds and targets, when
  to record, where, and what the capturing agent must not do
slug: findings-capture-protocol
type: standards
status: active
updated_at: 2026-08-25T00:00:00Z
version: "0.6"
---

# Findings Capture Protocol

Retro evidence does not survive being recalled later, and in a
multi-agent phase (implementing a groomed plan, where each work order
may go to a different spawned agent) no single session holds it. This
protocol is how that evidence is captured at the moment of friction.

This document is the single statement of the protocol. Skills that
capture or harvest findings vendor it into their own
`references/shared/` and point at that copy; they do not restate it.
Do not copy it into a target project, and do not restate it per work
order or per issue — an item carries only the short standing
instruction its grooming skill stamps, which is self-sufficient on its
own.

## The finding

```
retro:<kind>[:<target>] <one line: what happened>

evidence:
  <the quote, the path:line, the command and its output, or expected
   vs observed>
context:
  <what was being attempted>                           (optional)
corrects:
  <finding-id>                                         (optional)
```

`kind` is required and `target` is required WHERE one exists. Together
they are what routes the finding, and they cost one word each at the
moment the agent already knows the answer.

### Record the evidence in full. Record none of the analysis.

Two rules, deliberately separate, because they have opposite reasons.

**No length limit.** Evidence does not survive this session — the
context compacts, the worktree is removed, the tree moves on. Whoever
aggregates a finding reads it cold, weeks later, with none of what the
capturing agent could see. **A finding SHALL carry at least one of**,
preferring the most specific available:

- the **exact quote** of the instruction, AC or spec text that misled,
  with where it is;
- a **`path:line`** anchor;
- the **command run and the output observed**;
- **expected versus observed**, as two concrete values.

A quote, a path and a command beat three sentences of prose, and are
usually shorter. A sensor that under-reports its reading is not being
disciplined; it is being lossy.

**No analysis.** Root cause, proposal, priority, and any sentence
beginning "we should" are all forbidden. They are derivable from the
evidence at any time, by an agent with the whole run in view and no
task of its own to finish. **The capturing agent is a sensor**, and one
that stops to analyse both derails its own work and reasons worse than
the aggregator will.

## The kinds

Nine. The first six are what an implementing agent notices about the
work it was given; the seventh is about the tooling that gave it; the
last two are not defects at all.

| kind | fires when |
|---|---|
| `gap` | information was needed that the work order and its listed references did not carry — had to open the blueprint, FRD or code to fill it. **Record even small occurrences**: each is direct evidence about whether items are self-sufficient enough to delegate |
| `decision` | a design decision had to be made that the item or blueprint should have pinned — a path, a default, a scope cut, an edge-case rule |
| `stale` | the item contradicted repo reality — a stale anchor, moved code, a referenced element that no longer matches |
| `unverifiable` | an AC or citation obligation could not be satisfied as declared — an unclear verified-by level, a test that cannot meaningfully exist, verification blocked on work outside the item |
| `sequencing` | a broken build inherited from a predecessor, or a blocker the dependency graph did not encode |
| `scope` | the item was mis-scoped — too large to review as one unit, trivially small, or overlapping another item's work |
| `tooling` | a skill, script, template or protocol misled, failed, or had no rule where one was needed |
| `proof` | **how a claim was established** — an observed pre-fix failure, a negative control, a prior risk checked and cleared |
| `environment` | a durable fact about a **dependency, tool or platform** the next agent will hit, which no work order could have carried |

**A clean run is fine — record nothing. Do not manufacture findings.**

### `proof` and `environment` are captured, never filed

Seven of the nine describe something that could be fixed. These two do
not, and treating them as defects gets both wrong:

- **`proof` is this standard being obeyed.** Evidence rules require a
  claim to name the command run and the result observed. An agent that
  runs a negative control before trusting a probe is doing exactly
  that, and routing the record as friction would suppress the rigour.
  It belongs in the report, attached to what it supports.
- **`environment` is nobody's defect.** A parser that exports no error
  type is not the work order's fault, not a spec's, and not a skill's.
  It belongs in the project's environmental-failure ledger, where the
  next agent will look for it.

Neither becomes an issue. Both are still worth the line it costs to
write them.

### `target` — what the finding is about

The skill or script (`retro:tooling:run-bead`), or the document or
item (`retro:gap:FRD-014`). Omit it only where there genuinely is not
one; an omitted target costs a cluster, not a finding.

### Which audience — apply the merge test

Findings serve two audiences and the difference is not visible in the
words. "The blueprint template has no rule for X" and "our blueprint
has no rule for X" are the same sentence about different artifacts.

> **Where would the fix land?**
>
> - **In this project's own repository** → the project's audience. It
>   becomes an issue in this project's tracker.
> - **In the repository that vends the skill** → the skill's audience.
>   It leaves through that project's outbox and is never filed here.

`tooling` is the skill audience; the six work-order kinds are the
project's. **A project SHALL NOT accumulate issues about skills it
cannot change.**

### A decision that changed what *correct* means is an ordinary finding

An amended acceptance criterion, a standard and an AC in tension, a
contract or interface change, a spec silent where the item required it
to speak — each already maps onto a kind above, respectively
`decision`, `decision` against two targets, `scope`, `gap`. Record it
under that kind, naming what was decided and what was rejected, and it
routes like every other finding.

**It carries no extra severity marker, and one SHALL NOT be
reintroduced.** A severity spelled inside the finding's own text
depends on an author remembering where the word goes. Measured on one
run: three sweeps produced three disagreeing counts (8, 11 and 17),
and that run's single most consequential record reached no audience at
all because the marker had been written into the `kind` slot, where
nothing matches it. What makes a finding load-bearing is legible from
its own text, and the kind is what routes it.

### Findings are append-only, so a correction supersedes

A finding SHALL NOT be edited or deleted after recording — they remain
on closed items permanently as raw provenance. An agent that finds its
own earlier finding wrong records a **new** one naming it:

```
corrects: <finding-id>
```

Without the field the relationship is invisible: an aggregator reading
both cold has only prose like "the line above" to go on.

## Finding ids are derived, never authored

`<item-id>#<ordinal>` — the ordinal being the finding's position among
that item's `retro:` records, ordered by creation. So `92u#2` is the
second finding on item `…-92u`.

The capturing agent assigns nothing, and cannot: the record's identity
does not exist until the write lands. The ordering is stable **because
findings are append-only**, which is what makes an ordinal safe to
cite.

## Where to record

- **Beads backend:** one `bd comment` per finding on the item's own
  issue, prefixed `retro:`.
- **Task-file backend:** append each finding as a bullet under a
  `## Retro findings` section at the end of the item's own file
  (per-item, so parallel agents never contend on a shared file).

## Harvesting

**One query over the phase's closure set**, and `bd export` rather than
`bd list` plus `bd show --json`: `bd list` defaults to open issues and
a drained closure set is almost entirely closed, while `bd show --json`
omits comments entirely even though its text output prints them.

```
bd export | jq -r 'select((.labels//[]) | index("<feature-slug>"))
  | .comments[]? | select(.text|startswith("retro:")) | .text'
```

(Task-file backend: each task file's Retro findings section instead.)

**A zero count across a drained set is a broken query, not a clean
run.** Stop, check the scope and the prefix spelling, and confirm the
items really are closed, before aggregating anything.

**What happens next is not this protocol's.** Routing, clustering,
deciding and filing belong to whichever retro consumes these findings;
this document ends where the harvest does.
