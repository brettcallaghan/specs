---
title: "ADR-002: Use EARS Notation for Requirements"
description: Adopt EARS (Easy Approach to Requirements Syntax) for writing clear, testable requirements
slug: adr-002-ears-for-requirements
type: adr
status: active
updated_at: 2026-07-29T00:00:00Z
---

# ADR-002: Use EARS Notation for Requirements

**Status:** active  
**Date:** 2026-07-16

## Context

Requirements in specification documents need to be unambiguous, testable, and
consistent across different authors. Free-form prose requirements often contain
implicit assumptions, ambiguous language, or missing conditions. This leads to
misinterpretation during implementation, incomplete test coverage, and disputes
about whether a requirement has been met. The project includes specifications at
multiple levels (features, components, containers) written by different team
members, creating a need for a standardized approach that reduces variability
and improves clarity.

## Decision

We will use EARS (Easy Approach to Requirements Syntax) notation for writing all
functional and non-functional requirements in specification documents. EARS
provides five syntax patterns: ubiquitous, event-driven, state-driven, optional
feature, and unwanted behavior requirements.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Free-form prose requirements | No learning curve<br>Natural language flexibility<br>Familiar to all team members | High ambiguity<br>Inconsistent structure<br>Hard to validate completeness<br>Difficult to trace to tests | Creates the problem we're trying to solve;<br>lacks the precision needed for implementation |
| RFC 2119 keywords (MUST/SHOULD/MAY) | Industry standard<br>Clear priority levels<br>Minimal syntax | Doesn't structure conditions or triggers<br>Still allows ambiguous prose<br>No guidance on requirement patterns | Addresses priority but not structure;<br>requirements can still be ambiguous despite using MUST/SHOULD |
| Gherkin-style Given/When/Then | Familiar to QA/testing teams<br>Direct mapping to BDD tests<br>Structured format | Too granular for high-level requirements<br>Verbose for simple cases<br>Scenario-focused rather than system-focused | Better suited for acceptance criteria and test cases than system requirements;<br>wrong level of abstraction |
| EARS notation | Standardized syntax patterns<br>Reduces ambiguity<br>Clear separation of conditions and requirements<br>Industry-proven approach | Learning curve for new users<br>Can feel rigid for simple cases<br>Requires discipline to maintain | N/A — selected |

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
- Requires discipline and review to maintain consistency (easy to fall back to
  prose)
- May create initial resistance from stakeholders accustomed to natural language
  requirements
- Simple requirements become more verbose (e.g., "The system shall log errors"
  becomes "While the system is operational, if an error occurs, the system shall
  log...")

## References

- `references/ears-guide.md` — EARS format reference and examples
- EARS notation originated from work by John Terzakis at Intel and Rolls-Royce
- Mavin, A., et al. (2009). "Easy Approach to Requirements Syntax (EARS)"
