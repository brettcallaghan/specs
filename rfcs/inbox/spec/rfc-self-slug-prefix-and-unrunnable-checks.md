---
title: "RFC: The feature blueprint template's slug prefix is rejected by
  the CLI the skill tells you to validate with, and four readiness
  conditions have no way to be run"
description: Reconcile the fb- prefix with the pattern the shipped
  profile enforces; give the mechanically-checkable readiness conditions
  a command; correct the install path for a session inside this
  repository; and stop the shipped references modelling the wikilink rule
  the skill forbids
slug: rfc-self-slug-prefix-and-unrunnable-checks
type: rfc
status: ready-for-review
updated_at: 2026-08-07T00:00:00Z
origin: self
skill: spec
skill_version: "1.7"
model: "claude-opus-5"
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.15.0"
---

# RFC: A slug prefix the CLI rejects, and readiness checks with no runner

## Summary

One run of `spec` authored a feature blueprint and an ADR from an
approved FRD. The skill's own tooling contradicts its own template in one
place that produces an invalid document, and goes silent in three places
where a decision was forced. The blocker is concrete: following
`feature-blueprint-template.md` literally yields a slug that
`spec check` — the command the skill's step 7 tells you to run — reports
as `slug/pattern-mismatch`. The remaining proposals give the readiness
conditions a way to be run, correct the install instruction for a session
inside this repository, and stop the shipped reference files modelling
the wikilink usage `SKILL.md:218` forbids.

## Provenance

| Field | Value |
|-------|-------|
| Origin | self (this repository critiqued its own skill) |
| Skill | `spec` v1.7 |
| Task | Author a feature blueprint and one ADR from [[frd-spec-reference-integrity]], then preflight both |
| Model | `claude-opus-5` |
| Effort | not surfaced to the agent this session |
| Harness | Claude Code, remote cloud session |
| Runs observed | 1 |
| Run date | 2026-08-07 |
| References invoked | `feature-blueprint-template.md` v1.4; `adr-template.md` v1.3; `ac-traceability-standards.md` v0.2.0; `blueprint-readiness-checklist.md` v0.6.0; `document-readiness-checklist.md` v0.3; `template-authoring-standards.md`; `cross-document-references.md` v0.2; `spec-cli.md`; `c4-guide.md` |
| Scripts / tools invoked | `spec check`, `spec next-id`, `spec slug-exists` (built from `cli/` at working-tree HEAD); `bd list`, `bd show`, `bd create` |
| Plugin version | 0.15.0 |
| Documents produced | [[feature-spec-reference-integrity]], [[adr-018]] |

## Evidence

1. **[misleading-text | blocker]** `feature-blueprint-template.md:5`
   gives the frontmatter as `slug: fb-<feature-name>`, and
   `feature-blueprint-template.md:130-132` states: "Slug format /
   `fb-<feature-name>`, lowercase kebab-case, matching the FRD: /
   `frd-checkout` → `fb-checkout`". `template-authoring-standards.md:135`
   agrees: "| Feature Blueprint | `fb-` | `[[fb-guest-checkout]]` |". The
   profile the CLI ships enforces the opposite —
   `cli/profile/profile.go:105` registers
   `` `^feature-[a-z0-9]+(-[a-z0-9]+)*$` `` for the `feature-blueprint`
   type — and its own comment records the conflict: "The registry assigns
   the `fb-` prefix, but this repository's one real feature blueprint
   (`specs/features/spec-index/spec-index-blueprint.md`) is slugged
   `feature-spec-validation-cli`. The pattern here follows the document
   that exists, not the template that was not." The agent did not follow
   the template: it read the existing blueprint and the profile source
   first and chose `feature-spec-reference-integrity`. Had it followed
   the template, `spec check` — which `SKILL.md` step 7 mandates before
   `ready-for-review` — would have reported `slug/pattern-mismatch`.

2. **[gap | snag]** `document-readiness-checklist.md:67` and `:69` mark
   two of its eight checks as "you — CLI planned": "| 6 | No prose line
   exceeds the width limit | you — CLI planned |" and "| 8 | Every
   `[[slug]]` reference resolves | you — CLI planned |". Neither the
   checklist nor `spec-cli.md` gives a command. Check 8 is not a simple
   grep: a correct run must exclude `[[slug]]` occurrences inside code
   spans and fences, which is the very rule AC-SPECREF-001.2 of the FRD
   being specified exists to state. The agent wrote ad-hoc Python for
   both, three times in one session — once for the blueprint and ADR,
   once after editing them, once for the RFC produced by `retro-spec` —
   each time re-deriving its own code-span masking. The same silence
   covers `blueprint-readiness-checklist.md` checks 4 and 8 (every FRD
   AC appears in the coverage table with exactly one proving delta; the
   legend and the delta blocks are in bijection), which are equally
   mechanical and equally unserved.

3. **[gap | snag]** `spec-cli.md:103-113`, "When it is not installed",
   gives one install path: `go install
   github.com/brettcallaghan/agent-plugins/cli/spec@latest`. In this
   session `spec` was absent from `PATH` (`command -v spec` returned
   nothing; the SessionStart hook installed `bd` and failed to clone
   strata). `@latest` resolves to a published module version, but the
   run was *inside the repository that builds the CLI*, and the profile
   the CLI applies — including the slug patterns of finding 1 — lives in
   the working tree. Validating a working-tree corpus with a published
   binary can pass a document the working-tree profile rejects, or the
   reverse. The agent built from source instead
   (`cd cli && go build -o <scratch>/spec ./spec`), a step the reference
   does not mention.

4. **[gap | snag]** `SKILL.md:240` requires that "Authoring a feature
   blueprint includes surveying the project's issue tracker to populate
   its Tracker Reconciliation section", and
   `feature-blueprint-template.md` (Tracker Reconciliation instruction
   block) says "Survey the project's issue tracker while authoring"
   without naming a query, and without asking the section to record how
   it was derived. The agent ran `bd list --status=open`, read 26 titles,
   and chose 7 to open with `bd show` by keyword judgement; the resulting
   section classifies 13 issues and states no derivation. A different
   agent surveying the same tracker would plausibly classify a different
   set, and a reader cannot tell an omission from a stale survey. This is
   the second observed occurrence: RFC 0010 P4 raised the same defect
   against the previous feature blueprint and proposed a derivation line
   as a repository convention.

5. **[misleading-text | snag]** `SKILL.md:216-218` states the rule:
   "`[[slug]]` is reserved for documents in the project's spec set:
   everything under `references/` — the templates, the guides, and the
   standards shipped with this skill — is cited by backticked path,
   never by wikilink." The shipped references model the opposite. Ten
   instances across six files, including
   `blueprint-readiness-checklist.md:113` (`- [[ac-traceability-standards]]
   — the test citation obligations the AC`),
   `ac-traceability-standards.md:152` (`- [[template-authoring-standards]]
   — EARS and referencing conventions`), and
   `document-readiness-checklist.md:89`, `:91`, `:94`, `:95`, `:112`.
   The agent copied the pattern into the `retro-spec` RFC it authored in
   the same session, writing `[[ac-traceability-standards]]` and
   `[[blueprint-readiness-checklist]]`, and caught them only by running
   its own hand-written check-8 script — neither resolves, because
   neither document is in the `specs/` corpus.

## Proposals

### P1 — Settle the feature blueprint slug prefix in one place [blocker]

**Problem.** Three artifacts disagree about one string.
`feature-blueprint-template.md:5` and `:131` say `fb-`,
`template-authoring-standards.md:135` says `fb-`, and
`cli/profile/profile.go:105` enforces `^feature-…`. The CLI wins,
because `SKILL.md` step 7 makes `spec check` the gate. An agent that
trusts the template produces a document that fails that gate; an agent
that trusts the CLI silently contradicts two documents. Finding 1.

**Change.** Pick one and make the other two follow. The evidence favours
`feature-`: the only two feature blueprints that exist
(`feature-spec-validation-cli`, `feature-spec-reference-integrity`) use
it, and the profile comment already records the decision to follow the
documents that exist. Concretely, WHERE `feature-` is chosen:

- `feature-blueprint-template.md:5` → `slug: feature-<feature-name>`
- `feature-blueprint-template.md:130-133` → "Slug format /
  `feature-<feature-name>`, lowercase kebab-case, matching the FRD: /
  `frd-checkout` → `feature-checkout`"
- `template-authoring-standards.md:135` → ``| Feature Blueprint |
  `feature-` | `[[feature-guest-checkout]]` |``, and the same at `:146`
  ("feature blueprints use `fb-`") and `:166` ("→ `fb-guest-checkout`")

WHERE `fb-` is chosen instead, the profile pattern and both existing
documents' slugs change, and every reference to them is rewritten — the
larger migration, which is why `feature-` is recommended.

**Evidence.** Finding 1. The next agent authoring a feature blueprint
gets a slug its own preflight accepts, without reading the Go source of
the profile to discover which of two documented prefixes is real.

### P2 — Give the mechanically-checkable readiness conditions a command [snag]

**Problem.** `document-readiness-checklist.md:67` and `:69` say "you —
CLI planned" for checks 6 and 8, and `blueprint-readiness-checklist.md`
checks 4 and 8 are stated as conditions with no runner at all. Every
agent therefore writes its own, and check 8's correctness turns on a
code-span masking rule that is easy to get subtly wrong — an agent that
omits it reports every backticked `[[slug]]` illustration as a broken
reference. Finding 2.

**Change.** Add a `scripts/check-document.sh <path>...` to the skill,
invoked from `SKILL.md` step 7 and named in both checklists' "Decided
by" column, implementing the four conditions that are mechanical:

- **width** — report every line over `[conventions].line_width` that is
  not a table row, a fenced-block line, or a line containing a URL, per
  `template-authoring-standards.md`'s stated exemptions;
- **references** — report every `[[slug]]` that resolves to no `slug:`
  in the corpus, after masking fenced blocks and inline code spans;
- **instruction blocks** — report any surviving `<!--` (check 7, today
  also manual);
- **AC coverage** (feature blueprints only) — report any FRD AC absent
  from the coverage table, any row naming no delta block, any row with
  other than exactly one proving delta, and any legend key that is not
  in bijection with the delta headings.

Until the script exists, replace "you — CLI planned" with the exact
one-liners, so two agents at least run the same check. WHERE the
reference-integrity feature ships, checks 6 and 8 move to `spec check`
and the script's first two conditions retire.

**Evidence.** Finding 2. A reviewer can re-run the same check the author
ran, which is the property `spec-cli.md:115-120` already argues for
("the cost is that the result is an assertion rather than a verifiable
one") — but currently only for checks 1–3.

### P3 — Say how to obtain the CLI when the working tree builds it [snag]

**Problem.** `spec-cli.md:103-113` offers `go install
…/cli/spec@latest` and a release archive. Both fetch a *published*
binary. A session inside this repository — every dogfooding run, and
every run of `spec` on the framework itself — needs the CLI built from
the working tree, because the profile it applies is under edit. Nothing
in the reference says so, and the failure is silent: the check passes or
fails against the wrong vocabulary. Finding 3.

**Change.** Add to `spec-cli.md`, in "When it is not installed":

> WHERE the working tree contains the CLI's own source (a `cli/`
> directory with a `spec` main package), build from it rather than
> installing a published version:
>
> ```
> cd cli && go build -o "$SCRATCH/spec" ./spec
> ```
>
> **Rationale:** the active profile — its document types, slug patterns
> and schema — is compiled into the binary. A published binary checks a
> working-tree corpus against a published vocabulary, so a document that
> conforms to the vocabulary under review can be reported as
> non-conforming, and the reverse.

**Evidence.** Finding 3, and finding 1 is the concrete instance: the
slug pattern that decides whether a feature blueprint passes lives in
the binary, so which binary you ran changes the answer.

### P4 — Ask Tracker Reconciliation to record its query [snag]

**Problem.** `SKILL.md:240` and the template's Tracker Reconciliation
instruction block both say to survey the tracker and neither says how,
nor asks the result to state how it was derived. A section that
classifies 13 issues and a section that classifies 13 of 500 look
identical. Finding 4, and the second occurrence of the defect RFC 0010
P4 raised against the previous blueprint.

**Change.** In `feature-blueprint-template.md`'s Tracker Reconciliation
instruction block, add:

> Name the query and the date at the head of the section — e.g.
> "Reconciled against `bd list --status=open` and
> `bd search <feature keywords>` on 2026-08-07." A groomer reading a
> dated derivation can re-run it; without one, a missing issue and a
> deliberate omission are indistinguishable.

and add to `blueprint-readiness-checklist.md` check 10 that the section
states its derivation. RFC 0010 P4 proposes the project-side half (the
convention for this repository's blueprints); this is the skill-side
half that carries it to every project.

**Evidence.** Finding 4. Two blueprints, two occurrences, same cause —
the template asks for a survey's output and not for its method.

### P5 — Stop the shipped references modelling the forbidden wikilink [snag]

**Problem.** `SKILL.md:218` forbids citing a skill-shipped reference by
wikilink; ten instances in six shipped references do exactly that
(finding 5). An agent learning the convention by example — which is what
a template set is for — learns the wrong one, and the error is invisible
until a reference-resolution check runs, which today is hand-written
(P2).

**Change.** Rewrite the ten instances as backticked paths, per
`cross-document-references.md`'s own rule that a path is written
relative to the file containing it — so
`blueprint-readiness-checklist.md:113` becomes
`` `ac-traceability-standards.md` `` (a sibling), and
`document-readiness-checklist.md:89` becomes
`` `template-authoring-standards.md` ``. Add one line to
`cross-document-references.md`, under "References Outside the Document
Set":

> This applies between shipped references too: a guide in `references/`
> cites its sibling by backticked path, not by wikilink. The rule is
> about the document set a slug resolves in, not about who is citing.

**Evidence.** Finding 5. The agent copied the wrong pattern into an RFC
in the same session; the fix removes the example that taught it.

## Considered and Not Proposed

- **The feature blueprint template's Definition of Done implies a
  `version` bump for every owning blueprint, which a container cannot
  carry** — `ContainerBlueprint` declares no `version` property and sets
  `additionalProperties: false`, so the clause is unsatisfiable for one
  of the four blueprint types. Not proposed here: it is a schema defect,
  tracked as `agent-plugins-e0t`, not misleading text in the skill. The
  previous blueprint recorded the same finding in its own DoD.
- **`adr-template.md` instruction 5 requires updating the `adrs/README.md`
  index and nothing verifies it.** The instruction is present and clear,
  and the agent followed it; proposing a check for an instruction that
  was not missed is speculative on one run. Worth revisiting if a future
  run forgets it.
- **The ADR template's Alternatives Considered table produces rows of
  500–700 characters**, which required checking
  `template-authoring-standards.md` to confirm table cells are exempt
  from the width limit. The exemption is stated ("WHEN a line contains a
  URL, code block content, or table cell data"), so this is a lookup,
  not a defect.
- **The FRD's own content defects** — a criterion that errors on every
  retired feature blueprint, examples citing a deleted property, a
  configuration requirement narrower than the checks that depend on it.
  These are project-document findings, filed as RFC 0013 by `retro-spec`
  in the same session, not skill defects.

## Non-Goals

- No change to what any readiness condition requires. P2 gives the
  conditions a runner; it does not relax or add one.
- No new document type, no change to the skill's document taxonomy.
- No position on whether `fb-` or `feature-` is the better prefix beyond
  the migration-cost argument in P1 — the point is that one of the three
  artifacts must yield.

## Adoption

P1 is independent and should land first: it is the only proposal whose
absence produces an invalid document, and it is a three-file text
change plus a decision.

P2 and P3 are independent of P1 and of each other. P3 is two paragraphs;
P2 is the only proposal with real implementation cost, and its first
two conditions are deliberately scoped to retire when the
reference-integrity feature ships `spec check`'s reference resolution —
so a triage that prefers to wait for that feature can adopt P2's
"replace the placeholder with the exact one-liners" half alone.

P4 depends on nothing, but should be resolved alongside RFC 0010 P4 so
the skill-side and project-side halves of the same rule land together
rather than contradicting each other.

P5 is independent and mechanical.

Backward compatibility: P1 invalidates the slug of any existing
`fb-`-prefixed feature blueprint. No such document exists in this
repository (both use `feature-`), but an adopting project that followed
the template has one, so P1 SHOULD ship with a note that the profile has
always enforced `feature-` and any `fb-` document was already failing
`spec check`. P2, P3, P4 and P5 are additive to guidance and invalidate
no existing document.

Expected effect: a feature blueprint authored from the template passes
its own preflight without the author reading the CLI's source, and the
preflight itself becomes something a reviewer can re-run.

## Resolution
