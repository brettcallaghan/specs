---
name: bead-bonanza
description: Process many beads autonomously by orchestrating subagents. Use
  when the user asks to work on multiple beads at once — e.g. "work on all open
  beads with label X", "address all P1 beads", "clear the bead backlog". Triages
  the queue, gates every bead for readiness (grooming under-specified beads via a
  subagent, or bouncing what can't be made ready), surfaces both explicit and
  implicit decisions up front, dispatches via the cheapest viable execution mode
  (doc fast lane / shared-surface batch / sequential), runs every bead in its
  own git worktree, streams per-bead progress, and lands each bead in a single
  commit.
version: "2.0"
---

# bead-bonanza

**You are the coordinator, and routing is the only thing you do.** Each step
below is a subprocess: you invoke it, read what it returns, and decide what
happens next. You never hold a subprocess's method.

**What you retain — the whole context budget:** the run epic id, the review
pass counter, the authorization set, where the wave loop is, and what the last
subprocess returned. Not the wave table, not any subprocess's reasoning, not
the evidence behind a verdict you were handed.

**Act on the verdict, not on the evidence.** A subprocess whose output you have
to reason about has pushed work back up the tier it was meant to absorb.

Read `CLAUDE.md` and `AGENTS.md` once at the start, before delegating.

**This skill is language-agnostic** — it hard-codes no build, test, lint or
format command. Validation comes from the checks the project declares in
`sf.toml`, run by `scripts/bb-checks.py`: exit 0 means every check passed,
anything else means something failed, and the output says what.

## The shape of a run

```
1     pre-flight ............................ pre-flight-check.sh
2     resolve scope ......................... sprint-scope
      └─ AUTHORIZATION GATE ................. HUMAN, once
┌─ 3+4 groom & plan .......................... sprint-plan → groom-beads
│  5   sprint ................................ coordinator + subagents
│      ├ create the sprint branch, once ..... coordinator
│      ├ 5a dispatch a wave ─┐ repeat until   sprint-wave → sprint-batch
│      ├ 5b merge            │ the ready      coordinator
│      └ 5c supervise       ─┘ front is dry   bb-status.sh
│  6   review ................................ sprint-review → declared reviewer
└───── findings admitted AND passes < 3 ? ─── yes ──┘
7     PR review ............................. HUMAN
8     retro ................................. sprint-retro
9     report ................................ sprint-report
10    post-flight ........................... sprint-cleanup.sh
```

**Two nested loops, both bounded**: the wave loop ends when the ready front is
dry, the review loop at 3 passes or no admitted findings. Neither exhausting is
a failure — what is left is filed and named.

## Run checklist

Copy this into your response at run start and check it off as you go. It lives
in your response, not a task tool — nothing to fall out of sync with `bd`.

```
bead-bonanza run:
- [ ] 0. pre-flight run, BB_* facts eval'd
- [ ] 1. `sprint-scope` invoked: run epic resolved, admission record
        (admitted / excluded / moved), authorization gate taken
- [ ] 1b. pre-flight re-run with the epic: checks baseline on the manifest
- [ ] 2. Dispatch plan: batch + tier + push mode per bead
- [ ] 2b. Sprint branch created (or reused), `branch` + `base_sha` in manifest
- [ ] 3. Execute waves via `sprint-wave`, merging each accepted branch
- [ ] 3b. Review via `sprint-review`; loop back to groom & plan when findings
        were admitted and the pass budget is not spent
- [ ] 4. Final report via `sprint-report` (incidents first), then `sprint-retro`
- [ ] 5. Post-flight: `scripts/sprint-cleanup.sh`, then cancel the durable trigger
```

**Most of it is checkable rather than assertable.** `bb-status.sh run <epic>`
verifies the mechanical half — counts, health set, lint clean *within this
run's queue*, no open human item, every ready bead tiered, every admitted bead
closed — then **names what it cannot verify**, so a clean report never implies
the whole checklist passed. Run it at every wave boundary and at the end.

---

## Opening a run — pre-flight, then `sprint-scope`

Three things happen before the first bead, in this order:

```bash
eval "$(scripts/pre-flight-check.sh)"              # 1. environment
#   2. invoke `sprint-scope` — it returns BB_EPIC
eval "$(scripts/pre-flight-check.sh "$BB_EPIC")"   # 3. + checks baseline
```

**1 and 3 are a script**, not a skill: run it, eval the stdout, read the stderr
report once. It names each blocker AND its remedy, so there is nothing to
weigh. `references/pre-flight.md` only when it blocks, and **a blocking
condition means the run does not start** — do not work around it.

**2 is `sprint-scope`.** It turns whatever the user gave into exactly one run
epic, takes the run's single authorization gate in the same interaction, and
dispatches nothing. What you need from it:

- `BB_EPIC` — its id is the run id and its descendants are the queue
  (`bd ready --parent <epic> --exclude-type=epic`).
- The admission record — admitted, **excluded** and why, moved and from where.
  A run that narrowed its own scope says so in the final report.
- The authorization set as decided, including the rows that took their default.

**The readiness gate is not here and not in `sprint-scope`** — it belongs to
grooming, because readiness is a property of a groomed bead and asking before
grooming answers a question about beads that are about to change.
## Groom & plan — invoke `sprint-plan`

**One dispatch, two returns.** `sprint-plan` grooms every bead that is not
Ready (invoking `groom-beads`), bounces what it cannot clear, and — while it
still holds every bead in context — partitions what came out:

1. **The readiness verdict** per bead. You dispatch only ready beads; bounced
   ones are raised into `bd human` by the skill and named in your report.
2. **The wave table**, plus one line per override of its own preview.

**They are one step because they are one read** — partitioning needs exactly
what grooming just acquired, so a second agent would re-read the whole queue to
learn what the first already knew. **One subagent, never two**; where nothing
needed grooming, it plans alone.

**Do not store the wave table.** It is derived from the ready front, and a
stored copy only drifts.

**Three things stay yours** — the run manifest, the swarm registration, and
the run's amendments. All three: `references/run-record.md`.

## The sprint branch — you create it, once, before the first worktree

**Immediately before dispatching wave 1**, not earlier. A run that bounced
every bead has nothing to branch for, and a branch cut before grooming records
a `base_sha` from before the grooming commits land.

```bash
br="$(bd show "$BB_EPIC" --json | jq -r '.[0].metadata.branch // empty')"
if [ -n "$br" ]; then git checkout "$br"          # RESUME — create nothing
else
  br="sprint/<epic-suffix>"
  git rev-parse --verify "$br" >/dev/null 2>&1 && { echo "branch $br taken"; exit 1; }
  git checkout -b "$br"
  bd update "$BB_EPIC" --set-metadata branch="$br" \
                       --set-metadata base_sha="$(git rev-parse HEAD)"
fi
```

**Read the manifest before you create anything.** Reuse is not an optimization
— it is the whole of what makes a resumed run land on the same branch. **A name
already taken by an unrelated branch stops the run here**, before any worktree
exists.

**`base_sha` is recorded at creation because that is the only moment it is
unambiguous.** Deriving it later from a merge-base assumes the base branch has
not moved, and on a long run it will have.

Three things consume it and none may create it: the review scopes its diff
`base_sha`..sprint head, every worktree branches from sprint HEAD, and the pull
request opens against it.
## Every bead runs in its own worktree

**This is universal, not opt-in.** Each dispatched bead gets a git worktree
named for its hash suffix (`jh6`, `92q.14` — not the full id), created from the
sprint branch and carrying a branch of the same name. Nothing is implemented in
the shared checkout except by the orchestrator itself.

**Presence is the in-flight signal.** A standing worktree means its bead is not
yet accepted — you remove it as the final act of accepting, after the branch
merges. `git worktree list` is the complete answer to what is still in flight;
there is no mtime heuristic anywhere.

**Hard gate — verify `BB_WORKTREE_BASEREF` is `head` before the first worktree
of EACH wave, and do not dispatch otherwise.** It is global user state no
repository owns, so it can change between waves. A worktree cut from the
default branch cannot see what the sprint branch already carries, and the
merge afterwards reads as a conflict rather than as a mistake.

**You merge, one at a time; subagents never do.** A chained batch merges by
fast-forwarding the LAST bead's branch, which carries its predecessors. Take
`bd merge-slot acquire` before each and release after — not to serialize, which
the single-merger rule already does, but so a stalled wave has a queryable
answer and a resumed session cannot merge alongside the original.

**A `--ff-only` refusal means different things in a sequential and a concurrent
wave, and reading it wrong stops a healthy one.** That rule, the base-ref
remedy, how `bd` behaves inside a worktree, and the create/merge/remove
lifecycle: `references/subagent-prompts.md`, "Worktree mechanics".
## Execute the waves — invoke `sprint-wave`

Per wave, from the wave table: **invoke `sprint-wave`**. It dispatches a
subagent per batch, verifies every return against the tracker and git rather
than against the subagent's account, and returns one line per bead —
`completed` / `blocked` / `skipped`, each with a sha — plus the push verdict.

**Then you merge**, one branch at a time, removing each worktree as you accept
its bead. `sprint-wave` merges nothing.


**Re-query the ready front at every wave boundary** — `bd ready --parent <epic>
--exclude-type=epic`. **That query is the order, not the stored table**: closed
beads leave it, beads the last wave unblocked enter it, and a bead added
mid-run is in scope by construction.

**A tripped circuit breaker stops the run, not just the wave.** Three
consecutive blocked or failed beads means something systemic, and the next wave
will meet it too.

**How a wave goes out** — concurrency entry conditions, tier routing and
escalation, push mode: `references/dispatch-policy.md`.

### The run's health — the one thing nothing else can express

A stuck agent and a slow one look identical in git and in bead status, so
health is **asserted** rather than derived: you set it, and `bb-status.sh run`
prints the evidence for and against beside it.

```sh
bd set-state <epic> health=<value> --reason "<why>"
```

Closed set — at run start and every wave boundary:

| value | meaning |
|---|---|
| `healthy` | beads are closing and nothing is failing repeatedly |
| `slow` | nothing closed recently, but agents are alive |
| `stalled` | nothing closed recently **and** no live agent |
| `degraded` | progressing, but with repeated failures or escalations |
| `blocked` | cannot proceed without a human; pair with `bd human` |

`slow` and `stalled` are the distinction the dimension exists for. The label
**replaces** rather than accumulates, so there is one current value and no
history litter. **An unset health reads as repairable** — a watchdog seeing no
label cannot tell a healthy run from an abandoned one.
### Watching it, and picking it back up

**`bd swarm status <epic>`** is the durable answer to "where is this run" —
computed from the beads, so it survives any crash. On a **nested** epic it is
one level deep and reports sub-epics rather than leaf beads; fall back to the
re-queried ready front plus a walked count, and say which you used.

**Arm a `Monitor`** on the progress file for any subagent expected past 5
minutes (any non-trivial bead, given full test runs), `persistent`, stopped
with `TaskStop` when the wave drains — and, before the first dispatch, the
durable out-of-session trigger, cancelled at post-flight. Both:
`references/progress-streaming.md`.

**When something dies**, take the one you are in: a subagent past 5 minutes →
`references/progress-streaming.md`; one that crashed mid-ceremony →
`references/recovery.md`; **the orchestrator itself** → invoke
`sprint-recovery` — a skill rather than a file, because the session that needs
it may be fresh with nothing loaded and no path to resolve.
### When the run is done

**`bb-status.sh run <epic>`** — computed from state, not from your recollection
of what you dispatched. A run is done when every bead the manifest's `admitted`
list names is closed or accounted for as blocked-with-a-reason. **Beads
excluded at admission are reported alongside**: a run that finished its
admitted set while three sat out is done over a *narrowed* scope and must say
so.

It walks the epic's descendants rather than calling `bd epic status`, which
counts direct children only and on a nested epic answers about sub-epics.
`bd list --parent` is direct-children-only too; only `bd ready --parent` is
transitive, and it omits closed beads.

Exit 0 clear, 1 repairable (no admitted set recorded), 2 stop.

## Review the run's own diff — invoke `sprint-review`

When the ready front is dry, **before the report**: invoke `sprint-review`. It
runs the reviewer the project declared in `sf.toml`'s `[review]`, scopes it to
`base_sha`..sprint head, files every finding as a bead, and classifies each as
in or out of the run's authorized scope.

**The loop is yours, and only the loop.** `sprint-review` owns finding things;
you own the counter, the budget, and the decision to go round again — findings
admitted **and** passes under budget means re-enter, anything else means done.

**Re-enter at groom & plan, never at dispatch.** Review *files*; it does not
groom. An admitted finding earns its readiness like any other bead, or bounces
like one.

**The budget is 2 loops — three passes** unless the gate said otherwise, and
**findings from the human's pull-request review do not consume it**: that
counter bounds the machine reviewing itself. **Keep it in the manifest as
`review_passes`**, not in your head — a resumed run that restarted it would
loop forever.

**Exhausting it is not a failure.** What is left stays filed and is named in
the report. A third pass that still finds things is information for the human,
not a reason to keep going.

**No reviewer declared** returns a skip, and the report says so. Do not
substitute your own review.
## Report — invoke `sprint-report`

After the queue drains (or the circuit breaker trips), invoke `sprint-report`.
**Pass it the run epic id and nothing else** — everything it needs is on that
epic, which is what lets a resumed session write a report for a run it did not
hold.

It leads with the **incidents**, then the results table, routing evidence,
follow-up beads, exclusions, and what the run could not verify, and records its
path in the manifest.

**Read it before you hand it over** and add nothing you did not already know —
you own that it is true, not that it is written.

**A report is always produced**, including for a run that stopped early. A run
that ends without one has failed, whatever else it achieved.

### Then the retro — invoke `sprint-retro`

**After the report, whose tables it reads**, and **on every run** — including
one that closed nothing. A sprint that bounced its whole queue produced no code
and a great deal of evidence about grooming; a retro that only fires on success
cannot see the failures.

It harvests every finding, routes each by its stamped kind, reasons them into
candidate issues, puts those to the user once, and files what they accept —
beads here, `.sf/outbox/` files for the skills it ran on.

**Two of the nine kinds never become issues.** `proof` (how a claim was
established) stays in the report; `environment` (a dependency's behaviour no
bead could have carried) goes to the environmental-failure ledger. Measured at
7 of 73 findings on a real run, so expect them.

**Incidents are its best material** — each marks a place the corpus failed to
answer a question a bead needed answered, which is a finding against a
*document* rather than a complaint about the run.
## Post-flight — `scripts/sprint-cleanup.sh`

The last act, after the report and the retro:

```sh
scripts/sprint-cleanup.sh <epic>            # add --dry-run to see it first
```

It prunes worktrees, releases the merge slot, closes the swarm molecule, sets
terminal health and pushes the tracker, printing the command behind every
removal so a reader can tell a teardown from a claim of one.

**Exit 1 means something was left standing on purpose** — a worktree whose bead
still reads `dispatch_state=running` is a live worker or a dead one, and no
script can tell those apart. Resolve each before treating the sprint as torn
down. **Do not force it.**

**Then cancel the durable trigger** armed at first dispatch — a harness action
no script can do. One left armed wakes an agent into a finished sprint, which
trains the reader to ignore it.
---

## Reference files

Load the one that matches where you are. Read it in full rather than
previewing; each is one hop from here.

**The steps themselves are named in order above** — each is a skill you invoke
or a script you run, and the section for each says which.

| Working on | Load |
|---|---|
| Pre-flight blocked and you need the remedy | `references/pre-flight.md` |
| A project with no declared `[[checks]]`, `[models]` or `[review]` | `references/project-setup.md` |
| The manifest, the swarm registration, the run's amendments | `references/run-record.md` |
| Concurrency entry conditions, tier routing, push mode | `references/dispatch-policy.md` |
| Dispatching an execution subagent | `references/subagent-prompts.md` |
| How to call a script, and what it leaves behind | `references/scripts.md` |
| A subagent past 5 minutes; the Monitor and the durable trigger | `references/progress-streaming.md` |
| A subagent that crashed or timed out mid-ceremony | `references/recovery.md` |
| Resuming a bead interrupted mid-implementation; the full principle list | `references/operating-principles.md` |
| Resuming a run whose state you do not hold | the **`sprint-recovery`** skill |
| The Definition of Ready the readiness gate enforces | `references/shared/work-order-definition-of-ready.md` |

**Scripts are executed, not read** — `references/scripts.md` is how to call
one, so nothing sends you into a script to learn its interface.

## Operating principles

Four bind **you**. The rest belong to the subprocess that owns them, and the
full list is in `references/operating-principles.md`.

- **Don't read source files.** Reading a file to decide something is a
  subagent's job; you route.
- **Nothing escalates to a human mid-run.** Not an ambiguous acceptance
  criterion, not a contract change, not a silent spec. Decide, record it as an
  **incident** — `retro: incident —` on the bead, `amend: incident on <id> —`
  on the epic — and the report names it first. A recorded decision is reviewed
  at the pull request and revised in a later run; that is cheaper than a night
  blocked on a question nobody is awake to answer. This binds you as much as
  your subagents.
- **Findings serve two audiences.** A finding about **this project** becomes a
  bead here. A finding about **a skill this run used** leaves through
  `.sf/outbox/<skill>/` and is never filed in this project's backlog — a
  project that accumulates proposals it cannot act on learns to ignore the
  channel.
- **Auto-chain waves.** The gate already asked, so emit the wave's report and
  **dispatch the next in the same turn** — the report is a progress notice, not
  a checkpoint. The turn ends when the queue is dry, a `decision-required` bead
  is reached, or the circuit breaker trips. **Ending a turn with a non-empty
  queue and none of those is a defect, not a pause**: nobody is waiting to
  restart it.

## Retro this skill

**`sprint-retro` owns it, mechanically.** A finding about this skill is
captured as `retro:tooling:<name>` and routing picks it up by kind — nobody
decides whether the run "felt like it struggled". A run that felt clean can
still have produced several.
