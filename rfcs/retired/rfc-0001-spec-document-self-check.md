---
title: "RFC: A document self-check, and a home for non-blueprint documents"
description: The authoring workflow gates feature blueprints only, so a
  standards document shipped with EARS violations the same skill mandates
slug: rfc-0001-spec-document-self-check
type: rfc
status: retired
origin: self
skill: spec
skill_version: "1.3"
model: claude-opus-5
effort: high
runs_observed: 1
severity: blocker
plugin_version: "0.7.0"
---

# RFC: A document self-check, and a home for non-blueprint documents

## Summary

`spec` v1.3 was run to author a standards document
(`docs/standards/skill-best-practices.md`). The skill's authoring workflow
has exactly one verification step, and it applies only to feature
blueprints — so a standards document can be completed, and would have been
shipped, carrying three violations of the EARS rule the same skill
mandates two steps earlier. The workflow is also silent on where a new
document goes and on whether a document's slug must match its filename.
The thesis: `spec` gates the one document type that has a downstream
consumer enforcing a gate, and leaves every other type ungated.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code, remote (web) session |
| Task the skill was applied to | Authoring a standards document capturing skill-authoring requirements learned from a prior review |
| References / scripts / templates invoked | `standards-template.md` v1.1, `ears-guide.md` (status `draft`, unversioned), `template-authoring-standards.md` v0.12 — all `plugin/skills/spec/references/` |
| Run dates | 2026-07-30 |

## Evidence

1. **[gap · blocker]** No rule at `spec` authoring-workflow step 6 for any
   document type other than a feature blueprint.
   `plugin/skills/spec/SKILL.md:210` reads "**Preflight feature
   blueprints.** Before setting a feature blueprint to
   `ready-for-review`, run the decision sweep and self-check against
   `references/blueprint-readiness-checklist.md`". Steps 1–5 and 7 contain
   no verification instruction. Following the workflow literally, the
   agent produced a finished standards document containing three compound
   requirements — each joining two `SHALL` clauses with "and" — which
   `plugin/skills/spec/references/ears-guide.md:125` explicitly forbids:
   "**One requirement per sentence.** Never join two SHALLs with 'and.'"
   The violations were caught only because the agent independently wrote
   an ad-hoc parser to count modal verbs per paragraph; nothing in the
   skill asked for that. The three statements as authored were:

   - "A bundled script SHALL handle its own error conditions and SHALL NOT
     defer them to the calling agent."
   - "WHERE a skill prescribes a validator, it SHALL instruct the agent to
     fix what the validator reports and re-run it, and SHALL NOT permit
     proceeding while it fails."
   - "WHERE two skills must share a reference document, the shared file
     SHALL be authored once and vendored into each consumer by a
     generator, and the generated copies SHALL be verifiable against their
     source."

2. **[gap · snag]** No rule at `spec` authoring-workflow steps 1–2 for
   where a new document is written. `SKILL.md:188` says "**Pick the
   template** from `references/` (table below) and copy it as the starting
   point for the new document" and `SKILL.md:190` covers the embedded
   instructions; neither states a destination path. The standards
   template's own instruction block specifies the slug
   (`standards-template.md:29`) but not the location. The agent took the
   path from the human's request; absent one it would have guessed.

3. **[gap · snag]** No rule for the relationship between a document's
   `slug` and its filename. `standards-template.md:29` requires "a
   descriptive slug ending in an approved role noun: -standards, -policy,
   -guidelines, or -constraints". The requested filename was
   `skill-best-practices.md`, which ends in none of them. Resolving this
   required loading `template-authoring-standards.md` — a document NOT in
   the bundle the References table prescribes for standards work
   (`SKILL.md:244` lists only `standards-template.md` +
   `ears-guide.md`) — to find that "A reference resolves against the
   target document's frontmatter `slug` — never a file path." The agent
   spent a hop outside the prescribed bundle to answer a question the
   template raised.

4. **[gap · polish]** The section list in `standards-template.md:15-18`
   reads:

   ```
   Sections (complete each, then delete its instruction block)
     - <Major Category>
     - <Another Major Category>
     - References
   ```

   For a template whose sections are author-defined, the list enumerates
   placeholders and conveys nothing. It contrasts with fixed-section
   templates (ADR: Context / Decision / Alternatives Considered /
   Consequences / References), where the same list is a genuine map.

## Proposals

### P1 — Add a document self-check step that applies to every type [blocker]

**Problem.** The authoring workflow's only verification step
(`SKILL.md:210`) is scoped to feature blueprints, because that is the type
with a downstream consumer (`groom-feature-blueprint`) that enforces a
gate. Every other type — ADR, FRD, standards, contract, the architecture
blueprints — is authored with no check at all. Finding 1 shows the
consequence: the skill mandates EARS at step 3 and then never asks whether
the finished document complies, so a document violating its own mandated
syntax reaches `ready-for-review`.

**Change.** Insert a new step before the retro offer, and add a matching
reference file `references/document-readiness-checklist.md` holding the
type-independent checks:

> **N. Self-check the finished document.** Before setting any document to
> `ready-for-review`, run `references/document-readiness-checklist.md`
> against it. Fix what it reports and re-run; do not set
> `ready-for-review` while any check fails. (Feature blueprints
> additionally run the blueprint readiness checklist in step 6 — that one
> is about groomability, this one about conformance.)

The checklist file's contents, all mechanically verifiable:

```
1. Frontmatter parses as YAML and declares title, description, slug,
   type, status, updated_at.
2. The slug matches the registry's rule for this type.
3. status is a value from the controlled vocabulary.
4. Every normative statement uses an EARS pattern.
5. No requirement joins two modal verbs with "and".
6. No prose line exceeds 80 characters (tables, URLs, fences exempt).
7. Every template instruction comment block has been deleted.
8. Every [[slug]] reference resolves to an existing document.
```

**Evidence.** Finding 1. Under the same model, check 5 turns the three
compound requirements from something an agent has to think to look for
into something the workflow asks for by name. Checks 1, 6 and 7 are the
same defects this repository has already hit in other files, so the
checklist is not speculative.

### P2 — State where a document is written [snag]

**Problem.** Finding 2: steps 1–2 name the template and the embedded
instructions but never the destination, so location is decided by whoever
happens to have an opinion.

**Change.** Extend `SKILL.md:188` step 1:

> 1. **Pick the template** from `references/` (table below) and copy it as
>    the starting point for the new document. **Write it under the
>    project's spec root** — `docs/` unless the project's `AGENTS.md` or
>    `CLAUDE.md` names another — in a subdirectory for its type
>    (`docs/adrs/`, `docs/standards/`, `docs/blueprints/`, …). WHEN the
>    project has an established location for that type, follow it; the
>    frontmatter index does not depend on location, but humans do.

**Evidence.** Finding 2. Removes a guess from every authoring run in a
project that has not already established a convention.

### P3 — Say that the slug is independent of the filename [snag]

**Problem.** Finding 3. The standards template raises a slug constraint
that appears to conflict with a requested filename, and the document that
resolves the conflict is outside the prescribed bundle.

**Change.** Append to the Slug block at `standards-template.md:29-32`:

> The slug is independent of the filename. A reference resolves against
> this document's frontmatter `slug`, never its path, so a file may be
> named for how humans will look for it while the slug carries the
> registry-conforming identifier.

**Evidence.** Finding 3. Answers the question where it is raised, at the
cost of one sentence, instead of a hop into a document the References
table did not send the agent to.

### P4 — Drop or replace placeholder section lists in templates [polish]

**Problem.** Finding 4: a section list made of placeholders costs tokens
and conveys nothing.

**Change.** In templates whose sections are author-defined, replace the
enumerated list with a structural note — for `standards-template.md`:

> Structure: one `##` per major requirement category, `###` per
> subcategory, then a References section. Categories are author-defined;
> name them for the domain, not for this template.

Templates with fixed sections (ADR, FRD, contract, work order) keep their
enumerated list, which is a real map.

**Evidence.** Finding 4. Narrow, and worth doing while the section lists
are being reviewed for other reasons.

## Considered and Not Proposed

- **Clarify whether step 7's retro offer covers a document authored from
  non-spec inputs.** `SKILL.md:224` triggers the offer WHEN a document
  "consumed upstream documents"; this run consumed
  `docs/plans/plan.md` and `docs/plans/plan2.md`, which are process
  artifacts, not spec types. The ambiguity
  was noticed but did not mislead — the retro was requested explicitly, so
  no wrong action followed. No run evidence; revisit if a future run
  actually skips a warranted retro because of it.
- **Add `template-authoring-standards.md` to the standards row's bundle.**
  Would have resolved finding 3, but it also enlarges a bundle the skill
  deliberately keeps small, and P3 fixes the specific gap at its source
  for one sentence. Rejected as the more expensive of two fixes.
- **Note that `ears-guide.md` carries no `version`.** Its frontmatter has
  `status: draft` and no version field, so this RFC's Provenance cannot
  cite one. Real, but it is a defect in that document rather than friction
  the run hit, and `template-authoring-standards` already requires
  versions on versioned standards — a direct fix, not an RFC.

## Non-Goals

- Changing what EARS requires. The syntax rules are correct; the gap is
  that nothing checks compliance.
- Changing the blueprint readiness checklist. It gates groomability and is
  a different concern from conformance; P1 adds a second checklist rather
  than widening that one.
- Introducing tooling. Every check in P1 is stated so an agent can run it
  by reading; a script would be a welcome follow-on but is not required
  for the proposals to land.

## Adoption

P1 is independent and carries the blocker; adopt it alone if nothing else.
P2 and P3 are independent one-paragraph edits. P4 is cosmetic and can ride
with any of them.

Files touched: `plugin/skills/spec/SKILL.md` (P1, P2), a new
`plugin/skills/spec/references/document-readiness-checklist.md` (P1),
`plugin/skills/spec/references/standards-template.md` (P3, P4). P1 and P2
change `spec`'s workflow, so its `version` bumps MINOR; P3 and P4 change
template normative content, so `standards-template.md`'s instruction-block
`Version:` bumps MINOR.

Backward compatibility: additive. Existing documents are unaffected;
existing projects gain a check they did not have. No behavior depends on
the model, so the proposals should hold across tiers — though the value of
P1 rises as model capability falls, since a less thorough agent is less
likely to invent the check on its own.

Expected effect: a document of any type reaches `ready-for-review` having
been checked against the rules its own skill mandates, rather than only
feature blueprints being checked against groomability.

## Resolution

**Triaged 2026-07-31.** 4 proposals: 3 adopted (already implemented), 1 deferred.

| Proposal | Outcome |
|----------|---------|
| P1 — document self-check step | **adopted — already implemented.** Landed before this triage as `spec` step 7 "Self-check the finished document" plus `references/document-readiness-checklist.md` (8 checks). |
| P2 — state where a document is written | **adopted — already implemented.** Landed as the destination rule in `spec` step 1. |
| P3 — slug is independent of the filename | **adopted — already implemented.** Landed in `references/standards-template.md` (Version 1.2). |
| P4 — drop/replace placeholder section lists | **deferred** → `theme:template-ergonomics`. Scope is unsettled (which of ~10 templates) and no gate strategy exists; bead is `deferred` / `blocked-ungroomable`. |

**Files changed by this triage:** none — P1–P3 were already in the tree
when triage ran, and P4 was not implemented.

**Version bumps:** none from this RFC.

**Note on vocabulary.** The skill's classification set is
adopt-now / defer / reject, and P1–P3 fit none of them cleanly: they were
approved and are present, but this triage did not implement them. They
are recorded as "adopted — already implemented" rather than forced into
"rejected — already addressed", which would wrongly imply the proposals
were unsound. The gap between those two phrasings is worth a retro.
