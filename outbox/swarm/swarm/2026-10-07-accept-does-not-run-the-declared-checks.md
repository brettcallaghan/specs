---
title: "swarm dispatch accept runs the declared checks on the merged head before reporting the dispatch accepted"
description: A dispatch's claim that all checks pass is accepted as testimony; a format failure on the dispatch's own commit reached the sprint branch.
slug: issue-accept-does-not-run-the-declared-checks
audience: skill
target: swarm
priority: 2
updated_at: 2026-10-07T06:36:09Z

findings: ["c4s1.12#1"]
recurrence: 0
report: docs/retros/2026-10-07-c4s1-c4-diagrams-parser.md

run: specs-c4s1
branch: sprint/specs-c4s1
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "dev-specs-c4s1-1 opus, dev-specs-c4s1-2 sonnet, dev-specs-c4s1-3 haiku, dev-specs-c4s1-4 haiku; coordinator, plan, review and retro agents on the session default model; effort unrecorded"
---

## What is wrong, and what it costs

`swarm dispatch accept` merges a dispatch on the agent's report that its checks pass. Dispatch dev-specs-c4s1-4 reported "All project checks pass (format, ...)" while the declared `format` check failed on its own commit 0a4f37e. Cost: a check failure reached the sprint branch, and was caught only because the coordinator happened to re-run the checks on the combined tree; it then needed a run-introduced bead and a hand fix.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `c4s1.12#1` — "retro:tooling:swarm-dispatch The light-tier dispatch dev-specs-c4s1-4 reported 'All project checks pass (format, ...)' but the declared format check fails on its own commit 0a4f37e; testimony, not measurement. Fixed by the coordinator with gofmt -w (whitespace only)."

## Proposed change

After merging, `swarm dispatch accept` runs the project's declared checks (`bb-checks.py`) on the merged head and reports the dispatch accepted only if the result is no worse than the run's checks baseline; otherwise it exits non-zero and names the failing checks. Done when accepting a dispatch whose commit fails a declared check exits non-zero.
