---
title: "RFC: Implementing Gherkin adoption — a host component told to
  read a vocabulary it may not know, a severity table nothing implements,
  and a warm-index line the deltas never drew"
description: Name the seam where identity enters a host component, give
  every condition an acceptance criterion enumerates a rule identifier
  and a severity, state which side of the warm-index line each new record
  and check sits on, name the registration behind every "registered"
  thing, and state fold obligations at the grain the deltas carry them
slug: rfc-0023-implementing-gherkin-adoption
type: rfc
status: ready-for-review
updated_at: 2026-08-15T00:00:00Z
origin: self
skill: bead-bonanza
skill_version: "1.5"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "0.18.0"
---

# RFC: Implementing Gherkin adoption

## Summary

[[feature-gherkin-adoption]] was implemented as twelve delta work orders
under epic `agent-plugins-wvd`, each by a separate agent holding no
context but its own item. Every one landed; none was bounced or blocked.
The blueprint was, by the measure that matters here, a good one — the
deltas were correctly ordered, the anchors mostly correct, and no agent
had to renegotiate a boundary.

What the twelve agents did have to do, repeatedly, was decide things the
blueprint stated as if they were settled. Three patterns account for
most of it: a host component was given a behavior line phrased in the
vocabulary the same blueprint forbids it to know; an acceptance
criterion's Examples table enumerated conditions no delta ever assigned
a rule to; and a feature that adds records to a cached index never said
which of them survive the cache. The seven proposals below close those
at their source. Five change the way a feature blueprint states a delta;
two amend [[frd-gherkin-adoption]] itself.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins |
| Audience | project — resolved in this repository, which is what filing under `rfcs/inbox/feature-blueprint/` records; the generated `Rfc` schema has no `audience` property |
| Phase executed | Implemented a groomed plan — twelve delta work orders plus the Definition of Done item, each in its own worktree |
| Documents exercised | [[feature-gherkin-adoption]]; [[frd-gherkin-adoption]]; [[module-document-core]], [[module-spec-profile]], [[container-spec-cli]]; [[adr-021]]; [[gherkin-guide]]; `ac-traceability-standards.md` v0.4.0 |
| Documents produced | The folds into the three architecture blueprints and [[contract-specs-configuration]]; the retirement of [[feature-gherkin-adoption]] |
| Skill(s) followed | bead-bonanza v1.5 (orchestration and per-bead dispatch); groom-beads v1.3 and groom-feature-blueprint v1.3 produced the work orders in the prior phase |
| Templates / standards read | feature-blueprint-template v1.4; `ac-traceability-standards.md` v0.4.0; `testing-standards` |
| Plugin version | 0.18.0 |
| Session date | 2026-08-15 |
| Evidence channel | `retro:` comments recorded in flight on all thirteen items, harvested at the Definition of Done item per the findings capture protocol |

## Evidence

Findings are numbered for citation by the proposals. Each names the item
that recorded it.

1. **A host component was told to read a vocabulary it is forbidden to
   know (GHK).** The GHK delta's behavior list says the component "SHALL
   emit one criterion record per `Examples` block **that carries an
   identifier tag**, and one criterion record for a `Scenario Outline`
   whose `Examples` blocks carry none". Eleven lines above, the same
   delta's Intent says the component "holds the whole of what is true of
   Gherkin and none of what is true of this framework: it reads tags as
   strings, attaches no meaning to any of them". What an identifier tag
   is, is exactly the meaning it may not attach. The implementer resolved
   it by having `Options` carry two caller-supplied functions resolving a
   set of tags to an identity — a seam the blueprint does not mention.

2. **An acceptance criterion was assigned to a delta that could not
   prove it under the obvious implementation (GHK).** The AC Coverage
   table names GHK as the proving delta for AC-GHK-003.2, a feature
   record carrying its identifier and code. Identity is vocabulary the
   host cannot derive. It was provable in GHK's own tests only because
   the implementer had already chosen the `Options` seam of finding 1;
   had they returned tags alone and left identity to DISC — the obvious
   reading of the boundary — the criterion would have been unprovable in
   that item, with nothing in the item saying so.

3. **`FeatureCode` is required on every record and derivable from
   nothing (DISC).** The MODEL delta declares `FeatureCode` and the
   blueprint requires it on every feature record. Neither the XPT delta
   nor the registration it widens says how one is derived:
   `FeatureTagRules` registers `SharedCaptures` as an unordered list, so
   no registered value identifies which capture is the code. The
   implementer derived it structurally from the matched identifier — the
   segment before the number — and recorded the consequence: a profile
   whose identifier pattern has a different shape silently gets an empty
   code.

4. **An acceptance criterion's Examples table enumerates two conditions
   no rule implements (VTAG).** AC-GHK-006.6 is a `Scenario Outline`
   whose table has five rows. Two of them — "a Feature description that
   is absent" (warning) and "a criterion carrying no verification level"
   (warning) — correspond to no rule identifier anywhere in the
   blueprint's VTAG or VCON behavior lists, and to none of the thirteen
   `gherkin/` rules that shipped (`cli/core/finding/finding.go` lines
   116–153). The VCON delta nonetheless restates all five conditions
   verbatim and cites AC-GHK-006.6, and the AC Coverage table names VCON
   as its proving delta. AC-GHK-006.6 is satisfied for three of its five
   rows.

5. **Severity was unspecified for eight of the thirteen rules (VTAG,
   VCON).** AC-GHK-006.6's table names severities for three implemented
   conditions. The six VTAG rules `identifier-malformed`,
   `identifier-multiple`, `criterion-code-mismatch`, `owner-missing`,
   `owner-mismatch` and `owner-unresolved`, and the two VCON rules
   `scenario-prose` and `examples-partially-tagged`, are named by no
   severity in the FRD or the blueprint. Two different agents chose
   independently — `error` for the six, then `warning` for
   `scenario-prose` and `error` for `examples-partially-tagged`.

6. **The blueprint never drew the warm-index line, and four items hit
   it.** (a) DISC: "A source that fails to parse SHALL be recorded as a
   failure against its own file" had nowhere to land — a `.feature`
   source produces no `IndexEntry` and `gherkin/parse` is VCON's — so
   `Corpus.FeatureFailures()` was added as a seam, cold-build-only, with
   the note that "the blueprint should have said which side of the
   warm-cache line these sit on". (b) VTAG: `specmodel.Feature` and
   `.Criterion` retain the resolved identity and never the raw tags
   several identity rules need, so all nine recompute Gherkin from source
   on every `Check` **and every `CheckFor`** — walking the whole corpus
   for a single-document check, wider than the AC-SPEC-009.2
   single-document budget the rest of Document Core holds itself to.
   (c) VCON: `gherkin/parse` was invisible to any walk over features,
   because a source that fails to parse produces no feature at all, and
   had to read the recompute's second return. (d) CHG: the delta's
   Anchors name "`cli/core/corpus` (`Open`, `Build`) for the two sides of
   the comparison", but `Open`'s warm path rebuilds silently on a hash
   mismatch and returns what it just built, discarding the earlier state
   `--changed` exists to report — precisely the case that matters.
   `corpus.Load` was used instead.

7. **The record model did not carry what a declared check needs (VTAG).**
   `gherkin/examples-partially-tagged` must know which `Examples` blocks
   belong to one divided `Scenario Outline`. No persisted or in-memory
   structure retained it. `gherkin.Feature.PartialOutlines` was added
   from the VTAG item — a touch outside that item's named anchors
   (`cli/core/corpus`) — because without it VTAG's own
   `criterion-untagged` would double-report and VCON could not implement
   its rule at all.

8. **Three deltas refer to "a registered tag" that no delta registers
   (DISC).** MDS: "every fenced block in a document body whose info
   string matches **a registered tag**". GREF: "a fenced block whose info
   string names **a registered Gherkin tag**". `FeatureTagRules`, the
   registration XPT widens, has no field for it. The implementer put it
   in `cli/core/gherkin` as `FenceTag`/`FenceTags()` on the grounds that
   it names a notation rather than a vocabulary, and recorded that if
   GREF or EMIT expected it in the profile or in `cli/core/markdown`,
   they would have to move it.

9. **An anchor claimed work by a delta whose own behavior list did not
   carry it (XPT).** XPT's Anchors end: "the generated
   `specmodel.ProfileRegistration` that MODEL extends". MODEL's Data
   section enumerates `feature.strata`, `index.strata` and
   `finding.strata`, and MODEL did not touch `schemas/profile.strata`.
   Extending `ProfileRegistration` with `feature_tags` therefore fell to
   XPT, pulling a schema regeneration and four testdata `input_hash`
   updates into an item scoped as a host-interface widening.

10. **Anchor lists omitted mandatory consequential edits (MODEL).**
    MODEL's Anchors name three `.strata` files and
    `scripts/generate-schemas.sh`. Two further edits were mandatory:
    `cli/core/specindex.New` had to initialise `Features` so the index
    emits `[]` rather than `null` (the module's own lockstep rule), and
    the four `index-*` snapshot manifests had to be regenerated because
    `Fingerprint` folds the profile's JSON Schema into `input_hash`, so
    growing the generated bundle changes every recorded hash.

11. **Two fold obligations named in the Definition of Done appeared in
    no delta's Fold line, and neither landed.** The blueprint's DoD
    requires [[container-spec-cli]]'s "Key Decisions' [[adr-021]] line no
    longer says no component parses acceptance criteria". No delta's Fold
    line carries it — QRY's is "Command Surface records feature and
    criterion and their flags; Change Guidance's verification gates cover
    them", EMIT's and CHG's name only their flags. At retirement that
    line still read: "acceptance-criteria extraction is a follow-on
    feature, so no component here parses acceptance criteria yet."
    Separately, [[module-document-core]]'s Components prose still read
    "Nothing in Corpus calls `cli/core/markdown` or `cli/core/gherkin`
    yet" — a sentence DISC made false and which no Fold line named. Both
    were repaired by the DoD item.

12. **An acceptance criterion's scope over the forms it governs was
    unstated, twice.** (a) AC-GHK-007.4 reads "When a query names an
    identifier no record declares / Then the CLI reports that nothing
    declares it / And the CLI exits non-zero". The QRY delta adds five
    query forms: a positional `<id>`, `--code`, `--of`, `--tag`, and no
    argument at all. The implementer read the exit-non-zero rule as
    governing the positional lookup only and treated a filter matching
    nothing as a valid empty selection, recording that "a future reader
    could reasonably invert it". (b) AC-GHK-002.3's five Examples rows
    are all criterion tags (`@AC-CHK-003.1`, `@AC-chk-003.1`,
    `@AC-CHK-3.1`, `@AC-CHK-003`, `@AC-SPEC-001.12`). The VTAG delta
    generalizes it to "an identifier tag not matching its registered
    pattern", covering feature identifiers too — and the implementer
    found the generalization undeliverable: distinguishing a malformed
    identifier from an unrelated project tag (which AC-GHK-002.8 requires
    be ignored) needs the pattern's literal prefix, and the registered
    `FeaturePattern` `^[A-Z]+-[A-Z0-9]+-[0-9]{3}$` has none. Malformed
    feature identifiers are undetectable under this framework's own
    registration; malformed criterion identifiers are detectable because
    `CriterionPattern` begins with the literal `AC-`.

13. *(Context, not a proposal.)* This repository's own [[gherkin-guide]]
    failed the new checks on first run: three illustrative snippets — a
    standalone-file owner-tag example, a partial-tagging excerpt and a
    `Rule` excerpt — were extracted as real features and reported
    `owner-mismatch` and `feature-untagged`. They were relabelled as
    `text` fences, matching the document's own existing convention for
    non-checkable illustrations. The guide is now self-consistent and no
    document change is proposed.

## Proposals

### P1 — Where a host component needs vocabulary, the delta names the seam

**Problem.** GHK's Intent and GHK's behavior list contradict each other
in eleven lines (finding 1), and the AC Coverage table then assigns GHK a
criterion about identity (finding 2). The same shape recurs in DISC,
where `FeatureCode` is required of every record and derivable from no
registered value (finding 3). In all three the blueprint wrote the
requirement in the reader's vocabulary because that is the only way to
say it briefly, and left the implementer to invent the mechanism by which
a vocabulary-free component obtains it.

**Change.** Add to the feature-blueprint conventions: WHERE a delta in a
host component states a behavior in terms the component may not itself
know, the delta SHALL name the seam through which that knowledge arrives,
in one line beside the behavior. For GHK that is:

> Identity arrives through the caller: `Parse` takes a function
> resolving a set of tags to an identifier, and this component never
> inspects a tag's spelling. "An `Examples` block that carries an
> identifier tag" means one for which that function returns an
> identifier.

And a companion rule for the coverage table: an acceptance criterion
SHALL NOT name as its proving delta a component that cannot satisfy it
without such a seam, unless the seam is stated.

**Evidence.** Findings 1, 2, 3. The three items that hit this each
invented a different mechanism; naming the seam once makes the boundary
argument reviewable at blueprint time, which is when it is cheap.

### P2 — Every condition an acceptance criterion enumerates gets a rule identifier and a severity

**Problem.** AC-GHK-006.6's table has five rows and three of them
correspond to a shipped rule (finding 4). The blueprint's VCON delta
restates all five and cites the criterion, and the AC Coverage table
records VCON as proving it — so nothing in the reconciliation surfaced
the two missing rows. Separately, eight of the thirteen rules that did
ship carried no stated severity, and two agents chose independently
(finding 5).

**Change.** Two changes, one to the FRD and one to the blueprint
conventions.

Amend [[frd-gherkin-adoption]]: either add the two rules AC-GHK-006.6's
table promises — a `gherkin/feature-description-missing` at `warning` and
a `gherkin/criterion-level-missing` at `warning` — or remove those two
rows and let the criterion describe what the CLI does. The rows are worth
keeping: both conditions are real authoring defects the guide already
warns about.

Add to the feature-blueprint conventions: WHERE a delta introduces
findings, its behavior list SHALL state a rule identifier **and a
severity** for each, and every row of any severity or condition table the
delta cites SHALL map to one of those rule identifiers. A row that maps
to nothing is either a delta the blueprint is missing or a criterion the
FRD should not carry.

**Evidence.** Findings 4, 5. A one-to-one map between an AC's rows and a
delta's rule identifiers is mechanically checkable at review time, and it
is the only thing that would have caught AC-GHK-006.6 before the DoD.

### P3 — A delta that adds a record or a check states which side of the warm-index line it sits on

**Problem.** This feature adds records to a cached, fingerprinted index
and adds checks that read them, and the blueprint says nothing about the
cache anywhere except to require that `Fingerprint` fold feature sources
in. Four items hit the omission from four directions (finding 6), plus a
fifth where the record model turned out not to retain what a check needs
(finding 7). The costliest is VTAG: nine rules recompute the whole corpus
on every single-document `CheckFor`, exceeding a budget another FRD
states, and the blueprint gave no signal that the record shape and the
check inputs needed reconciling.

**Change.** Add a required line to any delta that adds a record to the
index or a check that reads one:

> **Warm path:** <what a warm open serves for this delta, what must be
> recomputed, and against which latency budget>.

For MODEL that would have read: the persisted record retains the resolved
identity only, so any check needing the raw tags recomputes from source.
For DISC: parse failures are not persisted and are recomputed. For VTAG:
nine rules recompute; the single-document budget of AC-SPEC-009.2 does
not hold for them, and that is accepted. For CHG: the comparison needs
the *unrebuilt* persisted index, which is `corpus.Load`, not
`corpus.Open`.

**Evidence.** Finding 6 (four items), finding 7. Every one of these was
discovered by an implementer at the point of writing code, and each was
resolved correctly but independently; the last of them contradicted its
own item's coordination note.

### P4 — The Definition of Done states fold obligations at the grain the deltas carry them

**Problem.** The blueprint's DoD enumerates fold obligations per
document, at a finer grain than any delta's Fold line — and two
obligations it names belong to no delta at all (finding 11). Both
survived twelve items and were caught only because the DoD item read the
documents rather than the checklist. A third class is worse: the stale
Components sentence in [[module-document-core]] was a consequence of DISC
that the DoD did not name either, so nothing but a full read would have
found it.

**Change.** Require that every clause of the Definition of Done's fold
list appear in exactly one delta's Fold line, and add a residual delta
where a clause belongs to none:

> Each fold obligation below is discharged by the delta named beside it.
> An obligation naming no delta is discharged by the Definition of Done
> item.

And require a delta whose change makes a statement in an owning blueprint
false to say so in its Fold line, not only what the blueprint gains —
DISC's would have read "…and the Components prose no longer says nothing
in Corpus calls `cli/core/markdown` or `cli/core/gherkin`".

**Evidence.** Finding 11. Two of the three DoD documents needed repair at
retirement, both for lines the DoD explicitly listed.

### P5 — Every "registered" thing a delta refers to names the registration that holds it

**Problem.** MDS and GREF both condition their behavior on a fenced
block's info string matching "a registered tag", and no delta registers
one (finding 8). The implementer placed it in the component that seemed
most apt and recorded that two other deltas might have expected it
elsewhere.

**Change.** Add to the feature-blueprint conventions: WHERE a delta's
behavior depends on a registered value, the delta SHALL name the
registration that holds it — the interface member, the profile, or the
delta that adds it — and where no registration holds it yet, adding it is
part of some delta's behavior list.

**Evidence.** Finding 8. Two deltas depended on a registration a third
had to invent; naming it costs one clause and removes the guess.

### P6 — An anchor list is complete, and a cross-delta anchor claim is backed by that delta's own behavior list

**Problem.** XPT's Anchors assert that MODEL extends
`specmodel.ProfileRegistration`; MODEL's Data section does not, and MODEL
did not (finding 9). MODEL's own Anchors omitted two mandatory edits, one
of them an index-emission rule the module blueprint states and the other
a snapshot regeneration forced by how `Fingerprint` is composed (finding
10).

**Change.** Two clauses. First: an Anchors entry that attributes work to
another delta SHALL correspond to a line in that delta's own behavior or
data list; grooming can check this mechanically. Second: Anchors SHALL
name the files a delta's stated behavior forces to change, not only the
files it is about — including the generated artifacts and committed
fixtures a regeneration invalidates.

**Evidence.** Findings 9, 10. Finding 10's second half is durable
knowledge about this repository — a schema change invalidates every
`index-*` snapshot manifest — and
`rfcs/inbox/feature-blueprint/rfc-0019-implementing-markdown-links.md`
already proposes recording it; this is corroboration from a second
feature, not a competing proposal.

### P7 — An acceptance criterion states its scope across the forms it governs

**Problem.** AC-GHK-007.4 governs "a query", and QRY adds five query
forms with materially different failure semantics; the implementer chose
and recorded that a future reader could reasonably invert the choice
(finding 12a). AC-GHK-002.3's Examples rows are all criterion tags, and
the blueprint generalized the criterion to identifiers of both kinds — a
generalization that cannot be delivered under this framework's own
registered `FeaturePattern`, because the mechanism needs a literal prefix
the pattern has not got (finding 12b).

**Change.** Amend [[frd-gherkin-adoption]]:

- AC-GHK-007.4 SHALL name the forms it governs — "a query that names an
  identifier" — and a sibling criterion SHALL state what a filter
  matching nothing does, since that is the case the current wording
  leaves open.
- AC-GHK-002.3 SHALL either add feature-identifier rows to its Examples
  table, or a sibling criterion SHALL state that malformedness is
  detectable only for identifier patterns carrying a literal prefix, so
  the limitation is a specified property rather than an implementation
  surprise.

**Evidence.** Finding 12. Both were recorded as silent decisions by their
implementers, which is precisely the class of decision an FRD exists to
remove.

## Considered and Not Proposed

- **AC-GHK-001.6 arriving as a `manual` acceptance criterion.** The
  coverage table declared it `manual`; DISC shipped it with a unit test
  (`TestAFeatureSourceIsFoundWhereverBeneathTheSpecRootItSits`) and the
  level was corrected to `unit` at retirement. This is the defect
  `rfcs/inbox/frd/rfc-0021-gherkin-frd-unverifiable-and-implied-work.md`
  P3 already proposes fixing, now with implementation evidence that the
  criterion was automatable all along. Corroboration, not a second
  proposal.
- **Two acceptance criteria cited in prose rather than behind the `AC:`
  marker.** AC-GHK-002.8 and AC-GHK-003.5 each had a test whose comment
  opened `// AC-GHK-002.8: …` rather than carrying a citation line;
  both were repaired at retirement. Both are negative criteria, whose
  tests are named `TestNoFindingFor…` and whose comments naturally read
  as explanations — but `ac-traceability-standards.md` v0.4.0 is
  unambiguous about the marker, and one agent writing two of them is
  carelessness the existing guidance already warns against, not a gap.
- **`FindingDetail` variants reused across rules.** Five Gherkin variants
  serve thirteen rules; `gherkin/criterion-empty` borrowed
  `FindingMissingProperty` and `gherkin/scenario-prose` borrowed
  `FindingIdentifierPattern`. Both implementers recorded this as a
  deliberate choice consistent with the package's stated design — the
  detail union is keyed by the fix, not by the rule. No defect.
- **The cucumber parser exporting no error type.** Extracting a position
  from a parse failure needs reflection on an unexported slice type plus
  a regexp on each message; `errors.As` and `errors.Unwrap` are useless
  against it. Durable and non-obvious, but it is a property of a
  third-party library rather than of any document here, and it is now
  recorded in [[module-document-core]] where the next reader of that
  package will find it.
- **The worktree path asymmetry two items hit.** `Read` and `Grep`
  against the shared-checkout path succeed silently while `Edit`, `Write`
  and `Bash` refuse it, so a wrong-root read is discovered only at the
  first mutation. Recorded by PREG and EMIT independently. This is run
  mechanics, not a project document — it belongs to the sprint retro and
  to `retro-skill`, and is left for them.

## Non-Goals

- **No change to the deltas as shipped.** Every finding here was
  resolved correctly in the implementation and folded into the owning
  architecture blueprint. The proposals prevent the next blueprint from
  requiring the same twelve judgement calls; they do not reopen this one.
- **No stricter readiness gate on feature blueprints.** This blueprint
  passed every readiness and content check, and would pass them again
  with these proposals adopted. P1 through P6 add information a delta
  carries, not a check it must survive — with the single exception of
  P2's row-to-rule map, which is checkable precisely because it is
  mechanical.
- **No proposal about the FRD's Gherkin notation itself.** The notation
  worked. Fifty-five criteria were extracted, tagged and queried by the
  tool the feature built, and this document's evidence about
  AC-GHK-006.6 and AC-GHK-002.3 is legible only because the criteria are
  machine-readable.
- **No revision of [[adr-021]].** All fourteen of its constraints held
  under implementation; nothing measured here contradicts it.

## Adoption

P1 through P6 are independently adoptable and change the feature-blueprint
conventions only; adopting any one of them requires no change to a
shipped document. P7 amends [[frd-gherkin-adoption]] and is independent
of the rest.

P2 has two halves with different owners: the blueprint-convention half
stands alone, and the FRD half — adding two rules or removing two rows —
is a change to an approved FRD and to the CLI if the rules are added.
Adopting the convention half without the FRD half leaves AC-GHK-006.6
partially satisfied but makes the next such gap visible at review.

Sequencing: P2's convention half and P4 both concern reconciliation at
the Definition of Done and are best adopted together, since P4's
per-delta attribution is what makes P2's row-to-rule map checkable
without reading every document.

**Backward compatibility.** P1, P3, P4, P5 and P6 add clauses a delta
carries; every existing blueprint stays valid, and the four retired ones
([[feature-spec-validation-cli]], [[feature-spec-reference-integrity]],
[[feature-spec-markdown-links]], [[feature-gherkin-adoption]]) are not
revisited — a retired blueprint's job is done. The next blueprint to be
authored, from [[frd-spec-structure-conformance]], acquires the clauses
as it is written. P2's convention half is the only one that can
fail an existing document: a blueprint citing a condition table with an
unmapped row would newly be incomplete, and the migration is to map the
row or amend the criterion — which is the proposal's point. P7 amends
one FRD in place; adding rows to an Examples table is append-only and
breaks no citation, and adding a sibling criterion takes the next unused
number per `ac-traceability-standards.md`, so no existing `AC:` citation
retargets.

## Resolution

_Awaiting review._
