---
title: "C4 Model"
description: A subcommand that emits the corpus's C4 model — its systems,
  containers, modules and components as a tree, and the relationships
  between them as one edge list — assembled from the architecture
  blueprints and their diagrams
slug: frd-spec-c4
type: frd
status: approved
sources: []
---

# C4 model

## Overview

A corpus that follows [[c4-standards]] describes an architecture, but no
artifact states it. The elements are scattered across system, container,
module, and component blueprints and the record blocks inside them. The
relationships between them exist only as `Rel(...)` lines in Mermaid
diagrams. A reader who wants to know what the system is made of and what
depends on what must open every blueprint and reassemble the model by
hand. That reader can be a person, an agent, or the Architecture Catalogue
named in [[container-spec-cli]], and gets a different answer each time the
corpus changes.

This feature adds `specs c4`, which emits that model as one structured
artifact. Elements are nested by containment. Modules are held apart,
because they may be composed into several containers. Relationships form a
flat edge list that carries the document and line where each was drawn.

The model is a projection of the index. The index says which documents are
architecture blueprints and where each record block sits. The command reads
back only what the index points at: each record block through its span,
decoded into the element record's Go type, and the Mermaid C4 diagrams,
parsed per [[adr-029]].

The diagram parser and node resolution belong to [[frd-spec-c4-diagrams]],
which also checks diagrams for conformance. This feature depends on that
one and adds the model built from what it resolves.

### Prerequisites

The index doesn't yet carry one thing this feature projects: recorded
record blocks. They're a prerequisite, not part of this feature.

Recording each registered block's tag and span against its heading is
AC-SPECSTR-001.2 of [[frd-spec-structure]], which waits on the tag
registration in FEAT-SPECSTR-004 of [[frd-spec-structure-conformance]].

The index carries where a record is, not what it says. This command
decodes the content itself. Today the index records no record block at all:
`specs index` on this repository reports zero, including the `component`
block in [[container-spec-cli]]'s roster. Without recorded record blocks,
the tree holds only promoted elements.

This feature reads each document's type-specific frontmatter from the
index: `system`, `modules`, `entity_name`, `kind`, `visibility`,
`provider`, and `contracts_consumed`. The index carries every property a
document's type defines (AC-SPEC-002.1 of [[frd-spec-index]]).

### Out of scope

- People and external systems as elements. An edge touching one is
  reported as an excluded relationship.
- Dynamic and deployment diagrams. They contribute nothing to the model.
- Contract tags in the reference graph. A `contract:` tag isn't a
  wikilink, so `specs refs` on a contract doesn't list the diagrams that
  tag it.
- Other output formats, such as Structurizr DSL or Backstage descriptors.

## Terminology

This feature shares these terms with [[frd-spec-c4-diagrams]], which
defines them, and uses them unchanged: architecture blueprint, element,
entity name, C4 diagram, drawing document, scope of a drawing document,
the diagram statement kinds, drawing, and reserved tag. This feature adds
the following:

- **Drawing document**, as used for edges: per [[adr-009]], it holds the
  edges among its element's children, and draws its element's outward
  edges in a Context view.
- **Relationship**: a dependency edge between two elements, read from a
  relationship statement in a C4 diagram. The source depends on the
  target. A relationship is identified by its source, target, and
  description.
- **Excluded relationship**: a relationship statement that the model
  doesn't admit, recorded with the reason it was excluded.
- **Orphaned element**: a container whose blueprint names no system, or
  a component that no container or module declares in its roster.

## Requirements

### FEAT-C4-001: Element tree

**Acceptance Criteria:**

```gherkin
@FEAT-C4-001
Feature: Element Tree

  As an agent or architect reading an unfamiliar corpus
  I want every architectural element in one tree that mirrors C4 containment
  So that I can see what the system is made of without opening each blueprint

  Containment follows the edges the corpus already declares: a container's
  system property, a host's modules list, and a component's place in its
  parent's roster section. Nothing is inferred from directory layout. A
  module may have several hosts, so modules form their own collection,
  each host lists them by slug, and each module's contained_by is derived
  from those lists.

  An element is one slug, however many documents declare it. Its
  properties come from its blueprint where it has one, and otherwise from
  the declaration whose path sorts first.

  @AC-C4-001.1
  Scenario: Containers nest under the system they declare
    Given a container blueprint whose system property names a system blueprint
    When the user runs specs c4
    Then that container appears among the containers of that system

  @AC-C4-001.2
  Scenario: A container that declares no system is orphaned
    Given a container blueprint with no system property
    When the user runs specs c4
    Then that container appears in the model's orphaned containers

  @AC-C4-001.3
  Scenario: Components nest under the parent whose roster declares them
    Given a container or module whose roster section holds a component record block or a component blueprint link
    When the user runs specs c4
    Then that component appears among the components of that container or module

  @AC-C4-001.4
  Scenario: A module appears once and its hosts are derived
    Given two container blueprints whose modules each name the same module
    When the user runs specs c4
    Then the module appears exactly once in the model's modules
    And each of the two containers lists the module's slug among its modules
    And the module's contained_by names both containers

  @AC-C4-001.5
  Scenario: An element lists every declaration
    Given an element declared by its own blueprint and by a record block in its parent's roster
    When the user runs specs c4
    Then the element's declarations list both, each with its document's path and the span of the frontmatter or the record block

  @AC-C4-001.6
  Scenario: An element with a blueprint takes its properties from the blueprint
    Given an element whose blueprint and record block disagree on its description
    When the user runs specs c4
    Then the element carries the blueprint's description

  @AC-C4-001.7
  Scenario: Element properties are copied from the declaration
    Given an element whose declaration sets entity_name, description, technology and contracts_consumed
    When the user runs specs c4
    Then the element in the model carries each of those values unchanged

  @AC-C4-001.8 @negative
  Scenario: No element is created from a diagram alone
    Given a C4 diagram node whose label matches no declared element
    When the user runs specs c4
    Then the model contains no element for that node

  @AC-C4-001.9
  Scenario: Record blocks in two documents yield one element
    Given two record blocks declaring the same element slug, in two documents, and no blueprint for it
    When the user runs specs c4
    Then the model holds one element with that slug, listing both declarations and taking its properties from the one whose path sorts first

  @AC-C4-001.10
  Scenario: Provided contracts are derived from each contract's provider
    Given a contract whose provider names a container
    When the user runs specs c4
    Then the container's contracts_provided lists that contract

  @AC-C4-001.11
  Scenario: A component no roster declares is orphaned
    Given a component blueprint that no container or module declares in its roster
    When the user runs specs c4
    Then that component appears in the model's orphaned components
```

### FEAT-C4-002: Relationships from blueprint diagrams

**Acceptance Criteria:**

```gherkin
@FEAT-C4-002
Feature: Relationships From Blueprint Diagrams

  As an agent planning a change
  I want every relationship between elements, with where it was drawn
  So that I can see what depends on the element I am about to change

  Relationships are read from the C4 diagrams in the two homes the
  nodes-and-edges decision gives an edge: the parent that composes both
  endpoints, and an element's own Context view for its outward edges. A
  diagram node resolves to an element as the diagram conformance feature
  specifies: by entity name, within the drawing document's scope.
  Direction always means dependency: the source depends on the target.

  @AC-C4-002.1
  Scenario Outline: Rel arguments become relationship properties
    Given a C4 diagram containing the statement <statement>
    And both aliases match elements in scope
    When the user runs specs c4
    Then the model holds a relationship with description "Reads documents", "<technology>" and "<detail>"

    Examples:
      | statement                                               | technology              | detail           |
      | Rel(a, b, "Reads documents")                            | no technology           | no detail        |
      | Rel(a, b, "Reads documents", "filesystem")              | technology filesystem   | no detail        |
      | Rel(a, b, "Reads documents", "filesystem", "At start")  | technology filesystem   | detail At start  |

  @AC-C4-002.2
  Scenario: Directional layout variants are the same relationship
    Given a C4 diagram containing Rel_U, Rel_D, Rel_L or Rel_R from a to b
    When the user runs specs c4
    Then the model holds a relationship from a's element to b's element

  @AC-C4-002.3
  Scenario: Rel_Back reverses the dependency
    Given a C4 diagram containing Rel_Back from a to b
    When the user runs specs c4
    Then the model holds a relationship whose source is b's element and whose target is a's element

  @AC-C4-002.4
  Scenario: A relationship drawn in both of its homes is one relationship
    Given the same source, target and description drawn in a parent's diagram and in a child's Context view
    When the user runs specs c4
    Then the model holds one relationship listing both drawings, each with its path, span, diagram type and aliases

  @AC-C4-002.5
  Scenario: A relationship drawn only in a Context view is admitted
    Given an orphaned container whose Context view draws a relationship to another declared container
    When the user runs specs c4
    Then the model holds that relationship

  @AC-C4-002.6
  Scenario Outline: A relationship records its level from its endpoints
    Given a relationship from a "<source kind>" to a "<target kind>"
    When the user runs specs c4
    Then the relationship's level is "<level>"

    Examples:
      | source kind | target kind | level     |
      | system      | system      | system    |
      | container   | system      | container |
      | container   | container   | container |
      | component   | container   | component |
      | component   | module      | component |
      | component   | component   | component |

  @AC-C4-002.9
  Scenario: Drawings of one relationship that disagree keep the first
    Given one relationship drawn twice with different technology arguments
    When the user runs specs c4
    Then the relationship carries the technology of the drawing that sorts first by path and then by span
    And both drawings are listed in its drawn_in

  @AC-C4-002.10
  Scenario Outline: Edges the model does not admit are excluded
    Given a relationship statement where "<condition>"
    When the user runs specs c4
    Then the model's excluded relationships hold that statement with reason "<reason>", its path and its span
    And the model's relationships do not include it

    Examples:
      | condition                                           | reason            |
      | an endpoint is an external node                     | external_endpoint |
      | an endpoint is a passive store                      | passive_store     |
      | an endpoint's label matches no element in scope     | undeclared_node   |
      | an endpoint's label matches two elements in scope   | ambiguous_node    |

  @AC-C4-002.11 @negative
  Scenario: Dynamic and deployment diagrams contribute no relationships
    Given a C4Dynamic or C4Deployment diagram containing Rel statements
    When the user runs specs c4
    Then no relationship and no excluded relationship is recorded from that diagram

  @AC-C4-002.12 @negative
  Scenario: Diagrams outside architecture blueprints contribute no relationships
    Given a C4 diagram in a feature blueprint
    When the user runs specs c4
    Then no relationship and no excluded relationship is recorded from that diagram
```

### FEAT-C4-003: Reserved relationship tags

**Acceptance Criteria:**

```gherkin
@FEAT-C4-003
Feature: Reserved Relationship Tags

  As an architect documenting how two elements interact
  I want to state on the diagram what Mermaid has no argument for
  So that the model records whether a call is asynchronous and which contract governs it

  The tag syntax, and the checks that keep tags well formed, belong to the
  diagram conformance feature. This feature records what a well-formed
  tag means in the model.

  @AC-C4-003.1
  Scenario Outline: The interaction tag sets the interaction
    Given the statement Rel(a, b, "Uses", $tags="<tags>")
    When the user runs specs c4
    Then the relationship's interaction is "<interaction>"

    Examples:
      | tags  | interaction |
      | sync  | sync        |
      | async | async       |

  @AC-C4-003.2
  Scenario: A contract tag names the governing contract
    Given the statement Rel(a, b, "Uses", $tags="contract:contract-orders")
    When the user runs specs c4
    Then the relationship's contract is contract-orders with source tag

  @AC-C4-003.3
  Scenario: An untagged relationship derives its contract when exactly one fits
    Given a relationship with no contract tag
    And exactly one contract in the source's contracts_consumed names the target as its provider
    When the user runs specs c4
    Then the relationship's contract is that contract's slug with source derived

  @AC-C4-003.4 @negative
  Scenario: No contract is derived when more than one fits
    Given a relationship with no contract tag
    And two contracts in the source's contracts_consumed name the target as their provider
    When the user runs specs c4
    Then the relationship carries no contract

  @AC-C4-003.5
  Scenario: Unreserved tags pass through
    Given the statement Rel(a, b, "Uses", $tags="async+audited")
    When the user runs specs c4
    Then the relationship's tags are exactly audited
```

### FEAT-C4-004: Output

**Acceptance Criteria:**

```gherkin
@FEAT-C4-004
Feature: Output

  As a pipeline publishing the architecture to a catalogue
  I want a versioned, deterministic artifact in the format my consumer reads
  So that a diff of two runs shows a change in the architecture and nothing else

  The artifact's shape is declared once in Strata and emitted as JSON Schema,
  so a consumer validates against the same definition the CLI emits from.
  Emitting the model never fails on corpus content: what cannot be placed is
  reported in the artifact, and judging it is the diagram checks' job.

  YAML is the default, as [[cli-output-standards]] requires of every
  structured answer, and --json selects JSON. There is no other format flag.

  Under --root the model keeps its shape. The root's ancestors appear only
  as the path down to it, holding no other children, and an excluded
  relationship
  is kept when its statement sits in a document inside the subtree. A root
  naming no element is a fact about the corpus, not a failed invocation, so
  it is answered with an empty model that says the root is not declared.

  @AC-C4-004.1
  Scenario: The artifact is YAML by default
    When the user runs specs c4 with no format flag
    Then standard output is a YAML document valid against the C4 model schema

  @AC-C4-004.2
  Scenario: The json flag selects JSON
    When the user runs specs c4 with --json
    Then standard output is a JSON document valid against the C4 model schema
    And it holds the same model as the YAML output for the same corpus

  @AC-C4-004.3
  Scenario: The artifact names its schema version and workspace
    When the user runs specs c4
    Then the artifact carries a schema_version and the workspace name from the configuration

  @AC-C4-004.4
  Scenario: Output is byte-identical for an unchanged corpus
    Given a corpus that has not changed between two runs
    When the user runs specs c4 twice
    Then the two outputs are byte-identical

  @AC-C4-004.5
  Scenario Outline: Root narrows the model to one element's subtree
    Given a corpus with two elements of kind "<kind>"
    When the user runs specs c4 with --root naming one of them
    Then the model holds that element and "<descendants>", and only relationships whose endpoints are both inside that set

    Examples:
      | kind      | descendants                                     |
      | system    | its containers and everything they hold         |
      | container | its components and the modules composed into it |
      | module    | its components and the modules composed into it |
      | component | nothing else                                    |

  @AC-C4-004.6
  Scenario: An unknown root is reported
    When the user runs specs c4 with --root naming a slug that is not an element
    Then standard output is a model holding no element or relationship, with a property stating that the root is not declared
    And the command exits with status zero

  @AC-C4-004.7
  Scenario: Excluded relationships do not fail the command
    Given a corpus whose diagrams hold excluded relationships
    When the user runs specs c4
    Then the command exits with status zero
```

## References

- [[adr-029]]: relationships are read from blueprint diagrams
- [[frd-spec-c4-diagrams]]: the diagram parser and node resolution that
  this feature builds on
- [[adr-009]]: an edge's two homes, the parent and the Context view
- [[adr-008]]: record blocks declare un-promoted elements
- [[c4-standards]]: levels, naming, diagram types, and arrow direction
- [[frd-spec-index]]: the index this feature projects, and the unverified
  severity of AC-SPEC-006.3
- [[frd-spec-structure]] and [[frd-spec-structure-conformance]]: record
  block parsing and registration, which this feature depends on
- [[cli-output-standards]]: YAML by default, `--json`, and stable order.
  Its command table gains a `c4` row when the command lands.
- [[cli-exit-status-standards]]: why a root naming no element exits zero
- [[container-spec-cli]]: the CLI and the Architecture Catalogue that
  consumes its artifacts
- `schemas/c4.strata`: the declared shape of the emitted model
