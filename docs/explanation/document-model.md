# How the document model works

Specs uses a small, closed set of document types. Each one has a single job,
and together they separate what to build from how to build it. For the list
of types and what each one captures, see
[Document types](../reference/document-types.md).

## A closed set of types

The `spec` skill doesn't create new document types. If a piece of work
doesn't fit one of the existing types, stop and reconsider the work instead
of creating a new type.

## What and how

Product documents describe *what* to build. Every feature starts with an FRD,
which holds the user story and the acceptance criteria and never says how to
build the feature.

Architecture and engineering documents describe *how* to build the system,
and the constraints that the design must respect. They use the
[C4 model](https://c4model.com/), which describes software at the system,
container, and component levels. C4 works for a single application. It's
most useful across an enterprise, where it gives every team the same
vocabulary for describing many systems.

These documents stop at systems, containers, modules, and components. Code
is the last level of the C4 hierarchy, and only Work Orders, which are
temporary, can refer to it. The code documents itself.

## Temporary documents that keep the architecture current

A Feature Blueprint and its Work Orders connect a requirement to the code
that implements it. Both are temporary: you can keep them for audit purposes
or delete them after the change merges.

When the change merges, its architecture changes are copied into the
blueprints that own them. The `groom-feature-blueprint` skill makes sure that
this happens: each Work Order updates the owning blueprint along with the
code and tests, and the last Work Order verifies that every change was
copied. As a result, the architecture specs stay current.

Nobody writes a Work Order by hand. Grooming generates each Work Order to
meet a Definition of Ready. That definition requires a capable planner to
document everything that an executor, whether less capable or equally
capable, needs.

## Where Standards and ADRs apply

You can define Standards and ADRs at any level of the C4 hierarchy: the
system landscape (enterprise), a system, a container or module, or a
component. A Standards document or ADR applies at the level where you define
it and at every level below. For example:

- A system-level standard applies to the system and to every container in
  it.
- A component-level ADR, such as a choice of external framework, applies
  only to that component.

The document's location and the documents that it refers to show its level.
Frontmatter doesn't record the level.

## This repository specifies itself

The FRDs, ADRs, blueprints, and Standards in [`specs/`](../../specs/) are the
ones that produced the CLI in [`cli/`](../../cli/). This repository uses its
own framework to test whether the framework works.
