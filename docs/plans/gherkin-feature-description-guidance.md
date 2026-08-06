# Writing Gherkin `Feature` descriptions in an FRD

**Status: proposed.** This is drafted as a section for
`plugin/skills/spec/references/frd-template.md`, to be folded in when
[[adr-018]] is approved. It is kept here rather than in the template so
that the `spec` skill does not start instructing agents to author
Gherkin while the decision is still draft.

Under [[adr-018]] a requirement maps to a `Feature`, and the `Feature`
description is the one free-text slot attached to it. This is what to
put there.

---

## What the description is for

Three things, in this order. The first is required; the others are
included when they have something to say.

**1. The user story** — three lines, `As` / `I want` / `So that`. Keep
the role specific: "As an adopting team" earns its line, "As a user"
does not.

**2. A design note** — one or two sentences stating the constraint that
governs this capability and that no single criterion states. This is
where rationale belongs when it is context rather than a rule.

**3. A scope fence** — only where the requirement has a tempting
adjacent behaviour it must *not* have.

Worked example:

```gherkin
@frd-spec-validation-cli
Feature: Corpus Discovery and Membership

  As an adopting team
  I want the tool to work out for itself which files are specifications
  So that I need not enumerate documents or restructure my directories

  Membership is determined by content, not location: a `.md` file whose
  frontmatter parses and declares a type the active profile knows. A file
  that looks like a specification but has no frontmatter is recorded as an
  exclusion rather than dropped, so an agent can notice what the CLI
  cannot judge.

  This requirement reads frontmatter and path only. Resolving references
  and walking body structure are follow-on features.

  @AC-SPEC-001.1
  Scenario Outline: A directory invocation admits every typed markdown file
```

## Rules

**State constraints, not justification.** "Membership is by content, not
path" is a constraint — it tells a reader what they may not change.
"This was added because reviewers kept missing it" is history, and
history is what makes prose rot. If a sentence would still be true after
the feature ships and would still change someone's decision, keep it.

**If it is a rule, make it a `Rule:` instead.** Rationale that states an
assertion governing several criteria belongs in a `Rule:` name, where the
parser can see it and the criteria it governs sit beneath it:

```gherkin
  Rule: A build failure is the checker's job, not the indexer's

    @AC-SPEC-002.4
    Scenario: An unparseable document does not abort the index
```

A `Rule` names an assertion that can be true or false — "there can be
only one". It is not a topic. `Rule: Configuration` is wrong; that is a
`Feature`.

**Never begin a line with a Gherkin keyword.** A description line
starting `Given`, `When`, `Then`, `Scenario`, `Scenario Outline`,
`Rule`, `Examples`, `Background` or `Feature` terminates the description
and becomes structure. Measured: a sentence starting "Scenario:" created
a second, untagged scenario in a document that otherwise parsed cleanly.

**Never use `#`.** In Gherkin `#` opens a comment, not a markdown
heading. Measured: `# Why this exists` was stripped out of the
description entirely and moved to the document's flat comment list. It
does not error — it silently disappears.

**Do not restate the criteria.** If the description says what the
scenarios below already say, delete it. Length is not the measure;
several of the strongest descriptions in the worked variant are two
paragraphs.

## What the format supports

All of the following was verified against
`github.com/cucumber/gherkin/go/v42`:

| | Supported? |
|---|---|
| Multiple paragraphs, separated by blank lines | **Yes** — blank lines are preserved verbatim in `Feature.Description` |
| Markdown emphasis, `code`, `[[slug]]` links | **Yes** — passed through as literal text, uninterpreted |
| Markdown bullet (`-`) and numbered lists | **Yes** |
| Blockquotes, table pipes as text | **Yes** |
| Markdown headings (`#`) | **No** — silently becomes a Gherkin comment |
| Lines wrapping freely | **Yes** — unlike steps, description lines wrap normally |

Because the description is captured as raw text, use markdown sparingly.
A consumer reading the AST gets the characters you typed, not rendered
output.

## Where it ends up

`Feature.Description` is a single string on the AST node, so a structured
store receives one text field per requirement, explicitly scoped to that
requirement rather than inferred from position in a markdown file.

**It is not carried on pickles.** `gherkin.Pickles()` compiles to one row
per example and those rows have no description field. Load descriptions
from the AST walk; the pickle stream will not have them.

## References

- [[adr-018]] — constraint 1, the requirement-to-`Feature` mapping
- `docs/spec-index-frd-gherkin.md` — nine worked descriptions
- `docs/gherkin-frd-comparison.md` — what the trial showed
- `docs/plans/gherkin-extraction-findings.md` — the extraction contract
