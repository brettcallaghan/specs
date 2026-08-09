---
title: Document Readiness Checklist
description: The type-independent self-check every spec document passes
  before it is set to ready-for-review — frontmatter, slug, status, EARS
  conformance, line width, instruction blocks, reference resolution
slug: document-readiness-checklist
type: standards
status: active
updated_at: 2026-08-09T00:00:00Z
version: "0.4"
---

# Document Readiness Checklist

## Contents

- Who decides — which checks the `spec` CLI settles and which are yours
- Frontmatter checks #1–#3 — does the metadata conform? (required
  properties, slug, status vocabulary)
- Normative-statement checks #4–#5 — are the requirements well formed?
  (EARS coverage, compound modal verbs)
- Finished-document checks #6–#8 — is the document finished? (line
  width, instruction blocks, reference resolution)

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

## Who decides

Run `spec check <path>` first. It settles checks 1–3 and 8 outright: it
parses the frontmatter, validates every property against the schema for
the document's `type`, matches the slug against that type's registered
pattern, reports a slug already claimed elsewhere in the corpus, and
resolves every `[[slug]]` reference — in the body and in the frontmatter
properties that hold slugs — against the corpus. Read its findings; each
names the rule, the span, and the value that violated it.

**A single-document check reports the links your edit broke — if it
broke them outright.** `spec check <path>` resolves that document's own
references against the whole corpus, and also reports every *other*
document whose `[[slug]]` links no longer resolve, each at its own line:
rename an ADR's slug, check that one file, and it names the documents
left citing the old one. `reference/unresolved` is the one rule the
narrow form does not scope to the file you named, because the document
that breaks a reference is never the document that holds it.

**Retiring a document is not covered by that.** A live document citing a
retired one is `reference/retired-target`, and that rule *is* still
scoped to the file you name — so checking the document you just retired
reports nothing and prints `reference/retired-target passed`. Changing a
document's `type` behaves the same way for `reference/type-mismatch`.
After either edit, run the corpus-wide `spec check`; the narrow form
will not tell you.

What no form of the check can tell you is whether a broken reference is
*yours*. It has no memory of what those references used to resolve to,
so breakage already in the corpus is reported alongside breakage your
edit caused. In a corpus you keep green that distinction is empty —
anything reported is new. `spec refs <slug>` before you change, retire
or re-type a document answers the blast-radius question ahead of time,
and is the cheaper habit.

A `frontmatter/parse` finding (row 0 below) comes before all of them and
is not one of the eight: it means the frontmatter could not be read as a
document envelope, so no other check examined this document. Fix it and
re-run before reading anything else in the report. The usual cause is a
template placeholder left unfilled — `updated_at:
<YYYY-MM-DDTHH:MM:SSZ>` is not a date.

| # | Check | Decided by |
|---|-------|------------|
| 0 | Frontmatter is readable at all | `spec check` |
| 1 | Required properties present and parseable | `spec check` |
| 2 | Slug matches the pattern for this type; unique | `spec check` |
| 3 | Status is in the controlled vocabulary | `spec check` |
| 4 | Every normative statement uses an EARS pattern | you |
| 5 | No requirement joins two modal verbs | you |
| 6 | No prose line exceeds the width limit | you — CLI planned |
| 7 | No template instruction blocks survive | you |
| 8 | Every `[[slug]]` reference resolves | `spec check` |

**A clean `spec check` is not a passing document.** It means row 0 and
checks 1–3 and 8 passed. Checks 4, 5 and 7 are judgement the tool does
not make, and 6 is specified but not built — a silent tool is not a
passing grade for them.

Check 8 has one part the CLI still leaves to you: it resolves the slug
half of a reference and does not interpret a **heading anchor**, so
`[[container-x#Some Heading]]` passes on `container-x` alone. If you
wrote an anchor, confirm the heading exists. `spec-cli.md` lists the
other reference cases a clean run says nothing about.

WHERE the CLI is not installed, every check below is still defined and
still runs; you carry out 1–3 by hand instead, against the property
definitions and prefix registry the checks cite. Say which happened —
"`spec check` reports no findings" and "I read the frontmatter and it
looks right" are different claims, and only the first can be re-run by a
reviewer. `spec-cli.md` covers installation.

---

## Frontmatter checks

1. The frontmatter parses as YAML and declares every required property:
   title, description, slug, type, status and updated_at. The property
   definitions are owned by [[template-authoring-standards]].
2. The slug follows the registry rule for this document's type — the
   Slug Conventions of [[template-authoring-standards]], including the
   type-prefix format and the standards type's role-noun exception.
3. The status is a value from the controlled vocabulary, which
   [[template-authoring-standards]] defines by pointing at
   [[frontmatter-standards]]. A status outside that vocabulary is an
   error, not a local convention.

## Normative-statement checks

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
   and the notation it must use. `spec check` decides this one; run it
   rather than reading the links. Two of its findings are worth reading
   rather than only counting: a `reference/retired-target` **warning**
   means the link resolves but the target has been replaced — cite the
   successor it names instead — and a `reference/type-mismatch` means
   the property resolved to a document of the wrong type.

---

## Running the checklist

Run it before setting a document to `ready-for-review`, fix what it
reports, then run it again — a fix for one check often breaks another
(rewrapping a line for check 6 can bury the compound requirement that
check 5 looks for). Re-run `spec check <path>` on each pass too: it is
one subprocess, and an edit made to satisfy check 5 can break check 1
by disturbing the frontmatter above it. Do not set `ready-for-review`
while any check fails.

## References

- `spec-cli.md` — the CLI that settles checks 1–3 and 8, and what it
  does not yet check
- `template-authoring-standards.md` (same directory) — frontmatter
  properties, slug conventions, line width, instruction blocks
- `frontmatter-standards.md` — the `status` controlled vocabulary
- `cross-document-references.md` — how a `[[slug]]` resolves
- `ears-guide.md` — the EARS patterns checks 4 and 5 test for
- `blueprint-readiness-checklist.md` — the separate groomability gate
  that feature blueprints also run
