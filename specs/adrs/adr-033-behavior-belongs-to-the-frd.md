---
title: "ADR-033: Behavior Belongs to the FRD"
description: A blueprint cites an AC id only as the reason for a rule, and
  element-to-AC traceability is derived from test citations rather than
  listed in blueprints
slug: adr-033
type: adr
status: approved
---

# ADR-033: Behavior belongs to the FRD

**Status:** approved  
**Date:** 2026-10-31

## Context

An FRD states a feature's behavior as acceptance criteria, and tests cite
those criteria by AC id ([[ac-traceability-standards]]). Architecture
blueprints written from the earlier templates restate that behavior, list the
features an element implements, and list the ACs it satisfies.

A behavior stated in both an FRD and a blueprint has two owners, and the two
copies diverge. The FRD is the copy that tests cite, so the blueprint's copy
is the one that goes stale. An AC list in a blueprint duplicates what the test
suite already records: each test's `AC:` citation names its criteria, and the
element's `test_paths` say which element the test belongs to.

Some rules look arbitrary without the criterion that motivates them. A
blueprint needs a way to give that reason without restating the behavior.

## Decision

We will keep behavior and acceptance criteria in the FRD. A blueprint cites an
AC id only as the reason for a rule that would look arbitrary without it.
Which element implements which criterion is derived from the test citations
under each element's `test_paths`, not listed in blueprints.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Blueprints list the ACs they implement | The element-to-AC map is in one place | The list drifts as tests move<br>A second owner of traceability | Duplicates the test citations, which are the durable record |
| Blueprints restate behavior as guarantees in their own words | A blueprint reads on its own | Two owners of one behavior diverge | The FRD is the copy tests cite, so the blueprint's copy goes stale |
| The FRD owns behavior; traceability comes from test citations | One owner per behavior<br>Traceability comes from artifacts that tests already maintain | Answering "which criteria does this element implement?" takes a query over test citations | N/A — selected |

## Consequences

**Positive:**
- Each behavior has one owner, and it's the one that tests cite.
- Element-to-AC traceability stays correct as tests move, because it comes
  from the tests.

**Negative:**
- Deriving traceability needs every active Component and Module Blueprint to
  declare `test_paths`.
- A behavior that a Feature Blueprint decides and no criterion states must be
  lifted into the FRD before the Feature Blueprint retires.

## References

- [[ac-traceability-standards]] — the `AC:` citation format and coverage
  obligations
- [[adr-020]] — FRDs evolve in place, keyed on AC ids
- [[adr-032]] — blueprints state rules, not code facts
- [[blueprint-content-standards]] — rule 4 applies this decision
