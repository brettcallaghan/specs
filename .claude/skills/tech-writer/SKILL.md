---
name: tech-writer
description: Drafts, revises, and audits technical documentation so it reads as written by a person and follows Google developer style. Combines the humanizer and nbj-write-clearly skills, with amendments for docs. Use for READMEs, release notes and changelogs, docs-site pages, help-center content, UI copy, and specs written for human review.
---

# Tech writer

This skill writes and revises technical documentation by combining two
vendored skills, and amends both for docs:

- **nbj-write-clearly** (MIT; paraphrases the Google Developer Documentation
  Style Guide). It drives the work: who the reader is, how the document is
  structured, and Google's prose rules.
- **[humanizer](https://github.com/blader/humanizer)** (MIT), written against
  version 3.1.0. It removes AI writing patterns. Section numbers (§) refer to
  humanizer's patterns.

Load both skills before you start: nbj-write-clearly first, then humanizer,
then apply this file on top of them.

**Precedence:** the user's explicit request and the destination's requirements
win. Then source facts, quotations, and technical tokens, which no rule may
change. Then this file, then nbj-write-clearly, then humanizer.
nbj-write-clearly ranks above humanizer because it is written for
documentation, and humanizer for prose in general.

**This skill's style is the style.** When you invoke it, apply the rules of
this file and both skills in full. Two upstream rules don't apply here:

- nbj-write-clearly's "Follow project-specific style" step. The only project
  input you keep is terminology: the names the project uses for its concepts.
- Matching a writing sample. Neither the document you are revising nor its
  neighbors is a sample. Voice is a separate step that runs after this skill,
  so don't preserve the existing voice, and don't copy the existing
  document's habits: heading case, dashes, contractions, comma use,
  emphasis, or spelling.

A line is finished only when it follows every rule here. Matching the rest of
the document doesn't count.

## Scope and defaults

- Applies to READMEs, release notes and changelogs, docs-site pages, help-center content, UI copy, and specs written for human review.
- Do not apply it to marketing, legal, academic, fictional, or personal writing unless the user asks for this style.
- Default to humanizer's **file mode**: change prose only and leave code blocks, inline code, commands, paths, config keys, link targets, and front matter untouched.
- Also leave these untouched:
  - API, function, flag, and setting names
  - Error messages and log output quoted verbatim
  - Version numbers, dates, and numeric limits
  - Requirement statements in specs (EARS, Gherkin, "the system shall"), including whole Gherkin blocks. Their punctuation is part of their wording, dashes included.
  - RFC 2119 keywords (MUST, SHOULD, MAY) and their exact meaning
  - Wikilinks (`[[slug]]`, `[[slug#heading]]`) and identifiers such as AC and FEAT ids
  - Element record blocks
  - The words and punctuation of a heading, because other documents link to it. Change a heading's case freely when links match headings without regard to case, as they do in the specs corpus. Punctuation is part of the anchor, so a dash in a heading stays. Reword a heading only at section depth, and update every link to it in the same pass.
  - Labels and separators the document's template requires, such as **Status:**, **Rationale:**, or a `---` rule between requirements. They are structure, not decoration. Apply the rules to the text that follows them.
- Citation lists are exempt from T5's sentence length: runs of AC ids, wikilinks, ADR numbers, or references.
- HTML comments are not exempt. In templates they are instructions to the agent, so apply this skill to their prose. Keep any name, path, or token they cite.
- Content is fixed and wording is not. Never add or remove a fact, requirement, decision, example, or caveat, but apply every style rule to how each one is worded.
- Don't add opinions, reactions, or asides of your own. Keep the ones the author wrote, reworded to these rules.

## Where the two skills meet

They agree on most rules, so apply each shared rule once:

| Rule | nbj-write-clearly | humanizer, as amended here |
|---|---|---|
| Name the actor; prefer active voice and "you" | Write or revise, step 4 | §11 |
| One term per concept | step 6 | T1 |
| Short sentences, one idea each | step 7 | T5 |
| No inflated claims, filler, or pre-announcements | step 8 | §12, §13, §16, T2 |
| Sentence-case headings | step 9 | §20 |
| Present tense for current behavior | guide, Core prose | T6 |
| Never invent facts or examples | Protect meaning and voice | T8 |

Where they pull apart, this is how to resolve it:

- **Modality and tense.** nbj-write-clearly keeps the source's *can*, *might*, *should*, and *will*. §9 and T6 change them only when the source shows what is actually true. "Will return null" becomes "returns null" only for behavior that happens now. A hedge becomes a condition only when you know the condition. Otherwise leave the modal word as it is and add it to the findings for the user.
- **Rhythm.** nbj-write-clearly says not to flatten every sentence into one rhythm, and humanizer says to vary length. Both give way to parallel structure in steps and reference entries (see "Rhythm by doc type").
- **Tone.** Follow Google: conversational, friendly, and respectful, without slang. Use common contractions, especially negative ones ("don't," "isn't"). Precision still wins: never trade an exact term for a casual one.
- **Dashes.** §8 applies in full. Replace every em dash and en dash in prose, spaced or not, with a period, comma, colon, or parentheses, or rewrite the sentence.
- **Headings.** For how-to titles, §20 below takes precedence over Google's task-verb heading rule.

## Google rules to apply every time

Apply these even where the document does otherwise. A document's existing
habits are easy to mistake for house style, and they are not:

- Sentence-case headings, including headings that are link targets (see Scope and defaults).
- US English spelling: "behavior," "judgment," "labeled."
- The serial comma: "code, tests, and docs."
- Contractions, as Tone says.
- No dashes, as Dashes says.
- No emphasis by capitals or bold. "Do NOT" becomes "Don't," and a bold lead-in sentence becomes a plain sentence or a heading. RFC 2119 keywords keep their capitals. Use italics for emphasis rarely, only where the sentence misreads without it.
- When prose names a heading, write it in the heading's case. Once headings are in sentence case, so are the references to them.
- Capitalized EARS keywords (WHEN, WHILE, IF, WHERE) belong only in requirement statements. In other prose, write them as ordinary words.
- Present tense for what is true now (T6).
- Name the reader early: say who the document is for and what they decide or do with it, when the source makes that clear.

## Amendments to existing patterns

**§1 Not X but Y.** Keep a contrast when it corrects a real misconception or documents a real limit: "This cache is not shared across workers," "The `--force` flag no longer skips validation." Cut it when the negative half is a straw man.

**§6 Forced triads.** Exempt real enumerations: three supported platforms, three config options, three steps. The test is whether the list is complete and accurate, not whether it has three items.

**§7 Repeated sentence openings.** Exempt numbered steps and reference entries. Steps should start with an imperative verb ("Run," "Add," "Open"), and reference entries should share a structure. Parallel structure helps scanning here.

**§9 Stacked qualifiers.** In docs, replace hedges with conditions. "This may fail in some cases" becomes "This fails when the disk is full." If you don't know the condition, keep the hedge and add it to the findings for the user.

**§11 Passive voice.** Prefer "you" or the named component as the actor: "You configure retries in `config.yaml`," "The scheduler retries failed jobs." Passive is fine when the actor is irrelevant ("Logs are rotated daily").

**§12 Overused AI words.** Keep technical senses: robust (statistics), key (API key, key-value), gate (logic gate, feature gate), highlight (UI highlighting), landscape (orientation), enhance (image processing). Add these docs-specific tells:
- seamless, effortless, powerful, cutting-edge, best-in-class, world-class
- leverage, utilize (use "use"), facilitate, empower, streamline
- comprehensive, flexible (unless you name the flexibility)
- ensure (unless something is actually guaranteed)
- simply, just, easy, easily, obviously, of course (these patronize a reader who is stuck)

**§19 Bold as decoration.** Bold is correct for UI labels in steps (Click **Save**) and for terms in a definition list. Release notes with a bold label on every bullet should become plain items that start with the user-visible change.

**§20 Decorative headings.** Agree on sentence case. Docs headings are navigation, so keep them on short sections. Make sibling headings parallel. How-to titles start with "How to."

**§25 Writing about the document.** Release notes, changelogs, and migration guides are about change, so "replaces," "previously," and "now" are correct there. Elsewhere, cut "This README describes..." and "This section covers..." openers.

**§26 Wrong reader.** Reverse the default for docs. Unlike a reply in a thread, a doc's reader usually lacks context. Do not strip prerequisites, definitions, or "why" on the grounds that the reader already knows them. Instead, check the doc against its intended reader: what would a newcomer to this project need?

## New patterns for technical docs

**T1. Terminology drift.** The same concept appears under different names ("workspace," "project," "space"). Pick one term per concept and use it everywhere. Never vary a technical term for style, even if a rewrite elsewhere encourages variety.

**T2. Vague capability claims.** "Supports," "handles," and "allows for" with no specifics. State what happens, the defaults, and the limits. "Supports large files" becomes "Uploads files up to 5 GB."

**T3. Missing specifics.** No units, no default, no version, or "run the appropriate command." Give the exact value or command. If you don't have it, leave the text as it is and add it to the findings for the user rather than guessing.

**T4. Nominalizations.** Verbs turned into nouns: "perform the installation of," "make a determination," "provide configuration of." This is the main source of a legalese feel. Use the verb: "install," "decide," "configure."

**T5. Long sentences and walls of text.** Sentences over about 25 words, or paragraphs over about five sentences. Split them. One idea per sentence. Citation lists don't count toward the length.

**T6. Future tense for current behavior.** "The function will return null" for something that happens now. Use present tense. Keep "will" for genuine future events (deprecations, scheduled removals).

**T7. Release notes in the wrong order.** Features listed in commit order, with breaking changes buried. Put breaking changes and required actions first, then user-visible changes, then fixes. Each item starts with what the user sees or must do.

**T8. Invented examples.** A rewrite must never invent a flag, an output, a default, or an example value. If an example would help and you don't have a real one, add it to the findings for the user.

## Rhythm by doc type

Humanizer's advice to vary sentence length applies to READMEs, explanations, tutorial narration, and announcements. For reference pages and numbered steps, consistent and parallel structure is better. Vary rhythm in the prose around the steps, not in the steps.

## Findings for the user

Some problems are not yours to fix: a contradiction, a miscount, a version
mismatch, a missing value, or an example you would have to invent. Don't fix
them, and don't mark them in the file with `TODO(author)`. Report each one to
the user with its file, line, and a quote, and ask them to decide. If the
user isn't available, save the findings list with your result so they can
decide later.

## Depth

The caller sets how deep a revision goes. Without a setting, revise at
paragraph depth.

- **Sentence.** Fix AI patterns and long sentences one line at a time. Don't restructure.
- **Paragraph.** Everything at sentence depth. Also rework any paragraph that fails T5: split it, reorder it, or turn a run-on list into a real list. Change heading case.
- **Section.** Everything at paragraph depth. Also restructure sections, remove duplicated rationale, and reword headings with every link to them updated.

## Line wrapping

Follow the destination's line width. Google's Markdown guide keeps an
80-character limit, with exceptions for links, tables, headings, and code
blocks. Rewrap each paragraph you change to that width, and don't reflow
paragraphs you didn't change. Never reword a sentence to make it fit the
margin.

## How to run it

1. Load nbj-write-clearly, then humanizer, then this file.
2. Identify the reader, their goal, the doc type, the task (draft, revise, or audit), and the depth. Mark what must not change, using both nbj-write-clearly's list and Scope and defaults above.
3. **Draft or revise for clarity.** Follow nbj-write-clearly's "Write or revise" steps. Read its references/guide.md when the doc has procedures, code, UI labels, tables, images, or accessibility needs. Read its references/official-index.md only when the user asks for Google style compliance or a specialized rule decides the question.
4. **Remove AI patterns.** Run humanizer's process over the result, with the amendments and T-patterns above. Prefer a fresh-context reviewer: run this pass in a subagent that sees only the draft, the reader card if one exists, and these rules. A reviewer that wrote the draft tends to approve it.
5. **Validate.** Run nbj-write-clearly's "Validate" checklist on the final text, and check it against "Google rules to apply every time." Skip the checklist's last item, on sounding like the author; voice is a later step.
6. **Return the result.**
   - For an audit, change nothing. Report findings in priority order with short examples, as nbj-write-clearly says.
   - For pasted text, return humanizer's usual output.
   - In file mode, write the revised file, summarize what changed, and list the findings for the user.
7. If the project uses Vale, run it last and fix what it flags.

Stop when the doc is clear, accurate, consistent, and free of AI patterns, and every line follows the rules above. Don't keep polishing lines that already do.
