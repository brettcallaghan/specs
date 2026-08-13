---
title: "FRD: Spec Corpus Index and Core Conformance Checks"
description: A command-line tool that indexes a specification corpus into
  machine-readable JSON and reports deterministic frontmatter and metadata
  conformance findings, so CI, agents, and architecture tooling read one
  authoritative artifact
slug: frd-spec-validation-cli
type: frd
status: approved
updated_at: 2026-08-04T00:00:00Z
sources: []
---

# Spec Corpus Index and Core Conformance Checks

## Overview

Specification documents in this framework carry machine-checkable
structure — frontmatter properties, typed slugs, `[[slug]]` references,
fenced record blocks, required sections — and nothing checks it. The
rules are enforced by agents reading documents and reporting what they
found. That is expensive, and it is unverifiable: an agent that claims a
readiness checklist passed produces the same output as one that ran it.
A prototype over this repository's own 50 documents found 26 broken
references and one document whose frontmatter does not parse, none of
which any reviewer had noticed.

This feature delivers a command-line tool that reads a specification
corpus and produces two things: a **machine-readable index** of every
document's metadata and location, and a **report of conformance
findings** against the metadata and membership rules the framework
already defines. It is the first increment of a larger tool: later
features add document **structure** (headings, section spans, record
blocks, hygiene) and cross-document **references** to the same index and
the same finding engine. This sprint's design is chosen to accommodate
that growth — the index schema is versioned and additive, the finding
record is uniform across every check, and the corpus walk established
here is the one those later checks reuse.
It serves three audiences from one engine. A team adopting the framework
points it at their spec root and gets a failing build when documents
drift. An agent runs one command instead of reading forty files, and
cites the output instead of asserting a result. An organization pushes
the index to a central service and sees its whole architecture — the
payoff of modeling systems, containers, and components rigorously in the
first place, which is felt at enterprise scale rather than in a single
application.

The index is the primary artifact, not a by-product. Every check in this
document is a query over it, and consumers outside this framework
consume the index without running any check at all. That ordering
governs the requirements below: indexing never fails, and validation is
a separate consumer that may.

**Out of scope for this feature.** Each of these is deferred
deliberately, not overlooked:

- **Reference integrity.** Parsing `[[slug]]` references, resolving them
  against the corpus, emitting inbound and outbound references in the
  index, and reporting broken links are a follow-on feature. This FRD
  indexes a document's frontmatter and location but does not read its
  body for references.
- **Document structure and hygiene.** Extracting headings and their
  section spans, parsing record blocks, and checking required sections,
  line width, leftover scaffolding, and placeholders are a follow-on
  feature. This FRD does not walk a document's headings or body
  structure.
- **Modifying documents.** This feature only reads and reports.
  Deriving `updated_at` from git history and rewriting frontmatter is a
  follow-on feature.
- **Acceptance-criteria and feature-definition validation.** Out of
  scope for this feature. [[adr-020]] settles the notation and assigns
  the extraction to a follow-on feature; the technology decision this
  feature defers to an ADR had to preserve the ability to parse and
  validate structured feature definitions later, and does.
- **Requirement quality judgment.** Whether a requirement is testable,
  whether a qualifier is vague, whether a design carries enough intent —
  these stay with a reviewing agent. This feature extracts the spans an
  agent must read; it does not judge them.
- **Rendering or diagram generation**, and export to any third-party
  architecture format.
- **Choosing an implementation language, runtime, or distribution
  channel.** REQ-SPEC-009 states the latency, portability, and
  installation requirements; the choice that satisfies them is recorded
  in a separate ADR.

## Terminology

- **Spec corpus** — the set of documents the CLI treats as
  specifications. Membership is determined by content, not location: a
  `.md` file with parseable frontmatter declaring a known `type`.
- **Spec root** — the configured directory the CLI walks to discover the
  corpus.
- **Workspace** — a named collection of documents from one repository,
  identified in configuration so that indexes produced by separate
  repositories can be combined without slug collisions.
- **Index** — the machine-readable JSON artifact describing every
  discovered document: its frontmatter properties, its file path, and the
  reason any candidate file was excluded. Later features add headings and
  their spans, extracted record blocks, and inbound and outbound
  references to the same artifact under the same schema.
- **Profile** — the configuration bundle that supplies document-type
  vocabulary: which types exist, what each requires, the slug pattern
  for each, and which required sections each declares.
- **Finding** — one reported conformance result, carrying a severity, a
  rule identifier, a location, and a message.
- **Severity** — one of `error` (a conformance failure), `warning` (a
  probable problem that does not fail a build), or `unverified` (a check
  the CLI was asked to perform and could not run).
- **Span** — a contiguous region of a document, carrying a byte offset
  and byte length as its authoritative extent, plus the line and column
  of its start and end for readability. Byte offsets are required
  because spans address regions *within* a line as well as whole
  sections — a malformed `slug` value occupies part of a line, not the
  line itself.
- **Warm index** — a persisted index whose inputs are unchanged since it
  was built, so a query or single-document check can be answered without
  re-reading the corpus (REQ-SPEC-010).

## Requirements

### REQ-SPEC-001: Corpus Discovery and Membership

**User Story:** As an adopting team, I want to point the tool at my
repository and have it work out which files are specifications, so that
I do not have to enumerate documents or restructure my directories.

**Acceptance Criteria:**

- **AC-SPEC-001.1:** WHEN the CLI is invoked against a directory, the
  CLI SHALL treat every `.md` file beneath it whose frontmatter parses
  and declares a `type` known to the active profile as a member of the
  corpus.
  - *Example:* Invoked against `docs/`, it admits
    `docs/adrs/ADR-011-targeted-verification-gates.md`, which declares
    `type: adr`.
- **AC-SPEC-001.2:** WHEN the CLI is invoked against a single file path,
  the CLI SHALL check that file and resolve its references against the
  full corpus discovered from the configured spec root.
  - *Example:* Checking one ADR still reports a `[[slug]]` in it that
    matches no document anywhere in `docs/`.
- **AC-SPEC-001.3:** IF a `.md` file beneath the spec root has no
  frontmatter, THEN the CLI SHALL record it in the index as excluded
  with the reason `no-frontmatter` rather than omitting it silently.
  - *Example:* `docs/plans/plan.md` appears in the index's exclusions
    list with reason `no-frontmatter`.
- **AC-SPEC-001.4:** IF a `.md` file declares a `type` the active
  profile does not define, THEN the CLI SHALL record it as excluded with
  the reason `unknown-type` and report the declared value.
  - *Example:* A file declaring `type: rfc` under a profile with no
    `rfc` type is listed with `"type": "rfc"`, so a reviewing agent can
    judge whether the profile or the document is wrong.
- **AC-SPEC-001.5:** The CLI SHALL NOT read files outside the configured
  spec root when resolving corpus membership.
- **AC-SPEC-001.6:** The CLI SHALL consider only files whose name ends
  in `.md` as candidate corpus members, and SHALL pass over any other
  file silently rather than recording it as an exclusion.
  - *Example:* `docs/specs/specs.json` and `docs/adrs/README.rst` are
    ignored without appearing in the exclusions list, while
    `docs/adrs/ADR-004-module-blueprint-and-atomic-component.md` is
    considered.

*Rationale:* Membership by content rather than path is what lets the
tool run against a repository it did not design. Recording exclusions
rather than dropping them is what lets an agent notice that a document
which *should* be a specification is missing frontmatter — a judgment
the CLI cannot make.

### REQ-SPEC-002: Document Index

**User Story:** As an architecture owner, I want one machine-readable
artifact describing every specification document and how they relate, so
that CI, agents, and a central catalogue all read the same source of
truth instead of each re-deriving it.

**Acceptance Criteria:**

- **AC-SPEC-002.1:** WHEN the CLI builds an index, the CLI SHALL emit a
  JSON document containing, for every corpus member, its frontmatter
  properties and its file path.
  - *Example:* The entry for `adr-011` carries its six frontmatter
    properties and its path
    `docs/adrs/ADR-011-targeted-verification-gates.md`. Its headings and
    outbound references are added to the same entry by later features.
- **AC-SPEC-002.2:** The index SHALL declare a schema version
  identifier.
  - *Example:* The emitted JSON opens with
    `"schema_version": "1.0"`. Later features add properties under this
    same version line without a breaking bump.
- **AC-SPEC-002.3:** The index SHALL declare the workspace identifier
  from configuration for every document it contains.
  - *Example:* Every entry carries `"workspace": "agent-plugins"`, so
    merging this index with another repository's cannot collide two
    documents that both use the slug `adr-001`.
- **AC-SPEC-002.4:** WHEN a document in the corpus cannot be parsed, the
  CLI SHALL index every other document and record the failure against
  the offending document rather than aborting the index.
  - *Example:* `work-order-template.md` fails YAML parsing at line 28;
    the index still contains the other 49 documents and lists that file
    with its parse error and position.
- **AC-SPEC-002.5:** The CLI SHALL exit zero after building an index
  that contains parse failures.

*Rationale:* AC-SPEC-002.4 and AC-SPEC-002.5 together are what make the
index usable as an architecture artifact. A repository with three broken
documents must still contribute its other nine hundred to an
organization's graph; a build failure is the checker's job, not the
indexer's. The index is deliberately additive: this sprint fills
frontmatter and path, and later features populate the remaining
properties without changing the schema's shape or its version.

### REQ-SPEC-003: Metadata Conformance

**User Story:** As a reviewer, I want frontmatter errors reported
mechanically, so that review attention goes to whether a document says
the right thing rather than whether it is shaped correctly.

**Acceptance Criteria:**

- **AC-SPEC-003.1:** WHEN the CLI checks a document, the CLI SHALL
  validate its frontmatter against the JSON Schema the active profile
  registers for that document's `type`.
  - *Example:* A document declaring `type: adr` but omitting
    `description` produces one finding naming the missing property.
- **AC-SPEC-003.2:** WHEN the CLI validates a frontmatter property whose
  schema constrains its textual format, the CLI SHALL validate the raw
  text of the property as written in the file.
  - *Example:* `updated_at: 2026-07-29` fails an RFC 3339 date-time
    check, even though a YAML parser would coerce it to a valid
    timestamp before any check could see it.
- **AC-SPEC-003.3:** WHEN two corpus members declare the same slug, the
  CLI SHALL report a finding against each, naming the other's path.
  - *Example:* Two documents both declaring `slug: adr-012` each produce
    a finding citing the other's file path.
- **AC-SPEC-003.4:** The CLI SHALL compare slugs for uniqueness without
  regard to letter case.
  - *Example:* `frd-checkout` and `FRD-Checkout` are reported as
    duplicates, because a case-insensitive filesystem cannot hold both
    as distinct files.
- **AC-SPEC-003.5:** WHEN a document's slug does not match the pattern
  the active profile registers for its `type`, the CLI SHALL report a
  finding naming both the actual slug and the expected pattern.
  - *Example:* `slug: adr-011-targeted-verification-gates` under a
    profile registering `^adr-[0-9]{3}$` for type `adr` produces a
    finding quoting both.
- **AC-SPEC-003.6:** WHEN a corpus member is indexed with a parse error
  rather than parsed frontmatter, the CLI SHALL report a finding
  carrying that error's message and span, and SHALL NOT report the
  document as having passed any check.
  - *Example:* A document copied from a template and not yet completed
    carries `updated_at: <YYYY-MM-DDTHH:MM:SSZ>`, which is not an RFC
    3339 date-time. The check run reports it and exits non-zero, rather
    than reporting four passed checks against a document nothing was
    able to read.

*Rationale for AC-SPEC-003.6:* every other criterion in this requirement
presupposes parsed frontmatter, so before this was written the worst-formed
document in a corpus was the one the checker was guaranteed to be silent
about — the severity of a defect and the tool's willingness to mention it
ran in opposite directions. REQ-SPEC-002's rationale already assigned this
work: an index containing parse failures exits zero because "a build
failure is the checker's job, not the indexer's". This is the criterion
that makes the checker do that job. A file with no frontmatter at all
remains an exclusion under AC-SPEC-001.3 and is *not* covered here; it
never claimed to be a specification, whereas a member with a parse error
did and got the claim wrong.

*Note on numbering:* REQ-SPEC-004 (reference integrity) and REQ-SPEC-005
(structure and hygiene conformance) covered behavior deferred to
follow-on features. Their numbers are intentionally left unused here
rather than reassigned: the retained requirements REQ-SPEC-006, -008,
and -009 are cited by ADRs and keep their numbers, so renumbering around
the gap would break those citations. The gap is deliberate.

### REQ-SPEC-006: Findings, Severities, and Diagnostics

**User Story:** As an agent acting on the output, I want findings that
name the rule and cite where it is defined, so that I can fix the
problem without loading the standards documents that define it.

**Acceptance Criteria:**

- **AC-SPEC-006.1:** Every finding the CLI reports SHALL carry a stable
  rule identifier, a severity, a file path, a span, and a message
  stating the violated constraint and the value that violated it.
  - *Example:* `slug/pattern-mismatch  error
    docs/adrs/ADR-011-targeted-verification-gates.md:4:7`
    "'adr-011-targeted-verification-gates' does not match pattern
    '^adr-[0-9]{3}$'" — self-contained, so the fixing agent needs no
    further lookup.
- **AC-SPEC-006.3:** WHEN the CLI is asked to perform a check it cannot
  run, the CLI SHALL report that check with severity `unverified`.
  - *Example:* Asked to validate mermaid diagrams with no diagram
    validator available, it emits one `unverified` finding per diagram
    naming the missing dependency.
- **AC-SPEC-006.4:** The CLI SHALL NOT report a check it could not run
  as passed.
- **AC-SPEC-006.5:** WHEN the `check` subcommand is invoked with the
  `--json` option, the CLI SHALL write the full finding set as JSON to
  standard output.
- **AC-SPEC-006.6:** WHEN a check run produces at least one finding of
  severity `error`, the CLI SHALL exit with a non-zero status.
- **AC-SPEC-006.7:** WHEN a check run produces findings only of severity
  `warning` or `unverified`, the CLI SHALL exit zero unless invoked with
  the `--strict` option.
- **AC-SPEC-006.8:** WHEN the `check` subcommand is invoked with the
  `--only` option naming a subset of checks, the CLI SHALL run only the
  named checks.
  - *Example:* An agent running `check --only slug-uniqueness` gets a
    report covering that check alone.
- **AC-SPEC-006.9:** The CLI SHALL NOT report a check that was not
  requested as passed.
  - *Example:* A `--only slug-uniqueness` report distinguishes "slug
    uniqueness passed" from "frontmatter schema was not checked".

*Rationale:* AC-SPEC-006.1's message requirement is what allows the
`spec` skill to stop preloading verification procedure. A diagnostic
that quotes the violated constraint and the offending value is
self-contained: the agent fixes it without opening the standard that
defines the rule.

### REQ-SPEC-007: Corpus Queries

**User Story:** As an authoring agent, I want to ask the corpus direct
questions, so that I can pick a conforming identifier and understand
impact without reading every document.

**Acceptance Criteria:**

- **AC-SPEC-007.1:** WHEN the `next-id` subcommand is invoked for a
  document type the active profile registers as numbered, the CLI SHALL
  return one greater than the highest number in use for that type.
  - *Example:* With ADRs numbered 001 through 013 in the corpus,
    `next-id adr` returns `014`.
- **AC-SPEC-007.2:** WHEN the `slug-exists` subcommand is invoked with a
  slug, the CLI SHALL report whether a corpus member declares it and, IF
  one does, the path of that member.
  - *Example:* `slug-exists adr-004` reports the slug in use and names
    `docs/adrs/ADR-004-module-blueprint-and-atomic-component.md`.

*Rationale:* Both queries answer from frontmatter and path alone, so
they belong in this sprint. Queries that need more of the index — "which
documents reference this slug", "return the spans of headings matching a
selection" — are deferred to the later features that populate those
parts of the index.

### REQ-SPEC-008: Configuration

**User Story:** As an adopting team, I want to declare my own document
types and conventions, so that the tool enforces my framework rather
than the one it shipped with.

**Acceptance Criteria:**

- **AC-SPEC-008.1:** The CLI SHALL read its configuration from a TOML
  file at the repository root.
  - *Example:* `specs.toml` declares `root = "docs/specs"` and the
    workspace identifier.
- **AC-SPEC-008.2:** The configuration SHALL determine the spec root,
  the workspace identifier, and the registered document types with their
  slug patterns and the JSON Schema that validates each type's
  frontmatter.
  - *Example:* A `[[type]]` table for `adr` declares
    `slug_pattern = "^adr-[0-9]{3}$"` and points at the schema
    REQ-SPEC-003 validates against. Later features add each type's
    required headings, its line-width limit, and its reference-bearing
    properties to this same file; the config schema is additive.
- **AC-SPEC-008.3:** WHEN no configuration file is present, the CLI
  SHALL report the absence and exit non-zero rather than assuming a
  layout.
  - *Example:* Run in a repository with no `specs.toml`, it names the
    file it looked for and the path it looked in.
- **AC-SPEC-008.4:** IF the configuration is present but invalid, THEN
  the CLI SHALL report the offending key and its position.
- **AC-SPEC-008.5:** IF the configuration is present but invalid, THEN
  the CLI SHALL NOT fall back to a default value for the offending key.
- **AC-SPEC-008.6:** The CLI SHALL ship a configuration for this
  framework's own document types that an adopting repository can use
  without authoring one.

### REQ-SPEC-009: Invocation, Performance, and Portability

**User Story:** As a platform engineer, I want to run this in locked-down
CI and in an editing loop that fires on every file write, so that neither
the build nor the author waits on it.

**Acceptance Criteria:**

- **AC-SPEC-009.1:** WHEN the CLI answers a corpus query against a
  1,000-document corpus with a warm index, the CLI SHALL return within
  50 milliseconds at the 95th percentile.
  - *Example:* An agent asking `slug-exists frd-checkout` before writing
    a document waits imperceptibly.
- **AC-SPEC-009.2:** WHEN the CLI checks a single document against a
  warm index for a 1,000-document corpus, the CLI SHALL complete within
  100 milliseconds at the 95th percentile.
  - *Example:* A validation hook firing on every document write does not
    make editing feel slow.
- **AC-SPEC-009.3:** WHEN the CLI builds an index for a 1,000-document
  corpus with no cache, the CLI SHALL complete within 1 second at the
  95th percentile.
- **AC-SPEC-009.4:** The CLI SHALL run on a machine with no language
  runtime, interpreter, or package manager installed beyond the CLI
  itself. (manual)
  - *Example:* It runs in a container image containing only a shell and
    the CLI binary.
- **AC-SPEC-009.5:** The CLI SHALL NOT retrieve executable content from
  a network location during installation. (manual)
  - *Example:* Installing from a package registry mirrored inside a
    network that blocks all other outbound traffic succeeds.
- **AC-SPEC-009.6:** The CLI SHALL be installable from at least one
  package registry that supports on-premises mirroring. (manual)
- **AC-SPEC-009.7:** WHEN the CLI runs against identical corpus content
  on Linux, macOS, and Windows, the CLI SHALL report the same findings
  in the same order.
  - *Example:* A corpus checked out with CRLF line endings on Windows
    yields byte-identical JSON output to the same corpus on Linux.
- **AC-SPEC-009.8:** The CLI SHALL emit file paths in the index and in
  findings using forward slashes as separators, relative to the
  repository root.

*Rationale:* AC-SPEC-009.1 through AC-SPEC-009.3 come from the two
invocation patterns that dominate: an agent asking one question, and a
write-time hook checking one file. AC-SPEC-009.5 exists because the
common packaging pattern of downloading a binary during install fails in
exactly the networks that most need this tool, and adds a supply-chain
surface an adopting security review will reject.

### REQ-SPEC-010: Index Output and Persistence

**User Story:** As a CI pipeline and as an authoring agent, I want the
index written to one known file that I can publish and that later
invocations reuse, so that the artifact pushed to a central catalogue is
the same one that answers my next query without re-reading the corpus.

**Acceptance Criteria:**

- **AC-SPEC-010.1:** WHEN the CLI is asked to build the index, the CLI
  SHALL write the JSON to a single file named `specs.json` inside the
  configured spec root.
  - *Example:* With `root = "docs/specs"` in the configuration, running
    `index` with no output option writes `docs/specs/specs.json` and the
    CLI prints that relative path.
- **AC-SPEC-010.2:** WHEN invoked with an output-path option naming a
  location, the CLI SHALL write the index to that location instead of
  the default.
  - *Example:* `--out build/spec-index.json` writes the artifact under
    `build/` for a CI job that publishes everything in that directory.
- **AC-SPEC-010.3:** WHEN invoked with the output-path option set to
  `-`, the CLI SHALL write the index to standard output and SHALL NOT
  create or modify the default file.
  - *Example:* `--out -` lets an agent pipe the index straight into
    `jq` without leaving a file behind.
- **AC-SPEC-010.4:** WHEN the CLI writes the index to a file, the CLI
  SHALL replace any existing file atomically, so that a concurrent
  reader observes either the old complete index or the new one and never
  a partially written file.
  - *Example:* A `PostToolUse` hook rebuilding the index while a query
    reads it never yields the query a truncated document.
- **AC-SPEC-010.5:** WHEN a corpus query or single-document check runs
  and the index file exists with inputs unchanged since it was written,
  the CLI SHALL answer from that persisted index — the warm index of
  REQ-SPEC-009 — without re-walking the corpus.
  - *Example:* `slug-exists frd-checkout` run twice in a row reads
    `specs.json` the second time and returns within the REQ-SPEC-009.1
    budget.
- **AC-SPEC-010.6:** IF the index file is absent or any corpus input has
  changed since it was written, THEN the CLI SHALL rebuild the index
  before answering rather than serving a stale result.
  - *Example:* Editing one ADR and immediately querying returns an
    answer reflecting the edit, because the changed input invalidated
    the warm index.

*Rationale:* The index artifact and the check report are different
outputs: REQ-SPEC-006.5 writes the *findings* of a check run to standard
output, while this requirement governs where the *index* — the primary
artifact of REQ-SPEC-002 — is persisted. One file serves both of the
index's roles: the deliverable a CI job publishes to a catalogue, and
the warm cache that satisfies the latency budgets of REQ-SPEC-009.
Placing `specs.json` inside the spec root keeps the artifact beside the
documents it describes, and because discovery reads only `.md` files
(AC-SPEC-001.6) the index is never mistaken for a corpus member. A fixed
default name means every skill body, CI config, and hook can refer to
`specs.json` without configuration, while the output-path option keeps
the CI-publish and pipe-to-stdout cases first-class. Whether an adopting
repository commits the file or ignores it is the adopter's choice; the
CLI requires neither.

### REQ-SPEC-011: Command Surface

**User Story:** As a human or agent developer integrating the tool, I
want a small, named set of subcommands, so that skills, hooks, and CI
scripts call a stable interface rather than reverse-engineering it from
behavior.

**Acceptance Criteria:**

- **AC-SPEC-011.1:** The CLI SHALL expose an `index` subcommand that
  builds the index and writes it per REQ-SPEC-010, accepting the
  output-path option named there.
  - *Example:* `index` writes the default `specs.json`; `index --out -`
    streams it to standard output.
- **AC-SPEC-011.2:** The CLI SHALL expose a `check` subcommand that runs
  conformance checks over the corpus, or over a single document when
  given a path, and reports findings per REQ-SPEC-006, accepting the
  `--json`, `--strict`, and `--only` options.
  - *Example:* `check` validates the whole corpus;
    `check docs/adrs/ADR-004-module-blueprint-and-atomic-component.md
    --json` validates one document and emits JSON.
- **AC-SPEC-011.3:** The CLI SHALL expose a `slug-exists` subcommand
  answering the query of AC-SPEC-007.2.
  - *Example:* `slug-exists frd-checkout`.
- **AC-SPEC-011.4:** The CLI SHALL expose a `next-id` subcommand
  answering the query of AC-SPEC-007.1.
  - *Example:* `next-id adr` returns `014`.
- **AC-SPEC-011.5:** WHEN invoked with no subcommand or with `--help`,
  the CLI SHALL list the available subcommands and their options rather
  than performing a default action.
  - *Example:* Bare invocation prints the four subcommands above with a
    one-line synopsis of each.

*Rationale:* The subcommand vocabulary is the tool's public interface,
so it is fixed here rather than left to implementation. Later features
extend `check` with additional checks and add their own query
subcommands, but they do not rename or repurpose the four established
here. The name of the executable itself is a distribution decision
recorded separately; this requirement fixes only the subcommands beneath
it.

## References

- [[adr-015]] — records Go as the implementation language, deferred to
  by REQ-SPEC-009
- [[adr-020]] — settles the acceptance-criteria notation, whose
  extraction is a follow-on feature rather than part of this one
- [[adr-002]] — EARS, the notation these acceptance criteria are
  written in
- `plugin/skills/spec/references/template-authoring-standards.md` —
  frontmatter properties, slug conventions, and the frontmatter index
  this feature implements
- `plugin/skills/spec/references/document-readiness-checklist.md` — the
  checks whose mechanical metadata subset this feature absorbs
- `plugin/skills/spec/references/frontmatter-standards.md` — the status
  vocabulary and element record schemas
- `docs/plans/spec-validation-cli-brief.md` — the research and prototype
  results behind this feature, including the corpus measurements cited
  in the Overview
