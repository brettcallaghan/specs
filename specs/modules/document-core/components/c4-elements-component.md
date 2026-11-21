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

C4 Elements owns the judgment of what a diagram node names, so the C4
checks and the C4 model read one answer. It reads the index and the
architecture vocabulary the profile registers, decides which blueprints and
roster record blocks declare elements, works out which elements each
drawing document's diagrams may name, and resolves every node and
relationship statement against that scope. It names no document type,
property or heading of its own.

### Owns

The element set: which members and record blocks declare an element, and
how declarations of one slug merge.

Each drawing document's scope and roster.

Node resolution, and the resolution of each relationship statement to a
drawing or to an unresolved statement with its reason.

The grouping of drawings into relationships.

The architecture vocabulary's types, which the `Profile` extension point
names.

### Hides

How a roster heading is matched and how merged declarations are ordered.
Callers see elements, nodes, drawings and relationships.

### Does not own

Parsing Mermaid statements — [[module-document-core#C4 Diagrams]] owns it.

Assembling the C4 model — [[component-c4-model]] owns it.

The C4 checks, and reporting a malformed roster record —
[[component-corpus]] owns them.

The architecture vocabulary itself — [[module-spec-profile]] registers it.

### Composed into

[[module-document-core]] — a Go package the Corpus checks and the C4 Model
call; see that blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of `cli/core/c4elements`, entered through `Build`.

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

Dependency Inversion (Martin, 1996) — the component reads sources and record
blocks only through the functions its caller passes in `Input`. It serves
the warm corpus, which reads a member off disk only when asked, and tests
that supply sources in memory.

### Invariants

The component SHALL read a member's bytes only through `Input.Source`.
Enforced by: review only.

The component SHALL check an unresolved statement's reasons in one fixed
order, source endpoint before target. Enforced by:
`TestUnresolvedStatementsCarryOneReasonInOrder`.

The component SHALL hold no literal naming registered vocabulary, under
[[module-document-core]]'s vocabulary invariant. Enforced by:
`TestTheExtensionPointNamesNoRegisteredVocabulary` in
`cli/core/corpus/profile_test.go`.

### Extension points

A new architecture property — a member of `ArchitectureRules` the profile
registers, with `ValidateRegistration` checking it against the union
schema. Not by a property name in this package.

The C4 levels are closed: they're notation, fixed in `ElementLevel`.

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

[[adr-029]] — relationships are read from blueprint diagrams, with reserved
tags.

[[adr-030]] — the architecture vocabulary arrives as a profile
registration.

[[adr-008]] — un-promoted elements are declared by record blocks.

[[adr-009]] — a parent owns the edges among its children.

## References

[[module-document-core]] — the module this component is composed into.
