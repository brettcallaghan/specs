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

## Developer / worker tier

Owns a bucket of beads and one worktree. **Never merges.**

| | |
|---|---|
| **Open** | write the dispatch record (`dispatch_id`, model, tier, worktree, time); take the worktree; read the manifest and checks baseline |
| **Work** | `run-bead` per bead, in order, one commit each |
| **Close** | write the terminal state; hand the branch to the merge queue; return one line per bead |

The record is what makes a standing worktree readable without inference:
present + `running` is a live agent, present + `returned` is an orphaned tree,
absent + `running` is a dead agent.

**Status:** the ceremony exists as prompt templates; it is not yet a skill. ⬜

---

## Bead tier

`run-bead` ✅ — claim, read in full, implement in scope, checks, caveat,
`retro:` comments, close-and-commit as one, return one line.

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

