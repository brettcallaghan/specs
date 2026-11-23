---
title: "C4 Model"
entity_name: C4 Model
technology: Go package
description: Assembly of the corpus's C4 model (the element tree, the module
  collection, the edges derived over the whole corpus, the relationships, and
  the excluded relationships) from the resolved C4 elements and relationship
  statements, and its narrowing to a root. Change it to change where a C4
  element is placed, what a C4 element or relationship carries, or how the
  model is narrowed to a root.
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

C4 Model assembles the C4 model that the corpus declares. The model is one
record holding the element tree (each system's containers and each container's
components), the module collection, the relationships between C4 elements, and
the diagram statements excluded from those relationships. Each C4 element in
the model is a system, container, module, or component that a corpus document
declares.

C4 Model decides nothing about which C4 elements exist or what a diagram node
names. It returns the model as `specmodel.C4Model`, a struct type generated
from `schemas/c4.strata`, and doesn't render it. It names no document type,
property, or heading of its own: it reads them through the vocabulary that the
active profile registers.

### Owns

Where each C4 element is placed: among the systems, under its parent, among
the modules, or in an orphaned collection. An orphaned collection is a
top-level list for a container that names no declared system, or for a
component that no Components section declares.

The edges that the model derives over the whole corpus: the hosts that
contain a module, read from the modules each host composes, and the
contracts a C4 element provides. A host is a container or module that
composes a module.

Which declaration and which resolved C4 relationship statement a C4 element's
or a relationship's properties are taken from. A declaration is a blueprint's
front matter or a record block in another document. A resolved C4
relationship statement is one whose two endpoints each name exactly one C4
element.

How the reserved relationship tags (`sync`, `async` and `contract:<slug>`)
are read into an interaction and a governing contract.

Narrowing the model to one root C4 element and its descendants.

### Hides

How ties between several declarations or resolved C4 relationship statements
are broken. Callers see one C4 element per slug (a C4 element's identifier)
and one relationship per group of resolved C4 relationship statements.

### Does not own

Which C4 elements exist and what a node or statement resolves to —
[[component-c4-elements]] owns it.

The model's shape, declared in `schemas/c4.strata` — [[module-document-core]]
owns the generated models.

Rendering the model — [[container-spec-cli]] owns the `c4` command.

### Composed into

[[module-document-core]] — a Go package that the Corpus component calls to
answer the `c4` query; see that blueprint for this component's
relationships.

## Interface

### Surface

The exported identifiers of the `cli/core/c4model` package. The entry point
is the `Build` function, and it returns a `specmodel.C4Model`.

### Provides

`Build` SHALL place every C4 element by the edges the corpus declares, never
by directory layout or a diagram.

`Build` SHALL NOT create a C4 element from a diagram node.

WHERE several declarations or relationship statements could supply a property,
`Build` SHALL take it from the C4 element's blueprint where it has one, and
otherwise from the one that sorts first by path and then by span.

WHEN the model is narrowed to a root, its derived properties SHALL keep
their whole-corpus values.

`Build` SHALL order every collection of C4 elements and of slugs by slug.

### Requires

WHEN `Build` is called, the caller SHALL supply in the `Input` struct the same
`c4elements.Input` value that C4 Elements resolves from.

WHEN `Build` is called, the caller SHALL supply in the `Input` struct the
configured workspace name.

## Design

### Patterns and principles

Functional Core, Imperative Shell (Bernhardt, "Boundaries", 2012) — `Build`
is a pure assembly over the records it's given. The caller does the reading,
through the `Source` and `Decode` function fields it passes in
`Input.Elements`. `Build` serves the `c4` query and the tests that build
models from in-memory corpora.

### Invariants

`Build` SHALL NOT create a C4 element from a diagram alone. Enforced by:
`TestNoElementIsCreatedFromADiagramAlone`.

`Build` SHALL keep derived properties whole when narrowing. Enforced by:
`TestNarrowingKeepsDerivedPropertiesWhole` for `contained_by`; review only for
`contracts_provided`.

The component SHALL hold no literal that names registered vocabulary, as
[[module-document-core]]'s vocabulary invariant requires. Enforced by:
`TestTheExtensionPointNamesNoRegisteredVocabulary` in
`cli/core/corpus/profile_test.go` for string literals naming document types,
statuses, reference properties and criterion qualifiers; review only
for the rest.

### Extension points

A new C4 element or relationship property — add a field in `schemas/c4.strata`
and copy it from the same declaration or relationship statement as the existing
properties. Not by a second resolution of nodes in this package.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN the model needs to know what a node or a statement names, it SHALL
read the answer from the `c4elements.Set` that C4 Elements builds rather than
resolve it again.

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
