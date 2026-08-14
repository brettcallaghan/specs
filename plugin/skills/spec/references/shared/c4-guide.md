<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/standards/c4-guide.md (version 0.2)
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Mapping Architecture to C4
description: How to map concrete architectural elements to C4 levels, which
  document and diagram each belongs in, and when to model an element vs
  mention it in prose
slug: c4-standards
type: standards
status: draft
updated_at: 2026-07-29T00:00:00Z
version: "0.2"
---

# Mapping Architecture to C4

## Contents

- The Abstraction Levels — system, container, module, component
- Element Declarations: Model in Prose, Views in Diagrams
- Nodes and Edges: Who Owns What
- Naming and Relationship Conventions — acronyms, arrow direction, message
  buses and async
- Diagram Types and Placement — container context, deployment view, and the
  Mermaid authoring rules
- Mapping Concrete Elements to C4 — where queues, shared libraries, IaC
  modules, sidecars and data products belong
- Model as Element, or Mention in Prose?
- Contract Visibility Tiers
- Lifecycle: Documentation Follows the Realization
- References

This guide defines how this framework uses the [C4 model](https://c4model.com)
in its architecture document types — system blueprint,
container blueprint, module blueprint, component blueprint, and
contract. The framework
is used across many projects and enterprises; the rules here are
conventions of the framework, not of any one codebase. It exists to answer one
question quickly: **"where does this thing go?"** — which C4 level a
concrete element maps to, which document describes it, and whether it
belongs on a diagram at all.

Diagrams are **views of a model**. The model is the set of elements and
relationships declared across the blueprint documents; every diagram is a
projection of that model, scoped and filtered for one audience. A diagram
SHALL NOT show an element or relationship that the documents do not
describe.

---

## The Abstraction Levels

One-line definitions, plus the rules this framework adds. (The definitions
are anchors for the mapping rules below — when a mapping is disputed,
argue from the definition.)

**Software system** — the highest level: a set of containers built and
run as a unit that delivers capabilities to users or other systems.

- System boundaries SHOULD be **domain-centric**: a system bundles
  related capabilities within one business domain or subdomain.
  Organizations that assign team ownership along domain lines arrive at
  the same boundaries indirectly — one system = the containers one team
  builds and runs as a unit.
- In practice, accommodate system boundaries however the enterprise has
  drawn them. Blueprints document the architecture as it is.
- Real systems accumulate history. A container that does not belong in
  its system (by the domain lens above) stays where it actually runs,
  and the system blueprint documents the misplacement: mark it "not for
  external use" and name the system it should belong to.

**Container** — an application or data store that must be *running* for
the system to work: API server, web app, worker, database, queue. Not a
Docker container. The container is the unit of deployment; components
inside it share its process space.

- Deployment is a separate concern: the same container may deploy
  differently per environment. Deployment diagrams map containers onto
  infrastructure; the container abstraction itself stays topology-free.
- A repository MAY document standalone containers without a system
  blueprint: omit `system` in the container blueprint. The system
  level is introduced when a second container needs shared context worth
  modeling. WHEN there is no system blueprint, the container blueprint's
  Infrastructure and Deployment section is authoritative for that
  container's deployment.

**Component** — C4 defines a component as a grouping of related
functionality behind a well-defined interface, living inside a
container, not separately deployable. In this framework a component is
the **atomic leaf**: an internal building block of a SINGLE container,
with a defined interface and no internal components of its own. It
changes in lockstep with its container — it is not vended or versioned
independently. Its home is a component blueprint (or, un-promoted, a
record block in its container's Components section).

**Module** — this framework's named level between container and
component, and the deliberate extension of strict C4:

> A module is a unit of software with a defined interface, composed into
> containers at build or deploy time, or living within a container at
> runtime — that crosses a boundary of independent change (it is shared
> or published across containers) and/or has internal components of its
> own (one level of nesting).

The extension admits build-time and deploy-time units — shared
libraries, SDKs, UI packages, micro-frontends, IaC modules, plugins. A
module is composed INTO a container; it is not separately deployable and
owns no runtime data. Its **visibility** — `internal` (unpublished,
evolves in the repo in lockstep with consumers) or `public` (published
to a registry, consumers span versions independently) — governs the
change rules for work touching it. Its home is a module blueprint. This
paragraph is the extension's definition of record.

**Code** — classes, functions, files. Not documented in blueprints; the
code and its inline documentation are the source of truth.

---

## Element Declarations: Model in Prose, Views in Diagrams

This framework separates the model from its views without a modeling
DSL — it deliberately stops short of Structurizr or fenced declaration
blocks. Prose is the declaration format: every modeled element is
declared once, in its home document (see the mapping table below), and
every diagram is rendered from declarations that already exist.

Each declaration SHALL capture these facts:

| Fact | Purpose |
|------|---------|
| **Name** | Canonical; used verbatim on every diagram and in references |
| **Technology** | For containers and components: language, framework, store or package type |
| **Brief description** | One line; doubles as the element's description text on diagrams |
| **Detailed description** | The fuller prose a reader of the document needs; never rendered on diagrams |

Elements are referenced by document link or heading anchor. A promoted
element (one with its own blueprint) is referenced by its blueprint slug
(`[[component-<name>]]` / `[[module-<name>]]`). An un-promoted element is
referenced by a wikilink heading anchor into its parent's roster —
`[[container-<name>#Element Name]]` — where each element is declared under
its own `### <Name>` heading in the parent's Components section. The
roster heading persists after promotion (it then carries the blueprint
link in place of the record block), so an anchor reference keeps
resolving without an immediate rewrite when an element is promoted.
Diagrams always use the display name. A code-level element that is NOT a
modeled component — a schema, domain type, event, flag, or table — is
written in backticks (`ElementName`), not referenced this way.

WHERE an element has its own blueprint, that blueprint IS the
declaration: the document title carries the name, the frontmatter
`description` is the brief description, and the Overview section is
the detailed description. Documents that reference the element repeat
only its name, technology, and brief description, linking the
blueprint for detail.

WHERE a child element has NOT been promoted to its own blueprint, its
declaration in the parent's roster (a container's or module's Components
section) carries a machine-readable **record block** instead of a
blueprint link — so every modeled node is declared in prose whether or
not it has its own document. The record block's format, its schema, and
the block-XOR-link promotion rule are defined in
[[template-authoring-standards]] (Embedded Element Records). Declaring every
node this way keeps the model machine-readable: the element vocabulary
and node set are parseable without a modeling DSL, and diagrams can be
rendered from the declarations.

WHEN a diagram and a declaration disagree, the declaration is
authoritative — fix the view.

---

## Nodes and Edges: Who Owns What

The model is a tree of nodes with edges between them, and the framework
files each fact at exactly one altitude:

- **A parent owns the edges among its children.** A system owns the
  relationships between its containers (the Containers section); a
  container owns the relationships among its components (the Components
  section); a module owns the relationships among its internal
  components. Edges live with the parent, declared once.
- **An element owns its own outward edges only through a Context view.**
  System Context and Container Context show the subject among its direct
  neighbors. Components and modules deliberately DO NOT get a context
  view: a component's neighbors are owned by its container, and a shared
  module has different neighbors in every host — so both point to their
  host container(s) for relationships rather than restating them.
- **Each element is light on the level below.** A container is light on
  component internals (the component/module blueprint owns them); a
  component or module is light on code (the code is the source of truth,
  and an agent reads it to discover collaborators).
- **Contract edges are the exception, and they belong to the element.**
  Every blueprint — at all four levels — carries `contracts_provided`
  and `contracts_consumed`. The two rules above govern *neighbor* edges,
  which have no single correct value for a shared element. A contract
  does: it is a named document with one identity, so "this element
  publishes it" and "this element depends on it" are true wherever the
  element is composed. A module MAY therefore declare a consumed
  contract even though it has no Context view.

So every document sits at its own altitude: it owns its children as
nodes and the edges among them, names the contracts it provides and
consumes, points UP to its parent for its remaining outward edges, and
points DOWN to code for internal detail.

---

## Naming and Relationship Conventions

C4 leaves several conventions to its adopters. This framework adopts the
opinionated ones below; they govern how nodes are named and how edges
are drawn, across every diagram type.

### Acronyms

- Element names SHALL NOT use **internal acronyms** — abbreviations
  meaningful only inside one team or enterprise.
- Only common, industry-standard acronyms (API, SDK, HTTP, SQL) are
  allowed in a name.
- An internal acronym MAY appear in the element's **description**, or in
  **parentheses after the name** WHERE there is room — never as the name
  itself.

RATIONALE: diagrams are read by newcomers and outsiders; an internal
acronym on a box is a decode step they cannot perform.

### Relationship Direction

A relationship (arrow) tells a story about the most important
dependencies or data flow, and points in **one** direction.

- `<source-title> -> <target-title>` SHALL read like a sentence: "My Web
  App requests widgets from My Headless API."
- The arrow SHALL point in the direction of the **dependency**: client
  -> server, consumer -> publisher.
- EXCEPTION: in a Dynamic diagram the arrow MAY point in the direction
  of data flow or process sequence instead.

### Message Buses and Async

- A message bus (broker, queue, topic) SHALL NOT appear as an element
  UNLESS the bus itself is the subject of the diagram.
  - RATIONALE: a diagram where many containers point at one centralized
    bus is incomprehensible — the bus becomes a hub that hides the
    actual dependencies rather than revealing them.
- Instead, draw the relationship from the **consumer to the producer**
  (consumer -> publisher, per the dependency rule above), as if the bus
  were not there.
- WHERE space allows, name the transport technology and/or topic in the
  relationship label.

This refines the "Queue / topic" row of the mapping table below: the
annotated consumer -> producer relationship is the default; modeling the
bus as its own Container is reserved for the case WHERE the bus is what
the diagram is about.

---

## Diagram Types and Placement

Brief on purpose — the value of this table is the placement column, not
the definitions.

| Diagram | Shows | Lives in |
|---------|-------|----------|
| System Context | One system, its users, neighboring systems | system blueprint |
| Container | All containers in one system | system blueprint |
| **Container Context** | One container + direct neighbors only (defined below) | container blueprint |
| Component | Components inside one container | container blueprint |
| Module components | Internal components inside one module | module blueprint |
| Deployment (per environment) | Containers mapped to infrastructure | system blueprint (authoritative) |
| Container deployment view | One container + its exclusive infra (defined below) | container blueprint (optional) |
| Dynamic | One runtime flow, step-ordered | Any blueprint, where a flow needs it |
| System Landscape | Many systems | Outside these documents (EA level) |

### Container Context Diagram

A **Container Context Diagram** is this framework's named convention for a
container-scoped context view — the analogue of a System Context
diagram, one level down. It contains:

- the subject container, centered;
- every container, system, or person with a **direct** relationship to
  it, in either direction — crossing system boundaries where the
  neighbor is external (an external neighbor MAY appear as its
  containing system when container-level detail adds nothing);
- nothing else. No transitive neighbors, no siblings that do not
  communicate with the subject.

It introduces no new model elements: it is a filtered view of the same
model the Container diagram projects, and every relationship shown SHALL
also exist in the documents (typically governed by a contract). In
Structurizr this is a container view with include filters scoped to the
subject's afferent and efferent couplings; in Mermaid, a `C4Container`
diagram drawn to this rule.

### Container deployment view

An optional, container-scoped filter of the system's deployment model:
the subject container plus only the infrastructure nodes that exist
exclusively for it (its Redis, its DLQ, its bucket). The system
blueprint's per-environment deployment diagrams remain authoritative for
shared topology; this view defers to them and SHALL NOT contradict them.

### Mermaid Authoring Rules

Rules for authoring the Mermaid C4 diagrams embedded in blueprints:

- Element and relationship declarations (`Component(...)`,
  `Container(...)`, `Rel(...)`, and friends) SHALL be written on a
  single line. Quoted strings in these declarations cannot span lines;
  wrapping them breaks the diagram. The 80-character line limit does
  not apply inside diagram fences ([[template-authoring-standards]]
  exempts code-block content).
- Keep each element's description text short — it is the element's
  declared brief description, verbatim, not a place for extra prose.
- Parse-check diagrams before finishing (RECOMMENDED). A broken C4
  diagram reports `Expecting 'RBRACE', got 'EOF'` anchored at the end
  of the diagram, far from the offending line, so authoring blind is
  risky.
  - `mermaid.parse` needs a DOM even though it renders nothing. Import
    it only after assigning a jsdom `window`, `document`, `navigator`,
    `Element`, `HTMLElement` and `SVGElement` onto `globalThis`;
    otherwise every diagram fails with `DOMPurify.addHook is not a
    function`, valid or not. That failure is environmental — it is NOT
    the mislocated syntax error described above, and the diagrams are
    not at fault.
  - Any Mermaid-rendering preview works too and needs no setup.

---

## Mapping Concrete Elements to C4

The table below maps recurring real-world elements to their C4 level and
home document. "Contract" in the last column means the interface gets a
contract document (see visibility tiers below).

| Concrete element | C4 mapping | Documented in |
|------------------|-----------|---------------|
| Web application (SPA, server-rendered) | Container | Its container blueprint |
| API server / backend service | Container | Its container blueprint |
| Background worker / queue consumer | Container | Its container blueprint |
| Scheduled job runner / cron | Container | Its container blueprint |
| Mobile app | Container | Its container blueprint |
| CLI tool (incl. codegen tooling) | Container | Its container blueprint |
| Serverless function (or cohesive set) | Container | Its container blueprint |
| Database / cache / bucket you **use** | Container (passive) | Element on owner's diagrams + Data section of owning container blueprint; **no blueprint of its own** |
| Database engine you **implement** | System of containers | Its own system blueprint |
| Queue / topic | Annotated consumer → producer relationship by default; Container (a data store) only WHEN the bus is the diagram's subject (see Message Buses and Async) | Message schema + semantics → contract |
| REST / GraphQL / gRPC API | Interface of a container | contract (shape in spec artifact) |
| Data product / published view / shared dataset | Interface of a container or system | contract |
| Shared library / SDK | **Module** | Its module blueprint; appears as a component in each consuming container's diagram |
| Generated client / stubs | Output of codegen; not modeled | The governing spec artifact belongs to the contract |
| UI component package / design-system package | **Module** | Its module blueprint |
| Design system's docs site / tokens API | Container(s) | Their container blueprint(s) |
| Micro-frontend package | **Module**, composed by the shell container | Its module blueprint |
| Module federation | Technology choice within containers | Prose in the affected container blueprints |
| IaC module (Terraform, CDK, Helm chart) | **Module** (deploy-time) | Its module blueprint |
| Platform / control plane (IDP, operators) | System — owned by the platform team | Platform team's system blueprint; your obligations → standards docs |
| Sidecar / mesh proxy | Deployment concern | Deployment diagrams + standards docs; **not** on container/component diagrams |
| Logging / observability platform | System — owned by its team | Consumers: prose mention + standards doc; **not** a diagram element on every system |
| Complex internal component, composite (rules engine) | **Module** (internal, composite) | Its module blueprint |
| Complex internal component, atomic (pricing calculator) | Component | Its component blueprint |
| Repeated unit built to one design (plugins, transpilers) | **Module** or Component (pattern) | One blueprint governs all instances; module WHEN shared/composite |
| External SaaS (Stripe, Auth0) | External system | Element on context diagrams; reference the vendor's published contract by URL — do NOT author contract documents for third-party interfaces |

When an element does not fit: argue from the definitions above, and
prefer the level whose *boundary* matches the element's unit of change
and ownership. IF it still does not fit, THEN flag the gap rather than
forcing it — do not invent a new abstraction silently.

---

## Model as Element, or Mention in Prose?

Not everything real deserves a box. Diagrams stay legible only if
elements are rationed.

An element SHALL be modeled (a box on a diagram, declared in a
document) when **any** of:

- your team owns, builds, or operates it;
- it is a direct communication partner of the subject, and the
  relationship carries a contract;
- its presence or absence changes a decision a reader of that document
  would make.

An element SHOULD be mentioned in prose (or covered by a standards
document) instead when **any** of:

- it is ambient — every container uses it the same way (logging
  platform, authN provider, service mesh, artifact registry). Ambient
  obligations belong in a cross-cutting requirements document that
  blueprints reference;
- it is a technology choice inside a container (ORM, module federation,
  web framework) rather than a runtime neighbor;
- it sits below the document's abstraction level (a load balancer on a
  component diagram, a helper class anywhere).

Litmus tests: *"Would this box change any reader's decision?"* — if no,
prose. *"Would this box appear on nearly every diagram in the
organization?"* — if yes, it is ambient: prose plus a standard, and its
own system blueprint owned by its team.

---

## Contract Visibility Tiers

Every interface between elements has a contract; only some contracts
get standalone documents.

| Tier | Boundary crossed | Documented as |
|------|------------------|---------------|
| **public** | The system boundary — consumed by other teams | contract document; strongest stability, versioning, deprecation policy |
| **internal** | Container boundaries within one system | contract document, marked internal; SHALL NOT be consumed outside the system |
| **private** | A container and infrastructure it exclusively owns (its DB schema, its cache keys), or in-process | Inline in the container blueprint's Data section; **no standalone document** |

The filing rule: a contract gets a standalone document WHEN the
interface crosses a boundary of **independent change** — two codebases
or teams evolving against it on their own schedules. A private schema is
still a contract, but its two sides change in lockstep in one codebase,
so it is documented where it lives.

The same three-tier spectrum classifies composed units, not just
contracts — the framework applies one rule of independent change across
both:

| Tier | Contract | Composed unit |
|------|----------|---------------|
| **private** | inline in the container's Data section | a **component** — internal to one container, lockstep |
| **internal** | contract, marked internal | a **module**, `visibility: internal` — shared in-repo, unpublished |
| **public** | contract, public | a **module**, `visibility: public` — published, consumers independent |

A component maturing into a shared-then-published module tracks the same
path a private schema takes when promoted to an internal then public
contract: each step widens the boundary of independent change and
tightens the stability obligations.

WHEN a private interface gains a second independently-built consumer,
THEN it SHALL be promoted (to internal or public) and extracted into a
contract document. Publishing a view over private data is such a
promotion: the view becomes a contract; the underlying schema stays
private.

The machine-readable spec artifact (OpenAPI, SDL, proto, DDL) is
normative for **shape**; the contract document is normative for
**semantics** (idempotency, ordering, SLOs, versioning policy). Both
change under the same versioning policy.

Contract documents are authored only for interfaces the organization
provides. Third-party interfaces (external SaaS APIs) are governed by
the vendor's published contract — architecture documents SHALL reference
those by URL and SHALL NOT restate them as contract documents.

---

## Lifecycle: Documentation Follows the Realization

Capabilities migrate between vending mechanisms as they mature. The
document type follows the structure, not the other way around:

- A component becomes shared across containers, or grows internal
  components → it has become a module: author a module blueprint and
  fold the component's content into it (superseded_by the module).
- An internal module is published to a registry → its visibility becomes
  public: backwards-compat, deprecation windows, and recorded change
  history now apply.
- A module gains a worker and a queue → it has become a container (or
  system): author the new blueprint, retire the module blueprint
  (superseded_by the successor). Its contract usually survives the move
  unchanged — which is why contracts are separate documents.
- An internal contract gains an external consumer → promote to public;
  the stability obligations tighten.
- A module stops being shared and has no internal components → its
  blueprint MAY be folded back into the single consuming container and
  retired (it has become a component).

WHEN the architecture changes, the affected blueprints, contracts, and
diagrams SHALL be updated in the same change — these are living
documents describing the system as it is.

---

## References

- [[template-authoring-standards]] — frontmatter, slugs, EARS, references
- [C4 model — abstractions](https://c4model.com/abstractions)
- [C4 model — queues and topics](https://c4model.com/abstractions/queues-and-topics)
