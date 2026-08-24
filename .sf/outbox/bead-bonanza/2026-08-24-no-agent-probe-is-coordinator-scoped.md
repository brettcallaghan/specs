---
title: "`pre-flight-check.sh`'s no-agent probe answers only for the coordinator"
description: State that `BB_ASSUME_NO_AGENT` is scoped to the session that ran the probe, and require each dispatching agent to re-probe in its own session
slug: issue-no-agent-probe-is-coordinator-scoped
audience: skill
target: bead-bonanza
priority: 2
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#9"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

`pre-flight-check.sh` already models a missing dispatch capability as
`BB_ASSUME_NO_AGENT=1` and blocks a run on it. It ran clean here, because it
ran in the **coordinator's** session, where the Agent tool is available — and
`sprint-wave`, dispatched as a subagent, had no such tool. The capability is
per-agent, not per-session. The cost is a run-level gate that reads green while
the condition it exists to catch is present two levels down, which is worse
than no gate: it licenses the run to proceed.

## Evidence

- `qozp#9` — "retro:tooling:sprint-wave — sprint-wave could not dispatch a subagent: no Task/Agent tool was exposed to it, so it created the worktree as orchestrator and ran sprint-batch -> run-bead itself inside it. The skill's whole premise is 'dispatches a subagent per batch', and it degraded silently into doing the work rather than stopping."

## Proposed change

State in the pre-flight's own output and in `references/scripts.md` that
`BB_ASSUME_NO_AGENT` is scoped to the session that ran the probe and does not
clear any subagent, and require each dispatching agent to re-probe in its own
session before it dispatches. Done: the pre-flight emits the scope caveat
alongside the result, and the reference no longer reads as a run-level
clearance.
