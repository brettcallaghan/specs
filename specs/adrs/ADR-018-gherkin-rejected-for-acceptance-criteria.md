---
title: "ADR-018: Reject Gherkin as the Acceptance-Criteria Notation"
description: Measurement shows the Gherkin parser validates its own grammar
  and not criterion well-formedness, so EARS prose stays; the AC index the
  proposal was reaching for is built from the existing notation instead
slug: adr-018
type: adr
status: draft
updated_at: 2026-08-05T00:00:00Z
---

# ADR-018: Reject Gherkin as the Acceptance-Criteria Notation

**Status:** draft  
**Date:** 2026-08-05

## Context

[[adr-014]] deferred Gherkin rather than judging it, on the grounds that
adoption was not needed by the tooling then being specified, and closed
by naming the way to settle it: author real acceptance criteria both
ways and compare, rather than deciding in the abstract. That comparison
has now been run against this repository's own corpus, and it changes
the answer from "later" to "no".

Three adoption levels were on the table: change nothing; embed
` ```gherkin ` blocks in an FRD; or put criteria in `.feature` files
that an FRD references. The motivating goals were a parseable and
validatable criterion format, an official Go parser, metadata
extraction for an AC lookup index agents could query, and optionality
toward future tooling.

Everything below was measured in this environment with
`github.com/cucumber/gherkin/go/v42` (v42.0.1, MIT), not taken from
documentation.

### Measured: what the parser actually validates

This is the finding that decides the ADR. [[adr-014]] framed Gherkin's
upside as "a parser, not a regex, decides whether a criterion is well
formed". That is not what the parser does.

| Input | Result |
|---|---|
| Scenario with **zero steps** (`Given`/`When`/`Then` bare) | **Accepted**, 0 steps |
| Tag `@NOT-AN-AC-ID @typo` | **Accepted** — any non-whitespace text is a valid tag |
| The **same** `@AC-SPEC-001.1` tag on two scenarios | **Accepted**, no duplicate detection |
| Scenario with **no tag at all** | **Accepted**, `err = nil` |
| `Wen` instead of `When` | Rejected: `(4:5): expected: #EOF, #TableRow, #DocStringSeparator, #StepLine, … got '    Wen b happens'` |

The parser enforces Gherkin's grammar. It has no concept of an
acceptance criterion, so every check that matters here — is the tag a
well-formed AC identifier, is it unique in the corpus, is there a tag at
all, does the scenario say anything — is custom Go we write ourselves,
identically, on either side of this decision. An untagged criterion is
*not* detectable by the parser; it is detectable by code we write, which
is equally true of the prose we have.

What Gherkin genuinely buys is narrower than the case made for it: a
reliable block boundary, a tag list, and a line and column for every
scenario. That is real and it is more than a regex gives. It is not
criterion validation.

### Measured: the corpus this would be applied to

Every acceptance criterion in the three FRDs under `specs/features/`
was classified by EARS pattern — 73 criteria:

| Pattern | Count | Share |
|---|---|---|
| Event-driven (`WHEN`) | 43 | 58% |
| **Ubiquitous (bare `SHALL`)** | **19** | **26%** |
| **Ubiquitous negative (`SHALL NOT`)** | **6** | **8%** |
| Conditional (`IF`/`THEN`) | 5 | 6% |
| State-driven (`WHILE`), optional (`WHERE`) | 0 | 0% |

[[adr-014]] identified the ubiquitous and `SHALL NOT` forms as the two
patterns Gherkin expresses worst. **They are 34% of this corpus.** A
ubiquitous criterion has no trigger for a `When` to carry, so it becomes
a `Then`-only scenario — legal Gherkin, but a scenario is meant to be an
*example*, and these are invariants. AC-SPEC-001.5, "The CLI SHALL NOT
read files outside the configured spec root", is not an example of
anything; it is a prohibition, and the substitution demotes it to
ordinary prose inside a `Then` step. Nine criteria (12%) contain
`SHALL NOT`, all of which are greppable today and none of which would be
after adoption.

The pattern Gherkin fits best, event-driven, is a clear majority — but a
notation that is awkward for a third of the corpus is not an improvement
over one that handles all of it.

### Measured: Markdown with Gherkin is not available in Go

The clean form of option 2 would be MDG — the official Markdown dialect,
`.feature.md`, where steps are list items and tags are backticked. It
would remove the wrapper problem entirely. It is
**JavaScript-only**: supported in `@cucumber/gherkin` v19+, with other
languages pending feedback.

So option 2 in Go means fenced blocks and a synthetic header. Measured:

- A bare fragment (`Scenario:` with no `Feature:`) fails to parse —
  three errors, including `unexpected end of file`.
- With a synthetic `Feature: synthetic` line prepended, it parses, and
  every reported line number is shifted by the wrapper's size.

An extractor must therefore inject a header, subtract the offset, and
add the fence's own start line, before any diagnostic can cite the
markdown source. That is tractable — but it is bespoke code we own
forever, and what is being parsed is not quite Gherkin.

### Measured: dependency and runtime cost

Cheap, and not a factor either way. Parsing a four-scenario document
takes **37.5 µs**, three orders of magnitude inside the REQ-SPEC-009
budgets. The parser links in for **+1.2 MB** (a 1.44 MB baseline binary
became 2.65 MB) and adds three modules — `cucumber/gherkin/go/v42`,
`cucumber/messages/go/v34`, `google/uuid`.

One recurring cost is not cheap. Gherkin releases **one version number
across all eleven language implementations**, and Go encodes the major
version in the import path. The release history shows v36.1.0 in
November through v42.0.1 on 2026-08-05 — **six major bumps in nine
months**, most driven by changes in other languages. Each is a
source-level import-path edit that `go get -u` will never surface.

### Checked: syntax highlighting

Partial, and decaying. GitHub does highlight ` ```gherkin ` fences via
Linguist, but the grammar is unmaintained — `Rule:` is mis-coloured in
rendered markdown and PR diffs, reported 2026-01-30 and still
unanswered, with the reporter noting the highlighter appears abandoned.
VS Code has **no built-in** Gherkin support; every reader needs a
third-party extension, and none of them highlight Gherkin inside a
markdown fence as well as they do a `.feature` file. Prose EARS in
markdown renders correctly everywhere today.

### Checked: option 3 contradicts an approved FRD

`.feature` files are not admissible to this framework's corpus as it is
currently defined:

- **AC-SPEC-001.6** (approved) makes only files ending in `.md`
  candidate corpus members.
- **AC-SPEC-001.3** excludes any file with no frontmatter — Gherkin has
  no frontmatter mechanism ([[adr-003]]).
- No frontmatter means no `slug`, so a `.feature` file cannot be the
  target of a `[[slug]]` reference ([[adr-012]]), and cannot carry
  `status` or `updated_at`.

Option 3 therefore requires either a second parallel corpus with its own
discovery, membership, indexing and reference rules, or an amendment to
an approved FRD — before considering that the criteria then live in a
file the FRD's reviewer does not see in the same diff.

### The goal that motivated this does not require it

The strongest item on the original list was a CLI that extracts criteria
and serves them to agents as an index. That needs **an AC identifier, a
span, and a parent requirement** — all three already available. The
identifier format and its extraction regex are specified in
`plugin/skills/spec/references/ac-traceability-standards.md`;
[[frd-spec-structure-conformance]] already extracts headings and section
spans, which is the same walk over the same documents. The index is
buildable now, in the notation we have, and is being blocked by a
notation question it does not depend on.

### Scope

EARS appears in 22 files across `plugin/` and `specs/`. Adopting Gherkin
for acceptance criteria and nothing else splits [[adr-002]] into two
scopes — Gherkin for FRD criteria, EARS still governing standards
documents, contracts, blueprints and every other normative statement —
so both readiness checklists gain a second syntax and an author must
know which surface they are on.

## Decision

We will not adopt Gherkin. Acceptance criteria stay prose EARS with
`AC-<PREFIX>-<nnn>.<n>` identifiers, and this ADR supersedes
[[adr-014]], converting its deferral into a decision on the merits for
the acceptance-criteria use case.

Two obligations follow, so this is a decision rather than a refusal:

1. The AC index SHALL be built over the existing notation — extracting
   each criterion's identifier, parent requirement, span and `(manual)`
   marker into the corpus index, and exposing a query subcommand.
2. Criterion structure SHALL be checked mechanically rather than by
   agent judgment: a criterion bullet without a well-formed identifier,
   a duplicate identifier, and an identifier out of sequence within its
   requirement are conformance findings.

The notation-agnostic extraction constraint from [[adr-014]] carries
forward unchanged.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Gherkin blocks embedded in an FRD** (level 2) | Reliable block boundary, tag list and line/column per criterion<br>`@AC-SPEC-001.1` is a valid tag, so identifiers need no change<br>Tag inheritance Feature→Rule→Scenario works in Go and arrives pre-merged<br>`Rule:` maps onto REQ, giving REQ→AC a grammar<br>`Scenario Outline` + `Examples` would structure the mandatory example<br>Keeps one artifact | Validates Gherkin's grammar, **not** criterion well-formedness — untagged, unstructured, duplicate-tagged and empty scenarios all parse clean<br>34% of this corpus is ubiquitous or `SHALL NOT`, the two forms it fits worst; `SHALL NOT` stops being greppable<br>MDG is JavaScript-only, so this needs synthetic wrappers and offset arithmetic in perpetuity<br>Import path churns roughly every six weeks on other languages' releases<br>Splits [[adr-002]] across 22 files<br>~3× the lines for the same content<br>Invites reading specification text as executable tests | The validation benefit that justified it does not exist; the remaining benefit is position data, which does not pay for a notation change of this blast radius |
| **`.feature` files referenced by an FRD** (level 3) | Real `.feature` files highlight properly in editors<br>Directly consumable by a runner if one is ever adopted<br>No FRD template change | Everything above, plus:<br>Contradicts AC-SPEC-001.6 and AC-SPEC-001.3 of an **approved** FRD — not `.md`, no frontmatter<br>No slug, so unreachable by `[[slug]]` ([[adr-012]]); no `status`, no `updated_at`<br>Either duplicates the criteria against the FRD's prose — the exact drift the CLI exists to eliminate — or removes them from the FRD, so the document can no longer be reviewed in one diff | Highest cost of the three and it breaks the corpus model; the duplication defeats the purpose |
| **Structured AC record blocks** — a fenced block per criterion, schema authored in Strata | Validates the thing actually in question: identifier format, required fields, enumerated EARS pattern, uniqueness<br>Reuses the record-block mechanism of [[adr-008]] and the generation pipeline of [[adr-016]]<br>Zero new dependencies; Go structs generated for free | Bespoke — no ecosystem, no editor support, nobody else's tooling reads it<br>Far more verbose than a sentence<br>Would make an FRD substantially less readable as prose, which is its primary job | Strictly more validating than Gherkin at lower cost, and the closest call here — rejected because an FRD is read by humans and agents as narrative, and this trades that away for validation of fields a linter can check in prose |
| **Gauge's markdown spec format** | Markdown-native — headings and bullets, no fenced blocks or wrappers<br>Renders and highlights everywhere | Looser than Gherkin, so *less* structure to validate, not more<br>No clean embeddable Go parsing library for the spec format<br>Smaller ecosystem | Gives up Gherkin's structure without buying back anything this framework needs |
| **ReqIF** (OMG requirements interchange) | The actual standard for machine-readable requirements interchange<br>Tool support in DOORS, Polarion | XML, tool-oriented, assumes a requirements-management suite<br>Not authorable or reviewable in a git diff | Wrong scale by orders of magnitude for a markdown spec framework |
| **Keep prose EARS, and build the AC index and structural checks over it** | No notation change; [[adr-002]] keeps one scope across 22 files<br>Delivers the actual goal — the AC lookup index — immediately<br>Identifier format, uniqueness and sequence become mechanical findings, which is most of what Gherkin was wanted for<br>Ubiquitous and `SHALL NOT` forms keep their grammatical marking<br>No new dependency, no import-path churn | Extraction stays regex-based over markdown bullets, so reformatting a criterion can silently drop it from the index<br>Criterion *content* quality stays agent judgment<br>Forgoes an off-the-shelf grammar and whatever a future runner would have given | N/A — selected |

## Consequences

**Positive:**
- [[adr-002]] keeps a single scope: one notation for every normative
  statement, across all 22 files that reference it.
- The `SHALL NOT` prohibition keeps its grammatical marking, so every
  prohibition in the corpus stays findable with a plain text search.
- The AC index is unblocked and is built from documents that exist, with
  no migration of 73 criteria first.
- Identifier well-formedness, uniqueness and sequence become
  deterministic findings — the largest share of what Gherkin was wanted
  for, obtained without it.
- No dependency whose import path changes every six weeks on another
  language's release schedule.

**Negative:**
- AC extraction stays regex-based over markdown bullets and inherits
  regex brittleness: reformatting a criterion can remove it from the
  index silently, where a parser would have failed loudly.
- Criteria have no formal grammar, so whether a criterion is *good* —
  atomic, testable, one behaviour — stays agent judgment and stays
  unauditable.
- The mandatory `*Example:*` beneath each criterion stays free prose;
  `Scenario Outline` with an `Examples` table would have given
  data-heavy criteria a real structure, and nothing replaces that.
- Forgoes the option value of an off-the-shelf grammar: if an adopting
  project later runs Cucumber or Reqnroll, its feature files and these
  FRDs will be separate artifacts stating overlapping things.
- The framework carries a hand-rolled AC extractor and its schema
  indefinitely, rather than a maintained upstream parser.

**Revisit when** an adopting project runs a Gherkin runner against a
product and wants its feature files to *be* the criteria — that is the
case where the notation earns its cost, because the scenarios become
executable rather than decorative. Two weaker triggers: MDG reaching the
Go implementation, which removes the wrapper and offset objection; or
measurement showing the regex extractor silently dropping criteria in
practice. The 34% ubiquitous-and-negative share is the number to re-run
first — it is a property of this corpus, and a project whose criteria
are overwhelmingly event-driven faces a materially different trade.

## References

- [[adr-014]] — superseded by this ADR; deferred the same question and
  specified the both-ways comparison this ADR carries out
- [[adr-002]] — EARS as the notation for all normative statements, whose
  scope this ADR declines to split
- [[adr-015]] — chose Go partly for the official in-process Gherkin
  parser; that capability is confirmed here and simply not needed
- [[adr-008]], [[adr-016]] — the record-block and Strata mechanisms
  behind the structured-record alternative
- [[frd-spec-validation-cli]] — AC-SPEC-001.3 and AC-SPEC-001.6, the
  corpus membership rules option 3 contradicts
- [[frd-spec-structure-conformance]] — the heading and span extraction the AC
  index builds on
- `plugin/skills/spec/references/ac-traceability-standards.md` — the
  identifier format and extraction regex this decision keeps
- [cucumber/gherkin](https://github.com/cucumber/gherkin) — parser,
  release history, and the eleven language implementations sharing one
  version number
- [MARKDOWN_WITH_GHERKIN.md](https://github.com/cucumber/gherkin/blob/main/MARKDOWN_WITH_GHERKIN.md)
  — the MDG dialect and its JavaScript-only support
- [github/community#185937](https://github.com/orgs/community/discussions/185937)
  — the unmaintained Linguist Gherkin grammar
