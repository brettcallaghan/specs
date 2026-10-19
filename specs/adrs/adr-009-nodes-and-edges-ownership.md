---
title: "ADR-009: Nodes-vs-Edges Ownership — Parents Own Child Edges"
description: File every relationship at one altitude; parents own the edges among their children, and leaves stay light on edges and code
slug: adr-009
type: adr
status: active
---

# ADR-009: Nodes-vs-Edges Ownership — Parents Own Child Edges

**Status:** active  
**Date:** 2026-07-29

## Context

Splitting architectural elements across more, smaller documents
([[adr-004]], [[adr-007]]) risked fragmenting the relationship map. If each
component document described its own collaborators, the view of how the
system fits together would be scattered across many files. Each copy would
also drift as it aged independently. We needed one unambiguous rule for
where relationships live.

## Decision

We will file every relationship at exactly one altitude. A **parent owns the
edges among its children**: a system owns the relationships between its
containers, a container owns those among its components, and a module owns
those among its internal components. An element documents its **own outward
edges only through a Context view** (System Context, Container Context).
Components and modules deliberately omit a context view and point to their
host containers instead. A shared module has different neighbors in every
host, so no single list of outward edges is correct. Each document is also
**light on the level below**: a component or module points to code as the
source of truth for the collaborators an agent discovers while working.

**This rule governs neighbor edges, which are the hierarchical composition
relationships, and not contract edges.** A neighbor relationship varies by
host, which is why the parent owns it. A contract doesn't vary: it's a named
document with one identity, so "this element provides it" and "this element
depends on it" hold wherever the element is composed. Therefore every
blueprint at all four levels carries its own `contracts_consumed`, and a
contract names its `provider`. A module or component that declares a
consumed contract isn't an exception to this decision.

**Nor does it govern the parent-child membership edge itself.** "A container
composes this module" is one fact with one authoritative home, and that home
depends on how many parents the child can have. Every C4 element has exactly
one parent: a container declares its `system`, and a component is declared
in its parent's roster section, where a reader meets it. A module is this
framework's addition to C4, and it's the one element composed into several
parents. Each parent lists the module in its own `modules`, because adopting
a module changes the host, not the module. The module's set of hosts is
derived from those lists. This decision is about the edges *among* a
parent's children, and the parent declares the membership edge too. A
hand-maintained host list on the module would be a second copy of an edge
that each parent already declares, and two copies is the condition this
decision exists to prevent.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Duplicate a component's immediate collaborators into its own document for standalone readability | Each document reads on its own | Duplicated edges drift as copies age<br>A shared module has different neighbors in each host, so no single list is correct<br>More to maintain | The drift and ambiguity outweigh the gain in standalone readability |
| Parents own child edges;<br>leaves stay light on outward edges and code | The relationship map lives in one place<br>Minimal duplication and drift<br>Code stays the source of truth for collaborators | A component document isn't fully standalone:<br>a reader follows one link to the container for relationships | N/A — selected |

## Consequences

**Positive:**
- The relationship map isn't fragmented: it lives with the parent
- Minimal duplication means minimal drift
- Code remains authoritative for a unit's collaborators, and agents read code
  anyway

**Negative:**
- A component or module document isn't self-contained for relationships. It
  points to its host container. This cost is accepted because the link is one
  hop and the alternative drifts.

## References

- [[adr-004]] — the split that raised the fragmentation risk
- [[adr-007]] — the granularity this principle keeps coherent
- `references/c4-guide.md` — Nodes and Edges: Who Owns What
