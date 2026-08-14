---
title: "RFC: The Gherkin adoption FRD carries one criterion nothing can
  verify, one clause that implies a subsystem, and a warning it would
  raise against all 56 of its own criteria"
description: Make AC-GHK-001.6 a behavior or mark it manual, give prose
  criterion extraction a requirement of its own, decide what an
  owning-document tag that resolves to nothing does, and settle whether
  the FRD warns about a tag its own guide calls optional
slug: rfc-0019-gherkin-frd-unverifiable-and-implied-work
type: rfc
status: ready-for-review
updated_at: 2026-08-14T00:00:00Z
audience: project
origin: agent-plugins
plugin_version: "0.17.0"
---

# RFC: Unverifiable and implied work in the Gherkin adoption FRD

## Summary

A feature blueprint was authored from [[frd-gherkin-adoption]] — the
first document in this corpus to write its own acceptance criteria in
Gherkin, and by some distance the most thorough FRD here: 56 criteria,
nine requirements, a rationale under four of them. It still left the
blueprint deciding five things no criterion states, and one of those
decisions — how a prose criterion is recognised and attributed — is a
whole delta of work implied by the second clause of a single criterion.
The proposals below close the gaps at their source rather than in the
blueprint, where they die when the blueprint retires.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins |
| Phase executed | Authored a feature blueprint from an approved FRD |
| Documents exercised | [[frd-gherkin-adoption]]; [[adr-020]]; [[container-spec-cli]], [[module-document-core]], [[module-spec-profile]]; [[contract-specs-configuration]]; [[testing-standards]]; `plugin/skills/spec/references/gherkin-guide.md`; `schemas/feature.strata`, `schemas/index.strata` |
| Documents produced | [[feature-gherkin-adoption]] |
| Skill(s) followed | spec v1.9; groom-feature-blueprint (Phase 1 gate only) |
| Templates / standards read | feature-blueprint-template v1.4; `blueprint-readiness-checklist.md` v0.6.0; `document-readiness-checklist.md` v0.4; `ac-traceability-standards.md` v0.2.0 |
| Plugin version | 0.17.0 |
| Session date | 2026-08-14 |

## Evidence

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
   built). None is inferable from the clause.

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
   its own 56 criteria.** AC-GHK-006.6 declares, at line 577:
   `| a criterion carrying no verification level    | warning  |`,
   while `plugin/skills/spec/references/gherkin-guide.md:132` records
   the same tag as
   `| `@level-<unit\|integration\|e2e>` | on a `Scenario` | optional |`.
   Not one criterion in [[frd-gherkin-adoption]] carries a `@level-*`
   tag — the only two occurrences of the string in the document are the
   out-of-scope note at line 41 and the vocabulary table at line 210.
   The document that defines the check is the document the check fires
   56 times against, and the guide calls the tag optional.

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

### P1 — Give prose criterion extraction a requirement of its own

**Problem.** AC-GHK-009.3's second clause ("`Then the CLI records
criteria from both notations`", line 725) is the only requirement text
covering a capability that reads a second notation out of markdown. A
criterion of a coexistence requirement is not where a parser is
specified, and the three decisions listed in finding 2 had to be made
by the blueprint's author with nothing to constrain them.

**Change.** Add FEAT-GHK-010, *Prose Criterion Extraction*, carrying
criteria for: which markdown form the CLI recognises as a prose
criterion; that a prose criterion has no name; that it is attributed to
the feature its identifier's code and number name; and that it carries
a content hash on the same terms as a Gherkin criterion so change
detection spans both notations. Leave AC-GHK-009.3 as the coexistence
statement it is — no finding merely because a document uses prose — and
cross-reference the new requirement from it.

**Evidence.** Finding 2. The next author derives the PROSE delta from
requirement text instead of inventing it, and the three decisions
become reviewable before the blueprint rather than inside it.

### P2 — Decide what an owning-document tag that resolves to nothing does

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

### P3 — Settle whether a verification level is optional or warned about

**Problem.** `gherkin-adoption-frd.md:577` makes an absent level a
`warning`; `gherkin-guide.md:132` calls the tag `optional`. Both are
current, and the FRD would raise 56 warnings against itself on the day
the check ships.

**Change.** Pick one and make both documents say it. The recommended
resolution is to keep the warning and tag the FRD's own criteria,
because the level is what [[testing-standards]] and the blueprint's AC
Coverage table both consume — in which case `gherkin-guide.md`'s tag
table changes `optional` to `yes, on every criterion` and the FRD's 56
criteria gain `@level-*` tags. The alternative — drop the row from
AC-GHK-006.6's severity table — is cheaper and loses the only machine
signal that a criterion was never assigned a verification level.

**Evidence.** Finding 4. Whichever way it resolves, the first corpus
the check runs against stops being a counterexample to it.

### P4 — Make AC-GHK-001.6 a behavior, or mark it manual

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

### P5 — Resolve the note that claims another FRD must change

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

- No change to the Gherkin notation, the tag vocabulary, or [[adr-020]]'s
  fourteen constraints. Every proposal here adds or repairs requirement
  text within the shape those already fix.
- No weakening of a downstream check to make authoring easier. P3 may
  resolve toward tagging 56 criteria rather than dropping a warning,
  and that is the direction preferred.
- No renumbering of existing criteria, so every citation already
  written stays valid.

## Adoption

P1 through P5 are independently adoptable and touch one document each
except P3, which changes [[frd-gherkin-adoption]] and
`plugin/skills/spec/references/gherkin-guide.md` together and must land
as one change.

All five are backward compatible in the sense that matters here: every
existing criterion identifier keeps its meaning, and no document
currently conforming stops conforming. P1 adds a requirement rather
than altering one. P2 adds a criterion. P3, resolved as recommended,
adds a tag to 56 criteria — a mechanical edit with no identifier
churn — and changes one table row in the guide. P4 rewrites one
criterion's body while keeping its identifier, so the one citation to
it (this feature's blueprint) stays valid and needs its coverage row's
level changed from `manual` to `unit`. P5 rewrites one prose note.

[[feature-gherkin-adoption]] is `ready-for-review` and not yet groomed,
so adopting P1, P2 and P4 before approval costs one blueprint revision
rather than a re-groom: P1 adds a delta whose design the blueprint
already carries, P2 adds a criterion the blueprint already implements,
and P4 changes one coverage row. Adopting them after grooming costs the
work orders derived from those deltas as well.

The expected effect is that the next feature blueprint authored from
this FRD makes no design decision the FRD leaves open, and that the
first corpus the new checks run against conforms to them.

## Resolution
