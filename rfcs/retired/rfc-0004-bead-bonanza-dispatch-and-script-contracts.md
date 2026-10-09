---
title: "RFC: bead-bonanza — dispatch fallback, script contracts,
  baseline gates"
description: Five evidence-backed fixes from a three-bead run — a
  fallback when no subagent tool exists, published usage lines for the
  bundled scripts, a validation-order fix in bb-progress.sh, a probe
  before the telemetry command, and a baseline verdict for the gate
  profile
slug: rfc-0004-bead-bonanza-dispatch-and-script-contracts
type: rfc
status: retired
origin: self
skill: bead-bonanza
skill_version: "1.2"
model: claude-opus-5
effort: high
runs_observed: 1
severity: blocker
plugin_version: "0.7.0"
---

# RFC: bead-bonanza — dispatch fallback, script contracts, baseline gates

## Summary

`bead-bonanza` v1.2 was run over three groomed beads
(`agent-plugins-fiu`, `agent-plugins-x8t`, `agent-plugins-8bx`) in this
repo, editing `plugin/skills/spec/`. All three closed and every gate
passed, so the run succeeded — but it succeeded by improvising around
the skill in two places and by guessing at a contract the skill
explicitly tells the reader not to look up. The session harness exposed
no `Agent` tool, so the skill's entire dispatch apparatus was
unavailable and the run had to invent a whole-queue main-loop mode; and
the three bundled scripts are declared "executed, not read" while their
argument shapes are published nowhere, which produced two failed
invocations and left a directory sitting at the path the progress file
needed. The five proposals give the skill a declared fallback when it
cannot delegate, publish the scripts' usage lines, fix the validation
order in `bb-progress.sh`, probe before the telemetry command, and make
the gate profile's baseline an explicit Phase 0 output.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code, remote (web) session |
| Task the skill was applied to | Three groomed beads amending the `spec` skill |
| References / scripts / templates invoked | See the list below the table |
| Run dates | 2026-07-30 |

The beads were `agent-plugins-fiu`, `agent-plugins-x8t` and
`agent-plugins-8bx`, which added an authoring-workflow self-check step
and its new checklist reference, noted slug/filename independence in the
standards template, and named a destination for newly authored
documents. Invoked during the run:

- `plugin/skills/bead-bonanza/SKILL.md` v1.2 — the skill under retro.
- `plugin/skills/bead-bonanza/scripts/bb-progress.sh` — progress events.
- The skill's own `references/shared/work-order-definition-of-ready.md`
  — the Phase 1.5 readiness gate.
- Repo gates from the project's profile: `scripts/check-skills.sh`,
  `scripts/sync-shared-references.sh`, `scripts/dogfood-install.sh`.
- Not invoked: `bb-verify-skip.sh` (no bead returned `skipped`) and
  `bb-close-bead.sh` (see Considered and Not Proposed).

## Evidence

1. **[gap · blocker] No rule for a harness that exposes no subagent
   tool.** The skill's premise is delegation —
   `plugin/skills/bead-bonanza/SKILL.md:18-19` "Each bead runs in a
   fresh subagent so context doesn't compound" — and its tier routing
   is expressed directly in terms of the tool:
   `plugin/skills/bead-bonanza/SKILL.md:452-453` "**Dispatch at the
   resolved model** — pass `model: \"<resolved>\"` on the `Agent` tool
   call for that bead's subagent." No `Agent` tool existed in this
   session. The only single-agent escape the skill offers is scoped to
   an individual bead, not to the harness:
   `plugin/skills/bead-bonanza/SKILL.md:457-460` "**Main-loop beads** —
   a bead at the orchestrator's own tier with a note saying \"do not
   delegate\" is implemented by the orchestrator directly". None of the
   three beads carried such a note. The agent improvised: it declared
   whole-queue main-loop execution, kept the per-bead ceremony, and
   collapsed the bucket plan to sequential-with-batched-gates — none of
   which the skill sanctions. A consequence the skill does not
   anticipate: Phase 4's escalation table
   (`plugin/skills/bead-bonanza/SKILL.md:596-616`), which the skill
   calls "the only direct evidence a run produces about whether
   grooming predicted capability correctly", was reported as "none"
   even though one model ran all three beads and no tier was honored at
   all. A reader cannot distinguish that from calibrated routing.

2. **[gap · blocker] The bundled scripts' argument shapes are published
   nowhere, and the skill forbids reading them.**
   `plugin/skills/bead-bonanza/SKILL.md:751` is the whole contract the
   caller is given: "| Emitting a progress event | run
   `scripts/bb-progress.sh` |" — no parameters. Eleven lines later,
   `plugin/skills/bead-bonanza/SKILL.md:761` closes the door on
   looking: "The three `scripts/` entries are **executed, not read**."
   The agent guessed, and guessed wrong twice in one invocation: it
   passed the progress *file* path where the script wants the *run
   dir*, and a free-text field where the script wants `key=value`. The
   usage line exists only inside the file the reader is told not to
   open (`bb-progress.sh:16-21`). Recovering it cost two failed calls
   and, via finding 3, a manual `rmdir`.

3. **[script-defect · snag] `bb-progress.sh` creates the run directory
   before it validates its arguments, so a rejected invocation leaves
   state that breaks every later correct one.** Exact command:

   ```
   bb-progress.sh "$RUN_DIR/agent-plugins-fiu.progress.jsonl" \
     agent-plugins-fiu claimed "main-loop dispatch, tier advanced"
   ```

   Observed: `bb-progress: extra fields must be key=value, got
   'main-loop dispatch, tier advanced'`, exit 2 — *and* a new directory
   at `$RUN_DIR/agent-plugins-fiu.progress.jsonl`. Expected: rejection
   with no filesystem side effect. The cause is ordering: the `mkdir
   -p "$run_dir"` at `bb-progress.sh:46-50` runs before the
   `key=value` check at `bb-progress.sh:60-63`. Because the bad
   run-dir argument was the progress path itself, the directory now
   occupies exactly the path `bb-progress.sh:70` appends to, so every
   subsequent *correct* invocation for that bead fails and cannot
   recover by retrying. Minimal repro:

   ```
   RD=/tmp/bb-repro
   bb-progress.sh "$RD/some-bead.progress.jsonl" some-bead claimed "a b"
   # -> exit 2, but $RD/some-bead.progress.jsonl now exists as a dir
   bb-progress.sh "$RD" some-bead claimed k=v
   # -> "$RD/some-bead.progress.jsonl: Is a directory", exit 1
   ```

   Both halves were run and both behaved as shown. Exit 1 is also
   outside the script's own documented contract, which lists only 0 and
   2 (`bb-progress.sh:23-25`).

4. **[script-defect · snag] Phase 0's mandatory telemetry command
   assumes a standalone `dolt` CLI that a `bd` install need not
   provide.** `plugin/skills/bead-bonanza/SKILL.md:74-84` is emphatic —
   "Just kill it — don't ask, don't suggest" — and prescribes `dolt
   config --global --add metrics.disabled true`, with `dolt config
   --global --list` as the skip-probe. On this machine, with `bd`
   1.1.2 installed and working, both commands fail: `dolt: command not
   found`. Observed: command not found, nothing configured. Expected
   per the skill: the config written, or a clean skip. The skill offers
   no alternative, so the agent proceeded on the environment's
   pre-exported `BD_DISABLE_TELEMETRY=1` — a fallback the skill never
   mentions. Repro: install `bd` 1.1.2 via the Go installer, then run
   `dolt config --global --list`.

5. **[gap · snag] No rule for a gate profile that is already red on an
   untouched tree.** Phase 0 step 1
   (`plugin/skills/bead-bonanza/SKILL.md:58-73`) loads the profile and
   warns "Do not invent gates or a lineup from memory", and
   `references/project-setup.md` covers a *missing* profile. Nothing
   covers a profile that is present and failing before any bead is
   touched. This repo's profile declares `<test>` as
   `scripts/check-skills.sh --strict && ...`; on the untouched tree
   `scripts/check-skills.sh --strict` exits 1, because
   `plugin/skills/bead-bonanza/SKILL.md` is itself 770 lines against a
   500-line guidance and `--strict` promotes that warning to an error.
   The agent had to decide unilaterally that a red declared gate was
   pre-existing rather than disqualifying, and to run the non-strict
   form per bead instead. The closest existing rule is the
   environmental-failure ledger at
   `plugin/skills/bead-bonanza/SKILL.md:714-716` — "Suites that fail —
   or silently PASS for environmental reasons — in the local sandbox
   get diagnosed ONCE" — which is scoped to *environmental* causes and
   does not reach a genuine, reproducible, repo-owned failure.

## Proposals

### P1 — Declare a main-loop fallback when delegation is unavailable [blocker]

**Problem.** Every dispatch mode in Phase 2.5 and Phase 3, and the tier
routing at `plugin/skills/bead-bonanza/SKILL.md:452-453`, presume an
`Agent` tool. When none exists the skill goes silent, and the only
adjacent rule
(`plugin/skills/bead-bonanza/SKILL.md:457-460`) is scoped to a bead
carrying a "do not delegate" note. Finding 1.

**Change.** Add a step to Phase 0, after the run-scratch step:

> 7. **Confirm subagent dispatch is available.** Every dispatch mode
>    below assumes an `Agent` tool. WHEN the harness exposes none, do
>    not abandon the run — the queue degrades to **whole-queue
>    main-loop execution**: the orchestrator implements every bead
>    itself at the full per-bead ceremony (claim, gates, retro, atomic
>    close, single commit), the buckets collapse to sequential order
>    with gates batched per bucket, and tiers become advisory rather
>    than routed. Say so once at run start.

and a sentence to Phase 4's tier-escalation block, after the existing
"State `none` explicitly" paragraph:

> WHEN the run executed in whole-queue main-loop mode, say that
> alongside the table: one model ran every bead, so "none" records that
> no routing occurred — it is not evidence about calibration in either
> direction.

**Evidence.** Finding 1. Under the same model the next run declares the
fallback instead of inventing one, and its Phase 4 report can no longer
be read as a calibration result when no routing happened.

### P2 — Publish each bundled script's usage line [blocker]

**Problem.** `plugin/skills/bead-bonanza/SKILL.md:751-753` names the
three scripts with no parameters, and
`plugin/skills/bead-bonanza/SKILL.md:761` tells the reader they are
"executed, not read" — so the only place the argument shape exists is
the file the reader is told not to open. Finding 2.

**Change.** Carry the usage into the table rows:

> | Emitting a progress event | run `scripts/bb-progress.sh <run-dir> <bead-id> <phase> [key=value ...]` |
> | Checking a `skipped` claim against the tree | run `scripts/bb-verify-skip.sh <bead-id> [worktree-path]` |
> | Closing a bead: close + export + stage + commit | run `scripts/bb-close-bead.sh <bead-id> <retro-file>` |

(The last two signatures must be read off those scripts' own headers
when the change is applied; the point is that the row carries the
signature.) Then amend the sentence at
`plugin/skills/bead-bonanza/SKILL.md:761` to close the loop:

> The three `scripts/` entries are **executed, not read** — their
> signatures are above, and each script prints its usage when invoked
> with no arguments, so a forgotten parameter never requires opening
> the file.

**Evidence.** Finding 2. The two failed invocations in this run were
both shape errors that a published signature would have prevented, and
the first of them is what triggered finding 3.

### P3 — Validate all args in `bb-progress.sh` before creating [snag]

**Problem.** `bb-progress.sh:46-50` creates the run directory before
the `key=value` validation at `bb-progress.sh:60-63`, so an invocation
the script itself rejects still mutates the filesystem — and when the
mistaken run-dir argument is the progress path, the created directory
permanently blocks the append at `bb-progress.sh:70`. Finding 3.

**Change.** Move the directory creation below the argument loop, and
guard the append. Delete lines 46-51 and replace line 70 with:

```bash
if [ ! -d "$run_dir" ]; then
  # Create it rather than failing: a subagent that starts before the
  # orchestrator's mkdir landed should still record its own progress.
  mkdir -p "$run_dir" || die "cannot create run dir: $run_dir"
fi
[ -w "$run_dir" ] || die "run dir is not writable: $run_dir"

out="$run_dir/$bead.progress.jsonl"
[ ! -e "$out" ] || [ -f "$out" ] || die "progress path is not a file: $out"
printf '%s\n' "$json" >> "$out" || die "cannot append to $out"
```

This also brings the failure back inside the documented contract at
`bb-progress.sh:23-25`, which lists only exit 0 and exit 2.

**Evidence.** Finding 3. With the reorder, the repro's first command
exits 2 having touched nothing, and the second command — the correct
one — succeeds.

### P4 — Probe for the `dolt` CLI, and name the fallback [snag]

**Problem.** `plugin/skills/bead-bonanza/SKILL.md:74-84` mandates `dolt
config --global --add metrics.disabled true` and offers `dolt config
--global --list` as the skip-check. Both fail with `dolt: command not
found` on a working `bd` 1.1.2 install that ships no standalone `dolt`
binary, and the skill names no alternative. Finding 4.

**Change.** Insert before the fenced command in step 2:

> Probe first — `command -v dolt` — because a `bd` install need not
> ship a standalone `dolt` CLI (observed with `bd` 1.1.2). WHEN there
> is no `dolt` on PATH, skip the config write and the deny-list edit,
> export `BD_DISABLE_TELEMETRY=1` into the run environment instead,
> and say which path was taken in the same one-line note.

Triage should confirm that `BD_DISABLE_TELEMETRY=1` covers the Dolt
event calls the current step is aimed at; this run observed only that
it is the switch this repo's environment already sets.

**Evidence.** Finding 4. The next run on a `dolt`-less install takes a
declared path instead of stepping over a failed mandatory instruction.

### P5 — Record the gate profile's baseline verdict in Phase 0 [snag]

**Problem.** Phase 0 step 1 loads the gate profile but never runs it,
so the first time `<test>` is executed it is already mixed with the
run's own changes. When the declared gate is red on a clean tree — as
`scripts/check-skills.sh --strict` is in this repo — the skill has no
rule, and the ledger principle at
`plugin/skills/bead-bonanza/SKILL.md:714-716` does not reach it,
because that principle is scoped to *environmental* failures. Finding
5.

**Change.** Append to Phase 0 step 1:

> Then run the profile's `<test>` once against the untouched tree and
> record the verdict as the run's **baseline**, in the ledger
> alongside the environmental entries. A gate that is red before any
> bead is touched is a pre-existing failure, not this run's: per-bead
> checks are judged against that baseline, not against green, and the
> Phase 4 report names it. A baseline-red gate is not a reason to stop
> — but it must never silently absorb a regression this run caused, so
> record the exact failing check, not just "red".

**Evidence.** Finding 5. The next run in this repo starts from a
recorded "`--strict` red at baseline: bead-bonanza SKILL.md 770 lines"
instead of re-deciding mid-run whether the queue is disqualified.

## Considered and Not Proposed

- **`bb-close-bead.sh` commits, which no-commit runs cannot use** — the
  orchestrator running this session forbade commits, so the close
  ceremony was performed by hand. Excluded on instruction as a
  constraint of this run rather than a skill defect, and the exclusion
  is right on the merits: the ordering the script encodes is also
  stated in prose at `plugin/skills/bead-bonanza/SKILL.md:656-660`
  ("the close happens via `bd update --status closed --body-file
  <retro-file>` BEFORE staging"), so nothing was actually missing.
- **Stale Phase 0 cross-references** — this skill's `SKILL.md:147` and
  `:529` both say "Phase 0 step 5" for the `TaskCreate` ban, which
  is step 4 (`:110`); step 5 (`:116`) is Push mechanics. Separately,
  `:447` cites "Phase 0 step 1a", but step 1 (`:58`) has bullets, not
  sub-numbers. Both pointers are wrong, but neither led this run to a
  wrong action — Phase 0 was read in full before Phase 1 — so they fail
  the misleading-text shape test, which requires an instruction that
  produced a wrong action. Recorded here so triage can fix them
  cheaply if it wants to.
- **`bead-bonanza/SKILL.md` is 770 lines against a 500-line guidance**
  — true, and it is the direct cause of the baseline-red gate in
  finding 5, but on its own it is a property of this repo's check
  rather than guidance that misled anyone. P5 proposes the general
  rule; splitting or trimming the skill is a separate decision with no
  run evidence behind it.
- **Phase 0 step 6's draft-PR question** — could not be asked (the
  session is non-interactive) and pushing was forbidden. Not a defect:
  the step is already conditioned on "If the repo has PR-triggered CI",
  and the constraint came from the run, not the skill.
- **Phase 1.5, Phase 2 and Phase 2.5 lean paths** — the skill's own
  guidance to skip the decisions doc and the formal dispatch plan for a
  small, fully-groomed queue worked exactly as written and saved real
  ceremony. No finding; recorded because it was examined.

## Non-Goals

- Changing the dispatch mode taxonomy, the bias toward batching, or the
  `light | standard | advanced | frontier` tier vocabulary. P1 adds a
  fallback for harnesses that cannot delegate; it does not alter how
  routing works where delegation is available.
- Restructuring or shortening `bead-bonanza/SKILL.md`.
- Any change to the three beads this run implemented, or to the `spec`
  skill edits they landed — those are settled work, not RFC material.
- Editing the `bead-bonanza` skill as part of this retro. Resolution is
  `triage-rfc`'s, after human approval.

## Adoption

All five proposals are independent and can be adopted or rejected
separately. P1, P2, P4 and P5 touch only
`plugin/skills/bead-bonanza/SKILL.md`; P3 touches only
`plugin/skills/bead-bonanza/scripts/bb-progress.sh`. P2 and P3 are
complementary rather than coupled — P2 removes the most likely trigger
for the bad invocation, P3 removes its lasting consequence — and either
is worth landing without the other.

Backward compatibility: every change is additive. P1's fallback only
engages where no `Agent` tool exists, so runs in delegating harnesses
are unaffected. P3 strictly narrows the cases in which the script
mutates the filesystem; no currently-succeeding invocation changes
behavior. P4 adds a branch that only fires where the current command
already fails. P5 adds one Phase 0 command and one ledger line.

If P1 and P2 land together, the two improvisations this run had to make
both become declared paths, and the remaining three remove a
filesystem trap, a dead mandatory command, and an ambiguity about
whose failure a red gate is.

## Resolution

**Triaged 2026-07-31.** 5 proposals: 4 adopted, 1 deferred.

| Proposal | Outcome | Bead |
|----------|---------|------|
| P1 — main-loop fallback when delegation is unavailable | **adopted.** The skill now defines the fallback (sequential, full per-bead ceremony, parallel dispatch modes dropped, context as the scarce resource) and requires Phase 4 to report "no routing occurred" rather than `none`. | `agent-plugins-vtl` |
| P2 — publish each bundled script's usage line | **adopted.** All three usage lines are now in the skill, verified against the scripts, with the script named authoritative on disagreement. | `agent-plugins-3rf` |
| P3 — validate args before creating the run directory | **adopted.** Every argument is validated before any filesystem write; the RFC's repro no longer leaves a stray directory. | `agent-plugins-fme` |
| P4 — probe for the `dolt` CLI | **adopted.** Phase 0 probes first and treats an absent `dolt` as a normal configuration, not a run failure. | `agent-plugins-cr2` |
| P5 — record the gate profile's baseline verdict | **deferred** → `agent-plugins-1a9`, `theme:gate-baseline`. Entangled with an unsettled question: the standard states the body budget in CHARACTERS while `check-skills.sh` measures LINES. Bead is `deferred` / `blocked-needs-info` pending that decision. | `agent-plugins-1a9` |

**2026-08-10:** bead `agent-plugins-1a9` was never persisted to `.beads/issues.jsonl` and does not exist in any branch's database; the deferral was lost with its container. The blocking question (body budget in characters vs `check-skills.sh` measuring lines) is resolved: the binding constraint is the 5,000-token re-injection cap, so the threshold is a size in bytes. Both halves of P5 are now tracked by `agent-plugins-92q.2` (check-skills.sh: threshold skill bodies on bytes) and `agent-plugins-92q.3` (pre-flight-check.sh: mechanize Phase 0 on macOS and Linux).

**Files changed:** `plugin/skills/bead-bonanza/SKILL.md`,
`plugin/skills/bead-bonanza/scripts/bb-progress.sh`.

**Version bumps:** `bead-bonanza` 1.2 → 1.3.
