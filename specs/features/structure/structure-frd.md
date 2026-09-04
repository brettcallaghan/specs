---
title: "FRD: Spec Corpus Structure Extraction"
description: Records each document's headings with their section spans
  and its parsed record blocks in the shared index, and answers a
  heading query, so section-level tooling reads structure rather than
  re-parsing the corpus
slug: frd-spec-structure
type: frd
status: draft
updated_at: 2026-09-04T00:00:00Z
sources:
  - "frd-spec-index"
---

# Spec Corpus Structure Extraction

## Overview

A document's structure is data: its headings, the span of content
beneath each, and the record blocks that declare elements inside it.
Nothing extracts any of it today, so every capability that needs to
address a place inside a document — rather than the document as a
whole — has nothing to address it with.

This feature produces that data. It walks each corpus member, records
its headings with their section spans and its parsed record blocks in
the index [[frd-spec-index]] defines, and answers a query for the spans
of headings matching a selection. It introduces no new artifact and no
new engine, and it judges nothing: whether a document omits a required
section, exceeds a line width or still carries its template's
scaffolding is [[frd-spec-structure-conformance]]'s question, asked of
the data this feature records.

That division is what lets the two ship apart, and it matters now
because a consumer is waiting. [[frd-spec-anchor-references]] resolves
the anchor half of a wikilink against a target document's headings, and
the corpus already carries nine such references that nothing can check
until the headings exist. Anchor resolution needs the data and none of
the judgment, so binding it to the conformance checks would make a
reference check wait on a line-width rule.

The heading half of this feature needs no configuration, which is why it
can land first: a heading is markdown, not a registered vocabulary. The
record-block half does need the tag registration
[[frd-spec-structure-conformance]] specifies, and waits for it.

**Out of scope for this feature.** Each of these is deferred
deliberately, not overlooked:

- **Required sections, hygiene and line width.** Specified in
  [[frd-spec-structure-conformance]], which checks documents against the
  structure this feature records.
- **The structure vocabulary a project configures.** Required headings,
  record-block tags, placeholder patterns and the line-width limit are
  declared through the configuration [[frd-spec-structure-conformance]]
  specifies.
- **Resolving an anchor against a heading.** Specified in
  [[frd-spec-anchor-references]], which consumes what this feature
  records.
- **Judging what a section says.** Whether a requirement is testable or
  a design carries enough intent stays with a reviewing agent. This
  feature extracts the spans that agent reads.

## Terminology

- **Section span** — the span of the content beneath a heading, from the
  end of the heading line to the start of the next heading of the same
  or higher level.
- **Record block** — a fenced code block carrying structured data about
  an element declared inside a document, per [[adr-008]].

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

*Note:* AC-SPECSTR-001.1 stands alone — a heading is markdown, and its
span is computed from the document. AC-SPECSTR-001.2 waits on the
record-block tag registration [[frd-spec-structure-conformance]]
specifies, because a block is only a record block where the active
profile says its info string is one.

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

## References

- [[frd-spec-index]] — the index, finding engine, corpus discovery and
  configuration this feature extends
- [[frd-spec-structure-conformance]] — checks documents against the
  structure this feature records, and registers the record-block tag
  vocabulary AC-SPECSTR-001.2 parses against
- [[frd-spec-anchor-references]] — resolves an anchor against the
  headings this feature records
- [[adr-008]] — the element record blocks this feature extracts
- [[adr-021]] — Gherkin, the notation these features and acceptance
  criteria are written in
- [[gherkin-guide]] — how to write them
