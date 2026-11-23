---
title: "Markdown Links"
entity_name: Markdown Links
technology: Go package
description: Choosing the target path that each resolved wikilink points
  at, and replacing each `[[...]]` wikilink, brackets included, with a
  Markdown link to that target. Change it to change how a target or an
  anchor fragment is chosen, or what a substitution replaces.
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
so that the document can be read outside the corpus. A wikilink is a
`[[slug]]` reference to another corpus document. The work has two halves,
one function each. The `Resolve` function settles each reference's target:
it picks the path that the Markdown link points at. Each reference arrives
already resolved, meaning an earlier step has looked its slug up in the
corpus. `Resolve` takes the target from the caller's link map (a
slug-to-path map) when the map names the slug, and otherwise from the cited
document's own path. When a reference settles neither way, `Resolve`
reports it instead of guessing. The `Substitute` function then replaces
each `[[...]]` construct (the wikilink, brackets included) whose target
settled, and copies everything else through unchanged.

### Owns

The target each reference settles to, including its anchor fragment (the
`#` part that points at a heading).

The `links/unmapped` and `links/unresolved` findings: the diagnostics
`Resolve` raises when a link map leaves out a corpus document, and when a
slug names no corpus document and no link map entry.

Replacing a wikilink construct with a Markdown link.

### Hides

How a target path is made relative to the citing document, and how the span
a construct covers is computed by widening the reference's span (the text
between the brackets) by the brackets on each side.

### Does not own

Extracting and resolving references — [[component-references]] owns it.

The heading identifier a fragment is derived from —
[[component-markdown-structure]] owns it.

Reading the document and writing the result — [[container-spec-cli]] owns
the `links` command.

### Composed into

[[module-document-core]] — as a Go package that the `links` command
calls; see that Module Blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of the `cli/core/mdlinks` package. Callers enter
through the `Resolve` and `Substitute` functions.

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
filtered through the `references.OutsideGherkinFences` function.

## Design

### Patterns and principles

Functional Core, Imperative Shell (Bernhardt, "Boundaries", 2012) —
`Resolve` and `Substitute` are the core: pure functions over the document's
bytes and its reference records. The shell is Command Surface
([[container-spec-cli#Command Surface]]), the Spec CLI's command layer,
which reads the document and writes the result to standard output.
The component serves the `links` command, which writes no file.

### Invariants

The component SHALL perform no file or network I/O. Enforced by: review
only.

`Substitute` SHALL derive the span each construct covers from the reference's
`Span` field, never by scanning the document again. Enforced by:
`TestNothingButTheWikilinksIsAltered` for the result; review only for the
mechanism.

The component SHALL derive every anchor fragment through the
`markdown.Identifier` function, holding no derivation of its own. Enforced by:
`TestTheAnchorRuleKeepsUnicodeLettersAndDigitsAndDropsTheRest` and
`TestAResolvedAnchorTakesItsFragmentFromTheResolvedHeading` for the
fragments; review only for the mechanism.

### Extension points

A new way to settle a target — a new step in `Resolve` that runs before
the target is derived from the cited document's path, as the link-map
lookup does. Not by a decision inside `Substitute`.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN the shape of a reference's span changes, the function that
derives the span a construct covers, `constructSpan`, SHALL change in the same
change.

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
