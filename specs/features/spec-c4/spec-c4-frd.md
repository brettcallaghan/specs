---
title: "FRD: C4 Model"
description: A subcommand that emits the corpus's C4 model — its systems,
  containers, modules and components as a tree, and the relationships
  between them as one edge list — assembled from the architecture
  blueprints and their diagrams, plus a check that keeps those diagrams
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

The same parser gives the corpus its first diagram check. A diagram whose
nodes do not match the elements the blueprints declare is today invisible
until a human notices; the check makes that mismatch a finding at the
offending line.

The index does not yet carry two things this feature projects, and both
are prerequisites rather than part of it:

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

## Terminology

- **Architecture blueprint** — a corpus member of type
  `system-blueprint`, `container-blueprint`, `module-blueprint` or
  `component-blueprint`.
- **Element** — a system, container, module or component declared either
  by its own architecture blueprint or by a record block in its parent's
  roster section.
- **C4 diagram** — a fenced `mermaid` block whose first statement is
  `C4Context`, `C4Container` or `C4Component`. `C4Dynamic` and
  `C4Deployment` blocks are not C4 diagrams for this feature.
- **Drawing document** — the document a C4 diagram sits in. Per
  [[adr-009]] it owns the edges among its element's children, and it
  draws its element's outward edges in a Context view.
- **Scope of a drawing document** — the elements a diagram node in it may
  name: the document's own element; the components its roster declares
  and the modules that name it in `contained_by`; and every system and
  container in the corpus.
- **Diagram node** — an element statement in a C4 diagram, such as
  `Container(alias, "Label", ...)`, identified within its diagram by its
  alias.
- **Relationship** — a dependency edge between two elements, read from a
  `Rel` statement in a C4 diagram: the source depends on the target.
- **Unplaced edge** — a `Rel` statement the model does not admit, because
  an endpoint is a person or external system, or a diagram node matches
  no element or more than one.
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
  Scenario Outline: Each element records how it is declared
    Given an element declared by "<declaration>"
    When the user runs specs c4
    Then the element's declaration kind is "<kind>" and carries the path and span of that declaration

    Examples:
      | declaration                  | kind         |
      | its own architecture blueprint | blueprint  |
      | a record block in a roster     | record-block |

  @AC-C4-001.6
  Scenario: Element properties are copied from the declaration
    Given an element whose declaration sets title, description, technology, contracts_provided and contracts_consumed
    When the user runs specs c4
    Then the element in the model carries each of those values unchanged

  @AC-C4-001.7 @negative
  Scenario: No element is created from a diagram alone
    Given a C4 diagram node whose label matches no declared element
    When the user runs specs c4
    Then the model contains no element for that node
```

### FEAT-C4-002: Relationships From Parent Diagrams

**Acceptance Criteria:**

```gherkin
@FEAT-C4-002
Feature: Relationships From Parent Diagrams

  As an agent planning a change
  I want every relationship between elements, with where it was drawn
  So that I can see what depends on the element I am about to change

  A relationship is read from the C4 diagrams of the document that owns it,
  per the parent-owns-child-edges rule. A diagram node is matched to an
  element by its label, which the C4 standard requires to be the element's
  display name verbatim, and only among the elements in the scope of the
  drawing document. Scoping is what lets two containers each hold a
  component named Parser. Direction always means dependency: the source
  depends on the target.

  A relationship is identified by its source, target and description. Two
  drawings that agree on all three are one relationship, whatever else
  they say.

  People, external systems and deployment detail are out of scope for this
  version. An edge touching one is reported as unplaced rather than dropped,
  so its absence from the model is visible.

  @AC-C4-002.1
  Scenario Outline: Rel arguments become relationship properties
    Given a C4 diagram containing the statement <statement>
    And both aliases match declared elements
    When the user runs specs c4
    Then the model holds a relationship with description "<description>", technology "<technology>" and detail "<detail>"

    Examples:
      | statement                                             | description | technology | detail  |
      | Rel(a, b, "Reads documents")                          | Reads documents |        |         |
      | Rel(a, b, "Reads documents", "filesystem")            | Reads documents | filesystem |     |
      | Rel(a, b, "Reads documents", "filesystem", "At start") | Reads documents | filesystem | At start |

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
  Scenario: One relationship drawn in two diagrams is one relationship
    Given the same source, target and description drawn in a parent's internal diagram and in a child's context diagram
    When the user runs specs c4
    Then the model holds one relationship listing both drawings, each with its path, span, diagram type and aliases

  @AC-C4-002.5
  Scenario Outline: A relationship records its level from its endpoints
    Given a relationship from a "<source kind>" to a "<target kind>"
    When the user runs specs c4
    Then the relationship's level is "<level>"

    Examples:
      | source kind | target kind | level     |
      | container   | container   | container |
      | container   | system      | container |
      | component   | component   | component |
      | component   | module      | component |
      | component   | container   | component |

  @AC-C4-002.8
  Scenario: A component label resolves within the drawing document's scope
    Given two containers whose rosters each declare a component titled Parser
    And a C4Component diagram in the first container's blueprint with a node labelled Parser
    When the user runs specs c4
    Then that node resolves to the first container's Parser component

  @AC-C4-002.9
  Scenario: Drawings of one relationship that disagree keep the first
    Given one relationship drawn twice with different technology arguments
    When the user runs specs c4
    Then the relationship carries the technology of the drawing that sorts first by path and then by position
    And both drawings are listed in its drawn_in

  @AC-C4-002.6
  Scenario Outline: Edges the model does not admit are unplaced
    Given a Rel statement where "<condition>"
    When the user runs specs c4
    Then the model's unplaced list holds that statement with reason "<reason>", its path and its span
    And the model's relationships do not include it

    Examples:
      | condition                                          | reason            |
      | an endpoint is a Person or an external System      | external_endpoint |
      | an endpoint's label matches no element in scope    | undeclared_node   |
      | an endpoint's label matches two elements in scope  | ambiguous_node    |

  @AC-C4-002.7 @negative
  Scenario: Dynamic and deployment diagrams contribute no relationships
    Given a C4Dynamic or C4Deployment diagram containing Rel statements
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
  nothing at render time.

  A contract tag is not a wikilink, so the reference graph does not record
  it: specs refs on a contract does not list the diagrams that tag it.
  Recording it there is out of scope here.

  @AC-C4-003.1
  Scenario Outline: The interaction tag sets the interaction
    Given a Rel statement tagged "<tags>"
    When the user runs specs c4
    Then the relationship's interaction is "<interaction>"

    Examples:
      | tags  | interaction |
      | sync  | sync        |
      | async | async       |

  @AC-C4-003.6
  Scenario Outline: Positional and named tags are read alike
    Given the statement <statement>
    When the user runs specs c4
    Then the relationship's interaction is async and its tags are exactly audited

    Examples:
      | statement                                                      |
      | Rel(a, b, "Uses", "HTTP", "", "", "async+audited")             |
      | Rel(a, b, "Uses", "HTTP", $tags="async+audited")               |

  @AC-C4-003.2
  Scenario: A contract tag names the governing contract
    Given a Rel statement tagged contract followed by a contract's slug
    When the user runs specs c4
    Then the relationship's contract is that slug with source tag

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
    Given a Rel statement tagged async plus audited
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
  reported in the artifact, and judging it is the check's job.

  YAML is the default, as for every other query command, and --json selects
  JSON. There is no other format flag.

  @AC-C4-004.1
  Scenario: The artifact is YAML by default
    When the user runs specs c4 with no format flag
    Then standard output is a YAML document valid against the C4 model schema

  @AC-C4-004.7
  Scenario: The json flag selects JSON
    When the user runs specs c4 with --json
    Then standard output is a JSON document valid against the C4 model schema
    And it holds the same model as the YAML output for the same corpus

  @AC-C4-004.2
  Scenario: The artifact names its schema version and workspace
    When the user runs specs c4
    Then the artifact carries a schema_version and the workspace name from the configuration

  @AC-C4-004.3
  Scenario: Output is byte-identical for an unchanged corpus
    Given a corpus that has not changed between two runs
    When the user runs specs c4 twice
    Then the two outputs are byte-identical

  @AC-C4-004.4
  Scenario Outline: Root narrows the model to one element's subtree
    Given a corpus with two elements of kind "<kind>"
    When the user runs specs c4 with --root naming one of them
    Then the model holds that element and "<descendants>", and only relationships whose endpoints are both inside that set

    Examples:
      | kind      | descendants                                     |
      | system    | its containers and everything they hold         |
      | container | its components and the modules composed into it |
      | module    | its components                                  |
      | component | nothing else                                    |

  @AC-C4-004.5
  Scenario: An unknown root is a usage error
    When the user runs specs c4 with --root naming a slug that is not an element
    Then the command exits with the usage error status and names the slug

  @AC-C4-004.6
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

  The check is one rule, c4/diagram, run by specs check alongside the
  existing rules. Every finding carries the span of the offending statement.
  Mermaid diagrams that are not C4 diagrams are outside what the parser
  reads, so the rule reports them as unverified rather than as passing.

  @AC-C4-005.1
  Scenario Outline: Diagram defects are findings
    Given a C4 diagram where "<defect>"
    When the user runs specs check on that document
    Then the report holds a c4/diagram finding of severity "<severity>" at that statement's span

    Examples:
      | defect                                                   | severity |
      | a statement does not parse                               | error    |
      | a Container or Component label matches no element in scope | error  |
      | a label matches more than one element in scope           | error    |
      | a BiRel statement draws a bidirectional relationship     | error    |
      | two Rel statements join the same pair of aliases         | error    |
      | one Rel statement is tagged both sync and async          | error    |
      | a contract tag names a slug that is not a contract       | error    |
      | a contract tag names a contract the source does not consume or the target does not provide | warning |
      | two drawings of one relationship disagree on technology, detail or tags | warning |
      | a tags value contains a comma, which joins nothing       | warning  |

  @AC-C4-005.2
  Scenario: A Mermaid diagram the parser does not read is unverified
    Given a mermaid block that is not a C4 diagram, including a C4Dynamic or C4Deployment block
    When the user runs specs check on that document
    Then the report holds a c4/diagram finding of severity unverified naming the diagram type

  @AC-C4-005.3
  Scenario: A quoted argument may contain parentheses and commas
    Given a C4 diagram containing Component(x, "X", "Go package (planned), internal", "desc")
    When the user runs specs check on that document
    Then the report holds no c4/diagram finding for that statement
```

## References

- [[adr-029]] — relationships are read from the parent's diagram
- [[adr-009]] — parents own the edges among their children
- [[adr-008]] — record blocks declare un-promoted elements
- [[c4-standards]] — levels, naming, diagram types and arrow direction
- [[frd-spec-index]] — the index this feature projects
- [[frd-spec-structure]] and [[frd-spec-structure-conformance]] — record
  block parsing and registration, which this feature depends on
- [[container-spec-cli]] — the CLI and the Architecture Catalogue that
  consumes its artifacts
- `schemas/c4.strata` — the declared shape of the emitted model
