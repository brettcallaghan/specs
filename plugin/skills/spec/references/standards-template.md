---
title: <Document Title>
description: <Brief summary of the requirements scope>
slug: <descriptive-slug ending in -standards/-policy/-guidelines/-constraints>
type: standards
status: draft
---

<!--
Template: Standards Document

Sections (complete each, then delete its instruction block)
  - <Major category>
  - <Another major category>
  - References

Purpose
  This template is for capturing collections of requirements that define 
  standards, policies, or constraints across systems, teams, or the 
  enterprise. This is the "Standards" spec type — the name is shorthand 
  for the whole family of cross-cutting standards, policies, invariants, 
  and constraints. Requirements are organized hierarchically and 
  expressed using EARS (Easy Approach to Requirements Syntax).

Reader profile (apply with the tech-writer skill)
- Reader: an engineer or agent checking whether their work complies, often
  from a review comment or a link in a spec.
- Knows: their own work, not the rule.
- Wants: the rule, its scope, and what compliant looks like.
- Budget: a minute per rule.
- Tone: firm and unambiguous, not scolding. Give the reason for any rule that
  isn't obviously necessary, so readers can apply it to cases it doesn't name.

Slug
  This type has no fixed prefix (it is the registry's one exception). Use 
  a descriptive slug ending in an approved role noun: -standards, -policy, 
  -guidelines, or -constraints (e.g. security-standards, 
  data-retention-policy).

  The slug is independent of the filename. A reference resolves against 
  this document's frontmatter slug, never its path (see 
  cross-document-references.md), so a file may be named for how humans 
  will look for it while the slug carries the registry-conforming 
  identifier.

When to use
  - Defining enterprise-wide standards (coding standards, security 
    policies, operational requirements)
  - Capturing cross-cutting requirements that span multiple features or 
    systems
  - Documenting regulatory or compliance requirements
  - Establishing architectural constraints or design principles
  - Capturing system-wide invariants (privacy, storage, retention)
  - Creating reusable requirement sets that multiple features depend on

When NOT to use
  - Feature-specific requirements → use FRD (Feature Requirements 
    Document) with EARS acceptance criteria instead
  - Architectural decisions with trade-offs → use ADR (Architecture 
    Decision Record)
  - Implementation plans → use technical design documents

Relationship to other document types
  - FRDs reference standards documents when features must comply with 
    standards: "This feature SHALL comply with [[security-standards]]"
  - FRDs capture feature-local requirements; standards documents 
    capture standards and invariants that span features or systems
  - ADRs explain WHY an approach was chosen; standards documents 
    define WHAT must be satisfied

Structure
  - Organize requirements hierarchically using markdown headers (#, ##, ###)
  - Group related requirements under common headings
  - Each requirement should be a single, testable statement using EARS
  - Include optional rationale to explain the "why" behind requirements
  - Separate requirements with horizontal rules (---) for visual clarity

Authoring guidance
  - One requirement per statement (never combine multiple SHALLs with "and")
  - Be specific enough to write a failing test
  - Use SHALL for mandatory, SHOULD for recommended, MAY for optional
  - Avoid vague terms: "appropriate," "reasonable," "adequate"
  - Include rationales for non-obvious requirements to justify their 
    necessity
  - Rationales are especially valuable for change requests and PR reviews
-->

# <Document Title>

<!--
Replace with a 2-3 sentence introduction that explains:
- What domain or scope these requirements cover
- Who these requirements apply to (systems, teams, processes)
- Authority or source (if applicable): regulatory, policy, architectural
-->

---

## <Major category>

<!--
Use level-2 headers (##) for major requirement categories.

Examples:
- ## Data Handling
- ## Security Controls
- ## Performance Requirements
- ## Integration Standards
- ## Error Handling
-->

### <Subcategory>

<!--
Use level-3 headers (###) for subcategories within a major category.

Examples:
- ### Authentication
- ### Data Validation
- ### Logging Requirements

After the subcategory header, provide requirements using EARS syntax.
Each requirement should be followed by an optional rationale.

Format:
  [EARS requirement statement]

  **Rationale:** [Explanation of why this requirement exists]

  ---

The rationale is optional but SHOULD be included when:
- The requirement is not obviously necessary
- The requirement addresses a past incident or known risk
- The requirement has significant cost or complexity implications
- Understanding "the why" helps evaluators assess change requests

Good rationales explain:
- What problem this prevents
- What value this provides
- What would happen if violated
- Historical context (if applicable)

After each requirement, use a horizontal rule (---) to separate it from 
the next requirement for visual clarity.
-->

The system SHALL <requirement statement>.

**Rationale:** <Explanation of why this requirement exists, what problem 
it prevents, or what value it provides.>

---

WHEN <triggering event>, the system SHALL <required action>.

**Rationale:** <Explanation.>

---

The system SHALL NOT <prohibited action>.

**Rationale:** <Explanation.>

---

IF <condition>, the system SHALL <conditional requirement>.

---

WHILE <state holds>, the system SHALL <required behavior>.

**Rationale:** <Explanation.>

---

### <Another subcategory>

<!--
Continue with additional subcategories as needed.
-->

The system SHOULD <recommended behavior>.

**Rationale:** <Explanation of why this is recommended but not mandatory.>

---

The system MAY <optional behavior>.

---

## <Another major category>

<!--
Continue with additional major categories as needed.

Remember:
- Remove all HTML comment blocks after reading and completing sections
- Use EARS patterns consistently (WHEN, WHILE, IF, SHALL, SHALL NOT)
- Capitalize modal verbs (SHALL, SHOULD, MAY) per RFC 2119
- Reference related documents using [[document-slug]] notation
- Keep line length under 80 characters (except for URLs or tables)
-->

### <Subcategory>

The system SHALL <requirement>.

---

## References

<!--
List related documents, source documents, regulatory standards, or other 
references that informed these requirements.

Use [[document-slug]] notation for internal documents.
Use full URLs or standard identifiers for external references.

Examples:
- [[security-policy]] — Enterprise security policy
- [[data-retention-standards]] — Data retention requirements
- [GDPR Article 5](https://gdpr-info.eu/art-5-gdpr/) — Principles relating 
  to processing of personal data
- NIST SP 800-53 — Security and Privacy Controls
-->

- <Additional references>
