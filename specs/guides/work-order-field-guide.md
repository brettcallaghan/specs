---
title: Work Order Field Guide
description: How to fill a work order's fields well — the three text
  fields kept distinct, the dispatch metadata conventions, and the
  tier-routing rubric
slug: work-order-field-guide
type: guide
status: active
---

# Work Order field guide

## Contents

- The three text fields (Problem / Design / Acceptance) and what separates them
- Dispatch metadata conventions — `tier`, `tier_reason`, `est_size`, `c4_*`
- Why tiers, not model names — and the orchestrator's model lineup
- Tier-routing rubric — which signals push toward `light` / `standard` /
  `advanced` / `frontier`
- Calibration — escalations as routing evidence, and why over-routing is
  invisible
- Estimates vs. tiers — what `est_size` answers that `tier` doesn't

This guide accompanies `plugin/skills/spec/references/work-order-template.md`
and [[work-order-definition-of-ready]]. The template defines the shape, and
the Definition of Ready defines the gate. This guide covers how to fill in the
fields well.

## The three text fields (keep them distinct)

Agents habitually collapse these three fields into one blob. Keeping them
separate is most of what makes a Work Order good, because each field answers
a different question for the isolated agent.

| Section | Question | Mutability | Contents |
|---------|----------|------------|----------|
| `## Problem` | **What & why** | Set at creation, rarely changed | The problem statement. Self-contained. Concrete anchors (paths, symbols, endpoints, behavior). For a bug: reproduction steps, and expected and actual behavior. |
| `## Design` | **How** | Mutable: evolves as the agent learns | Chosen approach, key decisions, constraints, trade-offs. May hedge. Keeps the agent from re-deriving architecture under token pressure. |
| `## Acceptance` | **Done?** | Stable once agreed | Outcome-focused checkboxes. Observable end states, not steps. The shared contract between the isolated agent and the orchestrator or verifier. |

Rule of thumb: if a line tells you *what to type*, it belongs in `## Design`. If
it tells you *what must be true when finished*, it belongs in `## Acceptance`.
If it tells you *what the problem is*, it belongs in `## Problem`.

## Dispatch metadata conventions

The frontmatter fields below let an orchestrator route Work Orders and pack
them into batches. Use stable, lowercase, snake_case keys. This framework
standardizes on these conventions:

- `tier` — the capability tier this order needs. It's a closed, **ordered**
  set (`light | standard | advanced | frontier`, from least to most capable),
  and it's never null. This is the routing field (see the rubric below).
- `tier_reason` — one line that says what about the order drove the choice.
  A retro can audit it.
- `model` — an OPTIONAL override, used only when a *named* model is genuinely
  required (for a model-specific capability or an explicit user
  instruction). A normal order carries `tier` and no `model`. See "Why tiers,
  not model names" below.
- `est_size` — `XS | S | M | L | XL`. It lets the orchestrator pack batches by
  effort, independently of tier. It correlates with tier but isn't the same.
- `c4_changes` — the components whose code this order edits. This is the
  collision key for dispatch isolation: two orders that both change the same
  component collide, even when neither blocks the other.
- `c4_impacts` — the components affected downstream but not edited.
- `c4_docs` — the C4 doc files that may need updating as a result.

How these fields are physically stored is up to the materializing skill:
`groom-beads` maps them onto the bead's metadata JSON and validates the result
against its bead schema. These conventions keep every grooming pass uniform,
instead of letting each agent invent its own field names.

## Why tiers, not model names

A Work Order is portable and long-lived: it may be groomed today and
dispatched next quarter, in a project whose model lineup differs from yours.
A named model is the one routing value that doesn't travel. It churns with
every release, and it forces the same edit across a schema, a validator, and
every skill that reads it.

The tier is the durable claim ("this order needs advanced reasoning"). The
model is a local binding. Each project declares that binding once, in
whatever configuration its orchestrator reads, and the orchestrator resolves
`tier` to a model at dispatch. The binding lives in a config file instead of
in prose because it's a routing fact that a script resolves, not guidance
that an agent reads.

When a project's lineup has fewer than four models, map several tiers onto
one model instead of inventing tier names, because the orchestrator's
escalation ladder depends on the ordering.

## Tier-routing rubric

**The goal is the lowest tier that can complete the order correctly.**
Capability, cost, and speed are different dimensions. Only capability is a
correctness concern: an under-capable model produces wrong work, while an
over-capable one produces right work at unnecessary cost. So route on
capability need, and then stop. Don't reach for a higher tier as insurance.

Route on signals that are visible at grooming time. This is a rubric, not a
formula, so explain the call in `tier_reason`.

**`frontier` / `advanced`** — reach for the top of the ladder when the order
has any of these signals:
- High ambiguity or novelty: the spec is the hard part, and the agent must
  make non-obvious design decisions.
- Large blast radius: the order touches many C4 components, or a high-impact
  one (the `c4_changes` and `c4_impacts` counts are your proxy).
- Deep reasoning: tricky concurrency, security-sensitive logic, algorithmic
  work, or cross-cutting refactors.
- High cost of being wrong: data migrations, auth, billing, or anything hard
  to undo.

Prefer `advanced`. Reserve `frontier` for orders that combine several of these
signals, or where a wrong answer is expensive to detect.

**`standard`** — the default landing spot, when the order is:
- Well specified, with clear acceptance criteria.
- Localized to one or two components.
- Mostly execution, not invention: it mirrors an existing pattern in the
  codebase.

**`light`** — when:
- The work is mechanical and low-ambiguity (rote edits, boilerplate, config,
  simple test additions), and
- Throughput or cheap parallel fan-out matters more than deep reasoning.

**Defaults and guardrails:**
- No Work Order gets a null tier. If you're genuinely unsure, assign
  `standard` and say so in `tier_reason`.
- Blast radius can push an otherwise simple order to a higher tier: a "small"
  change to a load-bearing component isn't small in risk.
- Escalate on failure, and never downgrade mid-run. An orchestrator that has
  to raise an order's tier to get it done records that (see below). Grooming
  that had to be corrected is the signal that this rubric is miscalibrated.

## Calibration: the tier is a prediction, and it is checked

`tier` is grooming's *prediction* of the capability an order needs.
Dispatch tests that prediction. An orchestrator that raises an
order's tier mid-run — because the assigned tier failed gates or visibly
flailed — SHALL report the correction: the order, the assigned tier, the
tier that completed it, and what failed.

Those corrections are the only direct evidence this rubric produces, and
they run in one direction:

- **Under-routing** (had to escalate) is visible and self-correcting: the
  run reports it, and grooming can adjust.
- **Over-routing** (a `frontier` order that a `light` model would have
  finished) is invisible: the work succeeds, and nobody learns anything. You
  pay for it silently, on every run, forever.

So treat a run with *zero* escalations across a large queue as a signal worth
examining, not as a clean bill of health. It's consistent with perfect
calibration, and equally consistent with routing everything too high.
`retro-spec` and `retro-skill` both take escalation counts as input.

## A note on estimates vs. tiers

`est_size` answers "how much work" and drives packing (don't cram ten `L`
orders into one batch). `tier` answers "how much capability per unit work." A
large but mechanical order can be `L` and `light`. A tiny but subtle order can
be `S` and `frontier`. Grooming should set both.
