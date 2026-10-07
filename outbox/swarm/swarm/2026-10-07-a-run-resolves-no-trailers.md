---
title: "A run resolves no `trailers`, so `--trailers-from` is never usable"
description: Resolve trailers at scope so every close can pass --trailers-from
slug: issue-a-run-resolves-no-trailers
audience: skill
target: swarm
priority: 2
updated_at: 2026-10-07T06:05:00Z
findings: []
recurrence: 2
report: docs/retros/2026-10-07-sra-c4-prerequisites-record-blocks.md
run: specs-sra
branch: claude/project-thread-es7x8o
skill_version: swarm plugin as installed 2026-10-07 (swarm CLI 1cea7e43)
model: dispatches resolved through manifest.models (standard->sonnet, light->haiku); served model and effort unrecorded
---

**What is wrong, and what it costs.** Third run running: the manifest never
carries `trailers`, closes run without `--trailers-from`, and no commit carries
the session's required attribution trailers.

**Evidence.** `swarm manifest get --run specs-sra --key trailers` → `ok: false,
value: null`; `git log --format='%h|%(trailers:only,unfold)|' b626128..HEAD` →
ten empty blocks.

**Proposed change.** As filed for `59ic`: the scope step resolves `trailers`
onto the manifest, and the close step always passes `--trailers-from`. Done
when a run's commits carry them.
