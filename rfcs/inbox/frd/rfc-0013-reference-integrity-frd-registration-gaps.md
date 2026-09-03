---
title: "RFC: The reference-integrity FRD requires checks whose
  configuration it never asks a profile to carry, and illustrates two
  criteria with a property the schema no longer has"
description: Require the active profile to register the lifecycle facts
  AC-SPECREF-002.6 depends on; say what a defect's position means when
  the registration is compiled in; and sweep an FRD's citations when the
  vocabulary it depends on changes underneath it
slug: rfc-0013
type: rfc
status: ready-for-review
updated_at: 2026-08-07T00:00:00Z
origin: self
skill: spec
skill_version: "1.7"
model: "claude-opus-5"
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "0.15.0"
---

# RFC: Registration gaps and stale citations in the reference-integrity FRD

## Summary

A feature blueprint and one ADR were authored from
[[frd-spec-reference-integrity]]. The FRD is unusually thorough — 25
acceptance criteria, each with a worked example, and a rationale under
every requirement — and it still left the blueprint inventing three
configuration keys and one position semantics that no criterion asks
for, because REQ-SPECREF-004 enumerates what the profile registers and
its enumeration is narrower than what REQ-SPECREF-002 checks. Separately,
two criteria illustrate themselves with a frontmatter property that PR
#14 deleted four days after the FRD was approved, and a third claims to
add a configuration property that PR #14 had already added. The
proposals close the configuration gap and give the project a rule for
keeping an FRD's citations true when the vocabulary beneath it moves.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins |
| Phase executed | Authored a feature blueprint and one ADR from an approved FRD |
| Documents exercised | [[frd-spec-reference-integrity]]; [[container-spec-cli]], [[module-document-core]], [[module-spec-profile]]; [[contract-specs-configuration]]; [[document-relationship-standards]]; [[adr-012]], [[adr-015]], [[adr-016]]; `schemas/frontmatter.strata`, `schemas/profile.strata`, `schemas/index.strata` |
| Documents produced | [[feature-spec-reference-integrity]], [[adr-018]] |
| Skill(s) followed | spec v1.7 |
| Templates / standards read | feature-blueprint-template v1.4; adr-template v1.3; `blueprint-readiness-checklist.md` v0.6.0; document-readiness-checklist v0.3; `ac-traceability-standards.md` v0.2.0 |
| Plugin version | 0.15.0 |
| Session date | 2026-08-07 |

## Evidence

1. AC-SPECREF-002.6 requires a `warning` WHEN "a document at a status
   the profile treats as live references a document at `status:
   retired`", naming "its `superseded_by` value when it has one"
   (`specs/features/reference-integrity/reference-integrity-frd.md`
   line 273). REQ-SPECREF-004 is the requirement that says what the
   configuration determines, and AC-SPECREF-004.1 scopes it to "which
   frontmatter properties are reference-bearing and which pairs of
   properties form a paired relationship", plus
   `typed_reference_properties`. Nothing asks the profile to register
   which statuses are live, which are retired, or which property names
   a successor — yet [[module-document-core]]'s Change Guidance forbids
   the engine from holding any of those: "A capability that cannot be
   expressed without naming a specific document type belongs in a
   profile, not here." The blueprint had to invent three registration
   keys (`live_statuses`, `retired_statuses`, `successor_property`) and
   add them to `schemas/profile.strata` in its MODEL delta and to the
   `Profile` interface in its XPT delta, none of which any criterion
   asks for.

2. AC-SPECREF-004.3 requires that a misconfigured pairing be reported
   with "the offending key and its position and SHALL NOT fall back to
   a default" (same file, line 418). A position implies a file. In v1
   there is none: [[contract-specs-configuration]] states the
   registration surface is "Internal — not an extension point outside
   this repository, and not reachable from `specs.toml` in v1", and
   [[module-spec-profile]] confirms the registration is compiled in
   ("everything is compiled in: the schema through `//go:embed` and the
   per-type table as a Go map"). The blueprint pinned a resolution the
   FRD does not authorize — report the finding at the `[profile].name`
   value's span in `specs.toml`, as the only file position a compiled-in
   registration has — and recorded it as a decision rather than a
   criterion.

3. AC-SPECREF-001.7's example reads "A document's `related: ["adr-012",
   "adr-002"]` records two outbound references" (line 201) and
   AC-SPECREF-002.3's reads "`related: ["work-order-template"]` reports
   as unresolved" (line 258). `related` was removed from the frontmatter
   envelope by RFC-0012 P5, adopted and merged in PR #14 on 2026-08-06 —
   the day the FRD was approved. No document may declare it, so both
   criteria are illustrated by an input the schema now rejects. The
   blueprint's PREG delta had to work out the real property set from
   `schemas/frontmatter.strata` instead.

4. AC-SPECREF-004.1's example states the feature "adds one:
   `typed_reference_properties`, mapping a reference-bearing property to
   the document `type` its target must declare" (lines 402–404). The
   property already existed: PR #14 added it to
   `schemas/profile.strata` and to the generated
   `specmodel.DocumentTypeRegistration` as the adopted form of RFC-0012
   P2, whose Resolution records that the typed-target check "had nowhere
   to be registered" and that this was the fix. This feature is the
   property's first consumer, not its author. The claim had to be
   checked against the schema before the MODEL delta could be written.

5. AC-SPECREF-004.2's example enumerates the shipped reference-bearing
   properties as "`superseded_by`, `sources`, `frd`, `system`,
   `contained_by`, `contracts_provided` and `contracts_consumed`" (lines
   417–420). `schemas/frontmatter.strata` also declares `provider: Slug`
   and `consumers: [Slug]?` on `Contract` — a contract's two
   load-bearing edges. The blueprint registered them anyway and recorded
   the divergence from the FRD's example in the PREG delta, because
   omitting them would leave those edges unchecked.

6. AC-SPECREF-002.4 required, unconditionally, that a document at
   `status: retired` declare a resolvable `superseded_by`. A scan of
   this corpus while authoring the LIFE delta found its only retired
   document is [[feature-spec-index]] — a feature blueprint,
   which the feature blueprint template retires when "the feature is
   merged and the Definition of Done is complete", not because anything
   replaced it. The criterion would have reported an `error` against
   every retired feature blueprint the framework produces. It was raised
   as the blueprint's sole Open Question, which held the blueprint at
   `draft`, and the FRD's owner resolved it by deleting the criterion.

## Proposals

### P1 — Register the lifecycle facts REQ-SPECREF-002 checks

**Problem.** REQ-SPECREF-002 checks two facts that are pure vocabulary —
which statuses mean a document is live or retired, and which property
names a successor — and REQ-SPECREF-004, the requirement that decides
what configuration carries, never mentions either. AC-SPECREF-004.1
says only: "The configuration SHALL determine which frontmatter
properties are reference-bearing and which pairs of properties form a
paired relationship". The engine may not hold these facts
([[module-document-core]] Change Guidance), and the configuration is not
asked for them, so the design has a hole exactly where the two
requirements meet. The blueprint filled it by inventing keys.

**Change.** Add to REQ-SPECREF-004:

> **AC-SPECREF-004.4:** The configuration SHALL determine which
> `status` values the profile treats as live, which it treats as
> retired, and which frontmatter property names a retired document's
> replacement.
> - *Example:* The shipped configuration registers `retired` as its
>   only retired status, every other status as live, and
>   `superseded_by` as the successor property, so AC-SPECREF-002.6 can
>   quote a replacement without the engine knowing either name.

and extend AC-SPECREF-004.2's example to name the same three
registrations, so the shipped configuration is specified as a whole
rather than in two places that can drift.

**Evidence.** Finding 1. The next agent authoring against this FRD reads
what the profile must register instead of deriving it from a check's
wording, and the three keys arrive as a specified surface rather than as
a blueprint's invention that grooming has to take on trust.

### P2 — Say what "position" means for a defect in a compiled-in registration

**Problem.** AC-SPECREF-004.3 demands "the offending key and its
position", and in v1 the registration it validates has no file to have a
position in. The criterion is written as though the profile were
authored data, which [[contract-specs-configuration]] explicitly says it
is not until a key is promoted. An implementer either invents a
position, drops the requirement, or reports a Go source location no
reader can act on.

**Change.** Amend AC-SPECREF-004.3 to state the v1 answer and keep the
forward path open:

> **AC-SPECREF-004.3:** IF a pairing names a property that is not
> registered as reference-bearing, THEN the CLI SHALL report the
> offending key and SHALL NOT fall back to a default.
> - *Example:* A pairing of `children` with `parent` where only
>   `parent_slug` is registered names the key, rather than silently
>   checking nothing.
> - *Note:* WHILE the active profile is compiled in and unreachable
>   from `specs.toml` ([[contract-specs-configuration]]), the finding is
>   positioned at the `[profile].name` value that selected it — the only
>   file position such a registration has. WHEN a registration becomes
>   authored data, the position is the offending key's own span.

**Evidence.** Finding 2. The blueprint's pinned decision becomes a
stated criterion, and the note is where a future promotion of the
registration surface to `specs.toml` will read as an obligation rather
than a surprise.

### P3 — Sweep an FRD's citations when the vocabulary beneath it changes

**Problem.** Three of the FRD's examples describe a world that stopped
existing between approval and consumption: two cite `related`, deleted
by RFC-0012 P5 in PR #14 (finding 3); one claims to add
`typed_reference_properties`, added by RFC-0012 P2 in the same PR
(finding 4); one enumerates the reference-bearing properties without
`provider` and `consumers` (finding 5). Nothing was careless here — PR
#14 was a deliberate, well-triaged vocabulary change that named its
blast radius carefully, enumerating 48 files carrying `related`. What it
did not do is revisit the *acceptance criteria* that describe the
vocabulary, because no rule says to. [[document-relationship-standards]]
governs how a property is introduced or removed and says nothing about
the documents whose requirements cite it.

**Change.** Add to [[document-relationship-standards]], under
"When a Frontmatter Relationship Property Is Justified":

> WHEN a frontmatter relationship property is introduced or removed,
> every `approved` or `active` FRD whose acceptance criteria name that
> property SHALL be updated in the same change.
>
> **Rationale:** An acceptance criterion is read months after it is
> written, by an agent that treats it as current. A criterion
> illustrated with a property the schema rejects costs that agent a
> schema lookup at best and a wrong design at worst — and the FRD is
> the one document type whose examples are load-bearing, because a
> feature blueprint is authored from them rather than from the schema.

and fix the four sites in
`specs/features/reference-integrity/reference-integrity-frd.md`: replace
the two `related` examples with a property the envelope still carries
(`sources` for AC-SPECREF-001.7; a `contracts_consumed` value naming no
member for AC-SPECREF-002.3), restate AC-SPECREF-004.1's example as
consuming `typed_reference_properties` rather than adding it, and add
`provider` and `consumers` to AC-SPECREF-004.2's enumeration.

**Evidence.** Findings 3, 4 and 5. The rule generalizes past this FRD:
[[frd-spec-anchor-references]] and [[frd-spec-slug-rename]] both cite
the same vocabulary and are still `draft`, so the next such change
strands them too unless the sweep is written down.

## Considered and Not Proposed

- **Restoring a successor obligation for retired documents, scoped by
  document type.** Finding 6 is the strongest single piece of evidence
  in this retro, and the fix is obvious — register the obligation per
  type and omit `feature-blueprint`. It is not proposed because the
  FRD's owner, shown the conflict during this session, chose to delete
  AC-SPECREF-002.4 outright and revisit later. Proposing it now would
  relitigate a decision made hours ago with the evidence in hand. The
  FRD's Note on numbering records the identifier as unused and the
  reason, so a future revision starts from a written record rather than
  from this RFC.
- **Requiring an FRD's examples to be machine-checked against the
  active schema.** It would have caught findings 3, 4 and 5
  mechanically, and it is the obviously stronger form of P3. Not
  proposed: an example is prose, the properties inside it are not
  marked up, and building an extractor that distinguishes
  `related: ["adr-012"]` in an example from the same string in a
  rationale is a feature, not a rule. P3's obligation is what can be
  written today; a checker is a reasonable follow-on once the
  reference-integrity feature makes frontmatter properties addressable.
- **AC-SPECREF-001.5, -001.6, -002.2 and -004.3 each join two modal
  verbs in one sentence**, which `document-readiness-checklist.md`
  check 5 forbids ("No requirement joins two modal verbs with 'and'").
  The FRD is `approved` and would fail its own readiness checklist. Not
  proposed as a change to the machinery: the checklist already states
  the rule plainly, so this is authoring that skipped an existing check
  rather than guidance that misled. Worth fixing in the document; not
  worth a proposal.
- **The FRD's Overview cites "29 documents, 256 body references once
  backticked illustrations are excluded" as of 2026-08-06.** A scan on
  2026-08-07 found 30 slug-bearing documents. Counts date by
  construction and the FRD stamps its scan with a date, which is the
  correct handling; no change proposed.

## Non-Goals

- No new document type, and no change to the reference-integrity
  feature's scope. P1 and P2 add configuration surface the blueprint
  already designed; they do not add a check.
- No weakening of any acceptance criterion. P2 removes the word
  "position" from the normative sentence and puts the v1 answer in a
  note, which is a narrowing of an unsatisfiable demand, not of the
  check.
- No re-opening of the retired-successor question — see Considered and
  Not Proposed.

## Adoption

P1 and P2 are independent of each other and of P3; each touches only
`specs/features/reference-integrity/reference-integrity-frd.md` except
for P3's rule, which also touches
`specs/standards/document-relationship-standards.md`.

P1 and P2 SHOULD land together with the feature blueprint's approval,
because [[feature-spec-reference-integrity]] is already `approved` and
its MODEL, XPT and CMD deltas assert exactly what P1 and P2 would
specify. Adopting them makes the blueprint's design traceable to
criteria rather than to its own judgement; declining them leaves three
registration keys and one position rule in a document that retires.
Neither changes what the blueprint would be groomed into.

Backward compatibility: P1 adds an acceptance criterion, so existing
documents stay valid and the new AC identifier is unused today. P2
rewords AC-SPECREF-004.3 without changing what it forbids; no test cites
it yet, since the feature is unimplemented. P3's rule is prospective and
its four citation fixes are corrections to examples, which
`ac-traceability-standards.md` makes non-normative — no citation or test
depends on them.

Expected effect: the next agent authoring or grooming against this FRD
finds the configuration surface specified rather than inferred, and
stops having to check each example against the schema before trusting
it.

## Resolution
