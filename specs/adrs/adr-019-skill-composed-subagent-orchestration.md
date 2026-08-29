---
title: "ADR-019: Orchestrate Work Orders with a Skill Composing Native
  Subagents"
description: bead-bonanza composes the Agent primitive rather than
  adopting agent teams, background sessions, /batch or dynamic workflows,
  because the durability and ceremony this workload needs are not what
  those mechanisms provide
slug: adr-019
type: adr
status: draft
updated_at: 2026-08-09T02:30:00Z
---

# ADR-019: Orchestrate Work Orders with a Skill Composing Native Subagents

**Status:** draft
**Date:** 2026-08-09

## Context

The `bead-bonanza` skill processes a queue of work orders autonomously:
it triages, gates each for readiness, dispatches an agent per order,
runs the project's gates, and lands each order as exactly one commit
with its tracker close folded in. A run lasts hours and spans dozens of
agents.

Claude Code ships several mechanisms for running work in parallel —
subagents, agent teams, background sessions, `/batch`, and dynamic
workflows — so a hand-rolled orchestration skill needs justifying. It is
otherwise an obvious candidate for deletion by a future maintainer who
notices the overlap.

Four constraints shape the choice, and all four are properties of where
this actually runs rather than preferences:

1. **The harness is Claude Code on the web**, a remote container.
   Terminal-only surfaces do not exist there.
2. **The orchestrator's context will be compacted mid-run, whether or
   not it plans for it.** `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE=80` is set in
   this environment (verified in-session), and a compacted session
   re-injects an invoked skill's body capped at 5,000 tokens.
3. **The container is ephemeral and the tracker's database is
   gitignored**, so the committed `.beads/issues.jsonl` export is the
   only artifact carrying tracker state between sessions.
4. **Each work order must land atomically** — tracker close folded into
   the implementation commit, one commit per order — because a
   half-closed order is worse than an unstarted one.

A run in this repository exercised every one of these: the orchestrator
was interrupted and lost its context, sat dead for 2h16m, and three work
orders subsequently closed with no progress ledger and no retro anchor.
That run is the evidence base for this decision and for the RFCs filed
against the skill.

## Decision

We will implement orchestration as a **skill that composes Claude Code's
native subagent primitive**, using native mechanisms wherever they
exist — the `Agent` tool, the per-invocation `model` parameter for tier
routing, and `isolation: worktree` for filesystem isolation — and
hand-rolling only what has no native equivalent: the run directory, the
JSONL progress ledger, the per-order ceremony, the crash-recovery
procedure, and the circuit breaker.

We will not adopt agent teams, background sessions, `/batch`, or dynamic
workflows as the orchestration substrate.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Agent teams** | Shared task list with dependencies, file-locked claiming and auto-unblocking — the closest native fit for a work queue | Experimental and disabled by default; documented as having no session resumption with in-process teammates; task status can lag, blocking dependents; higher token cost (each teammate is a separate instance) | The one thing it adds over plain subagents is a dependency-aware task list, which `bd` already provides *durably across containers*. Adopting it means two trackers, and the less durable one holds the run state (constraint 3) |
| **Agent view / background sessions** | Real isolation per session; durable across the parent's lifetime | Sessions are local: they run on the user's machine | Unavailable in the harness this runs in (constraint 1) |
| **`/batch`** | Purpose-built for fanning one change across many worktree-isolated agents | Splits *one* change into 5–30 agents that each open a pull request | Inverse shape. This workload is many independent orders landing on one branch, one commit each — not one change becoming many PRs |
| **Dynamic workflows** | A script holds the loop and intermediate results, so the orchestrator's context stays small — literally this skill's design principle, implemented by the platform; the runtime tracks each agent's result to make a run resumable | No mid-run user input, so a decision gate cannot pause a run; resume works only within the same Claude Code session, and exiting starts the next run fresh; replay re-runs every agent that started after the first incomplete one | Resume is weaker than what the run directory plus `bd` already give (constraint 3), and replay is **not idempotent for an agent that commits** — a re-run order would land twice. The decision gate is load-bearing and cannot move inside the script. Availability on the web is undocumented and **untested** |
| **Orchestrator's own loop, no delegation** | No dispatch machinery at all; nothing to maintain | Every order's context accumulates in one window; compaction then destroys the procedure mid-run | Fails constraint 2 outright. The skill keeps this only as a *fallback* for harnesses that expose no `Agent` tool, not as the default |
| **Skill composing native subagents** (chosen) | Fresh context per order; native model routing and worktree isolation; durable on-disk run state the orchestrator can re-read after compaction; ceremony and recovery expressed as procedure rather than as a runtime that must survive | The skill is a large document that must itself survive compaction; the ledger and recovery path are ours to maintain; nothing enforces the ceremony except the procedure | N/A — selected |

## Consequences

**Positive:**

- Orchestration state lives on the run epic — its `--metadata` manifest and
  its `amend:` comments — so an orchestrator that loses its context recovers
  by reading the tracker rather than by replaying agents that already
  committed. (Amended: this was originally a run directory on disk, which did
  not survive a reclaimed container.)
- `bd` stays the single tracker. There is no second task list to reconcile,
  and work-order state survives container reclaim through the Dolt remote.
  (Amended: originally through a committed JSONL export, which is now
  untracked and is not the transport.)
- The per-order ceremony — red-before-green on defect orders, retro
  before close, atomic close folded into the implementation commit — is
  expressible because we own the dispatch step. No native mechanism has
  a hook at that point.
- Native primitives are used where they exist, so improvements to
  subagent dispatch, model routing and worktree isolation arrive for
  free.
- The approach degrades rather than fails: a harness with no `Agent`
  tool runs the same ceremony in the orchestrator's own loop.

**Negative:**

- **The skill must fit under the 5,000-token re-injection cap, and
  today it does not.** At 869 lines it truncates at line 406 on
  re-injection, dropping the dispatch rule, the retro-anchor rule and
  the published script interfaces — the material an orchestrator needs
  most after losing context. This is the standing cost of expressing
  orchestration as a document rather than as a runtime, and it is
  tracked as `agent-plugins-8ef`.
- The progress ledger, the run directory and the recovery procedure are
  ours to maintain and to keep correct as the harness changes. A native
  mechanism that later covers them would make this maintenance
  redundant.
- Nothing mechanically enforces the ceremony. A dispatch that skips it
  produces a work order that looks closed and has no recoverable
  record; detection has to be added deliberately (`agent-plugins-8ef`
  and the acceptance check proposed in the skill's RFCs).
- We forgo the workflow runtime's cost accounting and its progress view,
  and the agent-team task list's automatic unblocking. Where those
  matter, they are re-implemented or done without.
- The decision is coupled to harness behaviour that is undocumented in
  places. If dynamic workflows become available on the web *and* gain
  cross-session resume and a mid-run input gate, the strongest reason
  for the chosen option weakens and this ADR should be revisited.

## References

- [[adr-013]] — work order as the atomic unit this skill orchestrates
- [[adr-017]] — the plugin/prerequisite split that keeps skills pure
  markdown, which is why orchestration is expressed as a skill at all
- `plugin/skills/bead-bonanza/SKILL.md` — the skill this decision
  governs
- `rfcs/inbox/bead-bonanza/` — three RFCs from the run cited in Context,
  including the compaction-truncation finding behind the first Negative
- Claude Code documentation, read 2026-08-08:
  `code.claude.com/docs/en/agents`, `/sub-agents`, `/workflows`,
  `/agent-teams`, `/agent-view`, `/worktrees`, `/context-window`,
  `/claude-code-on-the-web`
