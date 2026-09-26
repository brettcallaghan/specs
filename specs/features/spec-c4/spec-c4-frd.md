---
title: "FRD: C4 Model"
description: A subcommand that emits the corpus's C4 model — its systems,
  containers, modules and components as a tree, and the relationships
  between them as one edge list — assembled from the architecture
  blueprints and their diagrams, plus checks that keep those diagrams
  consistent with the elements the blueprints declare
slug: frd-spec-c4
type: frd
status: draft
updated_at: 2026-09-26T00:00:00Z
sources: []
---

# C4 Model

## Overview

A corpus that follows [[c4-standards]] describes an architecture, but no
artifact states it. The elements are scattered across system, container,
module and component blueprints and the record blocks inside them; the
relationships between them exist only as `Rel(...)` lines in Mermaid
diagrams. A reader — human, agent or the Architecture Catalogue named in
[[container-spec-cli]] — who wants to know what the system is made of and
what depends on what must open every blueprint and reassemble the model
by hand, and gets a different answer each time the corpus changes.

This feature adds `specs c4`, which emits that model as one structured
artifact: elements nested by containment, modules held apart because they
may be composed into several containers, and relationships as a flat edge
list carrying the document and line each was drawn at. It is a
projection of the index — no document is read that the index has not
already read — with Mermaid C4 diagrams as the one new input, parsed per
[[adr-029]].

The same parser gives the corpus its first diagram checks. A diagram
whose nodes do not match the elements the blueprints declare is today
invisible until a human notices; the checks make that mismatch a finding
at the offending line.

### Prerequisites

The index does not yet carry two things this feature projects. Both are
prerequisites rather than part of it:

- **Parsed record blocks.** Parsing them is AC-SPECSTR-001.2 of
  [[frd-spec-structure]], which waits on the tag registration in
  FEAT-SPECSTR-004 of [[frd-spec-structure-conformance]]. Today the index
  records no record block at all — `specs index` on this repository
  reports zero, including the `component` block in
  [[container-spec-cli]]'s roster. Without them the tree holds only
  promoted elements.
- **Each document's type-specific frontmatter.** An index entry carries
  the base envelope only — `IndexEntry.frontmatter` is typed
  `DocumentFrontmatter` — so `system`, `contained_by`, `kind`,
  `visibility` and the contract edges never reach it. The entry for
  [[container-spec-cli]] omits the `contracts_provided` its file
  declares. Carrying them is an addition to the index, which the
  additive rule in [[container-spec-cli]] admits under `schema_version`
  1.0.

### Out of scope

- **People and external systems** as elements. An edge touching one is
  reported as unplaced.
- **Dynamic and deployment diagrams.** They contribute nothing to the
  model and are reported as unverified by the checks.
- **Agreement between a diagram node's technology or description and its
  element's declaration.** [[c4-standards]] requires both to match; this
  feature matches nodes by name only.
- **Contract tags in the reference graph.** A `contract:` tag is not a
  wikilink, so `specs refs` on a contract does not list the diagrams that
  tag it.
- **Other output formats**, such as Structurizr DSL or Backstage
  descriptors.

## Terminology

- **Architecture blueprint** — a corpus member of type
  `system-blueprint`, `container-blueprint`, `module-blueprint` or
  `component-blueprint`.
- **Element** — a system, container, module or component declared either
  by its own architecture blueprint or by a record block in its parent's
  roster section.
- **Display name** — an element's name as diagrams draw it: a record
  block's `title`, or a blueprint's `title` with its `<Type> Blueprint: `
  prefix removed. "Container Blueprint: Spec CLI" displays as "Spec CLI".
- **C4 diagram** — a fenced `mermaid` block in an architecture blueprint
  whose first statement is `C4Context`, `C4Container` or `C4Component`.
- **Drawing document** — the architecture blueprint a C4 diagram sits in.
  Per [[adr-009]] it holds the edges among its element's children, and it
  draws its element's outward edges in a Context view.
- **Scope of a drawing document** — the elements a diagram node in it may
  name: the document's own element; the components its roster declares
  and the modules that name it in `contained_by`; and every system and
  container in the corpus.
- **Diagram node** — an element statement in a C4 diagram, identified
  within its diagram by its alias. The statements are:

  | Kind | Macros |
  |---|---|
  | Element node | `System`, `Container`, `Component`, and their `Db` and `Queue` forms |
  | Passive store | a `Db` or `Queue` node that names no element in scope — a data store the drawing element uses, which [[c4-standards]] gives no blueprint |
  | External node | `Person`, and every `_Ext` form |
  | Boundary | `Boundary`, `Enterprise_Boundary`, `System_Boundary`, `Container_Boundary` — groups nodes and is not itself a node |
  | Relationship | `Rel`; `Rel_U`, `Rel_Up`, `Rel_D`, `Rel_Down`, `Rel_L`, `Rel_Left`, `Rel_R`, `Rel_Right`; `Rel_Back`; `BiRel` |
  | Ignored | `title`, `UpdateElementStyle`, `UpdateRelStyle`, `UpdateLayoutConfig` |

- **Relationship** — a dependency edge between two elements, read from a
  relationship statement in a C4 diagram: the source depends on the
  target. It is identified by its source, target and description.
- **Unplaced edge** — a relationship statement the model does not admit.
- **Reserved tag** — a relationship tag with meaning to this feature:
  `sync`, `async`, or `contract:` followed by a slug.

## Requirements

### FEAT-C4-001: Element Tree

**Acceptance Criteria:**

```gherkin
@FEAT-C4-001
Feature: Element Tree

  As an agent or architect reading an unfamiliar corpus
  I want every architectural element in one tree that mirrors C4 containment
  So that I can see what the system is made of without opening each blueprint

  Containment follows the edges the corpus already declares: a container's
  system property, a module's contained_by list, and a component's place in
  its parent's roster section. Nothing is inferred from directory layout.
  A module may have several hosts, so modules form their own collection and
  each host lists them by slug rather than nesting a copy.

  @AC-C4-001.1
  Scenario: Containers nest under the system they declare
    Given a container blueprint whose system property names a system blueprint
    When the user runs specs c4
    Then that container appears among the containers of that system

  @AC-C4-001.2
  Scenario: A container that declares no system is standalone
    Given a container blueprint with no system property
    When the user runs specs c4
    Then that container appears in the model's standalone containers

  @AC-C4-001.3
  Scenario: Components nest under the parent whose roster declares them
    Given a container or module whose roster section holds a component record block or a component blueprint link
    When the user runs specs c4
    Then that component appears among the components of that container or module

  @AC-C4-001.4
  Scenario: A module appears once and is listed by each host
    Given a module blueprint whose contained_by names two containers
    When the user runs specs c4
    Then the module appears exactly once in the model's modules
    And each of the two containers lists the module's slug among its modules

  @AC-C4-001.5
  Scenario: A promoted element records the blueprint that declares it
    Given an element with its own architecture blueprint
    When the user runs specs c4
    Then the element's declaration kind is blueprint and carries the blueprint's path and status

  @AC-C4-001.6
  Scenario: An unpromoted element records the record block that declares it
    Given an element declared by a record block in a roster section
    When the user runs specs c4
    Then the element's declaration kind is record-block and carries the parent document's path and the block's span

  @AC-C4-001.7
  Scenario: Element properties are copied from the declaration
    Given an element whose declaration sets title, description, technology, contracts_provided and contracts_consumed
    When the user runs specs c4
    Then the element in the model carries each of those values unchanged

  @AC-C4-001.8 @negative
  Scenario: No element is created from a diagram alone
    Given a C4 diagram node whose label matches no declared element
    When the user runs specs c4
    Then the model contains no element for that node

  @AC-C4-001.9
  Scenario: A slug declared twice yields one element
    Given two declarations of the same element slug, in two documents
    When the user runs specs c4
    Then the model holds one element with that slug, taken from the declaration whose path sorts first
```

### FEAT-C4-002: Relationships From Blueprint Diagrams

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
  diagram node is matched by display name, and only among the elements in
  the scope of its drawing document. Scoping is what lets two containers
  each hold a component named Parser. Direction always means dependency:
  the source depends on the target.

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
    Given a standalone container whose Context view draws a relationship to another declared container
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

  @AC-C4-002.7
  Scenario: A label resolves within the drawing document's scope
    Given two containers whose rosters each declare a component titled Parser
    And a C4Component diagram in the first container's blueprint with a node labelled Parser
    When the user runs specs c4
    Then that node resolves to the first container's Parser component

  @AC-C4-002.8
  Scenario: A promoted element is matched by its display name
    Given a container blueprint titled Container Blueprint: Spec CLI
    And a C4 diagram node labelled Spec CLI
    When the user runs specs c4
    Then that node resolves to the container

  @AC-C4-002.9
  Scenario: Drawings of one relationship that disagree keep the first
    Given one relationship drawn twice with different technology arguments
    When the user runs specs c4
    Then the relationship carries the technology of the drawing that sorts first by path and then by position
    And both drawings are listed in its drawn_in

  @AC-C4-002.10
  Scenario Outline: Edges the model does not admit are unplaced
    Given a relationship statement where "<condition>"
    When the user runs specs c4
    Then the model's unplaced list holds that statement with reason "<reason>", its path and its span
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
    Then no relationship and no unplaced entry is recorded from that diagram

  @AC-C4-002.12 @negative
  Scenario: Diagrams outside architecture blueprints contribute no relationships
    Given a C4 diagram in a feature blueprint
    When the user runs specs c4
    Then no relationship and no unplaced entry is recorded from that diagram
```

### FEAT-C4-003: Reserved Relationship Tags

**Acceptance Criteria:**

```gherkin
@FEAT-C4-003
Feature: Reserved Relationship Tags

  As an architect documenting how two elements interact
  I want to state on the diagram what Mermaid has no argument for
  So that the model records whether a call is asynchronous and which contract governs it

  Tags are written with Mermaid's tags parameter, positionally or as the
  named $tags argument, several joined by a plus sign. Mermaid parses a
  tagged relationship without error, so the convention costs an author
  nothing at parse time.

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
    And exactly one contract appears in both the source's contracts_consumed and the target's contracts_provided
    When the user runs specs c4
    Then the relationship's contract is that contract's slug with source derived

  @AC-C4-003.4 @negative
  Scenario: No contract is derived when more than one fits
    Given a relationship with no contract tag
    And two contracts appear in both the source's contracts_consumed and the target's contracts_provided
    When the user runs specs c4
    Then the relationship carries no contract

  @AC-C4-003.5
  Scenario: Unreserved tags pass through
    Given the statement Rel(a, b, "Uses", $tags="async+audited")
    When the user runs specs c4
    Then the relationship's tags are exactly audited

  @AC-C4-003.6
  Scenario Outline: Positional and named tags are read alike
    Given the statement <statement>
    When the user runs specs c4
    Then the relationship's interaction is async and its tags are exactly audited

    Examples:
      | statement                                                      |
      | Rel(a, b, "Uses", "HTTP", "", "", "async+audited")             |
      | Rel(a, b, "Uses", "HTTP", $tags="async+audited")               |
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
  reported in the artifact, and judging it is the checks' job.

  YAML is the default, as for every other query command, and --json selects
  JSON. There is no other format flag.

  Under --root the model keeps its shape. The root's ancestors appear only
  as the path down to it, holding no other children, and an unplaced entry
  is kept when its statement sits in a document inside the subtree.

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
    Then the command names the slug on standard error and exits non-zero

  @AC-C4-004.7
  Scenario: Unplaced edges do not fail the command
    Given a corpus whose diagrams hold unplaced edges
    When the user runs specs c4
    Then the command exits with status zero
```

### FEAT-C4-005: Diagram Conformance

**Acceptance Criteria:**

```gherkin
@FEAT-C4-005
Feature: Diagram Conformance

  As an author editing a blueprint
  I want my C4 diagrams checked against the elements the corpus declares
  So that a diagram and its declarations cannot drift apart unnoticed

  Each defect is its own rule in the c4 namespace, run by specs check
  alongside the existing rules and selectable with --only. Every finding
  carries the span of the offending statement. The duplicate-edge rule
  compares statements within one diagram only, and treats a to b and b to a
  as different pairs.

  Scenario Outline: Each diagram defect is reported under its own rule
    Given a C4 diagram where "<defect>"
    When the user runs specs check on that document
    Then the report holds a "<rule>" finding of severity "<severity>" at that statement's span

    @AC-C4-005.1
    Examples: a statement the parser cannot read
      | defect                                   | rule     | severity |
      | a statement is not one the parser reads  | c4/parse | error    |

    @AC-C4-005.2
    Examples: a node that names no element in scope
      | defect                                                         | rule              | severity |
      | an element node's label matches no element in scope            | c4/undeclared-node | error   |

    @AC-C4-005.3
    Examples: a node that names several elements in scope
      | defect                                                  | rule              | severity |
      | a label matches more than one element in scope          | c4/ambiguous-node | error    |

    @AC-C4-005.4
    Examples: a bidirectional arrow
      | defect                  | rule             | severity |
      | a BiRel statement       | c4/bidirectional | error    |

    @AC-C4-005.5
    Examples: an edge drawn twice in one diagram
      | defect                                                  | rule              | severity |
      | two relationship statements join the same ordered pair  | c4/duplicate-edge | error    |

    @AC-C4-005.6
    Examples: contradictory interaction tags
      | defect                                     | rule                    | severity |
      | one statement is tagged both sync and async | c4/interaction-conflict | error   |

    @AC-C4-005.7
    Examples: a contract tag naming no contract
      | defect                                             | rule                   | severity |
      | a contract tag names a slug that is not a contract | c4/contract-unresolved | error    |

    @AC-C4-005.8
    Examples: a contract tag the endpoints do not declare
      | defect                                                                                     | rule                 | severity |
      | a contract tag names a contract the source does not consume or the target does not provide | c4/contract-mismatch | warning  |

    @AC-C4-005.9
    Examples: drawings of one relationship that disagree
      | defect                                                                   | rule                | severity |
      | two drawings of one relationship disagree on technology, detail or tags  | c4/drawing-conflict | warning  |

    @AC-C4-005.10
    Examples: tags joined with a comma
      | defect                                  | rule             | severity |
      | a tags value contains a comma           | c4/tag-separator | warning  |

  @AC-C4-005.11
  Scenario: A Mermaid diagram the parser does not read is unverified
    Given a mermaid block in an architecture blueprint that is not a C4 diagram, including a C4Dynamic or C4Deployment block
    When the user runs specs check on that document
    Then the report holds a c4/parse finding of severity unverified naming the diagram type

  @AC-C4-005.12
  Scenario: A quoted argument may contain parentheses and commas
    Given a C4 diagram containing Component(x, "X", "Go package (planned), internal", "desc")
    When the user runs specs check on that document
    Then the report holds no c4 finding for that statement

  @AC-C4-005.13 @negative
  Scenario: A passive store is not an undeclared node
    Given a C4 diagram whose ContainerDb node names no element in scope
    When the user runs specs check on that document
    Then the report holds no c4/undeclared-node finding for that node
```

## References

- [[adr-029]] — relationships are read from blueprint diagrams
- [[adr-009]] — an edge's two homes, the parent and the Context view
- [[adr-008]] — record blocks declare un-promoted elements
- [[c4-standards]] — levels, naming, diagram types and arrow direction
- [[frd-spec-index]] — the index this feature projects, and the unverified
  severity of AC-SPEC-006.3
- [[frd-spec-structure]] and [[frd-spec-structure-conformance]] — record
  block parsing and registration, which this feature depends on
- [[cli-output-standards]] — YAML by default, `--json`, and stable order;
  its command table gains a `c4` row when the command lands
- [[container-spec-cli]] — the CLI and the Architecture Catalogue that
  consumes its artifacts
- `schemas/c4.strata` — the declared shape of the emitted model
