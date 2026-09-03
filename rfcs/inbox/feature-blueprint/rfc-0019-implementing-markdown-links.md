---
title: "RFC: Implementing markdown links — a behavior with no criterion,
  a likeness that contradicts one, and citations that survived the
  documents they name moving into the corpus"
description: Assign the unreadable-operand path its own acceptance
  criteria; make the command surface state which corpus-open helper a
  subcommand takes; require updated_at to move when a document does;
  convert a backticked path to a wikilink when its target enters the
  corpus; and record that a schema regeneration invalidates committed
  snapshot fixtures
slug: rfc-0019
type: rfc
status: ready-for-review
updated_at: 2026-08-14T07:30:00Z
origin: self
skill: bead-bonanza
skill_version: "1.5"
model: "claude-opus-5"
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "0.17.0"
---

# RFC: Implementing markdown links

## Summary

The six work orders of `agent-plugins-ia5` were implemented against
[[frd-spec-markdown-links]], the (now retired) markdown-links feature
blueprint, and the two architecture blueprints its deltas fold into. The
documents held up well: all five deltas landed, all 31 acceptance
criteria carry citations, `spec check --strict` exits 0, and nothing
declared a durable edge into the transient blueprint — so retiring it
broke no check, which is `rfcs/inbox/feature-blueprint/rfc-0015` P1
already working. What they did not do is cover their own edges. One
required behavior has no acceptance criterion at all, so two snapshot
cases ship citing nothing. One Behavior bullet states a likeness to four
existing subcommands whose shared helper does the exact thing a
criterion in the same delta forbids. Five commits edited two durable
blueprints without moving either one's `updated_at`, because no rule
says they must. And a concurrent epic moved thirteen documents into the
corpus without converting the backticked paths that name them, leaving
38 citations pointing at files that are not there — invisible to every
check, because a backticked path is by construction the thing `spec
check` does not resolve.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins |
| Phase executed | Implemented the six work orders of `agent-plugins-ia5` (FND, TGT, MDL, MAP, CMD, DOD) |
| Documents exercised | [[frd-spec-markdown-links]]; `specs/features/markdown-links/markdown-links-blueprint.md`; [[module-document-core]], [[container-spec-cli]]; [[testing-standards]], [[ac-traceability-standards]], [[frontmatter-standards]], [[cross-document-references]]; [[adr-015]], [[adr-018]]; [[frd-spec-reference-integrity]] |
| Documents produced | The five deltas' folds into the two architecture blueprints; the blueprint's retirement; this RFC |
| Skill(s) followed | bead-bonanza v1.5; retro-spec v1.5.0 |
| Templates / standards read | feature-blueprint-template v1.4; frd-template v1.2; work-order-template v1.5; rfc-template v1.3; `testing-standards.md` v0.1; `ac-traceability-standards.md` v0.3.0; `frontmatter-standards.md` v0.2; `cross-document-references.md` v0.3 |
| Plugin version | 0.17.0 |
| Session date | 2026-08-14 |

### Completeness of the evidence base

Every one of the five implementing items — `agent-plugins-ia5.1`
through `.5` — recorded `retro:` comments per
[[findings-capture-protocol]] (three, four, three, four and five
respectively), and all five are represented below. No item went
unrecorded, so unlike
`rfcs/inbox/feature-blueprint/rfc-0015` this RFC rests on a uniform
base. Findings 3, 4 and 6 are the DoD verification's own, not any
implementer's.

## Evidence

1. **A required behavior has no acceptance criterion.** The blueprint's
   CMD delta states "IF the named document cannot be read, THEN the
   `links` subcommand SHALL report an error finding and write no
   rewritten document"
   (`specs/features/markdown-links/markdown-links-blueprint.md`, CMD
   Behavior) — the one bullet in that list carrying no AC in
   parentheses. [[frd-spec-markdown-links]] gives the analogous link-map
   path two criteria, AC-SPECMD-005.8 ("IF the link map cannot be read,
   or a record declares no `path`, THEN the CLI SHALL report a finding
   naming the fault") and AC-SPECMD-005.9, and gives the operand path
   none. The CMD implementer minted a rule id for it anyway
   (`document/unreadable`, `cli/cmd/finding.go`) and wrote two snapshot
   cases, `cli/testdata/cases/links-unreadable-operand/case.toml` and
   `links-no-strict-flag/case.toml`, each declaring `cites = []`. Its
   `retro:` comment: "No AC-SPECMD id exists for 'the named document
   cannot be read' … the FRD assigns it none."

2. **A Behavior bullet states a likeness that a criterion in the same
   delta forbids.** The blueprint's CMD delta: "The `links` subcommand
   SHALL open the corpus warm, as `check`, `refs`, `slug-exists` and
   `next-id` do." Those four open through `openCorpus`
   (`cli/cmd/root.go:170`), which on a cold cache runs
   `_ = corpus.Write(c.Index(), dest, nil)` — stated in
   [[container-spec-cli]]'s Data section as "The reading commands **also
   write it when they had to build it**". Two bullets above, the same
   delta says "The `links` subcommand SHALL create, modify and delete no
   file (AC-SPECMD-005.3)." Neither the blueprint nor the FRD reconciles
   them. The CMD implementer resolved it unilaterally, adding
   `openCorpusReadOnly` (`cli/cmd/links.go:153`) and a cold-cache
   regression test.

3. **Five commits edited two durable architecture blueprints without
   moving either one's `updated_at`.** `82d46f8`, `56ec2bf`, `daa7f3c`,
   `59d8c05` and `b008ff5` all touch
   `specs/modules/document-core/document-core-module.md` and/or
   `specs/containers/spec-cli/spec-cli-container.md`. The former still declares
   `updated_at: 2026-08-09T00:00:00Z` and `version: "0.3"`; the latter
   `updated_at: 2026-08-12T00:00:00Z`. [[frontmatter-standards]] defines
   the property once, in a table — "| `updated_at` | datetime | Last
   modification timestamp | RFC 3339 format …" (line 88) — and states
   no obligation to move it when the document changes. The retired
   reference-integrity blueprint compensated per-document inside its own
   Definition of Done ("`version` bumped 0.2 → 0.3", written out twice);
   this blueprint's Definition of Done did not, and the bumps did not
   happen.

4. **38 backticked path citations across `specs/` name documents that
   are no longer at those paths.** `agent-plugins-eaa.2` and `eaa.3`
   moved thirteen shipped documents out of
   `plugin/skills/*/references/` into `specs/standards/` and
   `specs/guides/` and inverted the vendoring, making the authored
   copies corpus members citable by slug. The citations did not follow.
   Two of them belong to this phase's own documents: the blueprint's
   Definition of Done cited
   `` `plugin/skills/spec/references/ac-traceability-standards.md` ``
   and [[frd-spec-markdown-links]] still cites
   `` `plugin/skills/spec/references/cross-document-references.md` ``
   twice (Terminology; References). [[cross-document-references]] itself
   offers `` `ears-guide.md` `` as an example of "the framework's own
   reference guides shipped with a skill" (References Outside the
   Document Set) — a document that is now `specs/guides/ears-guide.md`.
   `spec check --strict` exits 0 over all of it, correctly: the standard
   scopes wikilinks to corpus members and backticked paths to everything
   else, and no check resolves a backticked path.

5. **Regenerating `schemas/finding.strata` invalidated four committed
   snapshot fixtures, and no document says it can.** The FND
   implementer: "Regenerating `schemas/finding.strata` changes the
   embedded profile schema bytes … which feed `corpus.inputHasher`'s
   'profile-schema' field — so four committed CLI snapshot fixtures …
   failed on a stale `input_hash` until
   `scripts/regenerate-cli-snapshots.sh` was run." The chain's first
   link is documented — [[container-spec-cli]]'s Data section lists "the
   active profile's schema and per-type table" among the hash's inputs —
   and its consequence is not: neither [[testing-standards]]'s Snapshot
   Case Structure, nor `schemas/README.md`, nor the blueprint's own FND
   Anchors ("`scripts/generate-schemas.sh` regenerates
   `cli/core/specmodel`, which it rewrites from scratch on every run")
   says that a schema change alone can make every committed fixture
   stale.

6. **AC-SPECMD-005.3 is declared `integration` and proven by a Go test,
   not a snapshot case.** It is the one row of the blueprint's AC
   Coverage table whose citation does not come from the harness its
   level names: `cli/cmd/links_test.go`'s
   `TestLinksCreatesModifiesAndDeletesNoFile` digests the whole tree
   before and after the run, which a case cannot do —
   [[testing-standards]] requires a snapshot case to "consist of input
   files and expected artifacts, and SHALL NOT contain executable test
   code", and the same standard's Snapshot Cases section says "The
   observable behavior of the CLI SHALL be verified by snapshot cases
   rather than by unit tests of the command surface."

7. **A term the FRD defines was re-derived from the work order's
   paraphrase of it.** The TGT implementer recorded that the item's
   Design phrase "slug names a corpus document" "isn't tied explicitly
   to `ReferenceResolution`'s three states (resolved / non-member /
   unresolved)", and chose `resolved` only. [[frd-spec-markdown-links]]
   already settles it in Terminology — "**Corpus document** — a document
   beneath the spec root, indexed, and citable by slug" — which excludes
   a non-member. The choice was right; the FRD was not consulted for it.

## Proposals

### P1 — Give the unreadable-operand path its own acceptance criteria

**Problem.** [[frd-spec-markdown-links]] REQ-SPECMD-005 covers every
fault in the *link map* and none in the *document being rewritten*.
AC-SPECMD-005.8 and AC-SPECMD-005.9 are the pair for the map; there is
no text anywhere in the FRD for the operand, though the blueprint
requires the behavior and the CLI implements it. The consequence is
concrete and shipped: two snapshot case manifests declare `cites = []`,
so a behavior with a rule id, two cases and a `--help`-visible surface
is invisible to the coverage obligation [[ac-traceability-standards]]
states ("Every acceptance criterion of an FRD SHALL be cited by at
least one test").

**Change.** Append to REQ-SPECMD-005's acceptance criteria, at the next
unused numbers per [[ac-traceability-standards]]'s append-only rule:

> - **AC-SPECMD-005.11:** IF the document named for rewriting cannot be
>   read, THEN the CLI SHALL report a finding of severity error naming
>   the file and the fault.
> - **AC-SPECMD-005.12:** IF the document named for rewriting cannot be
>   read, THEN the CLI SHALL NOT emit a rewritten document.

Then set `cites = ["AC-SPECMD-005.11", "AC-SPECMD-005.12"]` in
`cli/testdata/cases/links-unreadable-operand/case.toml`. Leave
`links-no-strict-flag` at `cites = []`: it verifies the absence of an
option, which is a Behavior statement rather than a criterion.

**Evidence.** Finding 1. The next reader of the FRD sees the operand and
the map treated alike, and the next coverage sweep counts a behavior
that today it cannot see.

### P2 — Make a subcommand's corpus-open behavior an explicit choice

**Problem.** [[container-spec-cli]] describes corpus opening in two
places that a delta author reads as one. The Data section states the
persisting behavior as a property of a category — "The reading commands
**also write it when they had to build it**" — and the Container Context
states the exception as a property of one command — "which `links` opens
warm without persisting for exactly this reason". Nothing tells the
author of the *seventh* subcommand that "opens the corpus warm" names
two different helpers with different filesystem effects. That is exactly
the sentence this feature's blueprint wrote, and the contradiction with
AC-SPECMD-005.3 reached the implementer rather than the author.

**Change.** Add to [[container-spec-cli]]'s Change Guidance:

> - WHEN a subcommand is added that opens the corpus warm, the
>   blueprint SHALL state whether it persists a freshly built index.
>   `openCorpus` writes the index back on a cold cache and
>   `openCorpusReadOnly` does not, so "opens the corpus warm" alone
>   does not settle whether the run creates a file — and a subcommand
>   bound by a criterion that forbids creating one SHALL name the
>   read-only helper.

**Evidence.** Finding 2. The delta author writing the next subcommand
either names the helper or notices they cannot, instead of stating a
likeness that a criterion two bullets away forbids. Independent of P1.

### P3 — Require `updated_at` to move when the document does

**Problem.** [[frontmatter-standards]] introduces `updated_at` as a
required property and defines it as "Last modification timestamp" (line
88), and never states that changing a document obliges you to change it.
Every consumer treats the value as current — triage judges "is this
already fixed?" from it, `document-readiness-checklist` reads it as
freshness — while nothing makes it so. Five commits demonstrated the
gap in one feature. The workaround in use is per-blueprint: the
reference-integrity blueprint's Definition of Done spelled out "`version`
bumped 0.2 → 0.3" for each folded document, which works only for as
long as each blueprint author remembers to write it, and this one did
not.

**Change.** Add to [[frontmatter-standards]], Required Properties, after
the property table:

> WHEN a document's content changes, `updated_at` SHALL be set to the
> time of that change, in the same change.
>
> WHERE a document carries the optional `version` property, a change
> that alters what the document requires or describes SHALL also
> increment it; a purely editorial change SHALL NOT.
>
> **Rationale:** every consumer of `updated_at` — RFC triage asking
> whether a defect is already fixed, a readiness review asking whether a
> document has been revisited since the decision it records — reads it
> as the answer to "when did this last change?". A value that only
> sometimes answers that reads as fresh when it is not, which is worse
> than an absent one.

**Evidence.** Finding 3. Backward compatible: it constrains future
changes and leaves every existing document valid. It also removes the
need for each blueprint's Definition of Done to restate the bump
per folded document, which is what failed here.

### P4 — Convert a backticked path when its target enters the corpus

**Problem.** [[cross-document-references]] partitions citations by
whether the target is in the specification set — wikilink if it is,
backticked path if it is not — and says nothing about a document
crossing that line. It is not hypothetical: thirteen documents crossed
it in this run, and the standard's own examples crossed with them
(References Outside the Document Set still offers `` `ears-guide.md` ``
as a document "shipped with a skill", and `specs/guides/ears-guide.md`
is a corpus member). Because the rule's own design makes a backticked
path unresolvable by `spec check`, nothing reports the 38 stale
citations, and `spec check --strict` exits 0 over a corpus that carries
them.

**Change.** Add to [[cross-document-references]], References Outside the
Document Set:

> WHEN a document that was referenced by backticked path becomes a
> corpus member, every backticked-path citation of it SHALL be converted
> to a wikilink in the same change.
>
> **Rationale:** the two syntaxes partition citations by whether the
> target is in the set, so a document entering the set makes every
> existing citation of it wrong. No check reports this — a backticked
> path is by construction the citation `spec check` does not resolve —
> so the move itself is the only moment the correction can be made
> reliably.

The 38 existing instances are repair work for the epic that moved the
documents, not for this rule; file them as their own item.

**Evidence.** Finding 4, including the two citations this phase's own
documents carry. Independent of P1–P3.

### P5 — Record that a schema change invalidates committed fixtures

**Problem.** The fixtures under `cli/testdata/cases/**/expected/` are
compared byte for byte and carry an `input_hash` that folds the active
profile's schema. So a change confined to `schemas/*.strata` — touching
no command, no rule, no expected message — can make four unrelated cases
fail. [[testing-standards]]'s Snapshot Case Structure describes the case
layout and the regeneration script without naming any input outside
`cli/cmd` that can invalidate one, and `schemas/README.md` names no
consequence beyond regeneration of `cli/core/{schema,specmodel}`. The
FND implementer hit it, diagnosed it, and recorded that "neither the
bead nor the FND delta's 'regeneration is safe' note mentions this".

**Change.** Add to [[testing-standards]], Snapshot Case Structure:

> A case's expected artifacts depend on the active profile's schema as
> well as on the command surface, because the emitted index carries an
> `input_hash` folding that schema. WHEN `schemas/*.strata` changes, the
> snapshot regeneration script SHALL be run and its diff reviewed in the
> same change, even where no command behavior changed.

and add the same fact, one sentence, to `schemas/README.md`'s
regeneration instructions.

**Evidence.** Finding 5. The next schema change budgets for the
regeneration instead of discovering four red cases it did not cause.
Independent of P1–P4.

## Considered and Not Proposed

- **Re-declaring AC-SPECMD-005.3 at a reachable level, or proposing a
  rule for it.** Finding 6 is the third instance of the shape
  `rfcs/inbox/feature-blueprint/rfc-0015` P2 already proposes a rule for
  — "using the harness that owns that level", plus "WHERE a criterion is
  declared at a level other than the one its proving delta's harness
  would suggest, the coverage table SHALL name the test that verifies
  it". That RFC is still `ready-for-review` with an empty Resolution.
  Recorded here as corroborating evidence for it rather than
  re-proposed; the criterion is genuinely proven, by a genuinely
  integration-level test, and only the harness differs.
- **Requiring a work order to quote the FRD term it paraphrases.**
  Finding 7 is real friction, but its cause is how the item was written,
  not what the FRD says — the FRD's Terminology already answers the
  question the implementer had to reason out. That is a work-order
  authoring concern and belongs to the `retro-skill` lens
  (`work-order-template` v1.5), not to this project's documents.
- **Naming the three `ReferenceResolution` states in
  [[frd-spec-markdown-links]].** Proposed by the TGT implementer.
  Rejected: `ReferenceResolution` is an implementation type of
  [[module-document-core]], and an FRD that named it would bind the
  requirement to the model rather than to the behavior. The FRD's
  Terminology is the right layer and already correct.
- **Importing the `[[` / `]]` delimiter constants rather than
  duplicating them in `cli/core/mdlinks`.** The MDL implementer flagged
  the duplication deliberately; both copies are unexported and the
  packages should not share unexported symbols. No document is wrong,
  and no document would have prevented it.
- **Documenting that a frontmatter-origin reference is not rewritten in
  [[frd-spec-markdown-links]].** The MDL implementer read the FRD's
  Terminology ("**Wikilink** — a slug citation in `[[slug]]` or
  `[[slug#Heading Text]]` form") as scoping substitution to body-origin
  references, and it does: every REQ-SPECMD-001 criterion is phrased
  over "wikilink", and AC-SPECMD-001.5 forbids altering anything else.
  The reading is supported, and the fact is now recorded where it
  belongs — [[module-document-core]]'s Markdown Links section.

## Non-Goals

- No new document type, and no change to what `spec check` checks. P4
  proposes an authoring obligation, not a new check; the reason a check
  cannot cover it is stated in the proposal.
- No repair of the 38 stale backticked citations here. The RFC's
  authority ends at the proposal, and the repair belongs to the epic
  that moved the documents.
- No change to the verification level of any existing acceptance
  criterion; finding 6 is routed to an open RFC rather than acted on.

## Adoption

P1 through P5 are independently adoptable and touch disjoint files: P1
`specs/features/markdown-links/markdown-links-frd.md` and one case
manifest; P2 `specs/containers/spec-cli/spec-cli-container.md`; P3
`specs/standards/frontmatter-standards.md`; P4
`specs/standards/cross-document-references-standards.md`; P5
`specs/standards/testing-standards.md` and `schemas/README.md`.

Each is backward compatible. P1 appends criteria at unused numbers and
renumbers nothing, per [[ac-traceability-standards]]. P2 and P4 add
obligations that bind future changes; every existing document stays
valid, and P4 names the 38 existing instances as separate repair work
rather than as a migration this rule performs. P3 leaves existing
frontmatter valid and, if adopted, is worth applying once to the two
architecture blueprints finding 3 names. P5 documents an existing
behavior of the fixtures and changes none of them.

Adopting P3 and P4 together is worth more than either alone: both are
about a document's state going stale with nothing reporting it, and both
are enforced only by an author reading the standard, so a reviewer who
learns one learns where to look for the other.

## Resolution
