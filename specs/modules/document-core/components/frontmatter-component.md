---
title: "Frontmatter"
entity_name: Frontmatter
technology: Go package
description: Locating a document's frontmatter block and parsing it into a
  value tree that keeps each property's raw text and position. Change it to
  change where a block begins and ends, what a parsed property carries, or
  how a value is read from a block that doesn't parse.
slug: component-frontmatter
type: component-blueprint
status: ready-for-review
kind: standard
contracts_consumed: []
source_paths:
  - "cli/core/frontmatter"
test_paths:
  - "cli/core/frontmatter"
---

# Frontmatter

## Responsibility

Frontmatter splits a Markdown document's frontmatter block (the YAML
between `---` lines at the top of the file) from its body. It parses the
block once into a value tree of generic Go maps, slices, and scalars. Schema
validation reads that tree directly. Typed decoding converts the same tree to
JSON and decodes that into the Go record types generated from the project's
schemas. Each property keeps its raw source text beside its parsed value,
because a check has to validate what the author wrote, not what a YAML
parser inferred from it.

### Owns

Where a frontmatter block begins and ends.

The parsed block: for each top-level property, its name, the positions of
its key and value, its raw text, and its parsed value. A sequence-valued
property also carries the position of each sequence element.

Inline comments: a `#` comment after an unquoted scalar value on the same
line, which YAML reads as the end of the value.

Reading one top-level scalar value, line by line, from a block that doesn't
parse as YAML.

### Hides

The YAML library (`goccy/go-yaml`), and how its token positions become
byte offsets. Callers see byte-accurate spans (the `specmodel.Span` type: a
byte offset and length, plus start and end line and column) and a tree of
values that JSON can carry.

### Does not own

Validating a block against a schema, and decoding it into the typed
envelope (the generated Go record for the document's type, which carries the
properties every document type shares) — [[component-schema-registry]] owns
both.

Which properties a document type declares, and which of them hold
references to other documents — [[module-spec-profile]] registers them.

### Composed into

[[module-document-core]] — a Go package that the module's parsing and model
components call; see that blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of the `cli/core/frontmatter` package. Callers
enter through the functions `Parse`, `Locate`, and `TopLevelScalarAt` (or
its value-only form, `TopLevelScalar`). An error from `Parse` is a located
error: a `*specmodel.ParseError` that holds a message and the span of the
text it is about.

### Provides

`Parse` SHALL return exactly one of a parsed document, no document and no
error for a file without frontmatter, or no document and a located error.

`Parse` SHALL NOT panic on any input.

IF a block parses to something other than a mapping, THEN `Parse` SHALL
return a located error.

A property's raw text SHALL be its value's source text as written, not a
re-rendering of the parsed value.

The value tree SHALL hold only values JSON can carry.

A sequence-valued property SHALL carry one item per element, each with the
element's own span.

IF a value can't be read unambiguously from its own line, THEN
`TopLevelScalar` SHALL return no value rather than a guess.

### Requires

WHEN `Parse` or `Locate` is called, the caller SHALL pass the document from
its first byte, because a block is recognized only there.

## Design

### Patterns and principles

Single Source of Truth, for knowledge in code (Hunt and Thomas, 1999, "The
Evils of Duplication") — the function `Locate` is the one place that decides
where a block begins and ends, and `Parse` and `TopLevelScalarAt` call it. It
also serves callers that need where a block begins and ends, without its
values: [[component-corpus]], to find where a document's body starts, and
[[component-c4-elements]], to find a blueprint's frontmatter and body.

### Invariants

The component SHALL decide where a block begins and ends only in `Locate`.
Enforced by: `TestParseAndLocateAgreeOnTheRegion` and
`TestTheImplementationCallsTheSharedScanner` for `frontmatter.go`, and by
review only for `scalar.go`, which that test doesn't read.

The component SHALL derive every byte offset from the block's own bytes,
not from the YAML library's token offsets. Enforced by:
`TestAMultiByteCharacterEarlierInTheBlockDoesNotShiftALaterProperty` and
`TestSpansSurviveCommentsAndMultiByteText`.

The component SHALL read inline comments from the YAML lexer's tokens that
`Parse` parsed, so that the value tree and the comments agree on which
token is which. Enforced by: review only.

`Parse` SHALL return malformed YAML as a located error. Enforced by:
`TestMalformedYAMLIsALocatedErrorRatherThanAPanic`.

### Extension points

A new way to read a block that didn't parse — a function that calls
`Locate`, as `TopLevelScalarAt` does. Not by re-deriving the block's
delimiters.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN code anywhere in the module needs to know where a frontmatter block
begins and ends, it SHALL call `Locate` rather than search for the
delimiters itself.

WHEN a test of a position is added, it SHOULD place a multi-byte character
before the property it checks, because an ASCII fixture can't expose an
offset that counts codepoints instead of bytes.

### Verification gates

#### Verify always

`cd cli && go test ./core/frontmatter/`

#### Verify conditionally

WHEN what a parsed property carries changes, run
`cd cli && go test ./core/...`.

## Key decisions

[[adr-003]] — documents carry their metadata as YAML frontmatter.

[[adr-015]] — `goccy/go-yaml` is the YAML parser.

## References

[[module-document-core]] — the module this component is composed into.
