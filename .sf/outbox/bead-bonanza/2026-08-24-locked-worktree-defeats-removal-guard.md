---
title: "`git worktree remove` cannot remove a locked worktree, so the guard is lost"
description: Document unlock-then-remove for a harness-locked worktree and name `--force` as the wrong reach, since it discards the dirty-tree check the step exists to run
slug: issue-locked-worktree-defeats-removal-guard
audience: skill
target: bead-bonanza
priority: 2
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#2"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

`references/worktree-mechanics.md` mandates raw `git worktree remove` over
`bd worktree remove` specifically to preserve the dirty-tree check, arguing
that `bd`'s version can only proceed via `--force`, which skips every check.
But a harness-created worktree is **locked** — `git worktree list` showed
`locked` on all four of wave 1's — and plain `git worktree remove` refuses a
locked worktree. The orchestrator is forced to `--force` anyway, losing exactly
the guard the rule was written to keep. The rule silently converts into its own
opposite. A clean counter-case is on the record: wave 2's
`sprint-wave`-created worktree `/home/user/.bb-worktrees/dev-qozp-5` removed
**without** `--force`, so the lock comes from the harness, not from the skill's
mechanics.

## Evidence

- `qozp#2` — "retro:tooling:bead-bonanza — the worktree-removal rule collides with harness-created worktrees, and the collision silently drops the guard the rule exists to keep. references/worktree-mechanics.md says to use raw 'git worktree remove', NOT 'bd worktree remove', specifically so the dirty-tree check is preserved."

## Proposed change

In the removal step, state that a locked worktree is unlocked first —
`git worktree unlock <path> && git worktree remove <path>` — and that `--force`
is not the escape hatch for a lock, because it discards the dirty-tree check
the step exists to run. Done: `worktree-mechanics.md` carries the
unlock-then-remove form and names `--force` as the wrong reach.
