# bead-bonanza Operating Principles

Principles referenced from SKILL.md's Operating principles list, in full.
Read the summary line in SKILL.md first; come here for the rule's complete
statement.

## Contents

**Scoped to the coordinator.** A principle the per-bead ceremony enforces is
stated where that ceremony lives, not here — one commit per bead, and where an
interrupted bead's uncommitted work stays, both belong to the **`run-bead`**
skill and are enforced by the close script.

- Findings harvest — how to wire `retro:` capture; the harvest pipeline
  itself lives in the vendored findings-capture protocol, not here
- Guard against foreign dirt — how the orchestrator stages a commit, when a
  tracker write may happen, and whose claims it may reset
- Principles the coordinator applies but does not carry — the ones SKILL.md
  summarises only where they change a turn's routing

## Findings harvest

Wiring findings capture into a run and harvesting what it wrote are two
different steps. The harvest is the one that's easy to get wrong: `bd list`
defaults to open issues, and a drained closure set is almost entirely
closed, so a literal reading returns findings for the handful of beads that
never got closed and silently drops nearly everything else.

### Wiring

Every dispatch prompt — sequential, batch and fast-lane alike — instructs the
agent to record `retro:`-prefixed `bd comment`s on the bead it is working:
one per friction point, 1-3 sentences, friction only, zero analysis.

This is unconditional. It is not gated on the project having adopted anything
and there is no file to check for first — gate it and a project that carries
no retro tooling silently produces no findings at all, which is the one
outcome the channel exists to prevent. The comment channel IS the tooling.

It is the same channel `retro-spec`, `groom-feature-blueprint` and the
work-order template use — the findings-capture protocol, vendored at
`shared/findings-capture-protocol.md`. Keep it wired regardless
of how the harvest is done; do not remove it anywhere.

### Harvesting

The harvest pipeline (`bd export | jq`, why not `bd list` + `bd show --json`,
and the non-zero-count assertion) is the aggregation section of
`shared/findings-capture-protocol.md` — read it there rather than
here. This skill restates none of it.

## Guard against foreign dirt

### Working-tree dirt is the orchestrator's hazard

A subagent does not face it — inside its own worktree, "everything here" is
exactly its bead's work. **The orchestrator does face it.** It works in the
shared checkout, where a
diagnosis agent's uncommitted fixes, a regenerated fixture, or a half-applied
edit can be sitting alongside its own work. It therefore commits by explicit
pathspec (`git commit -- <files>`). The same applies on the main-loop fallback
path (SKILL.md), where the orchestrator implements every bead in the shared
checkout because there are no subagents to isolate.

### Tracker writes are never deferred

The tracker is a database synced to `origin refs/dolt/data`, not a file in the
working tree, and the close ceremony stages code only. No tracker write can
ride into another agent's commit, so **the orchestrator may groom a discovered
bead, bake a decision, or wire a dependency edge at any time, including
mid-wave.**

Worktrees are not what makes this safe, and must not be relied on for it: every
worktree resolves the same database through git's common directory — precisely
why `bd` works inside one.

### The claim-reset rule

The same "don't touch what another agent owns" principle applies to
tracker *claims*. Claiming is atomic — `bd update <id> --claim` sets status
and assignee in one write and is idempotent when the claim is already yours —
so the rule no longer has to cover the race between reading a status and
setting it. What it still covers is ownership: a bead claimed by a **different
assignee** belongs to a live sibling, and a subagent may not take it. The
orchestrator, never a subagent, owns reconciling claims across waves.

## Principles the coordinator applies but does not carry

Summarised in SKILL.md only where they change a turn's routing; stated here in
full.

- **Surface implicit decisions, don't bury them.** A bead with an open decision
  is bounced at the readiness gate. A choice a subagent still makes silently —
  one the bead's Design didn't pin — belongs in its `retro:` comment, not the
  diff.
- **Capture decisions in the bead, not in chat.** The bead description is the
  durable record; a session is not.
- **Fail loudly, recover quietly.** A blocked subagent is logged and the run
  moves on; the circuit breaker is what stops it. A crashed one follows
  `recovery.md` — the worktree and the bead are the anchors.
- **Route by tier, and report every escalation.** A bead's tier is grooming's
  *prediction* of the capability the work needs. Assign nothing below it;
  escalate one tier on repeated failure; never downgrade to save cost mid-run.
  Every escalation is a required row in the report — it is the only routing
  evidence a run produces. Full rules: `dispatch-policy.md`.
- **Project rules win over skill defaults.** `sf.toml` is the source of truth
  for what runs per bead. Follow the project's commit conventions. The skill is
  opinionated, not authoritative.
- **Never re-diagnose an environmental failure per wave.** Pre-flight captured
  the baseline once and every subagent prompt carries it as fact. The dangerous
  case is a check that silently PASSES without exercising the change — a stale
  prebuilt binary, a skipped suite, a cached result — so where a check *can*
  pass without touching the change, force a clean rebuild before trusting it.
- **Tracker writes are never deferred.** Groom a discovered bead, bake a
  decision, or wire a dependency edge at any time, mid-wave included. Only the
  working tree needs care; see "Guard against foreign dirt" above.
