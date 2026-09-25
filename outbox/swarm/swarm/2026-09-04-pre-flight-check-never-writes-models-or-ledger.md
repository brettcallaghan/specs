---
title: "pre-flight-check.sh writes the models and ledger keys its own stage requires"
description: The script writes neither key, so `manifest validate` reports both absent for the life of the run and the derived stage stays pinned at pre-flight.
slug: issue-pre-flight-check-never-writes-models-or-ledger
audience: skill
target: swarm
priority: 0
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#3"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

`pre-flight-check.sh` claims in its own header to write the model lineup to the
manifest, and writes neither `models` nor `ledger`. Because the manifest's
**stage is derived from the keys present**, the two missing keys pin a run at
stage `pre-flight` forever — this run reported `pre-flight` with `branch`,
`base_sha`, `admitted` and `standing_auth` all written and two beads already
committed. Two costs follow. `swarm manifest validate` reports two permanent
absences on every healthy run, which trains a reader to ignore its output; and
the run's **model lineup and effort become unrecoverable**, so this run's retro
had to file every issue with `model:` unrecorded — the one field the issue
template calls load-bearing, because findings differ systematically by the model
that produced them.

## Evidence

- `d7w5.2.1#3` — "pre-flight-check.sh never writes two keys its own stage
  requires. Measured: `grep -c ledger pre-flight-check.sh` -> 0, and the only two
  `manifest_merge` calls (lines 764, 807) write
  checks_baseline/checks_baseline_status/checks_ids and concurrency_cap. Its
  header at line 21 claims it writes 'the model lineup' to the manifest, but no
  'models' key exists on the run epic. Consequence: `swarm manifest validate`
  reports models and ledger absent forever, and because stage is derived from
  the keys present, this run's stage is pinned at 'pre-flight' even with branch,
  base_sha, admitted and standing_auth all written and two beads already
  committed."

Confirmed still true at retro time on this run:
`swarm manifest get --run agent-plugins-d7w5 --key models --key effort` →
`models: ok:false, value:null, "models is not set"`; `effort: ok:true,
value: unrecorded`.

## Proposed change

`pre-flight-check.sh` writes `models` and `ledger` in the same
`manifest_merge` that already writes `concurrency_cap`, from the values it has
resolved by that point: `models` from the project's `[models]` tier mapping as
resolved for this run, and `ledger` from the ledger location the stage
establishes. Where a value genuinely cannot be resolved, it is written as an
explicit recorded null with a reason rather than left absent, so `validate` can
distinguish "resolved to nothing" from "never written".

Done is: on a run that has completed pre-flight, `swarm manifest get --run
<epic> --key models --key ledger` returns `ok: true` for both, and the derived
stage advances past `pre-flight` once dispatch begins.
