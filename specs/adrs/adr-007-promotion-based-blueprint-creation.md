---
title: "ADR-007: Create Blueprints by Promotion at a Seam"
description: A unit starts as a record-block bullet in its container and graduates to its own blueprint at a self-contained seam
slug: adr-007
type: adr
status: active
---

# ADR-007: Create Blueprints by Promotion at a Seam

**Status:** active  
**Date:** 2026-07-29

## Context

The framework needed a policy for when a unit earns its own blueprint and when
it stays described inside its container. The original trigger was "warrants
standalone documentation," which in practice meant shared-ness. That rule left
most units embedded, so an agent that changed one unit had to load the whole
container's context. The opposite extreme, a separate document for every unit,
would make trivial glue costly to maintain and prone to drift, and would
fragment the container's relationship map. We wanted document granularity to
track the actual unit of change without either penalty.

## Decision

We will create blueprints by **promotion**. A component or module starts as a
bullet, with a machine-readable record block, in its container's Components
roster. It graduates to its own blueprint when it is a self-contained **seam**.
A seam meets all of these conditions:

- It has an interface that others depend on.
- It has enough internal behavior that a change to it is a self-contained unit
  of work that an agent can reason about locally.
- It can be documented so that it stands alone.

The container always carries the full roster. Promotion is a lockstep update
across every container that composes the unit.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Conservative "warrants standalone documentation" (shared-ness) | Fewest documents | Most units stay embedded;<br>a bead loads the whole container;<br>granularity doesn't track the unit of change | Fails the context-isolation goal that motivated the split |
| One blueprint per unit, always | Maximal granularity;<br>smallest context payload | Drift tax on trivial glue;<br>fragments the relationship map;<br>high authoring friction | Over-documents;<br>cheap to generate but not to keep true |
| Promotion at a self-contained seam, with a triviality floor | Granularity tracks the unit of change (and so do bead collision keys, model routing, and context isolation)<br>Trivial glue stays a cheap record-block bullet | Authors must judge whether a unit is a seam | N/A — selected |

## Consequences

**Positive:**
- Document granularity tracks the unit of change that an agent works on.
- Adoption is gradual: a unit starts embedded and graduates when it qualifies.
- Units that aren't promoted still appear in the element database through their
  record blocks ([[adr-008]]).

**Negative:**
- "Is this a seam?" is a judgment call.
- Promotion requires a link to the new blueprint from the roster of every
  container that composes the unit, in the same change. The record block can
  stay. If its fields disagree with the blueprint, `specs check` reports it.

## References

- [[adr-004]] — the types this policy applies to
- [[adr-008]] — record blocks make promotion mechanical
- `references/c4-guide.md`; template: `container-blueprint-template.md`
