---
name: sprint-plan
description: Groom a sprint's queue and partition it in one pass — bring every
  bead to Ready (invoking groom-beads for what fails, bouncing what grooming
  cannot clear) and return the readiness verdict plus the collision edges that
  serialize the run, with every bead assigned. Use when a sprint orchestrator
  is preparing a run epic's queue for dispatch, or when asked to gate beads for
  readiness, classify them as trivial / isolated / shared-surface, compute C4
  co-edit and sequencing overlap, group beads into batches behind one agent, or
  emit the `bd dep add --file` edge document a run is built from. Takes the run
  epic, resolves its queue from the graph, and dispatches nothing.
version: "0.3"
user-invocable: false
---

# sprint-plan

Take a run epic's queue from wherever it is to **groomed and partitioned**, in
one dispatch. Two returns: the readiness verdict per bead, and the collision
edges.

**They are one step because they are one read.** Partitioning needs exactly the
knowledge grooming just acquired — every bead's scope, surfaces and
dependencies — so splitting them means reading the whole queue twice. This is
read-heavy analytic work, which is why it runs in a subagent rather than in the
orchestrator, which stays small.

You are grooming and planning, not dispatching and not implementing. Nothing
here creates a worktree, edits a bead's code, or merges anything.

## Interface

**In** — all of it supplied by the caller; infer none of it:

| Input | Notes |
|-------|-------|
| The run epic | Its id is the run id, and its `--metadata` is the manifest the orchestrator already wrote |
| The **bead id list** | Explicit ids, and the epic they came from. Re-derive with `bd ready --parent <epic> --exclude-type=epic` rather than from any filter — admission decided that membership and may have excluded beads deliberately |
| Checks script and checks baseline | A gate recorded `fail` in the baseline sets "no findings beyond the baseline" as the bar, not exit 0 |
| Model lineup | Resolves a bead's tier (`metadata.execution_suggested_model`, else `metadata.tier`) to a model; leave it unset where the project declares none |
| Standing authorization | What the user authorized without further asking (PR, scope, review budget) |
| The grooming partition preview | The candidate grouping, WHEN one was produced |

WHEN an input is missing, say so in the return rather than substituting a
guess — a plan built on an inferred queue is the failure the enumeration rule
exists to prevent.

**Export `BEADS_ACTOR=plan-<run-id>` before your first `bd` call** — you write
labels and assignees, and unset they land under the harness default.

**Out** — two things, and nothing else:

1. **The readiness verdict** — per bead: ready, or NOT-READY with the gap a
   human must close. Bounced beads carry the **`human`** label, so the verdict
   is queryable and not merely reported.
2. **The plan — one edge file**, validated before you return its path:
   `references/partitioning.md`. Plus **one line per override** of the grooming
   partition preview saying what you changed and why.

**The partition is not a document — it is in the graph.** Which beads may not
run at once is a **collision edge**; which agent holds a bead is its
`assignee`; what runs next is `bd ready` re-queried. The orchestrator
dispatches from that front continuously, so there is no schedule to keep in
step with a graph that moves under it.

## The readiness gate — run it before you classify anything

A sprint dispatches **only groomed beads.** Every later phase already
assumes grooming's outputs: the ready front's ordering comes from the
dependency graph grooming wires; this gate's own cheap check reads the
open-decision verdict grooming reaches (DoR check 9); the classification below
reads the tier assignment grooming makes.

**Raise every bounced bead into `bd`'s human queue.** Add the **`human`**
label, which is what `bd human list` and `bd human stats` read. The bounce then
has a queryable state rather than living only in a report line the user has to
notice — and it survives the session that produced it, which a report line does
not. Surface `bd human list` at the point the report prints the bounce list, so
the two cannot disagree.

**Resolve by removing the label, never with `bd human respond`.** The answer
goes on the bead as a `bd comment`, then `bd update <id> --remove-label human`
once the gap is actually closed, and the bead re-enters the readiness gate on
the next pass — no full re-run. Two independent
reasons not to use `respond`: it is documented to add a comment **and close**
the bead, and a bounced bead is not finished; and on bd 1.2.2 it fails outright
— `bd human respond <id> -r "..."` errors with `cannot resolve issue ID
"<id>": storage is nil`, reproduced in two separate embedded workspaces, while
`list` and `stats` work in both.

`bd gate --await human:<prompt>` is the other candidate and is **not** used
here: gates are created by a formula step's `gate` field, and a sprint runs
no formula. A human item stopping the chain is the behaviour that matters, and
the label already delivers it. An ungroomed queue makes every one of them
unreliable — and handing a fresh, zero-context subagent an under-specified
bead is the single most expensive failure mode a run has: it detonates late,
after other beads have built on it.

So before dispatch, gate the queue against the **Definition of
Ready** — the checklist owned by the `spec` skill
(`references/shared/work-order-definition-of-ready.md`) and enforced by
`groom-beads`, which materializes work orders as beads. Those skills are the
authority; do not restate the checklist here.

### 1. Cheap check (read-only)

Most runs pass in seconds because the user groomed first. Per bead, confirm the
Definition of Ready's checks hold — read them from
`references/shared/work-order-definition-of-ready.md`, not from memory — plus
one signal it does not carry: `bd lint` is clean for the bead, WHERE the
installed `bd` has it. Parallelize; this is read-only. A bead failing check 9
(an open decision) is bounced, not questioned mid-run.

A queue that is entirely ready makes this phase a one-line no-op for *grooming*:
say so and hand back — the orchestrator still dispatches one subagent for
planning alone.

### 2. Groom what fails — invoke `groom-beads`

If any bead fails the check, do NOT dispatch it and do NOT fix it inline —
grooming reads source and wires graphs, which belongs in a subagent (see "don't
read source files in the orchestrator"). Dispatch a **grooming subagent** over
the ungroomed subset (template below); it runs `groom-beads` to the Definition
of Ready and returns a per-bead readiness report plus a **partition preview** —
its candidate grouping under dependencies and C4 collisions. **Its final act is
to invoke `sprint-plan`**, which turns that preview into the run's edges:
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

Subagents discover work while the run executes. Some belongs in this run; some
is scope creep wearing a matching label. **The discovering subagent only files
it**; whether it is admitted is the coordinator's, and where it is admitted it
comes back through this gate like any other bead.

**The test is necessity, not resemblance:** emergent work is in-scope only when
a bead *already in the run* cannot be correctly and completely finished without
it — an unblocking prerequisite, a piece the assigned bead's own acceptance
criteria cannot be satisfied without, or coverage the assigned scope already
implied should exist. **Everything else is out of scope even if it matches the
run's filter.** Adjacent cleanups, refactors "while I'm here", a new capability
no assigned bead's AC demands: matching the label is necessary and not
sufficient.

Groom an admitted one to Ready like the rest, and wire the `blocks` edge so it
orders ahead of the bead it unblocks. A groomed prerequisite that still fails
the decisions check is bounced with the `human` label, exactly as at the front
of the run.

**Ballooning is a stop signal.** WHEN "prerequisites" multiply — a discovered
blocker that is really a new epic, or more than a handful off a single bead —
say so in your return rather than grooming them all. That means the original
scope was mis-groomed, which is a human's call.

The grooming subagent prompt template is in `references/grooming-prompt.md`.

## Bias toward batching

**A batch is a collision cluster behind one agent.** Beads that would collide
if dispatched side by side are chained by collision edges and given ONE
assignee, so that agent takes them in order while everything else dispatches
around them. Your job is to find the largest safe batch, not to default to
one-bead-per-subagent for caution. A bead alone is the *exception*.

**Batching does not buy fewer check runs.** Every bead runs the project's
declared checks before its own close whether batched or not; a batch changes
who does the work and what runs alongside it, never what gets validated.

**Worktrees do not make batching redundant, because the collision is at the
merge, not between the agents.** Isolation is never grounds to relax the rules
below — why, in `references/partitioning.md`.

## Classify each bead

- **`trivial`** — pure documentation (md/mdx edits), single-line code swap,
  comment/regex-pin additions. No fixture regeneration needed; targeted unit
  tests suffice. Examples: "document a trade-off in docs/", "add an explanatory
  comment", "swap one equivalent stdlib call for another".
- **`isolated`** — shares no C4 component and no file with another queued bead.
  No `BLOCKS` / `RELATED` link to another queued bead.
- **`shared-surface`** — shares a `c4_changes` component with another queued
  bead, or touches files (especially snapshot/golden fixtures) another queued
  bead also touches, or has explicit `RELATED` links to another queued bead.

**Overlap comes from C4 metadata first, file paths second.** Grooming already
computed the blast radius onto each bead — `c4_changes` (components it edits)
and `c4_impacts` (downstream components affected but not edited). Read it; don't
re-derive it from prose. The `jq` and a worked example are in
`references/partitioning.md`.

1. **C4 metadata — primary.** Read `.metadata.c4_changes` and
   `.metadata.c4_impacts` from `bd list --json` for every queued bead.
   - Two beads sharing a `c4_changes` component are **shared-surface** whatever
     their file paths say — the same component through different files still
     collides. **Emit a collision edge between them, in the edge file** so they
     cannot both be ready at once, and give the cluster one assignee so one
     agent takes it in order. Shape and marking: `references/partitioning.md`.
   - Bead B whose `c4_impacts` intersects bead A's `c4_changes` consumes a
     contract A is changing. That is a **sequencing** constraint, not a batching
     one, and it holds with zero file overlap. It is a real `blocks` edge, not a
     collision edge, and the two are marked differently so a reader can tell
     them apart. WHEN you find one grooming missed, wire it in the edge file and
     say so in your return — never as silent batch order.
   - Disjoint changes with no impact intersection is isolated *architecturally*
     only — layer 2 still runs.
2. **File paths — fallback and corroborator.** Scan each bead's description for
   fixture names + source paths. It is the whole signal where `c4_*` is absent,
   and it runs alongside C4 where both exist. WHEN the two disagree — disjoint
   components, overlapping files — take the **union** as the collision set and
   note the discrepancy in the plan; do not silently prefer one signal.
3. **The grooming partition preview — starting point.** WHEN one was produced,
   start from it instead of re-deriving the grouping, and record whether you
   accepted or overrode it and why. You are the source of truth on packing;
   the preview is an input, not an instruction.

**Dependency ORDER is not yours to compute.** The graph already holds it, and
the orchestrator reads it by re-querying `bd ready --parent <epic>
--exclude-type=epic`. What you decide is **packing** — which beads would
collide, so that a collision edge chains them behind one agent instead of
letting them reach the ready front together.

Where the repo documents no C4 architecture, grooming skips the tagging and
`c4_*` is absent — file scanning is then the whole signal, and absent metadata
means *unknown*, never "no collisions".

**Say which class each edge you emit is.** A co-edit collision and a sequencing
constraint have different consequences; the marking that distinguishes them is
in `references/partitioning.md`.

## Choose a grouping per cluster

A grouping decides **how many beads share a subagent**, not how they are
isolated — isolation is settled above and is the same everywhere. Pick the
broadest one that's safe: you're maximizing what can run concurrently, not
minimizing check runs — those are per bead regardless.

**A batch subagent holds exactly one worktree at a time.** Holding two makes
worktree presence ambiguous and breaks the in-flight signal for every reader;
never pack a batch in a way that assumes otherwise.

1. **Batch** — ONE subagent processes a cluster sequentially, one commit per
   bead. Per bead, before its close: any regeneration, **targeted tests
   (non-negotiable)**, then the project's declared checks. Two clusters qualify
   and they take the SAME prompt, differing only in the surface it names:

   - `shared-surface` beads on one fixture/code area — that area is the
     surface, and the targeted tests are the ones covering it.
   - `trivial` beads — surface "documentation and chores; no shared code
     surface". Batched to stop them taking a dispatch slot each, not because
     they validate any cheaper. Cap ~6 per subagent.

   **Threshold:** batch when **≥2** beads share a surface and decisions are
   independent (bead A's choice doesn't change bead B's implementation). Even
   at 2 the gain is real: the second bead sees the first's work instead of
   racing it.

   **Don't over-split:** if 7 beads cluster on one surface and only 1–2 of them
   are "larger" (substantial new features), they can still go in one shared
   batch. Split only when bead A's implementation would conflict with bead B's,
   or when one bead's choice changes another's plan — never for caution alone.

2. **Alone** — reserved for: a single bead with no batch peers, OR a bead whose
   surface diverges from every other queued bead, OR explicit user override.
   Not the default. If you find yourself putting >1 bead in a row here,
   reconsider — they probably belong in a shared batch.

   Beads grouped this way need **no** collision edge and get their own
   assignee, so they reach the ready front together and dispatch concurrently.

**No batch gets a merge bead.** The coordinator merges each dispatch's branch
as that dispatch returns; nothing schedules that, and a plan creates no beads.

Tier is not a batching decision of yours: carry each bead's tier
(`metadata.execution_suggested_model`, else `metadata.tier`) into the plan's
node metadata as grooming assigned it, and leave the resolution and any
escalation to the orchestrator at dispatch.

## The dev pool is capped

**Never plan more concurrent dev assignees than the harness can run:
N_dev ≤ `concurrency_cap` − 1.** Ceremony agents are serialized by the graph
and share ONE slot between them, so count dev assignees only — the epic gate's
reference carries that rule and the `jq` that implements it. Breadth beyond
N_dev is more beads per dev, never more devs: a spawn past the cap is refused,
and the unspawned devs' queues stay ready, correctly assigned and permanently
undrained, with nothing reporting it. Read `concurrency_cap` off the manifest
rather than assuming a number; its source, its default and the plan-time check
are in `references/partitioning.md`.

## Assign every planned bead, then gate the epic

**The partition is what makes an assignee computable**, so assignment happens
here and not at the readiness gate. Per batch: `bd update <id> --assignee
dev-<run-id>-<batch-id>` on every bead it holds, the id derived in the
coordinator's dispatch-policy reference and never minted. Ceremony beads take
that same reference's roster ids. **The assignee is what groups a batch and
the collision edges are what order it.** There is no schedule to write down:
both live in the graph, which is re-queried rather than remembered.

Then run the epic's own gate, `references/shared/epic-definition-of-ready.md`
— a bead nobody owns and a dev pool over the concurrency cap, neither visible
to the per-bead gate. Report the bead or count each names. It runs **after**
the assign step: nothing earlier in a run writes an assignee.

## Return the plan

**One artifact** — an edge file, loaded with `bd dep add --file`. **You create
no beads**: the work beads were admitted before you were invoked, and
`sprint-scope` already built the ceremony epics and the target, so edges
between beads that already exist are the whole of what a plan contributes.
Emit it, diff it, validate it, return its path: `references/partitioning.md`.

Three rules that make the plan worth returning:

- **The plan names beads, and says which epic they came from.** The queue is
  `bd ready --parent <epic> --exclude-type=epic`, re-derivable from one durable
  id rather than from a filter nobody recorded.
- **The manifest is not yours.** It is `--metadata` on the run epic, holding
  the orchestrator's measurements. Read it; never rewrite it. If it is missing
  or malformed, say so in the return instead of supplying one.
- **Add an environmental-ledger entry only on evidence.** An entry goes to the
  orchestrator, which records it in the manifest. Every one names the command
  that proved it and the result observed; an untested claim is recorded as
  `unverified` with what would settle it, not omitted.

Do not ask for approval — the orchestrator prints the plan and proceeds, and
the user can interrupt.

Amendments after this point are the **orchestrator's** — an edge added, a bead
admitted or bounced, a tier escalated. Do not re-partition a run in flight, and
do not expect to be invoked twice.
