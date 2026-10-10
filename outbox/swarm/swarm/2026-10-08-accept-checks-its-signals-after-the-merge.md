---
title: "swarm dispatch accept evaluates its per-bead signals before it merges and deletes the dispatch branch"
description: accept merged a dispatch and deleted its branch, then failed it on the committed and findings signals, leaving nothing a re-accept could resolve.
slug: issue-accept-checks-its-signals-after-the-merge
audience: skill
target: swarm
priority: 1
updated_at: 2026-10-08T22:04:20Z

findings: ["qqnx#4"]
recurrence: 0
report: docs/retros/2026-10-08-qqnx-spec-cli-qa-round-2-fixes.md

run: "the sprint run recorded in report"
branch: sprint/<run-epic>
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "<withheld by session policy>"
---

## What is wrong, and what it costs

`swarm dispatch accept` checks its per-bead acceptance signals only after it has merged the dispatch branch and deleted it. The signals are `committed`, meaning one `<id> P<n>` commit subject per bead, and `findings`, meaning at least one `retro:` comment. When a signal fails, the merge has already landed and the branch is gone. A re-accept then fails because the dispatch branch cannot be resolved. The run that hit this had to reword the commit, add an empty close commit for the second bead, add the missing comments, and recreate the branch by hand. Only after all that did a third accept succeed. Any dispatch that legitimately fixes two beads in one change triggers it, and so does any dispatch whose agent forgot a comment.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `qqnx#4` — "retro:tooling:swarm A dispatch that fixes two beads in one commit fails accept's 'committed' signal (one '<id> P<n>' subject per bead) and 'findings' (no retro: comment) after the merge has already happened and the dispatch branch is deleted; re-accept then fails 'bb/dev-qqnx-9 could not be resolved'. evidence: accept exit 1 acceptance_failures [<bead-a>: committed, findings, <bead-b>: committed, findings]; second accept exit 2 rev-parse 128. Fixed by rewording to fc569da + empty close commit 793b724, adding retro:proof comments, recreating the branch at 793b724; third accept exit 0 landed true."

On the sprint branch, `git show --stat 793b724` lists no files. It is the empty close commit for `<bead-b>`.

## Proposed change

`swarm dispatch accept` evaluates every per-bead signal it can compute from the dispatch branch and the tracker before it merges. On any failure it exits non-zero naming the bead and signal, and leaves the branch and worktree untouched, so the coordinator can repair them and re-run accept. Done when accepting a dispatch whose single commit names two beads exits non-zero, the sprint branch has not moved, and `bb/<dispatch>` still resolves.
