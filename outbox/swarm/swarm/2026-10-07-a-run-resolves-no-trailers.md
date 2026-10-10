---
title: "A run resolves no `trailers`, so `--trailers-from` is never usable"
description: Resolve trailers at scope so every close can pass --trailers-from
slug: issue-a-run-resolves-no-trailers
audience: skill
target: swarm
priority: 2
updated_at: 2026-10-08T22:04:20Z
findings: ["qqnx.1#7", "qqnx#2"]
recurrence: 4
report: docs/retros/2026-10-07-sra-c4-prerequisites-record-blocks.md
run: "the sprint run recorded in report"
branch: claude/project-thread-es7x8o
skill_version: swarm plugin as installed 2026-10-07 (swarm CLI 1cea7e43)
model: dispatches resolved through manifest.models (standard->sonnet, light->haiku); served model and effort unrecorded
---

**What is wrong, and what it costs.** Third run running: the manifest never
carries `trailers`, closes run without `--trailers-from`, and no commit carries
the session's required attribution trailers.

**Evidence.** `swarm manifest get --run <run-epic> --key trailers` → `ok: false,
value: null`; `git log --format='%h|%(trailers:only,unfold)|' b626128..HEAD` →
ten empty blocks.

**Proposed change.** As filed for an earlier run: the scope step resolves `trailers`
onto the manifest, and the close step always passes `--trailers-from`. Done
when a run's commits carry them.

**Recurrence: two later runs (2026-10-08).** The first run's
record measured 31 of 34 commits without the trailer. In the second, the
first close hit the same refusal. `qqnx.1#7`: "`--trailers-from <later-run>`
exited 2 with "<later-run> carries no `trailers` in its manifest". The session
attribution trailers were therefore written into the commit message by hand."
`qqnx#2`: "--trailers-from <later-run> exited 2 because the manifest carried no
trailers; the coordinator has now set manifest.trailers." Once the coordinator
had set the trailers by hand, every later close carried them. All 22 commits in
`4676bc4..74ee559` carry both trailers, as measured by this run's retro.
