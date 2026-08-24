---
title: "A bare `incident:` comment is harvested by nothing"
description: Name the two admissible incident prefixes where the coordinator is told to record one, and require coordinator verification to carry a `retro:proof:` prefix
slug: issue-bare-incident-comment-is-harvested-by-nothing
audience: skill
target: bead-bonanza
priority: 1
updated_at: 2026-08-24T13:58:35Z

# Evidence — what this issue rests on
findings: ["qozp#8 (kind drift, same class)"]
recurrence: 0
report: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-qozp
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: <fill in — the model that ran this run> (retro)
---

## What is wrong, and what it costs

The protocol admits two forms for an incident: `retro:<kind>:<target> …
incident` on the bead, and (per `run-record.md`) `amend: incident …` on the
epic. `bead-bonanza` prescribes `amend:` but never names the incident forms at
the point a coordinator records one, and this run's coordinator wrote a bare
`incident: …` comment. Measured against both harvests over this run's comment
set: `sprint-retro`'s `startswith("retro:")` misses it, and `sprint-report`'s
`^(retro:[a-z]+:[^ ]+ incident|amend: incident)` misses it. So does an
unprefixed coordinator-verification comment on `qozp.18`. The cost is total
loss: the run's P0 incident and its independent confirmation were invisible to
both downstream ceremonies, and were recovered here only because a human
handoff named them.

## Evidence

- `qozp#8 (kind drift, same class)` — "retro:process:wave-1-bypassed-sprint-wave — wave 1's four dispatches were made by the coordinator directly … rather than by invoking the sprint-wave skill as SKILL.md prescribes." (`process` is not one of the nine kinds; the same prefix-drift class as the bare `incident:` comment.)

## Proposed change

**Reduce the admissible forms from two to one, then state it where the
coordinator reads.**

`run-bead/SKILL.md:245` already has this right — "an incident is a severity,
not a kind": an ordinary `retro:<kind>:<target>` comment carrying the word
` incident ` after the prefix, harvested by the same query as every other
finding and routed by its kind. There is one collection process already, and
nothing to split off.

The second form is the problem. `run-bead/SKILL.md:242` also prescribes
`amend: incident on {ID}` on the epic — an epic-level *pointer* to a finding
the bead-level harvest already carries. It buys nothing and costs the second
alternation in `sprint-report`'s matcher
(`^(retro:[a-z]+:[^ ]+ incident|amend: incident)`), which is the ambiguity that
made a bare `incident:` look plausible to write. **Drop the pointer form**; the
bead is the single home.

Then state the surviving form in `bead-bonanza`, at the point a coordinator is
told to record one — today the rule lives only in `run-bead`, which a
coordinator never reads. Say explicitly that a bare `incident:` is harvested by
nothing. Require coordinator verification of a review finding to be recorded as
`retro:proof:<bead>` rather than unprefixed.

Done: one admissible incident form, `sprint-report`'s matcher is a single
alternation, and every comment a coordinator is instructed to write carries a
prefix the harvest matches.
