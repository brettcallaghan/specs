---
title: "Severity cannot override the scope gate, so a run ships its own regression"
description: Add a classification row putting a P0 defect introduced by the run's own merged diff in scope regardless of the scope-growth gate
slug: issue-severity-cannot-override-the-scope-gate
audience: skill
target: sprint-review
priority: 1
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

Section 4's in-scope test requires **both** a defect in what the run changed
**and** gate authorization for scope growth. `admit_review_findings` resolved
to `false` from the skill default, so the second condition failed for every
finding regardless of severity — and the run therefore merged and shipped
`872f390`, which review then confirmed carries a silent data-loss path that
wipes `admitted`, `excluded`, `branch`, `base_sha`, `standing_auth`,
`checks_ids`, `checks_baseline`, `review_passes` and `review_budget` from a run
manifest while printing a success line. That is a P0 regression **the run
itself introduced**, declined by a gate whose purpose is to stop the run
growing into work nobody asked for — a different thing entirely. The gate is
right to refuse new work; it should not be the mechanism by which a run ships a
defect it created.

## Evidence

- `qozp#9 (context)` — "retro:tooling:sprint-wave — sprint-wave could not dispatch a subagent: no Task/Agent tool was exposed to it, so it created the worktree as orchestrator and ran sprint-batch -> run-bead itself inside it."

## Proposed change

Add a third row to section 4's classification table: a defect **the run's own
merged diff introduced**, at P0, is in scope regardless of the gate — the run
either fixes it or does not merge the commit that carries it. Leave every other
severity and every pre-existing defect on the conservative default. Done: a P0
regression attributable to a bead in `manifest.admitted` classifies in-scope
with `admit_review_findings: false`.
