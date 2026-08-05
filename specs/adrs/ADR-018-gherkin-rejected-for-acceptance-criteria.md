---
title: "ADR-018: Reject Gherkin as the Acceptance-Criteria Notation"
description: An extract-and-parse design over fenced Gherkin blocks works and
  gives a cleaner criterion enumeration than markdown bullets, but the margin
  does not pay for migrating the notation; EARS prose stays and the AC index
  is built over it
slug: adr-018
type: adr
status: draft
updated_at: 2026-08-05T00:00:00Z
---

# ADR-018: Reject Gherkin as the Acceptance-Criteria Notation

**Status:** draft  
**Date:** 2026-08-05

## Context

[[adr-014]] deferred Gherkin rather than judging it, and closed by naming
the way to settle it: author real acceptance criteria both ways and
compare, rather than deciding in the abstract. That comparison has now
been run against this repository's own corpus.

Three adoption levels were on the table: change nothing; embed
` ```gherkin ` blocks in an FRD; or put criteria in `.feature` files that
an FRD references. The motivating goals were a structured, parseable
criterion format, an official Go parser, metadata extraction feeding an
AC lookup index agents can query, and optionality toward future tooling.

**This is a close call, and the record should say so.** The proposed
design works. It was built and run, not reasoned about. What follows
separates what Gherkin genuinely delivers from what it costs, and the
decision turns on a margin, not on a defect.

Everything below was measured in this environment with
`github.com/cucumber/gherkin/go/v42` (v42.0.1, MIT).

### Measured: the extract-and-parse design works

The design under evaluation is: extract ` ```gherkin ` blocks from
markdown, wrap each in a synthetic `Feature:` header, parse with the
official Go library, walk the AST, and let the **spec CLI** — not the
parser — enforce that every scenario carries a tag matching
`^@AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`.

That was implemented. The extractor is roughly forty lines: match the
fence, strip the block's indentation, record the fence's line number,
prepend one wrapper line, and map every reported line back with
`fence_line + (parser_line - 1)`. Against a block containing one good
criterion and three defective ones, it produced:

```
.md line  7 (parser line  3)  "Directory invocation"   ok  @AC-SPEC-001.1
.md line 11 (parser line  7)  "THIS ONE HAS NO TAG"    FINDING: no tag at all
.md line 15 (parser line 11)  "Duplicate identifier"   FINDING: dup of line 7
.md line 19 (parser line 15)  "Malformed tag"          FINDING: no well-formed
                                                       AC tag (has [@not-an-ac])
```

Correct source line numbers, and all three defect classes detected.
**The design is sound and the objections to its mechanics do not hold.**
Specifically:

- **MDG is not needed.** The official Markdown dialect (`.feature.md`)
  is JavaScript-only, but its absence costs one wrapper line and one
  subtraction, not an offset map. This was overweighted in an earlier
  draft of this ADR.
- **The parser not validating criteria is not an objection.** It
  validates Gherkin's grammar; semantic validation of AC identifiers is
  the CLI's job under this design, by intent. What the parser
  contributes is a **clean enumeration of criterion nodes with
  positions**, and that is the thing actually worth having.

### Measured: what that enumeration is worth against prose

This is the real question, and the honest answer is *better, but by
less than it first appears.*

Gherkin's advantage is that the AST defines the denominator: every
`Scenario` is a node, so a criterion with no tag is *present and
countable*, where a prose criterion that was never given an identifier
is just a sentence among sentences. That is a genuine reduction in
heuristics.

But the same silent-loss hole exists on both sides. Free prose written
inside a Gherkin block is **silently absorbed**, not rejected:

| Input | Result |
|---|---|
| Two loose sentences above the first `Scenario:` | `err = nil`; absorbed into `Feature.Description`; scenario count unchanged |
| A loose sentence between `Scenario:` and its first step | `err = nil`; absorbed into `Scenario.Description` |

So a criterion someone forgot to make a scenario vanishes just as
quietly as one someone forgot to make a bullet. It is recoverable — the
CLI can assert those `Description` fields are empty — but that is the
same *shape* of check the prose side needs ("every non-blank line in the
Acceptance Criteria span is a conforming bullet"), written against a
different field.

The prose side's version is genuinely more heuristic, because an AC
section legitimately contains other things: the corpus holds **61
`*Example:*` sub-bullets** and **9 `*Rationale:*` paragraphs** inside
requirement sections, all of which a bullet-conformance check must
tolerate. Gherkin gives non-normative prose a cleaner home (`#`
comments, which are a separate AST channel).

**Net: Gherkin's enumeration is cleaner and needs fewer conventions to
be trusted. It is not categorically safer.** That margin is what the
rest of this decision is weighed against.

### Measured: the corpus this would be applied to

Every acceptance criterion in the three FRDs under `specs/features/` —
73 criteria:

| Pattern | Count | Share |
|---|---|---|
| Event-driven (`WHEN`) | 43 | 58% |
| Ubiquitous (bare `SHALL`) | 19 | 26% |
| Ubiquitous negative (`SHALL NOT`) | 6 | 8% |
| Conditional (`IF`/`THEN`) | 5 | 6% |
| State-driven (`WHILE`), optional (`WHERE`) | 0 | 0% |

[[adr-014]] identified the ubiquitous and `SHALL NOT` forms as the two
Gherkin expresses worst; they are **34% of this corpus**. The ubiquitous
share is the softer half — a `Then`-only scenario is legal and reads
acceptably, though a scenario is meant to be an *example* and these are
invariants. The harder half is the **nine criteria (12%) containing
`SHALL NOT`**: inside a `Then` step a prohibition becomes ordinary
prose, and every prohibition in the corpus stops being findable by plain
text search.

**Checked and found nothing:** modality was expected to be a problem,
because Gherkin has no way to mark RFC 2119 levels and its idiom uses
"should" for mandatory behaviour. All 73 criteria are `SHALL` or
`SHALL NOT` — zero `SHOULD`, zero `MAY`. There is no modality objection
and this ADR does not make one.

### The criterion/example distinction does not survive

`plugin/skills/spec/references/frd-template.md` separates two things
deliberately: the EARS sentence carries the requirement, and the
`*Example:*` beneath it is explicitly illustrative — "IF the example and
the criterion conflict, THEN the criterion governs." The corpus has 61
such examples.

A Gherkin scenario collapses them. Its steps are either general (then
they are the criterion and the example is gone) or concrete (then they
are the example and no general statement governs it). The normative
content migrates into the scenario *name*, which is a title rather than
a sentence with a modal verb. `Scenario Outline` with an `Examples`
table would structure data-heavy cases better than prose does — a real
gain — but it does not restore the governing/illustrative split.

### Costs that do not depend on the above

- **Import path churn.** Gherkin ships one version across eleven
  language implementations and Go encodes the major in the import path.
  v36.1.0 in November through v42.0.1 on 2026-08-05 — six major bumps in
  nine months, mostly driven by other languages. Each is a source-level
  edit `go get -u` will never surface.
- **Splits [[adr-002]].** EARS appears in 22 files across `plugin/` and
  `specs/`. Gherkin for FRD criteria leaves EARS governing standards
  documents, contracts, blueprints and every other normative statement,
  so both readiness checklists gain a second syntax.
- **Migration.** 73 criteria, plus `frd-template.md`, `ears-guide.md`,
  `ac-traceability-standards.md` and the `spec` skill.
- **Highlighting is partial and decaying.** GitHub highlights
  ` ```gherkin ` via Linguist, but that grammar is unmaintained —
  `Rule:` is mis-coloured in rendered markdown and PR diffs, reported
  2026-01-30 and unanswered. VS Code has no built-in Gherkin support.
- **Runtime cost is negligible** and not a factor: 37.5 µs per parse,
  +1.2 MB linked, three added modules.

### Option 3 additionally contradicts an approved FRD

`.feature` files are inadmissible to the corpus as defined:
**AC-SPEC-001.6** admits only files ending in `.md`; **AC-SPEC-001.3**
excludes files with no frontmatter, and Gherkin has no frontmatter
mechanism ([[adr-003]]). No frontmatter means no `slug`, so no
`[[slug]]` reference ([[adr-012]]), no `status`, no `updated_at`. Option
3 needs a second parallel corpus or an amendment to an approved FRD —
before considering that its criteria then live outside the diff its
reviewer reads.

### The motivating goal does not depend on the choice

The AC lookup index needs an identifier, a span, and a parent
requirement. All three are available in either notation. The identifier
format and its extraction regex are already specified in
`plugin/skills/spec/references/ac-traceability-standards.md`, and
[[frd-spec-structure-conformance]] already performs the heading and
section-span walk it builds on. The index is buildable now, and was
being sequenced behind a notation question it does not depend on.

## Decision

We will not adopt Gherkin. Acceptance criteria stay prose EARS with
`AC-<PREFIX>-<nnn>.<n>` identifiers. This ADR supersedes [[adr-014]],
converting its deferral into a decision on the merits.

The decision rests on proportion, not on a defect in the alternative:
the enumeration Gherkin adds is cleaner than markdown-bullet extraction
but not categorically safer, and it does not pay for a notation
migration that splits [[adr-002]] across 22 files, rewrites 73 criteria,
costs the `SHALL NOT` marking on 12% of them, and dissolves the
criterion/example distinction the FRD template rests on.

Two obligations follow, so this is a decision rather than a refusal:

1. The AC index SHALL be built over the existing notation — extracting
   each criterion's identifier, parent requirement, span and `(manual)`
   marker into the corpus index, and exposing a query subcommand.
2. Criterion structure SHALL be checked mechanically rather than by
   agent judgment: a bullet in an Acceptance Criteria span without a
   well-formed identifier, a duplicate identifier, and an identifier out
   of sequence within its requirement are conformance findings. This
   check is what closes most of the enumeration gap above, and its cost
   is a fraction of the migration.

The notation-agnostic extraction constraint from [[adr-014]] carries
forward unchanged.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Gherkin blocks embedded in an FRD** (level 2) | The design works — a ~40-line extractor plus the official parser yields correct source line numbers and catches untagged, malformed and duplicate identifiers<br>The AST defines the denominator, so an untagged criterion is a countable node rather than an unremarkable sentence<br>`@AC-SPEC-001.1` is a valid tag, so identifiers need no change<br>Tag inheritance Feature→Rule→Scenario works in Go, pre-merged in the AST<br>`Rule:` maps onto REQ<br>`Scenario Outline` + `Examples` would structure data-heavy criteria better than prose<br>Keeps one artifact; opens the door to a runner later | Loose prose inside a block is silently absorbed into `Description`, so the enumeration gap is narrowed, not closed<br>34% of this corpus is ubiquitous or `SHALL NOT`; the 12% that are prohibitions stop being greppable<br>Collapses the criterion/example distinction that governs 61 examples<br>Splits [[adr-002]] across 22 files<br>Migrates 73 criteria and four reference documents<br>Import path churns roughly every six weeks on other languages' releases | The closest option, and rejected on margin rather than merit: the enumeration it adds over a bullet-conformance check is real but small, and the rest of the cost is a whole-notation migration |
| **`.feature` files referenced by an FRD** (level 3) | Real `.feature` files highlight properly in editors<br>Directly consumable by a runner<br>No FRD template change | Everything above, plus:<br>Contradicts AC-SPEC-001.6 and AC-SPEC-001.3 of an **approved** FRD<br>No slug, so unreachable by `[[slug]]`; no `status`, no `updated_at`<br>Either duplicates the FRD's criteria — the drift the CLI exists to eliminate — or removes them from the FRD, so it can no longer be reviewed in one diff | Highest cost of the three and it breaks the corpus model |
| **Gherkin for the `*Example:*` only** — criterion stays EARS prose; the example becomes a tagged block | The one place a second notation coexists naturally, because examples are explicitly non-normative — so Gherkin's executable connotation is *correct* there<br>Structures 61 free-prose examples; `Examples` tables for data-heavy cases<br>EARS keeps every normative statement, so [[adr-002]] is not split | Still a mixed notation, with two grammars in one document<br>Validates the illustration, not the criterion<br>Adds the dependency and its import churn for the least load-bearing part of the FRD | The strongest hybrid, and worth revisiting on its own if example quality ever becomes the observed problem — but examples are illustrative, and structuring them is not what motivated this |
| **Structured AC record blocks** — a fenced block per criterion, schema authored in Strata | Validates identifier format, required fields, enumerated EARS pattern and uniqueness directly<br>Reuses [[adr-008]] record blocks and the [[adr-016]] generation pipeline<br>Zero new dependencies; Go structs generated free | Bespoke — no ecosystem, no editor support<br>Far more verbose than a sentence<br>Would make an FRD substantially less readable as prose | More validating than Gherkin at lower cost, but an FRD is read as narrative by humans and agents, and this trades that away |
| **Gauge's markdown spec format** | Markdown-native — no fenced blocks or wrappers<br>Renders and highlights everywhere | Looser than Gherkin, so *less* structure to validate<br>No clean embeddable Go library for the spec format | Gives up Gherkin's structure without buying back anything needed here |
| **ReqIF** (OMG requirements interchange) | The actual standard for machine-readable requirements interchange<br>Tool support in DOORS, Polarion | XML, assumes a requirements-management suite<br>Not authorable or reviewable in a git diff | Wrong scale by orders of magnitude |
| **Keep prose EARS, and build the AC index and structural checks over it** | No notation change; [[adr-002]] keeps one scope across 22 files<br>Delivers the AC index immediately, over documents that exist<br>Identifier well-formedness, uniqueness and sequence become mechanical findings — most of the enumeration benefit at a fraction of the cost<br>`SHALL NOT` keeps its grammatical marking<br>Criterion and example keep their governing/illustrative split<br>No new dependency, no import-path churn | Extraction stays regex-based over markdown bullets, so a criterion not written as a conforming bullet is invisible rather than reported<br>The bullet-conformance check must tolerate 61 `*Example:*` sub-bullets and 9 `*Rationale:*` paragraphs, so it carries conventions a grammar would not<br>Criterion *content* quality stays agent judgment<br>Forgoes an off-the-shelf grammar and a future runner | N/A — selected |

## Consequences

**Positive:**
- [[adr-002]] keeps a single scope across all 22 files that reference
  it, and both readiness checklists keep one syntax to test against.
- `SHALL NOT` keeps its grammatical marking, so every prohibition in the
  corpus stays findable by plain text search.
- The criterion/example distinction survives: 61 examples stay
  explicitly illustrative under a governing normative sentence.
- The AC index is unblocked and needs no migration of 73 criteria first.
- Identifier well-formedness, uniqueness and sequence become
  deterministic findings — most of what Gherkin was wanted for.
- No dependency whose import path changes on another language's release
  schedule.

**Negative:**
- AC extraction stays regex-based over markdown bullets. A criterion not
  written as a conforming bullet is **invisible rather than reported** —
  the one place Gherkin's AST was genuinely better, and this decision
  accepts that gap.
- The bullet-conformance check must tolerate `*Example:*` sub-bullets
  and `*Rationale:*` paragraphs, so it encodes document conventions that
  a grammar would have enforced structurally. Those conventions can
  drift.
- Criteria have no formal grammar, so whether a criterion is *good* —
  atomic, testable, one behaviour — stays agent judgment.
- Data-heavy examples stay free prose; an `Examples` table would have
  structured them, and nothing replaces that.
- Forgoes the option value of an off-the-shelf grammar: an adopting
  project that later runs Cucumber will keep its feature files and these
  FRDs as separate artifacts stating overlapping things.
- The framework carries a hand-rolled AC extractor indefinitely rather
  than a maintained upstream parser.

**Revisit when** an adopting project runs a Gherkin runner against a
product and wants its feature files to *be* the criteria — that is where
the notation earns its cost, because the scenarios become executable
rather than decorative. Two weaker triggers, both of which would narrow
the margin this decision rests on: the bullet-conformance check of
obligation 2 proving insufficient in practice, with criteria observed
reaching `approved` FRDs unreported; or the corpus shifting toward
event-driven criteria, since the 34% ubiquitous-and-negative share is a
property of *this* corpus and is the first number to re-measure.

## References

- [[adr-014]] — superseded by this ADR; deferred the same question and
  specified the both-ways comparison carried out here
- [[adr-002]] — EARS as the notation for all normative statements, whose
  scope this ADR declines to split
- [[adr-015]] — chose Go partly for the official in-process Gherkin
  parser; that capability is confirmed here and simply not exercised
- [[adr-008]], [[adr-016]] — the record-block and Strata mechanisms
  behind the structured-record alternative
- [[frd-spec-validation-cli]] — AC-SPEC-001.3 and AC-SPEC-001.6, the
  corpus membership rules option 3 contradicts
- [[frd-spec-structure-conformance]] — the heading and span extraction
  the AC index builds on
- `plugin/skills/spec/references/frd-template.md` — the criterion/example
  split this decision preserves
- `plugin/skills/spec/references/ac-traceability-standards.md` — the
  identifier format and extraction regex this decision keeps
- [cucumber/gherkin](https://github.com/cucumber/gherkin) — parser,
  release history, and the eleven implementations sharing one version
- [MARKDOWN_WITH_GHERKIN.md](https://github.com/cucumber/gherkin/blob/main/MARKDOWN_WITH_GHERKIN.md)
  — the MDG dialect and its JavaScript-only support
- [github/community#185937](https://github.com/orgs/community/discussions/185937)
  — the unmaintained Linguist Gherkin grammar
