---
title: "swarm ready holds back a bead whose closed dependency's commit is not yet in its dispatch branch"
description: swarm ready reported a bead ready because its dependency was closed in the tracker, though the dispatch worktree did not contain the dependency's commit.
slug: issue-ready-ignores-whether-a-dependency-commit-is-in-the-dispatch-branch
audience: skill
target: swarm
priority: 2
updated_at: 2026-10-08T22:04:20Z

findings: ["qqnx.9#1"]
recurrence: 0
report: docs/retros/2026-10-08-qqnx-spec-cli-qa-round-2-fixes.md

run: "the sprint run recorded in report"
branch: sprint/<run-epic>
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "<withheld by session policy>"
---

## What is wrong, and what it costs

`swarm ready --agent <dispatch>` decides a bead is ready from the tracker alone. A bead counts as ready once its dependencies read `closed`. When the dependency was worked in another concurrent dispatch, its commit reaches the sprint branch on accept. It does not reach a dispatch worktree that was cut before that accept. The dependent bead is then handed to an agent whose tree lacks the code it depends on. The agent either builds on a stale tree or has to notice the gap and rebase on its own. The dispatch skill does not tell it to do either.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped. The finding was recorded as kind `sequencing`. The merge test sends it here because the fix lands in swarm.

- `qqnx.9#1` — "retro:sequencing The tracker reported <bead-id> ready (its dependency <dep-id> closed), but the dispatch worktree bb/dev-qqnx-3 does not contain <dep-id>'s commit: closed in the tracker is not the same as present in this branch, and the bead's Ordering section depends on the latter. evidence: git merge-base --is-ancestor f9165d7 HEAD → exit 1 on bb/dev-qqnx-3 (HEAD 7b22b57); git log sprint/<run-epic> shows f9165d7 <dep-id> P2 headings matches carry a paste-ready reference. swarm ready --run <run-epic> --agent dev-qqnx-3 listed <bead-id> state: ready."

## Proposed change

For each closed dependency that belongs to the run, `swarm ready --agent <dispatch>` checks that the dependency's commit is an ancestor of the dispatch branch's HEAD. If it is not, ready reports the bead as `needs-rebase`, naming the missing commit, rather than `ready`. The dispatch skill then rebases its worktree onto the sprint branch before claiming. Done when, in the situation above, `swarm ready` lists `<bead-id>` as `needs-rebase` naming `f9165d7`, and lists it as `ready` after `git rebase sprint/<run-epic>` inside the worktree.
