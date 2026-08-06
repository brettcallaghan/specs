---
title: "FRD: Spec Corpus Index and Core Conformance Checks (Gherkin variant)"
description: A Gherkin rewrite of the acceptance criteria in
  specs/features/spec-index/spec-index-frd.md, produced to compare the two
  notations on a real document rather than in the abstract
slug: frd-spec-validation-cli-gherkin
type: frd
status: draft
updated_at: 2026-08-06T00:00:00Z
sources: []
related:
  - "adr-018"
---

<!--
COMPARISON ARTIFACT — not a corpus member.

This file lives in docs/, outside the spec root declared in specs.toml, so
`spec check` does not see it and its slug cannot collide with the original's.

It reproduces specs/features/spec-index/spec-index-frd.md under the mapping
ADR-018 constraint 1 fixes:

  REQ        -> Feature:          (one fenced block each)
  User Story -> the Feature description, with a design note and scope fence
  AC         -> Scenario / Scenario Outline, tagged with its identifier
  a business rule grouping several ACs -> Rule:
  *Example:* -> an Examples row

Overview, Terminology and the numbering note remain markdown prose — they
are FRD-level and Gherkin has no node above Feature. The nine per-requirement
*Rationale:* paragraphs are gone: their durable content moved into the
Feature descriptions, and the one that stated a real rule became a Rule:.
See docs/plans/gherkin-feature-description-guidance.md.

Every block below was parsed with github.com/cucumber/gherkin/go/v42 and the
extractor described in docs/plans/gherkin-extraction-findings.md: 9 blocks,
50 scenarios, all 52 AC identifiers, no duplicates, no untagged scenarios.
Findings from the exercise are in docs/gherkin-frd-comparison.md.
-->

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
already defines.

The index is the primary artifact, not a by-product. Every check in this
document is a query over it, and consumers outside this framework
consume the index without running any check at all. That ordering
governs the requirements below: indexing never fails, and validation is
a separate consumer that may.

*(Overview abridged for the comparison; the original section is
unchanged and carries the full out-of-scope list.)*

## Terminology

*(Unchanged from the original — see
`specs/features/spec-index/spec-index-frd.md`.)*

## Requirements

### REQ-SPEC-001: Corpus Discovery and Membership

**Acceptance Criteria:**

```gherkin
@frd-spec-validation-cli
Feature: Corpus Discovery and Membership

  As an adopting team
  I want the tool to work out for itself which files are specifications
  So that I need not enumerate documents or restructure my directories

  Membership is determined by content, not location: a `.md` file whose
  frontmatter parses and declares a type the active profile knows. A file
  that looks like a specification but has no frontmatter is recorded as an
  exclusion rather than dropped, so an agent can notice what the CLI
  cannot judge.

  This requirement reads frontmatter and path only. Resolving references
  and walking body structure are follow-on features.

  @AC-SPEC-001.1
  Scenario Outline: A directory invocation admits every typed markdown file
    Given the file "<file>" beneath the invoked directory
    And its frontmatter parses and declares "type: <type>"
    And the active profile knows that type
    When the CLI is invoked against the directory
    Then the CLI treats the file as a member of the corpus

    Examples:
      | file                                             | type |
      | docs/adrs/ADR-011-targeted-verification-gates.md | adr  |

  @AC-SPEC-001.2
  Scenario: A single-file invocation still resolves against the full corpus
    When the CLI is invoked against a single file path
    Then the CLI checks that file
    And the CLI resolves its references against the full corpus
    And a [[slug]] matching no document anywhere is still reported

  @AC-SPEC-001.3
  Scenario Outline: A file with no frontmatter is recorded, not dropped
    Given the file "<file>" beneath the spec root has no frontmatter
    When the CLI builds the index
    Then the index lists it as excluded with the reason "<reason>"
    And the CLI does not omit it silently

    Examples:
      | file               | reason         |
      | docs/plans/plan.md | no-frontmatter |

  @AC-SPEC-001.4
  Scenario Outline: An unknown type is recorded with its declared value
    Given the file "<file>" declares "type: <type>"
    And the active profile does not define that type
    When the CLI builds the index
    Then the index lists it as excluded with the reason "<reason>"
    And the entry reports the declared value "<type>"

    Examples:
      | file            | type | reason       |
      | docs/rfc-001.md | rfc  | unknown-type |

  @AC-SPEC-001.5 @negative
  Scenario: Files outside the spec root are never read
    When the CLI resolves corpus membership
    Then the CLI does not read files outside the configured spec root

  @AC-SPEC-001.6
  Scenario Outline: Only .md files are candidates; others pass over silently
    Given the candidate file "<file>"
    When the CLI resolves corpus membership
    Then the CLI treats it as "<outcome>"

    Examples:
      | file                                        | outcome               |
      | docs/adrs/ADR-004-module-blueprint.md       | considered            |
      | docs/specs/specs.json                       | ignored, not excluded |
      | docs/adrs/README.rst                        | ignored, not excluded |
```

### REQ-SPEC-002: Document Index

**Acceptance Criteria:**

```gherkin
@frd-spec-validation-cli
Feature: Document Index

  As an architecture owner
  I want one machine-readable artifact describing every specification document
  So that CI, agents and a central catalogue read one source of truth

  The index is the primary artifact, not a by-product: every check in this
  document is a query over it, and consumers outside this framework read it
  without running any check at all. A repository with three broken documents
  must still contribute its other nine hundred to an organization's graph.

  The schema is deliberately additive. This increment fills frontmatter and
  path; later features populate the rest without changing its shape or its
  version.

  @AC-SPEC-002.1
  Scenario: The index carries frontmatter and path for every member
    When the CLI builds an index
    Then the CLI emits a JSON document
    And that document contains, for every corpus member, its frontmatter properties and its file path

  @AC-SPEC-002.2
  Scenario Outline: The index declares a schema version identifier
    When the CLI builds an index
    Then the emitted JSON declares "<property>": "<value>"

    Examples:
      | property       | value |
      | schema_version | 1.0   |

  @AC-SPEC-002.3
  Scenario Outline: Every entry carries the configured workspace identifier
    Given configuration declaring the workspace "<workspace>"
    When the CLI builds an index
    Then every entry carries "workspace": "<workspace>"

    Examples:
      | workspace     |
      | agent-plugins |

  Rule: A build failure is the checker's job, not the indexer's

    @AC-SPEC-002.4
    Scenario: An unparseable document does not abort the index
      Given a document in the corpus that cannot be parsed
      When the CLI builds an index
      Then the CLI indexes every other document
      And the CLI records the failure against the offending document
      And the CLI does not abort the index

    @AC-SPEC-002.5
    Scenario: Building an index containing parse failures exits zero
      Given an index that contains parse failures
      When the CLI finishes building it
      Then the CLI exits zero
```

### REQ-SPEC-003: Metadata Conformance

**Acceptance Criteria:**

```gherkin
@frd-spec-validation-cli
Feature: Metadata Conformance

  As a reviewer
  I want frontmatter errors reported mechanically
  So that review attention goes to whether a document says the right thing

  Every criterion here except AC-SPEC-003.6 presupposes parsed frontmatter.
  A file with no frontmatter at all is an exclusion under AC-SPEC-001.3 and
  is not covered here: it never claimed to be a specification, whereas a
  member with a parse error claimed and got the claim wrong.

  @AC-SPEC-003.1
  Scenario Outline: Frontmatter is validated against the profile's schema
    Given a document declaring "type: <type>"
    And the active profile registers a JSON Schema for that type
    When the CLI checks the document
    Then the CLI validates its frontmatter against that schema
    And omitting "<property>" produces one finding naming it

    Examples:
      | type | property    |
      | adr  | description |

  @AC-SPEC-003.2
  Scenario Outline: Textual formats are validated against the raw text
    Given a frontmatter property "<property>: <raw>"
    And the schema constrains its textual format
    When the CLI validates that property
    Then the CLI validates the raw text as written in the file
    And "<raw>" fails the "<format>" check

    Examples:
      | property   | raw        | format         |
      | updated_at | 2026-07-29 | RFC 3339 date-time |

  @AC-SPEC-003.3
  Scenario Outline: A duplicate slug is reported against each holder
    Given two corpus members both declaring "slug: <slug>"
    When the CLI checks the corpus
    Then the CLI reports a finding against each
    And each finding names the other's path

    Examples:
      | slug    |
      | adr-012 |

  @AC-SPEC-003.4
  Scenario Outline: Slug uniqueness ignores letter case
    Given the corpus members "<first>" and "<second>"
    When the CLI compares slugs for uniqueness
    Then the CLI reports them as duplicates

    Examples:
      | first        | second       |
      | frd-checkout | FRD-Checkout |

  @AC-SPEC-003.5
  Scenario Outline: A slug pattern mismatch names actual and expected
    Given a document of type "<type>" declaring "slug: <slug>"
    And the active profile registers the pattern "<pattern>" for that type
    When the CLI checks the document
    Then the CLI reports a finding naming the actual slug "<slug>"
    And the finding names the expected pattern "<pattern>"

    Examples:
      | type | slug                                | pattern        |
      | adr  | adr-011-targeted-verification-gates | ^adr-[0-9]{3}$ |

  @AC-SPEC-003.6 @negative
  Scenario: A parse error is reported and never counted as a pass
    Given a corpus member indexed with a parse error rather than parsed frontmatter
    When the CLI checks the corpus
    Then the CLI reports a finding carrying that error's message and span
    And the CLI does not report the document as having passed any check
    And the check run exits non-zero
```

*Note on numbering:* REQ-SPEC-004 (reference integrity) and REQ-SPEC-005
(structure and hygiene conformance) covered behavior deferred to
follow-on features. Their numbers are intentionally left unused here
rather than reassigned: the retained requirements REQ-SPEC-006, -008,
and -009 are cited by ADRs and keep their numbers, so renumbering around
the gap would break those citations. The gap is deliberate.

### REQ-SPEC-006: Findings, Severities, and Diagnostics

**Acceptance Criteria:**

```gherkin
@frd-spec-validation-cli
Feature: Findings, Severities, and Diagnostics

  As an agent acting on the output
  I want findings that name the rule and quote what violated it
  So that I can fix the problem without loading the standards that define it

  A diagnostic is self-contained. Because it quotes the violated constraint
  and the offending value, the fixing agent needs no further lookup and the
  `spec` skill need not preload verification procedure.

  @AC-SPEC-006.1
  Scenario: Every finding is self-contained
    When the CLI reports a finding
    Then the finding carries a stable rule identifier
    And the finding carries a severity, a file path and a span
    And the message states the violated constraint and the value that violated it

  @AC-SPEC-006.3
  Scenario Outline: A check that cannot run is reported as unverified
    Given the CLI is asked to perform the check "<check>"
    And it cannot run because "<reason>"
    When the check run completes
    Then the CLI reports that check with severity "unverified"
    And the finding names the missing dependency

    Examples:
      | check            | reason                          |
      | mermaid-diagrams | no diagram validator is available |

  @AC-SPEC-006.4 @negative
  Scenario: A check that could not run is never reported as passed
    Given a check the CLI could not run
    When the CLI reports the run
    Then the CLI does not report that check as passed

  @AC-SPEC-006.5
  Scenario: The --json option writes the full finding set to stdout
    When the check subcommand is invoked with the "--json" option
    Then the CLI writes the full finding set as JSON to standard output

  @AC-SPEC-006.6
  Scenario: An error-severity finding fails the run
    Given a check run producing at least one finding of severity "error"
    When the run completes
    Then the CLI exits with a non-zero status

  @AC-SPEC-006.7
  Scenario Outline: Warnings pass unless --strict is given
    Given a check run producing findings only of severity "<severity>"
    When the run completes with the option "<option>"
    Then the CLI exits "<status>"

    Examples:
      | severity   | option   | status   |
      | warning    |          | zero     |
      | unverified |          | zero     |
      | warning    | --strict | non-zero |

  @AC-SPEC-006.8
  Scenario Outline: --only runs the named subset of checks
    When the check subcommand is invoked with "--only <check>"
    Then the CLI runs only the named check
    And the report covers that check alone

    Examples:
      | check           |
      | slug-uniqueness |

  @AC-SPEC-006.9 @negative
  Scenario: A check that was not requested is never reported as passed
    Given a run invoked with "--only slug-uniqueness"
    When the CLI reports the run
    Then the CLI does not report the unrequested checks as passed
    And the report distinguishes "slug uniqueness passed" from "frontmatter schema was not checked"
```

### REQ-SPEC-007: Corpus Queries

**Acceptance Criteria:**

```gherkin
@frd-spec-validation-cli
Feature: Corpus Queries

  As an authoring agent
  I want to ask the corpus direct questions
  So that I can pick a conforming identifier without reading every document

  Both queries answer from frontmatter and path alone. Queries that need
  more of the index — which documents reference a slug, the spans of
  headings matching a selection — are follow-on features.

  @AC-SPEC-007.1
  Scenario Outline: next-id returns one greater than the highest in use
    Given the active profile registers "<type>" as numbered
    And the corpus holds that type numbered up to "<highest>"
    When the next-id subcommand is invoked for that type
    Then the CLI returns "<next>"

    Examples:
      | type | highest | next |
      | adr  | 013     | 014  |

  @AC-SPEC-007.2
  Scenario Outline: slug-exists reports use and names the holder
    When the slug-exists subcommand is invoked with "<slug>"
    Then the CLI reports whether a corpus member declares it
    And where one does, the CLI reports "<path>"

    Examples:
      | slug    | path                                                |
      | adr-004 | docs/adrs/ADR-004-module-blueprint-and-atomic.md    |
```

### REQ-SPEC-008: Configuration

**Acceptance Criteria:**

```gherkin
@frd-spec-validation-cli
Feature: Configuration

  As an adopting team
  I want to declare my own document types and conventions
  So that the tool enforces my framework rather than the one it shipped with

  The configuration schema is additive: later features add each type's
  required headings, its line-width limit and its reference-bearing
  properties to the same file without breaking an existing one.

  @AC-SPEC-008.1
  Scenario Outline: Configuration is read from a TOML file at the root
    When the CLI reads its configuration
    Then the CLI reads "<file>" at the repository root

    Examples:
      | file       |
      | specs.toml |

  @AC-SPEC-008.2
  Scenario: Configuration determines root, workspace and registered types
    When the CLI reads its configuration
    Then the configuration determines the spec root
    And the configuration determines the workspace identifier
    And the configuration determines the registered document types with their slug patterns and the JSON Schema that validates each type's frontmatter

  @AC-SPEC-008.3
  Scenario: A missing configuration file is reported, not assumed away
    Given no configuration file is present
    When the CLI runs
    Then the CLI reports the absence, naming the file and the path
    And the CLI exits non-zero rather than assuming a layout

  @AC-SPEC-008.4
  Scenario: An invalid configuration reports the offending key
    Given the configuration is present but invalid
    When the CLI reads it
    Then the CLI reports the offending key and its position

  @AC-SPEC-008.5 @negative
  Scenario: An invalid configuration never falls back to a default
    Given the configuration is present but invalid
    When the CLI reads it
    Then the CLI does not fall back to a default value for the offending key

  @AC-SPEC-008.6
  Scenario: A configuration for this framework's own types ships with it
    When an adopting repository installs the CLI
    Then a configuration for this framework's document types is available
    And the repository can use it without authoring one
```

### REQ-SPEC-009: Invocation, Performance, and Portability

**Acceptance Criteria:**

```gherkin
@frd-spec-validation-cli
Feature: Invocation, Performance, and Portability

  As a platform engineer
  I want to run this in locked-down CI and in an editing loop on every write
  So that neither the build nor the author waits on it

  The budgets come from the two invocation patterns that dominate: an agent
  asking one question, and a write-time hook checking one file.

  Installation must succeed inside a network that blocks outbound traffic.
  That is why nothing executable is retrieved at install time — the common
  download-a-binary pattern fails in exactly the networks that most need
  this tool, and adds a supply-chain surface a security review will reject.

  Scenario Outline: Latency budgets against a 1,000-document corpus
    Given a corpus of 1,000 documents
    And the index is "<index state>"
    When the CLI performs "<operation>"
    Then the CLI completes within "<budget>" at the 95th percentile

    @AC-SPEC-009.1
    Examples: an agent asking one question
      | operation      | index state | budget          |
      | a corpus query | warm        | 50 milliseconds |

    @AC-SPEC-009.2
    Examples: a write-time hook checking one file
      | operation               | index state | budget           |
      | a single-document check | warm        | 100 milliseconds |

    @AC-SPEC-009.3
    Examples: a cold build
      | operation          | index state | budget   |
      | building the index | no cache    | 1 second |

  @AC-SPEC-009.4 @manual
  Scenario: The CLI runs with no language runtime installed
    Given a machine with no language runtime, interpreter, or package manager installed beyond the CLI itself
    When the CLI runs
    Then the CLI runs successfully

  @AC-SPEC-009.5 @manual @negative
  Scenario: Installation retrieves no executable content from a network
    When the CLI is installed
    Then the CLI does not retrieve executable content from a network location during installation

  @AC-SPEC-009.6 @manual
  Scenario: The CLI is installable from a mirrorable registry
    When an adopting organization installs the CLI
    Then the CLI is installable from at least one package registry that supports on-premises mirroring

  @AC-SPEC-009.7
  Scenario Outline: Identical content yields identical findings everywhere
    Given identical corpus content on "<platform>"
    When the CLI runs against it
    Then the CLI reports the same findings in the same order
    And the emitted JSON is byte-identical across platforms

    Examples:
      | platform |
      | Linux    |
      | macOS    |
      | Windows  |

  @AC-SPEC-009.8
  Scenario: Paths are emitted forward-slashed and repository-relative
    When the CLI emits a file path in the index or in a finding
    Then the CLI uses forward slashes as separators
    And the path is relative to the repository root
```

*Note:* AC-SPEC-009.1, .2 and .3 share one `Scenario Outline` above,
each carrying its identifier on its own tagged `Examples` block. Tagging
the scenario instead would have collapsed three criteria into one — see
the comparison document.

### REQ-SPEC-010: Index Output and Persistence

**Acceptance Criteria:**

```gherkin
@frd-spec-validation-cli
Feature: Index Output and Persistence

  As a CI pipeline and as an authoring agent
  I want the index written to one known file that later invocations reuse
  So that what CI publishes is what answers my next query

  The index artifact and the check report are different outputs.
  AC-SPEC-006.5 writes a check run's findings to standard output; this
  requirement governs where the index itself is persisted. One file serves
  both of the index's roles — the deliverable CI publishes, and the warm
  cache behind REQ-SPEC-009's budgets.

  Whether an adopting repository commits that file or ignores it is the
  adopter's choice; the CLI requires neither.

  @AC-SPEC-010.1
  Scenario Outline: The index is written to one known file by default
    Given configuration declaring root "<root>"
    When the CLI is asked to build the index with no output option
    Then the CLI writes the JSON to "<path>"
    And the CLI prints that relative path

    Examples:
      | root       | path                  |
      | docs/specs | docs/specs/specs.json |

  @AC-SPEC-010.2
  Scenario Outline: An output-path option redirects the index
    When the CLI is invoked with "--out <location>"
    Then the CLI writes the index to "<location>" instead of the default

    Examples:
      | location              |
      | build/spec-index.json |

  @AC-SPEC-010.3 @negative
  Scenario: An output path of "-" streams to stdout and writes no file
    When the CLI is invoked with "--out -"
    Then the CLI writes the index to standard output
    And the CLI does not create or modify the default file

  @AC-SPEC-010.4
  Scenario: The index file is replaced atomically
    When the CLI writes the index to a file
    Then the CLI replaces any existing file atomically
    And a concurrent reader observes either the old complete index or the new one, never a partially written file

  @AC-SPEC-010.5
  Scenario: A warm index answers without re-walking the corpus
    Given the index file exists with inputs unchanged since it was written
    When a corpus query or single-document check runs
    Then the CLI answers from that persisted index
    And the CLI does not re-walk the corpus

  @AC-SPEC-010.6
  Scenario Outline: A stale or absent index is rebuilt before answering
    Given the index file is "<state>"
    When a query runs
    Then the CLI rebuilds the index before answering
    And the CLI does not serve a stale result

    Examples:
      | state                              |
      | absent                             |
      | present but an input has changed   |
```

### REQ-SPEC-011: Command Surface

**Acceptance Criteria:**

```gherkin
@frd-spec-validation-cli
Feature: Command Surface

  As a developer integrating the tool
  I want a small, named set of subcommands
  So that skills, hooks and CI scripts call a stable interface

  The subcommand vocabulary is this tool's public interface. Later features
  extend `check` with additional checks and add their own query subcommands,
  but they never rename or repurpose the four established here.

  The name of the executable itself is a distribution decision recorded
  separately; this requirement fixes only the subcommands beneath it.

  @AC-SPEC-011.1
  Scenario Outline: The index subcommand builds and writes the index
    When the "index" subcommand is invoked with "<option>"
    Then the CLI builds the index and writes it per REQ-SPEC-010
    And the destination is "<destination>"

    Examples:
      | option  | destination     |
      |         | specs.json      |
      | --out - | standard output |

  @AC-SPEC-011.2
  Scenario Outline: The check subcommand runs conformance checks
    When the "check" subcommand is invoked with "<target>"
    Then the CLI runs conformance checks over "<scope>"
    And the CLI reports findings per REQ-SPEC-006
    And the CLI accepts the --json, --strict and --only options

    Examples:
      | target                      | scope           |
      |                             | the whole corpus |
      | docs/adrs/ADR-004-module.md | one document     |

  @AC-SPEC-011.3
  Scenario: The slug-exists subcommand answers the AC-SPEC-007.2 query
    When the "slug-exists" subcommand is invoked
    Then the CLI answers the query of AC-SPEC-007.2

  @AC-SPEC-011.4
  Scenario: The next-id subcommand answers the AC-SPEC-007.1 query
    When the "next-id" subcommand is invoked
    Then the CLI answers the query of AC-SPEC-007.1

  @AC-SPEC-011.5 @negative
  Scenario Outline: A bare or --help invocation lists the subcommands
    When the CLI is invoked with "<invocation>"
    Then the CLI lists the available subcommands and their options
    And the CLI does not perform a default action

    Examples:
      | invocation      |
      | no subcommand   |
      | --help          |
```

## References

- [[adr-018]] — the decision this variant implements
- `specs/features/spec-index/spec-index-frd.md` — the original this is
  compared against
- `docs/gherkin-frd-comparison.md` — what the exercise showed
