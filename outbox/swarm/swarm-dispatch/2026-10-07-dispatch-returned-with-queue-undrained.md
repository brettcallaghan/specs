---
title: "swarm-dispatch: a dispatch returns only when its queue is drained"
description: Require a drained queue before a dispatch hands back
slug: issue-dispatch-returned-with-queue-undrained
audience: skill
target: swarm-dispatch
priority: 2
updated_at: 2026-10-07T06:05:00Z
findings: []
recurrence: 0
report: docs/retros/2026-10-07-sra-c4-prerequisites-record-blocks.md
run: specs-sra
branch: claude/project-thread-es7x8o
skill_version: swarm plugin as installed 2026-10-07 (swarm CLI 1cea7e43)
model: dispatches resolved through manifest.models (standard->sonnet, light->haiku); served model and effort unrecorded
---

**What is wrong, and what it costs.** `dev-sra-1` returned after one of its
three beads; the coordinator had to notice and resume it, costing a round trip
on the critical path of the gate beads.

**Evidence.** Coordinator-reported; no durable transcript (see could-not-verify).
Commits: `e602237` (`f6tl.5`) 05:10, then `07fa867` / `4638d28` after resume.

**Proposed change.** `swarm-dispatch`'s hand-back step requires `bd list
--assignee <dispatch-id> --status open,in_progress` to be empty (or each
remaining bead blocked/skipped with its reason) before the agent returns.
