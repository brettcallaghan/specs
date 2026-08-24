---
name: sprint-plan
description: Groom a sprint's queue and partition it in one pass — bring every
  bead to Ready (invoking groom-beads for what fails, bouncing what grooming
  cannot clear) and return both the readiness verdict and the wave table of
  batches the run dispatches. Use when a sprint orchestrator is preparing a run
  epic's queue for dispatch, or when asked to gate beads for readiness,
  classify them as trivial / isolated / shared-surface, compute C4 co-edit and
  sequencing overlap, choose a dispatch mode per batch, or produce a sprint's
  wave table. Takes the run epic, resolves its queue from the graph, and
  dispatches nothing.
version: "0.2"
user-invocable: false
---

# sprint-plan

Take a run epic's queue from wherever it is to **groomed and partitioned**, in
one dispatch. Two returns: the readiness verdict per bead, and the wave table.

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
| Model lineup | Resolves a bead's tier (`metadata.execution_suggested_model`, else `metadata.tier`) to a model; leave the `Tier` column's model blank where the project declares none |
| Standing authorization | What the user authorized without further asking (push mode, PR, scope) |
| The grooming wave preview | The candidate partition, WHEN one was produced |

WHEN an input is missing, say so in the return rather than substituting a
guess — a plan built on an inferred queue is the failure the enumeration rule
exists to prevent.

**Out** — two things, and nothing else:

1. **The readiness verdict** — per bead: ready, or NOT-READY with the gap a
   human must close. Bounced beads carry the **`human`** label, so the verdict
   is queryable and not merely reported.
2. **The wave table** — plus **one line per override** of the grooming wave
   preview saying what you changed and why.

The orchestrator dispatches ready beads from the table and does **not** store
it. The table is derived from the ready front; a stored copy only drifts, and
recovery re-derives it rather than reading one.

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
bead is the single most expensive failure mode a run has: it detonates
mid-wave, after later beads have built on it.

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
of Ready and returns a per-bead readiness report plus a **wave preview** — its
candidate partition under dependencies and C4 collisions. **Its final act is to
invoke `sprint-plan`**, which turns that preview into the run's wave table:
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
  Definition of Ready. (A groomed prerequisite that still fails DoR check 9
  stops for the user, per the standing-authorization terms — and is raised into
  the human queue with the `human` label.) Grooming and wiring the edge are tracker writes, which need not wait
  for a wave boundary.
- **Out-of-scope** → leave it filed and linked, do NOT dispatch it, and surface
  it in the final report for the user to triage. The standing authorization
  finishes the launched scope; it never enlarges it.

**Ballooning is a stop signal.** If "prerequisites" multiply — a discovered
blocker that is really a new epic, or more than a handful of new in-scope beads
off a single bead — stop and surface it. That means the original scope was
mis-groomed, which is a human's call, not an autonomous expansion.

The grooming subagent prompt template is in `references/grooming-prompt.md`.

## Bias toward batching

**Batching is how a wave gets to run in parallel.** Beads that would collide
if dispatched side by side go behind ONE agent that takes them in order, which
frees the rest of the wave to dispatch concurrently around them. Left
un-batched they force the whole wave sequential. Your job is to find the
largest safe batch, not to default to one-bead-per-subagent for caution.
Sequential is the *exception*, not the rule.

**Batching does not buy fewer check runs.** Every bead runs the project's
declared checks before its own close whether batched or not — those are the
quality bar for one commit, sized to be affordable every time. A batch changes
who does the work and what can run alongside it, never what gets validated.

**Worktrees do not make batching redundant, because the collision is at the
merge.** A worktree isolates two agents from each other perfectly — they never
see each other's edits. Then both branches come back to the orchestrator, the
second one lands on a file the first already rewrote, and under `--ff-only` it
refuses and stops the wave. Batching is what removes that: the two beads go
behind one agent, in order, each branching from the last, so what returns is
already reconciled. Isolation is never grounds to relax the rules below.

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
`references/run-record.md`.

1. **C4 metadata — primary.** Read `.metadata.c4_changes` and
   `.metadata.c4_impacts` from `bd list --json` for every queued bead.
   - Two beads sharing a `c4_changes` component are **shared-surface** whatever
     their file paths say — the same component through different files still
     collides. They MAY batch; they SHALL NOT share a concurrent wave.
     **The hazard is a *merge* collision:** two branches changing the same
     component conflict when the second one merges, and under `--ff-only` that
     stops the wave. Worktrees isolate the agents but not the merge, so
     isolation is never grounds to relax this rule.
   - Bead B whose `c4_impacts` intersects bead A's `c4_changes` consumes a
     contract A is changing. That is a **sequencing** constraint, not a batching
     one, and it holds with zero file overlap. **It belongs in the graph:** a
     `blocks` edge is honoured by every re-query of the ready front, where an
     ordering noted only in this table is honoured once, by whoever reads the
     table. WHEN you find one grooming missed, say so in your return so the
     orchestrator can wire it — do not silently encode it as batch order.
   - Disjoint changes with no impact intersection is isolated *architecturally*
     only — layer 2 still runs.
2. **File paths — fallback and corroborator.** Scan each bead's description for
   fixture names + source paths. It is the whole signal where `c4_*` is absent,
   and it runs alongside C4 where both exist. WHEN the two disagree — disjoint
   components, overlapping files — take the **union** as the collision set and
   note the discrepancy in the plan; do not silently prefer one signal.
3. **The grooming wave preview — starting partition.** WHEN one was produced,
   start from it instead of re-deriving the partition, and record whether you
   accepted or overrode it and why. You are the source of truth on packing;
   the preview is an input, not an instruction.

**Dependency ORDER is not yours to compute.** The graph already holds it, and
the orchestrator reads it by re-querying `bd ready --parent <epic>
--exclude-type=epic` at every wave boundary. What you decide is **packing** —
which of the beads that are simultaneously ready can safely share a wave, and
which must be batched behind one agent because they would collide.

`bd swarm validate <epic>` gives a wave PREVIEW worth starting from, with two
conditions on trusting it, both measured:

- It scopes dependency analysis to the epic given. An edge to a bead outside
  that epic is reported as a warning and then ignored — it placed a blocked
  bead in Wave 1. **If it warns about an outside dependency, do not use its
  partition.**
- It is one level deep. On a nested run epic it lists sub-epics as work items
  and never sees their leaves. **If the run epic is nested, do not use its
  partition.**

In either case derive packing from the re-queried front instead. Say in your
return which you used.

Where the repo documents no C4 architecture, grooming skips the tagging and
`c4_*` is absent — file scanning is then the whole signal, and absent metadata
means *unknown*, never "no collisions".

The two collision classes read differently in the plan's `Reason` column: a
merge collision names what is shared (`same c4_changes: auth-service`); a
sequencing constraint names what is consumed and from whom (`consumes
auth-service, changed by demo-9h0`).

## Choose a dispatch mode per batch

Modes decide **how many beads share a subagent**, not how they are isolated —
isolation is settled above and is the same everywhere. Pick the broadest one
that's safe: you're maximizing what can run concurrently, not minimizing check
runs — those are per bead regardless.

**A batch subagent holds exactly one worktree at a time** — create, implement,
close, hand back, then the next. Holding two at once makes worktree presence
ambiguous and breaks the in-flight signal for every reader. Never pack a batch
in a way that assumes otherwise.

1. **Batch** — ONE subagent processes a cluster sequentially, one commit per
   bead. Per bead, before its close: any regeneration, **targeted tests
   (non-negotiable)**, then the project's declared checks. Two clusters qualify
   and they take the SAME prompt, differing only in the surface it names:

   - `shared-surface` beads on one fixture/code area — that area is the
     surface, and the targeted tests are the ones covering it.
   - `trivial` beads — surface "documentation and chores; no shared code
     surface". Cap ~6 per subagent.

   **There is no faster lane.** Both run the full declared set before every
   close — the dominant cost — so a trivial cluster is batched to stop it
   taking a dispatch slot per bead, not because it is cheaper.

   **Threshold:** batch when **≥2** beads share a surface and decisions are
   independent (bead A's choice doesn't change bead B's implementation). Even
   at 2 the gain is real: the pair stops blocking the wave, and the second bead
   sees the first's work instead of racing it.

   **Don't over-split:** if 7 beads cluster on one surface and only 1–2 of them
   are "larger" (substantial new features), they can still go in one shared
   batch. Splitting buys you nothing if decisions don't conflict. Only split
   when bead A's implementation would conflict with bead B's, or when decisions
   overlap such that one bead's choice changes another bead's plan.

2. **Concurrent wave** — `isolated` beads dispatched at the same time rather
   than one after another, each in its own worktree. Every bead already runs in
   a worktree, so this buys wall-clock and nothing else. A wave is the unit of
   parallelism and waves stay sequential, so put **no two beads sharing a
   `c4_changes` component in one wave** — the orchestrator re-verifies that and
   the base ref before it dispatches, and a wave that cannot pass the check is a
   planning error, not a dispatch-time surprise.

4. **Plain sequential** — reserved for: a single bead with no batch peers, OR a
   bead whose surface diverges from every other queued bead, OR explicit user
   override. Not the default. If you find yourself defaulting here for >1 bead
   in a row, reconsider — they probably belong in a shared batch.

Tier is not a batching decision of yours: carry each bead's tier
(`metadata.execution_suggested_model`, else `metadata.tier`) into
the wave table's `Tier` column as grooming assigned it, and leave the
resolution and any escalation to the orchestrator at dispatch.

## Return the plan

**The wave table is not written to disk.** It is derivable from the graph, so a
stored copy only drifts out of agreement with the beads it describes. Return
it; the orchestrator prints it and dispatches from the re-queried ready front.
The table's format — including the `Reason` column's required shape — is in
`references/run-record.md`.

Three rules that make the table worth returning:

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

Do not ask for approval — the orchestrator prints the table and proceeds, and
the user can interrupt.

Amendments after this point are the **orchestrator's** — a wave resequenced, a
bead admitted or bounced, a tier escalated. Do not re-partition a run in
flight, and do not expect to be invoked twice.
