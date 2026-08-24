# Dispatch policy — concurrency, tier, push

**The circuit breaker stops the RUN, not just the wave.** Three consecutive
blocked or failed beads is systemic — a broken base, a wrong assumption in the
plan, an environment fault — and the next wave meets it too. Stopping the wave
alone hands the same failure to a fresh set of agents.

Read this at the wave boundary, when you are deciding how a wave goes out.
`sprint-wave` executes the wave; this file is the policy it executes under.

## Contents

- Concurrent dispatch — the entry conditions that must ALL hold before two
  beads go out at once, and what to do when one does not
- Tier selection per bead — resolving a bead's tier to a model, the escalation
  rule, and why a downgrade never happens mid-run
- Push policy — `defer_push` versus `push_per_bead`, and who pushes

### Concurrent dispatch

**Concurrent dispatch is ENABLED.** Several beads may be dispatched at once
within a wave, each in its own worktree, merged as each returns.

Note what is and is not concurrent. **A wave is the unit of parallelism**, and
waves remain sequential — a wave boundary is where dependency edges resolve and
where the base ref advances. Concurrency is *within* a wave, across beads that
the dispatch plan established are independent.

**A wave is the current ready front, packed.** Re-query
`bd ready --parent <epic> --exclude-type=epic` at every boundary; what comes
back is what dependencies now allow, and the plan says only how to pack it.
Never advance by walking a stored wave list — a bead that closed early, a bead
blocked by something discovered mid-run, and a bead admitted since the plan was
made all make a stored position wrong, and the graph is right about all three.

Entry conditions, both required and both about **safety**: **no shared
`c4_changes` component** between any two beads in the wave, and the base-ref
check re-verified for this wave. Both are yours to check at dispatch, whatever
the plan says.

**There is no minimum bead count, and do not reintroduce one** — a count in a
list of safety conditions reads as one, and nothing about two concurrent beads
is less safe than three. Whether a wave is *worth* dispatching concurrently is
ordinary judgment; whether it is *safe* is the entry conditions above.

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
- **Record the routing** — the dispatch plan's wave table carries a
  `Tier` column for every bead.

### Push policy

- **`push-per-bead`** — subagent pushes after each commit. Best for sequential
  mode (recovery is easier on a per-bead failure).
- **`defer-push`** — subagent commits per bead, orchestrator pushes once at
  batch end. Default for fast-lane and shared-surface batch (one process;
  failures are observable).

For a single-bead "wave," `push-per-bead` and `defer-push` are equivalent —
don't bother labeling. Just push when the subagent returns.
