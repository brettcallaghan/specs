---
title: "RFC: Reconcile the document-relationship vocabulary with the
  derived reference graph"
description: Adopt a standard for declaring cross-document relationships
  once and deriving the reverse, and apply it to the framework's
  frontmatter — rename the feature blueprint's FRD edge, drop `updates`
  and `creates`, trim `related`, mark `depends_on` aspirational, derive
  `children` from `parent_slug`, and replace the C4 composition rosters
  with child back-pointers
slug: rfc-0012
type: rfc
status: retired
updated_at: 2026-08-06T00:00:00Z
origin: self
skill: spec
skill_version: "1.7"
model: "claude-opus-4-8"
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "0.14.0"
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
and the four C4 composition rosters with derivation from the child. The engine and the index schema are untouched; every
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
7. The composition rosters — `containers` (`SystemBlueprint`),
   `components` and `modules` (`ContainerBlueprint`), `components`
   (`ModuleBlueprint`) — carry exactly one non-empty value across the
   corpus's three blueprint documents: `container-spec-cli` lists
   `module-document-core` and `module-spec-profile` in `modules`.
   `components` is present and empty on all three; `containers` appears
   on no document, there being no system blueprint yet. In the other
   direction nothing records that `container-spec-cli` composes those two
   modules: `ContainerBlueprint.system` exists but is declared on 0
   documents, and `ModuleBlueprint` and `ComponentBlueprint` have no
   parent property at all — although [[adr-009]] already tells components
   and modules to "point to their host container(s) instead" of carrying
   a context view.

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
P2–P8 are that standard applied to the current vocabulary.

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

### P8 — Replace the composition rosters with child back-pointers

**Problem.** The composition rosters — `containers` on a system,
`components` and `modules` on a container, `components` on a module —
are the last hand-maintained mirrors of an edge the index derives, and
P7's reasoning applies to them unchanged. The originating review deferred
them for a mechanical reason rather than a principled one: three of the
four C4 levels have no back-pointer to derive a roster from (Evidence 7),
so removing a roster there would delete the edge rather than relocate it.
Supplying the missing back-pointers removes the only obstacle.

**Change.** Give every C4 child the property its parent's roster stands
in for, then drop the rosters. In `schemas/frontmatter.strata`:

- `ContainerBlueprint.system: Slug?` — unchanged. It is already the
  container's back-pointer to the system that composes it, and its
  absence already records the documented fact that a container is
  standalone rather than an oversight.
- `ComponentBlueprint.container: Slug?` — new. Names the container **or
  module** the component is composed into. Singular by the standard's
  singularity tie-breaker: a component promoted under [[adr-007]] is
  promoted out of exactly one parent's roster. Optional for the same
  reason `system` is — a parent with no blueprint of its own is a real
  state, and absence records it.
- `ModuleBlueprint.contained_by: [Slug]?` — new. Names the containers,
  or the modules, that compose this module. A list because a shared
  module is composed into more than one host — the fact [[adr-009]] cites
  when it denies modules a context view and tells them to point at their
  host container(s) instead, a direction the vocabulary has never given
  them a property for. Modules are the one level where singularity does
  not settle which end holds the edge, since both ends are lists; it is
  settled by ADR-009's instruction and by uniformity, so that one
  derivation rule covers all three levels.
- Remove `SystemBlueprint.containers`, `ContainerBlueprint.components`,
  `ContainerBlueprint.modules`, and `ModuleBlueprint.components`. Each
  parent's roster is then derived from inbound `system`, `container`, and
  `contained_by` edges, exactly as P7 derives `children`.

Register `container` and `contained_by` as reference-bearing for their
types in the shipped configuration (AC-SPECREF-004.2). Pair neither with
anything: under P1's standard a pairing is reserved for a document that
must be read without the index, and these are the halves the index
derives. Update `system-blueprint-template.md`,
`container-blueprint-template.md`, `module-blueprint-template.md` and
`component-blueprint-template.md`, and the relationship-property listing
in `frontmatter-standards.md`.

**Migration.** One non-empty roster exists corpus-wide (Evidence 7).
`module-document-core` and `module-spec-profile` each gain
`contained_by: ["container-spec-cli"]`; four roster properties — one
`modules`, three empty `components` — are deleted. `containers` migrates
nothing.

**What this does not change.** [[adr-008]]'s block-XOR-link invariant and
[[adr-012]]'s roster-heading convention govern the parent's **body**, not
its frontmatter: a child is a fenced record block XOR a `[[slug]]`
blueprint link under a `### <Display Name>` heading in the parent's
roster section. That is where a reader meets the children, and where a
promoted component's heading anchor keeps resolving. Only the frontmatter
copy of the membership list goes; the roster section is untouched.

**Evidence.** 6, 7. This supersedes the deferral this RFC first recorded
under Considered and Not Proposed, whose stated blocker — that the
child-to-parent back-pointer does not exist at every level — is what the
change above supplies. Afterwards no C4 parent declares its children in
frontmatter, and every composition edge is declared once, on the child,
by the author who composed it.

## Considered and Not Proposed

- **Removing `superseded_by`** — unused in the corpus today, but
  load-bearing: it drives the retired-target checks (AC-SPECREF-002.4,
  AC-SPECREF-002.6) and names a *typed* successor a prose citation cannot
  convey. It clears the standard's first test. Keep.
- **Removing `contracts_provided` / `contracts_consumed`** — [[adr-009]]
  explicitly keeps these on the element because a contract's identity is
  host-independent; they are typed edges a machine consumer reads. Keep.
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
at a time. P6 is a one-line schema annotation. P8 lands after P7, whose
derivation it repeats at the four C4 levels and whose ADR-009 companion
note covers it: it touches all four blueprint templates and the three
blueprint documents in the corpus, and its back-pointers must be in place
before the rosters are removed, so both halves belong in one change. Each
proposal that removes a property carries a mechanical migration, because
the data already lives in prose (`related`, `creates`, `updates`), in the
`plan` (`updates`), or in the counterpart the index derives from
(`children`, the composition rosters). Expected overall effect: the
frontmatter an author maintains by hand shrinks to the edges the index
cannot derive, and every relationship in the corpus has one authoritative
source.

## Resolution

**Triaged 2026-08-06.** 8 proposals: 8 adopted, 4 of them with
refinement. None deferred, none rejected. The human directed at the
outset that every proposal was to be adopted — this RFC came out of
their own planning session, and the vocabulary is being simplified
before `reference-integrity` is implemented so that the feature is not
built to maintain frontmatter the framework should not carry.

| Item | Outcome | Bead |
|------|---------|------|
| P1 — adopt the Document Relationship Standard | **adopted with refinement.** The document was already at `status: approved`, not `draft` as the RFC states, so the transition made was `approved` → `active`. Version left at `0.1`: this is its first release, not a revision. | `agent-plugins-l7k` |
| P2 — rename `FeatureBlueprint.sources` to `frd: Slug` | **adopted with refinement.** The rename landed as proposed. The refinement is the "check that its target declares `type: frd`" half, which had nowhere to be registered — see *On the typed-target check* below. | `agent-plugins-1h1` |
| P3 — drop `updates` | **adopted with refinement.** Extent check found two sites the RFC did not name: the template's own Contracts instruction block ("list them in `updates`"), and [[frontmatter-standards]], which did not merely mention the property but *mandated* it ("the template SHALL include the `updates` and `creates` properties"). Removing the property without those would have left a standard requiring a property the schema rejects. | `agent-plugins-f48` |
| P4 — drop `creates` and its readiness check | **adopted as proposed**, with the same two extra sites as P3. Both copies of `blueprint-readiness-checklist` were updated through the authored copy plus `scripts/sync-shared-references.sh`; the vendored copy is generated and was not hand-edited. | `agent-plugins-1rx` |
| P5 — remove `related` from the envelope | **adopted with refinement.** See *On P5's extent and sufficiency* below — the blast radius was 48 files rather than 11, and one document would have lost an edge. | `agent-plugins-2b1` |
| P6 — mark `depends_on` aspirational | **adopted as option (a): dropped.** Landed first as option (b), the RFC's recommendation; reversed in review — see *Reversed in review* below. `SystemBlueprint.depends_on` is gone. | `agent-plugins-yvo` |
| P7 — derive `children` from `parent_slug`; drop the pairing | **adopted with refinement.** See *On P7* below — neither property exists in this vocabulary, so what was removed was the framework's own worked example of a pairing its new standard forbids, not a live configuration. | `agent-plugins-0qm` |
| P8 — replace the composition rosters with child back-pointers | **adopted with refinement.** The four rosters were removed as proposed, and the migration was exactly what Evidence 7 predicted: two `contained_by` additions, four roster deletions. The back-pointers landed differently — one `contained_by: [Slug]?` on a shared `CompositionEdge` struct, inherited by both module and component, rather than the proposed singular `ComponentBlueprint.container`. See *Reversed in review* below. | `agent-plugins-4yv` |

**Provenance check.** `skill: spec` at `skill_version: "1.7"` — current at
triage. `plugin_version: 0.14.0` — current at triage. No cited text had
moved or been fixed since; nothing was already-addressed.

**Self-origin, disclosed.** This RFC carries `origin: self` and was
authored in a separate planning session by the model named in its `model`
field. The triaging agent is a different model, so this is not the
same-agent self-adoption the skill requires disclosure of — but it is
still this repository critiquing its own vocabulary, and the human gate
did the substantive gatekeeping, including directing the adopt-now
classification before triage began.

**Dedupe.** No match in RFCs 0001–0011. One adjacent deferred bead:
`agent-plugins-6tp` (RFC 0006 P4) asks whether `sources` can record an
input that is not a spec document. P2 does not resolve it, but removes
the overload that made it harder to reason about — after P2, `sources`
names one relationship rather than two.

**On P5's extent and sufficiency.** The RFC scoped its evidence to the 29
documents under `specs/`. Enumerating `^related:` inside YAML frontmatter
across the whole repository found **48 files**: 12 under `specs/`, 15
RFCs, 13 templates, and 8 skill reference documents.
`scripts/validate-corpus.py` validates the RFC queue and the skill
reference files against the same bundled schema, whose document types set
`additionalProperties: false`, so a partial migration would have failed
the corpus gate rather than left a harmless remnant. The retired RFCs and
every template carried `related: []` and migrated losslessly.

The sufficiency check found one real gap in the claim that every value is
already cited in prose. It holds for 47 of the 48:
`specs/architecture/module-spec-profile.md` listed `frd-spec-validation-cli`
in `related` and cited it nowhere in its body, so removing the property
would have deleted the edge rather than derived it. That document's
`## References` section gained the citation — along with `adr-008` and
`adr-012`, cited in its body but not under References — before the
property was removed.

**On P7.** `children` and `parent_slug` are not properties of this
vocabulary. Neither appears in `schemas/frontmatter.strata`; every
corpus-side occurrence is an *example* — in this FRD's glossary, in
AC-SPECREF-002.5, in AC-SPECREF-004.2, and in the `PropertyPairing`
doc-comment in `schemas/profile.strata`. (The `parent_slug` in the work
order template is a beads field in a different namespace and was left
alone.) So there was no shipped pairing to remove. What the framework was
doing was presenting, as its own canonical example, a pairing that P1's
standard forbids. Those four sites now carry an adopter-side example and
state that the spec-driven-dev profile registers no pairing. The pairing
*capability* is untouched: AC-SPECREF-004.1 and AC-SPECREF-004.3 stand,
because an adopting team may still need one.

**On the typed-target check** *(human-approved at the step-4 gate).* P2
asks for a check that a feature blueprint's `frd` resolves to a document
of type `frd`. Nothing could register that: `reference_properties` is a
list of names and carries no target type. The adopted form is therefore
slightly wider than P2's text — `schemas/profile.strata` gained
`typed_reference_properties`, a property→type map, and the
reference-integrity FRD gained **AC-SPECREF-002.10** for the check itself.
The human was shown this trade-off and chose it over registering the
property without the check.

**Reversed in review.** Both reversals came from the repository owner
reviewing the diff on PR #14, and both overrule a choice made earlier in
this triage — the second overrules one the same human had approved at the
step-4 gate, on the stronger evidence of seeing the schema itself.

- **`depends_on` dropped rather than annotated (P6).** The RFC recommended
  option (b), keep-and-annotate, on the grounds that the property is the
  only home for a system-to-system edge. The owner's objection is that
  this is not a home worth having: a slug list is an ineffective way to
  express what one system needs from another, the blueprint's "Consumed
  Systems and Contracts" section already says it with the detail a reader
  needs, and a repository spanning more than one system is rare enough
  that nobody would maintain the property anyway. `SystemBlueprint` now
  carries no relationship property beyond `ContractEdges`. The Document
  Relationship Standard's third test — "only home for an edge" — used
  `depends_on` as its worked example and no longer cites a live instance;
  the test stands for an adopting profile, which is now what it says.
- **`contained_by` is the module level's alone (P8).** This took two
  rounds, and the second corrected a premise rather than a placement.
  The RFC proposed a singular `ComponentBlueprint.container`; triage first
  generalized it to a `contained_by` list on both module and component,
  reasoning that any unit worth its own document is worth reusing. The
  owner rejected that generalization on C4 itself: **every C4 element has
  exactly one parent** — a container belongs to one system, a component to
  one container — and Module is this framework's *addition* to C4, whose
  entire purpose is to be composed into more than one container. So
  multi-containment is not a property of promotion; it is the definition
  of a module.

  As landed: `contained_by: [Slug]?` exists on `ModuleBlueprint` and
  nowhere else, documented as the only multi-parent edge in the
  vocabulary and as the reason the module level exists.
  `ComponentBlueprint` carries no parent property at all — a component's
  single parent is recorded where a reader meets the child, in the
  parent's roster section, which links each promoted child by `[[slug]]`
  under its own heading per ADR-008 and ADR-012, and the index derives the
  edge from that inbound link. `ContainerBlueprint.system` is unchanged.
  The shared `CompositionEdge` struct triage introduced in the first round
  was removed with the premise that motivated it.

**On the strata toolchain.** Regenerating the derived artifacts needs a
`strata` newer than the 0.7.0 on PATH, which drops struct inheritance
under `additionalProperties: false` and would have silently emptied every
document type (open bead `agent-plugins-j7f`). Building from the attached
checkout at HEAD fixed that but introduced an unrelated emitter change —
generated `UnmarshalJSON` methods enforcing required fields — which moved
two snapshot findings from `frontmatter/missing-property` to
`frontmatter/parse` and failed the CLI snapshot suites. That is a
toolchain upgrade, not this RFC's change, so the regeneration was pinned
to strata `8997153`, the commit `agent-plugins-j7f` records as producing
byte-identical output to what is committed. The emitter change is worth
adopting deliberately; it is not adopted here.

**Files changed.** `schemas/frontmatter.strata`, `schemas/profile.strata`
and their generated artifacts (`cli/core/schema/`, `cli/core/specmodel/`,
`plugin/skills/spec/references/generated/`); 48 documents migrated off
`related`; 4 corpus documents migrated for P2, P3, P4 and P8;
`specs/adrs/ADR-009-nodes-and-edges-ownership.md`;
`specs/features/reference-integrity/reference-integrity-frd.md`;
`specs/standards/document-relationship-standards.md`; 9 templates and 8
standards/guides under `plugin/skills/`.

**Version bumps.** Templates: feature-blueprint 1.3→1.4, system-blueprint
1.4→1.5, container-blueprint 1.5→1.6, module-blueprint 0.7→0.8,
component-blueprint 1.5→1.6, contract 1.3→1.4, frd 1.1→1.2, work-order
1.2→1.3, rfc 1.1→1.2, skill-rfc 0.2→0.3. Standards and guides:
frontmatter-standards 0.1→0.2, blueprint-readiness-checklist 0.5.0→0.6.0,
template-authoring-standards 0.12→0.13, work-order-definition-of-ready
1.1→1.2, work-order-field-guide 1.1→1.2, document-readiness-checklist
0.2→0.3, c4-guide 0.1→0.2, cross-document-references 0.1→0.2, ears-guide
0.1→0.2. `document-relationship-standards` stays at 0.1 — first release.
Plugin 0.14.0 → 0.15.0.
