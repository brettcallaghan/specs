---
title: "RFC: Make the sprint report's citations resolve, and stop the batch
  template branching every bead from the same base"
description: Phase 4 cites findings by a number nothing checks, and the batch
  prompt branches each bead in a batch from repo HEAD instead of its
  predecessor
slug: rfc-self-report-citations-and-batch-base
type: rfc
status: ready-for-review
updated_at: 2026-08-12T23:05:00Z
origin: self
skill: bead-bonanza
skill_version: "1.3"
model: claude-opus-5
effort: unrecorded
runs_observed: 1
severity: blocker
plugin_version: "0.17.0"
---

# RFC: Make the sprint report's citations resolve, and stop the batch template branching every bead from the same base

## Summary

`bead-bonanza` ran a 17-bead sprint (run `1786562615-17`, epic
`agent-plugins-41z`) to completion: 16 beads closed, 1 blocked, 73 findings
harvested. The work landed. Two defects in the skill are nonetheless
blockers, and they share a shape — **the skill mandates a discipline and
then provides nothing that can detect its violation**. Phase 4 requires the
sprint report to cite its evidence, and all 13 citations in the report it
produced point at the wrong finding. The batch dispatch template tells each
subagent in a batch to branch from `HEAD`, which cannot advance until the
batch returns, so a batch's beads silently do not build on each other.

## Provenance

| Field | Value |
|-------|-------|
| Skill | `bead-bonanza`, SKILL.md frontmatter `version: "1.3"` |
| Version note | `agent-plugins-x8r` bumped it 1.3 → 1.4 in commit `8e8f7c3` during this same run; the version **exercised** by every bead was 1.3 |
| Orchestrator model | `claude-opus-5` |
| Subagent models | `sonnet` (waves 1a, 2A, 2B, 8ef.3), `opus` (d70, 8ef.1, 8ef.2, wave 4), `haiku` (wave 5) |
| Effort | **unrecorded** — see note below |
| Harness | Claude Code 2.1.228, remote execution environment |
| Runs observed | 1 |
| Task | 17 enumerated beads: defect fixes to `pre-flight-check.sh`, a redundancy-cut epic over `SKILL.md` and its references, two new private skills, and a version-bump sweep |
| Run dates | 2026-08-12, base sha `b8b36d1`, branch `claude/bead-bonanza-refinement-261qvk` (PR #20) |
| Evidence base | `rfcs/sprints/2026-08-12-bead-bonanza-hardening.md` (73 verbatim `retro:` findings plus the orchestrator's analysis), and this session's direct observation while resuming the drained run. Every finding number cited below indexes that file. |

References and scripts invoked, at the versions run:
`references/subagent-prompts.md`, `references/dispatch-plan.md`,
`references/recovery.md`,
`references/operating-principles.md`, `references/project-setup.md`,
`scripts/pre-flight-check.sh`, `scripts/bb-progress.sh`,
`scripts/bb-verify-skip.sh`, `scripts/bb-close-bead.sh`,
`scripts/bb-backup-run.sh` — all as committed at `b8b36d1`.

**On `effort`:** the run's `dispatch.md` manifest records the model lineup but
carries no effort field, and no artifact this run wrote records one. Recording
`unrecorded` rather than guessing, because the template's own rationale is that
a defect is model-relative and triage cannot scope a fix from a fabricated
value. That the manifest omits it is noted under Considered and Not Proposed.

## Evidence

### E1 — every evidence citation in the sprint report points at the wrong finding

Shape: **gap**. Severity: **blocker**.

Phase 4 requires the report to cite its evidence, and the framework's own
standard requires it: `specs/standards/skill-best-practices.md` states that
where a skill's workflow ends in a report to a human, that step "SHALL require
every factual claim to cite the artifact it derives from", and that citation is
"mechanically checkable: a sentence either names its source or it does not."

The report complies in form and fails in substance. Its analysis sections cite
findings by a **globally sequential number** (`(finding 15)`, `(findings 29,
30)`). Nothing checks that the number names the finding the sentence is about.

Measured 2026-08-12 by mapping each finding number to its owning bead from the
report's own `### \`agent-plugins-<id>\`` headings, then checking each citation
against the bead named in the same paragraph:

```
13 citations checked
 0 resolve to a finding belonging to the bead the claim is about
13 point at a finding belonging to a different bead
```

Two verified by hand:

- The report says *"`8ef.1` promises eighteen rules and names fourteen
  (finding 15)"*. Finding 15 belongs to `q1h` and reads *"decided silently —
  gave the moved probe its own numbered section…"*. The finding that actually
  says it is **37**.
- The report says *"`x8r` shipped `out1`–`out4` in its close commit (finding
  63)"*. Finding 63 belongs to `b5z` and reads *"Every cited line range was
  stale. Phase 2.5 was SKILL.md:494-726…"*. The correct finding is **73**.

The failure is structural, not clerical. A global counter is assigned only once
all 17 beads' findings are concatenated, while the analysis is written per
theme; the author is citing a number that does not exist until the document is
assembled. Every citation being wrong — rather than a scattered few — is
what a systematic offset looks like, and it is why "be careful" is not the
fix.

The consequence is precise: the report is the run's only durable evidence
record, `triage-rfc` defers on recurrence and therefore reads these findings
across runs, and a reader who follows a citation lands on unrelated text. A
citation gate that cannot fail launders an unchecked claim as a checked one —
the same defect class the standard names for scripts.

### E2 — the batch template branches every bead in a batch from the same base

Shape: **misleading-text**. Severity: **blocker**.

`references/subagent-prompts.md:178` (and the worked block at `:312`) instructs
a batch subagent:

> `git -C <repo> worktree add -b <suffix> <repo>/../.bb-worktrees/<suffix> HEAD`

Repo `HEAD` advances only when the **orchestrator** merges an accepted bead,
which happens after the whole batch returns. So every bead in a batch branches
from the same commit, and the beads do not build on each other.

This directly contradicts the worktree lifecycle in the same bundle, which
states that branching from `HEAD` — "the sprint branch as it stands *now*,
after the previous bead merged" — is what makes each bead build on the last.
The rule is right; the batch template's `HEAD` is not the `HEAD` that rule
means.

Observed: harmless in wave 1 (two beads; `d6x` rebased cleanly onto `txq`).
Wave 2A would have produced **four conflicting branches on one file** —
`q1h`, `td1`, `9p4` and `8mv` all edit `scripts/pre-flight-check.sh`. The
orchestrator detected it and overrode the template with explicit chaining
(`worktree add -b td1 <path> q1h`) for every subsequent batch; all five batches
then merged as single fast-forwards.

Still present at the version in the working tree after this run
(`subagent-prompts.md:178,312`), so `8ef.3`'s template refactor did not touch
it.

### E3 — capturing a gate's exit status ships the output file

Shape: **misleading-text**. Severity: **snag**.

The skill correctly warns to capture `$?` directly rather than through a pipe,
because `cmd | tail` reports `tail`'s status. The natural form of that advice
is `cmd >out 2>&1`, and the dispatch prompt instructed exactly that. Run inside
the bead's own worktree, `bb-close-bead.sh` stages that file — by design,
since everything in the tree is the bead's work.

Finding 73: *"the dispatch prompt instructed 'cmd >out 2>&1'… those redirect
files (out1..out4) were staged by bb-close-bead.sh… and shipped in the close
commit."* Confirmed in the tree: commit `bb8969f` removes `out1`–`out4`, and
its message records that four more were left untracked.

The ceremony is not at fault. The guidance is: it names a redirect target
without saying where it must live.

### E4 — the `Monitor` cap is not mentioned anywhere in the skill

Shape: **gap**. Severity: **snag**.

The skill instructs the orchestrator to arm a `Monitor` for any subagent
expected to run past 5 minutes. Observed this run: `Monitor` caps at 30 minutes
regardless of the requested `timeout_ms` — 3,600,000 ms was requested and
1,800,000 reported — and it expired three times, each requiring a manual
re-arm. Most beads ran 10–16 minutes and several batches ran over an hour, so
re-arming is the normal case, not an edge case. No rule at the Monitor step
for "the Monitor expired while the subagent is still running".

### E5 — Phase 4's trigger cancellation has no visible outcome

Shape: **gap**. Severity: **polish**.

`SKILL.md:699` step 6a does instruct the orchestrator to *"Cancel the durable
out-of-session trigger armed at first dispatch"*, and
`references/progress-streaming.md:112` gives the rationale. The rule exists,
so the orchestrator skipping it is an execution miss and out of scope.

What is in scope is that its omission is **invisible at the time**. Every other
Phase 4 step yields an artifact — a report, a commit, a closed bead. A live
trigger yields nothing until it fires later, into a different session. Observed:
the trigger for this run fired twice into a fully drained sprint and was
deleted manually by a resuming session, which is precisely the outcome
`progress-streaming.md` warns of.

### E6 — verification-by-reconstruction is not named as the standard

Shape: **gap**. Severity: **polish**.

`8ef.3` refactored the dispatch templates by factoring shared blocks into a
preamble the orchestrator concatenates. Inspecting that diff cannot show
whether the assembled prompt still carries every rule. Finding 57 records the
method that can: assemble preamble + template exactly as the orchestrator
sends it, then diff normalised word-sets in both directions. It found the only
loss was a list-numbering artifact. The skill gives no rule for verifying a
change to a template whose product is assembled rather than read.

## Proposals

### P1 — number findings per bead, and check that citations resolve

Severity: **blocker**. Shape: gap. Evidence: E1.

**Change.** In Phase 4's report step, number harvested findings **per bead**
rather than globally: `8ef.1#3`, `x8r#1`. A per-bead number is stable at the
moment the finding is harvested, so a claim can cite it while the analysis is
being written rather than after assembly.

Add `scripts/bb-check-citations.py`, run as the last step before the report is
committed, exiting non-zero when any citation fails to resolve. It parses the
`### \`<bead>\`` headings to map finding ids to beads, then for each
`(finding <id>)` in the analysis sections checks the id exists and belongs to a
bead named in the same paragraph. A working implementation exists — it
produced E1's measurement — and reports `DANGLING` (no such finding),
`MISMATCH` (wrong
bead) and `UNATTRIBUTED` (paragraph names no bead) separately.

State in the step that the report SHALL NOT be committed while the check
fails, matching how the skill already gates `bb-verify-skip.sh`.

### P2 — chain each bead in a batch onto its predecessor

Severity: **blocker**. Shape: misleading-text. Evidence: E2.

**Change.** Replace the batch template's worktree creation so that the first
bead branches from `HEAD` and each subsequent bead branches from the branch of
the bead before it:

> First bead: `git -C <repo> worktree add -b <id-1> <path-1> HEAD`
> Each later bead: `git -C <repo> worktree add -b <id-N> <path-N> <id-N-1>`
> — branch from the PREVIOUS bead's branch, not from `HEAD`. Inside a batch
> `HEAD` does not move: the orchestrator merges only after the batch returns,
> so branching every bead from `HEAD` gives beads that cannot see each other's
> work and conflict on any shared file.

Add the intent sentence, per the standard's rule that a prescribed sequence
should state the intent it serves — the orchestrator that overrode this
template did so because it understood the intent, and an orchestrator that
does not will follow the literal instruction.

### P3 — redirect gate output outside the worktree

Severity: **snag**. Shape: misleading-text. Evidence: E3.

**Change.** Wherever the skill or a dispatch template shows capturing a gate's
status by redirection, name a path outside the bead's tree — the run directory
is already passed to every subagent as `{RUN_DIR}`:

> `<gate-cmd> >"$RUN_DIR/<id>.<gate>.out" 2>&1; echo $?` — never redirect into
> the worktree: the close ceremony stages everything in that tree by design,
> so a file written there ships in the bead's commit.

### P4 — state the Monitor cap and the re-arm expectation

Severity: **snag**. Shape: gap. Evidence: E4.

**Change.** At the Monitor step, add:

> `Monitor` caps at 30 minutes regardless of the `timeout_ms` requested.
> Beads routinely run longer, so expect to re-arm: when a Monitor expires
> while a subagent is still live, re-arm it rather than treating expiry as a
> signal about the subagent. Expiry says nothing about the bead.

### P5 — make the trigger's disposition a line in the final report

Severity: **polish**. Shape: gap. Evidence: E5.

**Change.** Add the trigger's disposition to Phase 4's report contents, so the
one step whose omission produces no artifact produces one:

> Durable trigger: cancelled (`<id>`) / none armed.

This does not add a rule; it makes an existing rule's outcome checkable by the
same standard the report's other claims are held to.

### P6 — name verification-by-reconstruction for assembled prompts

Severity: **polish**. Shape: gap. Evidence: E6.

**Change.** Where the skill describes changing a dispatch template, add:

> A template whose product is assembled rather than read is verified by
> reconstruction, not by reading the diff: assemble preamble + template
> exactly as the orchestrator sends it and diff normalised word-sets in both
> directions against the pre-change text. The diff shows what moved; only the
> reconstruction shows what a subagent will actually receive.

## Considered and Not Proposed

- **The orchestrator did not cancel the durable trigger.** The rule exists at
  `SKILL.md:699` step 6a. The skill already says to do it, so the omission is
  the agent's, not the skill's, and is out of scope. Only the visibility gap
  is proposed (P5).
- **`bd show --json` silently omits comment bodies**, returning
  `comments_omitted: true` and a count, so a naive `jq '.comments[]'` harvest
  reads as a clean run. Real, and the orchestrator hit it — but the skill
  already mandates `bd export | jq` for the harvest, so no change follows. A
  second independent reason for an existing rule is not a defect in it.
- **Stale `file:line` citations in bead descriptions** were the run's dominant
  friction, recorded independently by six beads. This is a property of how the
  beads were groomed, not of `bead-bonanza`, and is filed separately against
  `groom-beads`.
- **`dispatch.md`'s manifest records the model lineup but no effort setting**,
  which is why this RFC's `effort` is `unrecorded`. Not proposed here because
  one run cannot show it changed an outcome; recorded so a second occurrence
  has something to recur against.
- **Zero tier escalations across 16 closed beads.** Consistent with good
  routing and equally consistent with routing everything too high, since
  over-routing succeeds silently. Not a finding in either direction — noted
  because the absence of evidence is the observation.
- **The `--gate-baseline` mechanism worked and is load-bearing.** Every
  subagent compared against the recorded failing baseline rather than
  expecting exit 0, and none tried to "fix" the known-red warning. Recorded as
  a design that paid off; no change proposed.

## Non-Goals

- Re-litigating the wave partition. The C4 metadata predicted the collision
  structure correctly and the plan never needed resequencing.
- The size of `SKILL.md`. It remains over the re-injection cap, which is
  tracked as beads, not as a skill defect.
- Correcting the citations in the already-committed sprint report. P1 changes
  what future reports do; whether to repair this one is triage's call.

## Adoption

P1 and P2 are independent and can land in either order. P1 adds a script and
a numbering change to one Phase 4 step; P2 edits two lines of one reference
file plus an intent sentence. P3–P6 are text-only.

P2 has a migration wrinkle worth naming: an orchestrator mid-run that has
already dispatched a batch under the old template must keep overriding it by
hand, as this run did. Nothing breaks; the fix simply does not apply
retroactively.

## Resolution

**Partial triage 2026-08-17. This RFC is NOT retired** — P3, P5 and P6 are
still undecided. Recorded now because the provenance check settled three
proposals outright, and that result should not have to be re-derived.

| Proposal | Outcome |
|---|---|
| P1a — number findings per bead rather than globally | **adopted — already implemented.** Present at `SKILL.md` Phase 4 ("Number findings per bead, not globally: `<bead-id>#<ordinal>`"), carrying this proposal's own rationale about a global counter not existing until assembly. Landed by earlier work, not by this triage. The proposal was right; nothing was bumped for it. |
| P1b — add `scripts/bb-check-citations.py` and gate the commit on it | **recommended defer; not yet approved.** With P1a landed the residual risk is much smaller — a per-bead id names its own bead, so a mis-citation is visible on its face rather than only to a script. A new script for what remains wants a second run's evidence. Theme `report-citations`. |
| P2 — chain each bead in a batch onto its predecessor | **adopted — already implemented.** `references/subagent-prompts.md:178–179` now reads "First bead: … HEAD" / "Each later bead: … `<id-N-1>`", which is this proposal verbatim. Landed mid-run by bead `agent-plugins-61f.2` and independently reported as fixed by `rfc-self-acceptance-signals-and-gate-baseline`. |
| P3 — redirect gate output outside the worktree | **recommended defer; not yet approved.** Triage could not find the defect site: no template or skill passage shows the `cmd >out 2>&1` pattern. The evidence describes an *orchestrator-authored* dispatch prompt, not skill text, so this is preventive guidance rather than a fix. The only exit-status advice in the bundle is `SKILL.md:332`, about a pipe to `tail` masking a push's status. Theme `close-ceremony`. |
| P4 — state the Monitor cap and the re-arm expectation | **adopted — already implemented.** `references/progress-streaming.md:82–87` carries the re-arm rule and the five-minute default. The *measured* 30-minute ceiling is still unrecorded; that residue is `rfc-self-acceptance-signals-and-gate-baseline` P5, which is where it should be decided rather than split across two RFCs. |
| P5 — make the trigger's disposition a line in the final report | **recommended adopt; not yet approved.** |
| P6 — name verification-by-reconstruction for assembled prompts | **recommended defer; not yet approved.** Sound, but it is skill-authoring guidance sitting in a skill that does not own the subject; it belongs against the skill-authoring standard. Theme `skill-authoring`. |

**Provenance check — the reason three rows above read "already
implemented".** `bead-bonanza` v1.3 cited; **v1.5 current at triage**, two
minor versions on. Each citation was re-read against the tree rather than
assumed, and three had been fixed in the interim. Under this skill's own test
these are *adopted — already implemented* rather than *rejected — already
addressed*: each would have been adopted had the tree been untouched, so
triage is ruling on the merits and a future triage should read them as settled
rather than as evidence that went stale.

**Dedupe.** P2's text is also the subject of
`rfc-self-worktree-composition-and-shared-gates` P1, which concerns the same
block for a different defect (composition with harness-managed isolation) and
is unaffected by this fix. No match in retired RFCs 0001–0023.

**Self-authorship, disclosed.** `origin: self`, `model: claude-opus-5`; this
session is configured for `claude-opus-5`. **The RFC was authored and triaged
by the same model.** The three already-implemented rows are the exception that
cuts the other way: they were settled by measurement against the tree and by a
later run's independent report, not by this triage's judgement.

**Files changed.** None — every decided proposal here was already present.

**Version bumps.** None attributable to this RFC.
