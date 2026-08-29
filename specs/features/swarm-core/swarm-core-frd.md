---
title: "FRD: Swarm CLI — the command contract, the manifest, and status"
description: The foundation of the swarm CLI — how every command explains
  itself and exits, the run manifest every other command reads through, and
  the first command that reads one
slug: frd-swarm-core
type: frd
status: approved
updated_at: 2026-08-28T00:00:00Z
sources: []
---

# Swarm CLI — the command contract, the manifest, and status

## Overview

A `bead-bonanza` run is a distributed system whose control plane is
documentation. Its invariants are prose, its state transitions are prose,
and its guards are paragraphs an agent has to remember at the moment it
matters. Most of what that prose encodes is not advice but a **trap** — a
way the obvious invocation returns the wrong answer, often at exit 0. The
inventory behind that claim, and the count at the time it was taken, are
in the capability audit this document's References name.

This document is the **smallest shippable piece of the CLI**: three
requirements that everything else is built on top of, and that are worth
having on their own.

- **The command contract** — how every command reports what it ran, and
  what its exit status means. This is not a convenience. It is what makes
  a wrapper usable by an agent that did not write it and cannot see
  inside it.
- **The manifest** — the run's only durable record, and the single
  component every other command reads a run through. Nothing else can be
  correct if this is not.
- **Status** — the first consumer of that component, and the largest
  derivation in the family. It is also the one command that is useful
  before any other exists: pointed at a run in flight today, it answers
  where that run is.

**Order is build order.** The contract governs how the other two behave,
the manifest is what status reads, and status is what proves the manifest
component works against a real run rather than against a fixture.

### Why the manifest is here and not with the rest of the run's record

Two reasons, and the second is the one that decides it.

**Every command that touches a run reads the manifest through one
component.** The command that opens a dispatch reads and writes the
branch; the command that reports where a run is reads the admission
record; review reads the pass counter; recovery reads all of it. That is
one reader, and its behaviour — what an absent key returns, what a write
preserves, what a wrong type does — is what this document specifies. A
document written later cannot define it, because by then several commands
already depend on it.

**Those readers need the data to be right, and only the write path can
make it so.** A reader that faithfully returns a branch recorded twice, a
review counter that went backwards, or a ledger whose entries were
reordered under citations addressing them by position is a correct reader
of corrupt data. The validation is not garnish on top of the record; it is
what lets everything downstream treat the record as trustworthy. So the
write path ships with the read path rather than being added once things
already depend on it.

### What this document does not do

**Orchestration.** The CLI spawns no agent, arms no monitor, schedules
nothing, and holds no state of its own. It is called from inside an
existing harness and returns.

**Judgment.** Whether a run should stop, whether a finding is in scope,
whether partial work is finishable — all stay with an agent. Status
reports; it does not decide.

**Building the run.** Nothing here creates a run epic, a phase chain or a
target. `swarm status` reads a run that already exists, which today means
one the existing skills built by hand.

### The shape of a command

**`swarm <noun> <verb>`, and `run` is not a noun.** Every command in this
CLI concerns a run, so a `run` segment before every verb would carry no
information — `swarm status`, not `swarm run status`. The word survives
only where a run is the *object* of the verb rather than a namespace, as
in a command that creates one.

**The vocabulary is the graph's, not the tracker's.** A unit of work is a
**bead**: one vertex in the dependency graph a run dispatches from. That
is the mental model the whole design rests on — work is a graph whose
edges the tracker enforces, not a queue somebody walks — and the word is
chosen to keep it in view. Another tracker's issue, story or ticket maps
onto a bead; every command goes through one interface so a second tracker
is possible later.

**The conventions below are this requirement's subject**, specified by
FEAT-SWARM-CORE-001 and carried by every command in the CLI:

| | |
|---|---|
| `--run <run-epic-id>` | required by every subcommand that reads or mutates run state. An absent or empty scope is refused, never treated as the whole repository |
| `--explain` | print the invocations the command would run, in order, and execute none of them. It reads nothing, so an argument only a prior step could supply is printed as that step rather than as a value |
| `--dry-run` | perform every read and no write. Unlike `--explain` it runs, so its report reflects the run's live state |
| `--json` | machine-readable output carrying the same facts as the text form, never a subset |
| exit `0` | the command answered, and its verdict is clear |
| exit `1` | the command answered, and what it found needs attention; the output names what would resolve it |
| exit `2` | stop — not the caller's to repair in place |
| exit `3` | the invocation itself is malformed: a missing scope, an unknown flag, a bad argument |

**`--explain` and `--dry-run` answer different questions**, which is why
both exist. `--explain` answers *what would you run?* and touches
nothing, so it is safe against a repository you have not looked at and
costs one process. `--dry-run` answers *what would change?*, which cannot
be known without reading the current state, so it performs every read the
real invocation would.

The exit vocabulary is the one the family's existing status script
already uses, unchanged, because every skill in the family already routes
on it.

## Terminology

| Term | Meaning |
|---|---|
| **run** | one execution of `bead-bonanza` over one run epic |
| **run epic** | the bead whose id is the run id and whose metadata is the run manifest |
| **bead** | one unit of work, and one vertex in the graph a run dispatches from. The graph is the model; another tracker's issue, story or ticket maps onto a bead |
| **manifest** | the JSON object in the run epic's metadata; the run's only durable record |
| **phase epic** | one of the epics chained beneath the run epic, in order: pre-flight, scope, plan, target, review, retro, post-flight |
| **target epic** | the flat phase epic whose task children are the run's work |
| **dispatch** | the act of handing a batch of beads to a subagent, which works them in its own worktree on its own branch and hands the branch back |
| **dispatch id** | one string that is at once the subagent's identity, the assignee on every bead in its batch, its tracker actor, and its branch name. Derived from the run epic id, never minted |
| **ready front** | the complete set of beads in the run that nothing is blocking right now |
| **drained** | a subagent holding nothing ready and nothing in progress. Two queries, because the ready query alone excludes in-progress work |
| **derivation** | an answer requiring a tree walk or a join across the tracker and version control that no single command returns |
| **silent failure** | a command that returns the wrong answer at exit 0 |

## Requirements

### FEAT-SWARM-CORE-001: The contract every command keeps

**Carried by every subcommand of the CLI:**

```text
swarm <noun> <verb> ... [--explain] [--dry-run] [--json]
```

**Specified once, carried everywhere.** These are properties of every
command rather than a capability of their own, so they are stated here
and cited elsewhere rather than repeated as criteria in each requirement.

**Why the exit status is a contract and not a detail.** Every skill in
the family routes on it — a coordinator reads it to decide whether to
repair, stop, or carry on, and it does so without reading the output. A
command that exits 1 where it meant 2 sends a run onward into a state it
should have stopped in, and nothing downstream can tell.

**Why the invocations are printed.** An agent that can see the command
behind a line can re-run it, argue with it, and work around it when the
underlying tool changes. An agent that cannot has to trust a wrapper it
did not write, which is the one thing this CLI must not require. This is
also the direct remedy for the evidence rule in `AGENTS.md`: a claim that
prints the command that established it does not need a paragraph telling
an agent to measure it.

**Replaces:** the "prints the command that established it" paragraphs in
`bead-bonanza/references/scripts.md` and `references/dispatching.md`, by
making the property universal instead of per-script; and the exit-code
tables restated in `references/dispatching.md` and `references/scripts.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-CORE-001
Feature: The contract every command keeps

  As an agent that has to route on a wrapper it did not write
  I want every command to report what it ran, what it would run, and what
  its exit status means
  So that I can re-run any single step and branch on the result without
  parsing prose

  Two measured hazards shape how the invocations are run rather than how
  they are printed. Reading an exit status through a pipe returns the
  reader's status, so a failing step reads as a pass; and a reader that
  closes a pipe early makes a successful writer report failure. The
  tracker also writes warnings to its error stream while its data goes to
  its output stream, so a wrapper that merges the two corrupts the answer
  it is about to parse.

  @AC-SWARM-CORE-001.1
  Scenario Outline: The exit status says what the caller should do next
    Given a command whose outcome is "<outcome>"
    When it returns
    Then its exit status is "<status>"

    Examples:
      | outcome                                                     | status |
      | it answered and its verdict is clear                        | 0      |
      | it answered and what it found needs attention               | 1      |
      | it found a condition the caller cannot repair in place      | 2      |
      | its own invocation was malformed                            | 3      |

  @AC-SWARM-CORE-001.2
  Scenario: A malformed invocation is distinguishable from a finding
    When a command is given an unknown flag
    Then it exits with the malformed-invocation status
    And that status is not the one it uses for a condition it found

  @AC-SWARM-CORE-001.3
  Scenario: An exit that asks for attention names what would resolve it
    Given a command exiting with the attention status
    When its output is read
    Then the output names what would resolve the condition

  @AC-SWARM-CORE-001.4
  Scenario: An explained invocation executes nothing it prints
    When a command is invoked in explain mode
    Then it prints the tool invocations it would run, in order
    And no tool is executed

  @AC-SWARM-CORE-001.5
  Scenario: An argument only a prior step could supply is printed as that step
    Given a command whose second invocation takes an argument the first produces
    When it is invoked in explain mode
    Then that argument is printed as the step that produces it rather than as a value

  @AC-SWARM-CORE-001.6
  Scenario: A dry run reads live state and writes nothing
    Given a command that would write
    When it is invoked in dry-run mode
    Then its report reflects the state the reads returned
    And nothing is written

  @AC-SWARM-CORE-001.7
  Scenario: A normal run prints the invocations as it executes them
    When a command runs normally
    Then each tool invocation is printed as it is executed

  @AC-SWARM-CORE-001.8
  Scenario: Every reported line carries the command that established it
    When a command reports a fact
    Then the line names the command a reader can re-run to reproduce it

  @AC-SWARM-CORE-001.9
  Scenario: The machine-readable form carries the same facts as the text form
    When a command is invoked in both forms against the same state
    Then every fact in the text form is present in the machine-readable one

  @AC-SWARM-CORE-001.10
  Scenario: A step's exit status is read from the step itself
    Given a step whose command fails while a reader consumes its output
    When the step's result is reported
    Then the step is reported as failed

  @AC-SWARM-CORE-001.11
  Scenario: A step that succeeds while its reader stops early is not reported as failed
    Given a step whose reader stops consuming output before the command finishes
    When the step's result is reported
    Then the step is reported as succeeded

  @AC-SWARM-CORE-001.12 @negative
  Scenario: The output and error streams of a tool are never merged
    Given a tool that writes a warning to its error stream and data to its output stream
    When the command parses the data
    Then the warning is not part of what is parsed

  @AC-SWARM-CORE-001.13
  Scenario: Explain output is a runnable transcript
    Given a command invoked in explain mode
    When its printed invocations are run by hand in order
    Then they perform what the command would have performed

  @AC-SWARM-CORE-001.15
  Scenario Outline: A refusal takes the status matching what the caller can do about it
    Given a command refusing because "<reason>"
    When it returns
    Then its exit status is "<status>"

    Examples:
      | reason                                                  | status |
      | the arguments it was given are not valid                | 3      |
      | the state it read is wrong and the caller can repair it  | 1      |
      | the state it read is wrong and the caller cannot         | 2      |

  @AC-SWARM-CORE-001.14 @negative
  Scenario: A command that mutates run state refuses an empty scope
    When a command that reads or mutates run state is invoked with an empty run scope
    Then it exits with the malformed-invocation status
    And nothing is read or written outside that scope
```

### FEAT-SWARM-CORE-002: Read and update the run manifest

**Commands:**

```text
swarm manifest get     --run <id> [--key <key>] [--json] [--full-result]
swarm manifest set     --run <id> --key <key> --value <value> [--if-absent] [--json]
swarm manifest require --run <id> --key <key>...
```

**`get` is a lookup, and answers with the value.** Every other command
here answers with a report, because a report is the answer: what was
written, what was refused, which rules held. A key lookup is not — its
caller wants the value, and an envelope it has to walk to reach one puts
a parser between an agent and a string. So `get --key <k>` prints the
value alone, `get` with no key prints the record as one object, and
`--full-result` is how a caller asks for the report instead. The exit
status carries the same three answers either way, which is what makes the
bare form a rendering rather than a loss.

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
    Given a key whose declared type is a list of bead ids
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
    When a bead appears in more than one of the admitted, excluded and moved lists
    Then the command exits non-zero naming that bead

  @AC-SWARM-CORE-002.13
  Scenario: Several keys written in one invocation are applied as one merge
    Given a manifest and four keys to write
    When they are written in one invocation
    Then the manifest is read and written once

  @AC-SWARM-CORE-002.14 @negative
  Scenario: A multi-key write that fails on one key writes none of them
    Given four keys to write, one of which is refused
    When they are written in one invocation
    Then the refused key is named
    And none of the four is stored

  @AC-SWARM-CORE-002.15
  Scenario: Validation reports every rule that fails, not the first
    Given a manifest breaking three of the rules above
    When it is validated
    Then all three are named
    And the command exits non-zero

  @AC-SWARM-CORE-002.16
  Scenario: A manifest that satisfies every rule validates clean
    Given a manifest whose keys are all present, correctly typed and mutually consistent
    When it is validated
    Then the command exits zero

  @AC-SWARM-CORE-002.17
  Scenario: Validation separates a key that is absent from one that is wrong
    Given a manifest missing a key whose absence is legitimate at this stage of a run
    When it is validated
    Then that key is reported as not yet written rather than as a failure

  @AC-SWARM-CORE-002.18
  Scenario: A precondition check names every missing key
    Given a caller requiring three keys of which two are absent
    When the requirement is checked
    Then both absent keys are named
    And the command exits non-zero

  @AC-SWARM-CORE-002.19
  Scenario: Reading, requiring and validating share one declaration of the keys
    Given a key added to the manifest's declared set
    When it is written, required and validated
    Then each reads that key's name, type and stage from the same declaration

  @AC-SWARM-CORE-002.20
  Scenario Outline: Reading one key answers with the value and nothing else
    Given a manifest carrying a key whose value is "<value>"
    When that one key is read
    Then the whole of the output is that value
    And a caller substituting the command receives "<value>"

    Examples:
      | value                        |
      | claude/branch-x              |
      | two  words, doubly spaced    |
      | a line,\nand a second line   |
      | --not-a-flag                 |

  @AC-SWARM-CORE-002.21
  Scenario: Reading with no key answers with the whole record
    Given a manifest carrying several keys
    When it is read with no key named
    Then the output is one object carrying every key the record holds

  @AC-SWARM-CORE-002.22
  Scenario Outline: A key read answers with its status whichever form was asked for
    Given a key that is "<state>"
    When it is read, with and without the full result
    Then both exit with "<status>"

    Examples:
      | state                       | status |
      | written                     | 0      |
      | declared and not written    | 1      |
      | outside the declared set    | 3      |
```


### FEAT-SWARM-CORE-003: Report a run's status from state

**Command:**

```text
swarm status --run <id> [--stall-after <duration>] [--json]
```

**The command reports measurements; it does not diagnose.** Whether a
subagent is still thinking cannot be observed from a shell — a stalled
one never finishes, so no event fires, and absence of an event is not an
event. Three things *are* observable: the number of standing worktrees,
the age of the last commit on the sprint branch, and the depth of the
ready front. So the report gives each with the value it had, gives the
threshold it compared against, and stops there.

**Which means it never says a run is stalled**, because it does not know
that and cannot. It says how long it has been since the last commit and
what else was true at that moment — *no commit for 3h12m; no worktrees
standing; 4 beads ready* — and leaves the reading to whoever can see the
harness. Asserting a stall would be the same defect as the health label
this replaces, one step further along: a confident value nobody measured.

**Per-agent drain is reported from two queries**, and both are required:
the ready query excludes in-progress work, so on its own it reports a
subagent mid-bead as finished.

**Replaces:** `bead-bonanza/references/supervision.md`, "Where is this run" and "When
the run is done"; the "The run's health — derived, never asserted" section
of `bead-bonanza/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-CORE-003
Feature: Report a run's status from state

  As a coordinator or a session resuming one
  I want progress, completion and per-agent drain measured from version
  control and the tracker
  So that a run making no progress is distinguishable from a finished one
  without the coordinator's cooperation

  This is the liveness backbone and the largest token saving in the
  family, because every part of it is a derivation. A run once read
  healthy through a five-hour outage in which nothing was dispatched,
  because health was a label written by the component that had stopped.
  Completion needs a walk the tracker does not offer: three queries look
  like they answer "every bead in this run" and each answers
  something else.

  The measurements are reported as measurements. This command has no way
  to tell a stalled run from a slow one, so it never claims to, and every
  criterion below is written against what it can observe.

  @AC-SWARM-CORE-003.1
  Scenario: A run making no progress is reported as the measurements that show it
    Given no worktrees are standing
    And the ready front is not empty
    And the last commit is older than the threshold in force
    Then the report states the standing worktree count, the timestamp and age of the last commit, and the depth of the ready front
    And it states the threshold it compared the commit age against
    And the command exits with the attention status

  @AC-SWARM-CORE-003.11 @negative
  Scenario: The report never asserts that a run has stalled
    Given a run whose last commit is older than the threshold in force
    When progress is reported
    Then the report states the elapsed time and what else was true at that moment
    And it does not state that the run has stalled

  @AC-SWARM-CORE-003.2
  Scenario: Progress is never read from a stored value
    Given a run epic carrying a health value in its metadata that contradicts its state
    When progress is reported
    Then the report reflects the measurements, not the stored value

  @AC-SWARM-CORE-003.3
  Scenario: Completion counts every bead in the run, however nested
    Given a run whose beads are not direct children of the run epic
    When completion is reported
    Then every bead in the run is counted

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
    When progress is reported
    Then the report states that whether a subagent is still working is not observable from outside the harness
    And it names the three measurements it gives instead

  @AC-SWARM-CORE-003.7
  Scenario: The live queue is compared against the admitted set
    Given beads present in the run that the manifest's admitted list does not name
    When status is reported
    Then those beads are named as additions to the run's scope

  @AC-SWARM-CORE-003.8
  Scenario: Beads excluded at admission are reported beside the completion verdict
    Given a run whose admitted work is all closed and whose admission excluded three beads
    When completion is reported
    Then the run is reported as done over a narrowed scope naming those three

  @AC-SWARM-CORE-003.10
  Scenario: A manifest key that is absent is reported as absent, not as an empty answer
    Given a run whose manifest records no admitted list
    When the scope comparison is reported
    Then it is reported as unestablished rather than as no additions
    And the command exits non-zero

  @AC-SWARM-CORE-003.9
  Scenario: A blocked bead whose worktree still stands is reported as unresolved
    Given a bead marked blocked whose dispatch's worktree still exists
    When status is reported
    Then that bead is reported as an unresolved conflict holding its evidence
```


## References

- `docs/research/swarm-cli/2026-08-25-capability-audit.md` — the audit
  these requirements are drawn from, including the inventory of traps the
  Overview refers to; §7 is the classification that decides what earns a
  command
- [[module-bead-bonanza]] — the module whose components call this CLI
- [[gherkin-guide]] — the notation these criteria are written in
