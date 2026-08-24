---
title: "The run manifest omits `models` and `effort`, so every issue's provenance is unfillable"
description: Make `models` and `effort` required keys of the coordinator's manifest write, and fail loudly when the pre-dispatch check finds either absent
slug: issue-the-manifest-omits-models-and-effort
audience: skill
target: bead-bonanza
priority: 2
updated_at: 2026-08-24T20:40:00Z

# Evidence — what this issue rests on
findings: []
recurrence: 0
report: rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md

# Provenance — who and what produced this
run: agent-plugins-so39
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: claude-opus-5 (retro agent `retro-so39`; effort unrecorded — this harness exposes no effort setting)
---

## What is wrong, and what it costs

`sprint-plan/references/run-record.md` lines 58-59 specify
`"models": {"light": …, "standard": …, "advanced": …, "frontier": …}` and
`"effort": "low|medium|high|xhigh|unrecorded"`, and
`bead-bonanza/references/run-record.md` names both as the coordinator's to
write, with an explicit instruction to write `unrecorded` where the harness
exposes no effort setting. `agent-plugins-so39`'s manifest carries **neither
key**. `sprint-retro`'s closing step requires exactly these two values ("an
issue without them ships with an unfillable field"), so every skill-audience
issue this retro writes would carry a guess. Run 1's manifest carried both,
which makes this a regression rather than an unimplemented rule. The cost is
that findings cannot be weighed by capability tier, which is the one thing the
`model` field exists for.

## Evidence

**No `retro:` finding records this omission.** It was measured while writing the
findings document and recorded there rather than as a bead comment, so this
issue carries an empty `findings:` list by necessity rather than by oversight —
the claim below is the whole of its evidence, and it is reproducible on the
epic.

- *(no finding id)* — measured on `agent-plugins-so39`:
  `bd show agent-plugins-so39 --json | jq '.[0].metadata | keys'` returns 13
  keys, and `models` and `effort` are not among them. The run's model lineup was
  recovered instead from the coordinator's `amend:` dispatch comment:
  `dev-so39-1` sonnet, `dev-so39-2` opus, `dev-so39-3` sonnet, `dev-so39-4`
  opus, `dev-so39-5` opus, `dev-so39-6` haiku; coordinator and ceremony agents
  `claude-opus-5`. Recorded in
  `rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md`,
  candidate C3 and the document's opening "Model lineup" note.

## Proposed change

Add `models` and `effort` to the coordinator's manifest write, and to whatever
pre-dispatch check reads the manifest, as **required** keys — failing loudly on
absence rather than letting the run proceed. `unrecorded` is already the
specified value for a harness with no effort setting, so there is no case where
the key can legitimately be missing. Done: a run whose manifest lacks either key
cannot reach dispatch, and a retro can fill an issue's `model` field from the
manifest alone.
