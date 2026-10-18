---
title: Gherkin Authoring Guide
description: How to write features and acceptance criteria in Gherkin under
  this framework — the document mapping, the tag vocabulary, what belongs in
  each description slot, and the constructs that fail silently
slug: gherkin-guide
type: standards
status: active
---

# Gherkin Authoring Guide

Features and acceptance criteria are written in Gherkin ([[adr-021]]).
EARS governs every other normative statement — standards documents,
contracts, blueprints ([[adr-002]]).

This guide is what to write. Read it before authoring a feature.

## Contents

- The mapping — which FRD concept becomes which Gherkin keyword
- Two admissible forms — fenced blocks and standalone files
- Tags — the required vocabulary
- Descriptions — what belongs in each slot
- Writing good scenarios
- Seven ways to break a document silently
- References

---

## The mapping

| Concept | Gherkin |
|---|---|
| The document | its own frontmatter — a fenced block inherits it |
| A feature's identity | a `@<PREFIX>-<CODE>-<nnn>` tag on the `Feature` |
| A requirement / feature | `Feature:` |
| Its user story | the `Feature` description |
| An acceptance criterion | `Scenario` or `Scenario Outline` |
| A business rule grouping several criteria | `Rule:` |
| A criterion's concrete example | an `Examples` row |
| Rationale | a design note in the description, or a `Rule:` name |

**A requirement is a `Feature`, not a `Rule`.** A `Rule` names a
business rule — an assertion that can be true or false. "There can be
only one" is a rule. "Configuration" is a topic, and a topic is a
`Feature`. Getting this backwards is the most common structural error,
and it cascades: with the requirement on `Rule:`, the only description
slot is spent and the user story has nowhere to go.

**A Gherkin document holds exactly one `Feature`.** Two `Feature:` lines
in one document do not parse. So one requirement means one fenced block,
or one `.feature` file.

---

## Two admissible forms

### Fenced blocks inside an FRD — the default

Use where the feature benefits from surrounding markdown: an overview,
terminology, references, document frontmatter.

````markdown
### FEAT-CHK-003: Payment Confirmation

**Acceptance Criteria:**

```gherkin
@FEAT-CHK-003
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

Use where a feature needs no surrounding prose, and for publication to
external tooling.

A feature file carries **no frontmatter**. It is a feature source, not a
document: its owning document is the one its `@frd-<slug>` tag names, and
its metadata is that document's. This is the one place that tag is
written — a fenced block sits inside its owning document already.

Shown here as plain text, not as a `gherkin` fence: written as one, this
snippet would itself be extracted as a fenced-block feature of this
guide, and its illustrative `@frd-checkout` tag would then contradict
the document it actually sits in.

```text
@frd-checkout @FEAT-CHK-003
Feature: Payment Confirmation
```

**A raw `---` YAML block does not parse.** Gherkin rejects it at line 1,
so there is no frontmatter to write even if a file wanted one.

**Gherkin has only `#` line comments.** There is no block-comment
delimiter — no `/* */`, no `<!-- -->`. Both were measured, and they fail
in different ways:

- Before the `Feature` line, `/*` is a parse error.
- **Inside the feature body, `/* */` is silently absorbed into the
  description.** It raises nothing; the delimiters and the text between
  them become part of the requirement's prose.

So a multi-line comment is several `#` lines. Nothing else is a comment,
and the thing that looks most like one in other languages ends up in
your specification text.

---

## Tags

Tags are how every record is addressed. They are required, and a missing
one is reported as an **error**, not a warning — findings carry a
severity, and this one fails a run.

| Tag | Where | Required |
|---|---|---|
| `@<PREFIX>-<CODE>-<nnn>` | on the `Feature` | yes — the feature's identifier |
| `@frd-<slug>` | on the `Feature`, in a `.feature` file only | yes there — names the owning document |
| `@AC-<CODE>-<nnn>.<n>` | on each `Scenario` | yes — one per criterion, deriving its code and number from the feature |
| `@negative` | on a prohibition | yes, where the criterion forbids |
| `@manual` | on a hand-verified criterion | yes, where it applies |
| `@level-<unit\|integration\|e2e>` | on a `Scenario` | optional |

A `Feature` in a fenced block therefore carries one tag: its identifier.
Ownership is not written down, because the block is inside the document
that owns it and the CLI reads the slug from that document's
frontmatter. A tag would restate it, and could come to disagree with it —
writing one that names a different document is an error. Only a
standalone file, having no document around it, names its owner.

A `Scenario` carries one identifier plus whatever qualifiers apply.

**Features share a code; a code and a number together identify one.**
`SPEC` covers eleven requirements of the spec-index feature. That is what
makes the code a grouping axis rather than an identity.

**Decide what an `Examples` block is before you tag it.** A
`Scenario Outline` can mean two different things, and only you know
which:

- *One criterion, several named groups of data* — tag the **scenario**.
  The blocks are presentation.
- *Several criteria sharing one shape* — tag each **`Examples` block**.
  Each block is its own criterion.

Getting this wrong in the second direction is the most dangerous
mistake in the notation: tag the scenario when you meant several
criteria, and all but the first cease to exist, in a document that
parses cleanly with every scenario tagged and no check firing. The CLI
cannot read your intent, so it reports only *partial* tagging — some
blocks tagged, others not. The rest is your judgment.

Tag each block when they are separate criteria. Shown here as plain
text, not as a `gherkin` fence — it is an excerpt with no `Feature:`
header of its own, and a `gherkin` fence would extract it as one that
carries no identifier tag:

```text
  Scenario Outline: Latency budgets against a 1,000-document corpus
    Given a corpus of 1,000 documents
    When the CLI performs "<operation>"
    Then the CLI completes within "<budget>" at the 95th percentile

    @AC-SPEC-009.1
    Examples: an agent asking one question
      | operation      | budget          |
      | a corpus query | 50 milliseconds |

    @AC-SPEC-009.2
    Examples: a write-time hook checking one file
      | operation               | budget           |
      | a single-document check | 100 milliseconds |
```

**Grouping is by code, not by a path-shaped slug.** Tags would permit a
path — `/`, `.` and `:` all parse — but a framework slug admits only
lowercase alphanumerics and hyphens, so a hyphenated slug cannot be split
back into segments unambiguously. The code carries the grouping instead,
and a criterion identifier derives from its feature's: `AC-CHK-003.1`
belongs to `FEAT-CHK-003` without anything having to be parsed.

**A code may itself be hyphen-joined segments** — `SWARM-DISP`,
`SWARM-RUN` — which is how one capability split across several documents
keeps a common first segment while each document's requirements stay
addressable on their own. Every segment begins with a letter, so the
trailing number can never be read as a segment and the split is
unambiguous in a way a slug's is not:

```text
FEAT-SWARM-DISP-001      prefix FEAT · code SWARM-DISP · number 001
AC-SWARM-DISP-001.1      the criterion, deriving both from its feature
```

Reach for one only where the shared segment is doing work. Two codes that
merely look alike are two codes, and a reader gains nothing from a common
prefix that groups nothing.

---

## Descriptions

Every `Feature`, `Rule` and `Scenario` has one free-text description
slot. Use them deliberately.

**`Feature` description** — three parts, in order. The first is
required; the others are included when they have something to say.

1. **The user story** — `As` / `I want` / `So that`. Keep the role
   specific: "As an adopting team" earns its line; "As a user" does not.
2. **A design note** — one or two sentences stating the constraint that
   governs this capability and that no single criterion states.
3. **A scope fence** — only where the feature has a tempting adjacent
   behaviour it must *not* have.

Shown here as plain text, not as a `gherkin` fence: the identifier it
illustrates is a real one, and written as a live block this excerpt would
be extracted as a second feature declaring it — a duplicate identifier
across the corpus.

```text
@FEAT-SPEC-001
Feature: Corpus Discovery and Membership

  As an adopting team
  I want the tool to work out for itself which files are specifications
  So that I need not enumerate documents or restructure my directories

  Membership is determined by content, not location: a `.md` file whose
  frontmatter parses and declares a type the active profile knows.

  This requirement reads frontmatter and path only. Resolving references
  and walking body structure are follow-on features.
```

**`Rule` name** — carries rationale that states an assertion governing
several criteria. If rationale can be phrased as a rule, it belongs
here, where the parser sees it and the criteria it governs sit beneath.

**A `Rule:` claims every scenario after it, whatever the indentation.**
Gherkin's indentation is decorative. Once a rule opens there is no way to
leave it, so any scenario the rules do not govern must be written
*before* the first `Rule:` line — dedenting one back to feature level
looks right and changes nothing.

Shown here as plain text, not as a `gherkin` fence, for the same reason
as the excerpt above — no enclosing `Feature:` header:

```text
  Rule: A build failure is the checker's job, not the indexer's

    @AC-SPEC-002.4
    Scenario: An unparseable document does not abort the index
```

**`Scenario` description** — leave it empty. *Any* non-empty scenario
description is reported as a probable untagged criterion, not only a
normative one, so there is nowhere in a scenario for rationale to sit.
Per-criterion rationale goes in the `Feature` description, in a `Rule:`
name, or in markdown beside the block.

Descriptions hold **multiple paragraphs** — blank lines are preserved —
and pass markdown through as literal text: emphasis, `code`, `[[slug]]`,
bullets, blockquotes. Use it sparingly; consumers receive raw characters,
not rendered output.

**A `[[slug]]` anywhere in a Gherkin source is a reference.** Gherkin
never sees it — it is ordinary text in a description, a scenario name, a
step, a rule name or an examples cell — but this framework records it and
resolves it exactly as it would in prose. Cite a document that does not
exist inside a scenario and the check reports it.

**So a wikilink cannot be illustrated inside a block.** Backticks are not
a code span here; there is no code span. An FRD whose subject is the
notation itself — anything about references, anchors or rewriting — must
describe a wikilink in a step ("a wikilink naming the slug X") and keep
the literal syntax in backticked markdown beside the block. Written
literally, an illustration either enters the index as a citation the
document never made, or fails the run as an unresolved reference.

---

## Writing good scenarios

**State constraints, not justification.** "Membership is by content, not
path" tells a reader what they may not change. "This was added because
reviewers kept missing it" is history, and history is what makes prose
rot.

**A step is one line, and may be as long as it needs to be.** There is
no continuation syntax. There is also no length limit — Gherkin accepts
a 412-character step, and this framework exempts fenced blocks from
`line_width` (AC-SPECSTR-002.5). **Never shorten a step to fit a
margin**: compressing "no runtime installed *beyond the CLI itself*"
into "no runtime installed" inverts the criterion, and no reviewer
diffing the file will see it. Decompose into more steps where that
reads better; otherwise write the long line.

**Prefer `Scenario Outline` where a criterion varies over data.** Three
prose clauses about `.md`, `.json` and `.rst` become one legible table.
Below two instances it is ceremony — a three-line `Examples` block
holding one row costs more than the sentence it replaced.

**An invariant is a `Then`-only scenario.** Criteria with no trigger are
legal and common. A four-clause invariant reads worse as a
`When`/`Then`/`And`/`And` pile than as a sentence; consider whether it
is really several criteria.

**Do not restate the criteria in prose.** If a description says what the
scenarios below already say, delete it.

---

## Seven ways to break a document silently

Each of these parses. None raises a Gherkin error. Together they are
most of what the CLI's semantic checks exist to catch.

| Mistake | What happens |
|---|---|
| Tagging a `Scenario Outline` at scenario level when its `Examples` blocks were meant as separate criteria | Every criterion but the first **disappears**; only partial tagging is machine-detectable |
| `#` used as a markdown heading in a description | The line is a Gherkin comment: it **vanishes** from the description |
| A description line beginning with a Gherkin keyword | Terminates the description and **becomes structure** — a sentence starting "Scenario:" creates a second, untagged scenario |
| Free prose above the first `Scenario` | Silently absorbed into the `Feature` description; a criterion written this way is not a criterion |
| A scenario with no steps | Parses as an empty scenario |
| A duplicate `@AC-*` tag on two scenarios | Both accepted; the duplicate is invisible to the parser |
| `/* */` used as a comment inside a feature body | **Absorbed into the description** — the delimiters and their contents become requirement prose |

The parser validates Gherkin's grammar and nothing else. It has no
concept of an acceptance criterion — every check above is the spec
CLI's, and every one of these mistakes is why they exist.

---

## References

- [[adr-021]] — the decision, and the binding constraints this guide
  implements
- [[adr-002]] — EARS, which governs every normative statement that is
  not a feature or an acceptance criterion
- `docs/research/gherkin/extraction-findings.md` — the measurements behind
  the silent-failure table
- [[frd-gherkin-adoption]] — a full FRD written this way
- [Gherkin reference](https://cucumber.io/docs/gherkin/reference/)
