---
title: "RFC: The corpus validator checks no specification document, and
  four reconciliation gaps it would not have caught"
description: Repoint the corpus validator at specs/ and fail a vacuous
  scan; teach the Rfc model to represent a project-audience RFC; close
  the manual-AC, tracker-reconciliation and unproven-delta gaps grooming
  surfaced in the spec-index blueprint and its FRD
slug: rfc-0010
type: rfc
status: ready-for-review
updated_at: 2026-08-04T00:00:00Z
origin: self
skill: groom-feature-blueprint
skill_version: "1.2"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.14.0"
related:
  - "feature-spec-validation-cli"
  - "frd-spec-validation-cli"
---

# RFC: The corpus validator checks no specification document

## Summary

Grooming the approved spec-index feature blueprint into 16 work orders
bounced at the readiness gate on a frontmatter defect — `delivery:
approved`, a value outside the `Delivery` enum — that CI should have
caught long before. Investigating why produced the headline finding:
`scripts/validate-corpus.py` still searches `docs/`, which the `specs/`
reorg emptied, so **zero of the repository's 27 specification documents
are validated**, while the script reports "checked 21 documents, 0
findings" and exits 0. The remaining proposals close four gaps the same
phase surfaced that no schema would catch: an `Rfc` model that cannot
represent the very document this retro produces, an FRD missing the
`(manual)` markers its own standards require, a Tracker Reconciliation
section naming a nonexistent issue while omitting a relevant one, and a
delta whose verification basis grooming had to infer.

**This is a project-audience RFC** in the sense
`retro-spec`'s `references/rfc-template.md` v1.1 defines — its proposals
change this project's own documents and tooling, not the framework's
skills. The `Rfc` model cannot record that fact; see finding 7 and P2.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins (self) |
| Phase executed | Groomed the spec-index feature blueprint into beads (epic `agent-plugins-bge`, 16 work orders) |
| Documents exercised | `specs/features/spec-index/spec-index-blueprint.md`, `specs/features/spec-index/spec-index-frd.md`, `specs/architecture/{container-spec-cli,module-document-core,module-spec-profile}.md`, `specs/standards/testing-standards.md` |
| Documents produced | Epic `agent-plugins-bge` + 16 work-order beads; `rfcs/findings-capture.md` |
| Skill(s) followed | `groom-feature-blueprint` v1.2, `groom-beads` v1.1, `retro-spec` v1.3.0 |
| Templates / standards read | `blueprint-readiness-checklist.md` v0.5.0, `feature-blueprint-template.md` v1.3, `work-order-template.md` v1.2, `ac-traceability-standards.md` v0.2.0, `work-order-definition-of-ready.md` v1.1, `findings-capture-protocol.md` v0.2, `rfc-template.md` v1.1 |
| Plugin version | 0.14.0 |
| Session date | 2026-08-04 |

The `skill` and `skill_version` properties record the skill whose run
produced the evidence, because the `Rfc` model requires them; this RFC
does **not** critique `groom-feature-blueprint` — a separate skill-lens
retro covers that. The `model` property is withheld: this session runs
under a harness policy forbidding the model identifier in repository
artifacts, and the placeholder follows the convention already
established by `rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md`.

## Evidence

1. **The corpus validator matches no specification document.**
   `scripts/validate-corpus.py` line 45 declares
   `SEARCH = ["docs/**/*.md", "rfcs/**/*.md",
   "plugin/skills/*/references/*.md"]`, under a comment on lines 42-44
   reading "The corpus proper is docs/". Enumerating those globs in this
   working tree: `docs/**/*.md` matches **1** file
   (`docs/session-bootstrap.md`); `specs/**/*.md` matches **27** files,
   of which **0** are matched by any `SEARCH` pattern. The script's
   summary reports `checked 21 documents, 0 findings, 6 excluded` — all
   21 are RFCs and plugin skill references.

2. **A frontmatter value outside its enum reached an approved
   blueprint.** The blueprint carried `delivery: approved`, while
   `schemas/vocabulary.strata` line 130 declares `enum Delivery` with
   members `"single-pr"` and `"staged"`, and the generated
   `cli/core/schema/spec-model.schema.json` correctly renders it as
   `{"enum": ["single-pr", "staged"], "type": "string"}`. The schema is
   right; nothing applied it. Re-injecting `delivery: approved` and
   re-running the validator reproduces `checked 21 documents, 0
   findings` and exit status 0. The defect was caught for the first time
   by `groom-feature-blueprint`'s Phase 1 readiness gate — one phase
   before implementation.

3. **The blueprint's Definition of Done carried a stale corpus path.**
   It read "The profile self-application gate runs the built profile
   against this repository's `docs/` corpus" — the same stale `docs/`
   reference as finding 1, in a second artifact. Because `docs/` now
   holds one non-specification file, a gate wired to it would have
   passed vacuously. Corrected in this session on author confirmation;
   recorded because it is the second instance of one root cause.

4. **The FRD does not carry the `(manual)` markers its own standards
   require.** The blueprint's AC Coverage table declares AC-SPEC-009.4,
   AC-SPEC-009.5 and AC-SPEC-009.6 verified at level `manual`, but
   `specs/features/spec-index/spec-index-frd.md` lines 383-393 state
   those criteria with no marker appended.
   `specs/standards/testing-standards.md` lines 266-268 make the marker
   the sole exemption mechanism: "IF an acceptance criterion carries the
   `(manual)` marker defined by
   `plugin/skills/spec/references/frd-template.md`, THEN it is exempt
   from the citation obligation above." As written, the three criteria
   are simultaneously unverifiable by any test and ineligible for
   exemption. Grooming could not resolve this without making an
   authoring decision, so it was carried onto the BUILD work order as an
   open obligation.

5. **Tracker Reconciliation is out of step with the tracker.** The
   section classifies `agent-plugins-vg1` as out-of-scope, but that
   issue was absent from the tracker at groom time — it existed only in
   `.beads/issues.jsonl` and had never been imported into the local
   store. Separately, `agent-plugins-55q` ("ADR: release and
   distribution of CLI binaries") is open and directly governs the BUILD
   delta's distribution-mechanism choice, and the section does not list
   it. Readiness check 10 passed regardless, because it requires only
   that *listed* issues be classified.

6. **One delta's verification basis had to be inferred.** The SP
   (#Spans) delta appears in **no** row of the AC Coverage table, and FM
   (#Frontmatter) appears only as a participating delta, never a proving
   one. Nothing in the blueprint states how SP is verified, so grooming
   derived it from `specs/architecture/module-document-core.md`'s Change
   Guidance ("A change to Spans additionally runs the property tests
   over the delimiter scanner") and wrote that reasoning into the work
   order by hand.

7. **The `Rfc` model cannot represent this RFC.** Filing this document
   failed the project's own gate with seven schema errors, including
   `Additional properties are not allowed ('audience' was unexpected)`
   and `'skill' is a required property`. `schemas/frontmatter.strata`
   lines 372-417 define `struct Rfc` with required `skill`,
   `skill_version`, `model`, `effort`, `runs_observed` and `severity` —
   every one a skill-critique concept — and no `audience`. But
   `retro-spec`'s `references/rfc-template.md` v1.1 requires
   `audience: project | skill-authors` and describes exactly this
   document class: "proposals change the origin project's own documents
   (its specs, blueprints, standards)". All nine retired RFCs
   (0001-0009) critique skills, so the gap had never been exercised
   until now. The frontmatter above is the contortion required to pass.

## Proposals

### P1 — Repoint the validator at `specs/`, and fail a vacuous scan

**Problem.** `scripts/validate-corpus.py` line 45 reads:

```python
SEARCH = ["docs/**/*.md", "rfcs/**/*.md",
          "plugin/skills/*/references/*.md"]
```

with the comment above it (lines 42-44) asserting "The corpus proper is
`docs/`; the RFC queue and the spec skill's own reference documents are
checked too". That was true before the `specs/` reorg and is false now.
The failure is silent in the worst way: the script still finds 21
documents and reports zero findings, so CI and any human reading the
output receive positive confirmation that the corpus is clean while no
specification document has been examined at all (finding 1). Every
frontmatter defect under `specs/` is invisible, which is how finding 2
reached an approved document.

**Change.** Two edits to `scripts/validate-corpus.py`:

1. Replace `SEARCH` with `["specs/**/*.md", "rfcs/**/*.md",
   "plugin/skills/*/references/*.md"]`, and update the comment on lines
   42-44 to name `specs/` as the corpus proper.
2. Add a guard after enumeration, before validation: IF no file matched
   by any pattern declares a `type` in the profile vocabulary, THEN the
   script SHALL report that its search patterns matched no specification
   document, naming the patterns, and exit non-zero. A vacuous pass
   SHALL NOT be reported as success.

The second edit makes the first durable: a future move of `specs/`
breaks the build loudly instead of degrading to a silent no-op.

**Evidence.** Findings 1, 2, 3. On the next run `delivery: approved`
fails CI at authoring time instead of surviving to the grooming gate,
and 27 previously-unchecked documents come under the schema. Expect the
first run after adoption to surface real findings — `agent-plugins-6n4`
(`work-order-template.md` frontmatter is not valid YAML) and
`agent-plugins-5nz` (26 dangling wikilinks) predict some of them.

### P2 — Teach the `Rfc` model to represent a project-audience RFC

**Problem.** `schemas/frontmatter.strata` lines 372-417 model `Rfc`
solely as a skill critique: `skill`, `skill_version`, `model`, `effort`,
`runs_observed` and `severity` are all required, and `audience` — which
`rfc-template.md` v1.1 makes the field that decides who resolves the
RFC — is not permitted at all. A retro of project documents has no
skill to name, no model-scoped reproduction claim, and no skill version
for triage to compare against. This RFC had to name the phase's skill in
`skill` and explain in prose that it is not critiquing it (finding 7).

**Change.** In `schemas/frontmatter.strata`, add an `audience` property
to `struct Rfc` over a new enum (`project` | `skill-authors`), and make
`skill`, `skill_version`, `model`, `effort` and `runs_observed`
optional, since none is meaningful for `audience: project`. Keep
`origin`, `severity` and `plugin_version` required — all three are
meaningful for both audiences. Then regenerate via
`scripts/generate-schemas.sh` and commit the generated artifacts.

WHERE the schema cannot express "these five are required when audience
is skill-authors", state that obligation in the RFC template's readiness
gate rather than weakening the model further; a conditional-requirement
mechanism is not worth introducing for one document type.

**Evidence.** Finding 7. The next `retro-spec` run files a
project-audience RFC that passes the gate as written, instead of
contorting its frontmatter and explaining the contortion in Provenance.
Backward compatible: RFCs 0001-0009 already populate every field this
proposal relaxes.

### P3 — Mark the three manual acceptance criteria in the FRD

**Problem.** `specs/standards/testing-standards.md` lines 266-268 grant
citation exemption only to a criterion carrying the `(manual)` marker,
and `specs/features/spec-index/spec-index-frd.md` lines 383-393 state
AC-SPEC-009.4, AC-SPEC-009.5 and AC-SPEC-009.6 without it, while the
blueprint's AC Coverage table declares all three `manual` (finding 4).
Two approved documents disagree, and only the FRD's author can resolve
it: grooming may not append a marker to an approved FRD, and no test can
produce the citation the standard demands.

**Change.** Append the `(manual)` marker to the criterion text of
AC-SPEC-009.4, AC-SPEC-009.5 and AC-SPEC-009.6 in
`specs/features/spec-index/spec-index-frd.md`, in the format
`frd-template.md` defines. This records the verification level the
blueprint already approved; it does not change what the criteria
require. IF the intent was instead that these be automatically
verifiable, THEN the blueprint's AC Coverage table is what changes and
the three rows move off `manual`.

**Evidence.** Finding 4. Work order `agent-plugins-bge.13` (BUILD)
currently carries this contradiction as an obligation it must resolve
before closing; adopting P3 discharges it upstream, where the decision
belongs.

### P4 — Reconcile Tracker Reconciliation, and record its derivation

**Problem.** The section classifies `agent-plugins-vg1`, absent from the
tracker at groom time, and omits `agent-plugins-55q`, which is open and
governs a decision the BUILD delta must make (finding 5). Nothing in the
section states when or how it was derived, so a groomer cannot
distinguish a deliberate omission from a stale one — and readiness check
10 passes either way, because it validates only the issues that are
listed. Grooming wired both anomalies into the final Definition-of-Done
work order as notes rather than resolving them silently.

**Change.** In `specs/features/spec-index/spec-index-blueprint.md`,
classify `agent-plugins-55q` (a prerequisite decision for BUILD, or an
explicit out-of-scope deferral), and either remove `agent-plugins-vg1`
or note that it is JSONL-only. Add one line at the head of the section
recording its derivation — e.g. "Reconciled against `bd list --status
open` on 2026-08-04." Adopting that derivation line as a convention for
future feature blueprints in this repository is the durable half; the
two corrections are the immediate half.

**Evidence.** Finding 5. A groomer reading a dated derivation line can
tell a stale section from a complete one and re-run the query, instead
of discovering mid-decomposition that a classified issue does not exist.

### P5 — State the verification basis for a delta with no AC row

**Problem.** The SP delta is named in no AC Coverage row, and the
blueprint says nothing about how it is verified (finding 6). That is a
legitimate design — a foundational component proven through its
consumers — but the reasoning lived only in the author's head. Grooming
reconstructed it from a different document
(`specs/architecture/module-document-core.md`'s Change Guidance) and
wrote it into the work order, which is the guesswork the readiness gate
exists to prevent. The gate does not catch it: check 4 requires every AC
to appear in the table, not every delta.

**Change.** In `specs/features/spec-index/spec-index-blueprint.md`, add
one sentence to the SP delta stating that it carries no acceptance
criteria of its own and is verified through its consumers plus the
property tests over the delimiter scanner that
`specs/architecture/module-document-core.md` requires. Adopt the general
rule for feature blueprints in this repository: WHERE a delta block is
named in no AC Coverage row, the delta SHALL state its verification
basis.

**Evidence.** Finding 6. The SP work order's "This delta carries no
acceptance criteria of its own" paragraph — written by grooming, not by
the blueprint — becomes a quotation from the spec rather than an
inference, and the next foundational delta gets the same treatment
without a groomer having to notice the absence.

## Considered and Not Proposed

- **Treat the transposed `status`/`delivery` values as a guidance
  defect.** The feature blueprint template documents both fields
  correctly (`feature-blueprint-template.md` line 14:
  `delivery: single-pr # single-pr | staged`), so this was authoring
  carelessness, not misleading guidance. The fix is automation (P1), not
  more prose — a warning nobody read is not improved by rewording.
- **Add a frontmatter-enum check to the blueprint readiness
  checklist.** That checklist is a framework document owned by the
  `spec` skill, not a project document; routed to the `retro-skill` lens
  rather than filed here.
- **Require every delta block to appear in the AC Coverage table.**
  Too strict: SP is legitimately proven through its consumers, and a
  synthetic AC row would weaken the table's meaning. P5 is the lighter
  fix that preserves the signal.
- **Promote the emitted index to a contract document.** Already a
  recorded scope decision deferred to `agent-plugins-h6c` with a stated
  rationale; not a retro finding.
- **Add a merge driver for `.beads/issues.jsonl`.** Real, and it bit
  this session adjacently, but already tracked as `agent-plugins-oo2`
  and not evidence from the documents this phase consumed.
- **Reconcile the `bd` export/import desync that hid
  `agent-plugins-vg1`.** A tracker-tooling defect, not a document
  defect; it belongs in the tracker, and finding 5 uses it only as
  evidence about the blueprint's section.

## Non-Goals

- No new document types, and no change to the feature blueprint or FRD
  templates — every proposal changes this project's own documents,
  schemas, or tooling.
- Not relaxing any consumer check. The readiness gate performed
  correctly: it caught finding 2 and refused to plan around it. These
  proposals move detection earlier; they do not make the gate lenient.
- Not changing what any acceptance criterion requires. P3 records a
  verification level already approved.
- Not introducing a conditional-requirement mechanism to the schema
  language for P2's optional fields.

## Adoption

P1 is independent and should land first — it is the only proposal that
prevents recurrence rather than correcting an instance, and its second
edit is what stops the same class of defect after the next directory
move. Expect its first CI run to go red on pre-existing corpus defects;
triage those as findings, not as a reason to revert. Touch point:
`scripts/validate-corpus.py`.

P2 is independent of P1 but should land before the next `retro-spec`
run, or that run repeats this RFC's contortion. It is the only proposal
requiring regeneration: touch points are `schemas/frontmatter.strata`,
`schemas/vocabulary.strata`, and the generated `cli/core/schema/`,
`cli/core/specmodel/` artifacts via `scripts/generate-schemas.sh`. Note
the interaction with P1: once P1 lands, `specs/` documents are validated
too, so P2's regeneration must pass the widened gate.

P3, P4 and P5 are independent of everything and of each other. Touch
points: `specs/features/spec-index/spec-index-frd.md` (P3) and
`specs/features/spec-index/spec-index-blueprint.md` (P4, P5). P3 should
land before work order `agent-plugins-bge.13` (BUILD) is dispatched,
since that item carries the contradiction as an open obligation; P4 and
P5 should land before `agent-plugins-bge.16` runs the
Definition-of-Done audit.

Backward compatibility: P1 changes no document and can only widen what
CI examines. P2 relaxes required properties and adds one — every
existing RFC stays valid, as RFCs 0001-0009 already populate all of
them. P3 appends a marker the citation standard already defines, leaving
every existing citation valid. P4 and P5 add text to one blueprint and
introduce two conventions for future blueprints; no existing document
becomes invalid, and neither convention is retroactively enforced.

Overall effect: the repository stops reporting a clean corpus it never
checked, the document type this retro produces becomes representable in
the corpus it is filed into, and the three defects that survived to
grooming — none of which a schema would catch — are closed at their
source rather than carried as notes on work orders.

## Resolution
