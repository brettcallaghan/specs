---
title: "ADR-011: Blueprints Declare Targeted Verification Gates"
description: Each blueprint names the minimal set of tests that must run when its element changes; agents run that set, CI runs the full suite
slug: adr-011
type: adr
status: active
---

# ADR-011: Blueprints declare targeted verification gates

**Status:** active  
**Date:** 2026-07-29

## Context

In large projects that run as orchestrated bead waves, running the full test
suite for every work item doesn't scale. It slows a parallel sprint run to a
crawl.

Agents can choose which tests to run for a change, but nothing anchors that
judgment. Nothing tells an isolated agent with no session context which tests
actually gate a change to a given architectural element. As a result, test
selection is inconsistent: some agents run too many tests, which is slow, and
others run too few, which is risky.

## Decision

Each blueprint's Change Guidance declares a **Verification gates** entry. The
entry names the specific, runnable tests that MUST run when that element
changes: the minimal, targeted set that an agent runs to prove a change before
handoff. CI and downstream agents still run the full suite. Feature blueprints
map the "Verified by" column of their AC Coverage onto these gates, so that the
targeted set is traceable to acceptance criteria.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Run the full suite on every change | Maximal safety per change. | Doesn't scale across many parallel work items.<br>Slows waves dramatically. | Prohibitive for throughput at the scale that the plugin targets. |
| Leave test selection entirely to agent judgment, with no per-element guidance | Flexible. | An isolated agent has no authoritative signal for which tests gate this element.<br>Selection is inconsistent and error-prone. | Under-specified: agents miss tests or run too many. |
| The blueprint declares the targeted gate, and CI runs the full suite | An authoritative, minimal set per element.<br>A fast agent loop.<br>CI still enforces full coverage. | Authors must curate and maintain the gate list for each element. | Not applicable; selected. |

## Consequences

**Positive:**

- The agent loop for each change is fast, so bead waves stay fast.
- The question "What must I run if I touch this element?" has an authoritative
  answer that you can find with `grep`.
- Full coverage is preserved because CI still runs every test.
- The gates are a direct input to groom-beads, which can attach the targeted
  set to each work item.

**Negative:**

- The gate list is another thing to keep current as tests move.
- A stale gate can under-test a change locally. CI still runs the full suite,
  which limits the risk.

## References

- [[adr-004]]: the blueprints that carry these gates
- [[adr-006]]: visibility gates the change rules; these gates gate the tests
- The Change Guidance sections of the component, module, and container
  blueprint templates; the groom-beads skill, which consumes these gates
