---
title: "Contract: <Interface Name>"
description: <One-sentence summary of the interface and the capability it
  vends>
slug: contract-<interface-name>
type: contract
status: draft
updated_at: <YYYY-MM-DDTHH:MM:SSZ>
provider: <container-or-system slug that provides this interface>
visibility: <public | internal>
consumers: []                        # known consumer slugs (containers/systems)
specs: []                            # paths/URLs of machine-readable artifacts
version: "1.0"
related: []                          # related documents (ADRs, standards)
---

<!--
Template: Contract
Version: 1.2
Source: Architecture documentation framework. C4 mapping rules and
  contract visibility tiers are defined in `c4-guide.md`.

Purpose
  The prose companion to a machine-readable interface specification. The
  spec artifact (OpenAPI, GraphQL SDL, protobuf, JSON Schema, Avro
  subject, SQL DDL for a published view) is normative for SHAPE; this
  document is normative for SEMANTICS — the guarantees, operational
  expectations, and change policy that schema languages cannot express.

Ownership
  A contract is owned by its PROVIDER, never by a relationship or a
  consumer. Consumers reference it; exactly one document governs the
  interface no matter how many consumers exist.

Independence and longevity
  A contract is deliberately separate from any blueprint because it
  outlives, and may precede, the implementations that realize it. A
  blueprint describes one design of a system, container, or module as it
  IS; a contract describes an interface that MAY be implemented many
  times over its life — re-implemented for modernization, or served by
  more than one provider for backwards compatibility. It MAY also be
  authored contract-first, before any provider exists — often the most
  consequential architectural decision. WHEN no provider exists yet,
  leave provider a draft slug and keep status: draft.

Visibility (see `c4-guide.md`)
  public   — crosses the system boundary; consumable by other teams.
  internal — crosses container boundaries within one system only.
             Internal contracts SHALL NOT be consumed outside their
             system.
  private  — a container's contract with infrastructure it exclusively
             owns (its database schema). Private contracts DO NOT get
             this document — they are documented inline in the container
             blueprint's Data section. WHEN a second independently-built
             consumer appears, promote to internal or public and author
             this document.

When to use
  - An interface crosses a boundary of independent change: two codebases
    or teams evolve against it on their own schedules
  - APIs (REST, GraphQL, gRPC), event streams (topics + schemas), data
    products (published views, shared datasets), webhook payloads

When NOT to use
  - A private schema between a container and its own store → container
    blueprint, Data section
  - A shared library's API surface → usually the module blueprint's
    Interface section; author a contract only when it is formally
    specified and consumed across team boundaries
  - A third-party (SaaS) interface → the vendor's published contract
    governs; consuming blueprints reference it by URL. Contract
    documents exist only for interfaces the organization provides
  - WHY the interface is designed this way → ADR, referenced below

Slug format
  contract-<interface-name>, lowercase kebab-case. Name the interface,
  not the consumer. Example: contract-notification-events

Status vocabulary
  draft | ready-for-review | approved | active | retired

This is a living document. Semantics changes are contract changes:
version them under the same policy as shape changes. Remove all HTML
comment blocks after completing each section.
-->

# <Interface Name>

## Overview

<!--
What capability this interface vends, who provides it, and who it is
for. 1 short paragraph. State the vending mechanism (REST API, Kafka
topic, published view, ...) and link the provider's blueprint.
-->

<What this interface vends and for whom.> Provided by
@<provider-slug>. Visibility: <public | internal>.

## Specification

<!--
Point at the machine-readable artifacts — never duplicate their content.
The listed artifact at the pinned version is the normative source of
truth for shape. Include registry coordinates where applicable (schema
registry subject, package name).

For a GraphQL or REST service, the schema that matches the running
implementation 100% comes from the service itself — prefer generating it
(GraphQL introspection, an OpenAPI export endpoint) over a hand-kept copy
that drifts. WHERE that is possible, record HOW to obtain the schema (the
command or endpoint) rather than pinning a static snapshot; the running
service is the source of truth for shape, this document for semantics.
-->

| Artifact | Location | Normative for |
|----------|----------|---------------|
| <OpenAPI / SDL / proto / DDL> | <repo path / registry URL / how to generate> | Shape |
| This document | — | Semantics, operations, change policy |

## Guarantees

<!--
The semantic guarantees consumers can build against — everything the
schema cannot say. Write each guarantee in EARS. Cover what applies:
idempotency, ordering, delivery semantics, consistency, retry behavior,
error semantics, pagination stability, timezone/encoding conventions.
-->

- WHEN <a request/message is received under condition>, the provider
  SHALL <guarantee>.
- The provider SHALL <always-true invariant, e.g. idempotency on
  <key>>.
- Events on <topic> SHALL be ordered <per what key>.
- IF <failure condition>, THEN the provider SHALL <error semantics the
  consumer can rely on>.
- The provider SHALL NOT <thing consumers must not have to defend
  against>.

## Operational Expectations

<!--
The operational half of the contract: what consumers may assume about
availability and limits, and what the provider requires of consumers.
Reference standards docs for anything organization-wide (authN scheme)
rather than restating them.
-->

- **Availability / SLO:** <target and measurement window>
- **Latency:** <target at percentile>
- **Rate limits / quotas:** <limits and behavior at the limit>
- **AuthN/AuthZ:** <scheme and required scopes; reference standard>
- **Consumer obligations:** consumers SHALL <e.g. handle out-of-order
  delivery, honor Retry-After, tolerate unknown fields>.

## Versioning and Change Policy

<!--
How this contract evolves. Both shape AND semantics changes are subject
to this policy.
-->

- **Versioning scheme:** <semver on the spec artifact / URL versioning /
  schema registry compatibility mode>
- **Compatible changes:** the provider MAY <e.g. add optional fields>
  without a major version change.
- **Breaking changes:** WHEN a breaking change is required, the provider
  SHALL <major version / new topic / parallel-run window>.
- **Deprecation:** WHEN a version is deprecated, it SHALL remain
  available for <window>, and the provider SHALL notify consumers via
  <channel>.

## Key Decisions

- [[adr-<nnn>]] — <one-line statement of the decision>.

## Open Questions

<!--
Draft-mode only. WHEN a question is resolved, record it as an ADR or fold
it into the relevant section, then delete it here. WHILE status is draft,
the section MAY instead state the literal text "None." (no checkbox) —
the absence of questions is information. Remove the section entirely when
status becomes active.
-->

- [ ] <open question> — <what is blocked on it, who decides>

## References

- `c4-guide.md` — C4 mapping rules and contract visibility tiers
- @<provider-slug> — provider's blueprint
- <Additional references>
