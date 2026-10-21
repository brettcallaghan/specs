---
title: "Spec Corpus Reference Integrity"
description: Parses cross-document references from a spec corpus, records
  them in the shared index, and reports links that resolve to nothing, so
  a reference that silently broke is caught before a reader follows it
slug: frd-spec-reference-integrity
type: frd
status: approved
sources:
  - "frd-spec-index"
---

# Spec corpus reference integrity

## Overview

Specification documents cross-reference one another, through `[[slug]]` links in
prose and through frontmatter properties that name related documents. When a
target is renamed, retired, or never existed, the link silently stops resolving,
and a reader follows it into nothing. A corpus that is clean exactly once, at
the moment someone looked, is the state this feature exists to end.

This feature adds reference awareness to the tool that [[frd-spec-index]]
establishes. It reads each document's body and reference-bearing frontmatter,
records the outbound and inbound references in the shared index, and reports
every reference that resolves to no corpus member. It adds no new artifact and
no new engine. It populates the reference properties of the index that
FEAT-SPEC-002 defines, and it emits findings through the engine that
FEAT-SPEC-006 defines. Corpus discovery, membership, output location, and the
warm index come unchanged from that FRD, and this FRD doesn't restate them.

That FRD's ordering governs this one. The index records what a document
references, including references that resolve to nothing, and never fails.
Reporting an unresolved reference is the checker's job. A repository whose links
are broken still contributes its graph to an organization's catalogue.

**Out of scope for this feature.** Each of these is deferred deliberately, not
overlooked:

- **Anchor resolution.** A reference may name a heading anchor, such as
  `[[container-spec-cli#Element Registry]]`, the form that [[adr-012]] mandates
  for un-promoted components. This feature resolves the slug half and doesn't
  interpret the anchor (AC-SPECREF-002.2). [[frd-spec-anchor-references]], a
  sub-feature of this one, specifies how to interpret it. The split is
  deliberate: anchor resolution needs heading data that this feature doesn't
  have, from [[frd-spec-structure]].
- **Heading extraction and section spans.** [[frd-spec-structure]] specifies
  these. Nothing in this feature reads a document's headings.
- **Suggesting a correction for a broken reference.** Findings state the fact
  and leave the fix to a reviewing agent.
- **Repairing or rewriting references.** This feature only reads and reports, as
  FEAT-SPEC-002's ordering requires. [[frd-spec-slug-rename]], a sub-feature of
  this one, specifies rewriting every reference to a slug when an author
  deliberately renames that slug. Both documents hold the boundary: a check
  never edits what it merely infers is wrong, and a rename edits only what an
  author explicitly instructed.
- **References to documents outside the specification set**: the backticked
  relative paths that [[cross-document-references]] governs. A path that
  resolves to no file is a broken reference by the same argument as a `[[slug]]`
  that resolves to no document. But it belongs to a different resolution domain
  (paths relative to the containing file, not slugs in a flat namespace), and
  this FRD doesn't specify it.
- **The shape of the reference data in the index.** Whether the index embeds
  references in each document's entry or collects them separately as edges is a
  schema decision that the index-schema ADR records, not this FRD. This FRD
  states what must be recorded and answerable, not how the index models it.
- **Resolving a reference across workspaces.** AC-SPEC-002.3 gives every entry a
  workspace identifier so that indexes from separate repositories merge into one
  graph. This feature, though, resolves only within the workspace a run targets,
  and it reports a reference to a slug that exists only in another repository's
  index as unresolved. The merged-graph case is a known follow-on, not a defect
  for the first team that tries it to discover.
- **Reporting an unreferenced document.** Inbound edges make an orphan free to
  detect. But whether a document *should* be reachable is a judgment about the
  corpus, not a conformance rule: a draft is expected to be uncited, and a
  warning on every new document is a warning nobody reads. AC-SPECREF-003.5
  answers the question on demand, and no check reports it.
- **Rendering the reference graph**, and export to any third-party format.

## Terminology

- **Body** — the content of a document after its frontmatter block. Reference
  extraction reads the body. Frontmatter contributes references only through the
  properties the profile registers as reference-bearing, never through prose in
  a frontmatter value.
- **Outbound reference** — a `[[slug]]` link parsed from a document's body, or a
  slug value in a frontmatter property that the active profile registers as
  reference-bearing.
- **Anchor** — the portion of a reference after a `#`, naming a heading within
  the target document. This feature reads a reference's slug half, and
  [[frd-spec-anchor-references]] specifies the anchor half.
- **Inbound reference** — the reverse edge: for a given document, every other
  corpus member whose outbound references name it.
- **Resolution** — matching a reference's slug against the slugs that corpus
  members declare, without regard to letter case, per AC-SPEC-003.4. A reference
  resolves against a target's frontmatter `slug`, never a file path.
- **Unresolved reference** — a reference whose slug matches no corpus member. A
  reference whose slug matches a file that the index recorded as excluded is
  unresolved too, but it's reported with that reason (AC-SPECREF-002.7), because
  the reader's next step differs. A reference whose slug matches a file that the
  index recorded as a parse failure isn't reported at all: the citing document
  isn't what's wrong, and the parse failure is the one finding (AC-SPEC-003.8).
- **Reference-bearing property** — a frontmatter property that the active
  profile declares as holding one or more document slugs (for example,
  `superseded_by` or `modules`). A profile MAY additionally declare the document
  `type` such a property's target must carry, making the property a *typed*
  reference.
- **Paired relationship** — two reference-bearing properties that the profile
  declares as counterparts, so that a reference in one document implies a
  matching reference back. An adopting profile that pairs `owner` with
  `owned_by` gets both halves checked. The specs profile registers no pairing:
  per [[document-relationship-standards]], a pairing is reserved for a
  relationship that a reader must see without the index, and every relationship
  this framework declares has its reverse derived.

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
`adr-006`, each carrying the byte offset and the line and column of the link
(AC-SPECREF-001.1). The sentence "write `` `[[security-standards]]` `` to cite a
standard" contributes no reference, because the link is an illustration
(AC-SPECREF-001.2). `adr-004`'s entry lists `adr-005`, `adr-007` and `adr-009`
among its inbound references, without any of them declaring the relationship in
frontmatter (AC-SPECREF-001.3). Line 3 of
`specs/adrs/adr-012-reference-conventions.md` carries
`description: References use [[slug]] wikilinks …`, which names the notation
rather than a document (AC-SPECREF-001.4). An FRD that cites `` `[[adr-021]]` ``
in its Requirements and again in its References carries two entries, so a
finding against either cites the line the reader must open. Ordering by position
is what makes the index byte-identical across platforms per AC-SPEC-009.7
(AC-SPECREF-001.5). A corpus with eleven broken links still produces a complete
index (AC-SPECREF-001.6). A document's `related: ["adr-012", "adr-002"]` records
two outbound references, each naming the property it came from
(AC-SPECREF-001.7). `` `[[#Terminology]]` `` written in a document that carries
a `Terminology` heading is listed among that document's own inbound references.
So `specs refs` answers with the section link that a reader renaming that
heading has to fix. `` `[[#Nowhere]]` `` written beside it is listed among none,
because it reaches no heading to break (AC-SPECREF-001.8).

*Rationale:* AC-SPECREF-001.2 and AC-SPECREF-001.4 apply the same rule to the
two places where a reference is written about rather than written: a backticked
example, and a frontmatter value describing the notation. Both occur in this
corpus. Both are false positives that a reader would have to dismiss by hand,
which is the failure mode that gets a checker ignored. AC-SPECREF-001.6 restates
FEAT-SPEC-002's ordering for this feature specifically, because a reference
graph is the part of the index most likely to tempt an implementer into failing
the build from the indexer.

AC-SPECREF-001.7 is stated positively because AC-SPECREF-001.1 covers only the
body, and everything downstream assumes that frontmatter references are in the
index. AC-SPECREF-002.3 checks them, AC-SPECREF-003.1 answers for them,
AC-SPECREF-002.5 compares a pairing across two documents, and AC-SPECREN-001.4
rewrites them from their recorded spans. Left implicit, the reasonable reading
of FEAT-SPECREF-001 is that the index holds prose links only. That reading
passes every criterion in this requirement and silently drops the half of the
graph that a catalogue reads first.

AC-SPECREF-001.8 covers what AC-SPECREF-001.3's "all other corpus members"
leaves open: the one reference whose target is the document that holds it. The
anchor-only form, which AC-SPECANCH-002.6 resolves against the containing
document, is a reference like any other (AC-SPECREF-002.12 already refuses to
call it broken). So a query that omitted it would answer "nothing references
this heading" about a document that visibly does. That answer is wrong rather
than failing, and no check sees it. That's why the criterion is stated here
rather than left to the queries that FEAT-SPECREF-003 and FEAT-SPECANCH-003
define over this graph. The conjunction is AC-SPECREF-002.12's: an empty slug
alone earns nothing, and an anchor that reaches no heading records no edge,
because a reference that reaches nothing has no second end.

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
`` `[[container-spec-cli#Element Registry]]` `` resolves if and only if a corpus
member declares `slug: container-spec-cli`. Whether that document has an
`Element Registry` heading isn't asked here (AC-SPECREF-002.2).
`related: ["work-order-template"]` reports as unresolved, because templates
carry placeholder slugs and aren't corpus members (AC-SPECREF-002.3). In an
adopting profile that pairs `owner` with `owned_by`, a document that declares
`owner: team-ledger`, where `team-ledger` declares no matching `owned_by`,
produces a finding on each. The profile this framework ships registers no
pairing, so the check is inert for its own corpus by design (AC-SPECREF-002.5).
An approved blueprint citing a retired ADR reports
`reference/retired-target  warning … "adr-009 is retired, superseded by
adr-013"` (AC-SPECREF-002.6). A reference to a file with no frontmatter reports
"target exists but is not a corpus member: excluded: no-frontmatter". A
reference to a document whose frontmatter doesn't parse reports nothing against
the citing document, and the parse failure is reported once, against the
document that has it (AC-SPECREF-002.7, AC-SPEC-003.8). `` `[[ADR-012]]` ``
resolves to the document that declares `slug: adr-012` (AC-SPECREF-002.8). An
agent changes an ADR's slug, and the write-time check on that one file reports
its own findings and the seven references elsewhere that now point at nothing,
each at its line in the document that holds it (AC-SPECREF-002.11). A dangling
reference in an unrelated document stays out of that check (AC-SPECREF-002.9).
Running `specs refs` on the slug before the change is what tells the agent in
advance. A Feature Blueprint whose `frd` names `adr-009` reports "frd expects
type 'frd', target 'adr-009' declares type 'adr'" (AC-SPECREF-002.10). The
write-time check on a live blueprint citing a retired ADR reports the
retired-target warning at the blueprint's own line, exactly as a corpus check
does (AC-SPECREF-002.11). `` `[[#Terminology]]` `` in a document that carries a
`Terminology` heading is reported by nothing, while `` `[[#Terminolgy]]` `` in
the same document is reported both as an unresolved reference and as an
unresolved anchor (AC-SPECREF-002.12).

*Rationale:* Severity is assigned here rather than left to implementation,
because the exit status of a build depends on it. AC-SPECREF-002.6 is a
`warning` because a reference to a retired document still resolves: the reader
reaches a real document that says it has been replaced. So failing a build on it
would punish a corpus for retiring anything. Because it emits `warning`, it also
makes AC-SPEC-006.7's `--strict` boundary reachable end to end.

AC-SPECREF-002.2 is a scoping rule wearing a parser's clothes. Deferring anchors
to [[frd-spec-anchor-references]] doesn't let this feature ignore that anchors
exist. A reference that carries one is still a reference to a document, and
reading `container-spec-cli#Element
Registry` as a slug would report every anchored reference in a corpus as broken.
That's the false-positive failure mode that AC-SPECREF-001.2 and
AC-SPECREF-001.4 exist to prevent, arriving by a different route. The criterion
states the minimum this feature must know about a convention it otherwise leaves
alone. No check here emits `unverified`. AC-SPEC-006.3 defines that severity,
and the sub-feature emits it.

AC-SPECREF-002.9 keeps a single-document check about that document. A write-time
hook runs it on the file just written, and a check that also reported other
documents' unrelated findings would fail that hook on breakage the author didn't
cause. What it does report from elsewhere is exactly what this author's edit
caused: a reference citing a slug the document gave up, or an anchor into the
document that no longer names a heading (AC-SPECREF-002.11). Excluding those
too, because the author can't fix them in the file in front of them, is the
tempting simplification and the wrong one: the commonest cause is an uncommitted
rename, whose fix is to revert it or to edit the citing documents. An author
told nothing at write time learns of the breakage only from a later corpus
check, or not at all where nothing runs one. The former slug is read from git
HEAD, because that's the version every citation was written against and an
uncommitted rename is the common case. Where there's no git work tree, there's
no HEAD to read. So the check reports what it can and says on standard error
what it skipped (AC-SPECREF-002.13), rather than refusing a check whose own
findings are still sound. A changed `type` or a retirement isn't reported
elsewhere: the citing reference still resolves, and those findings belong to a
check of the citing document. FEAT-SPECREF-003 remains the blast-radius question
asked before the change.

AC-SPECREF-002.11 is also the positive half for the document's own findings.
It's stated for every reference rule rather than for unresolved references
alone, because narrowing a check to one document narrows the documents it
reports on, never the rules it runs over that document. Under AC-SPEC-006.8's
accounting, `reference/retired-target passed` means "ran, found nothing". So a
narrowed check that skipped a rule would report as passed a rule whose findings
a check of the corpus raises in the same document.

AC-SPECREF-002.12 is what makes the slug-less anchor form usable in a corpus
that gates on `--strict`. [[adr-012]] mandates that form for a section of the
referring document. It carries no slug, so its slug half matches nothing, and
this feature's own AC-SPECREF-002.1 calls it broken in the same run that the
sub-feature's AC-SPECANCH-002.6 resolves it. That's two rules disagreeing about
one reference. The exemption is conjunctive rather than a blanket skip of
slug-less references. A reference whose anchor names no heading of its own
document reaches nothing and is still an error, so the only thing waved through
is a reference that resolves.

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

*Examples:* `specs refs adr-007` returns the documents that cite it, so an
author retiring it knows the blast radius (AC-SPECREF-003.1).
`specs refs frd-spec-reference-integrity --outbound` lists the documents this
FRD cites, so an agent about to edit it knows what it depends on without reading
it (AC-SPECREF-003.2). An agent pipes a `--json` result into a rewrite and opens
only the lines that carry the reference rather than each citing document. It
passes `--full` for the spans (AC-SPECREF-003.3). When
`specs refs frd-spec-reference-integrity` returns nothing, that's the answer
that a document is unreferenced, not a failure. It's the same contract that
`slug-exists` sets in AC-SPEC-007.2 (AC-SPECREF-003.5).
`specs refs frd-spec-index` answers with one entry per citing document, with
each reference reduced to the line it starts on and its text (AC-SPECREF-003.6).
`specs refs adr-012 --outbound` answers with one entry for each slug that
adr-012 cites, each carrying the path to read it at and the lines that cite it
(AC-SPECREF-003.9). `--json`, `--outbound`, and
`specs refs 'container-spec-cli#Command Surface'` each answer in the same
compact form, and each adds `--full` for the full one (AC-SPECREF-003.7). A
compact answer drops properties, never references (AC-SPECREF-003.8).

*Rationale:* This extends the command surface that FEAT-SPEC-011 fixed without
renaming or repurposing any of its four subcommands, which is the growth that
requirement anticipated. Both directions are exposed because the index carries
both (AC-SPECREF-001.3) and the two questions are asked at different moments:
"what breaks if I retire this" before a change, and "what does this depend on"
before an edit. Answering either by reading the corpus is the behavior the tool
exists to replace.

AC-SPECREF-003.6 exists because the common question is "which documents, and at
what lines". The full answer repeats on every reference what the query already
established (the slug asked about and the path it resolved to), and it spends
most of its lines on span offsets. So the compact form is the default, and
`--full` is for the caller that rewrites by offset. An outbound query's citing
document is the document queried. Grouping by it would give one group and drop
the one property an outbound query is for: where to read each dependency. So
AC-SPECREF-003.9 groups the outbound compact answer by the slug cited and keeps
its resolved path. Either compact answer is a named view rather than a change of
depth, which [[cli-output-standards]] forbids `--json` to make, and
AC-SPECREF-003.8 keeps it a view rather than a truncation.

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

*Examples:* [[contract-specs-configuration]] and `schemas/profile.strata`
declare `reference_properties` and `pairings` per document type. This feature
consumes those properties and adds one, `typed_reference_properties`, which maps
a reference-bearing property to the document `type` its target must declare
(AC-SPECREF-004.1, and AC-SPECREF-002.10). The shipped configuration marks
`superseded_by`, `sources`, `frd`, `system`, `modules`, `provider` and
`contracts_consumed` as reference-bearing. It registers no pairing, because
every relationship this framework declares has its reverse derived from inbound
edges, and [[document-relationship-standards]] reserves a pairing for the case
that doesn't. The configuration registers one typed reference: a Feature
Blueprint's `frd` must resolve to a document of type `frd` (AC-SPECREF-004.2). A
pairing of `children` with `parent`, where only `parent_slug` is registered,
names the key and line rather than silently checking nothing, which is the
vacuous-success failure (AC-SPECREF-004.3). While the active profile is compiled
in and unreachable from `specs.toml` ([[contract-specs-configuration]]), the
only file position such a registration has is the `[profile].name` value that
selected it, so the finding is positioned there. Once a registration becomes
authored data, the position is the offending key's own span. The shipped
configuration registers the whole status vocabulary, with `retired` as its only
retired status, every other status as live, and `superseded_by` as the successor
property. So AC-SPECREF-002.6 can quote a replacement without the engine knowing
either name (AC-SPECREF-004.4).

*Rationale:* Configuration is what makes the checks in FEAT-SPECREF-002 belong
to the adopting team rather than to this framework. A team whose documents
relate through `depends_on` and `required_by` gets the same paired-relationship
check without this repository knowing those names exist. AC-SPECREF-004.3 exists
because a misconfigured pairing is indistinguishable from a passing one: the
check runs, finds no registered property to compare, and reports success.

*Note on numbering:* this feature's requirements carry the `SPECREF` prefix
rather than continuing [[frd-spec-index]]'s sequence. FEAT-SPEC-004, the number
reference integrity held before the split, stays unused in that document rather
than reclaimed here, so the ADRs that cite FEAT-SPEC-006, -008 and -009 keep
resolving. The gap there and the new prefix here are the same decision.
AC-SPECREF-002.4 is unused for the same reason. It required a retired document
to declare a successor, which isn't true of every retirement, and it's removed
rather than renumbered.

## References

- [[frd-spec-index]] — the index, finding engine, corpus discovery, and
  configuration this feature extends
- [[frd-spec-anchor-references]] — the sub-feature that resolves the anchor half
  of a reference, whose slug half this feature resolves
- [[frd-spec-slug-rename]] — the sub-feature that rewrites every reference to a
  slug when the slug is deliberately renamed
- [[frd-spec-structure]] — supplies the document headings that the anchor
  sub-feature resolves an anchor against
- [[contract-specs-configuration]] — the configuration surface FEAT-SPECREF-004
  consumes, which declares `reference_properties` and `pairings`
- [[adr-012]] — the `[[slug]]` convention this feature parses and resolves, and
  the heading-anchor convention it defers
- [[adr-021]] — Gherkin, the notation these features and acceptance criteria are
  written in
- [[gherkin-guide]] — how to write them
- [[module-document-core]] — declares the References component this feature
  builds
- [[module-spec-profile]] — registers the lifecycle rules whose relational half
  AC-SPECREF-002.6 supplies
- [[cross-document-references]] — what a reference resolves against
