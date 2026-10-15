---
name: humanizer-tech-writer
description: Amendments to humanizer for technical docs. Use with humanizer when cleaning AI tells from READMEs, release notes, docs pages, or specs.
---

# Humanizer amendments for technical docs

This skill amends [humanizer](https://github.com/blader/humanizer) (MIT license),
written against version 3.1.0. Load humanizer first, then apply these amendments.
Section numbers (§) refer to humanizer's patterns.

**Precedence:** the project's own style guide and terminology table win, then
this file, then humanizer.

## Scope and defaults

- Applies to READMEs, release notes and changelogs, docs-site pages, and specs written for human review.
- Default to humanizer's **file mode**: change prose only and leave code blocks, inline code, commands, paths, config keys, link targets, and front matter untouched.
- Also leave these untouched:
  - API, function, flag, and setting names
  - Error messages and log output quoted verbatim
  - Version numbers, dates, and numeric limits
  - Requirement statements in specs (EARS, Gherkin, "the system shall")
  - RFC 2119 keywords (MUST, SHOULD, MAY) and their exact meaning
- Voice for these doc types is neutral and plain, as humanizer already says for technical text. Do not add opinions, reactions, or asides, except in a README intro or a blog-style release announcement, and only if the user's sample voice does that.

## Amendments to existing patterns

**§1 Not X but Y.** Keep a contrast when it corrects a real misconception or documents a real limit: "This cache is not shared across workers," "The `--force` flag no longer skips validation." Cut it when the negative half is a straw man.

**§6 Forced triads.** Exempt real enumerations: three supported platforms, three config options, three steps. The test is whether the list is complete and accurate, not whether it has three items.

**§7 Repeated sentence openings.** Exempt numbered steps and reference entries. Steps should start with an imperative verb ("Run," "Add," "Open"), and reference entries should share a structure. Parallel structure helps scanning here.

**§9 Stacked qualifiers.** In docs, replace hedges with conditions. "This may fail in some cases" becomes "This fails when the disk is full." If you don't know the condition, flag it for the user instead of keeping the hedge.

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

**T3. Missing specifics.** No units, no default, no version, or "run the appropriate command." Give the exact value or command. If you don't have it, insert `TODO(author): ...` rather than guessing.

**T4. Nominalizations.** Verbs turned into nouns: "perform the installation of," "make a determination," "provide configuration of." This is the main source of a legalese feel. Use the verb: "install," "decide," "configure."

**T5. Long sentences and walls of text.** Sentences over about 25 words, or paragraphs over about five sentences, in procedural or reference content. Split them. One idea per sentence.

**T6. Future tense for current behavior.** "The function will return null" for something that happens now. Use present tense. Keep "will" for genuine future events (deprecations, scheduled removals).

**T7. Release notes in the wrong order.** Features listed in commit order, with breaking changes buried. Put breaking changes and required actions first, then user-visible changes, then fixes. Each item starts with what the user sees or must do.

**T8. Invented examples.** A rewrite must never invent a flag, an output, a default, or an example value. If an example would help and you don't have a real one, flag it for the author.

## Rhythm by doc type

Humanizer's advice to vary sentence length applies to READMEs, explanations, tutorial narration, and announcements. For reference pages and numbered steps, consistent and parallel structure is better. Vary rhythm in the prose around the steps, not in the steps.

## How to run it

1. Load humanizer, then this file.
2. Prefer a fresh-context reviewer: run this pass in a subagent that sees only the draft, the reader card if one exists, and these rules. A reviewer that wrote the draft tends to approve it.
3. Return humanizer's usual output for pasted text. In file mode, write the revised file and summarize what changed, including any `TODO(author)` flags.
4. If the project uses Vale, run it last and fix what it flags.
