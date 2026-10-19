---
title: "ADR-010: Contracts Are Independent, Long-Lived Documents"
description: Keep contracts separate from blueprints because they outlive and may precede implementations; source live schemas from the running service
slug: adr-010
type: adr
status: active
---

# ADR-010: Contracts are independent, long-lived documents

**Status:** active  
**Date:** 2026-07-29

## Context

Contracts, which are interface specifications, were already separate documents
rather than sections of blueprints. However, the rationale wasn't recorded, and
the template didn't stress it. A contract's lifecycle differs fundamentally
from a blueprint's. A blueprint describes one design of a system, container, or
module as it currently IS. An interface, by contrast, can be implemented many
times over its life. For example, it might be re-implemented for
modernization, or served by more than one provider for backwards
compatibility. An interface can also exist before any provider does, and
contract-first design is often the most consequential architectural decision.
Separately, for GraphQL and REST services, a hand-maintained copy of the schema
silently drifts from the running implementation.

## Decision

We keep contracts as documents that are independent of any blueprint, and the
template states the rationale: a contract outlives, and can precede, the
implementations that realize it. For GraphQL and REST interfaces, the schema
that matches the implementation comes from the running service, through
introspection or an OpenAPI export endpoint. The template instructs authors to
record how to obtain the schema rather than pin a static snapshot that drifts.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Fold interface specs into the owning blueprint | Fewer documents.<br>Co-located with the provider. | Dies when the blueprint is superseded.<br>Can't precede a provider (contract-first).<br>Can't represent an interface with multiple implementations. | Couples an independent lifecycle to the wrong host. |
| Maintain a hand-kept schema snapshot in the repository | Always present offline. | Drifts from the running implementation,<br>and the drift is silent. | The running service is the only fully accurate source of the schema's shape. |
| Independent contract documents, with the schema sourced from the running service | The contract lifecycle is independent of any one design.<br>Contract-first design is expressible.<br>The schema always matches the implementation. | One more document to maintain.<br>Obtaining the schema might require the service to be running. | Not applicable; selected. |

## Consequences

**Positive:**

- Contracts survive re-implementation and can precede their providers.
- Contract-first design is a first-class, expressible workflow.
- Schemas for live services don't silently drift from the implementation.

**Negative:**

- A contract is a separate document to maintain over its long life.
- Retrieving the schema might require running the service or CI access to it.

## References

- `references/c4-guide.md`: Contract Visibility Tiers
- [[adr-005]]: the same visibility spectrum classifies contracts and composed
  units
- Template: `contract-template.md`
