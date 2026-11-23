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

Corpus turns the spec root (the one directory of specification documents
the CLI reads) into the index that every other capability reads. It walks
the root once, decides which files are corpus documents, has a parsing
component parse each one, and resolves the reference graph over the
finished index. It then persists the index, so the next CLI invocation can
serve it warm: answer from the persisted index instead of walking the root
again. Every query and check the CLI offers is a function or method of this
package, so every caller asks the same question the same way. It
orchestrates parsing and holds no parsing logic of its own.

### Owns

The walk of the spec root and the membership decision: which Markdown
files are entries (corpus documents, which the index records), which are
exclusions (candidates the index rejects, with the reason), and which
`.feature` files are feature sources.

The index: populating `specmodel.SpecIndex` (the struct generated from
the index schema), and the `schema_version` property it declares, whose
value is the `specindex.SchemaVersion` constant in `cli/core/specindex`.

The warm cache: where the index is written, when a persisted index is
served instead of rebuilt, and the input hash (a content hash over every
input the index is built from) that decides it.

The corpus queries and the check rules, and the order of what they return.

The `Profile` interface, through which the caller supplies a profile (the
document vocabulary the walk and the checks read), and validation of what a
profile registers.

### Hides

Whether an answer came from a fresh walk or a persisted index. A caller
asks the `Open` function for a corpus and gets the same answers either way.
`Open` reports which path it took, and no answer the corpus gives depends
on it. The hash inputs, the rebuild conditions, and the files a warm corpus
reads on demand are all internal.

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

[[module-document-core]] — a Go package that
[[container-spec-cli#Command Surface]] (the CLI's command layer, `cli/cmd`)
calls into; see that blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of the `cli/core/corpus` package, entered through
the `Build` and `Open` functions. The extension point is the `Profile`
interface in `cli/core/corpus/profile.go`.

### Provides

WHEN `Open` serves a persisted index, the component SHALL answer every query
and check exactly as `Build` over the same inputs would.

The component SHALL order entries, exclusions, references and features by
repository-relative path and then by position within the file.

The component SHALL return every check's findings in one list, ordered by
`finding.Report`.

IF `CheckFor` is given a path no corpus document occupies, THEN the
component SHALL return `ErrNotAMember` rather than an empty result.

IF a query is given a document type or qualifier the active profile doesn't
register, THEN the component SHALL return `ErrUnregisteredType` or
`ErrUnregisteredQualifier` rather than an empty result.

### Requires

WHEN `Build`, `Open` or `Fingerprint` is called, the caller SHALL supply the
spec root, the repository root, the workspace identifier and the active
profile in `Options`.

WHEN `CheckFor` is to report findings the named document's change caused
elsewhere, the caller SHALL pass the slugs that document declared at the
base revision.

## Design

### Patterns and principles

Service Layer (Fowler, *PoEAA*, 2002) — the `corpus` package is the
service layer: each application operation (build, open, query, check)
starts and ends in its `Build` or `Open` function or a method on the
`Corpus` type, with the parsing components behind it. It serves Command
Surface, which by [[container-spec-cli]]'s Change guidance may hold no
domain logic.

### Invariants

The component SHALL decide whether a persisted index is current by a
content hash over every source file's repository-relative path and bytes,
never by modification time or absolute path ([[adr-035]]).
**Rationale:** a fresh clone or checkout changes every modification time and
no bytes, and an edit can keep a file's modification time.
Enforced by:
`TestAModificationTimeChangeWithIdenticalContentIsStillServedWarm` and
`TestTheSameContentAtADifferentAbsolutePathIsServedWarm`.

`Fingerprint` SHALL hash exactly the sources and profile inputs the build
records a hash over. Enforced by:
`TestTheFingerprintMatchesTheHashTheBuildRecords` and
`TestTheBuildRecordsTheHashFingerprintComputesOverBothKindsOfSource`.

IF a persisted index wasn't built from exactly the current inputs, THEN
`Open` SHALL rebuild rather than serve it.
**Rationale:** a rebuild costs one walk of the corpus, and invalidating single
entries would add a way to serve a subtly wrong index ([[adr-035]]).
Enforced by: `TestAnAbsentIndexRebuilds`, `TestAChangedInputRebuilds` and
`TestAnIndexThatCannotBeTrustedRebuilds`.

WHILE serving a persisted index, the component SHALL answer reference and
heading queries without reading a document. Enforced by:
`TestAWarmCorpusServesTheGraphFromThePersistedIndex` and
`TestAWarmCorpusAnswersForHeadingsWithoutReadingTheDocument`.

The component SHALL read every source through one read function that rewrites
CRLF line endings to LF before any other component sees the bytes.
**Rationale:** with every read normalized, a checkout under Git's
`core.autocrlf` setting can't move a span or change the content hash. The
consequence is deliberate: where a file on disk has CRLF endings, a reported
byte offset indexes the normalized text, not the file. Line and column are
unaffected.
Enforced by: `TestCRLFCheckoutProducesIdenticalOutput`.

The component SHALL read only regular files beneath the spec root, through
an `os.Root` handle. Enforced by: `TestBuildReadsNothingOutsideTheSpecRoot`
and `TestBuildDoesNotFollowASymlinkOutOfTheSpecRoot`.

The component SHALL order what it returns by its own rule, independent of
directory iteration order. Enforced by:
`TestEntriesExclusionsAndFindingsAreOrderedDeterministically` and
`TestCheckOutputIsDeterministic`.

WHEN a check needs Gherkin facts the index doesn't record, the component
SHALL parse every Gherkin source once per `Check` or `CheckFor` call, with
the same extraction functions the build uses.
**Rationale:** the cost is real and grows with the corpus, not with the one
file `CheckFor` names. Confining the parse to a cold build (one that walks
the corpus) would leave these checks reporting nothing when the index is
served warm, the common case. That is a worse defect than the cost.
Enforced by: review only.

The component SHALL hold no parsing logic: every parse is a call into a
parsing component. Enforced by: review only.

### Extension points

A new query — a method on the `Corpus` type that reads the index. Not by a
filter or a projection written in Command Surface.

A new check rule — a function whose findings the `Check` and `CheckFor`
methods add to the report they build. Its rule identifier goes in
`schemas/rules.strata`, its detail variant in `schemas/finding.strata`, and
its finding constructor in [[module-document-core#Findings]]. The same
change SHALL add its identifier to the `checkRules` variable in
`cli/cmd/check.go`, so `--only` can name it and a clean run reports it as
passed. Not by a second walk of the spec root.

A new index property — an additive change to `schemas/index.strata`,
populated during the existing walk. Not by a separate file written beside
the index.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN the index starts recording something derived from an input other than
a source file's bytes, that input SHALL be folded into the input hash in the
same change.

WHEN a part of a profile's registration is read only by checks that run
against the live profile, it SHOULD stay out of the input hash, so a change
to it doesn't force a rebuild.

WHEN the index changes what it already carries, `specindex.SchemaVersion`
SHALL change in the same change, under [[module-document-core]]'s Change
guidance.

### Verification gates

#### Verify always

`cd cli && go test ./core/corpus/ ./core/specindex/`

#### Verify conditionally

WHEN a change alters what a command prints, run
`cd cli && go test ./cmd/ -run TestSnapshotCases`.

WHEN a change alters the walk, a check or what a profile registers, run
`cd cli && go test ./profile/` over this repository's corpus.

## Key decisions

[[adr-016]] — the index model is generated from `schemas/index.strata`.

[[adr-018]] — references are one top-level edge collection on the index,
with both directions derived when it's read.

[[adr-008]] — record blocks are recorded by tag and span, and decoded on
demand.

[[adr-030]] — the architecture vocabulary the C4 checks read arrives through
`Profile`.

[[adr-035]] — a persisted index is served only while a content hash of every
source file's repository-relative path and bytes matches the hash it records.

## References

[[module-document-core]] — the module this component is composed into.
