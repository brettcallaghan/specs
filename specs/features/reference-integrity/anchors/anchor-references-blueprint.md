---
title: "Feature Blueprint: Spec Corpus Anchor References"
description: Resolves a reference's anchor against its target's headings
  — one heading-identifier derivation shared with the link rewriter, two
  checks, and an anchor-qualified refs query
slug: feature-spec-anchor-references
type: feature-blueprint
status: draft
updated_at: 2026-09-04T00:00:00Z
delivery: staged
frd: frd-spec-anchor-references
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

Delta blocks in dependency order: FIND, SREF, AEX, HID, ARES, CHK, CMD.

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

### Heading identifiers in [[module-document-core]] (added) — HID

**Intent:** Two halves of the module need the same derivation: the
matcher, to compare an anchor with a heading, and the rewriter, to emit
a fragment. `cli/core/mdlinks` holds a private one today. Left as two,
they can disagree, and every disagreement is a link the checks call
resolved and a reader cannot follow. This delta gives the derivation one
home, in the component that owns heading semantics, and makes the
rewriter a consumer of it.

**Behavior:**
- `Markdown Structure` SHALL expose the heading identifier of a text:
  its rendered inline content lowercased, with every character that is
  not a letter, digit, hyphen, underscore or space removed, and each
  remaining space replaced by a hyphen.
- The derivation SHALL be a pure function of the text, deriving no
  positional suffix for a repeated heading.
- `Markdown Links` SHALL derive a fragment through that surface rather
  than through a derivation of its own.
- WHERE a reference's anchor resolves to a heading, `Markdown Links`
  SHALL derive the emitted fragment from that heading.
- WHERE a reference's anchor resolves to no heading, `Markdown Links`
  SHALL derive the emitted fragment from the anchor as authored.

**Anchors:** `cli/core/markdown`; `slugifyAnchor` and `anchorFragment`
in `cli/core/mdlinks/mdlinks.go`.

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

**Anchors:** `cli/cmd/refs.go`; `cli/cmd/registry.go`; the snapshot
cases under `cli/testdata/cases/`.

### Relied-upon elements (unchanged)

- [[frd-spec-structure-conformance]] — supplies the heading tree the
  ARES and CHK deltas match against; see Sequencing Constraints.
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
  - **Stage 2 — matching, checks and query:** HID, ARES, CHK, CMD.
    Anchors resolve, the checks report, `refs` answers by heading, and
    exit status becomes sensitive to anchors for the first time.
- The heading tree in `Markdown Structure` SHALL be built before the
  ARES and CHK deltas land. It does not exist today, so stage 2 depends
  on [[frd-spec-structure-conformance]] shipping its heading extraction
  first; a check with no heading source reports `unverified` for every
  anchored reference, and this repository's own gates run `spec check
  --strict`, which fails on `unverified`.
- The FIND delta SHALL land before CHK, which reports through the
  variants it declares.
- The SREF delta SHALL land before a frontmatter citation carrying an
  anchor is added to the corpus, since the schema rejects one until it
  lands.
- The HID delta SHALL land before ARES, which matches through the
  derivation it exposes.
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
    built on the Spans delimiter scanner. HID derives an identifier from
    text a heading already carries and computes no position of its own;
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
(extraction); HID = the heading-identifier derivation and its two
consumers; ARES = References (anchor resolution); CHK = the anchor
checks in `Corpus`; CMD = Command Surface in [[container-spec-cli]].

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-SPECANCH-001.1 | **AEX (proves)** | unit |
| AC-SPECANCH-001.2 | **AEX (proves)** | unit |
| AC-SPECANCH-001.5 | **AEX (proves)** | unit |
| AC-SPECANCH-001.6 | **SREF (proves)**, AEX | integration |
| AC-SPECANCH-002.1 | **CHK (proves)**, ARES, FIND | unit |
| AC-SPECANCH-002.2 | **ARES (proves)**, HID | unit |
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

- [ ] Every delta above is folded into its owning blueprint:
  - [ ] [[module-document-core]] — References, Markdown Structure,
    Markdown Links, Findings and Corpus updated
  - [ ] [[module-spec-profile]] — the registered frontmatter schema's
    citation type recorded
  - [ ] [[container-spec-cli]] — Command Surface updated for the
    anchor-qualified `refs` form
- [ ] New ADRs are referenced from the owning blueprints' Key Decisions
  (none authored by this feature)
- [ ] Every AC in the coverage table is verified at its declared level
  and cited from its tests per [[ac-traceability-standards]]
- [ ] Every issue listed as subsumed in Tracker Reconciliation is closed
  or re-scoped (none; `agent-plugins-bem` is a prerequisite and
  `agent-plugins-z8r` and `agent-plugins-im99` are out-of-scope)
- [ ] [[frd-spec-markdown-links]]'s revised fragment criteria are
  verified by the HID delta's tests, since this feature is what makes
  them true
- [ ] This blueprint's status → retired

## Open Questions

- [ ] Stage 2 is gated on a heading tree that does not exist, which
  contradicts the FRD's statement that reporting `unverified` is what
  lets this feature ship before the heading data does. Both cannot
  hold: `spec check --strict` fails on `unverified`, and this
  repository's own gates run it, so shipping stage 2 first turns the
  corpus's nine anchored references into a red build. Either gate stage
  2 on [[frd-spec-structure-conformance]]'s heading extraction
  (recommended, and what Sequencing Constraints currently states), or
  ship stage 2 first and carve the anchor rules out of `--strict` until
  headings arrive. The human reviewer decides; the FRD's example prose
  changes with it.

## References

- [[frd-spec-anchor-references]] — the FRD this blueprint realizes
- [[frd-spec-reference-integrity]] — the parent feature, whose
  extraction and resolution this design extends
- [[frd-spec-structure-conformance]] — supplies the heading tree stage 2
  matches against
- [[frd-spec-markdown-links]] — the rewriter HID makes a consumer of the
  shared derivation
- [[c4-standards]] — the container, module and component definitions the
  delta blocks are written against
- [[ac-traceability-standards]] — the citation obligation the coverage
  table feeds
