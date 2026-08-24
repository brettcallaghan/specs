---
title: "`bb-status.sh`'s human-queue check is repo-wide, not run-scoped"
description: Scope the human-queue check to the run's admitted set, the way the neighbouring checks already are
slug: issue-bb-status-human-queue-check-is-repo-wide
audience: skill
target: bead-bonanza
priority: 2
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#5"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

The check runs bare `bd human list` and failed the run's health with "human
queue: 1 item(s) awaiting a decision" for `agent-plugins-qozp.12`, a bead that
is **deferred and explicitly excluded at admission**. Every other check in the
same report scopes correctly — "bd lint: nothing flagged in this run's queue",
"every ready bead carries a tier". The cost is that a run can never read clean
while any deferred decision exists anywhere in the tracker, which trains the
operator to ignore the signal.

## Evidence

- `qozp#5` — "retro:tooling:bead-bonanza — bb-status.sh's human-queue check is not scoped to the run's admitted set. It reported 'FAIL human queue: 1 item(s) awaiting a decision' for agent-plugins-qozp.12 (DECISION D1), which is DEFERRED and explicitly excluded at admission."

## Proposed change

Scope the human-queue check to the run's admitted set, the way the neighbouring
checks already are. Done: a human-needed item outside `manifest.admitted` does
not fail the run's health.
