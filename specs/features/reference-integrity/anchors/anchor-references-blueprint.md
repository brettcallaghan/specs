---
title: "Feature Blueprint: Spec Corpus Anchor References"
description: Resolves a reference's anchor against its target's headings
  — one heading-identifier derivation shared with the link rewriter, two
  checks, and an anchor-qualified refs query
slug: feature-spec-anchor-references
type: feature-blueprint
status: retired
updated_at: 2026-09-05T00:00:00Z
delivery: staged
frd: frd-spec-anchor-references
plan: agent-plugins-rc14
---

# Spec Corpus Anchor References

**FRD:** [[frd-spec-anchor-references]]

## Overview

The anchor half of a reference is already extracted and carried in the
index; what is missing is everything downstream of it. This design adds
one derivation — the heading identifier a markdown renderer computes —
and makes it the single thing both the matcher and the link rewriter
use, so an anchor the checks call resolved is an anchor an exported copy
can target. On top of it, `References` gains anchor resolution, `Corpus`
gains two checks, and the command surface gains an anchor-qualified
`refs` query. The frontmatter schema is widened so a citation may carry
an anchor where a document's own slug may not, and the finding union
gains the two variants the checks report through.

Delta blocks in dependency order: FIND, SREF, AEX, MDL, ARES, CHK, CMD.

## Element Deltas

### Findings in [[module-document-core]] (modified) — FIND

**Intent:** The two anchor checks report defects the existing finding
variants cannot carry: an anchor that named no heading, and one that
named several. Both are structured — the reader's next step is to open a
heading — so each gets a variant rather than a message string, in
`schemas/finding.strata` alongside every other model.

**Behavior:**
- The `FindingDetail` union SHALL carry a variant for an anchor matching
  no heading, holding the anchor as authored and the target document.
- The `FindingDetail` union SHALL carry a variant for an ambiguous
  anchor, holding the anchor as authored and every matching heading with
  its span.
- WHEN a finding is raised for an anchor, the CLI SHALL report it under
  a rule identifier distinct from the unresolved-slug rule, so `--only`
  selects the anchor checks without selecting slug resolution.

**Anchors:** `schemas/finding.strata`; the generated
`cli/core/schema/spec-model.schema.json` and `cli/core/specmodel`;
the constructors in `cli/core/finding`.

### spec-driven-dev registrations in [[module-spec-profile]] (modified) — SREF

**Intent:** A reference-bearing frontmatter property is typed `Slug`
today, and the slug pattern admits no `#`, so a citation carrying an
anchor fails schema validation before anything resolves it. The
identity a document declares and a citation of someone else's identity
are two different types wearing one name; this delta separates them.

**Behavior:**
- The profile's frontmatter schema SHALL declare a citation type that
  admits a slug optionally followed by `#` and a non-empty anchor.
- Every reference-bearing property SHALL be typed as a citation.
- A document's own `slug` SHALL remain typed as a slug, so no document
  can declare an anchored identity.
- The citation type SHALL reject a trailing `#` with no anchor after it.

**Anchors:** `schemas/vocabulary.strata` (the `Slug` alias),
`schemas/frontmatter.strata` (the reference-bearing properties),
`scripts/generate-schemas.sh`.

### References in [[module-document-core]] (modified — extraction) — AEX

**Intent:** Extraction already splits a reference on its first `#` and
records the anchor as authored, from a body wikilink and from a
registered frontmatter property alike. This delta pins the semantics the
FRD now fixes and covers them with tests, so a later change cannot
quietly acquire an alias splitter or normalize on the way in.

**Behavior:**
- The component SHALL record an anchor as authored, without
  normalization.
- The component SHALL divide a reference at no vertical bar, so a bar in
  an anchor is recorded as ordinary text of that anchor.
- The component SHALL record an anchor from a reference-bearing
  frontmatter property exactly as it records one written in the body.
- WHEN a reference is written as an embed, the component SHALL record it
  as it records the equivalent link.

**Anchors:** `cli/core/references/references.go` (`reference`,
`bodyReferences`, `frontmatterReferences`).

### Markdown Links in [[module-document-core]] (modified) — MDL

**Intent:** `cli/core/mdlinks` holds a private heading-identifier
derivation, and [[frd-spec-structure]] exposes the one this feature
matches against. Left as two, they can disagree, and every disagreement
is a link the checks call resolved and a reader cannot follow. This
delta makes the rewriter a consumer of the shared derivation, and
settles which heading a fragment is derived from.

**Behavior:**
- `Markdown Links` SHALL derive a fragment through the heading
  identifier `Markdown Structure` exposes rather than through a
  derivation of its own.
- WHERE a reference's anchor resolves to a heading, `Markdown Links`
  SHALL derive the emitted fragment from that heading.
- WHERE a reference's anchor resolves to no heading, `Markdown Links`
  SHALL derive the emitted fragment from the anchor as authored.

**Anchors:** `slugifyAnchor` and `anchorFragment` in
`cli/core/mdlinks/mdlinks.go`.

### References in [[module-document-core]] (modified — matching) — ARES

**Intent:** Resolution answers what a reference reached; it answers only
the slug half today. This delta extends it to the anchor, against the
headings the index carries for the target document, and keeps the
existing seam: extraction knows a document, resolution knows a corpus,
and matching a heading is a corpus question.

**Behavior:**
- The component SHALL match an anchor against a heading by deriving a
  heading identifier from each and comparing them.
- WHERE exactly one heading matches, the component SHALL record the
  anchor as resolved to that heading.
- WHERE more than one heading matches, the component SHALL record the
  anchor as ambiguous and as resolved to no heading, carrying every
  match.
- WHERE a reference names an anchor and no slug, the component SHALL
  resolve the anchor against the document containing the reference.
- WHERE the target document's headings are unavailable, the component
  SHALL record the anchor as unverified rather than as resolved.
- The component SHALL resolve an embed exactly as it resolves the
  equivalent link.

**Anchors:** `cli/core/references/resolve.go`; the resolution states in
`schemas/index.strata`.

### Anchor checks in [[module-document-core]] (added) — CHK

**Intent:** The checks are what makes an anchor's breakage visible. They
report at the document a reader must open, which for a reference is the
referring document, and they add nothing where the slug already failed —
one mistake, one finding.

**Behavior:**
- WHEN an anchor matches no heading of a target whose headings are
  available, the CLI SHALL report a finding of severity `error` naming
  the target document and the anchor.
- WHEN an anchor matches more than one heading, the CLI SHALL report a
  finding of severity `error` naming every matching heading and its
  span.
- WHERE the target document's headings are unavailable, the CLI SHALL
  report the reference with severity `unverified` naming what is
  missing.
- WHERE a reference's slug resolves to no corpus member, the CLI SHALL
  report no anchor finding for it.
- The anchor checks SHALL run over every reference carrying an anchor in
  an ordinary corpus check, not behind an option.
- WHEN the CLI checks a single document, it SHALL report inbound anchor
  references to that document that no longer resolve, each at the span
  in the referring document.

**Anchors:** `cli/core/corpus/check.go`; the snapshot cases under
`cli/testdata/cases/`.

### Command Surface in [[container-spec-cli]] (modified) — CMD

**Intent:** `refs` answers what cites a document; the rename question an
anchor creates is narrower — what cites this heading. The argument
already has a place to put the anchor, since a reference's own notation
carries one.

**Behavior:**
- WHEN `refs` is given an argument qualified by an anchor, the CLI SHALL
  list every reference whose anchor resolves to that heading, each with
  its referring document and span.
- WHEN a reference's anchor matches more than one heading, the CLI SHALL
  list it for every heading it matches, marked as matching more than
  one.
- WHERE no reference names the given anchor, the CLI SHALL report that
  and exit zero.
- The `--help` text for the anchor-qualified form SHALL derive from the
  declaration that defines it.
- The command's output SHALL conform to [[cli-output-standards]], as
  SHALL the findings the CHK delta reports.

**Anchors:** `cli/cmd/refs.go`; `cli/cmd/registry.go`; the snapshot
cases under `cli/testdata/cases/`.

### Relied-upon elements (unchanged)

- [[frd-spec-structure]] — supplies the heading tree the ARES and CHK
  deltas match against, and the heading-identifier derivation ARES and
  MDL both use; see Sequencing Constraints.
- [[container-spec-cli#Command Surface]] — the warm corpus open that
  `check` and `refs` already share.
- `Corpus` in [[module-document-core]] — the index the resolved
  reference graph is read from, in both directions.
- `Spans` in [[module-document-core]] — every position this feature
  reports, including a matching heading's span.
- [[contract-specs-configuration]] — the configuration surface; this
  feature adds no key to it.

## Contracts

- No new contract document. The finding JSON gains two variants under
  the additive rule the generated schema already documents, and the
  index gains no shape it did not carry: an anchor remains one string
  recorded as authored.
- [[contract-specs-configuration]] is untouched. The citation type SREF
  introduces is internal to the profile's frontmatter schema, and no
  `specs.toml` key changes meaning.

## Tracker Reconciliation

- `agent-plugins-bem` — prerequisite: the relational reference checks
  read documents off disk on the warm single-document path, which is the
  path CHK extends with per-reference heading lookups. Landing anchor
  checks on top of it compounds a budget defect rather than inheriting
  it.
- `agent-plugins-z8r` — out-of-scope: whether a retired document is
  exempt from reference checks is a policy question for reference checks
  as a whole. The anchor checks follow whatever that decides and
  introduce no exemption of their own.
- `agent-plugins-im99` — out-of-scope: the feature-blueprint template's
  slug rule contradicts the profile. This document follows the profile's
  registered pattern; the template fix is not this feature's to make.

## Sequencing Constraints

- `delivery: staged`. Two stages, each a reviewable increment:
  - **Stage 1 — the anchor in the index:** FIND, SREF, AEX. The
    extraction semantics are pinned and tested, a frontmatter citation
    may carry an anchor, and the finding vocabulary exists. No check
    runs, no finding is raised, and no exit status changes.
  - **Stage 2 — matching, checks and query:** MDL, ARES, CHK, CMD.
    Anchors resolve, the checks report, `refs` answers by heading, and
    exit status becomes sensitive to anchors for the first time.
- The heading tree in `Markdown Structure` SHALL be built before the
  ARES and CHK deltas land. It does not exist today:
  [[frd-spec-structure]] specifies it, and AC-SPECSTR-001.1 — a heading
  recorded with the span of the content beneath it — is the whole of
  what this feature needs from it. Stage 2 lands after that criterion
  does. Shipping stage 2 first would report `unverified` for every
  anchored reference in the corpus, and this repository's own gates run
  `specs check --strict`, which fails on `unverified`.
- The FIND delta SHALL land before CHK, which reports through the
  variants it declares.
- The SREF delta SHALL land before a frontmatter citation carrying an
  anchor is added to the corpus, since the schema rejects one until it
  lands.
- The heading-identifier derivation [[frd-spec-structure]] exposes
  SHALL land before ARES, which matches through it, and before MDL,
  which derives a fragment through it.
- The ARES delta SHALL land before CHK and CMD, which read what it
  resolved.
- **Lockstep rules spanning delta blocks, and how they are satisfied:**
  - [[module-document-core]] requires the schema version and the emitted
    index to change together when the index model changes shape. No
    delta changes the index model's shape: the anchor stays one string,
    and the resolution states ARES records are additive under the
    additive-at-1.0 rule [[container-spec-cli]] states. The satisfying
    act is the recorded no-bump.
  - [[module-document-core]] requires a new parsing capability to be
    built on the Spans delimiter scanner. No delta here parses anything
    new: MDL consumes a derivation over text a heading already carries;
    every position this feature reports is the `Span` its source already
    holds.
  - [[container-spec-cli]] requires the snapshot regeneration script to
    be run and every diff reviewed when observable output changes. CHK
    and CMD each change observable output, and each lands with its
    regenerated snapshots in the same change rather than a shared
    regeneration pass at the end of stage 2.
- No backward-compatibility window applies. [[module-document-core]] and
  [[module-spec-profile]] are both `visibility: internal` and evolve in
  lockstep with their only consumer.

## AC Coverage

Legend: FIND = the finding variants for the two anchor rules; SREF =
the citation type in the profile's frontmatter schema; AEX = References
(extraction); MDL = the link rewriter's fragment, derived from the
resolved heading; ARES = References (anchor resolution); CHK = the anchor
checks in `Corpus`; CMD = Command Surface in [[container-spec-cli]].

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-SPECANCH-001.1 | **AEX (proves)** | unit |
| AC-SPECANCH-001.2 | **AEX (proves)** | unit |
| AC-SPECANCH-001.5 | **AEX (proves)** | unit |
| AC-SPECANCH-001.6 | **SREF (proves)**, AEX | integration |
| AC-SPECANCH-002.1 | **CHK (proves)**, ARES, FIND | unit |
| AC-SPECANCH-002.2 | **ARES (proves)**, MDL | unit |
| AC-SPECANCH-002.4 | **CHK (proves)**, ARES, FIND | unit |
| AC-SPECANCH-002.5 | **CHK (proves)**, ARES, FIND | unit |
| AC-SPECANCH-002.6 | **ARES (proves)** | unit |
| AC-SPECANCH-002.7 | **ARES (proves)**, AEX | unit |
| AC-SPECANCH-002.8 | **CHK (proves)**, ARES | unit |
| AC-SPECANCH-002.9 | **CHK (proves)**, CMD, FIND | integration |
| AC-SPECANCH-003.1 | **CMD (proves)**, ARES | integration |
| AC-SPECANCH-003.2 | **CHK (proves)**, CMD | integration |
| AC-SPECANCH-003.3 | **CMD (proves)** | integration |
| AC-SPECANCH-003.4 | **CMD (proves)**, ARES | integration |

## Key Decisions

- [[adr-012]] — the wikilink and heading-anchor conventions this feature
  resolves, including the slug-less form its FRD added. No new ADR: the
  matching rule is stated as an acceptance criterion in the FRD, which
  is durable, and this design chooses no alternative the FRD left open.
- [[adr-016]] — Strata is the single source for generated models, which
  is why FIND and SREF are schema deltas rather than hand edits to Go.
- [[adr-018]] — references are one edge collection on the index, with
  both directions derived from it; CMD answers by heading from the same
  collection rather than gaining an index of its own.

## Definition of Done

- [x] Every delta above is folded into its owning blueprint:
  - [x] [[module-document-core]] — References carries the five anchor
    semantics beside its existing resolution list (ARES); the extraction
    half's split-on-first-`#` semantics stand for AEX; Markdown
    Structure records `Identifier` and why anchor resolution matches
    through that one derivation; Markdown Links derives its fragment
    through it, with the relationship paragraph and the component
    diagram gaining the `mdlinks → md` edge (MDL); Findings' detail
    vocabulary gains both anchor variants (FIND); Corpus records both
    rules, their severities, the unverified case, the one-mistake rule
    and the corpus-wide exception (CHK)
  - [x] [[module-spec-profile]] — the registered frontmatter schema's
    citation type is recorded, along with the drift guard renamed to
    assert against it (SREF)
  - [x] [[container-spec-cli]] — Command Surface takes the
    `<slug>#<anchor>` argument form, the ambiguity mark, the empty
    answer, the `--outbound` refusal, and the YAML/JSON rendering
    [[cli-output-standards]] requires (CMD)
- [x] New ADRs are referenced from the owning blueprints' Key Decisions
  (none authored by this feature)
- [x] Every AC in the coverage table is verified at its declared level
  and cited from its tests per [[ac-traceability-standards]] — all 16.
  Eleven of the twelve declared `unit` carry an `// AC:` comment on a
  unit test: AC-SPECANCH-001.1/.2/.5 in
  `cli/core/references/references_test.go`; 002.2/.6/.7 in
  `cli/core/references/anchor_test.go`; 002.1/.4/.5/.8/.9 in
  `cli/core/corpus/anchor_check_test.go`. The five declared
  `integration` are cited by a snapshot case manifest's `cites`, which
  [[testing-standards]] accepts equally: `index-citation-anchor`
  (001.6), `check-anchor-defects` and `check-only-anchor-rule` (002.9,
  which is cited at both levels), `refs-anchor` and `refs-anchor-json`
  (003.1), `check-path-anchor-elsewhere` (003.2), `refs-anchor-no-match`
  (003.3), `refs-anchor-ambiguous` (003.4)
- [x] Every issue listed as subsumed in Tracker Reconciliation is closed
  or re-scoped (none; `agent-plugins-bem` closed as the prerequisite it
  was, and `agent-plugins-z8r` — deferred — and `agent-plugins-im99` —
  open — remain out-of-scope, neither touched by this feature)
- [x] [[frd-spec-markdown-links]]'s revised fragment criteria are
  verified by the MDL delta's tests, since this feature is what makes
  them true — AC-SPECMD-001.3 by
  `TestAResolvedAnchorTakesItsFragmentFromTheResolvedHeading` and
  `TestAMappedTargetCarriesTheResolvedHeadingsFragmentToo` in
  `cli/core/mdlinks/mdlinks_test.go`, AC-SPECMD-001.7 by the
  authored-anchor fallback tests in the same file and in
  `cli/core/mdlinks/anchor_edges_test.go`. Both branches are exercised
  by fixture rather than by corpus input, because the matcher resolves
  an anchor by identifier equality and so cannot produce a reference
  whose two fragment sources differ; the criterion is what stops a
  looser matcher from emitting fragments that reach nothing
- [x] `specs check --strict` reports no findings over this repository's
  corpus, with `reference/unresolved-anchor` and
  `reference/ambiguous-anchor` both listed `passed`. Eleven references
  in this corpus carry an anchor — six into
  `container-spec-cli#Command Surface`, five into component rosters of
  [[module-document-core]] — and every one resolves
- [x] This blueprint's status → retired

## Open Questions

None.

## References

- [[frd-spec-anchor-references]] — the FRD this blueprint realizes
- [[frd-spec-reference-integrity]] — the parent feature, whose
  extraction and resolution this design extends
- [[frd-spec-structure]] — supplies the heading tree stage 2 matches
  against, in AC-SPECSTR-001.1
- [[frd-spec-markdown-links]] — the rewriter MDL makes a consumer of
  the shared derivation
- [[c4-standards]] — the container, module and component definitions the
  delta blocks are written against
- [[cli-output-standards]] — what the anchor query and the anchor
  findings print, and in which rendering
- [[ac-traceability-standards]] — the citation obligation the coverage
  table feeds
