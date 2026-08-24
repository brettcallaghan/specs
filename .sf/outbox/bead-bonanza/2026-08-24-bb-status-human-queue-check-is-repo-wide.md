---
title: "`bb-status.sh`'s human-queue check is repo-wide, not run-scoped"
description: Scope the human-queue check to the run's admitted set, the way the neighbouring checks already are
slug: issue-bb-status-human-queue-check-is-repo-wide
audience: skill
target: bead-bonanza
priority: 2
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#5"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: <fill in — the model that ran this run> (coordinator)
---

## What is wrong, and what it costs

The check runs bare `bd human list` and failed the run's health with "human
queue: 1 item(s) awaiting a decision" for `agent-plugins-qozp.12`, a bead that
is **deferred and explicitly excluded at admission**. Every other check in the
same report scopes correctly — "bd lint: nothing flagged in this run's queue",
"every ready bead carries a tier". The cost is that a run can never read clean
while any deferred decision exists anywhere in the tracker, which trains the
operator to ignore the signal.

## Evidence

- `qozp#5` — "retro:tooling:bead-bonanza — bb-status.sh's human-queue check is not scoped to the run's admitted set. It reported 'FAIL human queue: 1 item(s) awaiting a decision' for agent-plugins-qozp.12 (DECISION D1), which is DEFERRED and explicitly excluded at admission."

## Status — DEFERRED pending a design

The owner deferred this on 2026-08-24: the mechanical fix below is correct
but too small a lens. `bd human` is the whole human-decision lifecycle —
flagging a decision, listing what is outstanding, answering it, withdrawing
it — and this run hit defects in three of those four steps, which suggests
the shape of the wrapper matters more than this one check. Reopen it behind
a design that settles how a run assigns a decision to a human and manages
its lifecycle; the fix below is then a consequence of that design rather
than a patch to a check.

What the design has to account for, all measured on bd 1.2.2:

- `bd human list` has no scope flag. It is repo-wide, and the intersection
  has to be computed by the caller.
- It counts CLOSED beads. `--status=open` fixes that half and already
  exists; a scratch database with one open and one closed human-labelled
  bead showed both under the bare call and only the open one under
  `--status=open`.
- `bd human list --json` returns `null`, not `[]`, when the queue is empty.
  A consumer that assumes a list breaks on the empty case.
- `bd human dismiss` fails with `storage is nil` while every other
  subcommand resolves the same id in the same shell. The working path is
  `bd update <id> --remove-label human`.
- There is no `bd human create`. A decision reaches the queue only by
  labelling: `bd update <id> --add-label human`.

## Proposed change

**Two defects, and only one needs code.**

`bd human list` counts CLOSED beads. Measured on bd 1.2.2 in a throwaway
database: a bead closed with `bd close` still appears in the listing, and in
this run `bd human list` printed `agent-plugins-qozp.12` with `Status: closed`
after it had been answered and closed. A flag already fixes this —
`bd human list --status=open` — and it is documented in the command's own
`--help`.

Scope needs an intersection, because there is no scope flag. Take the shape
from the `bd lint` check ten lines above, which solves the identical problem by
parsing ids and intersecting with the run's queue:

```python
run_ids  = manifest["admitted"] | descendants(epic)
queue    = {i["id"] for i in bd_json("human list --status=open") or []}
blocking = queue & run_ids
```

`or []` is load-bearing: `bd human list --json` returns **`null`**, not `[]`,
when the queue is empty (measured). Iterating that is the same class of trap
that produced the character-by-character `admitted` bug.

Report both halves, so the fix does not create a false signal in the other
direction:

```
FAIL  human queue: 1 item(s) awaiting a decision in this run: agent-plugins-qozp.12
OK    human queue: clear for this run (2 open elsewhere in the tracker)
```

Done: an open human-needed item outside `manifest.admitted` does not fail the
run's health, a closed one is never counted at all, and the OK line still says
how many exist elsewhere.
