---
title: "ADR-013: Work Order as the Name for the Atomic Unit of Executable Work"
description: The portable unit is a work order and its beads materialization is a bead; "work item" and "task" are retired as synonyms
slug: adr-013
type: adr
status: ready-for-review
---

# ADR-013: Work Order as the Name for the Atomic Unit of Executable Work

**Status:** ready-for-review
**Date:** 2026-07-30

## Context

The corpus uses four different names for the framework's atomic, portable
unit of executable work. The following table shows the counts measured
across the skills:

| Skill | "work order" | "work item" | "task" (as the container) |
|---|---|---|---|
| `spec` | 11 | 0 | 0 |
| `groom-beads` | 9 | 0 | 0 |
| `groom-feature-blueprint` | 6 | **8** | 0 |
| `retro-spec` | 0 | **6** | 0 |

`groom-feature-blueprint`, the skill whose entire job is to produce the
unit, uses the non-canonical name more often than the canonical one, and
`retro-spec` uses only the non-canonical name. Meanwhile, `spec` makes
**Work Order** one of its first-class document types.

The corpus also conflates two distinct concepts. `groom-beads`'s mapping
table already separates them clearly but doesn't name them:

- The **portable format** and its readiness bar, owned by `spec`. The format
  is tracker-agnostic, and it's what makes a unit of work dispatchable.
- The **materialization** of that format into a specific tracker, owned by
  `groom-beads`. The current tracker is Beads (`bd`).

A reader who sees "work item" can't tell which of the two is meant.

The open question that this ADR settles is whether to keep the two-word term
at all, or to adopt the more familiar SDLC word **task**.

## Decision

The portable unit is a **work order**. Its Beads materialization is a
**bead**. "Work item" is retired. "Task" keeps its existing meaning as a
value of the `kind` / `type` enum and is never used for the container.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Task** | Universally familiar across Jira, Linear, GitHub Projects, and Asana; short; natural in a verb phrase ("dispatch a task") | Already occupies a slot in the `kind` vocabulary (`epic \| feature \| task \| bug \| chore`), which produces "a task of kind task" alongside "a task of kind bug"; the enum is built into `bead.schema.json`, `validate_beads.py`, the Definition of Ready, and the work-order-to-bead mapping. Collides with the harness's own `TaskCreate` and `TodoWrite` tools, which this repository's `AGENTS.md` explicitly bans for tracking. Naming the core artifact "task" invites the confusion that rule exists to prevent. Has no readiness connotation: a task is a thing to do, but this concept is a unit that is *authorized, specified, and dispatchable*. | The enum collision alone disqualifies it, and the harness collision adds to it |
| **Work item** | The actual Azure DevOps term; already in use in two skills | Vaguer than work order, because an "item" is any entry in a list; it's the name that currently *causes* the drift; it still doesn't distinguish format from materialization | Vagueness is the defect being fixed |
| **Job** | Short; associated with dispatch ("job queue", "job runner"); agent-native | Implies batch compute rather than specified work; loses the authorization and specification connotation; "job" has no readiness bar in any tradition | Weaker meaning than work order, with no structural gain |
| **Order** | Short form of the chosen term | Collides with *ordering* and sequencing, which this framework discusses constantly ("ordering lives in the graph, not in prose") | Hard collision with an existing, important term |
| **Work order** | A real term of art from manufacturing, field service, and ITSM that means *a self-contained, authorized instruction to perform a specified piece of work*; carries the readiness connotation that the concept depends on; doesn't collide with the `kind` enum, harness tooling, or sequencing vocabulary; distinctive, so it's easy to search for and unambiguous in prose | Two words; slightly formal; not native to agile vocabulary | Not applicable (selected) |

The "not native to agile" objection is weak here, and arguably supports the
choice. This framework deliberately replaces SDLC ceremonies with
agent-native equivalents, and the manufacturing metaphor is the accurate one:
an assembly line that dispatches authorized work orders to stations matches
the mental model of an orchestrator that dispatches work to isolated
subagents.

## Consequences

**Positive:**

- The portable concept has one name and its materialization has another, and
  the boundary between them is explicit rather than implied by a mapping
  table.
- The term doesn't conflict with the `kind: task` enum value, which stays as
  is.
- The term doesn't conflict with harness task tooling in an environment where
  that tooling is present and deliberately unused.
- "Work order" is distinctive enough to search for, so a script can detect
  future drift. For details, see the consistency check in this repository's
  improvement plan.

**Negative:**

- Two words take longer to type and read than "task" would have.
- Newcomers meet an unfamiliar term of art. The on-ramp documentation
  mitigates this by defining the term once, at the start.
- Renaming the term across `groom-feature-blueprint` and `retro-spec` had a
  cost.

## References

- `plugin/skills/spec/references/work-order-template.md`: the format.
- `plugin/skills/spec/references/work-order-definition-of-ready.md`: the
  readiness bar.
- `plugin/skills/groom-beads/SKILL.md`: the work-order-to-bead mapping.
