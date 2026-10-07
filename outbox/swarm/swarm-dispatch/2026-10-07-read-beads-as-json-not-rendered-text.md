---
title: "swarm-dispatch: read a bead with `bd show --json`, not the text rendering"
description: Read beads with bd show --json so code in a bead survives verbatim
slug: issue-read-beads-as-json-not-rendered-text
audience: skill
target: swarm-dispatch
priority: 1
updated_at: 2026-10-07T06:05:00Z
findings: ["specs-sra.6#1", "specs-sra.1#2", "specs-sra.4#2"]
recurrence: 0
report: docs/retros/2026-10-07-sra-c4-prerequisites-record-blocks.md
run: specs-sra
branch: claude/project-thread-es7x8o
skill_version: swarm plugin as installed 2026-10-07 (swarm CLI 1cea7e43)
model: dispatches resolved through manifest.models (standard->sonnet, light->haiku); served model and effort unrecorded
---

**What is wrong, and what it costs.** `swarm-dispatch` says "Read each bead in
full with `bd show <id>`". On bd 1.3.1 the text output renders Markdown: it
drops triple backticks inside inline code and doubles `~`. Three of seven
review-filed beads were misread this run; `sra.6`'s agent concluded a correct
reproduction "does not reproduce" and rewrote it, and `sra.4`'s agent had to
guess RFC 6901 tokens from garbled rows. A bead's code examples are exactly the
text that must survive verbatim.

**Evidence.** The three findings quoted above. Measured by the retro, bd 1.3.1:
`bd show specs-sra.6 > f` → `description: |\n  yaml\n  foo: 1`, while
`bd show specs-sra.6 --json | jq -r '.[0].description'` carries
`` ```yaml `` (2 fences). `bd show specs-sra.4` → `"m~~n"`; `--json` → `"m~n"`.

**Proposed change.** In `swarm-dispatch` SKILL.md, replace "Read it in full —
`bd show <id>`" with `bd show <id> --json | jq -r '.[0] | .description,
.design, .acceptance_criteria, .notes'` (comments via `--include-comments`),
and say why in one sentence: the text form renders Markdown and corrupts code.
