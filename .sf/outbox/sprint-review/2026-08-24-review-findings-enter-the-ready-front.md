---
title: "`sprint-review` files findings into the ready front the gate just refused"
description: File each review finding deferred in section 3 and undefer only what section 4's classification puts in scope
slug: issue-review-findings-enter-the-ready-front
audience: skill
target: sprint-review
priority: 0
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#9 (context)"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "0.1"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

Section 3 files every finding with `bd create … --parent {EPIC}`, which lands
it **open**. Section 4 then classifies it, and the conservative default is out
of scope. But an open bead parented to the run epic is in
`bd ready --parent <epic>`, so all 11 of this run's review findings entered the
run's ready front on creation — including a P0 — and would have been silently
admitted at the next wave boundary, which is exactly what the gate had refused.
The coordinator caught it and deferred all 11 by hand with
`grooming_status=out-of-scope-this-run`. The cost of missing it is a run that
grows past its authorized scope through a side channel, with the gate reading
as honoured.

## Evidence

- `qozp#9 (context)` — "retro:tooling:sprint-wave — sprint-wave could not dispatch a subagent: no Task/Agent tool was exposed to it, so it created the worktree as orchestrator and ran sprint-batch -> run-bead itself inside it."

*Evidence note: this rests on the coordinator's `incident:` comment on the
epic, which neither downstream harvest reaches. It is quoted in full in the
report's §1 and corroborated by the tracker state — `qozp.18` through `.28` all
carry `deferred` and `grooming_status=out-of-scope-this-run`.*

## Proposed change

In section 3, file each finding **deferred**; in section 4, undefer only what
classification puts in scope. Filing stays free and losing a finding stays
impossible, but a finding no longer enters the run by default. Done: findings
created by `sprint-review` do not appear in `bd ready --parent {EPIC}` until
section 4 admits them.
