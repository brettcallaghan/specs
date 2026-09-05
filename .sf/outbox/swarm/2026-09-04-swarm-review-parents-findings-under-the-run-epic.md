---
title: "File review findings outside the run epic, or file them deferred"
description: swarm-review parents its findings under the run epic, which puts unadmitted work into the run's own dispatch set against an explicit admit_review_findings=false.
slug: issue-swarm-review-parents-findings-under-the-run-epic
audience: skill
target: swarm
priority: 0
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#10"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

`swarm-review` files each review finding as a bead parented under the **run
epic**. The run epic is also what `swarm ready --run` resolves the dispatch set
from, so every finding a review pass raises becomes a dispatchable bead in the
run that raised it — regardless of `standing_auth.admit_review_findings`. On
this run that authorization row was `false`, chosen explicitly by the human at
the gate, and nine unadmitted findings entered the ready set anyway. The cost is
that the gate's answer is enforced only by the coordinator noticing: an
orchestrator that auto-chains on the ready set dispatches work a human declined,
and the manifest's `admitted` list — the record of what was authorized — no
longer describes what runs.

## Evidence

- `d7w5.2.1#10` — "swarm-review parents its findings under the RUN epic, which
  puts unadmitted work straight into the run's own dispatch set. Measured: after
  pass 1, `swarm ready --run agent-plugins-d7w5` reported ready_count 9, every
  one a findings bead with dispatch \"\" and held false, despite
  `standing_auth.admit_review_findings=false`. Only the coordinator's manual
  diff of the ready set against `manifest.admitted` caught it; an orchestrator
  that auto-chained on the ready set would have dispatched all nine against the
  gate's explicit authorization."

## Proposed change

`swarm-review` stops parenting findings under the run epic. A finding bead is
filed against **its origin** — the bead or epic the defect was found in — and
carries a `discovered-from` edge to the bead whose diff raised it, so the audit
trail survives without the hierarchy edge that feeds dispatch. Where
`admit_review_findings` is `true`, the admission step is what moves a finding
into the run, and it records the move in `manifest.admitted` like any other
admission.

Done is: with `admit_review_findings=false`, `swarm ready --run <epic>`
immediately after a review pass reports a ready set containing no bead the
review pass created, and `manifest.admitted` is unchanged by the pass.
