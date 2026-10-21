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

[[adr-004]] splits the component and module types, so the framework needs a
crisp, durable test for which is which. One candidate boundary is
composite-ness: whether the unit has internal components. The concern that
changes how an agent behaves when it touches a unit is whether it must preserve
backwards compatibility, version carefully, and record change history. Having
sub-parts doesn't trigger that concern. Being shared or published to consumers
who evolve on their own schedules does. The framework already draws this line
for contracts, whose visibility tiers are private, internal, and public.

## Decision

We will classify composed units by their **boundary of independent change** —
the same test the framework uses for contract visibility. A **component** is
internal to one container and changes in lockstep with it. A **module** crosses
a boundary of independent change: it is shared or published across containers,
or it is composite with internal components of its own, or both. The result is
a single three-tier spectrum that spans contracts and composed units:
**private** (a component, a private schema), **internal** (an in-repo module,
an internal contract), and **public** (a published module, a public contract).

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Split on composite-ness (has internal components) | Intuitive: "a module is the bigger thing" | Misses the concern that matters: an atomic shared library has no sub-parts yet needs versioning and compatibility, and a composite unit used by one container doesn't | Would misclassify exactly the cases where the distinction matters most |
| Split on a bare shared / not-shared binary | Captures sharing | Loses the published-versus-in-repo distinction that governs freedom to make breaking changes<br>Not unified with contracts | Too coarse (see [[adr-006]]) |
| Boundary of independent change, unified with contract visibility | One rule learned once, spanning contracts and composed units<br>Directly predicts the change-rule behavior of a unit | "Independent change" is a subtler concept than "has sub-parts," and it must be taught | N/A — selected |

## Consequences

**Positive:**
- One conceptual model, learned once, spans contracts and composed units
- The classification directly predicts the change rules a unit is subject to
- It matches how teams mature code: internal first, published later

**Negative:**
- "Boundary of independent change" is a subtler test than "has sub-parts," so
  [[c4-standards]] must teach it explicitly, which it does

## References

- [[adr-004]] — the type split this test
  governs
- [[adr-006]] — how the tier is recorded
- [[c4-standards#Contract Visibility Tiers]] — the contract tiers and the
  composed-unit parallel
