---
title: "ADR-002: Use EARS Notation for Requirements"
description: Adopt EARS (Easy Approach to Requirements Syntax) for writing clear, testable requirements
slug: adr-002
type: adr
status: active
---

# ADR-002: Use EARS notation for requirements

**Status:** active  
**Date:** 2026-07-16

## Context

Requirements in specification documents need to be unambiguous, testable, and
consistent across different authors. Free-form prose requirements often contain
implicit assumptions, ambiguous language, or missing conditions. This leads to
misinterpretation during implementation, incomplete test coverage, and disputes
about whether a requirement has been met. Different team members write the
project's specifications, at multiple levels (features, components, and
containers). That calls for a standard approach that reduces variability and
improves clarity.

## Decision

We will use EARS (Easy Approach to Requirements Syntax) notation for writing all
functional and non-functional requirements in specification documents. EARS
provides five syntax patterns: ubiquitous, event-driven, state-driven, optional
feature, and unwanted behavior requirements.

### Scope

This decision governs every normative statement that isn't a feature or an
acceptance criterion: standards documents, contracts, blueprints, and
requirement-level statements wherever they're written.

Features and acceptance criteria are written in Gherkin, decided by
[[adr-021]] and authored per [[gherkin-guide]]. That narrows this ADR's
scope but doesn't retire it. The two surfaces are distinguished by what
they hold, so an author can tell which applies without consulting either
document:

| Surface | Notation |
|---|---|
| An FRD's features and acceptance criteria | Gherkin |
| Every other normative statement, in any document type | EARS |

The split runs along a real boundary. Acceptance criteria are
scenario-shaped, externally consumed, and destined to leave the repository
for product and QA tooling. Standards documents, contracts, and blueprints
are invariant-shaped, internally consumed, and stay in the repository.
Their modality also varies, where an acceptance criterion's doesn't.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Free-form prose requirements | No learning curve<br>Natural language flexibility<br>Familiar to all team members | High ambiguity<br>Inconsistent structure<br>Hard to validate completeness<br>Difficult to trace to tests | Creates the problem we're trying to solve;<br>lacks the precision needed for implementation |
| RFC 2119 keywords (MUST/SHOULD/MAY) | Industry standard<br>Clear priority levels<br>Minimal syntax | Doesn't structure conditions or triggers<br>Still allows ambiguous prose<br>No guidance on requirement patterns | Addresses priority but not structure;<br>requirements can still be ambiguous despite using MUST/SHOULD |
| Gherkin-style Given/When/Then | Familiar to QA/testing teams<br>Direct mapping to BDD tests<br>Structured format | Too granular for high-level requirements<br>Verbose for simple cases<br>Scenario-focused rather than system-focused | Rejected here for system requirements, which is the scope this ADR keeps. [[adr-021]] later adopted it for acceptance criteria, on the ground this row already names: that it suits them |
| EARS notation | Standardized syntax patterns<br>Reduces ambiguity<br>Clear separation of conditions and requirements<br>Industry-proven approach | Learning curve for new users<br>Can feel rigid for simple cases<br>Requires discipline to maintain | N/A (selected) |

## Consequences

**Positive:**
- Reduces ambiguity through standardized syntax patterns (IF/WHEN/WHERE/etc.)
- Clear separation of preconditions, triggers, and system responses
- Easier to validate requirement completeness (missing conditions become
  obvious)
- Improves traceability between requirements and test cases
- Industry-proven approach with established training materials and references
- Forces authors to think through conditions and edge cases

**Negative:**
- Learning curve for team members unfamiliar with EARS notation
- Can feel overly rigid when writing simple, obvious requirements
- Requires discipline and review to maintain consistency (it's easy to fall
  back to prose)
- May create initial resistance from stakeholders accustomed to natural language
  requirements
- Simple requirements become more verbose. For example, "The system shall log
  errors" becomes "While the system is operational, if an error occurs, the
  system shall log..."
- Authors must know which surface they're on, because acceptance criteria are
  written in Gherkin and everything else in EARS. Both readiness checklists
  carry two notations.

## References

- [[adr-021]]: Gherkin for features and acceptance criteria, which narrows
  this decision's scope to non-AC normative statements
- [[guide-ears]]: EARS format reference and examples
- [[gherkin-guide]]: the authoring guide for the surface this ADR no longer
  governs
- EARS notation originated from work by John Terzakis at Intel and Rolls-Royce
- Mavin, A., et al. (2009). "Easy Approach to Requirements Syntax (EARS)"
