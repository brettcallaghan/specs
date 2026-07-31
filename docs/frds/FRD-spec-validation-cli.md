---
title: "FRD: Spec Corpus Index and Conformance Checks"
description: A command-line tool that indexes a specification corpus into
  machine-readable JSON and reports deterministic conformance findings, so
  CI, agents, and architecture tooling read one authoritative artifact
slug: frd-spec-validation-cli
type: frd
status: approved
updated_at: 2026-07-31T00:00:00Z
sources: []
related:
  - "adr-012"
  - "adr-014"
---

# Spec Corpus Index and Conformance Checks

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
document's metadata, structure, and relationships, and a **report of
conformance findings** against the rules the framework already defines.
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

- **Modifying documents.** This feature only reads and reports.
  Deriving `updated_at` from git history and rewriting frontmatter is a
  follow-on feature.
- **Acceptance-criteria and feature-definition validation.** Deferred
  per [[adr-014]], which also records the reversibility constraint that
  decision places on this tool. Deferred is not abandoned: the
  technology decision this feature defers to an ADR must preserve the
  ability to parse and validate structured feature definitions later.
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
  markdown file with parseable frontmatter declaring a known `type`.
- **Spec root** — the configured directory the CLI walks to discover the
  corpus.
- **Workspace** — a named collection of documents from one repository,
  identified in configuration so that indexes produced by separate
  repositories can be combined without slug collisions.
- **Index** — the machine-readable JSON artifact describing every
  discovered document: its frontmatter, headings and their spans,
  extracted structured blocks, and its outbound and inbound references.
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
  sections — a `[[slug]]` reference occupies part of a line, not the
  line itself.
- **Record block** — a fenced code block carrying structured data about
  an element declared inside a document, per [[adr-008]].
- **Warm index** — a cached index whose inputs are unchanged since it was
  built, so a query can be answered without re-reading the corpus.

## Requirements

### REQ-SPEC-001: Corpus Discovery and Membership

**User Story:** As an adopting team, I want to point the tool at my
repository and have it work out which files are specifications, so that
I do not have to enumerate documents or restructure my directories.

**Acceptance Criteria:**

- **AC-SPEC-001.1:** WHEN the CLI is invoked against a directory, the
  CLI SHALL treat every markdown file beneath it whose frontmatter
  parses and declares a `type` known to the active profile as a member
  of the corpus.
  - *Example:* Invoked against `docs/`, it admits
    `docs/adrs/ADR-011-targeted-verification-gates.md`, which declares
    `type: adr`.
- **AC-SPEC-001.2:** WHEN the CLI is invoked against a single file path,
  the CLI SHALL check that file and resolve its references against the
  full corpus discovered from the configured spec root.
  - *Example:* Checking one ADR still reports a `[[slug]]` in it that
    matches no document anywhere in `docs/`.
- **AC-SPEC-001.3:** IF a markdown file beneath the spec root has no
  frontmatter, THEN the CLI SHALL record it in the index as excluded
  with the reason `no-frontmatter` rather than omitting it silently.
  - *Example:* `docs/plans/plan.md` appears in the index's exclusions
    list with reason `no-frontmatter`.
- **AC-SPEC-001.4:** IF a markdown file declares a `type` the active
  profile does not define, THEN the CLI SHALL record it as excluded with
  the reason `unknown-type` and report the declared value.
  - *Example:* A file declaring `type: rfc` under a profile with no
    `rfc` type is listed with `"type": "rfc"`, so a reviewing agent can
    judge whether the profile or the document is wrong.
- **AC-SPEC-001.5:** The CLI SHALL NOT read files outside the configured
  spec root when resolving corpus membership.

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
  properties, its file path, its headings with the span of the content
  beneath each, and a `references` property holding every `[[slug]]`
  parsed from its body, each with the span at which it occurs.
  - *Example:* The entry for `adr-011` carries its six frontmatter
    properties, the spans of its Context, Decision, Alternatives
    Considered, Consequences and References headings, and its outbound
    references to `adr-004` and `adr-006`.
- **AC-SPEC-002.2:** The index SHALL declare a schema version
  identifier.
  - *Example:* The emitted JSON opens with
    `"schema_version": "1.0"`.
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
- **AC-SPEC-002.5:** WHEN the CLI builds an index, the CLI SHALL emit
  inbound references for every document, derived from the outbound
  references of all other corpus members.
  - *Example:* `adr-004`'s entry lists `adr-005`, `adr-006` and
    `adr-007` as inbound, without any of them declaring the
    relationship in frontmatter.
- **AC-SPEC-002.6:** WHEN the CLI encounters a fenced block whose info
  string matches a tag the active profile registers, the CLI SHALL parse
  its content as the configured format and attach the parsed data to the
  containing document and heading.
  - *Example:* A profile registering the tag `component` with format
    `yaml` causes each component record block in a container blueprint
    to appear in the index as structured data under the heading that
    declares it.
- **AC-SPEC-002.7:** The CLI SHALL exit zero after building an index
  that contains parse failures.

*Rationale:* AC-SPEC-002.4 and AC-SPEC-002.7 together are what make the
index usable as an architecture artifact. A repository with three broken
documents must still contribute its other nine hundred to an
organization's graph; a build failure is the checker's job, not the
indexer's.

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

### REQ-SPEC-004: Reference Integrity

**User Story:** As a document author, I want every cross-reference
checked, so that a link that silently stopped resolving is caught before
a reader follows it into nothing.

**Acceptance Criteria:**

- **AC-SPEC-004.1:** WHEN a document body contains a `[[slug]]`
  reference that matches no corpus member's slug, the CLI SHALL report a
  finding with the file, line, and unresolved slug.
  - *Example:* `[[adr-011]]` in `ADR-004` reports as unresolved while
    that ADR's frontmatter declares
    `slug: adr-011-targeted-verification-gates`.
- **AC-SPEC-004.2:** The CLI SHALL NOT treat text inside fenced code
  blocks or inline code spans as a reference.
  - *Example:* The sentence "write `` `[[security-standards]]` `` to cite
    a standard" in a guide produces no finding, because the reference is
    an illustration.
- **AC-SPEC-004.3:** WHEN a reference names a heading anchor and the
  target document has no heading matching it, the CLI SHALL report a
  finding naming the target document and the missing anchor.
  - *Example:* `[[container-api#Componets]]` reports `no heading named
    'Componets' found in container-api`. The finding states the fact
    rather than proposing a correction — an agent resolves a plain
    "not found" faster than a guess it must then verify.
- **AC-SPEC-004.4:** WHEN a frontmatter property registered by the
  active profile as holding document references contains a value
  matching no corpus member's slug, the CLI SHALL report a finding
  naming the property and the value.
  - *Example:* `related: ["work-order-template"]` reports as
    unresolved, because templates carry placeholder slugs and are not
    corpus members.
- **AC-SPEC-004.5:** WHEN a document declares `status: retired` without
  a resolvable `superseded_by` reference, the CLI SHALL report a
  finding.
- **AC-SPEC-004.6:** WHEN the active profile registers a frontmatter
  property as one half of a paired relationship and the named document
  does not declare the matching counterpart, the CLI SHALL report a
  finding against both documents.
  - *Example:* A blueprint declaring `children: [component-ledger]`
    where `component-ledger` declares no `parent_slug` produces a
    finding on each.

### REQ-SPEC-005: Structure and Hygiene Conformance

**User Story:** As a reviewer, I want documents checked for skipped
sections and leftover authoring scaffolding, so that an unfinished
document cannot reach review looking finished.

**Acceptance Criteria:**

- **AC-SPEC-005.1:** WHEN the active profile registers required
  headings for a document's `type` and a document omits one, the CLI
  SHALL report a finding naming the missing heading.
  - *Example:* A document declaring `type: adr` with no "Alternatives
    Considered" heading produces a finding naming it.
- **AC-SPEC-005.2:** WHEN a document contains an HTML comment block
  outside a fenced code block and outside an inline code span, the CLI
  SHALL report a finding with its span.
  - *Example:* A finished ADR still carrying its template's Context
    instruction block is reported at the position the block opens, while
    a module blueprint whose prose describes the `<!--` delimiter inside
    backticks produces no finding.
- **AC-SPEC-005.3:** WHEN a prose line exceeds the configured width, the
  CLI SHALL report a finding naming the line and its length.
  - *Example:* A line of 85 characters is reported when the configured
    width is 80.
- **AC-SPEC-005.4:** The CLI SHALL measure line width in Unicode
  codepoints.
  - *Example:* A line containing an em dash counts that dash as one
    character, not as its three UTF-8 bytes.
- **AC-SPEC-005.5:** The CLI SHALL NOT report a width finding for a line
  inside a fenced code block, a table row, or a line containing a URL.
- **AC-SPEC-005.6:** WHEN a document whose `status` the active profile
  marks as finished contains a configured placeholder pattern, the CLI
  SHALL report a finding.
  - *Example:* A document at `status: approved` containing
    `<Feature Name>` is reported; the same text in a `draft` document is
    not.
- **AC-SPEC-005.7:** WHEN a registered structured block fails to parse
  as its configured format, the CLI SHALL report a finding with the
  parse error and its position within the document.

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
- **AC-SPEC-006.5:** WHEN invoked with a machine-readable output option,
  the CLI SHALL write the full finding set as JSON to standard output.
- **AC-SPEC-006.6:** WHEN a check run produces at least one finding of
  severity `error`, the CLI SHALL exit with a non-zero status.
- **AC-SPEC-006.7:** WHEN a check run produces findings only of severity
  `warning` or `unverified`, the CLI SHALL exit zero unless invoked with
  the strict option.
- **AC-SPEC-006.8:** WHEN invoked with an option naming a subset of
  checks, the CLI SHALL run only the named checks.
  - *Example:* An agent verifying only reference integrity gets a report
    covering links alone.
- **AC-SPEC-006.9:** The CLI SHALL NOT report a check that was not
  requested as passed.
  - *Example:* The same report distinguishes "links passed" from
    "frontmatter was not checked".

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

- **AC-SPEC-007.1:** WHEN asked for the next available identifier for a
  document type the active profile registers as numbered, the CLI SHALL
  return one greater than the highest number in use for that type.
  - *Example:* With ADRs numbered 001 through 013 in the corpus, the
    query returns `014`.
- **AC-SPEC-007.2:** WHEN asked whether a slug is in use, the CLI SHALL
  report whether a corpus member declares it and, IF one does, the path
  of that member.
- **AC-SPEC-007.3:** WHEN asked for the documents referencing a given
  slug, the CLI SHALL return every corpus member whose body or
  registered frontmatter properties reference it.
  - *Example:* Asked about `adr-007`, it returns the seven documents
    citing it, so an author retiring it knows the blast radius.
- **AC-SPEC-007.4:** WHEN asked for the spans of headings matching a
  configured selection, the CLI SHALL return each match as a file path,
  heading text, and the span of the content beneath it.
  - *Example:* A reviewing agent requests the normative sections of
    eleven standards documents and receives spans totalling a fraction
    of those documents' text, then reviews only that text.

*Rationale:* AC-SPEC-007.4 exists to make agent review auditable as well
as cheap. Two runs that review the same extracted spans have reviewed
the same text; two agents told to "review the requirements" have not.

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
  the workspace identifier, the registered document types with their
  slug patterns and required headings, the registered structured-block
  tags, and the line-width limit.
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
  itself.
  - *Example:* It runs in a container image containing only a shell and
    the CLI binary.
- **AC-SPEC-009.5:** The CLI SHALL NOT retrieve executable content from
  a network location during installation.
  - *Example:* Installing from a package registry mirrored inside a
    network that blocks all other outbound traffic succeeds.
- **AC-SPEC-009.6:** The CLI SHALL be installable from at least one
  package registry that supports on-premises mirroring.
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

## References

- [[adr-014]] — defers acceptance-criteria and feature-definition
  validation out of this feature, and records the constraint that
  criterion notation stays a replaceable input
- [[adr-012]] — the `[[slug]]` and heading-anchor conventions
  REQ-SPEC-004 checks
- [[adr-008]] — the element record blocks REQ-SPEC-002 extracts and
  REQ-SPEC-005 validates
- [[adr-002]] — EARS, the notation these acceptance criteria are
  written in
- `plugin/skills/spec/references/template-authoring-standards.md` —
  frontmatter properties, slug conventions, line width, and the
  frontmatter index this feature implements
- `plugin/skills/spec/references/cross-document-references.md` — what a
  reference resolves against
- `plugin/skills/spec/references/document-readiness-checklist.md` — the
  eight checks whose mechanical subset this feature absorbs
- `plugin/skills/spec/references/frontmatter-standards.md` — the status
  vocabulary and element record schemas
- `docs/plans/spec-validation-cli-brief.md` — the research and prototype
  results behind this feature, including the corpus measurements cited
  in the Overview
