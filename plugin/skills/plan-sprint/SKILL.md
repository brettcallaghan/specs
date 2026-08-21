---
name: plan-sprint
description: Partition a queue of groomed beads into the waves and buckets a
  bead-bonanza run executes. Use
  when a bead-bonanza orchestrator delegates Phase 2.5 planning — typically as
  the final act of the Phase 1.5 grooming subagent, which has already read
  every bead — or when asked to classify queued beads as trivial / isolated /
  shared-surface, compute C4 co-edit and sequencing overlap, choose a dispatch
  mode per bucket, or produce a bead-bonanza run's wave table. Takes the run
  epic, resolves its queue from the graph, and returns the wave table.
version: "0.1"
user-invocable: false
---

# plan-sprint

Turn a run epic's queue of **groomed** beads into a wave table, and return it
and returns it. This is read-heavy analytic work over every
queued bead — that is why it runs in a subagent rather than in the
orchestrator, which stays small.

You are planning, not dispatching and not implementing. Nothing here creates a
worktree, edits a bead's code, or spawns an agent.

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

**Out** — one thing, and nothing else:

**To the caller:** the wave table, plus **one line per override** of
   the grooming wave preview saying what you changed and why. Keep it to that;
   the orchestrator prints the table and dispatches from it.

## Bias toward batching

**Batching is how a wave gets to run in parallel.** Beads that would collide
if dispatched side by side go behind ONE agent that takes them in order, which
frees the rest of the wave to dispatch concurrently around them. Left
un-batched they force the whole wave sequential. Your job is to find the
largest safe batch, not to default to one-bead-per-subagent for caution.
Sequential is the *exception*, not the rule.

**Batching does not buy fewer check runs.** Every bead runs the project's
checks script before its own close whether batched or not — that script is the
quality bar for one commit, sized to be affordable every time. A batch changes
who does the work and what can run alongside it, never what gets validated.

**Worktrees do not make batching redundant.** A worktree isolates a bead; it
does not make two beads that rewrite the same file safe to run at once. That is
the collision batching removes, and it is why the batching rules below hold
even though every bead already has a worktree of its own.

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
     orchestrator can wire it — do not silently encode it as bucket order.
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

## Choose a dispatch mode per bucket

Modes decide **how many beads share a subagent**, not how they are isolated —
isolation is settled above and is the same everywhere. Pick the broadest one
that's safe: you're maximizing what can run concurrently, not minimizing check
runs — those are per bead regardless.

**A batch subagent holds exactly one worktree at a time** — create, implement,
close, hand back, then the next. Holding two at once makes worktree presence
ambiguous and breaks the in-flight signal for every reader. Never pack a bucket
in a way that assumes otherwise.

1. **Doc/chore fast lane** — `trivial` beads. ONE subagent processes the cluster
   sequentially. Per bead: targeted tests plus the project's checks script,
   before that bead's close. One commit per bead. Cap ~6 beads per fast-lane
   subagent.

2. **Shared-surface batch** — `shared-surface` beads that cluster on the same
   fixture/code area. ONE subagent processes the cluster sequentially. Per bead,
   before its close: any regeneration, **surface-relevant targeted tests
   (non-negotiable)**, then the project's checks script. One commit per bead.

   **Threshold:** use shared-surface batch when **≥2** beads share a surface and
   decisions are independent (bead A's choice doesn't change bead B's
   implementation). Even at 2 the gain is real: the pair stops blocking the
   wave, and the second bead sees the first's work instead of racing it.

   **Don't over-split:** if 7 beads cluster on one surface and only 1–2 of them
   are "larger" (substantial new features), they can still go in one shared
   batch. Splitting buys you nothing if decisions don't conflict. Only split
   when bead A's implementation would conflict with bead B's, or when decisions
   overlap such that one bead's choice changes another bead's plan.

3. **Concurrent wave** — `isolated` beads dispatched at the same time rather
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

Tier is not a bucketing decision of yours: carry each bead's tier
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
