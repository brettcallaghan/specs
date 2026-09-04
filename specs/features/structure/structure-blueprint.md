---
title: "Feature Blueprint: Spec Corpus Structure Extraction"
description: Builds the heading tree with section spans, attaches parsed
  record blocks to the headings that contain them, carries both in the
  index, and answers a heading query
slug: feature-spec-structure
type: feature-blueprint
status: draft
updated_at: 2026-09-04T00:00:00Z
delivery: staged
frd: frd-spec-structure
---

# Spec Corpus Structure Extraction

**FRD:** [[frd-spec-structure]]

## Overview

`Markdown Structure` already locates fenced blocks and computes every
position through `Spans`; its heading tree is the half that was never
built. This design builds it, gives `IndexEntry` somewhere to carry the
result, and adds the command that answers a heading query. Record blocks
ride the same tree — a parsed block attaches to the heading whose
section contains it — but they need a tag vocabulary no profile
registers yet, so the extension point for one is part of this design and
the registration that fills it is not.

Nothing here checks anything. No finding is raised and no exit status
changes, which is what lets a consumer depend on this feature without
inheriting a build gate.

Delta blocks in dependency order: MODEL, HEAD, IDX, CMD, XPT, REC.

## Element Deltas

### Structure in the index model (added) — MODEL

**Intent:** `IndexEntry` carries a path, a frontmatter envelope and a
parse error. Headings and record blocks have nowhere to go. The index
crosses a boundary, so the shape is declared in Strata and transpiled
like every other model rather than assembled in the emitter.

**Behavior:**
- `IndexEntry` SHALL carry the document's headings in document order.
- A recorded heading SHALL carry its text as authored, its level, the
  span of the heading line itself, and the span of the content beneath
  it.
- A recorded heading SHALL carry the record blocks attached to it, and
  `IndexEntry` SHALL carry those attached to the document rather than to
  any heading.
- A recorded record block SHALL carry its tag, its parsed content, and
  the span of the block as written. The parsed content is what a
  catalogue reads an element declaration from; measured on this corpus,
  thirteen blocks carry 1,991 bytes against a 433 KB index, beside the
  153 KB the parsed Gherkin features already occupy.
- The index SHALL carry no section content, only the span of it. A
  reader that wants the text reads it back through the span, which is
  what the CMD delta's emit form does.
- The `schema_version` SHALL remain `1.0`: every property this delta
  adds is an addition, which the container blueprint's additive rule
  admits without a bump.

**Anchors:** `schemas/index.strata`; the generated
`cli/core/schema/spec-model.schema.json` and `cli/core/specmodel`.

### Headings in [[module-document-core]] (added) — HEAD

**Intent:** The component's own blueprint records the heading tree as
not yet built. This delta builds it: every heading with the span of the
content beneath it, and the identifier a markdown renderer would derive
from it. The identifier lives here because it is a property of a
heading, and because the first thing to need it is this feature's own
query; [[frd-spec-anchor-references]] then matches against the same
derivation rather than a second one.

**Behavior:**
- The component SHALL return every heading of a document body in
  document order, each with its level and its text as authored.
- A heading's section span SHALL run from the end of its heading line to
  the start of the next heading of the same or a higher level, and to
  the end of the document where there is none.
- The component SHALL NOT report a heading written inside a fenced code
  block or an inline code span.
- Every span the component returns SHALL be computed through `Spans`.
- WHERE a document carries no heading, the component SHALL return an
  empty result rather than an error.
- The component SHALL expose the heading identifier of a text: its
  rendered inline content lowercased, with every character that is not
  a letter, digit, hyphen, underscore or space removed, and each
  remaining space replaced by a hyphen.
- The derivation SHALL be a pure function of the text, deriving no
  positional suffix for a repeated heading.

**Anchors:** `cli/core/markdown` (alongside `Fenced`); `goldmark` for
heading recognition, as `Fenced` already uses it for fences; `Of` in
`cli/core/span`.

### Corpus in [[module-document-core]] (modified) — IDX

**Intent:** `Corpus` orchestrates the parsing components and holds no
parsing logic of its own. This delta calls the heading extraction during
the walk and populates the entries, keeping the emitted index
byte-identical across platforms.

**Behavior:**
- WHEN the corpus is built, `Corpus` SHALL populate each member's
  headings from `Markdown Structure`.
- `Corpus` SHALL order every member's headings by position in the
  document.
- WHERE a member's frontmatter could not be parsed, `Corpus` SHALL still
  record that member's headings, since a heading does not depend on a
  frontmatter envelope.
- A warm corpus SHALL answer for headings from the index it opened,
  reading no document off disk.

**Anchors:** `cli/core/corpus/corpus.go`; the warm-open path in
`corpus.Open`.

### Command Surface in [[container-spec-cli]] (modified) — CMD

**Intent:** The query is what makes agent review auditable: two runs
given the same selection read the same spans. It follows the surface's
existing discipline — open warm, convert flags into one core call, and
carry the result back out with no domain logic in the command.

**Behavior:**
- The CLI SHALL expose a `headings` command answering a heading
  selection with every match's file path, heading text, heading
  identifier, level, and the span of the content beneath it.
- The selection SHALL be one operand written `<document>#<heading>`,
  either half of which MAY be omitted: `<document>` alone selects every
  heading in that document, `#<heading>` alone selects that heading in
  every corpus member, and an absent operand selects every heading in
  the corpus.
- WHERE the operand's document half holds a path separator or ends in
  `.md`, the CLI SHALL read it as a repository-relative path; otherwise
  it SHALL read it as a slug.
- The CLI SHALL match a heading selection on heading identifier.
- WHEN a selection matches a heading, the CLI SHALL also return every
  heading within that heading's section, each carrying its own level.
- WHEN the caller asks for it, the CLI SHALL emit the content each
  matched heading stands over, read through the recorded span.
- The CLI SHALL accept `--json`, as every other query command does.
- WHERE no heading matches, the CLI SHALL report that and exit zero.
- The command SHALL open the corpus warm, as `check`, `refs`,
  `slug-exists` and `next-id` do.
- The command SHALL be one flat command with no subcommand of its own,
  as every command on this surface is.
- The command's `--help` text SHALL derive from the declaration that
  defines it.

**Anchors:** `cli/cmd/registry.go` and the query commands beside it
(`cli/cmd/refs.go`, `cli/cmd/feature.go`); the snapshot cases under
`cli/testdata/cases/`.

### Profile extension point in [[module-document-core]] (modified) — XPT

**Intent:** A fenced block is a record block only where the active
profile says its info string is one. The host must learn that without
learning what a `component` block means, so the `Profile` interface
gains a registration and nothing else: the vocabulary that fills it is
[[frd-spec-structure-conformance]]'s.

**Behavior:**
- The `Profile` interface SHALL let a profile register record-block
  tags, each with the format its content is parsed as.
- The module SHALL NOT name any particular tag, so a profile that
  registers none yields a corpus with no record blocks rather than an
  error.
- WHERE a profile registers a tag whose format the module does not
  support, registration validation SHALL report it rather than
  defaulting.

**Anchors:** the `Profile` interface in `cli/core/corpus/profile.go`;
`corpus.ValidateRegistration`.

### Record blocks in [[module-document-core]] (added) — REC

**Intent:** `Fenced` already returns the blocks whose info string
matches a registered tag. What is missing is parsing the content in the
tag's declared format and attaching the result where a reader would look
for it: the heading whose section contains it.

**Behavior:**
- The component SHALL parse a registered block's content in the format
  its tag declares.
- The component SHALL attach a parsed block to the innermost heading
  whose section span contains it, and to the document where no heading
  does.
- WHERE a registered block's content does not parse, the component SHALL
  report a finding carrying the parse error at the position the block
  opens, and SHALL continue with the rest of the document.

**Anchors:** `Fenced` in `cli/core/markdown`; the finding constructors
in `cli/core/finding`.

### Relied-upon elements (unchanged)

- `Spans` in [[module-document-core]] — every position this feature
  records, heading and block alike.
- `Frontmatter` in [[module-document-core]] — the body offset a heading
  search starts from.
- [[container-spec-cli#Command Surface]] — the warm corpus open and the
  flags-to-one-core-call discipline CMD follows.
- [[contract-specs-configuration]] — the configuration surface; this
  feature adds no key to it.
- [[frd-spec-structure-conformance]] — registers the record-block tag
  vocabulary XPT makes registrable.

## Contracts

- No new contract document. The index gains properties under the
  additive rule [[container-spec-cli]] states for `schema_version` 1.0,
  which anticipates exactly this: later features attaching headings and
  extracted record blocks.
- [[contract-specs-configuration]] (revised, drafted in place) — the
  contract's description of the internal registration surface enumerates
  what a profile supplies and does not yet mention record-block tags.
  The revision is additive and to the internal half only, which that
  contract's change policy exempts WHILE both modules are
  `visibility: internal`.

## Tracker Reconciliation

- `agent-plugins-gz3` — out-of-scope: `IndexEntry.frontmatter` carries
  the shared envelope, so per-type properties are lost. MODEL adds
  properties beside `frontmatter` and must neither change how it is
  carried nor be read as fixing it.
- `agent-plugins-h6c` — out-of-scope: the index schema and versioning
  ADR. MODEL relies on the additive-at-1.0 rule
  [[container-spec-cli]] already states and adds no policy of its own.

## Sequencing Constraints

- `delivery: staged`. Two stages, each a reviewable increment:
  - **Stage 1 — headings:** MODEL, HEAD, IDX, CMD. The index carries
    every document's headings with their section spans and the query
    answers. This is the whole of what [[frd-spec-anchor-references]]
    needs, and it depends on no configuration.
  - **Stage 2 — record blocks:** XPT, REC. Blocks are parsed and
    attached once a profile registers a tag.
- The MODEL delta SHALL land before HEAD, IDX, CMD and REC, all of which
  populate or read the properties it declares.
- The HEAD delta SHALL land before IDX and CMD, which carry and answer
  for what it extracts, and before REC, which attaches to its headings.
- The XPT delta SHALL land before REC, which reads the registration it
  admits.
- Stage 2 SHALL land after the record-block tag registration
  [[frd-spec-structure-conformance]] specifies in FEAT-SPECSTR-004,
  because a registration point no profile fills leaves REC inert.
- **Lockstep rules spanning delta blocks, and how they are satisfied:**
  - [[module-document-core]] requires the schema version and the emitted
    index to change together when the index model changes shape. Both
    are MODEL's, and the version is deliberately not bumped: every
    property is an addition under the additive-at-1.0 rule, so the
    satisfying act is the recorded no-bump rather than an omission.
  - [[module-document-core]] requires a new parsing capability to be
    built on the Spans delimiter scanner, narrowed for a capability that
    needs to know which regions are not to be scanned. HEAD takes
    heading recognition from `goldmark` for that reason — a heading
    inside a fence is not a heading — and computes every position
    through `Spans`.
  - [[container-spec-cli]] requires the snapshot regeneration script to
    be run and every diff reviewed when observable output changes. CMD
    adds a command and IDX changes the emitted index; each lands with
    its regenerated snapshots in the same change.
- No backward-compatibility window applies. [[module-document-core]] and
  [[module-spec-profile]] are both `visibility: internal` and evolve in
  lockstep with their only consumer.

## AC Coverage

Legend: MODEL = the structure properties on the index model; HEAD =
`Markdown Structure` (headings, section spans and identifiers); IDX = `Corpus`
population of those properties; CMD = Command Surface in
[[container-spec-cli]]; XPT = the record-block tag registration on the
`Profile` extension point; REC = `Markdown Structure` (record-block
parsing and attachment).

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-SPECSTR-001.1 | **HEAD (proves)**, MODEL, IDX | unit |
| AC-SPECSTR-001.2 | **REC (proves)**, XPT, MODEL, IDX | unit |
| AC-SPECSTR-003.1 | **CMD (proves)**, HEAD, IDX | integration |
| AC-SPECSTR-003.2 | **CMD (proves)**, HEAD | integration |
| AC-SPECSTR-003.3 | **CMD (proves)**, IDX | integration |
| AC-SPECSTR-003.4 | **CMD (proves)**, HEAD | integration |
| AC-SPECSTR-003.5 | **CMD (proves)**, HEAD | integration |
| AC-SPECSTR-003.6 | **CMD (proves)** | integration |

## Key Decisions

- [[adr-016]] — Strata is the single source for generated models, which
  is why MODEL is a schema delta rather than a hand edit to Go.
- [[adr-008]] — element record blocks are the structured declarations
  REC parses and attaches.
- [[adr-015]] — Go and `goldmark`, which HEAD uses for heading
  recognition as `Fenced` already uses it for fences.

## Definition of Done

- [ ] Every delta above is folded into its owning blueprint:
  - [ ] [[module-document-core]] — Markdown Structure, Corpus and the
    Profile extension point updated, and the index model's vocabulary
    extended
  - [ ] [[container-spec-cli]] — Command Surface updated for the heading
    query
- [ ] [[contract-specs-configuration]]'s internal registration surface
  names the record-block tag registration
- [ ] New ADRs are referenced from the owning blueprints' Key Decisions
  (none authored by this feature)
- [ ] Every AC in the coverage table is verified at its declared level
  and cited from its tests per [[ac-traceability-standards]]
- [ ] Every issue listed as subsumed in Tracker Reconciliation is closed
  or re-scoped (none; both listed issues are out-of-scope)
- [ ] This blueprint's status → retired

## Open Questions

None.

## References

- [[frd-spec-structure]] — the FRD this blueprint realizes
- [[frd-spec-structure-conformance]] — registers the record-block tags
  XPT makes registrable, and checks documents against what this feature
  records
- [[frd-spec-anchor-references]] — the consumer stage 1 unblocks
- [[frd-spec-index]] — the index, corpus discovery and warm cache this
  feature extends
- [[c4-standards]] — the container, module and component definitions the
  delta blocks are written against
- [[ac-traceability-standards]] — the citation obligation the coverage
  table feeds
