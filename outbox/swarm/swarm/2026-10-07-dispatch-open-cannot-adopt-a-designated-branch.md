---
title: "`swarm dispatch open` should adopt an existing, harness-designated branch"
description: Let swarm dispatch open adopt a harness-designated existing branch
slug: issue-dispatch-open-cannot-adopt-a-designated-branch
audience: skill
target: swarm
priority: 2
updated_at: 2026-10-07T06:05:00Z
findings: []
recurrence: 1
report: docs/retros/2026-10-07-sra-c4-prerequisites-record-blocks.md
run: "the sprint run recorded in report"
branch: claude/project-thread-es7x8o
skill_version: swarm plugin as installed 2026-10-07 (swarm CLI 1cea7e43)
model: dispatches resolved through manifest.models (standard->sonnet, light->haiku); served model and effort unrecorded
---

**What is wrong, and what it costs.** In a cloud session the harness permits
pushing only one designated branch; `swarm dispatch open` refuses to adopt an
existing branch, so the coordinator hand-writes `branch`/`base_sha` onto the
manifest, bypassing whatever `open` validates.

**Evidence.** "sprint branch is the session's designated branch
claude/project-thread-es7x8o (recorded directly, base_sha b626128…), because
the harness permits pushing only that branch and swarm dispatch open refuses to
adopt an existing branch." Not reproduced by the retro.

**Proposed change.** `swarm dispatch open --branch <existing>` adopts an
existing branch, recording its current tip as `base_sha`. Done when that
invocation exits 0 and the manifest's `branch`/`base_sha` match.
