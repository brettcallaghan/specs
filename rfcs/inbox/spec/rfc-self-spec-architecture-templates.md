---
title: "RFC: Three architecture templates still prescribe retired @
  notation; the diagram parse-check does not work as documented"
description: The container, module and contract templates instruct
  authors to use @ notation that ADR-012 retired and a standard
  prohibits; the recommended mermaid parse-check fails without an
  undocumented DOM
slug: rfc-self-spec-architecture-templates
type: rfc
status: ready-for-review
updated_at: 2026-07-31T00:00:00Z
origin: self
skill: spec
skill_version: "1.5"
model: claude-opus-5
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.10.0"
related: []
---

# RFC: Retired @ notation in three templates; the parse-check does not work

## Summary

The `spec` skill was run once to author an application-level
architecture set: `container-spec-cli`, `module-document-core`,
`module-spec-profile` and `contract-specs-configuration`. Two defects
surfaced, both in the reference bundle rather than in `SKILL.md`. Three
templates instruct the author to reference documents with `@` notation
that [[adr-012]] retired and `cross-document-references.md` prohibits
with a SHALL NOT. Separately, the diagram parse-check the C4 guide
recommends fails on every valid diagram unless the author supplies a DOM
the guide does not mention. The thesis is that the architecture bundle
has not been swept since the reference-conventions decision landed, and
that its one runnable recommendation was never run.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code on the web (remote execution environment) |
| Task the skill was applied to | Authoring an application-level architecture set: one container blueprint, two module blueprints, one contract — 4 documents, 4 mermaid diagrams, 13 declared components |
| References / scripts / templates invoked | `container-blueprint-template.md` v1.3, `module-blueprint-template.md` v0.5, `contract-template.md` v1.2, `c4-guide.md` (**carries no version field** — `slug: guide-c4`, `status: draft`), `document-readiness-checklist.md` v0.1, `cross-document-references.md` v0.1, `template-authoring-standards.md` v0.12 — all under `plugin/skills/spec/references/` |
| Run dates | 2026-07-31 |

`effort` is recorded as `unknown` because the reasoning-effort setting
is not exposed to the running agent in this harness.

`c4-guide.md` has no `version` property, so this RFC cannot cite one for
the document that supplied both findings' surrounding context. That is
the same defect RFC 0001 recorded against `ears-guide.md` and set aside
as a direct fix rather than an RFC proposal; it recurs here.

## Evidence

1. **[misleading-text · blocker]** Three templates in the architecture
   bundle instruct the author to reference documents with `@` notation:
   - `plugin/skills/spec/references/container-blueprint-template.md:323`
     — "`- @<standard-slug> — <one-line scope>.`"
   - `plugin/skills/spec/references/module-blueprint-template.md:317`
     — "`- @<standard-slug> — <one-line scope>.`"
   - `plugin/skills/spec/references/contract-template.md:108`
     — "`@<provider-slug>. Visibility: <public | internal>.`"
   - `plugin/skills/spec/references/contract-template.md:202`
     — "`- @<provider-slug> — provider's blueprint`"

   All four contradict
   `plugin/skills/spec/references/cross-document-references.md:47-49`:
   "At-sign (`@`) notation SHALL NOT be used: in a coding agent's
   environment `@` denotes a filesystem mention resolved against the
   session root, which a slug is not." The decision behind that
   prohibition is [[adr-012]], which is `ready-for-review` and titled
   "Wikilink and Anchor Reference Conventions — … replaces `@` and the
   `#` element convention".

   **What happened:** authoring all four documents, the agent had to
   notice the contradiction and deviate from three templates, writing
   `[[adr-015]]`, `[[container-spec-cli]]` and backticked paths where
   the templates showed `@`. The Cross-Cutting Standards section of both
   blueprints and the Overview and References sections of the contract
   were each affected.

   **Why this is the same defect class as an already-adopted fix:** RFC
   0006 P1 corrected `adr-template.md`, where a slug instruction
   contradicted the prefix registry, and recorded that the template's
   form had won 11 times out of 14. The mechanism is identical — a
   template prescribing what a standard prohibits — and it survives in
   three more templates that P1 did not touch.

2. **[gap · snag]** `plugin/skills/spec/references/c4-guide.md:307-311`
   recommends a verification step that does not work as described:
   "Parse-check diagrams before finishing (RECOMMENDED) — via
   `mermaid.parse` in a small script, or any Mermaid-rendering preview.
   A broken C4 diagram reports `Expecting 'RBRACE', got 'EOF'` anchored
   at the end of the diagram, far from the offending line, so authoring
   blind is risky."

   **What happened:** the agent followed this literally, importing
   `mermaid` and calling `mermaid.parse` on each of the four diagrams.
   All four failed:

   ```
   FAIL  container-spec-cli.md [0] Container Context — Spec CLI
         DOMPurify.addHook is not a function. (In 'DOMPurify.addHook(…
   FAIL  container-spec-cli.md [1] Components — Spec CLI
         DOMPurify.addHook is not a function. …
   FAIL  module-document-core.md [0] Components — Document Core
         DOMPurify.addHook is not a function. …
   FAIL  module-spec-profile.md [0] Components — Spec-Driven Development Profile
         DOMPurify.addHook is not a function. …

   4 diagrams, 4 failed
   ```

   Expected: four passes, since all four diagrams are valid. Observed:
   four failures for an environmental reason. `mermaid.parse` requires a
   DOM even though it renders nothing. The agent had to add `jsdom` and
   assign `window`, `document`, `navigator`, `Element`, `HTMLElement`
   and `SVGElement` onto `globalThis` before importing mermaid, after
   which the same four diagrams reported `4 diagrams, 0 failed`.

   **Repro:** `bun add mermaid`, then import mermaid and call
   `mermaid.parse` on any valid C4 source with no DOM present.

   **Why this matters beyond one round-trip:** the surrounding sentence
   primes the reader to expect a confusing error located far from its
   cause. An agent that takes `DOMPurify.addHook is not a function` as
   that predicted confusing error would begin editing four correct
   diagrams.

3. **[gap · snag]** No rule at `spec` SKILL.md step 1 (`:188`) resolves
   whether the per-type subdirectory is per document *type* or per
   document *family*. The step says to write the document "in a
   subdirectory for its type (`docs/adrs/`, `docs/frds/`,
   `docs/standards/`, `docs/blueprints/`, …)". Three of the four
   enumerated directories are per type; `docs/blueprints/` is per
   family, covering four distinct types (system, container, component,
   module) plus feature blueprints.

   **What happened:** the agent placed the container blueprint and both
   module blueprints in `docs/blueprints/`, following the family
   reading, and created `docs/contracts/` for the contract, following
   the type reading — two different readings of one rule inside one
   run. A second agent could as defensibly produce `docs/containers/`
   and `docs/modules/`, or place the contract in `docs/blueprints/`.

   **Recurrence:** this is the second occurrence of this class in two
   runs of `spec` in this repository. RFC 0006 P5, adopted, added
   `docs/frds/` to the same enumeration — which fixed the missing entry
   but left the type-versus-family rule unstated, so the next new
   document type hits it again.

## Proposals

### P1 — Replace @ notation in the three architecture templates [blocker]

**Problem.** `container-blueprint-template.md:323`,
`module-blueprint-template.md:317`, and `contract-template.md:108` and
`:202` all show `@<slug>` reference syntax. That syntax is prohibited by
a SHALL NOT in `cross-document-references.md:47-49` and was retired by
[[adr-012]]. An author following any of these three templates produces a
reference that violates the framework's own standard, and — because `@`
resolves as a filesystem mention in a coding agent's environment — one
that means something else entirely to the agent reading it (finding 1).

**Change.** Four substitutions, each replacing `@` with the wikilink
form the standard requires:

- `container-blueprint-template.md:323`:
  `- @<standard-slug> — <one-line scope>.`
  → `- [[<standard-slug>]] — <one-line scope>.`
- `module-blueprint-template.md:317`: the same substitution.
- `contract-template.md:108`:
  `@<provider-slug>. Visibility: <public | internal>.`
  → `[[<provider-slug>]]. Visibility: <public | internal>.`
- `contract-template.md:202`:
  `- @<provider-slug> — provider's blueprint`
  → `- [[<provider-slug>]] — provider's blueprint`

Bump each instruction block's `Version`: container-blueprint 1.3 → 1.4,
module-blueprint 0.5 → 0.6, contract 1.2 → 1.3. All three are PATCH-
scale corrections but the framework versions instruction blocks with a
single integer pair, so MINOR is the available step.

Triage should also sweep the remaining templates for `@` rather than
trusting this list: these four are the occurrences the architecture run
touched, not necessarily every one in the bundle.

**Evidence.** Finding 1. Under the same model, an agent authoring any
architecture document produces conforming references without having to
notice that a template and a standard disagree — the same outcome RFC
0006 P1 achieved for `adr-template.md`.

### P2 — State the DOM requirement in the parse-check guidance [snag]

**Problem.** `c4-guide.md:307-311` recommends `mermaid.parse` in a small
script and omits that mermaid requires a DOM to parse. Followed
literally, it reports every valid diagram as failing (finding 2), in a
context that has just told the reader to expect a confusing,
mislocated error.

**Change.** Replace the parse-check bullet at `c4-guide.md:307-311`
with:

```
- Parse-check diagrams before finishing (RECOMMENDED). A broken C4
  diagram reports `Expecting 'RBRACE', got 'EOF'` anchored at the end
  of the diagram, far from the offending line, so authoring blind is
  risky.
  - `mermaid.parse` needs a DOM even though it renders nothing: import
    it after assigning a jsdom `window` and `document` onto
    `globalThis`, or the parse fails with `DOMPurify.addHook is not a
    function` on every diagram, valid or not.
  - Any Mermaid-rendering preview works too and needs no setup.
```

`c4-guide.md` has no `version` property to bump. Adding one is a
separate fix (see Considered and Not Proposed).

**Evidence.** Finding 2. Under the same model, an agent that decides to
parse-check gets a working check on the first attempt instead of four
false failures, and cannot mistake the environmental error for the
mislocated syntax error the same bullet describes.

### P3 — State whether the subdirectory is per type or per family [snag]

**Problem.** Step 1's enumeration mixes both readings — three per-type
directories and one per-family — and states neither rule, so one run
produced both (finding 3). RFC 0006 P5 added a missing entry to this
same enumeration without stating the rule, so the enumeration grows by
one each time a new type is authored.

**Change.** At `plugin/skills/spec/SKILL.md:188-194`, replace the
parenthetical enumeration with a stated rule:

```
   …in a subdirectory named for its type — `docs/adrs/`, `docs/frds/`,
   `docs/standards/`, `docs/contracts/` — except the four architecture
   blueprint types (system, container, component, module) and feature
   blueprints, which share `docs/blueprints/` because they are read
   together. WHEN the project has an established location for a type,
   follow it…
```

Bump `SKILL.md` frontmatter `version: "1.5"` → `"1.6"` (MINOR).

**Evidence.** Finding 3. Under the same model, an agent authoring a
contract and a container blueprint in one run places both without a
judgment call, and the next new document type resolves from the rule
rather than from a longer list.

## Considered and Not Proposed

- **`c4-guide.md` carries no `version` property.** Real — this RFC
  cannot cite a version for the document behind two of its three
  findings, and `template-authoring-standards` requires versions on
  versioned standards. Not proposed here because it is a defect in that
  document rather than friction the run hit, and it is a one-line direct
  fix. Identical to the `ears-guide.md` item RFC 0001 recorded and set
  aside the same way — which makes it the second occurrence, and worth
  a sweep of all guides rather than two one-line fixes.
- **The contract template's Operational Expectations assumes a service
  interface.** `contract-template.md:159-164` prescribes availability,
  latency, rate limits and authentication. For a configuration file
  format — a contract type the `c4-guide.md` mapping table admits under
  "data product / published view" — four of those five are meaningless,
  and the run wrote "availability, latency, rate limits and
  authentication do not apply" to fill the section. Not proposed
  because stating N/A is a reasonable outcome, the section still earned
  its keep through consumer and provider obligations, and one run
  authoring one non-service contract is thin evidence for changing a
  template used mostly for services.
- **The contract template has one `visibility` field, but the authored
  contract spans two tiers.** `contract-specs-configuration` documents a
  public `specs.toml` surface and an internal registration surface. The
  run set `visibility: public` and marked the internal half inline. Not
  proposed because the framework's position — one tier per contract, so
  a two-tier surface is two contracts — is defensible, and the run has
  no evidence that it is wrong. The tension is recorded in that
  document's Open Questions where a reviewer will see it.
- **Both module blueprints carry `version: "0.1"` while
  `visibility: internal`.** `module-blueprint-template.md:20` marks
  `version` "required WHEN visibility: public" and does not prohibit it
  otherwise. The run set it anyway. Author's choice, not a template
  defect.

## Non-Goals

- Changing the `@`-to-wikilink decision itself. [[adr-012]] settled it;
  P1 propagates it into templates that were missed.
- Changing where this repository's existing architecture documents
  live. P3 states the rule going forward; the four documents authored in
  this run already follow the reading it proposes.
- Auditing every template for `@` notation. P1 fixes the four
  occurrences this run touched and asks triage to sweep the rest, rather
  than asserting a completeness this run did not verify.

## Adoption

P1 touches three files under `references/` and is independent of the
others. P2 touches `c4-guide.md` alone. P3 touches `SKILL.md` alone.
None blocks another, and all three are single-substitution changes.

P1 is the urgent one: it is actively producing non-conforming references
in every architecture document authored from these templates, and the
same defect class has already been observed to win 11 times out of 14 in
`adr-template.md` before RFC 0006 P1 corrected it. P2 and P3 each cost
one round-trip per run rather than producing a wrong artifact.

Backward compatibility: P1 changes only instruction text and placeholder
syntax, so documents already authored are unaffected — though any
existing architecture document that followed the old templates now
carries prohibited `@` references and would fail a reference check. This
repository has none outside the four authored in this run, all of which
use the wikilink form.

Expected effect: an architecture document authored after P1 needs no
deviation from its template; a diagram parse-check after P2 works on the
first attempt; and a document type introduced after P3 resolves its
location from a rule rather than from an enumeration that has to grow.

## Resolution
