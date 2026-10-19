---
title: "ADR-004: Introduce a Module Blueprint; Component as the Atomic Leaf"
description: Split the component blueprint into an atomic component type and a new module type
slug: adr-004
type: adr
status: active
---

# ADR-004: Introduce a module blueprint; component as the atomic leaf

**Status:** active  
**Date:** 2026-07-29

## Context

The framework originally used one component blueprint for everything from a
small shared library to a complex composite subsystem. As the plugin began to
drive isolated, orchestrated agent execution (see the groom-beads skill), two
problems appeared:

- An agent working on a single work item needs a small, single-purpose
  document to load as context. One blueprint that mixes an atomic unit's
  interface, a composite's internal structure, and a shared artifact's
  versioning policy is too much.
- The concerns that a component carries vary widely. An atomic helper inside a
  container has none of the versioning or internal-structure concerns that a
  shared SDK or a composite plugin has.

Overloading one type made documents harder both to write and to use.

## Decision

We will split the single component blueprint into two document types:

- The **component blueprint** becomes the atomic leaf: an internal building
  block of one container, with an interface and change guidance, and with no
  internal components of its own.
- A new **module blueprint** covers units that have internal components, cross
  a boundary of independent change, or both.

The architecture taxonomy grows from eight document types to nine.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| One flexible component blueprint with optional sections gated by kind | Keeps the taxonomy small<br>No decision about which type to use | Overloads one type across three very different altitudes<br>A bead that loads a component document still pulls in composite and versioning sections that don't apply to an atomic change<br>The "component" superset obscures the C4 vocabulary | The gains in context size and clarity from a dedicated type outweigh the cost of one more type |
| Status quo: extract a component blueprint only when the unit is sophisticated or shared | Fewest documents | Most units stay embedded in the container blueprint, so a bead must load the whole container to change one unit<br>No home for composite structure once it's removed from the container | Fails the context-size goal for isolated agents |
| Add a module blueprint; the component becomes atomic | Small, single-purpose documents for each unit of change<br>Each type carries only the concerns that apply to it<br>The C4 "component" is truly atomic again | The taxonomy grows from 8 to 9 types<br>Authors must decide between component and module | N/A—selected |

## Consequences

**Positive:**

- Agents load minimal, relevant context for each unit of change.
- Each type carries only the concerns that apply to it. For example, a
  component has no versioning section, and a module can have one.
- The framework's "component" aligns with classic C4, where a component is
  atomic, and composite structure and versioning get a clear home in the
  module type.

**Negative:**

- The taxonomy, which is deliberately small, has one more document type.
- Authors must decide between component and module. The boundary test in
  [[adr-005]] reduces this cost.
- Adopting projects must migrate existing single-component blueprints to one
  type or the other.

## References

- [[adr-005]]: the test for which type a unit is.
- [[adr-006]]: how a module is classified.
- [[adr-007]]: when to create either blueprint.
- `references/c4-guide.md`: the abstraction levels, including the module
  extension.
- Templates: `module-blueprint-template.md` and
  `component-blueprint-template.md`.
