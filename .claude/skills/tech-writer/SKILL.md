---
name: tech-writer
description: Drafts, revises, and audits technical documentation in the Google developer documentation style guide, with this project's extensions. Use for READMEs, release notes and changelogs, docs-site pages, help-center content, UI copy, and specs written for human review.
---

# Tech writer

Use the Google developer documentation style guide
(developers.google.com/style). This file covers only what extends it: what a
rewrite must leave alone, project terms, rules that settle choices Google
leaves open, and patterns to check.

**Precedence:** the user's explicit request and the destination's requirements
win. Then source facts, quotations, and technical tokens, which no rule may
change. Then this file, then Google's guide.

**This skill's style is the style.** Apply Google's guide and this file in
full. The only project input you keep is terminology: the names the project
uses for its concepts. Don't copy the existing document's habits, such as
heading case, dashes, contractions, comma use, emphasis, spelling, or voice.

A line is finished only when it follows every rule here. Matching the rest of
the document doesn't count.

## Scope and defaults

- Applies to READMEs, release notes and changelogs, docs-site pages, help-center content, UI copy, and specs written for human review.
- Do not apply it to marketing, legal, academic, fictional, or personal writing unless the user asks for this style.
- When you revise a file in place, change prose only. Leave code blocks, inline code, commands, paths, config keys, link targets, and front matter untouched.
- Also leave these untouched:
  - API, function, flag, and setting names
  - Error messages and log output quoted verbatim
  - Version numbers, dates, and numeric limits
  - Requirement statements in specs (EARS, Gherkin, "the system shall"), including whole Gherkin blocks. Their punctuation is part of their wording, dashes included.
  - RFC 2119 keywords (MUST, SHOULD, MAY) and their exact meaning
  - Wikilinks (`[[slug]]`, `[[slug#heading]]`) and identifiers such as AC and FEAT ids
  - Element record blocks
  - In an existing document, the words and punctuation of a heading, because other documents link to it. Change a heading's case freely when links match headings without regard to case, as they do in the specs corpus. Punctuation is part of the anchor, so a dash in a heading stays. Reword a heading only when you update every link to it in the same pass. Headings in a new document follow every rule here.
  - Labels and separators the document's template requires, such as **Status:**, **Rationale:**, or a `---` rule between requirements. They are structure, not decoration. Apply the rules to the text that follows them.
- Citation lists are exempt from T5's sentence length: runs of AC ids, wikilinks, ADR numbers, or references.
- HTML comments are not exempt. In templates they are instructions to the agent, so apply this skill to their prose. Keep any name, path, or token they cite.
- Content is fixed and wording is not. Never add or remove a fact, requirement, decision, example, or caveat, but apply every style rule to how each one is worded.
- Don't add opinions, reactions, or asides of your own. Keep the ones the author wrote, reworded to these rules.

## Rules Google leaves open

- RFC 2119 keywords keep their capitals, although Google rules out capitals for emphasis.
- Capitalized EARS keywords (WHEN, WHILE, IF, WHERE) belong only in requirement statements. In other prose, write them as ordinary words.
- When prose names a heading, write it in the heading's case.
- Never trade an exact term for a casual one to make the tone conversational.

## Project terms

The spec skill's document types are proper names. Write them in title case
wherever they appear, including in sentence-case headings and in plurals:
ADR, FRD, Standards, System Blueprint, Container Blueprint, Component
Blueprint, Module Blueprint, Feature Blueprint, Contract, and Work Order.
Write "a Component Blueprint," not "a component blueprint."

This applies only when the word names the document type. "The API contract
between two services" and "a blueprint for the rollout" use ordinary words.

## Voice and tone

Google's guide sets the base tone. On top of it, keep one voice in every
document, and shift the tone to fit the reader.

Write like an experienced colleague who remembers being new:

- **Outcome-minded.** Tie each technical fact to what the reader sees, waits for, can now do, or must change.
- **Candid.** State known bugs, unsupported cases, and sharp edges where the reader will meet them. Don't bury them, and don't oversell.
- **A translator.** Expertise shows in how simple you make a hard thing look, not in how technical you sound.
- **Respectful of the reader's judgment.** Readers know their own setup. Give them what they need to decide, and don't decide for them.
- **No humor** in errors, troubleshooting, security, or breaking changes. Elsewhere, default to none.

Don't sound like a salesperson (benefits without mechanisms, claims nobody can
check), a consultant (frameworks and padding before the answer), a press
release ("We're excited to…"), or a lecturer (background the reader must sit
through before they can act).

A release note, off-voice and on-voice:

> This release introduces significant enhancements to the caching layer,
> providing a more robust and seamless experience across a variety of
> deployment scenarios.

> Cached pages expire after 10 minutes instead of 24 hours, so edits show up
> sooner. If you set `CACHE_TTL` yourself, nothing changes for you.

### Profile the reader

Before you write, profile the reader in your working notes, not in the
document:

```
Reader:  who, and how they arrive
Knows:   what they already know, and what they don't
Wants:   the one thing they came to do or decide
Budget:  how long they'll give this page
Tone:    how the voice adjusts for them
```

Start from the default for the doc type in `references/reader-profiles.md`
and adjust it for the project. When a document has several readers, profile
the one whose needs should set the tone. If you can't tell, write down your
assumption and proceed.

The profile sets the tone in every task. When you draft, it also sets the
content: add what the reader needs and lacks (a prerequisite, a term's
meaning, a default), cut what they already know, put the answer to "Wants"
first, and order the rest by the reader's next step unless a template fixes
the order. When you revise, content is fixed (see Scope and defaults), so the
profile changes only how things are worded.

Vary sentence length in narrative docs. Keep steps and reference entries
parallel. When unsure, choose clear over friendly, and friendly over clever.

## Punctuation

Choose the mark by what the words after it do. Ask these questions in order
and stop at the first yes.

1. **Could the second part stand as its own sentence, and is the link between
   the two only "and also"?** Use a period. Two short sentences beat one joined
   by a semicolon or a dash.
2. **Does the second part deliver what the first part promised: a list, an
   example, a definition, a cause, or the answer?** Use a colon. A test: you
   could write "namely" or "that is" in its place ("The check fails for one
   reason: the anchor doesn't exist."). Template separators such as `[[slug]]: text` and
   `**Term**: definition` are colons too.
3. **Is it an aside: does the sentence still work with the phrase removed?**
   Pick the mark by how much the aside matters.
   - **Commas** for an aside that reads as part of the sentence's flow ("The
     index, which is rebuilt on every run, holds no state.").
   - **Parentheses** for something the reader can skip: a unit, an
     abbreviation, a cross-reference, a minor clarification ("Lines wrap at 80
     columns (set in `specs.toml`)."). This is the default for asides in
     technical docs, and parentheses wrap code and links cleanly.
   - **A pair of em dashes** for an aside that carries weight and should stand
     out, or one that already contains commas, where commas around it would be
     ambiguous ("Every tool—the CLI, the plugin, and CI—reads the same
     config.").
4. **Does the sentence turn sharply, or sum up a list just before it?** A
   single em dash ("Exit codes, messages, and anchors—all three changed.").
   Use this rarely. A colon or a new sentence usually reads better.

Whatever the answer:

- At most one pair of em dashes per sentence, and rarely more than one dash
  sentence per section. If you reach for a second, use parentheses or split
  the sentence. Several dashes in one section is a sign that most of them
  belong in parentheses, commas, or a sentence of their own.
- Requirement statements, headings, code, and link targets keep their
  punctuation (see Scope and defaults).

## Patterns to check

**Not X but Y.** A sentence that says what something isn't before saying what
it is. Keep the negative half when a reader might actually believe it, or when
it states a real limit: "This cache is not shared across workers," "The
`--force` flag doesn't skip validation." Cut it when no reader would have
believed it, because then it only builds up the second half: "It's not just a
linter, it's a full validator" becomes "It validates the whole corpus."

**Forced triads.** Lists of three that exist for rhythm. Exempt real
enumerations: three supported platforms, three config options, three steps.
The test is whether the list is complete and accurate, not whether it has
three items.

**Repeated sentence openings.** Exempt numbered steps and reference entries.
Steps should start with an imperative verb ("Run," "Add," "Open"), and
reference entries should share a structure. Parallel structure helps scanning
here.

**Stacked qualifiers.** In docs, replace hedges with conditions. "This may
fail in some cases" becomes "This fails when the disk is full." Change a
modal word (*can*, *might*, *should*, *will*) only when the source shows what
is actually true. If you don't know the condition, keep the hedge.

**Writing about the document.** Openers that describe the text instead of
saying something: "This README describes..." or "This section covers...."
Cut them and start with the content. A heading already says what a section
covers. A sentence that introduces a table or list is different, and Google
asks for one.

## New patterns for technical docs

**T1. Terminology drift.** The same concept appears under different names ("config file," "settings file," "configuration"). Pick one term per concept and use it everywhere. Never vary a technical term for style, even if a rewrite elsewhere encourages variety. Check that the names really are one concept first: some pairs that look like synonyms aren't (an Xcode workspace contains projects).

**T2. Vague capability claims.** "Supports," "handles," and "allows for" with no specifics. State what happens, the defaults, and the limits. "Supports large files" becomes "Uploads files up to 5 GB."

**T3. Missing specifics.** No units, no default, no version, or "run the appropriate command." Give the exact value or command. If you don't have it, leave the text as it is rather than guessing.

**T4. Nominalizations.** Verbs turned into nouns: "perform the installation of," "make a determination," "provide configuration of." This is the main source of a legalese feel. Use the verb: "install," "decide," "configure."

**T5. Long sentences and walls of text.** Sentences over about 25 words, or paragraphs over about five sentences. Split them. One idea per sentence. Citation lists don't count toward the length.

**T6. Release notes in the wrong order.** Features listed in commit order, with breaking changes buried. Put breaking changes and required actions first, then user-visible changes, then fixes. Each item starts with what the user sees or must do.

**T7. Invented examples.** A rewrite must never invent a flag, an output, a default, or an example value. If an example would help and you don't have a real one, leave it out.

## Line wrapping

Follow the destination's line width. Google's Markdown guide keeps an
80-character limit, with exceptions for links, tables, headings, and code
blocks. Rewrap each paragraph you change to that width, and don't reflow
paragraphs you didn't change. Never reword a sentence to make it fit the
margin.

## How to run it

1. Identify the doc type and the task (draft, revise, or audit), and profile the reader (see Voice and tone). Mark what must not change, using Scope and defaults above.
2. **Draft or revise** to Google's guide and this file. When a question turns on a specific Google rule (word choice, a punctuation mark, numbers, dates, units, product names, capitalization), look it up on developers.google.com/style when browsing is available.
3. **Check the patterns.** Run "Patterns to check" and the T-patterns over the result. Prefer a fresh-context reviewer: run this pass in a subagent that sees only the draft, the reader profile, and these rules. A reviewer that wrote the draft tends to approve it.
4. **Validate.** Check the final text against Google's guide, "Rules Google leaves open," "Project terms," and the reader profile's tone, and check that nothing in Scope and defaults changed.
5. **Return the result.**
   - For an audit, change nothing. Report findings in priority order with short examples.
   - For pasted text, return the revised text.
   - For a file, write the revised file and summarize what changed.
6. If the project uses Vale, run it last and fix what it flags.

Stop when the doc is clear, accurate, and consistent, and every line follows the rules above. Don't keep polishing lines that already do.
