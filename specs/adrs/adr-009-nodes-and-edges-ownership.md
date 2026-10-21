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

Architectural elements are split across more, smaller documents
([[adr-004]], [[adr-007]]), which risks fragmenting the relationship map. If
each component document describes its own collaborators, the "how it all fits
together" view is scattered across many files and drifts as each copy ages
independently. We need a single, unambiguous rule for where relationships live.

## Decision

We will file every relationship at exactly one altitude. A **parent owns the
edges among its children**: a system owns the relationships between its
containers, a container among its components, a module among its internal
components. An element documents its **own outward edges only through a Context
view** (System Context, Container Context). Components and modules
deliberately omit a context view and point to their host containers instead: a
shared module has different neighbors in every host, so no single outward-edge
list is correct. Each document is also **light on the level below**: a component
or module points to code as the source of truth for the collaborators an agent
discovers while working.

**This rule governs neighbor edges — the hierarchical composition
relationships — and not contract edges.** A neighbor relationship varies by
host, and that is the whole reason the parent owns it. A Contract doesn't vary:
it's a named document with one identity, so "this element provides it" and
"this element depends on it" hold wherever the element is composed. Every
blueprint at all four levels therefore carries its own `contracts_consumed`,
and a Contract names its `provider`. A module or component that declares a
consumed Contract isn't an exception to this decision.

**Nor does it govern the parent-child membership edge itself.** "A
container composes this module" is one fact with one authoritative home,
and that home depends on how many parents the child can have. Every C4
element has exactly one: a container declares its `system`, and a component is
declared in its parent's roster section, where a reader meets it. A module is
this framework's addition to C4, and it's the one element composed into several
parents. Each parent lists it in its own `modules`, because adopting a module
changes the host, not the module. The module's set of hosts is derived from
those lists. The parent owns the edges *among* its children, which is what this
decision is about, and the membership edge is declared on the parent too. A
hand-maintained host list on the module would be a second copy of an edge each
parent already declares. Two copies are the condition this decision exists to
prevent.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Duplicate a component's immediate collaborators into its own document for standalone readability | Each document reads on its own | Duplicated edges drift as copies age<br>A shared module has different neighbors per host, so there is no single correct list<br>More to maintain | The drift and ambiguity outweigh the standalone-readability gain |
| Parents own child edges;<br>leaves stay light on outward edges and code | The relationship map lives in one place<br>Minimal duplication and drift<br>Code stays the source of truth for collaborators | A component document isn't fully standalone:<br>a reader follows one link to the container for relationships | N/A — selected |

## Consequences

**Positive:**
- The relationship map is never fragmented: it lives with the parent
- Minimal duplication means minimal drift
- Code remains authoritative for a unit's collaborators, which agents read
  anyway

**Negative:**
- A component or module document isn't self-contained for relationships: it
  points to its host container (accepted, because the link is one hop and the
  alternative drifts)

## References

- [[adr-004]] — the split that raises the fragmentation risk
- [[adr-007]] — the granularity this principle keeps coherent
- [[c4-standards#Nodes and Edges: Who Owns What]]
