---
title: "ADR-021: Adopt Gherkin for Features and Acceptance Criteria"
description: Features and acceptance criteria become tagged Gherkin, as
  fenced blocks inside an FRD and as standalone .feature files, because the
  AST yields typed records a central store can load where EARS prose yields
  one string; EARS keeps every other normative statement
slug: adr-021
type: adr
status: approved
---

# ADR-021: Adopt Gherkin for features and acceptance criteria

**Status:** approved  
**Date:** 2026-08-06

## Context

A prior draft ADR at number 014 deferred this question and named the way to
settle it: author real criteria both ways and compare. We ran that comparison
twice against this repository's own corpus. This ADR replaces that draft. The
draft is deleted rather than kept as a retired record, because it was never
ratified, and the decision log needs one document per question, not two. ADR
number 014 is left unused.

The first pass evaluated the notation against a local Markdown destination,
asking whether an FRD is easier to read, review, and lint. On that basis, it
came out narrowly against adoption.

That was the wrong destination. The stated direction is publication out of
these repositories into a central structured store for C4 models, features,
and acceptance criteria, into a knowledge graph fed the same content, or into
both. Beyond that, acceptance criteria are meant to live in product and QA
tooling that product managers, engineers, and QA share, rather than in a
source repository at all. [[frd-spec-index]] already carries this intent: "an
organization pushes the index to a central service and sees its whole
architecture."

Against a typed-record destination, the criteria invert. What matters is no
longer how a sentence reads in a diff, but how much structure survives
extraction, and prose loses that comparison decisively.

Everything in this section was measured with
`github.com/cucumber/gherkin/go/v42` (v42.0.1, MIT).

### Measured: Gherkin yields typed fields; prose yields one string

An acceptance criterion in a structured store needs columns: precondition,
trigger, expected outcome, polarity, verification level, example data, and
provenance. Prose EARS delivers one string, such as
`"WHEN the CLI is invoked against a directory, the CLI SHALL treat…"`, and
every column after that requires writing and maintaining an EARS parser over
natural language.

The Gherkin AST delivers those columns directly. Steps carry a typed
`KeywordType`, not just text:

```
scenario "Slug pattern mismatch is reported with actual and expected"
  keyword="Given"  keywordType=Context      text="a document of type <type>…"
  keyword="And"    keywordType=Conjunction  text="the profile registers…"
  keyword="When"   keywordType=Action       text="the CLI checks the document"
  keyword="Then"   keywordType=Outcome      text="a finding names the actual…"
```

`Context`, `Action`, and `Outcome` express precondition, trigger, and expected
result as an enum. A maintained parser produces them, rather than our regular
expressions over English.

Constraint 13 declines to persist this. The capability is real, and it's what
distinguishes a parsed notation from prose. However, nothing this framework
does operates below the level of a whole criterion. A record that carries
typed steps could be filled from Gherkin alone, which would make it a
Gherkin-shaped model in a framework that treats the notation as replaceable.
The AST keeps the capability available for the day a consumer needs it; the
manifest doesn't spend it in advance.

### Measured: `Scenario Outline` preserves the criterion/example split

An earlier draft of this ADR found that a Gherkin scenario collapses the
distinction that the FRD template rests on: the EARS sentence carries the
requirement, the `*Example:*` beneath it only illustrates, and the criterion
governs when they conflict. That finding was wrong. It was made against plain
`Scenario`, and the corpus's 61 examples are exactly the case that
`Scenario Outline` exists for:

```
general (parameterised) form:
  Given a document of type <type> declaring slug "<slug>"
  When the CLI checks the document
  Then a finding names the actual slug "<slug>"

Examples header: [type slug pattern]
  row 1 -> {"type":"adr","slug":"adr-011-…","pattern":"^adr-[0-9]{3}$"}
  row 2 -> {"type":"frd","slug":"checkout","pattern":"^frd-[a-z-]+$"}
```

The parameterized steps are the governing general statement, and the rows are
the illustrative instances. The split survives, and it also becomes queryable,
where today an example is an unparseable sentence.

### Measured: the compiled form is the fact table

`gherkin.Pickles()` compiles the AST into one fully resolved row per example,
with tags already inherited from `Feature` and `Rule`:

```
pickle "Slug pattern mismatch…"  tags=[@frd-spec-index @AC-SPEC-003.5
                                       @level-unit]  steps=5
pickle "Slug pattern mismatch…"  tags=[…]  steps=5   <- 2nd Examples row
pickle "Files outside the spec root are never read"
                                 tags=[@frd-spec-index @AC-SPEC-001.5
                                       @negative @level-unit]  steps=1
```

Two scenarios produce three rows. That's the denormalized load target that a
structured store needs, and the library produces it. The whole document also
serializes to JSON by default (`{feature, comments}`), and `cucumber/messages`
is a published, versioned interchange schema. Adopting it is better than
inventing one, and it's what external test-management tools already ingest.

### Measured: rationale has a home, attached to the node

The concern that Gherkin has nowhere to put the corpus's `*Rationale:*`
paragraphs doesn't hold. Every description field is populated and attached to
its node:

```
Feature.Name        = "Document Index"
Feature.Description = "As an architecture owner / I want one machine-readable
                       artifact… / The index is the primary artifact, not a
                       by-product…"
Rule.Name           = "A build failure is the checker's job, not the indexer's"
```

Under the mapping that constraint 1 fixes, the parts are placed as follows:

- `Feature.Description` carries the requirement's user story and design note.
- A `Rule` name carries rationale that states an assertion governing several
  criteria.
- `Scenario.Description` carries per-criterion rationale.

Each becomes a field with an explicit scope, where today the scope of a
`*Rationale:*` paragraph is inferred from its position.

### The extraction mechanics are settled

Gherkin is written in fenced ` ```gherkin ` blocks in `.md` files. Under
constraint 1, each block in an FRD declares its own `Feature`, so each block is
a complete Gherkin document and needs no wrapper. The extractor still supplies
a synthetic `Feature:` header for any block that omits one. Positions map back
with `fence_line + (parser_line - wrap)`, where `wrap` is 0 for a block that
declares its own `Feature` and 1 for a wrapped fragment.

This was verified against a block that contains one good criterion and three
defective ones. The extractor reported correct `.md` line numbers, and it
reported all of the untagged, malformed, and duplicate identifiers. It was
separately verified end to end into this project's `Span`, whose byte offsets
were confirmed by slicing the source. MDG (`.feature.md`) is JavaScript-only
and isn't needed. The full extraction contract and its five gotchas are
recorded in `docs/research/gherkin/extraction-findings.md`.

Semantic validation stays in the spec CLI by design. The parser enforces
Gherkin's grammar, and the CLI enforces the following:

- Every scenario carries exactly one tag that matches
  `^@AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`.
- Identifiers are unique and in sequence.
- A `Scenario` `Description` doesn't contain what should have been a
  criterion.

A `Feature` description is exempt from the last rule, because under
constraint 1 it legitimately carries the user story and design note.

### Tried on a real document

We rewrote all of [[frd-spec-index]] in Gherkin as a trial (52 criteria in
nine blocks) and parsed it with the extractor. The findings are in
`docs/research/gherkin/ears-vs-gherkin-comparison.md`. All nine blocks parse
and all 52 identifiers extract, so the design holds.

The trial also settled where the FRD's other prose goes. The overview, the
terminology, and the numbering note stay in Markdown, because they're
FRD-level and Gherkin has no node above `Feature`. Everything at requirement
level moved inside the blocks: nine user stories moved into `Feature`
descriptions, and eight `*Rationale:*` paragraphs moved either into those
descriptions as design notes or into a `Rule:` where they stated a real rule.
None remains as loose prose. `Feature` descriptions hold multiple paragraphs
and pass Markdown through verbatim, subject to the two silent failures that
constraint 1 guards against. The authoring rules are drafted in
[[gherkin-guide]].

The Consequences section records three costs, so that they aren't discovered
during migration:

- Steps can't wrap, so long criteria become long lines.
- A `Scenario Outline` tagged at scenario level silently loses criteria.
- The Requirements section grows by 54%.

### What is genuinely lost, and what it costs

Prose isn't strictly worse. The following table lists what the change loses:

| Lost | Severity |
|---|---|
| **Grammatical negation.** `SHALL NOT` is greppable today; in a `Then` step a prohibition is ordinary prose | Replaced by a required `@negative` tag: better for machines, weaker for humans. 9 criteria (12%) are affected |
| **Modality.** Gherkin can't mark RFC 2119 levels, and its idiom uses "should" for mandatory behavior | **None for acceptance criteria.** All 73 are `SHALL` or `SHALL NOT`, with zero `SHOULD` and zero `MAY`. EARS keeps standards documents, where modality does vary |
| **Invariants read oddly.** 26% of criteria are ubiquitous, and they become `Then`-only scenarios | Real for a human reader; clean as a record (`Outcome` with no `Context` or `Action`) |
| **FRD narrative flow.** The document becomes prose (overview, user stories, rationale) interleaved with fenced blocks | Real and permanent |
| **Free rationale placement.** Rationale must attach to a `Feature`, `Rule`, or `Scenario` | A constraint, and a good one |

### The mapping the migration follows

The deleted draft at number 014 established this mapping. This decision
carries it forward and extends it with the tag conventions that it adds:

| EARS pattern | Gherkin form |
|---|---|
| `WHEN <event>`, the system SHALL `<action>` | `When <event>` / `Then <action>` |
| `WHILE <state>`, the system SHALL `<action>` | `Given <state>` / `Then <action>` |
| `IF <condition>`, THEN the system SHALL `<action>` | `Given <condition>` / `When …` / `Then <action>` |
| The system SHALL `<action>` (ubiquitous) | `Then <action>` with no trigger; this form reads least naturally |
| The system SHALL NOT `<action>` | `Then <action>` plus a required `@negative` tag; Gherkin has no native negative form |
| `WHERE <feature>`, the system SHALL `<action>` | A `Rule:` grouping, or a `Given` precondition |
| A criterion and its `*Example:*` | `Scenario Outline` with parameterized steps and one `Examples` row per instance |
| `(manual)` marker | `@manual` tag |

### Why this splits [[adr-002]] along a real boundary

The earlier draft counted "splits EARS into two scopes across 22 files" as a
leading cost. Under this framing, the split is principled rather than
arbitrary. Acceptance criteria are scenario-shaped, consumed externally, and
destined to leave the repository for product and QA tooling. Standards
documents, contracts, and blueprints are invariant-shaped, consumed
internally, and stay. EARS governs the second set, unchanged and unsplit
within itself.

## Decision

We will adopt Gherkin as the notation for features and acceptance criteria,
in two admissible forms:

- **Fenced ` ```gherkin ` blocks inside an FRD:** the default. Use this form
  where the feature benefits from surrounding Markdown, such as an overview,
  terminology, references, and document frontmatter.
- **Standalone `.feature` files:** for a feature that needs no surrounding
  prose, and for publication to external tooling.

Both forms carry the same tagged content and produce the same extracted
records. The spec CLI parses them with the official Go parser and performs
semantic validation itself.

An earlier revision admitted only the first form, because a `.feature` file
can't carry frontmatter. That's true, because a raw `---` block doesn't
parse, but it was the wrong objection: a feature file isn't a document. It
holds features, its owning document is the one that its tag names, and it
needs no metadata of its own. The adoption FRD leaves open whether a feature
file should ever have metadata.

EARS remains the notation for every other normative statement: standards
documents, contracts, blueprints, and requirement-level statements. This
decision amends the scope of [[adr-002]] and doesn't retire it.

Binding constraints:

1. A requirement SHALL map to a `Feature`, its user story to that `Feature`'s
   description, and each acceptance criterion to a `Scenario` or
   `Scenario Outline`. `Rule:` SHALL be used only where a genuine business
   rule groups several criteria; it names an assertion, not a topic. Because a
   Gherkin document holds exactly one `Feature`, this means one fenced block,
   or one `.feature` file, per requirement. That description carries the user
   story, and MAY carry a design note and a scope fence. It SHALL NOT use `#`,
   which Gherkin reads as a comment, and no line in it may begin with a
   Gherkin keyword.
2. Every `Scenario` SHALL carry exactly one tag matching
   `^@AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`; the CLI SHALL report a finding
   otherwise.
3. WHERE a `Scenario Outline` carries more than one acceptance criterion,
   each identifier SHALL be tagged on its own `Examples` block rather than on
   the scenario.
4. A criterion expressing a prohibition SHALL carry `@negative`, and one
   verifiable only by hand SHALL carry `@manual`, replacing the `(manual)`
   text marker.
5. `Scenario` `Description` fields SHALL NOT contain normative statements;
   the CLI SHALL report prose there as a probable untagged criterion. A
   `Feature` description carries the user story and is exempt.
6. WHERE a criterion has concrete example data, it SHOULD be expressed as a
   `Scenario Outline` with an `Examples` table rather than as prose.
7. A step SHALL NOT be shortened to fit a line-width limit. No such limit
   applies inside a fenced block (AC-SPECSTR-002.5). The line is written at
   whatever width the criterion needs, and is decomposed into further steps
   only where that genuinely reads better.
8. The CLI SHALL be able to emit conforming `.feature` files from the
   extracted blocks, so publication to external tooling needs no second
   authoring format. One requirement yields one file.
9. Extraction SHALL read Gherkin only. The FRDs that still carry prose
   criteria SHALL be translated rather than read by a second extractor. A
   compatibility reader is transitional machinery for a transition of three
   documents, and it would outlive the transition in the model and in every
   consumer of it. The extracted record SHALL still state the notation it was
   written in, so the model stays notation-agnostic where that costs one
   property rather than a parser. This narrows the notation-agnostic
   constraint carried forward from ADR-014, which assumed a compatibility
   window that this corpus is too small to need.
10. Every `Feature` SHALL carry an identifier composed of a prefix, a short
    uppercase group code, and a number, such as `FEAT-GHK-001`, that
    identifies it independently of the document that holds it. Features SHALL
    share a code, which is what makes the code a grouping axis; a code and a
    number together identify one feature. Every acceptance-criterion
    identifier SHALL derive its code and number from its feature's. Grouping
    is by code rather than by a path-shaped slug, because the framework's slug
    vocabulary admits only lowercase alphanumerics and hyphens, so a
    hyphenated slug can't be split back into segments unambiguously.
11. A standalone `.feature` file SHALL be a feature source rather than a
    document. It carries no frontmatter, and its owning document is the one
    that its `@frd-<slug>` tag names. Corpus membership admits only `.md`
    files (AC-SPEC-001.6) and continues to; feature sources are discovered
    separately rather than admitted as members. A feature written as a fenced
    block SHALL NOT carry that tag. The block is written inside the document
    that owns it, whose frontmatter already declares the slug, so the tag
    would restate a fact that the CLI derives and could come to contradict it.
14. A wikilink written anywhere inside a Gherkin source SHALL be recorded as a
    reference. Gherkin treats `[[slug]]` as ordinary text in every position
    (description, scenario name, step, rule name, and examples cell), so
    nothing about the notation argues for treating it differently from a
    wikilink in prose.
12. The extracted record SHALL carry a content hash over the criterion's name
    and text, excluding node `Id`s and `Location`s, so that a criterion that
    moved but didn't change hashes identically. Measured: the same scenario
    shifted down a file hashes the same, and a one-word edit changes the hash.
13. The extracted record SHALL treat the criterion as the atomic unit, and
    carry its body as text rather than decompose it into typed steps and
    example rows. Decomposition models one notation, because those parts are
    recoverable from Gherkin and from almost nothing else, and no consumer
    that this framework has needs them. The index schema is additive, so
    decomposition can be introduced later; withdrawing it couldn't.

Constraint 12 makes "which criteria changed locally" answerable. Node `Id`s
are positional counters and shift on any edit above them, so they can't carry
identity. The hash is taken over the scenario name, its typed steps, and its
example rows.

Constraint 1 is the one the trial got wrong first. Mapping a requirement onto
`Rule:` produced blocks headed `Rule: Configuration`, which names a topic
where the keyword means an assertion. Because `Rule.Description` was then the
only slot available, the user story had nowhere to go but Markdown. Mapping
the requirement onto `Feature` puts the user story in the description slot
built for it, and frees `Rule:` for its purpose.

Constraint 3 exists because the trial hit it. Tagging one `Scenario Outline`
at scenario level made two criteria disappear from a document that parsed
cleanly, with every scenario tagged and no check firing. Constraint 7 exists
because the trial's first pass invented a width limit that doesn't apply and
compressed steps to fit it, which changed the meaning of AC-SPEC-009.4. The
hazard is the invented constraint, not the notation.

Constraint 8 is why the fenced form is the default rather than the only form.
Emitting `.feature` files from fenced blocks is mechanical, so the FRD stays
one reviewable document, and the external artifact is available on demand.
Level 3 pays the corpus-model cost to obtain something that level 2 can
generate: `.feature` files are inadmissible under AC-SPEC-001.6 and
AC-SPEC-001.3, and they carry no frontmatter, no slug, and no `status`.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **No adoption: EARS prose, with an AC index built over it** | No migration; [[adr-002]] keeps one scope<br>`SHALL NOT` stays greppable<br>FRD reads as continuous narrative<br>No new dependency or import churn | Yields one string per criterion; every typed column needs a hand-written EARS parser over English<br>Examples stay unparseable prose<br>Criterion enumeration stays a bullet-matching heuristic that must tolerate 61 `*Example:*` sub-bullets and 9 `*Rationale:*` paragraphs<br>Nothing downstream ingests it: no interchange schema, no tool support | Adequate for a Markdown corpus, inadequate for a typed store; it optimizes for the destination we're leaving |
| **`.feature` files as the only form** (level 3 alone) | Directly ingestible by external tooling with no export step<br>Proper editor highlighting | A feature that needs an overview, terminology, or references has nowhere to put them<br>Corpus membership admits only `.md` today, so every standalone file needs the amendment of constraint 11 | Not rejected; adopted alongside the fenced form. Rejected only as the sole form, because much of this corpus's value is the prose around a feature |
| **Structured AC record blocks:** a fenced block per criterion, with the schema authored in Strata | Total schema control, matching whatever the central store defines<br>Reuses [[adr-008]] record blocks and the [[adr-016]] pipeline<br>Zero new dependencies, generated Go structs | Bespoke: no external tool ingests it, so the product and QA tooling destination needs a translator anyway<br>No published interchange schema<br>Far more verbose; an FRD becomes barely readable | The strongest rival, and the right answer if the store were the only destination. Rejected because the stated endgame is shared product, engineering, and QA tooling, and interoperability is the whole value there |
| **Gherkin blocks embedded in an FRD** (level 2) | Typed `Context`/`Action`/`Outcome` steps instead of one string<br>`Scenario Outline` + `Examples` preserves the criterion/example split and makes examples queryable<br>`Pickles()` emits a denormalized fact table with tags inherited<br>`cucumber/messages` is a published, versioned interchange schema that external tools already read<br>Rationale attaches to `Rule`/`Scenario` nodes<br>AST enumeration makes an untagged criterion a countable node<br>One artifact; `.feature` emitted on demand | Migrates 73 criteria plus four reference documents<br>Loses grammatical negation to a `@negative` tag convention<br>26% ubiquitous criteria read oddly as `Then`-only scenarios<br>Import path carries the major version, which was bumped six times in nine months on other languages' releases<br>Linguist's Gherkin grammar is unmaintained; VS Code needs an extension<br>FRD narrative becomes prose interleaved with fenced blocks | N/A: selected |

## Consequences

**Positive:**

- Acceptance criteria extract as typed records (precondition, trigger,
  outcome, polarity, verification level, and example rows) without a
  hand-written parser over English.
- Example data becomes queryable structured rows rather than prose.
- The corpus gains a published interchange schema (`cucumber/messages`), so
  the central store and any external product or QA tool consume a format that
  already exists.
- Criterion enumeration is an AST walk, so an untagged criterion is a
  countable node rather than an unremarkable sentence.
- Rationale attaches to a node, which gives it explicit rather than positional
  scope.
- `.feature` export makes the eventual move of criteria into shared tooling a
  build step, not a migration.

**Negative:**

- Prohibitions lose grammatical marking. Finding them depends on the
  `@negative` tag convention holding, which is now a lint rather than a
  property of the sentence.
- Ubiquitous criteria, which are 26% of the corpus, are expressed as
  `Then`-only scenarios. That's legal, but it reads as an example of nothing.
- An FRD is no longer continuous prose; readers alternate between narrative
  and fenced blocks.
- The scope of [[adr-002]] narrows, so authors must know which surface
  they're on, and both readiness checklists carry two syntaxes.
- The Go import path changes with Gherkin's shared major version, roughly
  every six weeks, on releases driven by other languages.
- Gherkin's cultural weight implies executable tests. These scenarios are
  specification text and no runner executes them, so they can drift toward
  imperative detail without a runner to catch it.
- Syntax highlighting is partial: GitHub's Linguist grammar is unmaintained
  and miscolors `Rule:`, and VS Code requires an extension.
- A step can't wrap, so a criterion that takes 150 columns to state occupies
  one 150-column line. Gherkin imposes no length limit, and this framework
  already exempts fenced blocks from `line_width` (AC-SPECSTR-002.5), so
  nothing forces shortening. However, a long line reads worse in a narrow
  diff than a wrapped prose bullet does.
- A mistagged `Scenario Outline` loses criteria silently. Tagging the
  scenario rather than each `Examples` block collapses several criteria into
  one, in a document that parses, whose every scenario is tagged, and against
  which no check fires. Constraint 7 and its conformance check are the only
  defense.
- The Requirements section grows by about 54%, measured from 410 to 633
  lines, or from 7.9 to 12.2 lines per criterion. There are two causes: 25 of
  52 criteria need a `Scenario Outline` to hold a single example row, and the
  requirement-level prose that was Markdown now sits inside the blocks.

**Revisit when** the central store's schema is actually designed. If it
diverges materially from `cucumber/messages`, the structured-record
alternative regains its advantage, because we would then be translating out of
one fixed shape into another rather than authoring the target shape directly.

## References

- [[adr-002]]: EARS, whose scope this ADR narrows to non-AC normative
  statements.
- [[adr-015]]: chose Go partly for the official in-process Gherkin parser,
  which this decision exercises.
- [[adr-008]], [[adr-016]]: the record-block and Strata mechanisms behind the
  structured-record alternative.
- [[frd-spec-index]]: the central-publication intent this ADR optimizes for,
  and the corpus membership rules that level 3 contradicts.
- `plugin/skills/spec/references/frd-template.md`: the criterion/example
  split that `Scenario Outline` preserves.
- `plugin/skills/spec/references/ac-traceability-standards.md`: the
  identifier format carried forward as the tag vocabulary.
- [cucumber/gherkin](https://github.com/cucumber/gherkin): the parser and its
  release history.
- [cucumber/messages](https://github.com/cucumber/messages): the published
  interchange schema that the extracted records conform to.
- `docs/research/gherkin/extraction-findings.md`: the measured extraction
  contract, the AST type hierarchy, and the five integration gotchas.
- `docs/research/gherkin/ears-vs-gherkin-comparison.md`: the trial migration
  of a real FRD and what it cost. The trial document itself was removed once
  [[frd-gherkin-adoption]] became the worked example.
- [[gherkin-guide]]: how to write conforming Gherkin under this decision.
