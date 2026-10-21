---
title: "ADR-004: Introduce a Module Blueprint; Component as the Atomic Leaf"
description: Split the component blueprint into an atomic component type and a new module type
slug: adr-004
type: adr
status: active
---

# ADR-004: Introduce a Module Blueprint; component as the atomic leaf

**Status:** active  
**Date:** 2026-07-29

## Context

A single Component Blueprint that covers everything from a small shared
library to a complex composite subsystem comes under two pressures when the
plugin drives isolated, orchestrated agent execution (see the groom-beads
skill). First, an agent working a single work item needs a small,
single-purpose document to load as context. It doesn't need one blueprint
that mixes an atomic unit's interface with a composite's internal structure
and a shared artifact's versioning policy. Second, the concerns a "component"
carries vary enormously: an atomic in-container helper has none of the
versioning or internal-structure concerns that a shared SDK or a composite
plugin has. Overloading one type makes both authoring and consumption harder.

## Decision

We will split the single component blueprint into two document types. The
**component blueprint** becomes the atomic leaf: an internal building block
of one container, with an interface and change guidance and no internal
components of its own. A new **module blueprint** covers units that have
internal components and/or cross a boundary of independent change. The
architecture taxonomy grows from eight document types to nine.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| One flexible Component Blueprint with kind-gated optional sections | Keeps the taxonomy small<br>No "which type do I use?" decision | Overloads one type across three very different altitudes<br>A bead that loads a Component Blueprint still pulls in composite and versioning sections that don't apply to an atomic change<br>The "component" superset obscures the C4 vocabulary | A dedicated type's gains in context size and clarity outweigh the cost of one more type |
| Status quo: extract a Component Blueprint only when a unit is "sophisticated" or shared | Fewest documents | Most units stay embedded in the Container Blueprint, so a bead must load the whole container to change one unit<br>No home for composite structure once it's removed from the container | Fails the isolated-agent goal for context size |
| Add a Module Blueprint; component becomes atomic | Small, single-purpose docs per unit of change<br>Each type carries only the concerns that apply to it<br>C4 "component" is truly atomic again | Taxonomy grows from eight types to nine<br>Authors must decide where the boundary between component and module falls | N/A — selected |

## Consequences

**Positive:**
- Agents load minimal, relevant context for each unit of change.
- Each type carries only the concerns that apply to it. A component has no
  versioning section, and a module can have one.
- The framework's "component" is atomic, as in classic C4, and composite
  structure and versioning get a clear home in the Module Blueprint.

**Negative:**
- A deliberately small taxonomy has one more document type.
- Authors must decide between component and module. The crisp boundary test
  in [[adr-005]] makes that easier.

## References

- [[adr-005]] — the test for which type a unit is
- [[adr-006]] — how a module is classified
- [[adr-007]] — when to create either blueprint
- [[c4-standards#The Abstraction Levels]] — the abstraction levels, including
  the module extension
- Templates: `plugin/skills/spec/references/module-blueprint-template.md`,
  `plugin/skills/spec/references/component-blueprint-template.md`
