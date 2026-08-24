---
title: "A bare `incident:` comment is harvested by nothing"
description: Name the two admissible incident prefixes where the coordinator is told to record one, and require coordinator verification to carry a `retro:proof:` prefix
slug: issue-bare-incident-comment-is-harvested-by-nothing
audience: skill
target: bead-bonanza
priority: 1
updated_at: 2026-08-24T20:40:00Z

# Evidence — what this issue rests on
findings: ["qozp#8 (kind drift, same class)", "so39#1 (kind drift)", "so39#2 (kind drift)"]
recurrence: 1
report: rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md   # run 2; run 1's findings (`qozp#8`) are in rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-so39 (recurrence; first reported from agent-plugins-qozp)
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: claude-opus-5 (coordinator, both runs; effort unrecorded — this harness exposes no effort setting)
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

**Recurred in run 2 (`agent-plugins-so39`), with all three shapes back and the
same role behind each.** A bare `incident:` comment on the epic (`so39` comment
14) carried the review loop's whole disposition and the `--set-metadata`
incident; an unprefixed coordinator-verification comment (`so39.5` comment 1)
carried the independent measurement that falsified a claim the run had shipped;
and `retro:process:` appeared twice (`so39#1`, `so39#2`) — the same prefix drift
as run 1's `qozp#8`, and what makes 2 of run 2's 28 findings unroutable. Both
consequential records reached run 2's findings document only because its
dispatch brief named them; a retro running from the epic id alone, as
`sprint-retro` is designed to, would have seen neither.

## Evidence

- `qozp#8 (kind drift, same class)` — "retro:process:wave-1-bypassed-sprint-wave — wave 1's four dispatches were made by the coordinator directly … rather than by invoking the sprint-wave skill as SKILL.md prescribes." (`process` is not one of the nine kinds; the same prefix-drift class as the bare `incident:` comment.)
- `so39#1 (kind drift)` — "retro:process:duplicate-bead — the coordinator filed the SAME bead twice, 27 minutes apart …" (run 2; `process` again, again the coordinator.)
- `so39#2 (kind drift)` — "retro:process:coordinator-dispatched-directly incident — the coordinator dispatched wave 1 itself rather than invoking sprint-wave, deliberately and with the reason recorded rather than by omission." (run 2; carries the word `incident` behind an unroutable kind, so it is lost twice over.)
- *(no finding id — the shape itself)* — `agent-plugins-so39` comment 14, bare `incident:` prefix: "incident: review pass 1 filed 10 findings, admitted ZERO, and three of them (so39.4, so39.5, so39.6) are defects IN THE CORRECTIONS THIS RUN SHIPPED …" Matched by neither harvest.
- *(no finding id — the shape itself)* — `agent-plugins-so39.5` comment 1, unprefixed: "coordinator verification — CONFIRMED by independent measurement …" Matched by neither harvest.

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

**What the recurrence adds: put the fix where the coordinator is told to
record, not in the shared protocol reference.** Across both runs, every instance
is *coordinator-role* drift — no dev agent produced one, in either run, on any
of the three shapes. The devs read `run-bead`, which carries the rule and which
they obey; the coordinator does not, and no document it does read names the
form. So restating the rule in the shared findings-capture reference would not
have prevented a single observed instance. The change has to land in
`bead-bonanza`, inline at each point a coordinator is instructed to write a
comment.

Done: one admissible incident form, `sprint-report`'s matcher is a single
alternation, and every comment a coordinator is instructed to write carries a
prefix the harvest matches.
