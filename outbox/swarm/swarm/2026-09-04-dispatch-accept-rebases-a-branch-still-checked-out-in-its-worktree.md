---
title: "Rebase inside the dispatch worktree, or remove it first, when accepting a dispatch"
description: swarm dispatch accept rebases the dispatch branch while that branch is still checked out in the worktree the command itself created, so the second of two concurrent dispatches can never be accepted.
slug: issue-dispatch-accept-rebases-a-branch-still-checked-out-in-its-worktree
audience: skill
target: swarm
priority: 0
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#12"]
recurrence: 1                        # run agent-plugins-59ic hit it again on
                                     # 2026-09-05; see the second bullet under
                                     # Evidence and the second report below
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md
reports:                             # every run whose findings this issue rests on
  - docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md
  - docs/retros/2026-09-05-59ic-spec-corpus-anchor-references.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

`swarm dispatch accept` rebases the dispatch branch onto the sprint head before
merging it, and it runs that rebase from outside the **worktree the command
itself created** — while that worktree still has the branch checked out. Git
refuses to rebase a branch checked out elsewhere, so the accept exits 2 and the
dispatch cannot be accepted. The failure needs the sprint head to have moved
since the dispatch branched, which happens exactly when a **second** dispatch
was accepted first — that is, only under real concurrency, the thing the whole
worktree design exists to enable. A run with one dispatch at a time never sees
it; a run with two sees the first accept succeed and the second dead-end. The
cost is that the accept step of a concurrent run stops on its own tooling, and
the operator has to reach into the worktree and finish the rebase by hand
before the command will do the job it was invoked for — with the run's beads
closed, their commits unmerged and a worktree standing, which is the state the
accept step exists to clear.

## Evidence

- `d7w5.2.1#12` — "`swarm dispatch accept` cannot accept the SECOND of two
  concurrent dispatches: it tries to rebase the dispatch branch while that
  branch is still checked out in the worktree the command itself created.
  Measured, accepting dev-d7w5-2 after dev-d7w5-3 had advanced the sprint head
  to b9b9510: exit 2, 'rebasing bb/dev-d7w5-2 onto b9b9510 failed: plan: step
  \"rebase\": git: git rebase exited 128: fatal: 'bb/dev-d7w5-2' is already
  used by worktree at /home/user/.bb-worktrees/dev-d7w5-2'. The first dispatch
  accepted cleanly because the sprint head had not moved, so this reproduces
  only under actual concurrency — the thing the whole worktree design exists to
  enable. Workaround: run `git rebase <sprint-head>` inside the worktree, then
  re-run accept, which then reports case: serialized and succeeds. The command
  should rebase inside the worktree, or remove the worktree before rebasing,
  since it owns both."

- **Recurrence 1 — run `agent-plugins-59ic`, 2026-09-05, same defect and same
  workaround.** Recorded by that run's coordinator as an `amend:` comment on
  run epic `agent-plugins-59ic`: "`swarm dispatch accept` DEFECT 2 — its
  concurrent path runs `git rebase <target> <dispatch-branch>` while that
  branch is still checked out in its own worktree, which git refuses:
  \"fatal: 'bb/dev-59ic-c' is already used by worktree\". The rebase must run
  inside the worktree, or the worktree be removed first. Worked around by
  rebasing from inside /home/user/.bb-worktrees/dev-59ic-c, after which accept
  saw the serialized case and completed." Same failure text, same manual
  recovery, a different run and a different dispatch — so this is a reproducing
  defect rather than one run's bad afternoon, and the fix below is unchanged by
  it. Recorded as an `amend:` comment rather than a `retro:` finding, which is
  itself filed as
  `issue-coordinator-tooling-findings-are-invisible-to-the-harvest`; it
  therefore carries no finding id.

- The 2026-09-04 run's coordinator recorded the two accepts side by side on the run
  epic `agent-plugins-d7w5`, as its `PASS 2 ACCEPTANCE EVIDENCE` comment:
  "`swarm dispatch accept --no-push`: dev-d7w5-3: exit 0, merged true, ancestor
  true, worktree_removed true, branch_deleted true, beads_checked 1,
  acceptance_failures []. dev-d7w5-2: FIRST ATTEMPT exit 2 — 'rebasing
  bb/dev-d7w5-2 onto b9b9510 failed: git rebase exited 128: fatal:
  bb/dev-d7w5-2 is already used by worktree at
  /home/user/.bb-worktrees/dev-d7w5-2'. After `git rebase b9b9510` inside that
  worktree (exit 0) and `bb-checks.py` there (exit 0), the retry gave exit 0,
  merged true, ancestor true, worktree_removed true, branch_deleted true,
  beads_checked 3, acceptance_failures []." The retry reported `case:
  serialized`, so the serialization path itself is sound — only the working
  directory the rebase runs in is wrong.

## Proposed change

The rebase step of `swarm dispatch accept` stops operating on a branch it has
left checked out somewhere else. Two shapes are available and this issue does
not pick between them, because the command owns both ends and either closes the
defect:

- run the rebase **inside the dispatch worktree** (`git -C <worktree> rebase
  <sprint-head>`), which is where the branch is checked out; or
- **remove the worktree first** and rebase the branch from the sprint clone,
  since the accept already removes that worktree on success
  (`worktree_removed true`).

Whichever is taken, the checks the accept re-runs after a rebase must run
against the rebased tree, not the pre-rebase one — the manual recovery on this
run re-ran `bb-checks.py` inside the worktree after rebasing, and the command
must not skip that because it changed working directory.

Done is: with two dispatches accepted in sequence where the first advanced the
sprint head, the second `swarm dispatch accept` exits 0 on its **first**
attempt, reports `case: serialized`, and leaves `merged true`, `ancestor true`,
`worktree_removed true`, `branch_deleted true` — with no `git rebase exited 128
… already used by worktree` in its output.
