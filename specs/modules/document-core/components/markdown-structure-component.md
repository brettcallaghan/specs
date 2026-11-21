---
title: "Markdown Structure"
entity_name: Markdown Structure
technology: Go package
description: Headings with their section spans, fenced blocks by info
  string, and the heading identifier and heading-path rules every anchor and
  heading query uses. Change it to change what counts as a heading or a
  fence, how an identifier is derived, or which headings a path selects.
slug: component-markdown-structure
type: component-blueprint
status: ready-for-review
kind: standard
contracts_consumed: []
source_paths:
  - "cli/core/markdown"
test_paths:
  - "cli/core/markdown"
---

# Markdown Structure

## Responsibility

Markdown Structure reads the structure of a Markdown body: its headings,
each with the span of the content beneath it, and its fenced blocks, located
by info string. It also owns the one derivation of a heading identifier and
the one rule for which headings a heading path selects, so that a `headings`
query, an anchor and an emitted link fragment can't name different headings.
An agent that asks for one section's span and reads only that section
depends on this component.

### Owns

The heading tree: each heading's text, level, position and section span.

Fenced-block location by registered info string.

The heading identifier of a text, heading paths, and path selection.

### Hides

The Markdown parser, and how a prefixed heading is told apart from an
underlined one inside a container block. Callers see positions and pure
functions over text.

### Does not own

Recording headings and record blocks on the index, and decoding a record
block's content — [[component-corpus]] owns both.

Resolving an anchor against a target document — [[component-references]]
owns it.

### Composed into

[[module-document-core]] — a Go package the parsing and model components
call; see that blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of `cli/core/markdown`.

### Provides

`HeadingsFrom` and `FencedFrom` SHALL report every span in the coordinates
of the whole source they're given, whatever offset the search starts at.

`Identifier` SHALL be a pure function of its text, with no positional suffix
for a repeated heading.

IF a fence has no closing line, THEN `Fenced` SHALL return no block for it.

IF a body has no heading, THEN `Headings` SHALL return an empty result
rather than an error.

### Requires

WHEN a caller searches a document that carries frontmatter, the caller SHALL
start the search at the body's offset, because the frontmatter's closing
delimiter reads as an underlined heading.

## Design

### Patterns and principles

Single Source of Truth (Codd, 1970; Hunt and Thomas, 1999) — `Identifier`
and `SelectPaths` are the one derivation and the one selection rule. It
serves the three callers that exist: the `headings` query, anchor
resolution, and Markdown-link fragments.

### Invariants

The component SHALL recognize headings and fences through `goldmark`, never
through a scan of its own. Enforced by:
`TestAHeadingInsideAFencedBlockOrACodeSpanIsNotReported` and
`TestAFalseClosingRunShorterThanTheOpeningIsContent`.

The component SHALL compute every position through the Spans component.
Enforced by: `TestEverySpanIsComputedThroughSpans`.

No two headings' positions SHALL overlap, whatever form each is written in.
Enforced by: `TestNoTwoHeadingPositionsOverlapForAnyForm`.

`Identifier` SHALL derive no positional suffix. Enforced by:
`TestTheHeadingIdentifierIsPureAndDerivesNoPositionalSuffix`.

The component SHALL NOT depend on the generated index model; its path
functions take an `Outline`. Enforced by: review only.

### Extension points

A new construct to locate — a function beside `Fenced` that recognizes the
construct through `goldmark` and positions it through Spans. Not by a
regular-expression scan of the body.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN the heading identifier or the path selection rule changes, the change
SHALL be made in this component and nowhere else.

### Verification gates

#### Verify always

`cd cli && go test ./core/markdown/`

#### Verify conditionally

WHEN `Identifier` or `SelectPaths` changes, run
`cd cli && go test ./core/... ./cmd/`.

## Key decisions

[[adr-012]] — the heading-anchor conventions the identifier serves.

[[adr-015]] — `goldmark` is the Markdown parser.

## References

[[module-document-core]] — the module this component is composed into.
