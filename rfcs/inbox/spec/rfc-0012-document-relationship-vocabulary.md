---
title: "RFC: Reconcile the document-relationship vocabulary with the
  derived reference graph"
description: Adopt a standard for declaring cross-document relationships
  once and deriving the reverse, and apply it to the framework's
  frontmatter — rename the feature blueprint's FRD edge, drop `updates`
  and `creates`, trim `related`, mark `depends_on` aspirational, and
  derive `children` from `parent_slug`
slug: rfc-0012
type: rfc
status: draft
updated_at: 2026-08-06T00:00:00Z
origin: self
skill: spec
skill_version: "1.7"
model: "claude-opus-4-8"
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "0.14.0"
related:
  - "document-relationship-standards"
  - "blueprint-readiness-checklist"
  - "frd-spec-reference-integrity"
  - "frd-spec-validation-cli"
  - "adr-009"
  - "adr-012"
---

# RFC: Reconcile the relationship vocabulary with the derived graph

## Summary

The reference-integrity feature makes the corpus reference graph
bidirectional in the index: it parses every `[[slug]]` from a document's
body, derives every document's inbound edges from the outbound edges of
all others, and answers both directions from a query
([[frd-spec-reference-integrity]], AC-SPECREF-001.1, AC-SPECREF-001.3,
AC-SPECREF-003). The framework's frontmatter relationship vocabulary
predates that capability and was never reconciled with it. The result is a
set of properties that mirror a document's own `## References` links, plus
one paired-declaration rule that requires by hand a reverse edge the index
already computes.

This RFC proposes a standard — declare each relationship once and derive
the reverse — and applies it to the existing vocabulary: rename the
feature blueprint's mis-typed FRD edge, drop two transient properties that
are described in three places at once, trim the property that most purely
mirrors prose, mark an unused property honestly, and replace one pairing
with derivation. The engine and the index schema are untouched; every
proposal consumes the derived graph rather than changing it.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins (self) |
| Phase executed | Reviewed the frontmatter relationship vocabulary (`schemas/frontmatter.strata`) against the reference-integrity FRD and ADR-009 while grooming [[frd-spec-reference-integrity]] |
| Documents exercised | `schemas/frontmatter.strata`, `schemas/profile.strata`, `specs/features/reference-integrity/reference-integrity-frd.md`, `specs/features/spec-index/spec-index-frd.md`, `specs/adrs/ADR-009-nodes-and-edges-ownership.md`, `specs/adrs/ADR-012-reference-conventions.md` |
| Documents produced | `specs/standards/document-relationship-standards.md` (draft) |
| Skill(s) followed | `spec` v1.7 |
| Templates / standards read | `standards-template.md` v1.2, `rfc-template.md` v1.1 |
| Plugin version | 0.14.0 |
| Session date | 2026-08-06 |
| Corpus scanned | `specs/` on 2026-08-06 (29 documents) |

## Evidence

1. The reference-integrity FRD records the body reference graph in both
   directions in the index (AC-SPECREF-001.1, AC-SPECREF-001.3) and
   exposes the reverse as `spec refs <slug>` (AC-SPECREF-003), yet the
   frontmatter vocabulary still carries relationship properties whose
   values re-state a document's own prose citations. No standard reconciles
   the two — nothing tells an author that a mirrored relationship is now
   redundant, so the mirrors accumulate.
2. `related` (on the shared `DocumentFrontmatter` envelope) appears on 11
   of the 29 corpus documents; in each, every listed slug also appears in
   that document's `## References` section as a `[[slug]]` — the graph the
   index now derives outbound and inbound. `related` is the property that
   most purely mirrors prose, and no machine consumer keys on it beyond the
   generic reference resolution every `[[slug]]` already gets.
3. `depends_on` (`SystemBlueprint`) appears on 0 corpus documents;
   `superseded_by`, `system`, and `containers` are likewise unused today.
   `depends_on` is aspirational — its doc-comment describes a relationship
   nothing in the corpus tracks.
4. `FeatureBlueprint.sources` is typed `[Slug]` (a required list) but its
   doc-comment states it is the FRD the blueprint realizes, "one to one."
   `Frd.sources` is `[Slug]?` — many, optional, genuine provenance. One
   property name spans two relationships with different cardinality and
   different required target types, which the `ContractEdges` doc-comment
   in the same schema names as the way "a vocabulary drifts."
5. `updates` and `creates` (`FeatureBlueprint`) each appear on exactly one
   corpus document — the single feature blueprint. Both are also written
   in that blueprint's prose "Documentation Changes" section
   (`feature-blueprint-template.md` lines 288–295), and `creates` is
   cross-checked a third time by `blueprint-readiness-checklist` check #5
   ("contracts listed in `creates` exist with status `draft`"). The same
   change set is thus described in three places; `updates` has no consumer
   at all, and the authoritative record of what a transient feature changes
   is its `plan` (the beads epic or task-file directory the schema already
   points to).
6. The shipped configuration pairs `children` with `parent_slug`
   (AC-SPECREF-004.2), requiring both ends and reporting a finding against
   both when either is missing (AC-SPECREF-002.5) — although a promoted
   component has exactly one parent, and the index derives either end from
   the other. The pairing enforces by hand a reverse edge the tool computes
   for free.

## Proposals

### P1 — Adopt the Document Relationship Standard

**Problem.** No standard states that a relationship recorded once is enough
now that the index derives the reverse. Its absence is the root cause of
Evidence 1: authors mirror relationships into frontmatter because nothing
tells them the tool already holds the other half.

**Change.** Adopt `specs/standards/document-relationship-standards.md`
(authored in draft alongside this RFC), advancing it `draft` → `active`.
It establishes: declare each relationship once and derive the reverse; a
frontmatter relationship property is justified only if it carries a typed
relationship a consumer keys on, names a target not in prose, or is the
only home for the edge; the retaining end is chosen by singularity, then
intrinsic-over-host-varying, then authored-at-decision; and a paired
declaration is reserved for a document that must be read without the index.
P2–P7 are that standard applied to the current vocabulary.

**Evidence.** 1, 2. The next author adding a relationship property has a
rule to check it against instead of a precedent to copy.

### P2 — Rename `FeatureBlueprint.sources` to `frd: Slug`

**Problem.** `FeatureBlueprint.sources: [Slug]` is a required list
documented as a strict one-to-one edge to an FRD (Evidence 4), overloading
the name `Frd.sources` already uses for optional many-valued provenance.
The list type cannot express the 1:1 constraint, and the shared name
cannot express that this target must be an FRD.

**Change.** In `schemas/frontmatter.strata`, replace
`FeatureBlueprint.sources: [Slug]` with `frd: Slug` (single, required).
Register `frd` as a reference-bearing property for the feature-blueprint
type, and add a check that its target declares `type: frd`. Update
`feature-blueprint-template.md` and any skill prose that names `sources`
on a feature blueprint. `Frd.sources` is unchanged.

**Evidence.** 4. The cardinality and target type become expressible in the
schema and checkable, and one name stops spanning two relationships.

### P3 — Drop `updates` from `FeatureBlueprint`

**Problem.** `updates` is used once, duplicated in the blueprint's prose,
and read by no consumer (Evidence 5). The completion signal its
doc-comment claims — "the feature is not done until every document it names
has been updated" — is owned authoritatively by the `plan`, which tracks
done-ness natively.

**Change.** Remove `updates` from `FeatureBlueprint` in
`schemas/frontmatter.strata` and from `feature-blueprint-template.md`. The
documents a feature changes remain described in the blueprint's
"Documentation Changes" prose and tracked in its `plan`.

**Evidence.** 5. One transient, unconsumed, thrice-recorded property
removed; no information lost.

### P4 — Drop `creates` from `FeatureBlueprint` and remove its readiness check

**Problem.** `creates` records, in frontmatter, documents the feature
introduces — the same set the blueprint's prose lists and, per this
project's grooming workflow, the set grooming derives directly from the
blueprint when it reads it to raise create/update work orders for
container and component blueprints. `blueprint-readiness-checklist` check
#5 then re-verifies the frontmatter copy (Evidence 5). Describing the same
change in frontmatter, in prose, and in a checklist consumer is three
sites that must agree; the frontmatter copy earns its keep under none of
the standard's three tests.

**Change.** Remove `creates` from `FeatureBlueprint` in
`schemas/frontmatter.strata` and from `feature-blueprint-template.md`. Edit
`blueprint-readiness-checklist` check #5 in **both** copies —
`plugin/skills/spec/references/blueprint-readiness-checklist.md:63` and
`plugin/skills/groom-feature-blueprint/references/shared/blueprint-readiness-checklist.md:68`
— to delete the clause "; contracts listed in `creates` exist with status
`draft`", leaving "Every `[[slug]]` reference resolves to an existing
document." Grooming continues to read the blueprint prose to derive the
documentation work orders, so the create set is described once.

**Evidence.** 5. This reverses the "keep `creates`" recommendation from the
review that produced this RFC: keeping it preserved a readiness check but
kept the change described in a second place, and grooming already derives
the work from the blueprint. See Considered and Not Proposed.

### P5 — Remove `related` from the document envelope

**Problem.** `related` is the property that most purely mirrors prose:
every value in the corpus is already a `## References` `[[slug]]`, and no
consumer keys on it (Evidence 2). Under the standard it earns its keep
under none of the three tests.

**Change.** Remove `related` from `DocumentFrontmatter` in
`schemas/frontmatter.strata` and drop it from the shipped
`reference_properties`. The peer relationships it named are answered by
`spec refs <slug>` over the prose graph. Migration: 11 documents carry the
property; removing it loses nothing, since each value is already cited in
that document's `## References`. A reviewer confirming this per document
is the whole of the migration.

*This is the highest-blast-radius proposal — it touches the shared
envelope and 11 documents — and is independently adoptable. A team not
ready to drop `related` corpus-wide can instead adopt the narrower rule
from the standard: keep `related` only for a peer not otherwise cited in
prose, and stop mirroring the References section into it.*

**Evidence.** 2. The envelope stops carrying a hand-maintained copy of the
derived graph.

### P6 — Mark `depends_on` aspirational, or drop it

**Problem.** `depends_on` (`SystemBlueprint`) is unused (Evidence 3), so it
reads as a maintained relationship the corpus does not actually track. But
it is the one relationship here with no other home: a system→system
dependency is neither composition (`containers`) nor a contract
(`contracts_consumed`), and a top-level system has no parent to own it
under [[adr-009]]. It clears the standard's third test.

**Change.** Either (a) drop `depends_on` until a real system blueprint
needs it and re-introduce it then, or (b) keep the property and annotate
its doc-comment in `schemas/frontmatter.strata` as aspirational — recorded
on the system's own Context view, and unused in the corpus so far. Do not
pair it; derive "depended-on-by" from inbound edges. This RFC recommends
(b): the slot is legitimate and cheap, and dropping-then-adding churns the
schema.

**Evidence.** 3. The property stops being mistaken for a tracked
relationship, without losing the only valid home for a system-level edge.

### P7 — Derive `children` from `parent_slug`; drop the pairing

**Problem.** The `children` ↔ `parent_slug` pairing requires both ends
though the index derives either from the other, and a promoted component
has exactly one parent (Evidence 6). By the standard's singularity
tie-breaker, `parent_slug` is the retaining end.

**Change.** Keep `parent_slug` on the child (singular, intrinsic, set at
promotion) and derive the parent's `children` roster from inbound
`parent_slug` edges. Remove the `children` ↔ `parent_slug` pairing from the
shipped configuration (AC-SPECREF-004.2). Add a one-line companion note to
[[adr-009]] clarifying that deriving the membership roster from child
back-pointers preserves its "one authoritative home, no drift" intent —
the ADR's rule governs neighbour edges among children, not the
parent-child membership edge — so this refines its mechanism rather than
reversing it.

**Evidence.** 6. One pairing check that could only ever fail on redundant
hand-work is replaced by a derivation, and ADR-009's intent is stated
rather than left to inference.

## Considered and Not Proposed

- **Removing `superseded_by`** — unused in the corpus today, but
  load-bearing: it drives the retired-target checks (AC-SPECREF-002.4,
  AC-SPECREF-002.6) and names a *typed* successor a prose citation cannot
  convey. It clears the standard's first test. Keep.
- **Removing `contracts_provided` / `contracts_consumed`** — [[adr-009]]
  explicitly keeps these on the element because a contract's identity is
  host-independent; they are typed edges a machine consumer reads. Keep.
- **Removing the composition rosters** (`containers`, `components`,
  `modules`) — these are candidates for the same derive-from-child
  treatment as P7, but the child-to-parent back-pointer does not exist for
  every level today, so deriving them is not yet mechanical. Deferred, not
  rejected; revisit once the back-pointers are in place.
- **Keeping `creates` with its readiness check** — the recommendation the
  originating review first made. Rejected in favour of P4: it preserved a
  check but kept the create set described in frontmatter as well as prose,
  and grooming already derives the work orders from the blueprint, so the
  frontmatter copy is a third site to keep in agreement for no consumer the
  prose does not already serve.

## Non-Goals

- No new document types.
- No change to the reference-integrity engine or the index schema. Every
  proposal consumes the derived graph (AC-SPECREF-001.3, AC-SPECREF-003);
  none alters how it is built.
- No change to the `[[slug]]` prose reference conventions of [[adr-012]].
- No change to how `plan` tracks feature completion; P3 relies on it as-is.

## Adoption

P1 lands first — it is the standard P2–P7 apply. P2, P3, and P4 touch only
the feature-blueprint surface and its single corpus instance, are low risk,
and can land together. P5 (drop `related`) and P7 (derive `children`) touch
the shared envelope and the shipped configuration respectively, have the
widest blast radius, and are each independently adoptable — land them one
at a time. P6 is a one-line schema annotation. Each proposal that removes a
property carries a mechanical migration, because the data already lives in
prose (`related`, `creates`, `updates`), in the `plan` (`updates`), or in
the counterpart the index derives from (`children`). Expected overall
effect: the frontmatter an author maintains by hand shrinks to the edges
the index cannot derive, and every relationship in the corpus has one
authoritative source.

## Resolution

<!-- FILLED BY THE RESOLVING REVIEW — leave empty. -->
