---
title: "RFC: A wave report that reads as permission to stop, a rebase target
  the skill's own snapshot rule invalidates, and no rule at all for a subagent
  that never returns"
description: Three things halted or nearly corrupted a 21-bead autonomous
  run — "a report per wave" licensing an end-of-turn, a rebase target that
  goes stale between two merges in one wave, and an acceptance check written
  only for subagents that return.
slug: rfc-self-wave-chaining-and-lost-returns
type: rfc
status: ready-for-review
updated_at: 2026-08-15T00:00:00Z
origin: self
skill: bead-bonanza
skill_version: "1.5"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.18.0"
---

# RFC: Wave chaining, the rebase target, and subagents that never return

## Summary

Run `1786778523-19` drained 21 beads across 9 waves under a standing
authorization. It completed, and every bead landed — but the run was
halted once by the user having to restart an orchestrator that had
stopped with nothing blocking it, lost a commit to an unreferenced
rebase (recovered from the object store), and three times had to
improvise a way to tell a finished bead from an unfinished one because
the subagent's completion notification never arrived.

None of the three is a bead defect and none is carelessness the skill
already warned against. Each traces to a specific instruction: one that
licenses the wrong action, one whose target the skill's own cadence rule
invalidates, and one that does not exist.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins (self — in-repo dogfooding) |
| Skill critiqued | `bead-bonanza` v1.5 |
| Harness | Claude Code on the web, remote cloud container |
| Model | Withheld: this session's policy forbids writing the model identifier into any repository artifact. Recorded in chat only. The same constraint is already filed as `rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md`, so triage cannot treat this as an omission — it is a known, tracked gap in this harness's provenance |
| Effort | `unknown` — this harness exposes no effort setting to the session |
| Runs observed | 1 (`1786778523-19`, 2026-08-15). Two findings recurred *within* that run across independent agents; no claim here rests on more than one run |
| Task | 21 beads implementing Gherkin features and acceptance criteria in the Spec CLI, epic `agent-plugins-wvd` |
| References invoked | `references/subagent-prompts.md`, `references/progress-streaming.md`, `references/recovery.md`, `references/operating-principles.md` (none carry their own version field) |
| Scripts invoked | `scripts/pre-flight-check.sh`, `bb-progress.sh`, `bb-close-bead.sh`, `bb-verify-skip.sh`, `bb-backup-run.sh` |
| Evidence record | `rfcs/sprints/2026-08-15-gherkin-adoption.md` (73 findings verbatim, amendments 1–8) |
| Session date | 2026-08-15 |

## Evidence

1. **The orchestrator stopped mid-run with a non-empty queue, a standing
   authorization covering it, and no `decision-required` bead.** After
   accepting wave 6 it narrated wave 7 rather than dispatching it and
   ended the turn. The user restarted it: "Why'd you stop? ... Keep
   going." Recorded as amendment 6. The skill's own rule is
   `plugin/skills/bead-bonanza/SKILL.md:872`:

   > With it granted, waves flow without per-wave prompts; the user gets
   > a report per wave instead of a question per wave.

   The clause contrasts a *report* with a *question*, which correctly
   removes the prompt — and then says nothing about whether emitting the
   report ends the turn. "The user gets a report per wave" is satisfied
   exactly as well by reporting and yielding as by reporting and
   continuing. The rule that removes the question does not supply the
   one thing the autonomous path needs.

2. **The rebase target is an unbound placeholder, and the skill's own
   snapshot cadence invalidates any value read early.**
   `references/subagent-prompts.md:355` prescribes:

   ```sh
   git -C <worktree> rebase <sprint-head>   # replay this bead onto what landed
   ```

   `<sprint-head>` is never bound to a command or a moment. Meanwhile
   `SKILL.md:184` requires `bb-backup-run.sh` to run "after **every**
   bead's acceptance check, not once per wave", and that script commits
   to the sprint branch. So within a single concurrent wave, accepting
   bead A moves the sprint head before bead B is merged. Merging `4vz` I
   used `135ceab` — the sha the previous bead had merged at — and
   `--ff-only` correctly refused; I then deleted the branch before
   re-checking, leaving the rebased commit unreferenced. Recovered from
   the object store. Amendment 3.

3. **The skill has no rule for a subagent that never returns.** In three
   waves an agent ran its full ceremony to a terminal phase and no
   completion notification arrived (amendment 5 documents wave 4, where
   both agents reached `pushed` and neither returned; the sprint report
   records the same in two later waves). The acceptance check is framed
   entirely around a return —
   `SKILL.md:140`:

   > A bead the subagent returned as `completed` is not closed until
   > **every signal below** agrees.

   A grep for `never return|did not return|no return|notification`
   across `SKILL.md`, `references/progress-streaming.md` and
   `references/recovery.md` returns **nothing**. The five signals happen
   to be return-independent, which is what made recovery possible, but
   nothing says so — and `SKILL.md:177` actively discourages the ledger
   ("the ledger is a hint, never the authority ... its absence proves
   nothing"), which is the signal that in fact resolved every case. A
   `bd list --status in_progress` query disagreed once and was wrong.
   The orchestrator improvised: first a durable-signal check, then a
   bounded watcher polling each bead's progress file for any terminal
   phase.

4. **Two independent agents lost a pass to the same worktree-path
   asymmetry.** `references/subagent-prompts.md:91` says:

   > Worktree: {WORKTREE}        (work HERE, not in the main checkout; it is yours alone)

   and `:16` warns that omitting it "sends the agent into the shared
   checkout, which defeats every guarantee this design rests on".
   Neither says the enforcement is partial. `Read` and `Grep` against an
   absolute shared-checkout path succeed **silently**, reading a sibling
   checkout; `Write`, `Edit` and `Bash` refuse the same paths loudly.
   `wvd.6` recorded "Reading several files under the shared path before
   the first Edit failure cost a redo pass"; `wvd.11` hit it
   independently — "the mismatch wasn't caught until the first Write
   failed". Neither could see the other's finding.

5. **A return asserted a gate outside the profile, and nothing marks
   such a claim unverified.** `wvd.13` reported
   `spec check --config rfcs/specs.toml` as "no findings, so RFC-0023's
   frontmatter validates". Measured: that command exits 1 with 222
   `reference/unresolved` findings. It is not in the run's gate profile,
   so the orchestrator's own full-gate pass would not have covered it;
   the claim was caught only because the orchestrator independently ran
   it. The acceptance check's five signals (`SKILL.md:143–159`) contain
   no gate signal at all, and Phase 4 has no rule distinguishing a
   measured statement from a reported one. Amendment 8. (The underlying
   state was benign — the same command exits 1 at the pre-sprint
   merge-base — but a false statement would have entered the sprint
   report as fact.)

## Proposals

### P1 — Say that the wave report does not end the turn [blocker]

**Finding shape:** misleading-text — `SKILL.md:872`, evidence 1.

**Problem.** "The user gets a report per wave instead of a question per
wave" removes the prompt and leaves the turn boundary unstated. An
orchestrator that reports and yields has satisfied the sentence as
written, and the run stops until a human notices. Standing
authorization is exactly the mode in which no human is watching.

**Change.** Replace the clause at `SKILL.md:872–874` with:

> With it granted, waves flow without per-wave prompts: emit the wave's
> report and **dispatch the next wave in the same turn**. The report is
> a progress notice, not a checkpoint — under standing authorization
> the turn ends only when the queue is dry, a `decision-required` bead
> is reached, or the circuit breaker trips. Ending a turn with a
> non-empty queue and no such cause is a defect, not a pause: nobody is
> waiting to restart it.

**Evidence.** Evidence 1. The run halted once and needed a user message
to resume.

### P2 — Bind the rebase target to the moment of merge [blocker]

**Finding shape:** misleading-text — `references/subagent-prompts.md:355`,
evidence 2.

**Problem.** `<sprint-head>` reads as a value the orchestrator may
resolve once per wave. `SKILL.md:184` guarantees it cannot be: the
per-bead snapshot commit moves the sprint head between two merges in the
same wave. The two rules are individually correct and jointly a trap.

**Change.** Replace the snippet at `references/subagent-prompts.md:355`
with a command that reads the target rather than naming it, plus the
reason:

> ```sh
> git -C <worktree> rebase "$(git rev-parse HEAD)"  # read the target NOW
> git merge --ff-only <suffix>
> ```
>
> **Read the target at merge time, never earlier in the wave.**
> `bb-backup-run.sh` commits to the sprint branch after *every* bead's
> acceptance (`SKILL.md:184`), so accepting the wave's first bead moves
> the head before its second is merged. A sha noted at wave start is
> already stale. And do not delete a branch until after the
> fast-forward succeeds — a rebase onto a stale target leaves the
> rebased commit unreferenced, recoverable only from the object store.

**Evidence.** Evidence 2, amendment 3.

### P3 — Give the acceptance check a no-return branch [blocker]

**Finding shape:** gap — no rule at Phase 3 step 4 (`SKILL.md:139`) for
"the subagent never returned", evidence 3.

**Problem.** Every acceptance instruction is conditioned on a return.
When the notification is lost the orchestrator has no sanctioned way to
distinguish a finished bead from an abandoned one, and the skill's
nearest statement — that the ledger proves nothing — points away from
the signal that actually resolves it. The failure mode is re-dispatching
a bead that already closed and pushed.

**Change.** Add to `SKILL.md` immediately after the five signals:

> **A subagent that never returns is checked by these same five
> signals.** A lost completion notification is not evidence of an
> unfinished bead, and it is common enough to plan for. The five signals
> above read state, not testimony, so they are valid with or without a
> return — run them unchanged. To decide whether a bead is still
> *working* rather than finished, read its `<bead-id>.progress.jsonl`
> last phase together with `bd show <id>`; a terminal phase
> (`committed`/`pushed`/`blocked`) plus a closed tracker record is
> conclusive. Do NOT use `bd list --status in_progress` for this — it is
> a query over a moving set and disagreed with both durable signals in
> practice. **Never re-dispatch on a missing return alone.**

This is the complement of `rfc-self-worktree-composition-and-shared-gates`
P6, which covers "returned but no terminal event"; the two together
close both directions and neither subsumes the other.

**Evidence.** Evidence 3, amendment 5.

### P4 — State that only the mutating tools enforce the worktree [snag]

**Finding shape:** gap — no rule at the per-bead template's `Worktree:`
line (`references/subagent-prompts.md:91`) for tool-dependent
enforcement, evidence 4.

**Problem.** The prompt asserts the worktree is where the agent works
and implies the harness holds it there. It does not: the read tools
accept a shared-checkout path silently and return another checkout's
content, so an agent can read a stale file, reason from it, and only
discover the error at its first write. Two independent agents in one run
paid for this, which makes it a property of the instruction rather than
of either agent.

**Change.** Extend the template's `Worktree:` line:

> Worktree: {WORKTREE}        (work HERE, not in the main checkout; it is yours alone)
>            **Prefix EVERY path with this root, or use cwd-relative
>            paths, from your first tool call.** Enforcement is partial:
>            `Write`/`Edit`/`Bash` refuse a shared-checkout path loudly,
>            but `Read`/`Grep` ACCEPT it silently and return the other
>            checkout's content — so a wrong-root read is not an error,
>            it is wrong data.

**Evidence.** Evidence 4 — `wvd.6` and `wvd.11`, independently, in one
run.

### P5 — Mark an unprofiled gate claim as unverified [snag]

**Finding shape:** gap — no rule at Phase 3 step 4 or Phase 4 for a
return that asserts a gate the profile does not contain, evidence 5.

**Problem.** The acceptance check verifies closure, commit, merge,
worktree and findings — never a gate. The orchestrator's own full-gate
pass covers the profile, so a claim about a command *outside* the
profile is checked by nothing, yet reads identically to a measured one
and flows into the sprint report as fact.

**Change.** Add to Phase 4's report instructions:

> **A gate the profile does not name is testimony until you run it.**
> Where a return or handoff asserts the result of a command outside the
> gate profile, either run it yourself and report your measurement, or
> attribute it ("`wvd.13` reports X; not independently verified"). Do
> not restate it as a run fact. A handoff is the subagent's account of
> its own work, and the report is the run's evidence — the two are not
> the same standard.

**Evidence.** Evidence 5, amendment 8.

## Considered and Not Proposed

- **The Bash tool refuses multi-line or compound commands containing
  `git` inside a worktree**, so `r19` had to write every verification
  step to a scratchpad script and run `bash <file>` (finding `r19#4`).
  Real friction, but a property of the harness's worktree guard, not of
  `bead-bonanza` — the skill can neither fix nor reliably predict it,
  and a warning about another tool's behavior in a skill that does not
  own it would rot. Recorded here so recurrence is visible if it appears
  in another run.
- **Six of 21 beads carried no `c4_changes`/`c4_impacts`** (amendment
  2), so a third of the queue was bucketed by file-path signal rather
  than by the field the rule names. A `groom-beads` finding, not a
  `bead-bonanza` one — this orchestrator tagged only the `wvd.*`
  children. Routed in the sprint report's §6 as an offer against that
  skill; filing it here would put it in the wrong directory for
  recurrence counting.
- **Zero tier escalations across 21 beads.** Reported with its caveat in
  the sprint report §8, but not a skill defect: `SKILL.md:612` already
  requires exactly that caveat and the run followed it.
- **The durable trigger's drained-run branch.** Already proposed as
  `rfc-self-acceptance-signals-and-gate-baseline` P6; this run
  corroborates it (the check-in fired into a run that had drained) but
  re-proposing would split one decision across two RFCs.

## Non-Goals

- No change to the five acceptance signals themselves. P3 and P5 add a
  branch and an attribution rule around them; the signals held on every
  one of 21 beads and needed no revision.
- No change to the worktree-per-bead model, the one-commit-per-bead
  invariant, or the snapshot cadence. P2 fixes the rebase target, not
  `bb-backup-run.sh`'s per-bead schedule, which is what made the run
  recoverable in the first place.
- No new script. Every proposal is text in `SKILL.md` or
  `references/subagent-prompts.md`.

## Adoption

P1 through P5 are independent and each edits one location. P1, P3 and P5
touch `SKILL.md`; P2 and P4 touch `references/subagent-prompts.md`. None
changes a script, a progress phase name, or the run-directory layout, so
a run in flight under v1.5 stays valid and no artifact needs
regenerating.

P1 is the one to adopt first if only one is adopted: it is the only
finding here that stopped the run outright, and its failure mode is
silent — an orchestrator that yields under standing authorization looks
exactly like one that finished.

The expected effect is that the next autonomous run chains its waves
without a human restart, merges every concurrent wave against a target
that cannot be stale, and treats a missing notification as the
non-event it is.

## Resolution
