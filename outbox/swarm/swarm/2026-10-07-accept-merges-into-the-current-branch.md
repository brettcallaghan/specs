---
title: "swarm dispatch accept merges into the recorded sprint branch, not the main checkout's current branch"
description: accept fast-forwards a dispatch into whatever branch the main checkout has checked out instead of the run's recorded sprint branch, then fails its own ancestry check.
slug: issue-accept-merges-into-the-current-branch
audience: skill
target: swarm
priority: 1
updated_at: 2026-10-08T22:04:20Z

findings: ["c4s1#3", "qqnx#1"]
recurrence: 4
report: docs/retros/2026-10-07-c4s1-c4-diagrams-parser.md

run: "the sprint run recorded in report"
branch: sprint/<run-epic>
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "dev-<run-epic>-1 opus, dev-<run-epic>-2 sonnet, dev-<run-epic>-3 haiku, dev-<run-epic>-4 haiku; coordinator, plan, review and retro agents on the session default model; effort unrecorded"
---

## What is wrong, and what it costs

`swarm dispatch accept` merges the dispatch branch into the main checkout's CURRENT branch, not into the sprint branch recorded on the run epic (`manifest.branch`). When those differ — always, in a harness that mandates its own push branch, since `dispatch open` refuses an existing branch as the sprint branch — the merge lands on the wrong branch and accept then exits 2 because the dispatch branch is not an ancestor of the sprint branch. Cost: every run in such a harness fails its first accept and needs a hand repair (`git branch -f` plus a re-run) and a second sprint branch with `accept --no-push`.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `c4s1#3` — "retro:tooling:swarm swarm dispatch accept fast-forwarded the dispatch branch into the main checkout's CURRENT branch (claude/project-thread-u8o164, the harness-designated branch) rather than into the recorded sprint branch sprint/<run-epic>, then exited 2 because bb/dev-<run-epic>-1 was not an ancestor of sprint/<run-epic> (merged_sha read d21fdef). Repaired by git branch -f sprint/<run-epic> 6c43deb and re-running accept (exit 0). Related: dispatch open refuses an existing branch as the sprint branch, so a harness that mandates one push branch forces a second sprint branch plus accept --no-push."
- `qqnx#1` (a later run, 2026-10-08; record `docs/retros/2026-10-08-qqnx-spec-cli-qa-round-2-fixes.md`) — "retro:tooling:swarm swarm dispatch accept fast-forwarded the CHECKED-OUT branch rather than the run's target: with the coordinator checkout on claude/project-thread-nr65x0, accept for dev-qqnx-1 reported 'serialized ... fast-forwarded into sprint/<later-run>' but sprint/<later-run> stayed at 4676bc4 and claude/project-thread-nr65x0 moved to a22fd79; it exited 2 on the ancestry guard. Its --explain lists 'git rev-parse --verify HEAD^{commit} (the repository is on the target)' but nothing refuses when HEAD is not the target. Repaired by hand: git branch -f claude/project-thread-nr65x0 4676bc4; git checkout sprint/<later-run>; re-ran accept -> exit 0. Expected: accept refuses (exit 2, naming the target) when HEAD is not on the target branch, before any merge."
- Recurrence 4: the records of four later runs cite this issue, one of them noting "fourth run running". The latest of them, quoted above, adds that accept's report says "fast-forwarded into sprint/<later-run>" while that ref did not move.

## Proposed change

`swarm dispatch accept` merges into `manifest.branch` regardless of what the main checkout has checked out — checking it out first, or updating the ref directly when the merge is a fast-forward — and refuses with a message naming both branches if it cannot. Done when accept run from a checkout sitting on another branch advances the recorded sprint branch and leaves the checked-out branch untouched.
