---
title: "FRD: Spec Corpus Reference Integrity"
description: Parses cross-document references from a spec corpus, records
  them in the shared index, and reports links that resolve to nothing, so
  a reference that silently broke is caught before a reader follows it
slug: frd-spec-reference-integrity
type: frd
status: approved
updated_at: 2026-08-09T00:00:00Z
sources:
  - "frd-spec-validation-cli"
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
examined: no reference in the corpus names a heading anchor, no retired
document's successor has been resolved, and no paired relationship has
been checked from both ends. Those are not passing checks. They are
checks that have never run.

This feature adds reference awareness to the tool established by
[[frd-spec-validation-cli]]. It reads each document's body and
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
  [[frd-spec-structure-conformance]], and the corpus contains no anchor
  reference today — so shipping slug resolution first costs the corpus
  nothing and unblocks it sooner.
- **Heading extraction and section spans.** Specified in
  [[frd-spec-structure-conformance]]. Nothing in this feature reads a
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
  `superseded_by` or `contained_by`). A profile MAY additionally declare
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

**User Story:** As an architecture owner, I want every document's
outbound and inbound references recorded in the index, so that impact
analysis and a central catalogue read the relationship graph from the
same artifact instead of re-deriving it.

**Acceptance Criteria:**

- **AC-SPECREF-001.1:** WHEN the CLI builds the index, the CLI SHALL add
  to every corpus member's entry a `references` property holding every
  `[[slug]]` parsed from its body, each with the span at which it occurs.
  - *Example:* `adr-011`'s entry gains outbound references to `adr-004`
    and `adr-006`, each carrying the byte offset and line/column of the
    link.
- **AC-SPECREF-001.2:** The CLI SHALL NOT record text inside a fenced
  code block or an inline code span as a reference.
  - *Example:* The sentence "write `` `[[security-standards]]` `` to
    cite a standard" contributes no reference, because the link is an
    illustration.
- **AC-SPECREF-001.3:** WHEN the CLI builds the index, the CLI SHALL add
  inbound references for every document, derived from the outbound
  references of all other corpus members.
  - *Example:* `adr-004`'s entry lists `adr-005`, `adr-006` and
    `adr-007` as inbound, without any of them declaring the relationship
    in frontmatter.
- **AC-SPECREF-001.4:** The CLI SHALL NOT record a `[[slug]]` occurring
  in a frontmatter value as a reference; frontmatter contributes
  references only through the properties the active profile registers as
  reference-bearing.
  - *Example:* `specs/adrs/ADR-012-reference-conventions.md` line 3
    carries `description: References use [[slug]] wikilinks …`, which
    names the notation rather than a document. A parser reading the
    whole file rather than the body reports it as unresolved — the only
    unresolved reference a scan of this corpus produces today, and a
    false one.
- **AC-SPECREF-001.5:** WHEN a document references the same target more
  than once, the CLI SHALL record each occurrence with its own span, and
  SHALL order a document's outbound references by position in the
  document.
  - *Example:* An FRD citing `[[adr-002]]` in its Requirements and again
    in its References carries two entries, so a finding against either
    cites the line the reader must open. Ordering by position is what
    makes the index byte-identical across platforms per AC-SPEC-009.7.
- **AC-SPECREF-001.6:** WHEN a document's outbound references include
  one that resolves to no corpus member, the CLI SHALL record it in the
  index and SHALL exit zero.
  - *Example:* A corpus with eleven broken links still produces a
    complete index; the eleven are reported by `check`, per the ordering
    FEAT-SPEC-002 sets — indexing never fails, validation may.
- **AC-SPECREF-001.7:** WHEN the CLI builds the index, the CLI SHALL
  record every value held by a frontmatter property the active profile
  registers as reference-bearing as an outbound reference of that
  document, each with the span at which it occurs, distinguishable from
  a reference parsed from the body.
  - *Example:* A document's `related: ["adr-012", "adr-002"]` records
    two outbound references naming the property they came from, so
    AC-SPECREF-002.3 can name it in a finding, AC-SPECREF-003.1 can
    answer for a document cited only in frontmatter, and the inbound
    edges of AC-SPECREF-001.3 include the relational graph rather than
    only the prose one.

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

### FEAT-SPECREF-002: Reference Integrity Checks

**User Story:** As a document author, I want every cross-reference
checked, so that a link that silently stopped resolving is caught before
a reader follows it into nothing.

**Acceptance Criteria:**

- **AC-SPECREF-002.1:** WHEN a document body contains a `[[slug]]`
  reference that matches no corpus member's slug, the CLI SHALL report a
  finding of severity `error` with the file, span, and unresolved slug.
  - *Example:* `reference/unresolved  error
    specs/adrs/ADR-004-module-blueprint-and-atomic-component.md:88:12`
    "`[[adr-011]]` matches no document's slug". This was this
    repository's own state until the slugs were migrated to the
    number-only form [[adr-012]] decided; the check exists so the next
    such drift is caught by the build rather than by a prototype.
- **AC-SPECREF-002.2:** WHEN a reference names a heading anchor, the CLI
  SHALL resolve the reference on its slug portion alone, and SHALL NOT
  report a finding about the anchor.
  - *Example:* `[[container-spec-cli#Element Registry]]` resolves iff a
    corpus member declares `slug: container-spec-cli`; whether that
    document has an `Element Registry` heading is not asked here.
    Splitting on the first `#` is the whole of this feature's anchor
    handling — without it the slug half would be read as
    `container-spec-cli#Element Registry`, matching nothing, and every
    anchor reference in a corpus would report as broken.
- **AC-SPECREF-002.3:** WHEN a frontmatter property registered by the
  active profile as reference-bearing contains a value matching no
  corpus member's slug, the CLI SHALL report a finding of severity
  `error` naming the property and the value.
  - *Example:* `related: ["work-order-template"]` reports as unresolved,
    because templates carry placeholder slugs and are not corpus
    members.
- **AC-SPECREF-002.5:** WHEN the active profile registers a frontmatter
  property as one half of a paired relationship and the named document
  does not declare the matching counterpart, the CLI SHALL report a
  finding of severity `error` against both documents.
  - *Example:* In an adopting profile that pairs `owner` with
    `owned_by`, a document declaring `owner: team-ledger` where
    `team-ledger` declares no matching `owned_by` produces a finding on
    each. No pairing is registered by the profile this framework ships,
    so this check is inert for its own corpus by design.
- **AC-SPECREF-002.6:** WHEN a document at a status the profile treats
  as live references a document at `status: retired`, the CLI SHALL
  report a finding of severity `warning` naming the retired target and
  its `superseded_by` value when it has one.
  - *Example:* An approved blueprint citing a retired ADR reports
    `reference/retired-target  warning … "adr-009 is retired,
    superseded by adr-013"`, which does not fail a build but is the one
    finding that tells an author what to cite instead.
- **AC-SPECREF-002.7:** WHEN a reference's slug matches a file the index
  recorded as excluded or as a parse failure rather than a corpus
  member, the CLI SHALL report the finding with that reason rather than
  as a slug matching nothing.
  - *Example:* A reference to a document whose frontmatter does not
    parse reports "target exists but is not a corpus member: parse
    failure at line 28" — one lookup for the reader instead of a search
    for a document that is right there.
- **AC-SPECREF-002.8:** The CLI SHALL resolve a reference against corpus
  slugs without regard to letter case.
  - *Example:* `[[ADR-012]]` resolves to the document declaring
    `slug: adr-012`, matching the case-insensitive slug uniqueness
    AC-SPEC-003.4 already establishes — the two rules must agree, or a
    corpus can hold a reference that resolves to two documents that
    cannot both exist.
- **AC-SPECREF-002.9:** WHEN the CLI checks a single document, the CLI
  SHALL report inbound references to that document that no longer
  resolve at the severity AC-SPECREF-002.1 assigns, each cited at the
  span in the referring document, in addition to findings against the
  document named.
  - *Example:* An agent changes an ADR's slug and the write-time check
    on that one file reports the seven documents now pointing at
    nothing, at their own lines. Without this, the edit most likely to
    break references is the one a single-document check cannot see: the
    seven referrers were not the file that was written, so nothing
    examines them until CI.
- **AC-SPECREF-002.10:** WHEN the active profile declares a required
  target `type` for a reference-bearing property and the resolved
  document declares a different `type`, the CLI SHALL report a finding
  of severity `error` naming the property, the type expected and the
  type the target declares.
  - *Example:* A feature blueprint whose `frd` names `adr-009` reports
    "frd expects type 'frd', target 'adr-009' declares type 'adr'". The
    schema can require that `frd` holds exactly one slug; only the index
    knows what that slug resolves to, so this half of the constraint has
    to live here.
- **AC-SPECREF-002.11:** WHEN the CLI checks a single document, the CLI
  SHALL report every reference finding against the document that holds
  the reference, whichever document that is, in addition to findings
  against the document named — the retired target of AC-SPECREF-002.6,
  the incomplete pairing of AC-SPECREF-002.5 and the type mismatch of
  AC-SPECREF-002.10, on the same terms AC-SPECREF-002.9 states for an
  unresolved reference.
  - *Example:* An agent sets an ADR to `status: retired` and the
    write-time check on that one file reports the four live documents
    still citing it, at their own lines. Under AC-SPECREF-002.9 alone
    that same check reported nothing and printed
    `reference/retired-target passed`: retiring a document does not stop
    its references resolving, so the edit whose entire purpose is to
    change what those references *mean* was invisible to the check an
    author runs before flipping status.

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

### FEAT-SPECREF-003: Reference Queries

**User Story:** As an authoring agent, I want to ask which documents
reference a slug, so that I understand the blast radius of retiring or
renaming a document without reading every file.

**Acceptance Criteria:**

- **AC-SPECREF-003.1:** The CLI SHALL expose a `refs` subcommand that,
  given a slug, returns every corpus member whose body or registered
  frontmatter properties reference it.
  - *Example:* `spec refs adr-007` returns the four documents citing it,
    so an author retiring it knows the blast radius.
- **AC-SPECREF-003.2:** WHEN the `refs` subcommand is invoked with an
  option selecting the outbound direction, the CLI SHALL return the
  documents the named slug references instead of those referencing it.
  - *Example:* `spec refs frd-spec-reference-integrity --outbound`
    lists the nine documents this FRD cites, so an agent about to edit
    it knows what it depends on without reading it.
- **AC-SPECREF-003.3:** WHEN the `refs` subcommand is invoked with the
  `--json` option, the CLI SHALL emit each returned reference with its
  file path and span.
  - *Example:* An agent pipes the result into a rewrite, opening only
    the lines that carry the reference rather than each citing document.
- **AC-SPECREF-003.4:** The `refs` subcommand SHALL answer from the warm
  index within the budget AC-SPEC-009.1 sets for a corpus query.
- **AC-SPECREF-003.5:** WHEN no corpus member references the named slug,
  the CLI SHALL report that and exit zero.
  - *Example:* `spec refs frd-spec-reference-integrity` returning
    nothing is the answer that a document is unreferenced, not a
    failure — the same contract `slug-exists` sets in AC-SPEC-007.2.

*Rationale:* This extends the command surface FEAT-SPEC-011 fixed without
renaming or repurposing any of its four subcommands, which is the growth
that requirement anticipated. Both directions are exposed because the
index carries both (AC-SPECREF-001.3) and the two questions are asked at
different moments — "what breaks if I retire this" before a change,
"what does this depend on" before an edit. Answering either by reading
the corpus is the behavior the tool exists to replace.

### FEAT-SPECREF-004: Reference Configuration

**User Story:** As an adopting team, I want to declare which of my
frontmatter properties hold references and which form pairs, so that the
tool checks my conventions rather than assuming a fixed set.

**Acceptance Criteria:**

- **AC-SPECREF-004.1:** The configuration SHALL determine which
  frontmatter properties are reference-bearing and which pairs of
  properties form a paired relationship, extending the configuration
  file of FEAT-SPEC-008 without a separate file.
  - *Example:* `reference_properties` and `pairings` are already
    declared per document type in [[contract-specs-configuration]] and
    in `schemas/profile.strata`; this feature is the first consumer of
    properties the configuration schema already carries. It adds one:
    `typed_reference_properties`, mapping a reference-bearing property
    to the document `type` its target must declare (AC-SPECREF-002.10).
- **AC-SPECREF-004.2:** The CLI SHALL ship a configuration registering
  this framework's own reference-bearing properties and pairings, which
  an adopting repository can use without authoring one.
  - *Example:* The shipped configuration marks `superseded_by`,
    `sources`, `frd`, `system`, `contained_by`,
    `contracts_provided` and `contracts_consumed` as reference-bearing,
    and registers **no** pairing — every relationship this framework
    declares has its reverse derived from inbound edges, and
    [[document-relationship-standards]] reserves a pairing for the case
    that does not. It registers one typed reference: a feature
    blueprint's `frd` must resolve to a document of type `frd`.
- **AC-SPECREF-004.3:** IF a pairing names a property that is not
  registered as reference-bearing, THEN the CLI SHALL report the
  offending key and its position and SHALL NOT fall back to a default.
  - *Example:* A pairing of `children` with `parent` where only
    `parent_slug` is registered names the key and line, rather than
    silently checking nothing — the vacuous-success failure this
    project has now observed in three independently built tools.

*Rationale:* Configuration is what makes the checks in FEAT-SPECREF-002
belong to the adopting team rather than to this framework. A team whose
documents relate through `depends_on` and `required_by` gets the same
paired-relationship check without this repository knowing those names
exist. AC-SPECREF-004.3 exists because a misconfigured pairing is
indistinguishable from a passing one: the check runs, finds no
registered property to compare, and reports success.

*Note on numbering:* this feature's requirements carry the `SPECREF`
prefix rather than continuing [[frd-spec-validation-cli]]'s sequence.
FEAT-SPEC-004, the number reference integrity held before the split, is
left unused in that document rather than reclaimed here, so the ADRs
citing FEAT-SPEC-006, -008 and -009 keep resolving. The gap there and the
new prefix here are the same decision. AC-SPECREF-002.4 is unused for
the same reason: it required a retired document to declare a successor,
which is not true of every retirement, and it is removed rather than
renumbered.

## References

- [[frd-spec-validation-cli]] — the index, finding engine, corpus
  discovery, and configuration this feature extends
- [[frd-spec-anchor-references]] — the sub-feature that resolves the
  anchor half of a reference this feature resolves the slug half of
- [[frd-spec-slug-rename]] — the sub-feature that rewrites every
  reference to a slug when the slug is deliberately renamed
- [[frd-spec-structure-conformance]] — supplies the document headings
  the anchor sub-feature resolves an anchor against
- [[contract-specs-configuration]] — the configuration surface
  FEAT-SPECREF-004 consumes, whose `reference_properties` and `pairings`
  are already declared
- [[adr-012]] — the `[[slug]]` convention this feature parses and
  resolves, and the heading-anchor convention it defers
- [[adr-002]] — EARS, the notation these acceptance criteria are written
  in
- [[module-document-core]] — declares the References component this
  feature builds, and records that it is not yet built
- [[module-spec-profile]] — declares the Lifecycle Rules component whose
  relational half AC-SPECREF-002.6 supplies
- `plugin/skills/spec/references/cross-document-references.md` — what a
  reference resolves against
