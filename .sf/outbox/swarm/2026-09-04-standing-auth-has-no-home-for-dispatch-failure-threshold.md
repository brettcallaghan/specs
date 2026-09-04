---
title: "Give `dispatch_failure_threshold` a field on StandingAuth"
description: The gate authorizes seven rows and StandingAuth declares six, so the threshold is resolved at the gate and stored nowhere — a resumed coordinator silently falls back to the flag default.
slug: issue-standing-auth-has-no-home-for-dispatch-failure-threshold
audience: skill
target: swarm
priority: 1
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#2"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

`swarm-scope`'s gate table puts seven rows to the human, and
`swarmmodel.StandingAuth` declares six fields. The seventh —
`dispatch_failure_threshold` — is resolved at the gate, handed to `swarm ready
--threshold`, and then discarded. The threshold is the circuit breaker: it is
what stops a run that keeps returning dispatches having completed nothing. A
coordinator resumed from the epic cannot recover the authorized value, falls
back to the flag default of 3, and **has no way to tell that it did** — so a run
authorized with a tighter breaker silently loses it at exactly the moment a
resumed coordinator is most likely to need it.

## Evidence

- `d7w5.2.1#2` — "swarm-scope's gate table lists seven rows including
  dispatch_failure_threshold, but swarmmodel.StandingAuth declares only six
  fields and has no home for it. The threshold is resolved at the gate and then
  passed to `swarm ready --threshold`, so a resumed coordinator cannot recover
  the authorized value from the manifest and silently falls back to the flag
  default of 3."

## Proposed change

Add `dispatch_failure_threshold` as a seventh field on
`swarmmodel.StandingAuth`, carrying a value and a source like every other row,
and have the gate write it with the rest of the standing authorization. `swarm
ready` reads it from the manifest when `--threshold` is not given, rather than
falling back to the flag default.

Done is: after the gate, `swarm manifest get --run <epic> --key standing_auth`
reports seven rows including `dispatch_failure_threshold` with its value and
source; and a coordinator resumed with no flag runs the breaker at the
authorized value rather than at 3.
