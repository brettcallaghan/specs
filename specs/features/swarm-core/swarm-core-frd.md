---
title: "FRD: Swarm CLI — the command contract, the manifest, and status"
description: The foundation of the swarm CLI — how every command explains
  itself, the run manifest every other command reads through, and the first
  command that reads one
slug: frd-swarm-core
type: frd
status: draft
updated_at: 2026-08-28T00:00:00Z
sources: []
---

# Swarm CLI — the command contract, the manifest, and status

## Overview

A `bead-bonanza` run is a distributed system whose control plane is
documentation. Its invariants are prose, its state transitions are prose,
and its guards are paragraphs an agent has to remember at the moment it
matters. Measured across the ten sprint-family skills: 79 distinct
capabilities, of which **73 carry a recorded trap** — a way the obvious
invocation produces the wrong result.

This document is the **smallest shippable piece of the CLI**: three
requirements that everything else in the family is built on top of, and
that are worth having on their own.

- **The command contract** — every subcommand prints the invocations
  behind its answer, and can print them without running them. This is not
  a convenience. It is what makes a wrapper usable by an agent that did
  not write it and cannot see inside it.
- **The manifest** — the run's only durable record, and the single
  component every other command reads a run through. Nothing else can be
  correct if this is not.
- **Run status** — the first consumer of that component, and the largest
  derived answer in the family. It is also the one command in the CLI that
  is useful before any other exists: pointed at a run in flight today, it
  answers where the run is.

**Order is build order.** The contract governs how the other two behave,
the manifest is what status reads, and status is what proves the manifest
component works against a real run rather than against a fixture.

### Why the manifest comes first, and not with the rest of the run's record

Two reasons, and the second is the one that decides it.

**Every command in the family reads the manifest through one component.**
`dispatch open` reads and writes the branch, `queue front` and `run
status` read the admission record, review reads the pass counter, recovery
reads all of it. That is one reader, and its behaviour — what an absent
key returns, what a write preserves, what a wrong type does — is what this
document specifies. A later document cannot define it, because by then several
commands already depend on it.

**Those readers need the data to be right, and only the write path can
make it so.** A reader that faithfully returns a branch recorded twice, a
review counter that went backwards, or a ledger whose entries were
reordered under citations addressing them by position is a correct reader
of corrupt data. The validation is not garnish on top of the record; it is
what lets everything downstream treat the record as trustworthy. So the
write path ships with the read path, here, rather than being
added once things already depend on it.

### The rest of the CLI

| document | what it covers |
|---|---|
| [[frd-swarm-dispatch]] | the dispatch loop — closing finished phases, the dispatch front, opening a dispatch, accepting one back, and proving a push landed |
| [[frd-swarm-cli]] | the run's structure and record — building a run's shape, partitioning and gating the queue, readiness, claims, findings and recovery |

Both depend on this document and neither is depended on by it.

### What this document does not do

**Orchestration.** The CLI spawns no agent, arms no monitor, schedules
nothing, and holds no state of its own. It is called from inside an
existing harness and returns.

**Judgment.** Whether a run should stop, whether a finding is in scope,
whether partial work is finishable — all stay with an agent. Status
reports; it does not decide.

**Building the run.** Nothing here creates a run epic, a phase chain or a
target. `run status` reads a run that already exists, which today means
one the existing skills built by hand.

### The shape of a command

**`swarm <noun> <verb>`.** The nouns are the things a run has — `run`,
`phase`, `queue`, `dispatch`, `branch` — and the verb is the action. This
document defines the conventions; every command in the family carries
them.

**The vocabulary is tracker-neutral on purpose.** A unit of work is a
**work item**, not the name one tracker gives it. The tracker in use today
is the only one implemented, and every command goes through one interface
so a second is possible later; nothing in these requirements names a
tracker's own vocabulary except where a measured behaviour is being pinned.

| | |
|---|---|
| `--run <run-epic-id>` | required by every subcommand that reads or mutates run state. An absent or empty scope is refused, never treated as the whole repository |
| `--explain` | print the resolved tracker and version-control invocations in order and run none of them (FEAT-SWARM-CORE-001) |
| `--json` | machine-readable output; the same facts as the text form, never a subset |
| `--dry-run` | on mutating subcommands: perform every read and no write |
| exit `0` | the command's own verdict is clear |
| exit `1` | repairable in place, and the output names the repair |
| exit `2` | stop; not repairable by the caller |
| exit `3` | invocation error — a missing scope, an unknown flag, a malformed argument |

The exit vocabulary is the one the family's existing status script already
uses, unchanged, because every skill in the family already routes on it.

## Terminology

| Term | Meaning |
|---|---|
| **run** | one execution of `bead-bonanza` over one run epic |
| **run epic** | the work item whose id is the run id and whose metadata is the run manifest |
| **work item** | one unit of work in the tracker. The tracker-neutral name for what this project's tracker calls a bead and others call an issue, story or ticket |
| **manifest** | the JSON object in the run epic's metadata; the run's only durable record |
| **phase epic** | one of the epics chained beneath the run epic, in order: pre-flight, scope, plan, target, review, retro, post-flight |
| **target epic** | the flat phase epic whose task children are the run's work |
| **dispatch** | one agent working one queue in one worktree on one branch. A dispatch carries one or many work items |
| **dispatch id** | one agent's whole identity — its assignee, its tracker actor, and its branch name. Derived from the run epic id, never minted |
| **ready front** | the complete set of work items in the run that nothing is blocking right now |
| **drained** | an agent holding nothing ready and nothing in progress. Two queries, because the ready query alone excludes in-progress work |
| **derived answer** | an answer requiring a tree walk or a join across the tracker and version control that no single command returns |
| **silent failure** | a command that returns the wrong answer at exit 0 |

## Requirements

### FEAT-SWARM-CORE-001: Show the commands behind every answer

**Flag, carried by every subcommand of the CLI:**

```text
swarm <noun> <verb> ... --explain
```

**Specified once, carried everywhere.** It is a property of every
command in every document of this family rather than a capability of its
own, so it is stated here and cited elsewhere rather than repeated as a
criterion in each requirement.

An agent that can see the invocation behind a line can re-run it, argue
with it, and work around it when the underlying tool changes. An agent
that cannot has to trust a wrapper it did not write, which is the one
thing this CLI must not require. This is also the direct remedy for the
evidence rule in `AGENTS.md`: a claim that prints the command that
established it does not need a paragraph telling an agent to measure it.

**Replaces:** the "prints the command that established it" paragraphs in
`bead-bonanza/references/scripts.md` and `references/dispatching.md`, by
making the property universal instead of per-script.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-CORE-001
Feature: Show the commands behind every answer

  As an agent that has to troubleshoot a wrapper it did not write
  I want every subcommand to print the tool invocations behind its answer
  So that I can re-run any single step rather than trust the whole result

  Two measured hazards shape how the invocations are run rather than how
  they are printed. Reading an exit status through a pipe returns the
  reader's status, so a failing step reads as a pass; and a reader that
  closes a pipe early makes a successful writer report failure. The
  tracker also writes warnings to its error stream while its data goes to
  its output stream, so a wrapper that merges the two corrupts the answer
  it is about to parse.

  @AC-SWARM-CORE-001.1
  Scenario: An explained invocation runs nothing
    When a subcommand is invoked in explain mode
    Then it prints the tool invocations it would run, in order, with every argument resolved
    And no tool is executed

  @AC-SWARM-CORE-001.2
  Scenario: A normal run prints the invocations as it executes them
    When a subcommand runs normally
    Then each tool invocation is printed as it is executed

  @AC-SWARM-CORE-001.3
  Scenario: Every reported line carries the command that established it
    When a subcommand reports a fact
    Then the line names the command a reader can re-run to reproduce it

  @AC-SWARM-CORE-001.4
  Scenario: A step's exit status is read from the step itself
    Given a step whose command fails while a reader consumes its output
    When the step's result is reported
    Then the step is reported as failed

  @AC-SWARM-CORE-001.5
  Scenario: A step that succeeds while its reader stops early is not reported as failed
    Given a step whose reader stops consuming output before the command finishes
    When the step's result is reported
    Then the step is reported as succeeded

  @AC-SWARM-CORE-001.6 @negative
  Scenario: The output and error streams of a tool are never merged
    Given a tool that writes a warning to its error stream and data to its output stream
    When the subcommand parses the data
    Then the warning is not part of what is parsed

  @AC-SWARM-CORE-001.7
  Scenario: Explain output is a runnable transcript
    Given a subcommand invoked in explain mode
    When its printed invocations are run by hand in order
    Then they perform what the subcommand would have performed
```


### FEAT-SWARM-CORE-002: Read and update the run manifest

**Commands:**

```text
swarm run manifest get     --run <id> [--key <key>] [--json]
swarm run manifest set     --run <id> --key <key> --value <value> [--if-absent] [--json]
swarm run manifest require --run <id> --key <key>...
```

**This is the component every other command reads a run through**, so
its behaviour on an absent key, a wrong type and a concurrent write is
not an implementation detail — it is the contract each of them is written
against.

**The validation is logical as well as structural.** The manifest's key
names are a contract, its types are a contract, and several of its keys
carry rules that no single write can check on its own — a branch written
twice, a review counter that went backwards, a ledger whose entries were
reordered under citations that address them by position. Every such rule
is a criterion below, because the alternative is prose in five skills.

**Two measured constraints shape the implementation**, both re-measured on
the tracker in use (bd 1.2.2) while this document was written:

```text
$ bd update <id> --set-metadata num=5 --set-metadata b=true --set-metadata f=1.5 \
                 --set-metadata n=null --set-metadata arr='[1,2]' \
                 --set-metadata obj='{"k":1}' --set-metadata txt=hello
$ bd show <id> --json | jq -c '.[0].metadata'
{"b":true,"f":1.5,"n":null,"arr":"[1,2]","num":5,"obj":"{\"k\":1}","txt":"hello"}

$ printf '%s' '{"num":9,"obj2":{"a":1}}' > m.json && bd update <id> --metadata @m.json
$ bd show <id> --json | jq -c '.[0].metadata'
{"b":true,"f":1.5,"n":null,"arr":"[1,2]","num":9,"obj":"{\"k\":1}","txt":"hello","obj2":{"a":1}}

$ printf '%s' '{"obj2":{"b":2}}' > m2.json && bd update <id> --metadata @m2.json
$ bd show <id> --json | jq -c '.[0].metadata.obj2'
{"b":2}
```

First: the per-key write parses JSON scalars — a number stays a number,
`true` stays a boolean, `null` stays null — but stores arrays and objects
as strings. So the list-valued and object-valued keys (`admitted`,
`excluded`, `moved`, `checks_ids`, `ledger`, `models`, `standing_auth`)
cannot be written that way at all.

Second: the whole-object write merges at the **top level** and replaces a
**nested** value wholesale — `obj2` went from `{"a":1}` to `{"b":2}` and
lost `a`. So it is the write path for the list-valued keys, and a caller
updating one field inside `standing_auth` or `models` must read that
object, edit it, and write the whole of it back.

**Replaces:** `references/run-record.md`, "The two things you own" item 1
and the whole of "Run manifest"; the manifest-reading preambles in
`run-batch/SKILL.md`, `sprint-review/SKILL.md` and `sprint-plan/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-CORE-002
Feature: Read and update the run manifest

  As any component recording or reading a run's durable state
  I want manifest keys read and written without destroying keys another
  component wrote
  So that the run's only durable record survives being updated

  The manifest is a bespoke convention layered on a generic metadata
  field: the run epic's metadata IS the run record, and several
  components write different keys into the same object at different
  times. A whole-object write from any one of them destroys the others.
  The key names are equally a contract, because dispatch reads them back
  rather than being handed them, and a renamed key reads back as absent
  while the run proceeds without it.

  @AC-SWARM-CORE-002.1
  Scenario: Writing one key preserves the others
    Given a manifest carrying keys written by other components
    When one key is updated
    Then the other keys retain their values

  @AC-SWARM-CORE-002.2 @negative
  Scenario: An unknown key is refused rather than stored
    When a key outside the manifest's declared set is written
    Then the command exits non-zero
    And the manifest is unchanged

  @AC-SWARM-CORE-002.3
  Scenario: A required key that is absent is reported as absent, not empty
    Given a manifest with no checks baseline recorded
    When the baseline is read
    Then the result is reported as absent rather than as an empty baseline

  @AC-SWARM-CORE-002.4
  Scenario: The review pass counter survives a resumed run
    Given a manifest recording spent review passes
    When a resumed run reads the counter
    Then it reports the spent count rather than zero

  @AC-SWARM-CORE-002.5
  Scenario: A list-valued key reads back as a list
    Given a key whose declared type is a list of work item ids
    When it is written and read back
    Then the value is a list of the same ids rather than one string

  @AC-SWARM-CORE-002.6 @negative
  Scenario: The branch and its base are written once
    Given a manifest already recording a branch and a base commit
    When a different branch or base is written
    Then the command exits non-zero
    And the recorded values are unchanged

  @AC-SWARM-CORE-002.7 @negative
  Scenario: The review pass counter never decreases and never exceeds its budget
    Given a manifest recording two spent passes against a budget of two
    When a lower count or a third pass is written
    Then the command exits non-zero

  @AC-SWARM-CORE-002.8 @negative
  Scenario: The environmental ledger is append-only
    Given a ledger whose entries are cited elsewhere by position
    When a write would remove or reorder an entry
    Then the command exits non-zero

  @AC-SWARM-CORE-002.9 @negative
  Scenario: A value of the wrong type for its key is refused
    When a key whose declared type is a number is written with text
    Then the command exits non-zero
    And the manifest is unchanged

  @AC-SWARM-CORE-002.10 @negative
  Scenario: A write that would clobber a concurrent change is refused
    Given a key that changed after this caller read the manifest
    When that caller writes the manifest back
    Then the command exits non-zero naming the key that moved

  @AC-SWARM-CORE-002.11
  Scenario: A setting the harness does not expose is recorded as unrecorded
    Given a run whose harness exposes no effort setting
    When the manifest is written
    Then the effort key is present with the value unrecorded rather than absent

  @AC-SWARM-CORE-002.12 @negative
  Scenario: The admission record's three lists do not overlap
    When a work item appears in more than one of the admitted, excluded and moved lists
    Then the command exits non-zero naming that item
```


### FEAT-SWARM-CORE-003: Report a run's status from state

**Command:**

```text
swarm run status --run <id> [--stall-after <duration>] [--json]
```

**Health is three measurements and a threshold, and the command reports
all four.** There is no way from a shell to observe whether an agent is
still thinking: a stalled agent never finishes, so no event fires and
absence of an event is not an event. The three observable inputs are the
number of standing worktrees, the age of the last commit on the sprint
branch, and the depth of the ready front. The command states each with the
value it had, states the threshold it applied, and names the thing it
could not observe — so a reader who disagrees with the classification can
still act on the facts.

**Per-agent drain is reported here and defined in [[frd-swarm-dispatch]].**
Status says which agents still hold work; the dispatch document's queue
command is where "drained" means something, because that is the command
an agent runs about itself. Status is useful before that command exists —
it reports the agents holding work from the same two queries — and cites
it rather than defining the term twice.

**Replaces:** `bead-bonanza/references/supervision.md`, "Where is this run" and "When
the run is done"; the "The run's health — derived, never asserted" section
of `bead-bonanza/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-CORE-003
Feature: Report a run's status from state

  As a coordinator or a session resuming one
  I want health, completion and per-agent drain answered from version
  control and the tracker
  So that a stalled run is distinguishable from a finished one without
  the coordinator's cooperation

  This is the liveness backbone and the largest token saving in the
  family, because every part of it is a derived answer. A run once read
  healthy through a five-hour outage in which nothing was dispatched,
  because health was a label written by the component that had stopped.
  Completion needs a walk the tracker does not offer: three queries look
  like they answer "every work item in this run" and each answers
  something else.

  @AC-SWARM-CORE-003.1
  Scenario: A stall is reported as the measurements that produced it
    Given no worktrees are standing
    And the ready front is not empty
    And the last commit is older than the stall threshold
    Then the report states the standing worktree count, the timestamp and age of the last commit, and the depth of the ready front
    And it states the threshold it compared the commit age against
    And the run's health is reported as stalled
    And the command exits non-zero

  @AC-SWARM-CORE-003.2
  Scenario: Health is never read from a stored value
    Given a run epic carrying a health value in its metadata that contradicts its state
    When health is reported
    Then the reported value reflects the state, not the stored value

  @AC-SWARM-CORE-003.3
  Scenario: Completion counts every work item in the run, however nested
    Given a run whose work items are not direct children of the run epic
    When completion is reported
    Then every work item in the run is counted

  @AC-SWARM-CORE-003.4
  Scenario: Each agent holding work is named
    Given two dispatches, one of which has drained
    When status is reported
    Then the dispatch still holding work is named
    And the drained one is reported as drained

  @AC-SWARM-CORE-003.5
  Scenario: A tracker that cannot answer is not reported as an empty answer
    Given a tracker query that fails
    When status is reported
    Then the affected line is reported as unknown
    And the command exits non-zero

  @AC-SWARM-CORE-003.6
  Scenario: What the report cannot observe is named
    When health is reported
    Then the report states that whether an agent is still working is not observable
    And it names the three measurements the classification rests on instead

  @AC-SWARM-CORE-003.7
  Scenario: The live queue is compared against the admitted set
    Given work items present in the run that the manifest's admitted list does not name
    When status is reported
    Then those work items are named as additions to the run's scope

  @AC-SWARM-CORE-003.8
  Scenario: Work items excluded at admission are reported beside the completion verdict
    Given a run whose admitted work is all closed and whose admission excluded three items
    When completion is reported
    Then the run is reported as done over a narrowed scope naming those three

  @AC-SWARM-CORE-003.10
  Scenario: A manifest key that is absent is reported as absent, not as an empty answer
    Given a run whose manifest records no admitted list
    When the scope comparison is reported
    Then it is reported as unestablished rather than as no additions
    And the command exits non-zero

  @AC-SWARM-CORE-003.9
  Scenario: A blocked work item whose worktree still stands is reported as unresolved
    Given a work item marked blocked whose dispatch's worktree still exists
    When status is reported
    Then that work item is reported as an unresolved conflict holding its evidence
```


## References

- [[frd-swarm-dispatch]] — the dispatch loop, built on this document
- [[frd-swarm-cli]] — the run's structure and record, built on this document
- `docs/research/swarm-cli/2026-08-25-capability-audit.md` — the audit
  these requirements are drawn from; §7 is the classification that decides
  what earns a command and §8 is the stack rank
- [[module-bead-bonanza]] — the module whose components call this CLI
- [[gherkin-guide]] — the notation these criteria are written in
