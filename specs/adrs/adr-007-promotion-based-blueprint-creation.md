---
title: "ADR-007: Create Blueprints by Promotion at a Seam"
description: A unit starts as a record-block bullet in its container and graduates to its own blueprint at a self-contained seam
slug: adr-007
type: adr
status: active
---

# ADR-007: Create blueprints by promotion at a seam

**Status:** active  
**Date:** 2026-07-29

## Context

The framework needed a policy that decides when a unit gets its own blueprint
and when it stays described inside its container. The original trigger was
"warrants standalone documentation," which in practice meant shared-ness. That
trigger left most units embedded, so an agent that changes one unit had to load
the context of the whole container. The opposite extreme, a separate document
for every unit, adds maintenance cost and drift to trivial glue code, and it
fragments the container's relationship map. We wanted document granularity to
track the actual unit of change without either cost.

## Decision

We create blueprints by **promotion**. A component or module starts as a
bullet with a machine-readable record block in its container's Components
roster. It graduates to its own blueprint when it's a self-contained **seam**,
which means that all of the following are true:

- It has an interface that other units depend on.
- It has enough internal behavior that a change to it is a self-contained unit
  of work that an agent can reason about locally.
- It can be documented so that the document stands alone.

The container always carries the full roster. Promotion is a lockstep update
across every container that composes the unit.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Conservative "warrants standalone documentation" (shared-ness) | Fewest documents | Most units stay embedded.<br>A bead loads the whole container.<br>Granularity doesn't track the unit of change. | Fails the context-isolation goal that motivated the split. |
| One blueprint per unit, always | Maximal granularity.<br>Smallest context payload. | Adds drift cost to trivial glue.<br>Fragments the relationship map.<br>High authoring friction. | Over-documents.<br>Documents are cheap to generate but not to keep accurate. |
| Promotion at a self-contained seam, with a triviality floor | Granularity tracks the unit of change, and with it bead collision keys, model routing, and context isolation.<br>Trivial glue stays a cheap record-block bullet. | Authors must judge whether a unit is a seam. | Not applicable; selected. |

## Consequences

**Positive:**

- Document granularity tracks the unit of change that an agent actually works
  on.
- Adoption is gradual: a unit starts embedded and graduates when it qualifies.
- Units that aren't promoted still stay in the element database through their
  record block ([[adr-008]]).

**Negative:**

- Whether a unit is a seam is a judgment call.
- Promotion requires linking the new blueprint from every composing
  container's roster in the same change. The record block can stay, and
  `specs check` reports a block whose fields disagree with the blueprint.

## References

- [[adr-004]]: the types that this policy applies to
- [[adr-008]]: record blocks make promotion mechanical
- `references/c4-guide.md`; template: `container-blueprint-template.md`
