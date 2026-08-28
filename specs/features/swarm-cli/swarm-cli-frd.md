---
title: "FRD: Swarm CLI — the run's structure and record"
description: The run's own structure and record — the commands that build a
  run, plan and gate its queue, capture its findings, and reconstruct it when
  nobody is holding it
slug: frd-swarm-cli
type: frd
status: draft
updated_at: 2026-08-28T00:00:00Z
sources: []
---

# Swarm CLI — the run's structure and record

## Overview

A `bead-bonanza` run is a distributed system whose control plane is
documentation. Its invariants are prose, its state transitions are prose,
and its guards are paragraphs an agent has to remember at the moment it
matters. Measured across the ten sprint-family skills: 79 distinct
capabilities, of which **73 carry a recorded trap** — a way the obvious
invocation produces the wrong result.

**This is the run's own structure and record**: building its shape,
partitioning and assigning its queue, gating readiness, claiming an item,
recording a finding, harvesting the findings at the end, and
reconstructing a run whose session is gone.

**It is built on [[frd-swarm-core]]** — the command conventions, and the
manifest component every command here reads a run through — and it sits
beside [[frd-swarm-dispatch]], the loop that sends work out and takes it
back. It waits on neither beyond the core.

**This is the least urgent of the three, not the least valuable.** Five of
these seven run exactly once per run, so a defect in them costs one
careful reading rather than being met at every re-query.

**One requirement number is a gap.** 002 named `swarm run manifest`, now
FEAT-SWARM-CORE-002. Its identifier is retired rather than reassigned, so
a note referring to it still resolves to the thing it meant.

Its value is the same three things, in the same order:

- **Stabilize.** A silent wrong answer becomes a non-zero exit. Three of
  these are cases where a *correct* agent still loses: a claim on an
  unassigned item silently rewrites its owner at exit 0; a plan document
  whose edge orientation is reversed inverts every dependency at exit 0;
  an assignee read from the tracker's owner display passes on an epic
  where nothing is assigned, because that display shows the creator.
- **Simplify.** An invariant enforced by a command is prose that can be
  deleted. Each requirement names the prose it retires.
- **Optimize.** A derived answer an agent assembles by hand costs N
  queries and N results held in the caller's context. Reconstruction and
  the findings harvest are the two largest here.

### What this document assumes rather than specifies

**The command conventions and the explain flag** are [[frd-swarm-core]]'s,
and every command here carries them: a run scope that is an argument
rather than a convention, `--explain`, `--json`, `--dry-run`, and the exit
vocabulary.

**The manifest** is that document's too. Commands here read and write it
through the one component it specifies — `run open` writes the admission
record, the retro writes its counts — rather than reaching into the run
epic's metadata for themselves.

**Judgment.** Whether a work item is genuinely ready, whether an unwired
edge is a real dependency, whether a review finding is in scope, whether a
cluster of findings is one problem, and whether partial work is finishable
all stay with an agent. A command that guessed at them would be worse than
the paragraph it replaced.

**Orchestration.** The CLI spawns no agent, arms no monitor, schedules
nothing, and holds no state outside the tracker.

**Single commands that already work.** Reading one work item, adding a
comment, creating one — these do what they say. Wrapping them adds a
layer and hides a tool the agent already knows behind one it does not.

### Where each command runs

The annotation is the requirement that specifies it; a step annotated with
another document's requirement belongs to that document.

```text
run epic ................................. swarm run open              001
├ pre-flight ............................. swarm run manifest set      core
├ scope
│   └ AUTHORIZATION GATE ................. (human; taken by sprint-scope)
├ groom & plan ─┐ ........................ swarm queue reconcile       004
│               │                          swarm item bounce           004
│               │                          swarm queue partition load  003
│               │                          swarm queue assign          003
│               │                          swarm queue gate            003
├ TARGET ───────┤ FLAT: tasks only ....... (the dispatch loop, its own doc)
│  ├ dispatch ┐ │
│  ├ merge    │ │
│  └ supervise┘ │
├ review ───────┘ ........................ swarm run manifest get      core
├ retro .................................. swarm finding harvest       007
│                                          swarm finding route         007
└ post-flight

inside a dispatch, run by the dev agent .. swarm item claim            005
                                           swarm finding record        006
at any point, holding no context ......... swarm run recover           008
```

`core` marks the steps [[frd-swarm-core]] specifies rather than this
document.

Read as a table:

| Lifecycle step | Caller | Commands | Frequency |
|---|---|---|---|
| pre-flight | coordinator | `run manifest set` (core) | twice per run |
| scope | `sprint-scope` | `run open` (001), `run manifest set` (core) | once |
| groom & plan | `sprint-plan` | `queue reconcile`, `item bounce` (004); `queue partition load`, `queue assign`, `queue gate` (003) | once per pass |
| inside a dispatch | dev agent | `item claim` (005), `finding record` (006) | once per item, plus one per finding |
| review | `sprint-review` | `run manifest get` and `set` for the pass counter (core) | once per pass |
| retro | `sprint-retro` | `finding harvest`, `finding route` (007), `run manifest set` (core) | once |
| resume | a fresh session | `run recover` (008) | on every resume |

**The vocabulary is tracker-neutral on purpose.** A unit of work is a
**work item**, not the name one tracker gives it. The tracker in use today
is the only one implemented, and every command goes through one interface
so a second is possible later; nothing in these requirements names a
tracker's own vocabulary except where a measured behaviour is being pinned.

## Terminology

| Term | Meaning |
|---|---|
| **run** | one execution of `bead-bonanza` over one run epic |
| **run epic** | the work item whose id is the run id and whose metadata is the run manifest |
| **work item** | one unit of work in the tracker. The tracker-neutral name for what this project's tracker calls a bead and others call an issue, story or ticket |
| **phase epic** | one of the epics chained beneath the run epic, in order: pre-flight, scope, plan, target, review, retro, post-flight |
| **target epic** | the flat phase epic whose task children are the run's work |
| **manifest** | the JSON object in the run epic's metadata; the run's only durable record |
| **dispatch id** | one agent's whole identity — its assignee, its tracker actor, and its branch name. Derived from the run epic id, never minted |
| **ready front** | the complete set of work items in the run that nothing is blocking right now |
| **collision edge** | a blocking edge between two work items that must not run at once because they share a component or fixture. Not a real dependency; both ends carry a collision mark naming the surface |
| **bounce** | move a work item out of the run's dispatchable set by labelling it for a human, leaving it open. Not a close and not a deferral |
| **finding** | one record of friction captured at the moment it happened, carrying a kind that routes it |
| **derived answer** | an answer requiring a tree walk or a join across the tracker and version control that no single command returns |
| **silent failure** | a command that returns the wrong answer at exit 0 |

## Requirements

### FEAT-SWARM-RUN-001: Build a run's epic shape

**Command:**

```text
swarm run open --title <title> --admit <work-item-or-epic-id>... [--second-target <epic-id>] [--dry-run] [--json]
```

Returns the run epic id and the target epic id. **The judgment stays with
`sprint-scope`** — which items to admit, which to exclude, and whether to
adopt an epic the user already owns are decided before this command is
called, and it refuses rather than deciding.

**Replaces:** `sprint-scope/SKILL.md`, "Build the shape, not just the
epic" and its "Assign every phase epic to its own agent id" paragraph;
`bead-bonanza/references/run-record.md`, "The run's epic shape".

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-RUN-001
Feature: Build a run's epic shape

  As a coordinator opening a run
  I want the run epic, its phase chain, its flat target and its
  authorization gate created as one act
  So that a shape that cannot dispatch is refused rather than built

  This runs once per run and still carries this document's largest blast
  radius. The tracker refuses a task blocking an epic and an epic
  blocking a task, so a phase can only be gated by an epic and the
  authorization gate can only be a task child. Build the shape wrong and
  nothing dispatches, with no error anywhere to say why.

  The shape is the run's schedule, so the criteria below specify it
  rather than leaving it to the implementation: seven phase epics in one
  order, the work as flat task children of the target, and the gate as a
  task child of scope.

  @AC-SWARM-RUN-001.1
  Scenario: The created shape dispatches exactly the target's open leaves
    When a run shape is created and the phases before the target are closed
    Then the run's ready front contains exactly the target epic's open task children

  @AC-SWARM-RUN-001.2
  Scenario: A partially built shape is not left behind
    Given the creation fails partway
    When the command reports the failure
    Then the run epic and its phases are either all present or all absent

  @AC-SWARM-RUN-001.3
  Scenario Outline: Every phase epic is assigned to its role identity
    Given a run whose epic id yields the run id "r7k"
    When a run shape is created
    Then the "<phase>" epic carries the assignee "<assignee>"

    Examples:
      | phase       | assignee        |
      | pre-flight  | coordinator-r7k |
      | scope       | coordinator-r7k |
      | plan        | plan-r7k        |
      | target      | coordinator-r7k |
      | review      | review-r7k      |
      | retro       | retro-r7k       |
      | post-flight | coordinator-r7k |

  @AC-SWARM-RUN-001.4
  Scenario: The authorization gate holds the whole chain
    Given a run shape whose authorization gate is unresolved
    When the ready front is queried
    Then it is empty

  @AC-SWARM-RUN-001.5
  Scenario: A non-flat target is refused
    When a run shape is created with a sub-epic beneath the target epic
    Then the command exits non-zero naming the offending child

  @AC-SWARM-RUN-001.6
  Scenario Outline: Each phase blocks the phase after it
    When a run shape is created
    Then the "<earlier>" epic blocks the "<later>" epic
    And that edge joins two epics rather than an epic and a task

    Examples:
      | earlier    | later       |
      | pre-flight | scope       |
      | scope      | plan        |
      | plan       | target      |
      | target     | review      |
      | review     | retro       |
      | retro      | post-flight |

  @AC-SWARM-RUN-001.7
  Scenario: The authorization gate is a task child of the scope epic
    When a run shape is created
    Then the gate is held against a task child of the scope epic rather than against a phase epic

  @AC-SWARM-RUN-001.8
  Scenario: Work that decomposes becomes a second target beside the first
    Given an admitted epic whose own children are epics
    When a run shape is created
    Then that epic becomes a second target epic under the run epic
    And it is not nested beneath the first target epic

  @AC-SWARM-RUN-001.9 @negative
  Scenario: A run epic that already exists is not rebuilt
    Given a run epic that already carries a phase chain
    When the command is asked to build a shape over it
    Then it exits non-zero
    And no epic is created, moved or reassigned

  @AC-SWARM-RUN-001.10
  Scenario: The admission record is written as the shape is created
    When a run shape is created
    Then the manifest records which items were admitted, which were excluded with a reason, and which were moved and from where

  @AC-SWARM-RUN-001.11 @negative
  Scenario: A work item is never moved out of an epic that already owns it
    Given an admitted work item that already belongs to another epic
    When a run shape is created without an explicit instruction to move it
    Then the work item keeps its existing parent
    And it is recorded as excluded with that reason
```

### FEAT-SWARM-RUN-002 — moved

`swarm run manifest` is specified in [[frd-swarm-core]] as
FEAT-SWARM-CORE-002. It moved because every command in the family reads a
run through that one component, so its behaviour cannot be defined by the
last document to be written. This identifier is retired and is not reused.
### FEAT-SWARM-RUN-003: Partition the queue, assign it, and gate the epic

**Commands:**

```text
swarm queue partition load --run <id> --file <edges.jsonl> [--dry-run] [--json]
swarm queue assign         --run <id> --file <assignments.jsonl> [--dry-run] [--json]
swarm queue gate           --run <id> [--json]
```

Three subcommands and one requirement, because **the gate is the check on
the other two**: an unassigned work item and a mis-oriented edge are both
invisible until something asserts the front and the ownership after the
writes land.

**On a missing concurrency cap the gate refuses rather than defaulting.**
The cap is resolved once at pre-flight, from the harness, and recorded on
the manifest; a default substituted at gate time would be a number nobody
measured reading as one somebody did — and the gate exists precisely to
catch a pool that will silently never drain. The place a default belongs
is pre-flight's resolution step, where the harness's documented value is
taken and the resolution is recorded with its source.

**Replaces:** `sprint-plan/references/partitioning.md`, the whole of
"Collision edges", "The edge file" and its two validators, and
"`concurrency_cap`"; `sprint-plan/SKILL.md`, "Assign every planned bead,
then gate the epic"; the four mechanical checks of
`groom-beads/references/epic-definition-of-ready.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-RUN-003
Feature: Partition the queue, assign it, and gate the epic

  As a planner preparing a run for dispatch
  I want collision edges loaded, every work item assigned, and the epic
  checked for swarmability as one act
  So that the ready front is safe to dispatch whole

  These three are one piece of work because the third checks the first
  two. A collision edge is a policy expressed as a dependency, and three
  ways of getting it wrong are silent: the orientation is that the first
  endpoint waits for the second, so reversing it inverts the dependency
  at exit zero; the collision mark must be stored bare, since a quoted
  value stores its quotes and then matches no other item while looking
  correct in every listing; and a cluster must be chained rather than
  fanned, since a fan leaves the rest of the cluster ready at once and
  defeats the point. Assignment is now the whole handover, so an
  unassigned item is one no agent will ever see. The epic gate is what
  catches that, and it has a silent failure of its own: an assignee read
  from the tracker's owner display passes on an epic where nothing is
  assigned, because that display shows the creator.

  Two of the epic gate's six checks stay with the planner and are not
  specified here: whether an unwired edge is a real dependency, and
  whether anything the run needs can be derived only by the planner.

  @AC-SWARM-RUN-003.1
  Scenario: Two work items sharing a collision mark are never ready together
    Given two work items separated by a collision edge
    When the ready front is queried
    Then at most one of them appears in it

  @AC-SWARM-RUN-003.2
  Scenario: An inverted edge is detected rather than loaded silently
    Given an edge file whose orientation would invert a dependency
    When the file is loaded and the front asserted
    Then the command exits non-zero naming the offending edge

  @AC-SWARM-RUN-003.3
  Scenario: A collision mark is stored bare
    When a collision mark is written
    Then reading it back returns the value without added quoting

  @AC-SWARM-RUN-003.4
  Scenario: A cluster is chained rather than fanned
    Given three work items sharing one surface
    When the partition is loaded
    Then exactly one of the three is in the ready front

  @AC-SWARM-RUN-003.5
  Scenario: An unassigned child is named by the epic gate
    Given a target epic with one unassigned child
    When the epic is gated for swarming
    Then that child's id is named
    And the command exits non-zero

  @AC-SWARM-RUN-003.6
  Scenario: The epic gate reads assignment from the tracker's data, not its display
    Given a work item whose creator is set and whose assignee is not
    When the epic is gated for swarming
    Then that item is reported as unassigned

  @AC-SWARM-RUN-003.7
  Scenario: A dev pool larger than the concurrency cap is refused
    Given planned dev assignees exceeding the cap the manifest records
    When the epic is gated for swarming
    Then the count and the cap are both named
    And the command exits non-zero

  @AC-SWARM-RUN-003.8
  Scenario: Both ends of a collision edge carry the same mark
    Given a loaded collision edge
    When the marks are read back from the tracker
    Then both endpoint items carry the same collision value

  @AC-SWARM-RUN-003.9
  Scenario: A collision mark with no edge to justify it is named
    Given a work item carrying a collision mark that no loaded edge accounts for
    When the partition is validated
    Then that item is named
    And the command exits non-zero

  @AC-SWARM-RUN-003.10 @negative
  Scenario: A sequencing edge carries no collision mark
    Given an edge recorded as a sequencing constraint
    When the marks are read back
    Then neither endpoint carries a collision mark

  @AC-SWARM-RUN-003.11
  Scenario: A collision chain that contradicts a real dependency is refused
    Given two work items in one cluster already joined by a real dependency
    When a collision chain would order them the other way round
    Then the command exits non-zero naming both items

  @AC-SWARM-RUN-003.12
  Scenario: A file that fails validation wires no edge at all
    Given an edge file whose second line names a work item that does not exist
    When the file is loaded
    Then no edge from that file is wired
    And the offending line is named

  @AC-SWARM-RUN-003.13
  Scenario: Loading the same file twice changes nothing
    Given an edge file that has already loaded
    When it is loaded again
    Then the graph is unchanged
    And the command exits zero

  @AC-SWARM-RUN-003.14
  Scenario: An edge that would strengthen an existing weaker edge is reported
    Given two work items already joined by a non-blocking edge
    When the file asks for a blocking edge between them
    Then the command exits non-zero naming both items and the existing edge type

  @AC-SWARM-RUN-003.15
  Scenario: Assignees are derived from the run epic id and never minted
    Given a run whose epic id yields the run id "r7k" and a second batch
    When the batch is assigned
    Then every work item in it carries the assignee "dev-r7k-2"

  @AC-SWARM-RUN-003.16
  Scenario: The epic gate counts closed and deferred children
    Given a target epic with one closed child carrying no assignee
    When the epic is gated for swarming
    Then that child is included in the check rather than skipped

  @AC-SWARM-RUN-003.17 @negative
  Scenario: A missing concurrency cap is reported as unmeasurable
    Given a manifest recording no concurrency cap
    When the epic is gated for swarming
    Then the check is reported as unmeasurable
    And no default cap is substituted
    And the command exits non-zero

  @AC-SWARM-RUN-003.18
  Scenario: The pool count includes dev assignees only
    Given a run carrying four dev assignees and four ceremony assignees
    When the epic is gated for swarming
    Then the reported pool count is four
```

### FEAT-SWARM-RUN-004: Return a readiness verdict for a queue

**Commands:**

```text
swarm queue reconcile --run <id> --verdicts <file> [--json]
swarm item bounce     <work-item-id> --run <id> --gap <text>
swarm item unbounce   <work-item-id> --run <id> --answer <text>
swarm queue human     --run <id> [--json]
```

**A bounce is not a close and not a deferral.** The item stays open, gains
the label the human queue is read from, and is parented to a late phase so
it surfaces at the end. Closing it would say the run did the work;
deferring it would say the run chose not to and would take it out of the
run's outstanding count, and it is precisely an item the run is still
waiting on. The resolution is the label's removal plus the answer recorded
on the item, after which it re-enters the gate.

**Replaces:** the readiness-gate and bounce sections of
`sprint-plan/SKILL.md`; the human-channel row and its two measured tracker
notes in `references/run-record.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-RUN-004
Feature: Return a readiness verdict for a queue

  As a planner gating a queue before dispatch
  I want every work item in the queue to carry a verdict, and unready
  items removed from the front
  So that an item with no verdict is never mistaken for a ready one

  The mechanical half of readiness is accounting rather than judgment:
  whether every item in the queue came back with a verdict at all. The
  judgment half stays with the grooming agent, which owns the checklist.
  Bouncing has its own silent failure: an unready item marked only with
  a grooming annotation stays dispatchable, because the front reads
  status and not annotations.

  @AC-SWARM-RUN-004.1
  Scenario: Every work item in the queue is accounted for
    Given a queue and a set of verdicts
    When the verdicts are reconciled against the queue
    Then any work item without a verdict is named
    And the command exits non-zero

  @AC-SWARM-RUN-004.2
  Scenario: A bounced work item leaves the dispatch front
    Given a work item bounced as not ready
    When the dispatch front is queried
    Then that item is not in it

  @AC-SWARM-RUN-004.3
  Scenario: A bounced work item is visible to a human
    Given a work item bounced as not ready
    When the human queue is listed
    Then that item appears in it

  @AC-SWARM-RUN-004.4
  Scenario: A resolved bounce returns the work item to the gate
    Given a bounced work item whose gap has been closed
    When the bounce is resolved
    Then the item re-enters the readiness gate on the next pass

  @AC-SWARM-RUN-004.5 @negative
  Scenario: A bounce neither closes nor defers the work item
    Given a work item bounced as not ready
    When its status is read
    Then it is open
    And it is neither closed nor deferred

  @AC-SWARM-RUN-004.6
  Scenario: The human queue omits a work item that has closed
    Given a bounced work item that has since been closed
    When the human queue is listed
    Then that item does not appear in it

  @AC-SWARM-RUN-004.7
  Scenario: Resolving a bounce records the answer on the work item
    Given a bounced work item
    When the bounce is resolved with an answer
    Then the answer is recorded on the item
    And the item is not closed

  @AC-SWARM-RUN-004.8
  Scenario: A bounced work item is counted in the run's outstanding work
    Given a run whose only unfinished item is bounced
    When the run's outstanding work is counted
    Then that item is counted
    And the run is not reported as drained
```

### FEAT-SWARM-RUN-005: Claim a work item as the acting agent

**Command:**

```text
swarm item claim <work-item-id> --actor <dispatch-id> [--run <id>] [--json]
```

**Replaces:** step 1 of the per-item ceremony in `run-batch/SKILL.md`, and
the "Why the export matters" table in `references/dispatch-policy.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-RUN-005
Feature: Claim a work item as the acting agent

  As a dev agent starting a work item
  I want the claim refused unless the item is mine and my identity is
  explicit
  So that a claim can never quietly move an item's ownership

  The underlying claim is a silent failure in exactly one case, and it is
  the common one: on an item with no assignee, with no actor in the
  environment, the claim succeeds at exit zero and writes the version
  control tool's configured user name into the assignee. Nothing errors,
  and the run's ownership view decays from that point on.

  A claim also has to set the status and the assignee together. Setting
  the status alone leaves an item that reads as in progress and is
  claimed by nobody, which is the state no query can attribute.

  @AC-SWARM-RUN-005.1
  Scenario: A claim is refused when the work item belongs to another agent
    Given a work item assigned to a different agent
    When a claim is attempted
    Then the command exits non-zero
    And the item's assignee is unchanged

  @AC-SWARM-RUN-005.2
  Scenario: A claim on an unassigned work item does not silently take ownership
    Given a work item with no assignee
    When a claim is attempted without an explicit actor
    Then the command exits non-zero

  @AC-SWARM-RUN-005.3
  Scenario: A claim sets the status and the assignee in one write
    Given a work item assigned to the acting agent
    When it is claimed
    Then its status is in progress
    And its assignee is the acting agent

  @AC-SWARM-RUN-005.4
  Scenario: Re-claiming a work item the actor already holds changes nothing
    Given a work item already claimed by the acting agent
    When it is claimed again
    Then the command exits zero
    And the item's status and assignee are unchanged

  @AC-SWARM-RUN-005.5 @negative
  Scenario: The actor is never taken from version control configuration
    Given an environment whose version control tool has a configured user name
    When a claim is attempted without an explicit actor
    Then that name is not written to the work item
```

### FEAT-SWARM-RUN-006: Record and read a finding

**Commands:**

```text
swarm finding record <work-item-id> --kind <kind> [--target <target>] --evidence <text|@file> [--corrects <finding-id>]
swarm finding list   --run <id> [--item <work-item-id>] [--json]
```

**Replaces:** step 6 of `run-batch/SKILL.md` and the kind table in it; the
"Where to record" and "Finding ids are derived" sections of
`specs/standards/findings-capture-protocol.md` become a validated write
rather than a rule an agent applies.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-RUN-006
Feature: Record and read a finding

  As an agent capturing friction, or a retro reading it
  I want a finding's kind validated when it is written
  So that a finding cannot be recorded in a form nothing will ever route

  This is the cheapest stabilization in the family. A finding routes by
  its kind, so a kind outside the nine routes nowhere and is invisible to
  every later sweep. Two consecutive runs lost their single most
  consequential finding that way, and the count of such findings was
  under-reported because the sweep that counted them had the same defect.

  What the command does not check is the finding's content. Whether the
  evidence is the right evidence, and whether the agent has drifted into
  analysis, are judgment and stay with the capturing agent and the retro.

  @AC-SWARM-RUN-006.1
  Scenario: A finding with an unknown kind is refused
    When a finding is recorded with a kind outside the declared set
    Then the command exits non-zero
    And nothing is written

  @AC-SWARM-RUN-006.2
  Scenario: A finding requiring a target is refused without one
    When a finding of a kind that requires a target is recorded without one
    Then the command exits non-zero

  @AC-SWARM-RUN-006.3
  Scenario: A finding is readable back with its work item and its ordinal
    Given findings recorded across several work items
    When they are read back
    Then each carries the item it was recorded on and its position within that item

  @AC-SWARM-RUN-006.4 @negative
  Scenario: A finding with no evidence is refused
    When a finding is recorded carrying no evidence
    Then the command exits non-zero
    And nothing is written

  @AC-SWARM-RUN-006.5 @negative
  Scenario: A finding whose only anchor is a line number is refused
    When a finding's evidence cites a path and a line number and nothing else
    Then the command exits non-zero naming the durable anchors it would accept

  @AC-SWARM-RUN-006.6
  Scenario Outline: A target is required of the kind that routes outside the project
    When a finding of kind "<kind>" is recorded without a target
    Then the command "<outcome>"

    Examples:
      | kind         | outcome        |
      | tooling      | exits non-zero |
      | gap          | exits zero     |
      | decision     | exits zero     |
      | stale        | exits zero     |
      | unverifiable | exits zero     |
      | sequencing   | exits zero     |
      | scope        | exits zero     |
      | proof        | exits zero     |
      | environment  | exits zero     |

  @AC-SWARM-RUN-006.7 @negative
  Scenario: A recorded finding is never edited or deleted
    Given a finding already recorded on a work item
    When a change to it is attempted
    Then the command exits non-zero
    And the recorded finding is unchanged

  @AC-SWARM-RUN-006.8
  Scenario: A correction is a new finding naming the one it corrects
    Given a finding already recorded on a work item
    When a correcting finding is recorded against it
    Then both findings are readable
    And the second names the first

  @AC-SWARM-RUN-006.9 @negative
  Scenario: An ordinal is derived and is not accepted as input
    When a finding is recorded with an ordinal supplied by the caller
    Then the command exits non-zero
```

### FEAT-SWARM-RUN-007: Harvest and route a run's findings

**Commands:**

```text
swarm finding harvest --run <id> [--json]
swarm finding route   --run <id> [--json]
```

**Replaces:** sections 2 and 3 of `sprint-retro/SKILL.md`, and the
"Harvesting" section of the findings-capture protocol.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-RUN-007
Feature: Harvest and route a run's findings

  As a retro turning a run's friction into work
  I want every finding in the run collected and routed by kind
  So that the retro reads a complete set rather than whichever work items
  happened to still be open

  The harvest has a silent failure that produces a plausible answer: the
  natural listing query defaults to open items, and a drained run is
  almost entirely closed, so a literal reading returns findings for the
  handful of items that never closed and drops nearly everything else.
  A second query omits comments altogether while its text output prints
  them. Routing is a lookup applied without reading content, which is
  what keeps it viable on a run with seventy findings.

  Two kinds terminate at the record and produce no issue candidate.
  A proof finding is the evidence rule being obeyed — an observed
  pre-fix failure, a negative control — and routing it as friction would
  suppress the rigour. An environment finding is nobody's defect: a
  parser that exports no error type is not the run's fault, and it
  belongs in the project's environmental-failure ledger where the next
  agent will look for it.

  @AC-SWARM-RUN-007.1
  Scenario: Findings on closed work items are harvested
    Given a run whose work items are closed
    When findings are harvested
    Then findings recorded on those items are returned

  @AC-SWARM-RUN-007.2
  Scenario: A zero harvest on a drained run is reported as a failure
    Given a run whose work items are closed and which recorded findings
    When the harvest returns nothing
    Then the command exits non-zero

  @AC-SWARM-RUN-007.3
  Scenario: Each finding is routed by its kind alone
    When findings are routed
    Then each is assigned a destination determined solely by its kind

  @AC-SWARM-RUN-007.4
  Scenario Outline: Each kind has one destination
    Given a finding of kind "<kind>"
    When routing runs
    Then its destination is "<destination>"
    And an issue candidate is "<candidate>"

    Examples:
      | kind            | destination                      | candidate    |
      | gap             | the project audience             | produced     |
      | decision        | the project audience             | produced     |
      | stale           | the project audience             | produced     |
      | unverifiable    | the project audience             | produced     |
      | sequencing      | the project audience             | produced     |
      | scope           | the project audience             | produced     |
      | tooling         | the skill audience               | produced     |
      | proof           | the record it supports           | not produced |
      | environment     | the environmental-failure ledger | not produced |
      | an unknown kind | the unrouted list                | not produced |

  @AC-SWARM-RUN-007.5
  Scenario: The unrouted count is reported even when it is zero
    When routing runs
    Then the number of findings that matched no destination is reported

  @AC-SWARM-RUN-007.6
  Scenario: The harvest reaches every work item in the run, however nested
    Given a run whose work items are children of the target epic rather than of the run epic
    When findings are harvested
    Then findings on those items are returned

  @AC-SWARM-RUN-007.7
  Scenario: Filing a candidate records a back-link on the finding's work item
    Given a candidate filed from two findings
    When the filing is recorded
    Then each finding's work item carries a record naming the issue that was filed and the finding it came from
```

### FEAT-SWARM-RUN-008: Reconstruct a run nobody is holding

**Command:**

```text
swarm run recover [--run <id>] [--branch <name>] [--json]
```

**Three ways to identify the run, tried in order, and each is reported.**
An explicit `--run` is taken as a starting point and still confirmed
against the branch. Otherwise the checked-out branch name is matched
against the branch recorded in candidate run epics' manifests. Failing
that, the close-commit subjects on the branch carry work item ids in the
form `<id> P<n> <title>`, and any one of them resolves to its parent chain
and so to the run epic.

**The third way has a measured failure mode**: a branch that was
squash-merged carries no such subject, so the completed set cannot be read
from history. That is reported as unestablished, never as an empty
completed set — the two are the difference between a run that did nothing
and a run whose record this checkout cannot see.

**Replaces:** `sprint-recovery/SKILL.md` steps 1 through 5, including the
five-row classification table a human is currently asked to apply.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-RUN-008
Feature: Reconstruct a run nobody is holding

  As a session resuming a run with no context
  I want the run derived from its branch, its epic and the surviving
  worktrees
  So that a run survives the loss of the session that started it

  Every input is durable and every output is a derived answer, which is
  why this is both a large token saving and the mechanism by which a run
  survives at all. The per-item classification is a pure function of two
  inputs — an item's status and whether its dispatch's worktree still
  exists — currently written as a table for a human to apply.

  @AC-SWARM-RUN-008.1
  Scenario Outline: The run is identified from whichever durable source answers
    Given a checkout on a sprint branch and "<available>"
    When reconstruction runs
    Then it names the run epic
    And it names "<source>" as what identified it

    Examples:
      | available                                              | source              |
      | a run epic id supplied by the caller                   | the supplied id     |
      | a run epic whose manifest records this branch          | the manifest branch |
      | close commits carrying work item ids in their subjects | the branch history  |

  @AC-SWARM-RUN-008.2
  Scenario: Completed work items are read from the branch history
    When reconstruction runs
    Then every work item whose close commit is on the branch is reported completed

  @AC-SWARM-RUN-008.3
  Scenario: Stale worktree registrations are not reported as in flight
    Given a registered worktree whose directory no longer exists
    When reconstruction runs
    Then it is not counted as in flight

  @AC-SWARM-RUN-008.4
  Scenario: Every unfinished work item is classified with the signal that settled it
    When reconstruction runs
    Then each unfinished item carries both its classification and the signal it was derived from

  @AC-SWARM-RUN-008.5
  Scenario: What could not be established is named rather than omitted
    Given a source that cannot be read
    When reconstruction reports
    Then the affected facts are listed as unestablished

  @AC-SWARM-RUN-008.6
  Scenario Outline: Each unfinished work item is classified from its status and its worktree
    Given a work item whose status is "<status>" and whose dispatch's worktree is "<worktree>"
    When reconstruction runs
    Then that item is reported as "<classification>"

    Examples:
      | status      | worktree | classification                      |
      | in progress | present  | the agent died holding it           |
      | open        | present  | started and abandoned, or unstarted |
      | in progress | absent   | died before or without a worktree   |
      | open        | absent   | never started                       |
      | blocked     | either   | stopped on purpose                  |

  @AC-SWARM-RUN-008.7
  Scenario: Worktree registrations are cleared before the in-flight set is read
    When reconstruction runs
    Then stale registrations are cleared before the in-flight set is computed

  @AC-SWARM-RUN-008.8
  Scenario: The review pass counter is carried forward
    Given a manifest recording spent review passes
    When reconstruction reports the run
    Then it reports the spent count rather than zero

  @AC-SWARM-RUN-008.9 @negative
  Scenario: Reconstruction creates no branch
    Given a manifest recording a sprint branch
    When reconstruction runs
    Then no branch is created

  @AC-SWARM-RUN-008.10
  Scenario: A history with no close-commit subjects yields no completed set
    Given a branch whose commits were squashed and carry no work item ids
    And no other source identifies the completed items
    When reconstruction reports
    Then the completed set is reported as unestablished rather than as empty

  @AC-SWARM-RUN-008.11
  Scenario: The landing check is reported before anything else
    When reconstruction runs
    Then the branch's landing state is the first verdict in its output
```

## References

- [[frd-swarm-core]] — the command conventions every command here
  carries, and the manifest component each reads a run through
- [[frd-swarm-dispatch]] — the dispatch loop, which sends the work this
  document plans
- `docs/research/swarm-cli/2026-08-25-capability-audit.md` — the audit
  these requirements are drawn from; §7 is the classification that decides
  what earns a command and §8 is the stack rank
- [[module-bead-bonanza]] — the module whose components call this CLI
- [[adr-022]] — projects declare their checks in configuration, which is
  why no check command is hard-coded here
- [[findings-capture-protocol]] — the nine kinds FEAT-SWARM-RUN-006 validates
  and FEAT-SWARM-RUN-007 routes
- [[gherkin-guide]] — the notation these criteria are written in
