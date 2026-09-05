---
title: "Record coordinator tooling friction as a retro: finding, and harvest the run epic"
description: A coordinator's tooling defects are recorded as amend: comments on the run epic, which the retro's harvest query matches on neither the prefix nor the scope — so they reach the record only by being hand-carried in a dispatch prompt.
slug: issue-coordinator-tooling-findings-are-invisible-to-the-harvest
audience: skill
target: swarm
priority: 2
updated_at: 2026-09-05T00:00:00Z

# Evidence — what this issue rests on
findings: ["rc14.8#9"]               # the finding channel's own gap, from the
                                     # other direction: two findings routed to
                                     # a lens nobody ran
recurrence: 0
report: docs/retros/2026-09-05-59ic-spec-corpus-anchor-references.md

# Provenance — who and what produced this
run: agent-plugins-59ic
session: https://claude.ai/code/session_0185zfzFm9maHZjPX4FQqhbZ
branch: sprint/agent-plugins-59ic
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<derived, not observed — manifest.models holds only the tier→model map (standard→sonnet, advanced→opus) and nothing records what served each dispatch>"
---

## What is wrong, and what it costs

The findings capture protocol gives an implementing agent one channel for
tooling friction — a `retro:tooling:<target>` comment on the item it was
working — and the retro harvests that channel with one query over the run's
closure set. **The coordinator has no such channel.** Friction it hits is
orchestration friction, which belongs to no work order, so it lands as an
`amend:` comment on the run epic instead. That record fails the harvest twice
over: the prefix does not match `retro:`, and the run epic is not in its own
closure set.

The cost is that **the most serious tooling defects a run produces are the
ones the record cannot see.** On the run that surfaced this, four of them —
including a P0 defect in which `swarm dispatch accept` folded into the wrong
branch and reported success — were invisible to the harvest and reached the
retro only because the coordinator hand-wrote them into the retro agent's
dispatch prompt. A dispatch prompt dies with its agent, is audited by nobody,
and is not in the repository. A coordinator that did not think to hand them
over, or a retro reconstructed from durable sources after an interruption
(which is exactly what `swarm-recover` exists for), loses them silently — and
every run record's routing counts are understated by precisely the friction the
orchestrator hit.

## Evidence

- **The harvest's scope and prefix, and what they return.** The retro's query
  is `bd export | jq -r 'select(<run scope>) | .comments[]? |
  select(.text | startswith("retro:")) | .text'`. Run over this run's closure
  set it returned **42** findings across the nine work-order beads and **0**
  from `agent-plugins-59ic` itself. (Measured before the retro added its own
  `retro:`-prefixed bookkeeping comment to the epic; that comment is not a
  finding.)

- **What was on the epic instead.** Seven `amend:` comments, of which four are
  tooling defects, quoted in the run record's "What the coordinator recorded"
  section: `swarm dispatch accept` folding into whatever `HEAD` is on and
  reporting `merged: true`; `swarm dispatch accept`'s concurrent path rebasing
  a branch still checked out in its worktree; `swarm dispatch accept`'s push
  step blocked by the session's permission classifier; and two dispatches both
  regenerating `README.md`'s coverage table into a merge conflict. Each is now
  an issue or a ledger entry only because it was hand-carried.

- **`rc14.8#9` is the same gap from the other side.** That finding records two
  items an implementing agent routed to "the skill lens" — and nobody ran a
  skill retro for them, so they sat in a bead comment with no destination:
  "Both are recorded under rfc-0025's Considered and Not Proposed with the
  routing stated, so they are not lost, but no skill-lens RFC exists for them
  and this bead did not author one."

## Proposed change

Two halves of one mechanism, and the issue asks for both because either alone
leaves the gap open.

**The coordinator records tooling friction as a finding.** Orchestration
friction goes on the run epic as `retro:tooling:<tool>` with evidence in the
protocol's shape, and `amend:` keeps the job it does well — recording
amendments to scope and authorization. A defect the coordinator hits is the
same kind of thing as one an implementer hits and should not need a different
prefix to be seen.

**The harvest covers the run epic as well as its closure set.** The retro's
query selects the epic's own `retro:` comments alongside its descendants',
counting each under its bead of origin as it already does.

Done is: a retro run against a run epic — with an **empty** dispatch prompt,
carrying nothing the coordinator remembered to mention — finds the
coordinator's tooling findings in its own harvest, and its routing counts
include them.
