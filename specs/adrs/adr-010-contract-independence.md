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

Contracts (interface specifications) were already separate documents rather
than sections of blueprints. But nobody had recorded the rationale, and the
template didn't stress it. A contract's lifecycle differs from a blueprint's. A
blueprint describes one design of a system, container, or module as it
currently is. An interface may be implemented many times over its life:
re-implemented for modernization, or served by more than one provider for
backwards compatibility. It may also exist before any provider, and
contract-first design is often the most consequential architectural decision.

For GraphQL and REST services, a hand-maintained schema copy silently drifts
from the running implementation.

## Decision

We will keep contracts as documents independent of any blueprint, and state the
rationale in the template: a contract outlives, and may precede, the
implementations that realize it. For GraphQL and REST interfaces, the schema
that matches the implementation comes from the running service (introspection,
or an OpenAPI export endpoint). The template instructs authors to record how to
obtain it rather than pin a static snapshot that drifts.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Fold interface specs into the owning blueprint | Fewer documents;<br>co-located with the provider | Dies when the blueprint is superseded<br>Can't precede a provider (contract-first)<br>Can't represent an interface with multiple implementations | Couples an independent lifecycle to the wrong host |
| Maintain a hand-kept schema snapshot in the repo | Always present offline | Drifts from the running implementation,<br>and the drift is silent | The running service is the only 100%-accurate source of shape |
| Independent contract documents + schema sourced from the running service | The contract lifecycle is independent of any one design<br>Contract-first is expressible<br>The schema always matches the implementation | One more document to steward<br>Obtaining the schema may require the service running | N/A (selected) |

## Consequences

**Positive:**
- Contracts survive re-implementation and can precede their providers
- Contract-first design is a first-class, expressible workflow
- Schemas for live services don't silently drift from the implementation

**Negative:**
- A contract is a separate document to maintain over its (long) life
- Schema retrieval may require running the service or CI access to it

## References

- `references/c4-guide.md`: Contract visibility tiers
- [[adr-005]]: the same visibility spectrum classifies contracts and composed
  units
- Template: `contract-template.md`
