<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/guides/c4-guide.md
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: C4 Guide
description: How to map software to the C4 model's levels, name elements,
  draw relationships, and choose and author C4 diagrams, including in
  Mermaid
slug: guide-c4
type: guide
status: active
---

# C4 guide

## Contents

- The abstraction levels — system, container, component, code
- Naming and relationship conventions — acronyms, arrow direction, message
  buses and async
- Diagram types — the standard set, plus the container context and
  container deployment views
- Mermaid authoring rules
- Mapping concrete elements to C4 — where queues, shared libraries, IaC,
  sidecars and data products belong
- Model as element, or mention in prose?
- References

The [C4 model](https://c4model.com) describes software at four nested levels
of abstraction and draws each as a diagram. This guide answers one question
quickly: **"where does this thing go?"** The answer names the C4 level a
concrete element maps to, and whether it belongs on a diagram at all.

Diagrams are **views of a model**. The model is the set of elements and
relationships the architecture has. Every diagram is a projection of that
model, scoped and filtered for one audience. A diagram SHALL NOT show an
element or a relationship that the model doesn't have.

---

## The abstraction levels

Each level has a one-line definition. When a mapping is disputed, argue from
the definition.

**Software system** — the highest level: a set of containers built and run
as a unit that delivers capabilities to users or other systems.

- System boundaries SHOULD be **domain-centric**: a system bundles related
  capabilities within one business domain or subdomain. Organizations that
  assign team ownership along domain lines arrive at the same boundaries
  indirectly: one system is the containers that one team builds and runs as
  a unit.
- In practice, accommodate system boundaries however the enterprise has
  drawn them, and document the architecture as it is.
- Real systems accumulate history. A container that doesn't belong in its
  system (by the domain lens above) stays where it actually runs. Document
  the misplacement: mark it "not for external use" and name the system it
  should belong to.

**Container** — an application or data store that must be *running* for the
system to work: API server, web app, worker, database, or queue. It isn't a
Docker container. The container is the unit of deployment. Components inside
it share its process space.

- Deployment is a separate concern: the same container may deploy
  differently per environment. Deployment diagrams map containers onto
  infrastructure. The container abstraction itself stays topology-free.

**Component** — a grouping of related functionality behind a well-defined
interface. It lives inside a container and isn't separately deployable.

**Code** — classes, functions, and files. The code and its inline
documentation are the source of truth at this level, and C4 recommends
drawing it only when a tool generates the diagram.

---

## Naming and relationship conventions

C4 leaves several conventions to its adopters. These are the opinionated
ones worth adopting. They govern how nodes are named and how edges are drawn
in every diagram type.

### Acronyms

- Element names SHALL NOT use **internal acronyms** — abbreviations
  meaningful only inside one team or enterprise.
- Only common, industry-standard acronyms (API, SDK, HTTP, SQL) are allowed
  in a name.
- An internal acronym MAY appear in the element's **description**, or in
  **parentheses after the name** WHERE there is room — never as the name
  itself.

**Rationale:** Newcomers and outsiders read diagrams. An internal acronym on
a box is a decode step they can't perform.

### Relationship direction

A relationship (arrow) tells a story about the most important dependencies
or data flow, and points in **one** direction.

- `<source-title> -> <target-title>` SHALL read like a sentence: "My Web App
  requests widgets from My Headless API."
- The arrow SHALL point in the direction of the **dependency**: client ->
  server, consumer -> publisher.
- EXCEPTION: in a Dynamic diagram the arrow MAY point in the direction of
  data flow or process sequence instead.

### Message buses and async

- A message bus (broker, queue, topic) SHALL NOT appear as an element UNLESS
  the bus itself is the subject of the diagram.
  - **Rationale:** A diagram where many containers point at one centralized
    bus is incomprehensible: the bus becomes a hub that hides the actual
    dependencies instead of revealing them.
- Instead, draw the relationship from the **consumer to the producer**
  (consumer -> publisher, per the dependency rule above), as if the bus were
  not there.
- WHERE space allows, name the transport technology and/or topic in the
  relationship label.

---

## Diagram types

| Diagram | Shows |
|---------|-------|
| System Landscape | Many systems across an enterprise |
| System Context | One system, its users, neighboring systems |
| Container | All containers in one system |
| **Container Context** | One container and its direct neighbors only (defined below) |
| Component | Components inside one container |
| Deployment (per environment) | Containers mapped to infrastructure |
| **Container deployment view** | One container and its exclusive infrastructure (defined below) |
| Dynamic | One runtime flow, step-ordered |

### Container Context diagram

A **Container Context diagram** is a container-scoped context view: the
analogue of a System Context diagram, one level down. It contains:

- the subject container, centered;
- every container, system, or person with a **direct** relationship to it,
  in either direction — crossing system boundaries where the neighbor is
  external (an external neighbor MAY appear as its containing system when
  container-level detail adds nothing);
- nothing else. No transitive neighbors, no siblings that don't communicate
  with the subject.

It introduces no new model elements: it's a filtered view of the same model
the Container diagram projects. In Structurizr, it's a container view with
include filters scoped to the subject's afferent and efferent couplings. In
Mermaid, it's a `C4Container` diagram drawn to this rule.

### Container deployment view

The container deployment view is an optional, container-scoped filter of the
system's deployment model: the subject container plus only the
infrastructure nodes that exist exclusively for it (its Redis, its
dead-letter queue, its bucket). The system's per-environment deployment
diagrams remain authoritative for shared topology. This view defers to them
and SHALL NOT contradict them.

---

## Mermaid authoring rules

- Element and relationship declarations (`Component(...)`, `Container(...)`,
  `Rel(...)`, and friends) SHALL be written on a single line. Quoted strings
  in these declarations can't span lines, and wrapping them breaks the
  diagram. A line-length limit doesn't apply inside a diagram fence.
- Keep each element's description text short: it's the element's brief
  description, not a place for extra prose.
- Parse-check diagrams before finishing (RECOMMENDED). A broken C4 diagram
  reports `Expecting 'RBRACE', got 'EOF'` anchored at the end of the diagram,
  far from the offending line, so authoring blind is risky.
  - `mermaid.parse` needs a DOM even though it renders nothing. Import it
    only after assigning a jsdom `window`, `document`, `navigator`,
    `Element`, `HTMLElement` and `SVGElement` onto `globalThis`. Otherwise,
    every diagram fails with `DOMPurify.addHook is not a function`, valid or
    not. That failure is environmental. It isn't the mislocated syntax error
    described above, and the diagrams aren't at fault.
  - Any Mermaid-rendering preview works too and needs no setup.

---

## Mapping concrete elements to C4

| Concrete element | C4 mapping |
|------------------|-----------|
| Web application (SPA, server-rendered) | Container |
| API server / backend service | Container |
| Background worker / queue consumer | Container |
| Scheduled job runner / cron | Container |
| Mobile app | Container |
| CLI tool (incl. codegen tooling) | Container |
| Serverless function (or cohesive set) | Container |
| Database / cache / bucket you **use** | Container (a data store), owned by the container that uses it |
| Database engine you **implement** | A system of containers |
| Queue / topic | An annotated consumer → producer relationship by default; a Container (a data store) only when the bus is the diagram's subject |
| REST / GraphQL / gRPC API | The interface of a container, drawn as a relationship label |
| Shared library / SDK | A component in each container that uses it |
| Generated client / stubs | Not modeled: an output of codegen |
| Micro-frontend package | A component of the shell container that composes it |
| IaC module (Terraform, CDK, Helm chart) | Not on static diagrams: a deployment concern |
| Platform / control plane (IDP, operators) | A system owned by the platform team |
| Sidecar / mesh proxy | A deployment concern: on deployment diagrams only |
| Logging / observability platform | A system owned by its team, usually mentioned in prose rather than drawn |
| External SaaS (Stripe, Auth0) | An external system |

When an element doesn't fit, argue from the definitions above, and prefer
the level whose *boundary* matches the element's unit of change and
ownership. IF it still doesn't fit, THEN flag the gap rather than forcing
it — don't invent a new abstraction silently.

---

## Model as element, or mention in prose?

Not everything real deserves a box. Diagrams stay legible only if elements
are rationed.

An element SHALL be modeled (a box on a diagram) when **any** of:

- your team owns, builds, or operates it;
- it's a direct communication partner of the subject, and the relationship
  carries a contract;
- its presence or absence changes a decision a reader of the diagram would
  make.

An element SHOULD be mentioned in prose instead when **any** of:

- it's ambient — every container uses it the same way (logging platform,
  authN provider, service mesh, artifact registry). Ambient obligations
  belong in a cross-cutting standard that the architecture documents
  reference;
- it's a technology choice inside a container (ORM, module federation, web
  framework) rather than a runtime neighbor;
- it sits below the diagram's abstraction level (a load balancer on a
  component diagram, a helper class anywhere).

Two litmus tests apply. *"Would this box change any reader's decision?"* If
no, use prose. *"Would this box appear on nearly every diagram in the
organization?"* If yes, it's ambient: prose plus a standard.

---

## References

- [C4 model — abstractions](https://c4model.com/abstractions)
- [C4 model — diagrams](https://c4model.com/diagrams)
- [C4 model — queues and topics](https://c4model.com/abstractions/queues-and-topics)
- [Mermaid C4 diagrams](https://mermaid.js.org/syntax/c4.html)
