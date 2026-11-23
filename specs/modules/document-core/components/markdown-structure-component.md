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
each with the span of its section, and its fenced blocks, located by info
string. A section is the content beneath a heading, up to the next heading
at the same or a higher level. A span is a position in the source: a byte
offset and length, with line and column.

The component also owns the only derivation of a heading identifier: the
lowercase, hyphenated form of a heading's text that anchors match on. It
owns the only rule for which headings a heading path selects. A heading path
is the identifiers of a heading's enclosing headings followed by its own,
joined by `#`. Because each rule has one home, the `specs headings` command,
an anchor (the `#heading` part of a `[[slug#heading]]` reference), and a
link fragment the CLI emits can't name different headings. An agent that
asks for one section's span and reads only that section depends on this
component.

### Owns

The heading tree: each heading's text, level, position, and section span.

Locating fenced blocks whose info string exactly matches one of the tags a
caller passes in.

Deriving the heading identifier of a text, deriving heading paths, and
deciding which headings a path selects.

### Hides

The Markdown parser, and how a prefixed heading (`## Title`) is told apart
from an underlined one (text above a `---` or `===` line) inside a container
block, such as a blockquote or a list item. Callers see positions, as
`span.Span` values from the Spans component (package `cli/core/span`), and
pure functions over text.

### Does not own

Recording headings and record blocks (tagged fenced blocks that hold an
element's metadata) on the corpus index, and decoding a record block's
content — [[component-corpus]] owns both. The index model is the set of Go
types in `cli/core/specmodel`, generated from a schema.

Resolving an anchor against a target document — [[component-references]]
owns it.

### Composed into

[[module-document-core]] — a Go package that the module's parsing and model
components call. See that blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of the Go package `cli/core/markdown`.

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

Single Source of Truth (Codd, 1970; Hunt and Thomas, 1999) — the function
`Identifier` is the only identifier derivation, and the function
`SelectPaths` is the only path selection rule. They serve the
`specs headings` command, anchor resolution, emitted Markdown link
fragments, and the C4 Elements component, which matches headings by
identifier.

### Invariants

The component SHALL recognize headings and fences through `goldmark`, never
through a scan of its own. Enforced by: review only.
`TestAHeadingInsideAFencedBlockOrACodeSpanIsNotReported` and
`TestAFalseClosingRunShorterThanTheOpeningIsContent` cover cases that a scan
can get wrong, but a scan that handled them correctly would also pass.

The component SHALL compute every position through the Spans component.
Enforced by: `TestEverySpanIsComputedThroughSpans` for heading spans, which
compares each with the value `span.Of` returns; review only for fenced-block
spans.

No two headings' positions SHALL overlap, whatever form each is written in.
Enforced by: `TestNoTwoHeadingPositionsOverlapForAnyForm`.

`Identifier` SHALL derive no positional suffix. Enforced by:
`TestTheHeadingIdentifierIsPureAndDerivesNoPositionalSuffix`.

The component SHALL NOT depend on the generated index model; its path
functions take an `Outline`. Enforced by: review only.

### Extension points

A new construct to locate — a function beside the function `Fenced` that
recognizes the construct through `goldmark` and computes its position
through Spans. Not by a regular-expression scan of the body.

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

[[adr-015]] — the CLI parses Markdown with `goldmark`.

## References

[[module-document-core]] — the module this component is composed into.
