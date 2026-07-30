---
title: "ADR-004: Introduce a Module Blueprint; Component as the Atomic Leaf"
description: Split the component blueprint into an atomic component type and a new module type
slug: adr-004-module-blueprint-and-atomic-component
type: adr
status: active
updated_at: 2026-07-29T00:00:00Z
---

# ADR-004: Introduce a Module Blueprint; Component as the Atomic Leaf

**Status:** active  
**Date:** 2026-07-29

## Context

The framework originally used one component blueprint to cover everything
from a small shared library to a complex composite subsystem. As the plugin
began driving isolated, orchestrated agent execution (see the groom-beads
skill), two pressures surfaced. First, an agent working a single work item
needs a small, single-purpose document to load as context, not one blueprint
that mixes an atomic unit's interface with a composite's internal structure
and a shared artifact's versioning policy. Second, the concerns a "component"
carries vary enormously: an atomic in-container helper has none of the
versioning or internal-structure concerns a shared SDK or a composite plugin
has. Overloading one type made both authoring and consumption harder.

## Decision

We will split the single component blueprint into two document types. The
**component blueprint** becomes the atomic leaf: an internal building block
of one container, with an interface and change guidance and no internal
components of its own. A new **module blueprint** covers units that have
internal components and/or cross a boundary of independent change. The
architecture taxonomy grows from eight document types to nine.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| One flexible Component Blueprint with kind-gated optional sections | Keeps the taxonomy small<br>No "which type do I use?" decision | Overloads one type across three very different altitudes<br>A bead loading a component doc still pulls composite/versioning sections irrelevant to an atomic change<br>The "component" superset obscures the C4 vocabulary | The context-size and clarity wins of a dedicated type outweigh the cost of one more type |
| Status quo — extract a component blueprint only when "sophisticated" or shared | Fewest documents | Most units stay embedded in the container blueprint, so a bead must load the whole container to change one unit<br>No home for composite structure once it is removed from the container | Fails the isolated-agent context-size goal |
| Add a Module Blueprint; component becomes atomic | Small, single-purpose docs per unit of change<br>Each type carries only its relevant concerns<br>C4 "component" is truly atomic again | Taxonomy 8→9<br>A component-vs-module boundary decision to make | N/A — selected |

## Consequences

**Positive:**
- Agents load minimal, relevant context per unit of change
- Each type carries only the concerns that apply to it (a component has no
  versioning section; a module can)
- The framework's "component" aligns with classic C4 (atomic), while composite
  structure and versioning get a clear home in the module type

**Negative:**
- One more document type in a deliberately small taxonomy
- Authors must decide component vs module (mitigated by the crisp boundary
  test in [[adr-005]])
- Existing single-component blueprints in adopting projects need migration
  to one type or the other

## References

- [[adr-005]] — the test for which type a unit is
- [[adr-006]] — how a module is classified
- [[adr-007]] — when to create either blueprint
- `references/c4-guide.md` — the abstraction levels, including the module extension
- Templates: `module-blueprint-template.md`, `component-blueprint-template.md`
