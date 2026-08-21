<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/guides/work-order-field-guide.md (version 1.3)
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Work Order Field Guide
description: How to fill a work order's fields well — the three text
  fields kept distinct, the dispatch metadata conventions, and the
  tier-routing rubric
slug: work-order-field-guide
type: guide
status: active
updated_at: 2026-08-21T00:00:00Z
version: "1.3"
---

# Work Order Field Guide

## Contents

- The three text fields (Problem / Design / Acceptance) and what separates them
- Dispatch metadata conventions — `tier`, `tier_reason`, `est_size`, `c4_*`
- Why tiers, not model names — and the project's `sf.toml` `[models]` lineup
- Tier-routing rubric — which signals push toward `light` / `standard` /
  `advanced` / `frontier`
- Calibration — escalations as routing evidence, and why over-routing is
  invisible
- Estimates vs. tiers — what `est_size` answers that `tier` does not

Companion to `work-order-template.md` and [[work-order-definition-of-ready]]. The
template defines the shape and the DoR defines the gate; this guide is how to
fill the fields well.

## The three text fields (keep them distinct)

Agents habitually collapse these into one blob. Keeping them separate is most
of the quality battle, because each answers a different question for the
isolated agent.

| Section | Question | Mutability | Contents |
|---------|----------|------------|----------|
| `## Problem` | **What & why** | Set at creation, rarely changed | The problem statement. Self-contained. Concrete anchors (paths, symbols, endpoints, behavior). For a bug: repro + expected/actual. |
| `## Design` | **How** | Mutable — evolves as the agent learns | Chosen approach, key decisions, constraints, trade-offs. May hedge. Keeps the agent from re-deriving architecture under token pressure. |
| `## Acceptance` | **Done?** | Stable once agreed | Outcome-focused checkboxes. Observable end states, not steps. The shared contract between the isolated agent and the orchestrator/verifier. |

Rule of thumb: if a line tells you *what to type*, it belongs in `## Design`. If
it tells you *what must be true when finished*, it belongs in `## Acceptance`.
If it tells you *what the problem is*, it belongs in `## Problem`.

## Dispatch metadata conventions

The frontmatter fields below make a work order routable and packable by an
orchestrator. Use stable, lowercase, snake_case keys. The conventions this
framework standardizes on:

- `tier` — the capability tier this order needs. Closed, **ordered** set
  (`light | standard | advanced | frontier`, least to most capable). Never
  null. This is the routing field; see the rubric below.
- `tier_reason` — one line; what about the order drove the choice. Auditable
  by retro.
- `model` — OPTIONAL override, and only when a *named* model is genuinely
  required (a model-specific capability or an explicit user instruction).
  Normal orders carry `tier` and no `model`. See "Why tiers, not model
  names" below.
- `est_size` — `XS | S | M | L | XL`; lets the orchestrator pack waves by
  effort independently of tier. Correlated with tier but not identical.
- `c4_changes` — components whose code this order edits (the collision key for
  wave isolation: two orders that both change the same component collide even
  when neither blocks the other).
- `c4_impacts` — components affected downstream but not edited.
- `c4_docs` — the C4 doc files that may need updating as a result.

How these fields are physically stored is the materializing skill's concern —
`groom-beads` maps them onto the bead's metadata JSON, and validates the
result against its bead schema. The conventions here are what keep every
grooming pass uniform instead of each agent inventing its own field names.

## Why tiers, not model names

A work order is portable and long-lived: it may be groomed today and
dispatched next quarter, in a project whose model lineup differs from
yours. A named model is the one routing value that does not travel — it
churns with every release, and it forces the same edit across a schema, a
validator, and every skill that reads it.

The tier is the durable claim ("this order needs advanced reasoning"); the
model is a local binding. Each project declares that binding once, in
`sf.toml` at the repository root, and the orchestrator resolves `tier` → model
at dispatch:

```toml
[models]
# Least to most capable.
light = "<model>"
standard = "<model>"
advanced = "<model>"
frontier = "<model>"
```

It lives in a config file rather than in prose because it is a routing fact a
script resolves, not guidance an agent reads.

WHEN a project's lineup has fewer than four models, map several tiers onto
one model rather than inventing tier names — the ordering is what the
orchestrator's escalation ladder depends on.

## Tier-routing rubric

**The goal is the LOWEST tier that can complete the order correctly.**
Capability, cost, and speed are different dimensions, and only the first is
a correctness concern: an under-capable model produces wrong work, while an
over-capable one produces right work at unnecessary cost. So route on
capability need, then stop — do not reach for a higher tier as insurance.

Route on signals visible at grooming time. This is a rubric, not a formula —
explain the call in `tier_reason`.

**`frontier` / `advanced`** — reach for the top of the ladder when the order
has any of:
- High ambiguity or novelty — the spec is the hard part; the agent must make
  non-obvious design decisions.
- Large blast radius — touches many C4 components, or a high-impact one
  (`c4_changes` / `c4_impacts` counts are your proxy).
- Deep reasoning — tricky concurrency, security-sensitive logic, algorithmic
  work, cross-cutting refactors.
- High cost of being wrong — data migrations, auth, billing, anything hard to
  undo.

Prefer `advanced` and reserve `frontier` for orders that combine several of
these, or where a wrong answer is expensive to detect.

**`standard`** — the default landing spot, when the order is:
- Well-specified with clear acceptance criteria.
- Localized to one or two components.
- Mostly execution over invention — mirrors an existing pattern in the
  codebase.

**`light`** — when:
- The work is mechanical and low-ambiguity (rote edits, boilerplate, config,
  simple test additions), and
- Throughput / cheap parallel fan-out matters more than deep reasoning.

**Defaults & guardrails:**
- No work order gets a null tier — if genuinely unsure, assign `standard`
  and say so in `tier_reason`.
- Blast radius can override an otherwise-simple order upward: a "small" change
  to a load-bearing component is not small in risk.
- Escalate on failure, never downgrade mid-run. An orchestrator that has to
  raise an order's tier to get it done records that (see below); grooming
  that had to be corrected is the signal this rubric is mis-calibrated.

## Calibration: the tier is a prediction, and it is checked

`tier` is grooming's *prediction* of the capability an order needs.
Dispatch is where that prediction is tested. An orchestrator that raises an
order's tier mid-run — because the assigned tier failed gates or visibly
flailed — SHALL report the correction: the order, the assigned tier, the
tier that completed it, and what failed.

Those corrections are the only direct evidence this rubric produces, and
they run in one direction:

- **Under-routing** (had to escalate) is visible and self-correcting — the
  run reports it, and grooming can adjust.
- **Over-routing** (a `frontier` order a `light` model would have finished)
  is invisible: the work succeeds and nobody learns anything. It is paid
  for silently, on every run, forever.

So treat a *zero*-escalation run across a large queue as a signal worth
examining, not as a clean bill of health. It is consistent with perfect
calibration and equally consistent with routing everything too high.
`retro-spec` and `retro-skill` both take escalation counts as input.

## A note on estimates vs. tiers

`est_size` answers "how much work" and drives wave packing (don't cram ten `L`
orders into one wave). `tier` answers "how much capability per unit work." A
large but mechanical order can be `L` + `light`; a tiny but subtle order can
be `S` + `frontier`. Grooming should set both.
