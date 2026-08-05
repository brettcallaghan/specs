---
title: "RFC: The readiness checklist validates no frontmatter value and
  no verified-by level, and certified a blueprint defective in both"
description: Ten checks read the blueprint's frontmatter without
  validating any value against its vocabulary, and none reconciles a
  manual verified-by level against the FRD marker that alone exempts it
  from citation
slug: rfc-self-readiness-checklist-frontmatter-and-manual-acs
type: rfc
status: ready-for-review
updated_at: 2026-08-04T00:00:00Z
origin: self
skill: spec
skill_version: "1.6"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.14.0"
related: []
---

# RFC: The readiness checklist validates no frontmatter value

## Summary

Running `groom-feature-blueprint` v1.2 over an approved feature blueprint
exercised `blueprint-readiness-checklist.md` v0.5.0 as a Phase 1 gate —
the document the `spec` skill owns and runs as its own authoring
preflight. Two defects in the blueprint passed all ten checks. One was an
out-of-vocabulary frontmatter value (`delivery: approved`, where the
property admits only `single-pr` and `staged`), caught incidentally
because a *different* check happened to read a neighbouring property. The
other was an acceptance criterion declared verified at level `manual`
whose FRD lacks the `(manual)` marker that alone exempts it from the
citation obligation — an unsatisfiable requirement the gate certified as
plannable and grooming had to carry forward onto a work order.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code on the web (remote execution environment) |
| Task the skill was applied to | `blueprint-readiness-checklist.md` run as `groom-feature-blueprint`'s Phase 1 gate over `specs/features/spec-index/spec-index-blueprint.md` (51 acceptance criteria, 12 delta blocks) before decomposition into 16 work orders |
| References / scripts / templates invoked | `blueprint-readiness-checklist.md` v0.5.0, `ac-traceability-standards.md` v0.2.0, `feature-blueprint-template.md` v1.3, `frd-template.md` (via `ac-traceability-standards`), `groom-feature-blueprint` v1.2 |
| Run dates | 2026-08-04 |

The `model` frontmatter property is withheld: this session runs under a
harness policy forbidding the model identifier in repository artifacts.
The placeholder follows the convention established by
`rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md`, the open
RFC covering that gate conflict. `effort` is `unknown` because the
harness does not expose it to the running agent.

## Evidence

1. **[gap · blocker] No check validates a frontmatter value against its
   vocabulary.** The blueprint carried `delivery: approved`. The
   `delivery` property admits exactly two values —
   `feature-blueprint-template.md:9` declares
   `delivery: single-pr  # single-pr | staged. The delivery` —
   and the origin project generates a JSON Schema enforcing it
   (`schemas/vocabulary.strata:130` `enum Delivery`, rendered as
   `{"enum": ["single-pr", "staged"]}`). All ten checks in
   `blueprint-readiness-checklist.md:42-85` passed on this property:
   checks 1-3 read `status` only, and checks 4-10 examine body sections.
   The defect surfaced only because check 2
   (`blueprint-readiness-checklist.md:43`, "The blueprint's status is
   `approved`.") read the adjacent `status` property, found
   `ready-for-review`, and stopped the run — at which point the agent
   noticed the transposition while reporting it. Had `status` been
   correct, `delivery: approved` would have reached decomposition, where
   `groom-feature-blueprint`'s rule at `SKILL.md:313-316` ("The delivery
   unit is the blueprint's `delivery` declaration: `single-pr` (assumed
   when the field is absent) or `staged`") has no branch for a third
   value.

2. **[gap · blocker] No check reconciles a `manual` verified-by level
   against the FRD's `(manual)` marker.** The blueprint's AC Coverage
   table declared AC-SPEC-009.4, AC-SPEC-009.5 and AC-SPEC-009.6
   verified at level `manual`. `ac-traceability-standards.md:106-109`
   states: "IF an acceptance criterion can only be verified manually,
   THEN the FRD SHALL record that designation by appending the marker
   `(manual)` to the criterion text, and the criterion is exempt from
   the citation obligation." The FRD appends no such marker to any of
   the three. Check 4
   (`blueprint-readiness-checklist.md:55-62`) validates that every FRD
   acceptance criterion appears in the coverage table and that each row
   names a real delta — it does not look at the `Verified by` column at
   all. So the gate certified as plannable a blueprint in which three
   criteria are simultaneously unverifiable by any test
   (`ac-traceability-standards.md:87-88`: "Every acceptance criterion of
   an FRD SHALL be cited by at least one test before the implementing
   change merges") and ineligible for the only exemption. Grooming could
   not resolve this without making an authoring decision, so it wrote
   both resolution paths onto the BUILD work order as an open obligation
   and flagged it — deferring a contradiction the gate should have
   caught one phase earlier.

## Proposals

### P1 — Add a frontmatter validity check [blocker]

**Problem.** `blueprint-readiness-checklist.md:42-85` defines ten
checks. Three read `status`; none validates any other frontmatter
property against its documented vocabulary, though
`feature-blueprint-template.md` documents a closed vocabulary for
`delivery` (line 9) and for `status`. The checklist's own preamble
(lines 24-29) positions it as "the single definition of 'ready to
groom'" run both as an authoring preflight and as a groom-time gate —
so a frontmatter defect escapes both consumers (finding 1).

**Change.** Insert as check 11 in the content-checks section:

> 11. Every frontmatter property whose value is drawn from a closed
>     vocabulary holds a value from that vocabulary, and every required
>     property is present. WHERE the project publishes a machine-readable
>     schema for its document types, validating the blueprint against it
>     satisfies this check; otherwise compare each controlled-vocabulary
>     property against the values `feature-blueprint-template.md`
>     documents for it.

Add to the decision-sweep section a one-line note that this check is
mechanical and should be run first, since it is the cheapest of the
eleven and its failure invalidates reasoning built on the frontmatter
(the delivery model, in particular, drives grooming's whole sequencing
approach).

**Evidence.** Finding 1. On the next run `delivery: approved` fails
check 11 explicitly and is reported as an invalid value rather than
being noticed sideways because a neighbouring property happened to be
wrong too — and a blueprint whose `status` is correct no longer carries
an invalid delivery model into decomposition.

### P2 — Reconcile manual verified-by levels against the FRD [blocker]

**Problem.** Check 4 (`blueprint-readiness-checklist.md:55-62`)
validates the AC Coverage table's rows and delta references but ignores
the `Verified by` column. `ac-traceability-standards.md:106-109` makes
the FRD's `(manual)` marker the sole exemption from the universal
citation obligation at lines 87-88. Nothing connects the two, so a
blueprint may declare a criterion `manual` while its FRD leaves it
citable-by-obligation and uncitable-in-fact (finding 2). This is
precisely a "design decision the blueprint does not record" — the
category check 7 exists to catch — but check 7's decision sweep
(lines 87-107) enumerates paths, defaults, scope cuts and edge cases,
never verification levels.

**Change.** Extend check 4 with a third clause:

> Every row whose `Verified by` level is `manual` names an acceptance
> criterion that carries the `(manual)` marker in its FRD, per
> [[ac-traceability-standards]]. A `manual` level without the marker is
> a content gap: the criterion is exempt from citation only once the FRD
> records the designation.

Add the corresponding line to the decision sweep's list of typical
silent decisions: "whether a criterion is verifiable only manually — and
if so, whether the FRD records the `(manual)` marker that exempts it".

**Evidence.** Finding 2. The next run bounces the three unmarked
criteria into Open Questions at Phase 1, where the FRD's author resolves
them, instead of grooming carrying an unsatisfiable obligation onto a
work order and hoping its implementer notices.

## Considered and Not Proposed

- **Require every delta block to appear in the AC Coverage table.** The
  run surfaced a delta (SP) named in no row, whose verification basis
  the agent had to infer from a different document. But a foundational
  component proven through its consumers is legitimate design, and a
  synthetic AC row would weaken the table. The fix belongs in the
  blueprint that omitted its verification basis, and is filed in the
  project-audience RFC `rfcs/inbox/feature-blueprint/rfc-0010`.
- **Require the Tracker Reconciliation section to be complete with
  respect to the live tracker.** The run found the section naming an
  issue absent from the tracker and omitting an open one relevant to a
  delta. Not proposed here: check 10 deliberately validates only listed
  issues, and requiring completeness would make the checklist depend on
  a live tracker query it cannot perform. Filed against the blueprint
  instead (`rfc-0010` P4).
- **Extend check 10's classification vocabulary.** A real finding from
  this run, but the missing piece is the *failure response* in
  `groom-feature-blueprint`'s Phase 1, not the vocabulary in this
  checklist. Filed against that skill.
- **Have the checklist mandate a specific validation tool.** Rejected:
  the checklist is framework-level and must hold for projects with no
  schema tooling. P1's WHERE clause keeps it tool-agnostic.

## Non-Goals

- No change to `ac-traceability-standards.md`. P2 makes the readiness
  gate enforce a rule that standard already states; the standard itself
  is correct as written.
- No new frontmatter properties, and no change to the vocabularies
  `feature-blueprint-template.md` documents.
- Not making the checklist depend on any project-specific tooling,
  schema format, or tracker.

## Adoption

P1 and P2 are independent. Both touch
`plugin/skills/spec/references/blueprint-readiness-checklist.md` — P1
adds a check to the content-checks list (lines 53-85) and a note to the
decision sweep; P2 extends check 4 and adds one line to the same sweep.
They collide textually in the decision-sweep section, so applying them
together in one edit is simpler than sequencing them, but neither
depends on the other's content.

Both are additive to the checklist, which means every blueprint
currently marked `ready-for-review` or `approved` becomes subject to two
checks it was never run against. That is the intended effect — this
run's blueprint would have failed both — but it should be stated at
adoption rather than discovered: expect existing approved blueprints to
fail on re-grooming until their authors reconcile them. The checklist's
version should bump from 0.5.0, and both consumers (`spec`'s authoring
preflight and `groom-feature-blueprint`'s Phase 1) inherit the change
with no edit of their own, since both reference the document rather than
restating it.

Expected effect: the two defect classes that reached decomposition on
this run — an invalid frontmatter value and an unsatisfiable citation
obligation — are caught at authoring time, which is the phase that can
actually fix them.

## Resolution
