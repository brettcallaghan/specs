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
consistent across authors. Free-form prose requirements often contain implicit
assumptions, ambiguous language, or missing conditions. The result is
misinterpretation during implementation, incomplete test coverage, and disputes
about whether a requirement is met. Different team members write the project's
specifications at several levels (features, components, containers). The
project needs a standard approach that reduces that variability and makes
requirements clearer.

## Decision

We will use EARS (Easy Approach to Requirements Syntax) notation for writing all
functional and non-functional requirements in specification documents. EARS
provides five syntax patterns: ubiquitous, event-driven, state-driven, optional
feature, and unwanted behavior requirements.

### Scope

This decision governs every normative statement that isn't a feature or an
acceptance criterion: Standards documents, Contracts, blueprints, and
requirement-level statements wherever they're written.

Features and acceptance criteria are written in Gherkin, as [[adr-021]]
decides and [[gherkin-guide]] describes. That narrows this ADR's scope but
doesn't retire it. The two surfaces differ in what they hold, so an author can
tell which notation applies without consulting either document:

| Surface | Notation |
|---|---|
| An FRD's features and acceptance criteria | Gherkin |
| Every other normative statement, in any document type | EARS |

The split runs along a real boundary. Acceptance criteria are
scenario-shaped and externally consumed, and they leave the repository for
product and QA tooling. Standards documents, Contracts, and blueprints are
invariant-shaped and internally consumed, and they stay. Their modality also
varies, where an acceptance criterion's doesn't.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Free-form prose requirements | No learning curve<br>Flexibility of natural language<br>Familiar to all team members | High ambiguity<br>Inconsistent structure<br>Hard to validate for completeness<br>Hard to trace to tests | Creates the problem this decision solves<br>Lacks the precision that implementation needs |
| RFC 2119 keywords (MUST, SHOULD, MAY) | Industry standard<br>Clear priority levels<br>Minimal syntax | Doesn't structure conditions or triggers<br>Still allows ambiguous prose<br>No guidance on requirement patterns | Addresses priority but not structure<br>A requirement that uses MUST or SHOULD can still be ambiguous |
| Gherkin-style Given/When/Then | Familiar to QA and testing teams<br>Maps directly to BDD tests<br>Structured format | Too granular for high-level requirements<br>Verbose for simple cases<br>Focused on scenarios rather than the system | Rejected for system requirements, which are the scope this ADR keeps. [[adr-021]] adopts it for acceptance criteria, on the ground this row already names: it suits them |
| EARS notation | Standard syntax patterns<br>Reduces ambiguity<br>Separates conditions from requirements<br>Proven in industry | Learning curve for new users<br>Can feel rigid for simple cases<br>Takes discipline to keep consistent | N/A — selected |

## Consequences

**Positive:**
- Standard syntax patterns, with keywords such as IF, WHEN, and WHERE, reduce
  ambiguity.
- Preconditions, triggers, and system responses are clearly separated.
- Requirements are easier to check for completeness, because missing
  conditions become obvious.
- Requirements trace to test cases more easily.
- The approach is proven in industry and has established training materials
  and references.
- EARS makes authors think through conditions and edge cases.

**Negative:**
- Team members who don't know EARS notation face a learning curve.
- EARS can feel too rigid for simple, obvious requirements.
- Keeping requirements consistent takes discipline and review, because it's
  easy to fall back to prose.
- Stakeholders used to natural-language requirements may resist at first.
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
- [[gherkin-guide]] — the authoring guide for the surface this ADR doesn't
  govern
- EARS notation originated from work by John Terzakis at Intel and Rolls-Royce
- Mavin, A., et al. (2009). "Easy Approach to Requirements Syntax (EARS)"
