---
title: Findings Capture Protocol
description: How agents record retro evidence in-flight during
  multi-agent phases — the finding format, when to record, where, and
  how the aggregated retro harvests them
slug: findings-capture-protocol
type: standards
status: active
updated_at: 2026-07-23T00:00:00Z
version: "0.2"
---

# Findings Capture Protocol

Retro evidence does not survive being recalled later, and in a
multi-agent phase (implementing a groomed plan, where each work order
may go to a different spawned agent) no single session holds it. This
protocol is how that evidence is captured at the moment of friction.

This document is deliberately self-contained: an agent implementing
one work order needs ONLY this file — not the `retro-spec` skill that
consumes the findings, and not the `groom-feature-blueprint` skill that produced
the item. Work orders reference the copy of this document vendored
into the project (`rfcs/findings-capture.md`).

## The finding

One finding per friction point, recorded with zero analysis:

> What happened + the item or spec text that caused it (quoted or
> precisely located) — or the explicit note that no guidance existed
> where it was needed.

Friction only. No root-causing, no proposals, no RFCs: the capturing
agent is a sensor, and analysis is the aggregated retro's job. A
finding SHOULD be 1-3 sentences.

## When to record

The implementing agent SHALL record a finding WHENEVER it:

- **needed information not in the work order or its listed
  references** — had to open the blueprint, FRD, or code to fill a
  gap the item should have carried. Record even small occurrences:
  each is direct evidence about whether items are self-sufficient
  enough to delegate.
- **made a design decision** the item or blueprint should have
  pinned (a path, a default, a scope cut, an edge-case rule).
- **found the item contradicting repo reality** — stale anchors,
  moved code, a referenced element that no longer matches.
- **could not satisfy an AC or citation obligation as declared** —
  unclear verified-by level, a test that cannot meaningfully exist,
  verification blocked on work outside the item.
- **hit a sequencing surprise** — inherited a broken build from a
  predecessor, or blocked on work the dependency graph did not
  encode.
- **found the item mis-scoped** — too large to review as one unit,
  trivially small, or overlapping another item's work.

A clean run is fine — record nothing. Do not manufacture findings.

## Where to record

- **Beads backend:** one `bd comment` per finding on the item's own
  issue, prefixed `retro:`.
- **Task-file backend:** append each finding as a bullet under a
  `## Retro findings` section at the end of the item's own file
  (per-item, so parallel agents never contend on a shared file).

Findings remain on closed items permanently as raw provenance; they
SHALL NOT be deleted after aggregation.

## Aggregation (for the retro-running agent)

The plan's final Definition-of-Done item runs ONE retro for the whole
implementation phase (the `retro-spec` skill; never a retro per work
item). It harvests the findings across the feature's closure set —
`bd list --label <feature-slug>` then each issue's `retro:` comments,
or each task file's Retro findings section — adds its own
DoD-verification friction, and treats that pool as the phase's
evidence.
