---
title: "FRD: Swarm CLI — the coordinator's tested action surface"
description: A command-line tool that performs the highest-value actions a
  bead-bonanza run takes, so a run's correctness comes from an exit code
  rather than from an agent remembering a paragraph
slug: frd-swarm-cli
type: frd
status: draft
updated_at: 2026-08-28T00:00:00Z
sources: []
---

# Swarm CLI

## Overview

A `bead-bonanza` run is a distributed system whose control plane is
documentation. Its invariants are prose, its state transitions are
prose, and its guards are paragraphs an agent has to remember at the
moment it matters. Measured across the ten sprint-family skills: 79
distinct capabilities, of which **73 carry a recorded trap** — a way the
obvious invocation produces the wrong result.

The traps are not evenly bad, and one class decides this feature's shape.
**In five of the seven highest-value capabilities a correct agent still
loses**: `bd epic close-eligible` ignores its positional argument and
closes every eligible epic in the database at exit 0; `bd update --claim`
on an unassigned bead silently rewrites the assignee; a check written
against `bd show`'s `Owner:` line passes on an epic where not one bead is
assigned, because that line prints `created_by`. No amount of care fixes
these. **A wrapper is the only way to get a non-zero exit at all.**

This feature delivers a CLI that performs those actions. Its value is
three things, in this order:

- **Stabilize.** A silent wrong answer becomes a non-zero exit. This is
  the whole justification for the top of the list, and it costs almost
  no tokens.
- **Simplify.** An invariant enforced by a command is prose that can be
  deleted. Three runs of this repository's own skills were spent writing
  and rewriting that prose.
- **Optimize.** A derived answer an agent assembles by hand costs N
  queries and N results held **in coordinator context** — the scarcest
  resource in a run, since it must survive the whole of it. The same
  answer as one command is one line.

**The order of the requirements below is the order of that value**, taken
from the stack rank in `docs/research/swarm-cli/2026-08-25-capability-audit.md`
§8. It is not arbitrary and it is not alphabetical: FEAT-SWARM-001 through
007 are the seven capabilities that are coordinator-side, run-ending, and
not mechanized today. 008 through 013 are the next tier. **014 and 015 sit
outside that ranking**: 014 is a property every other subcommand carries
rather than a capability of its own, and 015 was separated out of 011
because a claim is an act on a bead and a finding is a record about one.

**Out of scope for this feature.** Each deferred deliberately:

- **Anything already mechanized.** Thirteen capabilities are scripts
  today. Porting them to Go buys portability, one `--explain` convention
  and a test surface; it does not buy a capability the run lacks. That is
  a follow-on feature, and this one must not absorb it.
- **Orchestration.** The CLI spawns no agent, arms no monitor, schedules
  nothing, and holds no state of its own. It is called from inside an
  existing harness and returns.
- **Judgment.** Classifying a review finding's scope, reasoning findings
  into candidates, resolving a rebase conflict, and deciding whether
  partial work is finishable stay with an agent. A command that guessed
  at them would be worse than the paragraph it replaced.
- **A tracker abstraction beyond what these actions need.** Bead
  operations go behind an interface so a second tracker is possible
  later, but this feature ships one implementation and no plugin
  mechanism.
- **Wrapping single commands that already work.** `bd ready`, `bd show`,
  `bd comment`, `bd create` and `git log` do what they say. Wrapping them
  hides a tool the agent knows behind one it does not.

### When each capability runs

The run's shape **is** the dependency graph, and every capability below
attaches to one place in it. The annotation is the command a coordinator
calls at that step; the number is the requirement that specifies it.

```text
run epic ................................. swarm run open              003
├ pre-flight ............................. swarm run manifest set      006
├ scope
│   └ AUTHORIZATION GATE ................. (human; taken by sprint-scope)
├ groom & plan ─┐ ........................ swarm queue reconcile       013
│               │                          swarm bead bounce           013
│               │                          swarm queue partition load  007
│               │                          swarm queue assign          007
│               │                          swarm queue gate            007
├ TARGET ───────┤ FLAT: tasks only ....... swarm run branch ensure     008
│  ├ dispatch ┐ │  repeat                  swarm epic close-finished   001
│  ├ merge    │ │  until the               swarm queue front           010
│  └ supervise┘ │  front is dry            swarm dispatch open         008
│               │                          swarm dispatch accept       002
│               │                          swarm run status            004
├ review ───────┘ ........................ swarm run manifest get      006
├ retro .................................. swarm finding harvest       012
│                                          swarm finding route         012
└ post-flight ............................ swarm branch verify-landed  005

inside a dispatch, run by the dev agent .. swarm bead claim            015
                                           swarm finding record        011
at any point, holding no context ......... swarm run recover           009
```

Read as a table, with how often each runs and who calls it:

| Lifecycle step | Caller | Commands | Frequency |
|---|---|---|---|
| pre-flight | coordinator | `run manifest set` (006) | twice per run |
| scope | `sprint-scope` | `run open` (003), `run manifest set` (006) | once |
| groom & plan | `sprint-plan` | `queue reconcile`, `bead bounce` (013); `queue partition load`, `queue assign`, `queue gate` (007) | once per pass |
| sprint branch | coordinator | `run branch ensure` (008) | once; reused on resume |
| dispatch loop | coordinator | `epic close-finished` (001), `queue front` (010), `dispatch open` (008), `run status` (004) | every re-query |
| a dispatch's return | coordinator | `dispatch accept` (002), which ends in `branch verify-landed` (005) | once per dispatch |
| inside a dispatch | dev agent | `bead claim` (015), `finding record` (011) | once per bead, plus one per finding |
| review | `sprint-review` | `run manifest get` / `set` for the pass counter (006) | once per pass |
| retro | `sprint-retro` | `finding harvest`, `finding route` (012), `run manifest set` (006) | once |
| post-flight | coordinator | `branch verify-landed` (005) | once |
| resume | a fresh session | `run recover` (009) | on every resume |

**Two capabilities have no lifecycle row, and it is not an omission.**
014 is a flag every subcommand above carries. 006 appears at five
different steps because the manifest is the run's only durable record and
every phase writes part of it.

### The shape of a command

**`swarm <noun> <verb>`.** The nouns are the things a run has — `run`,
`epic`, `queue`, `dispatch`, `branch`, `bead`, `finding` — and the verb is
the action. Each requirement below states its own commands before its
criteria.

Conventions that hold across every subcommand, each derived from a trap in
the audit's §3 rather than proposed on taste:

| | |
|---|---|
| `--run <run-epic-id>` | required by every subcommand that reads or mutates run state. An absent or empty scope is refused, never treated as the whole repository |
| `--explain` | print the resolved `bd` and `git` invocations in order and run none of them (FEAT-SWARM-014) |
| `--json` | machine-readable output; the same facts as the text form, never a subset |
| `--dry-run` | on mutating subcommands: perform every read and no write |
| exit `0` | the command's own verdict is clear |
| exit `1` | repairable in place, and the output names the repair |
| exit `2` | stop; not repairable by the caller |
| exit `3` | invocation error — a missing scope, an unknown flag, a malformed argument |

The exit vocabulary is `bb-status.sh`'s, unchanged, because every skill in
the family already routes on it.

## Terminology

| Term | Meaning |
|---|---|
| **run** | one execution of `bead-bonanza` over one run epic |
| **run epic** | the bead whose id is the run id and whose `--metadata` is the run manifest |
| **phase epic** | one of the epics chained beneath the run epic, in order: pre-flight, scope, plan, target, review, retro, post-flight |
| **target epic** | the flat phase epic whose task children are the run's work |
| **manifest** | the JSON object in the run epic's `--metadata`; the run's only durable record |
| **dispatch** | one agent working one queue in one worktree on one branch. A dispatch carries one or many beads |
| **dispatch id** | one agent's whole identity — its `assignee`, its `BEADS_ACTOR`, and its branch name. Derived from the run epic id, never minted |
| **ready front** | `bd ready` over the run epic — the complete set of beads in the run that nothing is blocking right now |
| **dispatch front** | the ready front with the run's dispatch filter applied, which is the set that may be handed to agents |
| **collision edge** | a `blocks` edge between two beads that must not run at once because they share a component or fixture. Not a real dependency; both ends carry `metadata.collision` naming the surface |
| **drained** | an agent holding nothing ready and nothing in progress. Two queries, because the ready query alone excludes in-progress work |
| **bounce** | move a bead out of the run's dispatchable set by labelling it for a human, leaving it open. Not a close and not a deferral |
| **derived answer** | an answer requiring a tree walk or a join across `bd` and `git` that no single command returns |
| **landing check** | one fetch plus one comparison of a local branch against its remote counterpart |
| **silent failure** | a command that returns the wrong answer at exit 0 |

## Requirements

### FEAT-SWARM-001: Close this run's finished phase epics

**Command:**

```text
swarm epic close-finished --run <run-epic-id> [--dry-run] [--json]
```

**Replaces:** `bead-bonanza/references/supervision.md`, "Re-querying the
ready front is two steps, in order"; `references/dispatching.md`, the
first bullet of "Re-verify at every re-query"; `references/run-record.md`,
"Closing a ceremony epic is `bd epic close-eligible`, scoped BY HAND";
and the paragraph in `bead-bonanza/SKILL.md` beginning "Re-query before
every dispatch".

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-001
Feature: Close this run's finished phase epics

  As a coordinator advancing a run
  I want to close the phase epics this run has finished and only those
  So that the beads they gate enter the ready front without my closing
  another run's epics to do it

  This is the sharpest capability in the inventory because it fails
  destructively in one direction and silently in the other. Wrong one
  way, it closes every eligible epic in the database. Skipped the other
  way, a finished phase never closes, the beads it gates never become
  ready, and an empty front reads as "the run is done" rather than "the
  run has stalled". It runs at every re-query of the front, so both
  failures are reachable many times per run.

  The target epic closes on the same rule as the ceremony phases and is
  not a special case: the chain gates review behind the target, so a
  target that never closes holds the rest of the run exactly as an
  unclosed ceremony phase does.

  @AC-SWARM-001.1
  Scenario: Only epics belonging to the named run are closed
    Given eligible epics exist both inside and outside the named run
    When the command closes finished epics for that run
    Then only the epics within that run are closed
    And the epics outside it are left open

  @AC-SWARM-001.2
  Scenario: The command names every epic it closed
    When the command closes one or more epics
    Then each closed epic's id appears in the output

  @AC-SWARM-001.3
  Scenario: A dry run closes nothing
    When the command runs in preview mode
    Then it reports the epics it would close
    And no epic's status changes

  @AC-SWARM-001.4
  Scenario: A run with no eligible epic reports so and succeeds
    Given no epic in the run is eligible to close
    When the command runs
    Then it reports that nothing was eligible
    And it exits zero

  @AC-SWARM-001.5 @negative
  Scenario: An unscoped invocation is refused
    When the command is invoked without a run scope
    Then it exits non-zero
    And no epic's status changes

  @AC-SWARM-001.6 @negative
  Scenario: An epic with an open child is not closed
    Given an epic in the run carrying one open child
    When the command runs
    Then that epic is left open
    And the open child's id is named as the reason

  @AC-SWARM-001.7
  Scenario: Eligible epics outside the run are named rather than passed over silently
    Given an eligible epic outside the named run
    When the command runs
    Then that epic is reported as eligible and out of scope
    And it is not closed

  @AC-SWARM-001.8
  Scenario: The target epic closes on the same rule as a ceremony phase
    Given every task child of the target epic is closed
    When the command runs
    Then the target epic is closed

  @AC-SWARM-001.9
  Scenario: The beads a closure released are named
    Given closing an epic makes beads it gated ready
    When the command runs
    Then those beads' ids are reported as newly ready

  @AC-SWARM-001.10
  Scenario: A tracker that cannot answer is not reported as nothing eligible
    Given the eligibility query fails
    When the command runs
    Then it reports the query as unanswered
    And it exits non-zero
```

### FEAT-SWARM-002: Accept a returned dispatch

**Commands:**

```text
swarm dispatch accept --run <id> --dispatch <dispatch-id> [--no-push] [--dry-run] [--json]
```

`accept` is the close ceremony the coordinator performs on a dispatch's
behalf, and it is one command because the steps are ordered guards on each
other rather than a sequence a caller may reorder:

```text
1  drain check          the dispatch holds nothing ready and nothing in progress
2  pre-merge acceptance the five signals per bead, deferring 3 and 4
3  name the case        serialized or concurrent, before the refusal is read
4  merge                --ff-only; in the concurrent case rebase onto the target read now, then --ff-only
5  ancestry guard       the branch is an ancestor of the target
6  remove               the worktree, then the branch
7  post-merge acceptance the five signals per bead, with none deferred
8  land                 push the sprint branch and prove it landed (FEAT-SWARM-005)
```

**Four things stop an acceptance, not one.** A rebase conflict is the one
the prose describes; the others are a dispatch that has not drained, a
bead failing an acceptance signal at exit 2, and a fast-forward refused in
the serialized case. Each leaves the worktree standing, because a standing
worktree is the in-flight signal.

**Replaces:** `references/worktree-mechanics.md`, "The worktree lifecycle"
steps 2 through 5 and the whole of "Merging"; `references/dispatching.md`,
"Verify against state, never against testimony" step 3 and "The push, and
proving it landed"; the paragraphs in `bead-bonanza/SKILL.md` beginning
"YOU merge, as each dispatch returns" and "Removal goes behind an ancestry
check".

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-002
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
  blocks the bead and leaves the evidence standing so an agent holding
  that bead's context can be dispatched into the worktree.

  @AC-SWARM-002.1
  Scenario: A worktree is not removed unless its branch is an ancestor of the target
    Given a dispatch branch that is not an ancestor of the target after the merge step
    When the command is asked to accept that dispatch
    Then the worktree is left in place
    And the branch is not deleted
    And the command exits non-zero

  @AC-SWARM-002.2
  Scenario: The refusal case is named before the refusal is acted on
    Given a fast-forward is refused
    When the command reports the refusal
    Then it states whether the branch was cut from the current target or from an earlier one

  @AC-SWARM-002.3
  Scenario: A concurrent branch is rebased onto the target read at merge time
    Given a dispatch branch cut from an earlier target than the current one
    When the command accepts that dispatch
    Then the branch is rebased onto the target as it stands at that moment
    And the resulting merge is a fast-forward

  @AC-SWARM-002.4
  Scenario: A rebase conflict blocks the bead and preserves the evidence
    Given a rebase that conflicts
    When the command reports it
    Then a sequencing finding is recorded naming both beads and the conflicting files
    And the unmerged bead is marked blocked
    And its worktree is left standing

  @AC-SWARM-002.5
  Scenario: A serialized branch that cannot fast-forward stops rather than retries
    Given dispatches were serialized and a fast-forward is refused
    When the command reports it
    Then it exits non-zero without retrying the merge
    And the merge is never re-attempted with fast-forward-only relaxed

  @AC-SWARM-002.6 @negative
  Scenario: A step does not run when the step before it failed
    Given the merge step exits non-zero
    When the command reports the failure
    Then no worktree is removed
    And no branch is deleted
    And no push is attempted

  @AC-SWARM-002.7
  Scenario: A conflict report names every conflicting path
    Given a rebase that conflicts across several files
    When the command reports it
    Then every conflicting path is named in the output

  @AC-SWARM-002.8
  Scenario: A dispatch whose agent has not drained is not accepted
    Given the dispatch id still holds a bead that is ready or in progress
    When acceptance is attempted
    Then the command exits non-zero naming what the agent still holds
    And nothing is merged or removed

  @AC-SWARM-002.9 @negative
  Scenario: The dispatch branch is not deleted before the fast-forward succeeds
    Given a rebase that completed and a fast-forward that has not yet run
    When the command is interrupted at that point
    Then the dispatch branch still exists

  @AC-SWARM-002.10 @negative
  Scenario: A worktree is never force-removed
    Given a worktree whose removal is refused by the version control tool
    When the command reports it
    Then the removal is not retried with force
    And the command exits non-zero

  @AC-SWARM-002.11
  Scenario: Every bead is re-checked after the merge and the removal
    Given a merge and a removal that both succeeded
    When acceptance completes
    Then each bead in the dispatch is checked again with no signal deferred

  @AC-SWARM-002.12
  Scenario: The push runs only after every bead has passed its post-merge check
    Given one bead in the dispatch fails its post-merge check
    When the command reports it
    Then no push is attempted
    And the failing signal is named

  @AC-SWARM-002.13
  Scenario: A dry run performs every read and no write
    When acceptance runs in preview mode
    Then it reports the ordered steps it would take
    And no merge, removal, deletion or push occurs
```

### FEAT-SWARM-003: Build a run's epic shape

**Command:**

```text
swarm run open --title <title> --admit <bead-or-epic-id>... [--second-target <epic-id>] [--dry-run] [--json]
```

Returns the run epic id and the target epic id. **The judgment stays with
`sprint-scope`** — which beads to admit, which to exclude, and whether to
adopt an epic the user already owns are decided before this command is
called, and it refuses rather than deciding.

**Replaces:** `sprint-scope/SKILL.md`, "Build the shape, not just the
epic" and its "Assign every phase epic to its own agent id" paragraph;
`references/run-record.md`, "The run's epic shape".

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-003
Feature: Build a run's epic shape

  As a coordinator opening a run
  I want the run epic, its phase chain, its flat target and its
  authorization gate created as one act
  So that a shape that cannot dispatch is refused rather than built

  This runs once per run and still ranks third, on blast radius alone.
  The tracker refuses a task blocking an epic and an epic blocking a
  task, so a phase can only be gated by an epic and the authorization
  gate can only be a task child. Build the shape wrong and nothing
  dispatches, with no error anywhere to say why.

  The shape is the run's schedule, so the criteria below specify it
  rather than leaving it to the implementation: seven phase epics in one
  order, the work as flat task children of the target, and the gate as a
  task child of scope.

  @AC-SWARM-003.1
  Scenario: The created shape dispatches exactly the target's open leaves
    When a run shape is created and the phases before the target are closed
    Then the run's ready front contains exactly the target epic's open task children

  @AC-SWARM-003.2
  Scenario: A partially built shape is not left behind
    Given the creation fails partway
    When the command reports the failure
    Then the run epic and its phases are either all present or all absent

  @AC-SWARM-003.3
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

  @AC-SWARM-003.4
  Scenario: The authorization gate holds the whole chain
    Given a run shape whose authorization gate is unresolved
    When the ready front is queried
    Then it is empty

  @AC-SWARM-003.5
  Scenario: A non-flat target is refused
    When a run shape is created with a sub-epic beneath the target epic
    Then the command exits non-zero naming the offending child

  @AC-SWARM-003.6
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

  @AC-SWARM-003.7
  Scenario: The authorization gate is a task child of the scope epic
    When a run shape is created
    Then the gate is held against a task child of the scope epic rather than against a phase epic

  @AC-SWARM-003.8
  Scenario: Work that decomposes becomes a second target beside the first
    Given an admitted epic whose own children are epics
    When a run shape is created
    Then that epic becomes a second target epic under the run epic
    And it is not nested beneath the first target epic

  @AC-SWARM-003.9 @negative
  Scenario: A run epic that already exists is not rebuilt
    Given a run epic that already carries a phase chain
    When the command is asked to build a shape over it
    Then it exits non-zero
    And no epic is created, moved or reassigned

  @AC-SWARM-003.10
  Scenario: The admission record is written as the shape is created
    When a run shape is created
    Then the manifest records which beads were admitted, which were excluded with a reason, and which were moved and from where

  @AC-SWARM-003.11 @negative
  Scenario: A bead is never moved out of an epic that already owns it
    Given an admitted bead that already belongs to another epic
    When a run shape is created without an explicit instruction to move it
    Then the bead keeps its existing parent
    And it is recorded as excluded with that reason
```

### FEAT-SWARM-004: Report a run's status from state

**Command:**

```text
swarm run status --run <id> [--agent <dispatch-id>] [--stall-after <duration>] [--json]
```

**Health is three measurements and a threshold, and the command reports
all four.** There is no way from a shell to observe whether a subagent is
still thinking: a stalled agent never finishes, so no event fires and
absence of an event is not an event. The three observable inputs are the
number of standing worktrees, the age of the last commit on the sprint
branch, and the depth of the ready front. The command states each with the
value it had, states the threshold it applied, and names the thing it
could not observe — so a reader who disagrees with the classification can
still act on the facts.

**Replaces:** `references/supervision.md`, "Where is this run" and "When
the run is done"; the "The run's health — derived, never asserted" section
of `bead-bonanza/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-004
Feature: Report a run's status from state

  As a coordinator or a session resuming one
  I want health, completion and per-agent drain answered from git and the
  tracker
  So that a stalled run is distinguishable from a finished one without
  the coordinator's cooperation

  This is the liveness backbone and the largest token saving in the
  feature, because every part of it is a derived answer. A run once read
  healthy through a five-hour outage in which nothing was dispatched,
  because health was a label written by the component that had stopped.
  Completion needs a walk the tracker does not offer: three queries look
  like they answer "every bead in this run" and each answers something
  else. Drain needs two queries because the ready query alone excludes
  in-progress work and so reports a live agent as finished.

  @AC-SWARM-004.1
  Scenario: A stall is reported as the measurements that produced it
    Given no worktrees are standing
    And the ready front is not empty
    And the last commit is older than the stall threshold
    Then the report states the standing worktree count, the timestamp and age of the last commit, and the depth of the ready front
    And it states the threshold it compared the commit age against
    And the run's health is reported as stalled
    And the command exits non-zero

  @AC-SWARM-004.2
  Scenario: Health is never read from a stored label
    Given a run epic carrying a health value in its metadata that contradicts its state
    When health is reported
    Then the reported value reflects the state, not the stored value

  @AC-SWARM-004.3
  Scenario: Completion counts every bead in the run, however nested
    Given a run whose work beads are not direct children of the run epic
    When completion is reported
    Then every work bead in the run is counted

  @AC-SWARM-004.4
  Scenario: An agent working a bead is not reported as drained
    Given an agent holding one bead in progress and none ready
    When that agent's drain state is reported
    Then it is reported as still holding work

  @AC-SWARM-004.5
  Scenario: A tracker that cannot answer is not reported as an empty answer
    Given a tracker query that fails
    When status is reported
    Then the affected line is reported as unknown
    And the command exits non-zero

  @AC-SWARM-004.7
  Scenario: What the report cannot observe is named
    When health is reported
    Then the report states that whether an agent is still working is not observable
    And it names the three measurements the classification rests on instead

  @AC-SWARM-004.8
  Scenario: The live queue is compared against the admitted set
    Given beads present in the run that the manifest's admitted list does not name
    When status is reported
    Then those beads are named as additions to the run's scope

  @AC-SWARM-004.9
  Scenario: Beads excluded at admission are reported beside the completion verdict
    Given a run whose admitted beads are all closed and whose admission excluded three beads
    When completion is reported
    Then the run is reported as done over a narrowed scope naming those three

  @AC-SWARM-004.10
  Scenario: A blocked bead whose worktree still stands is reported as unresolved
    Given a bead marked blocked whose dispatch's worktree still exists
    When status is reported
    Then that bead is reported as an unresolved conflict holding its evidence
```

AC-SWARM-004.6 is retired: the requirement that every reported line names
the command behind it applies to every subcommand, not to status alone,
and moved to AC-SWARM-014.3. The identifier is not reused.

### FEAT-SWARM-005: Prove a push landed

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
@FEAT-SWARM-005
Feature: Prove a push landed

  As a coordinator finishing a dispatch or a run
  I want the branch compared against its remote after every push
  So that a rejected push is not indistinguishable from a finished run

  This scores lowest in the feature on prose deleted and highest on
  failure removed, because a rejected push leaves every other signal
  reading finished: the tree is clean, no worktrees stand, the beads are
  closed in the tracker. One fetch and one comparison is the only thing
  that separates work that landed from work that is stranded on a
  container about to be reclaimed.

  @AC-SWARM-005.1
  Scenario: Unpushed commits are reported as unpushed
    Given the local branch is ahead of its remote
    When the landing check runs
    Then it names the number of unpushed commits
    And it exits non-zero

  @AC-SWARM-005.2
  Scenario: A branch with no remote counterpart is not reported as landed
    Given the branch has no counterpart on the remote
    When the landing check runs
    Then it exits non-zero

  @AC-SWARM-005.3
  Scenario: A rejected push is surfaced with its cause
    Given a push that the remote rejects
    When the command reports the result
    Then the rejection's cause is included in the output

  @AC-SWARM-005.4
  Scenario: A landed branch reports clean
    Given every local commit is present on the remote
    When the landing check runs
    Then it exits zero

  @AC-SWARM-005.5
  Scenario: The remote is read afresh before the comparison
    Given a stale local record of where the remote branch points
    When the landing check runs
    Then the remote is fetched before the comparison is made

  @AC-SWARM-005.6
  Scenario: The landing check needs no run scope
    Given a checkout with no run epic known
    When the landing check runs
    Then it reports the branch's landing state

  @AC-SWARM-005.7
  Scenario: A push refused for one path is reported as a whole rejected push
    Given a remote that refuses one of the paths in the push
    When the command reports the result
    Then no commit in that push is reported as landed
    And the refused path is named
```

### FEAT-SWARM-006: Read and update the run manifest

**Commands:**

```text
swarm run manifest get     --run <id> [--key <key>] [--json]
swarm run manifest set     --run <id> --key <key> --value <value> [--if-absent] [--json]
swarm run manifest require --run <id> --key <key>...
```

**The validation is logical as well as structural.** The manifest's key
names are a contract, its types are a contract, and several of its keys
carry rules that no single write can check on its own — a branch written
twice, a review counter that went backwards, a ledger whose entries were
reordered under citations that address them by position. Every such rule
is a criterion below, because the alternative is prose in five skills.

**Two measured constraints shape the implementation**, both re-measured
on bd 1.2.2 while this document was written:

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

First: `--set-metadata key=value` parses JSON scalars — a number stays a
number, `true` stays a boolean, `null` stays null — but stores arrays and
objects as strings. So the list-valued and object-valued keys (`admitted`,
`excluded`, `moved`, `checks_ids`, `ledger`, `models`, `standing_auth`)
cannot be written that way at all.

Second: `--metadata @file` merges at the **top level** and replaces a
**nested** value wholesale — `obj2` went from `{"a":1}` to `{"b":2}` and
lost `a`. So it is the write path for the list-valued keys, and a caller
updating one field inside `standing_auth` or `models` must read that
object, edit it, and write the whole of it back.

**Replaces:** `references/run-record.md`, "The two things you own" item 1
and the whole of "Run manifest"; the manifest-reading preambles in
`run-batch/SKILL.md`, `sprint-review/SKILL.md` and `sprint-plan/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-006
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

  @AC-SWARM-006.1
  Scenario: Writing one key preserves the others
    Given a manifest carrying keys written by other components
    When one key is updated
    Then the other keys retain their values

  @AC-SWARM-006.2 @negative
  Scenario: An unknown key is refused rather than stored
    When a key outside the manifest's declared set is written
    Then the command exits non-zero
    And the manifest is unchanged

  @AC-SWARM-006.3
  Scenario: A required key that is absent is reported as absent, not empty
    Given a manifest with no checks baseline recorded
    When the baseline is read
    Then the result is reported as absent rather than as an empty baseline

  @AC-SWARM-006.4
  Scenario: The review pass counter survives a resumed run
    Given a manifest recording spent review passes
    When a resumed run reads the counter
    Then it reports the spent count rather than zero

  @AC-SWARM-006.5
  Scenario: A list-valued key reads back as a list
    Given a key whose declared type is a list of bead ids
    When it is written and read back
    Then the value is a list of the same ids rather than one string

  @AC-SWARM-006.6 @negative
  Scenario: The branch and its base are written once
    Given a manifest already recording a branch and a base commit
    When a different branch or base is written
    Then the command exits non-zero
    And the recorded values are unchanged

  @AC-SWARM-006.7 @negative
  Scenario: The review pass counter never decreases and never exceeds its budget
    Given a manifest recording two spent passes against a budget of two
    When a lower count or a third pass is written
    Then the command exits non-zero

  @AC-SWARM-006.8 @negative
  Scenario: The environmental ledger is append-only
    Given a ledger whose entries are cited elsewhere by position
    When a write would remove or reorder an entry
    Then the command exits non-zero

  @AC-SWARM-006.9 @negative
  Scenario: A value of the wrong type for its key is refused
    When a key whose declared type is a number is written with text
    Then the command exits non-zero
    And the manifest is unchanged

  @AC-SWARM-006.10 @negative
  Scenario: A write that would clobber a concurrent change is refused
    Given a key that changed after this caller read the manifest
    When that caller writes the manifest back
    Then the command exits non-zero naming the key that moved

  @AC-SWARM-006.11
  Scenario: A setting the harness does not expose is recorded as unrecorded
    Given a run whose harness exposes no effort setting
    When the manifest is written
    Then the effort key is present with the value unrecorded rather than absent

  @AC-SWARM-006.12 @negative
  Scenario: The admission record's three lists do not overlap
    When a bead appears in more than one of the admitted, excluded and moved lists
    Then the command exits non-zero naming that bead
```

### FEAT-SWARM-007: Partition the queue, assign it, and gate the epic

**Commands:**

```text
swarm queue partition load --run <id> --file <edges.jsonl> [--dry-run] [--json]
swarm queue assign         --run <id> --file <assignments.jsonl> [--dry-run] [--json]
swarm queue gate           --run <id> [--json]
```

Three subcommands and one requirement, because **the gate is the check on
the other two**: an unassigned bead and a mis-oriented edge are both
invisible until something asserts the front and the ownership after the
writes land.

**On a missing concurrency cap the gate refuses rather than defaulting.**
The cap is resolved once at pre-flight, from the harness, and recorded on
the manifest; a default substituted at gate time would be a number nobody
measured reading as one somebody did — and the gate exists precisely to
catch a pool that will silently never drain. The place a default belongs
is pre-flight's resolution step, where `20` is Claude Code's documented
value and the resolution is recorded with its source.

**Replaces:** `sprint-plan/references/partitioning.md`, the whole of
"Collision edges", "The edge file" and its two validators, and
"`concurrency_cap`"; `sprint-plan/SKILL.md`, "Assign every planned bead,
then gate the epic"; the four mechanical checks of
`groom-beads/references/epic-definition-of-ready.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-007
Feature: Partition the queue, assign it, and gate the epic

  As a planner preparing a run for dispatch
  I want collision edges loaded, every bead assigned, and the epic
  checked for swarmability as one act
  So that the ready front is safe to dispatch whole

  These three are one piece of work because the third checks the first
  two. A collision edge is a policy expressed as a dependency, and three
  ways of getting it wrong are silent: the orientation is that the first
  endpoint waits for the second, so reversing it inverts the dependency
  at exit zero; the collision mark must be stored bare, since a quoted
  value stores its quotes and then matches no other bead while looking
  correct in every listing; and a cluster must be chained rather than
  fanned, since a fan leaves the rest of the cluster ready at once and
  defeats the point. Assignment is now the whole handover, so an
  unassigned bead is one no agent will ever see. The epic gate is what
  catches that, and it has a silent failure of its own: an assignee read
  from the tracker's owner display passes on an epic where nothing is
  assigned, because that display shows the creator.

  Two of the epic gate's six checks stay with the planner and are not
  specified here: whether an unwired edge is a real dependency, and
  whether anything the run needs can be derived only by the planner.

  @AC-SWARM-007.1
  Scenario: Two beads sharing a collision mark are never ready together
    Given two beads separated by a collision edge
    When the ready front is queried
    Then at most one of them appears in it

  @AC-SWARM-007.2
  Scenario: An inverted edge is detected rather than loaded silently
    Given an edge file whose orientation would invert a dependency
    When the file is loaded and the front asserted
    Then the command exits non-zero naming the offending edge

  @AC-SWARM-007.3
  Scenario: A collision mark is stored bare
    When a collision mark is written
    Then reading it back returns the value without added quoting

  @AC-SWARM-007.4
  Scenario: A cluster is chained rather than fanned
    Given three beads sharing one surface
    When the partition is loaded
    Then exactly one of the three is in the ready front

  @AC-SWARM-007.5
  Scenario: An unassigned child is named by the epic gate
    Given a target epic with one unassigned child
    When the epic is gated for swarming
    Then that child's id is named
    And the command exits non-zero

  @AC-SWARM-007.6
  Scenario: The epic gate reads assignment from the tracker's data, not its display
    Given a bead whose creator is set and whose assignee is not
    When the epic is gated for swarming
    Then that bead is reported as unassigned

  @AC-SWARM-007.7
  Scenario: A dev pool larger than the concurrency cap is refused
    Given planned dev assignees exceeding the cap the manifest records
    When the epic is gated for swarming
    Then the count and the cap are both named
    And the command exits non-zero

  @AC-SWARM-007.8
  Scenario: Both ends of a collision edge carry the same mark
    Given a loaded collision edge
    When the marks are read back from the tracker
    Then both endpoint beads carry the same collision value

  @AC-SWARM-007.9
  Scenario: A collision mark with no edge to justify it is named
    Given a bead carrying a collision mark that no loaded edge accounts for
    When the partition is validated
    Then that bead is named
    And the command exits non-zero

  @AC-SWARM-007.10 @negative
  Scenario: A sequencing edge carries no collision mark
    Given an edge recorded as a sequencing constraint
    When the marks are read back
    Then neither endpoint carries a collision mark

  @AC-SWARM-007.11
  Scenario: A collision chain that contradicts a real dependency is refused
    Given two beads in one cluster already joined by a real dependency
    When a collision chain would order them the other way round
    Then the command exits non-zero naming both beads

  @AC-SWARM-007.12
  Scenario: A file that fails validation wires no edge at all
    Given an edge file whose second line names a bead that does not exist
    When the file is loaded
    Then no edge from that file is wired
    And the offending line is named

  @AC-SWARM-007.13
  Scenario: Loading the same file twice changes nothing
    Given an edge file that has already loaded
    When it is loaded again
    Then the graph is unchanged
    And the command exits zero

  @AC-SWARM-007.14
  Scenario: An edge that would strengthen an existing weaker edge is reported
    Given two beads already joined by a related edge
    When the file asks for a blocking edge between them
    Then the command exits non-zero naming both beads and the existing edge type

  @AC-SWARM-007.15
  Scenario: Assignees are derived from the run epic id and never minted
    Given a run whose epic id yields the run id "r7k" and a second batch
    When the batch is assigned
    Then every bead in it carries the assignee "dev-r7k-2"

  @AC-SWARM-007.16
  Scenario: The epic gate counts closed and deferred children
    Given a target epic with one closed child carrying no assignee
    When the epic is gated for swarming
    Then that child is included in the check rather than skipped

  @AC-SWARM-007.17 @negative
  Scenario: A missing concurrency cap is reported as unmeasurable
    Given a manifest recording no concurrency cap
    When the epic is gated for swarming
    Then the check is reported as unmeasurable
    And no default cap is substituted
    And the command exits non-zero

  @AC-SWARM-007.18
  Scenario: The pool count includes dev assignees only
    Given a run carrying four dev assignees and four ceremony assignees
    When the epic is gated for swarming
    Then the reported pool count is four
```

### FEAT-SWARM-008: Open a dispatch — the sprint branch and its worktree

**Commands:**

```text
swarm run branch ensure --run <id> [--name <branch>] [--json]
swarm dispatch open     --run <id> --dispatch <dispatch-id> [--base <ref>] [--json]
```

**This requirement is the open side only.** Removing a worktree and
deleting its branch belong to `swarm dispatch accept`, because removal is
the final act of accepting and its guard is the merge that precedes it.
Splitting the pair across two requirements is what let a removal run after
a merge exited 128.

**The base is an argument, which is what retires a blocking pre-flight
gate.** Under the harness's own worktree isolation the base came from a
global setting no repository owns, the path was chosen for you, and the
only available check was an assertion after the fact. With the coordinator
creating the worktree the base is named in the command and there is
nothing to verify.

**Replaces:** `references/worktree-mechanics.md`, "The sprint branch" and
"The base-ref trap" in full, and step 1 of "The worktree lifecycle"; the
`BB_WORKTREE_BASEREF` block in `pre-flight-check.sh`; the "Hard gate"
paragraph in `bead-bonanza/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-008
Feature: Open a dispatch — the sprint branch and its worktree

  As a coordinator isolating a dispatch
  I want the sprint branch and each dispatch's worktree created from a
  base I name
  So that a worktree can never be cut from the wrong base

  This is the largest single reduction in prose in the feature, and it
  depends on the coordinator creating the worktree rather than the
  harness. Where the harness creates it, the base comes from a global
  setting no repository owns, the path is chosen for you, and the only
  available check is an assertion after the fact. Where the coordinator
  creates it, the base is an argument and there is nothing to verify.

  @AC-SWARM-008.1
  Scenario: A worktree is created from the base the caller names
    When a worktree is created for a dispatch
    Then its head is the named base

  @AC-SWARM-008.2
  Scenario: An existing sprint branch is reused rather than recreated
    Given a manifest already recording a sprint branch
    When the branch is ensured
    Then the recorded branch is checked out
    And no new branch is created

  @AC-SWARM-008.3
  Scenario: A branch name already taken by unrelated work stops the run
    Given the intended sprint branch name exists and is unrelated
    When the branch is ensured
    Then the command exits non-zero
    And no branch is created or moved

  @AC-SWARM-008.4
  Scenario: One worktree and one branch are created per dispatch
    When a worktree is created for a dispatch carrying several beads
    Then exactly one branch is created
    And it is named for the dispatch

  @AC-SWARM-008.5
  Scenario: The base sha is recorded at creation
    When the sprint branch is created
    Then the manifest records the sha it was created from

  @AC-SWARM-008.6 @negative
  Scenario: A worktree is never created inside the repository's working tree
    When a worktree is created for a dispatch
    Then its path is outside the repository's working tree

  @AC-SWARM-008.7
  Scenario: The worktree path and branch name follow from the dispatch id alone
    Given two sessions holding the same run epic id and the same dispatch id
    When each computes the worktree path and branch name
    Then both compute the same values

  @AC-SWARM-008.8
  Scenario: Stale worktree registrations are cleared before the list is read
    Given a registered worktree whose directory no longer exists
    When a dispatch is opened
    Then the stale registration is cleared before any decision reads the list

  @AC-SWARM-008.9
  Scenario: The base defaults to the sprint branch as it stands now
    Given no base is named
    When a worktree is created for a dispatch
    Then its head is the sprint branch's head read at that moment

  @AC-SWARM-008.10
  Scenario: A dispatch whose worktree already exists is not recreated
    Given a dispatch whose worktree exists and holds uncommitted work
    When the dispatch is opened again
    Then the existing worktree is reported and left untouched

  @AC-SWARM-008.11 @negative
  Scenario: No global setting decides the base
    When a worktree is created for a dispatch
    Then no configuration outside the repository is read to determine its base
```

### FEAT-SWARM-009: Reconstruct a run nobody is holding

**Command:**

```text
swarm run recover [--run <id>] [--branch <name>] [--json]
```

**Three ways to identify the run, tried in order, and each is reported.**
An explicit `--run` is taken as a starting point and still confirmed
against the branch. Otherwise the checked-out branch name is matched
against the `branch` recorded in candidate run epics' manifests. Failing
that, the close-commit subjects on the branch carry bead ids in the form
`<id> P<n> <title>`, and any one of them resolves to its parent chain and
so to the run epic.

**The third way has a measured failure mode**: a branch that was
squash-merged carries no such subject, so the completed set cannot be read
from history. That is reported as unestablished, never as an empty
completed set — the two are the difference between a run that did nothing
and a run whose record this checkout cannot see.

**Replaces:** `sprint-recovery/SKILL.md` steps 1 through 5, including the
five-row classification table a human is currently asked to apply.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-009
Feature: Reconstruct a run nobody is holding

  As a session resuming a run with no context
  I want the run derived from its branch, its epic and the surviving
  worktrees
  So that a run survives the loss of the session that started it

  Every input is durable and every output is a derived answer, which is
  why this is both a large token saving and the mechanism by which a run
  survives at all. The per-bead classification is a pure function of two
  inputs — a bead's status and whether its dispatch's worktree still
  exists — currently written as a table for a human to apply.

  @AC-SWARM-009.1
  Scenario Outline: The run is identified from whichever durable source answers
    Given a checkout on a sprint branch and "<available>"
    When reconstruction runs
    Then it names the run epic
    And it names "<source>" as what identified it

    Examples:
      | available                                         | source              |
      | a run epic id supplied by the caller              | the supplied id     |
      | a run epic whose manifest records this branch     | the manifest branch |
      | close commits carrying bead ids in their subjects | the branch history  |

  @AC-SWARM-009.2
  Scenario: Completed beads are read from the branch history
    When reconstruction runs
    Then every bead whose close commit is on the branch is reported completed

  @AC-SWARM-009.3
  Scenario: Stale worktree registrations are not reported as in flight
    Given a registered worktree whose directory no longer exists
    When reconstruction runs
    Then it is not counted as in flight

  @AC-SWARM-009.4
  Scenario: Every unfinished bead is classified with the signal that settled it
    When reconstruction runs
    Then each unfinished bead carries both its classification and the signal it was derived from

  @AC-SWARM-009.5
  Scenario: What could not be established is named rather than omitted
    Given a source that cannot be read
    When reconstruction reports
    Then the affected facts are listed as unestablished

  @AC-SWARM-009.6
  Scenario Outline: Each unfinished bead is classified from its status and its worktree
    Given a bead whose status is "<status>" and whose dispatch's worktree is "<worktree>"
    When reconstruction runs
    Then that bead is reported as "<classification>"

    Examples:
      | status      | worktree | classification                      |
      | in progress | present  | the agent died holding it           |
      | open        | present  | started and abandoned, or unstarted |
      | in progress | absent   | died before or without a worktree   |
      | open        | absent   | never started                       |
      | blocked     | either   | stopped on purpose                  |

  @AC-SWARM-009.7
  Scenario: Worktree registrations are cleared before the in-flight set is read
    When reconstruction runs
    Then stale registrations are cleared before the in-flight set is computed

  @AC-SWARM-009.8
  Scenario: The review pass counter is carried forward
    Given a manifest recording spent review passes
    When reconstruction reports the run
    Then it reports the spent count rather than zero

  @AC-SWARM-009.9 @negative
  Scenario: Reconstruction creates no branch
    Given a manifest recording a sprint branch
    When reconstruction runs
    Then no branch is created

  @AC-SWARM-009.10
  Scenario: A history with no close-commit subjects yields no completed set
    Given a branch whose commits were squashed and carry no bead ids
    And no other source identifies the completed beads
    When reconstruction reports
    Then the completed set is reported as unestablished rather than as empty

  @AC-SWARM-009.11
  Scenario: The landing check is reported before anything else
    When reconstruction runs
    Then the branch's landing state is the first verdict in its output
```

### FEAT-SWARM-010: Answer what may be dispatched now

**Command:**

```text
swarm queue front --run <id> [--json]
```

**Replaces:** `references/dispatching.md`, "Dispatch continuously" and
"Circuit breaker"; the `--exclude-label human` paragraph in
`references/run-record.md`; the dispatch-list line in
`bead-bonanza/SKILL.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-010
Feature: Answer what may be dispatched now

  As a coordinator deciding what to send out
  I want the dispatch front returned with the run's filter applied, and
  systemic failure detected
  So that I never hand an agent a bead meant for a human, a bead already
  finished, or a bead somebody else is holding, and I stop when the next
  dispatch would meet the same fault

  The filter is a convention rather than a default: excluding beads
  awaiting a human answer belongs on the dispatch query and nowhere
  else, since a completion count that hides them reports a run as
  drained while a human still owes it an answer. An empty scope is the
  silent failure here, because an empty parent argument returns the
  whole repository rather than nothing.

  @AC-SWARM-010.1
  Scenario: A bead awaiting a human is excluded from the dispatch front
    Given a bead in the run labelled for human attention
    When the dispatch front is queried
    Then that bead is not in it

  @AC-SWARM-010.2
  Scenario: The same bead is included in a completion count
    Given the same bead
    When the run's outstanding work is counted
    Then that bead is counted

  @AC-SWARM-010.3
  Scenario: An empty or missing run scope is refused
    When the front is queried with no run scope
    Then the command exits non-zero
    And no beads are returned

  @AC-SWARM-010.4
  Scenario: Consecutive blocked or failed dispatches stop the run
    Given the configured number of consecutive dispatches have returned blocked or failed
    When the front is queried
    Then the command reports the run as stopped
    And exits non-zero

  @AC-SWARM-010.5 @negative
  Scenario: Epics are never in the dispatch front
    Given a run whose phase epics are open
    When the dispatch front is queried
    Then no epic is in it

  @AC-SWARM-010.6 @negative
  Scenario: A closed or deferred bead is not in the dispatch front
    Given one closed bead and one deferred bead in the run
    When the dispatch front is queried
    Then neither is in it

  @AC-SWARM-010.7
  Scenario: A bead a live dispatch is already holding is not offered again
    Given a bead assigned to a dispatch whose worktree is standing
    When the dispatch front is queried
    Then that bead is reported as held rather than as available

  @AC-SWARM-010.8
  Scenario: The front reaches beads nested below the run epic
    Given a run whose work beads are children of the target epic rather than of the run epic
    When the dispatch front is queried
    Then those beads are in it

  @AC-SWARM-010.9
  Scenario: The consecutive-failure count survives a resumed run
    Given a run that recorded two consecutive blocked dispatches before its session ended
    When a resumed session queries the front
    Then the count is reported as two rather than zero

  @AC-SWARM-010.10
  Scenario: The front is compared against the admitted set
    Given a bead in the front that the manifest's admitted list does not name
    When the front is queried
    Then that bead is named as an addition to the run's scope
```

### FEAT-SWARM-011: Record and read a finding

**Commands:**

```text
swarm finding record <bead-id> --kind <kind> [--target <target>] --evidence <text|@file> [--corrects <finding-id>]
swarm finding list   --run <id> [--bead <bead-id>] [--json]
```

**Replaces:** step 6 of `run-batch/SKILL.md` and the kind table in it; the
"Where to record" and "Finding ids are derived" sections of
`specs/standards/findings-capture-protocol.md` become a validated write
rather than a rule an agent applies.

AC-SWARM-011.4 and AC-SWARM-011.5 were criteria about claiming a bead.
They moved to FEAT-SWARM-015, where a claim is the subject rather than an
aside; both identifiers are retired here and are not reused.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-011
Feature: Record and read a finding

  As an agent capturing friction, or a retro reading it
  I want a finding's kind validated when it is written
  So that a finding cannot be recorded in a form nothing will ever route

  This is the cheapest stabilization in the feature. A finding routes by
  its kind, so a kind outside the nine routes nowhere and is invisible to
  every later sweep. Two consecutive runs lost their single most
  consequential finding that way, and the count of such findings was
  under-reported because the sweep that counted them had the same defect.

  What the command does not check is the finding's content. Whether the
  evidence is the right evidence, and whether the agent has drifted into
  analysis, are judgment and stay with the capturing agent and the retro.

  @AC-SWARM-011.1
  Scenario: A finding with an unknown kind is refused
    When a finding is recorded with a kind outside the declared set
    Then the command exits non-zero
    And nothing is written

  @AC-SWARM-011.2
  Scenario: A finding requiring a target is refused without one
    When a finding of a kind that requires a target is recorded without one
    Then the command exits non-zero

  @AC-SWARM-011.3
  Scenario: A finding is readable back with its bead and its ordinal
    Given findings recorded across several beads
    When they are read back
    Then each carries the bead it was recorded on and its position within that bead

  @AC-SWARM-011.6 @negative
  Scenario: A finding with no evidence is refused
    When a finding is recorded carrying no evidence
    Then the command exits non-zero
    And nothing is written

  @AC-SWARM-011.7 @negative
  Scenario: A finding whose only anchor is a line number is refused
    When a finding's evidence cites a path and a line number and nothing else
    Then the command exits non-zero naming the durable anchors it would accept

  @AC-SWARM-011.8
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

  @AC-SWARM-011.9 @negative
  Scenario: A recorded finding is never edited or deleted
    Given a finding already recorded on a bead
    When a change to it is attempted
    Then the command exits non-zero
    And the recorded finding is unchanged

  @AC-SWARM-011.10
  Scenario: A correction is a new finding naming the one it corrects
    Given a finding already recorded on a bead
    When a correcting finding is recorded against it
    Then both findings are readable
    And the second names the first

  @AC-SWARM-011.11 @negative
  Scenario: An ordinal is derived and is not accepted as input
    When a finding is recorded with an ordinal supplied by the caller
    Then the command exits non-zero
```

### FEAT-SWARM-012: Harvest and route a run's findings

**Commands:**

```text
swarm finding harvest --run <id> [--json]
swarm finding route   --run <id> [--json]
```

**Replaces:** sections 2 and 3 of `sprint-retro/SKILL.md`, and the
"Harvesting" section of the findings-capture protocol.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-012
Feature: Harvest and route a run's findings

  As a retro turning a run's friction into work
  I want every finding in the run collected and routed by kind
  So that the retro reads a complete set rather than whichever beads
  happened to still be open

  The harvest has a silent failure that produces a plausible answer: the
  natural listing query defaults to open issues, and a drained run is
  almost entirely closed, so a literal reading returns findings for the
  handful of beads that never closed and drops nearly everything else.
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

  @AC-SWARM-012.1
  Scenario: Findings on closed beads are harvested
    Given a run whose beads are closed
    When findings are harvested
    Then findings recorded on those beads are returned

  @AC-SWARM-012.2
  Scenario: A zero harvest on a drained run is reported as a failure
    Given a run whose beads are closed and which recorded findings
    When the harvest returns nothing
    Then the command exits non-zero

  @AC-SWARM-012.3
  Scenario: Each finding is routed by its kind alone
    When findings are routed
    Then each is assigned a destination determined solely by its kind

  @AC-SWARM-012.4
  Scenario Outline: Each kind has one destination
    Given a finding of kind "<kind>"
    When routing runs
    Then its destination is "<destination>"
    And an issue candidate is "<candidate>"

    Examples:
      | kind         | destination                       | candidate    |
      | gap          | the project audience              | produced     |
      | decision     | the project audience              | produced     |
      | stale        | the project audience              | produced     |
      | unverifiable | the project audience              | produced     |
      | sequencing   | the project audience              | produced     |
      | scope        | the project audience              | produced     |
      | tooling      | the skill audience                | produced     |
      | proof        | the record it supports            | not produced |
      | environment  | the environmental-failure ledger  | not produced |
      | an unknown kind | the unrouted list                | not produced |

  @AC-SWARM-012.5
  Scenario: The unrouted count is reported even when it is zero
    When routing runs
    Then the number of findings that matched no destination is reported

  @AC-SWARM-012.6
  Scenario: The harvest reaches every bead in the run, however nested
    Given a run whose work beads are children of the target epic rather than of the run epic
    When findings are harvested
    Then findings on those beads are returned

  @AC-SWARM-012.7
  Scenario: Filing a candidate records a back-link on the finding's bead
    Given a candidate filed from two findings
    When the filing is recorded
    Then each finding's bead carries a record naming the issue that was filed and the finding it came from
```

### FEAT-SWARM-013: Return a readiness verdict for a queue

**Commands:**

```text
swarm queue reconcile --run <id> --verdicts <file> [--json]
swarm bead bounce     <bead-id> --run <id> --gap <text>
swarm bead unbounce   <bead-id> --run <id> --answer <text>
swarm queue human     --run <id> [--json]
```

**A bounce is not a close and not a deferral.** The bead stays open,
gains the label the human queue is read from, and is parented to a late
phase so it surfaces at the end. Closing it would say the run did the
work; deferring it would say the run chose not to and would take it out
of the run's outstanding count, and it is precisely a bead the run is
still waiting on. The resolution is the label's removal plus the answer
recorded on the bead, after which the bead re-enters the gate.

**Replaces:** the readiness-gate and bounce sections of
`sprint-plan/SKILL.md`; the human-channel row and its two measured `bd`
notes in `references/run-record.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-013
Feature: Return a readiness verdict for a queue

  As a planner gating a queue before dispatch
  I want every bead in the queue to carry a verdict, and unready beads
  removed from the front
  So that a bead with no verdict is never mistaken for a ready one

  The mechanical half of readiness is accounting rather than judgment:
  whether every bead in the queue came back with a verdict at all. The
  judgment half stays with the grooming agent, which owns the checklist.
  Bouncing has its own silent failure: an unready bead marked only with
  a grooming annotation stays dispatchable, because the front reads
  status and not annotations.

  @AC-SWARM-013.1
  Scenario: Every bead in the queue is accounted for
    Given a queue and a set of verdicts
    When the verdicts are reconciled against the queue
    Then any bead without a verdict is named
    And the command exits non-zero

  @AC-SWARM-013.2
  Scenario: A bounced bead leaves the dispatch front
    Given a bead bounced as not ready
    When the dispatch front is queried
    Then that bead is not in it

  @AC-SWARM-013.3
  Scenario: A bounced bead is visible to a human
    Given a bead bounced as not ready
    When the human queue is listed
    Then that bead appears in it

  @AC-SWARM-013.4
  Scenario: A resolved bounce returns the bead to the gate
    Given a bounced bead whose gap has been closed
    When the bounce is resolved
    Then the bead re-enters the readiness gate on the next pass

  @AC-SWARM-013.5 @negative
  Scenario: A bounce neither closes nor defers the bead
    Given a bead bounced as not ready
    When its status is read
    Then it is open
    And it is neither closed nor deferred

  @AC-SWARM-013.6
  Scenario: The human queue omits a bead that has closed
    Given a bounced bead that has since been closed
    When the human queue is listed
    Then that bead does not appear in it

  @AC-SWARM-013.7
  Scenario: Resolving a bounce records the answer on the bead
    Given a bounced bead
    When the bounce is resolved with an answer
    Then the answer is recorded on the bead
    And the bead is not closed

  @AC-SWARM-013.8
  Scenario: A bounced bead is counted in the run's outstanding work
    Given a run whose only unfinished bead is bounced
    When the run's outstanding work is counted
    Then that bead is counted
    And the run is not reported as drained
```

### FEAT-SWARM-014: Show the commands behind every answer

**Flag, carried by every subcommand:**

```text
swarm <noun> <verb> ... --explain
```

**Not ranked with the twelve above, because it is not a capability.** It
is a property each of them carries, and it is specified once here rather
than repeated as a criterion in every requirement.

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
@FEAT-SWARM-014
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

  @AC-SWARM-014.1
  Scenario: An explained invocation runs nothing
    When a subcommand is invoked in explain mode
    Then it prints the tool invocations it would run, in order, with every argument resolved
    And no tool is executed

  @AC-SWARM-014.2
  Scenario: A normal run prints the invocations as it executes them
    When a subcommand runs normally
    Then each tool invocation is printed as it is executed

  @AC-SWARM-014.3
  Scenario: Every reported line carries the command that established it
    When a subcommand reports a fact
    Then the line names the command a reader can re-run to reproduce it

  @AC-SWARM-014.4
  Scenario: A step's exit status is read from the step itself
    Given a step whose command fails while a reader consumes its output
    When the step's result is reported
    Then the step is reported as failed

  @AC-SWARM-014.5
  Scenario: A step that succeeds while its reader stops early is not reported as failed
    Given a step whose reader stops consuming output before the command finishes
    When the step's result is reported
    Then the step is reported as succeeded

  @AC-SWARM-014.6 @negative
  Scenario: The output and error streams of a tool are never merged
    Given a tool that writes a warning to its error stream and data to its output stream
    When the subcommand parses the data
    Then the warning is not part of what is parsed

  @AC-SWARM-014.7
  Scenario: Explain output is a runnable transcript
    Given a subcommand invoked in explain mode
    When its printed invocations are run by hand in order
    Then they perform what the subcommand would have performed
```

### FEAT-SWARM-015: Claim a bead as the acting agent

**Command:**

```text
swarm bead claim <bead-id> --actor <dispatch-id> [--run <id>] [--json]
```

**Separated from FEAT-SWARM-011 deliberately.** A claim is an act on a
bead and a finding is a record about one; they shared a requirement only
because both are things a dev agent does. The two criteria that moved
here kept their meaning and took new identifiers.

**Replaces:** step 1 of the per-bead ceremony in `run-batch/SKILL.md`, and
the "Why the export matters" table in `references/dispatch-policy.md`.

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-015
Feature: Claim a bead as the acting agent

  As a dev agent starting a bead
  I want the claim refused unless the bead is mine and my identity is
  explicit
  So that a claim can never quietly move a bead's ownership

  The underlying claim is a silent failure in exactly one case, and it is
  the common one: on a bead with no assignee, with no actor in the
  environment, the claim succeeds at exit zero and writes the version
  control tool's configured user name into the assignee. Nothing errors,
  and the run's ownership view decays from that point on.

  A claim also has to set the status and the assignee together. Setting
  the status alone leaves a bead that reads as in progress and is claimed
  by nobody, which is the state no query can attribute.

  @AC-SWARM-015.1
  Scenario: A bead's claim is refused when the bead belongs to another agent
    Given a bead assigned to a different agent
    When a claim is attempted
    Then the command exits non-zero
    And the bead's assignee is unchanged

  @AC-SWARM-015.2
  Scenario: A claim on an unassigned bead does not silently take ownership
    Given a bead with no assignee
    When a claim is attempted without an explicit actor
    Then the command exits non-zero

  @AC-SWARM-015.3
  Scenario: A claim sets the status and the assignee in one write
    Given a bead assigned to the acting agent
    When it is claimed
    Then its status is in progress
    And its assignee is the acting agent

  @AC-SWARM-015.4
  Scenario: Re-claiming a bead the actor already holds changes nothing
    Given a bead already claimed by the acting agent
    When it is claimed again
    Then the command exits zero
    And the bead's status and assignee are unchanged

  @AC-SWARM-015.5 @negative
  Scenario: The actor is never taken from version control configuration
    Given an environment whose version control tool has a configured user name
    When a claim is attempted without an explicit actor
    Then that name is not written to the bead
```

## References

- `docs/research/swarm-cli/2026-08-25-capability-audit.md` — the audit
  these requirements are drawn from; §8 is the ranking this document
  preserves, and §7 is the classification that produced it
- [[module-bead-bonanza]] — the module whose components call this CLI
- [[adr-022]] — projects declare their checks in configuration, which is
  why no check command is hard-coded here
- [[findings-capture-protocol]] — the nine kinds FEAT-SWARM-011 validates
  and FEAT-SWARM-012 routes
- [[gherkin-guide]] — the notation these criteria are written in
