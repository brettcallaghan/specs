---
title: "FRD: Spec Corpus Structure and Hygiene Conformance"
description: Checks documents against the structure recorded in the
  shared index — required sections, line width, and leftover authoring
  scaffolding — so an unfinished document cannot reach review looking
  finished
slug: frd-spec-structure-conformance
type: frd
status: draft
updated_at: 2026-10-06T00:00:00Z
sources:
  - "frd-spec-structure"
---

# Spec Corpus Structure and Hygiene Conformance

## Overview

Specification templates prescribe a document's structure: required
sections, structured record blocks describing declared elements, a line
width, and no leftover authoring scaffolding once a document is
finished. None of that is checked mechanically today, so a document that
skipped a section or still carries its template's instruction blocks can
reach review looking finished.

This feature judges documents against the structure
[[frd-spec-structure]] records. It reports documents that omit a
required section, exceed the configured line width, or retain
placeholder or scaffolding text, and it declares the vocabulary those
checks read: the required headings of each document type, the
record-block tags, the placeholder patterns and the width limit. Like
[[frd-spec-reference-integrity]], it introduces no new artifact or
engine: it reads the structure properties of the index defined in
FEAT-SPEC-002 and emits findings through the engine defined in
FEAT-SPEC-006. Corpus discovery, membership, output location, and the
warm index are inherited unchanged and are not restated here.

**Out of scope for this feature.** Extracting the headings, section
spans and record blocks these checks read is specified in
[[frd-spec-structure]], which is a separate feature because the data has
a consumer that needs none of the judgment here: anchor resolution, in
[[frd-spec-anchor-references]]. Judging whether a section says the right
thing — whether a requirement is testable, whether a design carries
enough intent — stays with a reviewing agent; this feature checks that
the sections exist, not what they contain. Reference resolution is
specified in [[frd-spec-reference-integrity]].

## Terminology

- **Record block** — a fenced code block whose info string the active
  profile registers, carrying a YAML mapping about an element declared
  inside a document, per [[adr-008]]. The registration may name a JSON
  Schema the mapping must satisfy.
- **Placeholder pattern** — a configured textual pattern that marks
  unfilled template content (for example `<Feature Name>`), expected to
  be absent once a document reaches a finished status.

## Requirements

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
  Scenario: A record block that is not valid YAML is reported where it opens
    Given a registered record block whose content is not valid YAML
    When the CLI checks the document
    Then the CLI reports a finding carrying the parse error and its position within the document

  @AC-SPECSTR-002.8
  Scenario Outline: A record block that is not a mapping is reported
    Given a registered record block whose content is valid YAML holding "<value>"
    When the CLI checks the document
    Then the CLI reports a finding at the position the block opens, naming the kind of value found

    Examples:
      | value      |
      | a sequence |
      | a scalar   |
      | nothing    |

  @AC-SPECSTR-002.9
  Scenario: A record block that violates its registered schema is reported
    Given a tag registered with a JSON Schema
    And a record block of that tag whose mapping violates the schema
    When the CLI checks the document
    Then the CLI reports a finding at the position the block opens, naming the violated constraint and the property that violated it

  @AC-SPECSTR-002.10
  Scenario: A record block of a tag registered without a schema is checked only as a mapping
    Given a tag registered with no JSON Schema
    And a record block of that tag holding any valid YAML mapping
    When the CLI checks the document
    Then the CLI reports no record-block finding for it

  @AC-SPECSTR-002.11
  Scenario: Declarations of one element that disagree are reported
    Given an element declared by its blueprint and by a record block with the same slug
    And the two declarations give different values for one of entity_name, description or technology
    When the CLI checks the corpus
    Then the CLI reports a finding at the record block, naming the property and the blueprint's value
```

*Note:* AC-SPECSTR-002.6 reports the placeholder only at a finished
status — the same `<Feature Name>` text in a `draft` document is
unfinished work rather than a defect.

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
    Then the configuration determines, for each registered document type, its required headings, the registered record-block tags with their optional JSON Schemas, the placeholder patterns, and the line-width limit
    And it does so by extending the configuration file of FEAT-SPEC-008 rather than by a separate file

  @AC-SPECSTR-004.2
  Scenario: The shipped profile registers component record blocks against the element record schema
    Given the specs profile is active
    When the CLI checks a document holding a component record block
    Then the block is validated against the ElementRecord schema, so a block missing its slug is reported
```

## References

- [[frd-spec-structure]] — records the headings, section spans and
  record blocks these checks read
- [[frd-spec-index]] — the index, finding engine, corpus
  discovery, and configuration this feature extends
- [[frd-spec-anchor-references]] — resolves anchors against the headings
  [[frd-spec-structure]] records
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
