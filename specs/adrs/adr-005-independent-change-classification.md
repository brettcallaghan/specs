---
title: "ADR-005: Classify Composed Units by the Boundary of Independent Change"
description: Use the independent-change boundary — unified with contract visibility — to separate components from modules
slug: adr-005
type: adr
status: active
updated_at: 2026-08-05T00:00:00Z
---

# ADR-005: Classify Composed Units by the Boundary of Independent Change

**Status:** active  
**Date:** 2026-07-29

## Context

Having decided to split the component and module types
([[adr-004]]), we needed a crisp, durable
test for which is which. An initial cut used composite-ness — does the unit have
internal components? — as the boundary. But the concern that actually changes
how an agent behaves when it touches a unit (whether it must preserve backwards
compatibility, version carefully, and record change history) is not triggered by
having sub-parts. It is triggered by being shared or published to consumers who
evolve on their own schedules. The framework already draws exactly this line for
contracts, whose visibility tiers are private, internal, and public.

## Decision

We will classify composed units by their **boundary of independent change** —
the same test the framework uses for contract visibility. A **component** is
internal to one container and changes in lockstep with it. A **module** crosses
a boundary of independent change: it is shared or published across containers,
and/or it is composite with internal components of its own. This yields a single
three-tier spectrum spanning both contracts and composed units: **private** (a
component; a private schema), **internal** (an in-repo module; an internal
contract), and **public** (a published module; a public contract).

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Split on composite-ness (has internal components) | Intuitive — "a module is the bigger thing" | Misses the concern that matters: an atomic shared library has no sub-parts yet needs versioning and compatibility;<br>a composite unit used by one container does not | Would misclassify exactly the cases where the distinction matters most |
| Split on a bare shared / not-shared binary | Captures sharing | Loses the published-vs-in-repo distinction that governs breaking-change freedom;<br>not unified with contracts | Too coarse (see [[adr-006]]) |
| Boundary of independent change, unified with contract visibility | One rule learned once, spanning contracts and composed units<br>Directly predicts the change-rule behavior of a unit | "Independent change" is a subtler concept than "has sub-parts" and must be taught | N/A — selected |

## Consequences

**Positive:**
- One conceptual model spans contracts and composed units — learn it once
- The classification directly predicts the change rules a unit is subject to
- It matches how teams actually mature code: internal first, published later

**Negative:**
- "Boundary of independent change" is a subtler test than "has sub-parts," and
  requires the C4 guide to teach it explicitly (done)

## References

- [[adr-004]] — the type split this test
  governs
- [[adr-006]] — how the tier is recorded
- `references/c4-guide.md` — Contract Visibility Tiers and the composed-unit parallel
