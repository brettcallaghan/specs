---
title: "Give `swarm phase close` a path through a target epic that legitimately cannot close"
description: A partially-scoped feature epic stays open by design, and the phase chain then deadlocks — review, retro and post-flight each need --force on a run that did everything right.
slug: issue-phase-close-has-no-path-through-an-unclosable-target-epic
audience: skill
target: swarm
priority: 0
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#11"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

A run whose target is a **partially-scoped feature epic** finishes its admitted
set while the target epic stays legitimately open — closing it would report the
feature complete while its excluded stages are unlanded, which is exactly the
false-complete failure `swarm-scope` warns about. `bd` refuses to close any
phase epic blocked by that open target, so the phase chain deadlocks and every
remaining phase has to be forced. `--force` is not a workaround here, it is the
only path, and it is indistinguishable from forcing past a real failure: a run
that legitimately narrowed its scope and a run that abandoned work both close
their phases the same way, so the record cannot tell them apart afterwards.
A partially-scoped target is the normal case, not the exception.

## Evidence

- `d7w5.2.1#11` — "the phase chain deadlocks whenever the target epic
  legitimately cannot close. Measured: `bd close agent-plugins-d7w5.4` ->
  'cannot close agent-plugins-d7w5.4: blocked by open issues [agent-plugins-f6tl]
  (use --force to override)'. f6tl is open only because .5 .6 .7 are deferred,
  and closing it would falsely report the feature complete — the false-complete
  failure swarm-scope explicitly warns about. So review, retro and post-flight
  each need --force to close, on a run that did everything right. A run whose
  target is a partially-scoped feature epic is the normal case, not the
  exception, and `swarm phase close` has no path through it."

## Proposed change

`swarm phase close` gains a first-class path for this case rather than leaving
the operator to `--force`. When the only thing blocking a phase epic is the run's
own **target** epic, and every bead the manifest admitted is closed, the command
closes the phase and records on the phase epic *why* the target remains open —
naming the target and the excluded or deferred beads holding it. Anything else
blocking the phase still refuses as it does today.

Done is: on a run whose admitted set is fully closed and whose target epic is
open only because of excluded or deferred children, `swarm phase close` for
review, retro and post-flight each exit 0 without `--force`, and each phase epic
carries the recorded reason; and a phase blocked by an *unclosed admitted bead*
still refuses.
