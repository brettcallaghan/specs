---
title: "An acceptance test is marked Ready without ever being run against the case that motivated the bead"
description: Where an acceptance criterion states a mechanical completeness test and its motivating case is evaluable at grooming time, run it and record the result before marking the bead Ready
slug: issue-validate-an-acceptance-test-against-its-motivating-case
audience: skill
target: groom-beads
priority: 2
updated_at: 2026-08-24T20:40:00Z

# Evidence — what this issue rests on
findings: ["4tjk#2", "uva5#1"]
recurrence: 0
report: rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md

# Provenance — who and what produced this
run: agent-plugins-so39
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: unversioned — `plugin/skills/groom-beads/SKILL.md` carries no `version` field
model: claude-sonnet-4-5 (`4tjk#2`, dev-so39-3) and claude-opus-5 (`uva5#1`, dev-so39-4); grooming and coordination claude-opus-5. Effort unrecorded — this harness exposes no effort setting
---

## What is wrong, and what it costs

Two beads this run shipped with an acceptance test that had never been run
against the case the bead was written for. `4tjk`'s AC specified
`metadata.{0,40}(replace|entire metadata)` as its completeness test; it found 3
sites and missed a fourth carrying the same false claim in different words, at
`capability-tier.md:11`, outside the 40-character window — found only because
the agent widened the pattern as a sanity check before closing. `uva5`'s DESIGN
proposed two branch tests, and on the disputed case both fire, reproducing the
ambiguity the bead existed to remove. Neither cost this run anything, because
both agents were advanced-tier and both checked; a light-tier agent satisfying
the AC as written would have closed `4tjk` leaving the false claim in the corpus,
which is precisely the failure that bead exists to prevent.

## Evidence

- `4tjk#2` — "retro:stale:agent-plugins-4tjk The bead's own AC grep (metadata.{0,40}(replace|entire metadata)) found exactly 3 sites, but a 4th carried the same false claim in different words: plugin/skills/groom-beads/references/capability-tier.md:11 … It sits outside the 40-char window … the narrow AC pattern itself is what undercounted, not repo drift after grooming."
- `uva5#1` — "retro:decision:agent-plugins-uva5 sharpened the bead's proposed branch wording into two ORDERED tests, because the wording as proposed does not resolve the case it was written to resolve. … Applied to agent-plugins-qozp.2's roster citation, BOTH fire … Two branches that both fire on the disputed case reproduce the ambiguity the bead exists to remove."

## Proposed change

At grooming, **where an acceptance criterion states a mechanical completeness
test — a grep pattern, a count, a branch rule — and the case that motivated the
bead can be evaluated against the tree as it stands, run the test against that
case before marking the bead Ready, and record the result on the bead.** Where
it is a search pattern, state in the AC that the pattern is a floor rather than
a ceiling, so an agent that widens it and finds more is obeying the criterion
rather than exceeding it.

**Name the applicability, so the exemption is visible rather than assumed.**
The check is only available when both conditions hold:

1. **The criterion is mechanical.** A grep, a count, a branch rule — something
   with an answer. A criterion resting on judgement ("the rationale survives")
   has nothing to run, and grooming should not pretend otherwise.
2. **The motivating case is reachable now.** The case exists in the tree at
   grooming time and the test can be evaluated against it without the bead's own
   change having landed. Both cases above satisfy this: `4tjk`'s false-claim
   sites and `uva5`'s disputed citation were both present before either bead ran.
   An AC that tests a state only the change creates cannot be validated this way.

Where either condition fails, record *that* on the bead — one line naming which
condition failed — rather than silently skipping the step. Done: a bead whose AC
carries a mechanical test either carries the result of running it against its
motivating case, or carries the reason it could not be run.
