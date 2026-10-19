<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/guides/work-order-field-guide.md (version unversioned)
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
---

# Work order field guide

## Contents

- The three text fields (Problem, Design, Acceptance) and what separates
  them
- Dispatch metadata conventions: `tier`, `tier_reason`, `est_size`, and
  `c4_*`
- Why tiers, not model names, and the orchestrator's model lineup
- Tier-routing rubric: which signals push toward `light`, `standard`,
  `advanced`, or `frontier`
- Calibration: escalations as routing evidence, and why over-routing is
  invisible
- Estimates vs. tiers: what `est_size` answers that `tier` doesn't

This guide is a companion to `work-order-template.md` and
[[work-order-definition-of-ready]]. The template defines the shape, and the
Definition of Ready (DoR) defines the gate. This guide explains how to fill
the fields well.

## The three text fields (keep them distinct)

Agents habitually merge these fields into one block of text. Keeping them
separate does most of the work of a good work order, because each field
answers a different question for the isolated agent.

| Section | Question | Mutability | Contents |
|---------|----------|------------|----------|
| `## Problem` | What and why | Set at creation, rarely changed | The problem statement. Self-contained. Concrete anchors (paths, symbols, endpoints, behavior). For a bug, the repro and the expected and actual results. |
| `## Design` | How | Mutable. Evolves as the agent learns | Chosen approach, main decisions, constraints, and trade-offs. May hedge. Keeps the agent from re-deriving architecture under token pressure. |
| `## Acceptance` | Done? | Stable once agreed | Outcome-focused checkboxes. Observable end states, not steps. The shared contract between the isolated agent and the orchestrator or verifier. |

To place a line, ask what it tells you. If it tells you what to type, it
belongs in `## Design`. If it tells you what must be true when finished,
it belongs in `## Acceptance`. If it tells you what the problem is, it
belongs in `## Problem`.

## Dispatch metadata conventions

The frontmatter fields below make a work order routable and packable by an
orchestrator. Use stable, lowercase, snake_case keys. The framework
standardizes on these conventions:

- `tier`: the capability tier this order needs. It's a closed, ordered
  set (`light | standard | advanced | frontier`, least to most capable).
  Never null. This is the routing field. See the rubric below.
- `tier_reason`: one line saying what about the order drove the choice. A
  retro can audit it.
- `model`: OPTIONAL override, and only when a *named* model is genuinely
  required (a model-specific capability or an explicit user instruction). Normal
  orders carry `tier` and no `model`. See "Why tiers, not model names" below.
- `est_size`: `XS | S | M | L | XL`. It lets the orchestrator pack batches
  by effort independently of tier. It correlates with tier but isn't
  identical.
- `c4_changes`: components whose code this order edits. This is the
  collision key for dispatch isolation: two orders that both change the
  same component collide even when neither blocks the other.
- `c4_impacts`: components affected downstream but not edited.
- `c4_docs`: the C4 doc files that may need updating as a result.

The materializing skill decides how these fields are physically stored:
`groom-beads` maps them onto the bead's metadata JSON and validates the
result against its bead schema. These conventions keep every grooming pass
uniform, so agents don't each invent their own field names.

## Why tiers, not model names

A work order is portable and long-lived: it may be groomed today and
dispatched next quarter, in a project whose model lineup differs from
yours. A named model is the one routing value that doesn't travel. It
changes with every release, and each change forces the same edit across a
schema, a validator, and every skill that reads it.

The tier is the durable claim ("this order needs advanced reasoning"). The
model is a local binding. Each project declares that binding once, in
whatever configuration its orchestrator reads, and the orchestrator
resolves `tier` to a model at dispatch. The binding lives in a config file
rather than in prose because it's a routing fact that a script resolves,
not guidance that an agent reads.

When a project's lineup has fewer than four models, map several tiers onto
one model rather than inventing tier names. The orchestrator's escalation
ladder depends on the tier ordering.

## Tier-routing rubric

The goal is the lowest tier that can complete the order correctly.
Capability, cost, and speed are different dimensions, and only capability
affects correctness. An under-capable model produces wrong work. An
over-capable one produces right work at unnecessary cost. So route on
capability need, then stop. Don't reach for a higher tier as insurance.

Route on signals visible at grooming time. This is a rubric, not a
formula, so explain the call in `tier_reason`.

Reach for the top of the ladder, `frontier` or `advanced`, when the order
has any of these signals:

- High ambiguity or novelty: the spec is the hard part, and the agent must
  make non-obvious design decisions.
- Large blast radius: the order touches many C4 components, or a
  high-impact one (`c4_changes` and `c4_impacts` counts are your proxy).
- Deep reasoning: tricky concurrency, security-sensitive logic,
  algorithmic work, or cross-cutting refactors.
- High cost of being wrong: data migrations, auth, billing, or anything
  hard to undo.

Prefer `advanced` and reserve `frontier` for orders that combine several of
these, or where a wrong answer is expensive to detect.

`standard` is the default, for an order that is:

- Well specified, with clear acceptance criteria.
- Localized to one or two components.
- Mostly execution, not invention: it mirrors an existing pattern in the
  codebase.

`light` fits when both of these hold:

- The work is mechanical and has low ambiguity (rote edits, boilerplate,
  config, simple test additions).
- Throughput or cheap parallel fan-out matters more than deep reasoning.

These defaults and guardrails apply across tiers:

- No work order gets a null tier. If you're unsure, assign `standard` and
  say so in `tier_reason`.
- Blast radius can raise the tier of an otherwise simple order: a "small"
  change to a load-bearing component isn't small in risk.
- Escalate on failure, and never downgrade mid-run. An orchestrator that
  has to raise an order's tier to get it done records that (see below).
  Grooming that had to be corrected signals that this rubric is
  miscalibrated.

## Calibration: the tier is a prediction, and it is checked

`tier` is grooming's prediction of the capability an order needs, and
dispatch tests that prediction. An orchestrator that raises an order's
tier mid-run — because the assigned tier failed gates or visibly flailed —
SHALL report the correction: the order, the assigned tier, the tier that
completed it, and what failed.

Those corrections are the only direct evidence this rubric produces, and
they run in one direction:

- Under-routing (the order had to escalate) is visible and
  self-correcting: the run reports it, and grooming can adjust.
- Over-routing (a `frontier` order a `light` model would have finished)
  is invisible: the work succeeds and nobody learns anything. Its cost
  recurs silently, on every run, indefinitely.

So treat a zero-escalation run across a large queue as a signal to
examine, not as proof that routing is correct. Such a run is consistent
with perfect calibration and equally consistent with routing everything
too high. `retro-spec` and `retro-skill` both take escalation counts as
input.

## A note on estimates vs. tiers

`est_size` answers "how much work?" and drives packing: don't pack ten `L`
orders into one batch. `tier` answers "how much capability per unit of
work?" A large but mechanical order can be `L` and `light`, and a tiny but
subtle order can be `S` and `frontier`. Grooming should set both.
