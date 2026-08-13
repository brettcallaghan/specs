---
title: "RFC: Two acceptance signals the orchestrator cannot check as written,
  and a gate baseline that silently skips the gate that needs one"
description: Signal 5 names no command and the obvious one under-reports;
  --gate-baseline cannot parse a fenced test gate and records nothing for the
  known-red check
slug: rfc-self-acceptance-signals-and-gate-baseline
type: rfc
status: ready-for-review
updated_at: 2026-08-13T07:00:00Z
origin: self
skill: bead-bonanza
skill_version: "1.4"
model: claude-opus-5
effort: unrecorded
runs_observed: 1
severity: blocker
plugin_version: "0.17.0"
---

# RFC: Two acceptance signals the orchestrator cannot check as written, and a gate baseline that silently skips the gate that needs one

## Summary

`bead-bonanza` ran a 7-bead sprint (run `1786600009-7`, epic
`agent-plugins-61f`) to completion: 7 closed, 0 blocked, 0 skipped, 14 findings
harvested. The work landed and the spine held. The friction clusters in the
**verification** layer rather than the execution layer, and it has one shape:
**the skill states a check without stating how to perform it, and the obvious
way to perform it gives the wrong answer.** Phase 3's acceptance signal 5 names
no command, and the natural one reports zero findings on beads that have them.
Phase 0's `--gate-baseline` cannot parse this project's test gate and records
nothing for the one gate that is known-red — the exact gate the "no findings
beyond baseline" bar depends on.

## Provenance

| Field | Value |
|-------|-------|
| Skill | `bead-bonanza`, SKILL.md frontmatter `version: "1.4"` |
| Version note | The run itself bumped it 1.4 → 1.5 (bead `agent-plugins-61f.7`, commit `f6986f1`). Every passage cited below was verified **unchanged** between `b8e9436` and HEAD with `git diff b8e9436..HEAD -- plugin/skills/bead-bonanza/`, so the citations hold against both versions. Where a proposal concerns text this run *added*, it says so. |
| Orchestrator model | `claude-opus-5` |
| Subagent models | `opus` (planning, `61f.4`), `sonnet` (wave 1 batch, `61f.5`) |
| Effort | **unrecorded** — see note below |
| Harness | Claude Code, remote execution environment |
| Runs observed | 1 |
| Task | 7 beads implementing the approved proposals of a same-day RFC triage: 5 co-editing `bead-bonanza/SKILL.md`, 2 on the `spec`/`groom-beads` bundle |
| Run dates | 2026-08-13, base sha `b8e9436`, branch `claude/triage-rfc-self-report-anchors-7vlncg` |
| Evidence base | `rfcs/sprints/2026-08-13-rfc-triage-adoption.md` (14 verbatim `retro:` findings plus the run's own analysis), and this session's direct observation as orchestrator. Finding ids below use that report's per-bead form. |

References and scripts invoked, at the versions run:
`references/subagent-prompts.md`, `references/progress-streaming.md`,
`references/shared/work-order-definition-of-ready.md`,
`scripts/pre-flight-check.sh`, `scripts/bb-progress.sh`,
`scripts/bb-close-bead.sh`, `scripts/bb-backup-run.sh` — all as committed at
`b8e9436`. None of these files carries a version property; `bead-bonanza`
versions as a bundle and SKILL.md holds that version.

**On `effort`:** no artifact this run wrote records one, and the harness does
not expose it to the orchestrator. Recorded as `unrecorded` rather than
guessed. Noted with some irony: bead `agent-plugins-61f.7`, closed in this very
run, adds an effort field to the dispatch manifest — but the manifest is
written in Phase 2.5, before any bead is dispatched, so the rule cannot apply
to the
run that introduced it. The next run has no such excuse.

## Evidence

### E1 — signal 5 names no command, and the obvious one under-reports

Shape: **gap**. Severity: **blocker**.

`SKILL.md:156-158` states the fifth acceptance signal in full:

> 5. **Findings were recorded** — at least one `retro:` comment on the bead,
>    or the subagent's explicit "clean bead, no findings". A silent bead is
>    neither, and is not accepted.

No command is given. Every other signal in that list names one. Signal 2 gives
`git log --oneline --grep=<bead-id> -1`, signal 4 gives
`git worktree prune && git worktree list`. Signal 5 gives nothing, and the
obvious command returns the wrong answer:

```
$ bd show agent-plugins-61f.1 --json | jq '{comment_count}'
{"comment_count": 0}

$ bd export | jq -r 'select(.id=="agent-plugins-61f.1")
    | [.comments[]?|select(.text|startswith("retro:"))] | length'
2
```

Measured 2026-08-13 across all seven beads: `bd show --json` reported
`comment_count: 0` for **every** bead, including `61f.2` and `61f.6`, which
carried a `retro:` comment each at the time. The authoritative
`bd export | jq` pipeline found 14 across the run.

The consequence is not hypothetical. The orchestrator ran the acceptance check
after wave 1 and got `5 retro: comments: 0` on all five beads. Read literally
against the rule above, that is five beads "not accepted" — five false
acceptance failures. Only knowing independently that `bd show --json` omits
comment bodies turned that into the correct reading, which is that three beads
genuinely lacked findings and two did not.

The skill already knows this pipeline is the right one: Phase 4's sprint retro
says "Harvest with the `bd export | jq` pipeline in
`references/operating-principles.md`". But that instruction lives in
**Phase 4**, and signal 5 is checked **per bead in Phase 3**, potentially many
waves earlier. An orchestrator reaching signal 5 has not yet read it.

### E2 — `--gate-baseline` cannot parse a fenced test gate

Shape: **script-defect**. Severity: **blocker**.

Exact command, run at Phase 0:

```
$ .claude/skills/bead-bonanza/scripts/pre-flight-check.sh 1786600009-7 \
    --gate-baseline
```

Observed, in the stderr report's `Gate baseline` block:

```
Gate baseline
  Format: pass   (cd cli && gofmt -l .)
  Lint: pass   (scripts/check-skills.sh)
  Build: pass   (cd cli && go build ./...)
  Test (full suite): unparsed (no command in backticks)
```

Expected: a measured baseline for the test gate.

The cause is documented in the script itself. `pre-flight-check.sh:168` —
"Extract the command from an AGENTS.md gate bullet: the first backticked
span" — and `:142-143` restate it. This project's test gate is not a
backticked span; it is a fenced multi-line `bash` block, because the three
checks must run independently rather than chained with `&&`. The parser has no representation
for that, emits `unparsed` at `:443`, sets `$key=unparsed` at `:445`, and
`BB_GATE_BASELINE` is reported as `skipped`.

**Minimal repro:** in any project whose `## bead-bonanza gates` block expresses
a gate as a fenced code block rather than an inline backticked command, run
`pre-flight-check.sh <run-id> --gate-baseline` and read the `Gate baseline`
block.

Why this is a blocker rather than a cosmetic gap: the gate it failed to measure
is the **only known-red one**. `AGENTS.md`'s own ledger says
`scripts/check-skills.sh --strict` "exits 1 on a *warning*, not an error… The
bar is **no findings beyond this baseline**, not exit 0", and — pointedly —
"Measure the baseline at pre-flight rather than quoting a number from here."
The script is the mechanism that instruction depends on, and for this gate it
silently measures nothing. The orchestrator measured it by hand instead
(`0 error(s), 1 warning(s)`), which was an unscripted improvisation at the one
step the skill mechanized specifically so it would not be.

That hand-measured baseline then earned its keep: finding `61f.5#1` records that
a bead's drafted text pushed `groom-beads/SKILL.md` from 19,828 to 20,027 bytes,
crossing the threshold and producing a **second** warning. Without a baseline of
exactly one warning, a second warning is indistinguishable from the known-red
one.

### E3 — signal 4's repair rule contradicts its own removal duty

Shape: **misleading-text**. Severity: **snag**.

`SKILL.md:151-155` defines signal 4:

> 4. **The worktree is GONE** — `git worktree prune && git worktree list` no
>   longer lists it. **This is the load-bearing signal.** Removal is the
>   orchestrator's own final act of accepting the bead, so a surviving
>   worktree is a bead not accepted, whatever the subagent reported

and `SKILL.md:169-172` says what to do when a signal fails:

> Failures of 3
> or 4 are not the orchestrator's to repair inline: the merge or the
> acceptance never happened, so leave the worktree standing (it is the
> in-flight signal) and surface it.

These two contradict for the normal case. Signal 4 correctly says removal is the
orchestrator's own final act — so between a bead's return and that act, the
worktree *always* survives. The repair rule then forbids the orchestrator from
performing the very act signal 4 says is its own.

Observed on `agent-plugins-61f.4`, a single-bead sequential wave. The
orchestrator merged the branch and ran the acceptance check:

```
  3 merged:        yes
  4 worktree gone: NO
```

Followed literally, that is a signal-4 failure, "not the orchestrator's to
repair inline", to be surfaced as `acceptance-failed: worktree survives` and
excluded from the closed count. In fact nothing was wrong: the orchestrator had
simply not yet run `git worktree remove`. It did so, re-checked, and the signal
passed.

The rule is right for the case it was written for: a *batch* subagent that was
told to remove its own worktrees and left one standing. It does not distinguish
that from the orchestrator's own pending act, and the two are indistinguishable
from the output of `git worktree list`.

### E4 — the tier-batching cost test never reaches the planner

Shape: **gap**. Severity: **snag**.

`SKILL.md:599-605` gives the rule that decides whether batching is worth a tier
over-route:

> Split the bucket WHEN more than one bead would be dragged above its
> tier, OR when any bead would be dragged more than one tier. Weigh that
> against what batching saves: batching exists to avoid repeated full
> `<test>` runs, so WHERE Phase 0's gate baseline (`scripts/pre-flight-check.sh`)
> shows the suite completing in seconds, the saving is negligible and the
> correct tier wins

Bucketing is not the orchestrator's job. Phase 2.5 states: "Planning is
delegated; the manifest is not", and the partition is produced by the
`plan-sprint` skill in a subagent. That subagent reads `plan-sprint`, not
`bead-bonanza/SKILL.md` — so the cost test above is not in its context.

Observed: the planning subagent bucketed `61f.4` (tier `advanced`) and `61f.5`
(tier `standard`) into one batch at `advanced`, dragging `61f.5` a full tier up.
It had the measured baseline (the orchestrator passed it in the dispatch
prompt)
— and its returned rationale weighed batching's saving without applying the
seconds-vs-minutes test. The orchestrator caught it and split the wave
(amendment 2 in the run's `dispatch.md`); this run's gate suite is three shell
scripts finishing in seconds, so the batch would have bought nothing.

No rule at Phase 2.5 requires the orchestrator to pass the tier-batching cost
test to `plan-sprint`, and nothing in the enumerated hand-off list mentions it.

### E5 — the Monitor remedy landed unverified against the measured cap

Shape: **gap**. Severity: **snag**.

At the version run (1.4), the Monitor step gave no lifetime guidance at all:
finding `61f.6#1` records that `persistent`, `timeout_ms` and `TaskStop`
returned zero hits across `plugin/skills/bead-bonanza/`. The orchestrator
therefore chose the arming parameters itself, and measured this:

```
requested: timeout_ms = 3600000, persistent = true
returned:  "Monitor started (task bhjo96kee, timeout 1800000ms)"
```

3,600,000 ms requested, 1,800,000 ms granted — a 30-minute runtime cap, below
the `max 3600000` the Monitor tool's own schema documents for the `timeout_ms`
input. Both are consistent: the schema bounds the accepted input, the runtime
applies a lower ceiling.

The remedy this run added (bead `agent-plugins-61f.6`, now at
`references/progress-streaming.md`) prescribes `persistent: true` plus an
explicit `TaskStop`. **That remedy is unverified against the measured cap.** The
Monitor tool documents `timeout_ms` as "Ignored when persistent is true", yet
the runtime reported a timeout anyway. The orchestrator stopped the Monitor
after roughly 24 minutes of life — inside the 30-minute cap — so the run
produced
no evidence either way.

If `persistent: true` does not defeat the cap, the skill now prescribes a remedy
that does not work and relegates the re-arm instruction to a fallback, when it
is in fact the whole rule.

### E6 — no rule for a durable trigger that fires while the run is still live

Shape: **gap**. Severity: **polish**.

`references/progress-streaming.md` specifies arming the durable trigger and
cancelling it in Phase 4. It says nothing about the trigger *firing*: `grep -n
'fired\|fires\|drained'
plugin/skills/bead-bonanza/references/progress-streaming.md`
returns no matches.

Observed: this run's trigger fired at 06:18Z while the orchestrator was in
Phase 4 cleanup, with the run drained but not yet reported. It resolved
correctly only because the orchestrator had improvised a clause into the trigger
message — "If the run has already drained, cancel this trigger rather than
acting on it" — which no rule asked for. Without it, the woken turn's
instruction is to invoke `recover-sprint` on a finished sprint.

## Proposals

### P1 — name signal 5's command, and name the one that lies

Severity: **blocker**. Shape: gap. Evidence: E1.

**Change.** At `SKILL.md:156`, extend signal 5 so it carries a command like
every other signal in the list:

> 5. **Findings were recorded** — at least one `retro:` comment on the bead,
>    or the subagent's explicit "clean bead, no findings". A silent bead is
>    neither, and is not accepted. Check with the export pipeline, never with
>    `bd show`:
>
>    ```
>    bd export | jq -r 'select(.id=="<bead-id>")
>      | [.comments[]?|select(.text|startswith("retro:"))] | length'
>    ```
>
>    `bd show <id> --json` reports `comment_count: 0` even for a bead that has
>    comments, so checking this signal with it fails every bead in the run at
>    once. That is the same reason Phase 4's harvest mandates the export
>    pipeline; the check happens here, waves earlier.

### P2 — make `--gate-baseline` fail loudly on a gate it cannot parse

Severity: **blocker**. Shape: script-defect. Evidence: E2.

**Change, two parts.**

1. **Teach the parser fenced blocks.** In `scripts/pre-flight-check.sh`'s gate
   extractor (`:168`), WHERE no backticked span follows the label on the bullet
   line, look ahead for a fenced code block before the next bullet and take its
   body as the command.

2. **Stop treating `unparsed` as a benign outcome.** WHERE a gate still cannot
   be parsed, the script SHALL report it as a *blocking* verdict line rather
   than an informational one, naming the gate:

   > `BLOCKED: gate "Test (full suite)" could not be parsed, so no baseline was
   > measured for it. A gate with no baseline cannot be held to "no findings
   > beyond baseline". Measure it by hand and record it in dispatch.md before
   > dispatching, or express it as an inline backticked command in AGENTS.md.`

   Exit non-zero, as it already does for `worktree.baseRef`. An unmeasured
   known-red gate is the same class of hazard as a bad base ref: silent, and
   discovered only after beads have built on it.

### P3 — distinguish a pending removal from a failed one

Severity: **snag**. Shape: misleading-text. Evidence: E3.

**Change.** At `SKILL.md:169-172`, replace the blanket rule for signal 4:

> A signal-4 failure means the worktree survived **after the orchestrator
> performed its own removal** — check by doing the removal first
> (`git worktree remove <path>`), then re-running
> `git worktree prune && git worktree list`. A worktree standing before that
> act is not a failure; it is the act still pending, and it is the normal state
> between a bead's return and its acceptance. WHERE removal itself fails, or a
> worktree the orchestrator never created survives, that is the real signal-4
> failure: leave it standing (it is the in-flight signal) and surface it.
> Failures of 3 remain unrepairable inline.

### P4 — pass the tier-batching cost test to the planner

Severity: **snag**. Shape: gap. Evidence: E4.

**Change.** In Phase 2.5's enumerated hand-off to `plan-sprint`, add the cost
test to what the orchestrator passes, alongside the gate baseline it already
sends:

> - The tier-batching cost test, resolved for this run: state whether the
>   measured gate suite completes in **seconds** or in **minutes**. Batching
>   exists to avoid repeated full `<test>` runs, so where it is seconds the
>   saving is negligible and a bucket SHALL NOT drag any bead above its
>   assigned tier; where it is minutes, one bead one tier high is the cheaper
>   error.

The rule stays owned by `bead-bonanza`; what changes is that the agent applying
it receives it. This is cheaper than duplicating the rubric into `plan-sprint`,
and it keeps one source of truth.

### P5 — settle whether `persistent` defeats the cap

Severity: **snag**. Shape: gap. Evidence: E5.

**Change, two parts.**

1. **Record what is measured**, in `references/progress-streaming.md` beside the
   arming guidance:

   > Measured 2026-08-13 (Claude Code, remote environment): a Monitor requested
   > with `timeout_ms: 3600000` reported `timeout 1800000ms`. The tool schema
   > bounds the *input* at 3,600,000; the runtime applies a lower ceiling. Do
   > not infer a Monitor's lifetime from the value you requested.

2. **State the remedy's status honestly** until a run settles it:

   > Whether `persistent: true` escapes that ceiling is **unverified** — the run
   > that measured the cap stopped its Monitor inside it. Until a run observes a
   > persistent Monitor alive past 30 minutes, treat re-arming as the load-bearing
   > rule and `persistent: true` as an optimisation: WHEN a Monitor stops while
   > its subagent is still live, re-arm it; expiry is not information about the
   > bead.

Recording an unverified remedy as unverified is the point. The alternative —
leaving `persistent: true` reading as settled — is how the next run inherits a
belief nobody tested.

### P6 — give the durable trigger a drained-run branch

Severity: **polish**. Shape: gap. Evidence: E6.

**Change.** In `references/progress-streaming.md`, where the trigger's
self-contained message is specified, add a required clause:

> The message SHALL also tell the woken turn what to do when the run is already
> finished: "If the run has already drained, cancel this trigger rather than
> acting on it." A trigger can fire between the last bead closing and Phase 4
> cancelling it, and without this clause the woken turn's only instruction is to
> recover a sprint that needs no recovery.

## Considered and Not Proposed

- **The batch preamble branched every bead in a batch from `HEAD`.** Real,
  observed, and it forced this run to hand-apply a chaining override to the very
  batch carrying the fix (`61f.2#1` quotes the pre-fix text). Not proposed
  because bead `agent-plugins-61f.2` fixed it during this run; the text no
  longer exists. Recorded so a recurrence has something to attach to.
- **The Monitor step gave no lifetime guidance, so its default `timeout_ms` of
  5 minutes applied under a rule that fires at 5 minutes.** Fixed this run by
  bead `agent-plugins-61f.6`. Only the unverified-remedy residue is proposed
  (P5).
- **Four grooming defects in the beads this run executed** (`61f.5#1`,
  `61f.4#1`, `61f.4#4`, `61f.1#1`). All real, all recorded verbatim in the
  sprint report — but they are defects in how the beads were *groomed*, not in
  `bead-bonanza`, whose readiness gate correctly passed beads that were in fact
  Ready. They belong against `groom-beads` or the Definition of Ready, and the
  byte-budget one (`61f.5#1`) is the strongest candidate for its own RFC.
- **`updated_at` travelling with a version bump** (`61f.4#3`). A repository
  convention gap, not a skill defect: no `bead-bonanza` text claims to govern
  frontmatter dates. Recorded for whoever owns the versioning standard.
- **Zero tier escalations across 7 beads.** Consistent with good routing and
  equally consistent with routing everything too high, since over-routing
  succeeds silently. The skill's own rule reserves suspicion for runs of ≥10
  beads, and this run had 7 — so the observation supports nothing in either
  direction. Noted because the absence of evidence is the observation.
- **`bb-close-bead.sh`, `bb-progress.sh` and `bb-backup-run.sh` all behaved
  exactly as their published interfaces describe**, across 7 beads and 3
  snapshots, including the JSONL-export lag the ledger warns about. Recorded as
  a design that paid off; no change proposed.

## Non-Goals

- Re-litigating the wave partition. The file-collision analysis was correct, no
  wave was resequenced, and the one plan-level miss is P4.
- The size of `SKILL.md`. It remains over the re-injection cap and this run
  added ~1,500 bytes to it; that is tracked as beads, not as a skill defect.
- Changing the acceptance check's five-signal structure. The signals are the
  right ones; P1 and P3 make two of them checkable.

## Adoption

P1, P3, P4, P5 and P6 are text-only and mutually independent. P2 is the only one
touching a script, and its two parts are separable: part 2 (fail loudly) is a
few lines and delivers most of the safety, while part 1 (parse fenced blocks) is
the real fix. Adopting part 2 alone is a coherent outcome.

P1 and P3 both edit the Phase 3 acceptance-check list and should be sequenced
onto one another rather than run concurrently; everything else is disjoint.

No proposal invalidates an in-flight run: a run mid-execution under the current
text continues to work, and P2's stricter verdict applies at the next Phase 0.

## Resolution

<!-- Filled at triage. -->
