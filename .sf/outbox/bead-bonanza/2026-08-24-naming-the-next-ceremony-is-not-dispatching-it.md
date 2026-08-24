---
title: "A coordinator turn can end with nothing in flight, so the run stops between ceremonies"
description: State the turn-ending invariant where `bead-bonanza` describes ending a turn — a coordinator turn ends with work in flight, a question to the human, or the run's terminal state
slug: issue-naming-the-next-ceremony-is-not-dispatching-it
audience: skill
target: bead-bonanza
priority: 3
updated_at: 2026-08-24T20:40:00Z

# Evidence — what this issue rests on
findings: ["so39#12", "so39#10"]
recurrence: 0
report: rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md

# Provenance — who and what produced this
run: agent-plugins-so39
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: claude-opus-5 (coordinator; effort unrecorded — this harness exposes no effort setting)
---

## What is wrong, and what it costs

The coordinator ended a turn announcing the next ceremony without dispatching
it, with nothing in flight, and the run sat still until the human asked. The
mechanism that would have recovered it exists and works — the durable trigger
armed at first dispatch fired at 17:57:32 with the right instruction — so this
is not a missing signal. What it is: at a ceremony boundary, with the queue
drained and no work in flight, nothing in the skill states that announcing the
next step is not taking it, and a turn that ends there ends the run's forward
motion. A human noticing is what actually resumed it. The cost is bounded and
real: dead time at every stall, and an operator who learns that runs need
watching.

## Evidence

- `so39#12` — "retro:tooling:bead-bonanza incident — the coordinator STALLED at a ceremony boundary: it ended a turn announcing the next ceremony without dispatching it, with nothing in flight, and the run sat still until the human asked. … A run whose progress depends on a human noticing that nothing is happening has no liveness property at its ceremony boundaries."
- `so39#10` — "retro:proof:progress-streaming — CORRECTION to the coordinator's own stall finding … The durable recovery trigger armed at first dispatch (trig_01Tr5qJXx7ZnRFVvZohSZpwc) fired at 17:57:32 carrying exactly the right instruction … So the watchdog design works and needs no new mechanism."

## Proposed change

State the invariant where `bead-bonanza` describes ending a turn:

> A coordinator turn ends with **work in flight**, **a question to the human**,
> or **the run's terminal state**. Naming the next ceremony is not dispatching
> it — a turn that ends by announcing what comes next, with nothing dispatched
> and nothing running, has ended the run.

Done: the skill's turn-ending guidance names the three admissible end states,
and says plainly that an announcement is not one of them.

**Deliberately not proposed here: the recovery trigger's interval.** The same
findings support a second change — choosing the trigger's interval as an
explicit liveness budget, since it bounds dead time at a boundary rather than
preventing a stall. That is left out on purpose. The reporting project is
handling it instead as a background monitor on the bead event log (its rollout
plan item 14, `agent-plugins-qozp.15`), which reacts to the absence of events
rather than polling on a timer, and is the better mechanism of the two. The
invariant above stands on its own and does not depend on which one lands.
