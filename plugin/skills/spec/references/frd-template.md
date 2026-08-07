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
Version: 1.2
Source: Adapted from the 8090 Requirements Writing Guide
  (https://www.8090.ai/docs/opinions/requirements-writing-guide), modified to
  use EARS for acceptance criteria per [[template-authoring-standards]].

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
  2. Choose a short uppercase PREFIX derived from the feature name
     (e.g. CHK for Checkout) and use it in all REQ and AC IDs
  3. Complete each section below, following the guidance in comments
  4. Remove all HTML comment blocks after completing the document
  5. Set status using the controlled vocabulary:
     draft | ready-for-review | approved | active | retired
  6. Cross-reference other documents with [[slug]] notation rather than
     duplicating their content

Splitting and merging features
  - Split into separate FRDs when each feature delivers complete value
    independently, or when different roles own different parts
  - Keep in one FRD when requirements break without each other, complete one
    task together, or are describable in one sentence

ID conventions
  - Requirements: REQ-<PREFIX>-NNN (zero-padded sequence, e.g. REQ-CHK-001)
  - Acceptance criteria: AC-<PREFIX>-NNN.N (e.g. AC-CHK-001.2)
  - Never reuse a retired ID

Verification and traceability
  Automated tests cite the acceptance criteria they verify per
  [[ac-traceability-standards]] — AC identifiers are the durable
  traceability keys, which is why retired IDs are never reused.
  IF an acceptance criterion can only be verified manually, THEN append
  the marker "(manual)" to the end of the criterion text; manual
  criteria are exempt from the test citation obligation.

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

Acceptance criteria use EARS (Easy Approach to Requirements Syntax) with
capitalized keywords. Choose the pattern that fits each behavior:

  | Pattern      | Syntax                                          |
  |--------------|-------------------------------------------------|
  | Ubiquitous   | The system SHALL <action>                       |
  | Event-driven | WHEN <event>, the system SHALL <action>         |
  | State-driven | WHILE <state>, the system SHALL <action>        |
  | Conditional  | IF <condition>, THEN the system SHALL <action>  |
  | Negative     | The system SHALL NOT <action>                   |

Modality (RFC 2119): SHALL = mandatory, SHOULD = recommended, MAY = optional.
Keywords are always capitalized.

Examples
  Each acceptance criterion SHOULD be followed by exactly one concrete
  example that instantiates it with specific data (actor, values, formats,
  observed result). Writing the example forces precision — "which fields,
  exactly?" — at spec time, before implementation.

  Examples are illustrative, not normative: the EARS sentence carries the
  requirement; the example clarifies it. IF the example and the criterion
  conflict, THEN the criterion governs and the example SHALL be corrected.
  Keep to ONE canonical example per criterion; exhaustive edge-case
  enumeration belongs in tests, not in this document. Omit the example
  where the criterion is unambiguous without one.
-->

### REQ-<PREFIX>-001: <Requirement Name>

**User Story:** As a <role>, I want to <action>, so that I can <outcome>.

**Acceptance Criteria:**
- **AC-<PREFIX>-001.1:** WHEN <event>, the system SHALL <action>.
  - *Example:* <one concrete instance: specific actor, data, and observed
    result>.
- **AC-<PREFIX>-001.2:** WHILE <state>, the system SHALL <action>.
  - *Example:* <one concrete instance>.
- **AC-<PREFIX>-001.3:** IF <condition>, THEN the system SHALL <action>.
  - *Example:* <one concrete instance>.

### REQ-<PREFIX>-002: <Requirement Name>

**User Story:** As a <role>, I want to <action>, so that I can <outcome>.

**Acceptance Criteria:**
- **AC-<PREFIX>-002.1:** WHEN <event>, the system SHALL <action>.
  - *Example:* <one concrete instance>.
- **AC-<PREFIX>-002.2:** The system SHALL NOT <action>.

<!--
Add more REQ blocks as needed. Never reuse a retired ID.

Anti-example (do not do this):
  User Story: "As a user, I want payment to work."
  AC-CHK-003.1: "The system should show a message."
  AC-CHK-003.2: "Errors should be handled appropriately."
Why it fails: generic role, no outcome, criteria are untestable, and no EARS
trigger or condition is specified.

Good example:
  REQ-CHK-003: Payment Confirmation
  User Story: "As a customer, I want to receive confirmation after my
  payment is processed, so that I can know my order was placed successfully."
  AC-CHK-003.1: "WHEN payment processing succeeds, the system SHALL display
    a confirmation page with the order number and estimated delivery date."
    Example: "Alice completes checkout for 2 items totaling $45.00 with
    standard shipping. Payment succeeds. The confirmation page shows order
    number ORD-20260716-001 and estimated delivery July 23, 2026."
  AC-CHK-003.2: "IF payment processing fails, THEN the system SHALL return
    the user to the payment form with an error message describing the
    failure reason."
    Example: "Alice's card is declined. She is returned to the payment form
    with the message 'Payment declined — please try another card.'"
  AC-CHK-003.3: "WHEN the user navigates away during processing, the system
    SHALL complete the transaction and display the confirmation on their
    next visit."
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
