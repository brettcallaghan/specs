---
title: "swarm manifest counts review_budget in loops, as the authorization gate states it"
description: The gate authorizes "2 loops (three passes total)" but the manifest validator refuses review_passes=3 against review_budget=2, so the authorized third pass cannot be recorded.
slug: issue-review-budget-counts-passes-not-loops
audience: skill
target: swarm
priority: 2
updated_at: 2026-10-08T22:04:20Z

findings: ["qqnx#5"]
recurrence: 0
report: docs/retros/2026-10-08-qqnx-spec-cli-qa-round-2-fixes.md

run: specs-qqnx
branch: sprint/specs-qqnx
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "<withheld by session policy>"
---

## What is wrong, and what it costs

Two parts of swarm read the same `review_budget` number in two different units. `swarm-scope` puts the human the row "**Review loop budget** | `review_budget` | 2 loops (three passes total)". The manifest validator compares `review_passes` with `review_budget` directly and refuses `review_passes` greater than the budget. AC-SWARM-CORE-002.7 pins that reading: "Given a manifest recording two spent passes against a budget of two / When a lower count or a third pass is written / Then the command exits non-zero". So a run authorized for three passes cannot record its third. Its manifest then under-reports the review it actually ran. A resumed run reading `review_passes` would believe a pass is still owed.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `qqnx#5` — "retro:tooling:swarm review_budget semantics: authorized as '2 loops = 3 passes' and recorded review_budget=2, but swarm manifest set refuses review_passes=3 ('would be 3 against a budget of 2'), i.e. it treats the budget as total passes. Pass 3 ran anyway as authorized (read-only review); no fix loop follows it. evidence: swarm manifest set --run specs-qqnx --key review_passes --value 3 → exit 1."

Measured by this retro, reading swarm source at `1cea7e4`. `internal/manifest/validate.go` `checkCounter` reports `"review_passes" would be %d against a budget of %d` when `next > budget`. `plugin/skills/swarm-scope/SKILL.md` carries the "2 loops (three passes total)" row. `swarm manifest get --run specs-qqnx` reads `review_budget: 2`, `review_passes: 2`.

## Proposed change

The validator allows `review_passes` up to `review_budget + 1`, because the budget counts fix loops and the first pass is not a loop. AC-SWARM-CORE-002.7 changes to "Given a manifest recording three spent passes against a budget of two / When a lower count or a fourth pass is written". Done when `swarm manifest set --key review_passes --value 3` succeeds against `review_budget: 2`, and a value of 4 is refused.
