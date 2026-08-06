---
title: "ADR-014: Defer Gherkin as the Acceptance-Criteria Notation"
description: Prose EARS acceptance criteria with AC identifiers stay the
  authoring notation; validation tooling keeps AC extraction
  notation-agnostic so the choice stays reversible
slug: adr-014
type: adr
status: retired
updated_at: 2026-08-05T00:00:00Z
---

# ADR-014: Defer Gherkin as the Acceptance-Criteria Notation

**Status:** retired (superseded by [[adr-018]])  
**Date:** 2026-07-31

## Context

Acceptance criteria in this framework are prose sentences in EARS syntax,
carrying hand-written identifiers of the form `AC-<PREFIX>-<nnn>.<n>`
([[adr-002]], `plugin/skills/spec/references/frd-template.md`).
Tests cite those identifiers in line comments, and
`plugin/skills/spec/references/ac-traceability-standards.md`
specifies a validation tool that matches citations with the regex
`AC:\s*AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+`.

That design has one structural weakness: **nothing validates the criterion
itself.** An identifier is a string the author typed. A regex can prove a
citation resolves to an identifier that exists, but it cannot tell a
well-formed acceptance criterion from a malformed one, and it cannot
detect a criterion that was never given an identifier at all. The
framework compensates with agent review, which is judgment applied to a
problem that is partly structural.

Planned validation tooling for the spec document set (see
`docs/plans/spec-validation-cli-brief.md`) forces the question now,
because whatever that tool does for AC extraction is expensive to change
later. Gherkin is the obvious candidate: a standardized grammar for
scenarios with a tag mechanism that could carry AC identifiers, and
mature parsers that would replace regex matching with real parsing.

### What Gherkin's tagging actually provides

Established by running the official `gherkin-official` parser rather than
from documentation:

| Capability | Result |
|---|---|
| Tag syntax | `@` followed by any non-whitespace text. Whitespace is the only prohibition |
| Our identifier format | `@AC-CHK-003.1` parses as a valid tag — dots, hyphens and uppercase are all accepted, so the existing convention needs no change |
| Multiple tags | Whitespace-separated on one or more lines above the element they annotate |
| Attachment points | `Feature`, `Rule`, `Scenario`, `Scenario Outline`, `Examples` |
| Inheritance | Tags propagate down (Feature → Rule → Scenario) and arrive pre-merged in the AST, so a document-level tag lands on every criterion automatically |
| Position data | Every scenario carries a line and column, so a diagnostic can cite `file:line` |
| Parse target | Parses from a string, not only a file — so fenced ` ```gherkin ` blocks embedded in an FRD are parseable in-process |
| Structural validation | A malformed scenario is rejected with a position: a typo'd step keyword yields `(5:5): expected: #EOF, #TableRow, #DocStringSeparator, #StepLine, #TagLine, #Examples` |
| Fragments | A bare `Scenario:` or `Rule:` does **not** parse. Embedded blocks require a synthetic `Feature:` wrapper, which shifts line numbers and so requires an offset map to report errors against the markdown source |
| Official implementations | .NET, Java, JavaScript, Ruby, Go, Python, C, Perl, PHP, Dart, C++ — **not Rust**, where `cucumber-rs/gherkin` is a third-party reimplementation outside the upstream shared test suite |

Gherkin therefore satisfies the mechanical requirements. Two
non-mechanical problems remain, and they are why this ADR does not adopt
it.

**This is a substitution, not an addition.** EARS and Gherkin express
the same thing in different words. "WHEN payment processing succeeds,
the system SHALL display the confirmation page" and "When payment
processing succeeds / Then the confirmation page is displayed" are the
same criterion in two notations. Adopting Gherkin therefore means
Gherkin *replaces* EARS for features and acceptance criteria — the two
are never used for the same statement. The question this ADR answers is
which notation expresses an acceptance criterion, not whether to add a
second one.

The mapping between them is close enough to make the substitution
mechanical, which is itself an argument that one of them is redundant:

| EARS pattern | Gherkin equivalent |
|---|---|
| WHEN `<event>`, the system SHALL `<action>` | `When <event>` / `Then <action>` |
| WHILE `<state>`, the system SHALL `<action>` | `Given <state>` / `Then <action>` |
| IF `<condition>`, THEN the system SHALL `<action>` | `Given <condition>` / `When …` / `Then <action>` |
| The system SHALL `<action>` (ubiquitous) | `Then <action>` with no trigger — the one case Gherkin expresses less naturally |
| The system SHALL NOT `<action>` | No native negative form; expressed in the `Then` text |
| WHERE `<feature>`, the system SHALL `<action>` | A tag plus a `Rule:`, or a `Given` precondition |

Two EARS patterns have no clean Gherkin counterpart: the ubiquitous
form, which has no trigger for a `When` to carry, and the explicit
`SHALL NOT` prohibition, which becomes ordinary prose inside a `Then`
step and so loses the grammatical marking that makes prohibitions
greppable today.

The cost of the substitution is therefore narrower than "two notations"
but real: [[adr-002]] established EARS for *all* normative statements,
and adopting Gherkin would split that — Gherkin for feature acceptance
criteria, EARS still governing standards documents, contracts, and every
other normative sentence. One notation becomes two scopes, and an author
has to know which surface they are writing on.

**Scope coupling.** Adoption changes four surfaces together — the FRD
template, the AC traceability standards, the EARS guide and the `spec`
skill itself. None of that is required by the validation tooling now
being specified, and folding it in would expand that tool's requirements
document to cover a notation change it does not depend on.

## Decision

We will keep prose EARS acceptance criteria with `AC-<PREFIX>-<nnn>.<n>`
identifiers as the authoring notation, and defer Gherkin adoption.

To keep the deferral cheap to reverse: WHERE validation tooling extracts
acceptance criteria from a document, it SHALL treat the criterion
notation as a replaceable input rather than assuming prose, so that
adding a Gherkin extractor later changes one component instead of the
traceability design.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Adopt Gherkin as the AC notation** — criteria become ` ```gherkin ` blocks inside the FRD, tagged with their identifier | A parser, not a regex, decides whether a criterion is well formed<br>Identifiers become tags the parser extracts, so an untagged criterion is detectable<br>One artifact — no separate feature files to drift<br>The FRD template already requires an EARS statement plus a concrete example; scenario name and steps give that pair a grammar | Splits EARS into two scopes: Gherkin for acceptance criteria, EARS still governing every other normative statement<br>Loses the ubiquitous and SHALL NOT forms, which have no clean Gherkin counterpart<br>Rewrites `plugin/skills/spec/references/frd-template.md`, `plugin/skills/spec/references/ac-traceability-standards.md` and `plugin/skills/spec/references/ears-guide.md` together<br>Embedded blocks need a synthetic `Feature:` wrapper and an offset map for correct diagnostics<br>Invites the assumption that scenarios are executable tests when they are specification text | Not rejected on merit — deferred. Splitting the scope of [[adr-002]] is a prerequisite decision, and adoption is not needed by the tooling currently being specified |
| **Separate `.feature` files carrying `@AC-*` tags** — FRD keeps prose criteria; feature files mirror them | No template change<br>Traceability works through tags<br>Natural if a project already runs Cucumber | Two artifacts stating the same criteria, kept in sync by hand — precisely the drift the validation tooling exists to eliminate<br>Adds an artifact with no owner in the document taxonomy | The duplication defeats the purpose |
| **Gherkin for the example only** — prose EARS criterion stays; the mandatory `*Example:*` beneath it becomes a Gherkin block | Smallest change<br>Structures the example, which is currently free prose | Validates the illustration, not the criterion — the weakness in the Context section is untouched<br>The one place a second notation genuinely WOULD coexist with EARS, since the criterion above it stays prose | Cost of a mixed notation without the benefit |
| **Keep prose EARS criteria with identifiers** | No change to templates, guides, or the skill<br>Keeps the tooling's requirements scoped to validation<br>EARS remains the single normative notation, as [[adr-002]] established<br>Leaves the decision open for a deliberate later call | Identifiers stay hand-written and unvalidated — a criterion with no identifier is invisible to tooling<br>Criterion well-formedness stays a judgment call<br>Regex extraction is brittle relative to parsing | N/A — selected |

## Consequences

**Positive:**
- The validation tooling's requirements stay scoped to validation, and
  its FRD does not have to carry a notation change.
- EARS remains the single notation for normative statements, so
  [[adr-002]] needs no amendment and
  `plugin/skills/spec/references/document-readiness-checklist.md`
  checks 4 and 5 keep one syntax to test against.
- The scope of [[adr-002]] stays whole: one notation governs every
  normative statement, rather than one notation for acceptance criteria
  and another for everything else.
- The reversibility constraint on AC extraction is a small, concrete
  design requirement rather than an open-ended "keep it flexible".

**Negative:**
- Acceptance-criteria identifiers stay hand-written and unvalidated. A
  criterion that never received an identifier cannot be detected by
  tooling — only by a reviewer noticing its absence.
- Criterion well-formedness stays agent judgment, so it stays
  non-deterministic and unauditable in a way a parser would not be.
- AC extraction stays regex-based and inherits regex brittleness:
  reformatting a criterion can silently remove it from traceability.
- The question of whether Gherkin should replace EARS for acceptance
  criteria remains open and will have to be answered before this is
  revisited, not during.
- Notation-agnostic extraction is a design tax paid now for an option
  that may never be exercised.

**Revisit when** any of the following holds: an adopting project runs
Cucumber or another Gherkin runner and wants its feature files to be the
acceptance criteria; untagged or malformed criteria are observed reaching
`approved` FRDs often enough to be worth structural prevention; or the
substitution is judged on its own merits. The recommended way to settle
it is to author one real FRD both ways and compare, rather than deciding
in the abstract — paying particular attention to the ubiquitous and
SHALL NOT criteria, which are where the substitution loses the most.

## References

- [[adr-018]] — supersedes this ADR. It carries out the both-ways
  comparison recommended above and adopts Gherkin at level 2. The
  deciding factor is one this ADR did not weigh: against publication
  into a central structured store, the AST yields typed records where
  prose yields one string, and `Scenario Outline` preserves the
  criterion/example split rather than collapsing it
- [[adr-002]] — established EARS as the requirements notation; this ADR
  declines to displace it
- `plugin/skills/spec/references/ac-traceability-standards.md` —
  the citation format and the extraction regex this decision keeps
- `plugin/skills/spec/references/ears-guide.md` — the patterns
  acceptance criteria are written in
- `docs/plans/spec-validation-cli-brief.md` — the tooling work that
  raised the question, and the parser experiments summarized in Context
- [cucumber/gherkin](https://github.com/cucumber/gherkin) — the official
  parser and its language implementations
- [Gherkin reference](https://cucumber.io/docs/gherkin/reference/) — tag
  syntax and attachment points
- [cucumber-rs/gherkin](https://github.com/cucumber-rs/gherkin) — the
  third-party Rust implementation noted in Context
