---
title: "ADR-009: Nodes-vs-Edges Ownership — Parents Own Child Edges"
description: File every relationship at one altitude; parents own the edges among their children, and leaves stay light on edges and code
slug: adr-009
type: adr
status: active
---

# ADR-009: Nodes-vs-edges ownership — parents own child edges

**Status:** active  
**Date:** 2026-07-29

## Context

Architectural elements are now split across more, smaller documents
([[adr-004]], [[adr-007]]), which risks fragmenting the relationship map. If
each component document describes its own collaborators, the view of how
everything fits together is scattered across many files, and each copy drifts
as it ages. We need a single, unambiguous rule for where relationships live.

## Decision

We will file every relationship at exactly one altitude. A **parent owns the
edges among its children**: a system owns the relationships between its
containers, a container owns them among its components, and a module owns them
among its internal components. An element documents its **own outward edges
only through a context view** (System Context or Container Context). Components
and modules don't have a context view; they point to their host containers
instead. A shared module has different neighbors in every host, so no single
list of outward edges is correct. Each document is also **light on the level
below**: a component or module points to code as the source of truth for the
collaborators that an agent discovers while working.

**This rule governs neighbor edges, which are the hierarchical composition
relationships. It doesn't govern contract edges.** A neighbor relationship
varies by host, which is why the parent owns it. A contract doesn't vary: it's
a named document with one identity, so "this element provides it" and "this
element depends on it" hold wherever the element is composed. Every blueprint
at all four levels therefore carries its own `contracts_consumed`, and a
contract names its `provider`. A module or component that declares a consumed
contract isn't an exception to this decision.

**This rule also doesn't govern the parent-child membership edge itself.** "A
container composes this module" is one fact with one authoritative home, and
that home depends on how many parents the child can have. Every C4 element has
exactly one parent: a container declares its `system`, and a component is
declared in its parent's roster section, where a reader meets it. A module is
this framework's addition to C4, and it's the one element that can be composed
into several parents. Each parent lists the module in its own `modules`,
because adopting a module changes the host, not the module, and the module's
set of hosts is derived from those lists. The parent owns the edges *among* its
children, which is what this decision is about, and the parent declares the
membership edge too. A hand-maintained host list on the module would be a
second copy of an edge that each parent already declares, and two copies is
the condition that this decision exists to prevent.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Duplicate a component's immediate collaborators into its own document for standalone readability | Each document reads on its own | Duplicated edges drift as copies age<br>A shared module has different neighbors per host, so no single list is correct<br>More to maintain | The drift and ambiguity outweigh the gain in standalone readability |
| Parents own child edges;<br>leaves stay light on outward edges and code | The relationship map lives in one place<br>Minimal duplication and drift<br>Code stays the source of truth for collaborators | A component document isn't fully standalone:<br>a reader follows one link to the container for relationships | N/A—selected |

## Consequences

**Positive:**

- The relationship map isn't fragmented; it lives with the parent.
- Minimal duplication means minimal drift.
- Code remains authoritative for a unit's collaborators, which agents read
  anyway.

**Negative:**

- A component or module document isn't self-contained for relationships; it
  points to its host container. This is accepted because the link is one hop
  and the alternative drifts.

## References

- [[adr-004]]: the split that raised the fragmentation risk.
- [[adr-007]]: the granularity that this principle keeps coherent.
- `references/c4-guide.md`: the "Nodes and Edges: Who Owns What" section.
