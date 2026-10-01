---
title: "FRD: Spec Corpus Index and Core Conformance Checks"
description: A command-line tool that indexes a specification corpus into
  machine-readable JSON and reports deterministic frontmatter and metadata
  conformance findings, so CI, agents, and architecture tooling read one
  authoritative artifact
slug: frd-spec-index
type: frd
status: approved
updated_at: 2026-10-01T00:00:00Z
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
  scope for this feature. [[adr-021]] settles the notation and assigns
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
  channel.** FEAT-SPEC-009 states the latency, portability, and
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
  re-reading the corpus (FEAT-SPEC-010).
- **Latency target** — a stated elapsed-time figure for an operation,
  measured at the 95th percentile. A target is advisory: it is measured
  and reported wherever the suite runs, and a measurement over target is
  a signal to investigate rather than a failure, except in a reference
  environment.
- **Reference environment** — a machine an operator has explicitly
  declared to be the one a latency target is enforced on. Nothing
  declares itself one, and neither a shared CI runner nor a developer
  machine under unknown load is one by default.

## Requirements

### FEAT-SPEC-001: Corpus Discovery and Membership

**Acceptance Criteria:**

```gherkin
@FEAT-SPEC-001
Feature: Corpus Discovery and Membership

  As an adopting team
  I want the tool to work out for itself which files are specifications
  So that I need not enumerate documents or restructure my directories

  Membership is determined by content, not location: a .md file whose
  frontmatter parses and declares a type the active profile knows. That
  is what lets the tool run against a repository it did not design.

  An exclusion is recorded rather than dropped, so that an agent can
  notice that a document which should be a specification is missing
  frontmatter — a judgment the CLI cannot make.

  @AC-SPEC-001.1
  Scenario: A directory yields every conforming markdown file beneath it
    When the CLI is invoked against a directory
    Then the CLI treats every ".md" file beneath it whose frontmatter parses and declares a type known to the active profile as a member of the corpus

  @AC-SPEC-001.2
  Scenario: A single file is checked against the whole corpus
    When the CLI is invoked against a single file path
    Then the CLI checks that file
    And the CLI resolves its references against the full corpus discovered from the configured spec root

  @AC-SPEC-001.3
  Scenario: A markdown file with no frontmatter is recorded as an exclusion
    Given a ".md" file beneath the spec root that has no frontmatter
    When the CLI discovers the corpus
    Then the CLI records it in the index as excluded with the reason "no-frontmatter" rather than omitting it silently

  @AC-SPEC-001.4
  Scenario: A markdown file declaring an unknown type is recorded with the value it declared
    Given a ".md" file declaring a type the active profile does not define
    When the CLI discovers the corpus
    Then the CLI records it as excluded with the reason "unknown-type"
    And the CLI reports the declared value

  @AC-SPEC-001.5 @negative
  Scenario: Nothing outside the spec root is read
    When the CLI resolves corpus membership
    Then the CLI reads no file outside the configured spec root

  @AC-SPEC-001.6
  Scenario: A file that is not markdown is passed over silently
    Then the CLI considers only files whose name ends in ".md" as candidate corpus members
    And the CLI passes over any other file silently rather than recording it as an exclusion
```

*Examples:* invoked against `docs/`, the CLI admits
`docs/adrs/ADR-011-targeted-verification-gates.md`, which declares
`type: adr` (AC-SPEC-001.1). Checking one ADR still reports a wikilink in
it that matches no document anywhere in `docs/` (AC-SPEC-001.2).
`docs/plans/plan.md` appears in the index's exclusions list with reason
`no-frontmatter` (AC-SPEC-001.3). A file declaring `type: rfc` under a
profile with no `rfc` type is listed with `"type": "rfc"`, so a reviewing
agent can judge whether the profile or the document is wrong
(AC-SPEC-001.4). `docs/specs/specs.json` and `docs/adrs/README.rst` are
ignored without appearing in the exclusions list, while
`docs/adrs/ADR-004-module-blueprint-and-atomic-component.md` is
considered (AC-SPEC-001.6).

### FEAT-SPEC-002: Document Index

**Acceptance Criteria:**

```gherkin
@FEAT-SPEC-002
Feature: Document Index

  As an architecture owner
  I want one machine-readable artifact describing every specification document and how they relate
  So that CI, agents and a central catalogue all read the same source of truth instead of each re-deriving it

  The index is deliberately additive: this sprint fills frontmatter and
  path, and later features populate the remaining properties without
  changing the schema's shape or its version.

  Rule: A build failure is the checker's job, not the indexer's

    @AC-SPEC-002.1
    Scenario: The index carries every corpus member's frontmatter and path
      When the CLI builds an index
      Then the CLI emits a JSON document containing, for every corpus member, its file path and every frontmatter property its document type defines, not only the properties every type shares

    @AC-SPEC-002.2
    Scenario: The index declares a schema version
      Then the index declares a schema version identifier

    @AC-SPEC-002.3
    Scenario: Every entry declares its workspace
      Then the index declares the workspace identifier from configuration for every document it contains

    @AC-SPEC-002.4
    Scenario: An unparseable document does not abort the index
      Given a document in the corpus that cannot be parsed
      When the CLI builds an index
      Then the CLI indexes every other document
      And the CLI records the failure against the offending document rather than aborting the index

    @AC-SPEC-002.5
    Scenario: An index containing parse failures still exits zero
      When the CLI has built an index that contains parse failures
      Then the CLI exits zero
```

*Examples:* the entry for `adr-011` carries its six frontmatter
properties and its path
`docs/adrs/ADR-011-targeted-verification-gates.md`, and later features
add its headings and outbound references to the same entry
(AC-SPEC-002.1). The emitted JSON opens with `"schema_version": "1.0"`
(AC-SPEC-002.2). Every entry carries `"workspace": "agent-plugins"`, so
merging this index with another repository's cannot collide two
documents that both use the slug `adr-001` (AC-SPEC-002.3).
`work-order-template.md` fails YAML parsing at line 28; the index still
contains the other 49 documents and lists that file with its parse error
and position (AC-SPEC-002.4).

*Rationale:* AC-SPEC-002.4 and AC-SPEC-002.5 together are what make the
index usable as an architecture artifact. A repository with three broken
documents must still contribute its other nine hundred to an
organization's graph.

### FEAT-SPEC-003: Metadata Conformance

**Acceptance Criteria:**

```gherkin
@FEAT-SPEC-003
Feature: Metadata Conformance

  As a reviewer
  I want frontmatter errors reported mechanically
  So that review attention goes to whether a document says the right thing rather than whether it is shaped correctly

  @AC-SPEC-003.1
  Scenario: Frontmatter is validated against the schema for its type
    When the CLI checks a document
    Then the CLI validates its frontmatter against the JSON Schema the active profile registers for that document's type

  @AC-SPEC-003.2
  Scenario: A textual format is checked against the raw text in the file
    When the CLI validates a frontmatter property whose schema constrains its textual format
    Then the CLI validates the raw text of the property as written in the file

  @AC-SPEC-003.3
  Scenario: Two members declaring one slug each name the other
    Given two corpus members declaring the same slug
    When the CLI checks the corpus
    Then the CLI reports a finding against each, naming the other's path

  @AC-SPEC-003.4
  Scenario: Slug uniqueness ignores letter case
    Then the CLI compares slugs for uniqueness without regard to letter case

  @AC-SPEC-003.5
  Scenario: A slug not matching its type's pattern is reported with both
    Given a document whose slug does not match the pattern the active profile registers for its type
    When the CLI checks the document
    Then the CLI reports a finding naming both the actual slug and the expected pattern

  @AC-SPEC-003.6
  Scenario: A member indexed with a parse error is reported, never reported as passing
    Given a corpus member indexed with a parse error rather than parsed frontmatter
    When the CLI checks it
    Then the CLI reports a finding carrying that error's message and span
    And the CLI reports the document as having passed no check
```

*Examples:* a document declaring `type: adr` but omitting `description`
produces one finding naming the missing property (AC-SPEC-003.1).
`updated_at: 2026-07-29` fails an RFC 3339 date-time check, even though a
YAML parser would coerce it to a valid timestamp before any check could
see it (AC-SPEC-003.2). Two documents both declaring `slug: adr-012` each
produce a finding citing the other's file path (AC-SPEC-003.3).
`frd-checkout` and `FRD-Checkout` are reported as duplicates, because a
case-insensitive filesystem cannot hold both as distinct files
(AC-SPEC-003.4). `slug: adr-011-targeted-verification-gates` under a
profile registering `^adr-[0-9]{3}$` for type `adr` produces a finding
quoting both (AC-SPEC-003.5). A document copied from a template and not
yet completed carries `updated_at: <YYYY-MM-DDTHH:MM:SSZ>`, which is not
an RFC 3339 date-time; the check run reports it and exits non-zero,
rather than reporting four passed checks against a document nothing was
able to read (AC-SPEC-003.6).

*Rationale for AC-SPEC-003.6:* every other criterion in this requirement
presupposes parsed frontmatter, so before this was written the worst-formed
document in a corpus was the one the checker was guaranteed to be silent
about — the severity of a defect and the tool's willingness to mention it
ran in opposite directions. FEAT-SPEC-002's rationale already assigned this
work: an index containing parse failures exits zero because "a build
failure is the checker's job, not the indexer's". This is the criterion
that makes the checker do that job. A file with no frontmatter at all
remains an exclusion under AC-SPEC-001.3 and is *not* covered here; it
never claimed to be a specification, whereas a member with a parse error
did and got the claim wrong.

*Note on numbering:* FEAT-SPEC-004 (reference integrity) and FEAT-SPEC-005
(structure and hygiene conformance) covered behavior deferred to
follow-on features. Their numbers are intentionally left unused here
rather than reassigned: the retained requirements FEAT-SPEC-006, -008,
and -009 are cited by ADRs and keep their numbers, so renumbering around
the gap would break those citations. The gap is deliberate.

### FEAT-SPEC-006: Findings, Severities, and Diagnostics

**Acceptance Criteria:**

```gherkin
@FEAT-SPEC-006
Feature: Findings, Severities, and Diagnostics

  As an agent acting on the output
  I want findings that name the rule and cite where it is defined
  So that I can fix the problem without loading the standards documents that define it

  A diagnostic that quotes the violated constraint and the offending
  value is self-contained: the agent fixes it without opening the
  standard that defines the rule. That is what lets the spec skill stop
  preloading verification procedure.

  @AC-SPEC-006.1
  Scenario: Every finding is self-contained
    Then every finding the CLI reports carries a stable rule identifier, a severity, a file path, a span, and a message stating the violated constraint and the value that violated it

  @AC-SPEC-006.3
  Scenario: A check that cannot run is reported as unverified
    When the CLI is asked to perform a check it cannot run
    Then the CLI reports that check with severity "unverified"

  @AC-SPEC-006.4 @negative
  Scenario: A check that could not run is never reported as passed
    Then the CLI does not report a check it could not run as passed

  @AC-SPEC-006.5
  Scenario: The full finding set is available as JSON
    When the check subcommand is invoked with the JSON option
    Then the CLI writes the full finding set as JSON to standard output

  @AC-SPEC-006.6
  Scenario: An error fails the run
    Given a check run producing at least one finding of severity "error"
    Then the CLI exits with a non-zero status

  @AC-SPEC-006.7
  Scenario: Warnings and unverified findings pass unless strict
    Given a check run producing findings only of severity "warning" or "unverified"
    Then the CLI exits zero unless invoked with the strict option

  @AC-SPEC-006.8
  Scenario: A named subset of checks runs alone
    When the check subcommand is invoked with an option naming a subset of checks
    Then the CLI runs only the named checks

  @AC-SPEC-006.9 @negative
  Scenario: A check that was not requested is never reported as passed
    Then the CLI does not report a check that was not requested as passed
```

*Examples:* `slug/pattern-mismatch  error
docs/adrs/ADR-011-targeted-verification-gates.md:4:7`
"'adr-011-targeted-verification-gates' does not match pattern
'^adr-[0-9]{3}$'" — self-contained, so the fixing agent needs no further
lookup (AC-SPEC-006.1). Asked to validate mermaid diagrams with no
diagram validator available, the CLI emits one `unverified` finding per
diagram naming the missing dependency (AC-SPEC-006.3). An agent running
`check --only slug-uniqueness` gets a report covering that check alone
(AC-SPEC-006.8), and that report distinguishes "slug uniqueness passed"
from "frontmatter schema was not checked" (AC-SPEC-006.9).

### FEAT-SPEC-007: Corpus Queries

**Acceptance Criteria:**

```gherkin
@FEAT-SPEC-007
Feature: Corpus Queries

  As an authoring agent
  I want to ask the corpus direct questions
  So that I can pick a conforming identifier and understand impact without reading every document

  Both queries answer from frontmatter and path alone, which is why they
  belong here. Queries that need more of the index are deferred to the
  later features that populate those parts of it.

  @AC-SPEC-007.1
  Scenario: The next identifier for a numbered type is one past the highest in use
    When the next-id subcommand is invoked for a document type the active profile registers as numbered
    Then the CLI returns one greater than the highest number in use for that type

  @AC-SPEC-007.2
  Scenario: A slug query reports whether it is claimed, and by which document
    When the slug-exists subcommand is invoked with a slug
    Then the CLI reports whether a corpus member declares it
    And where one does, the CLI reports the path of that member
```

*Examples:* with ADRs numbered 001 through 013 in the corpus,
`next-id adr` returns `014` (AC-SPEC-007.1). `slug-exists adr-004`
reports the slug in use and names
`docs/adrs/ADR-004-module-blueprint-and-atomic-component.md`
(AC-SPEC-007.2).

### FEAT-SPEC-008: Configuration

**Acceptance Criteria:**

```gherkin
@FEAT-SPEC-008
Feature: Configuration

  As an adopting team
  I want to declare my own document types and conventions
  So that the tool enforces my framework rather than the one it shipped with

  @AC-SPEC-008.1
  Scenario: Configuration is one TOML file at the repository root
    Then the CLI reads its configuration from a TOML file at the repository root

  @AC-SPEC-008.2
  Scenario: The configuration carries the root, the workspace and the type vocabulary
    Then the configuration determines the spec root, the workspace identifier, and the registered document types with their slug patterns and the JSON Schema that validates each type's frontmatter

  @AC-SPEC-008.3
  Scenario: An absent configuration is reported rather than guessed at
    Given no configuration file is present
    When the CLI runs
    Then the CLI reports the absence
    And the CLI exits non-zero rather than assuming a layout

  @AC-SPEC-008.4
  Scenario: An invalid configuration names the offending key and its position
    Given a configuration that is present but invalid
    When the CLI reads it
    Then the CLI reports the offending key and its position

  @AC-SPEC-008.5 @negative
  Scenario: An invalid key is never quietly defaulted
    Given a configuration that is present but invalid
    When the CLI reads it
    Then the CLI falls back to no default value for the offending key

  @AC-SPEC-008.6
  Scenario: A configuration for this framework's own types ships with the CLI
    Then the CLI ships a configuration for this framework's own document types that an adopting repository can use without authoring one
```

*Examples:* `specs.toml` declares `root = "docs/specs"` and the workspace
identifier (AC-SPEC-008.1). A `[[type]]` table for `adr` declares
`slug_pattern = "^adr-[0-9]{3}$"` and points at the schema FEAT-SPEC-003
validates against; later features add each type's required headings, its
line-width limit and its reference-bearing properties to this same file,
the config schema being additive (AC-SPEC-008.2). Run in a repository
with no `specs.toml`, the CLI names the file it looked for and the path
it looked in (AC-SPEC-008.3).

### FEAT-SPEC-009: Invocation, Performance, and Portability

**Acceptance Criteria:**

```gherkin
@FEAT-SPEC-009
Feature: Invocation, Performance, and Portability

  As a platform engineer
  I want to run this in locked-down CI and in an editing loop that fires on every file write
  So that neither the build nor the author waits on it

  The three latency figures are targets rather than gates. They are
  stated against no hardware, and the same code measures 13.7 ms on
  Linux, 69.1 ms on macOS and 123.6 ms on Windows, so a shared runner's
  number measures the runner. Gated, they would also make any change that
  costs milliseconds for correctness' sake indistinguishable from a
  defect.

  Scenario Outline: Latency targets against a 1,000-document corpus
    Given a corpus of 1,000 documents
    When the CLI performs "<operation>"
    Then the CLI holds a latency target of "<target>" at the 95th percentile

    @AC-SPEC-009.1
    Examples: an agent asking one question
      | operation                           | target          |
      | a corpus query against a warm index | 50 milliseconds |

    @AC-SPEC-009.2
    Examples: a write-time hook checking one file
      | operation                                    | target           |
      | a single-document check against a warm index | 100 milliseconds |

    @AC-SPEC-009.3
    Examples: a build with nothing cached
      | operation                    | target   |
      | an index build with no cache | 1 second |

  @AC-SPEC-009.4 @manual
  Scenario: The CLI needs no runtime beyond itself
    Then the CLI runs on a machine with no language runtime, interpreter, or package manager installed beyond the CLI itself

  @AC-SPEC-009.5 @manual @negative
  Scenario: Installation retrieves no executable content from a network location
    Then the CLI retrieves no executable content from a network location during installation

  @AC-SPEC-009.6 @manual
  Scenario: The CLI is installable from a mirrorable registry
    Then the CLI is installable from at least one package registry that supports on-premises mirroring

  @AC-SPEC-009.7
  Scenario: Identical content yields identical findings on every platform
    When the CLI runs against identical corpus content on Linux, macOS and Windows
    Then the CLI reports the same findings in the same order

  @AC-SPEC-009.8
  Scenario: Paths are forward-slashed and repository-relative
    Then the CLI emits file paths in the index and in findings using forward slashes as separators, relative to the repository root

  @AC-SPEC-009.9
  Scenario: The latency targets are measured everywhere and enforced only in a reference environment
    Then the latency targets of AC-SPEC-009.1 through AC-SPEC-009.3 are measured and reported on every platform the test suite runs on
    And they fail a run that exceeds one only in a reference environment
```

*Examples:* an agent asking `slug-exists frd-checkout` before writing a
document waits imperceptibly (AC-SPEC-009.1), and a validation hook
firing on every document write does not make editing feel slow
(AC-SPEC-009.2). The CLI runs in a container image containing only a
shell and the CLI binary (AC-SPEC-009.4). Installing from a package
registry mirrored inside a network that blocks all other outbound traffic
succeeds (AC-SPEC-009.5). A corpus checked out with CRLF line endings on
Windows yields byte-identical JSON output to the same corpus on Linux
(AC-SPEC-009.7). A contended shared runner that measures 82 milliseconds
against the 50-millisecond query target reports the number and passes;
the same measurement on a machine declared a reference environment fails
(AC-SPEC-009.9).

*Rationale:* AC-SPEC-009.1 through AC-SPEC-009.3 come from the two
invocation patterns that dominate: an agent asking one question, and a
write-time hook checking one file. They are targets rather than gates
(AC-SPEC-009.9) because they are stated against no hardware and the same
code measures 13.7 ms on Linux, 69.1 ms on macOS and 123.6 ms on
Windows: the dominant cost on the warm path is one `stat` per document,
so a shared runner's number measures the runner. Gated, they also make
any change that costs milliseconds for correctness' sake — a decoder
that must distinguish an absent required property from a zero value, for
one — indistinguishable from a defect, and the margin a stated hardware
class would preserve is a few milliseconds wide. The targets stay
stated, measured and reported, so a real slowdown is still visible in
any run's log. AC-SPEC-009.5 exists because the common
packaging pattern of downloading a binary during install fails in
exactly the networks that most need this tool, and adds a supply-chain
surface an adopting security review will reject.

### FEAT-SPEC-010: Index Output and Persistence

**Acceptance Criteria:**

```gherkin
@FEAT-SPEC-010
Feature: Index Output and Persistence

  As a CI pipeline and as an authoring agent
  I want the index written to one known file that I can publish and that later invocations reuse
  So that the artifact pushed to a central catalogue is the same one that answers my next query without re-reading the corpus

  One file serves both of the index's roles: the deliverable a CI job
  publishes to a catalogue, and the warm cache that satisfies the latency
  targets of FEAT-SPEC-009. Placing it inside the spec root keeps the
  artifact beside the documents it describes, and because discovery reads
  only ".md" files it is never mistaken for a corpus member.

  Whether an adopting repository commits the file or ignores it is the
  adopter's choice; the CLI requires neither.

  @AC-SPEC-010.1
  Scenario: The index lands at one known path by default
    When the CLI is asked to build the index
    Then the CLI writes the JSON to a single file named "specs.json" inside the configured spec root

  @AC-SPEC-010.2
  Scenario: An output path overrides the default location
    When the CLI is invoked with an output-path option naming a location
    Then the CLI writes the index to that location instead of the default

  @AC-SPEC-010.3
  Scenario: An output path of a single dash streams the index and leaves no file
    When the CLI is invoked with the output-path option set to a single dash
    Then the CLI writes the index to standard output
    And the CLI neither creates nor modifies the default file

  @AC-SPEC-010.4
  Scenario: A concurrent reader never observes a half-written index
    When the CLI writes the index to a file
    Then the CLI replaces any existing file atomically, so that a concurrent reader observes either the old complete index or the new one and never a partially written file

  @AC-SPEC-010.5
  Scenario: A warm index answers without re-walking the corpus
    Given the index file exists with inputs unchanged since it was written
    When a corpus query or single-document check runs
    Then the CLI answers from that persisted index — the warm index of FEAT-SPEC-009 — without re-walking the corpus

  @AC-SPEC-010.6
  Scenario: A stale or absent index is rebuilt rather than served
    Given the index file is absent, or any corpus input has changed since it was written
    When a query runs
    Then the CLI rebuilds the index before answering rather than serving a stale result
```

*Examples:* with `root = "docs/specs"` in the configuration, running
`index` with no output option writes `docs/specs/specs.json` and the CLI
prints that relative path (AC-SPEC-010.1). `--out build/spec-index.json`
writes the artifact under `build/` for a CI job that publishes everything
in that directory (AC-SPEC-010.2). `--out -` lets an agent pipe the index
straight into `jq` without leaving a file behind (AC-SPEC-010.3). A
`PostToolUse` hook rebuilding the index while a query reads it never
yields the query a truncated document (AC-SPEC-010.4).
`slug-exists frd-checkout` run twice in a row reads `specs.json` the
second time and returns within the AC-SPEC-009.1 budget
(AC-SPEC-010.5). Editing one ADR and immediately querying returns an
answer reflecting the edit, because the changed input invalidated the
warm index (AC-SPEC-010.6).

*Rationale:* The index artifact and the check report are different
outputs: FEAT-SPEC-006.5 writes the *findings* of a check run to standard
output, while this requirement governs where the *index* — the primary
artifact of FEAT-SPEC-002 — is persisted. One file serves both of the
index's roles: the deliverable a CI job publishes to a catalogue, and
the warm cache that satisfies the latency budgets of FEAT-SPEC-009.
Placing `specs.json` inside the spec root keeps the artifact beside the
documents it describes, and because discovery reads only `.md` files
(AC-SPEC-001.6) the index is never mistaken for a corpus member. A fixed
default name means every skill body, CI config, and hook can refer to
`specs.json` without configuration, while the output-path option keeps
the CI-publish and pipe-to-stdout cases first-class. Whether an adopting
repository commits the file or ignores it is the adopter's choice; the
CLI requires neither.

### FEAT-SPEC-011: Command Surface

**Acceptance Criteria:**

```gherkin
@FEAT-SPEC-011
Feature: Command Surface

  As a human or agent developer integrating the tool
  I want a small, named set of subcommands
  So that skills, hooks and CI scripts call a stable interface rather than reverse-engineering it from behaviour

  The subcommand vocabulary is the tool's public interface, so it is
  fixed here rather than left to implementation. Later features extend
  the check subcommand and add their own query subcommands, but they do
  not rename or repurpose the four established here.

  The name of the executable itself is a distribution decision recorded
  separately; this requirement fixes only the subcommands beneath it.

  @AC-SPEC-011.1
  Scenario: An index subcommand builds and writes the index
    Then the CLI exposes an index subcommand that builds the index and writes it per FEAT-SPEC-010, accepting the output-path option named there

  @AC-SPEC-011.2
  Scenario: A check subcommand runs the conformance checks
    Then the CLI exposes a check subcommand that runs conformance checks over the corpus, or over a single document when given a path, and reports findings per FEAT-SPEC-006, accepting the JSON, strict and only options

  @AC-SPEC-011.3
  Scenario: A slug-exists subcommand answers the slug query
    Then the CLI exposes a slug-exists subcommand answering the query of AC-SPEC-007.2

  @AC-SPEC-011.4
  Scenario: A next-id subcommand answers the numbering query
    Then the CLI exposes a next-id subcommand answering the query of AC-SPEC-007.1

  @AC-SPEC-011.5
  Scenario: A bare invocation lists the surface rather than acting
    When the CLI is invoked with no subcommand or with the help option
    Then the CLI lists the available subcommands and their options rather than performing a default action
```

*Examples:* `index` writes the default `specs.json` and `index --out -`
streams it to standard output (AC-SPEC-011.1). `check` validates the
whole corpus, while
`check docs/adrs/ADR-004-module-blueprint-and-atomic-component.md --json`
validates one document and emits JSON (AC-SPEC-011.2).
`slug-exists frd-checkout` (AC-SPEC-011.3). `next-id adr` returns `014`
(AC-SPEC-011.4). Bare invocation prints the four subcommands above with a
one-line synopsis of each (AC-SPEC-011.5).

*Rationale:* The subcommand vocabulary is the tool's public interface,
so it is fixed here rather than left to implementation. Later features
extend `check` with additional checks and add their own query
subcommands, but they do not rename or repurpose the four established
here. The name of the executable itself is a distribution decision
recorded separately; this requirement fixes only the subcommands beneath
it.

## References

- [[adr-015]] — records Go as the implementation language, deferred to
  by FEAT-SPEC-009
- [[adr-021]] — Gherkin, the notation these features and acceptance
  criteria are written in, whose extraction is a follow-on feature
  rather than part of this one
- [[gherkin-guide]] — how to write them
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
