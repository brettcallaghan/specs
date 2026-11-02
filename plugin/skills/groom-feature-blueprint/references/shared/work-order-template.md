<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: plugin/skills/spec/references/work-order-template.md
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: "Work Order: <imperative summary of the unit of work>"
description: <One-sentence summary of the work this order encapsulates>
slug: work-order-<identifier>
                                     # No `type` property: see "Non-corpus"
                                     # below — do not add one.
kind: task                           # epic | feature | task | bug | chore
priority: 2                          # 0 (critical) .. 4 (backlog)
updated_at: <YYYY-MM-DDTHH:MM:SSZ>

# Provenance — where this work order came from
sources: []                          # feature-<slug> WHEN decomposed from a
                                     # feature blueprint; empty when standalone
delta: <KEY>                         # delta key from the blueprint's AC
                                     # Coverage legend (blueprint-sourced only)

# Dependency graph — typed edges to other work orders, by slug
parent_slug: <work-order-epic-slug>  # the epic this order belongs to
depends_on: []                       # orders that must complete first
blocks: []                           # orders this one blocks
discovered_from: <work-order-slug>   # order whose execution surfaced this one

# Dispatch metadata — what the orchestrator reads to route and pack
tier: standard                       # light | standard | advanced | frontier
                                     # (least to most capable) — never null.
                                     # Assign the LOWEST tier that can do the
                                     # work; the orchestrator resolves tier ->
                                     # model from the project's model lineup.
tier_reason: <one line — what about this order drove the tier choice>
# model: <name>                      # OPTIONAL override — set only when a
                                     # specific named model is required.
est_size: M                          # XS | S | M | L | XL
c4_changes: []                       # C4 components whose code this order edits
c4_impacts: []                       # affected downstream, not edited
c4_docs: []                          # C4 doc files that may need updating
---

<!--
Template: Work Order
Source: This specification framework.

Load before authoring
  - `shared/work-order-field-guide.md`: what each field holds.
  - `shared/work-order-definition-of-ready.md`: the readiness gate.

Sections (complete each, then delete its instruction block)
  - Problem
  - Design
  - Acceptance
  - Retro

Purpose
  A Work Order is the atomic, self-contained unit of executable work: the
  smallest thing that can be handed to a fresh agent with zero session
  context and completed correctly — right approach, verifiable done-ness —
  without asking anyone anything. One work order maps to exactly one tracker
  issue (one bead).

  The work order exists to make a groomed unit of work PORTABLE — a concept
  independent of any one tracker. Everything the executing agent needs rides
  in the order itself, never in a planning conversation that will not exist
  when the order is dispatched.

Reader profile (apply with the tech-writer skill)
- Reader: a fresh agent with zero session context, dispatched to do this one
  job.
- Knows: only what this work order says and the files it names.
- Wants: the job, how to prove it's done, and every fact needed to finish
  without asking.
- Budget: read in full before starting, then consulted while working.
- Tone: imperative and exact. Name files, commands, and values. Never point at
  a conversation the agent didn't see.

This document is GENERATED, not hand-authored
  Unlike every other spec type, a work order is NOT authored through the
  `spec` skill's authoring workflow. It is PRODUCED by a grooming skill:
    - `groom-beads` brings any unit of work (standalone bug/chore, or a
      backlog issue) to the Definition of Ready and materializes it.
    - `groom-feature-blueprint` decomposes an approved feature blueprint into
      work orders, then leans on `groom-beads` to finish them to the DoR.
  This template is the canonical FORMAT those skills target and validate
  against; the `spec` skill owns the format and the readiness bar, the
  grooming skills own production.

Non-corpus: no `type` frontmatter property
  Every other template in this framework but the issue template declares
  `type` from the DocumentType vocabulary (adr, frd, contract, standards,
  the blueprint kinds, guide), which is how the schema and
  `scripts/check-template-drift.py` govern it. A work order deliberately
  has no `type`: it materializes into a tracker issue (a bead) and stops
  existing as a document, so corpus membership — persisting, being cited
  by slug, appearing in the index — does not apply to it.
  `groom-beads`'s bead schema governs its actual shape.
  REVISIT WHEN a project has no tracker and work orders persist as files
  instead of materializing into beads (see "Storage" below) — corpus
  membership should be reconsidered then. Until that path exists, do NOT
  add `work-order` to the DocumentType vocabulary.

Storage — beads now, files later
  A work order is a LOGICAL schema, not (yet) a file on disk. Today it is
  materialized into a tracker issue: the body sections map to the bead's
  text fields (`description` / `design` / `acceptance`), the dependency
  graph maps to typed bead edges, and the dispatch-metadata frontmatter maps
  to the bead's metadata JSON. `groom-beads` owns that mapping — see its
  field guide. WHEN a project has no tracker, the same schema will serialize
  as a standalone markdown file (transient), but that path is deliberately
  out of scope for now.

Self-containment is the whole point
  The reason the dispatch metadata (tier, C4 blast radius, est_size,
  provenance, delta key) lives in the work order rather than only in a
  planning agent's head is that the dispatching agent has no head to read.
  A field left "to figure out later" fails, because "later" is inside an
  isolated agent that cannot ask.

When to use
  - Representing a single unit of work ready for isolated agent execution
  - The output shape of any grooming pass, whether the work is standalone
    or decomposed from a feature blueprint

When NOT to use
  - Authoring a feature's design (that is a feature blueprint) or its
    requirements (that is an FRD) — a work order is downstream of both
  - Recording a decision (ADR) or an interface ([[contract-]]…)
  - Hand-authoring: if you are writing one of these by hand rather than
    grooming it, you are probably at the wrong altitude — capture the work
    in the tracker and groom it

Reference syntax
  [[slug]]            — other documents (blueprints, FRDs, contracts, ADRs,
                     and other work orders by their work-order-<id> slug)
  [[container-<name>#Component Name]]
                   — a C4 component, by a wikilink heading anchor into
                     its parent's Components roster); c4_changes / c4_impacts
                     list the same component names
  `ElementName`    — code-level anchors: paths, symbols, endpoints, tables

Slug format
  wo-<identifier>, lowercase kebab-case. WHEN materialized into a
  tracker that assigns its own id, the tracker id is authoritative and this
  slug is the stable human handle.

UNRESOLVED — no lifecycle status field (deliberate deviation)
  `shared/frontmatter-standards.md` makes `status` a REQUIRED frontmatter
  property. This template omits it on purpose: a work order has two
  orthogonal lifecycles — grooming readiness and execution state — and both
  are owned by the tracker (bead `status` + `grooming_status`), not
  duplicated here, to avoid a two-copy drift problem. This is a known open
  question, not an oversight: WHEN the file-backed (no-tracker) path is
  designed, the status model SHALL be revisited. Until then, "is this order
  ready?" is answered by running the Definition of Ready, and the
  outcome is recorded on the tracker issue, not in this frontmatter.

Remove all HTML comment blocks after generating a work order.
-->

# <imperative summary of the unit of work>

## Problem

<!--
The IMMUTABLE problem statement: what & why. This is the field a fresh agent
reads first, so it SHALL survive the fresh-agent test — name concrete anchors
(file paths, symbols, endpoints, tables, user-facing behavior), never point at
a conversation ("as discussed", "the file above", "finish the earlier work").
Anchors SHALL be durable under edits to the same file: name a position by the
content at it — quote the heading or the first few words — and give a line
number, range or count only alongside that quote, never instead of it.

By `kind`:
  - bug   — steps to reproduce, expected vs. actual, where it manifests
  - feature / task — the concrete change and the user-facing intent
  - chore — the concrete change and what "done" looks like
  - epic  — the coherent goal only; an epic is not directly executable and
            carries child orders instead of Design / Acceptance below

WHEN this order was decomposed from a feature blueprint, link the blueprint
and FRD here ([[feature-<slug>]], [[frd-<slug>]]) so the agent can reach design
context without searching. Maps to the bead `description`.
-->

<The self-contained problem statement.>

## Design

<!--
The MUTABLE approach: how. The chosen approach and the decisions that would
otherwise be re-litigated by a time-pressured agent. Record the alternatives
weighed AND the one chosen; it SHALL NOT leave the choice open ("leaning X
because Y; fall back to Z if that doesn't work out" is an open decision and
fails the Definition of Ready). A fallback a fresh agent can resolve
mechanically from the repo is fine. Name existing code anchors;
do NOT prescribe new function signatures or file layouts — that is the
executing agent's choice. Omit for `kind: epic`. Maps to the bead `design`.

WHEN the owning blueprint's Design section names patterns
("Name each pattern with its progenitor, constraint, and need" in the
content standards), carry each pattern line the change touches into
Design, with its name, progenitor, and constraint. Leave out the need,
which the blueprint records.

If a line tells you WHAT TO TYPE, it belongs here. If it tells you WHAT MUST
BE TRUE when finished, it belongs in Acceptance. See the field guide.
-->

<The chosen approach and key decisions.>

## Acceptance

<!--
The DEFINITION OF DONE: outcome-focused, checkable, implementation-agnostic
boxes — observable end states, not steps. This is the shared contract by which
the isolated agent knows it is finished and the orchestrator/verifier confirms
it. Required (>=1 box) for `kind: feature` and `kind: bug`. Omit for
`kind: epic`. Maps to the bead `acceptance`.

WHEN this order changes a C4 component, include a box to update that
component's C4 doc (the c4_docs listed in frontmatter).

WHEN Design carries pattern lines from the owning blueprint, include a box
that the change conforms to each pattern's constraint.

  Good: - [ ] Expired tokens return 401
  Bad:  - [ ] Edit middleware.go   (a step — belongs in Design)
-->

- [ ] <observable end state>
- [ ] <observable end state>

## Retro

<!--
The standing instruction to the executing agent (often a subagent that sees
only this order) to record friction findings before closing. Reinforces the
retro pipeline: the implementer is a sensor for work-order quality. The exact
text and mechanism are backend-specific and STAMPED BY THE GROOMING SKILL at
materialization (a `bd comment` prefixed `retro:` for beads). It is short and
self-sufficient — it does not point at a protocol file, because an order that
sends its agent elsewhere to learn the format is an order that gets skipped.
Keep it verbatim as the grooming skill emits it; do not paraphrase.
-->

<Grooming stamps the backend-specific retro instruction here.>
