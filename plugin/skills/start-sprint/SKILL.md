---
name: start-sprint
description: Open a bead-bonanza run — probe the environment, resolve the run
  epic by admission, and gate its queue for readiness before anything is
  dispatched. Use when a bead-bonanza orchestrator is starting a run, when
  asked to set up or begin a sprint over a set of beads, or when a run must be
  re-opened after its scope changed. Turns whatever the user gave — an epic,
  several epics, bead ids, a label, a priority — into exactly one run epic,
  captures the manifest and checks baseline on it, and returns the facts
  dispatch needs. Returns those facts; it dispatches nothing.
version: "0.1"
user-invocable: false
---

# start-sprint

Open a run: measure the environment, resolve the **one epic** the run is
scoped by, and gate its queue. Phases 0, 1 and 1.5 of a bead-bonanza run.

**You return facts; you dispatch nothing.** The orchestrator that invoked you
takes the run from there.

## Inputs

| | |
|---|---|
| `{SCRIPTS}` | the absolute path to **bead-bonanza's** `scripts/` directory. The scripts are that skill's; you run them, you do not ship them |
| the user's scope | an epic, several epics, bead ids, a label, a priority — or nothing, in which case ask once |
| standing authorization | what the user pre-authorized, WHEN they did |

## What you return

- `BB_EPIC` — the run epic. Its id **is** the run id, and
  `bd ready --parent <epic> --exclude-type=epic` **is** the queue.
- The pre-flight `BB_*` facts, eval'd from the script's stdout.
- The admission record: which beads were admitted, which were **excluded** and
  why, which were **moved** and from where.
- The readiness verdict: which beads are groomed, and which were bounced with
  the gap a human must close.
- Anything that BLOCKED, named, with the run not started.

**Say what you could not establish** rather than omitting it. A run opened on
an unverified fact is the failure this phase exists to prevent.

## Phase 0 — Pre-flight

One script measures the environment; you act on what it returns. It runs
**twice**, and the split is not ceremony: admission (Phase 1) resolves the run
epic by writing to the tracker, and a tracker write must not happen before the
script has shown that a closure can be pushed.

```bash
eval "$({SCRIPTS}/pre-flight-check.sh)"              # environment
#   ... Phase 1 admission resolves the run epic ...
eval "$({SCRIPTS}/pre-flight-check.sh "$BB_EPIC")"   # + baseline
```

The second pass verifies the epic exists and IS an epic, counts what is ready
in scope, creates the run directory and captures the checks baseline. **The run
id is the epic id** — there is no generated id.

Eval the stdout — it sets `SCRIPTS` and the `BB_*` facts this
document refers to elsewhere. Read the stderr report once. Exit **0**
proceed; **1** blocking, named in the report and fixed before Phase 1 — a
missing `AGENTS.md` block means read `references/project-setup.md` and draft
it WITH the user, never invent checks or a lineup from memory; **2** you
called it wrong.

Those facts are ledger lines with a command behind them. **Do not re-assert
from memory what the script just measured** — push mechanics in particular
vary by container. WHEN `BB_PUSH_NEEDS_REFSPEC=yes`, push with an explicit
refspec (`git push <remote> <branch>`) every time, and never let a pipe to
`tail` mask a push's exit status. Verifying that a push **landed** is not
conditional on that fact and is not `git status -sb` — it is the end-of-wave
check the orchestrator performs on every wave.

**Commit trailers.** WHEN `CLAUDE.md`/`AGENTS.md` states a commit-trailer
convention (`Co-Authored-By`, a session URL, etc.), resolve it once here and
put it in the manifest's `trailers` key on the run epic — a session URL varies
per session, so it is per-run and never committed. Every dispatch prompt then
passes `--trailers-from {EPIC}`, and the close ceremony fetches the text from
the tracker and appends it to the commit body, so no subagent needs
`git commit --amend` to add them. Omit the flag entirely when the run resolved
no trailers.
WHEN the project states no trailer convention, leave `{TRAILERS_FILE}` unset
and omit the flag; the close ceremony then appends no trailers.

What no script can settle:

- **Pull requests are the project's business, not bead-bonanza's.** Whether a PR
  is opened, when it opens, and when its CI runs are the project's policy — and
  not every project uses a forge with pull requests at all. Do not ask, do not
  infer it from workflow files, and never treat a PR as cover for running fewer
  checks. **The project's declared checks run after every bead, always.** That
  is what it is for.

  WHERE the project does run CI on a pull request and the run ends with one
  becoming reviewable, the failures that surface then are the run's to fix, the
  same as any other gate failure.

- **Harness worktree directory.** Confirm it is gitignored: create the first
  worktree, then check `git status --porcelain` is clean. Checking only before
  that worktree exists tests nothing — the harness may place it inside the
  repo, and untracked worktrees there read as uncommitted work to anything
  that inspects `git status`.
- **Agent-tool availability — BLOCKING.** Whether you have an Agent/Task tool
  at all is a harness property, invisible from a shell. No script can check it,
  so **you** assert it: with no Agent tool, set `BB_ASSUME_NO_AGENT=1` in the
  pre-flight environment so the fact reaches the manifest like every other, and
  **stop**. Do not start a run.

  This is a blocking condition rather than a degraded mode because subtracting
  delegation leaves nothing bead-bonanza does: no subagents means no context
  isolation, which is its whole point, and with it go worktrees, batching, tier
  routing, concurrency, `plan-sprint` and the grooming subagent. What survives
  is the per-bead ceremony — genuinely valuable, and a different product.

  **Say so and hand over**: the **`run-bead`** skill takes one bead from
  claimed to committed with no orchestrator at all. Working a queue is then
  `bd ready`, then `run-bead` per bead, which is an honest description of what
  a no-Agent session can do.

**`BB_DOLT_WRITE` is a blocking fact where it reads `no`.** It means the
project syncs beads to a Dolt remote and this container cannot write it, so
every closure the run makes dies with the container — the exact loss the close
ceremony exists to prevent. Stop and fix the environment; do not run and hope.
`none` means no Dolt remote is configured, which is not a fault (the project
may track beads another way) and `yes` is the normal case. Pre-flight measures
it even where a project's own session hook also does; a green hook is not a
substitute for this fact.

`BB_WORKTREE_BASEREF` is the same kind of blocking fact — pre-flight now exits
non-zero when it is not `head`. See "Every bead runs in its own worktree"
above.

## Phase 1 — Admission: resolve the run epic

**A run is one epic.** Its id is the run id, its descendants are the queue, and
it is what recovery reads. The user hands over whatever they have; admission
turns that into the epic.

If they gave no scope at all, ask once: "Which beads? (epic / label / priority
/ explicit ID list)" — the only Phase-1 question allowed.

**Resolve a label or priority to a bead set first** (`bd ready --label <l>`,
`bd ready --priority 1`, `-n 50` to raise the limit). Then every input is beads
and/or epics, and these rules apply:

| what they gave | what you do |
|---|---|
| **one epic** | use it. Create nothing, move nothing |
| **beads with no parent** | create the run epic, adopt them. Nothing is lost — they had no parent |
| **beads that already belong to an epic** | **ask first**, naming each bead and the epic it would leave. On decline, EXCLUDE those beads and run without them |
| **several epics** | create the run epic over them, adopting the EPICS themselves. Each keeps its own children, so no bead moves and no source epic is emptied. Confirm — it gives those epics a parent they did not have, which restructures the project's top level |

**Never move a bead to make the run tidier.** Moving a bead out of its epic
makes that epic report complete when it is not: a feature epic with one closed
and one open child reads `1/1 complete (100%) — eligible for close` once the
open one is taken. That is a project's structure damaged by a transient run,
and there is no undo step that survives a crash.

**Declining is a normal outcome, not a failure.** The run proceeds with what was
admitted; the excluded beads are recorded and named in the final report. A user
must never have to choose between running bead-bonanza and keeping their
structure.

**Never flatten a nested epic**, and never adopt the leaf beads out of the
epics you were given. The queue is correct through nesting; only the wave
preview and the progress count degrade, and both are recomputable. Flattening
trades a real structure for a display detail, and emptying a source epic of its
open work is the false-complete failure above.

**What a nested run epic costs, stated once at run start:** `bd swarm validate`
and `bd swarm status` are one level deep, so they report sub-epics rather than
leaf beads. The wave preview comes from `plan-sprint` instead, and the progress
count from `bb-status.sh run <epic>`, which walks the tree. Dispatch is
unaffected — the re-queried ready front is transitive and honours dependency
edges that cross between the adopted epics.

**Cross-epic dependency edges are the reason several epics are being run
together, and `swarm validate` reports them as warnings and acts on none of
them.** Surface every such warning at run start rather than dropping it: an
edge from a bead in one adopted epic to a bead in another is exactly what makes
the wave preview untrustworthy, and it is what the ready front gets right.

Then print the scope: the epic id, the ready count, and any exclusions.

### Scope stays live, and the delta is reported

Membership of the epic IS the scope, so a bead added to it mid-run is in scope.
That is deliberate — it is why the run carries no frozen id list.

**On every start and restart:** recompute the queue, compare it against what
the manifest recorded, and state what changed before dispatching. Where the run
was admitted from a label or priority, re-run that filter too and offer new
matches for admission under the rules above. A run that silently grew is the
failure; one that grew and said so is the feature.

The Run checklist copied into your response (above) is the progress-visibility
mechanism, and it is sufficient. `bd` is the only tracker: a second progress
store is a correctness hazard, not redundancy.

## Phase 1.5 — Readiness gate (grooming)

bead-bonanza dispatches **only groomed beads.** Every later phase already
assumes grooming's outputs: Phase 1's `bd ready` ordering comes from the
dependency graph grooming wires; this gate's own cheap check reads the
`decision-required` label grooming sets (DoR check 9); the dispatch plan reads the
tier assignment grooming makes.

**Raise every bounced bead into `bd`'s human queue.** A bead this gate bounces
carries `decision-required`; also add the **`human`** label, which is what
`bd human list` and `bd human stats` read. The bounce then has a queryable
state rather than living only in a report line the user has to notice — and it
survives the session that produced it, which a report line does not. Surface
`bd human list` at the point the report prints the bounce list, so the two
cannot disagree.

**Resolve by removing the label, never with `bd human respond`.** The answer
goes on the bead as a `bd comment`, then `bd update <id> --remove-label human`
(and `decision-required` once the gap is actually closed), and the bead
re-enters the readiness gate on the next pass — no full re-run. Two independent
reasons not to use `respond`: it is documented to add a comment **and close**
the bead, and a bounced bead is not finished; and on bd 1.2.2 it fails outright
— `bd human respond <id> -r "..."` errors with `cannot resolve issue ID
"<id>": storage is nil`, reproduced in two separate embedded workspaces, while
`list` and `stats` work in both.

`bd gate --await human:<prompt>` is the other candidate and is **not** used
here: gates are created by a formula step's `gate` field, and bead-bonanza runs
no formula. A human item stopping the chain is the behaviour that matters, and
the label already delivers it. An ungroomed queue makes every one of them
unreliable — and handing a fresh, zero-context subagent an under-specified
bead is the single most expensive failure mode a run has: it detonates
mid-wave, after later beads have built on it.

So before dispatch, gate the queue against the **Definition of
Ready** — the checklist owned by the `spec` skill
(`references/shared/work-order-definition-of-ready.md`) and enforced by
`groom-beads`, which materializes work orders as beads. Those skills are the
authority; do not restate the checklist here.

### 1. Cheap check (orchestrator, read-only)

Most runs pass in seconds because the user groomed first. Per bead, confirm the
Definition of Ready's checks hold — read them from
`references/shared/work-order-definition-of-ready.md`, not from memory — plus
one signal it does not carry: `bd lint` is clean for the bead, WHERE the
installed `bd` has it. Parallelize; this is read-only. A bead failing check 9
(an open decision) is bounced, not questioned mid-run.

A queue that is entirely ready makes this phase a one-line no-op for *grooming*:
say so and hand back — the orchestrator still dispatches one subagent for
planning alone.

### 2. Groom what fails (delegate — don't groom in the orchestrator)

If any bead fails the check, do NOT dispatch it and do NOT fix it inline —
grooming reads source and wires graphs, which belongs in a subagent (see "don't
read source files in the orchestrator"). Dispatch a **grooming subagent** over
the ungroomed subset (template below); it runs `groom-beads` to the Definition
of Ready and returns a per-bead readiness report plus a **wave preview** — its
candidate partition under dependencies and C4 collisions. **Its final act is to
invoke `plan-sprint`**, which turns that preview into the run's wave table:
it has already read every bead, so a second agent would only re-derive the same
partition and leave two to reconcile.

When it returns, **re-run `bd ready`** with the run's filter: grooming may have
wired new dependencies or split a boulder into an epic + children, changing both
the ordering and the membership of the queue.

### 3. Bounce what grooming can't clear

Grooming, like this phase, bounces gaps rather than guessing. A bead it cannot
bring to Ready — needs a human decision it can't make, can't be sized without a
split, rests on an unexamined premise — does **not** enter the run. Exclude
these from the queue and surface them to the user at the end of this phase (one
line each + the gap). The run proceeds with the Ready beads; ungroomable beads
are the user's to resolve, not a subagent's.

Bounced beads are the run's **single decision point**: surfaced with their
open question, resolved or groomed before dispatch, then the run proceeds
with the Ready remainder — an *admission* decision made once, up front, not
mid-flight.

### 4. Admitting emergent work mid-run

During execution (Phase 3) subagents discover work. Some of it belongs in this
run; some is scope creep wearing a matching label. Telling them apart is the
point — pulling in the wrong thing is as real a failure as leaving a blocker
unaddressed. This is where the autonomy the run was launched for lives, and
where it stops.

**The test is necessity, not resemblance:** emergent work is in-scope only when
a bead *already in the run* cannot be correctly and completely finished without
it. These shapes qualify:

- **an unblocking prerequisite** — the assigned bead genuinely cannot be
  completed until this is done;
- **an acceptance-criteria decomposition** — a piece the assigned bead's own AC
  cannot be satisfied without (a sign the bead was under-groomed, larger than it
  looked);
- **an expected test-coverage gap** — coverage the assigned scope already
  implied should exist and doesn't.

Everything else — adjacent cleanups, refactors "while I'm here", a new
capability no assigned bead's AC demands — is **out of scope even if it matches
the run's filter.** Matching the label is necessary but not sufficient;
"required to finish a bead already in the run" is the bar.

**Mechanics.** The discovering subagent only *files* the bead (`bd create`) with
a `discovered-from` edge to its origin and a one-line note of which shape it is
(or "out-of-scope"); it never implements it. If its own bead is truly blocked by
it, it takes the Blocked path naming the new id. The **orchestrator** then
decides:

- **In-scope** → run it through this gate (groom → re-`bd ready`), wire the
  `blocks` edge so it orders ahead of the bead it unblocks, and dispatch it
  under the standing auto-chain authorization. No new prompt — this is the
  intended autonomy. Grooming closes decisions along with the rest of the
  Definition of Ready. (A groomed prerequisite that still lands a
  `decision-required` label stops for the user, per the standing-authorization
  terms — and is raised into the human queue with it.) Grooming and wiring the edge are tracker writes, which need not wait
  for a wave boundary.
- **Out-of-scope** → leave it filed and linked, do NOT dispatch it, and surface
  it in the final report for the user to triage. The standing authorization
  finishes the launched scope; it never enlarges it.

**Ballooning is a stop signal.** If "prerequisites" multiply — a discovered
blocker that is really a new epic, or more than a handful of new in-scope beads
off a single bead — stop and surface it. That means the original scope was
mis-groomed, which is a human's call, not an autonomous expansion.

The grooming subagent prompt template is in `references/grooming-prompt.md`.
