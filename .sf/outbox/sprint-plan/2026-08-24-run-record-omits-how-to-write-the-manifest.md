---
title: "`run-record.md` documents the manifest's shape but not how to write it"
description: Document that array and object manifest keys need `--metadata` with a read-modify-write merge, and show the `excluded` entry as one id per entry
slug: issue-run-record-omits-how-to-write-the-manifest
audience: skill
target: sprint-plan
priority: 1
updated_at: 2026-08-24T20:40:00Z

# Evidence — what this issue rests on
findings: ["qozp#4", "qozp#6", "4tjk#1 (correction, run 2)"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md   # run 1; the run-2 correction is in rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp (corrected from agent-plugins-so39)
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "0.2"
model: claude-opus-5 (coordinator, run 1; the run-2 correction measured by claude-sonnet-4-5 as dev-so39-3 and independently by claude-opus-5 as coordinator. Effort unrecorded — this harness exposes no effort setting)
---

## What is wrong, and what it costs

`references/run-record.md` shows the manifest's `admitted`, `excluded` and
`models` keys as JSON arrays and objects, and says nothing about writing them.
`bd`'s `--set-metadata` **JSON-parses its value**: scalars keep their type, and
**arrays and objects — not scalars — are what it stringifies**, so those are the
keys that need `--metadata` plus a read-modify-write merge.
`groom-beads/references/capability-tier.md` documents that trap for capability
tiers, in a different skill, where an orchestrator writing a manifest will not
read it. Separately, the `excluded` entry shape is `[{"id": …, "reason": …}]` —
one id per entry — and the natural thing to write when several beads share a
reason is `[{"ids": […], "reason": …}]`, which is valid JSON and wrong. Both
cost this run: the manifest's array keys were written with `--set-metadata`,
stored as strings, and the run's own status report became unreadable.

## Evidence

- `qozp#4` — "retro:tooling:bead-bonanza — the --set-metadata/--metadata array trap bit the RUN MANIFEST, and bb-status.sh reports the damage in a way that hides the cause."
- `qozp#6` — "retro:gap:run-record — the manifest's 'excluded' shape is easy to get wrong and nothing validates it. run-record.md documents [{\"id\":\"<bead-id>\",\"reason\":...}] — ONE id per entry. I wrote [{\"ids\":[...],\"reason\":...}] to group beads sharing a reason, which is the natural thing to write and reads fine as JSON."
- `4tjk#1 (correction, run 2)` — "retro:proof:groom-beads Reproduced all three measurements myself before writing the correction … Still-true trap: bd update probe-pj2 --set-metadata arr='[\"a\",\"b\"]' … includes \"arr\":\"[\\\"a\\\",\\\"b\\\"]\" -- --set-metadata stores the literal string, confirming why --metadata is reached for at all. bd version: 1.2.2 (dev)." Independently confirmed by the coordinator on `agent-plugins-so39.5`, seven `--set-metadata` pairs in one `bd update`: `num=5` → number, `b=true` → boolean, `f=1.5` → number, `n=null` → null, `s=hello` → string, `arr=["a","b"]` → **string**, `obj={"k":1}` → **string**.

## Proposed change

Add to `run-record.md`'s manifest section: **array and object** keys are written
with `--metadata` and a read-modify-write merge, never `--set-metadata`, which
stringifies them; and show the `excluded` entry as one id per entry with an
explicit "not `ids`" note against the grouped form. Done: an orchestrator can
write every documented manifest key correctly from `run-record.md` alone.

<!--
CORRECTED 2026-08-24 by the run-2 retro (`agent-plugins-so39`), before triage.
This file previously asserted "`bd`'s `--set-metadata` stores a value as a
**string**" as a universal, and that `--metadata` "**replaces** the whole
object". Both are false on bd 1.2.2 and both have been narrowed above:
`--set-metadata` JSON-parses scalars, and `--metadata` merges at the TOP LEVEL
while replacing a NESTED value wholesale — which is why a read-modify-write is
still required for the array and object keys. The issue's proposal is unchanged;
only its rationale was wrong. Measured evidence is the third bullet above.
-->
