# Dispatch policy — identity, concurrency, tier

**The circuit breaker stops the RUN, not just the batch.** Three consecutive
blocked or failed beads is systemic — a broken base, a wrong assumption in the
plan, an environment fault — and the next dispatch meets it too. Stopping one
batch alone hands the same failure to a fresh set of agents.

Read this before you dispatch. `dispatching.md` is the procedure; this file is
the policy it runs under.

## Contents

- Agent identity — the roster of agent ids, how each one is derived, and the
  `BEADS_ACTOR` export every dispatch prompt carries
- Concurrent dispatch — the entry conditions that must ALL hold before two
  beads go out at once, and what to do when one does not
- Tier selection per bead — resolving a bead's tier to a model, the escalation
  rule, and why a downgrade never happens mid-run

### Agent identity — the roster, and `BEADS_ACTOR`

**This section is the roster's only home.** Every dispatch prompt and every
skill that mints or reads an agent id cites it here rather than restating the
table; a second copy is what drifts.

**One string is an agent's whole identity** — the bead's `assignee` and the
agent's `$BEADS_ACTOR` are the same value, and there is no third place it is
recorded. Nothing is written at dispatch time to make that true, so nothing can
be forgotten or go stale.
Beads keeps no agent registry (an assignee is a plain string), so nothing
reconciles two spellings of the same agent, and one string is what makes
`bd ready --assignee <id>` that agent's queue.

**Every agent and subagent exports `BEADS_ACTOR=<its id>` before its first
`bd` call.** Not because it intends to claim, but because a claim may happen:
`run-batch` claims, a recovery path claims, a retry claims.

One kind per agent the coordinator spawns:

| agent | id | count | concurrent? |
|---|---|---|---|
| coordinator | `coordinator-<run-id>` | 1 | — |
| groom | `groom-<run-id>` | 1 | no |
| plan | `plan-<run-id>` | 1 | no |
| dev | `dev-<run-id>-<batch-id>` | N | **yes** |
| review | `review-<run-id>` | 1 | no |
| retro | `retro-<run-id>` | 1 | no |

**Only the `dev` pool is concurrent.** Every other kind runs alone, in a chain
the dependency graph already serializes, so they cost one slot between them
rather than one each — which is why a concurrency check counts *concurrent*
assignees rather than distinct ones. Counting distinct assignees
under-subscribes the dev pool by the number of ceremony roles.

**Ceremony agents are assigned beads like any other agent.** That is what makes
`bd ready --assignee groom-<run-id>` the groom agent's queue and keeps one
dispatch mechanism for the whole run. Each ceremony EPIC is assigned to its
phase's id at creation, and the phase's agent exports that same id before its
first `bd` call.

**Four phases have no row of their own, and it is not an omission.**
Pre-flight, scope, merging and post-flight are the coordinator's own steps —
the first and last are scripts it runs, merging is something it does directly,
and scope MINTS the run epic, so no `<run-id>` exists to derive an id from
until it returns. All four act as `coordinator-<run-id>`, which is also what
makes a resumed coordinator able to recompute the authorization gate's holder:
scope takes that gate, and a hold under an id the coordinator cannot recompute
is a hold nobody can release.
`sprint-scope` therefore exports its actor **after** it creates the run epic
and before any write that records a holder.

#### Ids are DERIVED from the run epic id, never minted

`<run-id>` is the run epic id with the project prefix dropped — `acme-r7k`
becomes `r7k` — and `<batch-id>` is the batch's position in the **run**,
numbered once across the whole run: the first batches take 1..n and batches
planned later continue from n+1.

**Never restarted.** Two batches numbered 1 in one run resolve to the same
`dev-<run-id>-1` — two agents sharing an id, and the recomputation below then
returns an ambiguous answer.

Both follow from the epic id and the partition, so a **resumed session
recomputes the same string** and finds its own queue. A randomly minted id
cannot be recomputed, and a resumed session loses it.

**The `<run-id>` segment is not optional.** Batch ids restart every run while
the tracker is shared, so a bare `dev-3` collides across runs with nothing
reporting it.

#### Why the export matters

Measured on `bd` 1.2.2, claiming with `bd update <id> --claim`:

| bead's `assignee` | `$BEADS_ACTOR` | result |
|---|---|---|
| the agent's id | the same id | assignee unchanged; status `in_progress` |
| the agent's id | unset | refused, exit 1: `issue already claimed by <id>` |
| unset | unset | assignee **silently becomes** the harness default — git `user.name` — exit 0 |

The last row is the failure the export prevents: nothing errors, and the
coordinator's ownership view decays. `bd ready --claim` writes the actor into
`assignee` the same way, and on an already-assigned ready set it claims
nothing at all and returns `[]`.

An unassigned bead is therefore the only bead a claim can move to the wrong
owner, which is why assigning every bead before dispatch and exporting
`BEADS_ACTOR` everywhere are belt and braces rather than alternatives.

### Concurrent dispatch

**Concurrent dispatch is ENABLED.** Several beads may go out of the ready set
at once, each in its own worktree, merged as each returns.

**The ready set IS what may run concurrently.** Re-query
`bd ready --parent <epic> --exclude-type=epic` before every dispatch; what
comes back is what dependencies and collision edges now allow. Never advance by
walking a stored list — a bead that closed early, a bead blocked by something
discovered mid-run, and a bead admitted since the plan was made all make a
stored position wrong, and the graph is right about all three.

Entry conditions, both required and both about **safety**: **no shared
`c4_changes` component** between any two beads dispatched at once — the
collision edges should already guarantee it, and you check rather than assume —
and the base-ref check re-verified. Both are yours at dispatch, whatever the
plan says.

**There is no minimum bead count, and do not reintroduce one** — a count in a
list of safety conditions reads as one, and nothing about two concurrent beads
is less safe than three. Whether concurrency is *worth* it is ordinary
judgment; whether it is *safe* is the entry conditions above.

### Tier selection per bead

Beads carry a **capability tier** from grooming — the readiness gate's
gate guarantees one (grooming assigns a default, never null), stored as
`metadata.execution_suggested_model`, falling back to `metadata.tier`. Both
are written by grooming; the first is `bd`'s own documented convention
(`docs/METADATA.md`), which is what makes a groomed bead routable by an
orchestrator that never read this skill, and the second is what 205 of this
tracker's existing beads carry. **Read the documented key first and fall back**
— that is the whole migration, and no bulk rewrite is required. Despite its
name, `execution_suggested_model` holds a TIER: the convention defines it as
"a capability-tier suggestion", to be substituted for a same-tier model on a
different provider. The tiers are ordered, least to most capable:

```
light  →  standard  →  advanced  →  frontier
```

Beads name a **tier**, not a model, so a backlog groomed months ago still
routes correctly. Resolve tier → model from the project's `## model
lineup` block in `sf.toml`, echoed by pre-flight. Honor these:

- **Collect during grooming** — you're already reading each bead's
  readiness signals; record its tier then. For a bulk query, read the
  metadata and **pass `--status all`**: `bd list` defaults to open issues, so
  a filter written without it silently under-counts a groomed backlog.

  ```sh
  bd list --status all --json | jq '[.[] | select(.metadata.tier == "standard")] | length'
  ```
- **Dispatch at the resolved model** — pass `model: "<resolved>"` on the
  `Agent` tool call for that bead's subagent. Absent a lineup, omit the
  param (the subagent inherits the session model) and say so once.
- **A named `model` on a bead is an override** — grooming sets it only
  when a specific model is genuinely required. It wins over the tier.
- **Main-loop beads** — a bead at the orchestrator's own tier with a
  note saying "do not delegate" is implemented by the orchestrator
  directly, not dispatched. It still follows the full per-bead ceremony
  (claim, gates, retro, atomic close, single commit).
- **Batching across tiers** — prefer batching beads that share a tier.
  When a batch mixes tiers, run the whole batch subagent at the
  **highest** tier among its members — never run a bead below its tier.
  Split the batch WHEN more than one bead would be dragged above its
  tier, OR when any bead would be dragged more than one tier. Weigh that
  against what batching saves: one dispatch instead of N, one worktree,
  and — for a shared surface — one agent holding that surface's context
  across every bead touching it. It does **not** save check runs; every
  bead runs the project's checks before its own close either way.
- **Escalate, never downgrade** — if a bead fails gates/review twice or
  its subagent visibly flails, re-dispatch that bead one tier up. Never
  re-dispatch below the assigned tier, and never "save cost" by
  downgrading mid-run.
- **Record every escalation** — the bead, the assigned tier, the tier
  that completed it, and what failed. This is a required line in the
  Phase 4 report (see "Tier escalations"), and it is the run's only
  direct evidence about whether grooming's routing is calibrated. WHEN
  `CLAUDE_CODE_SUBAGENT_MODEL` is set, report it as Phase 4's "Tier
  escalations" section directs.
- **Record the routing** — every planned bead keeps the tier grooming assigned
  it, in its own metadata; nothing re-writes it at dispatch.
