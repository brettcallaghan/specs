---
title: "RFC: The run directory is a measurable record of the run, and it
  contradicts two things the skill still assumes"
description: A file-level audit of a completed 31-bead run directory shows the
  per-bead retro FILE cost 14x the progress ledger per bead, the ledger's tail
  is 94% accurate but wrong exactly where it matters, and the one bead that
  lost its crash-recovery anchor is the one that needed it
slug: rfc-self-run-directory-as-evidence
type: rfc
status: ready-for-review
updated_at: 2026-08-12T00:00:00Z
origin: self
skill: bead-bonanza
skill_version: "1.3"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: minor
plugin_version: "0.14.0"
---

# RFC: The run directory is a measurable record of the run

## Context

Run `1786345424-bbh` drained 31 beads and left 67 files. Before clearing it at
the end of the run, the directory was audited as data rather than read as prose.
It answers questions the skill currently settles by assertion.

## Finding 1 — the artifact transition is visible, and the run directory dates it

Seven beads have a `<id>.retro.md` and no `<id>.handoff.md`. Twenty-three have a
handoff and no retro file. The two sets do not overlap at all, and the seven are
the earliest beads in the run.

That boundary is the moment the per-bead retro FILE was replaced by
`handoff.md`, visible in the artifacts without consulting git or the tracker. A
run directory therefore records not just what a run did but **which version of
the ceremony each bead ran under** — useful when a mid-run skill change is
suspected of causing a defect, which happened twice in this epic.

**Proposal (low cost):** the Phase 4 report should state the ceremony version
each bead ran under when the skill changed mid-run, derived from which artifacts
exist. Today a reader has to infer it.

## Finding 2 — the retro FILE cost 14x the ledger per bead, independently confirmed

| Artifact | Files | Total bytes | Bytes/bead |
|---|---|---|---|
| `<id>.retro.md` | 7 | 31,464 | 4,495 |
| `<id>.progress.jsonl` | 31 | 9,651 | 311 |

The seven retro files cost more than three times the entire progress ledger for
a run four times their size.

The epic already retired the retro file on a measurement from a *different* run
(10 files, 137 KB, 1 cited finding, against 34 `retro:` comments yielding 12 of
12). This is independent corroboration from a second run, and it strengthens the
case: the retirement was not an artifact of one unusual run.

**No action needed** — the file is already retired. Recorded so the decision is
not revisited on the assumption it rested on a single observation.

## Finding 3 — the ledger's tail is 94% accurate, and wrong exactly where it matters

Last recorded phase across 31 beads: `committed` 29, `retro-written` 1,
`implementing` 1.

So a naive "trust the last phase" reading is right 29 times out of 31. The two
exceptions are the orphaned dispatch that cost this epic 9.4 hours, and one
other incomplete bead — that is, the tail is accurate precisely when nothing is
wrong and misleading precisely when something is.

This is the strongest available argument for the existing rule that the ledger
is a hint and never the authority, and it is a better argument than the one the
skill currently gives, which is anecdotal. **Proposal:** state the 29/31 figure
where that rule is stated. A rule that is right 94% of the time needs a reason
to distrust it, or a reader will start trusting it.

## Finding 4 — the crash-recovery anchor was missing on the one bead that crashed

Exactly one bead has a progress ledger but neither a handoff file nor a retro
file: the bead whose dispatch was orphaned. Its ledger stops at `retro-written`.

The handoff file is described as a non-negotiable crash-recovery anchor written
BEFORE the close. In the single instance in this run where crash recovery was
actually needed, it was absent — the agent died between emitting the phase and
writing the file, or wrote neither.

**Proposal:** the phase and the file should not be separable. Either
`bb-progress.sh <id> retro-written` should refuse to record the phase unless the
handoff file exists, or the close ceremony should write a minimal handoff itself
when one is missing rather than trusting that a prior step produced it. The
current ordering makes the anchor most likely to be absent in exactly the
scenario it exists for.

## What this does not show

The directory says nothing about wall-clock, model tier, or retries — no
timestamps are recorded in the ledger beyond ordering. Any future proposal about
run *pacing* cannot be settled from this artifact set, and would need the ledger
to carry timestamps first. That is a deliberate prior decision, not an oversight,
and this RFC does not reopen it.

## Provenance

Derived by file-level audit of `.bonanza/1786345424-bbh/` (67 files, 328 KB)
immediately before the directory was cleared at the end of the run. All figures
are counts over that directory, reproducible from the snapshot committed on the
run branch. Model withheld by session policy.
