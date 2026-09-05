---
title: "State how a capability tier and a collision chain resolve when they disagree"
description: A collision edge serializes beads of different tiers onto one dispatch, and nothing says whether the tier gates the assignment or only informs it — so the resolved tier of a mixed batch is neither decided nor recorded.
slug: issue-tier-and-collision-chain-assignment-are-in-tension
audience: skill
target: swarm-plan
priority: 3
updated_at: 2026-09-05T00:00:00Z

# Evidence — what this issue rests on
findings: ["rc14.4#5", "rc14.8#9"]
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

A bead carries a capability tier in `metadata.tier`, and the run resolves that
tier to a model through the project's `[models]` table. A collision edge
serializes beads that write the same surface onto **one** dispatch, behind
**one** agent. When beads of different tiers land in the same collision chain,
those two mechanisms disagree, and no document says which wins: the partition
assigns by collision and the tier comes along for the ride.

The cost is that a bead's tier is either being recorded and ignored, or being
silently overridden, and a reader of the plan cannot tell which. On the run
that surfaced it the batch ran at the higher tier and nothing was lost — but
nothing about the partition guarantees that. The same shape with `light` and
`frontier` in one chain runs frontier-tier work on a light agent, or burns a
frontier agent on light work, with no signal either way. The implementing agent
noticed, could not resolve it from any document, and proceeded — which is the
right call and also the evidence that the rule is missing.

## Evidence

- `rc14.4#5` — "Tier was 'advanced' for this bead versus 'standard' for
  rc14.1 and rc14.2, all three assigned to the same dispatch (dev-59ic-a) via
  the shared collision mark
  `generated:cli/core/schema/spec-model.schema.json`. Proceeded rather than
  stopping, since the tracker had already assigned rc14.4 to this dispatch id
  and the dispatch prompt itself named it as part of the same serialized chain
  ('shared with rc14.2 and rc14.4'). Flagging in case tier is meant to gate
  which dispatch a bead lands on rather than only inform effort estimation — if
  so, the collision-chain assignment and the tier gate are in tension and a
  future swarm-plan run should reconcile them explicitly."

- Verified against the tracker on the finished run: `rc14.1` and `rc14.2` are
  `standard`, `rc14.4` is `advanced`, and all three carry
  `assignee: dev-59ic-a`.

- `rc14.8#9` — the run's Definition-of-Done item recorded this as one of two
  findings belonging to a skill lens nobody ran: "an `advanced`-tier bead
  landed on a dispatch also holding two `standard` ones because a shared
  collision mark … serialized them, and asked whether tier is meant to gate
  assignment — that is a question about swarm-plan, whose fix lands in the
  repository that vends the skill."

- The manifest records the tier→model map and every bead's tier, and **records
  no resolved tier per dispatch**: `swarm manifest get --run agent-plugins-59ic`
  returns `models: {advanced: opus, frontier: fable, light: haiku, standard:
  sonnet}` and no per-dispatch entry.

## Proposed change

`swarm-plan` states the precedence where the two mechanisms disagree, in one
sentence: **a collision edge serializes regardless of tier, and the batch runs
at the highest tier in it** — so a chain mixing `standard` and `advanced` runs
`advanced`, and no bead is ever worked below the tier it was assigned. Whatever
rule is chosen, the plan then **records the resolved tier per dispatch** on the
manifest beside the batch it names, so the report can say what each dispatch
actually ran at.

Done is: a mixed-tier collision chain produces a plan that names the tier its
dispatch will run at, and the manifest carries that tier — so a reader can tell
an override from an accident without re-deriving the partition.
