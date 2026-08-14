---
title: "RFC: Two wave collisions the C4 co-edit analysis cannot see — a shared
  interface that does not exist yet, and a bead that measures what another changes"
description: Wave packing tests for shared c4_changes, which is a file-level
  test, so two beads that build one interface and a bead that measures a tree
  another mutates were both judged safe to run concurrently.
slug: rfc-self-collisions-the-c4-check-cannot-see
type: rfc
status: ready-for-review
updated_at: 2026-08-14T09:45:00Z
origin: self
skill: plan-sprint
skill_version: "0.1"
model: claude-opus-5
effort: unrecorded
runs_observed: 1
severity: snag
plugin_version: "n/a (in-repo dogfooding)"
---

# RFC: Two wave collisions the C4 co-edit analysis cannot see

## Summary

`plan-sprint` packs concurrent waves by testing whether two beads share a
`c4_changes` component. That is a test about *files a bead edits*. Run
`1786686364-15` produced two collisions that this test cannot express, and both
were found after the fact rather than at planning time.

Neither cost the run a wrong result — one was caught at a code seam by the
downstream bead's dispatch prompt, the other by an orchestrator noticing a
measurement was about to run against a moving tree. Both were luck, and both
are mechanical enough to plan for.

## Provenance

| Field | Value |
|---|---|
| Harness | Claude Code (Claude Code on the web, remote execution environment) |
| Task the skill was applied to | Phase 2.5 planning for a 16-bead, 7-wave bead-bonanza run across two epics — a CLI feature and a documentation move — on branch `claude/validate-corpus-spec-cli-0t0cxo` |
| References / scripts / templates invoked | `plan-sprint` v0.1 (invoked from `bead-bonanza` v1.5 Phase 2.5); the wave table and `## Amendments` in `/tmp/claude/bead-bonanza/1786686364-15/dispatch.md` |
| Run dates | 2026-08-14 |
| Model lineup (from `dispatch.md` manifest) | light `haiku`, standard `sonnet`, advanced `opus`, frontier `fable`; `CLAUDE_CODE_SUBAGENT_MODEL` unset, so tier assignments were honoured |
| Testimony of record | `rfcs/sprints/2026-08-14-markdown-links-and-shipped-specs.md` — the run's sprint report. Findings below are cited by its `<bead-id>#<ordinal>` ids; where this RFC quotes one, the report is authoritative |

## Evidence

### E1 — two beads built one interface that did not exist yet [gap, snag]

The plan placed `agent-plugins-ia5.2` (target resolution) and
`agent-plugins-ia5.4` (link-map reading) in the same concurrent wave. The
recorded justification, from the run's wave table:

> | 2 | agent-plugins-ia5.4 | shared | concurrent-wave #2 | standard | defer |
> same c4_changes: container-spec-cli as ia5.5, which is in wave 4 |

They edit different components (`module-document-core` and
`container-spec-cli`), so the co-edit test passed. But `ia5.4` *produces* the
link map and `ia5.2` *consumes* one, and neither could see the other's types
while running. `ia5.2` recorded the consequence (`ia5.2#2`):

> With no link-map type in existence yet, it defined its own `mdlinks.LinkMap`
> (`map[string]string`) and leans on Go's nil vs non-nil map distinction to
> separate "no map supplied" from "a map supplied that resolves zero entries" —
> AC-SPECMD-006.3 requires the latter to still raise `links/unmapped`.

`ia5.4` independently built `cli/cmd/linkmap.go`. Nothing broke at the merge —
different packages, both gates green — but the semantics at the seam
(non-nil-empty vs nil) were settled twice, independently, and only agreed by
coincidence. `ia5.4` also split the map-reading fault into two rule ids where
the design named three faults together, so the consuming bead's expectations
were stale before it started.

The hazard was caught only because the orchestrator wrote it into `ia5.5`'s
dispatch prompt by hand, as an amendment:

> **Its dispatch prompt must carry this**: MAP must hand `mdlinks` a
> **non-nil** map even when the YAML decodes to zero entries, or the two
> components disagree about whether a map was given and `links/unmapped`
> silently stops firing.

No rule in `plan-sprint` produces that. A shared *interface* that does not
exist yet has no `c4_changes` entry to collide on, because neither bead is
editing it — they are both creating it.

### E2 — a bead that measures a surface collided with one that changes it [gap, snag]

Sprint report §3 records this pair. `agent-plugins-47b` builds a script that walks
`plugin/skills/*/references/**` and records a floor on how many files it
examined, so a glob that stops matching fails loudly. `agent-plugins-eaa.4`
moves a document into that same tree.

Both carry `c4_changes: []` — neither touches a modelled component — and they
co-edit not one file: `47b` only *reads* the tree. Every test `plan-sprint`
applies says they are safe to run concurrently. Running them concurrently would
have computed a file-count floor against a tree in motion, baking a number that
was true for no commit.

This was caught by hand while splitting wave 6, and recorded as an amendment:

> **a bead that measures a surface collides with a bead that changes it,
> without co-editing a single file.**

The wave was split into 6a (`ia5.6` + `eaa.4`) and 6b (`47b`) as a result.

## Proposals

### P1 — add a "produces/consumes an interface" collision test [snag]

**Change** — in `plan-sprint`, alongside the `c4_changes` co-edit test, add:

> **Interface seams.** Two beads collide WHEN one produces a type, format or
> API the other consumes, even when they change different components and share
> no file — because neither can see the other's definition while running, so
> each settles the seam's semantics independently. Detect it from the beads'
> own text rather than from C4: a Design that says it *reads* / *accepts* /
> *is handed* something another bead's Design says it *produces* / *emits* /
> *builds* is a seam.
>
> Resolve it one of two ways, and record which:
> - **Sequence them** — the producer first, so the consumer sees real types.
> - **Keep them concurrent and name the contract in BOTH dispatch prompts**,
>   in the same words, including the edge cases (empty vs absent, one error or
>   several). A seam settled twice agrees only by luck.
>
> A wave table row for a concurrent seam SHALL state the contract it pinned.

### P2 — treat "measures a surface" as a collision key, not only "changes" it [snag]

**Change** — extend the collision computation:

> `c4_changes` answers "what does this bead edit". It does not answer "what
> does this bead *measure*". A bead that counts, globs, snapshots, or asserts a
> floor over a tree collides with any bead that adds to or removes from that
> tree, and neither `c4_changes` nor a co-edit analysis can see it: the
> measuring bead edits nothing in that tree.
>
> WHEN a bead's Acceptance names a count, a floor, an inventory, or a
> "regenerate and review the diff" step over a path, record that path as a
> **measured surface** and treat it as a collision key against any bead whose
> `c4_changes`, Design or Acceptance writes under it. Sequence the measurer
> last.

### P3 — say in the wave table WHY a concurrent pair is safe, not only why a pair is separated [polish]

Every row in this run's wave table justified *separation*. No row asserted
*safety*, so a pair that collided in a way the tests do not model produced no
artifact anyone would review — the plan looked complete because the tests it
ran all passed.

**Change**:

> A `concurrent-wave` row SHALL state the basis on which its bead is safe
> beside its wave peers — not merely that no `c4_changes` overlap was found.
> "Disjoint surfaces: `<paths>` against `<paths>`" is the minimum. The absence
> of a detected collision is not a safety argument, and writing one down is
> where an undetected class becomes visible.

## Considered and Not Proposed

- **Requiring every concurrent wave to be size 1** (i.e. abandoning
  concurrency). Rejected: this run's concurrency was correct in 5 of 7 waves
  and is where its throughput came from. The fix is better collision keys, not
  fewer waves.
- **Deriving seams automatically from the code rather than the bead text.**
  Rejected as out of scope for a planner that runs before any code exists —
  at Phase 2.5 the only evidence is the beads' own Problem/Design/Acceptance.
- **Treating `c4_impacts` as a concurrency collision key.** Considered, because
  it is adjacent to E1. Rejected: this run showed `c4_impacts` already
  disagreeing with a wired `blocks` edge (`ia5.1` impacts `container-spec-cli`,
  which `ia5.4` changes, while the real edge runs `ia5.4 ← ia5.1`), so
  promoting it to a collision key would have produced a false collision here.
  The dispatch.md amendment records that the `blocks` edge correctly won.

## Non-Goals

- The dependency-edge derivation. Every one of the ten edges this run planned
  was verified against the tracker and none was fabricated; the wave table
  records that check. No finding.
- Tier assignment and the routing rubric — no escalations occurred, so this run
  produced no direct routing evidence either way.
- Anything about how beads are dispatched or closed; those belong to
  `bead-bonanza` and are filed separately.

## Adoption

P1 and P2 are independent and can land in either order; both add a collision
key to the same computation. P3 is a documentation requirement on the wave
table's output format and can land alone.

None changes the existing `c4_changes` test, so a plan that produced no
seam or measured-surface collision is packed identically to today. Expected
effect on this run: `ia5.2`/`ia5.4` either sequenced or dispatched with one
stated contract, and `47b`/`eaa.4` separated at planning time rather than by
an orchestrator noticing during wave 6.

## Resolution
