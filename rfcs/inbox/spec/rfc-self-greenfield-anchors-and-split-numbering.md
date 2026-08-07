---
title: "RFC: Greenfield anchors and split-FRD numbering — two silent gaps
  in the spec authoring templates"
description: The feature-blueprint and FRD templates give no rule for a
  first-implementation (greenfield) feature's code anchors, nor for
  identifier continuity when requirements are split out of an FRD
slug: rfc-self-greenfield-anchors-and-split-numbering
type: rfc
status: ready-for-review
updated_at: 2026-08-04T00:00:00Z
origin: self
skill: spec
skill_version: "1.0"
model: claude-opus-4-8[1m]
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "n/a (standalone install)"
---

# RFC: Greenfield anchors and split-FRD numbering gaps in the spec templates

## Summary

A single session ran the `spec` skill to re-scope
`FRD-spec-validation-cli`, split it into three FRDs, and author the
first — greenfield — feature blueprint for the Spec CLI. Two of the
skill's templates went silent where the run forced a decision: the
feature-blueprint template's **code-anchors** rule assumes existing code
to point at, but every delta was a first implementation with no code
yet; and the FRD template's **ID conventions** cover reuse but not the
numbering gaps that a split leaves behind. Both gaps were navigated by
improvising a convention. Three sharper frictions from the run were
confirmed already fixed in the current version and are recorded, not
proposed.

## Provenance

The run exercised `spec` **1.0** as installed standalone at
`~/.claude/skills/spec`. This repo's authoritative copy is
`plugin/skills/spec` at **1.6** (plugin 0.14.0). Both proposed findings
were verified to persist verbatim in 1.6, so citations below use the
repo-relative 1.6 paths triage will act on, noting the identical 1.0
line the run saw. The three frictions under Considered and Not Proposed
were confirmed fixed between 1.0 and 1.6.

| Field | Value |
|-------|-------|
| Harness | Claude Code (CLI) |
| Task the skill was applied to | Re-scope `FRD-spec-validation-cli`, split it into three FRDs, and author the greenfield feature blueprint `feature-spec-validation-cli` |
| References / templates invoked | `frd-template.md`, `feature-blueprint-template.md`, `blueprint-readiness-checklist.md`, `feature-blueprint` §Code anchors — all as shipped in `spec` 1.0; persistence checked against 1.6 |
| Run dates | 2026-08-04 |

## Evidence

1. **[gap · snag]**
   `plugin/skills/spec/references/feature-blueprint-template.md:133`
   "Delta blocks SHOULD name real, existing code anchors (packages, key
   symbols, migration directories)" and `:135` "They SHALL NOT prescribe
   new function signatures, new file layouts, or other structure that
   does not exist yet." — The feature was a first implementation: the
   owning blueprints ([[container-spec-cli]], [[module-document-core]],
   [[module-spec-profile]]) declare components (Spans, Corpus, Command
   Surface, …) but the packages they name (`cli/core/*`, `cli/profile`,
   `cli/cmd`) are unwritten. There is no existing code to anchor to, and
   naming the planned package the blueprint implies is arguably the very
   "structure that does not exist yet" the rule forbids. No rule covers
   the greenfield case, so all twelve delta blocks improvised the same
   ad-hoc anchor form — e.g. "new package under `cli/core` (e.g.
   `cli/core/span`)". Identical text at the run version:
   `feature-blueprint-template.md:131,133` in 1.0.

2. **[gap · polish]**
   `plugin/skills/spec/references/frd-template.md:63` "Never reuse a
   retired ID." — The rule addresses reuse but is silent on the
   numbering **gap** a split leaves. Splitting `FRD-spec-validation-cli`
   moved `REQ-SPEC-004`/`005` into two new FRDs while `REQ-SPEC-006`,
   `-008`, `-009` kept their numbers (ADRs cite them). Nothing said
   whether to renumber the survivors or leave the vacated numbers unused,
   so the run invented a "*Note on numbering:*" prose block to explain
   the deliberate gap so a reader would not read it as an omission.
   Identical rule at the run version: `frd-template.md:57` in 1.0; 1.6
   added a "Splitting and merging features" section at `:54` covering
   *when* to split, not identifier continuity.

## Proposals

### P1 — Add a greenfield rule to the "Code anchors" guidance [snag]

**Problem.** `feature-blueprint-template.md:133–135` tells the author to
"name real, existing code anchors" and "SHALL NOT prescribe … structure
that does not exist yet." A first-implementation feature has neither: no
code exists, and the planned package is the only useful starting point a
planner can be given. The result (Finding 1) is that every greenfield
anchor is improvised, inconsistently, with no blessing from the
template.

**Change.** Append a paragraph to the "Code anchors — anchors, not
signatures" block:

> WHERE the feature is a first implementation and the owning blueprint
> declares a component with no code yet, the anchor SHALL name the
> package or path that blueprint already implies (its `source_paths`,
> or the package its components map to), marked as not-yet-existing —
> e.g. "new package `cli/core/span`". Naming a location the blueprint
> already fixes is not prescribing structure; inventing a layout the
> blueprint does not imply still is.

**Evidence.** Finding 1. Under the same model, greenfield deltas would
carry a consistent, sanctioned anchor form instead of twelve ad-hoc
"new package under X (e.g. …)" phrasings.

### P2 — Add an identifier-continuity rule for split FRDs [polish]

**Problem.** `frd-template.md:63` forbids reusing a retired ID but says
nothing about the gaps a split produces (Finding 2), so the author must
decide renumber-vs-gap and invent a way to signal it.

**Change.** Add a bullet under "ID conventions":

> WHEN requirements are moved out of an FRD into a new FRD, the retained
> requirements SHALL keep their identifiers so external citations do not
> break, and the vacated numbers SHALL be left unused rather than
> reassigned. Note the deliberate gap in the document (a one-line note)
> so a reader does not read it as an omission.

**Evidence.** Finding 2. The next FRD split follows a stated rule and
produces a predictable note rather than an invented convention.

## Considered and Not Proposed

- **`@slug` cross-reference notation** (run version 1.0:
  `frd-template.md:45,201`, `SKILL.md:177,240`) — already retired in 1.6,
  whose `SKILL.md:213` now states "`[[slug]]` is reserved for
  documents." Following 1.0 literally would have diverged from the
  corpus (and from the `[[slug]]` parser this very feature builds), but
  the fix has landed. No action; the 1.0→1.6 change is evidence the
  retro→triage loop closed here.
- **`@guide-c4` in the feature-blueprint References boilerplate** (run
  version 1.0: `feature-blueprint-template.md:457`) — removed in 1.6
  (now `c4-guide.md:474` as a backticked path). In 1.0 it named a slug
  no corpus document declares, so following it would have produced a
  dangling reference that fails the template's own readiness check #5;
  fixed. No action.
- **Multi-token component `#`-reference requiring an explicit PascalCase
  id** (run version 1.0 reference-syntax:
  `feature-blueprint-template.md:104–113`) — the run hit this against
  existing components whose declarations carry no reference id (Schema
  Registry, Command Surface, Document Types), forcing prose workarounds.
  1.6 replaced the whole model (`:118–130`) with
  `[[container-<name>#Component Name]]` wikilink anchors and the rule
  "Delta-block headings NAME the component they change (it may be new,
  so it has no anchor yet)", which resolves both the id-less-existing
  component case and multi-word names. No action.

## Non-Goals

- Changing the "anchors, not signatures" principle: P1 narrows one
  greenfield case; it does not license prescribing signatures or file
  layouts a blueprint does not already imply.
- The `[[slug]]` reference model: already reworked in 1.6.

## Adoption

P1 and P2 are independent and touch different files
(`feature-blueprint-template.md`; `frd-template.md`), so they carry no
ordering. Both are additive — they supply a rule for a case the
templates currently leave unspecified — so brownfield deltas and
non-split FRDs are unaffected, and existing project copies see no
behavior change until re-authored. Both were verified against 1.6, the
version triage will edit. Expected effect: greenfield feature blueprints
and FRD splits stop requiring per-author improvisation on anchors and
numbering.

## Resolution
