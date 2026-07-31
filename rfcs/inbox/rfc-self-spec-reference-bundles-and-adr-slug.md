---
title: "RFC: Fix the ADR slug contradiction; bundle the always-required
  references"
description: The ADR template's slug format contradicts the prefix
  registry and has produced 11 non-conforming slugs; steps 4 and 7
  mandate two references that no authoring bundle loads
slug: rfc-self-spec-reference-bundles-and-adr-slug
type: rfc
status: ready-for-review
updated_at: 2026-07-31T00:00:00Z
origin: self
skill: spec
skill_version: "1.4"
model: claude-opus-5
effort: unknown
runs_observed: 2
severity: blocker
plugin_version: "0.9.0"
related: []
---

# RFC: Fix the ADR slug contradiction; bundle the always-required references

## Summary

The `spec` skill was run twice in one session: once to author
`docs/adrs/ADR-014-gherkin-for-acceptance-criteria.md`, once to author
`docs/frds/FRD-spec-validation-cli.md`. Both runs hit the same class of
friction — the skill's authoring bundles omit references that its own
workflow steps make mandatory, and one template contradicts the standard
it is supposed to implement. The ADR contradiction is not hypothetical:
it has already produced 11 non-conforming slugs and 26 dangling
references in this repository, tracked as bead `agent-plugins-5nz`. The
thesis is that both defects are the same shape — a rule that lives in a
document the authoring agent was never told to load.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code on the web (remote execution environment) |
| Task the skill was applied to | Authoring ADR-014 (a decision on Gherkin adoption), then FRD-spec-validation-cli (9 requirements, 58 acceptance criteria) |
| References / scripts / templates invoked | `adr-template.md` v1.2, `frd-template.md` v1.1, `ears-guide.md` (`guide-ears`, no version), `document-readiness-checklist.md` v0.1, `cross-document-references.md` v0.1, `template-authoring-standards.md` v0.12 — all under `plugin/skills/spec/references/` |
| Run dates | 2026-07-31 (both runs) |

`effort` is recorded as `unknown` because the reasoning-effort setting is
not exposed to the running agent in this harness. It is not omitted by
oversight; the value could not be observed.

## Evidence

1. **[misleading-text · blocker]**
   `plugin/skills/spec/references/adr-template.md:4`
   "`slug: adr-<nnn>-<short-description>`", reinforced at `:51`
   "`Use lowercase kebab-case: adr-<number>-<short-description>`" and
   `:52` "`Example: adr-003-local-tesseract-for-ocr`".
   This contradicts the prefix registry at
   `plugin/skills/spec/references/template-authoring-standards.md:130`,
   which registers the ADR slug form as `adr-<nnn>` with the example
   `[[adr-011]]`.
   **What happened:** authoring ADR-014, the agent had to notice the
   contradiction and deliberately deviate from the template, choosing
   `slug: adr-014` over the template's stated form.
   **What happened previously, without the deviation:** 11 of the 14
   ADRs in `docs/adrs/` carry the template's form
   (`adr-001-html-comments-for-instructions` through
   `adr-011-targeted-verification-gates`); only `adr-012`, `adr-013` and
   this run's `adr-014` carry the registry form. Every citation in the
   repository uses the bare number, so 26 references dangle. Filed as
   `agent-plugins-5nz`.

2. **[gap · blocker]** No rule at `spec` SKILL.md step 4
   (`plugin/skills/spec/SKILL.md:207`) or step 7 (`:229`) puts the
   references those steps mandate into any authoring bundle. Step 4
   requires "`[[document-slug]]` cross-references"; step 7 requires
   running `references/document-readiness-checklist.md`. The References
   table's authoring rows load neither: `:254` "Recording an
   architecture or technology decision | `references/adr-template.md`"
   and `:255` "Specifying a feature (user stories + acceptance criteria)
   | `references/frd-template.md` + `references/ears-guide.md`".
   **What happened:** authoring ADR-014, the agent wrote out-of-set
   wikilinks — `[[ac-traceability-standards]]` (4×), `[[guide-ears]]`
   (3×) and `[[document-readiness-checklist]]` (1×) — 8 occurrences
   across 3 targets. All violate
   `plugin/skills/spec/references/cross-document-references.md:124-125`:
   "The `[[document-slug]]` notation SHALL be reserved for documents in
   the specification set; it SHALL NOT be used for documents outside
   it." The error was found and corrected only after loading
   `cross-document-references.md` at the second hop, mid-authoring.
   **Why this is structural rather than carelessness:** the skill
   already names this failure mode at `:272` — "Read the bundle up
   front rather than following a pointer out of a template
   mid-authoring — a reference reached at the second hop tends to get
   previewed rather than read" — but applies it only to the table's
   rows, not to the two references its own workflow steps make
   unconditional.

3. **[gap · polish]** No rule at `spec` SKILL.md step 1 (`:188`) names a
   directory for FRDs. The step enumerates "(`docs/adrs/`,
   `docs/standards/`, `docs/blueprints/`, …)" while FRD is the second
   type in the taxonomy table at `:74`.
   **What happened:** the agent inferred `docs/frds/` from the step's
   stated rule ("a subdirectory for its type") and created it. The
   inference was probably right, but it was an inference, and a second
   agent could as easily produce `docs/frd/` or `docs/requirements/`.

## Proposals

### P1 — Match the ADR template's slug format to the registry [blocker]

**Problem.** `adr-template.md:4`, `:51` and `:52` instruct the author to
build a slug of the form `adr-<nnn>-<short-description>`. The registry
at `template-authoring-standards.md:130` registers `adr-<nnn>`. An
author following the template produces a slug no citation will resolve
against, because every ADR citation in this framework uses the bare
number. Finding 1 shows the template's form winning 11 times out of 14.

**Change.** In `plugin/skills/spec/references/adr-template.md`:

- Line 4, replace `slug: adr-<nnn>-<short-description>` with
  `slug: adr-<nnn>`.
- Lines 50-53, replace the "Slug format" block with:

  ```
  Slug format
    - The slug is the bare type prefix and number: adr-<nnn>
    - Example: adr-003
    - The slug is independent of the filename. Name the FILE
      descriptively — ADR-003-local-tesseract-for-ocr.md — so humans can
      find it; the slug stays adr-003 so [[adr-003]] resolves.
    - See template-authoring-standards.md, Prefix Registry.
  ```

- Bump the instruction block's `Version: 1.2` to `1.3` (MINOR).

**Evidence.** Finding 1. Under the same model, the next ADR authored
from this template carries a registry-conforming slug without the author
needing to notice a contradiction between two documents, only one of
which is in the bundle. This does not itself fix the 11 existing slugs —
that is `agent-plugins-5nz` — but it stops the template producing a
twelfth.

### P2 — Bundle the two always-required references [blocker]

**Problem.** Steps 4 and 7 apply to every document type, but the
References table treats `cross-document-references.md` and
`document-readiness-checklist.md` as optional rows an agent loads only
when it already knows it needs them. An agent authoring an ADR loads one
file; an agent authoring an FRD loads two. Neither bundle contains the
rules those two steps require, so the agent either violates them
(finding 2) or discovers them at the second hop the skill itself warns
against (`SKILL.md:272`).

**Change.** In `plugin/skills/spec/SKILL.md`, insert a paragraph
immediately before the References table (currently at `:250`):

```
Two references are required for EVERY document type and are therefore
not listed per row: `references/cross-document-references.md` (step 4 —
which references use `[[slug]]` and which use backticked paths) and
`references/document-readiness-checklist.md` (step 7 — the self-check
gate). Load both alongside the row for what you are authoring.
```

Then append to the trailing note at `:272`, after "…tends to get
previewed rather than read.":

```
That applies to the two always-required references above as much as to
the rows: a reference syntax rule read after the document is written is
read too late to prevent the error.
```

Bump `SKILL.md` frontmatter `version: "1.4"` to `"1.5"` (MINOR).

**Evidence.** Finding 2. Under the same model, an agent authoring any
document type has the out-of-set reference rule in context before
writing its first reference, so the 8 out-of-set wikilinks observed in
this run do not get written. It also front-loads the readiness checklist,
which the agent currently reaches only at step 7 — after the document
exists.

### P3 — Name the FRD directory in step 1's enumeration [polish]

**Problem.** Step 1 at `SKILL.md:188` enumerates three type
subdirectories and omits FRDs, leaving the second-most-common document
type to inference (finding 3).

**Change.** At `SKILL.md:191`, replace "(`docs/adrs/`,
`docs/standards/`, `docs/blueprints/`, …)" with "(`docs/adrs/`,
`docs/frds/`, `docs/standards/`, `docs/blueprints/`, …)". Covered by
P2's version bump if they land together; otherwise bump `version` to
`"1.5"`.

**Evidence.** Finding 3. The next agent authoring an FRD in a repository
with no established location writes to the same path this run chose,
rather than to a plausible variant.

## Considered and Not Proposed

- **The readiness checklist has no runnable implementation, so check 8
  costs a corpus-wide read.** Real friction — running check 8 by hand in
  this session required writing throwaway scripts — but it is the
  subject of `docs/frds/FRD-spec-validation-cli.md`, authored in this
  same run. Filing it here would duplicate that FRD, and the checklist's
  own text is not defective.
- **The 80-character limit is unsatisfiable on a line carrying a long
  backticked path.** Observed while converting finding 2's wikilinks: a
  47-character path plus surrounding prose cannot fit. Not a defect —
  `template-authoring-standards.md:47` already exempts lines where
  "line breaks would break functionality", and the run resolved it by
  restructuring the sentence instead. No change proposed.
- **`frd-template.md:69-71` marks manually-verifiable criteria with
  "(manual)".** Not exercised in this run — no criterion needed it — so
  there is no evidence either way.
- **The FRD template's example guidance.** `frd-template.md:139-149`
  says each criterion SHOULD carry one example and permits omission
  where the criterion is unambiguous. 18 of 58 criteria were written
  without one under that permission, and the guidance was not ambiguous
  in the run. No finding.

## Non-Goals

- Fixing the 11 existing non-conforming ADR slugs. P1 stops the template
  producing more; correcting the existing ones is `agent-plugins-5nz`.
- Changing what `cross-document-references.md` requires. The rule at
  `:124-125` is correct; P2 changes only when the authoring agent meets
  it.
- Restructuring the References table into bundles per workflow step.
  P2 adds a paragraph rather than reorganizing, because the table's
  row-per-task shape works for everything except the two unconditional
  references.

## Adoption

P1 and P2 are independent and touch different files —
`references/adr-template.md` and `SKILL.md` respectively. P3 touches the
same file as P2 and is naturally batched with it under one version bump;
if P2 is rejected, P3 still stands alone.

P1 is the more urgent of the two blockers: it is actively producing
defects (11 occurrences, 26 dangling references) and its fix is three
lines. P2 is preventive and its evidence is a single run.

Backward compatibility: P1 changes the slug an author will produce for
*new* ADRs, which means new ADRs stop matching the pattern of the 11
existing ones until `agent-plugins-5nz` lands. That inconsistency
already exists (`adr-012` and `adr-013` predate this RFC) and resolving
it in the other direction — changing the registry to match the template
— would invalidate 26 existing citations and the ADR-012 convention.

P2 adds roughly 1,700 tokens to every `spec` authoring run that does not
already load the two references. That cost is real and is the reason the
proposal is a pointer paragraph rather than inlining the rules.

Expected effect: an ADR authored after P1 needs no deviation from its
template, and a document of any type authored after P2 has the reference
syntax rule and the readiness gate in context before it is written
rather than after.

## Resolution
