---
title: "A batch's return line carries no findings count, so a silent agent is invisible until acceptance"
description: Add a per-bead findings count to the batch's return line, so a zero is a claim the batch asserts rather than an absence the orchestrator discovers
slug: issue-a-batch-return-line-carries-no-findings-count
audience: skill
target: sprint-batch
priority: 2
updated_at: 2026-08-24T20:40:00Z

# Evidence — what this issue rests on
findings: ["so39#4"]
recurrence: 0
report: rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md

# Provenance — who and what produced this
run: agent-plugins-so39
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "0.1"
model: claude-opus-5 (coordinator; the batch observed ran claude-haiku-4-5 at the light tier. Effort unrecorded — this harness exposes no effort setting)
---

## What is wrong, and what it costs

The light-tier batch — `dev-so39-6`, haiku, three beads — recorded **zero**
`retro:` comments, though the dispatch prompt asked for them explicitly. Two
readings, not equally likely: three one-line edits genuinely produced no
friction, or a light-tier agent dropped a step that is last in the ceremony and
has no output anyone waits for. It surfaced only at acceptance, via
`bb-status.sh bead` signal 5, exit 1 on all three — after the batch had returned
and its agent was gone. The cost is that the run's cheapest agents, running the
most beads, are exactly where the evidence channel silently closes; and
`bb-status.sh` failing three beads for it trains the operator to override the
signal.

## Evidence

- `so39#4` — "retro:gap:dispatch-policy — the light-tier trivial batch (dev-so39-6, haiku, three beads) recorded ZERO retro: comments, and bb-status.sh bead caught it: signal 5 'findings: none recorded', exit 1 on all three. The dispatch prompt asked for them explicitly. … Candidate fix: make the findings step produce something a return line must carry, so silence is visible at dispatch rather than at acceptance." (The other three batches this wave all recorded findings; the zero row is in section 1's harvest table of the run report.)

## Proposed change

Add a findings count to the batch's return line, per bead, so a zero is a value
the batch *asserts* rather than an absence the orchestrator discovers. A batch
that reports `findings: 0` for a bead has made a claim that can be questioned at
return, while the agent is still alive; a batch that reports nothing has not.
Done: every bead in a batch's return line carries a findings count, and an
orchestrator can challenge a zero before the agent is gone.

**Sequencing note.** Rollout plan item 9 deletes `prompt-batch.md`'s return-line
contract, so this change must be made against whatever replaces it, or made part
of item 9.
