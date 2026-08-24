---
title: "`bb-status.sh bead` fails a bead that legitimately produced no commit"
description: Have signals 2 and 3 read `metadata.outcome` and report OK-by-declaration where it is `declined-no-diff`, instead of FAIL
slug: issue-bb-status-fails-a-declined-bead
audience: skill
target: bead-bonanza
priority: 2
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#7"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

`agent-plugins-qozp.3`'s own acceptance permitted a decline branch — "OR the
bead closes as declined with an upstream report filed as a bead and linked to
this one" — and it took it: `bd setup claude` wrote nothing, so there was no
diff to commit. Signal 2 ("commit: none found") has no way to express that, so
a correct outcome reads as repairable and the coordinator has to override the
script by hand. The cost is an operator learning to override a failing signal,
which is how a real failure gets waved through next time.

## Evidence

- `qozp#7` — "retro:tooling:bb-status-declined-bead — bb-status.sh bead exits 1 on a bead that legitimately produced no commit. agent-plugins-qozp.3's own acceptance permitted a decline branch ('OR the bead closes as declined with an upstream report filed as a bead and linked to this one'), which it took: bd setup claude wrote nothing, so there was no diff to commit."

## Proposed change

Have signals 2 and 3 read `metadata.outcome`: where it is `declined-no-diff`
(set on `qozp.3` this run), report OK-by-declaration naming the outcome rather
than `FAIL`. Done: `bb-status.sh bead` on a bead carrying
`outcome=declined-no-diff` exits 0 and says why.
