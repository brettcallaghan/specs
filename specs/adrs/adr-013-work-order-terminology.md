---
title: "ADR-013: Work Order as the Name for the Atomic Unit of Executable Work"
description: The portable unit is a work order and its beads materialization is a bead; "work item" and "task" are retired as synonyms
slug: adr-013
type: adr
status: ready-for-review
---

# ADR-013: Work Order as the name for the atomic unit of executable work

**Status:** ready-for-review
**Date:** 2026-07-30

## Context

The framework's atomic, portable unit of executable work went by more than
one name across the corpus. Occurrences, measured in each skill before the
rename this ADR records:

| Skill | "work order" | "work item" | "task" (as the container) |
|---|---|---|---|
| `spec` | 11 | 0 | 0 |
| `groom-beads` | 9 | 0 | 0 |
| `groom-feature-blueprint` | 6 | **8** | 0 |
| `retro-spec` | 0 | **6** | 0 |

`groom-feature-blueprint`—the skill whose entire job is *producing* the
thing—used the non-canonical name more often than the canonical one.
`retro-spec` used only the non-canonical one. Meanwhile, `spec` promotes
**Work Order** to one of its first-class document types.

The corpus also conflates two distinct concepts. The framework already
draws the line clearly in `groom-beads`'s mapping table but doesn't name it:

- the **portable format** and its readiness bar, owned by `spec`, which is
  tracker-agnostic and is what makes a unit of work dispatchable; and
- the **materialization** of that format into a specific tracker, owned by
  `groom-beads`, currently the Beads `bd` tracker.

A reader who meets "work item" can't tell which of the two it means.

The open question is whether to keep the two-word term at all or to adopt
the more familiar SDLC word **task**.

## Decision

The portable unit is a **work order**. Its Beads materialization is a
**bead**. "Work item" is retired. "Task" is reserved for its existing
meaning as a value of the `kind` / `type` enum and is never used for the
container.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Task** | Universally familiar across Jira, Linear, GitHub Projects, and Asana. Short. Natural as a verb phrase ("dispatch a task") | Already occupies a slot in the `kind` vocabulary (`epic \| feature \| task \| bug \| chore`), which produces "a task of kind task" alongside "a task of kind bug". The enum is built into `bead.schema.json`, `validate_beads.py`, the Definition of Ready, and the work-order → bead mapping. Collides with the harness's own `TaskCreate` / `TodoWrite` tools, which this repo's `AGENTS.md` explicitly bans for tracking. Naming the core artifact "task" invites exactly the confusion that rule exists to prevent. Carries no readiness connotation: a task is a thing to do, whereas the whole point of this concept is a unit that is *authorized, specified, and dispatchable*. | The enum collision alone disqualifies it, and the harness collision compounds it |
| **Work item** | Azure DevOps' real term; was in use in two skills | Vaguer than work order: an "item" is any entry in a list. It's the name that *caused* the drift, and it still doesn't distinguish format from materialization | Vagueness is the defect being fixed |
| **Job** | Short. Suggests dispatch ("job queue," "job runner"). Agent-native | Implies batch compute rather than specified work. Loses the authorization and specification connotation. "Job" has no readiness bar in any tradition | Weaker semantics than work order for no structural gain |
| **Order** | Short form of the chosen term | Collides with *ordering* (sequencing), which this framework discusses constantly ("ordering lives in the graph, not in prose") | Hard collision with an existing load-bearing term |
| **Work order** | A real term of art from manufacturing, field service, and ITSM that means precisely *a self-contained, authorized instruction to perform a specified piece of work*. Carries the readiness connotation the concept depends on. Doesn't collide with the `kind` enum, with harness tooling, or with sequencing vocabulary. Distinctive, so it's greppable and unambiguous in prose | Two words. Slightly formal. Not native to agile vocabulary | N/A — selected |

The "not native to agile" objection is weak here and arguably a point in
favor. This framework deliberately replaces SDLC ceremonies with
agent-native equivalents, and the manufacturing metaphor is the accurate
one: an assembly line dispatching authorized work orders to stations is
exactly the mental model of an orchestrator dispatching to isolated
subagents.

## Consequences

**Positive:**
- One name for the portable concept and one for its materialization, with
  an explicit boundary between them instead of one implied by a mapping
  table.
- No ambiguity with the `kind: task` enum value, which stays as is.
- No ambiguity with harness task tooling in an environment where that
  tooling is present and deliberately unused.
- "Work order" is distinctive enough to grep, so future drift is
  mechanically detectable.

**Negative:**
- Two words take longer to type and read than "task" would.
- Newcomers meet an unfamiliar term of art. The on-ramp documentation
  mitigates this by defining it once, up front.
- A rename cost was paid across `groom-feature-blueprint` and `retro-spec`.

## References

- `plugin/skills/spec/references/work-order-template.md` — the format
- [[work-order-definition-of-ready]] — the readiness bar
- `plugin/skills/groom-beads/SKILL.md` — the work-order → bead mapping
