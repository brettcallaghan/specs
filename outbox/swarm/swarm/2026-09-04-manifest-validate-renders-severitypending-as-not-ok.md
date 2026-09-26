---
title: "`swarm manifest validate` distinguishes a pending key from a broken one"
description: SeverityPending renders as ok:false, so a healthy pre-dispatch manifest reports failures and a caller routing on ok cannot tell "not written yet" from "wrong".
slug: issue-manifest-validate-renders-severitypending-as-not-ok
audience: skill
target: swarm
priority: 2
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#4"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

The manifest model declares a `SeverityPending` / `RuleNotYetWritten` severity
precisely to say "this key is not written yet and becomes required at a later
stage". `swarm manifest validate` collapses it into `ok: false` alongside real
failures. A healthy pre-dispatch manifest therefore reports six failures of which
four are just "not written yet", and a caller routing on `ok` — which is what the
flag is for — cannot act on the distinction the severity exists to draw. The cost
is a validation command whose red state is the normal state, so an operator
reads past it and a genuinely broken manifest looks like every other run.

## Evidence

- `d7w5.2.1#4` — "`swarm manifest validate` renders SeverityPending as ok:false,
  so a healthy pre-dispatch manifest reports 6 failures of which 4 are just 'not
  written yet; becomes required at the retro stage'. A caller routing on ok
  cannot tell a pending key from a broken one, which is the distinction the
  RuleNotYetWritten severity exists to draw."

## Proposed change

`swarm manifest validate` renders a `SeverityPending` issue with `ok: true` and
reports it in a separate `pending` collection carrying the key, the stage at
which it becomes required, and the message. `ok: false` is reserved for a key
that is present and wrong, or absent past the stage that requires it.

Done is: `swarm manifest validate` against a healthy pre-dispatch manifest exits
with `ok: true` and a non-empty `pending` list; the same command against a
manifest holding a malformed declared key still reports `ok: false`.
