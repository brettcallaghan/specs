---
title: "swarm dispatch accept rebases a dispatch branch inside the worktree that holds it"
description: accept runs git rebase on a concurrent dispatch branch from the main checkout, which git refuses while the dispatch worktree holds that branch.
slug: issue-accept-rebases-a-branch-held-by-a-worktree
audience: skill
target: swarm
priority: 2
updated_at: 2026-10-07T06:36:09Z

findings: ["c4s1#8"]
recurrence: 0
report: docs/retros/2026-10-07-c4s1-c4-diagrams-parser.md

run: specs-c4s1
branch: sprint/specs-c4s1
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "dev-specs-c4s1-1 opus, dev-specs-c4s1-2 sonnet, dev-specs-c4s1-3 haiku, dev-specs-c4s1-4 haiku; coordinator, plan, review and retro agents on the session default model; effort unrecorded"
---

## What is wrong, and what it costs

When a second concurrent dispatch is accepted after another has advanced the sprint branch, `swarm dispatch accept` rebases the dispatch branch from the main checkout. Git refuses because the dispatch's worktree still has the branch checked out: `git rebase` exits 128 with "fatal: 'bb/dev-specs-c4s1-2' is already used by worktree". Cost: every concurrent dispatch accepted second fails and needs a hand rebase inside the worktree before accept can be re-run.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `c4s1#8` — "retro:tooling:swarm swarm dispatch accept could not rebase a concurrent dispatch branch: 'git rebase' exited 128, "fatal: 'bb/dev-specs-c4s1-2' is already used by worktree" — it rebases the branch from the main checkout while the dispatch's worktree still holds it. Repaired by running git rebase 36f09ee inside the worktree (clean), then accept exit 0."

## Proposed change

`swarm dispatch accept` runs the rebase inside the dispatch's own worktree (`git -C <worktree> rebase <sprint-head>`) when one holds the branch, and from the main checkout only when none does. Done when accepting a dispatch whose branch is behind the sprint branch and still held by its worktree succeeds without manual steps.
