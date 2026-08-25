---
title: "RFC: The worktree list is not the whole answer — a rejected push leaves
  a run indistinguishable from a finished one"
description: Every liveness signal the skill prescribes read "finished" while
  two commits sat unpushed behind a rejected push, because none of them compares
  local HEAD against the remote ref.
slug: rfc-self-worktree-absence-is-not-completion
type: rfc
status: ready-for-review
updated_at: 2026-08-14T09:45:00Z
origin: self
skill: recover-sprint
skill_version: "0.1"
model: claude-opus-5
effort: unrecorded
runs_observed: 1
severity: blocker
plugin_version: "n/a (in-repo dogfooding)"
---

# RFC: The worktree list is not the whole answer

## Summary

`recover-sprint` names `git worktree list` as the definitive liveness signal
and explicitly rejects the alternatives. In run `1786686364-15` that signal —
and every other one the skill offers — reported a finished run while two
commits sat unpushed behind a rejected push. The skill's own emphasis is what
makes this a defect rather than an omission: it tells a recovering agent that
the question is settled, so a correct recovery stops before checking the one
thing that would have shown the truth.

The recovery this RFC argues from did reach the right answer, but only by
running a check the skill does not list.

## Provenance

| Field | Value |
|---|---|
| Harness | Claude Code (Claude Code on the web, remote execution environment) |
| Task the skill was applied to | Recovering bead-bonanza run `1786686364-15` — 16 beads, 7 waves, two epics — in a session that no longer held the run's state, on branch `claude/validate-corpus-spec-cli-0t0cxo` |
| References / scripts / templates invoked | `recover-sprint` v0.1; `bead-bonanza` v1.5 (`scripts/bb-backup-run.sh`, `scripts/bb-close-bead.sh`); the run's `dispatch.md` and per-bead progress ledgers |
| Run dates | 2026-08-14 |
| Model lineup (from `dispatch.md` manifest) | light `haiku`, standard `sonnet`, advanced `opus`, frontier `fable`; `CLAUDE_CODE_SUBAGENT_MODEL` unset, so tier assignments were honoured |
| Testimony of record | `rfcs/sprints/2026-08-14-markdown-links-and-shipped-specs.md` — the run's sprint report. Findings below are cited by its `<bead-id>#<ordinal>` ids; where this RFC quotes one, the report is authoritative |

## Evidence

### E1 — every prescribed liveness signal reported a finished run that had stalled [misleading-text, blocker]

`recover-sprint/SKILL.md:112`:

> "3. **Establish what is in flight — `git worktree list` is the whole answer.**"

`SKILL.md:115`:

> "**A worktree that exists is a bead that is not finished.** The state is
> explicit, so nothing has to be inferred"

The skill then closes off the alternatives, correctly and for good reasons:

> `SKILL.md:128` — "**Do not infer liveness from file mtimes.** A dirty file's
> mtime bounds the last *write*, not the agent's liveness"
>
> "`bd list --status in_progress` is **not** a liveness signal either — a bead
> can be worked without ever being claimed"

At the moment of recovery, the run had stalled at wave 6b: `agent-plugins-47b`
committed `.github/workflows/checks.yml`, and the container's PAT lacks the
`workflow` scope, so the push was rejected outright — a rejection that fails
the whole push, not the offending file.

What the prescribed signals reported:

| Signal | Reading | Truth |
|---|---|---|
| `git worktree prune && git worktree list` | zero worktrees | the orchestrator had merged and removed 47b's worktree before pushing |
| working tree | clean | the work was committed, just not pushed |
| `bd show agent-plugins-47b` | `closed` | the close ceremony completed |
| `<id>.progress.jsonl` tail | `pushed` | the ledger records the subagent's own branch push, which succeeded |
| `git log --oneline origin/<branch>..HEAD` | **2 commits** | **the sprint branch never landed** |

Only the last row is true, and only it is absent from the skill. The
consequence: ~20 minutes in which a correct application of every documented
signal concludes "drained — nothing to do", and a recovery that stopped there
would have reported the run complete with two commits and a whole wave
unlanded.

The ledger row is the sharpest trap. Step 4 tells the agent to treat
`committed` as needing verification —

> `SKILL.md` step 4 — "**`committed`** — verify the recorded sha is an ancestor
> of `HEAD` (`git merge-base --is-ancestor <sha> HEAD`) before believing it."

— but offers no equivalent for `pushed`, the phase after it, even though
`pushed` is the one that can be true of a bead branch while the sprint branch
sits unlanded. The skill teaches distrust of the weaker claim and credulity
toward the stronger one.

## Proposals

### P1 — add the remote comparison as a required step, not an alternative signal [blocker]

The fix is not to weaken the worktree rule — it is correct about what it
measures. It is to say what it does *not* measure, and add the one check that
covers the gap.

**Change** — amend step 3's opening at `SKILL.md:112`:

> 3. **Establish what is in flight — `git worktree list` answers "is an agent
>    still working", and nothing else.**

**And** add a step 3b, before classification:

> 3b. **Establish what has LANDED — compare the branch against its remote.**
>
> ```sh
> git fetch "$REMOTE" "$BRANCH"
> git log --oneline "$REMOTE/$BRANCH..HEAD"    # must be empty
> ```
>
> An orchestrator that dies between committing and pushing, or whose push is
> **rejected**, leaves a state in which every other signal reads "finished":
> no worktrees (they are removed before the push), a clean tree, the beads
> closed in the tracker, and the progress ledger's last phase `pushed`. Only
> this comparison distinguishes a landed wave from an unlanded one.
>
> WHEN commits are unpushed, push them before concluding anything about the
> run, and record why they were not pushed — a rejection has a cause that will
> recur on the next wave. A credential refused for `.github/workflows/**`
> rejects the WHOLE push, so a single workflow file strands every commit
> behind it.

### P2 — extend step 4's per-phase distrust to `pushed` [blocker]

**Change** — in step 4's phase list, after the `committed` entry:

> - **`pushed`** — the subagent pushed its OWN bead branch, which is all the
>   per-bead prompt asks of it. It says nothing about whether the sprint branch
>   landed; in a concurrent wave the sprint-branch push is the orchestrator's
>   job and happens after this event is written. Verify with step 3b, never by
>   reading this phase.

### P3 — state the general rule the two proposals share [polish]

**Change** — add to step 5 ("Verify every claim against `bd` and `git`
yourself"):

> The signals in this skill are ordered by how much they can lie. A worktree's
> existence is state. A tracker status is a record someone wrote. A progress
> phase is testimony from an agent that has since stopped. And **the absence of
> a signal is not evidence of completion** — most of this skill's signals go
> quiet in exactly the same way whether a run finished or died, so a
> reconstruction built only from what is missing cannot tell those apart.

## Considered and Not Proposed

- **Having the orchestrator write a `landed` progress phase after the sprint
  push.** Attractive, and it would make the ledger self-sufficient. Not
  proposed here because it is a change to `bead-bonanza`'s ledger vocabulary
  rather than to `recover-sprint`, and because a recovering agent must not
  depend on a phase the dying orchestrator may never have reached — step 3b
  works regardless of what the ledger says.
- **Treating a non-empty `origin/<branch>..HEAD` as proof the run is alive.**
  Rejected: it proves the opposite is *not* established, nothing more. An
  orchestrator killed mid-push leaves the same state.
- **The snapshot/backup mechanism.** `bb-backup-run.sh` and the committed
  `.bonanza/<run-id>/` snapshot behaved correctly and the live-directory-first
  rule at step 1 was right: `$RUN_DIR` still existed, so the snapshot's
  staleness caveat never applied. No finding.
- **Step 2's "read `dispatch.md` in full, do not re-derive the queue".** Worked
  exactly as intended, including its instruction to reconcile `## Amendments`
  against the wave table — the amendments were where a mid-run discovered bead
  and a split wave were recorded. No finding.

## Non-Goals

- The cause of the rejected push. That is an environment fact and a
  `bead-bonanza` pre-flight gap; it is filed against that skill.
- Any change to how worktrees are created, merged or removed.
- Changing the run-identification steps or the snapshot format.

## Adoption

P1 and P2 are independent text additions and can land in either order; P3 is a
summarizing rule that reads best after both. All are additive — no existing
step's behavior changes, and a recovery of a healthy run does one extra fetch.

Expected effect: the 20-minute silent stall this run suffered becomes a check
that costs one fetch and reports the truth immediately.

## Resolution

**Partial triage 2026-08-17. This RFC is NOT retired** — P3 is recommended
for adoption but not yet approved.

| Proposal | Outcome |
|---|---|
| P1 — add the remote comparison as a required step | **adopted** as written. Step 3's opening now reads "`git worktree list` answers 'is an agent still working', and nothing else", and a step 3b runs `git fetch` + `git log --oneline "$REMOTE/$BRANCH..HEAD"` before classification, with the unpushed-commit and workflow-scope guidance intact. |
| P2 — extend step 4's per-phase distrust to `pushed` | **adopted** as written, placed directly after the `committed` entry. |
| P3 — state the general rule the two proposals share | **recommended adopt; not yet approved.** |

**Refinement applied to P1, and why.** This proposal and
`rfcs/inbox/bead-bonanza/rfc-self-push-scope-and-discovered-work-channel` P2
are **the same check written into two skills** — the same incident seen from
the recovering side and the running side. Implemented independently they
would drift into two rules, and a later triage would read the divergence as
two mechanisms. Each now names the other as the same check in the same form:
step 3b points at `bead-bonanza` Phase 3 step 7 and vice versa. The stronger
fix — stating it once in `references/shared/` and vendoring it — was not taken
in this pass because it widens the change beyond the approved set; it remains
the better long-term shape and should be revisited if a third caller appears.

**Provenance check.** `recover-sprint` v0.1 cited; v0.1 current at triage.
`SKILL.md:112` and `:115` were re-read and both were present verbatim. The
`pushed` phase was confirmed **absent** from step 4's phase list, exactly as
E1 reports.

**Dedupe.** No match in retired RFCs 0001–0023 — `recover-sprint` had no prior
RFC. Overlaps `rfc-self-push-scope-and-discovered-work-channel` P2 at one
point, handled above.

**Self-authorship, disclosed.** `origin: self`, `model: claude-opus-5`; this
session is configured for `claude-opus-5`. **The RFC was authored and triaged
by the same model.** Read these adoptions as this repository approving its own
proposal, not as independent corroboration. Both adopted proposals ratify a
check their own author improvised during the recovery being described — the
RFC says so itself ("The recovery this RFC argues from did reach the right
answer, but only by running a check the skill does not list").

**Files changed.** `plugin/skills/recover-sprint/SKILL.md`.

**Version bumps.** `recover-sprint` 0.1 → 0.2. Plugin 0.18.0 → 0.19.0.
