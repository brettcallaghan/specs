---
title: "swarm manifest validate ignores the run epic's grooming metadata keys"
description: The manifest shares the epic's metadata namespace with groom-beads, so validate reports grooming keys as undeclared on every groomed run epic.
slug: issue-manifest-validate-flags-grooming-metadata
audience: skill
target: swarm
priority: 3
updated_at: 2026-10-07T06:36:09Z

findings: ["c4s1#4"]
recurrence: 0
report: docs/retros/2026-10-07-c4s1-c4-diagrams-parser.md

run: "the sprint run recorded in report"
branch: sprint/<run-epic>
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "dev-<run-epic>-1 opus, dev-<run-epic>-2 sonnet, dev-<run-epic>-3 haiku, dev-<run-epic>-4 haiku; coordinator, plan, review and retro agents on the session default model; effort unrecorded"
---

## What is wrong, and what it costs

`swarm manifest validate` reports the run epic's own grooming metadata (`tier`, `c4_changes`, `grooming_status`, ...) as undeclared manifest keys, because the manifest is stored in the same metadata object that groom-beads writes on every bead. Cost: a run over a groomed epic can never validate clean, so the validator's verdict is ignored and a real undeclared key would go unnoticed.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `c4s1#4` — "retro:tooling:swarm swarm manifest validate reports the run epic's own grooming metadata (tier, c4_changes, grooming_status, ...) as undeclared manifest keys, because the manifest shares the epic's metadata namespace with groom-beads' bead metadata; a run over a groomed epic can never validate clean."

## Proposed change

`swarm manifest validate` treats the keys groom-beads writes (the bead-metadata schema's keys) as known non-manifest keys and does not report them; it still reports any key neither schema declares. Done when validate on a groomed run epic with a complete manifest exits 0.
