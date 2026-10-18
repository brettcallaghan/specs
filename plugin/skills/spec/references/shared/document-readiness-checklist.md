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

# Document Readiness Checklist

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

One list, every document type. The `spec` skill runs these checks before
any document is set to `ready-for-review`. The checks are defined only
here — the skill references this document rather than restating it, so
the authoring workflow and this checklist can never disagree about what
"conformant" means.

Feature blueprints additionally run
`blueprint-readiness-checklist.md`. The two do not overlap: that one
asks whether an approved design is *groomable*; this one asks whether a
document of any type *conforms* to the framework's authoring rules.

Most checks below enforce rules the framework already owns elsewhere.
Where that is so, the check cites its owner instead of repeating it — a
restated rule drifts from its source. Only checks 4, 5 and 7 are stated
here in this document's own voice.

---

## Two notations, two surfaces

Normative statements are written in two notations, and which one applies
is decided by *what the statement is*, not by the document type:

| Surface | Notation | Checks |
|---|---|---|
| An FRD's features and acceptance criteria — its fenced `gherkin` blocks | Gherkin ([[adr-021]], [[gherkin-guide]]) | 9 |
| Every other normative statement, in every document type — including an FRD's prose outside those blocks | EARS ([[adr-002]], `ears-guide.md`) | 4, 5 |

Most documents have only the second surface and run only checks 4 and 5.
An FRD has both: its blocks run check 9, and the prose around them still
runs 4 and 5. Applying the wrong list is the failure this table exists to
prevent — an EARS sentence inside a `gherkin` block does not parse, and a
`Then` step in a standards document is a criterion nothing traces.

---

## Who decides

Run `specs check <path>` first. It settles checks 1–3 and 8 outright,
and the machine-checkable half of check 9. It parses the frontmatter,
validates every property against the schema for the document's `type`,
matches the slug against that type's registered pattern, and reports a
slug already claimed elsewhere in the corpus. It also resolves every
`[[slug]]` reference against the corpus, both in the body and in the
frontmatter properties that hold slugs. Read its findings; each names
the rule, the span, and the value that violated it.

**A single-document check covers that document only.**
`specs check <path>` resolves that document's own references against the
whole corpus and reports every finding located in it, each at its own
line — an unresolved reference, a retired target, a type mismatch, a
slug or identifier it shares with another document. It reports nothing
located in another document, including a reference held elsewhere that
your edit left pointing at nothing.

So before you rename, retire or re-type a document, run
`specs refs <slug>`: it names every document citing it, which is every
document the change can break. After the change, a corpus-wide `specs check`
reports what is broken anywhere.

**Check the path you got back.** `specs check` on a path that names no
corpus member exits 2 and says so, naming the exclusion reason where the
file exists but was excluded. It never reports a clean document for a
path the corpus does not hold.

A `frontmatter/parse` finding (row 0 below) comes before all of them and
is not one of the nine: it means the frontmatter could not be read as a
document envelope, so no other check examined this document. Fix it and
re-run before reading anything else in the report. The usual cause is a
`status` or other enumerated value outside its vocabulary.

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

**A clean `specs check` is not a passing document.** It means row 0 and
checks 1–3, 8 and the machine-checkable half of 9 passed. Checks 4, 5 and
7 are judgement the tool does not make, and 6 is specified but not
built — a silent tool is not a passing grade for them.

Check 8 has one part the CLI still leaves to you: it resolves the slug
half of a reference and does not interpret a **heading anchor**, so
`[[container-x#Some Heading]]` passes on `container-x` alone. If you
wrote an anchor, confirm the heading exists. `spec-cli.md` lists the
other reference cases a clean run says nothing about.

WHERE the CLI is not installed, every check below is still defined and
still runs; you carry out 1–3 by hand instead, against the property
definitions and prefix registry the checks cite. Say which happened —
"`specs check` reports no findings" and "I read the frontmatter and it
looks right" are different claims, and only the first can be re-run by a
reviewer. `spec-cli.md` covers installation.

---

## Frontmatter checks

1. The frontmatter parses as YAML and declares every required property:
   title, description, slug, type and status. The property
   definitions are owned by [[template-authoring-standards]].
2. The slug follows the registry rule for this document's type — the
   Slug Conventions of [[template-authoring-standards]], including the
   type-prefix format and the standards type's role-noun exception.
3. The status is a value from the controlled vocabulary, which
   [[template-authoring-standards]] defines by pointing at
   [[frontmatter-standards]]. A status outside that vocabulary is an
   error, not a local convention.

## Normative-statement checks

These two cover the EARS surface. An FRD's fenced `gherkin` blocks are
check 9's; the prose around them is still these.

4. Every normative statement uses an EARS pattern. A sentence carrying a
   capitalized SHALL, SHOULD or MAY is a requirement, and a requirement
   with no EARS trigger — ubiquitous, event-driven, state-driven,
   optional-feature, unwanted-behavior or complex — is not testable.
   `ears-guide.md` holds the patterns.
5. No requirement joins two modal verbs with "and". One requirement per
   sentence: two SHALLs in one sentence cannot be tested, traced or
   disputed independently, so split them.

## Finished-document checks

6. No prose line exceeds the line-width limit set by
   [[template-authoring-standards]], which also states the exemptions
   for URLs, code-block content and table cells.
7. Every instruction comment block inherited from the template has been
   deleted — document-level and section-level alike. Search the finished
   document for `<!--`; a surviving block means it still carries
   authoring scaffolding a reader will mistake for content.
8. Every `[[slug]]` reference resolves to an existing document.
   [[cross-document-references]] owns what a reference resolves against
   and the notation it must use. `specs check` decides this one; run it
   rather than reading the links. Two of its findings are worth reading
   rather than only counting: a `reference/retired-target` **warning**
   means the link resolves but the target has been replaced — cite the
   successor it names instead — and a `reference/type-mismatch` means
   the property resolved to a document of the wrong type.

## Gherkin check

Run this one only on a document holding fenced `gherkin` blocks — in
practice an FRD. `gherkin-guide.md` owns every rule it tests for.

9. Every feature and criterion is well formed and identified. `spec
   check` settles the machine-checkable half: an untagged or malformed
   identifier, a duplicate one, a criterion whose code does not derive
   from its feature's, an empty scenario, prose in a `Scenario`
   description, and partially tagged `Examples` blocks. It reports the
   findings under `gherkin/*` rule identifiers.

   Three parts are yours, because the tool cannot read intent:

   - **An `Examples` block that should have been its own criterion.**
     Tag the scenario when you meant several criteria and all but the
     first cease to exist, in a document that parses cleanly and against
     which no check fires. Only *partial* tagging is machine-detectable.
   - **A prohibition with no `@negative`, or a hand-verified criterion
     with no `@manual`.** Gherkin has no grammatical negative, so the tag
     is the whole of the marking.
   - **A step shortened to fit a margin.** No width limit applies inside
     a fence; compressing a step to satisfy one silently changes what the
     criterion says.

---

## Running the checklist

Run it before setting a document to `ready-for-review`, fix what it
reports, then run it again — a fix for one check often breaks another
(rewrapping a line for check 6 can bury the compound requirement that
check 5 looks for). Re-run `specs check <path>` on each pass too: it is
one subprocess, and an edit made to satisfy check 5 can break check 1
by disturbing the frontmatter above it. Do not set `ready-for-review`
while any check fails.

## References

- `spec-cli.md` — the CLI that settles checks 1–3, 8 and most of 9, and
  what it does not yet check
- `template-authoring-standards.md` (same directory) — frontmatter
  properties, slug conventions, line width, instruction blocks
- `frontmatter-standards.md` — the `status` controlled vocabulary
- `cross-document-references.md` — how a `[[slug]]` resolves
- `ears-guide.md` — the EARS patterns checks 4 and 5 test for
- `gherkin-guide.md` — the tag vocabulary, description slots and silent
  failures check 9 tests for
- `blueprint-readiness-checklist.md` — the separate groomability gate
  that feature blueprints also run
