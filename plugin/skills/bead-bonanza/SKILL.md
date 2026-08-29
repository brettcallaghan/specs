---
name: bead-bonanza
description: Process many beads autonomously by orchestrating subagents. Use
  when the user asks to work on multiple beads at once — e.g. "work on all open
  beads with label X", "address all P1 beads", "clear the bead backlog". Triages
  the queue, gates every bead for readiness (grooming under-specified beads via a
  subagent, or bouncing what can't be made ready), surfaces both explicit and
  implicit decisions up front, dispatches the ready front continuously, runs
  every dispatch in its own git worktree, merges each branch as it returns, and
  lands each bead in a single commit.
version: "2.1"
---

# bead-bonanza

**You are the coordinator, and routing is the only thing you do.** Each step
below is a subprocess: you invoke it, read what it returns, and decide what
happens next. You never hold a subprocess's method.

**What you retain — the whole context budget:** the run epic id, the review
pass counter, the authorization set, where the dispatch loop is, and what the
last subprocess returned. Not a subprocess's reasoning, not the evidence behind
a verdict.

**Act on the verdict, not the evidence.** A subprocess whose output you have to
reason about has pushed work back up the tier it was meant to absorb.

Read `CLAUDE.md` and `AGENTS.md` once at the start, before delegating.

**This skill is language-agnostic** — it hard-codes no build, test, lint or
format command. Validation comes from the checks the project declares in
`sf.toml`, run by `scripts/bb-checks.py`: exit 0 means every check passed,
anything else means something failed, and the output says what.

## The shape of a run — it IS the graph

**Every phase is an EPIC under the run epic, chained `epic→epic` in order.** The
run's sequence is not a list you walk; it is edges the tracker enforces.

```
run epic .................. its id is the run id
├ pre-flight .............. pre-flight-check.sh
├ scope ................... sprint-scope
│   └ AUTHORIZATION GATE .. a TASK child, gated; HUMAN, once
├ groom & plan ─┐ ......... sprint-plan → groom-beads
├ TARGET ───────┤ FLAT: tasks only
│  ├ dispatch ┐ │ repeat    run-batch, one subagent per dispatch
│  ├ merge    │ │ until the coordinator, as each dispatch returns
│  └ supervise┘ │ front dry swarm status
├ review ───────┘ ......... sprint-review → declared reviewer
│   findings admitted AND passes < 3 ? — yes, re-enter at groom & plan
├ retro ................... sprint-retro — and it writes the run's retro
└ post-flight ............. sprint-cleanup.sh
```

**Epics because the type rules leave no other form.** Measured: `bd dep add`
refuses task→epic (`tasks can only block other tasks, not epics`) and
epic→task alike, so whatever blocks the target must be an epic. One `epic→epic`
edge then blocks **every bead beneath the target**, and a bead added there
mid-run inherits the block with nothing wired.

**The TARGET is flat — tasks only.** Work that genuinely decomposes into
sub-epics becomes a SECOND target under the run epic, never a sub-epic under
the first: `references/run-record.md`.

**Two nested loops, both bounded**: the dispatch loop ends when the ready front
is dry, the review loop at 3 passes or no admitted findings. Neither exhausting
is a failure — what is left is filed and named. **The retro writes the run's
one human-facing document**, so there is no separate report phase to keep in
agreement with it.

**A human's pull-request review is not a step here.** Whether one happens is
the project's business — the run ends at post-flight either way, and findings
from one do not consume the review budget.

## Run checklist

Copy this into your response at run start and check it off as you go — it lives
in your response, not a task tool.

```
run:
- [ ] 0. pre-flight run, BB_* facts eval'd
- [ ] 1. `sprint-scope`: run epic, admission record, authorization gate
- [ ] 1b. pre-flight re-run with the epic: checks baseline on the manifest
- [ ] 2. `sprint-plan`: readiness verdict; collision edges validated & loaded
- [ ] 2b. Sprint branch created (or reused), `branch` + `base_sha` in manifest
- [ ] 3. Dispatch the ready front; merge each branch as it returns, then re-check
- [ ] 3b. `sprint-review`; loop back to groom & plan if findings were admitted
        and the pass budget is not spent
- [ ] 4. `sprint-retro` — it writes the run's retro
- [ ] 5. Post-flight cleanup, then cancel the durable trigger
```

**Most of it is checkable rather than assertable.** `swarm status --run <epic>`
measures the mechanical half — the standing worktrees, the last commit and its
age, the ready front's depth, each dispatch's drain state, and completion
against the admission record. A measurement it could not take is reported as
unknown and exits non-zero rather than reading as clear, and the rest of the
checklist it does not see at all. Run it at every re-query and at the end.


## Opening a run — pre-flight, then `sprint-scope`

**The scripts are in THIS skill's `scripts/`, not the repository's** — your cwd
is the repository root, where `scripts/` is something else. The first call needs
the skill path; every call after uses `$SCRIPTS`, which pre-flight emits, and
every dispatch is handed it because a subagent cannot infer it.

```bash
SKILL=<this skill's directory>                     # e.g. .../plugin/skills/bead-bonanza
eval "$("$SKILL"/scripts/pre-flight-check.sh)"     # 1. environment; sets SCRIPTS
#   2. invoke `sprint-scope` — it returns BB_EPIC and BB_TARGET_EPIC
eval "$("$SCRIPTS"/pre-flight-check.sh "$BB_EPIC")" # 3. + checks baseline
```

**1 and 3 are a script**, not a skill: run it, eval the stdout, read the stderr
report once. It names each blocker AND its remedy, so there is nothing to
weigh. `references/pre-flight.md` only when it blocks, and **a blocking
condition means the run does not start** — never work around it.

**Given a run epic that already exists, you are RESUMING, not opening.** Skip
step 2 — scope has nothing left to resolve — and invoke `sprint-recovery`
instead, which reconstructs the run from its branch, its epic and the surviving
worktrees and hands the state back. Then re-enter at groom & plan: it grooms
only what is not already Ready and re-derives the partition from the current
front, so a resumed run is a narrower repeat rather than a duplicate.
**Never create a second sprint branch on a resume** — the manifest's `branch`
is the one the original run wrote.

**Run `swarm manifest validate --run <epic>` before you continue a run you did
not hold.** It checks the record against every rule at the stage it has
reached, so a run left half-written by an interrupted session is a status you
route on rather than a key you discover missing three dispatches later. A clear
status means the record is coherent; an attention status names the keys and the
rule each breaks, and those are yours to repair before dispatching.

**2 is `sprint-scope`.** Read three things off its return and route on them:
`BB_EPIC`, whose id is the run id and whose descendants are the queue;
`BB_TARGET_EPIC`, the flat target; and the admission record, because **a run that
narrowed its own scope says so in its record**.

## Groom & plan — dispatch ONE subagent running `sprint-plan`

**One subagent, two jobs, one context.** It invokes `groom-beads` for the
grooming and then plans the dispatch itself. You dispatch it once and never
dispatch grooming separately: `c4_changes` / `c4_impacts` are optional
metadata, so on a repo with no C4 docs the collision signal is each bead's
*description* — which grooming has just read. A second agent would read them
all again, and a metadata-only planner would under-detect collisions on
exactly the repos that have none.

It returns two things:

1. **The readiness verdict** per bead. Dispatch only the ready ones; the skill
   bounces the rest into `bd human` itself.
2. **A collision-edge file.** Diff it, re-run the validators the plan names,
   load it with `bd dep add --file <path>`, and record the path as
   `collision_edges` on the manifest. **It creates no beads** — the work beads
   already exist, so edges are all a plan contributes.

**Then assert the front rather than trusting the file**: an edge written the
wrong way round inverts a dependency at exit 0.

**The partition lives in the graph, not in a schedule you keep** — which is why
the ready front is safe to dispatch whole and there is nothing to re-derive.

**Two things stay yours**, and neither is ever delegated: the run manifest and
the run's amendments. Both: `references/run-record.md`.

## The sprint branch — you create it, once, before the first worktree

**Immediately before the first dispatch**, not earlier: a run that bounced
every bead has nothing to branch for, and a branch cut before grooming records
a `base_sha` predating the grooming commits.

**Read the manifest before you create anything** — `swarm manifest get --run
<epic> --key branch`. A run whose manifest already carries `branch` reuses it
and creates nothing, which is what makes a resumed run land where the original
did; **a name already taken by an unrelated branch stops the run here**, before
any worktree exists. The exact sequence, the one `swarm manifest set` that
records `branch` and `base_sha` together, and why `base_sha` is recorded at
creation rather than derived: `references/worktree-mechanics.md`.
## Every dispatch runs in its own worktree

**This is universal, not opt-in**, and it is **per dispatch, not per bead**: a
batch is one agent working serially, so a second tree isolates it from nobody
and a branch per bead buys nothing either — **one commit per bead is what
separates the beads.** Each dispatch gets one worktree and one branch, both
named for its dispatch id, cut from the sprint branch. Nothing is implemented
in the shared checkout except by you.

**Presence is the in-flight signal.** A standing worktree means its dispatch is
not yet accepted — removal is the final act of accepting, after its branch
merges. `git worktree list` says what is in flight; each tree's checked-out
branch names the agent holding it.

**Hard gate — `BB_WORKTREE_BASEREF` must read `head` before every worktree you
create.** It is global user state no repository owns, so it can change under
you, and a worktree cut from the default branch cannot see what the sprint
branch already carries.

**YOU merge, as each dispatch returns.** There is no merge queue, no merge bead
and no lock: you are one agent, so one-merge-at-a-time holds by construction
and a mechanism to enforce it buys nothing. One merge and one removal per
dispatch, then the acceptance check per bead.

**Removal goes behind an ancestry check, not behind an `&&`** — a chained
command whose merge exited 128 removed a worktree anyway. **A `--ff-only`
refusal means one thing where dispatches were serialized and another where they
went out concurrently, and reading it wrong stops a healthy run.** Both rules,
the base-ref remedy, the create/merge/remove lifecycle, and who resolves a
rebase conflict: `references/worktree-mechanics.md`.
## Dispatch the ready front — continuously

**`bd ready --parent <epic> --exclude-type=epic --exclude-label human` IS the
dispatch list.** The collision edges make it safe to dispatch whole — nothing
to freeze, wait on, or intersect. Dispatch what is in it and not already held, then re-query as
each batch returns, so a bead that became ready behind a slow one goes out at
once rather than waiting for its neighbours.

Each dispatch is one subagent running `run-batch` in one worktree, on one
branch named for the dispatch. **The procedure — filling in the prompt,
verifying skip claims, the per-bead acceptance check, the push, the circuit
breaker and reading a live agent — is `references/dispatching.md`.** Read it
before the first dispatch.

**Then you merge**, one branch at a time, removing each worktree as you accept
its dispatch. Subagents merge nothing and push nothing.

**Then re-run the acceptance check without `--pre-merge`** —
`"$SCRIPTS"/bb-status.sh bead <id> --worktree <path>`. The pre-merge run
deferred signals 3 and 4, which fail by construction before a merge, so this
run is the only thing that checks the merge and the removal happened. Exit 2
stops the run.

**Re-query before every dispatch, and run `bd epic close-eligible` first.** It
takes NO epic argument — a passed id is silently ignored, so the bare command
closes every eligible epic in the database, not just this run's. Scope it by
hand: `references/supervision.md`. **Omitting it is a silent deadlock, not a
slowdown**: an epic does not close on its own when its children do (measured),
so a finished ceremony phase stays open, the beads it gates never enter the
front, and an empty front reads as the run being done when it has stalled.
**That query is the order** — and a bead added mid-run is in scope by
construction, so before dispatching it diff the front against
`manifest.admitted` (`swarm status --run <epic>` names the difference) and
resolve any creep, admitted as an amendment or deferred out.

### Watching it, and knowing when it is done

**Arm a `Monitor`** over the dispatches in flight — `persistent`, an explicit
`timeout_ms`, stopped with `TaskStop` when the front dries — and, before the
first dispatch, the durable out-of-session trigger, cancelled at post-flight.

**A run is done when every bead the manifest's `admitted` list names is closed
or accounted for as blocked-with-a-reason** — `swarm status --run <epic>`
measures it and you route on the exit status, per [[frd-swarm-core]]'s command
contract. Beads excluded at admission are reported alongside, because a run
that finished its admitted set while three sat out is done over a *narrowed*
scope.

**Never write a health label of your own.** A signal whose only writer is the
thing it monitors reports "no failure" and "no monitor" identically. And the
command reports measurements, not a diagnosis: **do not call the run stalled**
on the strength of an elapsed time — say what was measured and the threshold
in force.

Where the run is, which of three deaths you are looking at, and what nothing
can measure from outside the harness: `references/supervision.md`.

## Review the run's own diff — invoke `sprint-review`

When the ready front is dry, **before the retro**. It files what it finds and
classifies each finding in or out of scope; it does not decide whether to loop.

**The loop is yours, and only the loop** — findings admitted **and** passes
under budget means re-enter, anything else means done.

**Re-enter at groom & plan, never at dispatch.** Review *files*; it does not
groom, so an admitted finding earns its readiness like any other bead.

**The budget is 2 loops — three passes** unless the gate said otherwise, and it
lives in the manifest as `review_passes` because a resumed run that restarted
the counter would never exhaust it. Why exhausting it is not a failure:
`references/supervision.md`.

## Retro — invoke `sprint-retro`

After the queue drains, or the circuit breaker trips, and **on every run** —
including one that closed nothing. A run that bounced its whole queue produced
no code and a great deal of evidence about grooming, and a retro that only
fires on success cannot see it.

It writes **the run's one human-facing document** and records its path on the
epic, which is what lets a later session find it from the epic id alone.

**Read it before you hand it over**: you own that it is true, not that it is
written. **A run ending without one has failed**, whatever else it achieved.

## Post-flight — the cleanup script

The last act, after the retro and the report:

```sh
"$SCRIPTS"/sprint-cleanup.sh <epic>         # add --dry-run to see it first
```

It prunes worktrees whose agent has drained and pushes the tracker, printing
the command behind each removal.

**Exit 1 means something was left standing on purpose** — a worktree whose
agent still holds a queue is a live worker or a dead one, and no script can
tell them apart. Resolve each before treating the sprint as torn down. **Do not
force it.**

**Then cancel the durable trigger** armed at first dispatch — a harness action
no script can do. One left armed wakes an agent into a finished sprint, which
trains the reader to ignore it.

## Reference files

Load the one that matches where you are, in full rather than previewing; each
is one hop from here. **The steps themselves are named in order above** — each
is a skill you invoke or a script you run, and its section says which.

| Working on | Load |
|---|---|
| Pre-flight blocked; the remedy | `references/pre-flight.md` |
| A project with no declared `[[checks]]`, `[models]` or `[review]` | `references/project-setup.md` |
| The manifest, the epic shape, amendments, the ledger | `references/run-record.md` |
| Concurrency entry and tier routing | `references/dispatch-policy.md` |
| Sending a dispatch out and verifying what returns | `references/dispatching.md` |
| Handing a ceremony phase its agent id | `references/dispatching.md` |
| Dispatching one bead or a batch | `references/prompt-dispatch.md` |
| Creating, merging and removing a worktree | `references/worktree-mechanics.md` |
| How to call a script, and what it leaves behind | `references/scripts.md` |
| Live progress: the Monitor and the trigger | `references/progress-streaming.md` |
| A subagent that crashed or timed out mid-ceremony | `references/recovery.md` |
| Where the run is; when it is done; which death you have | `references/supervision.md` |
| A bead interrupted mid-implementation; all principles | `references/operating-principles.md` |
| Resuming a run whose state you do not hold | the **`sprint-recovery`** skill, then re-enter at groom & plan |

**A prompt file IS the prompt** — send it whole, with no block to extract.
You resolve the placeholders: `references/dispatching.md`.

**Scripts are executed, not read** — `references/scripts.md` carries every
interface, so nothing sends you into a script's source.

## Operating principles

Four bind **you**. The rest belong to the subprocess that owns them, and the
full list is in `references/operating-principles.md`.

- **Don't read source files.** Reading a file to decide something is a
  subagent's job; you route.
- **Nothing escalates to a human mid-run.** Decide, record it as a finding of
  its kind, and the retro surfaces it. This binds you as much as your
  subagents: `references/operating-principles.md`.
- **Findings serve two audiences.** A finding about **this project** becomes a
  bead here. A finding about **a skill this run used** goes to the destination
  the project configured for them, and is never filed in this project's backlog
  — a project that accumulates proposals it cannot act on learns to ignore the
  channel.
- **Auto-chain dispatches.** The gate already asked, so emit the progress
  report and **dispatch what is ready in the same turn** — the report is a
  notice, not a checkpoint. The turn ends when the front is dry, a bead
  carrying `human` is reached, or the circuit breaker trips. **Ending a turn
  with a non-empty front and none of those is a defect, not a pause**: nobody
  is waiting to restart it.

## Retro this skill

**`sprint-retro` owns it, mechanically.** A finding about this skill is
captured as `retro:tooling:<name>` and routing picks it up by kind — nobody
decides whether the run "felt like it struggled".
