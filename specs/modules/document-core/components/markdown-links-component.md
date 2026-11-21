---
title: "Markdown Links"
entity_name: Markdown Links
technology: Go package
description: Settling the target each resolved wikilink points at, and
  substituting each `[[...]]` construct with the Markdown link it settles.
  Change it to change how a target or an anchor fragment is chosen, or
  what a substitution replaces.
slug: component-markdown-links
type: component-blueprint
status: ready-for-review
kind: standard
contracts_consumed: []
source_paths:
  - "cli/core/mdlinks"
test_paths:
  - "cli/core/mdlinks"
---

# Markdown Links

## Responsibility

Markdown Links turns a document's wikilinks into ordinary Markdown links,
so a document can be read outside the corpus. It works in two halves.
`Resolve` decides what each already-resolved reference should point at,
from a caller's link map where one names the slug or from the cited
document's own path, and reports a reference that settles neither way
instead of guessing. `Substitute` then replaces each `[[...]]` construct
whose target settled, and copies everything else through unchanged.

### Owns

The target each reference settles to, including its anchor fragment.

The `links/unmapped` and `links/unresolved` findings.

The substitution of a wikilink construct by a Markdown link.

### Hides

The path arithmetic that makes a target relative to the citing document,
and how a construct's extent is widened from a reference's span.

### Does not own

Extracting and resolving references — [[component-references]] owns it.

The heading identifier a fragment is derived from —
[[component-markdown-structure]] owns it.

Reading the document and writing the result — [[container-spec-cli]] owns
the `links` command.

### Composed into

[[module-document-core]] — a Go package the `links` command calls; see that
blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of `cli/core/mdlinks`, entered through `Resolve`
and `Substitute`.

### Provides

`Resolve` SHALL return one target per reference, aligned by index with its
input, and no target where none settled.

`Resolve` SHALL return every finding a run raises, not only the first.

`Substitute` SHALL copy every byte that no replaced reference covers
unchanged.

`Substitute` SHALL leave a reference that came from frontmatter untouched.

`Substitute` SHALL return the number of references it replaced.

### Requires

WHEN `Substitute` is called, the caller SHALL pass the same bytes the
references' spans were recorded against.

WHEN `Substitute` is called, the caller SHALL pass references already
filtered through `references.OutsideGherkinFences`.

## Design

### Patterns and principles

Functional Core, Imperative Shell (Bernhardt, 2012) — `Resolve` and
`Substitute` are pure functions over bytes and records; the command surface
is the shell that reads the document and writes the result. It serves the
`links` command, which writes no file.

### Invariants

The component SHALL perform no file or network I/O. Enforced by: review
only.

`Substitute` SHALL derive each construct's extent from the reference's
`Span`, never by scanning the document again. Enforced by:
`TestAWikilinkInsideACodeRegionIsNeverRewritten` for the result; review only
for the mechanism.

The component SHALL derive every anchor fragment through
`markdown.Identifier`, holding no derivation of its own. Enforced by:
`TestAResolvedAnchorTakesItsFragmentFromTheResolvedHeading`.

### Extension points

A new way to settle a target — a step in `Resolve` ahead of derivation, as
the link map is. Not by a decision inside `Substitute`.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN the shape of a reference's span changes, `Substitute`'s extent
derivation SHALL change in the same change.

### Verification gates

#### Verify always

`cd cli && go test ./core/mdlinks/`

#### Verify conditionally

WHEN target settling or substitution changes, run
`cd cli && go test ./cmd/ -run 'TestSnapshotCases|Links'`.

## Key decisions

[[adr-012]] — the `[[slug]]` and heading-anchor conventions this component
rewrites.

## References

[[module-document-core]] — the module this component is composed into.
