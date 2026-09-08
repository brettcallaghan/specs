---
title: "FRD: <Feature Name>"
description: <One-sentence summary of the feature and the user need it serves>
slug: frd-<feature-name>
type: frd
status: draft
updated_at: <YYYY-MM-DDTHH:MM:SSZ>
sources: []                          # slugs of documents this FRD is based on
---

<!--
Template: Feature Requirements Document (FRD)
Version: 1.3
Source: Adapted from the 8090 Requirements Writing Guide
  (https://www.8090.ai/docs/opinions/requirements-writing-guide), modified to
  use tagged Gherkin for features and acceptance criteria per [[adr-021]].

Sections (complete each, then delete its instruction block)
  - Overview
  - Terminology
  - Requirements
  - References

Purpose
  Capture the localized WHY (user story) and WHAT (acceptance criteria) for a
  single feature. An FRD describes what users need and how success is
  verified; it does not describe how the system is built internally.

When to use
  - Specifying a feature that delivers complete, independently testable value
  - Recording user stories and their verifiable acceptance criteria
  - Providing a stable requirements reference for design and implementation
    documents to trace back to

When NOT to use
  - Architectural or technology decisions (use an ADR instead)
  - Implementation design or component internals
  - System-wide invariants or cross-cutting standards
  - Sub-features: do not nest sub-features inside this document; author each
    sub-feature as its own FRD

Instructions
  1. Update the frontmatter with the actual feature details
  2. Choose a short uppercase CODE derived from the feature name
     (e.g. CHK for Checkout) and use it in all FEAT and AC IDs
  3. Complete each section below, following the guidance in comments
  4. Remove all HTML comment blocks after completing the document
  5. Set status using the controlled vocabulary:
     draft | ready-for-review | approved | active | retired
  6. Cross-reference other documents with [[slug]] notation rather than
     duplicating their content — but NOT inside a gherkin block, where a
     wikilink is recorded as a real reference and an illustrative one
     enters the index as a citation the document never made

Splitting and merging features
  - Split into separate FRDs when each feature delivers complete value
    independently, or when different roles own different parts
  - Keep in one FRD when requirements break without each other, complete one
    task together, or are describable in one sentence

ID conventions
  - Requirements: FEAT-<CODE>-NNN (zero-padded sequence, e.g. FEAT-CHK-001),
    written as a tag on the Feature
  - Acceptance criteria: AC-<CODE>-NNN.N (e.g. AC-CHK-001.2), written as a
    tag on the Scenario, deriving its code and number from its Feature's
  - Features share a CODE; a code and a number together identify one
  - Never reuse a retired ID

Verification and traceability
  Automated tests cite the acceptance criteria they verify per
  [[ac-traceability-standards]] — AC identifiers are the durable
  traceability keys, which is why retired IDs are never reused.
  IF an acceptance criterion can only be verified manually, THEN tag it
  @manual; manual criteria are exempt from the test citation obligation.
  A criterion stating a prohibition SHALL carry @negative, which is what
  replaces the grammatical marking SHALL NOT carried in prose.

Slug format
  - Use lowercase kebab-case with the frd- type prefix: frd-<feature-name>
  - Example: frd-checkout

This is a living document. Update when:
  - Requirements change
  - New information becomes available
  - Errors are discovered
  - Context shifts
-->

# <Feature Name>

## Overview

<!--
1-2 narrative paragraphs. Explain the feature's purpose and the user need it
serves. Focus on the problem solved and the value delivered. Do NOT describe
implementation; do NOT list requirements here.
-->

<Describe the feature's purpose, the user need, and the value delivered>

## Terminology

<!--
Define feature-specific ambiguous terms with brief, precise definitions.
Only include terms directly relevant to this feature; skip standard industry
vocabulary. Remove this section if no terms need defining.

Example:
  - Cart: The working collection of items a customer intends to purchase.
  - Order: A finalized purchase record created when checkout completes.
  - Fulfillment Hold: A temporary block on order processing pending manual
    review.
-->

- **<Term>** — <precise definition>.
- **<Term>** — <precise definition>.

## Requirements

<!--
Each requirement represents one cohesive capability and must be independently
testable.

Requirement quality bar:
  - User-centered: describe what users need, not how the system is built
  - Testable: every AC clear enough to write a test from
  - Atomic: one behavior per criterion

Notation: one fenced gherkin block per requirement. Read
[[gherkin-guide]] before writing the first one — it holds the tag
vocabulary, what belongs in each description slot, and the seven
constructs that parse cleanly while destroying meaning.

  | FRD concept                             | Gherkin        |
  |-----------------------------------------|----------------|
  | A requirement                           | Feature:       |
  | Its user story                          | the Feature description |
  | An acceptance criterion                 | Scenario / Scenario Outline |
  | A business rule grouping several ACs    | Rule:          |
  | A criterion's concrete example          | an Examples row |

A gherkin document holds exactly ONE Feature, so one block per
requirement is the only legal arrangement. A requirement is a Feature,
NOT a Rule — Rule: names an assertion that can be true or false ("there
can be only one"), and a topic is a Feature. Getting this backwards
leaves the user story with nowhere to go.

The Feature description carries the user story first, then a design note
stating a constraint no single criterion states, then a scope fence where
one is needed. It must NOT use "#" (Gherkin reads it as a comment and the
line vanishes) and no line in it may begin with a Gherkin keyword.

A Scenario description is reported as a probable untagged criterion, so
put per-criterion rationale in the Feature description, in a Rule: name,
or in markdown outside the block.

Examples
  WHERE a criterion carries concrete example data, prefer a Scenario
  Outline with an Examples table over prose. Below two instances that is
  ceremony — fold the data into Given steps instead, or leave the
  criterion general and put the illustration in markdown beside the block.

  Decide what an Examples block IS before tagging it. One criterion shown
  as several named data groups: tag the SCENARIO. Several criteria sharing
  one shape: tag each EXAMPLES BLOCK. Tag the scenario when you meant the
  second and every criterion but the first ceases to exist, in a document
  that parses cleanly with no check firing.

  Examples are illustrative, not normative: the steps carry the
  requirement. IF an example and the criterion conflict, THEN the
  criterion governs and the example SHALL be corrected.

Verify with the CLI rather than by reading: `specs check <path>` reports
untagged criteria, malformed or duplicate identifiers, empty scenarios,
prose in a Scenario description, and partially tagged Examples blocks.
-->

### FEAT-<CODE>-001: <Requirement Name>

**Acceptance Criteria:**

```gherkin
@FEAT-<CODE>-001
Feature: <Requirement Name>

  As a <role>
  I want to <action>
  So that I can <outcome>

  <A design note: the constraint that governs this capability and that no
  single criterion states. Omit where there is nothing to say.>

  @AC-<CODE>-001.1
  Scenario: <what this criterion establishes>
    When <event>
    Then <observable outcome>

  @AC-<CODE>-001.2
  Scenario: <what this criterion establishes>
    Given <state>
    Then <observable outcome>

  @AC-<CODE>-001.3
  Scenario: <what this criterion establishes>
    Given <condition>
    When <event>
    Then <observable outcome>
```

### FEAT-<CODE>-002: <Requirement Name>

**Acceptance Criteria:**

```gherkin
@FEAT-<CODE>-002
Feature: <Requirement Name>

  As a <role>
  I want to <action>
  So that I can <outcome>

  @AC-<CODE>-002.1
  Scenario Outline: <the criterion, stated once over its data>
    When the system receives "<input>"
    Then the system responds with "<result>"

    Examples:
      | input   | result   |
      | <value> | <value>  |
      | <value> | <value>  |

  @AC-<CODE>-002.2 @negative
  Scenario: <what this criterion forbids>
    Then <the system does not do the forbidden thing>
```

<!--
Add more requirement blocks as needed. Never reuse a retired ID.

Anti-example (do not do this):
  @FEAT-CHK-003
  Feature: Payment
    As a user
    I want payment to work
    So that I can pay

    Scenario: A message is shown
      Then the system shows a message
Why it fails: generic role, no outcome, the criterion is untestable and
carries no identifier tag, and "Payment" is a topic with no user need in it.

Good example:
  @FEAT-CHK-003
  Feature: Payment Confirmation

    As a customer
    I want confirmation after my payment is processed
    So that I know my order was placed

    @AC-CHK-003.1
    Scenario: A successful payment shows the order number
      When payment processing succeeds
      Then the confirmation page shows the order number and estimated delivery date

    @AC-CHK-003.2
    Scenario: A declined payment returns the customer to the form
      Given payment processing fails
      Then the customer is returned to the payment form with an error message describing the failure reason

    @AC-CHK-003.3
    Scenario: Navigating away does not lose the transaction
      When the customer navigates away during processing
      Then the system completes the transaction and displays the confirmation on their next visit
-->

## References

<!--
Optional section. Include references to:
- Source or related documents using [[slug]] notation
- External documentation or standards

Remove this section if no references are needed.
-->

- <Reference 1>
- <Reference 2>
