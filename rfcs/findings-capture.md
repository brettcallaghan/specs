<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: plugin/skills/retro-spec/references/findings-capture-protocol.md (version 0.3)
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->
<!-- Sync note (2026-08-11, agent-plugins-92q.18): this path is not
     actually covered by scripts/sync-shared-references.sh's MAP
     (verified: --check reports 11 vendored references and this file is
     not among them), so the header above is aspirational rather than
     enforced here. The Aggregation section was hand-synced to match
     source v0.3's corrected harvest command. The same correction was
     also sent upstream as an RFC against retro-spec
     (rfcs/outbox/retro-spec/); once adopted there, re-verify this copy
     still matches. -->

---
title: Findings Capture Protocol
description: How agents record retro evidence in-flight during
  multi-agent phases — the finding format, when to record, where, and
  how the aggregated retro harvests them
slug: findings-capture-protocol
type: standards
status: active
updated_at: 2026-08-11T00:00:00Z
version: "0.3"
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
item). It harvests the findings across the feature's closure set with
one `bd export | jq` pipeline, never `bd list` plus `bd show --json`:
`bd list` defaults to open issues, and a drained closure set is almost
entirely closed; `bd show --json` omits comments entirely even though
the text output prints them. `bd export` carries comments as a
`comments` array on the issue record and covers every status in one
pass:

```
bd export | jq -r 'select((.labels//[]) | index("<feature-slug>"))
  | .comments[]? | select(.text|startswith("retro:")) | .text'
```

(Task-file backend: each task file's Retro findings section instead.)
Assert the count is non-zero before aggregating — a harvest that
returns zero findings across a completed feature is a broken query,
not a clean run, and an aggregated retro built on it is worse than
none. Then add the DoD item's own verification friction, and treat
that pool as the phase's evidence.
