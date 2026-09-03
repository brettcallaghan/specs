---
title: "FRD: Swarm CLI — the dispatch loop"
description: The dispatch loop of the swarm CLI — the commands a coordinator
  runs to send work out, take it back, and prove the result landed
slug: frd-swarm-dispatch
type: frd
status: approved
updated_at: 2026-08-28T00:00:00Z
sources: []
---

# Swarm CLI — the dispatch loop

## Overview

A `bead-bonanza` run is a distributed system whose control plane is
documentation. Its invariants are prose, its state transitions are prose,
and its guards are paragraphs an agent has to remember at the moment it
matters. Nearly every capability that prose describes carries a **trap** —
a way the obvious invocation produces the wrong result. The inventory
behind that claim, and the counts at the time it was taken, are in the
capability audit this document's References name.

This document is **the dispatch loop**: the five commands that send work
out and take it back. The loop is where the run spends nearly all of its
time, where every measured incident happened, and where a wrong answer
ends a run rather than degrading one.

**It is built on [[frd-swarm-core]]**, which specifies the command
conventions every command here carries, the manifest component every
command here reads a run through, and `swarm status`. Nothing here is
buildable before that.

Its value is three things, in this order:

- **Stabilize.** A silent wrong answer becomes a non-zero exit. In three
  of these five a *correct* agent still loses: the tracker's epic-closing
  command ignores its positional argument and closes every eligible epic
  in the database at exit 0; an empty parent argument returns the whole
  repository rather than nothing; and a worktree removal chained after a
  merge with a shell conjunction runs even when the merge exited 128. No
  amount of care fixes these, and **a wrapper is the only way to get a
  non-zero exit at all.**
- **Simplify.** An invariant enforced by a command is prose that can be
  deleted. Each requirement names the prose it retires.
- **Optimize.** A derivation an agent performs by hand costs N queries
  and N results held **in coordinator context** — the scarcest resource in
  a run, since it must survive the whole of it. The same answer as one
  command is one line.

**The order is the cycle, not a ranking.** Closing what a run has
finished is what refills the ready set; then reading it, opening a
dispatch, accepting one back, and proving the result landed.

### What this document assumes rather than specifies

**The manifest, through the component [[frd-swarm-core]] specifies.**
These commands do not read the run epic's metadata for themselves; they
go through the one manifest reader, whose behaviour on an absent key, a
wrong type and a concurrent write that document defines. What they touch
is small and fixed:

| key | read by | written by |
|---|---|---|
| `branch` | `dispatch open` | `dispatch open`, once |
| `base_sha` | — | `dispatch open`, once |
| `admitted` | `ready` | — |
| `consecutive_dispatch_failures` | `ready` | `ready` |

**`consecutive_dispatch_failures` is the one key this document adds**, and
adding it is a schema change rather than a convention: the manifest's key
set is closed, an undeclared key is refused rather than stored, and a
key's name, type and stage all come from one declaration. So the key is
declared in `schemas/swarm/manifest.strata`'s `DispatchKeys` struct — an
optional `Int` alongside `branch` and `base_sha`, because it first has a
value when the run first dispatches. What it counts is settled in
FEAT-SWARM-DISP-002; the record it lives in is [[frd-swarm-core]]'s.

**Absent is not zero, and `ready` is what reconciles them.** The
manifest reader reports an unwritten key as absent rather than as an empty
value, which is what makes a resumed run's count recoverable at all. A
ready set query treats an absent count as zero — correct for a run that has
dispatched nothing — and it is the write, not the read, that makes the
count survive.

**The run's own structure**, which is [[frd-swarm-cli]]'s: building the
run's shape, partitioning and gating the queue, the readiness gate,
claims, findings, and reconstructing a run nobody is holding. Nothing
here waits on any of it.

**Judgment.** Classifying a review finding's scope, resolving a rebase
conflict, and deciding whether partial work is finishable stay with an
agent. A command that guessed at them would be worse than the paragraph it
replaced.

**Orchestration.** The CLI spawns no agent, arms no monitor, schedules
nothing, and holds no state of its own. It is called from inside an
existing harness and returns.

**Single commands that already work.** Reading one bead, adding a
comment, creating one — these do what they say and are well documented.
Wrapping them adds a layer, a version and a thing to learn, and hides a
tool the agent already knows behind one it does not.

### Where each command runs

The run's shape **is** the dependency graph, and every command below
attaches to one place in it. The annotation is the requirement that
specifies it; a step with no annotation is specified elsewhere in the
family.

```text
run epic
├ pre-flight
├ scope
│   └ AUTHORIZATION GATE
├ groom & plan ─┐
├ TARGET ───────┤ FLAT: tasks only
│  ├ dispatch ┐ │  repeat                  swarm phase close           001
│  ├ merge    │ │  until the               swarm ready                 002
│  └ supervise┘ │  ready set empties       swarm dispatch open         003
│               │                          swarm dispatch accept       004
│               │                          swarm status            core
├ review ───────┘
├ retro
└ post-flight ............................ swarm branch verify-landed  005

inside a dispatch, run by the dev agent .. swarm ready --agent   002
```

`core` marks the one step in this loop that [[frd-swarm-core]] specifies.

Read as a table:

| Lifecycle step | Caller | Commands | Frequency |
|---|---|---|---|
| the sprint branch | coordinator | the first `dispatch open` (003) creates it | once; reused on resume |
| every re-query of the ready set | coordinator | `phase close` (001), `ready` (002), and `status` from the core document | many times per run |
| sending a dispatch out | coordinator | `dispatch open` (003) | once per dispatch |
| inside a dispatch | dev agent | `ready --agent` (002) | once per bead |
| a dispatch's return | coordinator | `dispatch accept` (004), which ends in `branch verify-landed` (005) | once per dispatch |
| post-flight | coordinator | `branch verify-landed` (005) | once |

### The shape of a command

**`swarm <noun> <verb>`, with one bare verb.** The nouns this document
uses are `phase`, `dispatch` and `branch`. `swarm ready` carries none,
for the reason [[frd-swarm-core]] gives for `swarm status`: a segment
before the verb has to carry information, and every command here concerns
one run's graph, so naming that graph would say nothing. The word is the
tracker's own — `bd ready` answers the same question over the whole
repository that this answers over one run.

**The conventions are [[frd-swarm-core]]'s** and are not restated here.
`--explain`, `--dry-run`, `--json` and `--full-result` are carried by
every command in the CLI, so no signature below repeats them; `--run` is
shown, because whether a command needs a run scope at all is this
document's business and one command here does not. The four exit statuses
are core's too, and the criteria below name the one each condition takes
rather than saying only that it is non-zero.

**What these commands print is [[swarm-output-standards]]'s.** The answer
is a mapping of the properties the command established, rendered as YAML
and as JSON under `--json`, with the provenance behind `--full-result`;
a failing invocation renders a verdict per property instead. So a
criterion below saying a command *reports* something is saying that thing
is a property of the answer — not that it appears in prose the caller has
to parse.

**The vocabulary is the graph's, not the tracker's**, as core's is: a
unit of work is a **bead**, one vertex in the dependency graph a run
dispatches from. Another tracker's issue, story or ticket maps onto a
bead.

## Terminology

| Term | Meaning |
|---|---|
| **run** | one execution of `bead-bonanza` over one run epic |
| **run epic** | the bead whose id is the run id and whose metadata is the run manifest |
| **bead** | one unit of work, and one vertex in the graph a run dispatches from. The graph is the model; another tracker's issue, story or ticket maps onto a bead |
| **phase epic** | one of the epics chained beneath the run epic, in order: pre-flight, scope, plan, target, review, retro, post-flight |
| **target epic** | the flat phase epic whose task children are the run's work |
| **manifest** | the JSON object in the run epic's metadata; the run's only durable record |
| **dispatch** | the act of handing a batch of beads to a subagent, which works them in its own worktree on its own branch and hands the branch back. A dispatch carries one or many beads |
| **dispatch id** | one string that is at once the subagent's identity, the assignee on every bead in its batch, its tracker actor, and its branch name. Derived from the run epic id, never minted |
| **ready set** | the complete set of beads in the run that nothing is blocking right now |
| **dispatch set** | the ready set with the run's dispatch filter applied, which is the set that may be handed to agents |
| **drained** | an agent holding nothing ready and nothing in progress. Two queries, because the ready query alone excludes in-progress work |
| **landing check** | one fetch plus one comparison of a local branch against its remote counterpart |
| **derivation** | an answer requiring a tree walk or a join across the tracker and version control that no single command returns |
| **silent failure** | a command that returns the wrong answer at exit 0 |

## Requirements

### FEAT-SWARM-DISP-001: Close this run's finished phase epics

**Command:**

```text
swarm phase close --run <run-epic-id>
```

**Replaces:** `bead-bonanza/references/supervision.md`, "Re-querying the
ready set is two steps, in order"; `references/dispatching.md`, the
first bullet of "Re-verify at every re-query"; `references/run-record.md`,
"Closing a ceremony epic is `bd epic close-eligible`, scoped BY HAND";
and the paragraph in `bead-bonanza/SKILL.md` beginning "Re-query before
every dispatch".

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-DISP-001
Feature: Close this run's finished phase epics

  As a coordinator advancing a run
  I want to close the phase epics this run has finished and only those
  So that the beads they gate enter the ready set without my
  touching another run's epics to do it

  This is the sharpest capability in the inventory because it fails
  destructively in one direction and silently in the other. Wrong one
  way, it closes every eligible epic in the database. Skipped the other
  way, a finished phase never closes, the beads it gates never
  become ready, and an empty ready set reads as "the run is done" rather than
  "the run has stalled". It runs at every re-query of the ready set, so both
  failures are reachable many times per run.

  The target epic closes on the same rule as the ceremony phases and is
  not a special case: the chain gates review behind the target, so a
  target that never closes holds the rest of the run exactly as an
  unclosed ceremony phase does.

  Epics outside the run are not the run's business. They are neither
  closed nor reported, because a report naming them invites the next
  reader to act on them.

  @AC-SWARM-DISP-001.1
  Scenario: Only epics belonging to the named run are touched
    Given eligible epics exist both inside and outside the named run
    When the command closes finished epics for that run
    Then only the epics within that run are closed
    And the epics outside it are left open
    And no epic outside the run appears in the output

  @AC-SWARM-DISP-001.2
  Scenario: The command names every epic it closed
    When the command closes one or more epics
    Then each closed epic's id appears in the output

  @AC-SWARM-DISP-001.3
  Scenario: A dry run closes nothing
    When the command runs in preview mode
    Then it reports the epics it would close
    And no epic's status changes

  @AC-SWARM-DISP-001.4
  Scenario: A run with no eligible epic reports so and succeeds
    Given no epic in the run is eligible to close
    When the command runs
    Then it reports that nothing was eligible
    And it exits zero

  @AC-SWARM-DISP-001.5 @negative
  Scenario: An unscoped invocation is refused
    When the command is invoked without a run scope
    Then it exits with the malformed-invocation status
    And no epic's status changes

  @AC-SWARM-DISP-001.6 @negative
  Scenario: An epic with an open child is not closed
    Given an epic in the run carrying one open child
    When the command runs
    Then that epic is left open
    And the open child's id is named as the reason

  @AC-SWARM-DISP-001.7
  Scenario: The target epic closes on the same rule as a ceremony phase
    Given every task child of the target epic is closed
    When the command runs
    Then the target epic is closed

  @AC-SWARM-DISP-001.8
  Scenario: The beads a closure released are named
    Given closing an epic makes beads it gated ready
    When the command runs
    Then those beads' ids are reported as newly ready

  @AC-SWARM-DISP-001.9
  Scenario: A tracker that cannot answer is not reported as nothing eligible
    Given the eligibility query fails
    When the command runs
    Then it reports the query as unanswered
    And it exits with the stop status
```

### FEAT-SWARM-DISP-002: Answer what may be dispatched now, and what an agent still holds

**Command:**

```text
swarm ready --run <id> [--agent <dispatch-id>]
```

**One command, two callers.** Without `--agent` it is the **whole run's**
dispatchable set — every bead any agent could work right now, across every
dispatch. It is not the coordinator's own work, and the breadth is the
point: planning has already assigned each bead a dispatch id, so an answer
that names them is what tells a coordinator which dispatch ids have work
waiting and no agent standing over it — which is how it decides what to
spawn next. Narrow it and the coordinator can see only what it already
started.

With `--agent` it is one dispatch's own queue in order, plus that agent's
drain verdict — what a dev agent loops on, and what a coordinator asks
about an agent it is waiting for. They are the same query with one more
filter, so the filter set and its traps are implemented and tested once.

**What the breaker counts, because the unit decides whether it fires at
all.** The rule it mechanizes is stated in the skills as *three
consecutive beads returning blocked or failing*, and counted that way it
misfires on its own machinery: when a bead blocks with a dirty tree,
the batch ends and every remaining bead in that dispatch is returned
`blocked — not reached`. One genuine block in a batch of six therefore
produces six consecutive blocked beads and trips a threshold of three,
from a single fault — the opposite of the systemic signal the breaker
exists to catch.

So the unit here is **one dispatch return in which no bead
completed**. A dispatch that completed four of six and blocked two is
evidence the run is working; a dispatch that completed nothing is
evidence the next one will meet the same fault. A bead returned "not
reached" never counts on its own, because it is a record that the batch
stopped rather than that anything was attempted.

**Four recorded traps, and none of them is the query being long.** An
empty parent argument returns the whole repository rather than nothing. An
assignee filter alone spans the whole tracker, so work outside this run
reads as the agent's. Drain needs two queries, because the ready query
excludes in-progress work and so reports an agent mid-bead as finished.
And a payload that is not a list, counted with a naive length, answers for
whatever it is handed — an error body read back unguarded reports a queue
as long as its own text.

**The scope comparison is not this command's either.** A bead present in
the run that the manifest's `admitted` list does not name is a scope
addition, and `swarm status` already reports one — the coordinator asks
that question once per re-query, of the command that measures the whole
run, rather than twice.

**Three readiness signals are deliberately NOT this command's**, and
naming them is what stops the next reader assuming the ready set gates them.
`bd lint`'s findings and a bead's capability tier are settled while the
queue is groomed, not while it is dispatched — a bead reaching the ready set
without an acceptance criterion is a grooming defect, and the ready set is
the wrong place to discover it. **A ready bead carrying no tier is
dispatched at the advanced tier**, which is the resolution rule rather
than a reason to withhold the bead. And the open human-decision queue is
resolved by the end of the run's scope phase: a decision nobody made is
explicitly deferred, and a deferred bead is out of the run. What survives
into this command is the one filter that must hold at dispatch time — a
bead awaiting a human answer is not offered — because a bead can be
labelled for a human after planning gated the queue.

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
@FEAT-SWARM-DISP-002
Feature: Answer what may be dispatched now, and what an agent still holds

  As a coordinator deciding what to send out, or an agent asking what is
  left in its own queue
  I want the ready set returned with the run's filter applied, narrowable to
  one dispatch, and systemic failure detected
  So that I never hand an agent work meant for a human, work already
  finished, or work somebody else is holding, and I stop when the next
  dispatch would meet the same fault

  The filter is a convention rather than a default: excluding work
  awaiting a human answer belongs on the dispatch query and nowhere else,
  since a completion count that hides it reports a run as drained while a
  human still owes it an answer.

  @AC-SWARM-DISP-002.1
  Scenario: A bead awaiting a human is excluded from the dispatch set
    Given a bead in the run labelled for human attention
    When the dispatch set is queried
    Then that bead is not in it

  @AC-SWARM-DISP-002.2
  Scenario: The same bead is included in a completion count
    Given the same bead
    When the run's outstanding work is counted
    Then that bead is counted

  @AC-SWARM-DISP-002.3 @negative
  Scenario: An empty or missing run scope is refused
    When the ready set is queried with no run scope
    Then the command exits with the malformed-invocation status
    And no beads are returned

  @AC-SWARM-DISP-002.4
  Scenario: Consecutive dispatches that completed nothing stop the run
    Given the configured number of consecutive dispatches returned with no bead completed
    When the ready set is queried
    Then the command reports the run as stopped
    And it exits with the stop status

  @AC-SWARM-DISP-002.5
  Scenario: A dispatch that completed some of its work does not advance the breaker
    Given a dispatch that completed four beads and blocked two
    When the ready set is queried
    Then the failure count is reported as zero

  @AC-SWARM-DISP-002.6 @negative
  Scenario: Beads a batch never reached do not advance the breaker
    Given one dispatch in which a bead blocked and five were returned as not reached
    When the ready set is queried
    Then the failure count is reported as one rather than six

  @AC-SWARM-DISP-002.7 @negative
  Scenario: Epics are never in the dispatch set
    Given a run whose phase epics are open
    When the dispatch set is queried
    Then no epic is in it

  @AC-SWARM-DISP-002.8 @negative
  Scenario: A closed or deferred bead is not in the dispatch set
    Given one closed and one deferred bead in the run
    When the dispatch set is queried
    Then neither is in it

  @AC-SWARM-DISP-002.9
  Scenario: A bead a live dispatch is already holding is not offered again
    Given a bead assigned to a dispatch whose worktree is standing
    When the dispatch set is queried
    Then that bead is reported as held rather than as available

  @AC-SWARM-DISP-002.10
  Scenario: The ready set reaches beads nested below the run epic
    Given a run whose beads are children of the target epic rather than of the run epic
    When the dispatch set is queried
    Then those beads are in it

  @AC-SWARM-DISP-002.11
  Scenario: The failure count survives a resumed run
    Given a run that recorded two consecutive dispatches completing nothing before its session ended
    When a resumed session queries the ready set
    Then the count is reported as two rather than zero

  @AC-SWARM-DISP-002.12
  Scenario: An agent's own queue is the same query narrowed to that agent
    Given a dispatch holding three beads in a collision chain
    When the ready set is queried for that dispatch
    Then only beads assigned to that dispatch are returned
    And they are returned in the order their dependencies allow

  @AC-SWARM-DISP-002.13
  Scenario: An agent working a bead is not reported as drained
    Given an agent holding one bead in progress and none ready
    When the ready set is queried for that agent
    Then it is reported as still holding work

  @AC-SWARM-DISP-002.14
  Scenario: An agent's queue never reaches outside the run
    Given the same dispatch id assigned to a bead in another run
    When the ready set is queried for that agent
    Then that bead is not returned

  @AC-SWARM-DISP-002.15
  Scenario: An agent that was never given a queue is not reported as drained
    Given a dispatch id with no bead assigned to it anywhere in the run
    When the ready set is queried for that agent
    Then that dispatch is reported as not established rather than as drained
    And the command exits with the attention status

  @AC-SWARM-DISP-002.16
  Scenario: A payload that is not a list is not counted as a queue
    Given a tracker response that is an error body rather than a list
    When the ready set is queried
    Then the queue is not reported as having that response's length
    And the command exits with the stop status

  @AC-SWARM-DISP-002.17
  Scenario: The ready set spans every dispatch and names each bead's own
    Given a run whose ready beads are assigned across three dispatch ids
    When the dispatch set is queried with no agent named
    Then beads assigned to all three dispatches are returned
    And each is reported with the dispatch id it is assigned to
    And a bead assigned to no dispatch is reported as unassigned rather than omitted
```

### FEAT-SWARM-DISP-003: Open a dispatch — the sprint branch and its worktree

**Command:**

```text
swarm dispatch open --run <id> --dispatch <dispatch-id> [--base <ref>] [--branch <name>]
```

**The sprint branch has one caller, and it is this command.**
Establishing it is the first thing `dispatch open` does, and the step is
idempotent: the first dispatch of a run creates the branch and records it
in the manifest, every dispatch after that is cut from what the manifest
holds, and a resumed run lands on the branch the original wrote. There is
no second entry point, which is what makes the timing rule structural —
the branch cannot be cut before grooming has committed, because nothing
asks for it until the first dispatch does. `--branch` names it on the
invocation that creates it; absent, the name derives from the run epic id,
and once the manifest records a branch the flag is ignored in favour of
the recorded one.

**The manifest is what makes a branch this run's.** A branch carrying the
intended name that the manifest does not record was put there by something
other than this run — an earlier run, a human, a stale checkout — and
nothing in the repository distinguishes which. So the command refuses it
rather than committing a run's work onto it, and the coordinator decides
what to do. The converse is the same rule read forwards: a name the
manifest already records **is** this run's branch, and is reused without a
check.

**A run needing a branch before its first dispatch is not served here.**
Opening the pull request as a draft at the start of the run is that case,
and it wants a branch at a point where the run has nothing to put on one.
No command in this document creates one that early, and one that did would
be cutting from a base the run's own grooming has not yet reached.

**This requirement is the open side only.** Removing a worktree and
deleting its branch belong to `swarm dispatch accept`, because removal is
the final act of accepting and its guard is the merge that precedes it.
Splitting that pair across two requirements is what let a removal run
after a merge exited 128.

**The base is an argument, which is what retires a blocking pre-flight
gate.** Under the harness's own worktree isolation the base comes from a
global setting no repository owns, the path is chosen for you, and the
only available check is an assertion after the fact. With `dispatch open`
creating the worktree the base is named in the command and there is
nothing to verify.

**Three things the harness path costs, and this command takes back.** The
base, above. The path: the harness has been observed placing the worktree
INSIDE the repository, where it reads as uncommitted work in every
`git status` the run takes, so the run has to gitignore a directory it
does not control. And the branch name: the harness names it for its own
agent id, carrying no dispatch id, so `git worktree list` cannot say which
dispatch a tree belongs to and attribution has to go the long way round —
through the bead each tree's HEAD commit closed, and that bead's assignee.
Both are measured, and `bead-bonanza`'s worktree-mechanics reference is
where the measurements are recorded. Naming the branch for the dispatch id
makes the short way sound again; attribution through the commit stays
correct and stays the fallback, and stops being the only thing that
works.

**Replaces:** `references/worktree-mechanics.md`, "The sprint branch" and
"The base-ref trap" in full, and step 1 of "The worktree lifecycle"; the
`BB_WORKTREE_BASEREF` block in `pre-flight-check.sh`; the "Hard gate"
paragraph and the sprint-branch section of `bead-bonanza/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-DISP-003
Feature: Open a dispatch — the sprint branch and its worktree

  As a coordinator isolating a dispatch
  I want the sprint branch and each dispatch's worktree created from a
  base I name
  So that a worktree can never be cut from the wrong base

  This is the largest single reduction in prose in this document, and it
  depends on the coordinator creating the worktree rather than the
  harness. Where the harness creates it, the base comes from a global
  setting no repository owns, the path is chosen for you, and the only
  available check is an assertion after the fact. Where the coordinator
  creates it, the base is an argument and there is nothing to verify.

  @AC-SWARM-DISP-003.1
  Scenario: A worktree is created from the base the caller names
    When a worktree is created for a dispatch
    Then its head is the named base

  @AC-SWARM-DISP-003.2
  Scenario: An existing sprint branch is reused rather than recreated
    Given a manifest already recording a sprint branch
    When a dispatch is opened
    Then the recorded branch is the one the dispatch is cut from
    And no new sprint branch is created

  @AC-SWARM-DISP-003.3
  Scenario: A branch the manifest does not record stops the run
    Given a branch already exists carrying the name this run would use
    And the manifest records no sprint branch
    When a dispatch is opened
    Then the command exits with the attention status naming the existing branch
    And the output names passing an explicit branch name as what would resolve it
    And no branch is created or moved

  @AC-SWARM-DISP-003.4
  Scenario: One worktree and one branch are created per dispatch
    When a worktree is created for a dispatch carrying several beads
    Then exactly one branch is created
    And it is named for the dispatch

  @AC-SWARM-DISP-003.5
  Scenario: The base sha is recorded when the sprint branch is created
    When the sprint branch is created
    Then the manifest records the sha it was created from

  @AC-SWARM-DISP-003.6
  Scenario: Opening the first dispatch of a run creates the sprint branch
    Given a run whose manifest records no sprint branch
    When a dispatch is opened
    Then the sprint branch is created and recorded before the worktree is created

  @AC-SWARM-DISP-003.7 @negative
  Scenario: A worktree is never created inside the repository's working tree
    When a worktree is created for a dispatch
    Then its path is outside the repository's working tree

  @AC-SWARM-DISP-003.8
  Scenario: The worktree path and branch name follow from the dispatch id alone
    Given two sessions holding the same run epic id and the same dispatch id
    When each computes the worktree path and branch name
    Then both compute the same values

  @AC-SWARM-DISP-003.9
  Scenario: Stale worktree registrations are cleared before the list is read
    Given a registered worktree whose directory no longer exists
    When a dispatch is opened
    Then the stale registration is cleared before any decision reads the list

  @AC-SWARM-DISP-003.10
  Scenario: The base defaults to the sprint branch as it stands now
    Given no base is named
    When a worktree is created for a dispatch
    Then its head is the sprint branch's head read at that moment

  @AC-SWARM-DISP-003.11
  Scenario: A dispatch whose worktree already exists is not recreated
    Given a dispatch whose worktree exists and holds uncommitted work
    When the dispatch is opened again
    Then the existing worktree is reported and left untouched

  @AC-SWARM-DISP-003.12 @negative
  Scenario: No global setting decides the base
    When a worktree is created for a dispatch
    Then no configuration outside the repository is read to determine its base
```

### FEAT-SWARM-DISP-004: Accept a returned dispatch

**Command:**

```text
swarm dispatch accept --run <id> --dispatch <dispatch-id> [--no-push]
```

`accept` is the close ceremony the coordinator performs on a dispatch's
behalf, and it is one command because the steps are ordered guards on each
other rather than a sequence a caller may reorder:

```text
1  drain check          the dispatch holds nothing ready and nothing in progress
2  pre-merge acceptance the acceptance signals per bead, deferring the two a merge decides
3  name the case        serialized or concurrent, before the refusal is read
4  merge                fast-forward only; in the concurrent case rebase onto the target read now, then fast-forward
5  ancestry guard       the branch is an ancestor of the target
6  remove               the worktree, then the branch
7  post-merge acceptance the same signals with none deferred
8  land                 push the sprint branch and prove it landed (FEAT-SWARM-DISP-005)
```

**Four things stop an acceptance, not one.** A rebase conflict is the one
the prose describes; the others are a dispatch that has not drained, a
bead failing an acceptance signal at exit 2, and a fast-forward
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
@FEAT-SWARM-DISP-004
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
  blocks the bead and leaves the evidence standing so an agent
  holding that bead's context can be dispatched into the worktree.

  @AC-SWARM-DISP-004.1
  Scenario: A worktree is not removed unless its branch is an ancestor of the target
    Given a dispatch branch that is not an ancestor of the target after the merge step
    When the command is asked to accept that dispatch
    Then the worktree is left in place
    And the branch is not deleted
    And the command exits with the stop status

  @AC-SWARM-DISP-004.2
  Scenario: The refusal case is named before the refusal is acted on
    Given a fast-forward is refused
    When the command reports the refusal
    Then it states whether the branch was cut from the current target or from an earlier one

  @AC-SWARM-DISP-004.3
  Scenario: A concurrent branch is rebased onto the target read at merge time
    Given a dispatch branch cut from an earlier target than the current one
    When the command accepts that dispatch
    Then the branch is rebased onto the target as it stands at that moment
    And the resulting merge is a fast-forward

  @AC-SWARM-DISP-004.4
  Scenario: A rebase conflict blocks the bead and preserves the evidence
    Given a rebase that conflicts
    When the command reports it
    Then a sequencing finding is recorded naming both beads and the conflicting files
    And the unmerged bead is marked blocked
    And its worktree is left standing

  @AC-SWARM-DISP-004.5
  Scenario: A serialized branch that cannot fast-forward stops rather than retries
    Given dispatches were serialized and a fast-forward is refused
    When the command reports it
    Then it exits with the stop status without retrying the merge
    And the merge is never re-attempted with fast-forward-only relaxed

  @AC-SWARM-DISP-004.6 @negative
  Scenario: A step does not run when the step before it failed
    Given the merge step exits non-zero
    When the command reports the failure
    Then no worktree is removed
    And no branch is deleted
    And no push is attempted

  @AC-SWARM-DISP-004.7
  Scenario: A conflict report names every conflicting path
    Given a rebase that conflicts across several files
    When the command reports it
    Then every conflicting path is named in the output

  @AC-SWARM-DISP-004.8
  Scenario: A dispatch whose agent has not drained is not accepted
    Given the dispatch id still holds a bead that is ready or in progress
    When acceptance is attempted
    Then the command exits with the attention status naming what the agent still holds
    And nothing is merged or removed

  @AC-SWARM-DISP-004.9 @negative
  Scenario: The dispatch branch is not deleted before the fast-forward succeeds
    Given a rebase that completed and a fast-forward that has not yet run
    When the command is interrupted at that point
    Then the dispatch branch still exists

  @AC-SWARM-DISP-004.10 @negative
  Scenario: A worktree is never force-removed
    Given a worktree whose removal is refused by the version control tool
    When the command reports it
    Then the removal is not retried with force
    And the command exits with the attention status naming what the tool refused over

  @AC-SWARM-DISP-004.11
  Scenario: Every bead is re-checked after the merge and the removal
    Given a merge and a removal that both succeeded
    When acceptance completes
    Then each bead in the dispatch is checked again with no signal deferred

  @AC-SWARM-DISP-004.12
  Scenario: The push runs only after every bead has passed its post-merge check
    Given one bead in the dispatch fails its post-merge check
    When the command reports it
    Then no push is attempted
    And the failing signal is named

  @AC-SWARM-DISP-004.13
  Scenario: A dry run performs every read and no write
    When acceptance runs in preview mode
    Then it reports the ordered steps it would take
    And no merge, removal, deletion or push occurs
```

### FEAT-SWARM-DISP-005: Prove a push landed

**Commands:**

```text
swarm branch verify-landed [--branch <name>] [--remote <name>]
swarm branch push          [--branch <name>] [--remote <name>]
```

`push` is `verify-landed` with the push in front of it, and
`swarm dispatch accept` ends by calling `push`. A coordinator working the
dispatch loop therefore never invokes either by hand — accepting a
dispatch lands it.

**Three callers reach them directly, and none of the three has a dispatch
to accept.** Post-flight, where the loop is over and the last thing to
establish is that what the run produced is on the remote. A recovering
session, whose first act is the landing check and which does not yet know
which run it is in — which is why both are usable with **no run scope**.
And an acceptance taken with `--no-push`, where the push was deliberately
deferred and something has to perform it afterwards.

**Replaces:** `references/dispatching.md`, "The push, and proving it
landed"; step 1 of `sprint-recovery/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-DISP-005
Feature: Prove a push landed

  As a coordinator finishing a dispatch or a run
  I want the branch compared against its remote after every push
  So that a rejected push is not indistinguishable from a finished run

  This scores lowest in this document on prose deleted and highest on failure
  removed, because a rejected push leaves every other signal reading
  finished: the tree is clean, no worktrees stand, the beads are
  closed in the tracker. One fetch and one comparison is the only thing
  that separates work that landed from work that is stranded on a
  container about to be reclaimed.

  @AC-SWARM-DISP-005.1
  Scenario: Unpushed commits are reported as unpushed
    Given the local branch is ahead of its remote
    When the landing check runs
    Then it names the number of unpushed commits
    And it exits with the attention status

  @AC-SWARM-DISP-005.2
  Scenario: A branch with no remote counterpart is not reported as landed
    Given the branch has no counterpart on the remote
    When the landing check runs
    Then it exits with the attention status

  @AC-SWARM-DISP-005.3
  Scenario: A rejected push is surfaced with its cause
    Given a push that the remote rejects
    When the command reports the result
    Then the rejection's cause is included in the output

  @AC-SWARM-DISP-005.4
  Scenario: A landed branch reports clean
    Given every local commit is present on the remote
    When the landing check runs
    Then it exits zero

  @AC-SWARM-DISP-005.5
  Scenario: The remote is read afresh before the comparison
    Given a stale local record of where the remote branch points
    When the landing check runs
    Then the remote is fetched before the comparison is made

  @AC-SWARM-DISP-005.6
  Scenario: The landing check needs no run scope
    Given a checkout with no run epic known
    When the landing check runs
    Then it reports the branch's landing state

  @AC-SWARM-DISP-005.7
  Scenario: A push refused for one path is reported as a whole rejected push
    Given a remote that refuses one of the paths in the push
    When the command reports the result
    Then no commit in that push is reported as landed
    And the refused path is named
```

## References

- [[frd-swarm-core]] — the command conventions every command here
  carries, the manifest component each reads a run through, and
  `swarm status`
- [[swarm-output-standards]] — what every command here prints: the
  answer, the full result, and the style rules both are rendered under
- [[frd-swarm-cli]] — the run's shape, the partition and readiness gate,
  claims, findings and recovery
- `docs/research/swarm-cli/2026-08-25-capability-audit.md` — the audit
  these requirements are drawn from; §7 is the classification that decides
  what earns a command and §8 is the stack rank
- [[module-bead-bonanza]] — the module whose components call this CLI
- [[gherkin-guide]] — the notation these criteria are written in
