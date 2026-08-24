# Rolling coordination onto the beads graph

Implementation plan for items **1–14** of the ranked list in
[beads-coordination-design.md](beads-coordination-design.md). Each item below is
sized to become one bead, with the evidence already recorded in the design doc
and not repeated here.

Transient. Delete when the change lands.

---

## 7 · The run-epic shape, and the constraint that governs it

**In scope, and it changes how items 1 and 10 are justified** — both are revised
below rather than left as written.

### The shape

```
run epic  (container; NEVER swarmed)
├── preflight   (epic)  ─┐
├── scope       (epic)  ─┤  ceremony: epics, chained epic→epic
├── plan        (epic)  ─┤
├── TARGET      (epic)  ←─  swarmed. FLAT: tasks only.
├── review      (epic)  ─┤
├── retro       (epic)  ─┤
├── report      (epic)  ─┤
└── postflight  (epic)  ─┘
```

Forced, not chosen: `bd dep add` refuses `task→epic` and `epic→task`, so anything
that must block the target epic has to be an epic itself. One `epic→epic` edge
then blocks every bead under the target, and beads added mid-run inherit that
block structurally.

### The constraint: the swarmed target is FLAT, and swarm will not tell you

**Measured:**

| target shape | `bd swarm validate` reports |
|---|---|
| epic → 2 tasks | `Total issues: 2` · Wave 1: both leaves |
| epic → child epic → 2 tasks | `Total issues: 1` · Wave 1: **the child epic** |

A nested target makes swarm count the child epic as one issue and never see the
real beads beneath it. Progress is wrong, the ready-front analysis treats a
container as work, and `validate` — whose whole job is checking epic structure
for swarming — **reports the wrong total as though it were correct.**

So: **the target contains tasks only.** Ceremony epics are children of the run
epic, which is never swarmed. Work that genuinely decomposes into sub-epics
becomes a *second* target under the run epic, swarmed separately. It is an
admission rule too — scope creep arriving as an epic silently removes everything
beneath it from the swarm's view.

### Gates attach to task children, never to the ceremony epic

Measured: a `[steps.gate]` on an epic cannot be created (the gate bead is a task,
and gating an epic needs the refused `task→epic` edge). So the authorization gate
lives on a **task child** of the `scope` epic:

```
scope (epic)                 ← blocks the target epic
└── authorization (task)     ← bd gate create --blocks <this>
```

The epic cannot become close-eligible until its children close; the gated child
cannot close until a human resolves it. Verified end to end.

**Change.** The run shape, its creation, and the swarm call.

**Files.** `plugin/skills/bead-bonanza/SKILL.md` (the shape and the swarm call),
`plugin/skills/sprint-scope/SKILL.md` (run epic + ceremony epics + the
authorization gate on a task child — **and this is what WRITES `target_epic`
into the manifest**, because `sprint-plan`'s run-record forbids that skill from
writing the manifest at all: "The manifest is the orchestrator's, not yours …
Read it; never rewrite it"), `plugin/skills/sprint-plan/SKILL.md` (the
target is flat; emit it that way),
`plugin/skills/sprint-plan/references/run-record.md` (the `target_epic` key
*spec* — where the key is specified, not where it is written).

**Verify.** Build the shape in a scratch database; assert (a) every bead under
the target is absent from `bd ready` until pre-flight's epic closes via
`bd epic close-eligible`, (b) `bd swarm validate <target>` reports the true leaf
count, (c) a bead added to the target mid-run is blocked without being wired.

**Depends on.** 6 — the graph plan is what builds this shape. **1 is a hard
prerequisite**, not a nicety: see below.

**Size.** L.

---

## Revised: items 1 and 10 under item 7

**Item 1 returns to critical, and becomes a prerequisite.** With ceremony epics
in play the silent deadlock is live: ceremony beads all close, the ceremony epic
stays `open` because nothing closes it, the target stays blocked forever,
`bd ready` returns empty, and an is-the-front-empty termination check reads that
as **success**. The run reports done having done nothing. So item 1 is not
hygiene — **it must land before item 7, or item 7 ships a deadlock.**

**Item 10 reverts to the design's mechanism.** The deferred-review bead parks
under the **`review` or `report` ceremony epic** rather than taking a task→task
edge to the report bead. Same asymmetry — absent from `bd ready` until that phase
opens, present in `bd human list` immediately — and it needs no special-case
edge now that the late epics exist.

---

---

## Sequencing

```
1 ─────────────→ 7        1 is a HARD prerequisite: 7 without it ships a deadlock
2 ─┬─→ 3                  3 enforces what 2 and 4 decide
4 ─┘
5     independent

6 ─┬─→ 7                  the graph plan builds the run shape
   ├─→ 8, 9               both consume the graph 6 emits
   └─→ 11                 merge beads are graph nodes

7 ─────→ 10               deferred review parks under a late ceremony epic

12 ─┐
13 ─┼─ all three need the DISTRIBUTION decision below settled first
14 ─┘
```

**Two decisions gate work rather than blocking on evidence:**

- **D1 — where hooks live** (items 12–14). The `sf` plugin ships `skills/` only;
  it has no `hooks/` or `monitors/`. So either these land in this repository's
  `.claude/settings.json` (works here, travels nowhere) or in a new
  `plugin/hooks/hooks.json` (travels, but every adopter inherits them). This
  needs answering before 12–14 start, and it is a judgement, not a measurement.
- **D2 — one worker pool size** (item 2). The cap is 20 by default but 2 under
  the Workflow tool. Pick the number the skill assumes, or make it resolved at
  run time and recorded.

---

## Do first

### 1 · `bd epic close-eligible` in the coordinator loop

**Change.** The loop becomes two commands: `bd epic close-eligible`, then the
ready query. Both the loop statement in `bead-bonanza/SKILL.md` and the
supervision reference must agree — today only one of them describes the loop.

**Files.** `plugin/skills/bead-bonanza/SKILL.md`,
`plugin/skills/bead-bonanza/references/supervision.md`.

**Verify.** Close every child of an epic; assert the epic is still `open`; run
`bd epic close-eligible`; assert it closed. That is the whole behaviour.

**Risk.** None — the command is idempotent and closes only epics whose children
are all complete.

**Size.** XS. Two prose edits.

---

### 2 · Cap distinct assignees; record the cap in the manifest

**Change.** Planning must not create more distinct assignees than the harness
concurrency cap. Add `concurrency_cap` to the run manifest, sourced from
`CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` where set and defaulting to 20.

**Files.** `plugin/skills/sprint-plan/references/run-record.md` (the manifest
key spec — the *only* place manifest keys are specified; `bead-bonanza`'s
run-record covers run health and `base_sha`, not keys),
`plugin/skills/sprint-plan/SKILL.md` (the rule).

**Verify.**

```sh
bd list --parent {RUN_EPIC} --json \
  | jq -r '[.[].assignee] | map(select(. != null and . != "")) | unique | length'
# must be <= manifest.concurrency_cap
```

**Risk.** Picking the wrong default silently over- or under-subscribes. Mitigated
by recording the value rather than assuming it (D2).

**Size.** S.

---

### 3 · Two Definition-of-Ready checks

**Change.** Add to the epic-level DoR: (a) every bead in the run epic carries an
assignee; (b) the **concurrent** assignee count — the `dev` pool — is
≤ `concurrency_cap − 1`. Both fail loudly at plan time; both are invisible at
run time.

**Count the dev pool, not all distinct assignees** (see item 4's roster).
Ceremony agents are serialized by the graph and share one slot; counting them
would under-subscribe the pool by the number of ceremony roles.

**Files.** `plugin/skills/groom-beads/references/epic-definition-of-ready.md`,
and the readiness gate in `plugin/skills/sprint-plan/SKILL.md`.

**Verify.** Construct a run epic with one unassigned bead → the gate fails and
names the bead. Construct one with `cap+1` assignees → the gate fails and names
the count.

**Depends on.** 2 (for the cap) and 4 (for what an assignee *means*).

**Size.** S.

---

### 4 · `BEADS_ACTOR` in every agent and subagent

**Change.** Every dispatch prompt exports `BEADS_ACTOR=<agent id>` before any
`bd` call. Measured on `bd` 1.2.2, a claim with the variable unset silently
rewrites `assignee` on an **unassigned** bead (exit 0, the git `user.name`
fallback) and is refused loudly on an assigned one (exit 1, `issue already
claimed by <id>`). So this export and item 3's “every bead assigned before
dispatch” check are belt and braces rather than alternatives — the export
covers the loud arm, the assignment closes the silent one.

**The roster is wider than the design doc said, and item 7 is why.** Once
ceremony is beads, each ceremony epic is work someone does — so there is one
agent kind per agent the coordinator spawns:

| agent | id | count | concurrent? |
|---|---|---|---|
| coordinator | `coordinator-<run-epic-id>` | 1 | — |
| groom | `groom-<run-id>` | 1 | no |
| plan | `plan-<run-id>` | 1 | no |
| dev | `dev-<run-id>-<batch-id>` | N | **yes** |
| merge | `merge-<run-id>` | 1 | no |
| review | `review-<run-id>` | 1 | no |
| retro | `retro-<run-id>` | 1 | no |
| report | `report-<run-id>` | 1 | no |

**This interacts with item 2 and the interaction is easy to get wrong.** The
concurrency cap bounds *simultaneous* agents. Ceremony agents run one at a time,
in a chain the graph already serializes — so they cost one slot between them,
not seven. **Only the `dev` pool is concurrent**, so item 3's check should count
concurrent assignees, not distinct ones. Counting all distinct assignees would
under-subscribe the pool by the number of ceremony roles.

The practical rule: `N_dev ≤ cap − 1` (one slot for whichever ceremony agent is
running).

**Ceremony beads are assigned like any other**, which is what makes
`bd ready --assignee groom-<run-id>` the groom agent's queue and keeps the
dispatch mechanism uniform.

**Files.** `plugin/skills/bead-bonanza/references/prompt-per-bead.md`,
`prompt-batch.md`, the placeholder table in
`plugin/skills/sprint-wave/SKILL.md`, and — once item 7 lands — wherever each
ceremony agent is dispatched from.

**Verify.** In a scratch database, three arms — and the negative control must
start from an **unassigned** bead, which is the one entry point the failure
has:

1. Assign a bead to `dev-x` and claim it with `BEADS_ACTOR=dev-x`. Assert exit
   0 and the assignee unchanged.
2. Claim that same assigned bead with the variable unset. Assert exit 1,
   `issue already claimed by dev-x`, and the assignee unchanged — this arm is
   refused loudly, so it is not the control.
3. Claim an **unassigned** bead with the variable unset. Assert exit 0 and
   `assignee` silently set to the git `user.name` fallback. This is the
   negative control, because this is the failure that is silent.

**Risk.** None. Additive.

**Size.** S.

---

### 5 · `bd setup claude`, reviewed not applied

**Change.** `bd setup claude --check` reports the installed CLAUDE.md block is
stale. Regenerate and review.

**Files.** `CLAUDE.md`.

**Verify.** `bd setup claude --check` reports current.

**Risk. This one is not routine.** `CLAUDE.md` carries a great deal of
hand-written project content — the session-bootstrap instructions, the
multi-root hook explanation — that is not bd's to manage. **Do not blind-apply.**
Run it, read the whole diff, and keep any hand-written content the tool would
drop. If the tool cannot update its block without touching ours, the right
outcome is a bug report upstream and no change here.

**Size.** S, with a real chance it becomes "declined, reported upstream".

---

## The core rebuild

### 6 · Planning emits a `bd create --graph` JSON plan

**The keystone.** Planning's output stops being prose plus a wave table and
becomes one JSON document that builds the run.

**Change.** `sprint-plan` emits a graph plan; the coordinator loads it with
`bd create --graph`. Schema (measured against bd 1.2.x; corrected copy of
design §4a's schema, which shows the same fields with no orientation prose —
see `agent-plugins-fvnp`):

```json
{"nodes":[{"key":"…","title":"…","type":"epic|task","parent_key":"…",
           "assignee":"…","priority":2,"description":"…","metadata":{"tier":"…"}}],
 "edges":[{"from_key":"…","to_key":"…","type":"blocks"}]}
```

Without item 7 the emitted shape is flat: run epic, work beads as its children,
edges between them. The ceremony epics arrive with item 7 and the schema does
not change.

**`from_key` WAITS FOR `to_key` — the opposite of how `"type": "blocks"` reads
at a glance.** Writing an edge so the `from_key` node is the one that runs
*first* inverts every dependency in the run, exits 0, and produces a `bd
ready` front that still looks plausible. Measured in a scratch database:

```
$ cat edge-orientation.json
{"nodes": [{"key": "alpha", "title": "alpha bead"}, {"key": "beta", "title": "beta bead"}],
 "edges": [{"from_key": "alpha", "to_key": "beta", "type": "blocks"}]}
$ bd create --graph edge-orientation.json
Created 2 issues
  alpha -> scr-d4c
  beta -> scr-h6j
$ bd ready
  ○ scr-h6j ● P2 beta bead
$ bd blocked
  [● P2] scr-d4c: alpha bead
    Blocked by 1 open dependencies: [scr-h6j]
```

`alpha` is the `from_key` and comes out **blocked**; `beta` is the `to_key`
and comes out **ready**. Same result, independently, in
`plugin/skills/sprint-plan/references/run-record.md` ("The graph plan" —
merged earlier this run and now the source of truth this section agrees
with) and the negative control it documents. Assert the ready front after
loading rather than trusting the file.

**`parent_key` and `parent_id` are not interchangeable, and `bd` gives no
warning when they're confused.** `parent_key` references another node's
`key` **in this plan** — it is what builds the run's hierarchy. `parent_id`
is for hanging a plan off an id that **already exists** in the tracker.
Measured:

```
$ cat parent-id-trap.json
{"nodes": [{"key": "child_via_parent_id", "title": "child via parent_id", "parent_id": "some-plan-key"}], "edges": []}
$ bd create --graph parent-id-trap.json
Created 1 issues
  child_via_parent_id -> scr-odc
$ bd show scr-odc            # no PARENT section in the output
$ bd dep list scr-odc
scr-odc has no dependencies
```

Giving `parent_id` a plan key instead of a real tracker id is accepted with
**no warning** and produces an orphan: the node lands outside the intended
epic entirely, and every `--parent`-scoped query then under-reports the
queue. Use `parent_key` for every node this plan creates; reserve `parent_id`
for attaching to a bead that already exists.

**Files.** `plugin/skills/sprint-plan/SKILL.md` (the output contract),
`plugin/skills/sprint-plan/references/run-record.md` (where the plan file lives
and how it is recorded), `plugin/skills/bead-bonanza/SKILL.md` (the load step),
`plugin/skills/bead-bonanza/references/dispatch-policy.md` (the agent-id
roster — the plan writes an `assignee` on every node, and items 9, 11 and 13
all read those ids back, so the emitter must mint them from the roster rather
than invent a spelling).

**Verify.** Emit a plan for a real epic, load it into a scratch database, and
assert: node count matches, every edge present, every bead assigned, and
`bd ready` returns exactly the expected front. **Diff the plan before loading** —
being reviewable pre-load is most of the value.

**Resolved (was Open).** Whether `--graph` accepts gates or per-node metadata
is no longer unverified: it accepts per-node `metadata` (`map[string]string`
— a number or nested object aborts the load with a parse error before
anything is created) but it does **not** accept gates. Measured:

```
$ cat fields-probe.json
{"nodes": [{"key": "n1", "title": "field probe", "metadata": {"tier": "advanced"}, "gate": "human", "acceptance": "x", "notes": "y"}], "edges": []}
$ bd create --graph fields-probe.json
warning: graph plan node["n1"] has unknown field(s): [acceptance gate notes] (silently dropped — see 'bd create --graph' schema)
Created 1 issues
  n1 -> scr-9j0
$ bd show scr-9j0 --json | jq .metadata
{"tier": "advanced"}
```

`metadata` landed on the created bead; `gate`, `acceptance` and `notes` were
silently dropped — warned, not refused, exit 0. Accepted node fields in full:
`key title type parent_key parent_id assignee priority description labels
metadata`. Rejected (dropped with a warning, never an error): `id blocked_by
acceptance deps design estimate external_ref gate notes status`. Edges accept
only `from_key to_key type`. **The consequence for item 10: gates must be
created against already-loaded beads (`bd gate create`), never declared
inside the graph plan.**

**Risk.** A malformed plan fails at load having created part of the graph.
`bd create --graph` reports unknown *fields* as warnings and **silently drops
them** at exit 0 — so a typo in a key name is a silently missing assignee, not
an error. A malformed *value* (a non-string `metadata` entry, for instance) is
the opposite: it aborts the load before creating anything. The fail-open case
— a warning that still exits 0 — is the one a validator has to cover; a bad
value already fails loud on its own. Validate the plan against the known
field set (above) before loading.

**Validate flatness too, now rather than later** (see the item-7 constraint
above): no node parented to the target epic may be `type: "epic"`. This costs
one check in the emitter and prevents a class of failure that only shows up as
a quietly wrong swarm total once item 7 lands.

**Size.** L. This is the item with real design left in it.

---

### 8 · Collisions as edges; delete the wave table

**Change.** The C4 co-edit computation stops producing a wave table and starts
producing dependency edges in the graph plan. Everything describing waves goes.

**Files.** `plugin/skills/sprint-plan/SKILL.md` (the emitter — it is what
writes the edges and their policy label),
`plugin/skills/sprint-plan/references/run-record.md` (wave table section),
`plugin/skills/bead-bonanza/SKILL.md`,
`plugin/skills/bead-bonanza/references/dispatch-policy.md`,
`plugin/skills/sprint-recovery/SKILL.md`,
`plugin/skills/bead-bonanza/scripts/bb-status.sh` (its `wave` mode emits
wave-shaped status and goes with the table), and `sprint-wave` in its
entirety — this item is what deletes that skill.

**Label the policy edges.** A collision edge is a *policy* expressed as a
*dependency*, which is the conflation this design rejects elsewhere. Mark them
so a reader can tell them from real dependencies — a label on the edge's beads,
or a convention in the plan file.

**Verify.** Two beads sharing a `c4_changes` component appear in the plan with
an edge between them, and never appear in `bd ready` simultaneously.

**Depends on.** 6.

**Size.** L. Touches eight files and removes a skill.

---

### 9 · Assignment replaces batch bookkeeping

**Change.** "Is this batch done?" becomes `bd ready --assignee X` empty with
nothing `in_progress`. Delete `dispatch_state`, `dispatched_at` and `worktree`
metadata, the return-line contract, and the acceptance-signal machinery that
existed to check a subagent's testimony.

**Files.** `plugin/skills/sprint-batch/SKILL.md` (the writer — it is what sets
`dispatch_state`, `dispatched_at` and `worktree`),
`plugin/skills/sprint-wave/SKILL.md`,
`plugin/skills/bead-bonanza/references/recovery.md`, `scripts.md`,
`prompt-per-bead.md`, `prompt-batch.md` (it carries the return-line contract
this item deletes), `plugin/skills/bead-bonanza/SKILL.md` (it reads
`dispatch_state=running` to tell a live worker from a dead one),
`plugin/skills/sprint-recovery/SKILL.md`, and the scripts
`bb-status.sh` and `sprint-cleanup.sh`.

**Verify.** Assign three beads to one worker; close two; assert
`bd ready --assignee` returns one and the batch reads unfinished. Close the
third; assert empty and the batch reads finished.

**Careful.** `bb-status.sh` implements the five acceptance signals and is
depended on by recovery. Removing signals it computes is not a prose edit —
check each caller.

**Depends on.** 4 (agent ids must be stable first).

**Size.** L.

---

### 10 · Gates for authorization; the deferred-review channel

**Change.** The authorization gate becomes a real `bd gate create --type=human`.
Add the deferred-review channel: an agent that wants a human to look at a
decision later files a bead labelled `human`, blocking the report bead (see the
item-7 note above).

`decision-required` is **already gone** — measured, zero occurrences
repo-wide — so this item does not include retiring it.

**Files.** `plugin/skills/sprint-scope/SKILL.md` (authorization),
`plugin/skills/run-bead/SKILL.md` (how a worker files a deferred decision),
`plugin/skills/sprint-report/SKILL.md` (surface open `human` beads before the
report is handed over).

**Verify.** File a `human`-labelled bead blocking the report bead; assert absent
from `bd ready` and present in `bd human list`; `bd human respond`; assert the
report bead becomes ready.

**Size.** M.

---

### 11 · Merge beads and the merge slot

**Change.** A merge bead per batch, assigned to `merge-<run-id>`, depending on
every bead whose branch it lands. The merge agent holds the slot across the
merge. Replaces the merge-queue prose.

Register `merge-request` as a custom type — **but note it is local config**, so
it is a per-clone setup step and a clone that has not run it cannot create or
filter those beads. If that is unacceptable, use `task` and a label.

**Files.** `plugin/skills/bead-bonanza/references/worktree-mechanics.md`,
`plugin/skills/bead-bonanza/SKILL.md`,
`plugin/skills/sprint-plan/SKILL.md` — a merge bead is a graph node, so the
plan emitter is what creates it and writes its `merge-<run-id>` assignee;
`plugin/skills/bead-bonanza/references/dispatch-policy.md`, the roster's only
home and therefore where `merge-<run-id>` is minted; and
`plugin/skills/bead-bonanza/references/project-setup.md`, which is where a
per-clone setup step has to be written down — an unregistered `merge-request`
type is a clone that cannot create or filter these beads at all.

**Verify.** Two work beads and a merge bead depending on both: the merge bead is
absent from `bd ready` until both close, then present, assigned to the merge
agent.

**Known limitation to document, not fix.** The slot is repository-wide and
branch-unaware, so concurrent runs on different branches contend. Live with it;
record contention if it appears.

**Depends on.** 6 (merge beads are graph nodes).

**Size.** M.

---

## Harness integration — blocked on D1

All three need the distribution decision above. None should start until it is
made, because the answer changes where the files go and whether a fallback is
required.

### 12 · `Stop` hook: exit 2 while `bd ready` is non-empty

**Change.** A hook that refuses to let the coordinator stop while its run epic
has ready work. Makes the loop structural rather than remembered.

**Sketch.**

```sh
left=$(bd ready --parent "$RUN_EPIC" --exclude-type=epic --json | jq 'length')
[ "$left" -eq 0 ] || exit 2
```

**Files.** `plugin/hooks/hooks.json` (new) or this repository's own
`.claude/settings.json`, per D1;
`plugin/skills/bead-bonanza/scripts/pre-flight-check.sh` — **nothing writes
`$RUN_EPIC` today**, and pre-flight is what emits run facts into the
coordinator's shell (`BB_*`), so that is where the export has to come from;
and `plugin/skills/bead-bonanza/SKILL.md`, which carries the remembered loop
this hook makes structural.

**Verify.** With ready work outstanding, the coordinator cannot end its turn.
With none, it can.

**Risk — the highest in this plan.** A `Stop` hook that always exits 2 is an
infinite loop the user must interrupt. It needs a bound: a maximum
re-entry count, or a kill-switch file, or both. **Do not ship this without one**,
and test the bound before testing the happy path.

**Size.** M, mostly because of the bound.

---

### 13 · `SubagentStop` hook: verify the worker's queue

**Change.** A hook that rejects a worker's completion claim when its queue is
not empty. Moves acceptance from the coordinator verifying testimony to the
harness refusing the claim.

**Sketch.**

```sh
left=$(bd ready --assignee "$AGENT_ID" --json | jq 'length')
busy=$(bd list --assignee "$AGENT_ID" --status in_progress --json | jq 'length')
[ "$left" -eq 0 ] && [ "$busy" -eq 0 ] || exit 2
```

**Files.** `plugin/hooks/hooks.json` (new) or this repository's own
`.claude/settings.json`, per D1;
`plugin/skills/bead-bonanza/references/prompt-batch.md` and
`prompt-per-bead.md`, which are what export the agent id the hook reads — note
the sketch calls it `$AGENT_ID` while item 4 writes `BEADS_ACTOR`, so the two
names have to be reconciled or the hook reads an unset variable; and
`plugin/skills/bead-bonanza/references/dispatch-policy.md`, the roster's only
home and therefore where that id is minted.

**Verify.** A worker that stops with an open bead is refused; one with an empty
queue is allowed.

**Risk.** Same unbounded-loop hazard as 12, plus a worse failure: a worker
*blocked* on something it cannot resolve would be refused forever. The bound
must distinguish "not done" from "cannot proceed" — a blocked bead is not an
incomplete one.

**Depends on.** 4 (`AGENT_ID` must be reliable), 9 (the queue must be the
source of truth).

**Size.** M.

---

### 14 · Background monitor on the event log

**Change.** Declare a monitor that tails `$BB_EVENT_LOG`, replacing the
arm-a-Monitor-per-wave ceremony with an always-on declarative watch.

```json
[{ "name": "beads", "command": "tail -F $BB_EVENT_LOG",
   "description": "bead state changes" }]
```

**Files.** `plugin/monitors/monitors.json` (new) or this repository's own
config, per D1. Deletes most of
`plugin/skills/bead-bonanza/references/progress-streaming.md` — but **not the
two things that write the log the monitor tails**:
`plugin/skills/bead-bonanza/scripts/pre-flight-check.sh` sets `BB_EVENT_LOG`
(`emitq BB_EVENT_LOG "$EVENT_LOG"`, and emits `BB_EVENT_LOG=''` when the path
is unwritable), and the `on_update` hook that `progress-streaming.md` installs
into `.beads/hooks/` is what appends the lines — a per-project install, so there
is no committed path to point at beyond the reference that specifies it. Delete
either and the monitor tails a path nothing sets or a file nothing writes.

**Verify.** Close a bead; assert the line reaches the session without anything
having armed a watch.

**Risk.** Monitors are a plugin feature; if D1 puts hooks in the repo rather
than the plugin, this one may not have a home at all. Settle D1 first.

**Size.** S once D1 is settled.

---

## What this plan does not cover

- Nothing from 1–14. Item 7 is now planned above.
- **Items 15–19** — open questions. Each could change a decision here; none
  blocks starting.
- **Item 20** — the prototype. It is the point of all of this, and it should
  run *after* 1–6 and 8–11 land, before or alongside 12–14.

## Suggested first bead

**Items 1, 4 and 5 together** — all XS/S, no dependencies between them, and each
removes a silent failure. It is a half-day that makes the next items safe to
build on, and it exercises the plan's own assumptions before the large items
(6, 8, 9) commit anyone to a shape.
