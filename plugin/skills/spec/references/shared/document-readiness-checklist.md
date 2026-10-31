<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/standards/document-readiness-checklist.md
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Document Readiness Checklist
description: The type-independent self-check every spec document passes
  before it is set to ready-for-review — frontmatter, slug, status, EARS
  and Gherkin conformance, line width, instruction blocks, references
slug: document-readiness-checklist
type: standards
status: active
---

# Document readiness checklist

## Contents

- Who decides — which checks the `specs` CLI settles and which are yours
- Two notations, two surfaces — which of #4–#5 and #9 applies where
- Frontmatter checks #1–#3 — does the metadata conform? (required
  properties, slug, status vocabulary)
- Normative-statement checks #4–#5 — are the EARS requirements well
  formed? (EARS coverage, compound modal verbs)
- Finished-document checks #6–#8 — is the document finished? (line
  width, instruction blocks, reference resolution)
- Gherkin check #9 — are the features and acceptance criteria well
  formed?
- Architecture blueprint checks #10–#11 — does the blueprint state only
  what the code can't? (blueprint audit, code references, ownership,
  sections)

This list applies to every document type. The `spec` skill runs these
checks before any document is set to `ready-for-review`. Only this
document defines the checks. The skill cites it instead of restating
it, so the authoring workflow and this checklist can't disagree about
what "conformant" means.

Feature Blueprints also run [[blueprint-readiness-checklist]]. The two
checklists don't overlap: that one asks whether an approved design is
*groomable*, and this one asks whether a document of any type
*conforms* to the framework's authoring rules.

Most checks below enforce rules that the framework already owns
elsewhere. Where that's so, the check cites the owner instead of
repeating the rule, because a restated rule drifts from its source. Only
checks 4, 5, and 7 are stated in this document's own words.

---

## Two notations, two surfaces

Normative statements use one of two notations. *What the statement is*
decides which one applies, not the document type:

| Surface | Notation | Checks |
|---|---|---|
| An FRD's features and acceptance criteria (its fenced `gherkin` blocks) | Gherkin ([[adr-021]], [[gherkin-guide]]) | 9 |
| Every other normative statement, in every document type, including an FRD's prose outside those blocks | EARS ([[adr-002]], [[guide-ears]]) | 4, 5 |

Most documents have only the second surface and run only checks 4 and
5. An FRD has both: its blocks run check 9, and the prose around them
still runs checks 4 and 5. This table exists to prevent applying the
wrong list. An EARS sentence inside a `gherkin` block doesn't parse, and
a `Then` step in a Standards document is a criterion that nothing
traces.

---

## Who decides

Run `specs check <path>` first. It settles checks 1–3 and 8 outright,
and the machine-checkable half of check 9. It parses the frontmatter and
validates every property against the schema for the document's `type`.
It matches the slug against that type's registered pattern and reports a
slug that another document in the corpus already claims. It resolves
every `[[slug]]` reference against the corpus, both in the body and in
the frontmatter properties that hold slugs. Read its findings. Each one
names the rule, the span, and the value that violated it.

**A single-document check covers that document and the references your
change broke elsewhere.** `specs check <path>` resolves the document's
own references against the whole corpus. It reports every finding
located in the document, each at its own line: an unresolved reference
or anchor, a retired target, a type mismatch, or a slug or identifier it
shares with another document. It also reports two kinds of finding
located in other documents, each at its line in the citing document. The
first is a reference to a slug that the document declared at git HEAD
and no longer declares. The second is an anchor into the document that
no longer names one of its headings. It reports nothing else located in
another document. It doesn't report a retirement or a changed `type`
elsewhere, because the citing reference still resolves.

So before you rename, retire, or re-type a document, run
`specs refs <slug>`. It names every document that cites the slug, which
is the blast radius of the change. After the change, a corpus-wide
`specs check` reports what's broken anywhere.

**Check the path you got back.** When a path names no corpus member,
`specs check` exits 2 and says so. If the file exists but was excluded,
the message names the exclusion reason. It never reports a clean
document for a path that the corpus doesn't hold.

A `frontmatter/parse` finding (row 0 in the following table) comes
before all the others and isn't one of the eleven checks. It means the
frontmatter couldn't be read as a document envelope, so no other check
examined the document. Fix it and run the check again before you read
anything else in the report. The usual cause is a `status` or other
enumerated value outside its vocabulary.

| # | Check | Decided by |
|---|-------|------------|
| 0 | Frontmatter is readable at all | `specs check` |
| 1 | Required properties present and parseable | `specs check` |
| 2 | Slug matches the pattern for this type; unique | `specs check` |
| 3 | Status is in the controlled vocabulary | `specs check` |
| 4 | Every EARS normative statement uses an EARS pattern | you |
| 5 | No requirement joins two modal verbs | you |
| 6 | No prose line exceeds the width limit | you — CLI planned |
| 7 | No template instruction blocks survive | you |
| 8 | Every `[[slug]]` reference resolves | `specs check` |
| 9 | Every feature and criterion is well formed and identified | `specs check`, plus you |
| 10 | An architecture blueprint passes the blueprint audit | you |
| 11 | An architecture blueprint's code references resolve, its source paths don't overlap a sibling's, and its required sections are present | you — CLI planned |

**A clean `specs check` isn't a passing document.** It means that row
0, checks 1–3 and 8, and the machine-checkable half of check 9 passed.
Checks 4, 5, 7, and 10 are judgment calls the tool doesn't make, and
checks 6 and 11 are specified but not built. For those six, a silent tool
isn't a passing grade.

Check 8 covers heading anchors too: `specs check` resolves
`[[container-x#Some Heading]]` on `container-x`, and then matches the
anchor against that document's headings.
[[spec-cli-reference#What it does not check]] lists the reference cases
that a clean run says nothing about.

Where the CLI isn't installed, every check below is still defined and
still runs, and you carry out checks 1–3 by hand against the property
definitions and prefix registry that the checks cite. Say which
happened. "`specs check` reports no findings" and "I read the
frontmatter and it looks right" are different claims, and a reviewer
can re-run only the first. [[spec-cli-reference]] covers installation.

---

## Frontmatter checks

1. The frontmatter parses as YAML and declares every required property:
   title, description, slug, type, and status.
   [[template-authoring-standards]] owns the property definitions.
2. The slug follows the registry rule for the document's type. That rule
   is the "Slug conventions" section of [[template-authoring-standards]],
   including the type-prefix format and the Standards type's role-noun
   exception.
3. The status is a value from the controlled vocabulary, which
   [[template-authoring-standards]] defines by pointing at
   [[frontmatter-standards]]. A status outside that vocabulary is an
   error, not a local convention.

## Normative-statement checks

Checks 4 and 5 cover the EARS surface. An FRD's fenced `gherkin` blocks
belong to check 9, but the prose around them still runs these two.

4. Every normative statement uses an EARS pattern. A sentence that
   carries a capitalized SHALL, SHOULD, or MAY is a requirement. A
   requirement with no EARS trigger (ubiquitous, event-driven,
   state-driven, optional-feature, unwanted-behavior, or complex) isn't
   testable. [[guide-ears]] holds the patterns.
5. No requirement joins two modal verbs with "and". One requirement per
   sentence: two SHALLs in one sentence can't be tested, traced, or
   disputed independently, so split them.

## Finished-document checks

6. No prose line exceeds the line-width limit that
   [[template-authoring-standards]] sets. That document also states the
   exemptions for URLs, code-block content, and table cells.
7. Every instruction comment block inherited from the template is
   deleted, at document level and section level alike. Search the
   finished document for `<!--`. A surviving block is authoring
   scaffolding that a reader will mistake for content.
8. Every `[[slug]]` reference resolves to an existing document.
   [[cross-document-references]] owns what a reference resolves against
   and the notation it must use. `specs check` decides this check, so run
   it instead of reading the links. Read two of its findings instead of
   only counting them. A `reference/retired-target` **warning** means
   the link resolves but the target has been replaced: cite the
   successor it names instead. A `reference/type-mismatch` means the
   property resolved to a document of the wrong type.

## Gherkin check

Run this check only on a document that holds fenced `gherkin` blocks,
which in practice means an FRD. [[gherkin-guide]] owns every rule it
tests for.

9. Every feature and criterion is well formed and identified.
   `specs check` settles the machine-checkable half: an untagged or
   malformed identifier, a duplicate one, a criterion whose code doesn't
   derive from its feature's, an empty scenario, prose in a `Scenario`
   description, and partially tagged `Examples` blocks. It reports these
   findings under `gherkin/*` rule identifiers.

   Three parts are yours, because the tool can't read intent:

   - **An `Examples` block that should have been its own criterion.** If
     you tag the scenario when you meant several criteria, all but the
     first cease to exist. The document parses cleanly, and no check
     fires against it. Only *partial* tagging is machine-detectable.
   - **A prohibition with no `@negative`, or a hand-verified criterion
     with no `@manual`.** Gherkin has no grammatical negative, so the tag
     is the whole of the marking.
   - **A step shortened to fit a margin.** No width limit applies inside
     a fence. Compressing a step to satisfy one silently changes what
     the criterion says.

## Architecture blueprint checks

Run these checks only on a System, Container, Module, or Component
Blueprint. [[blueprint-content-standards]] owns every rule they test for.

10. The blueprint passes the blueprint audit in
    [[blueprint-content-standards]].
11. Every backticked path and symbol names something under the
    blueprint's `source_paths` or `test_paths`. A Component or Module
    Blueprint's `source_paths` don't overlap a sibling's. Every section
    the type's template requires is present.

---

## Running the checklist

Run the checklist before you set a document to `ready-for-review`, fix
what it reports, and then run it again. A fix for one check often breaks
another: rewrapping a line for check 6 can bury the compound requirement
that check 5 looks for. Run `specs check <path>` again on each pass too.
It's one subprocess, and an edit made to satisfy check 5 can break check
1 by disturbing the frontmatter above it. Don't set `ready-for-review`
while any check fails.

## References

- [[spec-cli-reference]] — the CLI that settles checks 1–3, 8, and most
  of 9, and what it doesn't yet check
- [[template-authoring-standards]] — frontmatter properties, slug
  conventions, line width, instruction blocks
- [[frontmatter-standards]] — the `status` controlled vocabulary
- [[cross-document-references]] — how a `[[slug]]` resolves
- [[guide-ears]] — the EARS patterns that checks 4 and 5 test for
- [[gherkin-guide]] — the tag vocabulary, description slots, and silent
  failures that check 9 tests for
- [[blueprint-readiness-checklist]] — the separate groomability gate
  that Feature Blueprints also run
- [[blueprint-content-standards]] — the rules and the audit that checks
  10 and 11 test for
