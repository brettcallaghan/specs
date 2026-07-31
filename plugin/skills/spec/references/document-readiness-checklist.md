---
title: Document Readiness Checklist
description: The type-independent self-check every spec document passes
  before it is set to ready-for-review — frontmatter, slug, status, EARS
  conformance, line width, instruction blocks, reference resolution
slug: document-readiness-checklist
type: standards
status: active
updated_at: 2026-07-30T00:00:00Z
version: "0.1"
related:
  - "template-authoring-standards"
  - "cross-document-references"
---

# Document Readiness Checklist

## Contents

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
   and the notation it must use.

---

## Running the checklist

Run it before setting a document to `ready-for-review`, fix what it
reports, then run it again — a fix for one check often breaks another
(rewrapping a line for check 6 can bury the compound requirement that
check 5 looks for). Do not set `ready-for-review` while any check
fails.

## References

- `template-authoring-standards.md` (same directory) — frontmatter
  properties, slug conventions, line width, instruction blocks
- `frontmatter-standards.md` — the `status` controlled vocabulary
- `cross-document-references.md` — how a `[[slug]]` resolves
- `ears-guide.md` — the EARS patterns checks 4 and 5 test for
- `blueprint-readiness-checklist.md` — the separate groomability gate
  that feature blueprints also run
