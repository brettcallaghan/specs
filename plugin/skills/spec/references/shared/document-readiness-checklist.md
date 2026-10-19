<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/standards/document-readiness-checklist.md (version unversioned)
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

- Who decides: which checks the `specs` CLI settles and which are yours
- Two notations, two surfaces: which of checks 4, 5, and 9 applies where
- Frontmatter checks 1 to 3: does the metadata conform? (required
  properties, slug, status vocabulary)
- Normative-statement checks 4 and 5: are the EARS requirements well
  formed? (EARS coverage, compound modal verbs)
- Finished-document checks 6 to 8: is the document finished? (line
  width, instruction blocks, reference resolution)
- Gherkin check 9: are the features and acceptance criteria well
  formed?

This checklist applies to every document type. The `spec` skill runs these
checks before any document is set to `ready-for-review`. The checks are
defined only here. The skill cites this document instead of restating it,
so the authoring workflow and this checklist can't disagree about what
"conformant" means.

Feature blueprints also run `blueprint-readiness-checklist.md`. The two
don't overlap. That one asks whether an approved design is groomable, and
this one asks whether a document of any type conforms to the framework's
authoring rules.

Most checks below enforce rules that the framework already owns elsewhere.
For those, the check cites its owner instead of repeating the rule, because
a restated rule drifts from its source. Only checks 4, 5, and 7 are stated
here, in this document's own voice.

---

## Two notations, two surfaces

Normative statements use two notations. The statement itself, not the
document type, decides which notation applies:

| Surface | Notation | Checks |
|---|---|---|
| An FRD's features and acceptance criteria (its fenced `gherkin` blocks) | Gherkin ([[adr-021]], [[gherkin-guide]]) | 9 |
| Every other normative statement in every document type, including an FRD's prose outside those blocks | EARS ([[adr-002]], `ears-guide.md`) | 4, 5 |

Most documents have only the second surface and run only checks 4 and 5.
An FRD has both: its blocks run check 9, and the prose around them still
runs checks 4 and 5. This table exists to prevent applying the wrong list.
An EARS sentence inside a `gherkin` block doesn't parse, and a `Then` step
in a standards document is a criterion that nothing traces.

---

## Who decides

Run `specs check <path>` first. It settles checks 1 to 3 and 8 outright,
and the machine-checkable half of check 9. It parses the frontmatter and
validates every property against the schema for the document's `type`. It
matches the slug against that type's registered pattern and reports a slug
that another document in the corpus already claims. It also resolves every
`[[slug]]` reference against the corpus, both in the body and in the
frontmatter properties that hold slugs. Read its findings: each one names
the rule, the span, and the value that violated it.

A single-document check covers that document only. `specs check <path>`
resolves the document's own references against the whole corpus. It
reports every finding located in that document, each at its own line: an
unresolved reference, a retired target, a type mismatch, or a slug or
identifier it shares with another document. It reports nothing located in
another document. That includes a reference held elsewhere that your edit
left pointing at nothing.

So before you rename, retire, or re-type a document, run
`specs refs <slug>`. It names every document that cites the slug, which
shows everything the change affects. After the change, a corpus-wide
`specs check` reports what's broken anywhere.

Check the path you got back. When a path names no corpus member,
`specs check` exits 2 and says so. If the file exists but was excluded, it
names the exclusion reason. It never reports a clean document for a path
the corpus doesn't hold.

A `frontmatter/parse` finding (row 0 below) comes before all the others and
isn't one of the nine checks. It means the CLI couldn't read the frontmatter
as a document envelope, so no other check examined this document. Fix it
and rerun before you read anything else in the report. The usual cause is a
`status` or other enumerated value outside its vocabulary.

| # | Check | Decided by |
|---|-------|------------|
| 0 | Frontmatter is readable at all | `specs check` |
| 1 | Required properties present and parseable | `specs check` |
| 2 | Slug matches the pattern for this type; unique | `specs check` |
| 3 | Status is in the controlled vocabulary | `specs check` |
| 4 | Every EARS normative statement uses an EARS pattern | you |
| 5 | No requirement joins two modal verbs | you |
| 6 | No prose line exceeds the width limit | you (CLI planned) |
| 7 | No template instruction blocks survive | you |
| 8 | Every `[[slug]]` reference resolves | `specs check` |
| 9 | Every feature and criterion is well formed and identified | `specs check`, plus you |

A clean `specs check` doesn't mean the document passes. It means row 0,
checks 1 to 3 and 8, and the machine-checkable half of check 9 passed.
Checks 4, 5, and 7 need judgment that the tool doesn't make, and check 6 is
specified but not built. For those four, a silent tool isn't a passing
grade.

Check 8 has one part that the CLI still leaves to you. It resolves the slug
half of a reference but doesn't interpret a heading anchor, so
`[[container-x#Some Heading]]` passes on `container-x` alone. If you wrote
an anchor, confirm that the heading exists. `spec-cli.md` lists the other
reference cases that a clean run says nothing about.

Where the CLI isn't installed, every check below is still defined and still
runs. You carry out checks 1 to 3 by hand instead, against the property
definitions and prefix registry that the checks cite. Say which one
happened. "`specs check` reports no findings" and "I read the frontmatter
and it looks right" are different claims, and a reviewer can rerun only the
first. `spec-cli.md` covers installation.

---

## Frontmatter checks

1. The frontmatter parses as YAML and declares every required property:
   title, description, slug, type, and status.
   [[template-authoring-standards]] owns the property definitions.
2. The slug follows the registry rule for this document's type. That rule
   is in the Slug conventions section of [[template-authoring-standards]],
   including the type-prefix format and the standards type's role-noun
   exception.
3. The status is a value from the controlled vocabulary.
   [[template-authoring-standards]] defines that vocabulary by pointing at
   [[frontmatter-standards]]. A status outside that vocabulary is an
   error, not a local convention.

## Normative-statement checks

These two checks cover the EARS surface. An FRD's fenced `gherkin` blocks
belong to check 9, but these checks still apply to the prose around them.

4. Every normative statement uses an EARS pattern. A sentence that carries
   a capitalized SHALL, SHOULD, or MAY is a requirement. A requirement with
   no EARS trigger (ubiquitous, event-driven, state-driven,
   optional-feature, unwanted-behavior, or complex) isn't testable.
   `ears-guide.md` holds the patterns.
5. No requirement joins two modal verbs with "and". Write one requirement
   per sentence. Two SHALLs in one sentence can't be tested, traced, or
   disputed independently, so split them.

## Finished-document checks

6. No prose line exceeds the line-width limit set by
   [[template-authoring-standards]], which also states the exemptions for
   URLs, code-block content, and table cells.
7. Every instruction comment block inherited from the template has been
   deleted, at document level and section level alike. Search the finished
   document for `<!--`. A surviving block means the document still carries
   authoring scaffolding that a reader will mistake for content.
8. Every `[[slug]]` reference resolves to an existing document.
   [[cross-document-references]] owns what a reference resolves against
   and the notation it must use. `specs check` decides this check, so run
   it instead of reading the links. Two of its findings are worth reading,
   not just counting. A `reference/retired-target` warning means the link
   resolves but the target has been replaced, so cite the successor it
   names instead. A `reference/type-mismatch` means the property resolved
   to a document of the wrong type.

## Gherkin check

Run this check only on a document that holds fenced `gherkin` blocks, which
in practice means an FRD. `gherkin-guide.md` owns every rule it tests for.

9. Every feature and criterion is well formed and identified.
   `spec check` settles the machine-checkable half: an untagged or
   malformed identifier, a duplicate one, a criterion whose code doesn't
   derive from its feature's, an empty scenario, prose in a `Scenario`
   description, and partially tagged `Examples` blocks. It reports the
   findings under `gherkin/*` rule identifiers.

   Three parts are yours, because the tool can't read intent:

   - An `Examples` block that should have been its own criterion. If you
     meant several criteria and tag the scenario, all but the first cease
     to exist. The document still parses cleanly, and no check fires
     against it. Only partial tagging is machine-detectable.
   - A prohibition with no `@negative`, or a hand-verified criterion with
     no `@manual`. Gherkin has no grammatical negative, so the tag is the
     only marking.
   - A step shortened to fit a margin. No width limit applies inside a
     fence, so compressing a step to satisfy one silently changes what the
     criterion says.

---

## Running the checklist

Run the checklist before you set a document to `ready-for-review`, fix what
it reports, and then run it again. A fix for one check often breaks
another. For example, rewrapping a line for check 6 can bury the compound
requirement that check 5 looks for. Rerun `specs check <path>` on each pass
too. It's one subprocess, and an edit made to satisfy check 5 can break
check 1 by disturbing the frontmatter above it. Don't set
`ready-for-review` while any check fails.

## References

- `spec-cli.md`: the CLI that settles checks 1 to 3, 8, and most of 9, and
  what it doesn't check yet
- `template-authoring-standards.md` (same directory): frontmatter
  properties, slug conventions, line width, and instruction blocks
- `frontmatter-standards.md`: the `status` controlled vocabulary
- `cross-document-references.md`: how a `[[slug]]` resolves
- `ears-guide.md`: the EARS patterns that checks 4 and 5 test for
- `gherkin-guide.md`: the tag vocabulary, description slots, and silent
  failures that check 9 tests for
- `blueprint-readiness-checklist.md`: the separate groomability gate that
  feature blueprints also run
