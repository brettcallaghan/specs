# Rolling coordination onto the beads graph

Implementation plan for items **1–6 and 8–14** of the ranked list in
[beads-coordination-design.md](beads-coordination-design.md). Each item below is
sized to become one bead, with the evidence already recorded in the design doc
and not repeated here.

Transient. Delete when the change lands.

---

## Item 7 is out of this plan's scope, but it is decided — and it carries a constraint

**Item 7** — ceremony as epics, target epic as payload, one `bd swarm` on the
target — is not planned here, but it is **wanted**: swarm is being adopted. This
plan is the groundwork under it, so two things follow.

### The constraint item 7 must carry: the swarmed target is FLAT

**Measured, and swarm gives no warning when this is violated:**

| target shape | `bd swarm validate` reports |
|---|---|
| epic → 2 tasks | `Total issues: 2` · Wave 1: both leaves |
| epic → child epic → 2 tasks | `Total issues: 1` · Wave 1: **the child epic** |

A nested target makes swarm count the child epic as a single issue and never see
the real beads beneath it. Progress is wrong, the ready-front analysis treats a
container as work, and `validate` — the command whose whole job is to check epic
structure for swarming — **reports the wrong total as though it were correct.**

So item 7's shape is: ceremony epics under the RUN epic (never swarmed), and a
**flat** target epic containing tasks only. Work that genuinely decomposes into
sub-epics becomes a *second* target under the run epic, swarmed separately.

**This reaches into item 6 now.** The graph plan that item 6 emits is what
creates the target's children, so the flatness rule is a plan-validation rule
from the start: **no node whose `parent_key` is the target may have
`type: "epic"`.** Cheaper to enforce in the emitter than to discover from a
wrong swarm total later. It is also an admission rule for mid-run additions —
scope creep arriving as an epic silently removes everything under it from the
swarm's view.

### And two items are justified differently while 7 is outstanding

**Item 1's severity drops from critical to hygiene.** The silent deadlock in the
design doc arises specifically from epic→epic ceremony gating: ceremony epics
that never close leave the target blocked while `bd ready` reads empty. Without
item 7 there are no ceremony epics, so that deadlock cannot occur yet. `bd epic
close-eligible` is still needed — the run epic must close, and `bd epic status`
is wrong until it does — but it is **not** the highest-severity item any more.
Do it anyway: it is one line, and it is a prerequisite the day item 7 lands.

**Item 10's deferred-review channel needs a different holder.** The design's
mechanism parks the `human`-labelled bead under a late ceremony epic. With no
ceremony epics, use a plain task→task edge instead: the human bead **blocks the
run's report bead**. Same effect — invisible to dispatch until the end, visible
in `bd human list` immediately — and it works with the graph as it stands today.

Everything else in 1–14 is independent of item 7.

---

## Sequencing

```
1 ─┐
2 ─┼─→ 3          (3 enforces what 2 and 4 decide)
4 ─┘
5     independent

6 ─────→ 8, 9     (both consume the graph 6 emits)
              10  (independent; needs the item-7 note above)
              11  independent

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
assignee; (b) the distinct-assignee count is ≤ `concurrency_cap`. Both fail
loudly at plan time; both are invisible at run time.

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
`bd` call. Agent ids per the design doc: `coordinator-<run-epic-id>`,
`dev-<run-id>-<batch-id>`, `merge-<run-id>`.

**Files.** `plugin/skills/bead-bonanza/references/prompt-per-bead.md`,
`prompt-batch.md`, and the placeholder table in
`plugin/skills/sprint-wave/SKILL.md`.

**Verify.** In a scratch database: assign a bead to `dev-x`, claim it with
`BEADS_ACTOR=dev-x`, assert the assignee is unchanged. Repeat without the
variable and assert it changed — the negative control is the point, because the
failure is silent.

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
`bd create --graph`. Schema (measured, design §4a):

```json
{"nodes":[{"key":"…","title":"…","type":"epic|task","parent_key":"…",
           "assignee":"…","priority":2,"description":"…"}],
 "edges":[{"from_key":"…","to_key":"…","type":"blocks"}]}
```

Without item 7 the emitted shape is flat: run epic, work beads as its children,
edges between them. The ceremony epics arrive with item 7 and the schema does
not change.

**Files.** `plugin/skills/sprint-plan/SKILL.md` (the output contract),
`plugin/skills/sprint-plan/references/run-record.md` (where the plan file lives
and how it is recorded), `plugin/skills/bead-bonanza/SKILL.md` (the load step).

**Verify.** Emit a plan for a real epic, load it into a scratch database, and
assert: node count matches, every edge present, every bead assigned, and
`bd ready` returns exactly the expected front. **Diff the plan before loading** —
being reviewable pre-load is most of the value.

**Open.** Whether `--graph` accepts gates or per-node metadata is unverified;
probe before relying on either.

**Risk.** A malformed plan fails at load having created part of the graph.
`bd create --graph` reports unknown fields as warnings and **silently drops
them** — so a typo in a key name is a silently missing assignee, not an error.
Validate the plan against the known field set before loading.

**Validate flatness too, now rather than later** (see the item-7 constraint
above): no node parented to the target epic may be `type: "epic"`. This costs
one check in the emitter and prevents a class of failure that only shows up as
a quietly wrong swarm total once item 7 lands.

**Size.** L. This is the item with real design left in it.

---

### 8 · Collisions as edges; delete the wave table

**Change.** The C4 co-edit computation stops producing a wave table and starts
producing dependency edges in the graph plan. Everything describing waves goes.

**Files.** `plugin/skills/sprint-plan/SKILL.md`,
`plugin/skills/sprint-plan/references/run-record.md` (wave table section),
`plugin/skills/bead-bonanza/SKILL.md`,
`plugin/skills/bead-bonanza/references/dispatch-policy.md`,
`plugin/skills/sprint-recovery/SKILL.md`, and `sprint-wave` in its entirety —
this item is what deletes that skill.

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

**Files.** `plugin/skills/sprint-batch/SKILL.md`,
`plugin/skills/sprint-wave/SKILL.md`,
`plugin/skills/bead-bonanza/references/recovery.md`, `scripts.md`,
`prompt-per-bead.md`, `plugin/skills/sprint-recovery/SKILL.md`, and the scripts
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
`plugin/skills/bead-bonanza/SKILL.md`.

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
`plugin/skills/bead-bonanza/references/progress-streaming.md`.

**Verify.** Close a bead; assert the line reaches the session without anything
having armed a watch.

**Risk.** Monitors are a plugin feature; if D1 puts hooks in the repo rather
than the plugin, this one may not have a home at all. Settle D1 first.

**Size.** S once D1 is settled.

---

## What this plan does not cover

- **Item 7** — the run-epic shape and swarm adoption. Decided but not planned
  here; its flatness constraint is recorded at the top and item 6 enforces it
  from the start.
- **Items 15–19** — open questions. Each could change a decision here; none
  blocks starting.
- **Item 20** — the prototype. It is the point of all of this, and it should
  run *after* 1–6 and 8–11 land, before or alongside 12–14.

## Suggested first bead

**Items 1, 4 and 5 together** — all XS/S, no dependencies between them, and each
removes a silent failure. It is a half-day that makes the next items safe to
build on, and it exercises the plan's own assumptions before the large items
(6, 8, 9) commit anyone to a shape.
