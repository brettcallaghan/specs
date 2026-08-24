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
to `false` from the skill default, so the second condition failed for **every
finding regardless of severity**, and the run merged and shipped `872f390` with
its review finding declined unread on the merits.

**The motivating example did not survive verification, and the structural
defect did.** Review filed that finding as a P0 silent-data-loss path on the
premise that `bd update --metadata` replaces the object; measured on bd 1.2.2
it merges at the top level, so no data was lost and the bead was downgraded to
P2. What remains is the shape of the decision: the gate refused the finding on
*scope*, having never weighed *severity*, so a genuine P0 in the run's own diff
would have been declined by exactly the same path, for exactly the same reason.
The gate is right to refuse new work; it should not be the mechanism by which a
run ships a defect it created — and this run demonstrated that mechanism
operating, with the severity question never reached.

## Evidence

- `qozp#9 (context)` — "retro:tooling:sprint-wave — sprint-wave could not dispatch a subagent: no Task/Agent tool was exposed to it, so it created the worktree as orchestrator and ran sprint-batch -> run-bead itself inside it."

## Proposed change

Add a third row to section 4's classification table: a defect **the run's own
merged diff introduced**, at P0, is in scope regardless of the gate — the run
either fixes it or does not merge the commit that carries it. Leave every other
severity and every pre-existing defect on the conservative default. Done: a P0
regression attributable to a bead in `manifest.admitted` classifies in-scope
with `admit_review_findings: false`.
