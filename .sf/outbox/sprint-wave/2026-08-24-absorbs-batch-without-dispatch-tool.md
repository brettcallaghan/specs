---
title: "`sprint-wave` absorbs the batch when it has no dispatch tool"
description: Probe for a dispatch tool first and return a blocker when none exists, instead of silently doing the batch's work and self-verifying it
slug: issue-sprint-wave-absorbs-batch-without-dispatch-tool
audience: skill
target: sprint-wave
priority: 1
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#9", "qozp#8"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "0.1"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

`sprint-wave`'s whole premise is "dispatches a subagent per batch". In wave 2
no Task/Agent tool was exposed to it, and instead of stopping it created the
worktree itself and ran `sprint-batch` → `run-bead` inline. It degraded
silently. The cost is the skill's entire reason for existing: the verification
it performs against an independent agent's account became self-verification,
so wave 2's returns were checked by the agent that produced them. A run cannot
tell from the outside that this happened.

## Evidence

- `qozp#9` — "retro:tooling:sprint-wave — sprint-wave could not dispatch a subagent: no Task/Agent tool was exposed to it, so it created the worktree as orchestrator and ran sprint-batch -> run-bead itself inside it. The skill's whole premise is 'dispatches a subagent per batch', and it degraded silently into doing the work rather than stopping."
- `qozp#8` — "retro:process:wave-1-bypassed-sprint-wave — wave 1's four dispatches were made by the coordinator directly (harness Agent calls with isolation:worktree, branches worktree-agent-<harness-id>) rather than by invoking the sprint-wave skill as SKILL.md prescribes."

## Proposed change

Make a dispatch-tool probe `sprint-wave`'s first act, before it creates a
worktree or reads the batch. Where no dispatch tool is available, stop and
return a named blocker — the wave is undispatchable — and do not perform the
batch's work. Done: a `sprint-wave` invoked with no dispatch tool returns a
blocker and leaves the tracker and the working tree untouched.
