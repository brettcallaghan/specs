---
title: "The checks-baseline inlining rule does not scale past one agent"
description: Replace verbatim inlining of the baseline in every subagent prompt with the verdict plus the command that reads the full text from the manifest
slug: issue-checks-baseline-inlining-does-not-scale
audience: skill
target: bead-bonanza
priority: 2
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#1"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

`references/prompt-per-bead.md` requires the run epic's `checks_baseline` be
"inlined verbatim" into every subagent prompt, with an explicit warning against
sending the one-word status instead. On this run the baseline text is ~6 KB and
wave 1 dispatched 4 agents, so obeying the rule literally costs ~24 KB of
duplicated prompt — a cost that grows linearly with the wave and is paid before
the agent has done anything. The coordinator deviated rather than pay it, which
means the rule as written is one an orchestrator will break under load.

## Evidence

- `qozp#1` — "retro:tooling:bead-bonanza — the checks-baseline inlining rule does not scale past one agent. prompt-per-bead.md requires the baseline block be 'inlined verbatim' into every subagent prompt, with an explicit warning against sending the one-word status instead. Measured on this run: the baseline text is ~6 KB, and wave 1 dispatches 4 agents, so obeying the rule literally costs ~24 KB of duplicated prompt."

## Proposed change

Replace the verbatim-inlining requirement with: inline the **verdict** (e.g.
"all 13 declared checks passed, exit status 0" plus "any failure you see is
yours") and the exact command that reads the full text from the manifest. Keep
the distinction the current rule protects by requiring an explicit "no baseline
was taken" line where none exists, so a missing block is never ambiguous with
an unstated one. Done: `prompt-per-bead.md` states the verdict-plus-command
form and the explicit no-baseline line.
