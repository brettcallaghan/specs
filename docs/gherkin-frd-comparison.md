# EARS vs Gherkin on a real FRD: what the exercise showed

`specs/features/spec-index/spec-index-frd.md` (the original, EARS prose)
rewritten as `docs/spec-index-frd-gherkin.md` (Gherkin blocks, per
[[adr-018]]). Same feature, same 52 acceptance criteria, same overview,
terminology and user stories — only the criteria notation differs.

Every block in the variant was parsed with
`github.com/cucumber/gherkin/go/v42` and run through the extractor from
`docs/plans/gherkin-extraction-findings.md`. **All 9 blocks parse and
all 52 AC identifiers extract, with no duplicates and no untagged
scenarios.** The numbers and the friction below are from doing it, not
from predicting it.

## The numbers

| | Original (EARS) | Variant (Gherkin) |
|---|---|---|
| Requirements section | 410 lines | 613 lines (**+50%**) |
| Lines per criterion | 7.9 | 11.8 |
| Criteria | 52 | 52 |
| Examples | 40 prose sentences | 37 rows across 25 `Scenario Outline`s |
| Lines over 80 cols | 0 | 9 (exempt — fenced blocks, AC-SPECSTR-002.5) |
| Machine-extractable fields per criterion | 1 (the sentence) | id, req, typed steps, example rows, rationale, tags |

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

ADR-018 claimed `Scenario Outline` preserves the split that
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
one-sentence `*Example:*`, is most of the +50%. The tabular form earns
its cost when a criterion has two or more instances; below that it is
ceremony.

## Finding 4 — rationale survives, but moves and loses granularity

The original scatters `*Rationale:*` paragraphs at the end of the
requirements they explain, and their scope is inferred from position.
In Gherkin they become `Rule.Description`, which sits **above** the
criteria and is explicitly scoped to the whole rule.

That is better for extraction and slightly worse to read: the rationale
for AC-SPEC-003.6 specifically now sits at the top of the whole
`Metadata Conformance` rule, ahead of the five criteria it does not
describe. Per-criterion rationale has a home (`Scenario.Description`)
but no obvious style — a paragraph between the scenario name and its
first `Given` reads oddly.

`REQ-SPEC-008` has no rationale at all in the original, which is fine
and should stay fine. An early version of my conformance check flagged
the empty `Rule.Description` as a problem; **rationale must be optional
in the check**, or every terse requirement gets a spurious finding.

## Finding 5 — one block per REQ works, and needs the wrapper

Each requirement carries one fenced block containing a bare `Rule:`.
That does **not** parse alone — it needs the synthetic `Feature:` header
— but with it, everything works, and the structure keeps each
requirement's criteria beside its user story where a reviewer expects
them.

Consequence: there is no `Feature` line, so there is no place for the
`@frd-spec-validation-cli` tag ADR-018 shows. **The FRD slug should come
from frontmatter, not from a Feature tag** — which is better anyway,
since it cannot drift from the document's own identity.

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
  whether wikilinks inside Gherkin blocks are resolved or ignored.

## What this changes about ADR-018

Nothing in the decision, but three things belong in the migration:

1. **Never shorten a step to fit a width limit** — none applies inside
   a fenced block. Write the long line, or decompose only where more
   steps genuinely read better. The first pass got this wrong and
   silently changed the meaning of AC-SPEC-009.4.
2. **AC identifiers on `Scenario Outline` go on `Examples` blocks.**
   Otherwise criteria disappear silently. The conformance check must
   read both tag positions.
3. **Rationale is optional**, and per-criterion rationale needs a stated
   convention or it will be written inconsistently.

The +50% document growth is real and is the honest price. It buys six
extractable fields per criterion where prose gives one — which is the
trade ADR-018 makes deliberately, now measured on a real document rather
than asserted.

## References

- [[adr-018]] — the decision this exercise tests
- `docs/spec-index-frd-gherkin.md` — the variant
- `specs/features/spec-index/spec-index-frd.md` — the original
- `docs/plans/gherkin-extraction-findings.md` — the extraction contract
