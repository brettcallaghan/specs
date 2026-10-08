---
title: "bb-close-bead.sh takes a flag to skip its tracker push, and the dispatch prompt's no-push rule names it"
description: The close script always ends with bd dolt push, while the dispatch prompt forbids a dev agent from pushing, so an agent that obeys the prompt cannot use the script.
slug: issue-close-script-pushes-the-tracker-the-dispatch-may-not-push
audience: skill
target: swarm
priority: 2
updated_at: 2026-10-08T22:04:20Z

findings: ["qqnx.1#7", "qqnx#2"]
recurrence: 0
report: docs/retros/2026-10-08-qqnx-spec-cli-qa-round-2-fixes.md

run: specs-qqnx
branch: sprint/specs-qqnx
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "<withheld by session policy>"
---

## What is wrong, and what it costs

`plugin/skills/swarm/scripts/bb-close-bead.sh` ends with step 5, `if bd dolt remote list 2>/dev/null | grep -q .; then bd dolt push ...`, and has no flag or environment variable to skip it. `plugin/skills/swarm/references/prompt-dispatch.md` tells the dev agent "You merge nothing and you push nothing." The two contradict each other on every close where a Dolt remote is configured. An agent that obeys the prompt cannot run the close script and has to repeat its steps 1–4 by hand. The script exists so that a bead's closure, its export and its commit land together, and that is the guarantee the hand-run loses. Where the project's own `Stop` hook already pushes the tracker, the per-close push also duplicates it.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `qqnx.1#7` — "retro:tooling:swarm bb-close-bead.sh has no way to skip its step-5 `bd dolt push`. This dispatch forbade pushing, so I ran steps 1-4 by hand and committed without the push. evidence: bb-close-bead.sh step 5 runs `if bd dolt remote list ... ; then bd dolt push` with no flag or environment variable to turn it off. The dispatch prompt said "You merge nothing and you push nothing (no git push, no bd dolt push)"."
- `qqnx#2` — "retro:tooling:swarm bb-close-bead.sh always ends with bd dolt push and has no flag to skip it, while the dispatch prompt says the dev agent pushes nothing; dev-qqnx-1 ran its steps 1-4 by hand."

Measured by this retro, reading swarm source at `1cea7e4`: `bb-close-bead.sh` contains `bd dolt push >/dev/null 2>&1` inside `if bd dolt remote list 2>/dev/null | grep -q .; then`, and no other condition guards it. `references/prompt-dispatch.md` contains "You merge nothing and you push nothing."

## Proposed change

`bb-close-bead.sh` accepts `--no-push` (or honours `BB_NO_TRACKER_PUSH=1`) to skip step 5 and print that it skipped. `prompt-dispatch.md`'s no-push sentence then tells the agent to close with that flag. Done when a dev agent following the prompt runs `bb-close-bead.sh --no-push` end to end, and no `bd dolt push` process is started.
