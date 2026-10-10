---
title: "swarm-plan instructs its findings in retro:<kind>:<target> form"
description: swarm-plan carries no findings-capture instruction, so the planner records findings with no kind and they route nowhere.
slug: issue-planner-findings-carry-no-kind
audience: skill
target: swarm-plan
priority: 2
updated_at: 2026-10-07T06:36:09Z

findings: ["c4s1#1", "c4s1#2", "c4s1#5", "c4s1#6", "c4s1#7", "c4s1#9", "c4s1#10"]
recurrence: 0
report: docs/retros/2026-10-07-c4s1-c4-diagrams-parser.md

run: "the sprint run recorded in report"
branch: sprint/<run-epic>
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "dev-<run-epic>-1 opus, dev-<run-epic>-2 sonnet, dev-<run-epic>-3 haiku, dev-<run-epic>-4 haiku; coordinator, plan, review and retro agents on the session default model; effort unrecorded"
---

## What is wrong, and what it costs

`swarm-plan`'s SKILL.md says nothing about how to record a finding: measured, `grep -n 'retro:' plugin/skills/swarm-plan/SKILL.md` in the swarm source (commit 1cea7e4) exits 1 with no match. The planner therefore wrote all seven of its findings as `retro: <text>` with no kind. Cost: those are 7 of the run's 9 unrouted findings (9 of 29, against 0, 1 and 1 on the three previous runs in this project), so the unrouted count — the taxonomy's health metric — measures one writer's missing instruction rather than drift, and each needed reading by hand to route.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `c4s1#1` — "retro: groom-beads scripts/bd-health.sh extracts epic ids with sed over `bd list --type epic --json`, matching every line shaped `"id": "..."` at any depth, so nested ids in metadata are fed to `bd swarm validate` as epic ids: 29 failures ("epic 'an earlier bead' not found", ...) from 11 epics, exit 1. `bd swarm validate <run-epic>` run directly exits 0, "Swarmable: YES"; the extraction should be `jq -r ".[].id"`."
- `c4s1#2` — "retro: both queued beads carried grooming_status=ready while the blueprint and ADR-030 they cite, and the FRD approval their // AC: citations require, exist only on origin/claude/project-thread-a9wsta (PR #47, unmerged); the run base is origin/main b626128. Neither the work-order DoR nor swarm-scope admission checks that a bead's cited documents exist on the dispatch base, so the gap surfaced only at plan time."
- `c4s1#5` — "retro: (pass 2 plan) all seven review-filed run-introduced beads (<run-epic>.3..9) arrived with no acceptance, no design, null metadata (no tier, no c4_*) and bug repro prose not under a '## Steps to Reproduce' heading, so bd lint flagged all 7; two carried unresolved options in prose (.6 'consider an empty slug a defect', .9 'rename the variant ... or fix the generator'). Grooming closed both (chose: no parse error for empty slug; rename variant, since the generator is the external strata compiler). swarm-review files beads that need a full groom pass every time; filing with a Steps-to-Reproduce heading and a chosen fix direction would make them near-ready."
- `c4s1#6` — "retro: (pass 2 plan) epic-definition-of-ready's dev-pool check counts distinct dev-* assignees over 'bd list --parent <epic> --status all', so closed beads of an already-returned dispatch count against concurrency_cap. Measured here: all-time 3 (dev-<run-epic>-1 holds only closed .1/.2), open-queue 2. Harmless at cap 20, but on a multi-pass run with a small cap it over-reports the concurrent pool; the concurrency question is about open beads."
- `c4s1#7` — "retro: (pass 2 plan) 'swarm ready --run <run-epic>' prints dispatch: "" for every unassigned bead, so it cannot name the next dispatch id before the plan assigns one; the plan derived dev-<run-epic>-2 and -3 by continuing batch 1's numbering. A 'next dispatch id' readout from swarm would remove the derivation."
- `c4s1#9` — "retro: (pass 3 plan) groom-beads scripts/bd-health.sh still exits 1 'hard problems found in: swarm validate' — 30 failures, all from passing epic TITLES (e.g. 'an earlier bead', 'Item 6: planning emits...') to bd swarm validate as ids; the scoped 'bd swarm validate <run-epic>' exits 0 'Swarmable: YES'. Same defect as the pass-1/2 retro comment, recurring every pass: a planner must ignore the health script's verdict and re-run validate by hand, which defeats its 'do not declare ready while it exits non-zero' rule."
- `c4s1#10` — "retro: (pass 3 plan) both review-filed run-introduced beads (.10, .11) again arrived with description only — no design, acceptance or Steps to Reproduce — and .11's fix direction ('neither in row.params nor a namedAs short name') would, read literally, accept $techn= on System/Person, which take no technology. Grooming had to measure both repros with a throwaway probe test (Rel(a,b,"x",$tags="sync",$tags="async") -> Raw=sync, 0 errors; Rel(a,b,"Uses",$tag="contract:missing") -> Tags.Present=false, 0 errors). swarm-review could file run-introduced beads with a measured repro and AC already shaped, since its reviewer has just probed them."

## Proposed change

Add to `swarm-plan` the findings-capture instruction the dispatch skill carries — record each finding as `retro:<kind>[:<target>] <one line>` with evidence, per the findings-capture protocol — and vendor the protocol to it if it is not already. Done when `grep -n 'retro:<kind>' plugin/skills/swarm-plan/SKILL.md` matches.
