---
title: "ADR-002: Use EARS Notation for Requirements"
description: Adopt EARS (Easy Approach to Requirements Syntax) for writing clear, testable requirements
slug: adr-002
type: adr
status: active
---

# ADR-002: Use EARS Notation for Requirements

**Status:** active  
**Date:** 2026-07-16

## Context

Requirements in specification documents need to be unambiguous, testable, and
consistent across different authors. Free-form prose requirements often contain
implicit assumptions, ambiguous language, or missing conditions. These defects
lead to misinterpretation during implementation, incomplete test coverage, and
disputes about whether a requirement has been met. The project includes
specifications at multiple levels (features, components, containers) written by
different team members, so it needs a standardized approach that reduces
variability and improves clarity.

## Decision

We will use EARS (Easy Approach to Requirements Syntax) notation for writing all
functional and non-functional requirements in specification documents. EARS
provides five syntax patterns: ubiquitous, event-driven, state-driven, optional
feature, and unwanted behavior requirements.

### Scope

This decision governs **every normative statement that isn't a feature or
an acceptance criterion**: standards documents, contracts, blueprints, and
requirement-level statements wherever they're written.

Features and acceptance criteria are written in Gherkin, as decided by
[[adr-021]] and authored according to [[gherkin-guide]]. That decision
narrows this ADR's scope but doesn't retire it. The two surfaces differ in
what they hold, so an author can tell which notation applies without
consulting either document:

| Surface | Notation |
|---|---|
| An FRD's features and acceptance criteria | Gherkin |
| Every other normative statement, in any document type | EARS |

The split follows a real boundary. Acceptance criteria are scenario-shaped,
externally consumed, and destined to leave the repository for product and QA
tooling. Standards documents, contracts, and blueprints are invariant-shaped,
internally consumed, and stay in the repository. Their modality also varies,
where an acceptance criterion's doesn't.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Free-form prose requirements | No learning curve<br>Natural language flexibility<br>Familiar to all team members | High ambiguity<br>Inconsistent structure<br>Hard to validate completeness<br>Difficult to trace to tests | Creates the problem we're trying to solve;<br>lacks the precision needed for implementation |
| RFC 2119 keywords (MUST/SHOULD/MAY) | Industry standard<br>Clear priority levels<br>Minimal syntax | Doesn't structure conditions or triggers<br>Still allows ambiguous prose<br>No guidance on requirement patterns | Addresses priority but not structure;<br>requirements can still be ambiguous despite using MUST/SHOULD |
| Gherkin-style Given/When/Then | Familiar to QA/testing teams<br>Direct mapping to BDD tests<br>Structured format | Too granular for high-level requirements<br>Verbose for simple cases<br>Scenario-focused rather than system-focused | Rejected here for *system* requirements, which is the scope this ADR keeps. [[adr-021]] later adopted it for acceptance criteria, on the ground this row already names — that it suits them |
| EARS notation | Standardized syntax patterns<br>Reduces ambiguity<br>Clear separation of conditions and requirements<br>Industry-proven approach | Learning curve for new users<br>Can feel rigid for simple cases<br>Requires discipline to maintain | N/A — selected |

## Consequences

**Positive:**
- Standardized syntax patterns (IF, WHEN, WHERE, and others) reduce
  ambiguity.
- Preconditions, triggers, and system responses are clearly separated.
- Requirement completeness is easier to validate, because missing conditions
  become obvious.
- Traceability between requirements and test cases improves.
- The approach is industry-proven, with established training materials and
  references.
- Authors must think through conditions and edge cases.

**Negative:**
- Team members unfamiliar with EARS notation face a learning curve.
- The notation can feel overly rigid for simple, obvious requirements.
- Consistency requires discipline and review, because authors can easily
  return to prose.
- Stakeholders accustomed to natural-language requirements might resist at
  first.
- Simple requirements become more verbose. For example, "The system shall log
  errors" becomes "While the system is operational, if an error occurs, the
  system shall log..."
- Authors must know which surface they're on, because acceptance criteria are
  written in Gherkin and everything else in EARS. Both readiness checklists
  carry two notations.

## References

- [[adr-021]] — Gherkin for features and acceptance criteria, which narrows
  this decision's scope to non-AC normative statements
- [[guide-ears]] — EARS format reference and examples
- [[gherkin-guide]] — the authoring guide for the surface this ADR no longer
  governs
- EARS notation originated from work by John Terzakis at Intel and Rolls-Royce
- Mavin, A., et al. (2009). "Easy Approach to Requirements Syntax (EARS)"
