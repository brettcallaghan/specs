---
title: "Exclude ceremony beads from `swarm status`'s scope_additions"
description: A ceremony bead is structurally never in manifest.admitted, so it always reads as scope creep and trains the reader to ignore the signal.
slug: issue-swarm-status-counts-ceremony-beads-as-scope-creep
audience: skill
target: swarm
priority: 1
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#8"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

`swarm status` computes `scope_additions` as beads under the run epic that are
not in `manifest.admitted`. Ceremony beads — the phase epics and the
authorization gate task the run creates for itself — are **structurally never**
in `admitted`, so they appear as scope additions on every run, including a run
with no creep at all. A signal that is dirty by construction is a signal a reader
learns to skip, and the one case it exists to catch — genuine unadmitted work
appearing in the run, which finding `d7w5.2.1#10` shows can happen nine beads at
a time — is then indistinguishable from the noise.

## Evidence

- `d7w5.2.1#8` — "`swarm status` reports ceremony beads as scope additions.
  scope_additions listed agent-plugins-d7w5.2.1 (the run's own authorization gate
  task, closed) alongside the three genuinely-excluded beads. A ceremony bead is
  structurally never in manifest.admitted, so it will always read as creep and
  trains the reader to ignore the signal."

## Proposed change

`swarm status` excludes the run's own ceremony beads — the phase epics and the
authorization gate task it created — from `scope_additions`, identifying them by
the same construction that creates them rather than by title matching. They
remain visible in the status output as the run's ceremony, reported separately
from creep.

Done is: on a clean run whose only non-admitted descendants are its own ceremony
beads, `swarm status` reports `scope_additions` empty.
