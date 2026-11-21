---
title: Gherkin Authoring Guide
description: How to write requirements and acceptance criteria in Gherkin —
  which concept maps to which keyword, what belongs in each description
  slot, and the constructs that parse cleanly while losing meaning
slug: gherkin-guide
type: guide
status: active
---

# Gherkin authoring guide

Gherkin is the notation of Cucumber and its relatives: a `Feature` holding
`Scenario`s, each written as `Given`/`When`/`Then` steps. This guide covers
using it to write requirements and their acceptance criteria, in a Markdown
document or in standalone `.feature` files.

Read this guide before you write a feature.

## Contents

- The mapping — which requirements concept becomes which Gherkin keyword
- Two forms — fenced blocks and standalone files
- Tags — identifiers and qualifiers
- Descriptions — what belongs in each slot
- Writing good scenarios
- Seven ways to break a document silently
- References

---

## The mapping

| Concept | Gherkin |
|---|---|
| A requirement | `Feature:` |
| Its identity | a tag on the `Feature` |
| Its user story | the `Feature` description |
| An acceptance criterion | `Scenario` or `Scenario Outline` |
| A business rule grouping several criteria | `Rule:` |
| A criterion's concrete example | an `Examples` row |
| Rationale | a design note in the description, or a `Rule:` name |

**A requirement is a `Feature`, not a `Rule`.** A `Rule` names a business
rule: an assertion that can be true or false. "There can be only one" is a
rule. "Configuration" is a topic, and a topic is a `Feature`. Swapping the two
is the most common structural error, and the error cascades: with the
requirement on `Rule:`, the only description slot is spent and the user story
has nowhere to go.

**A Gherkin document holds exactly one `Feature`.** Two `Feature:` lines in
one document don't parse, so one requirement means one fenced block or one
`.feature` file.

---

## Two forms

### Fenced blocks inside a Markdown document

Use a fenced block when the feature benefits from surrounding Markdown: an
overview, terminology, references, or document frontmatter.

````markdown
### Payment confirmation

```gherkin
@REQ-CHK-003
Feature: Payment Confirmation

  As a customer
  I want confirmation after my payment is processed
  So that I know my order was placed

  @AC-CHK-003.1
  Scenario: A successful payment shows the order number
    When payment processing succeeds
    Then the confirmation page shows the order number and delivery date
```
````

### Standalone `.feature` files

Use a standalone file when a feature needs no surrounding prose, and when you
publish a feature to external tooling.

**A raw `---` YAML block doesn't parse.** Gherkin rejects it at line 1, so a
feature file can't carry frontmatter. Any metadata it needs rides in tags.

**Gherkin has only `#` line comments.** It has no block-comment delimiter
(no `/* */` and no `<!-- -->`). Both fail in different ways:

- Before the `Feature` line, `/*` is a parse error.
- **Inside the feature body, `/* */` is silently absorbed into the
  description.** It raises nothing. The delimiters and the text between them
  become part of the requirement's prose.

So a multi-line comment is several `#` lines. Nothing else is a comment, and
the syntax that looks most like one in other languages ends up in your
specification text.

---

## Tags

Tags address every requirement and criterion. Give each `Feature` and each
criterion exactly one identifier tag, and add qualifier tags (prohibition,
manual verification, test level) as your conventions define them.

**Derive a criterion's identifier from its feature's.** `AC-CHK-003.1`
belongs to `REQ-CHK-003` without any lookup. Tags permit `/`, `.`, and `:`,
but an identifier whose segments each begin with a letter, with a trailing
number, splits unambiguously.

**Decide what an `Examples` block is before you tag it.** A
`Scenario Outline` can mean either of two things, and only you know which:

- *One criterion, several named groups of data*: tag the **scenario**. The
  blocks are presentation.
- *Several criteria sharing one shape*: tag each **`Examples` block**. Each
  block is its own criterion.

Getting the second case wrong is the most dangerous mistake in the notation.
If you tag the scenario when you meant several criteria, every criterion but
the first ceases to exist. The document still parses cleanly, every scenario
is tagged, and no parser complains. A checker can detect only *partial*
tagging (some blocks tagged, others not). The rest is your judgment.

When the blocks are separate criteria, tag each one:

```text
  Scenario Outline: Latency budgets against a 1,000-document corpus
    Given a corpus of 1,000 documents
    When the tool performs "<operation>"
    Then the tool completes within "<budget>" at the 95th percentile

    @AC-PERF-009.1
    Examples: an agent asking one question
      | operation      | budget          |
      | a corpus query | 50 milliseconds |

    @AC-PERF-009.2
    Examples: a write-time hook checking one file
      | operation               | budget           |
      | a single-document check | 100 milliseconds |
```

---

## Descriptions

Every `Feature`, `Rule`, and `Scenario` has one free-text description slot.
Use each one deliberately.

**`Feature` description**: three parts, in order. The first is required.
Include the others when they have something to say.

1. **The user story**: `As` / `I want` / `So that`. Keep the role specific.
   "As an adopting team" earns its line, and "As a user" doesn't.
2. **A design note**: one or two sentences stating the constraint that
   governs this capability and that no single criterion states.
3. **A scope fence**: only where the feature has a tempting adjacent behavior
   it must *not* have.

```text
Feature: Corpus Discovery and Membership

  As an adopting team
  I want the tool to work out for itself which files are specifications
  So that I need not enumerate documents or restructure my directories

  Membership is determined by content, not location: a `.md` file whose
  frontmatter parses and declares a type the tool knows.

  This requirement reads frontmatter and path only. Resolving references
  and walking body structure are follow-on features.
```

**`Rule` name**: carries rationale that states an assertion governing several
criteria. If you can phrase rationale as a rule, it belongs here, where the
parser sees it and the criteria it governs sit beneath it.

**A `Rule:` claims every scenario after it, whatever the indentation.**
Gherkin's indentation is decorative. Once a rule opens, nothing closes it, so
any scenario the rules don't govern must be written *before* the first
`Rule:` line. Dedenting a scenario back to feature level looks right and
changes nothing.

```text
  Rule: A build failure is the checker's job, not the indexer's

    @AC-IDX-002.4
    Scenario: An unparseable document does not abort the index
```

**`Scenario` description**: leave it empty. Prose under a scenario reads as
a criterion that someone forgot to write as one, so a scenario has nowhere
for rationale to sit. Per-criterion rationale goes in the `Feature`
description, in a `Rule:` name, or in Markdown beside the block.

Descriptions hold **multiple paragraphs** (blank lines are preserved) and
pass Markdown through as literal text: emphasis, `code`, bullets, and
blockquotes. Use Markdown sparingly, because consumers receive raw
characters, not rendered output. Backticks don't make a code span here,
because Gherkin has none.

---

## Writing good scenarios

**State constraints, not justification.** "Membership is by content, not
path" tells a reader what they may not change. "This was added because
reviewers kept missing it" is history, and history is what makes prose rot.

**A step is one line, and may be as long as it needs to be.** Gherkin has no
continuation syntax and no length limit: it accepts a 412-character step.
Exempt Gherkin from any line-width rule. **Never shorten a step to fit a
margin.** Compressing "no runtime installed *beyond the CLI itself*" into "no
runtime installed" inverts the criterion, and no reviewer diffing the file
will see it. Split a step into more steps where that reads better.
Otherwise, write the long line.

**Prefer `Scenario Outline` where a criterion varies over data.** Three prose
clauses about `.md`, `.json`, and `.rst` become one legible table. Below two
instances, an outline is ceremony: a three-line `Examples` block that holds
one row costs more than the sentence it replaced.

**Examples illustrate; steps require.** IF an example and the criterion
conflict, THEN the criterion governs and the example is corrected.

**An invariant is a `Then`-only scenario.** Criteria with no trigger are legal
and common. A four-clause invariant reads worse as a `When`/`Then`/`And`/`And`
pile than as a sentence. Consider whether it's really several criteria.

**Don't restate the criteria in prose.** If a description says what the
scenarios below already say, delete it.

### Review list

A checker can verify structure and identifiers. Review these by reading,
because no parser sees intent:

- [ ] Each `Examples` block meant as its own criterion carries its own tag.
- [ ] Each criterion carries the qualifier tags that apply to it.
- [ ] No step was shortened to fit a margin.
- [ ] None of the mistakes under Seven ways to break a document silently
  appears.

---

## Seven ways to break a document silently

Each of these mistakes parses, and none raises a Gherkin error.

| Mistake | What happens |
|---|---|
| Tagging a `Scenario Outline` at scenario level when its `Examples` blocks were meant as separate criteria | Every criterion but the first **disappears**. Only partial tagging is machine-detectable |
| `#` used as a Markdown heading in a description | The line is a Gherkin comment: it **vanishes** from the description |
| A description line beginning with a Gherkin keyword | Terminates the description and **becomes structure**: a sentence starting "Scenario:" creates a second, untagged scenario |
| Free prose above the first `Scenario` | Silently absorbed into the `Feature` description. A criterion written this way isn't a criterion |
| A scenario with no steps | Parses as an empty scenario |
| A duplicate identifier tag on two scenarios | Both accepted. The duplicate is invisible to the parser |
| `/* */` used as a comment inside a feature body | **Absorbed into the description**: the delimiters and their contents become requirement prose |

The parser validates Gherkin's grammar and nothing else. It has no concept of
an acceptance criterion, so catching these mistakes is the job of a checker
that does.

---

## References

- [Gherkin reference](https://cucumber.io/docs/gherkin/reference/)
