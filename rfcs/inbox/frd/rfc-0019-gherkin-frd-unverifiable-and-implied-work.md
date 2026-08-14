---
title: "RFC: The Gherkin adoption FRD carries one criterion nothing can
  verify, one clause that implies a subsystem, and a warning it would
  raise against every one of its own criteria"
description: Decide what an owning-document tag that resolves to nothing
  does, settle whether the FRD warns about a tag its own guide calls
  optional, make AC-GHK-001.6 a behavior or mark it manual, and resolve a
  note claiming another FRD must change
slug: rfc-0019-gherkin-frd-unverifiable-and-implied-work
type: rfc
status: ready-for-review
updated_at: 2026-08-14T00:00:00Z
origin: self
skill: spec
skill_version: "1.9"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "0.17.0"
---

# RFC: Unverifiable and implied work in the Gherkin adoption FRD

## Summary

A feature blueprint was authored from [[frd-gherkin-adoption]] — the
first document in this corpus to write its own acceptance criteria in
Gherkin, and by some distance the most thorough FRD here: 56 criteria
at the time, nine requirements, a rationale under four of them. It still
left the blueprint deciding five things no criterion states. The largest
— how a prose criterion is recognised and attributed — was a delta of
work implied by the second clause of one criterion, and review answered
it by withdrawing the capability rather than specifying it (see
Considered and Not Proposed, and the criterion count is 55 since). The
four proposals below close the remaining gaps at their source rather
than in the blueprint, where they die when the blueprint retires.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins |
| Audience | project — resolved in this repository, which is what filing under `rfcs/inbox/frd/` records; the generated `Rfc` schema has no `audience` property |
| Phase executed | Authored a feature blueprint from an approved FRD |
| Documents exercised | [[frd-gherkin-adoption]]; [[adr-021]]; [[container-spec-cli]], [[module-document-core]], [[module-spec-profile]]; [[contract-specs-configuration]]; [[testing-standards]]; `specs/standards/gherkin-guide.md`; `schemas/feature.strata`, `schemas/index.strata` |
| Documents produced | [[feature-gherkin-adoption]] |
| Skill(s) followed | spec v1.9; groom-feature-blueprint (Phase 1 gate only) |
| Templates / standards read | feature-blueprint-template v1.4; `blueprint-readiness-checklist.md` v0.6.0; `document-readiness-checklist.md` v0.4; `ac-traceability-standards.md` v0.2.0 |
| Plugin version | 0.17.0 |
| Session date | 2026-08-14 |

## Evidence

Line numbers below are against the revision observed, commit `ee3e332`.
The FRD was edited during the same review that produced this RFC, so a
verifier should search for the quoted text rather than trust an offset.

1. **AC-GHK-001.6 could not be assigned a verification level, because
   it describes no behavior of the CLI.** It reads, at
   `specs/features/gherkin-adoption/gherkin-adoption-frd.md:150-154`:

   > `@AC-GHK-001.6`
   > `Scenario: A standalone feature file sits with the documents of its feature`
   > `Given a feature owned by an FRD with the slug frd-checkout`
   > `When that feature is written as a standalone file`
   > `Then the file sits in the same directory as the FRD that owns it`

   Nothing in the tool places that file; a human does. The blueprint
   declared it `manual` and had to write a paragraph under the coverage
   table justifying which delta "proves" a convention. The criterion
   carries no `@manual` tag, so there is no marker in the FRD to
   reconcile that level against — the exact condition
   `rfcs/inbox/spec/rfc-self-readiness-checklist-frontmatter-and-manual-acs.md`
   P2 names as a blocker, arriving here from the FRD side.

2. **Prose criterion extraction is one clause of one criterion and cost
   a delta block.** AC-GHK-009.3 reads, at line 725: "`Then the CLI
   records criteria from both notations`". That is the whole of what
   the FRD says about reading a prose criterion. Authoring the PROSE
   delta forced three decisions the FRD does not state: which markdown
   form is a criterion (the bolded-identifier list item the three
   existing FRDs use), where its name comes from (nowhere — prose gives
   a criterion no name), and how it finds its feature (its identifier's
   code and number, with the feature record synthesized from the
   heading line, because Markdown Structure's heading tree is not
   built). None is inferable from the clause. Reviewed on PR #13 after
   this RFC was drafted, the capability was withdrawn outright — the
   remaining FRDs are translated instead — which is the resolution
   recorded under Considered and Not Proposed.

3. **An owning-document tag that resolves to nothing has no criterion.**
   AC-GHK-001.2 requires, at line 130: "`And the CLI attributes its
   features to the document their owning-document tag names`". FEAT-GHK-002
   requires the tag's presence (AC-GHK-002.10) and forbids it
   contradicting its containing document (AC-GHK-002.11), but no
   criterion says what happens when the slug it carries names no corpus
   member. Attribution is impossible in that case, and the blueprint
   invented a rule — a `gherkin/owner-unresolved` finding — to avoid
   emitting a feature attributed to a document that does not exist.

4. **The FRD specifies a warning it would raise against every one of
   its own criteria.** AC-GHK-006.6 declares, at line 577:
   `| a criterion carrying no verification level    | warning  |`,
   while `specs/standards/gherkin-guide.md:132` records
   the same tag as
   `| `@level-<unit\|integration\|e2e>` | on a `Scenario` | optional |`.
   Not one criterion in [[frd-gherkin-adoption]] carries a `@level-*`
   tag — the only two occurrences of the string in the document are the
   out-of-scope note at line 41 and the vocabulary table at line 210.
   The document that defines the check is the document the check fires
   55 times against, and the guide calls the tag optional.

5. **A note asserts an obligation on another FRD that the design does
   not discharge, and nothing tracks it.** At line 164: "`*Note:*
   AC-GHK-001.2 requires a change to [[frd-spec-validation-cli]]`",
   continuing that what AC-SPEC-001.6 "needs is a second discovery pass
   for feature sources, not an exception to corpus membership". The
   blueprint puts that discovery pass in this feature's own DISC delta
   and changes [[frd-spec-validation-cli]] in no way. The note and the
   design disagree, and the blueprint author resolved the disagreement
   silently in the design's favour.

## Proposals

### P1 — Decide what an owning-document tag that resolves to nothing does

**Problem.** No text exists where this rule should live. FEAT-GHK-002
covers the tag's presence, its uniqueness and its agreement with the
containing document, and stops there; AC-GHK-001.2 assumes the tag
names something.

**Change.** Add a criterion to FEAT-GHK-002:

> `@AC-GHK-002.12 @negative`
> `Scenario: An owning-document tag naming no corpus member is reported`
> `Given a Feature whose owning-document tag names a slug no corpus member declares`
> `When the CLI validates it`
> `Then the CLI reports a finding naming the slug`
> `And the CLI records no feature attributed to that document`

**Evidence.** Finding 3. The rule stops being a blueprint invention
that retires with the blueprint, and the second clause pins the part
that matters most — a feature is not indexed against a document that
does not exist.

### P2 — Settle whether a verification level is optional or warned about

**Problem.** `gherkin-adoption-frd.md:577` makes an absent level a
`warning`; `gherkin-guide.md:132` calls the tag `optional`. Both are
current, and the FRD would raise one warning per criterion against
itself on the day the check ships.

**Change.** Pick one and make both documents say it. The recommended
resolution is to keep the warning and tag the FRD's own criteria,
because the level is what [[testing-standards]] and the blueprint's AC
Coverage table both consume — in which case `gherkin-guide.md`'s tag
table changes `optional` to `yes, on every criterion` and the FRD's own
criteria gain `@level-*` tags. The alternative — drop the row from
AC-GHK-006.6's severity table — is cheaper and loses the only machine
signal that a criterion was never assigned a verification level.

**Evidence.** Finding 4. Whichever way it resolves, the first corpus
the check runs against stops being a counterexample to it.

### P3 — Make AC-GHK-001.6 a behavior, or mark it manual

**Problem.** AC-GHK-001.6 (lines 150-154, quoted in finding 1) states
where a human puts a file. An acceptance criterion is a claim about the
system under test, and this one has no observable behavior behind it,
so the coverage table could name no delta that proves it and no test
can cite it.

**Change.** Prefer restating it as the discovery behavior it is
standing in for:

> `@AC-GHK-001.6`
> `Scenario: A feature source is discovered beside the documents of its feature`
> `Given a .feature file in the same directory as the FRD that owns it`
> `When the CLI discovers feature sources`
> `Then the CLI parses that file without being given its location`

WHERE the convention itself is worth stating, it belongs in
`gherkin-guide.md` rather than in a criterion. IF the criterion is kept
as written, THEN it SHALL carry `@manual`, so that the marker
`ac-traceability-standards.md` requires exists for a blueprint's
`manual` level to reconcile against.

**Evidence.** Finding 1. The AC Coverage table loses its one row that
needed a paragraph of justification, and the manual-level reconciliation
that `rfc-self-readiness-checklist-frontmatter-and-manual-acs` P2
proposes becomes checkable on this feature.

### P4 — Resolve the note that claims another FRD must change

**Problem.** Line 164 states "`AC-GHK-001.2 requires a change to
[[frd-spec-validation-cli]]`" and the blueprint changes that document
in no way. One of the two is wrong, and the FRD gives a reader no way
to tell which.

**Change.** Replace the note's first sentence with the conclusion its
own remaining sentences reach: feature-source discovery is wholly this
feature's, [[frd-spec-validation-cli]] is untouched, and AC-SPEC-001.6
stands because a `.feature` file was never a corpus candidate. IF a
change to that FRD is in fact intended, THEN say which criterion it
gains and file the bead.

**Evidence.** Finding 5. A reader of either FRD gets one answer to
"does the spec-index feature change?" instead of two.

## Considered and Not Proposed

- **Giving prose criterion extraction a requirement of its own.** This
  RFC's original P1 proposed a FEAT-GHK-010 specifying which markdown
  form is a prose criterion, that it has no name, how it is attributed
  to a feature, and that it carries a content hash. Review rejected the
  premise rather than the wording: a compatibility reader is
  transitional machinery for a transition of three documents, and
  translating them costs less than a second extractor that would outlive
  the transition in the model and in every consumer of it. [[adr-021]]
  constraint 9 now narrows extraction to Gherkin only, AC-GHK-009.3 is
  withdrawn, and `agent-plugins-a7u` carries the translation. Finding 2
  still stands and its lesson is the general one: an eight-word clause
  of a coexistence criterion implied a subsystem, and nothing caught it
  until a blueprint costed it.
- **Requiring FEAT-GHK-007 to name the command surface.** The FRD says
  "a query names a feature identifier" and never `spec feature`, which
  left the blueprint to invent the whole subcommand surface and its
  flags. That is correct FRD discipline — WHAT, not HOW — and the
  blueprint is where a command surface belongs. No change.
- **Renumbering FEAT-GHK-002's criteria into order.** They run .1, .2,
  .10, .11, .3 … because .10 and .11 were appended beside the criterion
  they qualify rather than at the end. It cost a moment when building
  the coverage table and nothing else; identifiers are addresses, not a
  sequence, and renumbering would break any citation already written.
  Note that `agent-plugins-133` asks for a criterion-number *sequence*
  check, which would flag this — that check's design should decide
  whether ordering is in scope or only gaps are.
- **The 13 compound `SHALL … and SHALL …` bullets the blueprint's
  readiness check flagged.** These were written in the house style
  `spec-index-blueprint.md` already carries six of, so either the
  practice or `document-readiness-checklist.md` check 5 is wrong. That
  is a question about the framework's guidance, not about this
  project's FRD; it is routed to `retro-skill` rather than proposed
  here.

## Non-Goals

- No change to the Gherkin notation, the tag vocabulary, or [[adr-021]]'s
  fourteen constraints. Every proposal here adds or repairs requirement
  text within the shape those already fix.
- No weakening of a downstream check to make authoring easier. P3 may
  resolve toward tagging every criterion rather than dropping a warning,
  and that is the direction preferred.
- No renumbering of existing criteria, so every citation already
  written stays valid.

## Adoption

P1 through P4 are independently adoptable and touch one document each
except P2, which changes [[frd-gherkin-adoption]] and
`specs/standards/gherkin-guide.md` together and must land
as one change.

All five are backward compatible in the sense that matters here: every
existing criterion identifier keeps its meaning, and no document
currently conforming stops conforming. P1 adds a criterion. P2,
resolved as recommended, adds a tag to each of the FRD's criteria — a
mechanical edit with no identifier churn — and changes one table row in
the guide. P3 rewrites one criterion's body while keeping its
identifier, so the one citation to it (this feature's blueprint) stays
valid and needs its coverage row's level changed from `manual` to
`unit`. P4 rewrites one prose note.

[[feature-gherkin-adoption]] is `ready-for-review` and not yet groomed,
so adopting P1 and P3 before approval costs one blueprint revision
rather than a re-groom: P1 adds a criterion the blueprint already
implements, and P3 changes one coverage row. Adopting them after
grooming costs the work orders derived from them as well.

The expected effect is that the next feature blueprint authored from
this FRD makes no design decision the FRD leaves open, and that the
first corpus the new checks run against conforms to them.

## Resolution
