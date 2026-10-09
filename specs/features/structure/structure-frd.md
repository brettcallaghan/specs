---
title: "Spec Corpus Structure Extraction"
description: Records each document's headings with their section spans
  and its parsed record blocks in the shared index, and answers a
  heading query, so section-level tooling reads structure rather than
  re-parsing the corpus
slug: frd-spec-structure
type: frd
status: approved
updated_at: 2026-10-08T18:00:00Z
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
  Scenario: A registered record block is recorded against its document and heading
    Given a fenced block whose info string matches a tag the active profile registers
    When the CLI builds the index
    Then the index records the block's tag and span against the containing document and the innermost heading whose section contains it
    And the index does not carry the block's content
```

*Note:* AC-SPECSTR-001.1 stands alone — a heading is markdown, and its
span is computed from the document. AC-SPECSTR-001.2 waits on the
record-block tag registration [[frd-spec-structure-conformance]]
specifies, because a block is only a record block where the active
profile says its info string is one. The index records where a record
is, not what it says: a command that needs the record decodes it through
the span, and [[frd-spec-structure-conformance]] checks that it is valid.

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
  Scenario: The full answer returns each match with its section span
    When the CLI is asked, with the full option, for the spans of headings matching a given selection
    Then the CLI returns each match as a file path, the heading text, its identifier, its "anchor", its "ref", its level, and the span of the content beneath it

  @AC-SPECSTR-003.2
  Scenario: A selection matches a heading on its identifier
    Given a selection naming a heading
    When the CLI runs the query
    Then the CLI matches the selection against each heading by identifier, so that letter case and punctuation do not decide a match

  @AC-SPECSTR-003.3
  Scenario: A selection may name one document or none
    Given a selection qualified by a document
    When the CLI runs the query
    Then the CLI returns matches from that document alone
    And the CLI answers an unqualified selection from every corpus member

  @AC-SPECSTR-003.4
  Scenario: A matched heading carries the headings nested beneath it
    Given a selection matching one heading
    When the CLI runs the query
    Then the CLI returns every heading within that heading's section, each with its own level

  @AC-SPECSTR-003.5
  Scenario: The content beneath a match is returned on request
    Given a selection matching exactly one heading
    When the CLI is asked to emit the content that heading stands over
    Then the CLI writes the text of its section, read from its recorded span, as it stands in the document and with nothing added

  @AC-SPECSTR-003.6
  Scenario: A selection matching no heading is an answer rather than a failure
    Given a selection matching no heading
    When the CLI runs the query without the emit option
    Then the CLI reports that
    And the CLI exits zero

  @AC-SPECSTR-003.7 @negative
  Scenario: Emitting content for a selection matching several headings is refused
    Given a selection matching more than one heading
    When the CLI is asked to emit the content those headings stand over
    Then the CLI writes nothing to standard output
    And the CLI names every match by its ref on standard error, so that the caller can narrow the selection to one ref or request the matches without their content
    And the CLI exits with the invocation-error status

  @AC-SPECSTR-003.8
  Scenario: A selection naming a document no corpus member declares is an answer
    Given a selection qualified by a slug or path that no corpus member declares
    When the CLI runs the query without the emit option
    Then the CLI returns an empty list of matches and a property stating that the document is not declared
    And the CLI exits zero

  @AC-SPECSTR-003.9
  Scenario: A match's anchor is its heading path
    Given a heading nested beneath other headings
    When the CLI reports it
    Then its "anchor" is its heading path: the identifier of each enclosing heading from the outermost down, followed by its own, joined by "#"

  @AC-SPECSTR-003.10
  Scenario Outline: A selection may name a heading by a path of headings
    Given a document with a heading "<heading>" nested beneath a heading "<ancestor>"
    And another heading "<heading>" elsewhere in that document
    When the CLI runs the query for "<selection>"
    Then the CLI matches only the heading nested beneath "<ancestor>"

    Examples:
      | ancestor   | heading    | selection                                  |
      | Components | References | module-document-core#Components#References |

  @AC-SPECSTR-003.11
  Scenario: Each segment of a heading path is nested within the one before it at any depth
    Given a selection whose heading half holds more than one segment separated by "#"
    When the CLI runs the query
    Then the CLI matches exactly the headings whose heading path is those segments, where some heading's is
    And otherwise the CLI matches a heading named by the last segment whose enclosing headings include, in order from the outermost, a heading named by each earlier segment
    And the CLI matches each segment on its identifier as AC-SPECSTR-003.2 states

  @AC-SPECSTR-003.12
  Scenario: Every match carries a ref that pastes back unchanged
    When the CLI reports a match
    Then the match carries a "ref" of the form "<slug>#<anchor>", where the slug is the one its document declares and is the document's repository-relative path where it declares none
    And that ref is accepted unchanged as a heading selection, as the argument of a reference query, and inside a wikilink

  @AC-SPECSTR-003.13 @negative
  Scenario Outline: Emitting content for a selection matching no heading is refused
    Given a selection "<condition>"
    When the CLI is asked to emit the content of the heading it names
    Then the CLI writes nothing to standard output
    And the CLI writes a message to standard error naming the selection and stating that "<reason>"
    And the CLI exits with the invocation-error status

    Examples:
      | condition                                          | reason                                       |
      | naming a declared document and a heading it lacks  | the selection matches no heading             |
      | qualified by a document no corpus member declares  | no corpus member is the document it names    |

  @AC-SPECSTR-003.14
  Scenario: Documents no longer in effect are omitted unless asked for
    Given a corpus member at a status the active profile registers as no longer in effect
    When the CLI runs a query, or a refused emit names its matches, without the all option
    Then the CLI omits that member's headings
    And the CLI includes them when invoked with the all option

  @AC-SPECSTR-003.15
  Scenario: A document named explicitly is answered whatever its status
    Given a selection whose document half names, by slug or path, a corpus member at a status no longer in effect
    When the CLI runs the query without the all option
    Then the CLI returns that member's matching headings

  @AC-SPECSTR-003.16
  Scenario: The default answer is the compact view
    When the CLI is asked for the headings matching a selection without the full option
    Then the CLI returns each match as its "anchor", its "ref", its "level" and its "start_line", the line its heading starts on, and nothing else
    And the compact answer holds every match the full answer holds, in the same order

  @AC-SPECSTR-003.17 @negative
  Scenario: The full option is refused with the emit option
    When the CLI is asked to emit the content of a heading with the full option
    Then the CLI writes nothing to standard output
    And the CLI exits with the invocation-error status
```

*Note:* AC-SPECSTR-003.4 is what answers a roster question — the
components declared beneath a `Components` heading — without the caller
knowing how many there are or what they are called. AC-SPECSTR-003.5 is
what keeps the index free of section content: the span is recorded, and
the text is read back through it on demand. It is written raw, because
the caller asked for the document's own words, and raw text has no place
to say where one section ends and the next begins, which is why
AC-SPECSTR-003.7 refuses a selection matching several headings rather
than joining them; the full answer already lists every match, each with
its ref and span.

The same absence of an envelope is why AC-SPECSTR-003.13 refuses an
emit that matches nothing. AC-SPECSTR-003.6 and AC-SPECSTR-003.8 make
zero matches an answer because a structured answer can say so — an empty
`matches` and, for a document half, `declared: false`. Emitted text can
say neither: an empty standard output reads the same for a heading that
is absent, a document that is undeclared and a section that is empty.
The emit form's contract is exactly one heading, so zero is refused as
several is, and the message says which of the two causes it was.

A match's ref selects its heading alone except where its anchor is one
segment another heading of the document shares — a top-level heading of
a document with no H1 — or two siblings share text all the way up; such
a heading's content is read through the span the full answer returns,
and the refusal lists it apart from the refs that do name one heading
alone. The ref leads with the document's declared slug rather than its
anchor's first segment, which is the H1's identifier and not the slug,
so it is the form a nested wikilink anchor takes under
[[cross-document-references]]: a ref copied from an answer or a refusal
works as a selection, as a reference query's argument and as a citation
without editing (AC-SPECSTR-003.12). A member that declares no slug is
reached only by its path, so its ref leads with the path. A single
segment remains a valid selection. A full heading path takes precedence
over a deeper heading it is also a subsequence of, which is what lets a
pasted `document-core#references` name the top-level `References` alone
though `document-core#components#references` satisfies it too
(AC-SPECSTR-003.11); the precedence applies to a selection of more than
one segment only, so a one-segment selection is matched on its
identifier against every heading and names a top-level heading alone
only where no other heading shares its identifier.

AC-SPECSTR-003.14 applies the rule the list query follows: a member at a
status no longer in effect is withheld unless the caller asks for every
status, so a question about the live corpus is not answered mostly from
retired blueprints. AC-SPECSTR-003.15 is its limit — a caller who names
a retired document asked for it, and withholding it would answer a
question nobody put.

AC-SPECSTR-003.16 makes the compact view the default for the same
reason the reference query's is: the common question is "which headings,
and where", and most of a full match is section-span offsets. It is a
named view as [[cli-output-standards]] defines one — every match stays,
every kept value is whole — and the full option returns every field.
The emit form writes a section rather than an answer, so neither view
applies to it and AC-SPECSTR-003.17 refuses the combination rather than
ignoring one of the two options. The statuses in AC-SPECSTR-003.7,
AC-SPECSTR-003.8, AC-SPECSTR-003.13 and AC-SPECSTR-003.17 are those
[[cli-exit-status-standards]] assigns to an operand the command cannot
act on and to a subject the corpus does not hold.

## References

- [[frd-spec-index]] — the index, finding engine, corpus discovery and
  configuration this feature extends
- [[frd-spec-structure-conformance]] — checks documents against the
  structure this feature records, and registers the record-block tag
  vocabulary AC-SPECSTR-001.2 parses against
- [[frd-spec-anchor-references]] — resolves an anchor against the
  headings this feature records
- [[adr-008]] — the element record blocks this feature extracts
- [[cli-exit-status-standards]] — the exit statuses the structure query
  reports
- [[cli-output-standards]] — the named view the structure query's
  compact answer is
- [[adr-021]] — Gherkin, the notation these features and acceptance
  criteria are written in
- [[gherkin-guide]] — how to write them
