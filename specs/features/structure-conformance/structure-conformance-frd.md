---
title: "FRD: Spec Corpus Structure and Hygiene Conformance"
description: Extracts document headings, section spans, and record blocks
  into the shared index and checks required sections, line width, and
  leftover authoring scaffolding, so an unfinished document cannot reach
  review looking finished
slug: frd-spec-structure-conformance
type: frd
status: draft
updated_at: 2026-08-15T00:00:00Z
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
[[frd-spec-index]]. It walks each document's headings and body,
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

**Acceptance Criteria:**

```gherkin
@FEAT-SPECSTR-001
Feature: Structure Extraction in the Index

  As an architecture owner and a reviewing agent
  I want each document's headings, section spans and record blocks recorded in the index
  So that section-level tooling and targeted review read them from the shared artifact instead of re-parsing the corpus

  This requirement populates the structure properties of the index
  FEAT-SPEC-002 defines and reports through the finding engine
  FEAT-SPEC-006 defines. Corpus discovery, membership, output location and
  the warm index are inherited unchanged and are not restated.

  @AC-SPECSTR-001.1
  Scenario: Every corpus member's headings are recorded with their section spans
    When the CLI builds the index
    Then every corpus member's entry carries its headings, each with the span of the content beneath it

  @AC-SPECSTR-001.2
  Scenario: A registered record block is parsed and attached to its document and heading
    Given a fenced block whose info string matches a tag the active profile registers
    When the CLI builds the index
    Then the CLI parses that block's content as the format configured for the tag
    And the CLI attaches the parsed data to the containing document and heading in the index
```

### FEAT-SPECSTR-002: Required Sections and Hygiene

**Acceptance Criteria:**

```gherkin
@FEAT-SPECSTR-002
Feature: Required Sections and Hygiene

  As a reviewer
  I want documents checked for skipped sections and leftover authoring scaffolding
  So that an unfinished document cannot reach review looking finished

  A width finding is about prose. Code, tables and URLs are text whose
  width the author does not control, so the limit does not reach them.

  @AC-SPECSTR-002.1
  Scenario: A document omitting a required heading is reported
    Given the active profile registers required headings for a document's type
    And the document omits one of them
    When the CLI checks the document
    Then the CLI reports a finding naming the missing heading

  @AC-SPECSTR-002.2
  Scenario: An HTML comment block surviving in a document is reported
    Given a document containing an HTML comment block outside a fenced code block and outside an inline code span
    When the CLI checks the document
    Then the CLI reports a finding with the span of that block

  @AC-SPECSTR-002.3
  Scenario: A prose line over the configured width is reported
    Given a configured line width of 80 characters
    And a prose line of 85 characters
    When the CLI checks the document
    Then the CLI reports a finding naming the line and its length

  @AC-SPECSTR-002.4
  Scenario: Line width is measured in Unicode codepoints
    Given a prose line containing an em dash
    Then the CLI counts that dash as one character rather than as its three UTF-8 bytes

  @AC-SPECSTR-002.5 @negative
  Scenario Outline: Lines the width limit does not reach are not reported
    Given a line that exceeds the configured width and is "<kind>"
    When the CLI checks the document
    Then the CLI does not report a width finding for that line

    Examples:
      | kind                                     |
      | a line inside a fenced code block        |
      | an 88-character table row                |
      | a 120-character line that is a single URL |

  @AC-SPECSTR-002.6
  Scenario: A placeholder surviving into a finished document is reported
    Given a document whose status the active profile marks as finished
    And the document contains a configured placeholder pattern
    When the CLI checks the document
    Then the CLI reports a finding

  @AC-SPECSTR-002.7
  Scenario: A record block that does not parse is reported where it opens
    Given a registered record block that fails to parse as its configured format
    When the CLI checks the document
    Then the CLI reports a finding carrying the parse error and its position within the document
```

*Note:* AC-SPECSTR-002.6 reports the placeholder only at a finished
status — the same `<Feature Name>` text in a `draft` document is
unfinished work rather than a defect.

### FEAT-SPECSTR-003: Structure Queries

**Acceptance Criteria:**

```gherkin
@FEAT-SPECSTR-003
Feature: Structure Queries

  As a reviewing agent
  I want to request the spans of headings matching a selection
  So that I review only the normative text of many documents instead of reading them whole

  This query is what makes agent review auditable as well as cheap. Two
  runs that review the same extracted spans have reviewed the same text;
  two agents told to review the requirements have not.

  @AC-SPECSTR-003.1
  Scenario: A heading selection returns each match with its section span
    When the CLI is asked for the spans of headings matching a configured selection
    Then the CLI returns each match as a file path, heading text, and the span of the content beneath it
```

### FEAT-SPECSTR-004: Structure Configuration

**Acceptance Criteria:**

```gherkin
@FEAT-SPECSTR-004
Feature: Structure Configuration

  As an adopting team
  I want to declare my required sections, my record-block tags and my line width
  So that the tool enforces my templates rather than the ones it shipped with

  @AC-SPECSTR-004.1
  Scenario: One configuration file carries the structure vocabulary
    Then the configuration determines, for each registered document type, its required headings, the registered record-block tags with their formats, the placeholder patterns, and the line-width limit
    And it does so by extending the configuration file of FEAT-SPEC-008 rather than by a separate file
```

## References

- [[frd-spec-index]] — the index, finding engine, corpus
  discovery, and configuration this feature extends
- [[frd-spec-anchor-references]] — consumes the headings this feature
  extracts to resolve anchor references
- [[frd-spec-reference-integrity]] — resolves the slug half of a
  reference, and defers the anchor half to that sub-feature
- [[adr-008]] — the element record blocks this feature extracts and
  validates
- [[adr-021]] — Gherkin, the notation these features and acceptance
  criteria are written in
- [[gherkin-guide]] — how to write them
- `plugin/skills/spec/references/template-authoring-standards.md` —
  frontmatter properties, slug conventions, line width, and required
  sections
- `plugin/skills/spec/references/document-readiness-checklist.md` — the
  structure checks whose mechanical subset this feature absorbs
