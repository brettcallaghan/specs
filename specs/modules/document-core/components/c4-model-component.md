---
title: "C4 Model"
entity_name: C4 Model
technology: Go package
description: Assembly of the corpus's C4 model from the resolved elements
  and drawings — the element tree, modules, derived edges, relationships,
  exclusions and the subtree query. Change it to change where an element is
  placed, what an element or relationship carries, or how the model is
  narrowed to a root.
slug: component-c4-model
type: component-blueprint
status: ready-for-review
kind: standard
contracts_consumed: []
source_paths:
  - "cli/core/c4model"
test_paths:
  - "cli/core/c4model"
---

# C4 Model

## Responsibility

C4 Model assembles the C4 model the corpus declares: the element tree, the
module collection, the relationships and the statements excluded from them.
It decides nothing about which elements exist or what a node names, and
returns the generated `specmodel.C4Model` without rendering it. It names no
document type, property or heading of its own.

### Owns

Where each element is placed: under its parent, among the modules, or in an
orphaned collection.

The edges the model derives over the whole corpus: composed modules, the
hosts that contain a module, and the contracts an element provides.

Which declaration and which drawing an element's or a relationship's
properties are taken from.

The reading of reserved relationship tags into an interaction and a
governing contract.

Narrowing the model to one root and its descendants.

### Hides

How ties between several declarations or drawings are broken. Callers see
one element per slug and one relationship per group of drawings.

### Does not own

Which elements exist and what a node or statement resolves to —
[[component-c4-elements]] owns it.

The model's shape, declared in `schemas/c4.strata` — [[module-document-core]]
owns the generated models.

Rendering the model — [[container-spec-cli]] owns the `c4` command.

### Composed into

[[module-document-core]] — a Go package the Corpus calls to answer the `c4`
query; see that blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of `cli/core/c4model`, entered through `Build`; the
record it returns is `specmodel.C4Model`.

### Provides

`Build` SHALL place every element by the edges the corpus declares, never by
directory layout or a diagram.

`Build` SHALL NOT create an element from a diagram node.

WHERE several declarations or drawings could supply a property, `Build`
SHALL take it from the one that sorts first by path and then by span.

WHEN the model is narrowed to a root, its derived properties SHALL keep
their whole-corpus values.

`Build` SHALL order every collection of elements and of slugs by slug.

### Requires

WHEN `Build` is called, the caller SHALL supply in `Input` the same element
input C4 Elements resolves from, and the configured workspace name.

## Design

### Patterns and principles

Functional Core, Imperative Shell (Bernhardt, 2012) — `Build` is a pure
assembly over the records it's given; reading sources is the caller's,
through the functions in its input. It serves the `c4` query and the tests
that build models from in-memory corpora.

### Invariants

`Build` SHALL NOT create an element from a diagram alone. Enforced by:
`TestNoElementIsCreatedFromADiagramAlone`.

`Build` SHALL keep derived properties whole when narrowing. Enforced by:
`TestNarrowingKeepsDerivedPropertiesWhole`.

The component SHALL hold no literal naming registered vocabulary, under
[[module-document-core]]'s vocabulary invariant. Enforced by:
`TestTheExtensionPointNamesNoRegisteredVocabulary` in
`cli/core/corpus/profile_test.go`.

### Extension points

A new element or relationship property — a field in `schemas/c4.strata`,
copied from a declaration or a drawing by the first-by-path rule. Not by a
second resolution of nodes in this package.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN the model needs to know what a node or a statement names, it SHALL
read the answer from the element set rather than resolve it again.

### Verification gates

#### Verify always

`cd cli && go test ./core/c4model/`

#### Verify conditionally

WHEN placement or a property changes, run `cd cli && go test ./profile/`,
which builds this repository's own model.

WHEN what the `c4` command prints changes, run
`cd cli && go test ./cmd/ -run TestSnapshotCases`.

## Key decisions

[[adr-029]] — relationships are read from blueprint diagrams, with reserved
tags.

[[adr-030]] — the model is assembled in this module from a profile's
architecture registration.

[[adr-009]] — a parent owns the edges among its children.

## References

[[module-document-core]] — the module this component is composed into.
