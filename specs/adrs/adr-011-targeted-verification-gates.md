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

In large projects that run as orchestrated waves of beads, running the full
test suite for every work item doesn't scale. It makes a parallel sprint run
very slow.

Agents can choose which tests to run for a change, but nothing anchors that
choice. An isolated agent with no session context has no way to learn which
tests gate a change to a given architectural element. As a result, test
selection is inconsistent: some agents run more tests than needed, which is
slow, and others under-test, which is risky.

## Decision

We will have each blueprint's Change Guidance declare a **Verification gates**
entry. The entry names the specific, runnable tests that MUST run when that
element changes. These tests are the minimal, targeted set that an agent runs to
prove a change before handoff. CI and downstream agents still run the full
suite. Feature blueprints map the "Verified by" column of their AC Coverage
section onto these gates, so the targeted set is traceable to acceptance
criteria.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Run the full suite on every change | Maximal safety per change | Doesn't scale across many parallel work items;<br>slows waves dramatically | Throughput-prohibitive at the scale the plugin targets |
| Leave test selection entirely to agent judgment, with no per-element guidance | Flexible | An isolated agent has no authoritative signal for which tests gate this element;<br>selection is inconsistent and error-prone | Under-specified — misses tests or over-runs |
| The blueprint declares the targeted gate;<br>CI runs the full suite | An authoritative, per-element minimal set<br>Fast agent loop<br>Full coverage still enforced in CI | Authors must curate and maintain the gate list per element | N/A — selected |

## Consequences

**Positive:**
- The agent loop for each change is fast, so bead waves stay fast.
- Agents get a greppable, authoritative answer to "What must I run if I touch
  this element?"
- Full coverage is preserved because CI still runs everything.
- The gates are a direct input to groom-beads, which can attach the targeted
  set to each work item.

**Negative:**
- The gate list is another thing to keep current as tests move.
- A stale gate could under-test a change locally. Because CI still runs the
  full suite, the risk is limited.

## References

- [[adr-004]] — the blueprints that carry these gates
- [[adr-006]] — visibility determines the change rules; these gates determine
  the tests
- The Change Guidance sections of the component, module, and container
  blueprint templates
- The groom-beads skill, which consumes these gates
