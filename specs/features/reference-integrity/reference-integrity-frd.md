---
title: "FRD: Spec Corpus Reference Integrity"
description: Parses cross-document references from a spec corpus, records
  them in the shared index, and reports links that resolve to nothing, so
  a reference that silently broke is caught before a reader follows it
slug: frd-spec-reference-integrity
type: frd
status: approved
updated_at: 2026-10-07T00:00:00Z
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
  excluded or unparsable is unresolved too, but is reported differently
  (AC-SPECREF-002.7), because the reader's next step differs.
- **Reference-bearing property** — a frontmatter property the active
  profile declares as holding one or more document slugs (for example
  `superseded_by` or `modules`). A profile MAY additionally declare
  the document `type` such a property's target must carry, making the
  property a *typed* reference.
- **Paired relationship** — two reference-bearing properties the profile
  declares as counterparts, so that a reference in one document implies a
  matching reference back. An adopting profile that pairs `owner` with
  `owned_by` gets both halves checked. The spec-driven-dev profile
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

  A finding belongs to the document a reader must open to fix it, which
  for a reference is the referring document rather than the one that was
  edited. That is why the last two criteria widen a single-document check
  rather than narrowing it.

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
  Scenario: A slug matching an excluded or unparsable file is reported with that reason
    Given a reference whose slug matches a file the index recorded as excluded or as a parse failure rather than a corpus member
    When the CLI checks the corpus
    Then the CLI reports the finding with that reason rather than as a slug matching nothing

  @AC-SPECREF-002.8
  Scenario: Resolution ignores letter case
    Then the CLI resolves a reference against corpus slugs without regard to letter case

  @AC-SPECREF-002.9
  Scenario: A single-document check reports inbound references that no longer resolve
    When the CLI checks a single document
    Then the CLI reports inbound references to that document that no longer resolve, at the severity AC-SPECREF-002.1 assigns, each cited at the span in the referring document, in addition to findings against the document named

  @AC-SPECREF-002.10
  Scenario: A typed reference resolving to the wrong type is an error
    Given the active profile declares a required target type for a reference-bearing property
    And the resolved document declares a different type
    When the CLI checks the document
    Then the CLI reports a finding of severity error naming the property, the type expected and the type the target declares

  @AC-SPECREF-002.11
  Scenario: A single-document check reports every reference finding at the document holding the reference
    When the CLI checks a single document
    Then the CLI reports every reference finding against the document that holds the reference, whichever document that is, in addition to findings against the document named
    And that covers the retired target of AC-SPECREF-002.6, the incomplete pairing of AC-SPECREF-002.5 and the type mismatch of AC-SPECREF-002.10, on the same terms AC-SPECREF-002.9 states for an unresolved reference

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
adr-013"` (AC-SPECREF-002.6). A reference to a document whose
frontmatter does not parse reports "target exists but is not a corpus
member: parse failure at line 28" (AC-SPECREF-002.7). `` `[[ADR-012]]` ``
resolves to the document declaring `slug: adr-012` (AC-SPECREF-002.8). An
agent changes an ADR's slug and the write-time check on that one file
reports the seven documents now pointing at nothing, at their own lines
(AC-SPECREF-002.9). A feature blueprint whose `frd` names `adr-009`
reports "frd expects type 'frd', target 'adr-009' declares type 'adr'"
(AC-SPECREF-002.10). An agent sets an ADR to `status: retired` and the
write-time check on that one file reports the four live documents still
citing it, at their own lines (AC-SPECREF-002.11). `` `[[#Terminology]]` ``
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

AC-SPECREF-002.9 is the one place this feature widens a check's blast
radius rather than narrowing it, and it does so deliberately.
AC-SPEC-001.2 already resolves a single document's references against
the whole corpus, so the data is present; what was missing is the
reverse direction. A finding belongs to the document a reader must open
to fix it, which for broken inbound references is the referrer, not the
document that was edited. Every unresolved reference in the corpus is
already derived on any check, narrow or not, so this costs a narrowing
exception rather than a corpus walk and stays inside AC-SPEC-009.2's
single-document budget.

What it delivers is a superset of what this criterion asks for: a
narrowed check reports every unresolved reference, not only those that
had been inbound to the document named. Distinguishing the two needs
prior state the corpus does not keep — an edit invalidates the content
fingerprint, so the check reads a rebuilt graph with no memory of what
the references used to resolve to. In a corpus kept green the
distinction is empty, since anything reported is new; narrowing it to
exactly this criterion's set is deferred to the per-file digests of
`agent-plugins-h6c.1`.

AC-SPECREF-002.11 generalizes that widening to the other three reference
rules, and it exists because scoping it to unresolved references alone
was a defect rather than a boundary (`agent-plugins-bk4`). The argument
for AC-SPECREF-002.9 — a finding belongs to the document a reader must
open to fix it, which for a reference is the referrer — is a property of
*references*, not of unresolved ones. Retiring a document, or changing
its `type`, breaks a referring document exactly as renaming its slug
does; the reference still resolves, so nothing about the reader's
position changes. Leaving those two narrowed made the report state
something false rather than merely incomplete: `reference/retired-target
passed` means "ran, found nothing" under AC-SPEC-006.8's accounting,
while the same corpus checked whole reported findings the narrowing had
dropped. A check that reports a rule as passed on evidence it discarded
is worse than one that does not run the rule.

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
  Scenario: JSON output carries each reference's file path and span
    When the refs subcommand is invoked with the JSON option
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
```

*Examples:* `specs refs adr-007` returns the four documents citing it, so
an author retiring it knows the blast radius (AC-SPECREF-003.1).
`specs refs frd-spec-reference-integrity --outbound` lists the documents
this FRD cites, so an agent about to edit it knows what it depends on
without reading it (AC-SPECREF-003.2). An agent pipes a `--json` result
into a rewrite, opening only the lines that carry the reference rather
than each citing document (AC-SPECREF-003.3).
`specs refs frd-spec-reference-integrity` returning nothing is the answer
that a document is unreferenced, not a failure — the same contract
`slug-exists` sets in AC-SPEC-007.2 (AC-SPECREF-003.5).

*Rationale:* This extends the command surface FEAT-SPEC-011 fixed without
renaming or repurposing any of its four subcommands, which is the growth
that requirement anticipated. Both directions are exposed because the
index carries both (AC-SPECREF-001.3) and the two questions are asked at
different moments — "what breaks if I retire this" before a change,
"what does this depend on" before an edit. Answering either by reading
the corpus is the behavior the tool exists to replace.

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
