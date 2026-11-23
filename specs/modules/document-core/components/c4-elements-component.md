---
title: "C4 Elements"
entity_name: C4 Elements
technology: Go package
description: The corpus's C4 elements, the scope of C4 elements that each C4
  drawing document's diagram nodes may name, and the resolution of nodes to C4
  elements and of relationship statements to drawings or to unresolved C4
  relationships. Change it to change what declares a C4 element, what a
  diagram may name, or how a statement resolves.
slug: component-c4-elements
type: component-blueprint
status: ready-for-review
kind: standard
contracts_consumed: []
source_paths:
  - "cli/core/c4elements"
test_paths:
  - "cli/core/c4elements"
---

# C4 Elements

## Responsibility

C4 Elements decides what each diagram node names, so the Corpus component's
checks and the C4 Model component get the same answer. From the index and the
architecture vocabulary that the profile registers, it decides which
blueprints declare C4 elements and which record blocks in a Components section
do. Then it works out which C4 elements each C4 drawing document's diagrams may
name: that document's scope. Each node, other than one that draws an external
C4 element, is resolved against that scope, and each relationship statement
against the nodes of its diagram. Apart from the `slug` property, which every
document carries, the component names no document type, property, or heading
of its own.

### Owns

The set of C4 elements: which corpus documents and record blocks declare a C4
element, and how declarations of one slug merge.

Each C4 drawing document's scope and the C4 elements that its Components section
declares.

Node resolution, and the resolution of each relationship statement either to
a drawing or to an unresolved C4 relationship: a relationship statement with
an endpoint that doesn't name exactly one C4 element, which carries the
reason.

The grouping of drawings into relationships.

The Go types that define the shape of the architecture vocabulary, which the
`Profile` extension point names.

### Hides

How the heading of a Components section is matched. Callers see C4 elements,
nodes, drawings, and relationships.

### Does not own

Parsing Mermaid statements — [[module-document-core#C4 Diagrams]] owns it.

Assembling the C4 model — [[component-c4-model]] owns it.

The C4 checks, and reporting a malformed record block in a Components
section — [[component-corpus]] owns them.

The architecture vocabulary itself — [[module-spec-profile]] registers it.

### Composed into

[[module-document-core]] — a Go package that the Corpus component's checks
and the C4 Model component call; see that blueprint for this component's
relationships.

## Interface

### Surface

The exported identifiers of the `cli/core/c4elements` package, entered through
the `Build` function.

### Provides

The `Build` function SHALL return one C4 element per slug, carrying every
declaration of that slug ordered by path and then by span.

`Build` SHALL resolve every relationship statement to a drawing, to an
unresolved C4 relationship carrying exactly one reason, or, for a `BiRel`
statement, to neither.

`Build` SHALL group drawings whose source element, target element, and label
are equal into one relationship, wherever each is drawn.

IF the profile registers no architecture vocabulary, THEN `Build` SHALL
return a set with no C4 elements.

### Requires

WHEN `Build` is called, the caller SHALL supply in the `Input` struct a
function that returns a corpus document's bytes as the corpus reads them.

WHEN `Build` is called, the caller SHALL supply in the `Input` struct the
corpus's own record-block decoding function.

## Design

### Patterns and principles

Dependency Inversion (Martin, "The Dependency Inversion Principle",
*C++ Report*, 1996) — the component reads sources and record blocks only
through the functions that its caller passes in the `Input` struct. It serves
a corpus opened from the warm index, a corpus that reads a document from disk
only when asked, and it serves tests that supply sources in memory.

### Invariants

The component SHALL read a corpus document's bytes only through the
`Input.Source` field. Enforced by: review only.

The component SHALL check the reasons for an unresolved C4 relationship in one
fixed order, checking the source endpoint before the target. Enforced by:
`TestUnresolvedStatementsCarryOneReasonInOrder`.

The component SHALL hold no string literal that names registered vocabulary,
as [[module-document-core]]'s vocabulary invariant requires. Enforced by:
`TestTheExtensionPointNamesNoRegisteredVocabulary` in
`cli/core/corpus/profile_test.go` for document types, statuses,
reference-bearing properties and criterion qualifiers; review only for the
architecture vocabulary.

### Extension points

A new architecture property — a field of the `ArchitectureRules` struct that
the profile registers, which the `ValidateRegistration` function checks
against the union schema. Not by a property name hard-coded in this package.

The C4 levels are closed: they're C4 notation, fixed in the `ElementLevel`
type.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests that cover
its exported surface, each cited to the acceptance criteria it verifies.

## Change guidance

WHEN scope or resolution changes, this repository's own diagrams SHALL still
resolve every node, other than one that draws an external C4 element or a
passive store, to one C4 element, checked by the profile's self-application
gate in the same change.

### Verification gates

#### Verify always

`cd cli && go test ./core/c4elements/`

#### Verify conditionally

WHEN scope, resolution, or grouping changes, run
`cd cli && go test ./core/c4model/ ./core/corpus/ ./profile/`.

## Key decisions

[[adr-029]] — relationships are read from blueprint diagrams, with reserved
tags.

[[adr-030]] — the architecture vocabulary arrives as a profile registration.

[[adr-008]] — un-promoted C4 elements are declared by record blocks.

[[adr-009]] — a parent owns the edges among its children.

## References

[[module-document-core]] — the module this component is composed into.
