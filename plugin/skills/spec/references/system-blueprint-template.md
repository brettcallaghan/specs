---
title: "System Blueprint: <System Name>"
entity_name: <System Name>                 # the element, as diagrams draw it
technology: <language, framework or package type>   # OPTIONAL; shown on diagrams.
                                     # omit WHEN not meaningful
description: <One-sentence summary of the system and the capabilities it
  bundles — its brief description, shown on diagrams>
slug: system-<system-name>
type: system-blueprint
status: draft
updated_at: <YYYY-MM-DDTHH:MM:SSZ>
contracts_consumed: []               # contract slugs this system consumes
---

<!--
Template: System Blueprint
Version: 1.6
Source: Architecture documentation framework. C4 mapping rules are
  defined in `shared/c4-standards.md` — load it before authoring.

Sections (complete each, then delete its instruction block)
  - Overview
  - System Context
  - Containers
  - Public Contracts
  - Consumed Systems and Contracts
  - Deployment
  - Cross-Cutting Standards
  - Change Guidance
  - Key Decisions
  - Open Questions
  - References

Purpose
  Document one software system: the related capabilities it bundles, its
  context among neighboring systems and actors, the containers it is made
  of, the public contracts it provides, and how it deploys. This is the
  top-level architecture document for a system and the entry point for
  readers.

When to use
  - Documenting a system your team owns (one blueprint per system)
  - A capability has runtime elements of its own — it IS a system (or part
    of one); document it here, not in a component blueprint

When NOT to use
  - A single deployable unit inside a system → container blueprint
  - A shared library, package, or complex component → component blueprint
  - An interface's shape and semantics → `contract-template.md`
  - Recording a decision's trade-offs → ADR (reference it from Key
    Decisions below)

System boundary
  System boundaries SHOULD be domain-centric: a system bundles related
  capabilities within one business domain or subdomain. Team ownership
  drawn along domain lines yields the same boundaries indirectly. In
  practice, accommodate boundaries however the enterprise has drawn
  them — this blueprint documents the architecture as it is. WHERE
  history has left a container in a system it does not belong to, keep
  it in this blueprint and document the misplacement in the Containers
  section: mark it "not for external use" and name the system it
  should belong to. See `shared/c4-standards.md`.

Diagrams
  Diagrams are views rendered from the element declarations in this
  document and the blueprints it links — declare first, then draw. Embed
  Mermaid C4 diagrams (C4Context, C4Container, C4Deployment) or link
  rendered views from your modeling tool. Every element on a diagram
  SHALL be declared in prose or frontmatter; see `shared/c4-standards.md` for the
  declaration format and for what to model as an element vs mention in
  prose.

Slug format
  system-<system-name>, lowercase kebab-case. Example: system-notifications

Status vocabulary
  draft | ready-for-review | approved | active | retired

This is a living document. Update it when the architecture changes — a
system blueprint describes the system as it is, not as it was planned.
Remove all HTML comment blocks after completing each section. WHILE status
is draft, the Open Questions section MAY contain open questions or the
literal text "None."; remove the section when the document becomes active.
-->

# <System Name>

## Overview

<!--
1-2 paragraphs: the related capabilities this system bundles, who it
serves, and why it exists as a distinct system. Name the owning team. Do
not describe internal structure here — that is what the sections below
are for.
-->

<What this system does, for whom, and who owns it>

## System Context

<!--
A C4 System Context diagram: this system, the people who use it, and the
neighboring systems it exchanges data with. Follow the diagram with one
short prose paragraph per external relationship: direction, what crosses
the boundary, and why. Reference the governing contract with [[slug]] where
one exists (or the vendor's published contract URL for third-party
systems).
-->

```mermaid
C4Context
    title System Context — <System Name>
    Person(user, "<Actor>", "<role>")
    System(this, "<System Name>", "<brief description>")
    System_Ext(neighbor, "<Neighboring System>", "<brief description>")
    Rel(user, this, "<uses, how>")
    Rel(this, neighbor, "<calls / publishes>", "<protocol>")
```

<One paragraph per relationship: who depends on whom, what crosses the
boundary, which [[contract-slug]] governs it, and why the boundary is here.>

## Containers

<!--
The containers are DECLARED in the list below; the C4 Container diagram
is a view rendered from those declarations (see `shared/c4-standards.md`). Each
declaration captures: the container's name (canonical — used verbatim on
diagrams), its technology, and a brief description that doubles as the
diagram's description text.

The detailed description of an active container lives in its
[[container-<name>]] blueprint — do not restate internals here. Passive
containers (data stores, buckets, queues this system owns but did not
implement) have no blueprint of their own: give their detailed
description here (what they hold, which container owns the schema).
Their published surface, if any, is a contract.

Mark any container that exists here for historical reasons: state that
it SHALL NOT be consumed from outside the system, and name the system it
should belong to.
-->

```mermaid
C4Container
    title Containers — <System Name>
    Container(api, "<API Server>", "<technology>", "<brief description>")
    ContainerDb(db, "<Database>", "<technology>", "<brief description>")
    Rel(api, db, "reads/writes", "<protocol>")
```

- **<Container Name>** ([[container-<name>]]; <technology>) — <brief
  description, as it appears on diagrams>.
- **<Data Store Name>** (<technology>) — <brief description>. <Detailed
  description: what it holds, which container owns its schema.>

## Public Contracts

<!--
The interfaces this system provides for consumers OUTSIDE the system —
its integration surface. One line per contract, linking the [[contract-slug]].
Shape and semantics live in the contract documents; do not restate them.

Interfaces used only between this system's own containers are internal
contracts — list them in the providing container's blueprint, not here.
-->

- [[contract-<name>]] — <what it vends, to whom, via what mechanism>.

## Consumed Systems and Contracts

<!--
What this system depends on: neighboring systems, platform services, and
the specific contracts consumed. One line each. Cross-cutting platform
obligations (logging, authN, mesh) are standards, not dependencies — see
the next section.

Third-party (SaaS) interfaces are governed by the vendor's published
contract: reference it by URL. contract documents exist only for
interfaces the organization provides — SHALL NOT be authored for
third-party interfaces.
-->

- [[contract-<name>]] (provided by <system>) — <what we use it for>.
- <Vendor / SaaS product> — <what we use it for>; contract:
  <URL to the vendor's published specification>.

## Deployment

<!--
The authoritative deployment views for this system, one C4 Deployment
diagram per environment WHERE topology differs across environments.
Container blueprints MAY carry a container-scoped deployment view; those
are filtered views of the model shown here and defer to this section on
shared topology.
-->

### <Environment>

```mermaid
C4Deployment
    title Deployment — <System Name> (<environment>)
    Deployment_Node(cloud, "<Region / Account>", "<provider>") {
        Deployment_Node(cluster, "<Cluster / Runtime>", "<tech>") {
            ContainerInstance(api)
        }
    }
```

## Cross-Cutting Standards

<!--
The standards and platform obligations this system complies with, by
reference. Never restate a standard's requirements here.
-->

This system SHALL comply with:

- [[<standard-slug>]] — <one-line scope of the standard>.

## Change Guidance

<!--
Rules for feature work touching this system as a whole: the constraints
a feature blueprint — and the planning agent downstream of it — must
respect when a change spans containers. Write normative rules in EARS.

This section carries SYSTEM-LEVEL architectural constraints: invariants
that hold across containers, cross-container lockstep-update rules, and
boundary rules for what may talk to what. Constraints scoped to a single
container belong in that [[container-<name>]] blueprint's Change Guidance,
not here; organization-wide rules belong in a cross-cutting requirements
document referenced from Cross-Cutting Standards.

WHEN prose anywhere in this blueprint mentions a known defect or other
open tracker issue, it SHOULD cite the tracker ID in greppable form
(e.g. `proj-ab12`) — feature-blueprint authors harvest these IDs when
reconciling their deltas against the tracker.

End with a Verification gates entry naming the system-level suites
(integration, end-to-end, contract tests) that gate cross-container
changes. Feature blueprints map their AC Coverage "Verified by" column
onto these gates.
-->

- WHEN <a change spans containers>, <the artifacts that must change
  together> SHALL be updated in the same change.
- <A cross-container boundary or invariant rule any change must respect,
  in EARS.>
- **Verification gates:** <the system-level integration / end-to-end /
  contract-test suites that gate cross-container changes>.

## Key Decisions

<!--
One line per significant decision: restate the outcome, reference the ADR.
The trade-off analysis lives in the ADR, not here.
-->

- [[adr-<nnn>]] — <one-line statement of the decision>.

## Open Questions

<!--
Draft-mode only. Track unresolved architecture questions here while the
blueprint is in draft. WHEN a question is resolved, record the decision as
an ADR (or fold the answer into the relevant section) and delete it from
this list. WHILE status is draft, the section MAY instead state the
literal text "None." (no checkbox) — the absence of questions is
information. Remove the whole section when status becomes active.
-->

- [ ] <open question> — <what is blocked on it, who decides>

## References

- `shared/c4-standards.md` — C4 mapping rules and diagram conventions
- <Additional references>
