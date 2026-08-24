---
title: "`bb-status.sh` renders a mistyped manifest key as plausible wrong output"
description: Type-check each manifest key before use so a shape error is reported as a FAIL naming the key and observed type, not iterated into a character count
slug: issue-bb-status-renders-mistyped-manifest-key
audience: skill
target: bead-bonanza
priority: 1
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#4", "qozp#6"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

Given `admitted` stored as a string rather than a list, `bb-status.sh` iterated
it **character by character** and reported
`FAIL 116 admitted bead(s) not closed: [ " a g e n t - p l u g i n s …` and
`OK 347 bead(s) EXCLUDED at admission` — 116 and 347 being character counts.
Given `excluded` entries carrying `ids` rather than `id`, it read `.id`, found
nothing, and printed `None (reason); None (reason); None (reason)`. In both
cases the script produced confident, plausible-looking output over a shape
error. The cost is that the real cause is unguessable from the report, and a
reader who does not already know the trap will debug the wrong thing.

## Evidence

- `qozp#4` — "retro:tooling:bead-bonanza — the --set-metadata/--metadata array trap bit the RUN MANIFEST, and bb-status.sh reports the damage in a way that hides the cause."
- `qozp#6` — "retro:gap:run-record — the manifest's 'excluded' shape is easy to get wrong and nothing validates it. … bb-status.sh rendered it as 'None (reason); None (reason); None (reason)' — it read `.id`."

## Proposed change

Type-check each manifest key before use: a string where a list is expected, or
a list entry missing `id`, is a `FAIL` naming the key and the observed type,
not an iteration. Done: `bb-status.sh` run against a manifest with `admitted`
stored as a string reports a shape error rather than a character count.
