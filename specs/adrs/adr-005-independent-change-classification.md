---
title: "ADR-005: Classify Composed Units by the Boundary of Independent Change"
description: Use the independent-change boundary — unified with contract visibility — to separate components from modules
slug: adr-005
type: adr
status: active
---

# ADR-005: Classify Composed Units by the Boundary of Independent Change

**Status:** active  
**Date:** 2026-07-29

## Context

After [[adr-004]] split the component and module types, we needed a clear,
durable test for telling them apart. An initial version used composite-ness as
the boundary: does the unit have internal components? But having sub-parts
doesn't trigger the concern that changes how an agent behaves when it touches a
unit. That concern is whether the agent must preserve backwards compatibility,
version carefully, and record change history. What triggers it is being shared
with, or published to, consumers who evolve on their own schedules. The
framework already draws exactly this line for contracts, whose visibility tiers
are private, internal, and public.

## Decision

We will classify composed units by their **boundary of independent change**,
the same test that the framework uses for contract visibility. A **component**
is internal to one container and changes in lockstep with it. A **module**
crosses a boundary of independent change: it's shared or published across
containers, it's composite with internal components of its own, or both.

This test yields a single three-tier spectrum that spans both contracts and
composed units:

- **private**: a component; a private schema
- **internal**: an in-repo module; an internal contract
- **public**: a published module; a public contract

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Split on composite-ness (has internal components) | Intuitive — "a module is the bigger thing" | Misses the concern that matters: an atomic shared library has no sub-parts yet needs versioning and compatibility;<br>a composite unit used by one container does not | Would misclassify exactly the cases where the distinction matters most |
| Split on a bare shared / not-shared binary | Captures sharing | Loses the published-vs-in-repo distinction that governs breaking-change freedom;<br>not unified with contracts | Too coarse (see [[adr-006]]) |
| Boundary of independent change, unified with contract visibility | One rule learned once, spanning contracts and composed units<br>Directly predicts the change-rule behavior of a unit | "Independent change" is a subtler concept than "has sub-parts" and must be taught | N/A — selected |

## Consequences

**Positive:**
- One conceptual model spans contracts and composed units, so readers learn it
  once.
- The classification directly predicts the change rules that apply to a unit.
- It matches how teams mature code: internal first, published later.

**Negative:**
- "Boundary of independent change" is a subtler test than "has sub-parts," so
  the C4 guide must teach it explicitly. The C4 guide does so.

## References

- [[adr-004]] — the type split this test governs
- [[adr-006]] — how the tier is recorded
- `references/c4-guide.md` — Contract Visibility Tiers and the composed-unit
  parallel
