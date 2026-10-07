---
title: "swarm dispatch accept merges into the recorded sprint branch, not the main checkout's current branch"
description: accept fast-forwards a dispatch into whatever branch the main checkout has checked out instead of the run's recorded sprint branch, then fails its own ancestry check.
slug: issue-accept-merges-into-the-current-branch
audience: skill
target: swarm
priority: 1
updated_at: 2026-10-07T06:36:09Z

findings: ["c4s1#3"]
recurrence: 0
report: docs/retros/2026-10-07-c4s1-c4-diagrams-parser.md

run: specs-c4s1
branch: sprint/specs-c4s1
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "dev-specs-c4s1-1 opus, dev-specs-c4s1-2 sonnet, dev-specs-c4s1-3 haiku, dev-specs-c4s1-4 haiku; coordinator, plan, review and retro agents on the session default model; effort unrecorded"
---

## What is wrong, and what it costs

`swarm dispatch accept` merges the dispatch branch into the main checkout's CURRENT branch, not into the sprint branch recorded on the run epic (`manifest.branch`). When those differ — always, in a harness that mandates its own push branch, since `dispatch open` refuses an existing branch as the sprint branch — the merge lands on the wrong branch and accept then exits 2 because the dispatch branch is not an ancestor of the sprint branch. Cost: every run in such a harness fails its first accept and needs a hand repair (`git branch -f` plus a re-run) and a second sprint branch with `accept --no-push`.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `c4s1#3` — "retro:tooling:swarm swarm dispatch accept fast-forwarded the dispatch branch into the main checkout's CURRENT branch (claude/project-thread-u8o164, the harness-designated branch) rather than into the recorded sprint branch sprint/specs-c4s1, then exited 2 because bb/dev-specs-c4s1-1 was not an ancestor of sprint/specs-c4s1 (merged_sha read d21fdef). Repaired by git branch -f sprint/specs-c4s1 6c43deb and re-running accept (exit 0). Related: dispatch open refuses an existing branch as the sprint branch, so a harness that mandates one push branch forces a second sprint branch plus accept --no-push."

## Proposed change

`swarm dispatch accept` merges into `manifest.branch` regardless of what the main checkout has checked out — checking it out first, or updating the ref directly when the merge is a fast-forward — and refuses with a message naming both branches if it cannot. Done when accept run from a checkout sitting on another branch advances the recorded sprint branch and leaves the checked-out branch untouched.
