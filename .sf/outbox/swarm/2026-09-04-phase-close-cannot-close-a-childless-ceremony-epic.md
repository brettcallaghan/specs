---
title: "`swarm phase close` closes a childless ceremony epic"
description: A pre-flight phase epic has no children by construction, and phase close refuses it — leaving open the phase epic whose closure the chain depends on.
slug: issue-phase-close-cannot-close-a-childless-ceremony-epic
audience: skill
target: swarm
priority: 1
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#7"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

`swarm phase close` decides eligibility by looking for an open child to close,
and a ceremony phase epic can have none. The pre-flight phase epic is childless
**by construction**, so the command reports the epic ineligible and leaves it
open — and a phase epic left open blocks the chain, which is the exact failure
the command exists to prevent. The cost is that the run's own ceremony has to be
completed by hand with `bd close`, outside the command that is supposed to own
phase transitions, so the phase record is written by whichever operator happened
to notice rather than by the tool.

## Evidence

- `d7w5.2.1#7` — "`swarm phase close` cannot close a childless ceremony epic.
  The pre-flight phase epic has no children by construction, so it reported 'the
  tracker does not report this epic eligible, and no open child was found among
  its own' and stayed open; I closed it with `bd close` by hand. A phase epic
  left open blocks the chain, which is the failure mode the command exists to
  prevent."

## Proposed change

`swarm phase close` treats **zero children** as trivially satisfied rather than
as ineligible: a phase epic with no open children and nothing else blocking it
closes. The "no open child was found" branch stays only for the case where the
epic is genuinely blocked by something, and its message names what.

Done is: `swarm phase close` on a freshly created, childless pre-flight phase
epic exits 0 and closes it, with no `bd close` by hand and no `--force`.
