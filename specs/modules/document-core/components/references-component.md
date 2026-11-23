---
title: "References"
entity_name: References
technology: Go package
description: Reference extraction from document bodies and registered
  frontmatter properties, and resolution of each reference against the
  index. Change it to change what counts as a reference, how a slug or an
  anchor resolves, or which regions extraction skips.
slug: component-references
type: component-blueprint
status: ready-for-review
kind: standard
contracts_consumed: []
source_paths:
  - "cli/core/references"
test_paths:
  - "cli/core/references"
---

# References

## Responsibility

References finds every `[[slug]]` and `[[slug#Anchor]]` wikilink a
document makes and records where each one occurs. It reads the document's
body and its reference-bearing frontmatter properties: the ones the active
profile registers for the document's type. The active profile is the set of
document-type rules the CLI runs with.

References then resolves each reference against the index, the record of
every file the corpus walk read. A reference resolves to a member (a file
indexed as a spec document), to a file that isn't a member, or to nothing.
An anchor also resolves to the heading it names. The two halves are
separate because they know different things: extraction knows one document,
and resolution knows the whole corpus.

### Owns

Extraction: which text counts as a reference, which regions extraction
skips (the code-region mask), and the span (the byte range, lines, and
columns in the source) that each reference is recorded at.

Resolution: the resolution state each reference receives, and the heading
an anchor resolves to.

The split between the references the index keeps and the ones a renderer
rewrites, made by the `OutsideGherkinFences` function.

### Hides

How code regions are located, and how a slug is matched against a file that
isn't a member. Callers see only the `specmodel.Reference` struct, which
carries the reference's span and resolution.

### Does not own

How a heading identifier and a heading path are derived, and which headings
a path selects — [[component-markdown-structure]] owns it.

Which frontmatter properties hold references — [[module-spec-profile]]
registers them.

Reporting a broken reference as a finding — [[component-corpus]] owns the
reference checks.

### Composed into

[[module-document-core]] — a Go package that the Corpus component's walk
calls: `Extract` for each member and `.feature` file it reads, and `Resolve`
once after the last file. See that blueprint for this component's
relationships.

## Interface

### Surface

The exported identifiers of the `cli/core/references` package, entered
through the `Extract` and `Resolve` functions.

### Provides

`Extract` SHALL return a document's references ordered by position: the
registered frontmatter properties in declaration order, then the body in
the order it's written.

The component SHALL record a body reference's span over the text between
its brackets.

The component SHALL record a frontmatter reference's span over the value's
source text as written, quotes included.

WHEN `Resolve` returns, every reference it was given SHALL carry a
resolution, including a reference that reached nothing.

`OutsideGherkinFences` SHALL return its input without the body references
inside a Gherkin-tagged fence, in the same order.

### Requires

WHEN `Extract` is called, the caller SHALL pass the reference-bearing
property names the active profile registers for the document's type.

WHEN `Resolve` is called, the index SHALL already record every member's
headings.

## Design

### Patterns and principles

Separation of Concerns (Dijkstra, 1974) — extraction (`Extract`) works on
one document, and resolution (`Resolve`) works on one index. The split fits
how the corpus walk runs. The walk extracts each file's references as it
reads the file. It resolves them all after it reads the last file, because
the first file may cite the last.

### Invariants

`Extract` SHALL leave every reference's resolution unset. Enforced by:
`TestExtractionResolvesNothing`.

`Extract` SHALL read frontmatter references from the parsed frontmatter
block, never from an index entry. Enforced by: review only.
`TestExtractionReadsNoIndexEntry` checks only that the package doesn't
import `specindex`. The index entry type, `specmodel.IndexEntry`, is in the
`specmodel` package, which this package does import.

`Resolve` SHALL compute each resolution from the references and the index it
is given, reading no file. Enforced by: `TestEveryReferenceIsGivenAResolution`
and `TestResolutionDoesNotDependOnIndexOrder` for the result; review only for
reading no file.

The component SHALL select an anchor's headings through the heading paths
and selection rule of [[component-markdown-structure]], never a derivation
of its own. Enforced by:
`TestAnAnchorMatchesAHeadingByIdentifierNotByExactText` for the result;
review only for using no derivation of its own.

The component SHALL read the Gherkin fence tags from `gherkin.FenceTags`,
never from a literal. Enforced by: review only.

The code-region mask and `OutsideGherkinFences` SHALL share one `goldmark`
walk, so they agree on where a fence ends. Enforced by: review only.

### Extension points

A new kind of source — the caller passes the file's bytes to `Extract` with
no frontmatter, and `Extract` reads the whole file as body. The corpus walk
does this for `.feature` files. Not by a second wikilink scanner.

A new reference-bearing property — register it in the profile. Not by
naming the property in this package.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN the span a reference is recorded at changes, every consumer that
derives an extent from that span SHALL be updated in the same change.

### Verification gates

#### Verify always

`cd cli && go test ./core/references/`

#### Verify conditionally

WHEN a span or a resolution state changes, run
`cd cli && go test ./core/corpus/ ./core/mdlinks/`.

## Key decisions

[[adr-012]] — the `[[slug]]` and heading-anchor conventions this component
extracts.

[[adr-018]] — references are one top-level edge collection on the index,
with both directions derived from it.

## References

[[module-document-core]] — the module this component is composed into.
