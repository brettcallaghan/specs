---
title: "Spec Corpus Index and Core Conformance Checks"
description: Builds the first working slice of the Spec CLI — the index
  model, the framework-agnostic engine packages, the specs
  profile, and the command surface — realizing the index, metadata
  checks, queries and configuration of frd-spec-index
slug: feature-spec-index
type: feature-blueprint
status: retired
delivery: staged
frd: frd-spec-index
---

# Spec Corpus Index and Core Conformance Checks

## Overview

**FRD:** [[frd-spec-index]]

Today the design is complete but almost none of the code exists. The
Strata sources under `schemas/` generate the frontmatter, configuration
and profile-registration models, and three approved blueprints
([[container-spec-cli]], [[module-document-core]], [[module-spec-profile]])
describe the steady state. The Go packages those blueprints name
(`cli/core/*`, `cli/profile`, `cli/cmd`) are unwritten.

This feature builds the first working vertical slice:

- enough of [[module-document-core]] to discover a corpus and validate
  frontmatter
- enough of [[module-spec-profile]] to supply this framework's profile:
  one union schema plus a thin type table
- the command surface that wraps them and selects a profile from a
  registry
- a new **index model** in Strata that the emitted `specs.json` conforms
  to

Reference and structure checks are deferred to later features; the
packages built here are the ones those features extend.

The delta blocks in dependency order: **IDX** (index model in Strata)
→ **SP** (Spans) → **FIND** (Findings) → **FM** (Frontmatter) → **SR**
(Schema Registry) → **PROF** (specs profile) →
**CO** (Corpus discovery and index build) → **NUM** (Numbering) →
**CACHE** (index output and warm cache) → **CFG** (configuration loading
and checking) → **CLI** (command surface) → **BUILD** (static binary and
distribution).

## Element Deltas

### `SpecIndex` index model in Strata (added) — IDX

**Intent:** The emitted index (`specs.json`, AC-SPEC-010.1) has no data
model yet: the Strata sources cover frontmatter, config and profile
registration but not the artifact the CLI *writes*. Because the index is
consumed across a boundary (the Architecture Catalogue) and must carry a
versioned schema (AC-SPEC-002.2), its shape is defined once in Strata
and generated into the same JSON-Schema/Go/Markdown outputs as every
other model, per [[adr-016]]. The model is deliberately **additive**: it
holds only the frontmatter-and-path entries and exclusions this sprint
populates, with room for the later features to add headings and
references under the same `schema_version`.

**Data (Strata source, `schemas/`):** a new `.strata` file (e.g.
`index.strata`) declaring the index root and its members. Written in
the vocabulary of [[module-document-core]]'s Corpus "index model":
- A root struct carrying `schema_version` (String), `workspace` (String),
  a list of document entries, and a list of exclusions.
- A document-entry struct carrying the repository-relative `path`
  (String), the parsed `frontmatter` (the existing `DocumentFrontmatter`
  envelope or per-type struct), and an optional `parse_error` with its
  position for a member that failed to parse (AC-SPEC-002.4).
- An exclusion struct carrying `path`, a `reason` enum
  (`no-frontmatter` | `unknown-type`), and the declared `type` value
  where the reason is `unknown-type` (AC-SPEC-001.4).
- Behavior: the model SHALL declare `schema_version` as a required
  property whose value this sprint fixes at `"1.0"` (AC-SPEC-002.2).
- Behavior: entries and exclusions SHALL be lists whose element order
  the emitter controls, so cross-platform output can be made
  deterministic (AC-SPEC-009.7) without a model change.

**Anchors:** `schemas/*.strata`, `schemas/strata.config.json`,
`scripts/generate-schemas.sh` (regenerates
`cli/core/specmodel/`, `cli/core/schema/`, and the plugin markdown). The
generated Go structs land in the existing `specmodel` package the Corpus
emitter consumes.

### #Spans in [[module-document-core]] (added) — SP

**Intent:** Every position any parser reports is a `Span`, and every
parser locates delimited regions (frontmatter `---`, later comment and
fenced blocks) through one shared scanner, so positions are computed the
same way everywhere. This is the foundation the frontmatter parser and
all later parsers build on.

**Behavior:**
- The component SHALL define the `Span` type of [[frd-spec-index]]:
  a byte offset and byte length as the authoritative extent, plus start
  and end line and column for readability.
- The component SHALL provide a delimiter scanner that locates the region
  between an opening and closing delimiter and returns it as a `Span`.
- WHEN any component reports a position, it SHALL do so as a `Span`
  rather than a bare line number.

**Anchors:** new package under `cli/core` (the module blueprint names
`cli/core/corpus` for the `Profile` interface; Spans is its own package,
e.g. `cli/core/span`).

### #Findings in [[module-document-core]] (added) — FIND

**Intent:** One finding model and one severity vocabulary carry every
check's output, so the command surface renders one shape and later
features add checks without changing it. The finding JSON is a published
interface surface (an agent parses it), so its shape is defined in
Strata alongside the index model and generated, not hand-rolled in Go.
Establishing this early is what makes the AC-SPEC-006.1
self-contained-diagnostic guarantee uniform.

**Data (Strata source, `schemas/`):** a finding model (e.g.
`finding.strata`) generated into the same JSON-Schema/Go/Markdown outputs
as every other model, per [[adr-016]]:
- A base `Finding` struct carrying a stable rule identifier, a severity
  (an enum `error` | `warning` | `unverified`), a repository-relative
  path, a `Span`, and a message (AC-SPEC-006.1).
- A **discriminated union of finding variants** keyed by the rule (or a
  `kind` discriminator), each adding the structured fields that make the
  finding self-contained — e.g. a slug-pattern variant carrying the
  actual slug and the expected pattern, a missing-property variant
  carrying the property name, a duplicate-slug variant carrying the other
  document's path. The message is derivable from the variant, so a
  consumer can render or branch on structure rather than parse prose.
- A report envelope carrying, per check requested, whether it ran, so the
  JSON distinguishes "check passed" from "check not requested"
  (AC-SPEC-006.9).

**Behavior:**
- The component SHALL represent a check that was requested but could not
  run as `unverified`, never as passed (AC-SPEC-006.3, AC-SPEC-006.4).
- The component SHALL serialize the report both as human-readable text
  and as the generated finding JSON, so the command surface selects a
  rendering rather than formatting findings itself (AC-SPEC-006.5).
- *Scope note:* this sprint builds the `unverified` mechanism but
  registers no check that emits it (the mermaid example in AC-SPEC-006.3
  is a structure-feature concern), and it defines only the finding
  variants the sprint's checks raise; later features add variants to the
  union without changing the envelope.

**Anchors:** the Strata source under `schemas/` (generated into
`cli/core/specmodel`); a new package under `cli/core` (e.g.
`cli/core/finding`) for the severity rules and rendering.

### #Frontmatter in [[module-document-core]] (added) — FM

**Intent:** Split the frontmatter block off each document and parse it to
a generic model that preserves each property's raw text, feeding both
schema validation and typed decoding. Raw-text retention is what lets
AC-SPEC-003.2 validate what the author wrote rather than what YAML
inferred.

**Behavior:**
- The component SHALL extract the frontmatter region using the #Spans
  delimiter scanner and parse it to a generic document model.
- The component SHALL retain the raw source text of each property
  alongside its parsed value (AC-SPEC-003.2).
- WHEN a document's frontmatter fails to parse, the component SHALL
  return a located parse error rather than raising, so the caller records
  the failure and continues (AC-SPEC-002.4).

**Anchors:** new package under `cli/core` (e.g. `cli/core/frontmatter`);
[[adr-015]] names `goccy/go-yaml` as the parser.

### Schema Registry in [[module-document-core]] (added) — SR

**Intent:** Validate a document's frontmatter against the single union
schema the active profile supplies, dispatching by the document's `type`
to the matching branch of that union, and report every violation in one
pass. Document Core stays framework-agnostic: it compiles one schema from
whichever profile is active and never knows a `type` by name. There is no
per-type schema registration: a profile has one schema, not a registry
of them.

**Behavior:**
- The component SHALL accept one JSON Schema from the active profile — a
  discriminated union over the profile's document types — and compile it
  once.
- WHEN validating a document, the component SHALL validate its
  frontmatter against the branch of that union selected by the document's
  `type`, reporting every violation as a `Finding` with the JSON Pointer
  at which it occurred (AC-SPEC-003.1), validating the raw property text
  where the schema constrains textual format (AC-SPEC-003.2).
- WHEN a document's `slug` does not match the pattern the active profile
  supplies for its `type`, the component SHALL report a `Finding` naming
  both the actual slug and the expected pattern (AC-SPEC-003.5). The slug
  pattern is supplied by the profile's type table, not the union schema,
  because Strata cannot narrow the inherited `slug` per type.

**Anchors:** new package under `cli/core` (the implementing agent chooses
the name, distinct from the generated `cli/core/schema/*.json` outputs);
[[adr-015]] names `santhosh-tekuri/jsonschema`; the union schema is the
bundle `scripts/generate-schemas.sh` emits
(`cli/core/schema/spec-model.schema.json`).

### specs profile in [[module-spec-profile]] (added) — PROF

**Intent:** Supply this framework's vocabulary as one pluggable profile,
not a per-type registry. The profile carries a single union schema
covering every document type (the Strata bundle), a thin type table
giving each type's slug pattern and whether it is numbered, and the
convention defaults; and it exposes itself through Document Core's
`Profile` extension point. A different taxonomy is a different profile,
selected by name from the profile registry (CFG delta). That is the
customization path v1 supports, and it needs no per-type schema
plumbing.
Element records, required sections, reference properties and lifecycle
rules are added by later features; this delta supplies only what the
sprint's checks consume.

**Behavior (written in [[module-spec-profile]] vocabulary):**
- The profile SHALL expose one `New` constructor returning a value
  satisfying Document Core's `Profile` interface (AC-SPEC-008.6).
- The profile SHALL embed the bundled union schema via Go's `//go:embed`
  and supply it to the Schema Registry as the one schema all its document
  types validate against (AC-SPEC-003.1), so the binary carries the
  schema and reads no schema file at runtime (AC-SPEC-009.4).
- The profile SHALL supply, per document type in the `DocumentType`
  vocabulary, its slug pattern and whether it is numbered — a thin table,
  not a per-type schema (AC-SPEC-003.5, AC-SPEC-007.1).
- The profile SHALL supply the `line_width` default the CLI resolves when
  `specs.toml` omits `[conventions].line_width`
  ([[contract-specs-configuration]] resolution).
- The profile SHALL register under the name `specs` so an
  adopting repository activates it without authoring one (AC-SPEC-008.6).

**Anchors:** new package `cli/profile`; the embedded schema is the
generated `cli/core/schema/spec-model.schema.json`. The per-type table
maps onto the existing `slug_pattern` and `numbered` fields of
`DocumentTypeRegistration` in `cli/core/specmodel`; that struct's
`schema` field is unused in v1, since one union schema serves every type.
This collapses [[module-spec-profile]]'s separate Document Types "type
registry" component, which was reconciled at retirement (see Definition
of Done).

### #Corpus in [[module-document-core]] (added) — CO

**Intent:** Walk the spec root, decide membership from frontmatter and
declared type, record exclusions with their reason, detect duplicate
slugs, and assemble the `SpecIndex` every query reads. Corpus
orchestrates the parsers but holds no parsing logic of its own.

**Behavior:**
- WHEN invoked against a directory, the component SHALL treat every `.md`
  file beneath it whose frontmatter parses and declares a known `type` as
  a corpus member (AC-SPEC-001.1), and SHALL consider only `.md` files,
  passing over other files without recording them as exclusions
  (AC-SPEC-001.6).
- IF a `.md` file has no frontmatter or declares an unknown `type`, THEN
  the component SHALL record it as excluded with reason `no-frontmatter`
  or `unknown-type` respectively, reporting the declared type value in
  the latter case (AC-SPEC-001.3, AC-SPEC-001.4).
- The component SHALL NOT read files outside the configured spec root
  (AC-SPEC-001.5).
- WHEN a single file path is checked, the component SHALL check that file
  and resolve it against the full corpus discovered from the spec root
  (AC-SPEC-001.2).
- The component SHALL build a `SpecIndex` carrying, per member, its
  frontmatter and repository-relative path, plus `schema_version` and the
  configured `workspace` (AC-SPEC-002.1, AC-SPEC-002.3).
- WHEN a member cannot be parsed, the component SHALL record the failure
  against it and index every other document rather than aborting
  (AC-SPEC-002.4).
- WHEN two members declare the same slug, compared without regard to
  letter case, the component SHALL report a `Finding` against each naming
  the other's path (AC-SPEC-003.3, AC-SPEC-003.4).
- The component SHALL emit every path in the index and in findings using
  forward slashes, relative to the repository root (AC-SPEC-009.8), and
  SHALL order entries, exclusions and findings deterministically so the
  same corpus content yields identical output on any platform
  (AC-SPEC-009.7).

**Anchors:** `cli/core/corpus` (the package the blueprints name for the
`Profile` interface); consumes #Frontmatter, the Schema Registry,
#Findings, and the generated `specmodel.SpecIndex`.

### #Numbering in [[module-document-core]] (added) — NUM

**Intent:** Derive the next identifier for a numbered type from the built
index rather than from a filesystem walk, so the query costs a lookup.

**Behavior:**
- WHEN asked for the next identifier for a type the active profile
  registers as numbered, the component SHALL return one greater than the
  highest number in use for that type anywhere in the corpus, regardless
  of directory (AC-SPEC-007.1).

**Anchors:** new package under `cli/core` (e.g. `cli/core/numbering`);
reads the `SpecIndex` produced by #Corpus.

### #Corpus in [[module-document-core]] (added — output and cache) — CACHE

**Intent:** Persist the built index to one known file that doubles as the
warm cache, so a CI job publishes the same artifact a later query reads
without re-walking the corpus. This is the container blueprint's "Index
cache" data made concrete.

**Behavior:**
- WHEN asked to build the index, the CLI SHALL write it to `specs.json`
  inside the configured spec root by default (AC-SPEC-010.1), to a
  location named by the output-path option when given (AC-SPEC-010.2), or
  to standard output — leaving the default file untouched — when the
  option is `-` (AC-SPEC-010.3).
- WHEN writing the index to a file, the CLI SHALL replace any existing
  file atomically, so a concurrent reader never observes a partial file
  (AC-SPEC-010.4).
- The CLI SHALL detect staleness by a **content hash** of the index's
  inputs, not by modification time, so a fresh CI checkout whose files
  carry new mtimes but identical content is still served warm — the
  condition FEAT-SPEC-009.1's CI budget depends on.
- WHEN a query or single-document check runs and the persisted index's
  input hash is unchanged since it was written, the CLI SHALL answer from
  the persisted index without re-walking the corpus (AC-SPEC-010.5,
  AC-SPEC-009.1, AC-SPEC-009.2).
- IF the persisted index is absent or any corpus input's hash has
  changed, THEN the CLI SHALL rebuild it before answering (AC-SPEC-010.6),
  completing a cold 1,000-document build within the FEAT-SPEC-009.3 budget.
- The CLI SHALL produce identical output whether or not a warm index is
  present, so deleting `specs.json` is always safe.

**Anchors:** `cli/core/corpus` (emission and the content-hash staleness
check). This resolves [[module-document-core]]'s standing Open Question
on cache invalidation in favour of content hashing; that resolution folds
into the module blueprint at retirement.

### Command Surface in [[container-spec-cli]] (added — configuration) — CFG

**Intent:** Resolve and check `specs.toml`, select one profile by name
from a **profile registry**, and report configuration defects through the
same finding machinery as corpus defects. The profile registry is the
customization seam: it maps `[profile].name` to a `Profile`, holds
`specs` as the built-in, and is where a second taxonomy is wired in.
Each profile in the registry carries its own single schema.
Configuration is checked, not merely loaded: a malformed config is the
first thing an adopter gets wrong.

**Behavior (realizes [[contract-specs-configuration]]):**
- The command surface SHALL hold a registry of available profiles keyed
  by name, and SHALL activate exactly the one `[profile].name` selects,
  passing it the resolved configuration.
- The command surface SHALL read configuration from a TOML file at the
  repository root and resolve `[workspace].root` relative to it
  (AC-SPEC-008.1), determining the spec root, the workspace identifier,
  and — via the activated profile — the union schema and the slug
  patterns its checks apply (AC-SPEC-008.2).
- WHEN `specs.toml` is absent, the command surface SHALL report the path
  it looked for and exit non-zero, inferring no spec root (AC-SPEC-008.3).
- IF the configuration is present but invalid, THEN the command surface
  SHALL report each offending key and its span, and SHALL NOT fall back
  to a default for that key (AC-SPEC-008.4, AC-SPEC-008.5), reporting
  every defect in one run and validating against
  `cli/core/schema/specs-config.schema.json`.
- IF `[profile].name` names no profile in the registry, THEN the command
  surface SHALL report a finding listing the registered names.
- The command surface SHALL resolve absent `[conventions]` keys from the
  active profile ([[contract-specs-configuration]] resolution).

**Anchors:** new package `cli/cmd`; consumes the generated
`specmodel.SpecsConfig`, the generated config schema, and the profile
registry into which the PROF delta's `specs` profile registers.

### Command Surface in [[container-spec-cli]] (added — command tree) — CLI

**Intent:** Present the engine as the four-subcommand surface FEAT-SPEC-011
fixes, converting between the module's finding and index models and the
process's stdout and exit status. It holds no domain logic: a check
that lived here would be unavailable to a future API or MCP wrapper.

**Behavior:**
- The command surface SHALL expose the subcommands `index`, `check`,
  `slug-exists` and `next-id`, and SHALL list them with their options on
  `--help` or a bare invocation (AC-SPEC-011.1–011.5).
- `check` SHALL accept an optional path and the `--json`, `--strict` and
  `--only` options: `--json` writes the full finding set as JSON to
  stdout (AC-SPEC-006.5); `--only` names a subset of checks by their rule
  identifiers and runs only those (AC-SPEC-006.8); the report distinguishes
  a passed check from an un-requested one (AC-SPEC-006.9).
- WHEN a check run produces at least one `error`, the command surface
  SHALL exit non-zero (AC-SPEC-006.6); WHEN it produces only `warning` or
  `unverified` findings, it SHALL exit zero unless `--strict` is set
  (AC-SPEC-006.7).
- The `index` subcommand SHALL exit zero after building an index that
  contains parse failures (AC-SPEC-002.5), and SHALL accept the
  output-path option of the CACHE delta (AC-SPEC-011.1).
- `slug-exists` SHALL report whether a corpus member declares the given
  slug and, if so, its path (AC-SPEC-007.2); `next-id` SHALL return the
  next identifier for a numbered type (AC-SPEC-007.1, via #Numbering).
- WHEN a command or flag is added, its `--help` text SHALL derive from
  the same declaration that defines it (container Change Guidance).

**Anchors:** `cli/cmd`; [[adr-015]] names `cobra` for the command tree;
renders #Findings and reads the `SpecIndex`; snapshot cases live under
`cli/testdata/cases/`.

### Static binary and distribution in [[container-spec-cli]] (added) — BUILD

**Intent:** Ship the CLI as a per-platform statically linked binary that
installs without a runtime or network fetch, satisfying the portability
and supply-chain requirements that an adopting security review turns on.

**Behavior:**
- The CLI SHALL run on a machine with no language runtime, interpreter or
  package manager installed beyond the binary itself (AC-SPEC-009.4).
- Installation SHALL retrieve no executable content from the network
  (AC-SPEC-009.5), and the CLI SHALL be installable from at least one
  package registry that supports on-premises mirroring (AC-SPEC-009.6).
- The build SHALL produce the platform matrix by cross-compilation from a
  single source, and identical corpus content SHALL yield byte-identical
  output across Linux, macOS and Windows (AC-SPEC-009.7, in concert with
  the CO determinism behavior).

**Anchors:** Go build/release configuration and CI under `.github/workflows`;
the repository currently has no CI (see Tracker Reconciliation).

### Relied-upon elements (unchanged)

- Generated `cli/core/specmodel` structs (`DocumentFrontmatter`, per-type
  structs, `SpecsConfig`, `ProfileRegistration`, `DocumentTypeRegistration`)
  — the typed models FM, SR, PROF, CO and CFG decode into and register.
- `cli/core/schema/spec-model.schema.json` and `specs-config.schema.json`
  — the generated bundled schemas: the union schema SR validates against
  (embedded by PROF via `//go:embed`), and the config schema CFG
  validates `specs.toml` against.
- `scripts/generate-schemas.sh` and the Strata toolchain — IDX extends
  its inputs; every consumer relies on its `--check` gate to keep
  generated artifacts fresh.
- [[contract-specs-configuration]] — realized by CFG and PROF; its shape
  is unchanged by this feature.
- [[adr-015]] — Go, `cobra`, `goccy/go-yaml`, `santhosh-tekuri/jsonschema`.

## Contracts

- [[contract-specs-configuration]] (relied upon, no shape change) — this
  feature provides its first implementation: CFG realizes the `specs.toml`
  surface and PROF realizes the internal registration surface, both at the
  locations the contract already names. No revision is required.
- **v1 deliberately does not promote the emitted index to a contract
  document.** It crosses a boundary (the Architecture Catalogue) and IDX
  pins its shape in Strata, which by this framework's filing rule argues
  for a contract. But no consumer or catalogue exists yet, and the same
  filing rule promotes a schema to a Contract only when a second,
  independently-built consumer appears. The shape is additive and the
  generated JSON Schema is the interim machine-readable spec, so promotion,
  with a formal `schema_version` policy, is deferred to the
  index-schema ADR, to land with the first consumer. This is a scope
  decision, not an open question, and it does not block grooming or
  implementation.

## Tracker Reconciliation

- Out-of-scope (parent): tracks authoring this
  blueprint and the subsequent groom and parallel sprint run; it is the
  umbrella this feature is delivered under, not a delta it subsumes.
- Out-of-scope (deferred follow-up): "ADR: index
  schema and versioning." Promoting the emitted index to a contract
  document with a formal versioning policy is deferred until a consumer or
  catalogue exists; the sprint-1 shape is additive, carries no references,
  and is pinned in Strata at `schema_version` 1.0, so this does not block
  the feature. Not a prerequisite.
- Resolved: "ADR: index cache invalidation." The
  decision is **content hash**, pinned in the CACHE delta; the issue's ADR
  can be closed by recording that choice, or written up to document the
  mtime alternative that was rejected.
- Resolved: "Settle 'index' vs 'manifest' as the
  name of the corpus artifact." The name is **index**; the model is
  `SpecIndex` and the file is `specs.json`. The issue can be closed.
- Out-of-scope (relied upon): tracks [[adr-016]]
  (Strata as single source), already `ready-for-review`; IDX extends its
  generation pipeline rather than reopening the decision.
- Out-of-scope: "model references as a separate
  edge collection." A reference-feature concern; IDX's additive shape must
  not preclude it, but this feature indexes no references.
- Out-of-scope (coordination): `work-order-template.md`
  is not valid YAML. Its file
  (`plugin/skills/spec/references/work-order-template.md`) sits outside
  `specs/`, the corpus root this feature's self-application gate scans,
  so the gate cannot surface it as scoped. That scoping was a stale
  assumption from before the `specs/` reorg, corrected here. The CO
  parse-failure path would still record the defect if pointed at a root
  that included this file; this feature does not fix the document or
  widen the gate's root, and that issue's own acceptance criteria already
  call for a separate, generic regression check to cover it.
- Out-of-scope: dangling wikilinks and unpopulated `contracts_consumed`
  are reference-feature and corpus-data concerns this feature neither
  resolves nor regresses.

**Reconciled at retirement (2026-08-05).**

- The index-vs-manifest naming issue — **closed.** The artifact is named
  *index*: `SpecIndex` in `schemas/index.strata`, emitted as
  `specs.json`.
- The cache-invalidation issue — **closed**, taking the blueprint's
  stated default of recording the choice rather than authoring an ADR.
  Staleness is a content hash of the corpus inputs (`input_hash`, a
  `sha256:` digest carried in the emitted index); mtime was rejected
  because a fresh CI checkout carries new mtimes over identical content,
  which would cold-build every CI run and break the FEAT-SPEC-009.1
  budget the decision exists to protect. No second alternative was in
  contention, so the ADR would have recorded one rejected option and no
  open trade-off.
- The parent, the index-schema ADR, the [[adr-016]] issue, the
  edge-collection issue, the `work-order-template.md` issue and the
  dangling-wikilinks issue — remain open and out-of-scope; each was
  confirmed neither regressed nor silently absorbed. The [[adr-016]]
  issue is `in_progress` (ADR-016 is `ready-for-review`), which is its
  own lifecycle, not this feature's.
- The `contracts_consumed` issue — **exists in the tracker** ("Populate
  `contracts_consumed` on the committed blueprints", open). The claim in
  the grooming that it does not exist was wrong; its out-of-scope
  classification is unchanged and correct.
- "ADR: release and distribution of CLI binaries" —
  open, and deliberately so. The BUILD delta
  pre-empted three of its five questions by shipping the answer: the
  platform/arch matrix, cross-compilation from one source, and GitHub
  Releases as the channel. What that ADR still owns is the part BUILD did
  not decide: how an agent session obtains the binary given that this
  repository is distributed by clone, and whether checksums or signing
  apply. It is a distribution-policy decision, not a gap in this feature,
  and it blocks nothing here. Its sibling (`spec --version` and version
  stamping) is open for the same reason.
- Filed during implementation and left open: the snapshot harness
  treats a missing expected artifact as "case carries none"; `spec
  --version` and version stamping; no `version` on
  `ContainerBlueprint`; AC-SPEC-006.7's declared level unreachable. None
  blocks retirement.

## Sequencing Constraints

- `delivery: staged`. Three stages, each a reviewable increment:
  - **Stage 1 — engine (library):** IDX, SP, FIND, FM, SR, PROF, CO, NUM.
    Exercisable through Go tests and the profile self-application gate;
    no binary yet.
  - **Stage 2 — CLI (end-to-end):** CACHE, CFG, CLI. The binary indexes,
    checks and queries a real corpus.
  - **Stage 3 — distribution:** BUILD. Cross-platform binaries and CI.
- The IDX delta SHALL land before CO and CACHE, which populate and emit
  the `SpecIndex` it defines.
- The SP delta SHALL land before FM (FM builds on the delimiter scanner).
- The FIND delta SHALL land before SR and CO (both report through it).
- The FM delta SHALL land before SR and CO (both consume parsed
  frontmatter).
- The CO delta SHALL land before NUM and CACHE (both read the built
  index).
- The `Profile` extension-point interface (in package `cli/core/corpus`)
  — the surface through which a profile supplies its union schema, slug
  patterns and numbered types — is a foundational sliver: it SHALL be
  defined before the PROF delta implements it and before the CO and CFG
  deltas consume it, i.e. ahead of the CO delta's own discovery logic even
  though both live in that package. Grooming SHOULD lift it into its own
  early work item.
- The PROF delta SHALL land before CFG (which activates the profile by
  name) and is required for SR and CO to check a conforming corpus green.
- The CFG and CACHE deltas SHALL land before CLI, which dispatches to
  both.
- No backward-compatibility windows apply: [[module-document-core]] and
  [[module-spec-profile]] are `visibility: internal` and evolve in
  lockstep with their only consumer.

## AC Coverage

Legend: IDX = `SpecIndex` index model; SP = #Spans; FIND = #Findings;
FM = #Frontmatter; SR = Schema Registry; PROF = specs profile
(union schema + type table); CO = #Corpus; NUM = #Numbering;
CACHE = #Corpus (index output and warm cache); CFG = Command Surface
(configuration); CLI = Command Surface (command tree); BUILD = static
binary and distribution.

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-SPEC-001.1 | **CO (proves)** | unit |
| AC-SPEC-001.2 | **CO (proves)**, CLI | integration |
| AC-SPEC-001.3 | **CO (proves)** | unit |
| AC-SPEC-001.4 | **CO (proves)** | unit |
| AC-SPEC-001.5 | **CO (proves)** | unit |
| AC-SPEC-001.6 | **CO (proves)** | unit |
| AC-SPEC-002.1 | **CO (proves)**, IDX | unit |
| AC-SPEC-002.2 | **IDX (proves)**, CO | unit |
| AC-SPEC-002.3 | **CO (proves)**, IDX | unit |
| AC-SPEC-002.4 | **CO (proves)**, FM | unit |
| AC-SPEC-002.5 | **CLI (proves)**, CO | integration |
| AC-SPEC-003.1 | **SR (proves)**, PROF | unit |
| AC-SPEC-003.2 | **SR (proves)**, FM | unit |
| AC-SPEC-003.3 | **CO (proves)** | unit |
| AC-SPEC-003.4 | **CO (proves)** | unit |
| AC-SPEC-003.5 | **SR (proves)**, PROF | unit |
| AC-SPEC-006.1 | **FIND (proves)** | unit |
| AC-SPEC-006.3 | **FIND (proves)** | unit |
| AC-SPEC-006.4 | **FIND (proves)** | unit |
| AC-SPEC-006.5 | **CLI (proves)**, FIND | integration |
| AC-SPEC-006.6 | **CLI (proves)** | integration |
| AC-SPEC-006.7 | **CLI (proves)** | integration |
| AC-SPEC-006.8 | **CLI (proves)**, FIND | integration |
| AC-SPEC-006.9 | **FIND (proves)**, CLI | unit |
| AC-SPEC-007.1 | **NUM (proves)**, PROF | unit |
| AC-SPEC-007.2 | **CO (proves)**, CLI | integration |
| AC-SPEC-008.1 | **CFG (proves)** | unit |
| AC-SPEC-008.2 | **CFG (proves)**, PROF | unit |
| AC-SPEC-008.3 | **CFG (proves)** | integration |
| AC-SPEC-008.4 | **CFG (proves)** | unit |
| AC-SPEC-008.5 | **CFG (proves)** | unit |
| AC-SPEC-008.6 | **PROF (proves)**, CFG | integration |
| AC-SPEC-009.1 | **CACHE (proves)** | integration |
| AC-SPEC-009.2 | **CACHE (proves)** | integration |
| AC-SPEC-009.3 | **CO (proves)**, CACHE | integration |
| AC-SPEC-009.4 | **BUILD (proves)** | manual |
| AC-SPEC-009.5 | **BUILD (proves)** | manual |
| AC-SPEC-009.6 | **BUILD (proves)** | manual |
| AC-SPEC-009.7 | **BUILD (proves)**, CO | e2e |
| AC-SPEC-009.8 | **CO (proves)** | unit |
| AC-SPEC-010.1 | **CACHE (proves)** | unit |
| AC-SPEC-010.2 | **CACHE (proves)**, CLI | unit |
| AC-SPEC-010.3 | **CACHE (proves)**, CLI | unit |
| AC-SPEC-010.4 | **CACHE (proves)** | unit |
| AC-SPEC-010.5 | **CACHE (proves)** | integration |
| AC-SPEC-010.6 | **CACHE (proves)** | integration |
| AC-SPEC-010.7 | **CACHE (proves)**, CLI | integration |
| AC-SPEC-011.1 | **CLI (proves)**, CACHE | integration |
| AC-SPEC-011.2 | **CLI (proves)** | integration |
| AC-SPEC-011.3 | **CLI (proves)**, CO | integration |
| AC-SPEC-011.4 | **CLI (proves)**, NUM | integration |
| AC-SPEC-011.5 | **CLI (proves)** | integration |

## Key Decisions

- [[adr-016]] — Strata is the single source for generated models; the IDX
  delta adds the index model and the FIND delta adds the finding model
  under this decision.
- [[adr-015]] — Go, `cobra`, `goccy/go-yaml`, `santhosh-tekuri/jsonschema`
  are the implementation choices these deltas build on.
- **v1 validates frontmatter against one embedded union schema per
  profile, not a per-type schema registry.** Profiles stay pluggable and
  are selected from a registry by name; each profile carries a single
  union schema plus a slug/numbered table. This narrows
  [[module-spec-profile]]'s design (its Document Types "type registry"
  component is removed) and is recorded here for folding into that
  blueprint at retirement; a standalone ADR is warranted only if a second
  profile reopens the question (its `kind: pattern` Open Question).
- Cache invalidation is decided: **content hash**, pinned in the CACHE
  delta. The index-schema contract and its versioning policy are
  deferred until a consumer exists and do not block this feature; if the
  index is promoted to a contract later, that decision is authored as an
  ADR then.

## Definition of Done

Verified 2026-08-05.

- [x] Every delta above is folded into its owning blueprint:
  - [x] [[module-document-core]] — Spans, Frontmatter, Schema Registry
        (single-union validation), Corpus, Numbering, Findings confirmed
        as built; the `SpecIndex` index model and the `Finding` model
        recorded; the `Profile` extension point reconciled to supply one
        union schema (not a per-type schema); `version` bumped 0.1 → 0.2.
  - [x] [[module-spec-profile]] — the single-schema profile confirmed as
        built; the separate Document Types "type registry" component
        removed in favour of one union schema plus a slug/numbered table;
        `version` bumped 0.1 → 0.2; deferred components (Element Records,
        Lifecycle Rules) noted as not-yet-implemented.
  - [x] [[container-spec-cli]] — Command Surface and the profile registry
        built; the "Index cache" Data section reconciled with the CACHE
        delta; the emitted-index Open Question updated with the
        index-schema decision. **`version` was not bumped**: the
        `ContainerBlueprint` schema declares no `version` property and
        sets `additionalProperties: false`, so this DoD clause is
        unsatisfiable for a container as written (an open issue). The
        other two bumps stand.
- [x] The IDX index model and the FIND finding model exist in `schemas/`
      (`index.strata`, `finding.strata`), generated artifacts are
      regenerated, and `scripts/generate-schemas.sh --check` passes.
- [x] Every AC in the coverage table is verified at its declared level
      and cited from its tests per [[testing-standards]], **with two
      recorded exceptions**:
  - AC-SPEC-009.4, AC-SPEC-009.5 and AC-SPEC-009.6 carry the `(manual)`
    marker in [[frd-spec-index]] and are exempt from citation.
  - AC-SPEC-006.7 is declared `integration` but is cited only from the
    unit test of `exitCodeFor`. No snapshot case can reach it: every
    finding constructor this sprint registers emits `error` or
    `unverified`, and none emits `warning`, so no corpus produces the
    warning-only run the criterion describes (an open issue).
  - The remaining 48 of 51 are cited at or above their declared level;
    no citation is orphaned.
- [x] The profile self-application gate runs the built profile against
      this repository's `specs/` corpus: 26 entries, 1 exclusion, no
      findings.
- [x] Issues classified `prerequisite` in Tracker Reconciliation are
      resolved or their decisions recorded; `out-of-scope` issues are
      confirmed neither regressed nor silently absorbed.
- [x] This blueprint's status → retired.

## Open Questions

None. Three questions raised during authoring are resolved or scoped out:
the corpus artifact is named **index** (`SpecIndex`, `specs.json`);
cache staleness is detected by **content hash**; and promoting the
emitted index to a **contract document** with a formal versioning policy
is **deferred** until a consumer exists. That deferral is a scope
decision recorded in Contracts and Tracker Reconciliation, not a
plannability gap.

## References

- `c4-guide.md` — the C4 mapping rules and `#`-reference conventions the
  delta blocks are written against
- [[frd-spec-index]] — the requirements this blueprint realizes
- [[container-spec-cli]], [[module-document-core]], [[module-spec-profile]]
  — the owning blueprints the deltas fold into
- [[contract-specs-configuration]] — the configuration surface CFG and
  PROF realize
- `schemas/README.md` — the Strata source-of-truth pipeline the IDX delta
  extends
- [[testing-standards]] — the test-citation obligation the AC
  Coverage table feeds
