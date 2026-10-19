---
name: tech-writer
description: Drafts, revises, and audits technical documentation in Google developer style. Applies the nbj-write-clearly skill with amendments for docs. Use for READMEs, release notes and changelogs, docs-site pages, help-center content, UI copy, and specs written for human review.
---

# Tech writer

This skill writes and revises technical documentation by applying the
vendored **nbj-write-clearly** skill (MIT; it paraphrases the Google Developer
Documentation Style Guide) and amending it for docs. nbj-write-clearly drives
the work: who the reader is, how the document is structured, and Google's
prose rules. This file adds rules nbj-write-clearly leaves open and patterns
to check.

Load nbj-write-clearly before you start, then apply this file on top of it.

**Precedence:** the user's explicit request and the destination's requirements
win. Then source facts, quotations, and technical tokens, which no rule may
change. Then this file, then nbj-write-clearly.

**This skill's style is the style.** When you invoke it, apply the rules of
this file and nbj-write-clearly in full. nbj-write-clearly's "Follow
project-specific style" step doesn't apply. The only project input you keep is
terminology: the names the project uses for its concepts. Don't copy the
existing document's habits, such as heading case, dashes, contractions, comma
use, emphasis, or spelling. Voice is a separate step that runs after this
skill, so don't preserve the existing voice either.

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
  - The words and punctuation of a heading, because other documents link to it. Change a heading's case freely when links match headings without regard to case, as they do in the specs corpus. Punctuation is part of the anchor, so a dash in a heading stays. Reword a heading only when you update every link to it in the same pass.
  - Labels and separators the document's template requires, such as **Status:**, **Rationale:**, or a `---` rule between requirements. They are structure, not decoration. Apply the rules to the text that follows them.
- Citation lists are exempt from T5's sentence length: runs of AC ids, wikilinks, ADR numbers, or references.
- HTML comments are not exempt. In templates they are instructions to the agent, so apply this skill to their prose. Keep any name, path, or token they cite.
- Content is fixed and wording is not. Never add or remove a fact, requirement, decision, example, or caveat, but apply every style rule to how each one is worded.
- Don't add opinions, reactions, or asides of your own. Keep the ones the author wrote, reworded to these rules.

## Google rules to apply every time

Apply these even where the document does otherwise. A document's existing
habits are easy to mistake for house style, and they are not:

- Sentence-case headings, including headings that are link targets (see Scope and defaults).
- US English spelling: "behavior," "judgment," "labeled."
- The serial comma: "code, tests, and docs."
- A conversational, friendly, and respectful tone, without slang. Use common contractions, especially negative ones ("don't," "isn't"). Precision still wins: never trade an exact term for a casual one.
- Em dashes without spaces ("word—word"), only where "Punctuation" calls for one. Spaced dashes, en dashes in prose, and double hyphens (` -- `) used as dashes are always wrong.
- No emphasis by capitals or bold. "Do NOT" becomes "Don't," and a bold lead-in sentence becomes a plain sentence or a heading. RFC 2119 keywords keep their capitals. Use italics for emphasis rarely, only where the sentence misreads without it.
- When prose names a heading, write it in the heading's case. Once headings are in sentence case, so are the references to them.
- Capitalized EARS keywords (WHEN, WHILE, IF, WHERE) belong only in requirement statements. In other prose, write them as ordinary words.
- Present tense for what is true now (T6).
- Name the reader early: say who the document is for and what they decide or do with it, when the source makes that clear.

## Punctuation

Choose the mark by what the words after it do. Ask these questions in order
and stop at the first yes.

1. **Could the second part stand as its own sentence, and is the link between
   the two only "and also"?** Use a period. Two short sentences beat one joined
   by a semicolon or a dash.
2. **Does the second part deliver what the first part promised: a list, an
   example, a definition, a cause, or the answer?** Use a colon. A test: you
   could write "namely" or "that is" in its place. The part before the colon
   must be a complete sentence ("The check fails for one reason: the anchor
   doesn't exist."). Template separators such as `[[slug]]: text` and
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

Rules that hold whatever the answer:

- Em dashes take no spaces. Never use an en dash, a spaced dash, or ` -- ` in
  prose. En dashes belong only in number ranges, and even there "to" reads
  better in running text ("pages 4 to 9").
- At most one pair of em dashes per sentence, and rarely more than one dash
  sentence per section. If you reach for a second, use parentheses or split
  the sentence. Several dashes in one section is a sign that most of them
  belong in parentheses, commas, or a sentence of their own.
- Don't use a semicolon where a period works. Use one only to separate list
  items that already contain commas.
- Requirement statements, headings, code, and link targets keep their
  punctuation (see Scope and defaults).

## Patterns to check

**Not X but Y.** A sentence that says what something isn't before saying what
it is. Keep the negative half when a reader might actually believe it, or when
it states a real limit: "This cache is not shared across workers," "The
`--force` flag no longer skips validation." Cut it when no reader would have
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
is actually true. If you don't know the condition, keep the hedge and add it
to the findings for the user.

**Passive voice.** Prefer "you" or the named component as the actor: "You
configure retries in `config.yaml`," "The scheduler retries failed jobs."
Passive is fine when the actor is irrelevant ("Logs are rotated daily").

**Writing about the document.** Release notes, changelogs, and migration
guides are about change, so "replaces," "previously," and "now" are correct
there. Elsewhere, cut "This README describes..." and "This section covers..."
openers.

**Wrong reader.** A doc's reader usually lacks context. Don't strip
prerequisites, definitions, or "why" on the grounds that the reader already
knows them. Check the doc against its intended reader: what would a newcomer
to this project need?

## New patterns for technical docs

**T1. Terminology drift.** The same concept appears under different names ("config file," "settings file," "configuration"). Pick one term per concept and use it everywhere. Never vary a technical term for style, even if a rewrite elsewhere encourages variety. Check that the names really are one concept first: some pairs that look like synonyms aren't (an Xcode workspace contains projects).

**T2. Vague capability claims.** "Supports," "handles," and "allows for" with no specifics. State what happens, the defaults, and the limits. "Supports large files" becomes "Uploads files up to 5 GB."

**T3. Missing specifics.** No units, no default, no version, or "run the appropriate command." Give the exact value or command. If you don't have it, leave the text as it is and add it to the findings for the user rather than guessing.

**T4. Nominalizations.** Verbs turned into nouns: "perform the installation of," "make a determination," "provide configuration of." This is the main source of a legalese feel. Use the verb: "install," "decide," "configure."

**T5. Long sentences and walls of text.** Sentences over about 25 words, or paragraphs over about five sentences. Split them. One idea per sentence. Citation lists don't count toward the length.

**T6. Future tense for current behavior.** "The function will return null" for something that happens now. Use present tense. Keep "will" for genuine future events (deprecations, scheduled removals).

**T7. Release notes in the wrong order.** Features listed in commit order, with breaking changes buried. Put breaking changes and required actions first, then user-visible changes, then fixes. Each item starts with what the user sees or must do.

**T8. Invented examples.** A rewrite must never invent a flag, an output, a default, or an example value. If an example would help and you don't have a real one, add it to the findings for the user.

## Rhythm by doc type

Vary sentence length in READMEs, explanations, tutorial narration, and announcements. For reference pages and numbered steps, consistent and parallel structure is better. Vary rhythm in the prose around the steps, not in the steps.

## Findings for the user

Some problems are not yours to fix: a contradiction, a miscount, a version
mismatch, a missing value, or an example you would have to invent. Don't fix
them, and don't mark them in the file with `TODO(author)`. Report each one to
the user with its file, line, and a quote, and ask them to decide. If the
user isn't available, save the findings list with your result so they can
decide later.

## Line wrapping

Follow the destination's line width. Google's Markdown guide keeps an
80-character limit, with exceptions for links, tables, headings, and code
blocks. Rewrap each paragraph you change to that width, and don't reflow
paragraphs you didn't change. Never reword a sentence to make it fit the
margin.

## How to run it

1. Load nbj-write-clearly, then this file.
2. Identify the reader, their goal, the doc type, and the task (draft, revise, or audit). Mark what must not change, using both nbj-write-clearly's list and Scope and defaults above.
3. **Draft or revise.** Follow nbj-write-clearly's "Write or revise" steps with this file's rules. Read its references/guide.md when the doc has procedures, code, UI labels, tables, images, or accessibility needs. Read its references/official-index.md only when the user asks for Google style compliance or a specialized rule decides the question.
4. **Check the patterns.** Run "Patterns to check" and the T-patterns over the result. Prefer a fresh-context reviewer: run this pass in a subagent that sees only the draft, the reader card if one exists, and these rules. A reviewer that wrote the draft tends to approve it.
5. **Validate.** Run nbj-write-clearly's "Validate" checklist on the final text, and check it against "Google rules to apply every time." Skip the checklist's last item, on sounding like the author; voice is a later step.
6. **Return the result.**
   - For an audit, change nothing. Report findings in priority order with short examples, as nbj-write-clearly says.
   - For pasted text, return the revised text.
   - For a file, write the revised file, summarize what changed, and list the findings for the user.
7. If the project uses Vale, run it last and fix what it flags.

Stop when the doc is clear, accurate, and consistent, and every line follows the rules above. Don't keep polishing lines that already do.
