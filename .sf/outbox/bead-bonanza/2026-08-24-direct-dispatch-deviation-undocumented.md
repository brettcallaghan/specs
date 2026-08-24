---
title: "Direct dispatch instead of `sprint-wave` is undocumented and costly"
description: Name direct dispatch as a deviation in the wave step, state its two costs, and require the harness-id to dispatch-id mapping be recorded at dispatch time
slug: issue-direct-dispatch-deviation-undocumented
audience: skill
target: bead-bonanza
priority: 3
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#8"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

Wave 1's four dispatches were made by the coordinator directly, with harness
Agent calls, rather than by invoking `sprint-wave`. Two costs fell out and
nothing in `bead-bonanza` warns of either: worktrees and branches were named
`worktree-agent-<harness-id>` rather than by dispatch id, so "presence is the
in-flight signal" still worked but "named for its dispatch id" did not, and the
coordinator had to map harness ids to bead ids by hand; and `sprint-wave`'s
verification was performed by the coordinator, which is precisely the reasoning
the coordinator is designed not to hold. The run carries both paths, so the
comparison is measured rather than argued.

## Evidence

- `qozp#8` — "retro:process:wave-1-bypassed-sprint-wave — wave 1's four dispatches were made by the coordinator directly (harness Agent calls with isolation:worktree, branches worktree-agent-<harness-id>) rather than by invoking the sprint-wave skill as SKILL.md prescribes."

## Proposed change

In `bead-bonanza`'s wave step, state that a wave is dispatched by invoking
`sprint-wave` and that direct dispatch is a deviation, naming its two costs;
and give the deviation a fallback — where worktree names come from the harness,
record the harness-id → dispatch-id mapping in the dispatch record at dispatch
time, not by hand afterwards. Done: the wave step names the deviation and the
mapping requirement.
