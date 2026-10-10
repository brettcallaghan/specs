---
title: "swarm dispatch accept refuses a dispatch branch with no commits beyond its base while that dispatch's beads read closed"
description: accept read an empty dispatch branch as already merged and deleted its worktree and branch, though the dispatch's bead was closed with its commit elsewhere.
slug: issue-accept-treats-an-empty-dispatch-branch-as-merged
audience: skill
target: swarm
priority: 1
updated_at: 2026-10-07T06:36:09Z

findings: ["c4s1.9#1"]
recurrence: 0
report: docs/retros/2026-10-07-c4s1-c4-diagrams-parser.md

run: "the sprint run recorded in report"
branch: sprint/<run-epic>
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "dev-<run-epic>-1 opus, dev-<run-epic>-2 sonnet, dev-<run-epic>-3 haiku, dev-<run-epic>-4 haiku; coordinator, plan, review and retro agents on the session default model; effort unrecorded"
---

## What is wrong, and what it costs

When a dispatch agent commits outside its worktree, its dispatch branch stays at the sprint base. `swarm dispatch accept` reads that empty branch as already merged, removes the worktree and deletes the branch — although the dispatch's bead reads closed, so a commit for it must exist somewhere. Cost: the run nearly lost the bead's commit; only the coordinator's post-merge re-check (`bb-status.sh bead`) noticed, and the commit had to be found and fast-forwarded by hand.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `c4s1.9#1` — "retro:tooling:swarm-dispatch The dispatch agent (light tier) committed <bead-id> in the main checkout (then on claude/project-thread-u8o164) instead of its worktree /home/user/.bb-worktrees/dev-<run-epic>-3, so bb/dev-<run-epic>-3 stayed empty; swarm dispatch accept read the empty branch as merged, removed the worktree and deleted the branch, and only the post-merge re-check caught it. Repaired by fast-forwarding sprint/<run-epic> to 36f09ee. The agent recorded no retro: comment either."

## Proposed change

Before removing a worktree or deleting a dispatch branch, `swarm dispatch accept` checks whether the branch has any commit beyond the base it was opened from; if it has none while any bead assigned to that dispatch is closed, accept refuses (non-zero exit) and names the closed beads whose commits are missing, leaving worktree and branch in place. Done when an empty dispatch branch with a closed bead makes accept exit non-zero with the worktree intact.
