---
title: "RFC: The ceremony can be skipped with nothing detecting it, and
  the run directory's committed copy is a stale twin of the live one"
description: Three beads closed with no retro anchor because their briefs
  were paraphrased, and no orchestrator step checks; separately the
  committed backup shares the live run directory's id while missing two
  beads' trails, which misled a retro into reporting anchors as absent
slug: rfc-0018
type: rfc
status: retired
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
A reader who finds the copy has no way to know it is partial. Finding 2 —
two false conclusions in one retro, both from this cause.

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
> **The live run directory is authoritative; a snapshot never is.** Any
> procedure that reads the backup — recovery, a retro, a DoD audit —
> first checks whether the live directory still exists, and uses it when
> it does. Read the backup only when `$RUN_DIR` is gone, and then state
> the snapshot's as-of time in whatever you conclude from it.
>
> The snapshot is committed under `.bonanza/<run-id>/` while the run is
> live and deleted as the final step of Phase 4 (sibling RFC's P5), so
> "refresh after every bead" applies for the run's duration and the
> directory does not reach the merged tree.

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

**Triaged 2026-08-10; implemented across two `bead-bonanza` sprints and
resolved 2026-08-12.** 6 proposals: 6 adopted — 3 with refinement, 1 of
those still incomplete, and 1 since retired. None deferred, none
rejected. 2 triage-originated changes are recorded below the proposal
table.

Approved alongside two sibling RFCs in `rfcs/inbox/bead-bonanza/`; with
more than five proposals approved and most of them changing one file,
step 5's "Sequence before you edit" branch applied, so each proposal was
filed as a bead under epic `agent-plugins-92q` and run through
`groom-beads` rather than implemented inline.

| Item | Outcome | Bead |
|------|---------|------|
| P1 — verify the artifacts before recording a bead closed | **adopted with refinement**, and the refinement inverts the check. P1's test reads the run directory (`test -s "$RUN_DIR/$ID.retro.md"`, grep the progress ledger), but both files are written by the agent being checked, so the test is testimony verifying itself — exactly the failure it was written to catch. What landed is a five-signal acceptance check against git and tracker state: the bead reads closed in `bd`, a close commit exists on the bead's branch, that branch merged into the sprint branch `--ff-only`, **the worktree is gone** (the load-bearing signal — removal is the orchestrator's own act of accepting the bead, so a surviving worktree is a bead not accepted whatever the subagent reported), and at least one `retro:` comment exists or the subagent explicitly reported a clean bead. `SKILL.md` also names the two signals the check deliberately does not use, the `"phase":"committed"` ledger event and a per-bead retro file — which no longer exists, findings having moved to `retro:` comments. P1's disposition rules survived: a failing bead gets a Phase 4 row naming the failing signal instead of a commit sha, and is not counted among the closed. | `agent-plugins-92q.12`, `agent-plugins-92q.22` |
| P2 — make the run-directory backup self-describing | **adopted with refinement, and merged with the sibling RFC's P5.** `scripts/bb-backup-run.sh` writes `SNAPSHOT.md` beside the copy recording the run id, the as-of timestamp, the live directory's path, the run branch, the HEAD when it was taken, and every bead id with a progress ledger or handoff file. The "live directory is authoritative, a snapshot never is" rule landed verbatim in `references/recovery.md` and binds recovery, retros and DoD audits alike. Refinement: the refresh is keyed to P1's acceptance check as proposed, but the bead-id list is of handoff files rather than `.retro.md` files, because the per-bead retro file was retired in the same sprint. | `agent-plugins-92q.13`, `agent-plugins-xvl`, `agent-plugins-vme` |
| P3 — fit the operative procedure under the re-injection cap | **adopted with refinement — decomposed, and NOT yet complete.** See *On P3* below. The re-read instruction, the linter's byte threshold, the Phase 0 gate baseline and the reordering all landed. The size target did not: `SKILL.md`'s body is **61,153 bytes (~15,288 tokens)** against P3's ~20,000-byte target, and `scripts/check-skills.sh --strict` still exits 1 on it. Do not read this row as resolved. | `agent-plugins-92q.1`, `agent-plugins-92q.2`, `agent-plugins-92q.7`, `agent-plugins-92q.3` (Phase 0 baseline half); measurement gate `agent-plugins-8ef` deferred |
| P4 — state what the ledger does and does not prove | **adopted** as written. `references/recovery.md` carries "the ledger is a hint, never the authority" directly under the event schema, with all three negative statements — a trail ending before `committed` proves nothing, an absent progress file proves nothing, an mtime bounds the last write and not liveness — and the resolution order `bd show`, then `git log --grep`, then the ledger. It survived the second sprint untouched and is now cited by the Phase 3 acceptance check as the reason a `"phase":"committed"` event is not an acceptance signal. | `agent-plugins-92q.11` |
| P5 — define the resume path for an interrupted bead | **adopted, landed, and since retired.** The `WIP <bead-id>` exception and its squash-before-close rule landed as proposed. `agent-plugins-92q.30` then removed the condition that made it necessary: a bead's interrupted work sits in that bead's own worktree on its own branch, so there is nothing to persist with a WIP commit and nothing to squash. The resume path is now to re-enter the worktree and finish the work, and `references/recovery.md` says so explicitly. The proposal was correct for a shared working tree; it does not survive one worktree per bead. | `agent-plugins-92q.15`; retired by `agent-plugins-92q.30` / `agent-plugins-92q.33` |
| P6 — let the close script write the project's trailers | **adopted** as written, both halves. `bb-close-bead.sh` takes `--trailers-file <path>`, appended to the commit message body after a blank line, resolved once per run at Phase 0 and threaded into every dispatch template; and the ceremony now reads the landed sha from the branch (`git rev-parse --short HEAD`) rather than from the script's printout, so a rewrite between the commit and the ledger write cannot desync them. | `agent-plugins-92q.16` |

**Triage-originated changes.** From the review session rather than from
this RFC's testimony, numbered `T` so they are not read as its proposals.

| Item | Outcome | Bead |
|------|---------|------|
| T1 — Definition of Ready: require decisions to be closed, and narrow the sanctioned hedge in the Design check | **adopted.** The spec-owned `work-order-definition-of-ready` gained check 9; `groom-beads` gained the matching checklist line and parks a bead whose Design leaves a choice open as `blocked-needs-info` rather than guessing. | `agent-plugins-92q.4` |
| T2 — delete `SKILL.md` Phase 2: the readiness gate owns decisions, and refinement does not run inside the sprint | **adopted.** Phase 2 was 4,280 bytes inside the window that survives compaction, spent on work the Definition of Ready (T1) already requires to be finished before dispatch. Recorded here as triage-originated, though it also served P3's byte budget. | `agent-plugins-92q.5` |

**On P3.** Adopted with refinement because it was decomposed rather than
implemented as the single restructure it describes, and it is the one row
in this Resolution that is not finished.

*Why it was decomposed.* P3 proposes one restructure of a file whose
contents were about to change. Within the same epic,
`agent-plugins-92q.30` and `agent-plugins-92q.31` deleted whole passages
of `SKILL.md` outright — the shared-tree machinery, the tracker-write
deferral, the JSONL-as-transport rules — so a restructure performed
against the file as it stood would have been redone. The proposal was
split into the parts that were safe to land immediately, the reordering
that had to wait for the content-reducing beads, and a measurement gate
to verify the result.

*What landed.* `agent-plugins-92q.1` put the post-compaction re-read
instruction at the top of the file, above the truncation point — the half
P3 itself calls the cheapest and says can ship on its own.
`agent-plugins-92q.2` rewrote `scripts/check-skills.sh` to threshold skill
bodies on bytes derived from the 5,000-token re-injection cap instead of
on a 500-line guideline; the byte threshold immediately found a second
skill over the cap that the line count had never flagged
(`plugin/skills/spec/SKILL.md`, 22,497 bytes). `agent-plugins-92q.3`
carried P3's separate Phase 0 half: `pre-flight-check.sh --gate-baseline`
runs the gate profile on the untouched tree, records the result in
`dispatch.md`'s run manifest, and makes "no findings beyond the baseline"
the stated bar — retiring the hand-written exception this run repeated
into every prompt. `agent-plugins-92q.7` reordered the file so the
executable core comes first: worktrees, the merge policy, the phase loop,
the acceptance check, the close ceremony and the script interfaces are all
inside the surviving window, and the explanatory material moved to
`references/operating-principles.md` and `references/dispatch-plan.md`.

*What did not.* `agent-plugins-8ef` was the bead that would have measured
the fit, trimmed the residual and retired the KNOWN RED entry. It was
closed on 2026-08-12 by the repository owner's decision, to be revisited
in a dedicated analysis round rather than carried as a standing open bead.
So the finding that motivates P3 — Finding 7, that everything past the cap
leaves context mid-run — is **mitigated, not resolved**. Measured on the
tree this Resolution was written against: body 61,153 bytes against a
~20,000-byte cap, meaning roughly two thirds of the file is still dropped
on re-injection. The reordering means what is dropped is now explanatory
rather than operative, which is the whole of the improvement. The linter
baseline is `0 error(s), 2 warning(s)`, `RESULT: FAIL`, on `bead-bonanza`
and `spec`.

**On the P2/P5 merge.** Recorded in both Resolutions so neither reads as
unaddressed. This RFC's P2 explicitly refines the sibling RFC's P5 rather
than duplicating it: P5 established that a committed copy must exist, P2
established what it must declare and that it is never authoritative while
the live directory exists. One bead carried both.

**Self-authorship, disclosed.** This RFC carries `origin: self`: it was
authored inside this repository by an agent running the skill it
critiques, and triaged inside the same repository. Its `model` field is
withheld under the harness policy tracked by
`rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md`, so
whether the authoring agent and the triaging agent were the same model
**cannot be confirmed or excluded**. Nothing here is independent
corroboration — and this RFC is not even an independent run: it says so
itself, covering the same run `1786121603-ri` as its sibling, with
`runs_observed: 1`. Two of its proposals ratify decisions their own author
made. P5 sanctions the `WIP` commit that run improvised against the
skill's stated invariant, converting a rule-break into a rule; the sprint
then retired it on other grounds. P3's decomposition was decided by the
triage rather than by the RFC, and the measurement that would have tested
whether the restructure worked was deferred by the human — which is the
one place in these three Resolutions where a self-adopted proposal is
recorded as unfinished rather than as landed.

**Provenance check.** `skill: bead-bonanza` at `skill_version: "1.3"` and
`plugin_version: "0.15.0"` — both current when the RFC was filed and when
triage began. Every cited line was re-read against the tree before
classification; none had moved or been fixed since. Finding 5's measured
red gate was reproduced (`0 error(s), 1 warning(s)` at the time; the
threshold has since changed and the baseline is now 2 warnings).

**Dedupe.** No match in RFCs 0001–0015. The overlap with its two sibling
RFCs is confined to the P2/P5 merge above.

**Files changed.** `plugin/skills/bead-bonanza/SKILL.md`;
`plugin/skills/bead-bonanza/references/recovery.md`,
`operating-principles.md`, `dispatch-plan.md`, `subagent-prompts.md`;
`plugin/skills/bead-bonanza/scripts/bb-close-bead.sh`,
`bb-backup-run.sh` (new), `pre-flight-check.sh` (new);
`scripts/check-skills.sh`; `work-order-definition-of-ready.md` and
`work-order-template.md` in `plugin/skills/spec/references/` and their
vendored copies under `groom-beads`, `groom-feature-blueprint` and
`bead-bonanza`; `plugin/skills/groom-beads/SKILL.md`.

**Version bumps.** `work-order-definition-of-ready` 1.2→1.3,
`work-order-template` 1.3→1.4. Plugin 0.16.0→0.17.0, one MINOR bump
covering this RFC and its two siblings.

The `version` frontmatter of the changed shipped **skills** has been
bumped in `agent-plugins-x8r`: `bead-bonanza` 1.3→1.4 and `spec` 1.8→1.9.
The other four (`groom-beads` 1.1→1.2, `groom-feature-blueprint`
1.2→1.3, `retro-skill` 0.3→0.4, `retro-spec` 1.3.0→1.4.0) were bumped
for the SKILL.md edits this RFC's own closeout made, which had gone
unbumped since 2026-07-30.
