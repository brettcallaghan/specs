---
title: "ADR-018: Adopt Gherkin Blocks in FRDs for Acceptance Criteria"
description: Acceptance criteria become tagged Gherkin blocks inside the FRD
  because the AST yields typed records a central store can load, while EARS
  prose yields one string; EARS keeps every other normative statement
slug: adr-018
type: adr
status: draft
updated_at: 2026-08-06T00:00:00Z
---

# ADR-018: Adopt Gherkin Blocks in FRDs for Acceptance Criteria

**Status:** draft  
**Date:** 2026-08-06

## Context

A prior draft ADR at number 014 deferred this question and named the way
to settle it: author real criteria both ways and compare. That comparison
has been run, twice, against this repository's own corpus. This ADR
replaces that draft, which is deleted rather than retained as a retired
record — it was never ratified, and the decision log needs one document
per question, not two. ADR number 014 is left unused.

The first pass evaluated the notation against a **local markdown
destination** — is an FRD easier to read, review and lint — and came out
against adoption on margin.

That was the wrong destination. The stated direction is publication out
of these repositories into a **central structured store for C4 models,
features and acceptance criteria**, a knowledge graph fed the same
content, or both; and beyond that, acceptance criteria living in
product/QA tooling that product managers, engineers and QA share rather
than in a source repository at all. [[frd-spec-validation-cli]] already
carries this intent — "an organization pushes the index to a central
service and sees its whole architecture."

Against a typed-record destination the criteria invert. What matters is
no longer how a sentence reads in a diff but **how much structure
survives extraction**, and prose loses that comparison decisively.

Everything below was measured with
`github.com/cucumber/gherkin/go/v42` (v42.0.1, MIT).

### Measured: Gherkin yields typed fields; prose yields one string

An acceptance criterion in a structured store wants columns:
precondition, trigger, expected outcome, polarity, verification level,
example data, provenance. Prose EARS delivers **one string** —
`"WHEN the CLI is invoked against a directory, the CLI SHALL treat…"` —
and every column after that requires writing and maintaining an EARS
parser over natural language.

The Gherkin AST delivers those columns directly. Steps carry a typed
`KeywordType`, not just text:

```
scenario "Slug pattern mismatch is reported with actual and expected"
  keyword="Given"  keywordType=Context      text="a document of type <type>…"
  keyword="And"    keywordType=Conjunction  text="the profile registers…"
  keyword="When"   keywordType=Action       text="the CLI checks the document"
  keyword="Then"   keywordType=Outcome      text="a finding names the actual…"
```

`Context` / `Action` / `Outcome` is precondition / trigger / expected
result as an enum, produced by a maintained parser rather than by our
regexes over English.

### Measured: `Scenario Outline` preserves the criterion/example split

An earlier draft of this ADR found that a Gherkin scenario collapses the
distinction the FRD template rests on — the EARS sentence carries the
requirement, the `*Example:*` beneath it merely illustrates, and the
criterion governs when they conflict. **That finding was wrong.** It was
made against plain `Scenario`, and the corpus's 61 examples are exactly
the case `Scenario Outline` exists for:

```
general (parameterised) form:
  Given a document of type <type> declaring slug "<slug>"
  When the CLI checks the document
  Then a finding names the actual slug "<slug>"

Examples header: [type slug pattern]
  row 1 -> {"type":"adr","slug":"adr-011-…","pattern":"^adr-[0-9]{3}$"}
  row 2 -> {"type":"frd","slug":"checkout","pattern":"^frd-[a-z-]+$"}
```

The parameterised steps are the governing general statement; the rows
are the illustrative instances. The split not only survives — it becomes
**queryable**, where today an example is an unparseable sentence.

### Measured: the compiled form is the fact table

`gherkin.Pickles()` compiles the AST into one fully-resolved row per
example, with tags already inherited from Feature and Rule:

```
pickle "Slug pattern mismatch…"  tags=[@frd-spec-validation-cli @AC-SPEC-003.5
                                       @level-unit]  steps=5
pickle "Slug pattern mismatch…"  tags=[…]  steps=5   <- 2nd Examples row
pickle "Files outside the spec root are never read"
                                 tags=[@frd-spec-validation-cli @AC-SPEC-001.5
                                       @negative @level-unit]  steps=1
```

Three rows from two scenarios. That is the denormalised load target a
structured store wants, produced by the library. The whole document also
serialises to JSON out of the box (`{feature, comments}`), and
`cucumber/messages` is a **published, versioned interchange schema** —
adopting it beats inventing one, and it is what external test-management
tools already ingest.

### Measured: rationale has a home, attached to the node

The concern that Gherkin has nowhere for the corpus's `*Rationale:*`
paragraphs does not hold. `Rule.Description` and `Scenario.Description`
are populated and attached:

```
Rule.Name        = "Metadata Conformance"
Rule.Description = "Frontmatter errors are reported mechanically so review
                    attention goes to whether a document says the right…"
```

Group-level rationale (the corpus has rationale spanning several
criteria) attaches to the `Rule`; per-criterion rationale attaches to the
`Scenario`. Both become text columns. This is arguably better than
today, where a `*Rationale:*` paragraph's scope is inferred from
position.

### The extraction mechanics are settled

Fenced ` ```gherkin ` blocks in `.md` files, wrapped in a synthetic
`Feature:` header only when the block does not declare its own. The
extractor is ~40 lines; positions map back with
`fence_line + (parser_line - wrap)`. Verified against a block containing
one good criterion and three defective ones — correct `.md` line
numbers, and untagged, malformed and duplicate identifiers all reported;
and separately end to end into this project's `Span`, whose byte offsets
were confirmed by slicing the source. **MDG (`.feature.md`) is
JavaScript-only and is not needed.** The full extraction contract and
its five gotchas are recorded in
`docs/plans/gherkin-extraction-findings.md`.

Semantic validation stays in the spec CLI by design: the parser
enforces Gherkin's grammar, and the CLI enforces that every scenario
carries exactly one tag matching
`^@AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`, that identifiers are unique and in
sequence, and that `Feature`/`Scenario` `Description` fields do not
contain what should have been a criterion.

### Tried on a real document

The whole of [[frd-spec-validation-cli]] was rewritten in Gherkin as a
trial — 52 criteria, nine blocks — and parsed with the extractor
(`docs/spec-index-frd-gherkin.md`, findings in
`docs/gherkin-frd-comparison.md`). All nine blocks parse and all 52
identifiers extract, so the design holds. Three costs it surfaced are
recorded in Consequences below rather than discovered during migration:
steps cannot wrap so long criteria become long lines, a `Scenario
Outline` tagged at scenario level silently loses criteria, and the
Requirements section grows 50%.

### What is genuinely lost, and what it costs

Honest accounting, since prose is not strictly worse:

| Lost | Severity |
|---|---|
| **Grammatical negation.** `SHALL NOT` is greppable today; in a `Then` step a prohibition is ordinary prose | Replaced by a required `@negative` tag — machine-*better*, human-weaker. 9 criteria (12%) affected |
| **Modality.** Gherkin cannot mark RFC 2119 levels, and its idiom uses "should" for mandatory behaviour | **Nil for acceptance criteria** — all 73 are `SHALL`/`SHALL NOT`, zero `SHOULD`, zero `MAY`. EARS keeps standards documents, where modality does vary |
| **Invariants read oddly.** 26% of criteria are ubiquitous, becoming `Then`-only scenarios | Real for a human reader; clean as a record (`Outcome` with no `Context`/`Action`) |
| **FRD narrative flow.** The document becomes prose (overview, user stories, rationale) interleaved with fenced blocks | Real and permanent |
| **Free rationale placement.** Rationale must attach to a `Feature`, `Rule` or `Scenario` | A constraint, and a good one |

### The mapping the migration follows

Carried forward from the deleted draft at number 014, which established
it, and extended with the tag conventions this decision adds:

| EARS pattern | Gherkin form |
|---|---|
| `WHEN <event>`, the system SHALL `<action>` | `When <event>` / `Then <action>` |
| `WHILE <state>`, the system SHALL `<action>` | `Given <state>` / `Then <action>` |
| `IF <condition>`, THEN the system SHALL `<action>` | `Given <condition>` / `When …` / `Then <action>` |
| The system SHALL `<action>` (ubiquitous) | `Then <action>` with no trigger — the form that reads least naturally |
| The system SHALL NOT `<action>` | `Then <action>` plus a required `@negative` tag; Gherkin has no native negative form |
| `WHERE <feature>`, the system SHALL `<action>` | a `Rule:` grouping, or a `Given` precondition |
| criterion + its `*Example:*` | `Scenario Outline` with parameterised steps and one `Examples` row per instance |
| `(manual)` marker | `@manual` tag |

### Why this splits [[adr-002]] along a real boundary

The earlier draft counted "splits EARS into two scopes across 22 files"
as a leading cost. Under this framing the split is **principled rather
than arbitrary**: acceptance criteria are scenario-shaped, externally
consumed, and destined to leave the repository for product/QA tooling.
Standards documents, contracts and blueprints are invariant-shaped,
internally consumed, and stay. EARS governs the second set, unchanged
and unsplit within itself.

## Decision

We will adopt **Gherkin blocks embedded in FRDs** (adoption level 2) as
the acceptance-criteria notation. Criteria become ` ```gherkin ` blocks
tagged with their existing `AC-<PREFIX>-<nnn>.<n>` identifiers; the spec
CLI extracts them with the official Go parser and performs semantic
validation itself.

EARS remains the notation for every other normative statement —
standards documents, contracts, blueprints, and requirement-level
statements. [[adr-002]] is amended in scope, not retired.

Binding constraints:

1. Every `Scenario` SHALL carry exactly one tag matching
   `^@AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`; the CLI SHALL report a finding
   otherwise.
2. A criterion expressing a prohibition SHALL carry `@negative`, and one
   verifiable only by hand SHALL carry `@manual`, replacing the
   `(manual)` text marker.
3. `Feature` and `Scenario` `Description` fields SHALL NOT contain
   normative statements; the CLI SHALL report prose there as a probable
   un-tagged criterion.
4. WHERE a criterion has concrete example data, it SHOULD be expressed
   as a `Scenario Outline` with an `Examples` table rather than as prose.
5. The CLI SHALL be able to emit conforming `.feature` files from the
   extracted blocks, so publication to external tooling needs no second
   authoring format.
6. A step SHALL NOT be shortened to fit a line-width limit. No such
   limit applies inside a fenced block (AC-SPECSTR-002.5); the line is
   written at whatever width the criterion needs, and decomposed into
   further steps only where that genuinely reads better.
7. WHERE a `Scenario Outline` carries more than one acceptance
   criterion, each identifier SHALL be tagged on its own `Examples`
   block rather than on the scenario.
8. Extraction SHALL accept both notations while any FRD still carries
   prose criteria, which the notation-agnostic constraint carried
   forward from ADR-014 already requires.

Constraint 7 exists because the trial hit it: tagging one
`Scenario Outline` at scenario level made two criteria disappear from a
document that parsed cleanly, with every scenario tagged and no check
firing. Constraint 6 exists because the trial's first pass invented a
width limit that does not apply and compressed steps to fit it, changing
the meaning of AC-SPEC-009.4 — the hazard is the invented constraint,
not the notation.

Constraint 5 is why level 2 beats level 3 rather than compromising with
it: emitting `.feature` from fenced blocks is mechanical, so the FRD
stays one reviewable document *and* the external artifact is available
on demand. Level 3 pays the corpus-model cost — `.feature` files are
inadmissible under AC-SPEC-001.6 and AC-SPEC-001.3, carry no
frontmatter, no slug, no `status` — to obtain something level 2 can
generate.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **No adoption — EARS prose, with an AC index built over it** | No migration; [[adr-002]] keeps one scope<br>`SHALL NOT` stays greppable<br>FRD reads as continuous narrative<br>No new dependency or import churn | Yields **one string per criterion**; every typed column needs a hand-written EARS parser over English<br>Examples stay unparseable prose<br>Criterion enumeration stays a bullet-matching heuristic that must tolerate 61 `*Example:*` sub-bullets and 9 `*Rationale:*` paragraphs<br>Nothing downstream ingests it — no interchange schema, no tool support | Adequate for a markdown corpus, inadequate for a typed store; it optimises the destination we are leaving |
| **`.feature` files referenced by an FRD** (level 3) | Directly ingestible by external tooling with no export step<br>Proper editor highlighting | Inadmissible to the corpus: not `.md` (AC-SPEC-001.6), no frontmatter (AC-SPEC-001.3), so no slug, no `[[slug]]`, no `status`<br>Criteria leave the FRD, so it can no longer be reviewed in one diff<br>Needs a second parallel corpus or an approved-FRD amendment | Pays a real structural cost for an artifact constraint 5 generates from level 2 |
| **Structured AC record blocks** — a fenced block per criterion, schema authored in Strata | Total schema control, matching whatever the central store defines<br>Reuses [[adr-008]] record blocks and the [[adr-016]] pipeline<br>Zero new dependencies, generated Go structs | Bespoke: no external tool ingests it, so the product/QA-tooling destination needs a translator anyway<br>No published interchange schema<br>Far more verbose; an FRD becomes barely readable | The strongest rival, and the right answer if the store were the *only* destination — rejected because the stated endgame is shared product/engineering/QA tooling, and interop is the whole value there |
| **Gherkin blocks embedded in an FRD** (level 2) | Typed `Context`/`Action`/`Outcome` steps instead of one string<br>`Scenario Outline` + `Examples` preserves the criterion/example split and makes examples queryable<br>`Pickles()` emits a denormalised fact table with tags inherited<br>`cucumber/messages` is a published, versioned interchange schema external tools already read<br>Rationale attaches to `Rule`/`Scenario` nodes<br>AST enumeration makes an untagged criterion a countable node<br>One artifact; `.feature` emitted on demand | Migrates 73 criteria plus four reference documents<br>Loses grammatical negation to a `@negative` tag convention<br>26% ubiquitous criteria read oddly as `Then`-only scenarios<br>Import path carries the major version and bumped six times in nine months on other languages' releases<br>Linguist's Gherkin grammar is unmaintained; VS Code needs an extension<br>FRD narrative becomes prose interleaved with fenced blocks | N/A — selected |

## Consequences

**Positive:**
- Acceptance criteria extract as typed records — precondition, trigger,
  outcome, polarity, verification level, example rows — without a
  hand-written parser over English.
- Example data becomes queryable structured rows rather than prose.
- The corpus gains a published interchange schema (`cucumber/messages`),
  so the central store and any external product/QA tool consume a format
  that already exists.
- Criterion enumeration is an AST walk, so an untagged criterion is a
  countable node rather than an unremarkable sentence.
- Rationale attaches to a node, giving it explicit rather than
  positional scope.
- `.feature` export makes the eventual move of criteria into shared
  tooling a build step, not a migration.

**Negative:**
- Prohibitions lose grammatical marking; finding them depends on the
  `@negative` tag convention holding, which is now a lint rather than a
  property of the sentence.
- Ubiquitous criteria — 26% of the corpus — are expressed as `Then`-only
  scenarios, which is legal but reads as an example of nothing.
- An FRD is no longer continuous prose; readers alternate between
  narrative and fenced blocks.
- [[adr-002]]'s scope narrows, so authors must know which surface they
  are on, and both readiness checklists carry two syntaxes.
- The Go import path changes with Gherkin's shared major version,
  roughly every six weeks, on releases driven by other languages.
- Gherkin's cultural weight implies executable tests; these scenarios
  are specification text and no runner executes them, so they can drift
  toward imperative detail without a runner to punish it.
- Syntax highlighting is partial: GitHub's Linguist grammar is
  unmaintained and mis-colours `Rule:`; VS Code requires an extension.
- **A step cannot wrap**, so a criterion that takes 150 columns to state
  occupies one 150-column line. Gherkin imposes no length limit and this
  framework already exempts fenced blocks from `line_width`
  (AC-SPECSTR-002.5), so nothing forces shortening — but a long line
  reads worse in a narrow diff than a wrapped prose bullet does.
- **A mis-tagged `Scenario Outline` loses criteria silently.** Tagging
  the scenario rather than each `Examples` block collapses several
  criteria into one, in a document that parses, whose every scenario is
  tagged, and against which no check fires. Constraint 7 and its
  conformance check are the only defence.
- **The Requirements section grows about 50%** — measured 410 to 616
  lines, 7.9 to 11.8 lines per criterion — because 25 of 52 criteria
  need a `Scenario Outline` to hold a single example row.

**Revisit when** the central store's schema is actually designed. If it
diverges materially from `cucumber/messages`, the structured-record
alternative regains its advantage, because at that point we would be
translating out of one fixed shape into another rather than authoring
the target shape directly.

## References

- [[adr-002]] — EARS, whose scope this ADR narrows to non-AC normative
  statements
- [[adr-015]] — chose Go partly for the official in-process Gherkin
  parser; this decision exercises it
- [[adr-008]], [[adr-016]] — the record-block and Strata mechanisms
  behind the structured-record alternative
- [[frd-spec-validation-cli]] — the central-publication intent this ADR
  optimises for, and the corpus membership rules level 3 contradicts
- `plugin/skills/spec/references/frd-template.md` — the criterion/example
  split `Scenario Outline` preserves
- `plugin/skills/spec/references/ac-traceability-standards.md` — the
  identifier format carried forward as the tag vocabulary
- [cucumber/gherkin](https://github.com/cucumber/gherkin) — parser and
  release history
- [cucumber/messages](https://github.com/cucumber/messages) — the
  published interchange schema the extracted records conform to
- `docs/plans/gherkin-extraction-findings.md` — the measured extraction
  contract, the AST type hierarchy, and the five integration gotchas
- `docs/spec-index-frd-gherkin.md`, `docs/gherkin-frd-comparison.md` —
  the trial migration of a real FRD and what it cost
