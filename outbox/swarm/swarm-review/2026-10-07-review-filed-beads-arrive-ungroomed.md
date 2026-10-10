---
title: "swarm-review files run-introduced beads with design, acceptance, a Steps to Reproduce heading and a chosen fix direction"
description: Beads swarm-review files for admitted findings carry only description prose, so every one needs a full groom pass before it can be dispatched.
slug: issue-review-filed-beads-arrive-ungroomed
audience: skill
target: swarm-review
priority: 2
updated_at: 2026-10-07T06:36:09Z

findings: ["c4s1#5", "c4s1#10"]
recurrence: 1
report: docs/retros/2026-10-07-c4s1-c4-diagrams-parser.md

run: "the sprint run recorded in report"
branch: sprint/<run-epic>
skill_version: swarm 0.1.0 (source commit 1cea7e4)
model: "dev-<run-epic>-1 opus, dev-<run-epic>-2 sonnet, dev-<run-epic>-3 haiku, dev-<run-epic>-4 haiku; coordinator, plan, review and retro agents on the session default model; effort unrecorded"
---

## What is wrong, and what it costs

Run-introduced beads filed by `swarm-review` arrive with a description only: no design, no acceptance criteria, null grooming metadata, repro prose not under a `## Steps to Reproduce` heading, and sometimes unresolved options in the prose. All nine such beads in this run failed `bd lint` and needed a full groom pass, across two review loops, including re-measuring repros with probe tests the reviewer had already effectively run. This recurs an issue already recorded in agent-plugins ("sprint-review should write design and acceptance into the bead's own fields", from run e6xt). Cost: a groom-and-plan pass per review loop spent reconstructing what the reviewer knew.

## Evidence

Finding ids are `<bead>#<ordinal>` in agent-plugins' tracker, `specs-` prefix dropped.

- `c4s1#5` — "retro: (pass 2 plan) all seven review-filed run-introduced beads (<run-epic>.3..9) arrived with no acceptance, no design, null metadata (no tier, no c4_*) and bug repro prose not under a '## Steps to Reproduce' heading, so bd lint flagged all 7; two carried unresolved options in prose (.6 'consider an empty slug a defect', .9 'rename the variant ... or fix the generator'). Grooming closed both (chose: no parse error for empty slug; rename variant, since the generator is the external strata compiler). swarm-review files beads that need a full groom pass every time; filing with a Steps-to-Reproduce heading and a chosen fix direction would make them near-ready."
- `c4s1#10` — "retro: (pass 3 plan) both review-filed run-introduced beads (.10, .11) again arrived with description only — no design, acceptance or Steps to Reproduce — and .11's fix direction ('neither in row.params nor a namedAs short name') would, read literally, accept $techn= on System/Person, which take no technology. Grooming had to measure both repros with a throwaway probe test (Rel(a,b,"x",$tags="sync",$tags="async") -> Raw=sync, 0 errors; Rel(a,b,"Uses",$tag="contract:missing") -> Tags.Present=false, 0 errors). swarm-review could file run-introduced beads with a measured repro and AC already shaped, since its reviewer has just probed them."

## Proposed change

When `swarm-review` files an admitted (run-introduced) finding, it writes the reviewer's repro under a `## Steps to Reproduce` heading, a single chosen fix direction (no open alternatives), and acceptance criteria into the bead's acceptance field. Done when a bead filed by swarm-review passes `bd lint` and the work-order DoR without a groom pass, or the review step names which fields it could not fill.
