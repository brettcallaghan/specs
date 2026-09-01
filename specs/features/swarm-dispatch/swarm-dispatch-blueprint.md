---
title: "Feature Blueprint: Swarm Dispatch Loop"
description: Extend the swarm binary's tracker and version-control
  interfaces to cover the dispatch loop, build its five commands on top of
  them, and adopt each into the skills whose prose it replaces
slug: feature-swarm-dispatch
type: feature-blueprint
status: in-progress
updated_at: 2026-09-01T00:00:00Z
delivery: staged
plan: agent-plugins-o63s
frd: frd-swarm-dispatch
---

# Swarm Dispatch Loop

**FRD:** [[frd-swarm-dispatch]]

## Overview

[[frd-swarm-core]] left a binary that can read a run and report on it;
this feature gives it the verbs that move one. Nothing here is a new
architectural shape — every command is a plan of steps over the two
interfaces [[container-swarm-cli]] already declares, so the work divides
into widening those interfaces, building five commands on the widened
surface, and retiring the skill prose each command replaces.

The widening is where the risk is. The Tracker gains scoped epic
eligibility and the queries the ready set is derived from; Version Control
gains worktree creation, branch creation and deletion, rebase, merge and
push. Both are the components whose implementations are the only ones
allowed to name `bd` and `git`, so every decoder they gain is a place a
wrong answer could be returned at exit 0 — which is the failure this
container exists to remove.

Delta blocks in dependency order: **TRK**, **VCS**, **MFK** and **FIX**
build the capability layer; **PC**, **QF**, **AQ**, **CB**, **BR**,
**WT**, **MG**, **RM** and **LD** build the loop on top of it; **ADP-C**
and **ADP-D** adopt it into the skills.

## Element Deltas

### `Tracker` and `Beads Tracker` in [[container-swarm-cli]] (modified — the dispatch loop's operations)

**Intent:** the loop asks the tracker four questions the core feature did
not need: which epics under a run are eligible to close, what is ready
under a scope with the run's filters applied, what one assignee holds,
and what a bead's acceptance signals say. Each is stated in the run's
vocabulary on the interface and decoded once in the implementation, so
the commands above stay free of any tracker's spelling.

**Behavior:**
- The `Tracker` interface SHALL declare an operation returning the epics
  beneath a named run that are eligible to close, and that operation
  SHALL take the run scope as an argument rather than reading a default.
- The `Tracker` interface SHALL declare an operation returning the beads
  ready beneath a named scope, excluding epics, and an operation
  returning the beads one assignee holds beneath that same scope,
  partitioned by whether each is ready or in progress.
- WHEN an operation's response is not the shape its decoder expects, the
  `Beads Tracker` SHALL report the step as failed rather than decoding a
  length or a field from it.
- Every operation added here SHALL return a plan step, and SHALL NOT
  reach the Process Runner itself.

**Anchors:** `cli/swarm/tracker`, `cli/swarm/tracker/beads`.

### `Version Control` and `Git Version Control` in [[container-swarm-cli]] (modified — the dispatch loop's operations)

**Intent:** the loop creates and removes worktrees, creates and deletes
branches, rebases, merges fast-forward-only, fetches and pushes. Ancestry
comparison, worktree listing with its prune, and remote comparison
already exist; this widens the interface to the writes and states them in
the run's vocabulary.

**Behavior:**
- The `Version Control` interface SHALL declare operations for creating a
  worktree at a named path on a named new branch from a named base,
  removing a worktree, creating a branch from a named base, deleting a
  branch, rebasing a branch onto a named commit, merging
  fast-forward-only, fetching a remote, and pushing a branch.
- The worktree-removal operation SHALL NOT accept a force argument, so no
  caller can reach the forcing form.
- WHEN a merge or a rebase is refused, the `Git Version Control`
  implementation SHALL decode the refusal into the run's vocabulary —
  which paths conflict, and whether the refusal was a non-fast-forward —
  rather than returning the tool's text for a caller to parse.
- Each operation whose invocation writes SHALL declare whether re-running
  it is safe, per [[adr-026]].
- A step whose answer is carried by its exit status SHALL declare the
  statuses it answers with, so a legitimate "no" is not read as a
  failure.

**Anchors:** `cli/swarm/vcs`, `cli/swarm/vcs/git`.

### `schemas/swarm/manifest.strata` and the `Manifest` rule set (modified — the failure count)

**Intent:** the circuit breaker has to survive a resumed session, so its
count is durable. The manifest's key set is closed, so the key is
declared before anything writes it.

**Behavior:**
- `RunManifest` SHALL carry `consecutive_dispatch_failures`, an optional `Int`
  declared on `DispatchKeys`, so its stage is dispatch and its absence
  before the run's first dispatch is legitimate.
- The `Manifest` component SHALL refuse a write that decreases
  `consecutive_dispatch_failures` by more than a reset to zero, since the only
  legitimate decrease is a dispatch that completed something.
- WHEN `consecutive_dispatch_failures` is absent, the `Manifest` component SHALL
  report it as absent rather than as zero; reading an absent count as
  zero is the caller's, not the record's.

**Data:** the run manifest gains one key. `schemas/swarm/manifest.strata`,
the generated JSON Schema and Go model, and the validation rules are
updated in this delta, and the `core` end-to-end fixture is re-recorded
in it — both are lockstep rules in [[container-swarm-cli]]'s Change
Guidance and neither may land separately.

**Anchors:** `schemas/swarm/manifest.strata`, `cli/swarm/manifest`,
`cli/testdata/swarm-fixtures/core`.

### The end-to-end fixture variants (added)

**Intent:** the loop's commands cannot be proven against the `core`
fixture alone — several need a world it does not contain. Each variant is
derived as a branch from `core` rather than authored separately, so a
change to the core reaches every variant.

**Behavior:**
- Each variant SHALL be one directory under
  `cli/testdata/swarm-fixtures/`, holding `repository.bundle`,
  `tracker.jsonl` and a `fixture.toml` recording the contract version it
  was recorded against and its kind.
- The variants SHALL cover: a standing dispatch worktree with its branch;
  a dispatch branch cut from an earlier target than the current one; a
  branch whose rebase onto the target conflicts; a run whose phase epics
  are variously eligible and blocked by an open child; and a remote that
  refuses a push.
- A variant recorded to prove an older manifest shape is still readable
  SHALL NOT be re-recorded when the schema changes.

**Anchors:** `cli/testdata/swarm-fixtures/`, `cli/testdata/swarm-cases/`,
`cli/swarm/e2e`.

### `swarm phase close` in [[container-swarm-cli]] (added)

**Intent:** the tracker's own epic-closing command ignores its positional
argument and closes every eligible epic in the database at exit 0. The
command exists so that a scope is an argument the plan carries, and so
that closing nothing is a reported outcome rather than a silence.

**Behavior:**
- The command SHALL derive the epics it may close from the run epic's own
  descendants, and SHALL close only epics found by that walk.
- The command SHALL report each epic it closed as a property, and SHALL
  report the beads a closure released as ready.
- WHEN no epic in the run is eligible, the command SHALL report that as
  its answer and exit clear.
- WHEN an epic is left open because a child is open, the command SHALL
  report the open child's id as the reason.
- WHEN the eligibility query fails, the command SHALL report the query as
  unanswered and take the stop status, never reporting an unanswered
  query as nothing eligible.
- An epic SHALL be eligible WHEN it is open and carries no open child,
  which is the tracker's own rule; what this command adds is the
  restriction to the run's descendants, never a second rule.

**Anchors:** `cli/swarm/cli`, `cli/cmd`-style command registration under
`cli/swarm`.

### `swarm ready` in [[container-swarm-cli]] (added — the dispatch set)

**Intent:** what may be dispatched now, with the run's filters applied
inside the run's scope. Four measured traps live here, and each is a
filter or a guard rather than a longer query.

**Behavior:**
- The command SHALL refuse an empty or absent run scope with the
  malformed-invocation status, and SHALL NOT treat an absent scope as the
  whole repository.
- The ready set SHALL exclude epics, closed beads, deferred beads, and beads
  labelled for human attention.
- The ready set SHALL reach beads at any depth beneath the run epic, through
  the `Tracker`'s descendant walk rather than a single-level query, and
  SHALL span every dispatch in the run rather than any one of them —
  narrowing to a dispatch is `--agent`'s and nothing else's.
- Each bead in the ready set SHALL be reported with the dispatch id it is
  assigned to, since that is what tells a coordinator which dispatches
  have work waiting; a bead assigned to none SHALL be reported as
  unassigned rather than omitted.
- WHEN a bead is assigned to a dispatch whose worktree is standing, the
  command SHALL report it as held rather than as available.
- A dispatch's standing worktree SHALL be matched to its dispatch by the
  branch that worktree has checked out, and WHERE that branch carries no
  dispatch id, by the assignee of the bead its head commit closed.
- WHEN a tracker response is not the shape a queue takes, the command
  SHALL take the stop status rather than reporting a length derived from
  it.

**Anchors:** `cli/swarm/cli`, `cli/swarm/tracker`, `cli/swarm/vcs`.

### `swarm ready --agent` in [[container-swarm-cli]] (added — one dispatch's queue)

**Intent:** the same query with one more filter, so the filter set is
implemented once. It answers what one dispatch still holds, and separates
an agent that has finished from one that was never handed anything.

**Behavior:**
- WHEN `--agent` is passed, the command SHALL return only beads assigned
  to that dispatch within the run, in an order their dependencies allow.
- The drain verdict SHALL be derived from two queries — ready and in
  progress — so an agent holding a bead in progress and none ready is
  reported as still holding work.
- The agent's queue SHALL NOT reach outside the run, even where the same
  dispatch id is assigned to a bead in another run.
- WHEN no bead anywhere in the run is assigned to the named dispatch, the
  command SHALL report that dispatch as not established rather than as
  drained, and take the attention status.
- The drain derivation SHALL be one implementation, reused by
  `swarm dispatch accept`'s drain check rather than restated in it.

**Anchors:** `cli/swarm/cli`, `cli/swarm/tracker`.

### `swarm ready` in [[container-swarm-cli]] (added — the circuit breaker)

**Intent:** the breaker's unit decides whether it fires at all. Counted as
consecutive blocked beads it misfires on its own machinery, so the unit is
one dispatch return in which no bead completed.

**Behavior:**
- The command SHALL advance `consecutive_dispatch_failures` by one WHEN a
  dispatch returns having completed no bead, and SHALL reset it to zero
  WHEN a dispatch returns having completed at least one.
- A bead returned as not reached SHALL NOT advance the count on its own.
- The command SHALL read `consecutive_dispatch_failures` from the manifest
  and SHALL treat an absent value as zero, so a fresh run starts at zero
  and a resumed run continues from what it recorded.
- WHEN the count reaches the configured threshold, the command SHALL
  report the run as stopped and take the stop status.
- The threshold SHALL arrive as an argument to the command, and WHERE it is
  absent the command SHALL use 3.
- The command SHALL NOT read `sf.toml`, or any other configuration file, to
  resolve the threshold. Resolving it is the coordinator's, at the
  authorization gate where every other run policy is resolved, so the
  container gains no configuration-reading surface.

**Anchors:** `cli/swarm/cli`, `cli/swarm/manifest`.

### `swarm dispatch open` in [[container-swarm-cli]] (added — the sprint branch)

**Intent:** the sprint branch is established as the first step of opening a
dispatch, and nowhere else. Folding it in is what makes the timing rule
structural rather than remembered.

**Behavior:**
- WHEN the manifest records no branch, the command SHALL create the sprint
  branch and record it together with the commit it was cut from, in one
  manifest write.
- WHEN the manifest records a branch, the command SHALL use the recorded
  branch and create none, and SHALL ignore `--branch`.
- WHERE `--branch` is absent and no branch is recorded, the branch name
  SHALL be `sprint/<run-epic-id>`, which is the form the skills already
  create and resume onto.
- WHEN a branch already carries the name this run would use and the
  manifest records no branch, the command SHALL take the attention status,
  name the existing branch, name passing an explicit branch name as the
  resolution, and create or move nothing.
- The branch SHALL be created and recorded before the worktree is created.

**Anchors:** `cli/swarm/cli`, `cli/swarm/manifest`, `cli/swarm/vcs`.

### `swarm dispatch open` in [[container-swarm-cli]] (added — the worktree)

**Intent:** the coordinator creating the worktree is what makes the base an
argument, the path knowable and the branch name meaningful — three things
the harness's own isolation takes away.

**Behavior:**
- The command SHALL create exactly one worktree and one branch per
  dispatch, and SHALL name that branch for the dispatch id.
- The worktree path and the branch name SHALL be derived from the run epic
  id and the dispatch id alone, so two sessions compute the same values.
- The worktree path SHALL be outside the repository's working tree: a
  sibling of the repository root, `<parent>/.bb-worktrees/<dispatch-id>`,
  which is the location the skills already create and clean up — a
  different one would orphan every worktree standing at adoption. The
  dispatch branch SHALL be `bb/<dispatch-id>`, likewise the form already
  in use. Renaming either travels with the skill-family rename, not with
  this feature.
- WHERE `--base` is absent, the base SHALL be the sprint branch's head read
  at that moment.
- The command SHALL NOT read any configuration outside the repository to
  determine the base.
- Stale worktree registrations SHALL be cleared before any decision reads
  the worktree list.
- WHEN a worktree for the dispatch already exists, the command SHALL report
  it and leave it untouched.

**Anchors:** `cli/swarm/cli`, `cli/swarm/vcs`, `cli/swarm/vcs/git`.

### `swarm dispatch accept` in [[container-swarm-cli]] (added — drain, case and merge)

**Intent:** the first half of the close ceremony, up to and including the
merge. The fast-forward refusal means opposite things in the serialized and
concurrent cases, so the case is named before the refusal is read.

**Behavior:**
- The command SHALL check that the dispatch has drained before anything
  else, and WHEN it has not, SHALL take the attention status naming what the
  agent still holds and merge or remove nothing.
- The command SHALL determine whether the dispatch was serialized or
  concurrent before it reads a fast-forward refusal, and SHALL report which
  case it read.
- The case SHALL be decided by comparing the dispatch branch's merge-base
  with the target against the target's current head: equal is serialized,
  and anything else is concurrent. It SHALL NOT be inferred from how many
  worktrees are standing, which counts dispatches rather than measuring
  this one's base.
- WHEN the case is concurrent, the command SHALL rebase the dispatch branch
  onto the target as it stands at merge time and then fast-forward.
- WHEN the case is serialized and a fast-forward is refused, the command
  SHALL take the stop status without re-attempting the merge, and SHALL
  NOT re-attempt it with fast-forward-only relaxed.
- WHEN a rebase conflicts, the command SHALL record a sequencing finding
  naming both beads and every conflicting path, mark the unmerged bead
  blocked, leave the worktree standing, and resolve nothing itself.
- The finding SHALL be recorded through the `Tracker`'s comment operation
  in the run's existing finding convention, since no finding command
  exists and this feature waits on none. Adopting a dedicated one is
  [[frd-swarm-cli]]'s, and is a change of writer rather than of content.

**Anchors:** `cli/swarm/cli`, `cli/swarm/vcs`, `cli/swarm/tracker`.

### `swarm dispatch accept` in [[container-swarm-cli]] (added — ancestry, removal and landing)

**Intent:** the second half, whose whole point is that removal sits behind
the merge's result rather than beside it. A shell conjunction is not a
guard; the plan's ordering is.

**Behavior:**
- The command SHALL remove the worktree only WHEN the dispatch branch is an
  ancestor of the target, and WHEN it is not, SHALL leave the worktree in
  place, delete no branch, and take the stop status.
- The worktree SHALL be removed before its branch is deleted, and the branch
  SHALL NOT be deleted before the fast-forward has succeeded.
- WHEN the version control tool refuses a removal, the command SHALL NOT
  retry with force, and SHALL take the attention status naming what the
  tool refused over.
- WHERE a removal is refused because the worktree is locked, the resolution
  the command names SHALL be the unlock invocation, never the forcing form.
- After the merge and the removal, the command SHALL re-check every bead in
  the dispatch with no signal deferred, and SHALL attempt no push until
  every one has passed.
- WHERE `--no-push` is passed, the command SHALL complete every step before
  the push and report the branch as unlanded rather than pushing.
- WHEN run in dry-run mode, the command SHALL perform every read and no
  merge, removal, deletion or push.

**Anchors:** `cli/swarm/cli`, `cli/swarm/vcs`, `cli/swarm/tracker`.

### `swarm branch verify-landed` and `swarm branch push` in [[container-swarm-cli]] (added)

**Intent:** a rejected push leaves every other signal reading finished. One
fetch and one comparison is what separates work that landed from work
stranded on a container about to be reclaimed.

**Behavior:**
- Both commands SHALL be invocable with no run scope, so a recovering
  session can run the landing check before it knows which run it is in.
- WHERE `--branch` is absent, the branch SHALL be the one currently
  checked out; WHERE `--remote` is absent, the remote SHALL be `origin`.
  Both defaults are read from the repository, so neither command consults
  configuration outside it.
- The landing check SHALL fetch the remote before comparing, so a stale
  local record of the remote's position cannot report a branch as landed.
- WHEN the local branch is ahead of its remote, the command SHALL name the
  number of unpushed commits and take the attention status.
- WHEN the branch has no counterpart on the remote, the command SHALL take
  the attention status rather than reporting it as landed.
- WHEN a push is rejected, the command SHALL report the rejection's cause,
  and WHERE the remote refuses one of the push's paths, SHALL report no
  commit in that push as landed and name the refused path.
- `swarm branch push` SHALL be `verify-landed` with the push ordered before
  it, one implementation rather than two.

**Anchors:** `cli/swarm/cli`, `cli/swarm/vcs`, `cli/swarm/vcs/git`.

### [[module-bead-bonanza]] (modified — coordinator adoption)

**Intent:** an invariant enforced by a command is prose that can be deleted.
Each requirement in the FRD names the prose it retires; this delta performs
the retirement on the coordinator's own documents and script.

**Behavior:**
- The Coordinator SHALL reach the loop through `swarm phase close`,
  `swarm ready`, `swarm dispatch open` and `swarm dispatch accept`,
  and SHALL route on each command's exit status rather than on its text.
- The prose each command replaces SHALL be deleted in the same change that
  adopts it, not summarized beside it: `supervision.md`'s re-query ordering;
  `dispatching.md`'s "Dispatch continuously", "Circuit breaker", "What an
  agent is holding", "Verify against state" step 3 and "The push, and
  proving it landed"; `worktree-mechanics.md`'s sprint-branch, base-ref-trap
  and lifecycle sections and the whole of "Merging"; `run-record.md`'s
  epic-closing and `--exclude-label human` paragraphs; and the
  corresponding paragraphs of `bead-bonanza/SKILL.md`.
- `pre-flight-check.sh`'s `BB_WORKTREE_BASEREF` block SHALL be removed, since
  the base is an argument the dispatch names and there is nothing left to
  assert.
- The Coordinator's roster line and Change Guidance SHALL name the commands
  it now depends on.
- The Scope component SHALL resolve the breaker's threshold as a seventh
  authorization-gate row — `sf.toml`'s `[run]` `dispatch_failure_threshold` when the
  key is present, else 3 — and SHALL record that row's value and its source
  in `standing_auth` as every other row is. The Coordinator SHALL pass the
  resolved value to `swarm ready`.

**Anchors:** `plugin/skills/bead-bonanza/`,
`plugin/skills/bead-bonanza/scripts/pre-flight-check.sh`,
`plugin/skills/sprint-scope/SKILL.md`, `sf.toml`,
`specs/architecture/module-bead-bonanza.md`.

### [[module-bead-bonanza]] (modified — dispatch-side adoption)

**Intent:** the two components that run inside or after a dispatch read the
same commands the coordinator does, so a dev agent and a recovering session
stop carrying their own queries.

**Behavior:**
- The Developer SHALL resolve its queue with `swarm ready --agent`,
  and `run-batch/SKILL.md`'s queue-resolution preamble SHALL be deleted.
- Recovery SHALL take `swarm branch verify-landed` as its first act, and
  `sprint-recovery/SKILL.md`'s step 1 SHALL be replaced by it.
- `supervision.md`'s "Running without the hooks" fallback queries SHALL be
  replaced by the command that answers them.

**Anchors:** `plugin/skills/run-batch/`, `plugin/skills/sprint-recovery/`,
`plugin/skills/bead-bonanza/references/supervision.md`.

### Relied-upon elements (unchanged)

- [[container-swarm-cli#Swarm Command Surface]] — the five contract flags,
  the run-scope refusal, both renderings and the exit-status mapping, none
  of which any command here implements again.
- [[container-swarm-cli#Plan Runner]] — the ordering guard that makes "a
  step does not run when the step before it failed" a property of every
  command rather than of the ones that remembered.
- [[container-swarm-cli#Process Runner]] — the one gateway to an external
  tool, with separated streams and a status read from the step itself.
- [[container-swarm-cli#Manifest]] and
  [[container-swarm-cli#Manifest Store]] — the one reader and writer of the
  record; no command here touches the run epic's metadata directly.
- [[container-swarm-cli#Status Collector]] — the completion tally and the
  scope comparison, which this feature deliberately does not duplicate.
- [[contract-swarm-json]] — the `CommandResult` shape `--full-result`
  renders and the fact kinds a property's value takes.
- [[swarm-output-standards]] — what every command here prints.

## Contracts

None. Every command here reports through the fact kinds
[[contract-swarm-json]] already declares, and adds no field to
`schemas/swarm/output.strata`. WHERE implementation finds a property that
no existing kind carries, the Change Guidance lockstep rule applies —
the schema and the contract's version are updated in the same change —
and that is a change to an existing contract rather than a new one.

## Tracker Reconciliation

- `agent-plugins-offx.2` — prerequisite: this FRD is that task's
  deliverable, and check 1 of the readiness checklist requires the FRD to
  reach `approved` before grooming may start.
- `agent-plugins-offx.4` — subsumed by the **MG** and **RM** deltas: its
  first measured failure is the merge-then-removal chaining this
  requirement turns into a plan-ordered guard.
- `agent-plugins-y63n.7` — subsumed by the **RM** delta: a locked worktree
  defeating the removal guard is answered by the behavior line naming the
  unlock invocation as the resolution and refusing the forcing form.
- `agent-plugins-y63n.4` — subsumed by the **QF** delta: a repo-wide
  human-queue check is run-scoped by construction once the ready set refuses an
  absent scope and derives its set from the run's own descendants.
- `agent-plugins-bhl` — out-of-scope: this feature specifies how a
  concurrent dispatch's branch is merged, not whether or how widely a run
  may dispatch concurrently. The concurrency policy and its cap stay that
  bead's.
- `agent-plugins-89xd` — out-of-scope: arming the `Stop` ready-loop hook is
  orchestration, which the CLI does not do. No delta here arms, schedules
  or monitors anything.
- `agent-plugins-hg9t.6` — out-of-scope: **PC** closes whatever phase epics
  the run's shape declares. Whether that shape should carry seven of them
  is that bead's question and [[frd-swarm-cli]]'s.

## Sequencing Constraints

Delivery is `staged`, in three stages.

- **Stage 1 — the capability layer.** The **TRK**, **VCS**, **MFK** and
  **FIX** deltas SHALL land before any command delta, since every command
  is a plan over the operations they add.
- **Stage 2 — the loop.** The **PC**, **QF**, **AQ**, **CB**, **BR**,
  **WT**, **MG**, **RM** and **LD** deltas SHALL land after stage 1 and
  before stage 3.
- **Stage 3 — adoption.** The **ADP-C** and **ADP-D** deltas SHALL land
  after every command they name is built, because a skill instructing a
  coordinator to run a command that does not exist stops the run it was
  meant to help.
- The **QF** delta SHALL land before the **AQ** delta, which narrows its
  query, and before the **CB** delta, which reports through its answer.
- The **BR** delta SHALL land before the **WT** delta: the worktree's
  default base is the sprint branch the branch delta establishes.
- The **MG** delta SHALL land before the **RM** delta, whose every guard
  reads the merge's result.
- The **LD** delta SHALL land before the **RM** delta, which ends by
  calling it.
- The **MFK** delta SHALL update `schemas/swarm/manifest.strata`, the
  generated artifacts, the validation rules and the re-recorded `core`
  fixture in one change — [[container-swarm-cli]]'s Change Guidance couples
  all four, and they are assigned to this one delta so the rule needs no
  cross-delta reconciliation.
- Every tracker and git subcommand the loop's commands invoke SHALL be
  added in the **TRK** and **VCS** deltas, with its own decoder — the
  Change Guidance couples a new subcommand to its implementation, and a
  command delta that finds one missing extends **TRK** or **VCS** in the
  same change rather than reaching the Process Runner itself.
- WHERE a command delta adds a field to a command's result,
  `schemas/swarm/output.strata` SHALL be updated in that same delta, per the
  same Change Guidance rule.
- The **ADP-C** delta SHALL delete the prose each adopted command replaces
  in the same change that adopts the command; documentation travels with
  the change it describes.
- The fold of a delta into [[container-swarm-cli]] SHALL land **per stage**
  rather than per delta: each delta delivers its code and its tests, and one
  work order per stage folds that stage's deltas into the container blueprint,
  depending on every delta in that stage. This is the statement the grooming
  skill's lockstep rule requires in place of per-item delivery, and it exists
  so the container blueprint — one file every delta would otherwise edit —
  does not serialize the whole feature behind a single work order. The stage
  boundary is the fold boundary because it is a seam this blueprint already
  draws, and it bounds code-to-blueprint drift to one stage.
- The prose deletions in **ADP-C** and **ADP-D** SHALL NOT be deferred to a
  fold work order. They are this feature's deliverable rather than a
  description of it, and travel with the change that adopts each command.

## AC Coverage

Legend: **TRK** = `Tracker` and `Beads Tracker` delta; **VCS** =
`Version Control` and `Git Version Control` delta; **MFK** = manifest
failure-count delta; **FIX** = fixture-variants delta; **PC** =
`swarm phase close` delta; **QF** = `swarm ready` delta (— the
dispatch set); **AQ** = `swarm ready --agent` delta; **CB** =
`swarm ready` delta (— the circuit breaker); **BR** =
`swarm dispatch open` delta (— the sprint branch); **WT** =
`swarm dispatch open` delta (— the worktree); **MG** =
`swarm dispatch accept` delta (— drain, case and merge); **RM** =
`swarm dispatch accept` delta (— ancestry, removal and landing); **LD** =
`swarm branch verify-landed` and `swarm branch push` delta; **ADP-C** =
[[module-bead-bonanza]] delta (— coordinator adoption); **ADP-D** =
[[module-bead-bonanza]] delta (— dispatch-side adoption).

**FIX**, **ADP-C** and **ADP-D** carry no coverage row. The first supplies
the worlds the e2e rows are proven against and the other two retire prose;
all three are held by the Definition of Done rather than by an AC.

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-SWARM-DISP-001.1 | TRK, **PC (proves)** | e2e |
| AC-SWARM-DISP-001.2 | PC | e2e |
| AC-SWARM-DISP-001.3 | PC | unit |
| AC-SWARM-DISP-001.4 | PC | e2e |
| AC-SWARM-DISP-001.5 | PC | unit |
| AC-SWARM-DISP-001.6 | TRK, **PC (proves)** | e2e |
| AC-SWARM-DISP-001.7 | PC | e2e |
| AC-SWARM-DISP-001.8 | TRK, **PC (proves)** | e2e |
| AC-SWARM-DISP-001.9 | PC | unit |
| AC-SWARM-DISP-002.1 | QF | e2e |
| AC-SWARM-DISP-002.2 | QF — via relied-upon `Status Collector` | e2e |
| AC-SWARM-DISP-002.3 | QF | unit |
| AC-SWARM-DISP-002.4 | CB | unit |
| AC-SWARM-DISP-002.5 | CB | unit |
| AC-SWARM-DISP-002.6 | CB | unit |
| AC-SWARM-DISP-002.7 | QF | e2e |
| AC-SWARM-DISP-002.8 | QF | e2e |
| AC-SWARM-DISP-002.9 | VCS, **QF (proves)** | e2e |
| AC-SWARM-DISP-002.10 | TRK, **QF (proves)** | e2e |
| AC-SWARM-DISP-002.11 | MFK, **CB (proves)** | e2e |
| AC-SWARM-DISP-002.12 | AQ | e2e |
| AC-SWARM-DISP-002.13 | AQ | e2e |
| AC-SWARM-DISP-002.14 | AQ | e2e |
| AC-SWARM-DISP-002.15 | AQ | e2e |
| AC-SWARM-DISP-002.16 | TRK, **QF (proves)** | unit |
| AC-SWARM-DISP-002.17 | TRK, **QF (proves)** | e2e |
| AC-SWARM-DISP-003.1 | VCS, **WT (proves)** | e2e |
| AC-SWARM-DISP-003.2 | BR | e2e |
| AC-SWARM-DISP-003.3 | BR | e2e |
| AC-SWARM-DISP-003.4 | WT | e2e |
| AC-SWARM-DISP-003.5 | BR | e2e |
| AC-SWARM-DISP-003.6 | **BR (proves)**, WT | unit |
| AC-SWARM-DISP-003.7 | WT | unit |
| AC-SWARM-DISP-003.8 | WT | unit |
| AC-SWARM-DISP-003.9 | VCS, **WT (proves)** | unit |
| AC-SWARM-DISP-003.10 | WT | unit |
| AC-SWARM-DISP-003.11 | WT | e2e |
| AC-SWARM-DISP-003.12 | WT | unit |
| AC-SWARM-DISP-004.1 | MG, **RM (proves)** | e2e |
| AC-SWARM-DISP-004.2 | MG | unit |
| AC-SWARM-DISP-004.3 | VCS, **MG (proves)** | e2e |
| AC-SWARM-DISP-004.4 | TRK, **MG (proves)** | e2e |
| AC-SWARM-DISP-004.5 | MG | unit |
| AC-SWARM-DISP-004.6 | RM — via relied-upon `Plan Runner` | unit |
| AC-SWARM-DISP-004.7 | VCS, **MG (proves)** | unit |
| AC-SWARM-DISP-004.8 | AQ, **MG (proves)** | e2e |
| AC-SWARM-DISP-004.9 | RM | unit |
| AC-SWARM-DISP-004.10 | RM | unit |
| AC-SWARM-DISP-004.11 | TRK, **RM (proves)** | e2e |
| AC-SWARM-DISP-004.12 | RM | unit |
| AC-SWARM-DISP-004.13 | MG, **RM (proves)** | unit |
| AC-SWARM-DISP-005.1 | LD | e2e |
| AC-SWARM-DISP-005.2 | LD | e2e |
| AC-SWARM-DISP-005.3 | LD | e2e |
| AC-SWARM-DISP-005.4 | LD | e2e |
| AC-SWARM-DISP-005.5 | LD | unit |
| AC-SWARM-DISP-005.6 | LD | unit |
| AC-SWARM-DISP-005.7 | VCS, **LD (proves)** | unit |

## Key Decisions

No decision made while authoring this blueprint had alternatives worth an
ADR of its own: every structural choice was settled by an ADR the container
already carries, and every behavioral choice by the FRD. The ones binding
here, and what each binds:

- [[adr-024]] — a command builds a plan of step values, which is why every
  ordering guarantee above is a property of the plan rather than of the
  shell that would otherwise sequence it.
- [[adr-025]] — every external invocation is an argv vector through one
  gateway, which is what keeps the merge, the rebase and the removal out of
  a shell that could run the next one regardless.
- [[adr-026]] — no rollback: the irreversible writes are ordered last, which
  is the rule the removal-behind-ancestry and push-behind-post-merge
  orderings are instances of.
- [[adr-027]] — two test tiers over a committed fixture, which is why the
  coverage table's `unit` rows assert a plan and its `e2e` rows assert a
  world.
- [[adr-016]] — Strata as the one source for schemas, which is why the new
  manifest key is a schema change rather than a convention.

## Definition of Done

- [ ] Each stage's fold work order has landed, so [[container-swarm-cli]]'s
      component records match the code at every stage boundary
- [ ] `sf.toml`'s `[run]` documents `dispatch_failure_threshold` and `sprint-scope`'s
      authorization-gate table carries it as a seventh row
- [ ] Every delta above is folded into its owning blueprint:
  - [ ] [[container-swarm-cli]] — the `Tracker`, `Version Control` and
        their implementations' component records updated with the
        operations added; the command tree in the `Swarm Command Surface`
        record extended to the five commands; Data and Change Guidance
        updated for the new manifest key and the fixture variants
  - [ ] [[module-bead-bonanza]] — the Coordinator, Developer and Recovery
        component records updated to name the commands they now call, and
        the module's version raised
- [ ] `schemas/swarm/manifest.strata` carries
      `consecutive_dispatch_failures`, the generated artifacts are
      regenerated, and the `core` fixture is re-recorded against them
- [ ] [[contract-swarm-json]] is unchanged, or its version is raised in the
      same change as any schema change that removed, renamed or retyped
      anything
- [ ] Every AC in the coverage table is verified at its declared level and
      cited from its tests per [[ac-traceability-standards]]
- [ ] Every prose section named in the **ADP-C** and **ADP-D** deltas is
      deleted, and `scripts/check-skills.sh` passes over the result
- [ ] `agent-plugins-offx.4`, `agent-plugins-y63n.7` and
      `agent-plugins-y63n.4` are closed or re-scoped against the deltas
      that subsume them
- [ ] This blueprint's status → retired

## Open Questions

None.

## References

- [[frd-swarm-dispatch]] — the requirements this blueprint realizes
- [[frd-swarm-core]] — the command contract, the manifest component and
  `swarm status`, all of which this feature builds on unchanged
- [[container-swarm-cli]] — the container every delta above belongs to,
  and whose Change Guidance binds them
- [[module-bead-bonanza]] — the module whose components adopt these commands
- [[swarm-output-standards]] — what every command here prints
- [[contract-swarm-json]] — the result shape and the fact kinds a property
  takes
- [[c4-standards]] — the container, module and component definitions the
  delta blocks and component references are written against
- [[ac-traceability-standards]] — the test citation obligations the AC
  Coverage table feeds
