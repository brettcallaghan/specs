---
title: "Severity cannot override the scope gate, so a run ships its own regression"
description: Add a classification row putting a P0 defect introduced by the run's own merged diff in scope regardless of the scope-growth gate
slug: issue-severity-cannot-override-the-scope-gate
audience: skill
target: sprint-review
priority: 1
updated_at: 2026-08-24T20:40:00Z

# Evidence — what this issue rests on
findings: ["qozp#9 (context)", "so39#11 (same class, run 2)"]
recurrence: 1
report: rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md   # run 2; run 1's findings (`qozp#9`) are in rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md

# Provenance — who and what produced this
run: agent-plugins-so39 (recurrence; first reported from agent-plugins-qozp)
branch: claude/bead-bonanza-design-review-ne1plu
skill_version: "0.1"
model: claude-opus-5 (coordinator, both runs; effort unrecorded — this harness exposes no effort setting)
---

## What is wrong, and what it costs

Section 4's in-scope test requires **both** a defect in what the run changed
**and** gate authorization for scope growth. `admit_review_findings` resolved
to `false` from the skill default, so the second condition failed for **every
finding regardless of severity**, and the run merged and shipped `872f390` with
its review finding declined unread on the merits.

**The motivating example did not survive verification, and the structural
defect did.** Review filed that finding as a P0 silent-data-loss path on the
premise that `bd update --metadata` replaces the object; measured on bd 1.2.2
it merges at the top level, so no data was lost and the bead was downgraded to
P2. What remains is the shape of the decision: the gate refused the finding on
*scope*, having never weighed *severity*, so a genuine P0 in the run's own diff
would have been declined by exactly the same path, for exactly the same reason.
The gate is right to refuse new work; it should not be the mechanism by which a
run ships a defect it created — and this run demonstrated that mechanism
operating, with the severity question never reached.

**Recurred in run 2 (`agent-plugins-so39`), with the example run 1 lacked.**
Review pass 1 filed **10** findings and admitted **zero**.
`standing_auth.admit_review_findings` was again `{value:false,
source:"default"}` — nobody chose it, in either run. Three of the ten
(`agent-plugins-so39.4`, `.5`, `.6`) are defects **in the corrections this run
shipped**, on the exact thesis the run was opened to serve. `so39.5` is the
sharpest: the run shipped the false universal "`--set-metadata` stores every
value as a string" *inside the paragraph that fixed the previous false claim*,
and the coordinator confirmed the falsity by independent measurement — seven
`--set-metadata` pairs on bd 1.2.2, scalars keeping their JSON type. It is
deferred, unread on the merits, behind a *scope* decision that never reached the
severity question. Run 1's motivating example evaporated under verification;
run 2's did not.

## Evidence

- `qozp#9 (context)` — "retro:tooling:sprint-wave — sprint-wave could not dispatch a subagent: no Task/Agent tool was exposed to it, so it created the worktree as orchestrator and ran sprint-batch -> run-bead itself inside it."
- `so39#11 (same class, run 2)` — "retro:tooling:bd incident — 'bd search' is NOT non-functional; it EXCLUDES CLOSED ISSUES BY DEFAULT … A single negative observation was carried into a universal without the flag being checked. Third instance this run of the failure the run was opened to stop."
- *(no finding id)* — `agent-plugins-so39` comment 14, the review-loop incident, quoted in full in run 2's findings document section 7: "review pass 1 filed 10 findings, admitted ZERO, and three of them (so39.4, so39.5, so39.6) are defects IN THE CORRECTIONS THIS RUN SHIPPED — on the exact thesis the run was opened to serve. standing_auth.admit_review_findings is {value:false, source:'default'}: nobody chose it, and the skill's in-scope test needs gate authorization for scope growth, so severity cannot admit them."
- *(no finding id)* — `agent-plugins-so39.5` comment 1, the coordinator's independent measurement: `num=5` → number, `b=true` → boolean, `f=1.5` → number, `n=null` → null, `s=hello` → string, `arr=["a","b"]` → string, `obj={"k":1}` → string. bd 1.2.2, throwaway database.

## Proposed change

Add a third row to section 4's classification table: a defect **the run's own
merged diff introduced**, at P0, is in scope regardless of the gate — the run
either fixes it or does not merge the commit that carries it. Leave every other
severity and every pre-existing defect on the conservative default. Done: a P0
regression attributable to a bead in `manifest.admitted` classifies in-scope
with `admit_review_findings: false`.

**What the recurrence adds: a narrower and more defensible trigger than raw
severity.** Run 2's case is not merely severe — the defect was introduced by the
run's own merged diff *and* falsifies the run's own stated purpose. That
conjunction is a tighter test than "P0", it is mechanically checkable (the
finding's origin bead is in `manifest.admitted`, and the run epic states what the
run is for), and it admits strictly fewer findings than a severity threshold
would. Write the row on that conjunction rather than on severity alone, so the
gate keeps refusing ordinary scope growth.
