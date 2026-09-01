---
title: "Feature Blueprint: Swarm Core"
description: Build the swarm binary's command contract, its manifest
  component and its status command, prove them against a committed fixture,
  and adopt each command into the skills whose prose it replaces
slug: feature-swarm-core
type: feature-blueprint
status: retired
updated_at: 2026-08-29T23:59:59Z
delivery: staged
frd: frd-swarm-core
plan: agent-plugins-ih18
---

# Swarm Core

**FRD:** [[frd-swarm-core]]

## Overview

A new `swarm` binary joins the existing `cli` module and grows from the
bottom up: a process runner that invokes one external tool correctly, a
plan of step values that three interpreters read, a command surface that
turns a plan's result into output and an exit status, two tool
implementations behind role-named interfaces, and above those the
manifest and the status command. The Strata schemas that fix the JSON
shapes are data the manifest reads rather than code that repeats them. A
committed fixture — a `git bundle` and a tracker export — is what the
end-to-end cases run the real binary against. Last, each finished command
is adopted into the skills whose prose it replaces, one delta per
command, so a command that exists but nobody calls is never mistaken for
a command that landed.

In dependency order: **SKEL**, **PROC**, **PLAN**, **SURF**, **TRK**,
**VCS**, **SCHEMA**, **MAN**, **MANV**, **MANCMD**, **FIX**,
**ADOPT-CONV**, **ADOPT-MAN**, **COLL**, **REP**, **ADOPT-STAT**.

**The manifest command is adopted before the status command is built.**
That ordering is deliberate rather than incidental: adopting `swarm
manifest` into the skills puts it under real use while `swarm status` is
still being written, so whatever the contract gets wrong surfaces on the
first command rather than on both at once.

## Element Deltas

### [[container-swarm-cli]] (added — binary skeleton and build wiring)

**Intent:** stand the container up as a buildable, testable artifact
before any capability exists, so every later delta lands in a tree that
already compiles, formats, vets and cross-compiles on the matrices
[[adr-023]] puts it on. This delta delivers no behaviour a caller can
observe.

**Behavior:**
- The container SHALL build as a second binary from the existing `cli`
  module, with no new module and no second toolchain.
- The container SHALL be cross-compiled for every target the existing
  release matrix builds.
- WHEN the binary is invoked with no arguments, it SHALL print its
  command tree and exit with the malformed-invocation status.

**Anchors:** `cli/swarm/` for the new package tree; `cli/cmd/` for the
existing command-surface conventions to follow; `.github/workflows/cli.yml`
and `.github/workflows/release.yml` for the build matrices; `sf.toml` for
the declared checks.

### `Process Runner` in [[container-swarm-cli]] (added)

**Intent:** one component starts every external process, so the four
properties [[adr-025]] requires are structural rather than repeated at
call sites. It also becomes the seam the unit tier replaces, which is why
it ships before anything that would use it.

**Behavior:**
- The Process Runner SHALL invoke an external tool as an argv vector.
- The Process Runner SHALL NOT construct a command string for a shell
  interpreter.
- The Process Runner SHALL capture a tool's output stream and its error
  stream into separate buffers.
- The Process Runner SHALL read an invocation's exit status from the
  finished process.
- The Process Runner SHALL record each invocation's elapsed time.
- IF the named executable cannot be found, THEN the Process Runner SHALL
  report that distinctly from a non-zero exit.
- The Process Runner SHALL expose a fake implementation, in a test-only
  package, that replays recorded output, status and duration for a
  declared argv.

**Anchors:** `cli/swarm/exec/` for the runner and its fake; Go's
`os/exec` for the mechanism.

### `Plan Runner` in [[container-swarm-cli]] (added)

**Intent:** make the ordering guard, the timing and the three modes
properties of a data structure rather than of the code that happens to
build it ([[adr-024]]). A plan is a value, so the sequence a command
builds is assertable without running a subprocess — which is what makes
the guard itself testable.

**Behavior:**
- A step SHALL declare its invocation, its decoder, its failure policy,
  and whether re-running it is safe.
- WHEN a step whose failure policy is to stop returns a failure, the Plan
  Runner SHALL NOT run any later step in that plan.
- The Plan Runner SHALL offer exactly three interpreters over a plan:
  explain, dry-run and execute.
- WHEN a plan is explained, the Plan Runner SHALL print each step's
  invocation in order.
- WHEN a plan is explained, the Plan Runner SHALL execute none of the
  steps it printed.
- WHERE an argument is produced by an earlier step, explain output SHALL
  render it as that step rather than as a value.
- WHEN a plan is dry-run, the Plan Runner SHALL execute the steps
  declared as reads.
- WHEN a plan is dry-run, the Plan Runner SHALL refuse every step not
  declared as a read.
- WHEN a plan is executed, the Plan Runner SHALL print each step's
  invocation as it runs it.
- WHERE a step carries an alternative arm, explain output SHALL render
  both arms and name the condition that selects between them.
- The Plan Runner SHALL record, per step, the invocation, its exit
  status, its duration and a digest of each stream.

**Anchors:** `cli/swarm/plan/`; the Process Runner from the PROC delta.

### `Swarm Command Surface` in [[container-swarm-cli]] (added)

**Intent:** the process-level half of the contract — the flags every
command carries, the mapping from an outcome to one of four exit
statuses, and rendering in two forms that carry the same facts. It knows
nothing of the tracker or of git.

**Behavior:**
- The Command Surface SHALL accept `--explain`, `--dry-run` and `--json`
  on every subcommand.
- The Command Surface SHALL map an outcome to exit `0` when the verdict
  is clear, `1` when what was found needs attention, `2` when the caller
  cannot repair it in place, and `3` when the invocation itself is
  malformed.
- IF a subcommand that reads or mutates run state is invoked without a
  run scope, or with an empty one, THEN the Command Surface SHALL exit
  with the malformed-invocation status.
- WHEN a command exits with the attention status, its output SHALL name
  what would resolve the condition.
- The two syntaxes SHALL carry the same values, in the same order, for
  the same invocation.
- The full result SHALL conform to the shape [[contract-swarm-json]]
  declares, on every exit status.
- Each reported fact SHALL name the invocation that established it.

**Anchors:** `cli/swarm/main.go` and `cli/swarm/cli/`; `github.com/spf13/cobra`,
already a dependency of the module.

### `Tracker` and `Beads Tracker` in [[container-swarm-cli]] (added)

**Intent:** state what a run needs from an issue tracker in the run's own
vocabulary, and implement it once over `bd` with a decoder per
subcommand — which is where the measured shape non-uniformity is confined
([[adr-025]]).

**Behavior:**
- The Tracker interface SHALL name its operations in the run's
  vocabulary rather than in any tracker's.
- The Beads Tracker SHALL carry one decoder per `bd` subcommand it
  invokes.
- Each decoder SHALL normalize its own subcommand's absent, null and
  empty forms to one empty value for that subcommand.
- A component other than the Beads Tracker SHALL NOT invoke `bd`.
- The Beads Tracker SHALL reach beads only through the Process Runner.
- WHEN a bead set is requested for a run, the Beads Tracker SHALL return
  it transitively, so a bead nested below the run epic is included.

**Anchors:** `cli/swarm/tracker/` for the interface, `cli/swarm/tracker/beads/`
for the implementation and its decoders.

### `Version Control` in [[container-swarm-cli]] (added — with git)

**Intent:** the same seam for git, in the run's vocabulary — the sprint
branch, a dispatch's branch, a standing worktree — rather than in git's.

**Behavior:**
- The Version Control interface SHALL name its operations in the run's
  vocabulary rather than in git's.
- A component other than the Git Version Control SHALL NOT invoke `git`.
- The Git Version Control SHALL reach git only through the Process
  Runner.
- The Git Version Control SHALL clear stale worktree registrations
  before returning a registration list.
- The Git Version Control SHALL read a commit's subject, its timestamp
  and its ancestry without parsing a rendered log.

**Anchors:** `cli/swarm/vcs/` for the interface, `cli/swarm/vcs/git/` for the
implementation.

### `schemas/swarm/manifest.strata` and `schemas/swarm/output.strata` (added)

**Intent:** the manifest's declared key set and the output envelope are
**data**, generated into Go types and JSON Schema by the generator this
repository already runs ([[adr-016]]). That is what makes
AC-SWARM-CORE-002.19 true by construction: reading, requiring and
validating cannot each hold their own idea of the shape when there is one
generated declaration.

**Behavior:**
- Each manifest key SHALL declare its name, its type, and the stage of a
  run at which it becomes required.
- The output schema SHALL declare an envelope every command's result
  carries, holding the contract version, the invocations run with their
  statuses and durations, and the facts the command established.
- A fact the command could not establish SHALL be representable as
  unestablished, distinctly from an empty value of its type.
- The generated artifacts SHALL be produced by
  `scripts/generate-schemas.sh`.
- The generated artifacts SHALL NOT be hand-edited.

**Data:**
- **The run manifest** — the key set is fixed here rather than in code.
  Its shape is [[contract-swarm-json]]'s; the keys themselves are the
  ones `bead-bonanza/references/run-record.md` specifies today, carried
  over rather than redesigned by this feature.

**Anchors:** `schemas/` beside the existing `.strata` sources;
`scripts/generate-schemas.sh`; `cli/core/schema/` for the generated
output.

### `Manifest` in [[container-swarm-cli]] (added — read and write path)

**Intent:** one component every other command reads a run through, split
from where the record is kept so that moving the record is a small change
rather than a rewrite of the validation. This delta is the path — fetch,
merge, write, conditional write — and the rules that belong to a single
write; the rule set that spans keys is the MANV delta.

**Behavior:**
- The Manifest SHALL read and write the record only through a Manifest
  Store.
- WHEN one or more keys are written, the Manifest SHALL merge them into
  the record so that a key it did not set retains its value.
- WHEN several keys are written in one invocation, the Manifest SHALL
  read and write the record once.
- IF any key in a multi-key write is refused, THEN the Manifest SHALL
  store none of them.
- IF a key in a multi-key write is refused, THEN the Manifest SHALL name
  that key.
- IF a key is outside the declared set, THEN the Manifest SHALL refuse
  the write and leave the record unchanged.
- IF a value's type is not its key's declared type, THEN the Manifest
  SHALL refuse the write and leave the record unchanged.
- WHEN a key is absent, the Manifest SHALL report it as absent rather
  than as an empty value of its declared type.
- A list-valued key SHALL read back as a list.
- The Manifest Store interface SHALL offer a write conditional on the
  record not having changed since the fetch that preceded it.
- IF the record changed after this caller's fetch, THEN the write SHALL
  be refused.
- WHEN a conditional write is refused, the key that moved SHALL be named.
- The Beads Manifest Store SHALL reach beads through the Beads Tracker
  rather than invoking `bd`.

**Anchors:** `cli/swarm/manifest/` for the model and the store interface,
`cli/swarm/manifest/beads/` for the implementation.

### `Manifest` in [[container-swarm-cli]] (added — the logical rule set)

**Intent:** the rules no single write can check on its own, which are
what let everything downstream treat the record as trustworthy. Separated
from the path delta because they are a rule set that grows, testable
against a record value with no store behind it at all.

**Behavior:**
- The branch and the base commit SHALL each be writable once; a
  different value SHALL be refused and the recorded values left
  unchanged.
- The review pass counter SHALL NOT decrease.
- The review pass counter SHALL NOT exceed the recorded budget.
- A write removing or reordering an entry of the environmental ledger
  SHALL be refused, so the ledger is append-only.
- IF a bead appears in more than one of the admitted, excluded and moved
  lists, THEN validation SHALL name that bead.
- WHERE the harness exposes no effort setting, the effort key SHALL be
  present with the value `unrecorded` rather than absent.
- WHEN the record is validated, every failing rule SHALL be reported
  rather than the first.
- WHEN a key is absent and its stage does not yet require it, validation
  SHALL report it as not yet written rather than as a failure.
- WHEN a precondition names keys, every absent one SHALL be named.

**Anchors:** `cli/swarm/manifest/` beside the path delta; the stage
declarations from the SCHEMA delta.

### `swarm manifest` in [[container-swarm-cli]] (added)

**Intent:** the command surface over the manifest — the four subcommands
[[frd-swarm-core]] specifies, and the first commands in the binary a
caller can run.

**Behavior:**
- The container SHALL provide `swarm manifest get`, `set`, `require` and
  `validate`.
- `set` SHALL accept several key-and-value pairs in one invocation.
- WHEN `set` is given several pairs, it SHALL apply them as one merge.
- `require` SHALL report whether the named keys are present, and nothing
  about the rest of the record.
- `validate` SHALL report the whole record's health against every rule
  the MANV delta declares.
- Each subcommand SHALL be built as a plan, so explain, dry-run and
  execute come from the Plan Runner rather than from the command.

**Anchors:** `cli/swarm/cli/manifest.go`.

### The end-to-end fixture (added)

**Intent:** a real repository and a real tracker for the upper test tier,
restored from committed artifacts rather than generated by code
([[adr-027]]). This delta delivers the fixture, the version stamp that
makes its kind mechanically checkable, and the harness that materializes
a copy per case — not the cases themselves, which belong to the commands
they assert.

**Behavior:**
- The fixture SHALL be committed as a `git bundle` holding the
  repository and a tracker export holding the run.
- The fixture SHALL record the version of [[contract-swarm-json]] it was
  recorded against.
- The core fixture's recorded version SHALL equal the current contract
  version.
- IF the core fixture's recorded version differs from the current one,
  THEN a check SHALL fail.
- A fixture recorded to prove an older shape is still readable SHALL
  record a version older than the current one.
- IF such a fixture's recorded version has caught up with the current
  one, THEN a check SHALL fail.
- WHEN a case runs, the harness SHALL materialize the fixture into a
  temporary directory of that case's own.
- WHEN a case finishes, the harness SHALL delete that directory.
- The harness SHALL NOT restore a fixture in place.
- The harness SHALL NOT roll a fixture back.
- A variant fixture SHALL be a branch derived from the core fixture in
  the same bundle.
- WHERE a case needs the tracker and the tracker is absent, the harness
  SHALL report the case as skipped with the reason.
- WHERE a case needs the tracker and the platform is one the project
  declares the tracker present on, an absent tracker SHALL fail the case
  rather than skip it.

**Anchors:** `cli/testdata/swarm-fixtures/` for the bundle, the export
and the stamp; `cli/testdata/swarm-cases/` for the cases; `cli/swarm/e2e`
for the harness; `.github/workflows/cli.yml` for the platform
declaration. The harness is its own package rather than an extension of
`cli/cmd`'s snapshot harness: that one drives a corpus in place through
the `spec` binary's own entry point, where this one restores a git bundle
and a tracker export into a temporary world per case and runs a compiled
binary as a subprocess, so the two share the idea and no code.

### `Status Collector` in [[container-swarm-cli]] (added)

**Intent:** take a run's measurements and return them as a value,
formatting nothing and deciding nothing. It is the largest derivation in
the family, and separating it from the report is what lets both halves be
tested without a subprocess.

**Behavior:**
- The Status Collector SHALL walk the run's descendants itself rather
  than relying on a parent-scoped tracker query.
- The Status Collector SHALL count every bead in the run, however nested.
- The Status Collector SHALL determine a dispatch's drain state from two
  queries.
- The Status Collector SHALL report a dispatch holding a bead in progress
  as still holding work.
- The Status Collector SHALL measure the standing worktree count, the
  last commit's timestamp and age, and the size of the ready set.
- The Status Collector SHALL compare the live queue against the
  manifest's admitted list.
- The Status Collector SHALL name the beads present in one of those two
  and not the other.
- The Status Collector SHALL report a bead that is blocked and whose
  dispatch's worktree still exists.
- IF a measurement's source cannot answer, THEN the Status Collector
  SHALL record that measurement as unestablished rather than as an empty
  answer.
- The Status Collector SHALL NOT read any stored progress or health
  value.

**Anchors:** `cli/swarm/status/`; the Tracker and Version Control
interfaces, not their implementations.

### `Status Reporter` and `swarm status` in [[container-swarm-cli]] (added)

**Intent:** turn a measurement set into the report a caller reads and the
exit status it implies — and never into a diagnosis. This is the delta
that proves the manifest component against a real run.

**Behavior:**
- The Status Reporter SHALL state the threshold it compared the last
  commit's age against.
- The Status Reporter SHALL state that whether a subagent is still
  working is not observable from outside the harness.
- The Status Reporter SHALL name the measurements it gives in place of
  that observation.
- The Status Reporter SHALL NOT state that a run has stalled.
- WHEN the ready set is non-empty, no worktree stands, and the last
  commit is older than the threshold in force, `swarm status` SHALL exit
  with the attention status.
- WHEN a measurement is unestablished, `swarm status` SHALL report it as
  unknown.
- WHEN a measurement is unestablished, `swarm status` SHALL exit
  non-zero.
- WHEN a run's admitted beads are all closed and its admission excluded
  others, the report SHALL name the excluded beads beside the completion
  verdict.
- `swarm status` SHALL accept `--stall-after` to set the threshold.
- `swarm status` SHALL state the threshold in force whether it was given
  or defaulted.

**Anchors:** `cli/swarm/status/` beside the collector;
`cli/swarm/cli/status.go`.

### [[module-bead-bonanza]] (modified — the command conventions)

**Intent:** the exit vocabulary and the printed-invocation property stop
being per-script prose and become the CLI's contract, so the paragraphs
that taught them are deleted rather than left to disagree with the
binary. Nothing here calls a swarm command yet; this delta retires prose
about how commands behave.

**Behavior:**
- The module's reference documents SHALL cite [[frd-swarm-core]]'s
  command contract for the exit vocabulary rather than restating it.
- WHEN a bundled script's exit vocabulary is documented, it SHALL be
  documented as the same vocabulary the CLI uses, not as its own.

**Anchors:** `plugin/skills/bead-bonanza/references/scripts.md`, the exit-code
table and the "prints the command that established it" paragraph;
`plugin/skills/bead-bonanza/references/dispatching.md`, the same two.

### [[module-bead-bonanza]] (modified — manifest adoption)

**Intent:** replace every place a skill reads or writes the run manifest
by hand with a `swarm manifest` call, and delete the prose that taught
the hand method. This is where the feature's value is realized for the
manifest: until a skill calls the command, the validation exists and
nothing is validated by it.

**Behavior:**
- WHEN a component reads a manifest key, it SHALL invoke `swarm manifest
  get` rather than the tracker directly.
- WHEN a component writes manifest keys, it SHALL invoke `swarm manifest
  set` once for the whole write rather than once per key.
- The module's documents SHALL NOT restate the manifest's key set, its
  types or its merge semantics; they SHALL cite the command that
  enforces them.
- WHERE a component required manifest keys before proceeding, it SHALL
  invoke `swarm manifest require` and route on the exit status.
- The `Coordinator` SHALL invoke `swarm manifest validate` when resuming
  a run it does not hold.

**Anchors:** `plugin/skills/bead-bonanza/references/run-record.md`, "The two
things you own" item 1 and the whole "Run manifest" section;
`plugin/skills/run-batch/SKILL.md`, `plugin/skills/sprint-review/SKILL.md`
and `plugin/skills/sprint-plan/SKILL.md`, their manifest-reading
preambles.

### [[module-bead-bonanza]] (modified — status adoption)

**Intent:** replace the derived-health and done-ness prose, and the
script sections that compute them, with `swarm status`. The prose being
retired is the largest single block this feature removes, and it is also
where the wrong answer was most expensive — a run once read healthy
through a five-hour outage.

**Behavior:**
- WHEN the `Coordinator` asks where a run is, it SHALL invoke `swarm
  status` and route on its exit status.
- The module's documents SHALL NOT describe how to derive a run's
  progress, its completion or a dispatch's drain state; they SHALL cite
  the command that measures them.
- The module SHALL NOT describe a run as stalled on the strength of an
  elapsed time.
- The module SHALL report the measurements `swarm status` returns.

**Anchors:** `plugin/skills/bead-bonanza/references/supervision.md`, "Where is
this run" and "When the run is done";
`plugin/skills/bead-bonanza/SKILL.md`, "The run's health — derived, never
asserted"; `plugin/skills/bead-bonanza/scripts/bb-status.sh`, the `run`
subcommand the CLI supersedes.

### Relied-upon elements (unchanged)

- [[container-spec-cli]] — shares the `cli` module, its `go.mod`, its
  toolchain pin and its release workflow; this feature adds a binary
  beside it and changes nothing in it.
- [[module-document-core]] and [[module-spec-profile]] — in the same
  module, untouched; named so a regression in them is recognised as
  this feature's blast radius rather than a coincidence.
- [[testing-standards]] — the snapshot-case structure the end-to-end
  cases are written in, and the layer definitions the AC coverage table
  names.
- [[ac-traceability-standards]] — the citation obligation every test in
  this feature carries.
- `bd` and `git` on `PATH` — invoked, never vendored; their command
  surfaces are what the two implementations depend on.

## Contracts

- [[contract-swarm-json]] (new, draft) — the JSON every command emits
  under `--json` and the run manifest the CLI reads and writes; vended to
  the `bead-bonanza` skill family. Its shape is generated from the two
  Strata schemas the SCHEMA delta adds, and the Definition of Done flips
  it to active.

## Tracker Reconciliation

- `agent-plugins-offx.6` — subsumed by the **TRK** delta: it asks for a
  capability returning a run's beads transitively, with a test pinning
  both measured tracker behaviours. The Beads Tracker's transitive bead
  set is that capability, and its decoder tests are that pin.
- `agent-plugins-offx.4` — subsumed by the **PLAN** delta for the
  ordering invariant it names; the guard it asks for is a step's failure
  policy honoured by an interpreter. The other invariants it gestures at
  belong to commands outside this FRD and remain open.
- `agent-plugins-offx.14` — subsumed by this feature as a whole: it
  asks that the CLI be scoped as a thin capability layer rather than an
  orchestrator, and [[container-swarm-cli]] records that boundary. Close
  it against this blueprint rather than implementing it.
- `agent-plugins-offx.3` — out-of-scope: retiring `AGENTS.md`'s
  measure-at-the-point-of-use section depends on tool claims having
  become tests across the whole family, not only in these three
  commands. This feature makes it possible and does not discharge it.
- `agent-plugins-offx.17` and `agent-plugins-offx.21` — out-of-scope:
  the circuit breaker's unit and its manifest key belong to the dispatch
  document's queue command. This feature SHALL NOT add the key
  speculatively; the SCHEMA delta carries the keys the run record
  specifies today and no others.
- `agent-plugins-im99` — prerequisite for authoring only, already
  discharged: the feature-blueprint template prescribes a slug the
  profile rejects. This document uses the profile's form.

## Sequencing Constraints

Delivery is **staged**. Each stage below is a deliverable increment, and
its boundary is a constraint rather than a suggestion: the stage's
verification gates pass before the next stage begins.

- **Stage 1 — the substrate.** The SKEL delta SHALL land before every
  other delta. The PROC delta SHALL land before the PLAN delta. The PLAN
  and SURF deltas SHALL land before any delta that adds a command.
- **Stage 2 — the tools.** The PROC delta SHALL land before the TRK and
  VCS deltas. The TRK and VCS deltas MAY land in either order.
- **Stage 3 — the manifest.** The SCHEMA delta SHALL land before the MAN
  and MANV deltas. The MAN delta SHALL land before the MANV delta, which
  extends the record the MAN delta reads. The TRK delta SHALL land
  before the MAN delta, whose store reaches beads through it. The MANCMD
  delta SHALL land after MAN and MANV.
- **Stage 4 — the fixture.** The FIX delta SHALL land before any
  end-to-end case in any other delta. WHILE the FIX delta is
  outstanding, a command delta's acceptance rests on its unit tests, and
  its end-to-end case is added when the fixture exists.
- **Stage 5 — manifest adoption.** The ADOPT-CONV delta SHALL land after
  the SURF delta. The ADOPT-MAN delta SHALL land after the MANCMD and
  FIX deltas. The ADOPT-MAN delta SHALL land before the COLL delta, so
  the manifest command is under real use while the status command is
  built.
- **Stage 6 — status.** The COLL delta SHALL land after the TRK, VCS and
  MAN deltas, all of which it reads through. The REP delta SHALL land
  after the COLL delta.
- **Stage 7 — status adoption.** The ADOPT-STAT delta SHALL land after
  the REP delta.
- No adoption delta SHALL land before the command it adopts is verified
  at the level its AC coverage declares.
- The SCHEMA delta's generated artifacts and its `.strata` sources SHALL
  be updated in the same change, per [[container-swarm-cli]]'s lockstep
  rule; both are in that one delta, so no cross-delta coordination is
  required for it.
- WHEN the MANV delta changes the declared key set, the FIX delta's
  fixture SHALL be re-recorded in the same change. WHILE the FIX delta
  has not landed, that obligation does not yet apply.
- [[contract-swarm-json]] SHALL reach `active` only after every delta
  that emits or consumes one of its surfaces has landed.

## AC Coverage

Legend: **SKEL** = [[container-swarm-cli]] delta (— binary skeleton and
build wiring); **PROC** = `Process Runner` delta; **PLAN** = `Plan
Runner` delta; **SURF** = `Swarm Command Surface` delta; **TRK** =
`Tracker` and `Beads Tracker` delta; **VCS** = `Version Control` and
`Git Version Control` delta; **SCHEMA** = Strata schemas delta; **MAN** =
`Manifest` delta (— the read and write path); **MANV** = `Manifest`
delta (— the logical rule set); **MANCMD** = `swarm manifest` delta;
**FIX** = end-to-end fixture delta; **COLL** = `Status Collector` delta;
**REP** = `Status Reporter` and `swarm status` delta; **ADOPT-CONV** =
[[module-bead-bonanza]] delta (— the command conventions); **ADOPT-MAN**
= [[module-bead-bonanza]] delta (— manifest adoption); **ADOPT-STAT** =
[[module-bead-bonanza]] delta (— status adoption).

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-SWARM-CORE-001.1 | **SURF (proves)** | unit |
| AC-SWARM-CORE-001.2 | **SURF (proves)** | unit |
| AC-SWARM-CORE-001.3 | SURF, **REP (proves)** | e2e |
| AC-SWARM-CORE-001.4 | **PLAN (proves)**, SURF | unit |
| AC-SWARM-CORE-001.5 | **PLAN (proves)** | unit |
| AC-SWARM-CORE-001.6 | PLAN, **MANCMD (proves)** | e2e |
| AC-SWARM-CORE-001.7 | **PLAN (proves)**, SURF | unit |
| AC-SWARM-CORE-001.8 | PLAN, **SURF (proves)** | unit |
| AC-SWARM-CORE-001.9 | SCHEMA, **SURF (proves)** | unit |
| AC-SWARM-CORE-001.10 | **PROC (proves)** | unit |
| AC-SWARM-CORE-001.11 | **PROC (proves)** | unit |
| AC-SWARM-CORE-001.12 | **PROC (proves)** | unit |
| AC-SWARM-CORE-001.13 | PLAN, **FIX (proves)** | e2e |
| AC-SWARM-CORE-001.14 | **SURF (proves)** | unit |
| AC-SWARM-CORE-001.15 | **SURF (proves)** | unit |
| AC-SWARM-CORE-002.1 | **MAN (proves)** | unit |
| AC-SWARM-CORE-002.2 | SCHEMA, **MAN (proves)** | unit |
| AC-SWARM-CORE-002.3 | **MAN (proves)** | unit |
| AC-SWARM-CORE-002.4 | MAN, **MANV (proves)** | unit |
| AC-SWARM-CORE-002.5 | SCHEMA, **MAN (proves)** | unit |
| AC-SWARM-CORE-002.6 | **MANV (proves)** | unit |
| AC-SWARM-CORE-002.7 | **MANV (proves)** | unit |
| AC-SWARM-CORE-002.8 | **MANV (proves)** | unit |
| AC-SWARM-CORE-002.9 | SCHEMA, **MAN (proves)** | unit |
| AC-SWARM-CORE-002.10 | **MAN (proves)** | unit |
| AC-SWARM-CORE-002.11 | SCHEMA, **MANV (proves)** | unit |
| AC-SWARM-CORE-002.12 | **MANV (proves)** | unit |
| AC-SWARM-CORE-002.13 | MAN, **MANCMD (proves)** | integration |
| AC-SWARM-CORE-002.14 | **MAN (proves)**, MANCMD | unit |
| AC-SWARM-CORE-002.15 | **MANV (proves)** | unit |
| AC-SWARM-CORE-002.16 | MANV, **MANCMD (proves)** | e2e |
| AC-SWARM-CORE-002.17 | SCHEMA, **MANV (proves)** | unit |
| AC-SWARM-CORE-002.18 | **MANCMD (proves)** | unit |
| AC-SWARM-CORE-002.19 | **SCHEMA (proves)**, MAN, MANV, MANCMD | unit |
| AC-SWARM-CORE-003.1 | COLL, **REP (proves)** | e2e |
| AC-SWARM-CORE-003.2 | **COLL (proves)** | unit |
| AC-SWARM-CORE-003.3 | TRK, **COLL (proves)** | unit |
| AC-SWARM-CORE-003.4 | **COLL (proves)** | unit |
| AC-SWARM-CORE-003.5 | COLL, **REP (proves)** | unit |
| AC-SWARM-CORE-003.6 | **REP (proves)** | unit |
| AC-SWARM-CORE-003.7 | MAN, **COLL (proves)** | unit |
| AC-SWARM-CORE-003.8 | COLL, **REP (proves)** | e2e |
| AC-SWARM-CORE-003.9 | VCS, **COLL (proves)** | unit |
| AC-SWARM-CORE-003.10 | MAN, **REP (proves)** | unit |
| AC-SWARM-CORE-003.11 | **REP (proves)** | unit |

The three adoption deltas carry no acceptance criterion of their own.
[[frd-swarm-core]] specifies a CLI, not the skills that call it, so what
they deliver is verified by the Definition of Done's prose-retirement
entries rather than by an AC — which is why each is a delta block a
reviewer can see rather than a line inside another one.

## Key Decisions

- [[adr-023]] — Go, as a second binary in the existing `cli` module,
  which is what the SKEL delta wires up.
- [[adr-024]] — a command builds a plan of step values, and explain,
  dry-run and execute are three interpreters over it; the PLAN delta.
- [[adr-025]] — every external invocation is an argv vector through one
  runner, with separated streams and a per-subcommand decoder; the PROC,
  TRK and VCS deltas.
- [[adr-026]] — no rollback: validate before writing, order the
  irreversible writes last, declare each step's idempotence; why the MAN
  delta validates a whole multi-key write before applying any of it.
- [[adr-027]] — two test tiers, the committed fixture, and how that
  fixture evolves; the FIX delta.
- [[adr-016]] — Strata as the one source for schemas and generated
  models; the SCHEMA delta.

## Definition of Done

- [x] Every delta above is folded into its owning blueprint:
  - [x] [[container-swarm-cli]] — the Components roster reflects what
        was built, the diagram matches it, and Data names the fixture.
        Twelve components in the roster, twelve `###` sections and twelve
        `Component(...)` nodes in the diagram, one per built package
        under `cli/swarm`. Data gained its third entry, the end-to-end
        fixtures under `cli/testdata/swarm-fixtures/` — `core`,
        `status-attention` and `status-narrowed-scope`
  - [x] [[module-bead-bonanza]] — the components that now call the CLI
        say so, and the "Consumes" note carries no "planned". **The
        ADOPT-MAN adoption is incomplete by decision, and the module
        blueprint records the exception rather than claiming a clean
        sweep:** `scripts/pre-flight-check.sh` still merges into the run
        epic's metadata itself, so the four keys it writes bypass the
        command's validation. Every other reader and writer of the
        manifest in the module goes through `swarm manifest` — measured:
        the script's own `say` line is the last `bd show … .metadata.…`
        anywhere under `plugin/`
- [x] [[contract-swarm-json]] status: draft → active
- [x] The Strata sources and their generated artifacts are in step, and
      `scripts/generate-schemas.sh --check` is green — the `schemas`
      check in the project's declared set
- [x] New ADRs are referenced from the owning blueprints' Key Decisions
      — [[adr-023]] through [[adr-027]], all `approved`, all five named
      in [[container-swarm-cli]]'s Key Decisions
- [x] Every AC in the coverage table is verified at its declared level
      and cited from its tests per [[ac-traceability-standards]] — all
      45 carry a citation, and the FRD's AC set and this table's are the
      same 45. Thirty-nine are cited by a `// AC:` comment on a unit
      test; six by a snapshot case manifest's `ac` key, which
      [[testing-standards]] Acceptance-Criteria Traceability accepts
      equally. Each of the six declared `e2e` is cited from an
      end-to-end case. **One exception:** AC-SWARM-CORE-002.13 is
      declared `integration`, a level [[testing-standards]] does not
      define — it names two layers, unit tests and snapshot cases. The
      criterion is verified (`cli/swarm/cli/manifest_test.go`'s
      `TestSetAppliesSeveralKeysAsOneMerge` and
      `cli/swarm/manifest/beads/store_test.go`'s
      `TestSeveralKeysInOneInvocationAreOneReadAndOneWrite`, both across
      the command surface and the store with a fake process runner);
      what has no definition is the word this table declares it at
- [x] The prose each delta's "Anchors" names is deleted, not merely
      superseded — a paragraph left beside the command that replaced it
      is the drift this feature exists to end
- [ ] `agent-plugins-offx.6`, `agent-plugins-offx.4` and
      `agent-plugins-offx.14` are closed or re-scoped against the deltas
      that subsume them — **partially.** `offx.14` is closed: its
      boundaries are stated in [[frd-swarm-core]]'s "What this document
      does not do" and in [[container-swarm-cli]]'s Overview. `offx.6`
      and `offx.4` are re-scoped and stay open: the TRK delta shipped
      the transitive walk `offx.6` asked for, but neither the test
      pinning both measured `bd` behaviours nor the last run-scope
      `bd list --parent` call site went with it, and no delta here
      touches the guards `offx.4` is about
- [ ] The Windows leg has run at least once against the tracker install
      [[adr-027]] targets, and its result is recorded — **not met, by
      circumstance rather than by omission.** The leg exists
      (`.github/workflows/cli.yml`, the dolt and `bd` install steps
      under `if: matrix.os == 'windows-latest'`) and its pieces were
      measured individually, but no run has executed it: that workflow
      triggers only on a push to `main`, a pull request, or manual
      dispatch, and this feature's branch has neither merged nor opened
      one. The first non-draft pull request settles it
- [x] This blueprint's status → retired

## Open Questions

None.

## References

- [[frd-swarm-core]] — the requirements this design realizes
- [[container-swarm-cli]] — the container every delta but the adoptions
  changes
- [[module-bead-bonanza]] — the module the adoption deltas change
- [[c4-standards]] — the definitions the delta blocks are written against
- `docs/research/swarm-cli/2026-08-25-capability-audit.md` — the audit
  behind the requirements, and the measured tool behaviours the decoders
  pin
