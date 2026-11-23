---
name: tech-writer
description: Drafts, revises, and audits technical documentation in the project's style guide (by default, the Google developer documentation style guide), with its own extensions and a reader profile per doc type. Use for READMEs, release notes and changelogs, docs-site pages, help-center content, UI copy, specs written for human review, and skills and other docs written only for agents.
---

# Tech writer

Use the Google developer documentation style guide
(developers.google.com/style), unless the project names another style guide,
for example in a line of `AGENTS.md` or `CLAUDE.md`. Then use that one. This
file holds only what the style guide lacks. Don't apply it to marketing,
legal, academic, fictional, or personal writing unless the user asks.

**Precedence:** the user's explicit request and the destination's
requirements win. Then source facts, quotations, and technical tokens, which
no rule may change. Then this file, then the style guide.

Apply the style guide and this file in full. From the existing document, keep
only its terminology. Don't copy its habits, such as heading case, dashes,
contractions, comma use, emphasis, spelling, or voice. A line is finished when
it follows every rule here, not when it matches the rest of the document.
Stop when every line is finished.

## Accuracy

Every code example and command must run as written. Every API, flag, path,
and default must match the source. When unsure, read the source again or run
the command. Don't guess.

When you find a content problem, such as a contradiction, a stale claim, or a
path that doesn't exist, verify it against the source and fix it with the
smallest accurate change. Report it as a finding, with its `file:line` and a
quote, only when you can't confirm what is true or the fix would change a
decision. Never leave a marker in the document.

## Terminology

- Use one term for each concept, and one meaning for each term. Before you
  merge two names, check that they name one concept: an Xcode workspace
  contains projects.
- Choose unambiguous words. When a word has a common meaning other than the
  one you intend ("render," "member," "path," "extent"), qualify it ("heading
  path") or use another word.
- Use a project term only as the project's glossary defines it, if the
  project names one in `AGENTS.md` or `CLAUDE.md`. For a term the reader
  can't decode and the glossary doesn't define, use plain words first. If
  you need a term, qualify a common word by its domain ("link resolution,"
  "Gherkin qualifier tag") rather than give an ordinary word a private
  meaning ("settle," "drawing"). Define a term that only one document uses
  in that document. Propose a glossary entry only for a term that two or
  more documents use, and add it only when the user approves it.
- Name the kind of each code identifier on first mention in a section: "the
  `Build` function," "the `Span` type," "the `cli/core/corpus` package."
- Keep noun clusters to three words or fewer. Keep articles; don't drop them
  for brevity.
- Never trade an exact term for a casual one to make the tone conversational.

The spec skill's document types are proper names. Write them in title case
wherever they name the document type, including in sentence-case headings
and plurals: ADR, FRD, Standards, System Blueprint, Container Blueprint,
Component Blueprint, Module Blueprint, Feature Blueprint, Contract, and Work
Order. "A blueprint for the rollout" uses the ordinary word.

## Revising

When you revise a file in place, content is fixed and wording is not. Never
add or remove a fact, requirement, decision, example, or caveat, except as
this section allows.

Leave these untouched:

- Code blocks, inline code, commands, paths, config keys, and link targets
- API, function, flag, and setting names
- Error messages and log output quoted verbatim
- Version numbers, dates, and numeric limits
- Front matter, except a `description` (see below)
- Gherkin blocks
- RFC 2119 keywords (MUST, SHOULD, MAY, SHALL, SHALL NOT) and their strength
- Wikilinks (`[[slug]]`, `[[slug#heading]]`) and identifiers such as AC and
  FEAT ids
- Element record blocks
- Examples marked as wrong, such as an anti-pattern or a ❌ line
- Labels and separators a template requires, such as **Status:**,
  **Rationale:**, a `---` rule, or a ` — ` between a name and its text.
  They don't count against the dash guidance. Revise the text that follows
  them.
- In an existing document, a heading's words and punctuation, because other
  documents link to it. Change its case when links match headings without
  regard to case, as they do in the spec corpus, unless it names an element
  or repeats a Gherkin `Feature:` name. Reword it only when you update every
  link to it in the same pass.

**Requirement statements** (EARS, "the system shall") may take clarity edits:
define a term, name an identifier's kind, fix grammar, or split a statement
when each part repeats the condition verbatim. Never change the modal, the
actor, the trigger, the condition, the scope, or the EARS pattern. In a
blueprint, also correct a statement the source contradicts, and cite the
evidence (a file and symbol, or a test) in your report. In an FRD, Standards,
or ADR, report a contradiction as a finding.

**A front matter `description`** may take the same clarity edits. Keep its
form, such as "<what it owns>. Change it to <kinds of change>."

**An unlinked mention** of another document: add the link when `specs check`
or the project's link checker resolves it. Otherwise report it.

**Cuts.** Cut history (see Stable prose) and restate in the present tense any
fact it holds that is still true. In a doc written only for agents, cut what
its reader profile says the agent doesn't need, never an instruction,
constraint, or fact it needs to act correctly. No spec corpus document counts
as written only for agents. List every cut.

HTML comments are not exempt: in templates they instruct the agent. Revise
their prose and keep every name, path, or token they cite.

## Stable prose

Write prose that stays true when what it describes changes.

- Write about how things are, not how they were or will be. Leave out how a
  rule, design, or document came about. Only changelogs, release notes, and
  the decision statement an ADR's template requires may use past or future
  tense for this.
- Don't count what a list, table, directory, or codebase holds ("the eleven
  checks"). Write "each check" or "the checks below." Keep a count the design
  fixes, such as a protocol's two phases. Changelogs and release notes may
  count.
- Cite a rule, step, or section by its heading or a link, never by its number
  or position ("rule 14," "the table above").
- Don't restate what has a canonical source, such as a list kept elsewhere, a
  default set in code, or a version in a manifest. Link the source.
- Don't describe a mechanism that doesn't exist yet, such as a planned check.
- Leave out volatile measurements, such as line numbers, file sizes, and
  warning counts. Where one matters, give the command that measures it.

When you revise, a count or position you can't replace without changing a
fact becomes a finding.

## Capitals

- RFC 2119 keywords keep their capitals. Don't use capitals for emphasis.
- Capitalized EARS keywords (WHEN, WHILE, IF, THEN, WHERE) belong only in
  requirement statements and in prose that names the keyword. Elsewhere,
  write them as ordinary words.
- When prose names a heading, write it in the heading's case.

## Reader and tone

Before you write, profile the reader in your working notes, not in the
document. For a spec document type, start from the profile in the header
comment of the spec skill's template for that type. For anything else,
including a skill or another doc written only for agents, start from
`references/reader-profiles.md`. Adjust it for the project. With several
readers, profile the one whose needs should set the tone. If you can't tell,
write down your assumption and proceed.

The profile sets the tone in every task. When you draft, it also sets the
content: add what the reader needs and lacks, cut what they already know, and
put the answer to "Wants" first, unless a template fixes the order. When you
revise, a profile line that would add, cut, or move content becomes a
finding.

In every document, be outcome-minded (tie each fact to what the reader sees,
waits for, or must change), candid (state bugs, unsupported cases, and sharp
edges where the reader meets them), and plain (make the hard thing look
simple). Use no humor in errors, troubleshooting, security, or breaking
changes. When unsure, choose clear over friendly, and friendly over clever.

## Punctuation

Choose the mark by what the words after it do. Stop at the first yes:

1. The second part could stand alone, linked only by "and also": a period.
2. It delivers what the first part promised (a list, an example, a definition,
   a cause, the answer), so "namely" would fit: a colon.
3. It's an aside the sentence works without: commas if it reads as part of
   the flow, parentheses if the reader can skip it (the default in technical
   docs), and a pair of em dashes only if it carries weight or contains
   commas.
4. The sentence turns sharply or sums up a list just before it: a single em
   dash, rarely.

Use at most one pair of em dashes per sentence, and rarely more than one dash
sentence per section.

## Patterns to check

- **Not X but Y.** Keep the negative half only when a reader might believe
  it or it states a real limit ("This cache is not shared across workers").
- **Forced triads.** Lists of three that exist for rhythm. Real enumerations
  are exempt.
- **Repeated sentence openings.** Numbered steps and reference entries are
  exempt; steps start with an imperative verb.
- **Stacked qualifiers.** Replace a hedge with its condition ("fails when the
  disk is full"). Change a modal word only when the source shows what's true.
- **Writing about the document.** Cut "This section covers…" and start with
  the content. Keep a sentence that introduces a table or list.
- **T1. Terminology drift.** One concept under several names (see
  Terminology).
- **T2. Vague capability claims.** "Supports" or "handles" with no specifics.
  State what happens, the defaults, and the limits.
- **T3. Missing specifics.** No unit, default, version, or exact command. If
  you don't have it, leave the text rather than guess.
- **T4. Nominalizations.** "Perform the installation of" becomes "install."
- **T5. Long sentences and walls of text.** Over about 25 words, or
  paragraphs over about five sentences. Split them. Citation lists (runs of
  AC ids, wikilinks, or ADR numbers) are exempt.
- **T6. Release notes in the wrong order.** Breaking changes and required
  actions first, then user-visible changes, then fixes.
- **T7. Invented examples.** Never invent a flag, output, default, or value.
- **Asides of your own.** Don't add opinions or reactions. Keep the author's,
  reworded.

## Line wrapping

Follow the destination's line width (in the spec corpus, `line_width` in
`specs.toml`), or 80 characters when it sets none. Links, wikilinks that
contain spaces, tables, headings, and code blocks may run past it. Rewrap
each paragraph you change, don't reflow the rest, and never reword a sentence
to fit the margin.

## Cross-links

When you draft a document that names another document, link it: a `[[slug]]`
wikilink for the spec corpus, otherwise a relative Markdown link or a
backticked path, following the project's citation conventions.

## How to run it

1. Identify the document type and the task (draft, revise, or audit), profile
   the reader, and mark what must not change (see Revising).
2. Draft or revise to the style guide and this file. When a question turns
   on a specific rule (word choice, numbers, dates, units, capitalization),
   look it up in the style guide when browsing is available.
3. Review. Run Patterns to check over the result, then check Terminology,
   Capitals, the profile's tone, and every item in Revising against the
   original or the diff. An agent that reviews its own draft tends to approve
   it, so a separate agent runs this step and sees only the draft, the
   original or the diff, the reader profile, and this file. If you can't
   start one, for example because you are a subagent, return the draft and
   say that the review is still owed; the agent that dispatched you runs it.
4. Return the result.
   - For an audit, change nothing. Report findings in priority order with
     short examples.
   - For pasted text, return the revised text.
   - For a file, write it and report the sentences you changed, any exempt
     text you touched with the rule that allowed it, your findings, your
     cuts, and your judgment calls.
5. If the project uses Vale, run it last and fix what it flags.
