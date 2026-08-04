---
title: "FRD: Spec Corpus Reference Integrity"
description: Parses cross-document references from a spec corpus, records
  them in the shared index, and reports links that resolve to nothing, so
  a reference that silently broke is caught before a reader follows it
slug: frd-spec-reference-integrity
type: frd
status: draft
updated_at: 2026-08-04T00:00:00Z
sources: []
related:
  - "frd-spec-validation-cli"
  - "frd-spec-structure-conformance"
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

**Out of scope for this feature.** Heading extraction and section spans
are specified in [[frd-spec-structure-conformance]]; this feature
consumes a target document's headings only to resolve an anchor
reference against them, and depends on that FRD for the heading data
when an anchor is checked. Suggesting a correction for a broken
reference is out of scope — findings state the fact and leave the fix to
a reviewing agent.

## Terminology

- **Outbound reference** — a `[[slug]]` link, optionally naming a
  heading anchor, parsed from a document's body, or a slug value in a
  frontmatter property the active profile registers as reference-bearing.
- **Inbound reference** — the reverse edge: for a given document, every
  other corpus member whose outbound references name it.
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

### REQ-SPECREF-002: Reference Integrity Checks

**User Story:** As a document author, I want every cross-reference
checked, so that a link that silently stopped resolving is caught before
a reader follows it into nothing.

**Acceptance Criteria:**

- **AC-SPECREF-002.1:** WHEN a document body contains a `[[slug]]`
  reference that matches no corpus member's slug, the CLI SHALL report a
  finding with the file, line, and unresolved slug.
  - *Example:* `[[adr-011]]` in `ADR-004` reports as unresolved while
    that ADR's frontmatter declares
    `slug: adr-011-targeted-verification-gates`.
- **AC-SPECREF-002.2:** WHEN a reference names a heading anchor and the
  target document has no heading matching it, the CLI SHALL report a
  finding naming the target document and the missing anchor.
  - *Example:* `[[container-api#Componets]]` reports `no heading named
    'Componets' found in container-api`. The finding states the fact
    rather than proposing a correction — an agent resolves a plain
    "not found" faster than a guess it must then verify.
- **AC-SPECREF-002.3:** WHEN a frontmatter property registered by the
  active profile as reference-bearing contains a value matching no
  corpus member's slug, the CLI SHALL report a finding naming the
  property and the value.
  - *Example:* `related: ["work-order-template"]` reports as unresolved,
    because templates carry placeholder slugs and are not corpus
    members.
- **AC-SPECREF-002.4:** WHEN a document declares `status: retired`
  without a resolvable `superseded_by` reference, the CLI SHALL report a
  finding.
  - *Example:* A retired ADR whose `superseded_by` names a slug no
    corpus member declares is reported, so a reader is never sent to a
    replacement that does not exist.
- **AC-SPECREF-002.5:** WHEN the active profile registers a frontmatter
  property as one half of a paired relationship and the named document
  does not declare the matching counterpart, the CLI SHALL report a
  finding against both documents.
  - *Example:* A blueprint declaring `children: [component-ledger]`
    where `component-ledger` declares no `parent_slug` produces a
    finding on each.

### REQ-SPECREF-003: Reference Queries

**User Story:** As an authoring agent, I want to ask which documents
reference a slug, so that I understand the blast radius of retiring or
renaming a document without reading every file.

**Acceptance Criteria:**

- **AC-SPECREF-003.1:** WHEN asked for the documents referencing a given
  slug, the CLI SHALL return every corpus member whose body or
  registered frontmatter properties reference it.
  - *Example:* Asked about `adr-007`, it returns the seven documents
    citing it, so an author retiring it knows the blast radius.

### REQ-SPECREF-004: Reference Configuration

**User Story:** As an adopting team, I want to declare which of my
frontmatter properties hold references and which form pairs, so that the
tool checks my conventions rather than assuming a fixed set.

**Acceptance Criteria:**

- **AC-SPECREF-004.1:** The configuration SHALL determine which
  frontmatter properties are reference-bearing and which pairs of
  properties form a paired relationship, extending the configuration
  file of REQ-SPEC-008 without a separate file.
  - *Example:* The shipped configuration marks `related`,
    `superseded_by`, `children` and `parent_slug` as reference-bearing
    and pairs `children` with `parent_slug`.

## References

- [[frd-spec-validation-cli]] — the index, finding engine, corpus
  discovery, and configuration this feature extends
- [[frd-spec-structure-conformance]] — supplies the target-document
  headings that AC-SPECREF-002.2 resolves an anchor against
- [[adr-012]] — the `[[slug]]` and heading-anchor conventions this
  feature parses and resolves
- [[adr-002]] — EARS, the notation these acceptance criteria are written
  in
- `plugin/skills/spec/references/cross-document-references.md` — what a
  reference resolves against
