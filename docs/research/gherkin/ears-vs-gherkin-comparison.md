# EARS vs Gherkin on a real FRD: what the exercise showed

`specs/features/spec-index/spec-index-frd.md` (the original, EARS prose)
rewritten as a Gherkin variant of itself, per [[adr-020]]. Same feature,
same 52 acceptance criteria.

**The variant document was removed** once [[frd-gherkin-adoption]] became
the framework's worked example of Gherkin in an FRD. This record is kept
for its measurements and its findings, which is what the exercise was
for; the counts below describe the variant as it stood.

Overview, Terminology and the numbering note stay markdown — they are
FRD-level, and Gherkin has no node above `Feature`. Everything at
requirement level moved inside the blocks: the nine user stories into
`Feature` descriptions, and the eight `*Rationale:*` paragraphs either
into those descriptions or into a `Rule:` where they stated a real rule.
No `*Rationale:*` paragraph remains.

Every block in the variant was parsed with
`github.com/cucumber/gherkin/go/v42` and run through the extractor from
`docs/research/gherkin/extraction-findings.md`. **All 9 blocks parse and
all 52 AC identifiers extract, with no duplicates and no untagged
scenarios.** The numbers and the friction below are from doing it, not
from predicting it.

## The numbers

| | Original (EARS) | Variant (Gherkin) |
|---|---|---|
| Requirements section | 410 lines | 633 lines (**+54%**) |
| Lines per criterion | 7.9 | 12.2 |
| Criteria | 52 | 52 |
| Examples | 40 prose sentences | 37 rows across 25 `Scenario Outline`s |
| Lines over 80 cols | 0 | 9 (exempt — fenced blocks, AC-SPECSTR-002.5) |
| Machine-extractable fields per criterion | 1 (the sentence) | id, req, typed steps, example rows, tags |
| User stories | 9 markdown paragraphs | 9 `Feature` descriptions |
| Rationale paragraphs | 8 markdown, scope by position | 0 — absorbed into descriptions and one `Rule:` |

## Finding 1 — steps cannot wrap, but the width limit does not apply

Six of nine blocks failed on the first parse run, and every error was a
wrapped step:

```
(16:7): expected: #EOF, #TableRow, #DocStringSeparator, #StepLine,
        #TagLine, #ExamplesLine, #ScenarioLine, #RuleLine, #Comment,
        #Empty, got '      properties and its file path'
```

**A Gherkin step is one line. There is no continuation syntax.** That
part is real and immutable.

**The rest of what this section originally claimed was wrong.** It said
steps must fit in roughly 70 characters because `specs.toml` sets
`line_width = 80`, and that compressing them to fit was therefore an
unavoidable hazard. Neither holds:

- **Gherkin imposes no length limit at all.** A 412-character step parses
  without complaint.
- **This framework already exempts fenced code blocks from line width.**
  AC-SPECSTR-002.5 of [[frd-spec-structure-conformance]] states it
  outright — no width finding is reported for a line inside a fenced
  code block, a table row, or a line containing a URL — and
  `schemas/config.strata` documents the same exemption on the
  `line_width` property itself. A ` ```gherkin ` block is a fenced code
  block.

So the correct fix for a long criterion is to **write the long line**.
All ten steps that failed parse unchanged at full fidelity as single
lines. Their widths: nine exceed 80 columns, five exceed 100, one
reaches 150.

This matters because the wrong diagnosis produced real damage. Working
from a limit that did not apply, the first pass compressed *"a machine
with no language runtime, interpreter, or package manager installed
**beyond the CLI itself**"* down to *"a machine with no runtime,
interpreter or package manager"* — which is false, since the CLI is
installed. All ten steps have since been restored to the original
wording.

**What remains as a genuine cost** is much smaller than first reported:
a criterion that takes 150 columns to state occupies one long line
inside the block, which reads worse in a narrow diff than a wrapped
prose bullet does. That is a readability trade, not a fidelity hazard.

## Finding 2 — tag `Examples`, not the scenario, or criteria disappear

`REQ-SPEC-009` has three latency budgets — AC-SPEC-009.1, .2 and .3 —
which are obviously one `Scenario Outline` over three rows. Written the
natural way, with the tag on the scenario:

```gherkin
@AC-SPEC-009.1
Scenario Outline: Latency budgets
  ...
  Examples:
    | operation | budget |
    | query     | 50ms   |
    | check     | 100ms  |
    | build     | 1s     |
```

**Two acceptance criteria vanish.** The extractor reported 50 ids where
the original defines 52, and nothing else flagged it — the document
parses, every scenario is tagged, no check fires. It is a silent
requirements-coverage loss.

`Examples` blocks accept their own tags, and that is the fix:

```gherkin
Scenario Outline: Latency budgets against a 1,000-document corpus
  Given a corpus of 1,000 documents
  When the CLI performs "<operation>"
  Then the CLI completes within "<budget>" at the 95th percentile

  @AC-SPEC-009.1
  Examples: an agent asking one question
    | operation      | index state | budget          |
    | a corpus query | warm        | 50 milliseconds |

  @AC-SPEC-009.2
  Examples: a write-time hook checking one file
    | operation               | index state | budget           |
    | a single-document check | warm        | 100 milliseconds |
```

Verified: each `Examples` block carries its own tag in the AST, and
`Pickles()` emits one row per block with the right identifier inherited.
This restored the count to 52.

**Implication:** the AC-tag conformance check must look at
`Scenario.Tags` **and** `Examples.Tags`, and should probably require
that an Outline's identifiers live on its `Examples` blocks rather than
on the scenario. Worth adding to `agent-plugins-133`.

## Finding 3 — the criterion/example split holds up, and reads well

ADR-020 claimed `Scenario Outline` preserves the split that
`*Example:*` bullets provide. On a real document it does, and the
tabular form is genuinely better where a criterion varies over data.
AC-SPEC-001.6 is the clearest win — three prose clauses about `.md`,
`.json` and `.rst` become:

```gherkin
Examples:
  | file                                  | outcome               |
  | docs/adrs/ADR-004-module-blueprint.md | considered            |
  | docs/specs/specs.json                 | ignored, not excluded |
  | docs/adrs/README.rst                  | ignored, not excluded |
```

**But 25 of 52 criteria needed an Outline for a single example row.** A
three-line `Examples:` block carrying one row, to hold what was a
one-sentence `*Example:*`, is the largest single cause of the +54%. The
tabular form earns its cost when a criterion has two or more instances;
below that it is ceremony.

## Finding 4 — the first attempt misused `Rule:`, and that hid everything else

The variant originally mapped **REQ → `Rule:`**, producing blocks headed
`Rule: Configuration`, `Rule: Document Index`, `Rule: Command Surface`.

**That is wrong.** In Gherkin a `Rule` is a *business rule* — an
assertion that can be true or false, which the scenarios beneath it
illustrate. "There can be only one" is a rule. "Configuration" is a
topic. The mapping was made mechanically, because REQ is a grouping
level and `Rule` is a grouping level, without checking what the keyword
means.

The correct mapping follows this framework's own definition. The FRD
template says a requirement "represents one cohesive capability and must
be independently testable" — that is a **`Feature`**:

| FRD | Gherkin |
|---|---|
| FRD | frontmatter slug, and a `@frd-<slug>` tag on each Feature |
| REQ | `Feature:` |
| User Story | the `Feature` description — the canonical Cucumber idiom |
| AC | `Scenario` / `Scenario Outline` |
| a genuine business rule grouping several ACs | `Rule:` |
| `*Example:*` | an `Examples` row |
| rationale | a design note in the `Feature` description, or a `Rule:` name where it states a rule |

Fixing it resolved a problem the wrong mapping had created. With REQ
mapped to `Rule`, `Rule.Description` was the only description slot
available, rationale was put there, and the **user story had nowhere to
go** — so it stayed markdown prose and never became an extractable
field. With REQ mapped to `Feature`, the user story occupies
`Feature.Description`, which is exactly what that slot is for:

```gherkin
@frd-spec-validation-cli
Feature: Document Index

  As an architecture owner
  I want one machine-readable artifact describing every specification document
  So that CI, agents, and a central catalogue all read the same source of truth
```

And `Rule:` becomes available for its actual purpose. REQ-SPEC-002's
rationale contains a real rule, which now groups the two criteria that
depend on it:

```gherkin
  Rule: A build failure is the checker's job, not the indexer's

    @AC-SPEC-002.4
    Scenario: An unparseable document does not abort the index
    @AC-SPEC-002.5
    Scenario: Building an index containing parse failures exits zero
```

That is a quality gain, not just a correctness fix: it forces rationale
that *is* a rule to be stated as one, instead of buried in a paragraph.

## Finding 5 — one `Feature` per document, so one block per REQ

A Gherkin document holds **exactly one** `Feature`. Two in one document
fail to parse:

```
(5:1): expected: #EOF, #TableRow, … #RuleLine, #Comment, #Empty,
       got 'Feature: B'
```

With REQ → `Feature`, one fenced block per requirement is therefore not
a stylistic choice but the only legal arrangement — which happens to be
what a reader wants anyway, since it keeps each requirement's criteria
beside its user story.

Two consequences, both improvements on the first attempt:

- **No synthetic wrapper.** Each block now declares its own `Feature`,
  so the extractor's wrap offset is 0 and the conditional-wrapper
  gotcha does not arise for FRDs at all.
- **The `@frd-<slug>` tag has a home** on each `Feature`, and pickles
  inherit it. An earlier version of this document concluded there was
  nowhere to put it; that was a consequence of the wrong mapping.

For `.feature` emission (ADR-020 constraint 5), one FRD yields one file
per requirement rather than one file per FRD. That is normal for feature
files and needs no special handling.

## Finding 6 — what got harder to read

Honest notes from writing it:

- **Ubiquitous criteria** (26% of the corpus) become `Then`-only
  scenarios. `Scenario: Files outside the spec root are never read` with
  a single `Then` is fine; `AC-SPEC-006.1`, a four-clause invariant
  about what every finding carries, is a `When`/`Then`/`And`/`And` pile
  that reads worse than the sentence did.
- **Cross-references inside steps** are awkward. AC-SPEC-011.3 originally
  said "answering the query of AC-SPEC-007.2". As a step it is
  `Then the CLI answers the query of AC-SPEC-007.2` — a normative
  statement that points elsewhere, which Gherkin has no way to mark as a
  reference rather than prose.
- **`[[slug]]` inside a step** is just text. AC-SPEC-001.2 mentions a
  `[[slug]]`; the reference-integrity feature will need to decide
  whether wikilinks inside Gherkin blocks are resolved or ignored. The
  same question applies to `[[slug]]` inside a `Feature` description,
  which the variant now uses.

## Finding 7 — the `Feature` description absorbs more than the user story

The description is the only free-text slot attached to a requirement, and
it takes more than the three story lines. Verified against the parser:

| | Supported? |
|---|---|
| Multiple paragraphs, blank-line separated | **Yes**, preserved verbatim |
| Markdown emphasis, `code`, `[[slug]]`, bullets, blockquotes | **Yes**, as literal text |
| Markdown headings (`#`) | **No** — silently becomes a Gherkin comment and vanishes |
| A line starting with a Gherkin keyword | **No** — terminates the description and becomes structure |
| Lines wrapping freely | **Yes** — unlike steps, descriptions wrap normally |

So a description can carry the story, a design note, and a scope fence.
That is what the nine descriptions in the variant now do, and it is why
no `*Rationale:*` paragraph survives: the durable half of each one is a
constraint that belongs beside its requirement, and the archaeological
half was dropped.

The two unsupported cases are both silent. Neither errors — a `#` line
disappears and a keyword line changes the document's shape — which is why
they are stated as rules in
`plugin/skills/spec/references/gherkin-guide.md` rather than left to
be discovered.

**Caveat for the structured store:** descriptions live on the AST only.
`Pickles()` rows have no description field, so a loader that consumes
only pickles silently loses every `Feature`, `Rule` and `Scenario`
description.

## What this changes about ADR-020

Nothing in the decision, but three things belong in the migration:

1. **Map REQ to `Feature`, not `Rule`.** A `Rule` is a business rule;
   a requirement is a capability. Reserve `Rule:` for the cases where
   rationale genuinely states a rule that groups several criteria.
2. **Never shorten a step to fit a width limit** — none applies inside
   a fenced block. Write the long line, or decompose only where more
   steps genuinely read better. The first pass got this wrong and
   silently changed the meaning of AC-SPEC-009.4.
3. **AC identifiers on `Scenario Outline` go on `Examples` blocks.**
   Otherwise criteria disappear silently. The conformance check must
   read both tag positions.
4. **Rationale is optional**, and needs a stated convention. The one
   adopted here: if it states a rule governing several criteria it
   becomes a `Rule:`; otherwise it becomes a design note in the
   `Feature` description; if it is neither — history, or sprint-scoping
   — it is deleted.
5. **`Feature` descriptions need authoring guidance**, because two ways
   of breaking them are silent. Drafted in
   `plugin/skills/spec/references/gherkin-guide.md`.

The +54% document growth is real and is the honest price. It buys a
typed record per criterion — identifier, requirement, typed steps,
example rows, tags and span — where prose gives one string, plus a
described requirement where prose gave a story paragraph nothing could
read. That is the trade ADR-020 makes deliberately, now measured on a
real document rather than asserted.

## References

- [[adr-020]] — the decision this exercise tests
- [[frd-gherkin-adoption]] — the live worked example that replaced the
  trial variant
- `specs/features/spec-index/spec-index-frd.md` — the original
- `docs/research/gherkin/extraction-findings.md` — the extraction contract
- `plugin/skills/spec/references/gherkin-guide.md` — how to write
  conforming Gherkin
