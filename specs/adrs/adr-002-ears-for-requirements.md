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
consistent across authors. Free-form prose requirements often contain
implicit assumptions, ambiguous language, or missing conditions. These
problems lead to misinterpretation during implementation, incomplete test
coverage, and disputes about whether a requirement is met. The project
includes specifications at multiple levels (features, components, and
containers) written by different team members, so it needs a standard
approach that reduces variability and improves clarity.

## Decision

We use EARS (Easy Approach to Requirements Syntax) notation for all
functional and non-functional requirements in specification documents. EARS
provides five syntax patterns: ubiquitous, event-driven, state-driven,
optional feature, and unwanted behavior requirements.

### Scope

This decision governs every normative statement that isn't a feature or an
acceptance criterion. That includes standards documents, contracts,
blueprints, and requirement-level statements wherever they're written.

Features and acceptance criteria are written in Gherkin, as decided by
[[adr-021]] and authored according to [[gherkin-guide]]. That decision
narrows this ADR's scope; it doesn't retire it. The two surfaces differ in
what they hold, so an author can tell which notation applies without
consulting either document:

| Surface | Notation |
|---|---|
| An FRD's features and acceptance criteria | Gherkin |
| Every other normative statement, in any document type | EARS |

The split follows a real boundary. Acceptance criteria are scenario-shaped
and externally consumed, and they eventually leave the repository for product
and QA tooling. Standards documents, contracts, and blueprints are
invariant-shaped and internally consumed, and they stay in the repository.
Their modality also varies, whereas an acceptance criterion's modality
doesn't.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Free-form prose requirements | No learning curve<br>Flexibility of natural language<br>Familiar to all team members | High ambiguity<br>Inconsistent structure<br>Hard to validate completeness<br>Difficult to trace to tests | Creates the problem we're trying to solve;<br>lacks the precision that implementation needs |
| RFC 2119 keywords (MUST/SHOULD/MAY) | Industry standard<br>Clear priority levels<br>Minimal syntax | Doesn't structure conditions or triggers<br>Still allows ambiguous prose<br>No guidance on requirement patterns | Addresses priority but not structure;<br>requirements can still be ambiguous despite using MUST or SHOULD |
| Gherkin-style Given/When/Then | Familiar to QA and testing teams<br>Maps directly to BDD tests<br>Structured format | Too granular for high-level requirements<br>Verbose for simple cases<br>Focused on scenarios rather than on the system | Rejected here for *system* requirements, which is the scope this ADR keeps. [[adr-021]] later adopted it for acceptance criteria, for the reason that this row already names: it suits them |
| EARS notation | Standard syntax patterns<br>Reduces ambiguity<br>Clear separation of conditions and requirements<br>Proven in industry | Learning curve for new users<br>Can feel rigid for simple cases<br>Requires discipline to maintain | Not applicable (selected) |

## Consequences

**Positive:**

- Reduces ambiguity through standard syntax patterns, such as IF, WHEN, and
  WHERE.
- Clearly separates preconditions, triggers, and system responses.
- Makes requirement completeness easier to validate, because missing
  conditions become obvious.
- Improves traceability between requirements and test cases.
- Uses an approach that's proven in industry, with established training
  materials and references.
- Requires authors to think through conditions and edge cases.

**Negative:**

- Team members who are unfamiliar with EARS notation face a learning curve.
- Can feel too rigid for simple, obvious requirements.
- Requires discipline and review to keep consistent, because it's easy to
  fall back to prose.
- Stakeholders who are used to natural-language requirements might resist it
  at first.
- Simple requirements become more verbose. For example, "The system shall log
  errors" becomes "While the system is operational, if an error occurs, the
  system shall log..."
- Authors must know which surface they're on, because acceptance criteria are
  written in Gherkin and everything else in EARS. Both readiness checklists
  carry two notations.

## References

- [[adr-021]]: Gherkin for features and acceptance criteria, which narrows
  this decision's scope to normative statements that aren't acceptance
  criteria.
- [[guide-ears]]: EARS format reference and examples.
- [[gherkin-guide]]: the authoring guide for the surface that this ADR no
  longer governs.
- EARS notation originated from work by John Terzakis at Intel and
  Rolls-Royce.
- Mavin, A., et al. (2009). "Easy Approach to Requirements Syntax (EARS)."
