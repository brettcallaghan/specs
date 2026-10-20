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

The framework needs a policy for when a unit earns its own blueprint instead of
staying described inside its container. A conservative trigger, "warrants
standalone documentation" (essentially shared-ness), leaves most units
embedded. An agent that changes one unit then has to load the whole container's
context. The opposite extreme is a separate document for every unit. It would
create a maintenance and drift tax on trivial glue and would fragment the
container's relationship map. The goal is document granularity that tracks the
actual unit of change without either penalty.

## Decision

We will create blueprints by **promotion**. A component or module starts as a
bullet with a machine-readable record block in its container's Components
roster. It graduates to its own blueprint when it is a self-contained **seam**.
A seam has an interface others depend on, plus enough internal behavior that a
change to it is a self-contained unit of work an agent can reason about
locally. A seam can also be documented so as to stand alone. The container
always carries the full roster. Promotion is a lockstep update across every
container that composes the unit.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Conservative "warrants standalone documentation" (shared-ness) | Fewest documents | Most units stay embedded<br>A bead loads the whole container<br>Granularity doesn't track the unit of change | Fails the context-isolation goal behind the split |
| One blueprint per unit, always | Maximal granularity<br>Smallest context payload | Drift tax on trivial glue<br>Fragments the relationship map<br>High authoring friction | Over-documents<br>Cheap to generate but not to keep true |
| Promotion at a self-contained seam, with a triviality floor | Granularity tracks the unit of change (and bead collision keys, model routing, context isolation)<br>Trivial glue stays a cheap record-block bullet | Authors must judge whether a unit is a seam | N/A — selected |

## Consequences

**Positive:**
- Document granularity tracks the unit of change an agent actually works on
- A gentle adoption path: start embedded, graduate when warranted
- Even un-promoted units stay in the element database through their record
  block ([[adr-008]])

**Negative:**
- "Is this a seam?" is a judgment call
- Promotion requires linking the new blueprint from every composing
  container's roster in the same change. The record block may stay, and
  `specs check` reports a block whose fields disagree with the blueprint

## References

- [[adr-004]] — the types this policy
  applies to
- [[adr-008]] — record blocks make promotion mechanical
- `references/c4-guide.md`; template: `container-blueprint-template.md`
