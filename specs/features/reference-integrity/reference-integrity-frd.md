---
title: "FRD: Spec Corpus Reference Integrity"
description: Parses cross-document references from a spec corpus, records
  them in the shared index, and reports links that resolve to nothing, so
  a reference that silently broke is caught before a reader follows it
slug: frd-spec-reference-integrity
type: frd
status: approved
updated_at: 2026-10-08T18:00:00Z
sources:
  - "frd-spec-index"
---

# Spec Corpus Reference Integrity

## Overview

Specification documents cross-reference one another — `[[slug]]` links
in prose, and frontmatter properties that name related documents. When a
target is renamed, retired, or never existed, the link silently stops
resolving and a reader follows it into nothing. A prototype over this
repository's own corpus found 26 such broken references that no reviewer
had noticed.

Those 26 are gone: the slug migration [[adr-012]] decided repaired them,
and a scan of the same corpus on 2026-08-06 — 29 documents, 256 body
references once backticked illustrations are excluded — finds none
unresolved. That is the argument for this
feature rather than against it. The 26 were repaired by a one-off
prototype and a hand migration, and nothing now stands between the
corpus and the next such drift; a corpus that is clean exactly once, at
the moment someone looked, is the state this feature exists to end. The
same scan also shows how little of the relationship graph has ever been
examined: no anchor has ever been resolved against a heading, no retired
document's successor has been resolved, and no paired relationship has
been checked from both ends. Those are not passing checks. They are
checks that have never run.

This feature adds reference awareness to the tool established by
[[frd-spec-index]]. It reads each document's body and
reference-bearing frontmatter, records the outbound and inbound
references in the shared index, and reports every reference that
resolves to no corpus member. It does not introduce a new artifact or a
new engine: it populates the reference properties of the index defined
in FEAT-SPEC-002 and emits findings through the engine defined in
FEAT-SPEC-006. Corpus discovery, membership, output location, and the
warm index are inherited unchanged from that FRD and are not restated
here.

That FRD's ordering governs this one. The index records what a document
references, including references that resolve to nothing, and never
fails; reporting an unresolved reference is the checker's job. A
repository whose links are broken still contributes its graph to an
organization's catalogue.

**Out of scope for this feature.** Each of these is deferred
deliberately, not overlooked:

- **Anchor resolution.** A reference may name a heading anchor —
  `[[container-spec-cli#Element Registry]]`, the form [[adr-012]]
  mandates for un-promoted components. This feature resolves the slug
  half and does not interpret the anchor (AC-SPECREF-002.2).
  Interpreting it is specified in [[frd-spec-anchor-references]], a
  sub-feature of this one. The split is deliberate: anchor resolution
  needs heading data this feature does not have, from
  [[frd-spec-structure]], and every anchor reference in the
  corpus resolves on its slug half — so shipping slug resolution first
  costs the corpus nothing and unblocks it sooner.
- **Heading extraction and section spans.** Specified in
  [[frd-spec-structure]]. Nothing in this feature reads a
  document's headings.
- **Suggesting a correction for a broken reference.** Findings state
  the fact and leave the fix to a reviewing agent.
- **Repairing or rewriting references.** This feature only reads and
  reports, as FEAT-SPEC-002's ordering requires. Rewriting every
  reference to a slug when that slug is deliberately renamed is
  specified in [[frd-spec-slug-rename]], a sub-feature of this one. The
  boundary holds in both documents: a check never edits what it merely
  infers is wrong, and a rename edits only what an author explicitly
  instructed.
- **References to documents outside the specification set** — the
  backticked relative paths that
  `plugin/skills/spec/references/cross-document-references.md` governs.
  A path that resolves to no file is a broken reference by the same
  argument as a `[[slug]]` that resolves to no document, but it is a
  different resolution domain (paths relative to the containing file,
  not slugs in a flat namespace) and is not specified here.
- **The shape of the reference data in the index.** Whether references
  are embedded in each document's entry or collected separately as
  edges is a schema decision recorded in the index-schema ADR, not
  here. This FRD states what must be recorded and answerable, not how
  the index models it.
- **Resolving a reference across workspaces.** AC-SPEC-002.3 gives
  every entry a workspace identifier so indexes from separate
  repositories merge into one graph, but this feature resolves only
  within the workspace it is run against. A reference to a slug that
  exists only in another repository's index is reported as unresolved.
  The merged-graph case is a known follow-on, not a defect for the
  first team that tries it to discover.
- **Reporting an unreferenced document.** Inbound edges make an orphan
  free to detect, but whether a document *should* be reachable is a
  judgment about the corpus, not a conformance rule — a draft is
  expected to be uncited, and a warning on every new document is a
  warning nobody reads. AC-SPECREF-003.5 answers the question on
  demand; no check reports it.
- **Rendering the reference graph**, and export to any third-party
  format.

## Terminology

- **Body** — the content of a document after its frontmatter block.
  Reference extraction reads the body; frontmatter contributes
  references only through the properties the profile registers as
  reference-bearing, never through prose in a frontmatter value.
- **Outbound reference** — a `[[slug]]` link parsed from a document's
  body, or a slug value in a frontmatter property the active profile
  registers as reference-bearing.
- **Anchor** — the portion of a reference after a `#`, naming a heading
  within the target document. This feature reads a reference's slug
  half; [[frd-spec-anchor-references]] specifies the anchor half.
- **Inbound reference** — the reverse edge: for a given document, every
  other corpus member whose outbound references name it.
- **Resolution** — matching a reference's slug against the slugs corpus
  members declare, without regard to letter case, per AC-SPEC-003.4.
  A reference resolves against a target's frontmatter `slug`, never a
  file path.
- **Unresolved reference** — a reference whose slug matches no corpus
  member. A reference whose slug matches a file the index recorded as
  excluded is unresolved too, but is reported with that reason
  (AC-SPECREF-002.7), because the reader's next step differs. A reference
  whose slug matches a file the index recorded as a parse failure is not
  reported at all: the citing document is not what is wrong, and the
  parse failure is the one finding (AC-SPEC-003.8).
- **Reference-bearing property** — a frontmatter property the active
  profile declares as holding one or more document slugs (for example
  `superseded_by` or `modules`). A profile MAY additionally declare
  the document `type` such a property's target must carry, making the
  property a *typed* reference.
- **Paired relationship** — two reference-bearing properties the profile
  declares as counterparts, so that a reference in one document implies a
  matching reference back. An adopting profile that pairs `owner` with
  `owned_by` gets both halves checked. The specs profile
  registers no pairing: per [[document-relationship-standards]] a pairing
  is reserved for a relationship a reader must see without the index, and
  every relationship this framework declares has its reverse derived.

## Requirements

### FEAT-SPECREF-001: Reference Extraction in the Index

**Acceptance Criteria:**

```gherkin
@FEAT-SPECREF-001
Feature: Reference Extraction in the Index

  As an architecture owner
  I want every document's outbound and inbound references recorded in the index
  So that impact analysis and a central catalogue read the relationship graph from the same artifact instead of re-deriving it

  A reference written *about* rather than written is not a reference. The
  two places that happens are a backticked illustration and a frontmatter
  value describing the notation, and both are excluded.

  Indexing never fails. A document's outbound references are recorded
  whether or not they resolve, and reporting an unresolved one is the
  checker's job.

  @AC-SPECREF-001.1
  Scenario: Every wikilink in a document's body is recorded with its span
    When the CLI builds the index
    Then every corpus member's entry carries a references property holding every wikilink parsed from its body, each with the span at which it occurs

  @AC-SPECREF-001.2 @negative
  Scenario Outline: A wikilink inside a code region is not a reference
    Given a wikilink occurring inside "<region>"
    When the CLI builds the index
    Then the CLI does not record it as a reference

    Examples:
      | region                |
      | a fenced code block   |
      | an inline code span   |

  @AC-SPECREF-001.3
  Scenario: Inbound references are derived for every document
    When the CLI builds the index
    Then the CLI adds inbound references for every document, derived from the outbound references of all other corpus members

  @AC-SPECREF-001.4 @negative
  Scenario: A wikilink in a frontmatter value is not a reference
    Given a wikilink occurring in a frontmatter value
    When the CLI builds the index
    Then the CLI does not record it as a reference
    And frontmatter contributes references only through the properties the active profile registers as reference-bearing

  @AC-SPECREF-001.5
  Scenario: Repeated references are recorded separately and in document order
    Given a document referencing the same target more than once
    When the CLI builds the index
    Then the CLI records each occurrence with its own span
    And the CLI orders that document's outbound references by position in the document

  @AC-SPECREF-001.6
  Scenario: An unresolved reference is indexed rather than fatal
    Given a document whose outbound references include one that resolves to no corpus member
    When the CLI builds the index
    Then the CLI records that reference in the index
    And the CLI exits zero

  @AC-SPECREF-001.7
  Scenario: A reference-bearing frontmatter property contributes outbound references
    When the CLI builds the index
    Then the CLI records every value held by a frontmatter property the active profile registers as reference-bearing as an outbound reference of that document, each with the span at which it occurs, distinguishable from a reference parsed from the body

  @AC-SPECREF-001.8
  Scenario: A reference reaching a heading of its own document is an inbound reference of it
    Given a reference carrying an anchor and no slug
    When the CLI builds the index
    Then the CLI records it among the inbound references of the document holding it, where its anchor resolves against that document
    And the CLI records no inbound reference for it where its anchor does not
```

*Examples:* `adr-011`'s entry gains outbound references to `adr-004` and
`adr-006`, each carrying the byte offset and line/column of the link
(AC-SPECREF-001.1). The sentence "write `` `[[security-standards]]` ``
to cite a standard" contributes no reference, because the link is an
illustration (AC-SPECREF-001.2). `adr-004`'s entry lists `adr-005`,
`adr-006` and `adr-007` as inbound, without any of them declaring the
relationship in frontmatter (AC-SPECREF-001.3).
`specs/adrs/adr-012-reference-conventions.md` line 3 carries
`description: References use [[slug]] wikilinks …`, which names the
notation rather than a document (AC-SPECREF-001.4). An FRD citing
`` `[[adr-021]]` `` in its Requirements and again in its References
carries two entries, so a finding against either cites the line the
reader must open; ordering by position is what makes the index
byte-identical across platforms per AC-SPEC-009.7 (AC-SPECREF-001.5). A
corpus with eleven broken links still produces a complete index
(AC-SPECREF-001.6). A document's `related: ["adr-012", "adr-002"]`
records two outbound references naming the property they came from
(AC-SPECREF-001.7). `` `[[#Terminology]]` `` written in a document that
carries a `Terminology` heading is listed among that document's own
inbound references, so `specs refs` answers with the section link a reader
renaming that heading has to fix; `` `[[#Nowhere]]` `` written beside it
is listed among none, because it reaches no heading to break
(AC-SPECREF-001.8).

*Rationale:* AC-SPECREF-001.2 and AC-SPECREF-001.4 are the same rule
applied to the two places a reference is written about rather than
written: a backticked example, and a frontmatter value describing the
notation. Both were observed in this corpus, and both are false
positives a reader would have to dismiss by hand — the failure mode that
makes a checker get ignored. AC-SPECREF-001.6 restates FEAT-SPEC-002's
ordering for this feature specifically, because a reference graph is the
part of the index most likely to tempt an implementer into failing the
build from the indexer.

AC-SPECREF-001.7 is stated positively because AC-SPECREF-001.1 covers
only the body, and everything downstream assumes frontmatter references
are in the index: AC-SPECREF-002.3 checks them, AC-SPECREF-003.1 answers
for them, AC-SPECREF-002.5 compares a pairing across two documents, and
AC-SPECREN-001.4 rewrites them from their recorded spans. Left implicit,
the reasonable reading of FEAT-SPECREF-001 is that the index holds prose
links only — which passes every criterion in this requirement and
silently drops the half of the graph a catalogue reads first.

AC-SPECREF-001.8 says what AC-SPECREF-001.3's "all other corpus members"
leaves open, for the one reference whose target is the document holding
it. The anchor-only form AC-SPECANCH-002.6 resolves against the
containing document is a reference like any other — AC-SPECREF-002.12
already refuses to call it broken — so a query that omitted it would
answer "nothing references this heading" about a document that visibly
does. That answer is wrong rather than failing, and no check sees it,
which is why the criterion is stated here rather than left to the
queries FEAT-SPECREF-003 and FEAT-SPECANCH-003 define over this graph.
The conjunction is AC-SPECREF-002.12's: an empty slug alone earns
nothing, and an anchor reaching no heading records no edge, because a
reference that reaches nothing has no second end.

### FEAT-SPECREF-002: Reference Integrity Checks

**Acceptance Criteria:**

```gherkin
@FEAT-SPECREF-002
Feature: Reference Integrity Checks

  As a document author
  I want every cross-reference checked
  So that a link that silently stopped resolving is caught before a reader follows it into nothing

  Severity is assigned here rather than left to implementation, because
  it is what the exit status of a build depends on.

  A single-document check reports the findings located in that document,
  and of the findings located elsewhere only those of references that
  this document's own change left pointing at nothing. Unrelated breakage
  elsewhere stays out. Which documents an edit will break, asked before
  the edit, is a question for the refs query, not for the check.

  @AC-SPECREF-002.1
  Scenario: A body reference matching no corpus member is an error
    Given a document body containing a wikilink whose slug matches no corpus member's slug
    When the CLI checks the corpus
    Then the CLI reports a finding of severity error with the file, the span, and the unresolved slug

  @AC-SPECREF-002.2
  Scenario: A reference naming an anchor is resolved on its slug alone
    Given a reference naming a heading anchor
    When the CLI checks it
    Then the CLI resolves the reference on its slug portion alone

  @AC-SPECREF-002.3
  Scenario: A reference-bearing frontmatter value matching no corpus member is an error
    Given a frontmatter property the active profile registers as reference-bearing
    And its value matches no corpus member's slug
    When the CLI checks the document
    Then the CLI reports a finding of severity error naming the property and the value

  @AC-SPECREF-002.5
  Scenario: A pairing with no counterpart is an error against both documents
    Given the active profile registers a frontmatter property as one half of a paired relationship
    And the document it names does not declare the matching counterpart
    When the CLI checks the corpus
    Then the CLI reports a finding of severity error against both documents

  @AC-SPECREF-002.6
  Scenario: A live document citing a retired one is a warning naming the successor
    Given a document at a status the profile treats as live
    And it references a document at status retired
    When the CLI checks the corpus
    Then the CLI reports a finding of severity warning naming the retired target and its superseded_by value when it has one

  @AC-SPECREF-002.7
  Scenario Outline: A slug matching a file that is not a corpus member is told apart from one matching nothing
    Given a reference whose slug matches a file the index recorded as "<recorded>" rather than a corpus member
    When the CLI checks the corpus
    Then the CLI "<reports>"

    Examples:
      | recorded        | reports                                                                            |
      | excluded        | reports the finding with the exclusion's reason rather than as a slug matching nothing |
      | a parse failure | reports no finding against the citing document, the parse failure being the finding   |

  @AC-SPECREF-002.8
  Scenario: Resolution ignores letter case
    Then the CLI resolves a reference against corpus slugs without regard to letter case

  @AC-SPECREF-002.9 @negative
  Scenario: A single-document check reports no unrelated finding located in another document
    When the CLI checks a single document
    Then the CLI reports no finding located in a document other than the one named, except those AC-SPECREF-002.11 names
    And that holds for every finding of another document that does not arise from a reference to the document named

  @AC-SPECREF-002.10
  Scenario: A typed reference resolving to the wrong type is an error
    Given the active profile declares a required target type for a reference-bearing property
    And the resolved document declares a different type
    When the CLI checks the document
    Then the CLI reports a finding of severity error naming the property, the type expected and the type the target declares

  @AC-SPECREF-002.11
  Scenario: A single-document check reports its own reference findings and the references its change broke
    When the CLI checks a single document
    Then the CLI reports every reference finding located in that document, on the same terms as a check of the corpus
    And that covers the unresolved reference of AC-SPECREF-002.1, the retired target of AC-SPECREF-002.6, the incomplete pairing of AC-SPECREF-002.5 and the type mismatch of AC-SPECREF-002.10
    And the CLI also reports, located in each citing document, every unresolved-reference finding of a reference elsewhere citing a slug the document declared at git HEAD and no longer declares
    And the CLI also reports, located in each citing document, every unresolved-anchor finding of a reference elsewhere that resolves to the document
    And each such finding is the finding a check of the corpus reports for that reference

  @AC-SPECREF-002.13
  Scenario: A single-document check outside a git work tree reports the document's own findings
    Given a single-document check whose spec root is not in a git work tree, or where git cannot be run
    When the CLI checks the document
    Then the CLI writes a note to standard error that references elsewhere broken by the document's change were not checked
    And the CLI reports the document's own findings
    And the CLI exits with the status those findings decide

  @AC-SPECREF-002.12
  Scenario: An anchor-only reference reaching a heading of its own document is not unresolved
    Given a reference carrying an anchor and no slug
    When the CLI checks the corpus
    Then the CLI reports no unresolved-reference finding for it where its anchor resolves against the document holding the reference
    And the CLI reports it as an unresolved reference where its anchor does not
```

*Examples:* `reference/unresolved  error
specs/adrs/adr-004-module-blueprint-and-atomic-component.md:88:12`
"`[[adr-011]]` matches no document's slug" (AC-SPECREF-002.1).
`` `[[container-spec-cli#Element Registry]]` `` resolves iff a corpus
member declares `slug: container-spec-cli`; whether that document has an
`Element Registry` heading is not asked here (AC-SPECREF-002.2).
`related: ["work-order-template"]` reports as unresolved, because
templates carry placeholder slugs and are not corpus members
(AC-SPECREF-002.3). In an adopting profile that pairs `owner` with
`owned_by`, a document declaring `owner: team-ledger` where
`team-ledger` declares no matching `owned_by` produces a finding on each;
no pairing is registered by the profile this framework ships, so the
check is inert for its own corpus by design (AC-SPECREF-002.5). An
approved blueprint citing a retired ADR reports
`reference/retired-target  warning … "adr-009 is retired, superseded by
adr-013"` (AC-SPECREF-002.6). A reference to a file with no
frontmatter reports "target exists but is not a corpus member: excluded:
no-frontmatter", and a reference to a document whose frontmatter does not
parse reports nothing against the citing document — the parse failure is
reported once, against the document that has it (AC-SPECREF-002.7,
AC-SPEC-003.8). `` `[[ADR-012]]` ``
resolves to the document declaring `slug: adr-012` (AC-SPECREF-002.8). An
agent changes an ADR's slug and the write-time check on that one file
reports its own findings and the seven references elsewhere now pointing
at nothing, each at its line in the document holding it
(AC-SPECREF-002.11), while a dangling reference in an unrelated document
stays out of that check (AC-SPECREF-002.9); `specs refs` on the slug
before the change is what tells the agent in advance. A feature blueprint whose `frd`
names `adr-009` reports "frd expects type 'frd', target 'adr-009'
declares type 'adr'" (AC-SPECREF-002.10). The write-time check on a live
blueprint citing a retired ADR reports the retired-target warning at the
blueprint's own line, exactly as a corpus check does
(AC-SPECREF-002.11). `` `[[#Terminology]]` ``
in a document carrying a `Terminology` heading is reported by nothing,
while `` `[[#Terminolgy]]` `` in the same document is reported both as an
unresolved reference and as an unresolved anchor (AC-SPECREF-002.12).

*Rationale:* Severity is assigned here rather than left to
implementation because it is what the exit status of a build depends on.
AC-SPECREF-002.6 is a `warning` because a reference to a retired
document still resolves — the reader reaches a real document that tells
them it has been replaced — so failing a build on it would punish a
corpus for retiring anything. It is also the first check in this tool to
emit `warning` at all: AC-SPEC-006.7's `--strict` boundary has until now
been unreachable end to end, because every registered check emitted
`error`.

AC-SPECREF-002.2 is a scoping rule wearing a parser's clothes. Deferring
anchors to [[frd-spec-anchor-references]] does not let this feature
ignore that anchors exist: a reference that carries one is still a
reference to a document, and reading `container-spec-cli#Element
Registry` as a slug would report every anchored reference in a corpus as
broken — the false-positive failure mode AC-SPECREF-001.2 and
AC-SPECREF-001.4 exist to prevent, arriving by a different route. The
criterion states the minimum this feature must know about a convention
it otherwise leaves alone. No check here emits `unverified`; that
severity stays defined by AC-SPEC-006.3 and first gets used by the
sub-feature.

AC-SPECREF-002.9 keeps a single-document check about that document. A
write-time hook runs it on the file just written, and a check that also
reported other documents' unrelated findings would fail that hook on
breakage the author did not cause. What it does report from elsewhere is
exactly what this author's edit caused: a reference citing a slug the
document gave up, or an anchor into the document that no longer names a
heading (AC-SPECREF-002.11). Excluding those too, because the author
cannot fix them in the file in front of them, is the tempting
simplification and the wrong one: the commonest cause is an uncommitted
rename, whose fix is to revert it or to edit the citing documents, and
an author told nothing at write time learns of the breakage only from a
later corpus check — or not at all, where nothing runs one. The former
slug is read from git HEAD because that is the version every citation
was written against and an uncommitted rename is the common case; where
there is no git work tree there is no HEAD to read, so the check reports
what it can and says on standard error what it skipped
(AC-SPECREF-002.13), rather than refusing a check whose own findings are
still sound. A changed `type` or a retirement is not reported elsewhere:
the citing reference still resolves, and those findings belong to a
check of the citing document. FEAT-SPECREF-003 remains the blast-radius
question asked before the change.

AC-SPECREF-002.11 is also the positive half for the document's own
findings, stated for every reference rule rather than unresolved
references alone: narrowing a check to one document narrows the
documents it reports on, never the rules it runs over that document. Under AC-SPEC-006.8's accounting
`reference/retired-target passed` means "ran, found nothing", so a
narrowed check that skipped a rule would report as passed a rule whose
findings a check of the corpus raises in the same document.

AC-SPECREF-002.12 is what makes the slug-less anchor form [[adr-012]]
mandates for a section of the referring document usable in a corpus that
gates on `--strict`. That form carries no slug, so the slug half matches
nothing and this feature's own AC-SPECREF-002.1 calls it broken in the
same run the sub-feature's AC-SPECANCH-002.6 resolves it — two rules
disagreeing about one reference. The exemption is conjunctive rather
than a blanket skip of slug-less references: a reference whose anchor
names no heading of its own document reaches nothing and is still an
error, so the only thing waved through is a reference that resolves.

### FEAT-SPECREF-003: Reference Queries

**Acceptance Criteria:**

```gherkin
@FEAT-SPECREF-003
Feature: Reference Queries

  As an authoring agent
  I want to ask which documents reference a slug
  So that I understand the blast radius of retiring or renaming a document without reading every file

  Both directions are exposed because the index carries both and the two
  questions are asked at different moments — what breaks if I retire this,
  before a change, and what does this depend on, before an edit.

  @AC-SPECREF-003.1
  Scenario: A refs query returns every document that references a slug
    Then the CLI exposes a refs subcommand that, given a slug, returns every corpus member whose body or registered frontmatter properties reference it

  @AC-SPECREF-003.2
  Scenario: The outbound direction returns what the named slug references
    When the refs subcommand is invoked with an option selecting the outbound direction
    Then the CLI returns the documents the named slug references instead of those referencing it

  @AC-SPECREF-003.3
  Scenario: The full answer carries each reference's file path and span
    When the refs subcommand is invoked with the full option
    Then the CLI emits each returned reference with its file path and span

  @AC-SPECREF-003.4
  Scenario: A refs query answers from the warm index
    Then the refs subcommand answers from the warm index within the budget AC-SPEC-009.1 sets for a corpus query

  @AC-SPECREF-003.5
  Scenario: An unreferenced slug is an answer rather than a failure
    Given no corpus member references the named slug
    When the refs subcommand runs
    Then the CLI reports that
    And the CLI exits zero

  @AC-SPECREF-003.6
  Scenario: The default answer lists each referring document once
    When the refs subcommand is invoked without the full option or the outbound option
    Then the CLI returns the compact answer: one entry per referring document holding every reference that document makes in the answer
    And each reference carries only the properties [[cli-output-standards]] defines for the compact view of a reference

  @AC-SPECREF-003.7
  Scenario Outline: Either answer combines with every other form of the query
    When the refs subcommand is invoked with "<form>", with or without the full option
    Then the CLI returns the compact or the full answer, respectively, for the references that form selects

    Examples:
      | form                               |
      | the JSON option                    |
      | the outbound option                |
      | an argument qualified by an anchor |

  @AC-SPECREF-003.8 @negative
  Scenario: A compact answer omits no reference
    Given a refs query invoked with the full option
    When the same query is invoked without it
    Then the compact answer holds every reference the full answer holds

  @AC-SPECREF-003.9
  Scenario: The default outbound answer lists each cited slug once
    When the refs subcommand is invoked with the outbound option and without the full option
    Then the CLI returns one entry per cited slug, in the order each is first cited, holding every reference to that slug in the answer
    And each entry carries the "slug", the "path" of the document it resolves to, null where it resolves to none, a "resolution" holding its "state" only where it is not resolved, and its "lines"
    And each of its "lines" carries the "line" and "text" of one reference, its "origin" only for a reference in frontmatter, and "ambiguous_anchor" only where that is true
```

*Examples:* `specs refs adr-007` returns the four documents citing it, so
an author retiring it knows the blast radius (AC-SPECREF-003.1).
`specs refs frd-spec-reference-integrity --outbound` lists the documents
this FRD cites, so an agent about to edit it knows what it depends on
without reading it (AC-SPECREF-003.2). An agent pipes a `--json` result
into a rewrite, opening only the lines that carry the reference rather
than each citing document; it passes `--full` for the spans
(AC-SPECREF-003.3).
`specs refs frd-spec-reference-integrity` returning nothing is the answer
that a document is unreferenced, not a failure — the same contract
`slug-exists` sets in AC-SPEC-007.2 (AC-SPECREF-003.5).
`specs refs frd-spec-index` answers with one entry per citing document,
each reference reduced to the line it starts on and its text
(AC-SPECREF-003.6). `specs refs adr-012 --outbound` answers with one
entry for each slug adr-012 cites, each carrying the path to read it at
and the lines citing it (AC-SPECREF-003.9). `--json`, `--outbound` and
`specs refs 'container-spec-cli#Command Surface'` each answer in the same
compact form, and each adds `--full` for the full one (AC-SPECREF-003.7).
A compact answer drops properties, never references (AC-SPECREF-003.8).

*Rationale:* This extends the command surface FEAT-SPEC-011 fixed without
renaming or repurposing any of its four subcommands, which is the growth
that requirement anticipated. Both directions are exposed because the
index carries both (AC-SPECREF-001.3) and the two questions are asked at
different moments — "what breaks if I retire this" before a change,
"what does this depend on" before an edit. Answering either by reading
the corpus is the behavior the tool exists to replace.

AC-SPECREF-003.6 exists because the common question is "which documents,
and at what lines", and the full answer repeats on every reference what
the query already established — the slug asked about, the path it
resolved to — and spends most of its lines on span offsets. The compact
form is therefore the default, and `--full` is for the caller that
rewrites by offset. An outbound query's citing document is the
document queried, so grouping by it would give one group and drop the
one property an outbound query is for — where to read each dependency;
AC-SPECREF-003.9 therefore groups the outbound compact answer by the
slug cited and keeps its resolved path. Either compact answer is a
named view rather than a change of depth, which
[[cli-output-standards]] forbids `--json` to make, and AC-SPECREF-003.8
keeps it a view rather than a truncation.

### FEAT-SPECREF-004: Reference Configuration

**Acceptance Criteria:**

```gherkin
@FEAT-SPECREF-004
Feature: Reference Configuration

  As an adopting team
  I want to declare which of my frontmatter properties hold references and which form pairs
  So that the tool checks my conventions rather than assuming a fixed set

  @AC-SPECREF-004.1
  Scenario: One configuration file carries the reference vocabulary
    Then the configuration determines which frontmatter properties are reference-bearing and which pairs of properties form a paired relationship
    And it does so by extending the configuration file of FEAT-SPEC-008 rather than by a separate file

  @AC-SPECREF-004.2
  Scenario: A configuration for this framework's own properties ships with the CLI
    Then the CLI ships a configuration registering this framework's own reference-bearing properties and pairings, which an adopting repository can use without authoring one

  @AC-SPECREF-004.3
  Scenario: A pairing naming an unregistered property is reported rather than silently ignored
    Given a pairing that names a property not registered as reference-bearing
    When the CLI reads the configuration
    Then the CLI reports the offending key and its position
    And the CLI falls back to no default for it

  @AC-SPECREF-004.4
  Scenario: The configuration carries the lifecycle the retirement checks read
    Then the configuration determines the profile's status vocabulary, which of those statuses are live, which are retired, and which frontmatter property names a retired document's replacement
```

*Examples:* `reference_properties` and `pairings` are already declared
per document type in [[contract-specs-configuration]] and in
`schemas/profile.strata`; this feature is the first consumer of
properties the configuration schema already carries, and it adds one —
`typed_reference_properties`, mapping a reference-bearing property to the
document `type` its target must declare (AC-SPECREF-004.1, and
AC-SPECREF-002.10). The shipped configuration marks `superseded_by`,
`sources`, `frd`, `system`, `modules`, `provider` and
`contracts_consumed` as reference-bearing, and registers **no**
pairing — every relationship this framework declares has its reverse
derived from inbound edges, and [[document-relationship-standards]]
reserves a pairing for the case that does not. It registers one typed
reference: a feature blueprint's `frd` must resolve to a document of type
`frd` (AC-SPECREF-004.2). A pairing of `children` with `parent` where
only `parent_slug` is registered names the key and line, rather than
silently checking nothing — the vacuous-success failure this project has
now observed in three independently built tools (AC-SPECREF-004.3).
While the active profile is compiled in and unreachable from
`specs.toml` ([[contract-specs-configuration]]), the only file position
such a registration has is the `[profile].name` value that selected it,
so that is where the finding is positioned; once a registration becomes
authored data, the position is the offending key's own span. The shipped
configuration registers the whole status vocabulary, `retired` as its
only retired status, every other status as live, and `superseded_by` as
the successor property, so AC-SPECREF-002.6 can quote a replacement
without the engine knowing either name (AC-SPECREF-004.4).

*Rationale:* Configuration is what makes the checks in FEAT-SPECREF-002
belong to the adopting team rather than to this framework. A team whose
documents relate through `depends_on` and `required_by` gets the same
paired-relationship check without this repository knowing those names
exist. AC-SPECREF-004.3 exists because a misconfigured pairing is
indistinguishable from a passing one: the check runs, finds no
registered property to compare, and reports success.

*Note on numbering:* this feature's requirements carry the `SPECREF`
prefix rather than continuing [[frd-spec-index]]'s sequence.
FEAT-SPEC-004, the number reference integrity held before the split, is
left unused in that document rather than reclaimed here, so the ADRs
citing FEAT-SPEC-006, -008 and -009 keep resolving. The gap there and the
new prefix here are the same decision. AC-SPECREF-002.4 is unused for
the same reason: it required a retired document to declare a successor,
which is not true of every retirement, and it is removed rather than
renumbered.

## References

- [[frd-spec-index]] — the index, finding engine, corpus
  discovery, and configuration this feature extends
- [[frd-spec-anchor-references]] — the sub-feature that resolves the
  anchor half of a reference this feature resolves the slug half of
- [[frd-spec-slug-rename]] — the sub-feature that rewrites every
  reference to a slug when the slug is deliberately renamed
- [[frd-spec-structure]] — supplies the document headings the anchor
  sub-feature resolves an anchor against
- [[contract-specs-configuration]] — the configuration surface
  FEAT-SPECREF-004 consumes, whose `reference_properties` and `pairings`
  are already declared
- [[adr-012]] — the `[[slug]]` convention this feature parses and
  resolves, and the heading-anchor convention it defers
- [[adr-021]] — Gherkin, the notation these features and acceptance
  criteria are written in
- [[gherkin-guide]] — how to write them
- [[module-document-core]] — declares the References component this
  feature builds, and records that it is not yet built
- [[module-spec-profile]] — declares the Lifecycle Rules component whose
  relational half AC-SPECREF-002.6 supplies
- `plugin/skills/spec/references/cross-document-references.md` — what a
  reference resolves against
