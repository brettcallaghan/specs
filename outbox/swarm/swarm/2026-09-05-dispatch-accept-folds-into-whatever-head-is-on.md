---
title: "Check out the target before the fast-forward, or refuse when HEAD is not on it"
description: swarm dispatch accept's serialized arm fast-forwards whatever branch HEAD happens to be on, then reports the fold as merged with a merged_sha read from the untouched target.
slug: issue-dispatch-accept-folds-into-whatever-head-is-on
audience: skill
target: swarm
priority: 0
updated_at: 2026-09-05T00:00:00Z

# Evidence — what this issue rests on
findings: []                         # none: the run's coordinator recorded this
                                     # as an `amend:` comment on the run epic
                                     # rather than as a `retro:` finding, which
                                     # is itself filed as
                                     # issue-coordinator-tooling-findings-are-invisible-to-the-harvest
recurrence: 0
report: docs/retros/2026-09-05-59ic-spec-corpus-anchor-references.md

# Provenance — who and what produced this
run: agent-plugins-59ic
session: https://claude.ai/code/session_0185zfzFm9maHZjPX4FQqhbZ
branch: sprint/agent-plugins-59ic
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<derived, not observed — manifest.models holds only the tier→model map (standard→sonnet, advanced→opus) and nothing records what served each dispatch; the coordinator that observed the symptom is unrecorded>"
---

## What is wrong, and what it costs

`swarm dispatch accept`'s **serialized** arm never puts the repository on the
target branch before folding. It runs `git merge --ff-only <dispatch-branch>`,
which advances **whatever branch `HEAD` is on** — so a repository sitting on
some other branch has that branch fast-forwarded instead, and the run's target
stays at base. The command then reads what landed from the target *by name*,
which is a branch nothing wrote to, and reports the fold as a success with a
`merged_sha` naming the pre-merge target head. The concurrent arm does not have
this defect: it checks the target out before the fast-forward. Only the
serialized arm is affected, which is the arm every non-concurrent run takes.

The cost is a **silent write to the wrong branch, self-reported as a
success**. A caller reading `merged: true` has no way to know the fold did not
reach the target; the sprint branch it believes is advancing is not, and every
subsequent accept compounds against a branch carrying none of the run's work.
On the run that hit it, only the ancestry guard's independent check
(`ancestor: false`, exit 2) stopped it — the merge step itself reported no
problem. Recovery was manual: check the repository out onto the sprint branch,
reset the wrongly-advanced branch back to `base_sha`, re-run accept.

## Evidence

Read at the source against `brettcallaghan/swarm` at `1a19e8e`. Line numbers
are hints; the quoted strings are the durable anchors.

- **The target is checked out only on the concurrent arm.** `foldIn` in
  `internal/cli/accept.go` (line 603) builds its steps as:

  ```go
  steps := plan.Steps{}
  if !m.serialized {
      steps = append(steps, vc.Rebase("rebase", dispatchBranch, m.targetSHA)...)
      steps = append(steps, vc.Checkout("return-to-target", target)...)
  }
  ffSteps := vc.MergeFastForward("fast-forward", dispatchBranch)
  ```

  `grep -c 'vc.Checkout\|\.Checkout(' internal/cli/accept.go` returns **1** —
  the single occurrence is line 610, inside that `if !m.serialized` block. The
  serialized arm reaches `MergeFastForward` with the repository wherever the
  caller left it.

- **`MergeFastForward` advances the current branch, and says so.** Its own doc
  comment in `internal/vcs/git/git.go` (line 296) opens: "MergeFastForward
  advances **the current branch** to branch, refusing anything that is not a
  fast forward." The step it emits is
  `"merge", "--ff-only", branch` — no `-C`, no target, no checkout.

- **The reported sha is read from the target by name, not from what moved.**
  `landedSteps := vc.Resolve("merged-head", target)` at line 620. When the fold
  landed elsewhere, this resolves the target's *unchanged* head — which is
  exactly the stale `merged_sha` observed.

- **The precondition is documented and not enforced.** `--explain` prints, at
  line 164, `git rev-parse --verify HEAD^{commit}   (the repository is on the
  target)`. `readCase` does resolve `HEAD` — `headStep := vc.Resolve(
  "repo-head", "HEAD")` at line 483 — but uses it to *decide the merge case*,
  not to assert the invariant.

- **The observed symptom, recorded by the run's coordinator on the run epic
  `agent-plugins-59ic` and NOT measured by the author of this issue.** With the
  main checkout on `claude/swarm-agent-plugins-rc14-lclbkc`, accept ran
  `git merge --ff-only bb/dev-59ic-d` against **that** branch, left
  `sprint/agent-plugins-59ic` at base `225751e`, and still reported
  `merged: true` with `merged_sha` set to the pre-merge target head; the
  ancestry guard returned `ancestor: false`, exit 2. The source reading above
  is this issue's own measurement and accounts for that report exactly; the
  transcript itself was not reproduced.

## Proposed change

The serialized arm stops folding into a branch it has not established is the
target. Either shape closes it, and the command owns both ends:

- **Check the target out first**, exactly as the concurrent arm already does —
  move `vc.Checkout("return-to-target", target)` out of the `if !m.serialized`
  block so it runs on every path before `MergeFastForward`; or
- **Refuse when `HEAD` is not on the target.** `readCase` already resolves
  `HEAD` for its own purposes at line 483; compare it against the target and
  exit non-zero with the `--explain` precondition as the message, rather than
  printing that precondition and proceeding regardless.

Done is: with the repository checked out on some other branch, `swarm dispatch
accept` either lands the fold on the manifest's branch or exits non-zero. It
never returns `merged: true` for a fold that did not reach the target, and it
never returns a `merged_sha` read from a branch no step wrote to.
