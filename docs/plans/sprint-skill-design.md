# The sprint skill, reframed as a team

Evolves the role framing: coordinator → developer → bead, with a rule for what
belongs in each bucket and one contract every subprocess honours.

Working document. Status markers: ✅ exists · 🔨 exists but not extracted ·
⬜ does not exist yet.

---

## Objective

Implement many beads autonomously with as little human interaction as possible.
Hand the skill a large epic and have it run until the epic is complete, at high
quality.

## Hypotheses

**Humans orchestrate more work.** Agents run epics autonomously across sessions
with far less interaction. Not a director managing many teams — but it lets a
human operate at director level, and is a step toward an agentic director.

**Groomed beads are what make long autonomous sessions possible.** Every
subagent is a fresh agent with zero context, so grooming manufactures
*context-independence and correct ordering* — the two things the coordinator
actually consumes. A bead that reads fine to someone who was in the planning
session but is unworkable cold is the central failure mode, and it is invisible
until it detonates three waves deep.

**Beads is the harness.** Task management, memory and orchestration primitives
already exist and survive compaction, so this can be a skill rather than a
harness of our own.

**The coordinator's own skill must be the SMALLEST document in the system.**
It never pulls in another skill's context — it invokes, reads the result, and
routes. Today `bead-bonanza/SKILL.md` is the largest file here; in the target
state it should be the shortest thing a coordinator reads.

---

## Decided

1. **The review loop lives in the coordinator.** It owns the loop and the
   counter; a subagent owns finding things. The coordinator never holds the
   review *method*.
2. **Human interaction is front-loaded, and there are exactly four touchpoints**
   — see below.
3. **Review findings go through `groom-beads` like any other bead.** Review
   files; it does not groom. The loop re-enters at step 3 unconditionally.
4. **Two loops maximum — the process runs at most three times.**

---

---

## The human, and the rule that protects autonomy

Four touchpoints, and no others:

| when | what |
|---|---|
| **start** | the human invokes the skill with a scope |
| **immediately after** | the **authorization gate** — one interaction, all decisions at once |
| **after the loop** | PR review |
| **end** | reads the final report |

Everything between the gate and the PR is autonomous.

### The rule: no new questions after the gate

A run that stops mid-flight to ask something is not autonomous, however
reasonable the question. So:

> **After the authorization gate, an unauthorized decision is never a stop.**
> The coordinator takes the **conservative branch**, records the choice on the
> run epic as an `amend:` comment, and names it in the final report.

That rule is what makes the gate worth designing carefully: every question the
run could need must be *askable up front*, because it will not get a second
chance.

### The authorization set

Asked once, as one interaction, with a stated default for each so a human can
accept the lot in a sentence.

| decision | conservative default (taken if unauthorized) |
|---|---|
| **Push mode** | defer — the coordinator pushes per wave |
| **Pull request** | open as draft at start, mark ready at the end |
| **Moving beads** that already belong to another epic | **do not move** — exclude them from the run and report |
| **Scope growth** from review findings | file the beads, do **not** admit them into this run |
| **Tier escalation** on repeated failure | escalate one tier, once, and report it |
| **Review loop budget** | 2 loops (three passes total) |
| **Merge to the default branch** | never — that is the PR review's, not the run's |

The conservative default of every row is "do the smaller thing and say so",
which is the only default that is safe to apply without asking.

---

## The shape of a run

```
1  pre-flight ................................. script
2  resolve scope ............................. script + coordinator
   └─ AUTHORIZATION GATE ..................... HUMAN, once
┌─ 3  groom .................................. subagent
│  4  plan ................................... subagent
│  5  sprint ................................. coordinator + subagents
│     ├ 5a dispatch a wave  ─┐ repeat until
│     ├ 5b merge queue       │ the ready front
│     └ 5c supervise        ─┘ is dry
│  6  review ................................. subagents; coordinator loops
└──── findings admitted AND passes < 3 ? ──── yes ──┘
7  PR review ................................. HUMAN
8  retro ..................................... subagents + coordinator
9  report .................................... coordinator → HUMAN
10 post-flight ............................... script
```

**Two nested loops, and both are bounded.** The wave loop ends when the ready
front is dry; the review loop ends at 3 passes or no admitted findings,
whichever comes first.

**At loop exhaustion the run does not fail.** Remaining findings stay filed as
beads, are named in the report, and the PR proceeds. A third pass that still
finds things is information for the human, not a reason to keep going.

**The pass counter is durable.** It lives in the manifest on the run epic, not
in the coordinator's head — recovery has to know which pass it is on, and a
resumed run that restarts the counter would loop forever.

---

## The bucketing test

Two questions decide where a step goes, in this order.

### 1. Does the coordinator need to retain HOW it was done?

- **No** → it is a subprocess. The coordinator keeps the verdict and a handle,
  never the working detail.
- **Yes** → it stays in the coordinator, because extracting it would mean
  reading it back in anyway.

This is the whole reason the coordinator shrinks: *detail goes down, verdicts
come up.* A subprocess that returns its reasoning has failed to encapsulate.

### 2. What kind of thing produces the result?

| the result is… | executed by | why |
|---|---|---|
| computable from state | **script** | deterministic, re-runnable, costs no context, and can be re-run by anyone doubting it |
| judgment over prose or code | **subagent** | needs a fresh context window and returns a structured report |
| a decision whose consequences the human owns | **human** | scope changes, moves that restructure their tracker, merging to main |
| routing between the above | **coordinator** | this is the only thing it does |

**A script beats a subagent whenever both could work.** A script's output is
the same every time and cheap to re-run; a subagent's is neither. The
pre-flight step is the model: run it, read stdout, done — the coordinator needs
no understanding of what it checked.

---

## The subprocess contract

Every subprocess the coordinator invokes declares exactly five things. This is
what makes the coordinator's skill a table of calls rather than prose.

| field | meaning |
|---|---|
| **Input** | what the coordinator must provide. If it needs more than a handle and a couple of facts, the boundary is wrong |
| **Returns** | what lands in the coordinator's context — and nothing else |
| **Coordinator does** | the routing decision this return enables |
| **Durable side effect** | where it wrote, so recovery finds it without the coordinator |
| **On failure** | what a failure means, and whether the run stops |

**The coordinator holds handles, not copies.** The run epic id, not the
manifest. The wave table for the current wave, not every wave. Anything it can
re-derive from a handle, it re-derives.

---

## Coordinator — the steps

| # | step | executed by | status |
|---|---|---|---|
| 1 | Pre-flight | script | ✅ |
| 2 | Resolve scope | script + coordinator + **human** | ✅ |
| 3 | Groom | subagent | ✅ |
| 4 | Plan sprint | subagent | ✅ |
| 5a | Dispatch a wave | coordinator + subagents | 🔨 |
| 5b | Merge queue | subagent, serialized | ⬜ |
| 5c | Supervise | script + trigger | 🔨 |
| 6 | Review code | subagents review, coordinator loops | ⬜ |
| 7 | PR review | **human** | ⬜ |
| 8 | Retro | subagents + coordinator | 🔨 |
| 9 | Report | coordinator | 🔨 |
| 10 | Post-flight | script | ⬜ |

### 1 · Pre-flight

- **Input** — nothing.
- **Returns** — `BB_*` facts on stdout, eval'd. Human-readable report on stderr.
- **Coordinator does** — eval, then proceed or stop. It does not interpret;
  the script names each blocker.
- **Durable side effect** — none on this pass.
- **On failure** — exit 1 names a blocking condition. **The run does not
  start.** Do not work around it.

*The model for every other step: run it, read it, done.*

### 2 · Resolve scope

Turns whatever the user gave — an epic, several epics, bead ids, a label, a
priority — into **exactly one run epic**.

- **Input** — the user's scope; standing authorization.
- **Returns** — the run epic id; the admission record (admitted / **excluded**
  and why / **moved** and from where).
- **Coordinator does** — carries the epic id as the run id for everything
  after. Reports exclusions in the final report.
- **Durable side effect** — the epic exists; the manifest is `--metadata` on
  it; pre-flight's second pass writes the checks baseline there.
- **On failure** — a declined move is **not** a failure: those beads are
  excluded and the run proceeds narrower.

**This step carries the one human interaction of the run** — the authorization
gate, below. It is deliberately early: the human invokes, authorizes, and then
does not hear from the run again until the PR.

A move is confirmed by naming each bead and the epic it leaves, because moving
a bead out of an epic makes that epic falsely report complete. Unauthorized,
the conservative branch excludes those beads and the run proceeds narrower.

### 3 · Groom

- **Input** — the run epic.
- **Returns** — per bead: ready, or NOT-READY with the gap a human must close.
- **Coordinator does** — dispatches only ready beads. Bounced beads are raised
  into `bd human` and named in the report.
- **Durable side effect** — beads edited in place; `decision-required` +
  `human` labels on bounces.
- **On failure** — a bounced bead never blocks the run; it narrows it.

### 4 · Plan sprint

- **Input** — the run epic; the current ready front.
- **Returns** — the wave table: which simultaneously-ready beads can share a
  wave, which must be batched behind one worker.
- **Coordinator does** — dispatches from it. Does **not** store it.
- **Durable side effect** — none by design. The table is derived; a stored copy
  only drifts.
- **On failure** — no partition means dispatch one worker per bead.

**Order is not planned — it is queried.** `bd ready --parent <epic>
--exclude-type=epic` at every wave boundary is the order. Planning decides
*packing* only.

### 5a · Dispatch a wave

- **Input** — the wave's buckets; the manifest handle; `{SCRIPTS}`.
- **Returns** — one line per bead: completed / blocked / skipped, with a sha.
- **Coordinator does** — verifies every return against state (never testimony),
  then hands each branch to the merge queue.
- **Durable side effect** — worktrees, branches, commits, closed beads,
  dispatch records in bead metadata.
- **On failure** — 3 consecutive blocked/failed beads trips the circuit breaker
  and stops the wave.

### 5b · Merge queue ⬜

- **Input** — a branch, and the run epic.
- **Returns** — merged, or a conflict report naming what collided.
- **Coordinator does** — continues, or stops the wave on a real conflict.
- **Durable side effect** — the sprint branch advances; the merge slot records
  the holder.
- **On failure** — a conflict is a triage failure: the beads were not isolated.

**One merge worker at a time, strictly.** Concurrency here is the failure the
merge slot exists to prevent. The win is that conflict resolution stops
consuming the coordinator's context.

### 5c · Supervise

- **Input** — the run epic.
- **Returns** — health, live dispatches, standing worktrees, DoD.
- **Coordinator does** — sets health at each wave boundary; acts on a stall.
- **Durable side effect** — a `health:<value>` label and one event bead.
- **On failure** — an unset health is itself the finding: a watchdog cannot
  tell stalled from slow.

### 6 · Review code ⬜

**The largest genuinely new step.** Reviews everything the sprint produced,
files what it finds, and **loops back to step 3** when new work is admitted.

- **Input** — the sprint branch; the run epic; the base sha.
- **Returns** — findings, each already filed as a bead, classified in-scope vs
  out-of-scope.
- **Coordinator does** — admits in-scope findings **only where the gate
  allowed scope growth**, and re-enters at **step 3** so new beads are groomed
  like any other. Surfaces out-of-scope ones. **At most 2 loops — three passes
  total.**
- **Durable side effect** — new beads linked to the run epic; the **pass
  counter** in the manifest, so a recovered run does not restart it.
- **On failure** — findings that keep reshaping rather than converging means
  stop and surface, not another loop. Exhausting the budget is **not** a
  failure: remaining findings stay filed, are named in the report, and the PR
  proceeds.

Review **files**; it does not groom. That is why the loop re-enters at 3 rather
than 4 — a finding's bead meets the Definition of Ready the same way every
other bead does, through `groom-beads`, or it bounces the same way too.

This is the coordinator's own loop, and the reason the loop stays here while
the review method goes to a subagent: the counter, the budget and the decision
to re-enter are routing, and routing is the only thing the coordinator does.

### 7 · PR review — human ⬜

- **Input** — a PR against the sprint branch.
- **Returns** — approval, or comments.
- **Coordinator does** — re-enters at step 3 for anything actionable.
- **On failure** — n/a; this gate is the human's.

### 8 · Retro

- **Input** — the run epic.
- **Returns** — harvested `retro:` comments, coalesced; the team's retros; the
  coordinator's own.
- **Coordinator does** — contributes its own, then hands composition to a
  subagent which decides with a human which recommendations become beads.
- **Durable side effect** — new beads; RFCs where the finding is a spec's.
- **On failure** — an empty harvest is a finding about the run, not a pass.

### 9 · Report

- **Input** — everything above, by handle.
- **Returns** — the human-facing sprint report.
- **Coordinator does** — this is its own output, and the one thing it does not
  delegate.

### 10 · Post-flight ⬜

- **Input** — the run epic.
- **Returns** — what it removed, and what it deliberately left.
- **Coordinator does** — reports it.
- **Durable side effect** — worktrees pruned, merge slot released, swarm
  molecule closed, health set to its terminal value.
- **On failure** — never destructive on ambiguity: a standing worktree with a
  `running` dispatch record is left alone and reported.

---

## What the coordinator retains

The entire context budget, and nothing else belongs in it:

1. the run epic id
2. the current wave's table
3. the standing authorization
4. a running tally — closed, blocked, excluded, escalated
5. the review loop's iteration count

Everything else is re-derivable from (1). That is the test for whether the
split worked.

---

## Recovery — re-run with the same epic id

**`recover-sprint`** ✅ owns this. It stays a skill, and what follows is its
specification: it reconciles the run's residue and hands back to step 1. It is
a **precondition to re-entry**, not a second way to run a sprint.

**Recovery is re-entry.** Invoke the skill again with the same epic id. Every step is keyed on the epic and naturally skips what is
already done:

| step | why re-entry is a no-op or a narrower repeat |
|---|---|
| 1 pre-flight | re-probes; it asserts nothing about history |
| 2 resolve scope | the epic exists, so admission has nothing to resolve |
| 3 groom | only ungroomed beads are groomed |
| 4 plan | re-derived from the **current** ready front — this is why the wave table is never stored |
| 5 sprint | closed beads are not in the front, so they are not dispatched |
| 8–10 | operate on final state, which is what they wanted anyway |

**Where the simple story breaks — and it is exactly five things.** All of them
are residue from the interrupted run, and all of them are what
`bb-status.sh` already reports:

1. **A standing worktree with uncommitted work.** The bead reads open, so
   re-entry would re-dispatch it and race the partial work. Reconcile first:
   present + `running` is a live agent, present + `returned` is an orphaned
   tree, absent + `running` is a dead agent.
2. **A held merge slot.** A slot held by a dead agent blocks the new run.
   `bb-status.sh slot <epic>` says whether release is defensible.
3. **Unpushed commits.** Every other signal reads "finished" — no worktrees,
   clean tree, beads closed. Only `bb-status.sh wave` separates a landed wave
   from an unlanded one.
4. **The review pass counter.** The only genuinely non-idempotent step. It
   lives in the manifest for exactly this reason; a re-entry that reset it
   would loop forever.
5. **A bead claimed by an assignee that no longer exists.** `--claim` is
   idempotent for the same actor and blocking for a different one.

So recovery is: **reconcile the residue, then re-enter at step 1.** The
reconciliation is three script calls, which is why `recover-sprint` collapses
to running them and acting — it is a precondition to re-entry rather than a
second way to run the sprint.

---

## What each step leverages from beads

The skill is a skill rather than a harness because bd already provides this.
Making it explicit also shows where we would otherwise be hand-rolling.

| step | bd features |
|---|---|
| 1 pre-flight | `bd stats` / `bd ready` reachability · `bd dolt push` probe against `refs/dolt/*` |
| 2 resolve scope | `bd create --type=epic` · `bd update --parent` (admission) · `bd ready --parent --exclude-type=epic` (the queue) · `bd update --metadata` (the manifest) |
| 3 groom | `bd update --description/--design/--acceptance` · `bd dep add` (sequencing as real edges) · `bd lint` · `bd human` + `decision-required` label (bounces) |
| 4 plan | `bd ready --parent` (order) · `bd swarm validate` (wave **preview** only) · `metadata.tier` / `c4_changes` / `c4_impacts` (packing) |
| 5a dispatch | `bd update --claim` (atomic, sets status **and** assignee) · `--set-metadata` (the dispatch record) · re-queried `bd ready --parent` |
| 5b merge queue | `bd merge-slot create/check/acquire/release` |
| 5c supervise | `bd set-state health=<v>` (event bead + replacing label, atomically) · `bd swarm status` (computed, not stored) · `.beads/hooks/on_update` (the event stream) |
| 6 review | `bd create --discovered-from` (findings stay attached to their origin) · labels for in/out of scope |
| 7 PR review | none — this is the forge's |
| 8 retro | `bd export \| jq` over `retro:` comments · `bd create` for adopted actions |
| 9 report | `bd epic status` (flat) or a walked count (nested) · the manifest's admission record |
| 10 post-flight | `bd merge-slot release` · close the swarm molecule · `bd set-state health=<terminal>` · `bd dolt push` |
| **throughout** | `bd dolt push`/`pull` — the tracker is the durable store, so nothing the run learns lives only in a session |

**Two things we deliberately do not use.** `bd remember` memories are excluded
from `bd export` and are not synced, so durable context goes in an issue
instead. And **event beads are not used as a dispatch record** — measured, 100
of them made `bd dolt push` 8.5× slower and deleting them did not restore the
baseline, so the record is metadata.

---

## Developer / worker tier

One subagent, one bucket of beads, one worktree. **`run-batch`** ⬜ — the skill
that does not exist yet; today this is prompt templates.

The developer is the tier with no artifacts of its own until the dispatch
record gave it one. Its ceremony mirrors the coordinator's — open, work, close
— and its three hard rules are all *negative*, because a worker that reaches
past its boundary is the failure mode:

> **Never merge. Never work outside your worktree. Never touch a bead you were
> not assigned.**

### D1 · Open — accept the assignment

- **Input** — the bucket's bead ids in order; `{DISPATCH_ID}`; `{WORKTREE}`;
  the run epic; `{SCRIPTS}`; push mode.
- **Returns** — nothing yet. This step is a gate, not a report.
- **Does** — writes the dispatch record; verifies the worktree is its own and
  clean; verifies the **base ref** (branched from sprint HEAD, not the default
  branch); reads the manifest and the checks baseline off the epic.
- **Durable side effect** — `dispatch_id`, `dispatch_state=running`,
  `dispatch_model`, `dispatch_tier`, `worktree`, `dispatched_at` on every bead
  in the bucket.
- **On failure** — a wrong base ref is a **stop**: every commit the sprint
  branch carries would be silently dropped from this worktree. Report and do
  not start.
- **bd** — `bd update --set-metadata` · `bd show <epic> --json` for the manifest.

**Why the record comes first.** It is written *before* any work, so a worker
that dies in its first minute is still distinguishable from one that was never
dispatched.

### D2 · Work — `run-bead`, per bead, in order

- **Input** — one bead id, priority, title; the checks baseline; the trailers
  source.
- **Returns** — per bead: `completed` / `blocked` / `skipped`, with a sha.
- **Does** — invokes **`run-bead`**. The worker adds nothing to the per-bead
  ceremony; it sequences it.
- **Durable side effect** — one commit per bead carrying both the code and the
  closure; `retro:` comments; any beads filed for out-of-scope work found.
- **On failure** — a blocked bead does **not** stop the bucket. Leave its
  partial work exactly where it is — no stash, no WIP commit — and continue to
  the next bead. The standing worktree plus the record is the signal.
- **bd** — `bd update --claim` · `bd comment` (`retro:`, `caveat:`) ·
  `bd create --discovered-from` · `bd update --status closed` inside the close
  script.

**Order is the bucket's order.** The coordinator packed it; a worker
re-sequencing its own bucket defeats the packing.

### D3 · Close — hand off

- **Input** — what D2 returned.
- **Returns** — one line per bead, then one batch line:
  `batch-end: <n> completed, <n> blocked, <n> skipped @ <branch>`.
- **Does** — writes the terminal state on every assigned bead; confirms every
  bead is accounted for; **hands the branch to the merge queue** — it does not
  merge, and it does not remove its own worktree.
- **Durable side effect** — `dispatch_state=returned` and
  `terminal_state=completed|blocked|skipped` per bead.
- **On failure** — a bead with no terminal state is the one thing that must not
  happen: it is indistinguishable from a dead worker. Write it even when
  blocked — **especially** then.
- **bd** — `bd update --set-metadata`.

**The worker does not remove its worktree.** Removal is the coordinator's final
act of *accepting* the bead, after the merge. A worker that cleaned up after
itself would erase the in-flight signal before anyone read it.

### What makes the tier legible

| worktree | `dispatch_state` | reading |
|---|---|---|
| present | `running` | live worker |
| present | `returned` | finished; orphaned tree, coordinator removes it |
| absent | `running` | the worker died |
| absent | `returned` | done and cleaned up |

That table is why the tier needed a record at all: before it, a standing
worktree took three paragraphs of prose to interpret, and a batch holding one
worktree for six beads broke the acceptance signal outright.

## Bead tier

`run-bead` ✅ — claim, read in full, implement in scope, checks, caveat,
`retro:` comments, close-and-commit as one, return one line.

---
