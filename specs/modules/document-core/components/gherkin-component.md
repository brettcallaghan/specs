---
title: "Gherkin"
entity_name: Gherkin
technology: Go package
description: Parsing a Gherkin source into one feature with its criteria
  nested beneath it, the criterion content hash, and rendering a feature
  back out as a standalone source. Change it to change what a feature or
  criterion record carries from the parse, how a criterion is hashed, or
  how a feature is rendered.
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

Gherkin parses one Gherkin source into one feature with its criteria nested
beneath it. A source is either a fenced code block inside a Markdown document
or a standalone `.feature` file. A criterion is one acceptance criterion: a
`Scenario`, or one `Examples` block of a scenario in which at least one
`Examples` block carries its own criterion identifier. Each criterion carries
its text, its tags as written, and its position.

The parse is general, and the meaning of a tag isn't. Which tag identifies
a feature and which identifies a criterion are facts about one tag
vocabulary (the tag patterns a profile registers). They reach this
component as functions its caller supplies.

A Gherkin qualifier tag is a tag that the profile gives a meaning on a
criterion, such as marking it a prohibition or one that only a person can
verify. Which Gherkin qualifier tags count is the caller's to decide, from the
tags this component carries. The component never inspects a tag's spelling.
That is the module's split between general machinery and the vocabulary a
profile supplies, applied to a notation.

### Owns

The parse of a source into `Feature` and `Criterion` values (types in this
package). Each wraps a `specmodel.Feature` or `specmodel.Criterion` record
(the index types generated into `cli/core/specmodel`) with the tags the
record doesn't carry. It also owns the located parse failure, the
`ParseError` type.

The criterion content hash: a SHA-256 digest of a criterion's normalized
name and text, computed by the `ContentHashOf` function.

The info strings that mark a fenced block as Gherkin.

Rendering a feature back out as a standalone `.feature` source.

### Hides

The official Cucumber Gherkin parser, and the synthetic header: a bare
`Feature:` line prepended to a source that doesn't open with a `Feature`
keyword, so that the parser accepts it. Callers see positions in their own
document's coordinates.

### Does not own

What a tag means — [[module-spec-profile]] registers the tag vocabulary.

Finding sources in the spec root and attributing each feature to an owning
document — [[component-corpus]] owns both.

Running the identity and content checks over the parsed records —
[[component-corpus]] owns the checks.

### Composed into

[[module-document-core]] — a Go package that the Corpus component calls for
each source it finds in the spec root; see that blueprint for this
component's relationships.

## Interface

### Surface

The exported identifiers of the `cli/core/gherkin` package. Callers enter
it through the `Parse`, `ParseRegion` and `Render` functions.

### Provides

`Parse` and `ParseRegion` SHALL return either a feature or a `*ParseError`,
never both, and SHALL return neither for a source that holds only blank
lines, comments and tags.

`Parse` and `ParseRegion` SHALL NOT panic on any input.

`ParseRegion` SHALL report every position in the containing document's
coordinates.

The component SHALL NOT report a position inside the synthetic header.

A criterion's content hash SHALL cover its normalized name and text and
nothing else.

`Render` SHALL produce a source that `Parse`, under the same tag
vocabulary, reads back to a feature with the same tags, names,
descriptions, rules, text and content hashes, at fresh positions.

### Requires

WHEN `Parse` or `ParseRegion` is called, the caller SHALL supply in
`Options` the functions that resolve a set of tags to a feature identity and
to a criterion identifier, or leave either nil to resolve nothing.

## Design

### Patterns and principles

Strategy (Gamma et al., 1994) — the two resolution functions in the
`Options` struct are the strategies, and the caller chooses them. The
pattern serves the variations that exist: the Corpus component resolves
tags with the patterns the active profile registers, and a caller with no
vocabulary passes none.

### Invariants

The component SHALL NOT inspect a tag's spelling; it passes tags to the
caller's resolution functions as strings. Enforced by:
`TestParsingWithNoVocabularyStillYieldsRecords` for the parse result, and
review only for the package source.

The content hash SHALL exclude spans and the node identifiers the official
parser assigns. Enforced by: `TestACriterionThatOnlyMovedHashesIdentically`
and `TestReindentingOrCheckingOutWithCRLFLeavesTheHashUnchanged`.

`ContentHashOf` SHALL be the one hash computation, so every producer of a
criterion hashes it the same way. Enforced by: review only.

### Extension points

A new property on a feature or criterion record — a field in
`schemas/feature.strata` (the Strata schema that generates the `specmodel`
types), populated from what the parse establishes. Not by decomposing a
criterion's text into typed steps or example rows.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, each citing the acceptance criteria it verifies.

## Change guidance

WHEN a change needs to know what a tag means, the meaning SHALL arrive
through the caller's resolution functions, from what the profile registers,
and not through a tag spelling written in this package.

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
