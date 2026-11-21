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

Frontmatter splits a document's YAML block from its body and parses it once
into a generic value tree. That one tree serves both schema validation and,
converted to JSON, decoding into the generated typed records. Each property
keeps its raw source text beside its parsed value, because a check has to
validate what the author wrote, not what a YAML parser inferred from it.

### Owns

Where a frontmatter block begins and ends.

The parsed block: each top-level property's name, key and value positions,
raw text and value, with per-element positions for a sequence.

Comments that trail a value on its line, which cut the value short.

Reading one top-level scalar from a block that doesn't parse.

### Hides

The YAML library, and how its token positions become byte offsets. Callers
see byte-accurate `Span` values and a JSON-compatible tree.

### Does not own

Validating a block against a schema, and decoding it into a typed envelope
— [[component-schema-registry]] owns both.

Which properties a document type declares, and which hold references —
[[module-spec-profile]] registers them.

### Composed into

[[module-document-core]] — a Go package the parsing and model components
call; see that blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of `cli/core/frontmatter`, entered through `Parse`
and `Locate`.

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

Single Source of Truth (Codd, 1970; Hunt and Thomas, 1999) — `Locate` is
the one place a block's extent is decided, and `Parse` is written on it. It
serves the callers that need a block's extent without its values: the
Corpus walk for a member whose frontmatter didn't parse, and the C4
components.

### Invariants

The component SHALL decide a block's extent only in `Locate`. Enforced by:
`TestParseAndLocateAgreeOnTheRegion` and
`TestTheImplementationCallsTheSharedScanner`.

The component SHALL derive every byte offset from the block's own bytes,
not from the YAML library's token positions. Enforced by:
`TestAMultiByteCharacterEarlierInTheBlockDoesNotShiftALaterProperty` and
`TestSpansSurviveCommentsAndMultiByteText`.

The component SHALL read inline comments from the YAML lexer's tokens, so
no value or span `Parse` reports depends on them. Enforced by: review only.

`Parse` SHALL return malformed YAML as a located error. Enforced by:
`TestMalformedYAMLIsALocatedErrorRatherThanAPanic`.

### Extension points

A new read of a block that didn't parse — a function written on `Locate`,
as `TopLevelScalar` is. Not by re-deriving the block's delimiters.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN code anywhere in the module needs a frontmatter block's extent, it
SHALL call `Locate` rather than search for the delimiters itself.

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
