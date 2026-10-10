---
title: "RFC: The run directory is the recovery surface, and its location
  survives only in the orchestrator's context"
description: Persist the run directory and sprint plan where the recovery
  query already reaches, add an orchestrator-death recovery path, and
  back the in-session Monitor with a durable scheduled trigger
slug: rfc-0017
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

# RFC: Only context knows where the recovery surface lives

## Summary

A `bead-bonanza` run closed 8 beads across 7 subagents in ~1h40m, was
killed mid-batch by a user interrupt, and then sat dead for 2h16m. The
skill had already written everything needed to recover — a dispatch
plan, a decisions doc, seven retro anchors, and per-bead progress files
recording the exact phase each bead reached. The recovering orchestrator
never found any of it, because the run directory's path is generated at
runtime and echoed to the transcript, and the transcript is the one
store that did not survive. Recovery was instead reconstructed by hand
from git mtimes, `bd`, and session transcripts over roughly thirty
minutes. **This RFC is not "build a sprint plan" — the skill already
builds a good one. It is "make it findable, and add the path for when
the orchestrator is the thing that died."**

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins (self) |
| Skill critiqued | `bead-bonanza` v1.3 |
| Task | Implement the reference-integrity feature (one epic, 10 work orders + 4 prerequisite/adjacent beads) |
| Run directory | `/tmp/claude/bead-bonanza/1786121603-ri/` |
| Subagents dispatched | 7 (tiers: 1 haiku, 2 sonnet, 4 opus) |
| Beads closed by the run | `koi`, `tix`, `j7f`, `ml9`, `9kb.1`–`9kb.4` |
| References invoked | `references/recovery.md`; `references/subagent-prompts.md`; `scripts/bb-close-bead.sh`; AGENTS.md `## bead-bonanza gates` and `## model lineup` |
| Sibling skills in the same session | `groom-feature-blueprint` v1.2, `groom-beads` v1.1, `retro-spec`, `retro-skill` |
| Harness | Claude Code on the web (`remote_desktop`), remote execution container, multi-root workspace (`agent-plugins` + `strata`), CLI 2.1.224 |
| Model | Withheld: this session's harness supplies the model identifier under a policy forbidding it in artifacts pushed to a repository. Same handling as RFC-0010; tracked by `rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md` |
| Runs observed | 1 |
| Run date | 2026-08-07 (run 16:52Z–18:33Z; recovery 20:48Z–22:00Z) |
| Plugin version | 0.15.0 |

## Background — how a long-running session actually works

Every proposal below depends on three mechanics that are easy to get
wrong from the inside, so they are stated here once.

**There is no long-running agent. There are turns.** An agent is
instantiated when input arrives, thinks and acts, emits a final message,
and ceases. Between turns it holds nothing — no thread, no working
memory. Its entire state at the start of a turn is (a) the conversation
context the harness reconstructs and (b) whatever is on disk. A sprint
that spans hours is not one agent working; it is a chain of turns, each
one waking with reconstructed memory. **Waking with amnesia is the
normal condition, not the failure case** — it simply goes unnoticed
while the reconstruction happens to be complete.

**Only some things can start a turn.** Distinguish two capabilities that
are routinely conflated:

| Mechanism | Starts a turn | Carries state into it |
|---|---|---|
| Task notification (a child agent finishes or dies) | yes | no |
| Scheduled trigger (`send_later`, cron) | yes | yes — arbitrary text |
| `Monitor` on a file | yes | yes — the matched event |
| SessionStart / UserPromptSubmit hook | **no** | yes |
| `CLAUDE.md` / `AGENTS.md` | **no** | yes, but static |

A hook cannot wake an agent; it can only enrich a turn that is already
happening. A notification can wake an agent but tells it nothing about
what it was doing. **A durable sprint needs one of each**, and this run
had neither in durable form.

**Context restore is not guaranteed to return what you last had.** This
session's transcript file holds 434 records from the run under a second
`sessionId` created by a `/clear`, and the turn that resumed 2h16m later
was reconstructed from the *pre-clear* half. The orchestrator woke
holding a conversation that predated its own mandate. The mechanism
behind that selection is harness-internal and not observable from inside
the agent; what is observable, and what matters here, is that **context
is not a store you may rely on for anything you need after a gap.**

## Evidence

1. **[gap, blocker] The run directory's location is persisted nowhere.**
   `plugin/skills/bead-bonanza/SKILL.md:103-104` creates it:

   > ```
   > RUN_DIR="${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>"
   > mkdir -p "$RUN_DIR" && touch "$RUN_DIR/.ok" && echo "run dir: $RUN_DIR"
   > ```

   The `echo` is the entire persistence mechanism — a line of stdout in
   the orchestrator's context. The run id (`1786121603-ri`) is generated
   and unguessable. Measured after the fact:

   ```
   grep -c "1786121603-ri" .beads/issues.jsonl            -> 0
   git log --all -S "1786121603-ri" | wc -l               -> 0
   grep -rl "1786121603-ri" <repo> --exclude-dir=.git |wc -> 0
   ```

   Meanwhile the directory held: `dispatch.md` (the full wave plan),
   `decisions.md`, **seven** `<bead>.retro.md` anchors (4.3–17.5 KB
   each), and **eight** `<bead>.progress.jsonl` files. The recovering
   orchestrator found none of it and reconstructed sprint state by hand
   from `git log`, dirty-file mtimes, `bd show`, and raw session
   transcripts.

2. **[gap, blocker] The exact fact reconstructed by hand was sitting in a
   file.** `<bead-id>.progress.jsonl` contained, in full:

   > ```
   > {"bead":"<bead-id>","phase":"claimed"}
   > {"bead":"<bead-id>","phase":"implementing"}
   > ```

   against a completed bead's six-line trace ending
   `{"phase":"committed","sha":"7607b69"}`. "Claimed, implementing, never
   reached gates" is precisely the conclusion the orchestrator derived
   forensically from file modification times.

3. **[gap, blocker] `recovery.md` covers subagent death, not orchestrator
   death.** Its scope line (`references/recovery.md:3`) is "how it
   recovers when a [subagent dies]", and its procedure section is headed
   `## Subagent crash / mid-step timeout recovery` (line 19). Every step
   presumes a live orchestrator that knows the run. There is no section
   for the orchestrator itself dying, losing context, or re-entering a
   run in progress — which is what happened.

4. **[gap, blocker] The only wake mechanism is in-session and died with
   the turn.** `SKILL.md:612-615` and `references/recovery.md:62`
   instruct: "**Monitor any subagent past 5 minutes** … the orchestrator
   must arm a `Monitor` on its progress JSONL". The run complied —
   Monitor `btnrafelq` was armed and was relaying phase events correctly
   (`gates-passed` at 18:27:46, `committed` at 18:29:20). At 18:33:08.850
   a user interrupt recorded `[Request interrupted by user]` in the
   subagent. At 18:33:08.862 and .863 two notifications were **enqueued**:

   > `status: killed — Agent "Batch: RES + PREG (9kb.4, 9kb.5)" was stopped by user`
   > `status: killed — Monitor "batch RES+PREG phase events" stopped`

   Neither was ever dequeued — every prior notification in that
   transcript shows an `enqueue` followed ~50ms later by a `dequeue`;
   these two have none. The same interrupt stopped the orchestrator, so
   the mechanism that would have reported the death died with it. Result:
   2h16m of silence while the UI showed the task running.

5. **[gap, snag] No mechanism for amending the plan mid-sprint.**
   `dispatch.md` is written once at Phase 2.5 and never revised. During
   recovery the plan changed materially — wave 5 was planned as one
   shared batch (`9kb.6`, `9kb.7`, `9kb.9`), and was re-decided as
   strictly sequential because every remaining bead runs `bd` against a
   shared Dolt DB and commits to one working tree. There was nowhere in
   the skill's artifacts to record that, so it was appended to the epic's
   tracker notes instead — an improvised location the skill does not
   know about and the next run will not look in.

6. **[gap, snag] The run directory dies with the container.**
   `SKILL.md:103` roots it at `${TMPDIR:-/tmp}`. The project's own
   `CLAUDE.md` states the container "is reclaimed after a period of
   inactivity (or when the session ends)". The seven retro files are
   4.3–17.5 KB of real per-bead analysis and are the documented input to
   the end-of-sprint aggregated retro; they exist only in `/tmp`, and
   the run had already been idle 2h16m — well into reclaim territory —
   with them unbacked.

## Proposals

### P1 — Persist the run directory where recovery already looks [blocker]

**Shape.** Gap — no rule at Phase 0 for making `$RUN_DIR` discoverable
after the orchestrator's context is gone.

**Problem.** `SKILL.md:104`'s `echo "run dir: $RUN_DIR"` writes the only
copy of the path into the least durable store available. Findings 1 and
2: every artifact needed for mechanical recovery existed, and the
recovering agent could not reach them.

**Change.** At Phase 0 step 2, after creating the directory, persist the
pointer to **two** places that survive context loss, and make the
epic/tracker note the canonical one:

> ```bash
> RUN_DIR="${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>"
> mkdir -p "$RUN_DIR" && touch "$RUN_DIR/.ok"
> ln -sfn "$RUN_DIR" "${TMPDIR:-/tmp}/claude/bead-bonanza/current"
> ```
>
> Then record the pointer in the tracker, on the epic when the run has
> one and on every bead in scope when it does not:
>
> ```bash
> bd update <epic-id> --append-notes "bead-bonanza run <run-id>
> started <ts>. RUN_DIR=$RUN_DIR (also $TMPDIR/claude/bead-bonanza/current).
> Holds dispatch.md, decisions.md, <bead>.progress.jsonl, <bead>.retro.md."
> ```
>
> The tracker copy is canonical because it is committed to
> `.beads/issues.jsonl` and therefore survives container reclaim; the
> `current` symlink is the cheap local shortcut.

**Evidence.** Findings 1, 2.

**Backward compatibility.** Purely additive — two extra commands at
Phase 0. No existing run artifact changes shape, and a run that skips it
behaves exactly as today.

### P2 — Add an orchestrator-recovery entry procedure [blocker]

**Shape.** Gap — `references/recovery.md` has no section for the
orchestrator dying or resuming.

**Problem.** Finding 3: the file's scope is subagent death and every step
assumes a live orchestrator holding the run. Finding 4: the orchestrator
is exactly what died. The recovering agent had no procedure and invented
one, and the one it invented was wrong in a way it later had to correct
(it keyed liveness on `bd list --status in_progress`, which returns
empty while an unclaimed agent is actively working).

**Change.** Add to `references/recovery.md` a section
`## Orchestrator crash / context loss — resuming a run`, stating the
entry procedure in order:

> 1. **Find the run.** `bd show <epic-id>` for the RUN_DIR pointer (P1),
>    or `ls -l ${TMPDIR:-/tmp}/claude/bead-bonanza/current`. If neither
>    exists, the run predates P1: reconstruct from `git log` and
>    `bd list --status all`, and say so in the report.
> 2. **Read `dispatch.md`.** It carries the mandate, the standing
>    authorization, the wave table and the gate profile. This is the
>    plan; do not re-derive it.
> 3. **Establish liveness — git is primary, the tracker is
>    corroboration.** A bead may be worked without being claimed, so
>    `bd list --status in_progress` can be empty during live work:
>    ```bash
>    git -C <repo> status --porcelain | awk '{print $2}' \
>      | while read f; do stat -c '%Y %n' "<repo>/$f"; done | sort -rn | head -1
>    ```
>    newest dirty mtime **< ~5 min** → an agent is alive; do not
>    dispatch, do not commit its files. **> ~30 min** → orphaned work
>    from a killed child; recover it. **No dirty files** → idle.
> 4. **Classify each in-scope bead** by its `<bead>.progress.jsonl` tail:
>    absent → never started; `claimed`/`implementing` → died mid-work,
>    re-dispatch or finish; `gates-passed`/`retro-written` → finish the
>    close from `<bead>.retro.md`; `committed` → verify the sha and move
>    on.
> 5. **Verify every claim against `bd` and `git` yourself.** A completion
>    notification is not proof; one was lost entirely in this run.

**Evidence.** Findings 3, 4, and the corrected liveness rule the
recovering orchestrator had to discover mid-run.

**Backward compatibility.** New section in a reference file; changes no
existing step.

### P3 — Back the Monitor with a durable trigger [blocker]

**Shape.** Gap — the skill mandates an in-session wake mechanism and
names no durable one.

**Problem.** Finding 4. `Monitor` is the correct primary signal — it is
immediate and carries the event — but it is in-session and dies with the
turn that armed it. When both parent and child are killed together there
is no path back into the session, and the two kill notifications proved
it: enqueued, never dequeued, silence until a human asked.

**Change.** Amend `SKILL.md:612-615` to require both, and to state why:

> **Monitor any subagent past 5 minutes, and back it with a durable
> trigger.** The `Monitor` is the fast path; it is in-session and dies
> with the turn that armed it, so it cannot report its own death. Before
> the first dispatch of a run, also arm a recurring out-of-session
> check-in (`send_later`, or a cron trigger, whichever the harness
> offers) at 15–30 minute intervals, whose message states the RUN_DIR
> path and points at the orchestrator-recovery procedure. Cancel it in
> Phase 4 when the run drains. A run whose only wake mechanism is
> in-session cannot survive an interrupt that lands on the orchestrator.

**Evidence.** Finding 4. Verified in this session: a `send_later`
trigger armed at 21:13Z fired correctly at 21:31Z and drove a full
recovery check, and its stored binding is to the **cloud session id**
(`persistent_session_id`), not to a local transcript — which is why it
survives clears and interrupts.

**Backward compatibility.** Additive. Costs one trigger per run plus a
cancel in Phase 4.

### P4 — Make `dispatch.md` amendable, with a revision log [snag]

**Shape.** Gap — no rule for changing the plan after Phase 2.5.

**Problem.** Finding 5. Plans change for good reasons mid-run: a
constraint is discovered, a batch is found unsafe, a bead is bounced.
The skill writes `dispatch.md` once and never revisits it, so an
amendment has nowhere to live and the plan on disk silently stops
describing the run. The recovering orchestrator in this session recorded
its re-decision in tracker notes because the skill offered no home for
it — meaning `dispatch.md` now contradicts what actually happened.

**Change.** Add to Phase 2.5:

> `dispatch.md` is **append-only and amendable**. WHEN the plan changes
> mid-run — a wave resequenced, a bucket reclassified, a bead added or
> bounced, a tier escalated — append a dated entry under a
> `## Amendments` heading stating what changed and why, and update the
> wave table in place. The plan on disk SHALL describe the run as it is
> actually being executed; a stale plan is worse than none, because
> recovery trusts it (P2 step 2). Mirror one line of the amendment to
> the tracker note from P1 so a reader who has only the tracker knows
> the plan moved.

**Evidence.** Finding 5.

**Backward compatibility.** Additive; a run with no amendments produces
the same `dispatch.md` as today.

### P5 — Back up the run directory before it is reclaimed [snag]

**Shape.** Gap — no rule for run artifacts outliving the container.

**Problem.** Finding 6. The retro files are the documented input to the
end-of-sprint aggregated retro and represent the run's only per-bead
analysis, at 4.3–17.5 KB each. They live only under `${TMPDIR:-/tmp}`,
which the project's `CLAUDE.md` says is reclaimed after inactivity. This
run idled 2h16m with them unbacked.

**Change.** Add to the per-bead close ceremony, after `retro-written`:

> The retro is already concatenated into the bead's description by the
> close ceremony, so the tracker copy is durable for any bead that
> closed. Additionally, **as each wave completes**, copy the whole run
> directory into the repository under `.bonanza/<run-id>/`, **commit
> it**, and note the location on the epic.
>
> It must be **committed, not gitignored**. A gitignored copy lives in
> the working tree, which is reclaimed with the container exactly like
> `/tmp` — it would move the artifacts without making them durable. The
> whole point is surviving container reclaim, and only a pushed commit
> does that. The cost is small: this run's directory is 152 KB across 17
> files, dominated by seven retros of 4.3–17.5 KB.
>
> Copy the whole directory rather than a filtered subset. The progress
> files are a few hundred bytes each and are what make a post-mortem
> mechanical; deciding which retros "matter" is a judgment best not made
> while the evidence is being discarded.
>
> **Then delete it as the final step of the run.** `.bonanza/<run-id>/`
> is scratch state for one run: it earns its keep while the run is live
> and recoverable, and it must not reach the PR that carries the run's
> actual work. The last action of Phase 4, after the final report and
> before handing off, is `git rm -r .bonanza/<run-id>/` committed on its
> own. A run that ends with the directory still tracked has not
> finished.
>
> Do **not** gitignore `.bonanza/`. Ignoring it and committing it are
> mutually exclusive, and ignoring it silently removes the durability
> this proposal exists to provide.

**Where the artifacts live after cleanup.** Deleting the directory does
not lose the run: the commits that added it stay in the branch's
history, so a post-mortem reaches them with `git log --diff-filter=A --
'.bonanza/<run-id>/'` and `git show`. Durable during the run, reachable
afterwards, absent from the merged tree.

**Evidence.** Finding 6, and PR #16 review: "The `.bonanza` directory
should be cleaned up as a last step in the bonanza. We don't want to
merge these with our PRs."

**Backward compatibility.** Additive and cheap; a few file copies at
Phase 4.

## Status detection — the concrete methods, ranked

Consolidated from P2, and stated here because the run showed the ranking
is not obvious:

| Question | Method | Reliability |
|---|---|---|
| Is an agent alive right now? | newest dirty-file mtime via `git status --porcelain` + `stat` | **highest** — observed correct; caught a live agent at 71s and an orphan at 2h16m |
| Where did a bead die? | tail of `<bead>.progress.jsonl` | high — closed vocabulary, one line per phase |
| What was the plan / mandate? | `dispatch.md` | high, once discoverable (P1) |
| What is committed? | `git log --oneline`, `git rev-list --count @{u}..HEAD` | high |
| Which beads are done? | `bd list --label <slug> --status all` | high for closed; see below for in-flight |
| Is a bead claimed? | `bd show <id>` status | **low for liveness** — an agent can work without claiming; observed `open` while 12 files were being edited |

The last row is the trap. Bead status is a good record of *completion*
and a poor signal of *activity*.

## Trigger survey and recommendation

| Trigger | Wakes a turn | Survives interrupt / clear / restart | Verified this run | Recommend |
|---|---|---|---|---|
| Task notification (child completes or dies) | yes | **no** — lost when the parent turn dies | yes: two enqueued 18:33:08, never dequeued | **Adopt as primary.** Fastest, zero cost. Never as sole. |
| `Monitor` on progress JSONL | yes | **no** — in-session, dies with the turn | yes: `btnrafelq` killed with its parent | **Keep** for live per-phase visibility, per `recovery.md:62`. |
| `send_later` (one-shot scheduled) | yes | **yes** — bound to `persistent_session_id`, server-side | yes: armed 21:13Z, fired 21:31Z, drove a full recovery | **Adopt as backstop** (P3). |
| Cron / recurring trigger | yes | yes | not exercised | **Prefer over `send_later`** for a run of unknown length — it re-arms itself instead of depending on the agent to remember. Cancel at Phase 4. |
| SessionStart hook | **no** | n/a — enriches a turn | yes: `~/.claude/session-start-git-identity.sh` runs in this harness | **Adopt** for restoring the run pointer on every session start. Note it must be user-level: project-level `.claude/settings.json` hooks do not register in a multi-root workspace, per this project's `CLAUDE.md`. |
| UserPromptSubmit hook | **no** | n/a | not exercised | **Consider** — injects a compact computed status header per turn. Keep it ~10 lines; the full plan is read on demand. |
| Stop hook | no (fires at turn end) | n/a | yes: the repo's git-guard fired 3× | **Consider** as an invariant guard: "a run is active and this turn neither advanced it nor armed a wake." Would have caught this stall at 18:33 rather than 20:48. |

The layering that follows: **notification (fast) + cron (durable) +
SessionStart hook (state on wake)**. Any one can fail; in this run the
first did, and nothing stood behind it.

## Considered and Not Proposed

- **The orchestrator paraphrased the dispatch prompt from memory instead
  of reading `references/subagent-prompts.md`.** During recovery it
  hand-wrote briefs for `9kb.5`, `9kb.6` and `9kb.7`. Consequences:
  `9kb.6` was worked with the bead still at `open` because the brief
  omitted `bd update --claim`, and no `progress.jsonl` or `retro.md` was
  written for any of the three, so those beads have no recovery anchor
  at all. `SKILL.md:626-627` warns against exactly this: "Do not
  paraphrase a template from memory: the per-bead ceremony is ordered,
  and the ordering is what makes the crash-recovery path mechanical."
  The skill was right and was ignored. Recorded because the consequence
  is real and worth knowing; not proposed, because the fix for ignoring
  a warning is not another warning.
- **The recovering orchestrator's own heartbeat keyed liveness on
  `bd list --status in_progress`** and would have concluded "nothing in
  flight" while an agent edited twelve files. Caught by the heartbeat's
  own first firing and corrected in-run. That is a defect in an
  improvised artifact, not in `bead-bonanza`; its lesson is folded into
  P2 step 3.
- **The git-guard Stop hook cannot distinguish orphaned work from
  in-flight work.** It fired three times this session on three
  situations needing opposite responses. Real and worth fixing, but it
  is `~/.claude/stop-hook-git-check.sh`, not this skill. Noted for the
  hook's owner; the discriminator is the `stat` one-liner in P2 step 3.
- **Batching two beads into one subagent doubles the blast radius of a
  kill.** `9kb.4`+`9kb.5` were batched; the kill landed between them,
  leaving one committed and one orphaned. Not proposed as a rule change:
  the batch behaved correctly — it committed `9kb.4` on completion
  rather than deferring both — and with P1 and P2 in place the orphan
  would have been a two-command recovery. Batching is a throughput
  decision the dispatch plan already reasons about explicitly.
- **`decisions.md` and `dispatch.md` quality.** Both were good: the plan
  correctly identified that worktree-parallel was unavailable and why,
  carried a four-item environmental ledger, and flagged its own
  over-routing suspect. No finding.

## Non-Goals

- No change to the phase structure, the gate profile, the tier rubric,
  or the per-bead close ceremony. All performed correctly.
- No change to what `dispatch.md` contains. Finding 5 is about amending
  it, not about its fields — the existing fields (wave, bead, class,
  bucket, tier, model, push policy, reason, plus the gate profile and
  ledger) are what a recovering orchestrator needs, and this run proved
  it by needing exactly them.
- No proposal to stop using `Monitor`. P3 adds a backstop; it does not
  replace the fast path.

## Adoption

P1 and P2 are the pair that matters and **should land together**: P1
makes the artifacts findable, P2 says what to do once found. Either
alone is half a recovery path. P3 is independent and cheap. P4 and P5
are independent of everything.

All five are edits to `SKILL.md` and `references/recovery.md` plus two
shell lines in Phase 0; none requires a script change or alters any
existing artifact's shape. `bead-bonanza` would go 1.3 → 1.4.

Backward compatibility is stated per proposal; every proposal is
additive, and a run that ignores all five behaves exactly as today.

Expected effect: an orchestrator that wakes with no memory of a run can
find it in one `bd show`, read the plan, classify every bead by its
progress tail, and resume — instead of reconstructing the run from git
archaeology, or not noticing it exists at all.

## Resolution

**Triaged 2026-08-10; implemented across two `bead-bonanza` sprints and
resolved 2026-08-12.** 5 proposals: 5 adopted, 3 of them with refinement.
None deferred, none rejected. 1 triage-originated change is recorded below
the proposal table.

Approved alongside two sibling RFCs in `rfcs/inbox/bead-bonanza/`; with
more than five proposals approved and most of them changing one file,
step 5's "Sequence before you edit" branch applied, so each proposal was
filed as a bead under one epic and run through
`groom-beads` rather than implemented inline. The second sprint replaced
two mechanisms these proposals depended on — the tracker's transport and
the liveness signal — and the affected rows say so.

| Item | Outcome |
|------|---------|
| P1 — persist the run directory where recovery already looks | **adopted with refinement.** The problem is solved; the canonical store is not the one P1 named. P1 made a tracker note on the epic canonical *because* `.beads/issues.jsonl` was committed and therefore survived container reclaim. A later bead untracked the JSONL exports, voiding that rationale, so no tracker note is written. What replaced it, in order of authority: the run is derived from the branch — one run, one branch, one run directory, and `git log <default-branch>..HEAD` lists the beads that landed; `.bonanza/<run-id>/SNAPSHOT.md`, committed on that branch, names the run and its as-of time and survives a reclaim; `${TMPDIR:-/tmp}/claude/bead-bonanza/current`, a symlink `pre-flight-check.sh` repoints at the start of every run, is the cheap local shortcut, kept as P1 proposed. `dispatch.md` is written unconditionally and carries a run manifest. |
| P2 — add an orchestrator-recovery entry procedure | **adopted with refinement.** `references/recovery.md` gained the section, and steps 1, 2, 4 and 5 landed substantially as written. Step 3 did not survive: its mtime liveness table (`< ~5 min` alive, `> ~30 min` orphaned, via `git status --porcelain` and `stat`) landed as proposed and was then deleted when a later bead made every bead run in its own worktree. Worktree presence is now the in-flight signal and `git worktree list` is the whole answer, so the file states the opposite of P2's step 3: "Do not infer liveness from file mtimes." The heuristic P2 was right to prefer over `bd list --status in_progress` was replaced by a fact rather than a better heuristic. |
| P3 — back the Monitor with a durable trigger | **adopted** as written. `references/recovery.md`, "Back the Monitor with a durable trigger": the `Monitor` stays the fast path, an out-of-session check-in is armed before the first dispatch at 15–30 minute intervals carrying the run pointer and the recovery entry point, and it is cancelled in Phase 4. |
| P4 — make `dispatch.md` amendable, with a revision log | **adopted** as written, and slightly widened: `dispatch.md` is now written unconditionally rather than only for large runs, enumerates the beads in scope, and carries a run manifest alongside the `## Amendments` log. The recovery procedure's step 2 reconciles the amendments against the wave table, which is what makes an amendment load-bearing rather than decorative. |
| P5 — back up the run directory before it is reclaimed | **adopted with refinement, and merged with the sibling RFC's P2.** RFC 0018 (`rfc-0018-silent-ceremony-loss-and-the-stale-twin`) P2 refined this proposal rather than duplicating it, and the two were implemented as one bead: `scripts/bb-backup-run.sh` copies the run directory to `.bonanza/<run-id>/` and writes a self-describing `SNAPSHOT.md`. Two refinements to P5 as written: the snapshot is refreshed after **every bead's acceptance check**, not once per wave — a copy taken once goes stale within the half hour, which is the sibling's finding — and the live run directory is authoritative wherever both exist. Committed-not-gitignored and deleted as the final act of Phase 4 landed as proposed. |

**Triage-originated changes.** From the review session rather than from
this RFC's testimony, numbered `T` so they are not read as its proposals.

| Item | Outcome |
|------|---------|
| T1 — add `pre-flight-check.sh`: mechanize Phase 0 on macOS and Linux, and measure what it currently asserts | **adopted.** Phase 0 spent ~4,600 bytes of the compaction-surviving window describing shell commands; the script now performs them, including creating the run directory, repointing the `current` symlink P1 asks for, and capturing the gate baseline. |

**On the P5/P2 merge.** Recorded in both Resolutions so neither reads as
unaddressed. This RFC's P5 established *that* a copy must exist, be
committed, and be deleted at the end of Phase 4. The sibling RFC's P2
established what the copy must *declare* — an as-of time, the source
path, and which beads' artifacts it holds — and that a snapshot is never
authoritative while the live directory exists. One bead carried both;
neither proposal was decided twice, and neither was dropped.

**Self-authorship, disclosed.** This RFC carries `origin: self`: it was
authored inside this repository by an agent running the skill it
critiques, and triaged inside the same repository. Its `model` field is
withheld under the harness policy tracked by
`rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md`, so
whether the authoring agent and the triaging agent were the same model
**cannot be confirmed or excluded**. Nothing here is independent
corroboration. Two proposals ratify decisions their own author made
during the run being critiqued: P3, whose durable-trigger requirement
generalizes the `send_later` the authoring session armed and verified on
itself, and P4, whose amendment log sanctions the mid-run re-plan that
session had already improvised into tracker notes. Both are recorded as
adopted on the strength of a single self-observed run.

**Provenance check.** `skill: bead-bonanza` at `skill_version: "1.3"` and
`plugin_version: "0.15.0"` — both current when the RFC was filed and when
triage began. Every cited line was re-read against the tree before
classification; none had moved or been fixed since.

**Dedupe.** No match in RFCs 0001–0015. The two sibling RFCs triaged with
it overlap at exactly one point, the P5/P2 merge above.

**Files changed.** `plugin/skills/bead-bonanza/SKILL.md`;
`plugin/skills/bead-bonanza/references/recovery.md`, `dispatch-plan.md`,
`subagent-prompts.md`, `project-setup.md`;
`plugin/skills/bead-bonanza/scripts/pre-flight-check.sh` (new),
`bb-backup-run.sh` (new); `scripts/check-skills.sh`.

**Version bumps.** No versioned template or standard changed for this
RFC. Plugin 0.16.0→0.17.0, one MINOR bump covering this RFC and its two
siblings.

The `version` frontmatter of the changed shipped **skills** has been
bumped: `bead-bonanza` 1.3→1.4 and `spec` 1.8→1.9.
The other four (`groom-beads` 1.1→1.2, `groom-feature-blueprint`
1.2→1.3, `retro-skill` 0.3→0.4, `retro-spec` 1.3.0→1.4.0) were bumped
for the SKILL.md edits this RFC's own closeout made, which had gone
unbumped since 2026-07-30.
