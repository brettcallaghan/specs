---
title: "RFC: Work orders should anchor by content, not by line number"
description: Grooming produced file:line citations in 14 of 17 beads and every
  one had moved by dispatch, because earlier beads in the same run edited the
  same files
slug: rfc-self-line-anchors-rot-across-a-sprint
type: rfc
status: ready-for-review
updated_at: 2026-08-12T23:05:00Z
origin: self
skill: groom-beads
skill_version: "1.2"
model: claude-opus-5
effort: unrecorded
runs_observed: 1
severity: snag
plugin_version: "0.17.0"
---

# RFC: Work orders should anchor by content, not by line number

## Summary

`groom-beads` groomed 17 beads that `bead-bonanza` then executed in one sprint
over a single skill bundle (run `1786562615-17`). Fourteen of the seventeen
cited `file:line` positions in their Description or Design, and **every cited
position had moved by the time its bead was dispatched**, because earlier beads
in the same run edited the same files. The sprint's own analysis calls this its
dominant friction. Nothing was mis-edited — every implementer relocated by
content — but the Definition of Ready currently asks for "concrete anchors"
without distinguishing anchors that survive a sprint from anchors that cannot.

## Provenance

| Field | Value |
|-------|-------|
| Skill | `groom-beads`, SKILL.md frontmatter `version: "1.2"` |
| Gate enforced | `work-order-definition-of-ready.md` `version: "1.3"`, owned by the `spec` skill and vendored into `groom-beads` |
| Model (grooming) | `claude-opus-5` |
| Models (execution) | `sonnet`, `opus`, `haiku` — the beads were consumed by all three tiers |
| Effort | unrecorded — no artifact of this run records one |
| Harness | Claude Code 2.1.228, remote execution environment |
| Runs observed | 1 |
| Task | Grooming 17 beads for a redundancy-cut sprint over `plugin/skills/bead-bonanza/` |
| Run dates | 2026-08-12, PR #20 |
| Evidence base | `rfcs/sprints/2026-08-12-bead-bonanza-hardening.md`, verbatim findings. Every finding number cited below indexes that file. |

Disclosure: the grooming pass being critiqued here was performed by this same
session earlier in the day. The evidence is the implementers' findings, not the
groomer's self-assessment.

## Evidence

### E1 — every `file:line` anchor a bead carried was stale at dispatch

Shape: **gap**. Severity: **snag**.

Six beads recorded it independently, in their own words:

- `b13` (finding 26): *"the bead's cited line numbers
  (pre-flight-check.sh:269,:284,:455,:494-495; SKILL.md:1106-1113;
  bb-close-bead.sh:159-165) had all moved — four earlier-sprint beads (q1h,
  td1, 9p4, 8mv) touched pre-flight-check.sh. Located each effect by
  content…"*
- `8ef.1` (finding 38): *"Every line number in the description is stale, as
  warned…"*
- `8ef.2` (finding 44): *"every line number in the description had moved, but
  every named passage was still findable by content… The offsets are roughly
  uniform (-72 lines…)"*
- `8ef.3` (finding 53): *"the bead's Category-D line-number citations were
  mostly stale, as warned."*
- `new` (finding 58): *"Every line number in the Design was stale. recovery.md
  was 17,277 bytes / 251 lines, not 18,767 / 263 — the cited ranges… all
  shifted."*
- `b5z` (finding 63): *"Every cited line range was stale. Phase 2.5 was
  SKILL.md:494-726, not :559-818…"*

The Definition of Ready asks for this at check 1: a Problem statement "SHALL
name concrete anchors — file paths, function/class names, endpoints, tables,
user-facing behavior". A line number reads as the most concrete anchor
available, so a groomer reaching for concreteness reaches for it. The check
does not say that concreteness and durability are different properties.

The mitigation that worked was not in the skill: the orchestrator added
*"locate by content, not by line number, and report the drift"* to every
dispatch prompt. That it worked is why nothing was mis-edited — and why
this is a snag rather than a blocker. But it was improvised per-run by an
orchestrator
that happened to anticipate the problem.

The failure mode when it is *not* anticipated is not a mis-edit; it is silence.
An agent that trusts a stale range edits the wrong region, or reports the cited
content "already fixed" because what is at those lines now is something else.

## Proposals

### P1 — the Definition of Ready distinguishes durable from volatile anchors

Severity: **snag**. Shape: gap. Evidence: E1.

**Change.** In `work-order-definition-of-ready.md` check 1, after the existing
"concrete anchors" sentence, add:

> An anchor SHALL be durable under edits to the same file. File paths, symbol
> names, headings and short quoted strings are durable; line numbers and line
> ranges are not, and neither are byte counts or item counts. WHERE a work
> order names a position, it SHALL name the content at that position — quote
> the heading or the first few words — and MAY give a line number only as a
> convenience alongside it. In a queue where several orders touch one file,
> every line number is stale before the second order is dispatched.

The source lives in the `spec` skill and is vendored into `groom-beads` and
`groom-feature-blueprint` by `scripts/sync-shared-references.sh`; change the
source and regenerate, never the vendored copies.

### P2 — `groom-beads` rejects a bare line-number anchor at grooming time

Severity: **polish**. Shape: gap. Evidence: E1.

**Change.** Add a bare line reference to the anti-patterns `groom-beads`
actively rejects, alongside session-dependent language:

> A `file:line` citation with no quoted content — the position is stale as
> soon as any earlier order edits that file, and the reader has nothing to
> relocate by. Inline the content the line held.

This is the enforcement half of P1: the DoR defines ready, `groom-beads` is
where a work order is brought to it.

## Considered and Not Proposed

- **Requiring the orchestrator's "locate by content" line in every dispatch
  prompt.** It is a real mitigation and it worked, but it treats the symptom
  in the consumer rather than the defect in the artifact. If P1 lands, a
  well-groomed order does not need it. Recorded because if P1 is rejected,
  this becomes the fallback and belongs to `bead-bonanza`, not here.
- **Beads whose *counts* were wrong** — `8ef.1` promising eighteen rules and
  naming fourteen (finding 37), `8ef.2`'s B9 claiming a duplicate that occurs
  once (finding 46), `8ef.3`'s permission-entry claim that does not hold
  (finding 54). A related defect — an acceptance criterion keyed to a count is
  not checkable — but it is about asserting quantities without counting them,
  not about anchor durability, and one run is thin evidence for a second
  proposal. Recorded so a recurrence has something to attach to.
- **Beads whose premises the sprint itself invalidated** (`b13`'s facts
  depended on `8mv`'s outcome; `9ko`'s premise on `41z`'s). Real, and recorded
  in the sprint report — but the fix is a content-dependency layer in the
  dispatch plan, which is `bead-bonanza`/`plan-sprint` territory, not a
  grooming defect. The dependency graph correctly expressed the ordering.

## Non-Goals

- Changing how dependencies or tiers are expressed. Both worked this run.
- Retro-fitting the 17 beads already groomed; they are closed.

## Adoption

P1 is a paragraph in one spec-owned reference plus a regeneration run; the
vendored copies are mechanically verified by
`scripts/sync-shared-references.sh --check`, so drift is detectable. P2 is one
bullet in `groom-beads`. Neither invalidates an existing work order: an order
carrying a line number stays valid, it simply would not be groomed that way
again.

## Resolution

<!-- Filled at triage. -->
