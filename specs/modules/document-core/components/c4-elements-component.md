---
title: "C4 Elements"
entity_name: C4 Elements
technology: Go package
description: The corpus's architectural elements, the scope each drawing
  document's diagram nodes may name, and the resolution of nodes and
  relationship statements to elements. Change it to change what declares an
  element, what a diagram may name, or how a statement resolves.
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

C4 Elements decides what each diagram node names, so that the Corpus
component's C4 checks and the C4 Model component work from one answer. It
reads the corpus index and the architecture vocabulary that the active
profile registers. From them, it decides which blueprints and roster record
blocks declare elements. It then works out which elements each drawing
document's diagrams may name, and resolves every node and relationship
statement against that scope.

A drawing document is a corpus document whose type the profile registers as
declaring an element. Only drawing documents' diagrams are resolved. A roster
is the section, under a heading the profile registers, in which a parent
document lists its children. A record block in a roster is a fenced block
that holds one child's element fields. The component names no document type,
property, or heading of its own, apart from the `slug` property that every
document carries.

### Owns

The element set: which corpus documents and record blocks declare an
element, and how several declarations of one slug merge into one element.

Each drawing document's scope (the elements its diagram nodes may name)
and its roster (the children it declares).

Node resolution: matching each node's label to the elements in scope.
Relationship resolution: turning each relationship statement into a drawing
(a statement whose two endpoints each name one element) or into an
unresolved statement with its reason.

The grouping of drawings into relationships.

The Go types of the architecture vocabulary, such as the
`ArchitectureRules` struct, which package `corpus` aliases for its
`Profile` extension point.

### Hides

How a roster heading is matched and how merged declarations are ordered.
Callers see elements, nodes, drawings, and relationships.

### Does not own

Parsing Mermaid statements — [[module-document-core#C4 Diagrams]] owns it.

Assembling the C4 model — [[component-c4-model]] owns it.

The C4 checks, and reporting a malformed roster record —
[[component-corpus]] owns them.

The architecture vocabulary itself — [[module-spec-profile]] registers it.

### Composed into

[[module-document-core]] — a Go package that the Corpus component's checks
and the C4 Model component call; see that blueprint for this component's
relationships.

## Interface

### Surface

The exported identifiers of package `cli/core/c4elements`. Callers enter
through the `Build` function, which takes an `Input` struct and returns a
`*Set` and an error.

### Provides

`Build` SHALL return one element per slug, carrying every declaration of
that slug ordered by path and then by span.

`Build` SHALL resolve every relationship statement to a drawing, to an
unresolved statement carrying exactly one reason, or, for `BiRel`, to
neither.

`Build` SHALL group drawings whose source element, target element and label
are equal into one relationship, wherever each is drawn.

IF the profile registers no architecture vocabulary, THEN `Build` SHALL
return a set with no elements.

### Requires

WHEN `Build` is called, the caller SHALL supply in `Input` a function that
returns a member's bytes as the corpus read them, and the corpus's own
record-block decoding function.

## Design

### Patterns and principles

Dependency Inversion (Martin, 1996) — the component reads source bytes and
record blocks only through the functions its caller passes in the `Input`
struct's `Source` and `Decode` fields. This serves a corpus loaded from its
cached index, which reads a document's file only when asked, and tests that
supply sources in memory.

### Invariants

The component SHALL read a member's bytes only through `Input.Source`.
Enforced by: review only.

The component SHALL check an unresolved statement's reasons in one fixed
order, checking the source endpoint before the target for each reason.
Enforced by:
`TestUnresolvedStatementsCarryOneReasonInOrder`.

The component SHALL hold no string literal naming a document type, a
status, a reference-bearing property or a criterion qualifier that the
generated schema declares, under [[module-document-core]]'s vocabulary
invariant. Enforced by:
`TestTheExtensionPointNamesNoRegisteredVocabulary` in
`cli/core/corpus/profile_test.go`.

### Extension points

A new architecture property — a new field of the `ArchitectureRules`
struct, which the profile fills in, never a property name in this package.
The `corpus.ValidateRegistration` function checks that the profile's union
schema declares the property on each document type it applies to.

The C4 levels are closed: they're C4 notation, fixed as the constants of
the `ElementLevel` string type.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN scope or resolution changes, this repository's own diagrams SHALL
still resolve every node to one element, checked by the profile's
self-application gate in the same change.

### Verification gates

#### Verify always

`cd cli && go test ./core/c4elements/`

#### Verify conditionally

WHEN scope, resolution or grouping changes, run
`cd cli && go test ./core/c4model/ ./core/corpus/ ./profile/`.

## Key decisions

[[adr-029]] — relationships are read from blueprint diagrams, and a set of
relationship tags is reserved for edge properties.

[[adr-030]] — the architecture vocabulary arrives as a profile
registration.

[[adr-008]] — an element with no blueprint of its own is declared by a
record block in its parent's roster.

[[adr-009]] — a parent owns the edges among its children.

## References

[[module-document-core]] — the module this component is composed into.
