---
title: "FRD: Swarm CLI — the coordinator's tested action surface"
description: A command-line tool that performs the thirteen highest-value
  actions a bead-bonanza run takes, so a run's correctness comes from an
  exit code rather than from an agent remembering a paragraph
slug: frd-swarm-cli
type: frd
status: draft
updated_at: 2026-08-25T00:00:00Z
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
- **Optimize.** A derivation an agent performs by hand costs N queries
  and N results held **in coordinator context** — the scarcest resource
  in a run, since it must survive the whole of it. The same answer as one
  command is one line.

**The order of the requirements below is the order of that value**, taken
from the stack rank in `docs/research/swarm-cli/2026-08-25-capability-audit.md`
§8. It is not arbitrary and it is not alphabetical: FEAT-SWARM-001 through
007 are the seven capabilities that are coordinator-side, run-ending, and
not mechanized today. 008 through 013 are the next tier.

**Out of scope for this feature.** Each deferred deliberately:

- **Anything already mechanized.** Thirteen capabilities are scripts
  today. Porting them to Go buys portability, one `--explain` convention
  and a test surface; it does not buy a capability the run lacks. That is
  a follow-on feature, and this one must not absorb it.
- **Orchestration.** The CLI spawns no agent, arms no monitor, schedules
  nothing, and holds no state of its own. It is called from inside an
  existing harness and returns.
- **Judgment.** Classifying a review finding's scope, reasoning findings
  into candidates, and deciding whether partial work is finishable stay
  with an agent. A command that guessed at them would be worse than the
  paragraph it replaced.
- **A tracker abstraction beyond what these actions need.** Bead
  operations go behind an interface so a second tracker is possible
  later, but this feature ships one implementation and no plugin
  mechanism.
- **Wrapping single commands that already work.** `bd ready`, `bd show`,
  `bd comment`, `bd create` and `git log` do what they say. Wrapping them
  hides a tool the agent knows behind one it does not.

## Terminology

| Term | Meaning |
|---|---|
| **run** | one execution of `bead-bonanza` over one run epic |
| **run epic** | the bead whose id is the run id and whose `--metadata` is the run manifest |
| **target** | the flat epic beneath the run epic whose task children are the work |
| **manifest** | the JSON object in the run epic's `--metadata`; the run's only durable record |
| **dispatch id** | one agent's whole identity — its `assignee`, its `BEADS_ACTOR`, and its branch name. Derived from the run epic id, never minted |
| **ready front** | `bd ready` over the target, filtered — the complete set of beads nothing is blocking right now |
| **collision edge** | a `blocks` edge between two beads that must not run at once because they share a component or fixture. Not a real dependency; both ends carry `metadata.collision` naming the surface |
| **derivation** | an answer requiring a tree walk or a join across `bd` and `git` that no single command returns |
| **silent failure** | a command that returns the wrong answer at exit 0 |

## Requirements

### FEAT-SWARM-001: Close finished ceremony epics, scoped to this run

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-001
Feature: Close finished ceremony epics, scoped to this run

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

  @AC-SWARM-001.1
  Scenario: Only epics belonging to the named run are closed
    Given eligible epics exist both inside and outside the named run
    When the command closes finished ceremony epics for that run
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

  @AC-SWARM-001.5
  Scenario: An unscoped invocation is refused
    When the command is invoked without a run scope
    Then it exits non-zero
    And no epic's status changes
```

### FEAT-SWARM-002: Merge a returned dispatch and accept it

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-002
Feature: Merge a returned dispatch and accept it

  As a coordinator accepting a dispatch that has returned
  I want the merge, the refusal reading, and the worktree removal to be
  one guarded act
  So that a worktree is never removed while its branch is unmerged

  Two failures make this one action rather than three. A fast-forward
  refusal means opposite things depending on whether the dispatches were
  serialized or concurrent, and reading it wrong stops a healthy run.
  And ordering the removal after the merge with a shell conjunction is
  not a guard: a merge that exited non-zero was followed by a removal
  that succeeded, leaving a removed worktree beside an unmerged branch
  for ninety seconds, which is the one state the in-flight signal
  asserts cannot occur.

  @AC-SWARM-002.1
  Scenario: A worktree is not removed unless its branch is an ancestor of the target
    Given a dispatch branch that has not merged
    When the command is asked to accept that dispatch
    Then the worktree is left in place
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
```

### FEAT-SWARM-003: Build a run's epic shape

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

  @AC-SWARM-003.1
  Scenario: The created shape dispatches exactly the target's open leaves
    When a run shape is created and the phases before the target are closed
    Then the run's ready front contains exactly the target's open task children

  @AC-SWARM-003.2
  Scenario: A partially built shape is not left behind
    Given the creation fails partway
    When the command reports the failure
    Then the run epic and its phases are either all present or all absent

  @AC-SWARM-003.3
  Scenario: Every phase epic is assigned to its role identity
    When a run shape is created
    Then each phase epic carries the assignee derived from the run id

  @AC-SWARM-003.4
  Scenario: The authorization gate holds the whole chain
    Given a run shape whose authorization gate is unresolved
    When the ready front is queried
    Then it is empty

  @AC-SWARM-003.5
  Scenario: A non-flat target is refused
    When a run shape is created with a sub-epic beneath the target
    Then the command exits non-zero naming the offending child
```

### FEAT-SWARM-004: Report a run's status from state

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
  feature, because every part of it is a derivation. A run once read
  healthy through a five-hour outage in which nothing was dispatched,
  because health was a label written by the component that had stopped.
  Completion needs a walk the tracker does not offer: three queries look
  like they answer "every bead in this run" and each answers something
  else. Drain needs two queries because the ready query alone excludes
  in-progress work and so reports a live agent as finished.

  @AC-SWARM-004.1
  Scenario: A stalled run is reported as stalled
    Given no worktrees are standing
    And the ready front is not empty
    And the last commit is older than the stall threshold
    Then the run's health is reported as stalled
    And the command exits non-zero

  @AC-SWARM-004.2
  Scenario: Health is never read from a stored label
    Given a run epic carrying a health label that contradicts its state
    When health is reported
    Then the reported value reflects the state, not the label

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

  @AC-SWARM-004.6
  Scenario: Every reported line carries the command that established it
    When status is reported
    Then each line names the command a reader can re-run to reproduce it
```

### FEAT-SWARM-005: Prove a push landed

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
```

### FEAT-SWARM-006: Read and update the run manifest

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

  @AC-SWARM-006.2
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
```

### FEAT-SWARM-007: Partition the queue, assign it, and gate the epic

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
    Given a target with one unassigned child
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
```

### FEAT-SWARM-008: Create, verify and remove a dispatch's worktree

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-008
Feature: Create, verify and remove a dispatch's worktree

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
```

### FEAT-SWARM-009: Reconstruct a run nobody is holding

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-009
Feature: Reconstruct a run nobody is holding

  As a session resuming a run with no context
  I want the run derived from its branch, its epic and the surviving
  worktrees
  So that a run survives the loss of the session that started it

  Every input is durable and every output is a derivation, which is why
  this is both a large token saving and the mechanism by which a run
  survives at all. The per-bead classification is a pure function of two
  inputs — a bead's status and whether its dispatch's worktree still
  exists — currently written as a table for a human to apply.

  @AC-SWARM-009.1
  Scenario: The run is identified from the branch alone
    Given a checkout on a sprint branch and no other context
    When reconstruction runs
    Then it names the run epic

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
```

### FEAT-SWARM-010: Answer what may be dispatched now

**Acceptance Criteria:**

```gherkin
@FEAT-SWARM-010
Feature: Answer what may be dispatched now

  As a coordinator deciding what to send out
  I want the dispatch front returned with the run's filter applied, and
  systemic failure detected
  So that I never hand an agent a bead meant for a human, and I stop
  when the next dispatch would meet the same fault

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
```

### FEAT-SWARM-011: Record and read a finding

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

  @AC-SWARM-011.4
  Scenario: A bead's claim is refused when the bead belongs to another agent
    Given a bead assigned to a different agent
    When a claim is attempted
    Then the command exits non-zero
    And the bead's assignee is unchanged

  @AC-SWARM-011.5
  Scenario: A claim on an unassigned bead does not silently take ownership
    Given a bead with no assignee
    When a claim is attempted without an explicit actor
    Then the command exits non-zero
```

### FEAT-SWARM-012: Harvest and route a run's findings

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
  Scenario: Kinds that terminate produce no issue
    Given findings of the kinds that terminate at the record
    When routing runs
    Then no issue candidate is produced for them

  @AC-SWARM-012.5
  Scenario: The unrouted count is reported even when it is zero
    When routing runs
    Then the number of findings that matched no destination is reported
```

### FEAT-SWARM-013: Return a readiness verdict for a queue

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
```

## References

- `docs/research/swarm-cli/2026-08-25-capability-audit.md` — the audit
  these requirements are drawn from; §8 is the ranking this document
  preserves, and §7 is the classification that produced it
- [[module-bead-bonanza]] — the module whose components call this CLI
- [[adr-022]] — projects declare their checks in configuration, which is
  why no check command is hard-coded here
