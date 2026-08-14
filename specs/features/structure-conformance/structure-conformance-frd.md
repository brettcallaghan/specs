---
title: "FRD: Spec Corpus Structure and Hygiene Conformance"
description: Extracts document headings, section spans, and record blocks
  into the shared index and checks required sections, line width, and
  leftover authoring scaffolding, so an unfinished document cannot reach
  review looking finished
slug: frd-spec-structure-conformance
type: frd
status: draft
updated_at: 2026-08-04T00:00:00Z
sources: []
---

# Spec Corpus Structure and Hygiene Conformance

## Overview

Specification templates prescribe a document's structure: required
sections, structured record blocks describing declared elements, a line
width, and no leftover authoring scaffolding once a document is
finished. None of that is checked mechanically today, so a document that
skipped a section or still carries its template's instruction blocks can
reach review looking finished.

This feature adds structure awareness to the tool established by
[[frd-spec-validation-cli]]. It walks each document's headings and body,
records the headings with their section spans and the parsed record
blocks in the shared index, and reports documents that omit a required
section, exceed the configured line width, or retain placeholder or
scaffolding text. Like [[frd-spec-reference-integrity]], it introduces
no new artifact or engine: it populates the structure properties of the
index defined in FEAT-SPEC-002 and emits findings through the engine
defined in FEAT-SPEC-006. Corpus discovery, membership, output location,
and the warm index are inherited unchanged and are not restated here.

**Out of scope for this feature.** Judging whether a section says the
right thing — whether a requirement is testable, whether a design
carries enough intent — stays with a reviewing agent; this feature
extracts the spans the agent reads and checks that the sections exist,
not what they contain. Reference resolution is specified in
[[frd-spec-reference-integrity]]; anchor resolution against the headings
this feature extracts is specified in its sub-feature,
[[frd-spec-anchor-references]].

## Terminology

- **Section span** — the span of the content beneath a heading, from the
  end of the heading line to the start of the next heading of the same
  or higher level.
- **Record block** — a fenced code block carrying structured data about
  an element declared inside a document, per [[adr-008]].
- **Placeholder pattern** — a configured textual pattern that marks
  unfilled template content (for example `<Feature Name>`), expected to
  be absent once a document reaches a finished status.

## Requirements

### FEAT-SPECSTR-001: Structure Extraction in the Index

**User Story:** As an architecture owner and a reviewing agent, I want
each document's headings, section spans, and record blocks recorded in
the index, so that section-level tooling and targeted review read them
from the shared artifact instead of re-parsing the corpus.

**Acceptance Criteria:**

- **AC-SPECSTR-001.1:** WHEN the CLI builds the index, the CLI SHALL add
  to every corpus member's entry its headings, each with the span of the
  content beneath it.
  - *Example:* `adr-011`'s entry gains the spans of its Context,
    Decision, Alternatives Considered, Consequences and References
    headings.
- **AC-SPECSTR-001.2:** WHEN the CLI encounters a fenced block whose
  info string matches a tag the active profile registers, the CLI SHALL
  parse its content as the configured format and attach the parsed data
  to the containing document and heading in the index.
  - *Example:* A profile registering the tag `component` with format
    `yaml` causes each component record block in a container blueprint to
    appear in the index as structured data under the heading that
    declares it.

### FEAT-SPECSTR-002: Required Sections and Hygiene

**User Story:** As a reviewer, I want documents checked for skipped
sections and leftover authoring scaffolding, so that an unfinished
document cannot reach review looking finished.

**Acceptance Criteria:**

- **AC-SPECSTR-002.1:** WHEN the active profile registers required
  headings for a document's `type` and a document omits one, the CLI
  SHALL report a finding naming the missing heading.
  - *Example:* A document declaring `type: adr` with no "Alternatives
    Considered" heading produces a finding naming it.
- **AC-SPECSTR-002.2:** WHEN a document contains an HTML comment block
  outside a fenced code block and outside an inline code span, the CLI
  SHALL report a finding with its span.
  - *Example:* A finished ADR still carrying its template's Context
    instruction block is reported at the position the block opens, while
    a module blueprint whose prose describes the `<!--` delimiter inside
    backticks produces no finding.
- **AC-SPECSTR-002.3:** WHEN a prose line exceeds the configured width,
  the CLI SHALL report a finding naming the line and its length.
  - *Example:* A line of 85 characters is reported when the configured
    width is 80.
- **AC-SPECSTR-002.4:** The CLI SHALL measure line width in Unicode
  codepoints.
  - *Example:* A line containing an em dash counts that dash as one
    character, not as its three UTF-8 bytes.
- **AC-SPECSTR-002.5:** The CLI SHALL NOT report a width finding for a
  line inside a fenced code block, a table row, or a line containing a
  URL.
  - *Example:* An 88-character table row and a 120-character line whose
    length is a single URL both pass.
- **AC-SPECSTR-002.6:** WHEN a document whose `status` the active profile
  marks as finished contains a configured placeholder pattern, the CLI
  SHALL report a finding.
  - *Example:* A document at `status: approved` containing
    `<Feature Name>` is reported; the same text in a `draft` document is
    not.
- **AC-SPECSTR-002.7:** WHEN a registered record block fails to parse as
  its configured format, the CLI SHALL report a finding with the parse
  error and its position within the document.
  - *Example:* A `component` block containing malformed YAML is reported
    at the line the block opens, with the parser's message.

### FEAT-SPECSTR-003: Structure Queries

**User Story:** As a reviewing agent, I want to request the spans of
headings matching a selection, so that I review only the normative text
of many documents instead of reading them whole, and the review is
auditable.

**Acceptance Criteria:**

- **AC-SPECSTR-003.1:** WHEN asked for the spans of headings matching a
  configured selection, the CLI SHALL return each match as a file path,
  heading text, and the span of the content beneath it.
  - *Example:* A reviewing agent requests the normative sections of
    eleven standards documents and receives spans totalling a fraction
    of those documents' text, then reviews only that text.

*Rationale:* AC-SPECSTR-003.1 exists to make agent review auditable as
well as cheap. Two runs that review the same extracted spans have
reviewed the same text; two agents told to "review the requirements"
have not.

### FEAT-SPECSTR-004: Structure Configuration

**User Story:** As an adopting team, I want to declare my required
sections, my record-block tags, and my line width, so that the tool
enforces my templates rather than the ones it shipped with.

**Acceptance Criteria:**

- **AC-SPECSTR-004.1:** The configuration SHALL determine, for each
  registered document type, its required headings, the registered
  record-block tags with their formats, the placeholder patterns, and
  the line-width limit, extending the configuration file of FEAT-SPEC-008
  without a separate file.
  - *Example:* The shipped configuration sets a width of 80, registers
    the `component` tag with format `yaml`, and lists the required
    headings for each of this framework's document types.

## References

- [[frd-spec-validation-cli]] — the index, finding engine, corpus
  discovery, and configuration this feature extends
- [[frd-spec-anchor-references]] — consumes the headings this feature
  extracts to resolve anchor references
- [[frd-spec-reference-integrity]] — resolves the slug half of a
  reference, and defers the anchor half to that sub-feature
- [[adr-008]] — the element record blocks this feature extracts and
  validates
- [[adr-002]] — EARS, the notation these acceptance criteria are written
  in
- `plugin/skills/spec/references/template-authoring-standards.md` —
  frontmatter properties, slug conventions, line width, and required
  sections
- `plugin/skills/spec/references/document-readiness-checklist.md` — the
  structure checks whose mechanical subset this feature absorbs
