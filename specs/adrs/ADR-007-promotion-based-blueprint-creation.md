---
title: "ADR-007: Create Blueprints by Promotion at a Seam"
description: A unit starts as a record-block bullet in its container and graduates to its own blueprint at a self-contained seam
slug: adr-007-promotion-based-blueprint-creation
type: adr
status: active
updated_at: 2026-07-29T00:00:00Z
---

# ADR-007: Create Blueprints by Promotion at a Seam

**Status:** active  
**Date:** 2026-07-29

## Context

The framework needed a policy for when a unit earns its own blueprint versus
staying described inside its container. The original trigger was "warrants
standalone documentation" — essentially shared-ness — which left most units
embedded, forcing an agent that changes one unit to load the whole container's
context. The opposite extreme, a separate document for every unit, would create
a maintenance and drift tax on trivial glue and would fragment the container's
relationship map. We wanted doc granularity to track the actual unit of change
without either penalty.

## Decision

We will create blueprints by **promotion**. A component or module starts as a
bullet with a machine-readable record block in its container's Components
roster, and graduates to its own blueprint when it is a self-contained **seam**:
it has an interface others depend on, plus enough internal behavior that a
change to it is a self-contained unit of work an agent can reason about locally,
and it can be documented so as to stand alone. The container always carries the
full roster; promotion is a lockstep update across every container that composes
the unit.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Conservative "warrants standalone documentation" (shared-ness) | Fewest documents | Most units stay embedded;<br>a bead loads the whole container;<br>granularity does not track the unit of change | Fails the context-isolation goal that motivated the split |
| One blueprint per unit, always | Maximal granularity;<br>smallest context payload | Drift tax on trivial glue;<br>fragments the relationship map;<br>high authoring friction | Over-documents;<br>cheap to generate but not to keep true |
| Promotion at a self-contained seam, with a triviality floor | Granularity tracks the unit of change (and bead collision keys, model routing, context isolation)<br>Trivial glue stays a cheap record-block bullet | Authors must judge whether a unit is a seam | N/A — selected |

## Consequences

**Positive:**
- Document granularity tracks the unit of change an agent actually works on
- A gentle adoption path: start embedded, graduate when warranted
- Even un-promoted units stay in the element database via their record block
  ([[adr-008]])

**Negative:**
- "Is this a seam?" is a judgment call
- Promotion requires a lockstep roster update in every composing container,
  enforced by the record-block-XOR-link rule

## References

- [[adr-004]] — the types this policy
  applies to
- [[adr-008]] — record blocks make promotion mechanical
- `references/c4-guide.md`; template: `container-blueprint-template.md`
