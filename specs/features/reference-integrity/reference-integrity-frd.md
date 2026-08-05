---
title: "FRD: Spec Corpus Reference Integrity"
description: Parses cross-document references from a spec corpus, records
  them in the shared index, and reports links that resolve to nothing, so
  a reference that silently broke is caught before a reader follows it
slug: frd-spec-reference-integrity
type: frd
status: draft
updated_at: 2026-08-05T00:00:00Z
sources: []
related:
  - "frd-spec-validation-cli"
  - "frd-spec-structure-conformance"
  - "contract-specs-configuration"
  - "adr-012"
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
and a scan of the same corpus on 2026-08-05 — 27 documents, 208
references — finds none unresolved. That is the argument for this
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
in REQ-SPEC-002 and emits findings through the engine defined in
REQ-SPEC-006. Corpus discovery, membership, output location, and the
warm index are inherited unchanged from that FRD and are not restated
here.

That FRD's ordering governs this one. The index records what a document
references, including references that resolve to nothing, and never
fails; reporting an unresolved reference is the checker's job. A
repository whose links are broken still contributes its graph to an
organization's catalogue.

**Out of scope for this feature.** Each of these is deferred
deliberately, not overlooked:

- **Heading extraction and section spans.** Specified in
  [[frd-spec-structure-conformance]]. This feature consumes a target
  document's headings only to resolve an anchor reference against them,
  and depends on that FRD for the heading data. Until that data exists,
  an anchor check reports `unverified` rather than passing
  (AC-SPECREF-002.2).
- **Suggesting a correction for a broken reference.** Findings state
  the fact and leave the fix to a reviewing agent.
- **Repairing or rewriting references.** This feature only reads and
  reports, as REQ-SPEC-002's ordering requires.
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
- **Rendering the reference graph**, and export to any third-party
  format.

## Terminology

- **Body** — the content of a document after its frontmatter block.
  Reference extraction reads the body; frontmatter contributes
  references only through the properties the profile registers as
  reference-bearing, never through prose in a frontmatter value.
- **Outbound reference** — a `[[slug]]` link, optionally naming a
  heading anchor, parsed from a document's body, or a slug value in a
  frontmatter property the active profile registers as reference-bearing.
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
  `related` or `superseded_by`).
- **Paired relationship** — two reference-bearing properties the profile
  declares as counterparts, so that a reference in one document implies a
  matching reference back (for example `children` and `parent_slug`).

## Requirements

### REQ-SPECREF-001: Reference Extraction in the Index

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
    REQ-SPEC-002 sets — indexing never fails, validation may.

*Rationale:* AC-SPECREF-001.2 and AC-SPECREF-001.4 are the same rule
applied to the two places a reference is written about rather than
written: a backticked example, and a frontmatter value describing the
notation. Both were observed in this corpus, and both are false
positives a reader would have to dismiss by hand — the failure mode that
makes a checker get ignored. AC-SPECREF-001.6 restates REQ-SPEC-002's
ordering for this feature specifically, because a reference graph is the
part of the index most likely to tempt an implementer into failing the
build from the indexer.

### REQ-SPECREF-002: Reference Integrity Checks

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
- **AC-SPECREF-002.2:** WHEN a reference names a heading anchor and the
  headings of the target document are available, the CLI SHALL report a
  finding of severity `error` naming the target document and the missing
  anchor IF no heading matches it.
  - *Example:* `[[container-spec-cli#Componets]]` reports `no heading
    named 'Componets' found in container-spec-cli`. The finding states
    the fact rather than proposing a correction — an agent resolves a
    plain "not found" faster than a guess it must then verify.
- **AC-SPECREF-002.3:** IF a reference names a heading anchor and the
  headings of the target document are not available, THEN the CLI SHALL
  report that reference with severity `unverified` naming what is
  missing, and SHALL NOT report the anchor as resolved.
  - *Example:* Before [[frd-spec-structure-conformance]] ships there is
    no heading data at all, so every anchor reference in the corpus is
    reported `unverified`, per AC-SPEC-006.3 and AC-SPEC-006.4. The same
    path covers a target whose own frontmatter failed to parse, whose
    headings were therefore never extracted.
- **AC-SPECREF-002.4:** WHEN a frontmatter property registered by the
  active profile as reference-bearing contains a value matching no
  corpus member's slug, the CLI SHALL report a finding of severity
  `error` naming the property and the value.
  - *Example:* `related: ["work-order-template"]` reports as unresolved,
    because templates carry placeholder slugs and are not corpus
    members.
- **AC-SPECREF-002.5:** WHEN a document declares `status: retired` and
  either declares no `superseded_by` or declares one that resolves to no
  corpus member, the CLI SHALL report a finding of severity `error`
  distinguishing the two cases.
  - *Example:* A retired ADR with no `superseded_by` reports "retired
    with no superseded_by"; one whose `superseded_by` names a slug no
    corpus member declares reports "superseded_by 'adr-020' matches no
    document's slug". A reader is never sent to a replacement that does
    not exist, and never left to guess whether one was intended.
- **AC-SPECREF-002.6:** WHEN the active profile registers a frontmatter
  property as one half of a paired relationship and the named document
  does not declare the matching counterpart, the CLI SHALL report a
  finding of severity `error` against both documents.
  - *Example:* A blueprint declaring `children: [component-ledger]`
    where `component-ledger` declares no `parent_slug` produces a
    finding on each.
- **AC-SPECREF-002.7:** WHEN a document at a status the profile treats
  as live references a document at `status: retired`, the CLI SHALL
  report a finding of severity `warning` naming the retired target and
  its `superseded_by` value when it has one.
  - *Example:* An approved blueprint citing a retired ADR reports
    `reference/retired-target  warning … "adr-009 is retired,
    superseded by adr-013"`, which does not fail a build but is the one
    finding that tells an author what to cite instead.
- **AC-SPECREF-002.8:** WHEN a reference's slug matches a file the index
  recorded as excluded or as a parse failure rather than a corpus
  member, the CLI SHALL report the finding with that reason rather than
  as a slug matching nothing.
  - *Example:* A reference to a document whose frontmatter does not
    parse reports "target exists but is not a corpus member: parse
    failure at line 28" — one lookup for the reader instead of a search
    for a document that is right there.
- **AC-SPECREF-002.9:** The CLI SHALL resolve a reference against corpus
  slugs without regard to letter case.
  - *Example:* `[[ADR-012]]` resolves to the document declaring
    `slug: adr-012`, matching the case-insensitive slug uniqueness
    AC-SPEC-003.4 already establishes — the two rules must agree, or a
    corpus can hold a reference that resolves to two documents that
    cannot both exist.

*Rationale:* Severity is assigned here rather than left to
implementation because it is what the exit status of a build depends on.
AC-SPECREF-002.7 is a `warning` because a reference to a retired
document still resolves — the reader reaches a real document that tells
them it has been replaced — so failing a build on it would punish a
corpus for retiring anything. It is also the first check in this tool to
emit `warning` at all: AC-SPEC-006.7's `--strict` boundary has until now
been unreachable end to end, because every registered check emitted
`error`. AC-SPECREF-002.3 is the same argument in the other direction:
an anchor check that cannot see headings must say so rather than pass,
and the `unverified` severity of AC-SPEC-006.3 exists precisely for
this. It also means this feature does not block on
[[frd-spec-structure-conformance]] — it ships with the anchor check
honest about what it cannot yet do, and gains resolution when heading
data arrives.

### REQ-SPECREF-003: Reference Queries

**User Story:** As an authoring agent, I want to ask which documents
reference a slug, so that I understand the blast radius of retiring or
renaming a document without reading every file.

**Acceptance Criteria:**

- **AC-SPECREF-003.1:** The CLI SHALL expose a `refs` subcommand that,
  given a slug, returns every corpus member whose body or registered
  frontmatter properties reference it.
  - *Example:* `spec refs adr-007` returns the seven documents citing
    it, so an author retiring it knows the blast radius.
- **AC-SPECREF-003.2:** WHEN the `refs` subcommand is invoked with an
  option selecting the outbound direction, the CLI SHALL return the
  documents the named slug references instead of those referencing it.
  - *Example:* `spec refs frd-spec-reference-integrity --outbound`
    lists the four documents this FRD cites, so an agent about to edit
    it knows what it depends on without reading it.
- **AC-SPECREF-003.3:** WHEN the `refs` subcommand is invoked with the
  `--json` option, the CLI SHALL emit each returned reference with its
  file path and span.
  - *Example:* An agent pipes the result into a rewrite, opening only
    the lines that carry the reference rather than each citing document.
- **AC-SPECREF-003.4:** The `refs` subcommand SHALL answer from the warm
  index within the budget AC-SPEC-009.1 sets for a corpus query.

*Rationale:* This extends the command surface REQ-SPEC-011 fixed without
renaming or repurposing any of its four subcommands, which is the growth
that requirement anticipated. Both directions are exposed because the
index carries both (AC-SPECREF-001.3) and the two questions are asked at
different moments — "what breaks if I retire this" before a change,
"what does this depend on" before an edit. Answering either by reading
the corpus is the behavior the tool exists to replace.

### REQ-SPECREF-004: Reference Configuration

**User Story:** As an adopting team, I want to declare which of my
frontmatter properties hold references and which form pairs, so that the
tool checks my conventions rather than assuming a fixed set.

**Acceptance Criteria:**

- **AC-SPECREF-004.1:** The configuration SHALL determine which
  frontmatter properties are reference-bearing and which pairs of
  properties form a paired relationship, extending the configuration
  file of REQ-SPEC-008 without a separate file.
  - *Example:* `reference_properties` and `pairings` are already
    declared per document type in [[contract-specs-configuration]] and
    in `schemas/profile.strata`; this feature is the first consumer of
    properties the configuration schema already carries, not a change to
    it.
- **AC-SPECREF-004.2:** The CLI SHALL ship a configuration registering
  this framework's own reference-bearing properties and pairings, which
  an adopting repository can use without authoring one.
  - *Example:* The shipped configuration marks `related`,
    `superseded_by`, `children` and `parent_slug` as reference-bearing
    and pairs `children` with `parent_slug`, mirroring AC-SPEC-008.6.
- **AC-SPECREF-004.3:** IF a pairing names a property that is not
  registered as reference-bearing, THEN the CLI SHALL report the
  offending key and its position and SHALL NOT fall back to a default.
  - *Example:* A pairing of `children` with `parent` where only
    `parent_slug` is registered names the key and line, rather than
    silently checking nothing — the vacuous-success failure this
    project has now observed in three independently built tools.

*Rationale:* Configuration is what makes the checks in REQ-SPECREF-002
belong to the adopting team rather than to this framework. A team whose
documents relate through `depends_on` and `required_by` gets the same
paired-relationship check without this repository knowing those names
exist. AC-SPECREF-004.3 exists because a misconfigured pairing is
indistinguishable from a passing one: the check runs, finds no
registered property to compare, and reports success.

*Note on numbering:* this feature's requirements carry the `SPECREF`
prefix rather than continuing [[frd-spec-validation-cli]]'s sequence.
REQ-SPEC-004, the number reference integrity held before the split, is
left unused in that document rather than reclaimed here, so the ADRs
citing REQ-SPEC-006, -008 and -009 keep resolving. The gap there and the
new prefix here are the same decision.

## References

- [[frd-spec-validation-cli]] — the index, finding engine, corpus
  discovery, and configuration this feature extends
- [[frd-spec-structure-conformance]] — supplies the target-document
  headings that AC-SPECREF-002.2 resolves an anchor against
- [[contract-specs-configuration]] — the configuration surface
  REQ-SPECREF-004 consumes, whose `reference_properties` and `pairings`
  are already declared
- [[adr-012]] — the `[[slug]]` and heading-anchor conventions this
  feature parses and resolves
- [[adr-002]] — EARS, the notation these acceptance criteria are written
  in
- [[module-document-core]] — declares the References component this
  feature builds, and records that it is not yet built
- [[module-spec-profile]] — declares the Lifecycle Rules component whose
  relational half AC-SPECREF-002.5 and AC-SPECREF-002.7 supply
- `plugin/skills/spec/references/cross-document-references.md` — what a
  reference resolves against
