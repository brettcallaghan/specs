---
title: "No skill tells a coordinator to check before `bd create`, so it filed the same bead twice"
description: State a pre-create dedupe check where `bead-bonanza` instructs the coordinator to file a bead, and name the `--status all` flag that makes it work
slug: issue-a-pre-create-dedupe-step-needs-a-mechanism-that-works
audience: skill
target: bead-bonanza
priority: 1
updated_at: 2026-08-24T20:40:00Z

# Evidence — what this issue rests on
findings: ["so39#1", "so39#11", "so39#9"]
recurrence: 0
report: rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md

# Provenance — who and what produced this
run: agent-plugins-so39
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "2.0"
model: claude-opus-5 (coordinator; effort unrecorded — this harness exposes no effort setting)
---

## What is wrong, and what it costs

The coordinator filed the same bead twice 27 minutes apart — `agent-plugins-vcka`
(16:13:59Z) and `agent-plugins-xpn8` (16:41:04Z), same two call sites, same
pipeline, near-identical acceptance criteria. Nothing caught it until
`sprint-scope` did at admission, and scoping could not resolve it either, because
deduping is a structural change it may not make unilaterally. No skill instructs
an agent to check before `bd create`; `groom-beads` treats duplication as an
anti-pattern to *reject during grooming*, which is a different step in a
different role. Remembering what you filed is not available: the first filing
happened in a turn whose context the coordinator no longer held, which is the
normal condition for a long run. The cost is a duplicate that consumed a bounce,
a manifest `excluded` entry, and a `bd supersede` — and, next time, one that is
not caught.

## Evidence

- `so39#1` — "retro:process:duplicate-bead — the coordinator filed the SAME bead twice, 27 minutes apart: agent-plugins-vcka (16:13:59Z) and agent-plugins-xpn8 (16:41:04Z) … Candidate fix: a pre-create search is cheap and mechanical - state it where a skill instructs an agent to file a bead."
- `so39#11` — "retro:tooling:bd incident — 'bd search' is NOT non-functional; it EXCLUDES CLOSED ISSUES BY DEFAULT … The recorded fix for the duplicate filing therefore works after all, with one flag." (`bd search metadata` → 2 hits, both deferred; `bd search metadata --status all` → 6 hits, matching `jq -r '.title' export.jsonl | grep -ci metadata` → 6. bd 1.2.2 (dev).)
- `so39#9` — "retro:tooling:bd incident — 'bd search' is NON-FUNCTIONAL on this build … The pre-create dedupe step is still right; it needs a mechanism that exists." (the mechanism as previously — wrongly — written off; corrected by `so39#11`.)

## Proposed change

Where `bead-bonanza` instructs the coordinator to file a bead, state the
pre-create check and give it the flag that makes it work:

```sh
bd search "<two or three distinctive title words>" --status all
```

and inspect the hits before creating. Naming `--status all` is the whole of the
fix — the bare form silently hides every closed bead, which on a long run is
most of them, so a zero result from the default is not evidence that nothing
matches. Done: the coordinator's file-a-bead step names a search command that
returns closed issues, and a duplicate of a closed bead is visible before
`bd create` rather than at admission.
