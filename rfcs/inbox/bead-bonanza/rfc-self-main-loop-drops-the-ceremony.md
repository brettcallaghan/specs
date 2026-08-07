---
title: "RFC: The per-bead ceremony rides on the dispatch step, so the
  main-loop fallback silently drops it"
description: Every progress event and retro file is written by a filled-in
  subagent template; the main-loop fallback claims the ceremony runs
  "unchanged" but has no step that reads the template, so five beads
  closed with no ledger and no retro anchor and nothing detected it
slug: rfc-self-main-loop-drops-the-ceremony
type: rfc
status: ready-for-review
updated_at: 2026-08-07T00:00:00Z
origin: self
skill: bead-bonanza
skill_version: "1.3"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.15.0"
---

# RFC: The ceremony rides on dispatch, and main-loop has no dispatch

## Summary

The second half of `bead-bonanza` run `1786121603-ri` closed five beads
(`9kb.5`–`9kb.9`) with no progress ledger and no retro file for any of
them, while the first half wrote both for every bead. Nothing in the
skill noticed: the orchestrator recorded all five as closed, and later
cited a progress file that does not exist as proof a subagent was alive.
The cause is structural rather than an agent lapse — every progress
event and every retro write in this skill lives inside
`references/subagent-prompts.md`, a template the orchestrator fills in
**at dispatch**. `SKILL.md:474` and `SKILL.md:502` promise the main-loop
fallback runs the ceremony "unchanged", but that path never dispatches,
so nothing ever reads the template and the ceremony has no carrier.
These proposals give the ceremony an owner that survives the fallback,
and give the orchestrator an acceptance check so its absence is loud.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code on the web (remote container session) |
| Task the skill was applied to | The `agent-plugins-9kb` reference-integrity epic: 12 beads across two repos, on branch `claude/reference-integrity-blueprint-groom-xc7d30` |
| References / scripts / templates invoked | `bead-bonanza` SKILL.md v1.3; `references/subagent-prompts.md`; `references/recovery.md`; `scripts/bb-progress.sh`; `scripts/bb-close-bead.sh`; `skill-rfc-template` v0.3 |
| Run dates | 2026-08-07 (single continuous run, 17:21–22:24) |

**On the withheld model field.** This session's operating instructions
forbid writing the model identifier into any artifact committed to a
repository. The frontmatter placeholder is deliberate and matches the two
prior `bead-bonanza` RFCs in this directory. Triage needs the value to
scope the fix: it is in the session transcript, and the human copying
this RFC can fill it in.

**Relationship to the two prior RFCs in this directory.** This RFC
covers the **same run** as
`rfc-self-run-dir-discoverability-and-orchestrator-recovery.md` — that
one was written at 21:43 (commit `ad1c285`) about the run's first half
and its 2h16m death; this one observes the second half, after recovery.
It is therefore **not an independent recurrence** of anything in that
RFC, and `runs_observed` stays at 1. Findings that merely corroborate
its proposals are listed under Considered and Not Proposed rather than
re-proposed.

## Evidence

1. **[gap · blocker] Five beads closed with no ceremony artifacts, and
   no step in the skill checks.** `.bonanza/1786121603-ri/` contains a
   `.progress.jsonl` and a `.retro.md` for `tix`, `ml9`, `j7f`, and
   `9kb.1`–`9kb.4`. For `9kb.5` it holds a progress file with two events
   and **no** retro file; for `9kb.6`, `9kb.7`, `9kb.8`, `9kb.9` it holds
   **nothing at all**. All five beads are closed in `bd` (`bd list`
   reports only `9kb` and `9kb.10` still open) and all five have landing
   commits: `bb7ecc3`, `a43bcf9`, `da9d8bd`, `2d57aef`, `7761db5`. The
   orchestrator recorded each as complete. `SKILL.md:616-618` calls the
   retro file "the recovery anchor, and it is non-negotiable — without
   it, recovery is guesswork instead of a procedure", but the obligation
   is stated only against the subagent, and Phase 3 has no step where
   the orchestrator verifies the artifact before accepting an outcome.

2. **[gap · blocker] The ceremony is carried entirely by the dispatch
   template, which the main-loop fallback never reads.** Every
   `bb-progress.sh` call and the retro write appear only in
   `references/subagent-prompts.md` (lines 68, 70, 71, 77, 79-85, 89, 90,
   99, 145, 178), each as a `{RUN_DIR}` / `{SCRIPTS}` / `{ID}` template
   slot. `SKILL.md:637-643` instructs: "**Read that file at dispatch**
   and fill in every `{PLACEHOLDER}` before sending." The main-loop
   fallback at `SKILL.md:495-502` is entered when "some harnesses expose
   no `Agent` tool at all, and some sessions forbid it" — the condition
   that held for this session — and it says the orchestrator implements
   every bead "following the full per-bead ceremony unchanged (claim,
   gates, retro, atomic close, single commit)". That path has no
   dispatch and no send, so the only document containing the ceremony's
   commands is never opened. *Inference, marked as such:* the artifact
   boundary falling exactly at `9kb.5` — the first bead after the run was
   interrupted and resumed — is consistent with the run continuing in
   main-loop mode from that point, but the skill text above is the
   defect regardless of which bead the mode changed at.

3. **[gap · blocker] A stale ledger is affirmatively misleading, and no
   rule says so.** `agent-plugins-9kb.5.progress.jsonl` ends at
   `{"phase":"implementing"}` with mtime 18:29, yet bead `9kb.5` was
   committed at 21:06 (`bb7ecc3`) and four further beads landed after it.
   `references/recovery.md:40-55` presents the event sequence as the
   status surface without stating what a ledger that stops short of
   `committed` does and does not prove. Acting on it, the orchestrator
   read a trail ending at `implementing` as "positive proof of a live
   agent at a known phase" in a heartbeat message, and in the same
   message cited `agent-plugins-9kb.9.progress.jsonl` — a file that has
   never existed in this run directory (`git log --name-status --
   .bonanza/` shows one adding commit, `02c3cc8`, listing 17 paths, none
   of them `9kb.6`–`9kb.9`; the tree is clean, so nothing was deleted).

4. **[gap · snag] No rule for a bead that cannot land in one commit.**
   `SKILL.md:718-722` states "**One commit per bead, always.** Every
   closed bead = exactly one commit on the branch." Bead `9kb.5` landed
   in two: `e3f2777` "WIP agent-plugins-9kb.5 (PREG): profile
   registration, snapshots pending" at 20:52, then `bb7ecc3` at 21:06.
   The rule is absolute and has no interrupted-bead exception, so the
   agent improvised a WIP commit the invariant forbids rather than
   following a stated resume path.

5. **[gap · snag] The skill's own size makes a project gate permanently
   red, and the skill says nothing about that case.** `check-skills.sh`
   warns at 500 lines (`scripts/check-skills.sh:118`); `bead-bonanza`'s
   SKILL.md body is 859 lines, so `./scripts/check-skills.sh --strict`
   prints `0 error(s), 1 warning(s)` and `RESULT: FAIL` on a clean tree
   in this repo — the sole finding is `bead-bonanza` itself. The run's
   dispatch plan carries a hand-written exception for it (ledger entry 2
   in `.bonanza/1786121603-ri/dispatch.md`: "KNOWN RED, pre-existing …
   Bar is no NEW findings, not exit 0") that every prompt had to repeat.
   Nothing in Phase 0's gate-profile step tells the orchestrator how to
   baseline a gate the skill itself causes to fail.

## Proposals

### P1 — Give the ceremony an owner that survives the fallback [blocker]

**Problem.** The ceremony's commands exist only as `{PLACEHOLDER}` slots
in a file `SKILL.md:637-643` says to read "at dispatch". The main-loop
fallback (`SKILL.md:495-502`) has no dispatch, so it inherits the
promise of the ceremony without any of its text. Finding 2.

**Change.** Extract the ordered ceremony into a mode-independent
reference, `references/bead-ceremony.md`, holding the eight numbered
steps currently at `subagent-prompts.md:68-90` with `{RUN_DIR}`,
`{SCRIPTS}`, `{ID}` retained as slots. Have `subagent-prompts.md` cite
it rather than restate it, and replace the main-loop bullet at
`SKILL.md:499-503` with:

> - **Implement every bead in the orchestrator's own loop**, in queue
>   order. **Read `references/bead-ceremony.md` before the first bead
>   and follow its eight steps for each one, substituting `$RUN_DIR`
>   and `$SCRIPTS` for yourself** — there is no dispatch step in this
>   mode, so nothing else will put the ceremony in front of you. The
>   progress events and the retro file are written by you, at the same
>   points a subagent would write them.

**Evidence.** Findings 1, 2. Under the same model the main-loop path
then reaches the ceremony's commands through a step it actually
executes, instead of through a template it never opens.

### P2 — Verify the artifacts before recording a bead closed [blocker]

**Problem.** `SKILL.md:616-618` calls the retro file "non-negotiable"
but binds only the subagent. The orchestrator has no acceptance check,
so five beads were recorded closed with no anchor and the run's
aggregated retro lost its inputs silently. Finding 1.

**Change.** Insert a step in Phase 3, immediately after an outcome is
received and before the bead is written into the run table:

> **Acceptance check — run it for every bead, in every dispatch mode
> including main-loop.** A bead is not closed until its artifacts
> exist:
>
> ```bash
> test -s "$RUN_DIR/$ID.retro.md" &&
>   grep -q '"phase":"committed"' "$RUN_DIR/$ID.progress.jsonl"
> ```
>
> On failure the bead's row is `ceremony-incomplete`, not `closed`.
> Write the missing retro yourself from the subagent's returned summary
> (or your own, in main-loop mode), append the missing events with
> `bb-progress.sh`, and say so in the Phase 4 report. Do NOT reopen the
> bead or amend the commit — the work landed; only its record is
> missing.

**Evidence.** Finding 1. The next run cannot reach Phase 4 with five
beads' anchors missing and no one aware.

### P3 — State what the ledger does and does not prove [blocker]

**Problem.** `references/recovery.md:40-55` defines the event sequence
as the status surface without bounding its authority. A trail ending at
`implementing` was read as live-agent proof while the bead had been
committed three hours earlier, and an absent file was cited as present.
Finding 3.

**Change.** Add to `references/recovery.md`, directly under the event
schema:

> **The ledger is a hint, never the authority.** It is written by the
> agent doing the work, so it stops the instant that agent stops — and
> in main-loop mode it may never start. Therefore:
>
> - A trail ending before `committed` does **not** mean the bead is
>   unfinished. It means nothing was appended after that point.
> - An absent progress file does **not** mean no agent ran.
> - A file's mtime bounds the last *write*, not the agent's liveness.
>
> Resolve any status question in this order, stopping at the first
> answer: `bd show <id>` (the tracker), then `git log --oneline
> --grep=<id>` (the branch), then the ledger. Cite which one you used.
> Never report a ledger fact you have not just read from disk.

**Evidence.** Finding 3. Both errors — the stale trail read as liveness,
the absent file cited as present — are ruled out by a check that
consults `bd` and `git` first and forbids citing an unread file.

### P4 — Define the resume path for an interrupted bead [snag]

**Problem.** `SKILL.md:718-722` makes one-commit-per-bead absolute with
no interruption case, so an interrupted bead has no legal move. The run
produced a WIP commit the rule forbids. Finding 4.

**Change.** Append to that bullet:

> **WHEN a bead is interrupted mid-implementation** and its work must
> be persisted before the agent stops, commit it with a subject
> prefixed `WIP <bead-id>` and leave the bead `in_progress` — a WIP
> commit is the one sanctioned exception, and it is not a close. The
> agent that resumes the bead finishes the work, then squashes its own
> commit together with every `WIP <bead-id>` commit on the branch into
> the single landing commit before running the close ceremony. Record
> the interruption in the bead's retro.

**Evidence.** Finding 4. The next interrupted bead follows a stated
path, and the invariant the rest of the skill relies on holds at close
time.

### P5 — Split SKILL.md below the authoring guideline [snag]

**Problem.** At 859 body lines the skill trips this repo's own skill
linter, making `check-skills.sh --strict` red on a clean tree. This run
carried a hand-written baseline exception for it in every prompt; since
the line count is a static property of the committed file, any run in
this repo faces the same red gate until it is split. Finding 5.

**Change.** Move Phase 3's dispatch-mode detail and the Phase 4 report
specification into `references/dispatch-modes.md` and
`references/run-report.md`, leaving a one-paragraph summary and a
pointer for each — the pattern the skill already uses for
`recovery.md` and `subagent-prompts.md`, and complementary to P1's
extraction. Target under 500 body lines.

Separately, add to Phase 0's gate-profile step:

> WHEN a gate is already failing on an untouched tree, capture its
> output as the run's **baseline** before Phase 1 and record it in the
> dispatch plan's ledger. The bar for every bead is "no findings beyond
> the baseline", not "exit 0". A gate with no recorded baseline is
> expected to exit 0.

**Evidence.** Finding 5. The exception the run wrote by hand becomes a
step, and the linter stops reporting the skill that is meant to teach
authoring standards.

## Considered and Not Proposed

- **The dispatch plan was deviated from without a revision log.**
  `.bonanza/1786121603-ri/dispatch.md` plans wave 5 as a shared batch of
  `9kb.6`, `9kb.7`, `9kb.9` followed by `9kb.8` alone; the commits landed
  `.6` (21:34), `.7` (21:58), `.8` (22:17), `.9` (22:24) — a different
  order and, by their spacing, not as a batch. This is real, but it is
  exactly P4 of `rfc-self-run-dir-discoverability-and-orchestrator-
  recovery.md` ("Make `dispatch.md` amendable, with a revision log"),
  and it comes from the same run, so it is corroboration and not a
  second observation. Filing it again would make triage decide one
  proposal twice.
- **The orchestrator citing a progress file that never existed.** The
  fabricated citation was the agent's error, not the skill's. It is
  recorded in Finding 3 because P3's "never report a ledger fact you
  have not just read from disk" rule addresses it, but "the agent should
  have checked" is not a proposal — the fix for ignoring a check is not
  to add a warning about checking.
- **Making the Agent-tool availability check a hard stop.** Tempting,
  since main-loop mode is where the ceremony was lost, but the fallback
  is correct behavior and the run completed 12 beads through it. The
  defect is that the fallback lacks the ceremony's text, which P1 fixes;
  disabling the mode would trade a fixable gap for a dead run.
- **Requiring a timestamp field in progress events.** Would have made
  Finding 3's staleness self-evident, but `recovery.md:52-55` documents
  a deliberate decision against it ("the orchestrator's monitor
  notification is timestamped on arrival, so `ts` never earned its
  keep"), and P3's cross-check resolves the question without it.

## Non-Goals

- No change to the three scripts (`bb-progress.sh`, `bb-close-bead.sh`,
  `bb-verify-skip.sh`). All three behaved correctly wherever they were
  invoked; the defect is that two of them were not invoked at all.
- No change to the phase vocabulary or the event schema.
- No change to tier routing, gate profiles, or the close ceremony's
  ordering.

## Adoption

P1 and P2 land together — P2's acceptance check is only enforceable once
P1 guarantees a path that writes the artifacts in every mode, and P1
without P2 still fails silently if a future mode is added. P3 is
independent and touches only `references/recovery.md`; it is the
cheapest of the five and should land first regardless. P4 is independent
(`SKILL.md` only). P5's split is independent of all of them but should
follow P1, since P1 already moves the ceremony out of `SKILL.md`'s
orbit and changes the line budget; its Phase 0 baseline rule is
independent of the split and can land on its own.

Backward compatibility: P1 relocates text within the skill's own
reference set, so a project holding a pinned copy sees new file paths —
`SKILL.md`'s pointers move together with the files, and no project
document references `subagent-prompts.md` line numbers. P2's acceptance
check adds two shell commands per bead and no new dependency. P4 changes
what agents are permitted to do, not what they must undo, so
in-flight branches with WIP commits stay valid.

Expected effect: a run that degrades to main-loop still produces a
complete ledger and a retro anchor per bead; a run that loses them says
so in the Phase 4 report instead of reaching the aggregated retro with
five silent holes; and a recovering orchestrator answers "is this bead
done?" from the tracker rather than from a file that may never have
been written.

## Resolution
