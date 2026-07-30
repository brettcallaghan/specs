---
title: "ADR-009: Nodes-vs-Edges Ownership — Parents Own Child Edges"
description: File every relationship at one altitude; parents own the edges among their children, and leaves stay light on edges and code
slug: adr-009-nodes-and-edges-ownership
type: adr
status: active
updated_at: 2026-07-29T00:00:00Z
---

# ADR-009: Nodes-vs-Edges Ownership — Parents Own Child Edges

**Status:** active  
**Date:** 2026-07-29

## Context

With architectural elements split across more, smaller documents
([[adr-004]],
[[adr-007]]), we risked fragmenting the
relationship map. If each component document described its own collaborators,
the "how it all fits together" view would be scattered across many files and
would drift as each copy aged independently. We needed a single, unambiguous
rule for where relationships live.

## Decision

We will file every relationship at exactly one altitude. A **parent owns the
edges among its children**: a system owns the relationships between its
containers, a container among its components, a module among its internal
components. An element documents its **own outward edges only through a Context
view** (System Context, Container Context). Components and modules
deliberately omit a context view and point to their host container(s) instead —
a shared module has different neighbors in every host, so no single outward-edge
list is correct. Each document is also **light on the level below**: a component
or module points to code as the source of truth for the collaborators an agent
discovers while working.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Duplicate a component's immediate collaborators into its own document for standalone readability | Each document reads on its own | Duplicated edges drift as copies age<br>A shared module has different neighbors per host, so there is no single correct list<br>More to maintain | The drift and ambiguity outweigh the standalone-readability gain |
| Parents own child edges;<br>leaves stay light on outward edges and code | The relationship map lives in one place<br>Minimal duplication and drift<br>Code stays the source of truth for collaborators | A component document is not fully standalone —<br>a reader follows one link to the container for relationships | N/A — selected |

## Consequences

**Positive:**
- The relationship map is never fragmented — it lives with the parent
- Minimal duplication means minimal drift
- Code remains authoritative for a unit's collaborators, which agents read
  anyway

**Negative:**
- A component or module document is not self-contained for relationships; it
  points to its host container (accepted, because the link is one hop and the
  alternative drifts)

## References

- [[adr-004]] — the split that raised the
  fragmentation risk
- [[adr-007]] — the granularity this principle
  keeps coherent
- `references/c4-guide.md` — Nodes and Edges: Who Owns What
