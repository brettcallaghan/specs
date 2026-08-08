---
title: "RFC: The ceremony can be skipped with nothing detecting it, and
  the run directory's committed copy is a stale twin of the live one"
description: Three beads closed with no retro anchor because their briefs
  were paraphrased, and no orchestrator step checks; separately the
  committed backup shares the live run directory's id while missing two
  beads' trails, which misled a retro into reporting anchors as absent
slug: rfc-self-silent-ceremony-loss-and-the-stale-twin
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

# RFC: Silent ceremony loss, and a backup that lies by omission

## Summary

`bead-bonanza` run `1786121603-ri` closed 12 beads. Nine carry a
complete progress trail and a retro anchor; three — `9kb.5`, `9kb.6`,
`9kb.7` — carry none, and the run reached its final bead with nobody
aware. The skill does warn against the proximate cause
(`SKILL.md:625-627`, "Do not paraphrase a template from memory"), but
that warning **was not in the orchestrator's context when the decision
was made**: a skill body is re-injected after compaction capped at 5,000
tokens, which truncates this 869-line SKILL.md at line 406 and drops the
warning, the retro-anchor rule, and the script interfaces together
(Finding 7). So this RFC does not propose a louder warning — a warning
that is not present cannot be made louder. It proposes the check that
catches the loss, and the restructuring that keeps the procedure in
context. A second, independent defect compounded it: the run
directory's committed backup carries the same run id as the live
directory but was snapshotted mid-run and never refreshed, so a retro
reading the backup reported two beads' anchors as missing when both are
complete. A stale twin under an identical name is worse than no backup,
because it answers confidently and wrongly.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code on the web (remote container session) |
| Task the skill was applied to | The `agent-plugins-9kb` reference-integrity epic: 12 beads across two repos, on branch `claude/reference-integrity-blueprint-groom-xc7d30` |
| References / scripts / templates invoked | `bead-bonanza` SKILL.md v1.3; `references/subagent-prompts.md`; `references/recovery.md`; `scripts/bb-progress.sh`; `scripts/bb-close-bead.sh`; `skill-rfc-template` v0.3 |
| Run dates | 2026-08-07 (single continuous run, 17:21–22:24) |

**On the withheld model field.** This session's operating instructions
forbid writing the model identifier into any artifact committed to a
repository. The placeholder is deliberate and matches the two prior
`bead-bonanza` RFCs here. The value is in the session transcript; the
human copying this RFC can fill it in, and triage needs it to scope the
fix.

**Relationship to the two prior RFCs in this directory.** This RFC
covers the **same run** as
`rfc-self-run-dir-discoverability-and-orchestrator-recovery.md`, which
was written at 21:43 about the run's first half. This is therefore
**not an independent recurrence**, and `runs_observed` stays 1. P2 below
refines that RFC's P5 rather than duplicating it; the distinction is
stated there.

## Evidence

1. **[gap · blocker] Three beads closed with no retro anchor, and no
   step in the skill checks.** The live run directory
   `/tmp/claude/bead-bonanza/1786121603-ri/` holds a `.progress.jsonl`
   and a `.retro.md` for `tix`, `ml9`, `j7f`, `9kb.1`–`9kb.4`, `9kb.8`
   and `9kb.9`. For `9kb.5` it holds a progress file of two events and
   **no** retro; for `9kb.6` and `9kb.7` it holds **nothing**. All three
   are closed in `bd` (`bd show agent-plugins-9kb` reports 9/10
   complete) with landing commits `bb7ecc3`, `a43bcf9`, `da9d8bd`. The
   epic's own progress note records the cause: "the previous three
   hand-written briefs cost 9kb.5/.6/.7 their progress files and retro
   anchors". `SKILL.md:616-618` calls the retro file "the recovery
   anchor, and it is non-negotiable", and `SKILL.md:625-627` says "Do
   not paraphrase a template from memory" — but both bind the dispatch
   step, and **no step anywhere verifies the artifacts exist** before a
   bead is recorded closed or before Phase 4 aggregates them. Three
   beads' worth of anchors went missing across roughly three hours with
   no signal.

2. **[gap · blocker] The committed backup is a stale twin of the live
   run directory, under the same run id.** `.bonanza/1786121603-ri/`
   was committed at 21:56 (`02c3cc8`, 17 paths). The live directory now
   holds five files the backup does not:
   `agent-plugins-9kb.8.{progress.jsonl,retro.md,show.txt}` and
   `agent-plugins-9kb.9.{progress.jsonl,retro.md}` — beads that landed
   at 22:17 and 22:24, after the snapshot. Nothing in the backup says
   it is partial or when it was taken. A retro run at 22:30 read the
   backup, found no `9kb.8`/`9kb.9` artifacts, and reported both beads
   as having closed with no ceremony — then, compounding it, asserted
   that `agent-plugins-9kb.9.progress.jsonl` "has never existed in this
   run directory". It exists, with a complete seven-event trail ending
   `{"phase":"committed","sha":"7761db5"}` and `{"phase":"pushed"}`.
   Both errors are attributable to a backup that shares its name with
   the live directory and declares no as-of time.

3. **[gap · snag] A ledger that stops short of `committed` says nothing
   about the bead, and no rule states this.**
   `agent-plugins-9kb.5.progress.jsonl` ends at
   `{"phase":"implementing"}` with mtime 18:29, while bead `9kb.5` was
   committed at 21:06 (`bb7ecc3`) and four further beads landed after
   it. `references/recovery.md:40-55` presents the event sequence as the
   status surface without bounding its authority, and a heartbeat in
   this run read a trail ending at `implementing` as "positive proof of
   a live agent at a known phase".

4. **[gap · snag] No rule for a bead that cannot land in one commit.**
   `SKILL.md:718-722`: "**One commit per bead, always.** Every closed
   bead = exactly one commit on the branch." Bead `9kb.5` landed in two
   — `e3f2777` "WIP agent-plugins-9kb.5 (PREG): profile registration,
   snapshots pending" at 20:52, then `bb7ecc3` at 21:06. The rule is
   absolute with no interruption case, so the agent improvised a commit
   the invariant forbids rather than following a stated resume path.

5. **[gap · snag] The skill's own size makes a project gate red, and the
   skill says nothing about that case.** `scripts/check-skills.sh:118`
   warns above 500 body lines; `bead-bonanza`'s SKILL.md body is 859, so
   `./scripts/check-skills.sh --strict` prints `0 error(s), 1
   warning(s)` and `RESULT: FAIL` on a clean tree, with `bead-bonanza`
   the sole finding. This run carried a hand-written exception for it
   (`.bonanza/1786121603-ri/dispatch.md` ledger entry 2: "KNOWN RED,
   pre-existing … Bar is no NEW findings, not exit 0") repeated into
   every prompt. Since the line count is a static property of the
   committed file, any run in this repo meets the same red gate until
   the file is split. Phase 0's gate-profile step has no rule for
   baselining an already-failing gate.

6. **[script-defect · snag] `bb-close-bead.sh` writes no commit trailers,
   and the amend that adds them invalidates the sha the ledger just
   recorded.** The script commits with the bare subject `{ID} P{PRIORITY}
   {TITLE}` and no body, so a repository requiring `Co-Authored-By` /
   `Claude-Session` trailers must `git commit --amend` after step 6 —
   which rewrites the commit. Step 7 then records the pre-amend sha.
   Observed on `9kb.10`:
   `agent-plugins-9kb.10.progress.jsonl` ends
   `{"phase":"committed","sha":"14f0be3"}`, while the landing commit on
   the branch is `875593a`; `git merge-base --is-ancestor 14f0be3 HEAD`
   exits non-zero — the recorded sha is a dangling pre-amend object.
   Repro: run `bb-close-bead.sh`, `git commit --amend` to add any
   trailer, then compare the ledger's `sha` to `git rev-parse HEAD`.

7. **[gap · blocker] Compaction truncates the skill body at line 406, so
   the entire operative procedure leaves the orchestrator's context
   mid-run.** Claude Code re-injects an invoked skill's body after
   compaction "capped at 5,000 tokens per skill … oldest dropped first",
   and truncation keeps the START of the file
   (`code.claude.com/docs/en/context-window`). `SKILL.md` is 45,983
   bytes / 869 lines; measured, `head -c 20000 SKILL.md | wc -l` lands at
   **line 406**, `### Choose a dispatch mode per bucket`. Everything
   after that is dropped on re-injection, including:

   - `SKILL.md:624-627` — "**Read that file at dispatch** and fill in
     every `{PLACEHOLDER}` before sending. Do not paraphrase a template
     from memory";
   - `SKILL.md:616-618` — the retro anchor "is non-negotiable";
   - `SKILL.md:813-843` — the Reference-files table and the three
     published script interfaces, which the skill states are published
     precisely because the scripts are not to be read.

   `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE=80` is set in this cloud
   environment (verified in-session), so compaction fires at 80% whether
   or not the orchestrator plans for it. The three beads that lost their
   ceremony — `9kb.5`, `9kb.6`, `9kb.7` — are exactly the three
   dispatched after the run's context loss. The proximate cause recorded
   in the epic ("hand-written briefs") is correct; this is why the
   briefs were hand-written.

## Proposals

### P1 — Verify the artifacts before recording a bead closed [blocker]

**Problem.** `SKILL.md:616-618` and `SKILL.md:625-627` place the whole
obligation on the dispatch step and the subagent. When that step is done
wrong the loss is undetectable: three beads closed without anchors and
the run continued normally. Finding 1. Adding emphasis to the existing
warning is explicitly not the fix — the skill already carries it.

**Change.** Insert a step in Phase 3, immediately after an outcome is
received and before the bead is written into the run table:

> **Acceptance check — every bead, every dispatch mode.** A bead is not
> closed until its artifacts exist:
>
> ```bash
> test -s "$RUN_DIR/$ID.retro.md" &&
>   grep -q '"phase":"committed"' "$RUN_DIR/$ID.progress.jsonl"
> ```
>
> On failure the bead's row is `ceremony-incomplete`, not `closed`.
> Reconstruct the missing retro from the subagent's returned summary and
> the bead's diff, append the missing events with `bb-progress.sh`, and
> record the reconstruction in the Phase 4 report. Do NOT reopen the
> bead or amend the commit — the work landed; only its record is
> missing.

And in Phase 4, before the aggregated retro: list every bead whose
anchor was missing or reconstructed. An aggregated retro built from an
incomplete anchor set states which beads contributed nothing.

**Evidence.** Finding 1. Under the same model, a paraphrased brief still
costs that bead its trail, but it is caught at the next outcome instead
of at the end of the run — or, as happened here, not at all.

### P2 — Make the run-directory backup self-describing [blocker]

**Problem.** The backup and the live directory share a run id and differ
in content, and the backup carries no as-of time or completeness marker.
A reader who finds the committed copy has no way to know it is partial.
Finding 2 — two false conclusions in one retro, both from this cause.

**Change.** This refines P5 of
`rfc-self-run-dir-discoverability-and-orchestrator-recovery.md` ("Back
up the run directory before it is reclaimed"), which established *that*
the copy should exist. It does not say what the copy must declare, and
this run shows a once-taken copy going stale within 30 minutes. Add:

> The backup is written by `bb-backup-run.sh <run-dir> <dest>`, which
> copies the directory and writes `SNAPSHOT.md` beside it recording the
> as-of timestamp, the source path, and the bead ids whose
> `.retro.md` are present. Re-run it after **every** bead's acceptance
> check (P1), not once — a copy taken mid-run and never refreshed is
> indistinguishable from a complete one.
>
> **The live run directory is authoritative; the committed copy never
> is.** Any procedure that reads the backup — recovery, a retro, a DoD
> audit — first checks whether the live directory still exists, and
> uses it when it does. Read the backup only when
> `$RUN_DIR` is gone, and then state the snapshot's as-of time in
> whatever you conclude from it.

**Evidence.** Finding 2. A retro that consults the live directory first
does not report a complete seven-event trail as a file that never
existed.

### P3 — Fit the operative procedure under the re-injection cap [blocker]

**Problem.** Two problems with one fix, and the severe one is Finding 7,
not Finding 5. The skill body is re-injected after compaction capped at
5,000 tokens, truncating at line 406 and dropping the dispatch rule, the
retro-anchor rule and the script interfaces — the material an
orchestrator needs *most* after losing context, gone precisely when it
loses context. Separately (Finding 5) the same length trips this repo's
skill linter.

**Change.** Restructure so everything an orchestrator executes sits
before the cap, and everything explanatory moves out:

- **Above the cap:** Phase 3's execution loop, the close ceremony's
  ordering, the acceptance check from P1, the circuit breaker, and the
  Reference-files table with the three script interfaces.
- **Into `references/`:** Phase 1.5's grooming rationale, Phase 2's
  decision-doc format, the Phase 4 report specification, and the long
  Operating-principles prose — the pattern already used for
  `recovery.md` and `subagent-prompts.md`.
- Target under 5,000 tokens of body (~20,000 bytes), which also clears
  the linter's 500-line guidance.

Add near the TOP of the file, where truncation cannot reach it:

> **After any compaction, before dispatching anything:** re-read this
> file from Phase 3 onward and re-read `references/recovery.md` — a
> compacted session re-injects only this file's first ~5,000 tokens, so
> the ceremony, the script interfaces and the anti-paraphrase rule are
> not in your context even though you have "read the skill". Then
> `ls $RUN_DIR` and re-run the queue query to re-derive state from the
> durable source rather than from the summary.

And give Phase 2.5 the compaction trigger: compact deliberately at a
bucket boundary — never mid-bucket, never with a subagent in flight —
and check context after each bucket rather than letting the automatic
pass choose the moment. `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE=80` is set in
cloud environments, so it will choose one.

Separately, add to Phase 0's gate-profile step:

> WHEN a gate already fails on an untouched tree, capture its output as
> the run's **baseline** before Phase 1 and record it in the dispatch
> plan's ledger. The bar for every bead is "no findings beyond the
> baseline", not "exit 0". A gate with no recorded baseline is expected
> to exit 0.

**Evidence.** Finding 5. The exception this run wrote by hand becomes a
step, and the linter stops reporting the skill meant to teach authoring
standards.

### P4 — State what the ledger does and does not prove [snag]

**Problem.** `references/recovery.md:40-55` defines the event sequence
as the status surface without bounding its authority, so a trail ending
at `implementing` was read as proof of a live agent while the bead had
been committed three hours earlier. Finding 3.

**Change.** Add directly under the event schema in
`references/recovery.md`:

> **The ledger is a hint, never the authority.** It is written by the
> agent doing the work, so it stops the instant that agent stops.
> Therefore:
>
> - A trail ending before `committed` does **not** mean the bead is
>   unfinished — only that nothing was appended after that point.
> - An absent progress file does **not** mean no agent ran.
> - A file's mtime bounds the last *write*, not the agent's liveness.
>
> Resolve any status question in this order, stopping at the first
> answer: `bd show <id>`, then `git log --oneline --grep=<id>`, then the
> ledger. Name which one you used. Never report a ledger fact you have
> not just read from disk, and never from the backup while the live
> directory exists (P2).

**Evidence.** Finding 3, and the reading error in Finding 2 that the
last sentence forecloses.

### P5 — Define the resume path for an interrupted bead [snag]

**Problem.** `SKILL.md:718-722` makes one-commit-per-bead absolute with
no interruption case, so an interrupted bead has no legal move. Finding
4.

**Change.** Append to that bullet:

> **WHEN a bead is interrupted mid-implementation** and its work must be
> persisted before the agent stops, commit it with a subject prefixed
> `WIP <bead-id>` and leave the bead `in_progress` — a WIP commit is the
> one sanctioned exception, and it is not a close. The agent that
> resumes the bead finishes the work, then squashes its own commit with
> every `WIP <bead-id>` commit on the branch into the single landing
> commit before running the close ceremony. Record the interruption in
> the bead's retro.

**Evidence.** Finding 4. The next interrupted bead follows a stated
path, and the invariant holds at close time.

### P6 — Let the close script write the project's trailers [snag]

**Problem.** `bb-close-bead.sh` emits a bare one-line subject. Any
project whose convention requires commit trailers must amend after the
close, and the amend rewrites the commit the ledger recorded one step
later — so the ledger points at an object that is not on the branch.
Finding 6. Every bead in this run that pushed trailers has this drift.

**Change.** Give `bb-close-bead.sh` an optional `--trailers-file <path>`
whose contents are appended to the commit message body (blank line
first), and have Phase 0 resolve it once per run alongside the gate
profile, from the project's stated commit convention. Then reorder the
subagent template so no amend is needed: step 6 closes and commits with
trailers already in place, and step 7's recorded sha is final. WHERE a
project has no trailer convention, the flag is omitted and behavior is
unchanged.

Independently of the flag, change step 7 of
`references/subagent-prompts.md` to read the sha from the branch rather
than from the script's output:

> 7. `bb-progress.sh "$RUN_DIR" {ID} committed sha=$(git rev-parse
>    --short HEAD)` — read it from the branch, not from the script's
>    printout, so any rewrite between the two steps cannot desync it.

**Evidence.** Finding 6. The ledger's `committed` event names a commit
that is actually reachable, which is what P1's acceptance check and
every recovery path assume.

## Considered and Not Proposed

- **"The main-loop fallback drops the ceremony."** Proposed in the first
  draft of this RFC and **withdrawn**: the premise is refuted by this
  same run. `9kb.8` and `9kb.9` produced complete seven-event trails and
  14KB/9KB retro anchors at 22:17 and 22:24, so the ceremony demonstrably
  reaches beads dispatched from the real template in this session. The
  first draft reached the opposite conclusion by reading only the
  committed backup, which is Finding 2 — the withdrawal is itself
  evidence for P2.
- **A stronger warning against paraphrasing briefs.** The proximate
  cause of Finding 1 is hand-written briefs, and `SKILL.md:625-627`
  already forbids exactly that. An earlier draft of this RFC dismissed
  it as carelessness on that basis — that dismissal was **wrong**, and
  Finding 7 is why: the warning sits at line 625 of a file truncated at
  line 406 on re-injection, so it was absent from context at the moment
  it was needed. The conclusion survives for a different reason. Per the
  anti-pattern table, "the fix for ignoring a warning is not another
  warning"; here the fix for a *truncated* warning is not a louder one
  either. It motivates P1's detection and P3's restructuring, and
  nothing else.
- **The dispatch plan was deviated from without a revision log.**
  `.bonanza/1786121603-ri/dispatch.md` plans wave 5 as a shared batch of
  `9kb.6`, `9kb.7`, `9kb.9` then `9kb.8` alone; the commits landed `.6`
  (21:34), `.7` (21:58), `.8` (22:17), `.9` (22:24). Real, but it is P5
  of the run-dir RFC and comes from the same run — corroboration, not a
  second observation. Filing it again would make triage decide one
  proposal twice.
- **Requiring a timestamp in progress events.** Would make Finding 3's
  staleness self-evident, but `recovery.md:52-55` records a deliberate
  decision against it, and P4's cross-check resolves the question
  without it.

## Non-Goals

- No change to `bb-progress.sh` or `bb-verify-skip.sh`; P6 changes only
  `bb-close-bead.sh`, by adding an optional flag. All three behaved
  correctly wherever they were invoked; Finding 1's defect is that two
  were not invoked for three beads at all.
- No change to the phase vocabulary or the event schema.
- No change to tier routing, gate profiles, or the close ceremony's
  ordering.

## Adoption

**Land P3 first.** It is the causal root: while the operative procedure
sits past the re-injection cap, every other rule in this file is one
compaction away from being absent, including P1's acceptance check. The
post-compact re-read paragraph is the cheapest half and can ship on its
own, ahead of the restructuring.

P1 and P2 are independent of each other but reinforce: P1 detects a lost
anchor, P2 stops a stale copy from hiding one that exists. P2's "live
directory is authoritative" rule is two paragraphs and is what this run
got wrong twice, so it should not wait on the rest. P4 depends on
nothing and touches only `references/recovery.md`. P5 and P6 are
independent (`SKILL.md` and `bb-close-bead.sh` respectively).

Backward compatibility: P1 adds two shell commands per bead and no new
dependency. P2 adds a script and a `SNAPSHOT.md` artifact; existing
backups without one are, by the new rule, to be treated as
unknown-as-of and subordinate to a live directory. P3 moves text between
files that only this skill reads, and no project document cites
`SKILL.md` line numbers. P5 permits a commit shape that was previously
forbidden, so no in-flight branch becomes invalid.

Expected effect: a bead that loses its anchor is caught at the next
outcome instead of at the end of the run; an aggregated retro states
which beads contributed nothing rather than under-reporting silently;
and a reader of a preserved run directory knows what it is and when it
was taken.

## Resolution
