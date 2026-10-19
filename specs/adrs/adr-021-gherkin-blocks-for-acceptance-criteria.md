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

# ADR-021: Adopt Gherkin for Features and Acceptance Criteria

**Status:** approved  
**Date:** 2026-08-06

## Context

An earlier draft ADR, numbered 014, deferred this question and named the
way to settle it: write real criteria in both notations and compare them.
That comparison has now been run twice against this repository's own
corpus. This ADR replaces the draft. The draft is deleted rather than kept
as a retired record, because it was never ratified and the decision log
needs one document per question. ADR number 014 stays unused.

The first pass judged the notation against a **local Markdown
destination**: is an FRD easier to read, review, and lint? On that test,
adoption narrowly lost.

That was the wrong destination. The stated direction is to publish from
these repositories into a **central structured store for C4 models,
features, and acceptance criteria**, into a knowledge graph fed the same
content, or into both. Beyond that, acceptance criteria are meant to live
in product and QA tooling that product managers, engineers, and QA share,
outside any source repository. [[frd-spec-index]] already states this
intent: "an organization pushes the index to a central service and sees
its whole architecture."

For a typed-record destination, the criteria reverse. The question is no
longer how a sentence reads in a diff. It's **how much structure survives
extraction**, and on that question prose loses clearly.

All measurements in this ADR use `github.com/cucumber/gherkin/go/v42`
(v42.0.1, MIT).

### Measured: Gherkin yields typed fields; prose yields one string

A structured store needs an acceptance criterion as columns:
precondition, trigger, expected outcome, polarity, verification level,
example data, and provenance. Prose EARS delivers **one string**, such as
`"WHEN the CLI is invoked against a directory, the CLI SHALL treat…"`.
Every column beyond that requires someone to write and maintain an EARS
parser over natural language.

The Gherkin AST delivers those columns directly. Each step carries a
typed `KeywordType` as well as its text:

```
scenario "Slug pattern mismatch is reported with actual and expected"
  keyword="Given"  keywordType=Context      text="a document of type <type>…"
  keyword="And"    keywordType=Conjunction  text="the profile registers…"
  keyword="When"   keywordType=Action       text="the CLI checks the document"
  keyword="Then"   keywordType=Outcome      text="a finding names the actual…"
```

`Context`, `Action`, and `Outcome` are precondition, trigger, and
expected result as an enum. A maintained parser produces them, not our
own regular expressions over English.

**Constraint 13 declines to persist these typed steps.** The capability
is real, and it's what distinguishes a parsed notation from prose. But
nothing this framework does works below the level of a whole criterion.
A record that carried typed steps could be filled only from Gherkin,
which would make the model Gherkin-shaped in a framework that treats the
notation as replaceable. The AST keeps the capability available for when
a consumer needs it. The manifest doesn't commit to it in advance.

### Measured: `Scenario Outline` preserves the criterion/example split

An earlier draft of this ADR found that a Gherkin scenario collapses the
distinction the FRD template depends on. In that template, the EARS
sentence carries the requirement, the `*Example:*` beneath it only
illustrates, and the criterion governs when the two conflict. **That
finding was wrong.** It was made against a plain `Scenario`, and the
corpus's 61 examples are exactly the case that `Scenario Outline` exists
for:

```
general (parameterised) form:
  Given a document of type <type> declaring slug "<slug>"
  When the CLI checks the document
  Then a finding names the actual slug "<slug>"

Examples header: [type slug pattern]
  row 1 -> {"type":"adr","slug":"adr-011-…","pattern":"^adr-[0-9]{3}$"}
  row 2 -> {"type":"frd","slug":"checkout","pattern":"^frd-[a-z-]+$"}
```

The parameterized steps are the governing general statement, and the
rows are the illustrative instances. The split survives, and it also
becomes **queryable**. Today, an example is a sentence that nothing can
parse.

### Measured: the compiled form is the fact table

`gherkin.Pickles()` compiles the AST into one fully resolved row per
example. Each row already inherits its tags from the `Feature` and
`Rule`:

```
pickle "Slug pattern mismatch…"  tags=[@frd-spec-index @AC-SPEC-003.5
                                       @level-unit]  steps=5
pickle "Slug pattern mismatch…"  tags=[…]  steps=5   <- 2nd Examples row
pickle "Files outside the spec root are never read"
                                 tags=[@frd-spec-index @AC-SPEC-001.5
                                       @negative @level-unit]  steps=1
```

Two scenarios produce three rows. That's the denormalized load target a
structured store wants, and the library produces it. The whole document
also serializes to JSON without extra work (`{feature, comments}`).
`cucumber/messages` is a **published, versioned interchange schema**.
Adopting it is better than inventing one, and external test-management
tools already ingest it.

### Measured: rationale has a home, attached to the node

One concern was that Gherkin has no place for the corpus's
`*Rationale:*` paragraphs. That concern doesn't hold. The parser
populates every description field and attaches it to its node:

```
Feature.Name        = "Document Index"
Feature.Description = "As an architecture owner / I want one machine-readable
                       artifact… / The index is the primary artifact, not a
                       by-product…"
Rule.Name           = "A build failure is the checker's job, not the indexer's"
```

Constraint 1 fixes the mapping:

- `Feature.Description` carries the requirement's user story and design
  note.
- A `Rule` **name** carries rationale that states an assertion governing
  several criteria.
- `Scenario.Description` carries rationale for a single criterion.

Each becomes a field with an explicit scope. Today, a reader infers a
`*Rationale:*` paragraph's scope from where it sits.

### The extraction mechanics are settled

Gherkin goes in fenced ` ```gherkin ` blocks in `.md` files. Under
constraint 1, each block in an FRD declares its own `Feature`, so each is
a complete Gherkin document and needs no wrapper. The extractor still
supplies a synthetic `Feature:` header for any block that omits one.
Positions map back with `fence_line + (parser_line - wrap)`, where `wrap`
is 0 for a block that declares its own `Feature` and 1 for a wrapped
fragment.

Two tests verified this. The first used a block with one good criterion
and three defective ones. It produced correct `.md` line numbers and
reported the untagged, malformed, and duplicate identifiers. The second
ran end to end into this project's `Span`, and slicing the source
confirmed its byte offsets. **MDG (`.feature.md`) is JavaScript-only, and
this design doesn't need it.** For the full extraction contract and its
five gotchas, see `docs/research/gherkin/extraction-findings.md`.

By design, semantic validation stays in the spec CLI. The parser enforces
Gherkin's grammar. The CLI enforces the following:

- Every scenario carries exactly one tag that matches
  `^@AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`.
- Identifiers are unique and in sequence.
- A `Scenario` `Description` doesn't contain what should have been a
  criterion.

A `Feature` description is exempt from the last check, because under
constraint 1 it legitimately carries the user story and design note.

### Tried on a real document

As a trial, all of [[frd-spec-index]] was rewritten in Gherkin, giving
52 criteria in nine blocks, and parsed with the extractor. The findings
are in `docs/research/gherkin/ears-vs-gherkin-comparison.md`. All nine
blocks parse and all 52 identifiers extract, so the design holds.

The trial also settled where the FRD's other prose goes. The overview,
terminology, and numbering note stay in Markdown, because they belong to
the FRD as a whole and Gherkin has no node above `Feature`. Everything at
requirement level moved inside the blocks:

- Nine user stories moved into `Feature` descriptions.
- Eight `*Rationale:*` paragraphs moved either into those descriptions as
  design notes, or into a `Rule:` where they stated a real rule.

No loose prose remains. `Feature` descriptions hold multiple paragraphs
and pass Markdown through verbatim, subject to the two silent failures
that constraint 1 guards against. The authoring rules are drafted in
[[gherkin-guide]].

The trial found three costs, which Consequences records so that nobody
first discovers them during migration:

- Steps can't wrap, so long criteria become long lines.
- A `Scenario Outline` tagged at scenario level silently loses criteria.
- The Requirements section grows 54%.

### What is genuinely lost, and what it costs

Prose isn't strictly worse. The following table lists what adoption
loses and how much each loss matters:

| Lost | Severity |
|---|---|
| **Grammatical negation.** You can grep for `SHALL NOT` today. In a `Then` step, a prohibition is ordinary prose | A required `@negative` tag replaces it. That's better for machines and weaker for human readers. Affects 9 criteria (12%) |
| **Modality.** Gherkin can't mark RFC 2119 levels, and its idiom uses "should" for mandatory behavior | **None for acceptance criteria.** All 73 are `SHALL` or `SHALL NOT`, with no `SHOULD` and no `MAY`. EARS keeps standards documents, where modality does vary |
| **Invariants read oddly.** 26% of criteria are ubiquitous and become `Then`-only scenarios | Real for a human reader. Clean as a record: an `Outcome` with no `Context` or `Action` |
| **FRD narrative flow.** The document becomes prose (overview, user stories, rationale) interleaved with fenced blocks | Real and permanent |
| **Free rationale placement.** Rationale must attach to a `Feature`, `Rule`, or `Scenario` | A constraint, and a useful one |

### The mapping the migration follows

The deleted draft numbered 014 established this mapping. This decision
carries it forward and adds tag conventions:

| EARS pattern | Gherkin form |
|---|---|
| `WHEN <event>`, the system SHALL `<action>` | `When <event>` / `Then <action>` |
| `WHILE <state>`, the system SHALL `<action>` | `Given <state>` / `Then <action>` |
| `IF <condition>`, THEN the system SHALL `<action>` | `Given <condition>` / `When …` / `Then <action>` |
| The system SHALL `<action>` (ubiquitous) | `Then <action>` with no trigger. This form reads least naturally |
| The system SHALL NOT `<action>` | `Then <action>` plus a required `@negative` tag. Gherkin has no native negative form |
| `WHERE <feature>`, the system SHALL `<action>` | A `Rule:` grouping, or a `Given` precondition |
| Criterion plus its `*Example:*` | `Scenario Outline` with parameterized steps and one `Examples` row per instance |
| `(manual)` marker | `@manual` tag |

### Why this splits [[adr-002]] along a real boundary

The earlier draft counted "splits EARS into two scopes across 22 files"
as a leading cost. Under this framing, the split follows a **principle,
not an accident**. Acceptance criteria are scenario-shaped, consumed
outside the repository, and headed for product and QA tooling. Standards
documents, contracts, and blueprints are invariant-shaped, consumed
inside the repository, and stay there. EARS governs that second set,
unchanged and with no split inside it.

## Decision

We will adopt **Gherkin** as the notation for features and acceptance
criteria, in **two admissible forms**:

- **Fenced ` ```gherkin ` blocks inside an FRD.** This is the default.
  Use it when the feature benefits from surrounding Markdown: an
  overview, terminology, references, and document frontmatter.
- **Standalone `.feature` files.** Use them for a feature that needs no
  surrounding prose, and for publication to external tooling.

Both forms carry the same tagged content and produce the same extracted
records. The spec CLI parses both with the official Go parser and does
the semantic validation itself.

An earlier revision admitted only the first form, because a `.feature`
file can't carry frontmatter. That's true, since a raw `---` block
doesn't parse, but it was the wrong objection. A feature file isn't a
document. It holds features, its owning document is the one its tag
names, and it needs no metadata of its own. The adoption FRD leaves open
whether it should ever have any.

EARS remains the notation for every other normative statement: standards
documents, contracts, blueprints, and requirement-level statements. This
ADR amends the scope of [[adr-002]] and doesn't retire it.

The binding constraints are as follows:

1. A requirement SHALL map to a `Feature`, its user story to that
   `Feature`'s description, and each acceptance criterion to a
   `Scenario` or `Scenario Outline`. `Rule:` SHALL be used only where a
   real business rule groups several criteria. A rule names an
   assertion, not a topic. Because a Gherkin document holds exactly one
   `Feature`, each requirement gets one fenced block or one `.feature`
   file. The `Feature` description carries the user story, and MAY carry
   a design note and a scope fence. It SHALL NOT use `#`, which Gherkin
   reads as a comment, and no line in it may begin with a Gherkin
   keyword.
2. Every `Scenario` SHALL carry exactly one tag matching
   `^@AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`. Otherwise, the CLI SHALL report a
   finding.
3. WHERE a `Scenario Outline` carries more than one acceptance
   criterion, each identifier SHALL be tagged on its own `Examples`
   block rather than on the scenario.
4. A criterion that expresses a prohibition SHALL carry `@negative`. A
   criterion that can be verified only by hand SHALL carry `@manual`,
   which replaces the `(manual)` text marker.
5. `Scenario` `Description` fields SHALL NOT contain normative
   statements. The CLI SHALL report prose there as a probable untagged
   criterion. A `Feature` description carries the user story and is
   exempt.
6. WHERE a criterion has concrete example data, it SHOULD be expressed
   as a `Scenario Outline` with an `Examples` table rather than as prose.
7. A step SHALL NOT be shortened to fit a line-width limit. No such
   limit applies inside a fenced block (AC-SPECSTR-002.5). Write the line
   at whatever width the criterion needs. Split it into further steps
   only where that reads better.
8. The CLI SHALL be able to emit conforming `.feature` files from the
   extracted blocks, so that publishing to external tooling needs no
   second authoring format. One requirement yields one file.
9. Extraction SHALL read Gherkin only. The FRDs that still carry prose
   criteria SHALL be translated rather than read by a second extractor.
   A compatibility reader is transitional machinery for a transition of
   three documents, and it would outlive that transition in the model and
   in every consumer of the model. The extracted record SHALL still state
   the notation it was written in, so the model stays notation-agnostic
   where that costs one property rather than a parser. This narrows the
   notation-agnostic constraint carried forward from ADR-014, which
   assumed a compatibility window that this corpus is too small to need.
10. Every `Feature` SHALL carry an identifier, such as `FEAT-GHK-001`,
    made of a prefix, a short uppercase group code, and a number. The
    identifier names the feature independently of the document that
    holds it. Features SHALL share a code, which is what makes the code
    a grouping axis. A code and a number together identify one feature.
    Every acceptance-criterion identifier SHALL derive its code and
    number from its feature's identifier. Grouping uses a code rather
    than a path-shaped slug because the framework's slug vocabulary
    admits only lowercase alphanumerics and hyphens, so a hyphenated slug
    can't be split back into its segments without ambiguity.
11. A standalone `.feature` file SHALL be a feature source rather than a
    document. It carries no frontmatter, and its owning document is the
    one its `@frd-<slug>` tag names. Corpus membership admits only `.md`
    files (AC-SPEC-001.6), and that doesn't change. The CLI discovers
    feature sources separately instead of admitting them as members. A
    feature written as a fenced block SHALL NOT carry that tag. The block
    sits inside the document that owns it, and that document's
    frontmatter already declares the slug. The tag would restate a fact
    the CLI derives, and it could come to contradict that fact.
14. A wikilink written anywhere inside a Gherkin source SHALL be recorded
    as a reference. Gherkin treats `[[slug]]` as ordinary text in every
    position: description, scenario name, step, rule name, and examples
    cell. Nothing about the notation argues for treating it differently
    from a wikilink in prose.
12. The extracted record SHALL carry a content hash over the criterion's
    name and text, excluding node `Id`s and `Location`s, so that a
    criterion that moved but didn't change hashes identically. Measured:
    the same scenario shifted down a file hashes the same, and a
    one-word edit changes the hash.
13. The extracted record SHALL treat the criterion as the atomic unit,
    and carry its body as text rather than decomposing it into typed
    steps and example rows. Decomposition models one notation, because
    those parts can be recovered from Gherkin and from almost nothing
    else, and no consumer of this framework needs them. The index schema
    is additive, so decomposition can be added later. Withdrawing it
    later wouldn't be possible.

Constraint 12 makes it possible to answer "which criteria changed
locally." Node `Id`s are positional counters that shift on any edit above
them, so they can't carry identity. The hash is taken over the scenario
name, its typed steps, and its example rows.

The trial got constraint 1 wrong at first. Mapping a requirement onto
`Rule:` produced blocks headed `Rule: Configuration`, which is a topic,
where the keyword means an assertion. Because `Rule.Description` was then
the only slot available, the user story had nowhere to go except
Markdown. Mapping the requirement onto `Feature` puts the user story in
the description slot built for it, and frees `Rule:` for its intended
purpose.

Constraint 3 exists because the trial hit the problem. Tagging one
`Scenario Outline` at scenario level made two criteria disappear from a
document that parsed cleanly, in which every scenario was tagged and no
check fired. Constraint 7 exists because the trial's first pass invented
a width limit that doesn't apply and compressed steps to fit it. That
changed the meaning of AC-SPEC-009.4. The hazard is the invented
constraint, not the notation.

Constraint 8 is why the fenced form is the default and not the only
form. Emitting `.feature` files from fenced blocks is mechanical, so the
FRD stays one reviewable document *and* the external artifact is
available on demand. Level 3 pays the corpus-model cost to get something
level 2 can generate: `.feature` files are inadmissible under
AC-SPEC-001.6 and AC-SPEC-001.3, and carry no frontmatter, no slug, and
no `status`.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **No adoption — EARS prose, with an AC index built over it** | No migration; [[adr-002]] keeps one scope<br>`SHALL NOT` stays greppable<br>FRD reads as continuous narrative<br>No new dependency or import churn | Yields **one string per criterion**, so every typed column needs a hand-written EARS parser over English<br>Examples stay unparseable prose<br>Criterion enumeration stays a bullet-matching heuristic that must tolerate 61 `*Example:*` sub-bullets and 9 `*Rationale:*` paragraphs<br>Nothing downstream ingests it: no interchange schema and no tool support | Adequate for a Markdown corpus but inadequate for a typed store. It optimizes for the destination we're leaving |
| **`.feature` files as the only form** (level 3 alone) | External tooling can ingest it directly with no export step<br>Proper editor highlighting | A feature that wants an overview, terminology, or references has nowhere to put them<br>Corpus membership admits only `.md` today, so every standalone file needs the amendment in constraint 11 | Not rejected. It's **adopted alongside** the fenced form, and rejected only as the *sole* form, because much of this corpus's value is the prose around a feature |
| **Structured AC record blocks** — a fenced block per criterion, schema authored in Strata | Total schema control, matching whatever the central store defines<br>Reuses [[adr-008]] record blocks and the [[adr-016]] pipeline<br>No new dependencies, and generated Go structs | Bespoke: no external tool ingests it, so the product and QA tooling destination needs a translator anyway<br>No published interchange schema<br>Far more verbose, so an FRD becomes barely readable | The strongest rival, and the right answer if the store were the *only* destination. Rejected because the stated end goal is tooling that product, engineering, and QA share, and there interoperability is the whole value |
| **Gherkin blocks embedded in an FRD** (level 2) | Typed `Context`/`Action`/`Outcome` steps instead of one string<br>`Scenario Outline` + `Examples` preserves the criterion/example split and makes examples queryable<br>`Pickles()` emits a denormalized fact table with tags inherited<br>`cucumber/messages` is a published, versioned interchange schema that external tools already read<br>Rationale attaches to `Rule`/`Scenario` nodes<br>AST enumeration makes an untagged criterion a countable node<br>One artifact, with `.feature` emitted on demand | Migrates 73 criteria plus four reference documents<br>Loses grammatical negation to a `@negative` tag convention<br>26% ubiquitous criteria read oddly as `Then`-only scenarios<br>The import path carries the major version, which was bumped six times in nine months for other languages' releases<br>Linguist's Gherkin grammar is unmaintained, and VS Code needs an extension<br>FRD narrative becomes prose interleaved with fenced blocks | N/A — selected |

## Consequences

**Positive:**
- Acceptance criteria extract as typed records (precondition, trigger,
  outcome, polarity, verification level, and example rows) without a
  hand-written parser over English.
- Example data becomes queryable structured rows rather than prose.
- The corpus gains a published interchange schema (`cucumber/messages`),
  so the central store and any external product or QA tool consume a
  format that already exists.
- Criterion enumeration is an AST walk, so an untagged criterion is a
  countable node rather than a sentence nobody notices.
- Rationale attaches to a node, which gives it an explicit scope rather
  than one inferred from position.
- `.feature` export makes the eventual move of criteria into shared
  tooling a build step, not a migration.

**Negative:**
- Prohibitions lose grammatical marking. Finding them depends on authors
  following the `@negative` tag convention, which is now a lint rather
  than a property of the sentence.
- Ubiquitous criteria, 26% of the corpus, are expressed as `Then`-only
  scenarios. That's legal, but each one reads as an example of nothing.
- An FRD is no longer continuous prose. Readers alternate between
  narrative and fenced blocks.
- [[adr-002]]'s scope narrows, so authors must know which surface they're
  on, and both readiness checklists carry two syntaxes.
- The Go import path changes with Gherkin's shared major version, about
  every six weeks, on releases driven by other languages.
- Gherkin is strongly associated with executable tests. These scenarios
  are specification text and no runner executes them. They can drift
  toward imperative detail, and no failing runner will catch it.
- Syntax highlighting is partial. GitHub's Linguist grammar is
  unmaintained and colors `Rule:` incorrectly, and VS Code requires an
  extension.
- **A step can't wrap**, so a criterion that takes 150 columns to state
  occupies one 150-column line. Gherkin imposes no length limit, and this
  framework already exempts fenced blocks from `line_width`
  (AC-SPECSTR-002.5), so nothing forces shortening. But a long line reads
  worse in a narrow diff than a wrapped prose bullet does.
- **A mis-tagged `Scenario Outline` loses criteria silently.** Tagging
  the scenario rather than each `Examples` block collapses several
  criteria into one. The document still parses, every scenario is
  tagged, and no check fires. Constraint 7 and its conformance check are
  the only defense.
- **The Requirements section grows about 54%**, measured as 410 to 633
  lines, or 7.9 to 12.2 lines per criterion. There are two causes. First,
  25 of 52 criteria need a `Scenario Outline` to hold a single example
  row. Second, the requirement-level prose that was Markdown now sits
  inside the blocks.

**Revisit when** the central store's schema is designed. If it differs
materially from `cucumber/messages`, the structured-record alternative
regains its advantage. At that point we would be translating from one
fixed shape into another, rather than authoring the target shape
directly.

## References

- [[adr-002]] — EARS, whose scope this ADR narrows to normative
  statements other than acceptance criteria
- [[adr-015]] — chose Go partly for the official in-process Gherkin
  parser, which this decision uses
- [[adr-008]], [[adr-016]] — the record-block and Strata mechanisms
  behind the structured-record alternative
- [[frd-spec-index]] — the central-publication intent that this ADR
  optimizes for, and the corpus membership rules that level 3
  contradicts
- `plugin/skills/spec/references/frd-template.md` — the criterion/example
  split that `Scenario Outline` preserves
- `plugin/skills/spec/references/ac-traceability-standards.md` — the
  identifier format carried forward as the tag vocabulary
- [cucumber/gherkin](https://github.com/cucumber/gherkin) — parser and
  release history
- [cucumber/messages](https://github.com/cucumber/messages) — the
  published interchange schema that the extracted records conform to
- `docs/research/gherkin/extraction-findings.md` — the measured extraction
  contract, the AST type hierarchy, and the five integration gotchas
- `docs/research/gherkin/ears-vs-gherkin-comparison.md` — the trial
  migration of a real FRD and what it cost. The trial document itself was
  removed after [[frd-gherkin-adoption]] became the worked example.
- [[gherkin-guide]] — how to write conforming Gherkin under this decision
