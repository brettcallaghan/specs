---
title: "Add a `human` member to AuthSource so a row changed at the gate records true provenance"
description: AuthSource admits only "swarm.toml" and "default", so a standing-authorization row the human changed at the gate is recorded as a default — which is false.
slug: issue-authsource-has-no-human-member
audience: skill
target: swarm
priority: 1
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#1"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

`swarmmodel.AuthSource` is a two-member enum — `swarm.toml` and `default` — and
the authorization gate exists precisely so a **human** can change a row. There is
therefore no truthful encoding for the gate's own output: a row the human moved
is stored with `source: default`, asserting that nobody chose it. The manifest is
the record a resumed coordinator, a reviewer and a retro all read, and on this
run it says `pull_request: value=none source=default` when the skill default is
`draft` and `none` was the human's decision. The cost is that the manifest
silently misreports which rows a human touched, so nothing downstream can tell an
authorized deviation from an unexercised default — and the only correct account
of the run lives in a gate comment nobody is required to read.

## Evidence

- `d7w5.2.1#1` — "swarmmodel.AuthSource admits only \"swarm.toml\" and
  \"default\", so a row the HUMAN changed at the authorization gate cannot be
  recorded truthfully. This run's pull_request resolved to the skill default
  \"draft\" and the human chose \"none\"; the manifest now reads value=none
  source=default, which is false provenance. The gate exists for a human to
  change rows, so the enum needs a third member (e.g. \"human\")."

## Proposed change

Add a third member `human` to `swarmmodel.AuthSource`, and have the gate record
it on every row whose value the human changed at the gate. A row the human was
shown and left alone keeps the source it resolved from (`swarm.toml` or
`default`); only a row the human moved becomes `human`.

Done is: a gate interaction in which the human changes one row produces a
manifest where that row reads `source: human` and every other row keeps its
resolved source, and `swarm manifest validate` accepts the value.
