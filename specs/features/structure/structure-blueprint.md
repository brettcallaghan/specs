---
title: "Spec Corpus Structure Extraction"
description: Builds the heading tree with section spans, attaches parsed
  record blocks to the headings that contain them, carries both in the
  index, and answers a heading query
slug: feature-spec-structure
type: feature-blueprint
status: retired
delivery: staged
frd: frd-spec-structure
---

# Spec corpus structure extraction

**FRD:** [[frd-spec-structure]]

## Overview

`Markdown Structure` already locates fenced blocks and computes every position
through `Spans`. Its heading tree is the half not yet built. This design builds
that tree, gives `IndexEntry` a place to carry the result, and adds the command
that answers a heading query. Record blocks ride the same tree: a parsed block
attaches to the heading whose section contains it. They need a tag vocabulary
that no profile registers yet, so the extension point for one is part of this
design, and the registration that fills it isn't.

Nothing here checks anything. This design raises no finding and changes no exit
status, so a consumer can depend on this feature without inheriting a build
gate.

Delta blocks in dependency order: MODEL, HEAD, IDX, CMD, XPT, REC.

## Element deltas

### Structure in the index model (added) — MODEL

**Intent:** `IndexEntry` carries a path, a frontmatter envelope, and a parse
error, and has no place for headings or record blocks. The index crosses a
boundary, so this delta declares the shape in Strata and transpiles it like
every other model rather than assembling it in the emitter.

**Behavior:**
- `IndexEntry` SHALL carry the document's headings in document order.
- A recorded heading SHALL carry its text as authored, its level, the
  span of the heading line itself, and the span of the content beneath
  it.
- A recorded heading SHALL carry the record blocks attached to it, and
  `IndexEntry` SHALL carry those attached to the document rather than to
  any heading.
- A recorded record block SHALL carry its tag and the span of the block
  as written, and SHALL NOT carry its content. Strata has no open value
  type, so an arbitrary YAML record has no typed home in the generated
  model. A command that needs a record SHALL read it back through the
  span and decode it into its own Go type; `specs check` validates it
  against the record-block criteria of
  [[frd-spec-structure-conformance]].
- The index SHALL carry no section content, only the span of it. A
  reader that wants the text reads it back through the span, which is
  what the CMD delta's emit form does.
- The `schema_version` SHALL NOT change for this delta: every property it
  adds is an addition, which the container blueprint's additive rule
  admits without a bump.

**Anchors:** `schemas/index.strata`; the generated
`cli/core/schema/spec-model.schema.json` and `cli/core/specmodel`.

### Headings in [[module-document-core]] (added) — HEAD

**Intent:** The component's own blueprint records the heading tree as not yet
built. This delta builds it: every heading with the span of the content beneath
it, and the identifier a markdown renderer would derive from it. The identifier
lives here because it's a property of a heading, and because this feature's own
query is the first thing to need it. [[frd-spec-anchor-references]] then matches
against the same derivation rather than a second one.

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

**Intent:** `Corpus` orchestrates the parsing components and holds no parsing
logic of its own. This delta calls heading extraction during the walk, populates
the entries, and keeps the emitted index byte-identical across platforms.

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

**Intent:** The query makes agent review auditable: two runs given the same
selection read the same spans. The command follows the surface's existing
discipline: open warm, convert flags into one core call, and carry the result
back out with no domain logic in the command.

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
- The command's output SHALL conform to [[cli-output-standards]]: the
  matches render as YAML by default and as JSON under `--json`, and the
  emitted content is a document payload written as itself.
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

**Intent:** A fenced block is a record block only where the active profile says
its info string is one. The host must learn that without learning what a
`component` block means, so the `Profile` interface gains a registration and
nothing else. The vocabulary that fills it belongs to
[[frd-spec-structure-conformance]].

**Behavior:**
- The `Profile` interface SHALL let a profile register record-block
  tags, each with an optional JSON Schema the record is validated
  against. A record is YAML; there is no per-tag format.
- The module SHALL NOT name any particular tag, so a profile that
  registers none yields a corpus with no record blocks rather than an
  error.
- WHERE a profile registers a tag whose schema pointer resolves to
  nothing, registration validation SHALL report it rather than skipping
  validation.

**Anchors:** the `Profile` interface in `cli/core/corpus/profile.go`;
`corpus.ValidateRegistration`.

### Record blocks in [[module-document-core]] (added) — REC

**Intent:** `Fenced` already returns the blocks whose info string matches a
registered tag. Two things are missing: recording each one where a reader would
look for it (the heading whose section contains it), and a way to read its
record back.

**Behavior:**
- The component SHALL record a registered block's tag and span, and
  attach the record to the innermost heading whose section span contains
  it, and to the document where no heading does.
- The component SHALL expose one function that reads a recorded block's
  content through its span and decodes it as YAML into a caller-supplied
  Go value, so that every command and check decodes a record the same
  way.
- The index SHALL record a block whether or not its content is valid;
  judging the content is the check's job, and a block that fails it is
  still a place a reader can be sent.

**Anchors:** `Fenced` in `cli/core/markdown`; the finding constructors
in `cli/core/finding`.

### Relied-upon elements (unchanged)

- `Spans` in [[module-document-core]] — every position this feature
  records, heading and block alike.
- `Frontmatter` in [[module-document-core]] — the body offset a heading
  search starts from.
- [[container-spec-cli#Command Surface]] — the warm corpus open and the
  flags-to-one-core-call discipline CMD follows.
- [[contract-specs-configuration]] — the configuration surface. This feature
  adds no key to it.
- [[frd-spec-structure-conformance]] — registers the record-block tag
  vocabulary XPT makes registrable.

## Contracts

- No new contract document. The index gains properties under the additive rule
  that [[container-spec-cli]] states for `schema_version`. That rule anticipates
  exactly this case: later features attaching headings and extracted record
  blocks.
- [[contract-specs-configuration]] (revised, drafted in place) — the contract's
  description of the internal registration surface lists what a profile supplies
  and doesn't yet mention record-block tags. The revision is additive and
  touches only the internal half, which that contract's change policy exempts
  while both modules are `visibility: internal`.

## Tracker reconciliation

- Out-of-scope: `IndexEntry.frontmatter` carries the shared envelope, so
  per-type properties are lost. MODEL adds properties beside `frontmatter`. It
  must not change how `frontmatter` is carried, and it must not be read as
  fixing that loss.
- Out-of-scope: the index schema and versioning ADR. MODEL relies on the
  additive-at-1.0 rule that [[container-spec-cli]] already states and adds no
  policy of its own.

## Sequencing constraints

- `delivery: staged`. Two stages, each a reviewable increment:
  - **Stage 1 — headings:** MODEL, HEAD, IDX, CMD. The index carries
    every document's headings with their section spans and the query
    answers. This is the whole of what [[frd-spec-anchor-references]]
    needs, and it depends on no configuration.
  - **Stage 2 — record blocks:** XPT, REC. The component parses and
    attaches blocks once a profile registers a tag.
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
  - [[module-document-core]] requires the schema version and the emitted index
    to change together when the index model changes shape. Both belong to MODEL,
    which deliberately doesn't bump the version: every property is an addition
    under the additive-at-1.0 rule. The recorded no-bump is therefore the
    satisfying act, not an omission.
  - [[module-document-core]] requires a new parsing capability to be built on
    the Spans delimiter scanner, narrowed for a capability that needs to know
    which regions are not to be scanned. HEAD takes heading recognition from
    `goldmark` for that reason (a heading inside a fence isn't a heading) and
    computes every position through `Spans`.
  - [[container-spec-cli]] requires the snapshot regeneration script to be run
    and every diff reviewed when observable output changes. CMD adds a command,
    and IDX changes the emitted index. Each lands with its regenerated snapshots
    in the same change.
- No backward-compatibility window applies. [[module-document-core]] and
  [[module-spec-profile]] are both `visibility: internal` and evolve in
  lockstep with their only consumer.

## AC coverage

Legend: MODEL = the structure properties on the index model; HEAD =
`Markdown Structure` (headings, section spans, and identifiers); IDX =
`Corpus` population of those properties; CMD = Command Surface in
[[container-spec-cli]]; XPT = the record-block tag registration on the
`Profile` extension point; REC = `Markdown Structure` (record-block parsing
and attachment).

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

## Key decisions

- [[adr-016]] — Strata is the single source for generated models, which
  is why MODEL is a schema delta rather than a hand edit to Go.
- [[adr-008]] — element record blocks are the structured declarations
  REC parses and attaches.
- [[adr-015]] — Go and `goldmark`, which HEAD uses for heading
  recognition as `Fenced` already uses it for fences.

## Definition of done

- [x] Every delta above is folded into its owning blueprint:
  - [x] [[module-document-core]] — Markdown Structure, Corpus, and the
    Profile extension point updated, and the index model's vocabulary
    extended
  - [x] [[container-spec-cli]] — Command Surface updated for the heading
    query
- [x] [[contract-specs-configuration]]'s internal registration surface
  names the record-block tag registration
- [x] New ADRs are referenced from the owning blueprints' Key decisions
  (none authored by this feature)
- [x] Every AC in the coverage table is verified at its declared level
  and cited from its tests per [[ac-traceability-standards]]
- [x] Every issue listed as subsumed in Tracker reconciliation is closed
  or re-scoped (none: both listed issues are out of scope)
- [x] This blueprint's status → retired

## Open questions

None.

## References

- [[frd-spec-structure]] — the FRD this blueprint realizes
- [[frd-spec-structure-conformance]] — registers the record-block tags
  XPT makes registrable, and checks documents against what this feature
  records
- [[frd-spec-anchor-references]] — the consumer stage 1 unblocks
- [[frd-spec-index]] — the index, corpus discovery, and warm cache that this
  feature extends
- [[c4-standards]] — the container, module, and component definitions that the
  delta blocks are written against
- [[cli-output-standards]] — what the heading query prints, and in
  which rendering
- [[ac-traceability-standards]] — the citation obligation the coverage
  table feeds
