---
title: "Gherkin"
entity_name: Gherkin
technology: Go package
description: Parsing a Gherkin source into one feature record with its
  criteria nested beneath it, the criterion content hash, and rendering a
  feature back out as a standalone source. Change it to change what a
  feature or criterion record carries from the parse, how a criterion is
  hashed, or how a feature is rendered.
slug: component-gherkin
type: component-blueprint
status: ready-for-review
kind: standard
contracts_consumed: []
source_paths:
  - "cli/core/gherkin"
test_paths:
  - "cli/core/gherkin"
---

# Gherkin

## Responsibility

Gherkin parses one Gherkin source, whether a fenced block inside a document
or a standalone `.feature` file, into one feature record with its criteria
nested beneath it, each with its text, its tags as written and its
position. Parsing is general; meaning isn't. Which tag identifies a
feature, which identifies a criterion and which qualifiers count are facts
about one vocabulary, and they reach this component as functions its caller
supplies. It never inspects a tag's spelling, which is the module's
host-and-plugin split applied to a notation.

### Owns

The parse of a source into `specmodel.Feature` and `specmodel.Criterion`
records, and the located parse failure.

The criterion content hash.

The info strings that mark a fenced block as Gherkin.

Rendering a feature record back out as a standalone `.feature` source.

### Hides

The official parser and the synthetic header a source with no `Feature` is
parsed behind. Callers see positions in their own document's coordinates.

### Does not own

What a tag means — [[module-spec-profile]] registers the tag vocabulary.

Finding sources in the spec root and attributing each feature to an owning
document — [[component-corpus]] owns both.

Reporting identity and content findings over the parsed records —
[[component-corpus]] owns the checks.

### Composed into

[[module-document-core]] — a Go package the Corpus walk calls for each
source; see that blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of `cli/core/gherkin`, entered through `Parse` and
`ParseRegion`.

### Provides

`Parse` and `ParseRegion` SHALL return either a feature or a `*ParseError`,
never both.

`Parse` and `ParseRegion` SHALL NOT panic on any input.

`ParseRegion` SHALL report every position in the containing document's
coordinates.

The component SHALL NOT report a position inside the synthetic header.

A criterion's content hash SHALL cover its name and its normalized text and
nothing else.

`Render` SHALL produce a source that `Parse` reads back to the same feature.

### Requires

WHEN `Parse` or `ParseRegion` is called, the caller SHALL supply in
`Options` the functions that resolve a set of tags to a feature identity and
to a criterion identifier.

## Design

### Patterns and principles

Strategy (Gamma et al., 1994) — tag resolution is a pair of functions in
`Options`, chosen by the caller. It serves the variation that exists: the
profile's registered patterns in the Corpus walk, and parsing with no
vocabulary at all.

### Invariants

The component SHALL NOT inspect a tag's spelling; every tag is passed to the
caller's resolution functions as a string. Enforced by:
`TestParsingWithNoVocabularyStillYieldsRecords` for the result; review only
for the source.

The content hash SHALL exclude spans and parser node identifiers. Enforced
by: `TestACriterionThatOnlyMovedHashesIdentically` and
`TestReindentingOrCheckingOutWithCRLFLeavesTheHashUnchanged`.

`ContentHashOf` SHALL be the one hash computation, so every producer of a
criterion hashes it the same way. Enforced by: review only.

### Extension points

A new property on a feature or criterion record — a field in
`schemas/feature.strata`, populated from what the parse establishes. Not by
decomposing a criterion's text into typed steps or example rows.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN a change needs to know what a tag means, the meaning SHALL arrive
through the caller's resolution functions from the profile's registration,
not through a tag spelling in this package.

### Verification gates

#### Verify always

`cd cli && go test ./core/gherkin/`

#### Verify conditionally

WHEN a record's content or the hash changes, run
`cd cli && go test ./core/corpus/ ./cmd/`.

## Key decisions

[[adr-021]] — Gherkin is the notation for features and acceptance criteria,
read with the official Go parser.

## References

[[module-document-core]] — the module this component is composed into.
