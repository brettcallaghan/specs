---
title: "No dispatch tool is a hard blocker, not a degradation"
description: Probe for a dispatch tool before the first dispatch and refuse to run without one, instead of silently absorbing the work and self-verifying it
slug: issue-bead-bonanza-no-dispatch-tool-is-a-hard-blocker
audience: skill
target: bead-bonanza
priority: 0
updated_at: 2026-08-24T15:20:00Z

# Evidence — what this issue rests on
findings: ["qozp#9", "qozp#8", "qozp#11"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "0.1"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

A skill whose premise is "dispatches a subagent per batch" was invoked with no
dispatch tool exposed to it, and instead of stopping it created the worktree
itself and ran `sprint-batch` → `run-bead` inline. It degraded silently. The
cost is the verification boundary: the check that exists to test an
*independent* agent's account became self-verification, and nothing outside
the run can tell that happened.

`pre-flight-check.sh` already models the condition as `BB_ASSUME_NO_AGENT=1`
and blocks a run on it — and it ran clean, because it ran in the
**coordinator's** session, where the Agent tool is available. The capability is
per-agent, not per-session, so a run-level probe answered for the wrong agent.
A gate that reads green while the condition it exists to catch is present one
level down is worse than no gate: it licenses the run to proceed.

## Evidence

- `qozp#9` — "retro:tooling:sprint-wave — sprint-wave could not dispatch a subagent: no Task/Agent tool was exposed to it, so it created the worktree as orchestrator and ran sprint-batch -> run-bead itself inside it. The skill's whole premise is 'dispatches a subagent per batch', and it degraded silently into doing the work rather than stopping."
- `qozp#8` — "retro:process:wave-1-bypassed-sprint-wave — wave 1's four dispatches were made by the coordinator directly (harness Agent calls with isolation:worktree, branches worktree-agent-<harness-id>) rather than by invoking the sprint-wave skill as SKILL.md prescribes."

## Proposed change

**Supersedes two earlier issues.** This absorbs the `sprint-wave` dispatch
probe (that skill is deleted by the wave refactor, so a fix targeting it would
land nowhere) and the separate `pre-flight` scope-caveat issue. One probe,
stated once, at the point where the coordinator dispatches.

Probe for a dispatch tool in the coordinator's own session, before the first
dispatch. **With none, the run does not start** — this is a blocker with a
named remedy, in pre-flight's blocker vocabulary: *this skill orchestrates
subagents and this harness exposes no way to spawn one; run it in a harness
that does.* Not a warning, not a degraded mode, not a fallback to inline work:
a skill that cannot dispatch cannot do the thing it is for, and inline
execution silently removes every verification boundary the design rests on.

Where a ceremony agent is itself given work to dispatch, it re-probes in its
own session and refuses on the same terms, because the capability does not
travel from the session that measured it.

Done: `bead-bonanza` names the probe as a pre-flight blocker with its remedy;
a coordinator with no dispatch tool halts before creating a worktree, and no
agent anywhere in the run performs a dispatch's work inline.
