---
title: "ADR-011: Blueprints Declare Targeted Verification Gates"
description: Each blueprint names the minimal set of tests that must run when its element changes; agents run that set, CI runs the full suite
slug: adr-011
type: adr
status: active
---

# ADR-011: Blueprints Declare Targeted Verification Gates

**Status:** active  
**Date:** 2026-07-29

## Context

In large projects run as orchestrated bead waves, running the full test suite
for every work item does not scale — it slows a parallel sprint run to a crawl.

Agents have been given the agency to choose which tests to run for a change, but
that judgment is unanchored: nothing tells an isolated agent, with zero session
context, which tests actually gate a change to a given architectural element.
The result is inconsistent selection — some agents over-run (slow), others
under-test (risky).

## Decision

We will have each blueprint's Change Guidance declare a **Verification gates**
entry naming the specific, runnable tests that MUST run when that element
changes — the minimal, targeted set an agent runs to prove a change before
handoff. CI and downstream agents still run the full suite. Feature blueprints
map their AC Coverage "Verified by" column onto these gates, so the targeted set
is traceable to acceptance criteria.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Run the full suite on every change | Maximal safety per change | Does not scale across many parallel work items;<br>slows waves dramatically | Throughput-prohibitive at the scale the plugin targets |
| Leave test selection entirely to agent judgment, with no per-element guidance | Flexible | An isolated agent has no authoritative signal for which tests gate this element;<br>selection is inconsistent and error-prone | Under-specified — misses tests or over-runs |
| The blueprint declares the targeted gate;<br>CI runs the full suite | An authoritative, per-element minimal set<br>Fast agent loop<br>Full coverage still enforced in CI | Authors must curate and maintain the gate list per element | N/A — selected |

## Consequences

**Positive:**
- A fast per-change agent loop, so bead waves stay fast
- A greppable, authoritative answer to "what must I run if I touch this
  element?"
- Full coverage is preserved because CI still runs everything
- A direct input to groom-beads, which can attach the targeted set to each
  work item

**Negative:**
- The gate list is another thing to keep current as tests move
- A stale gate could under-test a change locally (mitigated because CI still
  runs the full suite)

## References

- [[adr-004]] — the blueprints that carry
  these gates
- [[adr-006]] — visibility gates the change
  rules; these gates the tests
- Change Guidance sections of the component, module, and container blueprint
  templates; the groom-beads skill (consumes these gates)
