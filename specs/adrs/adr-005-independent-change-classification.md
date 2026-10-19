---
title: "ADR-005: Classify Composed Units by the Boundary of Independent Change"
description: Use the independent-change boundary — unified with contract visibility — to separate components from modules
slug: adr-005
type: adr
status: active
---

# ADR-005: Classify composed units by the boundary of independent change

**Status:** active  
**Date:** 2026-07-29

## Context

After we decided to split the component and module types ([[adr-004]]), we
needed a clear, durable test for which is which. An initial version used
composite-ness as the boundary: does the unit have internal components?
However, having sub-parts doesn't trigger the concern that actually changes how
an agent behaves when it touches a unit. That concern is whether the agent must
preserve backwards compatibility, version carefully, and record change
history. Being shared or published to consumers who evolve on their own
schedules triggers it. The framework already draws exactly this line for
contracts, whose visibility tiers are private, internal, and public.

## Decision

We classify composed units by their **boundary of independent change**, which
is the same test that the framework uses for contract visibility:

- A **component** is internal to one container and changes in lockstep with
  it.
- A **module** crosses a boundary of independent change. It's shared or
  published across containers, composite with internal components of its own,
  or both.

This test yields a single three-tier spectrum that spans both contracts and
composed units:

- **Private**: a component; a private schema.
- **Internal**: an in-repo module; an internal contract.
- **Public**: a published module; a public contract.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Split on composite-ness (has internal components) | Intuitive: "a module is the bigger thing." | Misses the concern that matters. An atomic shared library has no sub-parts but needs versioning and compatibility;<br>a composite unit used by one container doesn't. | Would misclassify exactly the cases where the distinction matters most. |
| Split on a bare shared or not-shared binary | Captures sharing. | Loses the distinction between published and in-repo, which governs freedom to make breaking changes.<br>Not unified with contracts. | Too coarse (see [[adr-006]]). |
| Boundary of independent change, unified with contract visibility | One rule, learned once, spans contracts and composed units.<br>Directly predicts the change-rule behavior of a unit. | "Independent change" is a subtler concept than "has sub-parts" and must be taught. | Not applicable; selected. |

## Consequences

**Positive:**

- One conceptual model spans contracts and composed units, so you learn it
  once.
- The classification directly predicts the change rules that apply to a unit.
- It matches how teams actually mature code: internal first, published later.

**Negative:**

- "Boundary of independent change" is a subtler test than "has sub-parts," and
  the C4 guide must teach it explicitly. The guide now does.

## References

- [[adr-004]]: the type split that this test governs
- [[adr-006]]: how the tier is recorded
- `references/c4-guide.md`: Contract Visibility Tiers and the composed-unit
  parallel
