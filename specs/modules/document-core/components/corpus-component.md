---
title: "Corpus"
entity_name: Corpus
technology: Go package
description: The spec-root walk, corpus membership, the index and its warm
  cache, and the queries and checks that read them. Change it to add a
  corpus query, a check rule, an index property, or an input the cache must
  hash.
slug: component-corpus
type: component-blueprint
status: ready-for-review
kind: standard
contracts_consumed: []
source_paths:
  - "cli/core/corpus"
  - "cli/core/specindex"
test_paths:
  - "cli/core/corpus"
  - "cli/core/specindex"
---

# Corpus

## Responsibility

Corpus turns a spec root into the index every other capability reads. It
walks the root once, decides which files are members, has each member
parsed, resolves the reference graph over the finished index, and persists
the result so the next invocation can serve it warm. Every query and check
the CLI offers is a method here, so every caller asks the same question the
same way. It orchestrates parsing and holds no parsing logic of its own.

### Owns

The walk of the spec root and the membership decision: which files are
entries, which are exclusions and why, and which are feature sources.

The index: populating the generated `specmodel.SpecIndex`, and the
`schema_version` it declares, held in `cli/core/specindex`.

The warm cache: where the index is written, when a persisted index is
served, and the input hash that decides it.

The corpus queries and the check rules, and the order of what they return.

The `Profile` extension point and the validation of a profile's
registration.

### Hides

Whether an answer came from a fresh walk or a persisted index. A caller
asks `Open` for a corpus and can't tell the two apart; the hash inputs, the
rebuild conditions and the on-demand reads of a warm corpus are all
internal.

### Does not own

Parsing frontmatter, Markdown structure, Gherkin and diagrams —
[[component-frontmatter]], [[component-markdown-structure]],
[[component-gherkin]] and [[module-document-core#C4 Diagrams]] own it.

Extracting and resolving references — [[component-references]] owns it.

The finding model and the order findings are reported in —
[[module-document-core#Findings]] owns it.

The document vocabulary the walk and the checks read —
[[module-spec-profile]] owns it.

Rendering an answer and choosing an exit status — [[container-spec-cli]]
owns it, and also owns whether the emitted index becomes a Contract.

### Composed into

[[module-document-core]] — a Go package the command surface calls into; see
that blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of `cli/core/corpus`, entered through `Build` and
`Open`; the extension point is `Profile` in `cli/core/corpus/profile.go`.

### Provides

WHEN `Open` serves a persisted index, the component SHALL answer every query
and check exactly as `Build` over the same inputs would.

The component SHALL order entries, exclusions, references and features by
repository-relative path and then by position within the file.

The component SHALL return every check's findings in one `finding.Report`.

IF `CheckFor` is given a path no member occupies, THEN the component SHALL
return `ErrNotAMember` rather than an empty result.

IF a query is given a document type or qualifier the active profile doesn't
register, THEN the component SHALL return `ErrUnregisteredType` or
`ErrUnregisteredQualifier` rather than an empty result.

### Requires

WHEN `Build`, `Open` or `Fingerprint` is called, the caller SHALL supply the
spec root, the repository root, the workspace identifier and the active
profile in `Options`.

WHEN `CheckFor` is to report findings the named document's change caused
elsewhere, the caller SHALL pass the slugs that document declared at the
comparison base.

## Design

### Patterns and principles

Service Layer (Fowler, *PoEAA*, 2002) — `Corpus` is the service layer: each
application operation (build, open, query, check) starts and ends in one of
its methods, with the parsing components behind it. It serves the command
surface today, which by [[container-spec-cli]]'s Change guidance may hold no
domain logic.

### Invariants

The component SHALL decide whether a persisted index is current by a
content hash over every source file's repository-relative path and bytes,
never by modification time or absolute path. Enforced by:
`TestAModificationTimeChangeWithIdenticalContentIsStillServedWarm` and
`TestTheSameContentAtADifferentAbsolutePathIsServedWarm`.
**Rationale:** a fresh CI checkout gives every file a new modification
time and identical bytes, so a cache keyed on it would be cold on every CI
run, where the warm-index latency of AC-SPEC-009.1 is claimed.

`Fingerprint` SHALL hash exactly the sources and profile inputs the build
records a hash over. Enforced by:
`TestTheFingerprintMatchesTheHashTheBuildRecords` and
`TestTheBuildRecordsTheHashFingerprintComputesOverBothKindsOfSource`.

IF a persisted index wasn't built from exactly the current inputs, THEN
`Open` SHALL rebuild rather than serve it. Enforced by:
`TestAnAbsentIndexRebuilds`, `TestAChangedInputRebuilds` and
`TestAnIndexThatCannotBeTrustedRebuilds`.
**Rationale:** invalidation is whole-index, never per entry. A rebuild
costs one corpus walk, and a partial invalidation would save a fraction of
that while adding a way to be subtly wrong.

WHILE serving a persisted index, the component SHALL answer reference and
heading queries without reading a document. Enforced by:
`TestAWarmCorpusServesTheGraphFromThePersistedIndex` and
`TestAWarmCorpusAnswersForHeadingsWithoutReadingTheDocument`.

The component SHALL read every source through one choke point that rewrites
CRLF line endings to LF before any other component sees the bytes. Enforced
by: `TestCRLFCheckoutProducesIdenticalOutput`.
**Rationale:** so a checkout under `core.autocrlf` can't move a span or the
content hash. The consequence is deliberate: where the two differ, a
reported byte offset indexes the normalized text rather than the file on
disk, while line and column are unaffected.

The component SHALL read only regular files beneath the spec root, through
an `os.Root` handle. Enforced by: `TestBuildReadsNothingOutsideTheSpecRoot`
and `TestBuildDoesNotFollowASymlinkOutOfTheSpecRoot`.

The component SHALL order what it returns by its own rule, independent of
directory iteration order. Enforced by:
`TestEntriesExclusionsAndFindingsAreOrderedDeterministically` and
`TestCheckOutputIsDeterministic`.

WHEN a check needs Gherkin facts the index doesn't record, the component
SHALL parse every Gherkin source once per `Check` or `CheckFor` call, with
the same extraction functions the build uses. Enforced by: review only.
**Rationale:** the cost is real and grows with the corpus, not with the one
file `CheckFor` names. Confining the parse to the cold path would report
nothing in the common warm case, which is a worse defect than the cost.

The component SHALL hold no parsing logic: every parse is a call into a
parsing component. Enforced by: review only.

### Extension points

A new query — a method on `Corpus` that reads the index. Not by a filter or
a projection written in the command surface.

A new check rule — a function folded into the report `Check` and `CheckFor`
build, with its rule identifier in `schemas/rules.strata`, its detail variant
in `schemas/finding.strata`, and its constructor in
[[module-document-core#Findings]]. The same change SHALL add its identifier
to `checkRules` in `cli/cmd/check.go`, so `--only` can name it and a clean
run reports it as passed. Not by a second walk of the spec root.

A new index property — an additive change to `schemas/index.strata`,
populated on the existing walk. Not by a side file beside the index.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN the index starts recording something derived from an input other than
a source file's bytes, that input SHALL be folded into the input hash in the
same change.

WHEN a profile registration member is read only by checks that run against
the live profile, it SHOULD stay out of the input hash, so a change to it
doesn't force a rebuild.

WHEN the index changes what it already carries, `specindex.SchemaVersion`
SHALL change in the same change, under [[module-document-core]]'s Change
guidance.

### Verification gates

#### Verify always

`cd cli && go test ./core/corpus/ ./core/specindex/`

#### Verify conditionally

WHEN a change alters what a command prints, run
`cd cli && go test ./cmd/ -run TestSnapshotCases`.

WHEN a change alters the walk, a check or the registration, run
`cd cli && go test ./profile/` over this repository's corpus.

## Key decisions

[[adr-016]] — the index model is generated from `schemas/index.strata`.

[[adr-018]] — references are one top-level edge collection on the index,
with both directions derived when it's read.

[[adr-008]] — record blocks are recorded by tag and span, and decoded on
demand.

[[adr-030]] — the architecture vocabulary the C4 checks read arrives through
`Profile`.

## Open questions

- [ ] Whether to record as an ADR that the index cache is invalidated by
      content hash, never by modification time. The rule and its rationale
      are settled under Invariants; only the record is open. Blocks
      nothing; the module's owner decides.

## References

[[module-document-core]] — the module this component is composed into.
