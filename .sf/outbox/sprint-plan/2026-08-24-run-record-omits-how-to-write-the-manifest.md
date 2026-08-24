---
title: "`run-record.md` documents the manifest's shape but not how to write it"
description: Document that array and object manifest keys need `--metadata` with a read-modify-write merge, and show the `excluded` entry as one id per entry
slug: issue-run-record-omits-how-to-write-the-manifest
audience: skill
target: sprint-plan
priority: 1
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#4", "qozp#6"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "0.2"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

`references/run-record.md` shows the manifest's `admitted`, `excluded` and
`models` keys as JSON arrays and objects, and says nothing about writing them.
`bd`'s `--set-metadata` stores a value as a **string**; arrays need
`--metadata`, which **replaces** the whole object and therefore needs a
read-modify-write. `groom-beads/references/capability-tier.md` documents that
trap for capability tiers, in a different skill, where an orchestrator writing
a manifest will not read it. Separately, the `excluded` entry shape is
`[{"id": …, "reason": …}]` — one id per entry — and the natural thing to write
when several beads share a reason is `[{"ids": […], "reason": …}]`, which is
valid JSON and wrong. Both cost this run: the manifest was written with
`--set-metadata`, stored as strings, and the run's own status report became
unreadable.

## Evidence

- `qozp#4` — "retro:tooling:bead-bonanza — the --set-metadata/--metadata array trap bit the RUN MANIFEST, and bb-status.sh reports the damage in a way that hides the cause."
- `qozp#6` — "retro:gap:run-record — the manifest's 'excluded' shape is easy to get wrong and nothing validates it. run-record.md documents [{\"id\":\"<bead-id>\",\"reason\":...}] — ONE id per entry. I wrote [{\"ids\":[...],\"reason\":...}] to group beads sharing a reason, which is the natural thing to write and reads fine as JSON."

## Proposed change

Add to `run-record.md`'s manifest section: array and object keys are written
with `--metadata` and a read-modify-write merge, never `--set-metadata`, which
stores a string; and show the `excluded` entry as one id per entry with an
explicit "not `ids`" note against the grouped form. Done: an orchestrator can
write every documented manifest key correctly from `run-record.md` alone.
