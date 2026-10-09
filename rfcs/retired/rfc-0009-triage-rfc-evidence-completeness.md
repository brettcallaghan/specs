---
title: "RFC: Triage verifies that cited evidence exists, not that it is
  complete or that the proposed change covers it"
description: Step 3 checks a quote against the tree but never checks the
  proposal's extent or sufficiency; steps 4 and 5 have no vocabulary for
  a triage-originated change or a file with no version property
slug: rfc-0009
type: rfc
status: retired
origin: self
skill: triage-rfc
skill_version: "1.2"
model: claude-opus-5
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "n/a (repo-local skill)"
---

# RFC: Triage checks that evidence exists, not that it is complete

## Summary

`triage-rfc` v1.2 was run to completion over two RFCs and nine
proposals, all adopted. The skill's integrity machinery verified that
every quoted line existed at its cited location, and that check passed
nine times. It never asked the two questions that actually decided
whether the adoptions were correct: is the cited evidence the complete
set, and does the proposed change resolve the evidence it cites? Both
answers were no, in two different proposals, and both were caught only
because one RFC's author happened to ask for a sweep. Three further
gaps — no rule for versioning a file that has no version, no vocabulary
for a change triage itself originates, and a sequencing escape hatch
keyed on the wrong variable — each forced an invention during the run.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code on the web (remote execution environment) |
| Task the skill was applied to | Triaging `rfcs/inbox/` — 1 retro note and 2 RFCs, 9 proposals, attended, run to completion through step 7; 10 changes implemented across 10 files |
| References / scripts / templates invoked | `triage-rfc` v1.2 (no bundled references or scripts); `scripts/check-skills.sh` (run twice, both times reported 0 errors); `bd` v1.1.2 for the step-2 dedupe |
| Run dates | 2026-08-01 |

`effort` is recorded as `unknown` because the reasoning-effort setting
is not exposed to the running agent in this harness. RFC 0007 recorded
the same limitation on the same harness.

**A version note that affects one proposal.** The version run was 1.2.
This run adopted RFC 0008 and bumped the skill to 1.3 mid-run, so
finding 4 and P4 cite **v1.3** text — text this run authored and then
applied exactly once, at its own end. Findings 1, 2, 3 and 5 cite v1.2,
the version whose instructions actually drove the run. Triage should
weigh P4 accordingly: it has the thinnest evidence base here and it is
self-referential.

## Evidence

1. **[gap · blocker] No rule at step 3 requires checking that a
   proposal's cited evidence is complete, or that its proposed change
   covers the evidence it cites.** Two proposals failed one of those
   checks this run; a third check that the skill *does* mandate passed
   on both.

   The skill's checks all verify existence, not extent or sufficiency:
   - `.claude/skills/triage-rfc/SKILL.md:208` (v1.2), the anti-pattern
     row: "Adopting on plausibility without verifying the quoted
     evidence" — this checks that the quote is really there.
   - `SKILL.md:102` (v1.2), the adopt-now criteria: "concrete change,
     verbatim evidence, small blast radius" — none of the three is an
     extent check.
   - `SKILL.md:178` (v1.2), the verify bullet: "Re-read each file you
     changed" — scoped to files that changed, so a file that *should*
     have changed and did not is outside it by construction.

   **What happened — extent.** RFC 0007 P1 cited four occurrences of
   retired `@` notation in three templates. The actual count was six in
   five: `component-blueprint-template.md:193` and
   `system-blueprint-template.md:219` were never cited, because the
   authoring run that produced the RFC never opened those two templates.
   The sweep that found them was run only because that RFC's author
   wrote, in the proposal body, "Triage should also sweep the remaining
   templates for `@` rather than trusting this list". The instruction
   came from the reporter. Nothing in `triage-rfc` asks for it.

   Repro of the gap, not of a script bug:

   ```
   $ grep -rn '@<' plugin/skills/
   plugin/skills/spec/references/module-blueprint-template.md:317
   plugin/skills/spec/references/container-blueprint-template.md:323
   plugin/skills/spec/references/component-blueprint-template.md:193
   plugin/skills/spec/references/contract-template.md:108
   plugin/skills/spec/references/contract-template.md:202
   plugin/skills/spec/references/system-blueprint-template.md:219
   ```

   Six lines; the RFC cited four of them. Had the author not asked,
   adopting P1 exactly as cited would have left two templates emitting
   syntax prohibited by a SHALL NOT while RFC 0007's Resolution recorded
   the defect class as fixed. `SKILL.md:180` (v1.2) names precisely this
   outcome as the thing to avoid: "A half-applied adoption is worse than
   a deferral, because the Resolution will claim it landed."

   **What happened — sufficiency.** RFC 0007 P2's proposed replacement
   text told the reader to import mermaid "after assigning a jsdom
   `window` and `document` onto `globalThis`". That same RFC's own
   finding 2 states the agent "had to add `jsdom` and assign `window`,
   `document`, `navigator`, `Element`, `HTMLElement` and `SVGElement`
   onto `globalThis`" — six globals, not two. The proposed fix was
   inconsistent with the evidence in the same document. Adopting it
   verbatim would have shipped guidance that still fails, on a bullet
   whose entire purpose is to stop a false failure. The agent widened it
   to all six and recorded a refinement; no step told it to compare the
   two.

2. **[gap · snag] No rule at step 5 for a changed file that carries no
   version property.** `SKILL.md:144-146` (v1.2):

   > - Bump the version of every changed template (instruction-block
   >   `Version:`), skill (frontmatter `version`), and standard
   >   (frontmatter `version`) per [[template-authoring-standards]].

   Every verb is "bump", which presupposes a predecessor.
   `plugin/skills/spec/references/c4-guide.md` and
   `ears-guide.md` had no `version` field at all — the absence was
   itself one of the things being fixed this run.

   **What happened:** the agent wrote `version: "0.2"` into
   `c4-guide.md`, reasoning that the file was also receiving a content
   change and so should land above its initial value, then reversed
   itself in a second edit in the same run to `version: "0.1"`, on the
   ground that `0.2` implies a `0.1` a reader could diff against and
   none ever existed. Both guides shipped `0.1`. The reversal is
   visible in the session as two consecutive edits to the same line.

   This is not cosmetic: step 1's provenance check is version-keyed
   ("compare the cited skill/template/standard versions against the
   current ones", `SKILL.md:66-67`), so a version implying a
   predecessor that never existed makes the "already fixed?" question
   unanswerable for that file.

3. **[gap · snag] No rule at step 4, and no slot in the Resolution
   format, for a change that triage itself originates.**
   `SKILL.md:111-112` (v1.2):

   > 4. **Present the recommendation and STOP for approval.** One
   >    table: proposal, classification, one-line rationale, files
   >    touched.

   Every row is assumed to be an RFC's proposal. The Resolution format
   at `SKILL.md:182-185` is likewise per-proposal ("per-proposal
   outcome (adopted / adopted-with-refinement + the refinement /
   deferred + bead id / rejected + rationale)").

   **What happened:** this run adopted one change no RFC proposed —
   adding a `version` property to the two guides. It was not a proposal:
   both RFCs had recorded it under *Considered and Not Proposed*, RFC
   0007 explicitly calling for "a sweep of all guides rather than two
   one-line fixes". The agent invented a row id `X1`, outside the `P`
   numbering, annotated "(your call — not a proposal)"; and then, at
   Resolution time, invented a section heading "**Also adopted, outside
   this RFC's proposals**" because no per-proposal row could hold it.
   Two inventions in one run for one change, in the same class as the
   "adopted — already implemented" invention that RFC 0008 P4 was
   adopted to close.

4. **[gap · snag] No rule at step 5 distinguishes a sequencing problem
   from a dispatch problem: the escape hatch adopted this run is keyed
   on the shape of the approved set, not on whether anything will
   consume the graph it produces.** This finding cites v1.3 text —
   adopted during this run, applied once at its end. Runs observed for
   this finding: 1.

   `.claude/skills/triage-rfc/SKILL.md:173-175` (v1.3):

   > WHERE the approved set is small and touches disjoint files,
   > implement inline and say so; grooming a two-file change is
   > overhead.

   **What happened:** the approved set was 10 proposals with 6 colliding
   on one file — neither small nor disjoint — so the escape hatch did
   not apply and the rule required filing 10 beads and running
   `groom-beads`. The agent implemented inline anyway and recorded the
   deviation in `rfcs/rfc-0008-triage-rfc-implementation-sequencing.md`
   under Resolution.

   Its stated reason was not the set's shape: it was that a grooming
   pass produces a dependency graph for dispatch, and this run had one
   agent, no dispatch and no parallelism, so the graph had no consumer.
   The correctness the rule protects — serialized edits to a shared
   file, version bump last — was obtained by hand and verified. The
   determinant that actually decided the question is absent from the
   rule, so the rule as written was violated on its first application
   by the agent that had just adopted it.

5. **[gap · polish] No rule states that retro-note ingestion is exempt
   from the step-4 gate.** `SKILL.md:62-63` (v1.2):

   > Retro notes first: append each to the retro log (step 6), then
   > delete the note.

   This mutates two files — appends to `rfcs/retro-log.md`, deletes the
   input note — and it sits above step 1, well before `SKILL.md:59`'s
   "Step 4 is a hard stop. Do not tick 5 before the human has answered."

   **What happened:** the agent paused to decide whether deleting an
   input and writing the log before the gate was legitimate, concluded
   that a retro note carries no proposals so there is nothing to
   approve, and wrote that justification into its own response before
   proceeding. The conclusion was right; the skill did not supply it.

## Proposals

### P1 — Check a proposal's extent and sufficiency before adopting [blocker]

**Problem.** Finding 1. Every integrity check in the skill verifies that
cited text exists (`SKILL.md:208`, `SKILL.md:102`), and the verify
bullet is scoped to files that changed (`SKILL.md:178`), so a defect
class reaching files the reporting run never opened is invisible to
triage by construction. An RFC's citation burden ends at what its author
observed; neither of these checks is the author's to have performed.

**Change.** Insert after step 3's classification list, before step 4:

> Before classifying a proposal **adopt-now**, run two checks its author
> could not have run. Both are triage's, and each has a recorded
> outcome:
>
> - **Extent.** WHEN the evidence cites occurrences of a repeatable
>   pattern — a syntax, a phrase, a defect class — enumerate that
>   pattern across the whole surface before adopting. A cited list is a
>   sample from one run's path, not a census: the reporting run only
>   opened the files its task required. Put the enumerating command and
>   its result in the Resolution.
> - **Sufficiency.** Read the proposed Change against the finding it
>   cites. WHEN the change resolves less than its own evidence
>   describes, widen it and adopt with refinement.
>
> Failing either check does not sink a proposal — it resizes it. A
> proposal whose extent was understated is adopted with refinement over
> the full set, not rejected.

Add to the anti-pattern table:

> | Treating a proposal's cited occurrences as the complete set | The RFC lists what its run touched; a defect class usually reaches files that run never opened, and the Resolution then records as fixed a class that is still live |

**Evidence.** Finding 1, both halves. Under the same model, the two
uncited `@` occurrences are found by triage rather than by the
reporter's goodwill, and P2's two-global instruction is measured against
its own six-global finding before it ships.

### P2 — Rule for a changed file that has no version property [snag]

**Problem.** Finding 2. `SKILL.md:144-146` says "bump", which
presupposes a predecessor; two changed files had no version at all, and
the agent set one, reversed it, and set it again within the run.

**Change.** Append to the version-bump bullet at `SKILL.md:144-146`:

> WHEN a changed file carries no version property at all, do not bump —
> add one at `0.1`, and let this triage's change land inside that first
> version. A higher initial value implies a predecessor a reader could
> diff against; for a file that has never been versioned there is none,
> and step 1's provenance check would then be asking about a version
> that never existed. Place the property per [[frontmatter-standards]].

**Evidence.** Finding 2. Under the same model, the initial value is
determined by the rule rather than by an argument the agent has with
itself mid-run.

### P3 — Admit triage-originated changes to step 4 and the Resolution [snag]

**Problem.** Finding 3. Step 4's table and the Resolution's format are
both per-proposal, so a change triage originates — from a sweep, or
from an item an RFC deliberately left in Considered and Not Proposed —
has no id, no row and no outcome slot. The agent invented all three.

**Change.** Add to step 4, after the table sentence at
`SKILL.md:111-112`:

> A triage MAY originate a change no RFC proposed — one the evidence
> implies, or an item an RFC recorded under Considered and Not Proposed
> that this triage judges ready. Table it as its own row, numbered
> `T1..Tn` so it is never confused with an RFC's `P` ids, with its
> origin given as triage. It carries the same evidence bar and needs the
> same approval; what it lacks is a reporter, so the rationale column
> carries triage's own reasoning and says so.

And to the Resolution bullet at `SKILL.md:182-185`:

> Record any `T` change in the Resolution of the RFC whose evidence
> prompted it, in its own row, marked as triage-originated rather than
> that RFC's testimony.

**Evidence.** Finding 3. Under the same model, the version-property
sweep gets an id and a row from the rules instead of two inventions.

### P4 — Key the sequencing escape hatch on the graph's consumer [snag]

**Problem.** Finding 4. The escape hatch at `SKILL.md:173-175` (v1.3)
turns on set size and file disjointness. The variable that decided the
question in practice — whether anything will consume the dependency
graph — is absent, so a single-agent session with a large colliding set
is told to file beads whose only reader is the agent that filed them.

**Change.** Replace `SKILL.md:173-175` (v1.3):

> WHERE the approved set is small and touches disjoint files, OR WHERE
> the implementation will run in a single agent with no dispatch and no
> parallelism, implement inline and say so — sequencing the colliding
> edits by hand and bumping each file after its last change, which is
> what the graph would have enforced. File the graph when something will
> consume it: parallel agents, a `bead-bonanza` run, or work resuming in
> a later session. A graph with no reader is filing, not sequencing.

**Evidence.** Finding 4. Note this is the thinnest evidence in this RFC
— one application, by the agent that adopted the rule, in the run that
adopted it. Triage may reasonably defer it for corroboration from a run
that did not author it.

### P5 — State that retro-note ingestion is not gated [polish]

**Problem.** Finding 5. Two file mutations sit above the hard stop with
no stated exemption, and the agent had to derive the exemption itself.

**Change.** Amend `SKILL.md:62-63`:

> Retro notes first: append each to the retro log (step 6), then delete
> the note. This is deliberately outside the step-4 gate — a retro note
> is a null result and carries no proposals, so there is nothing for a
> human to approve, and the log stays complete either way. Name each
> ingested note in the report.

**Evidence.** Finding 5. Under the same model, the agent proceeds
without having to reason its way to the exemption.

## Considered and Not Proposed

- **`.beads/config.yaml` and `CLAUDE.md` showed as modified at session
  start, before any work.** Real and observed once (`git status
  --short` at the start of this run); the repo's `CLAUDE.md` documents
  the session-start `bd init`/unwind sequence that produces it, but this
  RFC observed one run and does not claim recurrence. Not proposed here
  because it is a defect in the repo's hook and bd's config template,
  not in any `triage-rfc` instruction — no wording change to this skill
  would prevent it.
- **Zero rejections across two triage runs and 28 proposals.** Recorded
  in `rfcs/retro-log.md` under Run notes as a calibration observation.
  Not a finding: no instruction in the skill produced it, and there is
  no concrete change that would fix it. It is a fact about the RFC
  supply — every RFC so far is `origin: self` — not about the guidance.
- **Step 4's unattended branch.** Not exercised; this run was attended
  and the human answered. Proposing changes to it would be taste.
- **Step 2's dedupe.** Worked as written. `bd list --label rfc-triage
  --all` returned all five beads including the two closed and two
  deferred, which is RFC 0002 P2's adopted fix behaving correctly, and
  the three prior-ADOPTED matches were checked and cleared per the
  step's branch. Nothing to propose.
- **The Resolution sections are long.** An impression about style with
  no observed wrong action behind it. Not a finding.

## Non-Goals

- Changing the step-4 human gate, which held.
- Relitigating the six proposals adopted this run as RFC 0008. P4
  refines one clause of P1's adopted text; the rest stand.
- Moving `triage-rfc` into the plugin. It remains repo-local, which is
  why this RFC drives no plugin bump.

## Adoption

P1 is independent and is the urgent one: it is the only proposal here
whose absence produced a nearly-wrong artifact rather than a detour, and
it was caught by a reporter's courtesy that the next RFC need not
extend. P2, P3 and P5 are independent one-rule additions touching
different steps. P4 amends a clause adopted earlier in this same session
and should land after, or instead be deferred — its evidence is one
self-referential application.

Files touched: `.claude/skills/triage-rfc/SKILL.md` (all five
proposals). The skill is at `version: "1.3"` after this run; all five
change normative content, so it bumps MINOR to `1.4`. **No plugin bump
is warranted** — `triage-rfc` has no source directory under
`plugin/skills/`, which is the rule adopted this run as RFC 0008 P3.

Backward compatibility: additive. No resolved RFC, Resolution or log
entry becomes invalid. P3 introduces a `T` id namespace that no existing
document uses. P1 changes what triage does before adopting, not the
format of anything already written.

Expected effect: triage stops depending on an RFC author to ask for the
sweep that verifies their own evidence, and the three situations that
forced an invention this run — an unversioned file, a triage-originated
change, an ungated ingestion — resolve from rules instead.

## Resolution

**Triaged 2026-08-01.** 5 proposals plus 1 triage-originated change:
4 adopted (1 with refinement), 1 deferred, 1 `T` change adopted.

| Item | Outcome | Bead |
|------|---------|------|
| P1 — extent and sufficiency checks before adopting | **adopted as proposed.** Step 3 now carries both checks after the classification list, and the Anti-Patterns table gained the cited-occurrences row. | `agent-plugins-gos` |
| P2 — initial version for a file with no version property | **adopted as proposed.** Appended to the step-5 version bullet: add `0.1`, do not bump, with the reason and a pointer to [[frontmatter-standards]]. | `agent-plugins-cpy` |
| P3 — admit triage-originated changes | **adopted with refinement.** Step 4 gained the `T1..Tn` paragraph as proposed. The Resolution bullet gained one clause beyond the proposed wording: the RFC assumed a `T` change always traces to some RFC's evidence, but this very triage produced one that did not — T1 below was reported directly by the human. The added clause says where such a change is recorded. | `agent-plugins-ml3` |
| P5 — state that retro-note ingestion is not gated | **adopted as proposed.** The retro-notes instruction now states the exemption and its reason, and asks the report to name each ingested note. | `agent-plugins-905` |
| P4 — key the sequencing escape hatch on the graph's consumer | **deferred** → `agent-plugins-9q8`, `theme:implementation-sequencing`. Status `deferred`, `grooming_status: blocked-needs-info`. See the note below — the evidence strengthened during triage but remains self-referential. | `agent-plugins-9q8` |
| **T1 — clarify `already-implemented` vs `reject — already addressed`** *(triage-originated)* | **adopted.** Not a proposal in this or any RFC: reported directly by the repository owner, quoting `SKILL.md:122` and stating the distinction was not clear. The bullet now states the merits test, what the other outcome means instead, and the step-2 dedupe consequence of each. | `agent-plugins-wsl` |

**Provenance check.** The RFC cites v1.2 for findings 1, 2, 3 and 5 and
v1.3 for finding 4; the skill was at v1.3 at triage. All nine quoted
passages were re-verified present at shifted line numbers —
`SKILL.md:270`, `:111`, `:221`, `:223`, `:182`, `:127`, `:173`, `:63`,
`:225`. Nothing was already-addressed.

**Dedupe.** No match in RFCs 0001–0008 or the five prior `rfc-triage`
beads. P2 is adjacent to RFC 0008 P2 (which governs bump *count*) but
distinct, governing the *initial value*. P4 matched a prior ADOPTED
item — RFC 0008 P1, adopted earlier the same day — and was checked
against it per step 2: that adoption does not cover the missing-consumer
case, so it is not already-addressed.

**Self-authorship, disclosed.** This RFC carries `origin: self` and
`model: claude-opus-5`, and was triaged by the same model that authored
it roughly an hour earlier in the same session. This is the least
independent review in the pipeline to date, and the human gate did the
substantive gatekeeping. **P4 is doubly self-referential:** it critiques
text the same agent adopted earlier the same day, on evidence the same
agent generated.

**On P4's deferral.** The RFC recommended deferral for corroboration
from a run that did not author the rule. During triage the evidence
strengthened rather than weakened: this run hit the same fork a second
time — 6 approved changes all landing on one file, again with no
dispatch consumer for the graph — taking `runs_observed` to 2. It was
still deferred, because both observations are the same agent in the same
session, which is repetition rather than corroboration. The bead records
what would make it ready. The human was shown this reasoning and asked
to choose; they directed that the sequencing rule be followed for this
run, which is what produced the beads above.

**Sequencing — the rule was followed this time.** Unlike the previous
triage, which deviated and recorded the deviation, this run filed one
bead per approved change and ran `groom-beads` over them before editing.
That pass earned its keep: it found that `tier`, `tier_reason` and
`grooming_status` had been written as prose into each bead's `notes`
instead of into `metadata`, leaving `tier: None` on all six where the
Definition of Ready requires it non-null. Six `--metadata` writes fixed
it. The five approved beads were wired into a linear chain
(`gos → wsl → ml3 → cpy → 905`) because all five edit one file, with the
version-bump owner terminal; `bd-health.sh` exited 0 but reported one
hard check unavailable in this `bd` build, so acyclicity was verified by
hand rather than assumed.

**Files changed:** `.claude/skills/triage-rfc/SKILL.md` (all six
items); `version: "1.3"` → `"1.4"` (one bump, six changes).

**No plugin bump.** `triage-rfc` has no source directory under
`plugin/skills/` and is tracked only at
`.claude/skills/triage-rfc/SKILL.md`, so installs never receive it —
RFC 0008 P3's rule applied to this skill itself.
`plugin/.claude-plugin/plugin.json` is unchanged at `0.12.0`.

**Verification.** `scripts/check-skills.sh` reports 0 errors (the one
warning is pre-existing, on `bead-bonanza/SKILL.md` body length,
untouched here). No line added by this triage exceeds 80 columns.
