---
title: "FRD: C4 Diagram Conformance"
description: Checks that keep the Mermaid C4 diagrams in architecture
  blueprints parseable and consistent with the elements the corpus
  declares — every node resolving to one element, every arrow a single
  dependency, every reserved relationship tag well formed
slug: frd-spec-c4-diagrams
type: frd
status: draft
updated_at: 2026-09-29T00:00:00Z
sources: []
---

# C4 Diagram Conformance

## Overview

Architecture blueprints draw their structure as Mermaid C4 diagrams, and
[[c4-standards]] requires every node on those diagrams to be an element
the documents declare, named exactly as declared. Nothing checks it. A
diagram that names a component nobody declares, draws an arrow both
ways, or breaks on a parenthesis inside a quoted argument is invisible
until a human notices, and by then the diagram and its declarations have
drifted.

This feature adds a C4 diagram parser and a family of `c4/*` rules to
`specs check`. Each rule reports one defect at the statement that
carries it. The parser and the node resolution it performs are also the
foundation [[frd-spec-c4]] builds its relationship model on, parsed per
[[adr-029]]; this feature stands without that one.

### Prerequisites

Resolving a node needs the elements it may name, and two of the checks
need an element's contract edges. The index does not yet carry either:

- **Recorded record blocks.** An un-promoted element is declared by a
  `component` record block, which the index records per AC-SPECSTR-001.2
  of [[frd-spec-structure]]. Today it records none. Without them a node
  naming an un-promoted element is reported as undeclared.
- **Each document's type-specific frontmatter.** `entity_name`,
  `contained_by` and the contract edges never reach an index entry,
  which carries the base envelope only. Scope and the contract checks
  depend on them.

### Out of scope

- **Agreement between a node's technology or description and its
  element's declaration.** [[c4-standards]] requires both to match; this
  feature matches nodes by entity name only.
- **Dynamic and deployment diagrams**, which are reported as unverified.
- **Diagrams outside architecture blueprints.** A feature blueprint may
  draw a proposed architecture that no document declares yet.

## Terminology

- **Architecture blueprint** — a corpus member of type
  `system-blueprint`, `container-blueprint`, `module-blueprint` or
  `component-blueprint`.
- **Element** — a system, container, module or component declared either
  by its own architecture blueprint or by a record block in its parent's
  roster section.
- **Entity name** — an element's `entity_name`, from its blueprint's
  frontmatter or its record block. It is the name diagrams draw, and is
  distinct from a blueprint's `title`, which names the document.
- **C4 diagram** — a fenced `mermaid` block in an architecture blueprint
  whose first statement is `C4Context`, `C4Container` or `C4Component`.
- **Drawing document** — the architecture blueprint a C4 diagram sits in.
- **Scope of a drawing document** — the elements a diagram node in it may
  name: the document's own element; the components its roster declares
  and the modules that name it in `contained_by`; and every system and
  container in the corpus.
- **Diagram statement** — one line of a C4 diagram. The parser reads:

  | Kind | Macros |
  |---|---|
  | Element node | `System`, `Container`, `Component`, and their `Db` and `Queue` forms |
  | Passive store | a `Db` or `Queue` node that names no element in scope — a data store the drawing element uses, which [[c4-standards]] gives no blueprint |
  | External node | `Person`, and every `_Ext` form |
  | Boundary | `Boundary`, `Enterprise_Boundary`, `System_Boundary`, `Container_Boundary` — groups nodes and is not itself a node |
  | Relationship | `Rel`; `Rel_U`, `Rel_Up`, `Rel_D`, `Rel_Down`, `Rel_L`, `Rel_Left`, `Rel_R`, `Rel_Right`; `Rel_Back`; `BiRel` |
  | Ignored | `title`, `UpdateElementStyle`, `UpdateRelStyle`, `UpdateLayoutConfig` |

- **Drawing** — one relationship statement. Two drawings are of the same
  relationship when they resolve to the same source and target and carry
  the same label.
- **Reserved tag** — a relationship tag with meaning to this framework:
  `sync`, `async`, or `contract:` followed by a slug. Tags are written in
  Mermaid's tags argument, positionally or as `$tags`, several joined by
  a plus sign.

## Requirements

### FEAT-C4-DIAG-001: C4 Diagram Parsing

**Acceptance Criteria:**

```gherkin
@FEAT-C4-DIAG-001
Feature: C4 Diagram Parsing

  As an author editing a blueprint
  I want every C4 diagram parsed as it will render
  So that a diagram that breaks is reported at the line that breaks it

  The parser reads the statement subset in the terminology and nothing
  else. A statement outside it is a parse error rather than something
  skipped, so the subset grows deliberately. Every finding in this
  feature carries the span of the statement it concerns and a rule
  identifier in the c4 namespace, selectable with --only.

  @AC-C4-DIAG-001.1
  Scenario: A statement the parser cannot read is an error
    Given a C4 diagram containing a statement outside the parsed subset
    When the user runs specs check on that document
    Then the report holds a c4/parse finding of severity error at that statement's span

  @AC-C4-DIAG-001.2
  Scenario: A quoted argument may contain parentheses and commas
    Given a C4 diagram containing Component(x, "X", "Go package (planned), internal", "desc")
    When the user runs specs check on that document
    Then the report holds no c4 finding for that statement

  @AC-C4-DIAG-001.3
  Scenario: A Mermaid diagram the parser does not read is unverified
    Given a mermaid block in an architecture blueprint that is not a C4 diagram, including a C4Dynamic or C4Deployment block
    When the user runs specs check on that document
    Then the report holds a c4/parse finding of severity unverified naming the diagram type

  @AC-C4-DIAG-001.4 @negative
  Scenario: Diagrams outside architecture blueprints are not checked
    Given a C4 diagram in a feature blueprint
    When the user runs specs check on that document
    Then the report holds no c4 finding for that diagram
```

### FEAT-C4-DIAG-002: Node Resolution

**Acceptance Criteria:**

```gherkin
@FEAT-C4-DIAG-002
Feature: Node Resolution

  As an author editing a blueprint
  I want every node on my diagrams to name exactly one declared element
  So that a diagram cannot show something the documents do not describe

  A node resolves by entity name, only among the elements in the scope of
  its drawing document. Scoping is what lets two containers each hold a
  component named Parser. External nodes and passive stores are not
  elements and are never reported as undeclared.

  @AC-C4-DIAG-002.1
  Scenario: A node resolves to the element whose entity name it carries
    Given a container blueprint whose entity_name is Spec CLI
    And a C4 diagram node labelled Spec CLI in a document whose scope includes it
    When the user runs specs check on that document
    Then the report holds no c4 finding for that node

  @AC-C4-DIAG-002.2
  Scenario: A label resolves within the drawing document's scope
    Given two containers whose rosters each declare a component with entity_name Parser
    And a C4Component diagram in the first container's blueprint with a node labelled Parser
    When the user runs specs check on that document
    Then the report holds no c4/ambiguous-node finding for that node

  @AC-C4-DIAG-002.3
  Scenario: A node naming no element in scope is an error
    Given a C4 diagram whose Container or Component node label matches no element in scope
    When the user runs specs check on that document
    Then the report holds a c4/undeclared-node finding of severity error at that node's span

  @AC-C4-DIAG-002.4
  Scenario: A node naming several elements in scope is an error
    Given a C4 diagram node whose label matches the entity names of two elements in scope
    When the user runs specs check on that document
    Then the report holds a c4/ambiguous-node finding of severity error at that node's span, naming both elements

  @AC-C4-DIAG-002.5 @negative
  Scenario Outline: External nodes and passive stores are not undeclared
    Given a C4 diagram whose "<node>" names no element in scope
    When the user runs specs check on that document
    Then the report holds no c4/undeclared-node finding for that node

    Examples:
      | node                |
      | Person node         |
      | System_Ext node     |
      | ContainerDb node    |
```

### FEAT-C4-DIAG-003: Relationship Conformance

**Acceptance Criteria:**

```gherkin
@FEAT-C4-DIAG-003
Feature: Relationship Conformance

  As an architect reading a diagram
  I want every arrow to state one dependency, once
  So that what the diagram shows is what the architecture is

  The C4 standard requires an arrow to point one way, in the direction of
  the dependency. The duplicate-edge rule compares statements within one
  diagram only and treats a to b and b to a as different pairs; the
  drawing-conflict rule compares drawings of one relationship across the
  whole corpus, since an edge may be drawn in both of its homes.

  @AC-C4-DIAG-003.1
  Scenario: A bidirectional arrow is an error
    Given a C4 diagram containing a BiRel statement
    When the user runs specs check on that document
    Then the report holds a c4/bidirectional finding of severity error at that statement's span

  @AC-C4-DIAG-003.2
  Scenario: An edge drawn twice in one diagram is an error
    Given a C4 diagram containing two relationship statements from alias a to alias b
    When the user runs specs check on that document
    Then the report holds a c4/duplicate-edge finding of severity error at the second statement's span

  @AC-C4-DIAG-003.3
  Scenario: Drawings of one relationship that disagree are a warning
    Given one relationship drawn in two diagrams with different technology arguments
    When the user runs specs check
    Then the report holds a c4/drawing-conflict finding of severity warning at the drawing that sorts second by path and then by position, naming the other
```

### FEAT-C4-DIAG-004: Reserved Relationship Tags

**Acceptance Criteria:**

```gherkin
@FEAT-C4-DIAG-004
Feature: Reserved Relationship Tags

  As an architect documenting how two elements interact
  I want the tags I write on a relationship checked
  So that a mistyped or contradictory tag is caught where I wrote it

  Mermaid parses a tagged relationship without error, so it will not catch
  a wrong tag; this feature does. Unreserved tags are not checked.

  @AC-C4-DIAG-004.1
  Scenario Outline: Positional and named tags are read alike
    Given the statement <statement>
    When the user runs specs check on that document
    Then the statement's tags are read as async and audited

    Examples:
      | statement                                              |
      | Rel(a, b, "Uses", "HTTP", "", "", "async+audited")     |
      | Rel(a, b, "Uses", "HTTP", $tags="async+audited")       |

  @AC-C4-DIAG-004.2
  Scenario: Contradictory interaction tags are an error
    Given the statement Rel(a, b, "Uses", $tags="sync+async")
    When the user runs specs check on that document
    Then the report holds a c4/interaction-conflict finding of severity error at that statement's span

  @AC-C4-DIAG-004.3
  Scenario: A contract tag naming no contract is an error
    Given the statement Rel(a, b, "Uses", $tags="contract:contract-missing") where no contract has that slug
    When the user runs specs check on that document
    Then the report holds a c4/contract-unresolved finding of severity error at that statement's span

  @AC-C4-DIAG-004.4
  Scenario: A contract tag the endpoints do not declare is a warning
    Given a statement tagged with a contract the source does not consume or the target does not provide
    When the user runs specs check on that document
    Then the report holds a c4/contract-mismatch finding of severity warning at that statement's span

  @AC-C4-DIAG-004.5
  Scenario: Tags joined with a comma are a warning
    Given the statement Rel(a, b, "Uses", $tags="async,audited")
    When the user runs specs check on that document
    Then the report holds a c4/tag-separator finding of severity warning at that statement's span
```

## References

- [[adr-029]] — relationships are read from blueprint diagrams, by a
  parser this CLI owns
- [[frd-spec-c4]] — the model built on this feature's parser and node
  resolution
- [[c4-standards]] — the diagram conventions these rules enforce
- [[frd-spec-structure]] — record blocks, which declare un-promoted
  elements
- [[frd-spec-index]] — the check engine, rule identifiers and the
  unverified severity of AC-SPEC-006.3
