---
title: "FRD: Swarm CLI — the dispatch loop"
description: The first wave of the swarm CLI — the commands a coordinator runs
  in the loop that sends work out, takes it back, and reports where the run is
slug: frd-swarm-dispatch
type: frd
status: draft
updated_at: 2026-08-28T00:00:00Z
sources: []
---

# Swarm CLI — the dispatch loop

## Overview

A `bead-bonanza` run is a distributed system whose control plane is
documentation. Its invariants are prose, its state transitions are prose,
and its guards are paragraphs an agent has to remember at the moment it
matters. Measured across the ten sprint-family skills: 79 distinct
capabilities, of which **73 carry a recorded trap** — a way the obvious
invocation produces the wrong result.

This document is the **first wave**: the seven commands that run inside the
dispatch loop. It is deliberately the smaller half. The loop is where the
run spends nearly all of its time, where every measured incident happened,
and where a wrong answer ends a run rather than degrading one — so it is
also the half that can be built, exercised against a real run, and
corrected before anything else is written.

Its value is three things, in this order:

- **Stabilize.** A silent wrong answer becomes a non-zero exit. In four of
  these seven capabilities a *correct* agent still loses: the tracker's
  epic-closing command ignores its positional argument and closes every
  eligible epic in the database at exit 0; an empty parent argument means
  the whole repository rather than nothing; a check written against the
  tracker's owner display passes on an epic where nothing is assigned,
  because that display shows the creator. No amount of care fixes these,
  and **a wrapper is the only way to get a non-zero exit at all.**
- **Simplify.** An invariant enforced by a command is prose that can be
  deleted. Each requirement names the prose it retires.
- **Optimize.** A derived answer an agent assembles by hand costs N queries
  and N results held **in coordinator context** — the scarcest resource in
  a run, since it must survive the whole of it. The same answer as one
  command is one line.

**The order is the grouping, not a ranking.** Two commands manage the run's
task state, four run the dispatch cycle in the order it happens, and the
last is a property the other six carry rather than a capability of its own.

**The second wave** — building a run's shape, reading and validating the
manifest, partitioning and assigning the queue, the readiness gate, claims,
findings and the retro's harvest, and reconstructing a run nobody is
holding — is [[frd-swarm-cli]]. Nothing here waits on it.

### What this wave assumes rather than specifies

**The manifest.** Three of these commands read the run manifest and one
writes to it: `branch` and `base_sha` are written by `dispatch open`, and
`admitted` and the consecutive-failure count are read by `run status` and
`queue front`. The manifest's key set, its types and its validation are
the second wave's. Until that lands, the one rule this wave carries is
that a write here **merges** rather than replaces, because other
components own keys in the same object.

**Judgment.** Classifying a review finding's scope, resolving a rebase
conflict, and deciding whether partial work is finishable stay with an
agent. A command that guessed at them would be worse than the paragraph it
replaced.

**Orchestration.** The CLI spawns no agent, arms no monitor, schedules
nothing, and holds no state of its own. It is called from inside an
existing harness and returns.

**Single commands that already work.** Reading one work item, adding a
comment, creating one — these do what they say and are well documented.
Wrapping them adds a layer, a version and a thing to learn, and hides a
tool the agent already knows behind one it does not.

### Where each command runs

The run's shape **is** the dependency graph, and every command below
attaches to one place in it. The annotation is the requirement that
specifies it; a step with no annotation belongs to the second wave.

```text
run epic
├ pre-flight
├ scope
│   └ AUTHORIZATION GATE
├ groom & plan ─┐
├ TARGET ───────┤ FLAT: tasks only ....... swarm run branch ensure     004
│  ├ dispatch ┐ │  repeat                  swarm phase close           001
│  ├ merge    │ │  until the               swarm queue front           003
│  └ supervise┘ │  front is dry            swarm dispatch open         004
│               │                          swarm dispatch accept       005
│               │                          swarm run status            002
├ review ───────┘
├ retro
└ post-flight ............................ swarm branch verify-landed  006

inside a dispatch, run by the dev agent .. swarm queue front --agent   003
```

Read as a table:

| Lifecycle step | Caller | Commands | Frequency |
|---|---|---|---|
| the sprint branch | coordinator | `run branch ensure` (004) | once; reused on resume |
| every re-query of the front | coordinator | `phase close` (001), `queue front` (003), `run status` (002) | many times per run |
| sending a dispatch out | coordinator | `dispatch open` (004) | once per dispatch |
| inside a dispatch | dev agent | `queue front --agent` (003) | once per work item |
| a dispatch's return | coordinator | `dispatch accept` (005), which ends in `branch verify-landed` (006) | once per dispatch |
| post-flight | coordinator | `branch verify-landed` (006) | once |

### The shape of a command

**`swarm <noun> <verb>`.** The nouns are the things a run has — `run`,
`phase`, `queue`, `dispatch`, `branch` — and the verb is the action.

**The vocabulary is tracker-neutral on purpose.** A unit of work is a
**work item**, not the name one tracker gives it. The tracker in use today
is the only one implemented, and every command goes through one interface
so a second is possible later; nothing in these requirements names a
tracker's own vocabulary except where a measured behaviour is being pinned.

Conventions that hold across every subcommand, each derived from a
recorded trap rather than proposed on taste:

| | |
|---|---|
| `--run <run-epic-id>` | required by every subcommand that reads or mutates run state. An absent or empty scope is refused, never treated as the whole repository |
| `--explain` | print the resolved tracker and version-control invocations in order and run none of them (FEAT-DISP-007) |
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
| **phase epic** | one of the epics chained beneath the run epic, in order: pre-flight, scope, plan, target, review, retro, post-flight |
| **target epic** | the flat phase epic whose task children are the run's work |
| **manifest** | the JSON object in the run epic's metadata; the run's only durable record |
| **dispatch** | one agent working one queue in one worktree on one branch. A dispatch carries one or many work items |
| **dispatch id** | one agent's whole identity — its assignee, its tracker actor, and its branch name. Derived from the run epic id, never minted |
| **ready front** | the complete set of work items in the run that nothing is blocking right now |
| **dispatch front** | the ready front with the run's dispatch filter applied, which is the set that may be handed to agents |
| **drained** | an agent holding nothing ready and nothing in progress. Two queries, because the ready query alone excludes in-progress work |
| **landing check** | one fetch plus one comparison of a local branch against its remote counterpart |
| **derived answer** | an answer requiring a tree walk or a join across the tracker and version control that no single command returns |
| **silent failure** | a command that returns the wrong answer at exit 0 |

## Requirements

### FEAT-DISP-001: Close this run's finished phase epics

**Command:**

```text
swarm phase close --run <run-epic-id> [--dry-run] [--json]
```

**Replaces:** `bead-bonanza/references/supervision.md`, "Re-querying the
ready front is two steps, in order"; `references/dispatching.md`, the
first bullet of "Re-verify at every re-query"; `references/run-record.md`,
"Closing a ceremony epic is `bd epic close-eligible`, scoped BY HAND";
and the paragraph in `bead-bonanza/SKILL.md` beginning "Re-query before
every dispatch".

**Acceptance Criteria:**

```gherkin
@FEAT-DISP-001
Feature: Close this run's finished phase epics

  As a coordinator advancing a run
  I want to close the phase epics this run has finished and only those
  So that the work items they gate enter the ready front without my
  touching another run's epics to do it

  This is the sharpest capability in the inventory because it fails
  destructively in one direction and silently in the other. Wrong one
  way, it closes every eligible epic in the database. Skipped the other
  way, a finished phase never closes, the work items it gates never
  become ready, and an empty front reads as "the run is done" rather than
  "the run has stalled". It runs at every re-query of the front, so both
  failures are reachable many times per run.

  The target epic closes on the same rule as the ceremony phases and is
  not a special case: the chain gates review behind the target, so a
  target that never closes holds the rest of the run exactly as an
  unclosed ceremony phase does.

  Epics outside the run are not the run's business. They are neither
  closed nor reported, because a report naming them invites the next
  reader to act on them.

  @AC-DISP-001.1
  Scenario: Only epics belonging to the named run are touched
    Given eligible epics exist both inside and outside the named run
    When the command closes finished epics for that run
    Then only the epics within that run are closed
    And the epics outside it are left open
    And no epic outside the run appears in the output

  @AC-DISP-001.2
  Scenario: The command names every epic it closed
    When the command closes one or more epics
    Then each closed epic's id appears in the output

  @AC-DISP-001.3
  Scenario: A dry run closes nothing
    When the command runs in preview mode
    Then it reports the epics it would close
    And no epic's status changes

  @AC-DISP-001.4
  Scenario: A run with no eligible epic reports so and succeeds
    Given no epic in the run is eligible to close
    When the command runs
    Then it reports that nothing was eligible
    And it exits zero

  @AC-DISP-001.5 @negative
  Scenario: An unscoped invocation is refused
    When the command is invoked without a run scope
    Then it exits non-zero
    And no epic's status changes

  @AC-DISP-001.6 @negative
  Scenario: An epic with an open child is not closed
    Given an epic in the run carrying one open child
    When the command runs
    Then that epic is left open
    And the open child's id is named as the reason

  @AC-DISP-001.7
  Scenario: The target epic closes on the same rule as a ceremony phase
    Given every task child of the target epic is closed
    When the command runs
    Then the target epic is closed

  @AC-DISP-001.8
  Scenario: The work items a closure released are named
    Given closing an epic makes work items it gated ready
    When the command runs
    Then those work items' ids are reported as newly ready

  @AC-DISP-001.9
  Scenario: A tracker that cannot answer is not reported as nothing eligible
    Given the eligibility query fails
    When the command runs
    Then it reports the query as unanswered
    And it exits non-zero
```

### FEAT-DISP-002: Report a run's status from state

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

**Per-agent drain is reported here and defined in FEAT-DISP-003.** Status
says which agents still hold work; the queue command is where "drained"
means something, because it is the command an agent runs about itself.

**Replaces:** `references/supervision.md`, "Where is this run" and "When
the run is done"; the "The run's health — derived, never asserted" section
of `bead-bonanza/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-DISP-002
Feature: Report a run's status from state

  As a coordinator or a session resuming one
  I want health, completion and per-agent drain answered from version
  control and the tracker
  So that a stalled run is distinguishable from a finished one without
  the coordinator's cooperation

  This is the liveness backbone and the largest token saving in the
  wave, because every part of it is a derived answer. A run once read
  healthy through a five-hour outage in which nothing was dispatched,
  because health was a label written by the component that had stopped.
  Completion needs a walk the tracker does not offer: three queries look
  like they answer "every work item in this run" and each answers
  something else.

  @AC-DISP-002.1
  Scenario: A stall is reported as the measurements that produced it
    Given no worktrees are standing
    And the ready front is not empty
    And the last commit is older than the stall threshold
    Then the report states the standing worktree count, the timestamp and age of the last commit, and the depth of the ready front
    And it states the threshold it compared the commit age against
    And the run's health is reported as stalled
    And the command exits non-zero

  @AC-DISP-002.2
  Scenario: Health is never read from a stored value
    Given a run epic carrying a health value in its metadata that contradicts its state
    When health is reported
    Then the reported value reflects the state, not the stored value

  @AC-DISP-002.3
  Scenario: Completion counts every work item in the run, however nested
    Given a run whose work items are not direct children of the run epic
    When completion is reported
    Then every work item in the run is counted

  @AC-DISP-002.4
  Scenario: Each agent holding work is named
    Given two dispatches, one of which has drained
    When status is reported
    Then the dispatch still holding work is named
    And the drained one is reported as drained

  @AC-DISP-002.5
  Scenario: A tracker that cannot answer is not reported as an empty answer
    Given a tracker query that fails
    When status is reported
    Then the affected line is reported as unknown
    And the command exits non-zero

  @AC-DISP-002.6
  Scenario: What the report cannot observe is named
    When health is reported
    Then the report states that whether an agent is still working is not observable
    And it names the three measurements the classification rests on instead

  @AC-DISP-002.7
  Scenario: The live queue is compared against the admitted set
    Given work items present in the run that the manifest's admitted list does not name
    When status is reported
    Then those work items are named as additions to the run's scope

  @AC-DISP-002.8
  Scenario: Work items excluded at admission are reported beside the completion verdict
    Given a run whose admitted work is all closed and whose admission excluded three items
    When completion is reported
    Then the run is reported as done over a narrowed scope naming those three

  @AC-DISP-002.9
  Scenario: A blocked work item whose worktree still stands is reported as unresolved
    Given a work item marked blocked whose dispatch's worktree still exists
    When status is reported
    Then that work item is reported as an unresolved conflict holding its evidence
```

### FEAT-DISP-003: Answer what may be dispatched now, and what an agent still holds

**Command:**

```text
swarm queue front --run <id> [--agent <dispatch-id>] [--json]
```

**One command, two callers.** Without `--agent` it is the coordinator's
dispatch front: what may go out now. With `--agent` it is one dispatch's
own queue in order, plus that agent's drain verdict — which is what a dev
agent loops on, and what a coordinator asks about an agent it is waiting
for. They are the same query with one more filter, so the filter set and
its four traps are implemented and tested once.

**Four recorded traps, and none of them is the query being long.** An
empty parent argument returns the whole repository rather than nothing. An
assignee filter alone spans the whole tracker, so work outside this run
reads as the agent's. Drain needs two queries, because the ready query
excludes in-progress work and so reports an agent mid-item as finished.
And a payload that is not a list, counted with a naive length, answers for
whatever it is handed — an error body read back unguarded reports a queue
as long as its own text.

**"Drained" and "never given anything" are different answers**, and the
command separates them. An agent with nothing ready and nothing in
progress has finished; an agent with nothing assigned to it in this run at
all was never handed a queue, which is a dispatch error and not a
completion.

**Replaces:** `references/dispatching.md`, "Dispatch continuously",
"Circuit breaker" and "What an agent is holding, and when it is done — as
a query"; the queue-resolution preamble in `run-batch/SKILL.md`; the
`--exclude-label human` paragraph in `references/run-record.md`; the
fallback queries in `references/supervision.md`, "Running without the
hooks".

**Acceptance Criteria:**

```gherkin
@FEAT-DISP-003
Feature: Answer what may be dispatched now, and what an agent still holds

  As a coordinator deciding what to send out, or an agent asking what is
  left in its own queue
  I want the front returned with the run's filter applied, narrowable to
  one dispatch, and systemic failure detected
  So that I never hand an agent work meant for a human, work already
  finished, or work somebody else is holding, and I stop when the next
  dispatch would meet the same fault

  The filter is a convention rather than a default: excluding work
  awaiting a human answer belongs on the dispatch query and nowhere else,
  since a completion count that hides it reports a run as drained while a
  human still owes it an answer.

  @AC-DISP-003.1
  Scenario: A work item awaiting a human is excluded from the dispatch front
    Given a work item in the run labelled for human attention
    When the dispatch front is queried
    Then that work item is not in it

  @AC-DISP-003.2
  Scenario: The same work item is included in a completion count
    Given the same work item
    When the run's outstanding work is counted
    Then that work item is counted

  @AC-DISP-003.3 @negative
  Scenario: An empty or missing run scope is refused
    When the front is queried with no run scope
    Then the command exits non-zero
    And no work items are returned

  @AC-DISP-003.4
  Scenario: Consecutive blocked or failed dispatches stop the run
    Given the configured number of consecutive dispatches have returned blocked or failed
    When the front is queried
    Then the command reports the run as stopped
    And exits non-zero

  @AC-DISP-003.5 @negative
  Scenario: Epics are never in the dispatch front
    Given a run whose phase epics are open
    When the dispatch front is queried
    Then no epic is in it

  @AC-DISP-003.6 @negative
  Scenario: A closed or deferred work item is not in the dispatch front
    Given one closed and one deferred work item in the run
    When the dispatch front is queried
    Then neither is in it

  @AC-DISP-003.7
  Scenario: A work item a live dispatch is already holding is not offered again
    Given a work item assigned to a dispatch whose worktree is standing
    When the dispatch front is queried
    Then that work item is reported as held rather than as available

  @AC-DISP-003.8
  Scenario: The front reaches work items nested below the run epic
    Given a run whose work items are children of the target epic rather than of the run epic
    When the dispatch front is queried
    Then those work items are in it

  @AC-DISP-003.9
  Scenario: The consecutive-failure count survives a resumed run
    Given a run that recorded two consecutive blocked dispatches before its session ended
    When a resumed session queries the front
    Then the count is reported as two rather than zero

  @AC-DISP-003.10
  Scenario: The front is compared against the admitted set
    Given a work item in the front that the manifest's admitted list does not name
    When the front is queried
    Then that work item is named as an addition to the run's scope

  @AC-DISP-003.11
  Scenario: An agent's own queue is the same query narrowed to that agent
    Given a dispatch holding three work items in a collision chain
    When the front is queried for that dispatch
    Then only work items assigned to that dispatch are returned
    And they are returned in the order their dependencies allow

  @AC-DISP-003.12
  Scenario: An agent working an item is not reported as drained
    Given an agent holding one work item in progress and none ready
    When the front is queried for that agent
    Then it is reported as still holding work

  @AC-DISP-003.13
  Scenario: An agent's queue never reaches outside the run
    Given the same dispatch id assigned to a work item in another run
    When the front is queried for that agent
    Then that work item is not returned

  @AC-DISP-003.14
  Scenario: An agent that was never given a queue is not reported as drained
    Given a dispatch id with no work item assigned to it anywhere in the run
    When the front is queried for that agent
    Then it is reported as never assigned rather than as drained
    And the command exits non-zero

  @AC-DISP-003.15
  Scenario: A payload that is not a list is not counted as a queue
    Given a tracker response that is an error body rather than a list
    When the front is queried
    Then the queue is not reported as having that response's length
    And the command exits non-zero
```

### FEAT-DISP-004: Open a dispatch — the sprint branch and its worktree

**Commands:**

```text
swarm dispatch open     --run <id> --dispatch <dispatch-id> [--base <ref>] [--json]
swarm run branch ensure --run <id> [--name <branch>] [--json]
```

**One step with two entry points, not two mechanisms.** Ensuring the
sprint branch is the first thing `dispatch open` does, and it is
idempotent: the first dispatch of a run creates the branch and records it,
every dispatch after that reuses what the manifest holds, and a resumed
run lands on the branch the original wrote. Folding it in is what makes
the timing rule structural — the branch cannot be cut before grooming has
committed, because nothing calls for it until the first dispatch does.

`run branch ensure` is that same step invoked alone. It earns its own
entry point for two callers: a run whose authorization set says to open
the pull request as a draft **at the start** needs the branch before any
dispatch exists, and a coordinator that wants to fail fast on a branch
name already taken would rather learn it before spawning anything.

**This requirement is the open side only.** Removing a worktree and
deleting its branch belong to `swarm dispatch accept`, because removal is
the final act of accepting and its guard is the merge that precedes it.
Splitting that pair across two requirements is what let a removal run
after a merge exited 128.

**The base is an argument, which is what retires a blocking pre-flight
gate.** Under the harness's own worktree isolation the base came from a
global setting no repository owns, the path was chosen for you, and the
only available check was an assertion after the fact. With the coordinator
creating the worktree the base is named in the command and there is
nothing to verify.

**Replaces:** `references/worktree-mechanics.md`, "The sprint branch" and
"The base-ref trap" in full, and step 1 of "The worktree lifecycle"; the
`BB_WORKTREE_BASEREF` block in `pre-flight-check.sh`; the "Hard gate"
paragraph and the sprint-branch section of `bead-bonanza/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-DISP-004
Feature: Open a dispatch — the sprint branch and its worktree

  As a coordinator isolating a dispatch
  I want the sprint branch and each dispatch's worktree created from a
  base I name
  So that a worktree can never be cut from the wrong base

  This is the largest single reduction in prose in the wave, and it
  depends on the coordinator creating the worktree rather than the
  harness. Where the harness creates it, the base comes from a global
  setting no repository owns, the path is chosen for you, and the only
  available check is an assertion after the fact. Where the coordinator
  creates it, the base is an argument and there is nothing to verify.

  @AC-DISP-004.1
  Scenario: A worktree is created from the base the caller names
    When a worktree is created for a dispatch
    Then its head is the named base

  @AC-DISP-004.2
  Scenario: An existing sprint branch is reused rather than recreated
    Given a manifest already recording a sprint branch
    When the branch is ensured
    Then the recorded branch is checked out
    And no new branch is created

  @AC-DISP-004.3
  Scenario: A branch name already taken by unrelated work stops the run
    Given the intended sprint branch name exists and is unrelated
    When the branch is ensured
    Then the command exits non-zero
    And no branch is created or moved

  @AC-DISP-004.4
  Scenario: One worktree and one branch are created per dispatch
    When a worktree is created for a dispatch carrying several work items
    Then exactly one branch is created
    And it is named for the dispatch

  @AC-DISP-004.5
  Scenario: The base sha is recorded when the sprint branch is created
    When the sprint branch is created
    Then the manifest records the sha it was created from

  @AC-DISP-004.6
  Scenario: Opening the first dispatch of a run ensures the sprint branch
    Given a run whose manifest records no sprint branch
    When a dispatch is opened
    Then the sprint branch is created and recorded before the worktree is created

  @AC-DISP-004.7
  Scenario: A manifest write preserves the keys it did not set
    Given a manifest carrying keys written by other components
    When the branch and base sha are recorded
    Then the other keys retain their values

  @AC-DISP-004.8 @negative
  Scenario: A worktree is never created inside the repository's working tree
    When a worktree is created for a dispatch
    Then its path is outside the repository's working tree

  @AC-DISP-004.9
  Scenario: The worktree path and branch name follow from the dispatch id alone
    Given two sessions holding the same run epic id and the same dispatch id
    When each computes the worktree path and branch name
    Then both compute the same values

  @AC-DISP-004.10
  Scenario: Stale worktree registrations are cleared before the list is read
    Given a registered worktree whose directory no longer exists
    When a dispatch is opened
    Then the stale registration is cleared before any decision reads the list

  @AC-DISP-004.11
  Scenario: The base defaults to the sprint branch as it stands now
    Given no base is named
    When a worktree is created for a dispatch
    Then its head is the sprint branch's head read at that moment

  @AC-DISP-004.12
  Scenario: A dispatch whose worktree already exists is not recreated
    Given a dispatch whose worktree exists and holds uncommitted work
    When the dispatch is opened again
    Then the existing worktree is reported and left untouched

  @AC-DISP-004.13 @negative
  Scenario: No global setting decides the base
    When a worktree is created for a dispatch
    Then no configuration outside the repository is read to determine its base
```

### FEAT-DISP-005: Accept a returned dispatch

**Command:**

```text
swarm dispatch accept --run <id> --dispatch <dispatch-id> [--no-push] [--dry-run] [--json]
```

`accept` is the close ceremony the coordinator performs on a dispatch's
behalf, and it is one command because the steps are ordered guards on each
other rather than a sequence a caller may reorder:

```text
1  drain check          the dispatch holds nothing ready and nothing in progress
2  pre-merge acceptance the acceptance signals per work item, deferring the two a merge decides
3  name the case        serialized or concurrent, before the refusal is read
4  merge                fast-forward only; in the concurrent case rebase onto the target read now, then fast-forward
5  ancestry guard       the branch is an ancestor of the target
6  remove               the worktree, then the branch
7  post-merge acceptance the same signals with none deferred
8  land                 push the sprint branch and prove it landed (FEAT-DISP-006)
```

**Four things stop an acceptance, not one.** A rebase conflict is the one
the prose describes; the others are a dispatch that has not drained, a
work item failing an acceptance signal at exit 2, and a fast-forward
refused in the serialized case. Each leaves the worktree standing, because
a standing worktree is the in-flight signal.

**Replaces:** `references/worktree-mechanics.md`, "The worktree lifecycle"
steps 2 through 5 and the whole of "Merging"; `references/dispatching.md`,
"Verify against state, never against testimony" step 3 and "The push, and
proving it landed"; the paragraphs in `bead-bonanza/SKILL.md` beginning
"YOU merge, as each dispatch returns" and "Removal goes behind an ancestry
check".

**Acceptance Criteria:**

```gherkin
@FEAT-DISP-005
Feature: Accept a returned dispatch

  As a coordinator accepting a dispatch that has returned
  I want the drain check, the merge, the refusal reading, the removal and
  the landing proof to be one guarded act
  So that a worktree is never removed while its branch is unmerged

  Two failures make this one action rather than eight. A fast-forward
  refusal means opposite things depending on whether the dispatches were
  serialized or concurrent, and reading it wrong stops a healthy run.
  And ordering the removal after the merge with a shell conjunction is
  not a guard: a merge that exited non-zero was followed by a removal
  that succeeded, leaving a removed worktree beside an unmerged branch
  for ninety seconds, which is the one state the in-flight signal
  asserts cannot occur.

  Resolving a conflict is not this command's. It records the conflict,
  blocks the work item and leaves the evidence standing so an agent
  holding that item's context can be dispatched into the worktree.

  @AC-DISP-005.1
  Scenario: A worktree is not removed unless its branch is an ancestor of the target
    Given a dispatch branch that is not an ancestor of the target after the merge step
    When the command is asked to accept that dispatch
    Then the worktree is left in place
    And the branch is not deleted
    And the command exits non-zero

  @AC-DISP-005.2
  Scenario: The refusal case is named before the refusal is acted on
    Given a fast-forward is refused
    When the command reports the refusal
    Then it states whether the branch was cut from the current target or from an earlier one

  @AC-DISP-005.3
  Scenario: A concurrent branch is rebased onto the target read at merge time
    Given a dispatch branch cut from an earlier target than the current one
    When the command accepts that dispatch
    Then the branch is rebased onto the target as it stands at that moment
    And the resulting merge is a fast-forward

  @AC-DISP-005.4
  Scenario: A rebase conflict blocks the work item and preserves the evidence
    Given a rebase that conflicts
    When the command reports it
    Then a sequencing finding is recorded naming both work items and the conflicting files
    And the unmerged work item is marked blocked
    And its worktree is left standing

  @AC-DISP-005.5
  Scenario: A serialized branch that cannot fast-forward stops rather than retries
    Given dispatches were serialized and a fast-forward is refused
    When the command reports it
    Then it exits non-zero without retrying the merge
    And the merge is never re-attempted with fast-forward-only relaxed

  @AC-DISP-005.6 @negative
  Scenario: A step does not run when the step before it failed
    Given the merge step exits non-zero
    When the command reports the failure
    Then no worktree is removed
    And no branch is deleted
    And no push is attempted

  @AC-DISP-005.7
  Scenario: A conflict report names every conflicting path
    Given a rebase that conflicts across several files
    When the command reports it
    Then every conflicting path is named in the output

  @AC-DISP-005.8
  Scenario: A dispatch whose agent has not drained is not accepted
    Given the dispatch id still holds a work item that is ready or in progress
    When acceptance is attempted
    Then the command exits non-zero naming what the agent still holds
    And nothing is merged or removed

  @AC-DISP-005.9 @negative
  Scenario: The dispatch branch is not deleted before the fast-forward succeeds
    Given a rebase that completed and a fast-forward that has not yet run
    When the command is interrupted at that point
    Then the dispatch branch still exists

  @AC-DISP-005.10 @negative
  Scenario: A worktree is never force-removed
    Given a worktree whose removal is refused by the version control tool
    When the command reports it
    Then the removal is not retried with force
    And the command exits non-zero

  @AC-DISP-005.11
  Scenario: Every work item is re-checked after the merge and the removal
    Given a merge and a removal that both succeeded
    When acceptance completes
    Then each work item in the dispatch is checked again with no signal deferred

  @AC-DISP-005.12
  Scenario: The push runs only after every work item has passed its post-merge check
    Given one work item in the dispatch fails its post-merge check
    When the command reports it
    Then no push is attempted
    And the failing signal is named

  @AC-DISP-005.13
  Scenario: A dry run performs every read and no write
    When acceptance runs in preview mode
    Then it reports the ordered steps it would take
    And no merge, removal, deletion or push occurs
```

### FEAT-DISP-006: Prove a push landed

**Commands:**

```text
swarm branch verify-landed [--branch <name>] [--remote <name>] [--json]
swarm branch push          [--branch <name>] [--remote <name>] [--json]
```

`push` is `verify-landed` with the push in front of it. Both are usable
with **no run scope**, which is what lets a recovering session run the
landing check as its first act, before it knows which run it is in.
`swarm dispatch accept` ends by calling `push`.

**Replaces:** `references/dispatching.md`, "The push, and proving it
landed"; step 1 of `sprint-recovery/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-DISP-006
Feature: Prove a push landed

  As a coordinator finishing a dispatch or a run
  I want the branch compared against its remote after every push
  So that a rejected push is not indistinguishable from a finished run

  This scores lowest in the wave on prose deleted and highest on failure
  removed, because a rejected push leaves every other signal reading
  finished: the tree is clean, no worktrees stand, the work items are
  closed in the tracker. One fetch and one comparison is the only thing
  that separates work that landed from work that is stranded on a
  container about to be reclaimed.

  @AC-DISP-006.1
  Scenario: Unpushed commits are reported as unpushed
    Given the local branch is ahead of its remote
    When the landing check runs
    Then it names the number of unpushed commits
    And it exits non-zero

  @AC-DISP-006.2
  Scenario: A branch with no remote counterpart is not reported as landed
    Given the branch has no counterpart on the remote
    When the landing check runs
    Then it exits non-zero

  @AC-DISP-006.3
  Scenario: A rejected push is surfaced with its cause
    Given a push that the remote rejects
    When the command reports the result
    Then the rejection's cause is included in the output

  @AC-DISP-006.4
  Scenario: A landed branch reports clean
    Given every local commit is present on the remote
    When the landing check runs
    Then it exits zero

  @AC-DISP-006.5
  Scenario: The remote is read afresh before the comparison
    Given a stale local record of where the remote branch points
    When the landing check runs
    Then the remote is fetched before the comparison is made

  @AC-DISP-006.6
  Scenario: The landing check needs no run scope
    Given a checkout with no run epic known
    When the landing check runs
    Then it reports the branch's landing state

  @AC-DISP-006.7
  Scenario: A push refused for one path is reported as a whole rejected push
    Given a remote that refuses one of the paths in the push
    When the command reports the result
    Then no commit in that push is reported as landed
    And the refused path is named
```

### FEAT-DISP-007: Show the commands behind every answer

**Flag, carried by every subcommand in this document:**

```text
swarm <noun> <verb> ... --explain
```

**Foundational rather than ranked.** It is a property each of the six
capabilities above carries, specified once here rather than repeated as a
criterion in every one of them. The second wave's commands carry it too,
and cite this requirement rather than restating it.

An agent that can see the invocation behind a line can re-run it, argue
with it, and work around it when the underlying tool changes. An agent
that cannot has to trust a wrapper it did not write, which is the one
thing this CLI must not require. This is also the direct remedy for the
evidence rule in `AGENTS.md`: a claim that prints the command that
established it does not need a paragraph telling an agent to measure it.

**Replaces:** the "prints the command that established it" paragraphs in
`references/scripts.md` and `references/dispatching.md`, by making the
property universal instead of per-script.

**Acceptance Criteria:**

```gherkin
@FEAT-DISP-007
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

  @AC-DISP-007.1
  Scenario: An explained invocation runs nothing
    When a subcommand is invoked in explain mode
    Then it prints the tool invocations it would run, in order, with every argument resolved
    And no tool is executed

  @AC-DISP-007.2
  Scenario: A normal run prints the invocations as it executes them
    When a subcommand runs normally
    Then each tool invocation is printed as it is executed

  @AC-DISP-007.3
  Scenario: Every reported line carries the command that established it
    When a subcommand reports a fact
    Then the line names the command a reader can re-run to reproduce it

  @AC-DISP-007.4
  Scenario: A step's exit status is read from the step itself
    Given a step whose command fails while a reader consumes its output
    When the step's result is reported
    Then the step is reported as failed

  @AC-DISP-007.5
  Scenario: A step that succeeds while its reader stops early is not reported as failed
    Given a step whose reader stops consuming output before the command finishes
    When the step's result is reported
    Then the step is reported as succeeded

  @AC-DISP-007.6 @negative
  Scenario: The output and error streams of a tool are never merged
    Given a tool that writes a warning to its error stream and data to its output stream
    When the subcommand parses the data
    Then the warning is not part of what is parsed

  @AC-DISP-007.7
  Scenario: Explain output is a runnable transcript
    Given a subcommand invoked in explain mode
    When its printed invocations are run by hand in order
    Then they perform what the subcommand would have performed
```

## References

- [[frd-swarm-cli]] — the second wave: the run's shape, its manifest, the
  partition and readiness gate, claims, findings and recovery
- `docs/research/swarm-cli/2026-08-25-capability-audit.md` — the audit
  these requirements are drawn from; §7 is the classification that decides
  what earns a command and §8 is the stack rank
- [[module-bead-bonanza]] — the module whose components call this CLI
- [[gherkin-guide]] — the notation these criteria are written in
