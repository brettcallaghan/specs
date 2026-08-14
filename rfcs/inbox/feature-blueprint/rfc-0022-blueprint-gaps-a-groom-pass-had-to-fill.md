---
title: "RFC: A blueprint that groomed cleanly still left three things for
  the groomer to invent — a component name, an enabling delta's done-ness,
  and an ordering the graph cannot express"
description: Declare the slug of any component a delta adds, mark a delta
  that proves no acceptance criterion as an enabling delta, and state
  sequencing constraints as orderings a dependency edge can carry
slug: rfc-0022-blueprint-gaps-a-groom-pass-had-to-fill
type: rfc
status: ready-for-review
updated_at: 2026-08-14T00:00:00Z
origin: self
skill: groom-feature-blueprint
skill_version: "1.3"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "0.17.0"
---

# RFC: Blueprint gaps a grooming pass had to fill

## Summary

[[feature-gherkin-adoption]] was groomed into thirteen work orders under
epic `agent-plugins-wvd`. It passed the readiness gate on all three
checks and the content checks on all seven — and grooming still had to
invent a C4 component name the vocabulary does not hold, decide what
done-ness means for a delta that proves no acceptance criterion, and
choose how to encode an ordering phrased so that no dependency edge can
carry it. Each of the three is a small, specific addition to the
blueprint that would have removed the judgement call.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins |
| Audience | project — resolved in this repository, which is what filing under `rfcs/inbox/feature-blueprint/` records; the generated `Rfc` schema has no `audience` property |
| Phase executed | Groomed an approved feature blueprint into a beads plan |
| Documents exercised | [[feature-gherkin-adoption]]; [[frd-gherkin-adoption]]; [[container-spec-cli]], [[module-document-core]], [[module-spec-profile]] (for the C4 vocabulary) |
| Documents produced | Epic `agent-plugins-wvd` and thirteen work orders (tracker records, not specification documents) |
| Skill(s) followed | groom-feature-blueprint v1.3; groom-beads v1.3 |
| Templates / standards read | `blueprint-readiness-checklist.md` v0.6.0; feature-blueprint-template v1.4; `work-order-definition-of-ready.md` v1.4; `work-order-field-guide.md` v1.2; `ac-traceability-standards.md` v0.2.0 |
| Plugin version | 0.17.0 |
| Session date | 2026-08-14 |

## Evidence

1. **Tagging C4 blast radius required minting a component name the
   corpus does not define.** The GHK delta's Anchors read "a new
   `cli/core/gherkin` package", and its Definition of Done entry says
   [[module-document-core]]'s "Components gains **Gherkin**". Neither
   states the component's slug. The corpus defines exactly twelve
   component slugs, in `component` record blocks across the three
   architecture blueprints — `component-corpus`, `component-references`,
   `component-markdown-structure` and nine others — and no
   `component-gherkin` among them. `groom-beads` requires a controlled
   vocabulary for `c4_changes` ("read the actual C4 docs and only use
   component names that exist"), so grooming had to write
   `component-gherkin` on the strength of a component name appearing in
   prose, and flag the mismatch it had itself created.

2. **One delta proves no acceptance criterion, and nothing said so.**
   The AC Coverage table names XPT as a participant in AC-GHK-002.1,
   AC-GHK-002.8 and AC-GHK-003.6, and as the proving delta for none —
   PREG and VTAG prove all three. That is correct: XPT widens an
   interface and the criteria are about what the widened interface
   carries. But the work order derived from it reaches the Definition of
   Ready with an empty acceptance list, which reads as a grooming
   defect rather than a property of the delta. Grooming had to author
   the sentence the blueprint could have carried: this item proves no
   criterion of its own, and its verification lands with PREG and VTAG.

3. **A sequencing constraint was phrased so that no dependency edge can
   express it.** Sequencing Constraints states: "The VTAG and VCON
   deltas SHALL land before or with QRY, which declares their rule
   identifiers to the run report." A `blocks` edge means strictly
   before; there is no edge meaning "before or concurrently". Grooming
   wired `QRY` blocked-by `VTAG` and `VCON`, which is safe and which
   serializes work the constraint explicitly permits to be concurrent.
   The phrasing recurs: [[feature-spec-reference-integrity]] carries
   "The CHK and LIFE deltas SHALL land before or with CMD", in the same
   position, and the feature-blueprint template does not use the phrase
   anywhere — so it is this corpus's habit rather than an inherited one.

4. *(Context, not a proposal.)* The work orders reached the finishing
   pass with no `## Design` field, which `work-order-definition-of-ready.md`
   check 3 requires, and all thirteen had to be authored during the
   groom-beads pass. The blueprint's delta blocks carry **Intent**,
   **Behavior** and **Anchors** and deliberately no implementation
   approach — the template's own "anchors, not signatures" rule. The gap
   is between the two documents rather than in either, and it is routed
   to `retro-skill` rather than proposed here.

## Proposals

### P1 — A delta that adds a component declares its slug

**Problem.** The GHK delta names a package path and a component name in
prose and no slug. `c4-guide.md` fixes component slugs as
`component-<name>` and the corpus carries them in `component` record
blocks, but a component that does not exist yet has no record block to
read, so the vocabulary a groomer validates against is missing exactly
the entry the blueprint is adding.

**Change.** Add to the feature-blueprint conventions: WHERE a delta
adds a C4 element, the delta SHALL state that element's slug alongside
its anchors, in the form the owning blueprint will carry once the fold
lands. For GHK that is one line:

> **Adds component:** `component-gherkin` — declared in
> [[module-document-core]] by this delta's fold.

**Evidence.** Finding 1. Grooming tags `c4_changes` from the blueprint
rather than from a guess, and the name in the plan matches the name the
fold will create rather than being reconciled later.

### P2 — Mark a delta that proves no acceptance criterion

**Problem.** Nothing in the blueprint distinguishes a delta that proves
criteria from one that only enables others to. The AC Coverage table
makes the distinction visible only by absence — XPT never appears in
bold — and absence is exactly what a reader does not notice.

**Change.** Add a line to any such delta's Intent, and a note under the
AC Coverage table listing them:

> This delta proves no acceptance criterion. It is an enabling delta;
> its verification lands with <the deltas that prove the criteria it
> participates in>.

**Evidence.** Finding 2. The derived work order carries the explanation
rather than an unexplained empty acceptance list, and a reviewer reading
the bead graph can tell an enabling delta from a mis-groomed one.

### P3 — State sequencing constraints as orderings an edge can carry

**Problem.** "SHALL land before or with QRY" has no representation in a
dependency graph, which is where the orchestrator reads ordering from.
Every groomer meeting the phrase makes the same choice, and the safe
choice — a strict edge — discards the concurrency the constraint was
written to permit.

**Change.** Adopt two forms and use nothing else. Either a strict
ordering, which becomes an edge:

> The VTAG delta SHALL land before QRY.

or an explicit statement that the two may be concurrent and what
actually couples them:

> QRY declares the rule identifiers VTAG and VCON introduce. The three
> MAY land concurrently; a wave containing QRY SHALL contain VTAG and
> VCON or follow them.

**Evidence.** Finding 3, in two blueprints. Grooming encodes the
constraint rather than interpreting it, and a constraint that permits
concurrency stops being silently serialized.

## Considered and Not Proposed

- **AC-GHK-001.6 arriving in a work order as an unverifiable
  acceptance line.** DISC's acceptance lists it at level `manual`, so
  an implementing agent sees a criterion it cannot write a test for.
  This is the same defect
  `rfcs/inbox/frd/rfc-0021-gherkin-frd-unverifiable-and-implied-work.md`
  P3 already proposes fixing, reaching the same conclusion from the
  grooming side rather than the authoring side. Corroboration, not a
  second proposal — re-proposing it would split the decision across two
  RFCs.
- **The missing `## Design` field on derived work orders.** Finding 4.
  Root-caused to the guidance rather than to this project's documents:
  `groom-feature-blueprint` Phase 2 enumerates what a work order carries
  and does not mention Design, while the Definition of Ready requires
  it. Routed to `retro-skill`.
- **The RFC number collision this retro hit.** `rfc-0019` was minted
  independently on two branches — this project's Gherkin RFC and main's
  markdown-links RFC — and the RFC queue's own check deliberately does
  not request `slug/duplicate`, so nothing caught it. The Gherkin RFC
  was renumbered to `rfc-0021`. Not proposed here because it is the
  third instance of one pattern this session (ADR-018 and ADR-020
  collided the same way) and the fix belongs with `spec next-id` and
  branch-parallel numbering rather than with any one document — it wants
  its own investigation, not a proposal buried in a blueprint RFC.

## Non-Goals

- No change to the readiness checklist. All ten checks passed and every
  finding here is something the checklist does not ask about — the fix
  is more information in the blueprint, not a stricter gate.
- No change to the AC Coverage row-validity rules. P2 adds a note
  alongside the table, not a new column or a new row form.
- No re-groom of `agent-plugins-wvd`. All three findings were resolved
  in the plan as it was built; adopting these proposals prevents the
  next blueprint from needing the same judgement calls.

## Adoption

P1, P2 and P3 are independent and touch one document each — except that
adopting them fully means editing both existing feature blueprints, so
that the convention is visible in the examples the next author reads.
[[feature-gherkin-adoption]] is `in-progress` and
[[feature-spec-reference-integrity]] is `retired`; a retired blueprint
is a record and SHOULD NOT be edited, so P3's second instance is
evidence rather than a target.

All three are backward compatible: each adds text and none changes an
identifier, a delta key, or a coverage row, so the plan already derived
from this blueprint stays valid and no work order needs re-deriving.

The expected effect is that the next blueprint groomed in this project
produces a bead graph with no invented C4 names, no unexplained empty
acceptance lists, and no ordering the groomer had to interpret.

## Resolution
