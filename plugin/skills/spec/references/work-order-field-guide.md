---
title: Work Order Field Guide
description: How to fill a work order's fields well — the three text
  fields kept distinct, the dispatch metadata conventions, and the
  model-routing rubric
slug: work-order-field-guide
type: guide
status: active
updated_at: 2026-07-30T00:00:00Z
version: "1.1"
related:
  - "work-order-template"
  - "work-order-definition-of-ready"
---

# Work Order Field Guide

Companion to `references/work-order-template.md` and [[work-order-definition-of-ready]]. The
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

- `model` — closed set (`opus | sonnet | fable`). Never null.
- `model_reason` — one line; what about the order drove the choice. Auditable
  by retro.
- `est_size` — `XS | S | M | L | XL`; lets the orchestrator pack waves by
  effort independently of model. Correlated with model but not identical.
- `c4_changes` — components whose code this order edits (the collision key for
  wave isolation: two orders that both change the same component collide even
  when neither blocks the other).
- `c4_impacts` — components affected downstream but not edited.
- `c4_docs` — the C4 doc files that may need updating as a result.

How these fields are physically stored is the materializing skill's concern —
`groom-beads` maps them onto the bead's metadata JSON, and validates the
result against its bead schema. The conventions here are what keep every
grooming pass uniform instead of each agent inventing its own field names.

## Model-routing rubric

Route on signals visible at grooming time. This is a rubric, not a formula —
explain the call in `model_reason`.

**Push toward the top model (`opus`)** when the order has any of:
- High ambiguity or novelty — the spec is the hard part; the agent must make
  non-obvious design decisions.
- Large blast radius — touches many C4 components, or a high-impact one
  (`c4_changes` / `c4_impacts` counts are your proxy).
- Deep reasoning — tricky concurrency, security-sensitive logic, algorithmic
  work, cross-cutting refactors.
- High cost of being wrong — data migrations, auth, billing, anything hard to
  undo.

**Route to the mid model (`sonnet`)** when the order is:
- Well-specified with clear acceptance criteria.
- Localized to one or two components.
- Mostly execution over invention — mirrors an existing pattern in the
  codebase.

**Route to the cheap/fast model (`fable`)** when:
- The work is mechanical and low-ambiguity (rote edits, boilerplate, config,
  simple test additions), and
- Throughput / cheap parallel fan-out matters more than deep reasoning.

**Defaults & guardrails:**
- No work order gets a null model — if genuinely unsure, assign the mid model
  and say so.
- Blast radius can override an otherwise-simple order upward: a "small" change
  to a load-bearing component is not small in risk.
- Keep the vocabulary closed. If the project's model lineup differs, adopt its
  names but preserve the tiering (top / mid / cheap-fast) and the reasoning.

## A note on estimates vs. models

`est_size` answers "how much work" and drives wave packing (don't cram ten `L`
orders into one wave). `model` answers "how much thinking per unit work." A
large but mechanical order can be `L` + `fable`; a tiny but subtle order can be
`S` + `opus`. Grooming should set both.
