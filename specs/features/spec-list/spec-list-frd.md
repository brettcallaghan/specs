---
title: "Corpus Listing"
description: A subcommand that summarizes every corpus member from its
  frontmatter, so an agent can decide which specifications to read without
  opening them or the index
slug: frd-spec-list
type: frd
status: approved
updated_at: 2026-09-07T00:00:00Z
sources: []
---

# Corpus Listing

## Overview

An agent arriving at a specification corpus has to decide which documents
to read. Today it has two routes and both are bad. It can list paths,
which costs almost nothing and tells it almost nothing — a filename is a
guess at a document's subject, and a corpus whose filenames happen to be
descriptive is a corpus that got lucky. Or it can read the documents,
which is authoritative and unaffordable: this repository's 57 members are
888,869 bytes, and the decision the agent is trying to make is *which of
them to spend context on*.

The index does not close the gap. `specs.json` carries every span,
heading and reference in the corpus, and at 936,833 bytes it is larger
than the corpus it describes — so the artifact built for machine
consumption is the most expensive way to ask what a corpus contains.

This feature adds the missing surface. Every member already declares a
`description` — a one- to two-sentence summary, required by
[[frontmatter-standards]] on every document type — and nothing reads it
back. `specs list` returns that summary, with the fields needed to act on
it, for every member at once. The measured payload for this corpus is
about 21 KB against 889 KB of documents, but the size is not the point:
the point is that one invocation returns a complete, typed, filterable
answer that no combination of shell tools returns reliably, and that the
agent's next step — opening the right document — is a path it was handed
rather than one it inferred.

The feature also gives `description` a consumer. A property that is
required, written once and never read back is a property that decays;
making it the payload of a command is what turns it into an interface
whose quality an author has a reason to care about.

## Terminology

- **Corpus member** — a markdown file beneath the spec root whose
  frontmatter parses and declares a type the active profile defines. The
  index records one entry per member.
- **Summary** — the fields `list` reports for one member: its path, slug,
  type, status, title and description.
- **Undescribable member** — a member the index recorded with a parse
  error rather than parsed frontmatter. It is a member, so it is
  reported, but it has no summary to report.
- **Registered type** — a document type the active profile defines.
  `--type` accepts these and nothing else.

## Requirements

### FEAT-SPECLIST-001: Corpus Summary

**Acceptance Criteria:**

```gherkin
@FEAT-SPECLIST-001
Feature: Corpus Summary

  As an agent orienting in an unfamiliar specification corpus
  I want one command that returns what every document is about
  So that I can choose which documents to read instead of reading all of them

  The summary is drawn entirely from frontmatter the index already holds,
  so this requirement adds no parsing and no new artifact. Its cost is
  proportional to the number of members, never to the length of the
  documents themselves.

  Nothing bounds a description's length: the schema constrains it to a
  string and the frontmatter standards ask for one to two sentences in
  prose, which no check enforces. So no criterion here names a maximum,
  and an implementation must not assume one.

  Reporting the path alongside the slug is deliberate: the caller's next
  action is opening a document, and a summary that forced a second
  command to resolve a slug to a path would have answered half a question.

  @AC-SPECLIST-001.1
  Scenario: Every corpus member is summarized
    Given a corpus whose members all declare parsable frontmatter
    When the caller runs the list command with no filters
    Then the answer carries one entry per member at a status other than retired

  @AC-SPECLIST-001.2
  Scenario: An entry carries the fields needed to act on it
    When the list command reports a member
    Then that member's entry carries its path, slug, type, status, title and description, in that order

  @AC-SPECLIST-001.3
  Scenario: A description is reported in full whatever its length
    Given a member whose description is the longest in the corpus
    When the list command reports that member
    Then the entry carries that description exactly as authored, neither truncated nor elided

  @AC-SPECLIST-001.4
  Scenario: A path is reported relative to the repository root
    When the list command reports a member
    Then that member's path is written relative to the repository root with forward slashes on every platform

  @AC-SPECLIST-001.5
  Scenario: An undescribable member is reported rather than omitted
    Given a member the index recorded with a parse error
    When the caller runs the list command with no filters
    Then the answer carries an entry naming that member's path and the parse failure, with no slug, type, status, title or description

  @AC-SPECLIST-001.6 @negative
  Scenario: A file that is not a corpus member is not reported
    Given a markdown file beneath the spec root that the index recorded as an exclusion
    When the caller runs the list command
    Then the answer carries no entry for that file

  @AC-SPECLIST-001.7
  Scenario: An empty corpus answers with an empty collection
    Given a corpus with no members
    When the caller runs the list command
    Then the answer carries an empty collection and the command exits zero
```

The exclusion in AC-SPECLIST-001.6 is a scope boundary rather than a
silence. A file beneath the spec root that declares no frontmatter, or
declares a type the profile does not define, is not a member and has no
summary to give; reporting it here would put a conformance judgement in a
query. `specs check` owns that judgement and reports it as a finding.

An undescribable member is different, and AC-SPECLIST-001.5 treats it
differently for a reason: the index recorded it *as a member*, so
omitting it would make the answer disagree with the corpus. An agent
reading a list that silently skipped it would conclude the document does
not exist.

### FEAT-SPECLIST-002: Filtering by Type

**Acceptance Criteria:**

```gherkin
@FEAT-SPECLIST-002
Feature: Filtering by Type

  As an agent that needs one category of document
  I want to narrow the listing to the types I care about
  So that the answer stays proportionate to the question as the corpus grows

  The filter accepts type values rather than a fixed set of per-type
  flags, so a profile that registers a type this feature never heard of
  is filterable on the day it registers it. This matches the repeatable
  or comma-separated form the check command's own narrowing flag
  established.

  Filtering is what keeps the command affordable. The answer's size grows
  with the corpus, so a large corpus is served by asking a narrower
  question rather than by the command shortening its reply.

  @AC-SPECLIST-002.1
  Scenario: A single type narrows the answer to that type
    Given a corpus containing members of several types
    When the caller filters by one registered type
    Then the answer carries every member of that type and no other member

  @AC-SPECLIST-002.2
  Scenario Outline: Several types are named in either accepted form
    Given a corpus containing members of several types
    When the caller names three registered types "<form>"
    Then the answer carries every member of those three types and no other member

    Examples:
      | form                                                        |
      | as three separate occurrences of the type filter            |
      | as one occurrence whose value is the three names comma-separated |

  @AC-SPECLIST-002.3
  Scenario: A type matching no member answers with an empty collection
    Given a corpus containing no member of a registered type
    When the caller filters by that type
    Then the answer carries an empty collection and the command exits zero

  @AC-SPECLIST-002.4 @negative
  Scenario: A type the profile does not define is rejected
    When the caller filters by a type the active profile does not register
    Then the command reports the unregistered value and the names of every registered type, and exits non-zero

  @AC-SPECLIST-002.5 @negative
  Scenario: An undescribable member is not reported under any type filter
    Given a member the index recorded with a parse error
    When the caller filters by any registered type
    Then the answer carries no entry for that member
```

AC-SPECLIST-002.4 is the one place this feature departs from
[[cli-output-standards]]' rule that a query matching nothing is an
answer. That rule is about the *world*: a corpus with no contracts in it
is a fact, and reporting it as a failure would make a caller branch on
the shape of the corpus. A type the profile does not define is not a fact
about the corpus — it is a malformed question, and answering it with an
empty collection would render a typo indistinguishable from an empty
category. The active profile already reports an unregistered
`[profile].name` by naming the registered alternatives rather than by
failing the lookup bare; this is that behaviour, one level down.

AC-SPECLIST-002.5 follows from what a filter means rather than from a
policy about parse failures. An undescribable member declares no type, so
no type names it. It is visible in the unfiltered listing, which is where
a caller looking for holes in the corpus will find it.

### FEAT-SPECLIST-003: Retired Documents

**Acceptance Criteria:**

```gherkin
@FEAT-SPECLIST-003
Feature: Retired Documents

  As an agent choosing which specifications to read
  I want retired documents left out unless I ask for them
  So that I do not pull a superseded specification into context and act on it

  A retired document is not in effect. It stays in the corpus so that
  references to it still resolve and so that its successor can be named,
  but an agent orienting over the corpus that reads one has been actively
  misled rather than merely charged for tokens.

  Retired is the only status treated this way. Draft, ready-for-review,
  approved, in-progress and active documents are all in flight or in
  force, and an agent is entitled to see them and to read the status
  reported beside each.

  @AC-SPECLIST-003.1
  Scenario: Retired documents are left out by default
    Given a corpus containing members at status retired and members at other statuses
    When the caller runs the list command without asking for retired documents
    Then the answer carries every member at a status other than retired and no retired member

  @AC-SPECLIST-003.2
  Scenario: Retired documents are included on request
    Given a corpus containing members at status retired
    When the caller runs the list command asking for documents at every status
    Then the answer carries every member, retired members included

  @AC-SPECLIST-003.3
  Scenario: Every other status is reported by default
    Given a corpus containing members at draft, ready-for-review, approved, in-progress and active
    When the caller runs the list command without asking for retired documents
    Then the answer carries every one of those members

  @AC-SPECLIST-003.4
  Scenario: Asking for every status composes with a type filter
    Given a corpus containing retired and active members of one registered type
    When the caller filters by that type and asks for documents at every status
    Then the answer carries both the retired and the active members of that type and no member of another type
```

### FEAT-SPECLIST-004: Rendering

**Acceptance Criteria:**

```gherkin
@FEAT-SPECLIST-004
Feature: Rendering

  As a caller that has to parse what it reads
  I want this command to print what every other command in the surface prints
  So that I need no per-command knowledge to consume its answer

  This requirement adds nothing. It states that the listing is an answer
  in the sense the output standards define, and is therefore rendered by
  the rules that already govern every subcommand rather than by a format
  of its own.

  @AC-SPECLIST-004.1
  Scenario: The answer is rendered as YAML by default
    When the caller runs the list command without asking for JSON
    Then the answer is rendered as YAML

  @AC-SPECLIST-004.2
  Scenario: The answer is rendered as JSON on request
    When the caller runs the list command asking for JSON
    Then the answer is rendered as JSON carrying the same values in the same order as the YAML rendering

  @AC-SPECLIST-004.3
  Scenario: Members are reported in the order the index holds them
    Given a corpus whose members are indexed in a fixed order
    When the caller runs the list command twice against an unchanged corpus
    Then both runs report the members in the index's own order and produce byte-identical output

  @AC-SPECLIST-004.4
  Scenario: A title carrying a colon and a space is quoted
    Given a member whose title contains a colon followed by a space
    When the answer is rendered as YAML
    Then that title is rendered double-quoted so that it reads back as written

  @AC-SPECLIST-004.5 @negative
  Scenario: The answer carries nothing but the members
    When the caller runs the list command
    Then the answer carries no count of the members, no heading naming the command and no summary of what was reported
```

AC-SPECLIST-004.4 is called out rather than left to the general rule
because it is the majority case, not an edge: 38 of this repository's 57
titles contain a colon followed by a space, which is exactly the sequence
that stops a plain scalar reading back as written.

### FEAT-SPECLIST-005: Descriptions as a Consumed Interface

**Acceptance Criteria:**

```gherkin
@FEAT-SPECLIST-005
Feature: Descriptions as a Consumed Interface

  As an author writing a specification
  I want to know that this document's description is what other agents will
  triage it by
  So that I write it to be selected from rather than as an introduction

  A description was previously written once and read by nobody. This
  command makes it the basis on which every other agent decides whether
  to open the document, which is a change in what the property is for and
  therefore a change the authoring guidance has to state.

  This requirement governs guidance, not enforcement. Whether a
  description is written well enough to choose from is a judgement no
  check can make, and a length rule would pass a useless description of
  the right size.

  @AC-SPECLIST-005.1 @manual
  Scenario: The frontmatter standards state the second audience
    Given the standards document governing frontmatter properties
    Then its guidance on the description property states that the listing command reports it and that it is written to be triaged from

  @AC-SPECLIST-005.2 @manual
  Scenario: The CLI reference documents the command
    Given the guide documenting the CLI's command surface
    Then it documents the listing command, its filters, and the fields an entry carries
```

## References

- [[cli-output-standards]] — the rendering rules FEAT-SPECLIST-004
  restates rather than redefines
- [[frontmatter-standards]] — where `description` is required, and the
  guidance FEAT-SPECLIST-005 amends
- [[container-spec-cli]] — the container this subcommand joins
- [[frd-spec-index]] — the index this command reads, and the artifact
  whose size motivates a separate listing surface
- [[spec-cli-reference]] — the guide FEAT-SPECLIST-005 amends
