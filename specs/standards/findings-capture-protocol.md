---
title: Findings Capture Protocol
description: How agents record retro evidence in-flight during
  multi-agent phases — the finding format, its kinds and targets, when
  to record, where, and what the capturing agent must not do
slug: findings-capture-protocol
type: standards
status: active
---

# Findings capture protocol

Retro evidence doesn't survive being recalled later. In a multi-agent
phase (implementing a groomed plan, where each Work Order may go to a
different spawned agent), no single session holds it. This protocol is
how that evidence is captured at the moment of friction.

This document is the single statement of the protocol. Skills that
capture or harvest findings vendor it into their own
`references/shared/` and point at that copy. They don't restate it.
Don't copy it into a target project, and don't restate it per Work
Order or per issue. An item carries only the short standing instruction
that its grooming skill stamps, and that instruction is self-sufficient.

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

`kind` is required, and `target` is required where one exists. Together
they route the finding, and each costs one word at the moment the agent
already knows the answer.

### Record the evidence in full. Record none of the analysis.

These are two rules, kept separate on purpose, because they have
opposite reasons.

**No length limit.** Evidence doesn't survive this session: the context
compacts, the worktree is removed, and the tree moves on. Whoever
aggregates a finding reads it cold, weeks later, with none of what the
capturing agent could see. **A finding SHALL carry at least one of**,
preferring the most specific available:

- the **exact quote** of the instruction, AC or spec text that misled,
  with where it is;
- a **`path:line`** anchor;
- the **command run and the output observed**;
- **expected versus observed**, as two concrete values.

A quote, a path, and a command beat three sentences of prose, and
they're usually shorter. A sensor that under-reports its reading isn't
being disciplined. It's being lossy.

**No analysis.** Root cause, proposal, priority, and any sentence that
begins "we should" are all forbidden. An agent with the whole run in
view and no task of its own to finish can derive them from the evidence
at any time. **The capturing agent is a sensor.** One that stops to
analyze both derails its own work and reasons worse than the aggregator
will.

## The kinds

There are nine kinds. The first six are what an implementing agent
notices about the work it was given. The seventh is about the tooling
that gave it the work. The last two aren't defects at all.

| kind | fires when |
|---|---|
| `gap` | the agent needed information that the Work Order and its listed references didn't carry, and had to open the blueprint, FRD, or code to fill it. **Record even small occurrences**: each one is direct evidence of whether items are self-sufficient enough to delegate |
| `decision` | the agent had to make a design decision that the item or blueprint should have pinned: a path, a default, a scope cut, or an edge-case rule |
| `stale` | the item contradicted repo reality: a stale anchor, moved code, or a referenced element that no longer matches |
| `unverifiable` | an AC or citation obligation couldn't be satisfied as declared: an unclear verified-by level, a test that can't meaningfully exist, or verification blocked on work outside the item |
| `sequencing` | a broken build inherited from a predecessor, or a blocker that the dependency graph didn't encode |
| `scope` | the item was mis-scoped: too large to review as one unit, trivially small, or overlapping another item's work |
| `tooling` | a skill, script, template, or protocol misled, failed, or had no rule where one was needed |
| `proof` | **how a claim was established**: an observed pre-fix failure, a negative control, or a prior risk checked and cleared |
| `environment` | a durable fact about a **dependency, tool, or platform** that the next agent will hit, and that no Work Order could have carried |

**A clean run is fine. Record nothing, and don't manufacture findings.**

### `proof` and `environment` are captured, never filed

Seven of the nine kinds describe something that could be fixed. These
two don't, and treating them as defects gets both of them wrong:

- **`proof` is this standard being obeyed.** Evidence rules require a
  claim to name the command run and the result observed. An agent that
  runs a negative control before trusting a probe is doing exactly
  that, and routing the record as friction would suppress the rigor. A
  `proof` finding belongs in the report, attached to what it supports.
- **`environment` is nobody's defect.** A parser that exports no error
  type isn't the fault of the Work Order, a spec, or a skill. An
  `environment` finding belongs in the project's environmental-failure
  ledger, where the next agent will look for it.

Neither becomes an issue. Both are still worth the line they cost to
write.

### `target` — what the finding is about

The target is the skill or script (`retro:tooling:swarm-dispatch`), or
the document or item (`retro:gap:FRD-014`). Omit it only where there
genuinely isn't one. An omitted target costs a cluster, not a finding.

### Which audience — apply the merge test

Findings serve two audiences, and the difference isn't visible in the
words. "The blueprint template has no rule for X" and "our blueprint
has no rule for X" are the same sentence about different artifacts.

> **Where would the fix land?**
>
> - **In this project's own repository** → the project's audience. It
>   becomes an issue in this project's tracker.
> - **In the repository that vends the skill** → the skill's audience.
>   It's reported to whoever vends the skill, and it's never filed here.

`tooling` is the skill audience. The six Work Order kinds are the
project's. **A project SHALL NOT accumulate issues about skills it
cannot change.**

### A decision that changed what *correct* means is an ordinary finding

Each of these already maps onto a kind above:

- an amended acceptance criterion: `decision`
- a standard and an AC in tension: `decision`, against two targets
- a contract or interface change: `scope`
- a spec that's silent where the item required it to speak: `gap`

Record it under that kind, naming what was decided and what was
rejected, and it routes like every other finding.

### Findings are append-only, so a correction supersedes

A finding SHALL NOT be edited or deleted after recording — they remain
on closed items permanently as raw provenance. An agent that finds its
own earlier finding wrong records a **new** one that names it:

```
corrects: <finding-id>
```

Without the field, the relationship is invisible: an aggregator that
reads both cold has only prose like "the line above" to go on.

## Finding ids are derived, never authored

A finding id is `<item-id>#<ordinal>`, where the ordinal is the
finding's position among that item's `retro:` records, ordered by
creation. So `92u#2` is the second finding on item `…-92u`.

The capturing agent assigns nothing, and it can't: the record's
identity doesn't exist until the write lands. The ordering is stable
**because findings are append-only**, and that stability is what makes
an ordinal safe to cite.

## Where to record

- **Beads backend:** one `bd comment` per finding on the item's own
  issue, prefixed `retro:`.
- **Task-file backend:** append each finding as a bullet under a
  `## Retro findings` section at the end of the item's own file
  (one file per item, so parallel agents never contend on a shared
  file).

## Harvesting

**Run one query over the phase's closure set**, with `bd export`
instead of `bd list` plus `bd show --json`, for two reasons. `bd list`
defaults to open issues, and a drained closure set is almost entirely
closed. `bd show --json` omits comments entirely, even though its text
output prints them.

```
bd export | jq -r 'select((.labels//[]) | index("<feature-slug>"))
  | .comments[]? | select(.text|startswith("retro:")) | .text'
```

(For the task-file backend, read each task file's Retro findings section
instead.)

**A zero count across a drained set is a broken query, not a clean
run.** Before you aggregate anything, stop, check the scope and the
prefix spelling, and confirm that the items really are closed.

**What happens next isn't this protocol's.** Routing, clustering,
deciding, and filing belong to whichever retro consumes these findings.
This document ends where the harvest does.
