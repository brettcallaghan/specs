---
title: "RFC: Sequencing at step 5, and the gaps steps 5-7 met on their
  first complete run"
description: Triage's implement step is silent on colliding proposals and
  never names the grooming skills; its publish and log steps met four
  situations their first execution produced
slug: rfc-0008
type: rfc
status: retired
origin: self
skill: triage-rfc
skill_version: "1.0"
model: claude-opus-5
effort: high
runs_observed: 2
severity: blocker
plugin_version: "0.9.0"
---

# RFC: Sequencing at step 5, and the gaps steps 5-7 met on first run

## Summary

`triage-rfc` step 5 says "Implement the approved adopt-now items" and
then describes what to do to *a file*. It says nothing about what to do
when fourteen approved proposals land on seven files, four of them on one
file — no ordering rule, no collision rule, and no mention of the two
skills this repository already owns for exactly that problem
(`groom-beads`, `bead-bonanza`). The gap is not theoretical: the human
had to supply the missing step by hand, in both runs observed.

Steps 5 through 7 also executed end-to-end for the first time in run 2.
Run 1 stopped at the step-4 gate, so every rule after it was unexercised
text. Its first real execution immediately produced four situations none
of those steps covers: a version bump owed once but claimed by four
proposals, a changed skill that does not ship in the plugin, a proposal
approved but already implemented, and a log file that did not exist.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code, remote (web) session |
| Task the skill was applied to | Triage of 5 RFCs / 19 proposals in `rfcs/inbox/`, attended, run to completion through step 7 |
| References / scripts / templates invoked | `triage-rfc` v1.0 (no bundled references or scripts); `groom-beads` v1.0 and `scripts/check-skills.sh` were reached for outside the skill's instruction |
| Run dates | 2026-07-30 (run 1, stopped at step 4), 2026-07-31 (run 2, complete) |
| Runs observed | 2. Finding 1 is evidenced by both runs; findings 2-6 by run 2 only, because steps 5-7 did not execute in run 1 |

Version cited is `1.0`, the version run. The skill was bumped to `1.1`
during run 2 by the adoptions that run implemented; that bump is
unrelated to the findings below.

## Evidence

1. **[gap · blocker] No rule at step 5 for approved proposals that
   collide on the same file, and no route to the grooming skills.**
   `.claude/skills/triage-rfc/SKILL.md:96-98` (v1.0) reads:

   > 5. **Implement** the approved adopt-now items:
   >    - Apply the changes; keep every touched file within its own
   >      conventions (80-column lines, EARS for normative statements).

   Every subsequent bullet is likewise phrased against a single change.
   Nothing in the step addresses ordering, batching, or two proposals
   editing one file, and the words `groom-beads` and `bead-bonanza` do
   not appear anywhere in the skill.

   Run 2 approved 14 proposals distributed as: 4 on
   `.claude/skills/triage-rfc/SKILL.md`, 3 on
   `plugin/skills/bead-bonanza/SKILL.md`, 2 each on
   `plugin/skills/groom-beads/scripts/bd-health.sh`,
   `plugin/skills/groom-beads/SKILL.md` and
   `plugin/skills/retro-skill/SKILL.md`, and 1 on
   `plugin/skills/bead-bonanza/scripts/bb-progress.sh`. Following step 5
   literally, the agent began implementing inline in classification
   order. The human interrupted with the parenthetical "think carefully
   about sequencing before you start (ahem … groom)", which is what
   caused a `groom-beads` pass to happen at all.

   **Recurrence is observed, not assumed.** After run 1 the same human
   wrote: "I just noticed the triage-rfc skill will do implementation
   but it doesn't use `groom-beads` or `bead-bonanza`! We need to
   improve this skill." Two runs, two human interventions supplying the
   same missing step.

   The grooming pass that resulted was not ceremonial. It found four
   defects in the beads the triage step had just filed — including that
   no proposal owned the single `plugin.json` bump, and that two beads
   owning a skill's version bump were positioned before other beads
   changing the same skill. Both would have produced a wrong result
   under inline implementation.

2. **[gap · snag] Step 5 does not say a file's version is bumped once,
   after every proposal touching it has landed.**
   `.claude/skills/triage-rfc/SKILL.md:99-101` (v1.0):

   > - Bump the version of every changed template (instruction-block
   >   `Version:`), skill (frontmatter `version`), and standard
   >   (frontmatter `version`) per [[template-authoring-standards]].

   Read per-proposal — which is the grain step 3 explicitly establishes
   ("Classify each proposal (per proposal, not per RFC)",
   `SKILL.md:82`) — this instructs a bump per approved proposal. Four
   proposals changed `triage-rfc/SKILL.md` in run 2, which literally
   read would advance one skill four minor versions for one release.
   The anti-pattern table at `SKILL.md:141` warns only in the opposite
   direction: "Forgetting version bumps on changed files".

   Observed consequence: the beads filed at step 5 each carried a
   "version bumped" acceptance box, and the grooming pass had to
   reassign bump ownership so that the owning bead was last in its
   file's chain. Nothing in step 5 asks for that.

3. **[misleading-text · snag] Step 7 treats "any skill changed" as
   grounds for a plugin bump, but not every skill ships in the plugin.**
   `.claude/skills/triage-rfc/SKILL.md:124-126` (v1.0):

   > 7. **Publish.** WHEN any skill, template, or standard changed, bump
   >    the plugin version in `plugin/.claude-plugin/plugin.json` (MINOR
   >    for normative changes). Distribution is plugin-only …

   `triage-rfc` is itself a skill and it is NOT in the plugin: it lives
   at `.claude/skills/triage-rfc/SKILL.md`, is tracked in git, and has
   no source under `plugin/skills/`. Verified during run 2:

   ```
   $ ls plugin/skills/triage-rfc/
   ls: cannot access 'plugin/skills/triage-rfc/': No such file or directory
   $ git ls-files .claude/skills/
   .claude/skills/triage-rfc/SKILL.md
   ```

   (The other six directories under `.claude/skills/` are gitignored
   dogfood copies generated from `plugin/skills/`.) Run 2 changed
   `triage-rfc` and, read literally, step 7 requires a plugin bump for
   it — announcing to every install a change they will never receive.
   The agent had to decide unaided that the repo-local skill does not
   count; the run's bump was justified by the three plugin skills that
   did change.

4. **[gap · snag] No classification exists for a proposal that was
   approved and is already implemented.** Step 3
   (`SKILL.md:82-92`, v1.0) offers exactly three outcomes: `adopt-now`,
   `defer`, `reject`. Step 1 supplies one adjacent phrasing,
   `SKILL.md:65-66`: a proposal against already-fixed text "is resolved
   `rejected — already addressed in <file> v<X.Y>` without further
   debate."

   Run 2's RFC 0001 (`spec`) P1-P3 fit none of them. They were sound,
   the human approved them, and they were already present in the tree —
   implemented in an earlier session, not by this triage. `adopt-now` is
   wrong (nothing to implement); `reject` is wrong and actively
   misleading, because the Resolution would then read as though the
   proposals had failed. The agent invented the phrase "adopted —
   already implemented" and flagged the invention inside the Resolution
   it wrote. A vocabulary an agent has to extend mid-run is a gap in the
   vocabulary.

5. **[gap · snag] No rule for an RFC the triaging agent authored.** All
   five RFCs in run 2 carried `origin: self` and were written by the
   same model now deciding them. The sharpest case: RFC 0005 was
   authored about that agent's own fabricated claim, and its P2
   sanctions the very routing decision the RFC itself used — adopting
   it retroactively legitimises the author's own choice.

   The skill has no rule here. Its integrity machinery all points
   outward at the RFC's author: "Adopting on plausibility without
   verifying the quoted evidence" and "Editing a received RFC's Evidence
   or Proposals" (`SKILL.md:138`, `SKILL.md:142`) both assume the
   reporter and the triager are different parties. The agent recorded
   the conflict voluntarily in RFC 0005's Resolution; nothing required
   it, and nothing would have surfaced its absence.

6. **[gap · snag] Step 6 assumes `rfcs/retro-log.md` exists and
   specifies no format.** `.claude/skills/triage-rfc/SKILL.md:120-123`
   (v1.0):

   > 6. **Log.** Append one line per ingested item to
   >    `rfcs/retro-log.md`: date, origin project, phase retroed, item
   >    (RFC number or "retro note"), outcome (counts
   >    adopted/deferred/rejected, or "null retro").

   The file did not exist — run 2 was this repository's first triage, so
   there was nothing to append to. Five named fields do not determine a
   format: the agent chose a markdown table plus a "Run notes" prose
   section, and invented column headers. The next run has the same
   latitude and no reason to make the same choices, which is how a log
   whose entire purpose is longitudinal comparison becomes
   heterogeneous by its second entry.

## Proposals

### P1 — Route step 5 through grooming when approvals collide [blocker]

**Problem.** Finding 1. Step 5 describes changing a file and is silent
on changing many, in an order. The repository already owns the two
skills that solve this and the skill names neither, so the human has
supplied the missing step in both runs observed.

**Change.** Insert after `SKILL.md:98` (v1.0), before the version-bump
bullet:

> - **Sequence before you edit.** Count the approved proposals and the
>   files they touch. WHEN more than one approved proposal changes the
>   same file, or more than about five are approved in total, do not
>   implement inline: file one bead per approved proposal (to the same
>   bar as a deferral bead, below) and run `groom-beads` over them to
>   wire the dependency graph. Two proposals editing one file SHALL be
>   serialized by a real edge; the proposal that owns that file's
>   version bump SHALL be last in its chain. Implement in the resulting
>   wave order.
>
>   The collision key is the FILE, not the RFC — proposals from
>   different RFCs routinely land in the same file, and two beads with
>   no dependency between them are still a collision when they edit the
>   same lines.
>
>   WHERE the approved set is small and touches disjoint files, implement
>   inline and say so; grooming a two-file change is overhead.
>
>   `bead-bonanza` MAY then execute the graph when the queue is large
>   enough to warrant orchestration, but this step does not require it —
>   the sequencing is the point, not the dispatch.

**Evidence.** Finding 1. Under this rule run 2's 14 approvals reach
grooming without a human prompt. The rule is conditional rather than
unconditional so that a two-proposal triage does not pay for machinery
it does not need — which is the objection that would otherwise sink it.

### P2 — Bump each changed file's version exactly once [snag]

**Problem.** Finding 2. The bump bullet reads per-proposal because step 3
sets the grain at the proposal. Four proposals on one file then imply
four bumps for one release.

**Change.** Replace the bullet at `SKILL.md:99-101` (v1.0):

> - Bump the version of every changed template (instruction-block
>   `Version:`), skill (frontmatter `version`), and standard
>   (frontmatter `version`) per [[template-authoring-standards]].
>   **Bump each file exactly once for the whole triage, no matter how
>   many proposals changed it** — the version announces the release, not
>   the proposal count. WHEN several proposals touch one file, the bump
>   belongs to whichever lands last; verify the bump after the final
>   change to that file, never before.

**Evidence.** Finding 2. Pairs with P1 but stands alone: a triage that
implements inline still needs the rule, and P1's grooming pass has to be
told which bead owns the bump.

### P3 — Scope step 7's plugin bump to skills that ship in the plugin [snag]

**Problem.** Finding 3. "WHEN any skill … changed" is false for a
repo-local skill, and `triage-rfc` is itself the counterexample.

**Change.** Amend `SKILL.md:124` (v1.0):

> 7. **Publish.** WHEN any skill, template, or standard **that ships
>    inside the plugin** changed, bump the plugin version in
>    `plugin/.claude-plugin/plugin.json` (MINOR for normative changes).
>    A skill ships in the plugin IFF it has a source directory under
>    `plugin/skills/`; a tracked skill living only under
>    `.claude/skills/` (such as `triage-rfc` itself) is repo-local, and
>    changing it does NOT on its own justify a bump, because installs
>    never receive it. WHEN a triage changed only repo-local skills, say
>    so and skip the bump.

**Evidence.** Finding 3, with the `ls` and `git ls-files` output above.

### P4 — Add "already implemented" as a recorded outcome [snag]

**Problem.** Finding 4. Three sound, approved proposals had no honest
classification, and `reject — already addressed` would have recorded
them as failures.

**Change.** Add to step 3's list at `SKILL.md:92` (v1.0):

> - **already-implemented** — the proposal is sound and its change is
>   already present in the tree, landed by earlier work rather than by
>   this triage. Distinct from `reject — already addressed`, which is
>   for a proposal whose premise no longer holds: this one was right and
>   nothing about it is being turned down. Record it in the Resolution
>   as "adopted — already implemented", name where the change lives, and
>   bump nothing.

**Evidence.** Finding 4. The distinction is load-bearing for the dedupe
in step 2, which reads prior Resolutions: a future triage matching this
theme must see that the idea was accepted, not that it was refused.

### P5 — Require a self-authored RFC to disclose it [snag]

**Problem.** Finding 5. The skill's integrity rules assume author and
triager are different parties. When they are the same, nothing marks it,
and RFC 0005 adopted a proposal that retroactively sanctioned its own
author's choice.

**Change.** Add to step 1 (Provenance check), after `SKILL.md:68`:

> WHEN the RFC's `origin` is `self` AND the model triaging it is the
> model named in its `model` field, the Resolution SHALL state that the
> RFC was authored and triaged by the same agent, and name any proposal
> that ratifies a decision its own author made. This does not disqualify
> the proposal — self-retro is how this repository improves — but an
> approval recorded without the disclosure reads as independent
> corroboration when there was none.

Add to the anti-pattern table at `SKILL.md:142`:

> | Adopting a self-authored proposal without recording that it was
> self-authored | The Resolution is the durable record; an undisclosed
> self-adoption is indistinguishable later from an independently
> reviewed one |

**Evidence.** Finding 5. Cheap and mechanically checkable — the
frontmatter already carries `origin` and `model`, so the condition is
readable without judgment.

### P6 — Give the retro log a format and a create-if-missing rule [snag]

**Problem.** Finding 6. Step 6 says "append" to a file that need not
exist, and five field names do not determine a layout.

**Change.** Amend `SKILL.md:120` (v1.0):

> 6. **Log.** Append one row per ingested item to `rfcs/retro-log.md`.
>    WHEN the file does not exist, create it with this header and no
>    other content; do not restyle an existing log to match:
>
>    ```markdown
>    # Retro Log
>
>    | Date | Origin | Retroed | Item | Outcome |
>    |------|--------|---------|------|---------|
>    ```
>
>    One row per item, newest at the bottom: date (`YYYY-MM-DD`), origin
>    project (`self` for this repo), the skill or phase retroed, the
>    item (`RFC <nnnn>` or `retro note`), and the outcome as per-proposal
>    counts (`N adopted, N deferred, N rejected`) or `null retro`.
>    Counts are per proposal, never per RFC. Any commentary goes below
>    the table under `## Run notes`, never inside a row.

**Evidence.** Finding 6. The log's only value is comparison across runs;
a format fixed at creation is what makes the second entry comparable to
the first.

## Considered and Not Proposed

- **Require `bead-bonanza` for every triage implementation.** Rejected:
  orchestration overhead is unjustified for the two- or three-proposal
  triages that will be the common case once the inbox drains. P1
  requires *sequencing* and leaves dispatch optional, which gets the
  correctness benefit without the cost.
- **Specify the RFC numbering width and starting number.** Run 2 was the
  first triage, so `SKILL.md:116-118`'s "this repo's next free number"
  had an empty set to draw from, and the agent chose `0001` with
  four-digit padding. Below the bar: the pattern `rfc-<nnnn>-` already
  implies the width, and `0001` from empty is not a real ambiguity. Noted
  only so a future run choosing `0000` is recognised as drift rather
  than a new decision.
- **Require an independent reviewer for self-authored RFCs.** Rejected
  as unimplementable here: this repository has one human and one model,
  so "independent" has no referent. P5 proposes disclosure, which is the
  part that can actually be honored.
- **Treat the step-4 table's own claims as needing citation.** The
  citation requirement adopted this run landed on `retro-skill` step 7
  and on the skill-authoring standard, and `triage-rfc`'s step 4 table
  is a report to a human that the standard now covers. Not proposed
  separately: the standard requirement already reaches it, and
  duplicating it here would be the N-near-identical-rules anti-pattern.
- **The step-2 dedupe defect (`bd list` hiding closed beads).** Real and
  hit again in run 2, but already adopted this run as RFC 0002 P2 and
  fixed in v1.1. Recorded so it is not re-proposed.

## Non-Goals

- Changing the step-4 human gate. It held in both runs and the
  unattended branch adopted this run extends it correctly; nothing here
  touches it.
- Moving `triage-rfc` into the plugin. P3 makes the repo-local status
  explicit rather than resolving whether it should ship — that is a
  separate decision, and pre-judging it here would smuggle a
  distribution change into a bug fix.
- Making grooming mandatory. P1 is conditional by design; a small triage
  should stay cheap.

## Adoption

P1 and P2 are coupled in practice — P1's grooming pass needs P2's rule
to know which bead owns a bump — but either may land alone. P3 through
P6 are independent of both and of each other.

Files touched: `.claude/skills/triage-rfc/SKILL.md` (all six). The skill
is at `version: 1.1` as of this run; all six proposals change normative
content, so it bumps MINOR. **No plugin bump is warranted by this RFC** —
which is P3's own rule applied to itself, `triage-rfc` being repo-local.

Backward compatibility: additive. No resolved RFC or existing log entry
becomes invalid. P6 explicitly does not restyle an existing log, and P4
adds an outcome without redefining the three that exist.

Expected effect: the implement step stops depending on a human noticing
that fourteen changes need an order, and the four situations that steps
5-7 met on their first execution stop being decisions the agent has to
invent under time pressure.

## Resolution

**Triaged 2026-08-01.** 6 proposals, all adopted; P1 with a refinement.

| Proposal | Outcome |
|----------|---------|
| P1 — route step 5 through grooming when approvals collide | **adopted with refinement.** Two refinements. First, placement: the proposal said to insert after the "Apply the changes" bullet, but its own heading is "Sequence before you edit" and the rule is useless once editing has begun, so the bullet is first in step 5. Second, the evidence had gone stale — see the provenance note below — so the text does not read as if it were introducing `groom-beads` to the skill for the first time. The rule itself is adopted as written, conditional trigger and all. |
| P2 — bump each changed file's version exactly once | **adopted as proposed.** Appended to the version-bump bullet. Exercised immediately: `contract-template.md` took two substitutions this triage and moved one version. |
| P3 — scope step 7's plugin bump to plugin-shipping skills | **adopted as proposed.** Step 7 now defines shipping as having a source directory under `plugin/skills/`, names `triage-rfc` as the counterexample, and requires a triage that changed only repo-local skills to say so and skip the bump. |
| P4 — add "already implemented" as a recorded outcome | **adopted as proposed.** Added to step 3's list, and to the step-3 line of the progress checklist so the two stay in step. |
| P5 — require a self-authored RFC to disclose it | **adopted as proposed.** Added to step 1 and to the anti-pattern table. Applied to both RFCs in this triage, including this one. |
| P6 — give the retro log a format and a create-if-missing rule | **adopted as proposed.** Step 6 now carries the canonical header, the per-row field list, and the instruction not to restyle an existing log. The existing log was not restyled. |

**Provenance check — one claim had gone stale.** The RFC cites
`skill_version: "1.0"` and its Adoption section assumes the skill is at
1.1. It was at **1.2** when this triage ran. All six quoted passages
were re-verified against 1.2 and all six survive at shifted line
numbers. One evidence claim did not: P1's finding states that "the
words `groom-beads` and `bead-bonanza` do not appear anywhere in the
skill." `groom-beads` had appeared at `SKILL.md:150` since v1.1, added
by RFC 0002 P4 — but in the deferral-bead bullet, for bringing a bead
up to the Definition of Ready, not for sequencing an implementation.
The finding's substance therefore holds and the proposal is adopted;
the claim as literally written was already false when triage read it.

**Dedupe.** No prior RFC or `rfc-triage` bead matches any of these six.
The two deferred beads are
untouched by this batch. The RFC's own Considered and Not Proposed
correctly identified the `bd list` defect as already adopted in RFC
0002 P2, and did not re-propose it.

**Self-authorship, disclosed — and one self-ratifying proposal.** This
RFC carries `origin: self` and `model: claude-opus-5`, and was triaged
by the same model. **P4 ratifies a decision its own author made:** the
phrase "adopted — already implemented" was invented by that agent in
RFC 0001's Resolution and flagged there as an invention; P4 makes it
sanctioned vocabulary. It is additive, cheap, and load-bearing for the
step-2 dedupe, but it is not independent corroboration and a reviewer
should weigh it knowing that. This disclosure is the first application
of P5, which this same RFC proposes.

**Sequencing, and a deviation recorded.** This triage was itself P1's
trigger case: 10 approved proposals, 6 of them colliding on
`.claude/skills/triage-rfc/SKILL.md`. P1 as adopted calls for filing
one bead per proposal and running `groom-beads`. That was not done.
The proposals were sequenced by hand instead — the five architecture
templates first as disjoint one-line substitutions, then the two
multi-proposal files, each bumped once after its final edit. The
reason: every change was a single known substitution executed by one
agent in one session with no dispatch and no parallelism, so the
grooming pass's product (a dependency graph for dispatch) had no
consumer, and this project's `CLAUDE.md` warns that beads created on a
branch conflict textually on merge. The correctness the rule protects —
serialized edits to a shared file, version bump last — was obtained
directly. This is a deviation from the rule as written on its first
run, and it is recorded here rather than smoothed over, because a rule
that its own author skips on day one is exactly what the next retro
should examine.

**Files changed:** `.claude/skills/triage-rfc/SKILL.md` (all six
proposals); `version: "1.2"` → `"1.3"` (one bump, six changes — P2's
own rule).

**No plugin bump is warranted by this RFC.** `triage-rfc` is
repo-local: it has no source directory under `plugin/skills/` and is
tracked only at `.claude/skills/triage-rfc/SKILL.md`. That is P3's rule
applied to itself. The `0.11.0` → `0.12.0` bump this session carries
belongs entirely to RFC 0007's changes to the `spec` skill, its
templates and its guides.

**Verification.** `scripts/check-skills.sh` reports 0 errors. No line
added by this triage exceeds 80 columns.
